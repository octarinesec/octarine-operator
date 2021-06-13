package common

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	accessTokenSecretName = "access-token-secret"

	testName1  = "test_1"
	testName2  = "test_2"
	testName3  = "test_2"
	testValue1 = "value_1"
	testValue2 = "value_2"
	testValue3 = "value_3"

	eventsGatewayHost = "test.com"
	eventsGatewayPort = 443
)

func compareEnvVars(t *testing.T, expected map[string]coreV1.EnvVar, actual []coreV1.EnvVar) {
	for _, envVar := range actual {
		expectedEnvVar, ok := expected[envVar.Name]
		require.True(t, ok)
		require.True(t, reflect.DeepEqual(expectedEnvVar, envVar))
	}
}

func TestWithDataPlaneCommonConfig(t *testing.T) {
	expected := map[string]coreV1.EnvVar{
		accessTokenVarName: {
			Name: accessTokenVarName,
			ValueFrom: &coreV1.EnvVarSource{
				SecretKeyRef: &coreV1.SecretKeySelector{
					LocalObjectReference: coreV1.LocalObjectReference{Name: accessTokenSecretName},
					Key:                  AccessTokenSecretKeyName,
				},
			},
		},
		apiSchemeVarName: {
			Name: apiSchemeVarName,
			ValueFrom: &coreV1.EnvVarSource{
				ConfigMapKeyRef: &coreV1.ConfigMapKeySelector{
					LocalObjectReference: coreV1.LocalObjectReference{Name: DataPlaneConfigmapName},
					Key:                  DataPlaneConfigmapApiSchemeKey,
				},
			},
		},
		apiHostVarName: {
			Name: apiHostVarName,
			ValueFrom: &coreV1.EnvVarSource{
				ConfigMapKeyRef: &coreV1.ConfigMapKeySelector{
					LocalObjectReference: coreV1.LocalObjectReference{Name: DataPlaneConfigmapName},
					Key:                  DataPlaneConfigmapApiHostKey,
				},
			},
		},
		apiPortVarName: {
			Name: apiPortVarName,
			ValueFrom: &coreV1.EnvVarSource{
				ConfigMapKeyRef: &coreV1.ConfigMapKeySelector{
					LocalObjectReference: coreV1.LocalObjectReference{Name: DataPlaneConfigmapName},
					Key:                  DataPlaneConfigmapApiPortKey,
				},
			},
		},
		accountVarName: {
			Name: accountVarName,
			ValueFrom: &coreV1.EnvVarSource{
				ConfigMapKeyRef: &coreV1.ConfigMapKeySelector{
					LocalObjectReference: coreV1.LocalObjectReference{Name: DataPlaneConfigmapName},
					Key:                  DataPlaneConfigmapAccountKey,
				},
			},
		},
		clusterVarName: {
			Name: clusterVarName,
			ValueFrom: &coreV1.EnvVarSource{
				ConfigMapKeyRef: &coreV1.ConfigMapKeySelector{
					LocalObjectReference: coreV1.LocalObjectReference{Name: DataPlaneConfigmapName},
					Key:                  DataPlaneConfigmapClusterKey,
				},
			},
		},
		apiAdapterVarName: {
			Name: apiAdapterVarName,
			ValueFrom: &coreV1.EnvVarSource{
				ConfigMapKeyRef: &coreV1.ConfigMapKeySelector{
					LocalObjectReference: coreV1.LocalObjectReference{Name: DataPlaneConfigmapName},
					Key:                  DataPlaneConfigmapApiAdapterKey,
				},
			},
		},
	}
	actual := NewEnvVarBuilder().
		WithCommonDataPlane(accessTokenSecretName).
		Build()

	compareEnvVars(t, expected, actual)
}

func TestWithCustomConfig(t *testing.T) {
	expected := map[string]coreV1.EnvVar{
		testName1: {
			Name:  testName1,
			Value: testValue1,
		},
		testName2: {
			Name:  testName2,
			Value: testValue2,
		},
	}

	envSlice := make([]coreV1.EnvVar, 0, len(expected))
	for _, envVar := range expected {
		envSlice = append(envSlice, envVar)
	}

	actual := NewEnvVarBuilder().
		WithCustom(envSlice...).
		Build()

	compareEnvVars(t, expected, actual)
}

func TestWithEventsGateway(t *testing.T) {
	eventsGatewaySpec := &cbcontainersv1.CBContainersEventsGatewaySpec{
		Host: eventsGatewayHost,
		Port: eventsGatewayPort,
	}
	expected := map[string]coreV1.EnvVar{
		eventGatewayHostVarName: {
			Name:  eventGatewayHostVarName,
			Value: eventsGatewaySpec.Host,
		},
		eventGatewayPortVarName: {
			Name:  eventGatewayPortVarName,
			Value: strconv.Itoa(eventsGatewaySpec.Port),
		},
	}

	actual := NewEnvVarBuilder().
		WithEventsGateway(eventsGatewaySpec).
		Build()

	compareEnvVars(t, expected, actual)
}

func TestWithSpecNoOverlap(t *testing.T) {
	envSpec := map[string]string{
		testName1: testValue1,
		testName2: testValue2,
	}
	eventsGatewaySpec := &cbcontainersv1.CBContainersEventsGatewaySpec{
		Host: eventsGatewayHost,
		Port: eventsGatewayPort,
	}
	expected := map[string]coreV1.EnvVar{
		eventGatewayHostVarName: {
			Name:  eventGatewayHostVarName,
			Value: eventsGatewaySpec.Host,
		},
		eventGatewayPortVarName: {
			Name:  eventGatewayPortVarName,
			Value: strconv.Itoa(eventsGatewaySpec.Port),
		},
		testName1: {
			Name:  testName1,
			Value: testValue1,
		},
		testName2: {
			Name:  testName2,
			Value: testValue2,
		},
	}

	actual := NewEnvVarBuilder().
		WithEventsGateway(eventsGatewaySpec).
		WithSpec(envSpec).
		Build()

	compareEnvVars(t, expected, actual)
}

func TestWithSpecOverlapping(t *testing.T) {
	eventsGatewaySpec := &cbcontainersv1.CBContainersEventsGatewaySpec{
		Host: eventsGatewayHost,
		Port: eventsGatewayPort,
	}
	envSpec := map[string]string{
		eventGatewayPortVarName: strconv.Itoa(eventsGatewaySpec.Port + 1),
		testName2:               testValue2,
	}
	expected := map[string]coreV1.EnvVar{
		eventGatewayHostVarName: {
			Name:  eventGatewayHostVarName,
			Value: eventsGatewaySpec.Host,
		},
		eventGatewayPortVarName: {
			Name:  eventGatewayPortVarName,
			Value: strconv.Itoa(eventsGatewaySpec.Port + 1),
		},
		testName2: {
			Name:  testName2,
			Value: testValue2,
		},
	}

	actual := NewEnvVarBuilder().
		WithEventsGateway(eventsGatewaySpec).
		WithSpec(envSpec).
		Build()

	compareEnvVars(t, expected, actual)
}

func TestMutationFalse(t *testing.T) {
	expected := map[string]coreV1.EnvVar{
		testName1: {
			Name:  testName1,
			Value: testValue1,
		},
		testName2: {
			Name:  testName2,
			Value: testValue2,
		},
	}
	expectedSlice := make([]coreV1.EnvVar, 0, len(expected))
	for _, envVar := range expected {
		expectedSlice = append(expectedSlice, envVar)
	}
	container := &coreV1.Container{
		Env: expectedSlice,
	}
	builder := NewEnvVarBuilder().
		WithCustom(expectedSlice...)

	MutateEnvVars(container, builder)
	compareEnvVars(t, expected, container.Env)
}

func TestMutationTrueSameSize(t *testing.T) {
	expected := map[string]coreV1.EnvVar{
		testName1: {
			Name:  testName1,
			Value: testValue1,
		},
		testName2: {
			Name:  testName2,
			Value: testValue2,
		},
	}
	expectedSlice := make([]coreV1.EnvVar, 0, len(expected))
	for _, envVar := range expected {
		expectedSlice = append(expectedSlice, envVar)
	}
	container := &coreV1.Container{
		Env: []coreV1.EnvVar{
			{
				Name:  testName1,
				Value: testValue1,
			},
			{
				Name:  testName3,
				Value: testValue3,
			},
		},
	}
	builder := NewEnvVarBuilder().
		WithCustom(expectedSlice...)

	MutateEnvVars(container, builder)
	compareEnvVars(t, expected, container.Env)
}

func TestMutationTrueDifferentSize(t *testing.T) {
	expected := map[string]coreV1.EnvVar{
		testName1: {
			Name:  testName1,
			Value: testValue1,
		},
		testName2: {
			Name:  testName2,
			Value: testValue2,
		},
	}
	expectedSlice := make([]coreV1.EnvVar, 0, len(expected))
	for _, envVar := range expected {
		expectedSlice = append(expectedSlice, envVar)
	}
	container := &coreV1.Container{
		Env: []coreV1.EnvVar{
			{
				Name:  testName3,
				Value: testValue3,
			},
		},
	}
	builder := NewEnvVarBuilder().
		WithCustom(expectedSlice...)

	MutateEnvVars(container, builder)
	compareEnvVars(t, expected, container.Env)
}

func TestMutateImageWithTag(t *testing.T) {
	expectedImage := "cbartifactory/test:1.0.0"
	expectedPullPolicy := coreV1.PullPolicy("IfNotPresent")
	imageSpec := cbcontainersv1.CBContainersImageSpec{
		Repository: "cbartifactory/test",
		Tag:        "1.0.0",
		PullPolicy: expectedPullPolicy,
	}
	container := &coreV1.Container{}
	MutateImage(container, imageSpec, "3.0.0")
	require.Equal(t, expectedImage, container.Image)
	require.Equal(t, expectedPullPolicy, container.ImagePullPolicy)
}

func TestMutateImageWithoutTag(t *testing.T) {
	expectedImage := "cbartifactory/test:3.0.0"
	expectedPullPolicy := coreV1.PullPolicy("IfNotPresent")
	imageSpec := cbcontainersv1.CBContainersImageSpec{
		Repository: "cbartifactory/test",
		Tag:        "",
		PullPolicy: expectedPullPolicy,
	}
	container := &coreV1.Container{}
	MutateImage(container, imageSpec, "3.0.0")
	require.Equal(t, expectedImage, container.Image)
	require.Equal(t, expectedPullPolicy, container.ImagePullPolicy)
}

const (
	expectedInitialDelay     = 1
	expectedTimeout          = 2
	expectedPeriod           = 3
	expectedSuccessThreshold = 4
	expectedFailureThreshold = 5
	expectedPort             = 8181
	expectedReadinessPath    = "/ready"
	expectedLivenessPath     = "/alive"
)

func TestMutateContainerHTTPProbes(t *testing.T) {
	httpProbesSpec := cbcontainersv1.CBContainersHTTPProbesSpec{
		CBContainersCommonProbesSpec: cbcontainersv1.CBContainersCommonProbesSpec{
			InitialDelaySeconds: expectedInitialDelay,
			TimeoutSeconds:      expectedTimeout,
			PeriodSeconds:       expectedPeriod,
			SuccessThreshold:    expectedSuccessThreshold,
			FailureThreshold:    expectedFailureThreshold,
		},
		ReadinessPath: expectedReadinessPath,
		LivenessPath:  expectedLivenessPath,
		Port:          expectedPort,
		Scheme:        coreV1.URISchemeHTTP,
	}

	expectedReadinessProbe := &coreV1.Probe{
		Handler: coreV1.Handler{
			HTTPGet: &coreV1.HTTPGetAction{
				Path:   expectedReadinessPath,
				Port:   intstr.FromInt(expectedPort),
				Scheme: coreV1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: expectedInitialDelay,
		TimeoutSeconds:      expectedTimeout,
		PeriodSeconds:       expectedPeriod,
		SuccessThreshold:    expectedSuccessThreshold,
		FailureThreshold:    expectedFailureThreshold,
	}
	expectedLivenessProbe := &coreV1.Probe{
		Handler: coreV1.Handler{
			HTTPGet: &coreV1.HTTPGetAction{
				Path:   expectedLivenessPath,
				Port:   intstr.FromInt(expectedPort),
				Scheme: coreV1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: expectedInitialDelay,
		TimeoutSeconds:      expectedTimeout,
		PeriodSeconds:       expectedPeriod,
		SuccessThreshold:    expectedSuccessThreshold,
		FailureThreshold:    expectedFailureThreshold,
	}

	container := &coreV1.Container{}
	MutateContainerHTTPProbes(container, httpProbesSpec)
	require.True(t, reflect.DeepEqual(expectedLivenessProbe, container.LivenessProbe))
	require.True(t, reflect.DeepEqual(expectedReadinessProbe, container.ReadinessProbe))
}

func TestMutateContainerFileProbes(t *testing.T) {
	fileProbesSpec := cbcontainersv1.CBContainersFileProbesSpec{
		CBContainersCommonProbesSpec: cbcontainersv1.CBContainersCommonProbesSpec{
			InitialDelaySeconds: expectedInitialDelay,
			TimeoutSeconds:      expectedTimeout,
			PeriodSeconds:       expectedPeriod,
			SuccessThreshold:    expectedSuccessThreshold,
			FailureThreshold:    expectedFailureThreshold,
		},
		ReadinessPath: expectedReadinessPath,
		LivenessPath:  expectedLivenessPath,
	}

	expectedReadinessProbe := &coreV1.Probe{
		Handler: coreV1.Handler{
			Exec: &coreV1.ExecAction{
				Command: []string{"cat", expectedReadinessPath},
			},
		},
		InitialDelaySeconds: expectedInitialDelay,
		TimeoutSeconds:      expectedTimeout,
		PeriodSeconds:       expectedPeriod,
		SuccessThreshold:    expectedSuccessThreshold,
		FailureThreshold:    expectedFailureThreshold,
	}
	expectedLivenessProbe := &coreV1.Probe{
		Handler: coreV1.Handler{
			Exec: &coreV1.ExecAction{
				Command: []string{"cat", expectedLivenessPath},
			},
		},
		InitialDelaySeconds: expectedInitialDelay,
		TimeoutSeconds:      expectedTimeout,
		PeriodSeconds:       expectedPeriod,
		SuccessThreshold:    expectedSuccessThreshold,
		FailureThreshold:    expectedFailureThreshold,
	}

	container := &coreV1.Container{}
	MutateContainerFileProbes(container, fileProbesSpec)
	require.True(t, reflect.DeepEqual(expectedLivenessProbe, container.LivenessProbe))
	require.True(t, reflect.DeepEqual(expectedReadinessProbe, container.ReadinessProbe))
}