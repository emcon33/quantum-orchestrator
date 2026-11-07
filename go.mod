module github.com/quantum-orchestrator/quantum-orchestrator

go 1.22

require (
	k8s.io/api v0.30.0
	k8s.io/apimachinery v0.30.0
	k8s.io/client-go v0.30.0
	sigs.k8s.io/controller-runtime v0.18.3
	github.com/spf13/cobra v1.8.1
	github.com/prometheus/client_golang v1.19.0
	go.opentelemetry.io/otel v1.26.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.26.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.26.0
	go.opentelemetry.io/otel/sdk v1.26.0
	github.com/aws/aws-sdk-go-v2 v1.30.0
	github.com/aws/aws-sdk-go-v2/service/braket v1.25.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.6.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/quantum/armquantum v0.7.0
	gopkg.in/yaml.v3 v3.0.1
)
