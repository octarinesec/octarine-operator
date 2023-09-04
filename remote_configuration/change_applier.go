package remote_configuration

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
)

// TODO: Move somewhere else

type Sensor struct {
	Version                 string `json:"version"`
	IsLatest                bool   `json:"is_latest" `
	SupportsRuntime         bool   `json:"supports_runtime"`
	SupportsClusterScanning bool   `json:"supports_cluster_scanning"`
	SupportsCndr            bool   `json:"supports_cndr"`
}

type CustomResourceChanger struct {
	SensorData                []Sensor
	OperatorCompatibilityData models.OperatorCompatibility
}

func (changer *CustomResourceChanger) ValidateChange(change ConfigurationChange, cr *cbcontainersv1.CBContainersAgent) (bool, string) {
	var versionToValidate string

	// If the change will be modifying the agent version as well, we need to check what the _new_ version supports
	if change.AgentVersion != nil {
		versionToValidate = *change.AgentVersion
	} else {
		// Otherwise the current agent must actually work with the requested features
		versionToValidate = cr.Spec.Version
	}

	if sensorAndOperatorCompatible, msg := changer.validateOperatorAndSensorVersionCompatibility(versionToValidate); !sensorAndOperatorCompatible {
		return false, msg
	}

	return changer.validateSensorAndFeatureCompatibility(versionToValidate, change)
}

func (changer *CustomResourceChanger) ApplyChangeToCR(change ConfigurationChange, cr *cbcontainersv1.CBContainersAgent) error {
	if isValid, msg := changer.ValidateChange(change, cr); !isValid {
		return fmt.Errorf("provided change cannot be applied to the custom resource with reason (%s)", msg)
	}

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
	if change.EnableRuntime != nil {
		cr.Spec.Components.RuntimeProtection.Enabled = change.EnableRuntime
	}
	if change.EnableCNDR != nil {
		if cr.Spec.Components.Cndr == nil {
			cr.Spec.Components.Cndr = &cbcontainersv1.CBContainersCndrSpec{}
		}
		cr.Spec.Components.Cndr.Enabled = change.EnableCNDR
	}

	return nil
}

func (changer *CustomResourceChanger) findMatchingSensor(sensorVersion string) (*Sensor, string) {
	for _, sensor := range changer.SensorData {
		if sensor.Version == sensorVersion {
			return &sensor, ""
		}
	}

	return nil, fmt.Sprintf("could not find sensor metadata for version %s", sensorVersion)
}

func (changer *CustomResourceChanger) validateOperatorAndSensorVersionCompatibility(sensorVersion string) (bool, string) {
	if err := changer.OperatorCompatibilityData.CheckCompatibility(sensorVersion); err != nil {
		return false, err.Error()
	}
	return true, ""
}

func (changer *CustomResourceChanger) validateSensorAndFeatureCompatibility(targetVersion string, change ConfigurationChange) (bool, string) {
	sensor, msg := changer.findMatchingSensor(targetVersion)
	if sensor == nil {
		return false, msg
	}

	if change.EnableClusterScanning != nil &&
		*change.EnableClusterScanning == true &&
		!sensor.SupportsClusterScanning {
		return false, fmt.Sprintf("sensor version %s does not support cluster scanning feature", targetVersion)
	}

	if change.EnableRuntime != nil &&
		*change.EnableRuntime == true &&
		!sensor.SupportsRuntime {
		return false, fmt.Sprintf("sensor version %s does not support runtime protection feature", targetVersion)
	}

	if change.EnableCNDR != nil &&
		*change.EnableCNDR == true &&
		!sensor.SupportsCndr {
		return false, fmt.Sprintf("sensor version %s does not support cloud-native detect and response feature", targetVersion)
	}

	return true, ""
}
