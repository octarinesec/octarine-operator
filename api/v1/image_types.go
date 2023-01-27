package v1

import (
	coreV1 "k8s.io/api/core/v1"
)

type CBContainersImageSpec struct {
	Repository string `json:"repository,omitempty"`
	Tag        string `json:"tag,omitempty"`
	// +kubebuilder:default:="IfNotPresent"
	PullPolicy coreV1.PullPolicy `json:"pullPolicy,omitempty"`
	// PullSecrets is a list of secret names, which will be used to pull the container image(s).
	//
	// The secrets must already exist.
	PullSecrets []string `json:"pullSecrets,omitempty"`
}
