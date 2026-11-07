// Package telemetry provides Prometheus metrics and OpenTelemetry tracing
// for the Quantum Orchestrator.
package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// JobSubmissions counts total job submissions by provider.
	JobSubmissions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "quantum_job_submissions_total",
			Help: "Total quantum jobs submitted, by provider.",
		},
		[]string{"provider"},
	)

	// JobCompletions counts completed jobs.
	JobCompletions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "quantum_job_completions_total",
			Help: "Total quantum jobs completed, by provider.",
		},
		[]string{"provider"},
	)

	// JobSubmitErrors counts submission failures.
	JobSubmitErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "quantum_job_submit_errors_total",
			Help: "Total quantum job submission errors, by provider.",
		},
		[]string{"provider"},
	)

	// JobDuration tracks job wall-clock duration.
	JobDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "quantum_job_duration_seconds",
			Help:    "Duration of quantum jobs from submission to completion.",
			Buckets: prometheus.ExponentialBuckets(1, 2, 16), // 1s to ~1h
		},
		[]string{"provider", "phase"},
	)

	// QueueWait tracks time spent in provider queue.
	QueueWait = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "quantum_job_queue_wait_seconds",
			Help:    "Time spent waiting in the provider queue.",
			Buckets: prometheus.ExponentialBuckets(1, 2, 14),
		},
		[]string{"provider"},
	)

	// CircuitDepth tracks circuit depth distribution.
	CircuitDepth = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "quantum_circuit_depth",
			Help:    "Distribution of circuit depths (number of layers).",
			Buckets: prometheus.LinearBuckets(1, 2, 50),
		},
		[]string{"provider"},
	)

	// QPUUtilization tracks active QPU time.
	QPUUtilization = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "quantum_qpu_utilization_ratio",
			Help: "Fraction of QPU time in use (0.0 to 1.0).",
		},
		[]string{"provider", "device"},
	)

	// CostEstimate tracks estimated cost per job.
	CostEstimate = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "quantum_cost_estimate_usd",
			Help:    "Estimated cost of quantum jobs in USD.",
			Buckets: prometheus.ExponentialBuckets(0.01, 10, 8),
		},
		[]string{"provider"},
	)

	// ReconcileDuration tracks controller reconcile loop duration.
	ReconcileDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "quantum_reconcile_duration_seconds",
			Help:    "Duration of controller reconcile loops.",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 12),
		},
		[]string{"controller"},
	)
)
