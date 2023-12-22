package v1

import (
	coreV1 "k8s.io/api/core/v1"
)

type CBContainersCndrSensorSpec struct {
	// +kubebuilder:default:=<>
	Labels map[string]string `json:"labels,omitempty"`
	// +kubebuilder:default:=<>
	DaemonSetAnnotations map[string]string `json:"daemonSetAnnotations,omitempty"`
	// +kubebuilder:default:=<>
	PodTemplateAnnotations map[string]string `json:"podTemplateAnnotations,omitempty"`
	// +kubebuilder:default:=<>
	Env map[string]string `json:"env,omitempty"`
	// +kubebuilder:default:={repository:"cbartifactory/cndr"}
	Image CBContainersImageSpec `json:"image,omitempty"`
	// +kubebuilder:default:={requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "1024Mi", cpu: "500m"}}
	Resources coreV1.ResourceRequirements `json:"resources,omitempty"`
	// +kubebuilder:default:={initialDelaySeconds: 240, timeoutSeconds: 1, periodSeconds: 30, successThreshold: 1, failureThreshold: 5, readinessPath: "/tmp/ready", livenessPath:  "/tmp/alive" }
	Probes CBContainersFileProbesSpec `json:"probes,omitempty"`
	// +kubebuilder:default:={port: 7071}
	Prometheus CBContainersPrometheusSpec `json:"prometheus,omitempty"`
	// +kubebuilder:default:=2
	VerbosityLevel *int `json:"verbosity_level,omitempty"`
	// +kubebuilder:default:="info"
	LogLevel string `json:"logLevel,omitempty"`
}

// CBContainersCndrSpec defines the desired state of CBContainersCndr
type CBContainersCndrSpec struct {
	// +kubebuilder:default:=false
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:default:=<>
	Sensor CBContainersCndrSensorSpec `json:"sensor,omitempty"`
	// +kubebuilder:default:="cbcontainers-company-code"
	CompanyCodeSecretName string `json:"companyCodeSecretName,omitempty"`
}
