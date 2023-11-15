package remote_configuration

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
)

// ApplyConfigChangeToCR will modify CR according to the values in the configuration change provided
// If sensorMetadata is provided, specific supported features will be enabled or disabled based on their compatibility with the requested agent version
func ApplyConfigChangeToCR(change models.ConfigurationChange, cr *cbcontainersv1.CBContainersAgent, sensorMetadata []models.SensorMetadata) {
	if change.AgentVersion != nil {
		cr.Spec.Version = *change.AgentVersion

		resetImageTagsInCR(cr)
		toggleFeaturesBasedOnCompatibility(cr, *change.AgentVersion, sensorMetadata)
	}
}

func resetImageTagsInCR(cr *cbcontainersv1.CBContainersAgent) {
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

func toggleFeaturesBasedOnCompatibility(cr *cbcontainersv1.CBContainersAgent, version string, sensorMetadata []models.SensorMetadata) {
	var sensorMetadataForVersion *models.SensorMetadata
	for _, sensor := range sensorMetadata {
		if sensor.Version == version {
			sensorMetadataForVersion = &sensor
			break
		}
	}
	if sensorMetadataForVersion == nil {
		return
	}

	trueRef, falseRef := true, false

	if sensorMetadataForVersion.SupportsClusterScanning {
		cr.Spec.Components.ClusterScanning.Enabled = &trueRef
	} else {
		cr.Spec.Components.ClusterScanning.Enabled = &falseRef
	}

	if sensorMetadataForVersion.SupportsClusterScanningSecrets {
		cr.Spec.Components.ClusterScanning.ClusterScannerAgent.CLIFlags.EnableSecretDetection = true
	} else {
		cr.Spec.Components.ClusterScanning.ClusterScannerAgent.CLIFlags.EnableSecretDetection = false
	}

	if sensorMetadataForVersion.SupportsRuntime {
		cr.Spec.Components.RuntimeProtection.Enabled = &trueRef
	} else {
		cr.Spec.Components.RuntimeProtection.Enabled = &falseRef
	}

	if cr.Spec.Components.Cndr == nil {
		cr.Spec.Components.Cndr = &cbcontainersv1.CBContainersCndrSpec{}
	}
	if sensorMetadataForVersion.SupportsCndr {
		cr.Spec.Components.Cndr.Enabled = &trueRef
	} else {
		cr.Spec.Components.Cndr.Enabled = &falseRef
	}
}
