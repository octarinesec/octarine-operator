package common

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/vmware/cbcontainers-operator/api/v1/common_specs"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func MutateEnvVars(container *coreV1.Container, desiredEnvsValues map[string]string, accessTokenSecretName string, eventsGatewaySpec *common_specs.CBContainersEventsGatewaySpec, customEnvsToAdd ...coreV1.EnvVar) {
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

func getDesiredEnvVars(desiredEnvsValues map[string]string, accessTokenSecretName string, eventsGatewaySpec *common_specs.CBContainersEventsGatewaySpec, customEnvsToAdd []coreV1.EnvVar) map[string]coreV1.EnvVar {
	desiredEnvVars := make(map[string]coreV1.EnvVar)
	for desiredEnvVarName, desiredEnvVarValue := range desiredEnvsValues {
		desiredEnvVars[desiredEnvVarName] = coreV1.EnvVar{Name: desiredEnvVarName, Value: desiredEnvVarValue}
	}
	envsToAdd := GetCommonDataPlaneEnvVars(accessTokenSecretName)
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

func getEventsGateWayEnvVars(eventsGatewaySpec *common_specs.CBContainersEventsGatewaySpec) []coreV1.EnvVar {
	return []coreV1.EnvVar{
		{Name: "OCTARINE_MESSAGEPROXY_HOST", Value: eventsGatewaySpec.Host},
		{Name: "OCTARINE_MESSAGEPROXY_PORT", Value: strconv.Itoa(eventsGatewaySpec.Port)},
	}
}

func MutateImage(container *coreV1.Container, desiredImage common_specs.CBContainersImageSpec, desiredVersion string) {
	desiredTag := desiredImage.Tag
	if desiredTag == "" {
		desiredTag = desiredVersion
	}
	desiredFullImage := fmt.Sprintf("%s:%s", desiredImage.Repository, desiredTag)

	container.Image = desiredFullImage
	container.ImagePullPolicy = desiredImage.PullPolicy
}

func MutateContainerHTTPProbes(container *coreV1.Container, desiredProbes common_specs.CBContainersHTTPProbesSpec) {
	mutateContainerCommonProbes(container)

	mutateHTTPProbe(container.ReadinessProbe, desiredProbes.ReadinessPath, desiredProbes)
	mutateHTTPProbe(container.LivenessProbe, desiredProbes.LivenessPath, desiredProbes)
}

func MutateContainerFileProbes(container *coreV1.Container, desiredProbes common_specs.CBContainersFileProbesSpec) {
	mutateContainerCommonProbes(container)

	mutateFileProbe(container.ReadinessProbe, desiredProbes.ReadinessPath, desiredProbes)
	mutateFileProbe(container.LivenessProbe, desiredProbes.LivenessPath, desiredProbes)
}

func mutateContainerCommonProbes(container *coreV1.Container) {
	if container.ReadinessProbe == nil {
		container.ReadinessProbe = &coreV1.Probe{}
	}

	if container.LivenessProbe == nil {
		container.LivenessProbe = &coreV1.Probe{}
	}
}

func mutateHTTPProbe(probe *coreV1.Probe, desiredPath string, desiredProbes common_specs.CBContainersHTTPProbesSpec) {
	if probe.Handler.HTTPGet == nil {
		probe.Handler = coreV1.Handler{
			HTTPGet: &coreV1.HTTPGetAction{},
		}
	}

	probe.HTTPGet.Path = desiredPath
	probe.HTTPGet.Port = intstr.FromInt(desiredProbes.Port)
	probe.HTTPGet.Scheme = desiredProbes.Scheme
	mutateCommonProbe(probe, desiredProbes.CBContainersCommonProbesSpec)
}

func mutateFileProbe(probe *coreV1.Probe, desiredPath string, desiredProbes common_specs.CBContainersFileProbesSpec) {
	if probe.Handler.Exec == nil {
		probe.Handler = coreV1.Handler{
			Exec: &coreV1.ExecAction{},
		}
	}

	probe.Handler.Exec.Command = []string{"cat", desiredPath}
	mutateCommonProbe(probe, desiredProbes.CBContainersCommonProbesSpec)
}

func mutateCommonProbe(probe *coreV1.Probe, desiredProbes common_specs.CBContainersCommonProbesSpec) {
	probe.InitialDelaySeconds = desiredProbes.InitialDelaySeconds
	probe.TimeoutSeconds = desiredProbes.TimeoutSeconds
	probe.PeriodSeconds = desiredProbes.PeriodSeconds
	probe.SuccessThreshold = desiredProbes.SuccessThreshold
	probe.FailureThreshold = desiredProbes.FailureThreshold
}
