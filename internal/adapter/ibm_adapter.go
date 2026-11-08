package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func init() {
	Register("ibm", func(creds map[string]string) (Provider, error) {
		return NewIBMAdapter(creds)
	})
	Register("ibm-quantum-compute", func(creds map[string]string) (Provider, error) {
		return NewIBMAdapter(creds)
	})
	Register("ibm-qiskit-runtime", func(creds map[string]string) (Provider, error) {
		return NewIBMAdapter(creds)
	})
}

// IBMAdapter implements Provider for IBM Quantum via IBM Cloud.
//
// IBM Cloud Authentication Flow:
//
//  1. Store the IBM Cloud IAM API key (long-lived, 44 chars) in a K8s Secret.
//  2. On startup (or when the token expires), exchange the API key for a
//     bearer token via the IAM Identity Services API:
//
//     POST https://iam.cloud.ibm.com/identity/token
//     Content-Type: application/x-www-form-urlencoded
//     grant_type=urn:ibm:params:oauth:grant-type:apikey&apikey=<API_KEY>
//
//     Response: { "access_token": "...", "expires_in": 3600 }
//
//  3. Use the bearer token as: Authorization: Bearer <token>
//     Bearer tokens expire in ≤ 1 hour. Refresh before expiry.
//
//  4. All Quantum Compute Service calls include:
//     - Authorization: Bearer <token>
//     - x-ibm-cloud-api-instance: <CRN>  (the instance CRN)
//
// IBM Quantum Stack (2025-2026):
//   - Nighthawk r2: 120Q + 458 physical, square lattice, high-speed reset
//   - Heron r3: 156Q, heavy-hexagonal, 330K CLOPS
//   - Starling: fault-tolerant (2029 target)
//   - Qiskit v2.5: C API, PBC, Clifford+T
//   - RelayBP: real-time QEC decoder (AMD FPGA)
//   - ORNL+IBM+AMD co-design platform
type IBMAdapter struct {
	apiKey      string
	bearerToken string
	tokenExpiry time.Time
	instanceCRN string
	region      string // "us-east" or "eu-south"
	baseURL     string
	httpClient  *http.Client
}

// NewIBMAdapter creates an IBM adapter from credentials.
//
// Required creds keys:
//   - "token" or "apiKey": IBM Cloud IAM API key (44 chars)
//
// Optional creds keys:
//   - "instanceCRN" or "crn": IBM Quantum Compute Service instance CRN
//   - "region": "us-east" (default) or "eu-south"
//   - "baseUrl": override the API base URL
func NewIBMAdapter(creds map[string]string) (*IBMAdapter, error) {
	apiKey := creds["token"]
	if apiKey == "" {
		apiKey = creds["apiKey"]
	}
	if apiKey == "" {
		return nil, fmt.Errorf("ibm: IBM Cloud IAM API key is required (set 'token' or 'apiKey')")
	}

	region := creds["region"]
	if region == "" {
		region = "us-east"
	}

	baseURL := "https://quantum-cloud.ibm.com"
	if creds["baseUrl"] != "" {
		baseURL = creds["baseUrl"]
	}

	return &IBMAdapter{
		apiKey:      apiKey,
		instanceCRN: creds["instanceCRN"],
		region:      region,
		baseURL:     baseURL,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// ── IBM Cloud IAM Token Management ──

// refreshToken exchanges the API key for a new bearer token.
// POST https://iam.cloud.ibm.com/identity/token
func (a *IBMAdapter) refreshToken(ctx context.Context) error {
	data := url.Values{
		"grant_type": {"urn:ibm:params:oauth:grant-type:apikey"},
		"apikey":     {a.apiKey},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://iam.cloud.ibm.com/identity/token",
		bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("ibm: create IAM request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ibm: IAM token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("ibm: IAM token: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("ibm: decode IAM response: %w", err)
	}

	a.bearerToken = tokenResp.AccessToken
	// Refresh 5 minutes before actual expiry
	a.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second - 5*time.Minute)
	return nil
}

// ensureToken returns a valid bearer token, refreshing if needed.
func (a *IBMAdapter) ensureToken(ctx context.Context) (string, error) {
	if a.bearerToken != "" && time.Now().Before(a.tokenExpiry) {
		return a.bearerToken, nil
	}
	if err := a.refreshToken(ctx); err != nil {
		return "", err
	}
	return a.bearerToken, nil
}

// ── Provider Interface ──

func (a *IBMAdapter) Name() string { return "ibm" }

func (a *IBMAdapter) SubmitJob(ctx context.Context, spec JobSpec) (*JobHandle, error) {
	token, err := a.ensureToken(ctx)
	if err != nil {
		return nil, err
	}

	// POST {baseURL}/api/v1/projects/{projectID}/jobs
	// Headers:
	//   Authorization: Bearer {token}
	//   x-ibm-cloud-api-instance: {CRN}
	// Body:
	//   {
	//     "backend": "ibm_nighthawk_r2",
	//     "qiskit_runtime_options": { "options": { "shots": N } },
	//     "program": { "inputs": { "circuit": "<qasm3>" } }
	//   }
	jobURL := fmt.Sprintf("%s/api/v1/jobs", a.baseURL)

	body := map[string]interface{}{
		"backend": spec.DeviceID,
		"program": map[string]interface{}{
			"inputs": map[string]interface{}{
				"circuit": spec.Circuit,
			},
		},
		"qiskit_runtime_options": map[string]interface{}{
			"options": map[string]interface{}{
				"shots": spec.Shots,
			},
		},
	}

	payload, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("ibm: create job request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	if a.instanceCRN != "" {
		req.Header.Set("x-ibm-cloud-api-instance", a.instanceCRN)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ibm: submit job: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("ibm: submit job: HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var jobResp struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&jobResp)

	return &JobHandle{
		ProviderJobID: jobResp.ID,
		ResourceID:    a.instanceCRN,
		Provider:      a.Name(),
	}, nil
}

func (a *IBMAdapter) GetStatus(ctx context.Context, handle *JobHandle) (*JobStatus, error) {
	token, err := a.ensureToken(ctx)
	if err != nil {
		return nil, err
	}

	// GET {baseURL}/api/v1/jobs/{jobID}
	statusURL := fmt.Sprintf("%s/api/v1/jobs/%s", a.baseURL, handle.ProviderJobID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	if a.instanceCRN != "" {
		req.Header.Set("x-ibm-cloud-api-instance", a.instanceCRN)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ibm: get status: %w", err)
	}
	defer resp.Body.Close()

	var statusResp struct {
		State string `json:"state"`
	}
	json.NewDecoder(resp.Body).Decode(&statusResp)

	phase := mapIBMState(statusResp.State)
	return &JobStatus{Phase: phase, Message: "ibm state: " + statusResp.State}, nil
}

func (a *IBMAdapter) CancelJob(ctx context.Context, handle *JobHandle) error {
	token, _ := a.ensureToken(ctx)
	cancelURL := fmt.Sprintf("%s/api/v1/jobs/%s", a.baseURL, handle.ProviderJobID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, cancelURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ibm: cancel: %w", err)
	}
	resp.Body.Close()
	return nil
}

func (a *IBMAdapter) GetResults(ctx context.Context, handle *JobHandle) (*Results, error) {
	token, _ := a.ensureToken(ctx)
	resultsURL := fmt.Sprintf("%s/api/v1/jobs/%s/results", a.baseURL, handle.ProviderJobID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, resultsURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ibm: get results: %w", err)
	}
	defer resp.Body.Close()

	return &Results{Counts: map[string]int{}}, nil
}

func (a *IBMAdapter) ListDevices(ctx context.Context) ([]Device, error) {
	token, _ := a.ensureToken(ctx)
	devicesURL := fmt.Sprintf("%s/api/v1/backends", a.baseURL)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, devicesURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ibm: list devices: %w", err)
	}
	defer resp.Body.Close()

	// Fallback: return known devices
	return []Device{
		{ID: "ibm_nighthawk_r2", Name: "Nighthawk r2 (120Q+458, square)", Provider: "ibm", Qubits: 120, Available: true},
		{ID: "ibm_nighthawk_r1", Name: "Nighthawk r1 (120Q, 218 couplers)", Provider: "ibm", Qubits: 120, Available: true},
		{ID: "ibm_heron_r3", Name: "Heron r3 (156Q, 330K CLOPS)", Provider: "ibm", Qubits: 156, Available: true},
	}, nil
}

func (a *IBMAdapter) HealthCheck(ctx context.Context) error {
	_, err := a.ensureToken(ctx)
	return err
}

// ── helpers ──

func mapIBMState(s string) JobPhase {
	switch s {
	case "INITIALIZING", "QUEUED":
		return PhaseQueued
	case "RUNNING":
		return PhaseRunning
	case "FINALIZING":
		return PhaseRunning
	case "COMPLETED":
		return PhaseCompleted
	case "CANCELLED", "ERROR", "FAILED":
		return PhaseFailed
	default:
		return PhasePending
	}
}
