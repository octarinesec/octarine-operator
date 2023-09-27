package remote_configuration_test

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/cbcontainers/remote_configuration"
	"github.com/vmware/cbcontainers-operator/cbcontainers/remote_configuration/mocks"
	"testing"
)

func TestValidatorConstructorReturnsErrOnFailures(t *testing.T) {
	expectedOperatorVersion := "5.0.0"

	testCases := []struct {
		name             string
		setupGatewayMock func(gateway *mocks.MockApiGateway)
	}{
		{
			name: "get compatibility returns err",
			setupGatewayMock: func(gateway *mocks.MockApiGateway) {
				gateway.EXPECT().GetCompatibilityMatrixEntryFor(expectedOperatorVersion).Return(nil, errors.New("some error")).AnyTimes()
			},
		},
		{
			name: "get compatibility returns nil",
			setupGatewayMock: func(gateway *mocks.MockApiGateway) {
				gateway.EXPECT().GetCompatibilityMatrixEntryFor(expectedOperatorVersion).Return(nil, nil).AnyTimes()
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGateway := mocks.NewMockApiGateway(ctrl)
			tC.setupGatewayMock(mockGateway)

			validator, err := remote_configuration.NewConfigurationChangeValidator(expectedOperatorVersion, mockGateway)

			assert.Nil(t, validator)
			assert.Error(t, err)
		})
	}
}

func TestValidateFailsIfSensorAndOperatorAreNotCompatible(t *testing.T) {
	testCases := []struct {
		name                  string
		versionToApply        string
		operatorCompatibility models.OperatorCompatibility
	}{
		{
			name:           "sensor version is too high",
			versionToApply: "5.0.0",
			operatorCompatibility: models.OperatorCompatibility{
				MinAgent: models.AgentMinVersionNone,
				MaxAgent: "4.0.0",
			},
		},
		{
			name:           "sensor version is too low",
			versionToApply: "0.9",
			operatorCompatibility: models.OperatorCompatibility{
				MinAgent: "1.0.0",
				MaxAgent: models.AgentMaxVersionLatest,
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			target := remote_configuration.ConfigurationChangeValidator{
				OperatorCompatibilityData: tC.operatorCompatibility,
			}

			change := models.ConfigurationChange{AgentVersion: &tC.versionToApply}
			cr := &cbcontainersv1.CBContainersAgent{}

			err := target.ValidateChange(change, cr)
			assert.Error(t, err)
		})
	}
}

func TestValidateSucceedsIfSensorAndOperatorAreCompatible(t *testing.T) {
	testCases := []struct {
		name                  string
		versionToApply        string
		operatorCompatibility models.OperatorCompatibility
	}{
		{
			name:           "sensor version is at lower end",
			versionToApply: "5.0.0",
			operatorCompatibility: models.OperatorCompatibility{
				MinAgent: "5.0.0",
				MaxAgent: "6.0.0",
			},
		},
		{
			name:           "sensor version is at upper end",
			versionToApply: "0.9",
			operatorCompatibility: models.OperatorCompatibility{
				MinAgent: "0.1.0",
				MaxAgent: "0.9.0",
			},
		},
		{
			name:           "sensor version is within range",
			versionToApply: "2.3.4",
			operatorCompatibility: models.OperatorCompatibility{
				MinAgent: "1.0.0",
				MaxAgent: "2.4",
			},
		},
		{
			name:           "operator supports 'infinite' versions",
			versionToApply: "5.0.0",
			operatorCompatibility: models.OperatorCompatibility{
				MinAgent: models.AgentMinVersionNone,
				MaxAgent: models.AgentMaxVersionLatest,
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			target := remote_configuration.ConfigurationChangeValidator{
				OperatorCompatibilityData: tC.operatorCompatibility,
			}

			change := models.ConfigurationChange{AgentVersion: &tC.versionToApply}
			cr := &cbcontainersv1.CBContainersAgent{}

			err := target.ValidateChange(change, cr)
			assert.NoError(t, err)
		})
	}
}
