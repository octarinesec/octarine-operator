package remote_configuration_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/cbcontainers/remote_configuration"
	"math/rand"
	"strconv"
	"testing"
)

func TestVersionIsAppliedCorrectly(t *testing.T) {
	originalVersion := "my-version-42"
	newVersion := "new-version"
	cr := cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Version: originalVersion}}
	change := models.ConfigurationChange{AgentVersion: &newVersion}

	remote_configuration.ApplyConfigChangeToCR(change, &cr, nil)
	assert.Equal(t, newVersion, cr.Spec.Version)
}

func TestMissingVersionDoesNotModifyCR(t *testing.T) {
	originalVersion := "my-version-42"
	cr := cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Version: originalVersion}}
	change := models.ConfigurationChange{AgentVersion: nil}

	remote_configuration.ApplyConfigChangeToCR(change, &cr, nil)
	assert.Equal(t, originalVersion, cr.Spec.Version)

}

func TestVersionOverwritesCustomTagsByRemovingThem(t *testing.T) {
	cr := cbcontainersv1.CBContainersAgent{
		Spec: cbcontainersv1.CBContainersAgentSpec{
			Version: "some-version",
			Components: cbcontainersv1.CBContainersComponentsSpec{
				Basic: cbcontainersv1.CBContainersBasicSpec{
					Enforcer: cbcontainersv1.CBContainersEnforcerSpec{
						Image: cbcontainersv1.CBContainersImageSpec{
							Tag: "custom-enforcer",
						},
					},
					StateReporter: cbcontainersv1.CBContainersStateReporterSpec{
						Image: cbcontainersv1.CBContainersImageSpec{
							Tag: "custom-state-repoter",
						},
					},
					Monitor: cbcontainersv1.CBContainersMonitorSpec{
						Image: cbcontainersv1.CBContainersImageSpec{
							Tag: "custom-monitor",
						},
					},
				},
				RuntimeProtection: cbcontainersv1.CBContainersRuntimeProtectionSpec{
					Resolver: cbcontainersv1.CBContainersRuntimeResolverSpec{
						Image: cbcontainersv1.CBContainersImageSpec{
							Tag: "custom-runtime-resolver",
						},
					},
					Sensor: cbcontainersv1.CBContainersRuntimeSensorSpec{
						Image: cbcontainersv1.CBContainersImageSpec{
							Tag: "custom-runtime-sensor",
						},
					},
				},
				Cndr: &cbcontainersv1.CBContainersCndrSpec{
					Sensor: cbcontainersv1.CBContainersCndrSensorSpec{
						Image: cbcontainersv1.CBContainersImageSpec{
							Tag: "custom-cndr-sensor",
						},
					},
				},
				ClusterScanning: cbcontainersv1.CBContainersClusterScanningSpec{
					ClusterScannerAgent: cbcontainersv1.CBContainersClusterScannerAgentSpec{
						Image: cbcontainersv1.CBContainersImageSpec{
							Tag: "custom-cluster-scanning-agent",
						},
					},
					ImageScanningReporter: cbcontainersv1.CBContainersImageScanningReporterSpec{
						Image: cbcontainersv1.CBContainersImageSpec{
							Tag: "custom-image-scanning-reporter",
						},
					},
				},
			},
		},
	}

	newVersion := "new-version"
	change := models.ConfigurationChange{AgentVersion: &newVersion}

	remote_configuration.ApplyConfigChangeToCR(change, &cr, nil)
	assert.Equal(t, newVersion, cr.Spec.Version)
	// To avoid keeping "custom" tags forever, the apply change should instead reset all such fields
	// => the operator will use the common version instead
	assert.Empty(t, cr.Spec.Components.Basic.Monitor.Image.Tag)
	assert.Empty(t, cr.Spec.Components.Basic.Enforcer.Image.Tag)
	assert.Empty(t, cr.Spec.Components.Basic.StateReporter.Image.Tag)
	assert.Empty(t, cr.Spec.Components.ClusterScanning.ImageScanningReporter.Image.Tag)
	assert.Empty(t, cr.Spec.Components.ClusterScanning.ClusterScannerAgent.Image.Tag)
	assert.Empty(t, cr.Spec.Components.RuntimeProtection.Sensor.Image.Tag)
	assert.Empty(t, cr.Spec.Components.RuntimeProtection.Resolver.Image.Tag)
	assert.Empty(t, cr.Spec.Components.Cndr.Sensor.Image.Tag)
}

func TestFeatureToggles(t *testing.T) {
	testCases := []struct {
		name                string
		sensorCompatibility models.SensorMetadata
		assert              func(agent *cbcontainersv1.CBContainersAgent)
	}{
		{
			name: "cluster scanner supported, should enable",
			sensorCompatibility: models.SensorMetadata{
				SupportsClusterScanning: true,
			},
			assert: func(agent *cbcontainersv1.CBContainersAgent) {
				require.NotNil(t, agent.Spec.Components.ClusterScanning.Enabled)
				assert.True(t, *agent.Spec.Components.ClusterScanning.Enabled)
			},
		},
		{
			name: "cluster scanner not supported, should disable",
			sensorCompatibility: models.SensorMetadata{
				SupportsClusterScanning: false,
			},
			assert: func(agent *cbcontainersv1.CBContainersAgent) {
				require.NotNil(t, agent.Spec.Components.ClusterScanning.Enabled)
				assert.False(t, *agent.Spec.Components.ClusterScanning.Enabled)
			},
		},
		{
			name: "secret scanning supported, should enable",
			sensorCompatibility: models.SensorMetadata{
				SupportsClusterScanningSecrets: true,
			},
			assert: func(agent *cbcontainersv1.CBContainersAgent) {
				assert.True(t, agent.Spec.Components.ClusterScanning.ClusterScannerAgent.CLIFlags.EnableSecretDetection)
			},
		},
		{
			name: "secret scanning not supported, should disable",
			sensorCompatibility: models.SensorMetadata{
				SupportsClusterScanningSecrets: false,
			},
			assert: func(agent *cbcontainersv1.CBContainersAgent) {
				assert.False(t, agent.Spec.Components.ClusterScanning.ClusterScannerAgent.CLIFlags.EnableSecretDetection)
			},
		},
		{
			name: "runtime protection supported, should enable",
			sensorCompatibility: models.SensorMetadata{
				SupportsRuntime: true,
			},
			assert: func(agent *cbcontainersv1.CBContainersAgent) {
				require.NotNil(t, agent.Spec.Components.RuntimeProtection.Enabled)
				assert.True(t, *agent.Spec.Components.RuntimeProtection.Enabled)
			},
		},
		{
			name: "runtime protection not supported, should disable",
			sensorCompatibility: models.SensorMetadata{
				SupportsRuntime: false,
			},
			assert: func(agent *cbcontainersv1.CBContainersAgent) {
				require.NotNil(t, agent.Spec.Components.RuntimeProtection.Enabled)
				assert.False(t, *agent.Spec.Components.RuntimeProtection.Enabled)
			},
		},
		{
			name: "CNDR supported, should enable",
			sensorCompatibility: models.SensorMetadata{
				SupportsCndr: true,
			},
			assert: func(agent *cbcontainersv1.CBContainersAgent) {
				require.NotNil(t, agent.Spec.Components.Cndr)
				require.NotNil(t, agent.Spec.Components.Cndr.Enabled)
				assert.True(t, *agent.Spec.Components.Cndr.Enabled)
			},
		},
		{
			name: "CNDR not supported, should disable",
			sensorCompatibility: models.SensorMetadata{
				SupportsCndr: false,
			},
			assert: func(agent *cbcontainersv1.CBContainersAgent) {
				require.NotNil(t, agent.Spec.Components.Cndr)
				require.NotNil(t, agent.Spec.Components.Cndr.Enabled)
				assert.False(t, *agent.Spec.Components.Cndr.Enabled)
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			version := "2.3.4"
			cr := &cbcontainersv1.CBContainersAgent{}
			change := models.ConfigurationChange{AgentVersion: &version}
			tC.sensorCompatibility.Version = version

			remote_configuration.ApplyConfigChangeToCR(change, cr, []models.SensorMetadata{tC.sensorCompatibility})

			tC.assert(cr)
		})
	}
}

// randomPendingConfigChange creates a non-empty configuration change with randomly populated fields in pending state
// the change is not guaranteed to be 100% valid
func randomPendingConfigChange() models.ConfigurationChange {
	var versions = []string{"2.12.1", "2.10.0", "2.12.0", "2.11.0", "3.0.0"}

	changeVersion := &versions[rand.Intn(len(versions))]

	return models.ConfigurationChange{
		ID:           strconv.Itoa(rand.Int()),
		AgentVersion: changeVersion,
		Status:       models.ChangeStatusPending,
	}
}
