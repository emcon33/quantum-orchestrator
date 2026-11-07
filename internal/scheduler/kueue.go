// Package scheduler provides Kueue integration for quantum job scheduling.
package scheduler

import (
	"context"
	"fmt"
)

// KueueScheduler wraps Kueue ResourceFlavor and ClusterQueue operations.
type KueueScheduler struct {
	// client is the Kueue client (in production, kueuev1beta1.Clientset)
	// For now, this is a stub.
}

// NewKueueScheduler creates a scheduler bound to a cluster.
func NewKueueScheduler() *KueueScheduler {
	return &KueueScheduler{}
}

// Submit enqueues a quantum job into the specified Kueue queue.
// In production, this creates a Kueue Job object that references the
// QuantumJob as the workload.
func (s *KueueScheduler) Submit(ctx context.Context, queue, namespace, name string) error {
	// 1. Create a Kueue Job CR with:
	//    - queue: <queue>
	//    - podTemplate: the classical component (if hybrid)
	//    - priorityClass: from QuantumJob.spec.scheduling
	//    - resourceFlavors: the QPU device class
	//
	// 2. Kueue will admit the job when resources are available.
	//
	// 3. The controller watches for the "admitted" condition and
	//    proceeds with provider submission.
	return fmt.Errorf("kueue: not yet implemented (wire up kueuev1beta1 client)")
}

// Cancel removes a job from the Kueue queue.
func (s *KueueScheduler) Cancel(ctx context.Context, namespace, name string) error {
	return nil
}
