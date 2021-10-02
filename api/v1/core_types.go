package v1

import coreV1 "k8s.io/api/core/v1"

type CBContainersBasicSpec struct {
	// +kubebuilder:default:=<>
	Monitor CBContainersMonitorSpec `json:"monitor,omitempty"`
	// +kubebuilder:default:=<>
	Enforcer CBContainersEnforcerSpec `json:"enforcer,omitempty"`
	// +kubebuilder:default:=<>
	StateReporter CBContainersStateReporterSpec `json:"stateReporter,omitempty"`
}

type CBContainersMonitorSpec struct {
	// +kubebuilder:default:=<>
	Labels map[string]string `json:"labels,omitempty"`
	// +kubebuilder:default:=<>
	DeploymentAnnotations map[string]string `json:"deploymentAnnotations,omitempty"`
	// +kubebuilder:default:=<>
	PodTemplateAnnotations map[string]string `json:"podTemplateAnnotations,omitempty"`
	// +kubebuilder:default:=<>
	Env map[string]string `json:"env,omitempty"`
	// +kubebuilder:default:={repository:"cbartifactory/monitor"}
	Image CBContainersImageSpec `json:"image,omitempty"`
	// +kubebuilder:default:={requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "256Mi", cpu: "200m"}}
	Resources coreV1.ResourceRequirements `json:"resources,omitempty"`
	// +kubebuilder:default:=<>
	Probes CBContainersHTTPProbesSpec `json:"probes,omitempty"`
}
