package adapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/braket"
	brakettypes "github.com/aws/aws-sdk-go-v2/service/braket/types"
)

func init() {
	Register("aws-braket", func(creds map[string]string) (Provider, error) {
		return NewBraketAdapter(creds)
	})
}

// BraketAdapter implements Provider for Amazon Braket.
type BraketAdapter struct {
	client   *braket.Client
	region   string
	project  string
}

// NewBraketAdapter creates a BraketAdapter from credentials.
// Expects creds keys: "region" (optional, default us-east-1), "project" (optional).
func NewBraketAdapter(creds map[string]string) (*BraketAdapter, error) {
	region := creds["region"]
	if region == "" {
		region = "us-east-1"
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("braket: load AWS config: %w", err)
	}

	client := braket.NewFromConfig(cfg)

	return &BraketAdapter{
		client:   client,
		region:   region,
		project:  creds["project"],
	}, nil
}

func (a *BraketAdapter) Name() string { return "aws-braket" }

func (a *BraketAdapter) SubmitJob(ctx context.Context, spec JobSpec) (*JobHandle, error) {
	input := &braket.CreateQuantumTaskInput{
		DeviceArn:    a.deviceARN(spec.DeviceID),
		Shots:        int32(spec.Shots),
		Parallelism:  brakettypes.ParallelismLevelSingle,
	}

	// Build the OpenQASM task specification.
	ov := &braket.OpenQASMTasksSpecification{
		Source: &spec.Circuit,
	}
	input.TaskSpec = &braket.TaskSpecification{
		OpenQASMTasksSpecification: ov,
	}

	// Result destination.
	if spec.ResultBucket != "" {
		s3Location := &braket.S3DataSource{
			BucketName: &spec.ResultBucket,
		}
		input.OutputS3Uri = &spec.ResultBucket
	}

	output, err := a.client.CreateQuantumTask(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("braket: create quantum task: %w", err)
	}

	return &JobHandle{
		ProviderJobID: *output.QuantumTaskArn,
		ResourceID:    *output.QuantumTaskArn,
		Provider:      a.Name(),
	}, nil
}

func (a *BraketAdapter) GetStatus(ctx context.Context, handle *JobHandle) (*JobStatus, error) {
	output, err := a.client.GetQuantumTask(ctx, &braket.GetQuantumTaskInput{
		QuantumTaskArn: &handle.ProviderJobID,
	})
	if err != nil {
		return nil, fmt.Errorf("braket: get task status: %w", err)
	}

	phase, _ := mapBraketStatus(output.Status)
	return &JobStatus{
		Phase:   phase,
		Message: fmt.Sprintf("braket task %s", output.Status),
	}, nil
}

func (a *BraketAdapter) CancelJob(ctx context.Context, handle *JobHandle) error {
	_, err := a.client.UpdateQuantumTask(ctx, &braket.UpdateQuantumTaskInput{
		QuantumTaskArn: &handle.ProviderJobID,
		NewStatus:      brakettypes.QuantumTaskStatusCancelled,
	})
	if err != nil {
		return fmt.Errorf("braket: cancel task: %w", err)
	}
	return nil
}

func (a *BraketAdapter) GetResults(ctx context.Context, handle *JobHandle) (*Results, error) {
	// In production, read from S3. For now, return a placeholder.
	return &Results{
		Counts:       map[string]int{},
		ShotsExecuted: 0,
	}, nil
}

func (a *BraketAdapter) ListDevices(ctx context.Context) ([]Device, error) {
	output, err := a.client.SearchDevices(ctx, &braket.SearchDevicesInput{})
	if err != nil {
		return nil, fmt.Errorf("braket: list devices: %w", err)
	}

	var devices []Device
	for _, d := range output.Results {
		if d == nil {
			continue
		}
		devices = append(devices, Device{
			ID:        *d.DeviceArn,
			Name:      valueOrEmpty(d.DeviceName),
			Provider:  a.Name(),
			Simulator: strings.Contains(*d.DeviceArn, "sv") || strings.Contains(*d.DeviceArn, "dm") || strings.Contains(*d.DeviceArn, "tn"),
			Available: true,
		})
	}
	return devices, nil
}

func (a *BraketAdapter) HealthCheck(ctx context.Context) error {
	_, err := a.client.GetDevice(ctx, &braket.GetDeviceInput{
		DeviceArn: aws.String("arn:aws:braket:" + a.region + ":000000000000:device/quantum-simulator/SV1"),
	})
	return err
}

// --- helpers ---

func (a *BraketAdapter) deviceARN(deviceID string) string {
	// Simplified: in production, resolve via SearchDevices or a QPUBackend CRD.
	return "arn:aws:braket:" + a.region + ":000000000000:device/" + deviceID
}

func mapBraketStatus(s brakettypes.QuantumTaskStatus) (JobPhase, error) {
	switch s {
	case brakettypes.QuantumTaskStatusCREATED, brakettypes.QuantumTaskStatusENVIRONMENT_SETUP:
		return PhasePending, nil
	case brakettypes.QuantumTaskStatusENVIRONMENT_READY, brakettypes.QuantumTaskStatusRUNNING:
		return PhaseRunning, nil
	case brakettypes.QuantumTaskStatusCOMPLETED:
		return PhaseCompleted, nil
	case brakettypes.QuantumTaskStatusFAILED, brakettypes.QuantumTaskStatusCANCELLED:
		return PhaseFailed, nil
	default:
		return PhaseQueued, nil
	}
}

func valueOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func aws(v string) *string { return &v }
