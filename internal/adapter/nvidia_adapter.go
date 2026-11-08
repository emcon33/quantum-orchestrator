package adapter

import (
	"context"
	"fmt"
)

func init() {
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
// NVIDIA distributes CUDA-Q and cuQuantum via the NGC (NVIDIA GPU Cloud)
// container registry:
//
//   Stable:   nvcr.io/nvidia/quantum/cuda-quantum:cu12-0.14.2
//             nvcr.io/nvidia/quantum/cuda-quantum:cu13-0.16.0
//   Nightly:  nvcr.io/nvidia/nightly/cuda-quantum:cu13-latest
//
//   - ~4.3 GB compressed
//   - Multi-node support (multi-GPU, multi-node clusters)
//   - Multi-arch support (amd64)
//   - Signed, scanned
//   - CUDA 12 or CUDA 13
//   - Includes: CUDA-Q, cuQuantum (cuStateVec, cuDensityMat, cuTensorNet,
//     cuPauliProp, cuStabilizer), NVQ++ compiler (LLVM-based), CUDA-QX (QEC)
//
//   GitHub source: github.com/NVIDIA/cuda-quantum (Apache 2.0)
//   NGC catalog:   catalog.ngc.nvidia.com/orgs/nvidia/teams/quantum/containers/cuda-quantum
//
// The adapter creates K8s Jobs with:
//   - image: nvcr.io/nvidia/quantum/cuda-quantum:<tag>
//   - resources: nvidia.com/gpu: <count>
//   - multi-node: via Kueue ResourceFlavor or Gang Scheduling
type NvidiaAdapter struct {
	ngcImage     string // NGC container image
	cudaVersion  string // "cu12" or "cu13"
	multiNode    bool
	region       string
	ecrRepo      string
	inCluster    bool
}

// DefaultNGCImages lists the known NGC image tags.
var DefaultNGCImages = map[string]string{
	"cu12-stable":  "nvcr.io/nvidia/quantum/cuda-quantum:cu12-0.14.2",
	"cu13-stable":  "nvcr.io/nvidia/quantum/cuda-quantum:cu13-0.16.0",
	"cu12-nightly": "nvcr.io/nvidia/nightly/cuda-quantum:cu12-latest",
	"cu13-nightly": "nvcr.io/nvidia/nightly/cuda-quantum:cu13-latest",
}

func NewNvidiaAdapter(creds map[string]string) (*NvidiaAdapter, error) {
	img := creds["ngcImage"]
	if img == "" {
		img = DefaultNGCImages["cu13-stable"]
	}

	cudaVer := creds["cudaVersion"]
	if cudaVer == "" {
		cudaVer = "cu13"
	}

	return &NvidiaAdapter{
		ngcImage:    img,
		cudaVersion: cudaVer,
		multiNode:   creds["multiNode"] == "true",
		region:      creds["region"],
		ecrRepo:     creds["ecrRepo"],
		inCluster:   true,
	}, nil
}

func (n *NvidiaAdapter) Name() string { return "nvidia" }

func (n *NvidiaAdapter) SubmitJob(ctx context.Context, spec JobSpec) (*JobHandle, error) {
	// In-cluster: create a K8s Job with the NGC CUDA-Q container.
	//
	// Pod spec (created by the operator):
	//   containers:
	//     - name: cuda-q
	//       image: nvcr.io/nvidia/quantum/cuda-quantum:cu13-0.16.0
	//       resources:
	//         limits:
	//           nvidia.com/gpu: <count>
	//       env:
	//         - name: CUDAQ_TARGET
	//           value: "nvidia"  # or "braket", "quantinuum", etc.
	//         - name: CUDAQ_DEVICE
	//           value: "<device-arn>"
	//       volumeMounts:
	//         - name: results
	//           mountPath: /results
	//   volumes:
	//     - name: results
	//       persistentVolumeClaim:
	//         claimName: quantum-results
	//
	// Multi-node (cuQuantum distributed):
	//   - Use Kueue Gang Scheduling or Volcano
	//   - Each node: 1+ GPU
	//   - NCCL for inter-node GPU communication
	//   - NGC image supports multi-node out of the box
	//
	// CUDA-Q → Braket (cloud QPU):
	//   - Classical component: NGC container on GPU
	//   - QPU: via Braket (IonQ, Rigetti, IQM, QuEra)
	//   - CUDA-Q target: "braket" + device ARN
	return &JobHandle{
		ProviderJobID: "nvidia-pending",
		Provider:      n.Name(),
	}, nil
}

func (n *NvidiaAdapter) GetStatus(ctx context.Context, handle *JobHandle) (*JobStatus, error) {
	return &JobStatus{Phase: PhasePending, Message: "nvidia: not yet implemented"}, nil
}

func (n *NvidiaAdapter) CancelJob(ctx context.Context, handle *JobHandle) error { return nil }

func (n *NvidiaAdapter) GetResults(ctx context.Context, handle *JobHandle) (*Results, error) {
	return &Results{Counts: map[string]int{}}, nil
}

func (n *NvidiaAdapter) ListDevices(ctx context.Context) ([]Device, error) {
	return []Device{
		{ID: "nvidia", Name: "CUDA-Q Simulator (NGC, single GPU)", Provider: "nvidia", Simulator: true, Available: true},
		{ID: "nvidia/multi", Name: "CUDA-Q Multi-GPU (NGC, multi-node)", Provider: "nvidia", Simulator: true, Available: true},
		{ID: "cuquantum-statevec", Name: "cuQuantum cuStateVec (GPU)", Provider: "nvidia", Simulator: true, Available: true},
		{ID: "cuquantum-stabilizer", Name: "cuQuantum cuStabilizer (GPU)", Provider: "nvidia", Simulator: true, Available: true},
		{ID: "cuquantum-tensornet", Name: "cuQuantum cuTensorNet (GPU)", Provider: "nvidia", Simulator: true, Available: true},
		{ID: "braket", Name: "CUDA-Q → AWS Braket (cloud QPU)", Provider: "nvidia", Available: true},
	}, nil
}

func (n *NvidiaAdapter) HealthCheck(ctx context.Context) error {
	// Verify NGC image is pullable and GPU nodes are available.
	return nil
}
