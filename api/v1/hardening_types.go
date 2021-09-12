package v1

import (
	coreV1 "k8s.io/api/core/v1"
)

type CBContainersHardeningSpec struct {
	EventsGatewaySpec CBContainersEventsGatewaySpec `json:"eventsGatewaySpec,required"`
	// +kubebuilder:default:=<>
	EnforcerSpec CBContainersHardeningEnforcerSpec `json:"enforcerSpec,omitempty"`
	// +kubebuilder:default:=<>
	StateReporterSpec CBContainersHardeningStateReporterSpec `json:"stateReporterSpec,omitempty"`
}

type CBContainersHardeningStateReporterSpec struct {
	// +kubebuilder:default:=<>
	Labels map[string]string `json:"labels,omitempty"`
	// +kubebuilder:default:=<>
	DeploymentAnnotations map[string]string `json:"deploymentAnnotations,omitempty"`
	// +kubebuilder:default:=<>
	PodTemplateAnnotations map[string]string `json:"podTemplateAnnotations,omitempty"`
	// +kubebuilder:default:={repository:"cbartifactory/guardrails-state-reporter"}
	Image CBContainersImageSpec `json:"image,omitempty"`
	// +kubebuilder:default:=<>
	Env map[string]string `json:"env,omitempty"`
	// +kubebuilder:default:={requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "256Mi", cpu: "200m"}}
	Resources coreV1.ResourceRequirements `json:"resources,omitempty"`
	// +kubebuilder:default:=<>
	Probes CBContainersHTTPProbesSpec `json:"probes,omitempty"`
}

type CBContainersHardeningEnforcerSpec struct {
	// +kubebuilder:default:=<>
	Labels map[string]string `json:"labels,omitempty"`
	// +kubebuilder:default:=<>
	DeploymentAnnotations map[string]string `json:"deploymentAnnotations,omitempty"`
	// +kubebuilder:default:=<>
	PodTemplateAnnotations map[string]string `json:"podTemplateAnnotations,omitempty"`
	// +kubebuilder:default:=1
	ReplicasCount *int32 `json:"replicasCount,omitempty"`
	// +kubebuilder:default:=<>
	Env map[string]string `json:"env,omitempty"`
	// +kubebuilder:default:=<>
	Prometheus CBContainersPrometheusSpec `json:"prometheus,omitempty"`
	// +kubebuilder:default:={repository:"cbartifactory/guardrails-enforcer"}
	Image CBContainersImageSpec `json:"image,omitempty"`
	// +kubebuilder:default:={requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "256Mi", cpu: "200m"}}
	Resources coreV1.ResourceRequirements `json:"resources,omitempty"`
	// +kubebuilder:default:=<>
	Probes CBContainersHTTPProbesSpec `json:"probes,omitempty"`
	// +kubebuilder:default:=5
	WebhookTimeoutSeconds int32 `json:"webhookTimeoutSeconds,omitempty"`
}
