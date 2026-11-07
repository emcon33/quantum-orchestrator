package adapter

import (
	"context"
	"fmt"
)

func init() {
	Register("ionq", func(creds map[string]string) (Provider, error) {
		return NewDirectAdapter("ionq", creds)
	})
	Register("rigetti", func(creds map[string]string) (Provider, error) {
		return NewDirectAdapter("rigetti", creds)
	})
	Register("quantinuum", func(creds map[string]string) (Provider, error) {
		return NewDirectAdapter("quantinuum", creds)
	})
	Register("iqm", func(creds map[string]string) (Provider, error) {
		return NewDirectAdapter("iqm", creds)
	})
}

// DirectAdapter is a thin REST client for a single direct QPU provider.
type DirectAdapter struct {
	provider string
	apiKey   string
	baseURL  string
}

func NewDirectAdapter(provider string, creds map[string]string) (*DirectAdapter, error) {
	apiKey := creds["apiKey"]
	if apiKey == "" {
		return nil, fmt.Errorf("%s: apiKey is required", provider)
	}

	baseURL := directBaseURL(provider)
	return &DirectAdapter{provider: provider, apiKey: apiKey, baseURL: baseURL}, nil
}

func (d *DirectAdapter) Name() string { return d.provider }

func (d *DirectAdapter) SubmitJob(ctx context.Context, spec JobSpec) (*JobHandle, error) {
	// POST {baseURL}/api/v2/jobs with OpenQASM or provider-native format
	return &JobHandle{ProviderJobID: "direct-pending", Provider: d.provider}, nil
}

func (d *DirectAdapter) GetStatus(ctx context.Context, handle *JobHandle) (*JobStatus, error) {
	return &JobStatus{Phase: PhasePending, Message: d.provider + ": not yet implemented"}, nil
}

func (d *DirectAdapter) CancelJob(ctx context.Context, handle *JobHandle) error { return nil }

func (d *DirectAdapter) GetResults(ctx context.Context, handle *JobHandle) (*Results, error) {
	return &Results{Counts: map[string]int{}}, nil
}

func (d *DirectAdapter) ListDevices(ctx context.Context) ([]Device, error) { return nil, nil }

func (d *DirectAdapter) HealthCheck(ctx context.Context) error { return nil }

func directBaseURL(provider string) string {
	switch provider {
	case "ionq":
		return "https://api.ionq.com"
	case "rigetti":
		return "https://api.rigetti.com"
	case "quantinuum":
		return "https://api.quantinuum.com"
	case "iqm":
		return "https://api.iqm.technology"
	default:
		return "https://api." + provider + ".com"
	}
}
