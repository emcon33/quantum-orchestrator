package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type QuantumProviderBindingSpec struct {
	// Cloud: aws, azure, direct.
	Cloud string `json:"cloud"`

	// Region.
	// +optional
	Region string `json:"region,omitempty"`

	// AWS account ID or Azure subscription ID.
	// +optional
	AccountID string `json:"accountId,omitempty"`

	// Azure workspace name / AWS Braket project.
	// +optional
	WorkspaceID string `json:"workspaceId,omitempty"`

	// Credential reference.
	Credential CredentialRef `json:"credential"`

	// Namespaces this binding applies to. Empty = all.
	// +optional
	Namespaces []string `json:"namespaces,omitempty"`

	// Tenant label selector.
	// +optional
	TenantSelector *metav1.LabelSelector `json:"tenantSelector,omitempty"`
}

type QuantumProviderBindingStatus struct {
	// Valid: whether the credentials are reachable.
	Valid bool `json:"valid,omitempty"`

	// LastVerified.
	LastVerified *metav1.Time `json:"lastVerified,omitempty"`

	// Message.
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=qpbinding;qpb
type QuantumProviderBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuantumProviderBindingSpec   `json:"spec,omitempty"`
	Status QuantumProviderBindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type QuantumProviderBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QuantumProviderBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&QuantumProviderBinding{}, &QuantumProviderBindingList{})
}
