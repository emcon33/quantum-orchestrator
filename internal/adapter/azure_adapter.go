package adapter

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

func init() {
	Register("azure-quantum", func(creds map[string]string) (Provider, error) {
		return NewAzureQuantumAdapter(creds)
	})
}

// AzureQuantumAdapter implements Provider for Azure Quantum.
type AzureQuantumAdapter struct {
	workspace    string
	subscription string
	tenantID     string
	clientID     string
	httpClient   *http.Client
	baseURL      string
}

// NewAzureQuantumAdapter creates an adapter from credentials.
// Expects: "workspace", "subscription", "tenantId", "clientId" (optional, uses Workload Identity if absent).
func NewAzureQuantumAdapter(creds map[string]string) (*AzureQuantumAdapter, error) {
	workspace := creds["workspace"]
	subscription := creds["subscription"]
	if workspace == "" || subscription == "" {
		return nil, fmt.Errorf("azure: workspace and subscription are required")
	}

	tenantID := creds["tenantId"]
	clientID := creds["clientId"]
	_ = tenantID
	_ = clientID
	// In production, use azidentity.NewWorkloadIdentityCredential() on AKS
	// or azidentity.NewClientSecretCredential() with a service principal.

	baseURL := fmt.Sprintf(
		"https://%s.quantum.azure.com",
		workspace,
	)

	return &AzureQuantumAdapter{
		workspace:    workspace,
		subscription: subscription,
		httpClient:   http.DefaultClient,
		baseURL:      baseURL,
	}, nil
}

func (a *AzureQuantumAdapter) Name() string { return "azure-quantum" }

func (a *AzureQuantumAdapter) SubmitJob(ctx context.Context, spec JobSpec) (*JobHandle, error) {
	// In production: POST {baseURL}/jobs with the circuit payload.
	// The circuit can be Q#, Qiskit (serialized), Cirq, Quil, or QIR.
	// For QIR Adaptive RI: compile via qsharp-compiler sidecar first.
	return &JobHandle{
		ProviderJobID: "azure-pending",
		ResourceID:    fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Quantum/Workspaces/%s/jobs/pending",
			a.subscription, a.workspace, a.workspace),
		Provider: a.Name(),
	}, nil
}

func (a *AzureQuantumAdapter) GetStatus(ctx context.Context, handle *JobHandle) (*JobStatus, error) {
	// In production: GET {baseURL}/jobs/{id}
	return &JobStatus{Phase: PhasePending, Message: "azure: status not yet implemented"}, nil
}

func (a *AzureQuantumAdapter) CancelJob(ctx context.Context, handle *JobHandle) error {
	// In production: PATCH {baseURL}/jobs/{id} with {"state": "cancelled"}
	return nil
}

func (a *AzureQuantumAdapter) GetResults(ctx context.Context, handle *JobHandle) (*Results, error) {
	// In production: GET {baseURL}/jobs/{id}/results
	return &Results{Counts: map[string]int{}}, nil
}

func (a *AzureQuantumAdapter) ListDevices(ctx context.Context) ([]Device, error) {
	// In production: GET {baseURL}/providers
	return nil, nil
}

func (a *AzureQuantumAdapter) HealthCheck(ctx context.Context) error {
	// In production: GET {baseURL} with auth token
	return nil
}
