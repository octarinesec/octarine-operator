package v1

import (
	coreV1 "k8s.io/api/core/v1"
)

type CBContainersStateReporterSpec struct {
	// +kubebuilder:default:=<>
	Labels map[string]string `json:"labels,omitempty"`
	// +kubebuilder:default:=<>
	DeploymentAnnotations map[string]string `json:"deploymentAnnotations,omitempty"`
	// +kubebuilder:default:=<>
	PodTemplateAnnotations map[string]string `json:"podTemplateAnnotations,omitempty"`
	// +kubebuilder:default:=<>
	Env map[string]string `json:"env,omitempty"`
	// +kubebuilder:default:={repository:"cbartifactory/guardrails-state-reporter"}
	Image CBContainersImageSpec `json:"image,omitempty"`
	// +kubebuilder:default:={requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "256Mi", cpu: "200m"}}
	Resources coreV1.ResourceRequirements `json:"resources,omitempty"`
	// +kubebuilder:default:=<>
	Probes CBContainersHTTPProbesSpec `json:"probes,omitempty"`
	// +kubebuilder:default:=<>
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// +kubebuilder:default:=<>
	Affinity *coreV1.Affinity `json:"affinity,omitempty"`
	// ImagePullSecrets is a list of image pull secret names, which will be used to pull the container image(s)
	// for the State Reporter Deployment.
	//
	// The secrets must already exist.
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
}

type CBContainersEnforcerSpec struct {
	// +kubebuilder:default:=<>
	Labels map[string]string `json:"labels,omitempty"`
	// +kubebuilder:default:=<>
	DeploymentAnnotations map[string]string `json:"deploymentAnnotations,omitempty"`
	// +kubebuilder:default:=<>
	PodTemplateAnnotations map[string]string `json:"podTemplateAnnotations,omitempty"`
	// +kubebuilder:default:=<>
	Env map[string]string `json:"env,omitempty"`
	// +kubebuilder:default:=1
	ReplicasCount *int32 `json:"replicasCount,omitempty"`
	// +kubebuilder:default:={port: 7071}
	Prometheus CBContainersPrometheusSpec `json:"prometheus,omitempty"`
	// +kubebuilder:default:={repository:"cbartifactory/guardrails-enforcer"}
	Image CBContainersImageSpec `json:"image,omitempty"`
	// +kubebuilder:default:={requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "256Mi", cpu: "200m"}}
	Resources coreV1.ResourceRequirements `json:"resources,omitempty"`
	// +kubebuilder:default:=<>
	Probes CBContainersHTTPProbesSpec `json:"probes,omitempty"`
	// +kubebuilder:default:=<>
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// +kubebuilder:default:=<>
	Affinity *coreV1.Affinity `json:"affinity,omitempty"`
	// +kubebuilder:default:=5
	WebhookTimeoutSeconds int32 `json:"webhookTimeoutSeconds,omitempty"`
	// +kubebuilder:default:=true
	EnableEnforcementFeature *bool `json:"enableEnforcementFeature,omitempty"`
	// +kubebuilder:validation:Enum=Ignore;Fail
	// +kubebuilder:default:=Ignore
	FailurePolicy string `json:"failurePolicy,omitempty"`
	// ImagePullSecrets is a list of image pull secret names, which will be used to pull the container image(s)
	// for the Enforcer Deployment.
	//
	// The secrets must already exist.
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
}
