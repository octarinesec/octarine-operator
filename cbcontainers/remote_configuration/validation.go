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

// EmptyConfigurationChangeValidator can be used when no validation should occur
type EmptyConfigurationChangeValidator struct {
}

func (emptyValidator *EmptyConfigurationChangeValidator) ValidateChange(change models.ConfigurationChange, cr *cbcontainersv1.CBContainersAgent) error {
	return nil
}

type ConfigurationChangeValidator struct {
	OperatorCompatibilityData models.OperatorCompatibility
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

func (validator *ConfigurationChangeValidator) ValidateChange(change models.ConfigurationChange, cr *cbcontainersv1.CBContainersAgent) error {
	var versionToValidate string

	// If the change will be modifying the agent version as well, we need to check what the _new_ version supports
	if change.AgentVersion != nil {
		versionToValidate = *change.AgentVersion
	} else {
		// Otherwise the current agent must actually work with the requested features
		versionToValidate = cr.Spec.Version
	}

	return validator.OperatorCompatibilityData.CheckCompatibility(models.Version(versionToValidate))
}
