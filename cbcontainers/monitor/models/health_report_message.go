package models

type WorkloadKind string
type WebhookType string

const (
	WorkloadKindDeployment = WorkloadKind("workload-kind-deployment")
	WorkloadKindDaemonSet  = WorkloadKind("workload-kind-daemonset")

	WebhookTypeValidating = WebhookType("webhook-type-validating")
	WebhookTypeMutating   = WebhookType("webhook-type-mutating")
)

type HealthReportMessage struct {
	Account           string
	Cluster           string
	Version           string
	EnabledComponents map[string]bool
	Workloads         map[string]WorkloadHealthReport
	Webhooks          map[string]WebhookHealthReport
}

type WorkloadHealthReport struct {
	Kind            WorkloadKind
	Status          []byte
	Labels          map[string]string
	Spec            WorkloadSpecReport
	ReplicasReports map[string]WorkloadReplicaHealthReport
}

type WorkloadSpecReport struct {
	Replicas   int32
	Containers map[string]ContainerHealthReport
}

type WorkloadReplicaHealthReport struct {
	Node   string
	Spec   WorkloadReplicaSpecReport
	Status []byte
}

type WorkloadReplicaSpecReport struct {
	Containers map[string]ContainerHealthReport
}

type ContainerHealthReport struct {
	Image string
}

type WebhookHealthReport struct {
	Type WebhookType
	Uid  string
}
