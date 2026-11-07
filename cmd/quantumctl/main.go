// quantumctl is a kubectl-like CLI for managing quantum workloads.
// It talks to the Kubernetes API; no cloud provider SDKs are needed.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	quantumv1alpha1 "github.com/quantum-orchestrator/quantum-orchestrator/api/v1alpha1"
)

var (
	kubeconfig  string
	namespace   string
	output      string
)

func main() {
	root := &cobra.Command{
		Use:   "quantumctl",
		Short: "kubectl-like CLI for quantum workloads",
		Long:  `quantumctl manages quantum computing jobs, sessions, and backends on Kubernetes.`,
	}

	root.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig")
	root.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "Namespace")
	root.PersistentFlags().StringVarP(&output, "output", "o", "table", "Output format: table, json, yaml")

	root.AddCommand(newSubmitCmd())
	root.AddCommand(newGetCmd())
	root.AddCommand(newLogsCmd())
	root.AddCommand(newResultsCmd())
	root.AddCommand(newDevicesCmd())
	root.AddCommand(newCostCmd())
	root.AddCommand(newDeleteCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func getClient() (client.Client, error) {
	config, err := restConfig()
	if err != nil {
		return nil, err
	}
	scheme := clientgoscheme()
	return client.New(config, client.Options{Scheme: scheme})
}

func restConfig() (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func clientgoscheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(s)
	_ = quantumv1alpha1.AddToScheme(s)
	return s
}

// --- commands ---

func newSubmitCmd() *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "submit -f <file>",
		Short: "Submit a quantum job from a YAML file",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}

			job := &quantumv1alpha1.QuantumJob{}
			if err := yaml.Unmarshal(data, job); err != nil {
				return fmt.Errorf("parse YAML: %w", err)
			}

			c, err := getClient()
			if err != nil {
				return err
			}

			job.Namespace = namespace
			if err := c.Create(context.TODO(), job); err != nil {
				return fmt.Errorf("create job: %w", err)
			}

			fmt.Printf("quantumjob/%s created\n", job.Name)
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "filename", "f", "", "Path to YAML file (required)")
	cmd.MarkFlagRequired("filename")
	return cmd
}

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [type] [name]",
		Short: "List or describe quantum resources",
		Args:  cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := getClient()
			if err != nil {
				return err
			}

			switch len(args) {
			case 0, 1:
				return listQuantumJobs(c, namespace)
			case 2:
				return getQuantumJob(c, namespace, args[1])
			}
			return nil
		},
	}
	return cmd
}

func newLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs <job-name>",
		Short: "Stream logs for a quantum job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// In production: attach to the classical pod (if hybrid)
			// or fetch provider-side logs via the adapter.
			fmt.Printf("Streaming logs for quantumjob/%s...\n", args[0])
			return nil
		},
	}
	return cmd
}

func newResultsCmd() *cobra.Command {
	var plot bool
	cmd := &cobra.Command{
		Use:   "results <job-name>",
		Short: "Display results for a completed quantum job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := getClient()
			if err != nil {
				return err
			}

			job := &quantumv1alpha1.QuantumJob{}
			if err := c.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: args[0]}, job); err != nil {
				return err
			}

			if job.Status.Phase != "Completed" {
				fmt.Printf("Job %s is in phase: %s\n", args[0], job.Status.Phase)
				return nil
			}

			fmt.Println("Results:")
			fmt.Printf("  Location: %s\n", job.Status.ResultLocation)
			if plot {
				fmt.Println("  (plotting would go here)")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&plot, "plot", false, "Plot the measurement distribution")
	return cmd
}

func newDevicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "devices",
		Short: "List available QPU backends",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := getClient()
			if err != nil {
				return err
			}

			backends := &quantumv1alpha1.QPUBackendList{}
			if err := c.List(context.TODO(), backends, client.InNamespace(namespace)); err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tPROVIDER\tDEVICE\tQUBITS\tAVAILABLE\tCOST")
			for _, b := range backends.Items {
				avail := "false"
				if b.Status.Available {
					avail = "true"
				}
				cost := "-"
				if b.Spec.Cost != nil {
					cost = fmt.Sprintf("$%.4f/%s", b.Spec.Cost.UnitCost, b.Spec.Cost.Billing)
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
					b.Name, b.Spec.Provider, b.Spec.DeviceID, b.Spec.Qubits, avail, cost)
			}
			w.Flush()
			return nil
		},
	}
	return cmd
}

func newCostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cost",
		Short: "Show cost summary for quantum jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := getClient()
			if err != nil {
				return err
			}

			jobs := &quantumv1alpha1.QuantumJobList{}
			if err := c.List(context.TODO(), jobs, client.InNamespace(namespace)); err != nil {
				return err
			}

			total := 0.0
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "JOB\tPROVIDER\tPHASE\tCOST (USD)\tAGE")
			for _, j := range jobs.Items {
				total += j.Status.CostEstimate
				age := time.Since(j.CreationTimestamp.Time).Round(time.Minute)
				fmt.Fprintf(w, "%s\t%s\t%s\t$%.4f\t%s\n",
					j.Name, j.Spec.Target.Provider, j.Status.Phase, j.Status.CostEstimate, age)
			}
			fmt.Fprintf(w, "\nTOTAL\t\t\t$%.4f\t\n", total)
			w.Flush()
			return nil
		},
	}
	return cmd
}

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <job-name>",
		Short: "Delete a quantum job (cancels at provider if active)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := getClient()
			if err != nil {
				return err
			}

			job := &quantumv1alpha1.QuantumJob{
				ObjectMeta: metav1.ObjectMeta{Name: args[0], Namespace: namespace},
			}
			if err := c.Delete(context.TODO(), job); err != nil {
				return err
			}
			fmt.Printf("quantumjob/%s deleted\n", args[0])
			return nil
		},
	}
	return cmd
}

// --- helpers ---

func listQuantumJobs(c client.Client, ns string) error {
	jobs := &quantumv1alpha1.QuantumJobList{}
	if err := c.List(context.TODO(), jobs, client.InNamespace(ns)); err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tPROVIDER\tDEVICE\tSHOTS\tPHASE\tAGE")
	for _, j := range jobs.Items {
		age := time.Since(j.CreationTimestamp.Time).Round(time.Minute)
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
			j.Name, j.Spec.Target.Provider, j.Spec.Target.Device, j.Spec.Shots, j.Status.Phase, age)
	}
	w.Flush()
	return nil
}

func getQuantumJob(c client.Client, ns, name string) error {
	job := &quantumv1alpha1.QuantumJob{}
	if err := c.Get(context.TODO(), types.NamespacedName{Namespace: ns, Name: name}, job); err != nil {
		return err
	}
	out, _ := json.MarshalIndent(job, "", "  ")
	fmt.Println(string(out))
	return nil
}
