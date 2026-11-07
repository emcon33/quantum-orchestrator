package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type QPUBackendSpec struct {
	// Provider name.
	Provider string `json:"provider"`

	// Device ID.
	DeviceID string `json:"deviceId"`

	// Number of qubits.
	Qubits int32 `json:"qubits"`

	// Gate set (e.g. ["H", "CX", "RZ", "RX"]).
	// +optional
	GateSet []string `json:"gateSet,omitempty"`

	// Error rates.
	// +optional
	ErrorRates *ErrorRates `json:"errorRates,omitempty"`

	// Availability: on-demand, reserved, subscription.
	// +kubebuilder:default=on-demand
	Availability string `json:"availability,omitempty"`

	// Cost model.
	// +optional
	Cost *CostModel `json:"cost,omitempty"`

	// URL for the provider dashboard.
	// +optional
	DashboardURL string `json:"dashboardUrl,omitempty"`
}

type ErrorRates struct {
	// Gate error rate.
	// +optional
	Gate float64 `json:"gate,omitempty"`
	// Readout error rate.
	// +optional
	Readout float64 `json:"readout,omitempty"`
	// T1 (microseconds).
	// +optional
	T1 float64 `json:"t1,omitempty"`
	// T2 (microseconds).
	// +optional
	T2 float64 `json:"t2,omitempty"`
}

type CostModel struct {
	// Billing: per-shot, per-minute, per-qubit-minute.
	Billing string `json:"billing,omitempty"`

	// Cost per unit in USD.
	// +optional
	UnitCost float64 `json:"unitCost,omitempty"`

	// Currency.
	// +kubebuilder:default=USD
	Currency string `json:"currency,omitempty"`
}

type QPUBackendStatus struct {
	// Available: true/false.
	Available bool `json:"available,omitempty"`

	// LastChecked.
	LastChecked *metav1.Time `json:"lastChecked,omitempty"`

	// Current queue depth at the provider.
	QueueDepth int32 `json:"queueDepth,omitempty"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=qpu;qpus
type QPUBackend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QPUBackendSpec   `json:"spec,omitempty"`
	Status QPUBackendStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type QPUBackendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QPUBackend `json:"items"`
}

func init() {
	SchemeBuilder.Register(&QPUBackend{}, &QPUBackendList{})
}
