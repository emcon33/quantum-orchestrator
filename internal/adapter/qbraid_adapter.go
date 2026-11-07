package adapter

import (
	"context"
	"fmt"
)

func init() {
	Register("qbraid", func(creds map[string]string) (Provider, error) {
		return NewQbraidAdapter(creds)
	})
}

// QbraidAdapter wraps the qBraid SDK as a universal runtime.
// It can be run as a sidecar container with the qBraid Python SDK,
// or as a gRPC service. This Go client calls the sidecar over HTTP.
type QbraidAdapter struct {
	apiKey  string
	baseURL string
}

func NewQbraidAdapter(creds map[string]string) (*QbraidAdapter, error) {
	apiKey := creds["apiKey"]
	if apiKey == "" {
		return nil, fmt.Errorf("qbraid: apiKey is required")
	}
	return &QbraidAdapter{
		apiKey:  apiKey,
		baseURL: creds["baseUrl"], // sidecar URL, e.g. http://transpiler.quantum-system.svc:8080
	}, nil
}

func (q *QbraidAdapter) Name() string { return "qbraid" }

func (q *QbraidAdapter) SubmitJob(ctx context.Context, spec JobSpec) (*JobHandle, error) {
	// POST to the qBraid sidecar: /v1/jobs
	// The sidecar handles transpilation + provider routing internally.
	return &JobHandle{ProviderJobID: "qbraid-pending", Provider: q.Name()}, nil
}

func (q *QbraidAdapter) GetStatus(ctx context.Context, handle *JobHandle) (*JobStatus, error) {
	return &JobStatus{Phase: PhasePending, Message: "qbraid: not yet implemented"}, nil
}

func (q *QbraidAdapter) CancelJob(ctx context.Context, handle *JobHandle) error { return nil }

func (q *QbraidAdapter) GetResults(ctx context.Context, handle *JobHandle) (*Results, error) {
	return &Results{Counts: map[string]int{}}, nil
}

func (q *QbraidAdapter) ListDevices(ctx context.Context) ([]Device, error) {
	// GET /v1/devices from the qBraid sidecar
	return nil, nil
}

func (q *QbraidAdapter) HealthCheck(ctx context.Context) error {
	// GET /v1/health
	return nil
}
