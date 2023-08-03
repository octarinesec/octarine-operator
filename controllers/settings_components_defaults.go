package controllers

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	coreV1 "k8s.io/api/core/v1"
)

func (r *CBContainersAgentController) setSettingsComponentsDefaults(settings *cbcontainersv1.CBContainersComponentsSettings) error {
	if settings.Proxy == nil {
		settings.Proxy = new(cbcontainersv1.CBContainersProxySettings)
	}
	if err := r.setProxySettingsComponentsDefaults(settings.Proxy); err != nil {
		return err
	}

	if settings.CreateDefaultImagePullSecrets == nil {
		settings.CreateDefaultImagePullSecrets = &trueRef
	}

	if settings.DaemonSetsTolerations == nil || len(settings.DaemonSetsTolerations) == 0 {
		settings.DaemonSetsTolerations = []coreV1.Toleration{
			{Operator: coreV1.TolerationOpExists},
		}
	}

	return nil
}

func (r *CBContainersAgentController) setProxySettingsComponentsDefaults(proxy *cbcontainersv1.CBContainersProxySettings) error {
	if proxy.Enabled == nil {
		proxy.Enabled = &falseRef
	}

	if proxy.NoProxySuffix == nil {
		noProxySuffix, err := GetDefaultNoProxyValue(r.Namespace)
		if err != nil {
			return err
		}
		proxy.NoProxySuffix = &noProxySuffix
	}

	return nil
}
