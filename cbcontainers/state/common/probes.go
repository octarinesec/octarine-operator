package common

import (
	"github.com/vmware/cbcontainers-operator/api/v1/common_specs"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

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
