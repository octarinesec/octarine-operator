package objects

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"strconv"
)

func mutateEnvVars(container *coreV1.Container, desiredEnvsValues map[string]string, accessTokenSecretName string, eventsGatewaySpec *cbcontainersv1.CBContainersHardeningEventsGatewaySpec, customEnvsToAdd ...coreV1.EnvVar) {
	desiredEnvVars := getDesiredEnvVars(desiredEnvsValues, accessTokenSecretName, eventsGatewaySpec, customEnvsToAdd)

	if !shouldChangeEnvVars(container, desiredEnvVars) {
		return
	}

	container.Env = make([]coreV1.EnvVar, 0, len(desiredEnvVars))
	for _, desiredEnvVar := range desiredEnvVars {
		container.Env = append(container.Env, desiredEnvVar)
	}
}

func shouldChangeEnvVars(container *coreV1.Container, desiredEnvVars map[string]coreV1.EnvVar) bool {
	if len(container.Env) != len(desiredEnvVars) {
		return true
	}

	for _, actualEnvVar := range container.Env {
		desiredEnvVar, ok := desiredEnvVars[actualEnvVar.Name]
		if !ok || !reflect.DeepEqual(actualEnvVar, desiredEnvVar) {
			return true
		}
	}

	return false
}

func getDesiredEnvVars(desiredEnvsValues map[string]string, accessTokenSecretName string, eventsGatewaySpec *cbcontainersv1.CBContainersHardeningEventsGatewaySpec, customEnvsToAdd []coreV1.EnvVar) map[string]coreV1.EnvVar {
	desiredEnvVars := make(map[string]coreV1.EnvVar)
	for desiredEnvVarName, desiredEnvVarValue := range desiredEnvsValues {
		desiredEnvVars[desiredEnvVarName] = coreV1.EnvVar{Name: desiredEnvVarName, Value: desiredEnvVarValue}
	}
	envsToAdd := commonState.GetCommonDataPlaneEnvVars(accessTokenSecretName)
	envsToAdd = append(envsToAdd, getEventsGateWayEnvVars(eventsGatewaySpec)...)
	envsToAdd = append(envsToAdd, customEnvsToAdd...)

	for _, dataPlaneEnvVar := range envsToAdd {
		if _, ok := desiredEnvVars[dataPlaneEnvVar.Name]; ok {
			continue
		}
		desiredEnvVars[dataPlaneEnvVar.Name] = dataPlaneEnvVar
	}
	return desiredEnvVars
}

func getEventsGateWayEnvVars(eventsGatewaySpec *cbcontainersv1.CBContainersHardeningEventsGatewaySpec) []coreV1.EnvVar {
	return []coreV1.EnvVar{
		{Name: "OCTARINE_MESSAGEPROXY_HOST", Value: eventsGatewaySpec.Host},
		{Name: "OCTARINE_MESSAGEPROXY_PORT", Value: strconv.Itoa(eventsGatewaySpec.Port)},
	}
}

func mutateImage(container *coreV1.Container, desiredImage cbcontainersv1.CBContainersHardeningImageSpec, desiredVersion string) {
	desiredTag := desiredImage.Tag
	if desiredTag == "" {
		desiredTag = desiredVersion
	}
	desiredFullImage := fmt.Sprintf("%s:%s", desiredImage.Repository, desiredTag)

	container.Image = desiredFullImage
	container.ImagePullPolicy = desiredImage.PullPolicy
}

func mutateContainerProbes(container *coreV1.Container, desiredProbes cbcontainersv1.CBContainersHardeningProbesSpec) {
	if container.ReadinessProbe == nil {
		container.ReadinessProbe = &coreV1.Probe{}
	}

	if container.LivenessProbe == nil {
		container.LivenessProbe = &coreV1.Probe{}
	}

	mutateProbe(container.ReadinessProbe, desiredProbes.ReadinessPath, desiredProbes)
	mutateProbe(container.LivenessProbe, desiredProbes.LivenessPath, desiredProbes)
}

func mutateProbe(probe *coreV1.Probe, desiredPath string, desiredProbes cbcontainersv1.CBContainersHardeningProbesSpec) {
	if probe.Handler.HTTPGet == nil {
		probe.Handler = coreV1.Handler{
			HTTPGet: &coreV1.HTTPGetAction{},
		}
	}

	probe.HTTPGet.Path = desiredPath
	probe.HTTPGet.Port = intstr.FromInt(desiredProbes.Port)
	probe.HTTPGet.Scheme = desiredProbes.Scheme
	probe.InitialDelaySeconds = desiredProbes.InitialDelaySeconds
	probe.TimeoutSeconds = desiredProbes.TimeoutSeconds
	probe.PeriodSeconds = desiredProbes.PeriodSeconds
	probe.SuccessThreshold = desiredProbes.SuccessThreshold
	probe.FailureThreshold = desiredProbes.FailureThreshold
}
