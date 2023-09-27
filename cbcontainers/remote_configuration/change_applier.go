package remote_configuration

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
)

// ApplyConfigChangeToCR will modify CR according to the values in the configuration change provided
func ApplyConfigChangeToCR(change models.ConfigurationChange, cr *cbcontainersv1.CBContainersAgent) {
	if change.AgentVersion != nil {
		cr.Spec.Version = *change.AgentVersion

		// We do not set the tag to the version as that would make it harder to upgrade manually
		// Instead, we reset any "custom" tags, which will fall back to the default (spec.Version)
		images := []*cbcontainersv1.CBContainersImageSpec{
			&cr.Spec.Components.Basic.Monitor.Image,
			&cr.Spec.Components.Basic.Enforcer.Image,
			&cr.Spec.Components.Basic.StateReporter.Image,
			&cr.Spec.Components.ClusterScanning.ImageScanningReporter.Image,
			&cr.Spec.Components.ClusterScanning.ClusterScannerAgent.Image,
			&cr.Spec.Components.RuntimeProtection.Sensor.Image,
			&cr.Spec.Components.RuntimeProtection.Resolver.Image,
		}
		if cr.Spec.Components.Cndr != nil {
			images = append(images, &cr.Spec.Components.Cndr.Sensor.Image)
		}

		for _, i := range images {
			i.Tag = ""
		}
	}
}
