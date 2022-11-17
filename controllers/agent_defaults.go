package controllers

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
)

func (r *CBContainersAgentController) setAgentDefaults(agentSpec *cbcontainersv1.CBContainersAgentSpec) error {
	if agentSpec.AccessTokenSecretName == "" {
		agentSpec.AccessTokenSecretName = defaultAccessToken
	}

	r.setGatewaysDefaults(&agentSpec.Gateways)

	if err := r.setComponentsDefaults(&agentSpec.Components); err != nil {
		return err
	}

	// The namespace field of the agent spec should always be populated, because it has a default value,
	// but just in case include this check here in case it turns out to be empty in the future.
	// By default all objects have the "cbcontainers-dataplane" as namespace.
	if agentSpec.Namespace == "" {
		agentSpec.Namespace = common.DataPlaneNamespaceName
	}

	return nil
}

func (r *CBContainersAgentController) setComponentsDefaults(components *cbcontainersv1.CBContainersComponentsSpec) error {
	if err := r.setBasicComponentsDefaults(&components.Basic); err != nil {
		return err
	}

	if err := r.setRuntimeProtectionComponentsDefaults(&components.RuntimeProtection); err != nil {
		return err
	}

	if err := r.setClusterScanningComponentsDefaults(&components.ClusterScanning); err != nil {
		return err
	}

	return nil
}
