package remote_configuration

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
)

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
	if compatibilityMatrix == nil {
		return nil, fmt.Errorf("compatibility matrix API returned no data but no error as well, cannot continue")
	}

	return &ConfigurationChangeValidator{
		OperatorCompatibilityData: *compatibilityMatrix,
	}, nil
}

type ConfigurationChangeValidator struct {
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

	return validator.validateOperatorAndSensorVersionCompatibility(versionToValidate)
}

func (validator *ConfigurationChangeValidator) validateOperatorAndSensorVersionCompatibility(sensorVersion string) error {
	if err := validator.OperatorCompatibilityData.CheckCompatibility(sensorVersion); err != nil {
		return invalidChangeError{msg: err.Error()}
	}
	return nil
}
