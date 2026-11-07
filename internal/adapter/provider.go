// Package adapter defines the QuantumProvider interface and all cloud/direct
// provider implementations.
package adapter

import (
	"context"
)

// Provider is the interface every quantum backend adapter must implement.
// The operator's controller loop interacts only with this interface;
// it never calls cloud SDKs directly.
type Provider interface {
	// Name returns the provider identifier (e.g. "aws-braket", "azure-quantum", "ionq").
	Name() string

	// SubmitJob submits a quantum job and returns a handle for tracking.
	SubmitJob(ctx context.Context, spec JobSpec) (*JobHandle, error)

	// GetStatus polls the job status from the provider.
	GetStatus(ctx context.Context, handle *JobHandle) (*JobStatus, error)

	// CancelJob cancels a running or queued job.
	CancelJob(ctx context.Context, handle *JobHandle) error

	// GetResults retrieves the measurement results for a completed job.
	GetResults(ctx context.Context, handle *JobHandle) (*Results, error)

	// ListDevices returns available devices for this provider.
	ListDevices(ctx context.Context) ([]Device, error)

	// HealthCheck verifies credentials and connectivity.
	HealthCheck(ctx context.Context) error
}

// JobSpec is the provider-agnostic job description.
type JobSpec struct {
	// Circuit is the OpenQASM 3 source.
	Circuit string `json:"circuit"`

	// DeviceID is the target device (e.g. "sv1", "ionq.simulator").
	DeviceID string `json:"deviceId"`

	// Shots is the number of measurement repetitions.
	Shots int `json:"shots"`

	// Hybrid config (classical component).
	Hybrid *HybridConfig `json:"hybrid,omitempty"`

	// Parameters for parameterized circuits.
	Parameters map[string]string `json:"parameters,omitempty"`

	// ResultBackend: s3, azure-blob, minio.
	ResultBackend string `json:"resultBackend,omitempty"`

	// ResultBucket.
	ResultBucket string `json:"resultBucket,omitempty"`
}

// HybridConfig describes the classical component of a hybrid job.
type HybridConfig struct {
	Image      string            `json:"image"`
	Entrypoint []string          `json:"entrypoint,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	Resources  map[string]string `json:"resources,omitempty"`
}

// JobHandle is an opaque reference to a submitted job.
type JobHandle struct {
	// ProviderJobID is the provider's job identifier.
	ProviderJobID string `json:"providerJobId"`

	// ResourceARN (AWS) or resourceID (Azure).
	ResourceID string `json:"resourceId,omitempty"`

	// Provider name for routing.
	Provider string `json:"provider"`

	// Namespace of the originating QuantumJob.
	Namespace string `json:"namespace"`

	// Name of the originating QuantumJob.
	Name string `json:"name"`
}

// JobPhase represents the lifecycle state of a quantum job.
type JobPhase string

const (
	PhasePending    JobPhase = "Pending"
	PhaseSubmitted  JobPhase = "Submitted"
	PhaseQueued     JobPhase = "Queued"
	PhaseRunning    JobPhase = "Running"
	PhaseCompleted  JobPhase = "Completed"
	PhaseFailed     JobPhase = "Failed"
	PhaseCancelled  JobPhase = "Cancelled"
)

// JobStatus is the provider-agnostic status.
type JobStatus struct {
	Phase         JobPhase `json:"phase"`
	Message       string   `json:"message,omitempty"`
	QueuePosition int      `json:"queuePosition,omitempty"`
	CostEstimate  float64  `json:"costEstimate,omitempty"`
}

// Results contains the measurement outcomes.
type Results struct {
	// Counts is a map of bitstring -> count.
	Counts map[string]int `json:"counts"`

	// ExpectationValues for observable measurements.
	// +optional
	ExpectationValues map[string]float64 `json:"expectationValues,omitempty"`

	// RawResultURI is the provider-specific result location.
	RawResultURI string `json:"rawResultUri,omitempty"`

	// ShotsExecuted.
	ShotsExecuted int `json:"shotsExecuted"`
}

// Device describes an available quantum device.
type Device struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Provider    string  `json:"provider"`
	Qubits      int     `json:"qubits"`
	GateSet     []string `json:"gateSet,omitempty"`
	Simulator   bool    `json:"simulator"`
	Available   bool    `json:"available"`
	CostPerUnit float64 `json:"costPerUnit,omitempty"`
	Billing     string  `json:"billing,omitempty"`
}

// Factory creates a Provider by name.
type Factory func(creds map[string]string) (Provider, error)

var registry = map[string]Factory{}

// Register adds a provider factory to the global registry.
func Register(name string, f Factory) {
	registry[name] = f
}

// New creates a Provider from the registry by name.
func New(name string, creds map[string]string) (Provider, error) {
	f, ok := registry[name]
	if !ok {
		return nil, &UnknownProviderError{Name: name}
	}
	return f(creds)
}

// Known returns the list of registered provider names.
func Known() []string {
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	return names
}

// UnknownProviderError is returned when a provider name is not registered.
type UnknownProviderError struct{ Name string }

func (e *UnknownProviderError) Error() string {
	return "unknown quantum provider: " + e.Name
}
