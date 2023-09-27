package remote_configuration_test

import (
	"github.com/stretchr/testify/assert"
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

	remote_configuration.ApplyConfigChangeToCR(change, &cr)
	assert.Equal(t, newVersion, cr.Spec.Version)
}

func TestMissingVersionDoesNotModifyCR(t *testing.T) {
	originalVersion := "my-version-42"
	cr := cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Version: originalVersion}}
	change := models.ConfigurationChange{AgentVersion: nil}

	remote_configuration.ApplyConfigChangeToCR(change, &cr)
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

	remote_configuration.ApplyConfigChangeToCR(change, &cr)
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
