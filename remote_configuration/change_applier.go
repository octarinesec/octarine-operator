package remote_configuration

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
)

// ApplyConfigChangeToCR will modify CR according to the values in the configuration change provided
func ApplyConfigChangeToCR(change models.ConfigurationChange, cr *cbcontainersv1.CBContainersAgent) {
	resetVersion := func(ptrToField *string) {
		if ptrToField != nil && *ptrToField != "" {
			*ptrToField = ""
		}
	}

	if change.AgentVersion != nil {
		cr.Spec.Version = *change.AgentVersion

		resetVersion(&cr.Spec.Components.Basic.Monitor.Image.Tag)
		resetVersion(&cr.Spec.Components.Basic.Enforcer.Image.Tag)
		resetVersion(&cr.Spec.Components.Basic.StateReporter.Image.Tag)
		resetVersion(&cr.Spec.Components.ClusterScanning.ImageScanningReporter.Image.Tag)
		resetVersion(&cr.Spec.Components.ClusterScanning.ClusterScannerAgent.Image.Tag)
		resetVersion(&cr.Spec.Components.RuntimeProtection.Sensor.Image.Tag)
		resetVersion(&cr.Spec.Components.RuntimeProtection.Resolver.Image.Tag)
		if cr.Spec.Components.Cndr != nil {
			resetVersion(&cr.Spec.Components.Cndr.Sensor.Image.Tag)
		}
	}
	if change.EnableClusterScanning != nil {
		cr.Spec.Components.ClusterScanning.Enabled = change.EnableClusterScanning
	}

	if change.EnableClusterScanningSecretDetection != nil {
		cr.Spec.Components.ClusterScanning.ClusterScannerAgent.CLIFlags.EnableSecretDetection = *change.EnableClusterScanningSecretDetection
	}

	if change.EnableRuntime != nil {
		cr.Spec.Components.RuntimeProtection.Enabled = change.EnableRuntime
	}

	if change.EnableCNDR != nil {
		if cr.Spec.Components.Cndr == nil {
			cr.Spec.Components.Cndr = &cbcontainersv1.CBContainersCndrSpec{}
		}
		cr.Spec.Components.Cndr.Enabled = change.EnableCNDR
	}
}
