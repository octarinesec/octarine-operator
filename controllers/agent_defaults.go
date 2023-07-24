package controllers

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
)

func (r *CBContainersAgentController) setAgentDefaults(agentSpec *cbcontainersv1.CBContainersAgentSpec) error {
	if agentSpec.AccessTokenSecretName == "" {
		agentSpec.AccessTokenSecretName = defaultAccessToken
	}

	r.setGatewaysDefaults(&agentSpec.Gateways)

	if err := r.setComponentsDefaults(&agentSpec.Components); err != nil {
		return err
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

	if err := r.setCndrComponentsDefaults(components.Cndr); err != nil {
		return err
	}

	if err := r.setSettingsComponentsDefaults(components.Settings); err != nil {
		return err
	}

	return nil
}
