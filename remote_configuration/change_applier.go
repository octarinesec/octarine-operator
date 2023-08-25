package remote_configuration

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
)

// TODO: Move somewhere else

type Sensor struct {
	Version                 string `json:"version" yaml:"version"`
	IsLatest                bool   `json:"is_latest" yaml:"isLatest"`
	SupportsRuntime         bool   `json:"supports_runtime" yaml:"supportsRuntime"`
	SupportsClusterScanning bool   `json:"supports_cluster_scanning" yaml:"supportsClusterScanning"`
	SupportsCndr            bool   `json:"supports_cndr" yaml:"supportsCndr"`
}

type TODO struct {
	SensorData                []Sensor
	OperatorCompatibilityData models.OperatorCompatibility
}

func (todo *TODO) ValidateChange(change ConfigurationChange, cr *cbcontainersv1.CBContainersAgent) (bool, string) {
	var versionToValidate string

	// If the change will be modifying the agent version as well, we need to check what the _new_ version supports
	if change.AgentVersion != nil {
		versionToValidate = *change.AgentVersion
	} else {
		// Otherwise the current agent must actually work with the requested features
		versionToValidate = cr.Spec.Version
	}

	if sensorAndOperatorCompatible, msg := todo.validateOperatorAndSensorVersionCompatibility(versionToValidate); !sensorAndOperatorCompatible {
		return false, msg
	}

	return todo.validateSensorAndFeatureCompatibility(versionToValidate, change, cr)
}

func (todo *TODO) findMatchingSensor(sensorVersion string) (*Sensor, string) {
	for _, sensor := range todo.SensorData {
		if sensor.Version == sensorVersion {
			return &sensor, ""
		}
	}

	return nil, fmt.Sprintf("could not find sensor metadata for version %s", sensorVersion)
}

func (todo *TODO) validateOperatorAndSensorVersionCompatibility(sensorVersion string) (bool, string) {
	if err := todo.OperatorCompatibilityData.CheckCompatibility(sensorVersion); err != nil {
		return false, err.Error()
	}
	return true, ""
}

func (todo *TODO) validateSensorAndFeatureCompatibility(targetVersion string, change ConfigurationChange, cr *cbcontainersv1.CBContainersAgent) (bool, string) {
	sensor, msg := todo.findMatchingSensor(targetVersion)
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

func applyChangesToCR(change ConfigurationChange, cr *cbcontainersv1.CBContainersAgent) {
	// TODO: Validation?

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
}
