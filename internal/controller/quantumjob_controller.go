// Package controller implements the reconcile loops for Quantum Orchestrator CRDs.
package controller

import (
	"context"
	"fmt"
	"time"

	quantumv1alpha1 "github.com/quantum-orchestrator/quantum-orchestrator/api/v1alpha1"
	"github.com/quantum-orchestrator/quantum-orchestrator/internal/adapter"
	"github.com/quantum-orchestrator/quantum-orchestrator/internal/telemetry"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	requeueInterval = 30 * time.Second
	finalizerName   = "quantum.orchestrator.io/finalizer"
)

// QuantumJobReconciler reconciles a QuantumJob object.
type QuantumJobReconciler struct {
	client.Client

	// ProviderCache lazily creates and caches Provider instances by name.
	ProviderCache ProviderCache
}

// ProviderCache stores Provider instances keyed by provider name.
type ProviderCache interface {
	Get(ctx context.Context, name string, namespace string) (adapter.Provider, error)
}

// SetupWithManager registers the controller with the manager.
func (r *QuantumJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&quantumv1alpha1.QuantumJob{}).
		Named("quantumjob").
		Complete(r)
}

// Reconcile is the main reconcile loop.
func (r *QuantumJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 1. Fetch the QuantumJob
	job := &quantumv1alpha1.QuantumJob{}
	if err := r.Get(ctx, req.NamespacedName, job); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 2. Handle deletion
	if !job.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, job)
	}

	// 3. Add finalizer if not present
	if !containsString(job.Finalizers, finalizerName) {
		job.Finalizers = append(job.Finalizers, finalizerName)
		if err := r.Update(ctx, job); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 4. Reconcile based on current phase
	switch job.Status.Phase {
	case "", "Pending":
		return r.submitJob(ctx, job)
	case "Submitted", "Queued":
		return r.pollStatus(ctx, job)
	case "Running":
		return r.pollStatus(ctx, job)
	case "Completed":
		return r.finalizeJob(ctx, job)
	case "Failed":
		return r.finalizeJob(ctx, job)
	default:
		return ctrl.Result{}, nil
	}
}

func (r *QuantumJobReconciler) submitJob(ctx context.Context, job *quantumv1alpha1.QuantumJob) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("submitting quantum job", "provider", job.Spec.Target.Provider)

	// Resolve provider
	provider, err := r.ProviderCache.Get(ctx, job.Spec.Target.Provider, job.Namespace)
	if err != nil {
		r.setStatus(job, "Failed", fmt.Sprintf("provider resolution: %v", err))
		return ctrl.Result{}, err
	}

	// Build job spec
	spec := adapter.JobSpec{
		Circuit:      r.resolveCircuit(job),
		DeviceID:     job.Spec.Target.Device,
		Shots:        int(job.Spec.Shots),
		Parameters:   job.Spec.Parameters,
		ResultBackend: "",
		ResultBucket:  "",
	}

	if job.Spec.Result != nil {
		spec.ResultBackend = job.Spec.Result.Backend
		spec.ResultBucket = job.Spec.Result.Bucket
	}

	if job.Spec.Hybrid != nil {
		spec.Hybrid = &adapter.HybridConfig{
			Image:      job.Spec.Hybrid.Image,
			Entrypoint: job.Spec.Hybrid.Entrypoint,
		}
	}

	// Submit
	handle, err := provider.SubmitJob(ctx, spec)
	if err != nil {
		r.setStatus(job, "Failed", fmt.Sprintf("submit: %v", err))
		telemetry.JobSubmitErrors.WithLabelValues(job.Spec.Target.Provider).Inc()
		return ctrl.Result{}, err
	}

	// Update status
	now := metav1.Now()
	job.Status.Phase = "Submitted"
	job.Status.ProviderJobID = handle.ProviderJobID
	job.Status.ProviderResourceID = handle.ResourceID
	job.Status.StartTime = &now
	job.Status.Message = "Job submitted to " + provider.Name()

	if err := r.Status().Update(ctx, job); err != nil {
		return ctrl.Result{}, err
	}

	telemetry.JobSubmissions.WithLabelValues(job.Spec.Target.Provider).Inc()
	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

func (r *QuantumJobReconciler) pollStatus(ctx context.Context, job *quantumv1alpha1.QuantumJob) (ctrl.Result, error) {
	provider, err := r.ProviderCache.Get(ctx, job.Spec.Target.Provider, job.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}

	handle := &adapter.JobHandle{
		ProviderJobID: job.Status.ProviderJobID,
		ResourceID:    job.Status.ProviderResourceID,
		Provider:      job.Spec.Target.Provider,
	}

	status, err := provider.GetStatus(ctx, handle)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Update phase if changed
	newPhase := string(status.Phase)
	if job.Status.Phase != newPhase {
		job.Status.Phase = newPhase
		job.Status.Message = status.Message
		job.Status.CostEstimate = status.CostEstimate

		if newPhase == "Completed" {
			now := metav1.Now()
			job.Status.CompletionTime = &now
			telemetry.JobCompletions.WithLabelValues(job.Spec.Target.Provider).Inc()

			// Fetch results
			results, err := provider.GetResults(ctx, handle)
			if err == nil {
				job.Status.ResultLocation = results.RawResultURI
			}
		}

		if err := r.Status().Update(ctx, job); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

func (r *QuantumJobReconciler) finalizeJob(ctx context.Context, job *quantumv1alpha1.QuantumJob) (ctrl.Result, error) {
	// Remove finalizer to allow deletion
	job.Finalizers = removeString(job.Finalizers, finalizerName)
	return ctrl.Result{}, r.Update(ctx, job)
}

func (r *QuantumJobReconciler) handleDeletion(ctx context.Context, job *quantumv1alpha1.QuantumJob) (ctrl.Result, error) {
	// Cancel job at provider if still active
	if job.Status.Phase != "Completed" && job.Status.Phase != "Failed" {
		provider, err := r.ProviderCache.Get(ctx, job.Spec.Target.Provider, job.Namespace)
		if err == nil {
			_ = provider.CancelJob(ctx, &adapter.JobHandle{
				ProviderJobID: job.Status.ProviderJobID,
				Provider:      job.Spec.Target.Provider,
			})
		}
	}
	job.Finalizers = removeString(job.Finalizers, finalizerName)
	return ctrl.Result{}, r.Update(ctx, job)
}

// --- helpers ---

func (r *QuantumJobReconciler) setStatus(job *quantumv1alpha1.QuantumJob, phase, msg string) {
	job.Status.Phase = phase
	job.Status.Message = msg
	_ = r.Status().Update(context.TODO(), job)
}

func (r *QuantumJobReconciler) resolveCircuit(job *quantumv1alpha1.QuantumJob) string {
	if job.Spec.Circuit == nil {
		return ""
	}
	if job.Spec.Circuit.Inline != "" {
		return job.Spec.Circuit.Inline
	}
	// TODO: fetch from ConfigMap/Secret ref
	return ""
}

func containsString(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(list []string, s string) []string {
	result := make([]string, 0, len(list))
	for _, item := range list {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

// Ensure QuantumJobReconciler implements the Reconciler interface.
var _ = types.NamespacedName{}
