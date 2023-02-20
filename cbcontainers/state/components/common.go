package components

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	coreV1 "k8s.io/api/core/v1"
)

// getImagePullSecrets returns a list of shared pull secrets, defined in the agent spec.
//
// Additional secrets can be provided via the variadic argument of this method.
func getImagePullSecrets(agentSpec *cbcontainersv1.CBContainersAgentSpec, additionalSecrets ...string) []coreV1.LocalObjectReference {
	var imagePullSecrets []coreV1.LocalObjectReference
	if agentSpec.Components.Settings.ShouldCreateDefaultImagePullSecrets() {
		imagePullSecrets = append(imagePullSecrets, coreV1.LocalObjectReference{Name: commonState.RegistrySecretName})
	}
	for _, secretName := range agentSpec.Components.Settings.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, coreV1.LocalObjectReference{Name: secretName})
	}
	for _, secretName := range additionalSecrets {
		imagePullSecrets = append(imagePullSecrets, coreV1.LocalObjectReference{Name: secretName})
	}
	return imagePullSecrets
}

// objectsDiffer compares the two LocalObjectReference slice for equality, ignoring the order of the elements.
func objectsDiffer(actual, desired []coreV1.LocalObjectReference) bool {
	if len(actual) != len(desired) {
		return true
	}

	actualMap := make(map[string]struct{})
	for _, a := range actual {
		actualMap[a.Name] = struct{}{}
	}
	for _, d := range desired {
		if _, ok := actualMap[d.Name]; !ok {
			return true
		}
	}
	return false
}
