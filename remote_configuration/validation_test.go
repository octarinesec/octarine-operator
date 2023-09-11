package remote_configuration_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/remote_configuration"
	"testing"
)

var (
	trueV    = true
	truePtr  = &trueV
	falseV   = false
	falsePtr = &falseV
)

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
