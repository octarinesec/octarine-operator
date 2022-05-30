package common

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func MutateContainerHTTPProbes(container *coreV1.Container, desiredProbes cbcontainersv1.CBContainersHTTPProbesSpec) {
	mutateContainerCommonProbes(container)

	mutateHTTPProbe(container.ReadinessProbe, desiredProbes.ReadinessPath, desiredProbes)
	mutateHTTPProbe(container.LivenessProbe, desiredProbes.LivenessPath, desiredProbes)
}

func MutateContainerFileProbes(container *coreV1.Container, desiredProbes cbcontainersv1.CBContainersFileProbesSpec) {
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

func mutateHTTPProbe(probe *coreV1.Probe, desiredPath string, desiredProbes cbcontainersv1.CBContainersHTTPProbesSpec) {
	if probe.HTTPGet == nil {
		probe.HTTPGet = &coreV1.HTTPGetAction{}
	}

	probe.HTTPGet.Path = desiredPath
	probe.HTTPGet.Port = intstr.FromInt(desiredProbes.Port)
	probe.HTTPGet.Scheme = desiredProbes.Scheme
	mutateCommonProbe(probe, desiredProbes.CBContainersCommonProbesSpec)
}

func mutateFileProbe(probe *coreV1.Probe, desiredPath string, desiredProbes cbcontainersv1.CBContainersFileProbesSpec) {
	if probe.Exec == nil {
		probe.Exec = &coreV1.ExecAction{}
	}

	probe.Exec.Command = []string{"cat", desiredPath}
	mutateCommonProbe(probe, desiredProbes.CBContainersCommonProbesSpec)
}

func mutateCommonProbe(probe *coreV1.Probe, desiredProbes cbcontainersv1.CBContainersCommonProbesSpec) {
	probe.InitialDelaySeconds = desiredProbes.InitialDelaySeconds
	probe.TimeoutSeconds = desiredProbes.TimeoutSeconds
	probe.PeriodSeconds = desiredProbes.PeriodSeconds
	probe.SuccessThreshold = desiredProbes.SuccessThreshold
	probe.FailureThreshold = desiredProbes.FailureThreshold
}
