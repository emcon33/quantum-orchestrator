package v1alpha1

// This file extends QuantumJobSpec with NVIDIA-specific fields
// for CUDA-Q, cuQuantum, and NVQLink workloads.

// NVIDIASpec holds NVIDIA-specific configuration for a QuantumJob.
// Embedded in QuantumJobSpec via the `nvidia` field.
type NVIDIASpec struct {
	// SDK: cudaq, cuquantum, cudaqx.
	// +kubebuilder:default=cudaq
	// +optional
	SDK string `json:"sdk,omitempty"`

	// CUDA-Q target device string (e.g. "nvidia", "nvidia/multi", "braket", "quantinuum").
	// +optional
	CudaqTarget string `json:"cudaqTarget,omitempty"`

	// CUDA-Q kernel source (inline C++ or Python).
	// +optional
	CudaqKernel *CudaqKernelRef `json:"cudaqKernel,omitempty"`

	// cuQuantum simulator engine.
	// +optional
	CuQuantum *CuQuantumConfig `json:"cuQuantum,omitempty"`

	// CUDA-QX QEC decoder configuration.
	// +optional
	QEC *QECDecoderConfig `json:"qec,omitempty"`

	// NVQLink Logical QPU reference (for real-time QEC workloads).
	// +optional
	NVQLinkRef *NVQLinkRef `json:"nvqlinkRef,omitempty"`

	// GPU requirements.
	// +optional
	GPU *GPURequirements `json:"gpu,omitempty"`
}

type CudaqKernelRef struct {
	// Inline CUDA-Q kernel (C++ or Python source).
	// +optional
	Inline string `json:"inline,omitempty"`

	// Reference to a ConfigMap/Secret.
	// +optional
	Ref *ObjectRef `json:"ref,omitempty"`

	// Language: cpp, python.
	// +kubebuilder:default=python
	// +optional
	Language string `json:"language,omitempty"`

	// Shots for cudaq.sample().
	// +optional
	Shots int32 `json:"shots,omitempty"`

	// Observables for cudaq.observe().
	// +optional
	Observables []string `json:"observables,omitempty"`
}

type CuQuantumConfig struct {
	// Engine: statevec, densitymat, tensornet, pauliprop, stabilizer.
	// +kubebuilder:default=statevec
	// +optional
	Engine string `json:"engine,omitempty"`

	// Number of GPUs for distributed simulation.
	// +kubebuilder:default=1
	// +optional
	GPUs int32 `json:"gpus,omitempty"`

	// Multi-node: number of nodes.
	// +kubebuilder:default=1
	// +optional
	Nodes int32 `json:"nodes,omitempty"`

	// Framework integration: qiskit, cirq, pennylane, qibo, quesp, quesp, quesp.
	// +optional
	Integration string `json:"integration,omitempty"`

	// Max qubits to simulate.
	// +optional
	MaxQubits int32 `json:"maxQubits,omitempty"`
}

type QECDecoderConfig struct {
	// Decoder: nv-qldpc-decoder, bp-osd, tensor-network, union-find.
	// +kubebuilder:default=nv-qldpc-decoder
	// +optional
	Decoder string `json:"decoder,omitempty"`

	// Code type: surface, color, ldpc, custom.
	// +optional
	Code string `json:"code,omitempty"`

	// Code distance.
	// +optional
	Distance int32 `json:"distance,omitempty"`

	// Decoder iterations.
	// +optional
	Iterations int32 `json:"iterations,omitempty"`

	// History depth.
	// +optional
	HistoryDepth int32 `json:"historyDepth,omitempty"`

	// Real-time mode: true = streaming decode via NVQLink.
	// +optional
	RealTime bool `json:"realTime,omitempty"`

	// Detector error model source.
	// +optional
	DEMSource string `json:"demSource,omitempty"`
}

type NVQLinkRef struct {
	// Name of the NVQLinkLogicalQPU CRD.
	Name string `json:"name"`

	// Namespace (defaults to the job's namespace).
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

type GPURequirements struct {
	// GPU model: a100, h100, h200, b200, gh200, rtx-pro-6000.
	// +optional
	Model string `json:"model,omitempty"`

	// Count.
	// +kubebuilder:default=1
	// +optional
	Count int32 `json:"count,omitempty"`

	// Minimum memory (GiB).
	// +optional
	MinMemoryGiB int32 `json:"minMemoryGiB,omitempty"`

	// CUDA version required.
	// +optional
	CUDAVersion string `json:"cudaVersion,omitempty"`

	// GPU feature requirements (e.g. "tensor-core", "fp64", "nvlink").
	// +optional
	Features []string `json:"features,omitempty"`
}
