package adapter

import (
	"context"
	"fmt"
)

func init() {
	// Register NVIDIA as a provider for CUDA-Q / cuQuantum workloads.
	// These typically run on GPU nodes in-cluster, but can also target
	// cloud QPUs via the CUDA-Q → Braket integration.
	Register("nvidia", func(creds map[string]string) (Provider, error) {
		return NewNvidiaAdapter(creds)
	})
	Register("nvidia-cudaq", func(creds map[string]string) (Provider, error) {
		return NewNvidiaAdapter(creds)
	})
	Register("nvidia-cuquantum", func(creds map[string]string) (Provider, error) {
		return NewNvidiaAdapter(creds)
	})
	Register("nvidia-nvqlink", func(creds map[string]string) (Provider, error) {
		return NewNvidiaAdapter(creds)
	})
}

// NvidiaAdapter handles NVIDIA CUDA-Q, cuQuantum, CUDA-QX, and NVQLink workloads.
//
// Unlike cloud providers (Braket, Azure), NVIDIA workloads typically run
// IN-CLUSTER on GPU nodes. The adapter:
//
//  1. CUDA-Q jobs: Creates a K8s Job/Pod with the CUDA-Q container
//     (nvcr.io/nvidia/cuda-quantum) + GPU resources. The job runs the
//     cudaq.sample() / cudaq.observe() kernel and writes results to a
//     shared volume or object storage.
//
//  2. cuQuantum jobs: Creates a GPU Pod with the cuQuantum container.
//     Supports multi-GPU / multi-node distributed simulation.
//
//  3. NVQLink jobs: References an existing NVQLinkLogicalQPU CRD.
//     The QPU + GPU are already provisioned; the adapter submits the
//     real-time QEC workload to the Logical QPU.
//
//  4. CUDA-Q → Braket: When the CUDA-Q target is "braket", the adapter
//     delegates to the BraketAdapter for QPU submission, while the
//     classical component runs on GPU in-cluster.
type NvidiaAdapter struct {
	// In-cluster: the adapter creates K8s Jobs with GPU resources.
	// Cloud: delegates to BraketAdapter for QPU access via CUDA-Q.
	region    string
	project   string
	ecrRepo   string // ECR repo for CUDA-Q hybrid job images
	inCluster bool
}

func NewNvidiaAdapter(creds map[string]string) (*NvidiaAdapter, error) {
	a := &NvidiaAdapter{
		region:    creds["region"],
		project:   creds["project"],
		ecrRepo:   creds["ecrRepo"],
		inCluster: true, // default: in-cluster GPU workloads
	}
	// If Braket credentials are present, we can also target cloud QPUs.
	if creds["braketRegion"] != "" {
		a.region = creds["braketRegion"]
	}
	return a, nil
}

func (n *NvidiaAdapter) Name() string { return "nvidia" }

func (n *NvidiaAdapter) SubmitJob(ctx context.Context, spec JobSpec) (*JobHandle, error) {
	// For in-cluster CUDA-Q / cuQuantum workloads:
	// 1. Create a K8s Job with:
	//    - image: nvcr.io/nvidia/cuda-quantum:latest (or cuQuantum)
	//    - GPU resources: nvidia.com/gpu: <count>
	//    - env: CUDAQ_TARGET, CUDAQ_DEVICE, etc.
	//    - initContainer: transpiler sidecar (if circuit format conversion needed)
	//    - volume: result output to PVC or S3 via s3fs
	//
	// 2. For NVQLink workloads:
	//    - Reference the NVQLinkLogicalQPU service
	//    - Submit via cudaq::device_call to the real-time GPU host
	//    - Stream QEC syndromes via RoCE/RDMA
	//
	// 3. For CUDA-Q → Braket (cloud QPU):
	//    - Package the CUDA-Q kernel as a container image
	//    - Submit via BraketAdapter.SubmitJob()
	//    - Classical component on GPU, QPU via Braket
	//
	// This is a stub; production implementation creates the K8s Job.
	return &JobHandle{
		ProviderJobID: "nvidia-pending",
		Provider:      n.Name(),
	}, nil
}

func (n *NvidiaAdapter) GetStatus(ctx context.Context, handle *JobHandle) (*JobStatus, error) {
	// For in-cluster: read the K8s Job status.
	// For NVQLink: query the Logical QPU service.
	// For Braket-delegated: delegate to BraketAdapter.
	return &JobStatus{Phase: PhasePending, Message: "nvidia: status not yet implemented"}, nil
}

func (n *NvidiaAdapter) CancelJob(ctx context.Context, handle *JobHandle) error {
	// Delete the K8s Job (in-cluster) or cancel via Braket (cloud).
	return nil
}

func (n *NvidiaAdapter) GetResults(ctx context.Context, handle *JobHandle) (*Results, error) {
	// Read from the result volume / S3 / MinIO.
	// CUDA-Q results: measurement probabilities, expectation values.
	// cuQuantum results: statevector, density matrix, tensor network.
	// NVQLink results: QEC syndrome log, logical error rate, correction ops.
	return &Results{Counts: map[string]int{}}, nil
}

func (n *NvidiaAdapter) ListDevices(ctx context.Context) ([]Device, error) {
	// In-cluster: list available GPU nodes with nvidia.com/gpu resources.
	// Cloud: list Braket devices accessible via CUDA-Q.
	return []Device{
		{
			ID:        "nvidia",
			Name:      "NVIDIA CUDA-Q Simulator (GPU)",
			Provider:  "nvidia",
			Simulator: true,
			Available: true,
		},
		{
			ID:        "nvidia/multi",
			Name:      "NVIDIA CUDA-Q Multi-GPU Simulator",
			Provider:  "nvidia",
			Simulator: true,
			Available: true,
		},
		{
			ID:        "cuquantum-statevec",
			Name:      "cuQuantum cuStateVec (GPU)",
			Provider:  "nvidia-cuquantum",
			Simulator: true,
			Available: true,
		},
		{
			ID:        "cuquantum-stabilizer",
			Name:      "cuQuantum cuStabilizer (GPU)",
			Provider:  "nvidia-cuquantum",
			Simulator: true,
			Available: true,
		},
		{
			ID:        "cuquantum-tensornet",
			Name:      "cuQuantum cuTensorNet (GPU)",
			Provider:  "nvidia-cuquantum",
			Simulator: true,
			Available: true,
		},
	}, nil
}

func (n *NvidiaAdapter) HealthCheck(ctx context.Context) error {
	// Check that GPU nodes are available and the CUDA-Q container can pull.
	return nil
}
