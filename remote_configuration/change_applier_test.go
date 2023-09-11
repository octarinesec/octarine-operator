package remote_configuration_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/remote_configuration"
	"math/rand"
	"strconv"
	"testing"
)

func TestFeatureTogglesAreAppliedCorrectly(t *testing.T) {
	type appliedChangeTest struct {
		name          string
		change        models.ConfigurationChange
		initialCR     cbcontainersv1.CBContainersAgent
		assertFinalCR func(*testing.T, cbcontainersv1.CBContainersAgent)
	}

	crVersion := "1.2.3"

	// generateFeatureToggleTestCases produces a set of tests for a single feature toggle in the requested change
	// The tests validate if each toggle state (true, false, nil) is applied correctly or ignored when it's not needed against the CR's state (true, false, nil)
	generateFeatureToggleTestCases :=
		func(feature string,
			changeFieldChanger func(*models.ConfigurationChange, *bool),
			crFieldChanger func(*cbcontainersv1.CBContainersAgent, *bool),
			crAsserter func(*testing.T, cbcontainersv1.CBContainersAgent, *bool)) []appliedChangeTest {

			var result []appliedChangeTest

			for _, crState := range []*bool{truePtr, falsePtr, nil} {
				crState := crState // Avoid closure issues

				// Validate that each toggle state works (or doesn't do anything when it matches)
				for _, changeState := range []*bool{falsePtr, truePtr} {
					changeState := changeState // Avoid closure issues

					cr := cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Version: crVersion}}
					crFieldChanger(&cr, crState)
					change := models.ConfigurationChange{}
					changeFieldChanger(&change, changeState)

					result = append(result, appliedChangeTest{
						name:      fmt.Sprintf("toggle feature (%s) from (%v) to (%v)", feature, prettyPrintBoolPtr(crState), prettyPrintBoolPtr(changeState)),
						change:    change,
						initialCR: cr,
						assertFinalCR: func(t *testing.T, agent cbcontainersv1.CBContainersAgent) {
							crAsserter(t, agent, changeState)
						},
					})
				}

				cr := cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Version: crVersion}}
				crFieldChanger(&cr, crState)
				// Validate that a change with the toggle unset does not modify the CR
				result = append(result, appliedChangeTest{
					name:      fmt.Sprintf("missing toggle feature (%s) with CR state (%v)", feature, prettyPrintBoolPtr(crState)),
					change:    models.ConfigurationChange{},
					initialCR: cr,
					assertFinalCR: func(t *testing.T, agent cbcontainersv1.CBContainersAgent) {
						crAsserter(t, agent, crState)
					},
				})
			}

			return result
		}

	var testCases []appliedChangeTest

	clusterScannerToggleTestCases := generateFeatureToggleTestCases("cluster scanning",
		func(change *models.ConfigurationChange, val *bool) {
			change.EnableClusterScanning = val
		}, func(agent *cbcontainersv1.CBContainersAgent, val *bool) {
			agent.Spec.Components.ClusterScanning.Enabled = val
		}, func(t *testing.T, agent cbcontainersv1.CBContainersAgent, b *bool) {
			assert.Equal(t, b, agent.Spec.Components.ClusterScanning.Enabled)
		})

	secretDetectionToggleTestCases := generateFeatureToggleTestCases("cluster scanning secret detection",
		func(change *models.ConfigurationChange, val *bool) {
			change.EnableClusterScanningSecretDetection = val
		}, func(agent *cbcontainersv1.CBContainersAgent, val *bool) {
			if val == nil {
				// Bail out, this value is not valid for the flag
				return
			}
			agent.Spec.Components.ClusterScanning.ClusterScannerAgent.CLIFlags.EnableSecretDetection = *val
		}, func(t *testing.T, agent cbcontainersv1.CBContainersAgent, b *bool) {
			if b == nil {
				// Bail out, this value is not valid for the flag
				return
			}
			assert.Equal(t, *b, agent.Spec.Components.ClusterScanning.ClusterScannerAgent.CLIFlags.EnableSecretDetection)
		})

	runtimeToggleTestCases := generateFeatureToggleTestCases("runtime protection",
		func(change *models.ConfigurationChange, val *bool) {
			change.EnableRuntime = val
		}, func(agent *cbcontainersv1.CBContainersAgent, val *bool) {
			agent.Spec.Components.RuntimeProtection.Enabled = val
		}, func(t *testing.T, agent cbcontainersv1.CBContainersAgent, b *bool) {
			assert.Equal(t, b, agent.Spec.Components.RuntimeProtection.Enabled)
		})

	cndrToggleTestCases := generateFeatureToggleTestCases("CNDR",
		func(change *models.ConfigurationChange, val *bool) {
			change.EnableCNDR = val
		}, func(agent *cbcontainersv1.CBContainersAgent, val *bool) {
			if agent.Spec.Components.Cndr == nil {
				agent.Spec.Components.Cndr = &cbcontainersv1.CBContainersCndrSpec{}
			}
			agent.Spec.Components.Cndr.Enabled = val
		}, func(t *testing.T, agent cbcontainersv1.CBContainersAgent, b *bool) {
			assert.Equal(t, b, agent.Spec.Components.Cndr.Enabled)
		})

	testCases = append(testCases, clusterScannerToggleTestCases...)
	testCases = append(testCases, secretDetectionToggleTestCases...)
	testCases = append(testCases, runtimeToggleTestCases...)
	testCases = append(testCases, cndrToggleTestCases...)

	t1 := testCases[0]
	t2 := testCases[1]
	t3 := testCases[2]
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(t1, t2, t3)

			target := remote_configuration.ChangeApplier{}

			target.ApplyConfigChangeToCR(testCase.change, &testCase.initialCR)
			testCase.assertFinalCR(t, testCase.initialCR)
		})
	}
}

func TestVersionIsAppliedCorrectly(t *testing.T) {
	originalVersion := "my-version-42"
	newVersion := "new-version"
	cr := cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Version: originalVersion}}
	change := models.ConfigurationChange{AgentVersion: &newVersion}

	target := remote_configuration.ChangeApplier{}

	target.ApplyConfigChangeToCR(change, &cr)
	assert.Equal(t, newVersion, cr.Spec.Version)
}

func TestMissingVersionDoesNotModifyCR(t *testing.T) {
	originalVersion := "my-version-42"
	cr := cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Version: originalVersion}}
	change := models.ConfigurationChange{AgentVersion: nil, EnableRuntime: truePtr}

	target := remote_configuration.ChangeApplier{}
	target.ApplyConfigChangeToCR(change, &cr)
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

	target := remote_configuration.ChangeApplier{}

	target.ApplyConfigChangeToCR(change, &cr)
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

func prettyPrintBoolPtr(v *bool) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%t", *v)
}

// randomPendingConfigChange creates a non-empty configuration change with randomly populated fields in pending state
// the change is not guaranteed to be 100% valid
func randomPendingConfigChange() models.ConfigurationChange {
	var versions = []string{"2.12.1", "2.10.0", "2.12.0", "2.11.0", "3.0.0"}

	csRand, runtimeRand, cndrRand, versionRand := rand.Int(), rand.Int(), rand.Int(), rand.Intn(len(versions))

	changeVersion := &versions[versionRand]

	var changeClusterScanning *bool
	var changeRuntime *bool
	var changeCNDR *bool

	switch csRand % 5 {
	case 1, 3:
		changeClusterScanning = truePtr
	case 2, 4:
		changeClusterScanning = falsePtr
	default:
		changeClusterScanning = nil
	}

	switch runtimeRand % 5 {
	case 1, 3:
		changeRuntime = truePtr
	case 2, 4:
		changeRuntime = falsePtr
	default:
		changeRuntime = nil
	}

	switch cndrRand % 5 {
	case 1, 3:
		changeCNDR = truePtr
	case 2, 4:
		changeCNDR = falsePtr
	default:
		changeCNDR = nil
	}

	return models.ConfigurationChange{
		ID:                    strconv.Itoa(rand.Int()),
		AgentVersion:          changeVersion,
		EnableClusterScanning: changeClusterScanning,
		EnableRuntime:         changeRuntime,
		EnableCNDR:            changeCNDR,
		Status:                models.ChangeStatusPending,
	}
}
