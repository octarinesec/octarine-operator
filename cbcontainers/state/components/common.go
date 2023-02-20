package components

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	coreV1 "k8s.io/api/core/v1"
)

// getSharedImagePullSecrets returns a list of shared pull secrets, defined in the agent spec.
//
// Each component can add more pull secret to the resulting array if such are defined.
func getSharedImagePullSecrets(agentSpec *cbcontainersv1.CBContainersAgentSpec) []coreV1.LocalObjectReference {
	var imagePullSecrets []coreV1.LocalObjectReference
	if agentSpec.Components.Settings.ShouldCreateDefaultImagePullSecrets() {
		imagePullSecrets = append(imagePullSecrets, coreV1.LocalObjectReference{Name: commonState.RegistrySecretName})
	}
	for _, secretName := range agentSpec.Components.Settings.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, coreV1.LocalObjectReference{Name: secretName})
	}
	return imagePullSecrets
}
