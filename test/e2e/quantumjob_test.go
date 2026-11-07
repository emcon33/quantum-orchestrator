// Package e2e tests the Quantum Orchestrator end-to-end.
// Requires a running Kubernetes cluster and a quantum provider account.
package e2e

import (
	"context"
	"testing"
	"time"

	quantumv1alpha1 "github.com/quantum-orchestrator/quantum-orchestrator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = ctrl.SetupSignalHandler // ensure imported

func TestBellState(t *testing.T) {
	// 1. Create a QuantumJob
	// 2. Wait for completion (timeout 5 min)
	// 3. Verify results: counts should be mostly "00" and "11"
	t.Skip("requires cluster + provider")
}

func TestSessionLifecycle(t *testing.T) {
	// 1. Create a QuantumSession
	// 2. Submit a child job
	// 3. Verify session status
	t.Skip("requires cluster + provider")
}

func TestProviderBinding(t *testing.T) {
	// 1. Create a binding
	// 2. Submit a job in the bound namespace
	// 3. Verify it routes to the correct account
	t.Skip("requires cluster + provider")
}

// --- helpers ---

func waitForPhase(c client.Client, ns, name, phase string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		job := &quantumv1alpha1.QuantumJob{}
		if err := c.Get(context.TODO(), types.NamespacedName{Namespace: ns, Name: name}, job); err != nil {
			return err
		}
		if job.Status.Phase == phase {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return context.DeadlineExceeded
}

var _ = metav1.Now
