package v1alpha1

// This file extends QuantumJobSpec with IBM-specific fields
// for Nighthawk, Heron, Starling, Qiskit v2.5, RelayBP, and PBC.

// IBMSpec holds IBM-specific configuration for a QuantumJob.
// Embedded in QuantumJobSpec via the `ibm` field.
type IBMSpec struct {
	// Processor family: nighthawk, heron, starling.
	// +kubebuilder:default=heron
	// +optional
	Processor string `json:"processor,omitempty"`

	// Processor revision: r1, r2, r3.
	// +optional
	Revision string `json:"revision,omitempty"`

	// Qiskit configuration.
	// +optional
	Qiskit *QiskitConfig `json:"qiskit,omitempty"`

	// IBM QEC configuration (RelayBP, logical qubits).
	// +optional
	QEC *IBMQECConfig `json:"qec,omitempty"`

	// HPC co-design platform (ORNL + IBM + AMD pattern).
	// +optional
	HPCCoDesign *HPCCoDesignConfig `json:"hpCoDesign,omitempty"`

	// Error mitigation strategy.
	// +optional
	ErrorMitigation *ErrorMitigationConfig `json:"errorMitigation,omitempty"`
}

type QiskitConfig struct {
	// Qiskit version (e.g. "2.5").
	// +optional
	Version string `json:"version,omitempty"`

	// Compilation mode: standard, pbc (Pauli-based computation), clifford-t.
	// +kubebuilder:default=standard
	// +optional
	CompilationMode string `json:"compilationMode,omitempty"`

	// Use Qiskit C API for transpilation (v2.5+).
	// +optional
	UseCAPI bool `json:"useCAPI,omitempty"`

	// Dynamic circuits enabled.
	// +kubebuilder:default=true
	// +optional
	DynamicCircuits bool `json:"dynamicCircuits,omitempty"`

	// Qiskit kernel / program source (Python).
	// +optional
	Program *ProgramRef `json:"program,omitempty"`

	// Preset pass manager: v1, v2, pbc, clifford-t.
	// +optional
	PresetPassManager string `json:"presetPassManager,omitempty"`
}

type ProgramRef struct {
	// Inline Python source (Qiskit program).
	// +optional
	Inline string `json:"inline,omitempty"`

	// Reference to a ConfigMap/Secret.
	// +optional
	Ref *ObjectRef `json:"ref,omitempty"`
}

type IBMQECConfig struct {
	// Decoder: relaybp, bp-osd, union-find, custom.
	// +kubebuilder:default=relaybp
	// +optional
	Decoder string `json:"decoder,omitempty"`

	// Decoder hardware: fpga (AMD), gpu, cpu.
	// +kubebuilder:default=fpga
	// +optional
	DecoderHardware string `json:"decoderHardware,omitempty"`

	// Code: surface, color, ldpc, concatenated, space-time, custom.
	// +optional
	Code string `json:"code,omitempty"`

	// Code distance.
	// +optional
	Distance int32 `json:"distance,omitempty"`

	// Number of logical qubits.
	// +optional
	LogicalQubits int32 `json:"logicalQubits,omitempty"`

	// Number of hard operations (circuits through logical qubits).
	// +optional
	HardOperations int32 `json:"hardOperations,omitempty"`

	// Real-time decoding: true = streaming via FPGA.
	// +optional
	RealTime bool `json:"realTime,omitempty"`

	// Algorithm-aware decoding.
	// +optional
	AlgorithmAware bool `json:"algorithmAware,omitempty"`

	// Hardware-aware compilation.
	// +optional
	HardwareAwareCompilation bool `json:"hardwareAwareCompilation,omitempty"`
}

// HPCCoDesignConfig for the ORNL + IBM + AMD co-design platform pattern.
// Pairs IBM QPU with AMD GPU/CPU HPC for hybrid QEC workloads.
type HPCCoDesignConfig struct {
	// Enable HPC co-design mode.
	Enabled bool `json:"enabled"`

	// GPU model: mi355x (AMD Instinct), a100, h100.
	// +optional
	GPU string `json:"gpu,omitempty"`

	// CPU: epyc-5th-gen, xeon.
	// +optional
	CPU string `json:"cpu,omitempty"`

	// Number of GPU nodes.
	// +kubebuilder:default=1
	// +optional
	GPUNodes int32 `json:"gpuNodes,omitempty"`

	// Low-latency interconnect: roce, nvlink, custom.
	// +optional
	Interconnect string `json:"interconnect,omitempty"`

	// Target latency (microseconds).
	// +optional
	TargetLatencyUs int32 `json:"targetLatencyUs,omitempty"`

	// Workload profile: qec-decoding, calibration, algorithm-development.
	// +optional
	Profile string `json:"profile,omitempty"`
}

type ErrorMitigationConfig struct {
	// Strategy: zero-noise, probabilistic-error-mitigation,
	// readout-mitigation, dynamic-circuits, none.
	// +kubebuilder:default=zero-noise
	// +optional
	Strategy string `json:"strategy,omitempty"`

	// HPC-powered mitigation (100x cost reduction).
	// +optional
	HPCPowered bool `json:"hpcPowered,omitempty"`

	// Readout error mitigation.
	// +optional
	ReadoutMitigation bool `json:"readoutMitigation,omitempty"`

	// Zero-noise extrapolation orders.
	// +optional
	ExtrapolationOrders []int32 `json:"extrapolationOrders,omitempty"`
}
