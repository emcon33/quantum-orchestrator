package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NVQLinkLogicalQPU defines a Logical QPU: a tightly coupled QPU + GPU real-time
// host connected by a low-latency interconnect (RoCE/RDMA). Used for real-time
// quantum error correction (QEC) and calibration workloads.
//
// Reference: NVIDIA NVQLink Architecture (Nov 2025)
//   - QPU + Quantum System Controller (QSC) + GPU host
//   - cudaq::device_call for microsecond-latency GPU callbacks
//   - Demonstrated: Quantinuum Helios + NVIDIA GH200 (5.4x error rate improvement)
type NVQLinkLogicalQPU struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NVQLinkSpec   `json:"spec,omitempty"`
	Status NVQLinkStatus `json:"status,omitempty"`
}

type NVQLinkSpec struct {
	// QPU configuration.
	QPU QPUConfig `json:"qpu"`

	// Real-time host (GPU) configuration.
	RealTimeHost RealTimeHostConfig `json:"realTimeHost"`

	// Interconnect configuration.
	Interconnect InterconnectConfig `json:"interconnect"`

	// QEC decoder to run on the GPU (e.g. "nv-qldpc-decoder", "bp-osd", "tensor-network").
	// +kubebuilder:default="nv-qldpc-decoder"
	// +optional
	QECDecoder string `json:"qecDecoder,omitempty"`

	// QEC code parameters.
	// +optional
	QECParameters *QECParameters `json:"qecParameters,omitempty"`

	// Credential for the QPU provider.
	Credential CredentialRef `json:"credential"`

	// Replicas: number of identical Logical QPUs to deploy.
	// +kubebuilder:default=1
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
}

type QPUConfig struct {
	// Provider: quantinuum, ionq, rigetti, iqm, custom.
	Provider string `json:"provider"`

	// Device ID.
	DeviceID string `json:"deviceId"`

	// Number of physical qubits.
	PhysicalQubits int32 `json:"physicalQubits"`

	// QPU type: superconducting, trapped-ion, photonic, neutral-atom, spin.
	// +optional
	QPUType string `json:"qpuType,omitempty"`

	// QPU system controller (QSC) type: RFSoC, FPGA, ASIC, custom.
	// +optional
	QSCTYPE string `json:"qscType,omitempty"`
}

type RealTimeHostConfig struct {
	// GPU model: gh200, h200, b200, a100, rtx-pro-6000.
	GPU string `json:"gpu"`

	// Number of GPUs.
	// +kubebuilder:default=1
	GPUs int32 `json:"gpus,omitempty"`

	// CPU architecture: arm64 (Grace), x86_64.
	// +optional
	CPUArch string `json:"cpuArch,omitempty"`

	// Memory (GiB).
	// +optional
	MemoryGiB int32 `json:"memoryGiB,omitempty"`

	// Container image for the real-time host runtime.
	// Default: nvcr.io/nvidia/cuda-quantum:latest
	// +optional
	Image string `json:"image,omitempty"`

	// Resources.
	// +optional
	Resources *ResourceRequirements `json:"resources,omitempty"`
}

type InterconnectConfig struct {
	// Technology: roce, rdma, nvlink, custom.
	// +kubebuilder:default=roce
	Technology string `json:"technology,omitempty"`

	// Target latency (microseconds).
	// +optional
	TargetLatencyUs int32 `json:"targetLatencyUs,omitempty"`

	// Bandwidth (Gbps).
	// +optional
	BandwidthGbps int32 `json:"bandwidthGbps,omitempty"`

	// NIC model (e.g. "connectx-7").
	// +optional
	NIC string `json:"nic,omitempty"`

	// Network namespace / VLAN.
	// +optional
	Network *NetworkConfig `json:"network,omitempty"`
}

type NetworkConfig struct {
	// +optional
	VLAN int32 `json:"vlan,omitempty"`
	// +optional
	Subnet string `json:"subnet,omitempty"`
	// +optional
	MTU int32 `json:"mtu,omitempty"`
}

type QECParameters struct {
	// Code: surface, color, ldpc, custom.
	// +optional
	Code string `json:"code,omitempty"`

	// Code distance.
	// +optional
	Distance int32 `json:"distance,omitempty"`

	// Number of QEC cycles per batch.
	// +optional
	Cycles int32 `json:"cycles,omitempty"`

	// Decoder iterations max.
	// +optional
	MaxIterations int32 `json:"maxIterations,omitempty"`

	// History depth for BP+OSD decoder.
	// +optional
	HistoryDepth int32 `json:"historyDepth,omitempty"`

	// Adaptive convergence: enable/disable.
	// +optional
	AdaptiveConvergence bool `json:"adaptiveConvergence,omitempty"`
}

type NVQLinkStatus struct {
	// Phase: Creating, Provisioning, Active, Degraded, Failed.
	Phase string `json:"phase,omitempty"`

	// Measured interconnect latency (microseconds).
	MeasuredLatencyUs int32 `json:"measuredLatencyUs,omitempty"`

	// Logical qubit error rate (from QEC).
	// +optional
	LogicalErrorRate float64 `json:"logicalErrorRate,omitempty"`

	// Active QEC cycles completed.
	QECCyclesCompleted int64 `json:"qecCyclesCompleted,omitempty"`

	// GPU utilization.
	// +optional
	GPUUtilization float64 `json:"gpuUtilization,omitempty"`

	// Ready replicas.
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Service name for in-cluster access.
	ServiceName string `json:"serviceName,omitempty"`

	Message string `json:"message,omitempty"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=nvqlink;nvq
// +kubebuilder:printcolumn:name="QPU",type=string,JSONPath=`.spec.qpu.deviceId`
// +kubebuilder:printcolumn:name="GPU",type=string,JSONPath=`.spec.realTimeHost.gpu`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Latency(μs)",type=integer,JSONPath=`.status.measuredLatencyUs`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type NVQLinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NVQLinkLogicalQPU `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NVQLinkLogicalQPU{}, &NVQLinkList{})
}
