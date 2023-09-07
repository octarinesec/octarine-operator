package remote_configuration

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
)

func ApplyChangeToCR(change models.ConfigurationChange, cr *cbcontainersv1.CBContainersAgent) {
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

// TODO: Move

type invalidChangeError struct {
	msg string
}

func (i invalidChangeError) Error() string {
	return i.msg
}

func NewConfigurationChangeValidator(operatorVersion string, api ApiGateway) (*ConfigurationChangeValidator, error) {
	compatibilityMatrix, err := api.GetCompatibilityMatrixEntryFor(operatorVersion)
	if err != nil {
		return nil, err
	}

	sensors, err := api.GetSensorMetadata()
	if err != nil {
		return nil, err
	}

	// TODO: Dereference
	return &ConfigurationChangeValidator{
		SensorData:                sensors,
		OperatorCompatibilityData: *compatibilityMatrix,
	}, nil
}

type ConfigurationChangeValidator struct {
	SensorData                []models.SensorMetadata
	OperatorCompatibilityData models.OperatorCompatibility
}

func (validator *ConfigurationChangeValidator) ValidateChange(change models.ConfigurationChange, cr *cbcontainersv1.CBContainersAgent) error {
	var versionToValidate string

	// If the change will be modifying the agent version as well, we need to check what the _new_ version supports
	if change.AgentVersion != nil {
		versionToValidate = *change.AgentVersion
	} else {
		// Otherwise the current agent must actually work with the requested features
		versionToValidate = cr.Spec.Version
	}

	if err := validator.validateOperatorAndSensorVersionCompatibility(versionToValidate); err != nil {
		return err
	}

	return validator.validateSensorAndFeatureCompatibility(versionToValidate, change)
}

func (validator *ConfigurationChangeValidator) findMatchingSensor(sensorVersion string) *models.SensorMetadata {
	for _, sensor := range validator.SensorData {
		if sensor.Version == sensorVersion {
			return &sensor
		}
	}

	return nil
}

func (validator *ConfigurationChangeValidator) validateOperatorAndSensorVersionCompatibility(sensorVersion string) error {
	if err := validator.OperatorCompatibilityData.CheckCompatibility(sensorVersion); err != nil {
		return invalidChangeError{msg: err.Error()}
	}
	return nil
}

func (validator *ConfigurationChangeValidator) validateSensorAndFeatureCompatibility(targetVersion string, change models.ConfigurationChange) error {
	sensor := validator.findMatchingSensor(targetVersion)
	if sensor == nil {
		return fmt.Errorf("could not find sensor metadata for version %s", targetVersion)
	}

	if change.EnableClusterScanning != nil &&
		*change.EnableClusterScanning == true &&
		!sensor.SupportsClusterScanning {
		return invalidChangeError{msg: fmt.Sprintf("sensor version %s does not support cluster scanning feature", targetVersion)}
	}

	if change.EnableRuntime != nil &&
		*change.EnableRuntime == true &&
		!sensor.SupportsRuntime {
		return invalidChangeError{msg: fmt.Sprintf("sensor version %s does not support runtime protection feature", targetVersion)}
	}

	if change.EnableCNDR != nil &&
		*change.EnableCNDR == true &&
		!sensor.SupportsCndr {
		return invalidChangeError{msg: fmt.Sprintf("sensor version %s does not support cloud-native detect and response feature", targetVersion)}
	}

	return nil
}
