package config_applier

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"testing"
)

var (
	trueV    = true
	truePtr  = &trueV
	falseV   = false
	falsePtr = &falseV
)

func TestFeatureTogglesAreAppliedCorrectly(t *testing.T) {
	type appliedChangeTest struct {
		name          string
		change        ConfigurationChange
		initialCR     cbcontainersv1.CBContainersAgent
		assertFinalCR func(*testing.T, *cbcontainersv1.CBContainersAgent)
	}

	// generateFeatureToggleTestCases produces a set of tests for a single feature toggle in the requested change
	// The tests validate if each toggle state (true, false, nil) is applied correctly or ignored when it's not needed against the CR's state (true, false, nil)
	generateFeatureToggleTestCases :=
		func(feature string,
			changeFieldSelector func(*ConfigurationChange) **bool,
			crFieldSelector func(agent *cbcontainersv1.CBContainersAgent) **bool) []appliedChangeTest {

			var result []appliedChangeTest

			for _, crState := range []*bool{truePtr, falsePtr, nil} {
				cr := cbcontainersv1.CBContainersAgent{}
				crFieldPtr := crFieldSelector(&cr)
				*crFieldPtr = crState

				// Validate that each toggle state works (or doesn't do anything when it matches)
				for _, changeState := range []*bool{truePtr, falsePtr} {
					change := ConfigurationChange{}
					changeFieldPtr := changeFieldSelector(&change)
					*changeFieldPtr = changeState

					expectedState := changeState // avoid closure issues
					result = append(result, appliedChangeTest{
						name:      fmt.Sprintf("toggle feature (%s) from (%v) to (%v)", feature, prettyPrintBoolPtr(crState), prettyPrintBoolPtr(changeState)),
						change:    change,
						initialCR: cr,
						assertFinalCR: func(t *testing.T, agent *cbcontainersv1.CBContainersAgent) {
							crFieldPostChangePtr := crFieldSelector(agent)
							assert.Equal(t, expectedState, *crFieldPostChangePtr)
						},
					})
				}

				// Validate that a change with the toggle unset does not modify the CR
				result = append(result, appliedChangeTest{
					name:      fmt.Sprintf("missing toggle feature (%s) with CR state (%v)", feature, prettyPrintBoolPtr(crState)),
					change:    ConfigurationChange{},
					initialCR: cr,
					assertFinalCR: func(t *testing.T, agent *cbcontainersv1.CBContainersAgent) {
						crFieldPostChangePtr := crFieldSelector(agent)
						assert.Equal(t, *crFieldPtr, *crFieldPostChangePtr)
					},
				})
			}

			return result
		}

	var testCases []appliedChangeTest

	clusterScannerToggleTestCases := generateFeatureToggleTestCases("cluster scanning",
		func(change *ConfigurationChange) **bool {
			return &change.EnableClusterScanning
		}, func(agent *cbcontainersv1.CBContainersAgent) **bool {
			return &agent.Spec.Components.ClusterScanning.Enabled
		})

	runtimeToggleTestCases := generateFeatureToggleTestCases("runtime protection",
		func(change *ConfigurationChange) **bool {
			return &change.EnableRuntime
		}, func(agent *cbcontainersv1.CBContainersAgent) **bool {
			return &agent.Spec.Components.RuntimeProtection.Enabled
		})

	cndrToggleTestCases := generateFeatureToggleTestCases("CNDR",
		func(change *ConfigurationChange) **bool {
			return &change.EnableCNDR
		}, func(agent *cbcontainersv1.CBContainersAgent) **bool {
			if agent.Spec.Components.Cndr == nil {
				agent.Spec.Components.Cndr = &cbcontainersv1.CBContainersCndrSpec{}
			}
			return &agent.Spec.Components.Cndr.Enabled
		})

	testCases = append(testCases, clusterScannerToggleTestCases...)
	testCases = append(testCases, runtimeToggleTestCases...)
	testCases = append(testCases, cndrToggleTestCases...)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			applyChangesToCR(testCase.change, &testCase.initialCR)
			testCase.assertFinalCR(t, &testCase.initialCR)
		})
	}
}

func TestVersionIsAppliedCorrectly(t *testing.T) {
	originalVersion := "my-version-42"
	newVersion := "new-version"
	cr := cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Version: originalVersion}}
	change := ConfigurationChange{AgentVersion: &newVersion}

	applyChangesToCR(change, &cr)
	assert.Equal(t, newVersion, cr.Spec.Version)
}

func TestMissingVersionDoesNotModifyCR(t *testing.T) {
	originalVersion := "my-version-42"
	cr := cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Version: originalVersion}}
	change := ConfigurationChange{AgentVersion: nil, EnableRuntime: truePtr}

	applyChangesToCR(change, &cr)
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
	change := ConfigurationChange{AgentVersion: &newVersion}

	applyChangesToCR(change, &cr)

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
