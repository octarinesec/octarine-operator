package common_specs

import (
	coreV1 "k8s.io/api/core/v1"
)

type CBContainersCommonProbesSpec struct {
	// +kubebuilder:default:=3
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty"`
	// +kubebuilder:default:=1
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty"`
	// +kubebuilder:default:=30
	PeriodSeconds int32 `json:"periodSeconds,omitempty"`
	// +kubebuilder:default:=1
	SuccessThreshold int32 `json:"successThreshold,omitempty"`
	// +kubebuilder:default:=3
	FailureThreshold int32 `json:"failureThreshold,omitempty"`
}

type CBContainersHTTPProbesSpec struct {
	CBContainersCommonProbesSpec `json:"commonSpec,omitempty"`
	// +kubebuilder:default:="/ready"
	ReadinessPath string `json:"readinessPath,omitempty"`
	// +kubebuilder:default:="/alive"
	LivenessPath string `json:"livenessPath,omitempty"`
	// +kubebuilder:default:=8181
	Port int `json:"port,omitempty"`
	// +kubebuilder:default:="HTTP"
	Scheme coreV1.URIScheme `json:"scheme,omitempty"`
}

type CBContainersFileProbesSpec struct {
	CBContainersCommonProbesSpec `json:"commonSpec,omitempty"`
	// +kubebuilder:default:="/tmp/ready"
	ReadinessPath string `json:"readinessPath,omitempty"`
	// +kubebuilder:default:="/tmp/alive"
	LivenessPath string `json:"livenessPath,omitempty"`
}
