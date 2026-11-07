package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type QuantumSimulatorSpec struct {
	// Engine: aer, braket-local, qsharp-sim, qsim, horqrux, qibo.
	Engine string `json:"engine"`

	// Max qubits supported.
	// +optional
	MaxQubits int32 `json:"maxQubits,omitempty"`

	// Container image.
	// +optional
	Image string `json:"image,omitempty"`

	// Resources.
	// +optional
	Resources *ResourceRequirements `json:"resources,omitempty"`

	// Replicas.
	// +kubebuilder:default=1
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// GPU enabled.
	// +optional
	GPU bool `json:"gpu,omitempty"`
}

type QuantumSimulatorStatus struct {
	// Phase: Creating, Running, Failed.
	Phase string `json:"phase,omitempty"`

	// Service name for in-cluster access.
	ServiceName string `json:"serviceName,omitempty"`

	// Ready replicas.
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=qsims
type QuantumSimulator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuantumSimulatorSpec   `json:"spec,omitempty"`
	Status QuantumSimulatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type QuantumSimulatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QuantumSimulator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&QuantumSimulator{}, &QuantumSimulatorList{})
}
