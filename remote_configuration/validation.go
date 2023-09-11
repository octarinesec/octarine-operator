package remote_configuration

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
)

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

	if change.EnableClusterScanningSecretDetection != nil &&
		*change.EnableClusterScanningSecretDetection == true &&
		!sensor.SupportsClusterScanningSecrets {
		return invalidChangeError{msg: fmt.Sprintf("sensor version %s does not support secret detection during cluster scanning feature", targetVersion)}
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
