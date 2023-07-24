package controllers

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	coreV1 "k8s.io/api/core/v1"
)

func (r *CBContainersAgentController) setSettingsComponentsDefaults(settings cbcontainersv1.CBContainersComponentsSettings) error {
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
