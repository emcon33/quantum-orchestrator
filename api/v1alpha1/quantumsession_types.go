package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type QuantumSessionSpec struct {
	Target    TargetSpec    `json:"target"`
	Credential CredentialRef `json:"credential"`

	// Duration of the session.
	// +kubebuilder:default="1h"
	Duration string `json:"duration,omitempty"`

	// Number of qubits to reserve.
	// +optional
	Qubits int32 `json:"qubits,omitempty"`

	// Priority.
	// +optional
	Priority int32 `json:"priority,omitempty"`
}

type QuantumSessionStatus struct {
	// Phase: Creating, Active, Idle, Expired, Failed
	Phase string `json:"phase,omitempty"`

	// ProviderSessionID.
	ProviderSessionID string `json:"providerSessionId,omitempty"`

	// QueuePosition at the provider.
	QueuePosition int32 `json:"queuePosition,omitempty"`

	// ChildJobs count.
	ChildJobs int32 `json:"childJobs,omitempty"`

	// StartTime.
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// ExpiryTime.
	ExpiryTime *metav1.Time `json:"expiryTime,omitempty"`

	// Message.
	Message string `json:"message,omitempty"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=qsession;qsessions
type QuantumSession struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuantumSessionSpec   `json:"spec,omitempty"`
	Status QuantumSessionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type QuantumSessionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QuantumSession `json:"items"`
}

func init() {
	SchemeBuilder.Register(&QuantumSession{}, &QuantumSessionList{})
}
