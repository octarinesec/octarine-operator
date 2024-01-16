package controllers

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	coreV1 "k8s.io/api/core/v1"
	"strings"
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

	httpProxyLen := 0
	if proxy.HttpProxy != nil {
		httpProxyLen = len(strings.TrimSpace(*(proxy.HttpProxy)))
	}

	httpsProxyLen := 0
	if proxy.HttpsProxy != nil {
		httpsProxyLen = len(strings.TrimSpace(*(proxy.HttpsProxy)))
	}

	// Don't set NoProxySuffix default value if we don't have any proxies defined
	if httpProxyLen+httpsProxyLen == 0 {
		return nil
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
