package remote_configuration_test

import (
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/remote_configuration"
	"github.com/vmware/cbcontainers-operator/remote_configuration/mocks"
	"testing"
)

var (
	trueV    = true
	truePtr  = &trueV
	falseV   = false
	falsePtr = &falseV
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
				gateway.EXPECT().GetSensorMetadata().Return([]models.SensorMetadata{}, nil).AnyTimes()
			},
		},
		{
			name: "get sensor metadata returns err",
			setupGatewayMock: func(gateway *mocks.MockApiGateway) {
				gateway.EXPECT().GetCompatibilityMatrixEntryFor(expectedOperatorVersion).Return(&models.OperatorCompatibility{}, nil).AnyTimes()
				gateway.EXPECT().GetSensorMetadata().Return(nil, errors.New("some error")).AnyTimes()
			},
		},
		{
			name: "get compatibility returns nil",
			setupGatewayMock: func(gateway *mocks.MockApiGateway) {
				gateway.EXPECT().GetCompatibilityMatrixEntryFor(expectedOperatorVersion).Return(nil, nil).AnyTimes()
				gateway.EXPECT().GetSensorMetadata().Return([]models.SensorMetadata{}, nil).AnyTimes()
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

func TestValidateFailsIfSensorDoesNotSupportRequestedFeature(t *testing.T) {
	testCases := []struct {
		name       string
		change     models.ConfigurationChange
		sensorMeta models.SensorMetadata
	}{
		{
			name: "cluster scanning",
			change: models.ConfigurationChange{
				EnableClusterScanning: truePtr,
			},
			sensorMeta: models.SensorMetadata{
				SupportsClusterScanning: false,
			},
		},
		{
			name: "cluster scanning secret detection",
			change: models.ConfigurationChange{
				EnableClusterScanningSecretDetection: truePtr,
			},
			sensorMeta: models.SensorMetadata{
				SupportsClusterScanningSecrets: false,
			},
		},
		{
			name: "runtime protection",
			change: models.ConfigurationChange{
				EnableRuntime: truePtr,
			},
			sensorMeta: models.SensorMetadata{
				SupportsRuntime: false,
			},
		},
		{
			name: "CNDR",
			change: models.ConfigurationChange{
				EnableCNDR: truePtr,
			},
			sensorMeta: models.SensorMetadata{
				SupportsCndr: false,
			},
		},
	}

	for _, tC := range testCases {
		version := "dummy-version"
		tC.sensorMeta.Version = version
		target := remote_configuration.ConfigurationChangeValidator{
			SensorData: []models.SensorMetadata{tC.sensorMeta},
		}

		t.Run(fmt.Sprintf("no version in change, %s not supported by current agent", tC.name), func(t *testing.T) {
			tC.change.AgentVersion = nil
			cr := &cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Version: version}}

			err := target.ValidateChange(tC.change, cr)

			assert.Error(t, err)
		})

		t.Run(fmt.Sprintf("change also applies agent version, %s not supported by that version", tC.name), func(t *testing.T) {
			tC.change.AgentVersion = &version
			cr := &cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Version: "some-other-verson"}}

			err := target.ValidateChange(tC.change, cr)

			assert.Error(t, err)
		})
	}
}

func TestValidateFailsIfSensorIsNotInList(t *testing.T) {
	sensorMetaWithoutTargetSensor := []models.SensorMetadata{{
		Version:                        "1.0.0",
		IsLatest:                       false,
		SupportsRuntime:                true,
		SupportsClusterScanning:        true,
		SupportsClusterScanningSecrets: true,
		SupportsCndr:                   true,
	}}
	operatorSupportsAll := models.OperatorCompatibility{
		MinAgent: models.AgentMinVersionNone,
		MaxAgent: models.AgentMaxVersionLatest,
	}
	unknownVersion := "1.2.3"

	validator := remote_configuration.ConfigurationChangeValidator{
		SensorData:                sensorMetaWithoutTargetSensor,
		OperatorCompatibilityData: operatorSupportsAll,
	}

	change := models.ConfigurationChange{
		AgentVersion: &unknownVersion,
	}
	cr := &cbcontainersv1.CBContainersAgent{}

	assert.Error(t, validator.ValidateChange(change, cr))
}

func TestValidateSucceedsIfSensorSupportsRequestedFeature(t *testing.T) {
	testCases := []struct {
		name       string
		change     models.ConfigurationChange
		sensorMeta models.SensorMetadata
	}{
		{
			name: "cluster scanning",
			change: models.ConfigurationChange{
				EnableClusterScanning: truePtr,
			},
			sensorMeta: models.SensorMetadata{
				SupportsClusterScanning: true,
			},
		},
		{
			name: "cluster scanning secret detection",
			change: models.ConfigurationChange{
				EnableClusterScanningSecretDetection: truePtr,
			},
			sensorMeta: models.SensorMetadata{
				SupportsClusterScanningSecrets: true,
			},
		},
		{
			name: "runtime protection",
			change: models.ConfigurationChange{
				EnableRuntime: truePtr,
			},
			sensorMeta: models.SensorMetadata{
				SupportsRuntime: true,
			},
		},
		{
			name: "CNDR",
			change: models.ConfigurationChange{
				EnableCNDR: truePtr,
			},
			sensorMeta: models.SensorMetadata{
				SupportsCndr: true,
			},
		},
	}

	for _, tC := range testCases {
		version := "dummy-version"
		tC.sensorMeta.Version = version
		target := remote_configuration.ConfigurationChangeValidator{
			SensorData: []models.SensorMetadata{tC.sensorMeta},
		}

		t.Run(fmt.Sprintf("no version in change, %s is supported by current agent", tC.name), func(t *testing.T) {
			tC.change.AgentVersion = nil
			cr := &cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Version: version}}

			err := target.ValidateChange(tC.change, cr)

			assert.NoError(t, err)
		})

		t.Run(fmt.Sprintf("change also applies agent version, %s is supported by that version", tC.name), func(t *testing.T) {
			tC.change.AgentVersion = &version
			cr := &cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Version: "some-other-verson"}}

			err := target.ValidateChange(tC.change, cr)

			assert.NoError(t, err)
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
				SensorData:                []models.SensorMetadata{{Version: tC.versionToApply}},
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
				SensorData:                []models.SensorMetadata{{Version: tC.versionToApply}},
				OperatorCompatibilityData: tC.operatorCompatibility,
			}

			change := models.ConfigurationChange{AgentVersion: &tC.versionToApply}
			cr := &cbcontainersv1.CBContainersAgent{}

			err := target.ValidateChange(change, cr)
			assert.NoError(t, err)
		})
	}
}
