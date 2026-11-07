package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// QuantumJobSpec defines the desired state of a QuantumJob.
type QuantumJobSpec struct {
	// Circuit is the quantum program in OpenQASM 3 format (inline) or a reference.
	// +optional
	Circuit *CircuitRef `json:"circuit,omitempty"`

	// Target specifies the quantum backend to run on.
	Target TargetSpec `json:"target"`

	// Shots is the number of measurement repetitions. Default: 1024.
	// +kubebuilder:default=1024
	// +optional
	Shots int32 `json:"shots,omitempty"`

	// Hybrid configuration for classical+quantum workloads.
	// +optional
	Hybrid *HybridSpec `json:"hybrid,omitempty"`

	// Credentials reference for provider auth.
	Credential CredentialRef `json:"credential"`

	// Scheduling hints (queue, priority, preemption policy).
	// +optional
	Scheduling *SchedulingSpec `json:"scheduling,omitempty"`

	// Result output configuration.
	// +optional
	Result *ResultSpec `json:"result,omitempty"`

	// Parameters for parameterized circuits.
	// +optional
	Parameters map[string]string `json:"parameters,omitempty"`
}

type CircuitRef struct {
	// Inline OpenQASM 3 source.
	// +optional
	Inline string `json:"inline,omitempty"`

	// Reference to a ConfigMap/Secret containing the circuit.
	// +optional
	Ref *ObjectRef `json:"ref,omitempty"`

	// Format of the circuit. Default: openqasm3.
	// +kubebuilder:default=openqasm3
	// +optional
	Format string `json:"format,omitempty"`
}

type TargetSpec struct {
	// Provider: aws-braket, azure-quantum, ionq, rigetti, quantinuum, local
	Provider string `json:"provider"`

	// Device identifier (e.g. "sv1", "ionq.simulator", "rigetti.qpu.cepheus-1-108q").
	Device string `json:"device"`

	// Use a local simulator instead of a cloud QPU.
	// +optional
	Simulator bool `json:"simulator,omitempty"`
}

type HybridSpec struct {
	// Container image for the classical component.
	Image string `json:"image"`

	// Entrypoint command.
	// +optional
	Entrypoint []string `json:"entrypoint,omitempty"`

	// Resource requirements for the classical component.
	// +optional
	Resources *ResourceRequirements `json:"resources,omitempty"`

	// Environment variables.
	// +optional
	Env []EnvVar `json:"env,omitempty"`
}

type ResourceRequirements struct {
	// +optional
	Limits map[string]string `json:"limits,omitempty"`
	// +optional
	Requests map[string]string `json:"requests,omitempty"`
}

type EnvVar struct {
	Name  string  `json:"name"`
	Value *string `json:"value,omitempty"`
	// +optional
	ValueFrom *EnvVarSource `json:"valueFrom,omitempty"`
}

type EnvVarSource struct {
	// +optional
	SecretRef *SecretKeyRef `json:"secretRef,omitempty"`
	// +optional
	ConfigMapRef *ConfigMapKeyRef `json:"configMapRef,omitempty"`
}

type SecretKeyRef struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type ConfigMapKeyRef struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type CredentialRef struct {
	// Name of the K8s Secret containing provider credentials.
	SecretRef SecretRef `json:"secretRef"`
}

type SecretRef struct {
	Name string `json:"name"`
	// +optional
	Key string `json:"key,omitempty"`
}

type SchedulingSpec struct {
	// Kueue queue name.
	// +optional
	Queue string `json:"queue,omitempty"`

	// Priority class name.
	// +optional
	PriorityClass string `json:"priorityClass,omitempty"`

	// Preemption policy: Never, LowerPriority, Any.
	// +kubebuilder:default=Never
	// +optional
	Preemption string `json:"preemption,omitempty"`

	// Max retry count on transient failures.
	// +kubebuilder:default=3
	// +optional
	Retries int32 `json:"retries,omitempty"`
}

type ResultSpec struct {
	// Output storage backend: s3, azure-blob, minio.
	// +optional
	Backend string `json:"backend,omitempty"`

	// Bucket / container name.
	// +optional
	Bucket string `json:"bucket,omitempty"`

	// Key prefix.
	// +optional
	Prefix string `json:"prefix,omitempty"`
}

// ObjectRef is a generic reference to a K8s object.
type ObjectRef struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
	// +optional
	Key string `json:"key,omitempty"`
}

// QuantumJobStatus defines the observed state of QuantumJob.
type QuantumJobStatus struct {
	// Phase: Pending, Scheduling, Submitted, Queued, Running, Completed, Failed, Cancelled
	// +kubebuilder:default=Pending
	Phase string `json:"phase,omitempty"`

	// ProviderJobID is the job identifier returned by the provider.
	// +optional
	ProviderJobID string `json:"providerJobId,omitempty"`

	// ProviderARN / Azure resource ID.
	// +optional
	ProviderResourceID string `json:"providerResourceId,omitempty"`

	// ResultLocation is the URI where results are stored.
	// +optional
	ResultLocation string `json:"resultLocation,omitempty"`

	// Message provides human-readable status detail.
	// +optional
	Message string `json:"message,omitempty"`

	// Timing.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// CostEstimate in USD.
	// +optional
	CostEstimate float64 `json:"costEstimate,omitempty"`

	// ObservedGeneration.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=qjob;qjobs
// +kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.target.provider`
// +kubebuilder:printcolumn:name="Device",type=string,JSONPath=`.spec.target.device`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type QuantumJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuantumJobSpec   `json:"spec,omitempty"`
	Status QuantumJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type QuantumJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QuantumJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&QuantumJob{}, &QuantumJobList{})
}
