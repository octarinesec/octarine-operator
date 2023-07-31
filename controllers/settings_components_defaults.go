package controllers

import (
	"fmt"
	"strings"

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
		const k8sAPIServiceDomain = "kubernetes.default.svc"

		// We use a DNS lookup to get an accurate IP list of the API server
		noProxyItems, err := netLookupHost(k8sAPIServiceDomain)
		if err != nil {
			return fmt.Errorf("unable to fetch Kubernetes API server addresses by querying %q: %w", k8sAPIServiceDomain, err)
		}

		// we should be able to connect to a <service-name>.svc.cluster.local without
		// the need for a proxy
		noProxyItems = append(noProxyItems, r.Namespace+".svc.cluster.local")
		noProxySuffix := strings.Join(noProxyItems, ",")
		proxy.NoProxySuffix = &noProxySuffix
	}

	return nil
}
