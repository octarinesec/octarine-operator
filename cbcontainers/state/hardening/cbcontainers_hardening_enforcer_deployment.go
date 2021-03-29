package hardening

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	EnforcerName = "Enforcer"
)

var (
	DesiredContainerPorts = []coreV1.ContainerPort{
		{Name: "https", ContainerPort: 443},
	}
)

type EnforcerK8sObject struct {
	CBContainersHardeningChildK8sObject
}

func NewEnforcerDeploymentK8sObject() *EnforcerK8sObject { return &EnforcerK8sObject{} }

func (obj *EnforcerK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: EnforcerName, Namespace: obj.cbContainersHardening.Namespace}
}

func (obj *EnforcerK8sObject) EmptyK8sObject() client.Object { return &appsV1.Deployment{} }

func (obj *EnforcerK8sObject) MutateK8sObject(k8sObject client.Object) (bool, error) {
	mutated := false
	enforcerSpec := obj.cbContainersHardening.Spec.EnforcerSpec

	deployment, ok := k8sObject.(*appsV1.Deployment)
	if !ok {
		return false, fmt.Errorf("expected Deployment K8s object")
	}
	template := deployment.Spec.Template

	mutated = obj.mutateStringsMap(deployment.Labels, enforcerSpec.DeploymentLabels) || mutated
	mutated = obj.mutateStringsMap(template.Labels, enforcerSpec.PodTemplateLabels) || mutated
	mutated = obj.mutateStringsMap(deployment.Annotations, enforcerSpec.DeploymentAnnotations) || mutated
	mutated = obj.mutateStringsMap(template.Annotations, enforcerSpec.PodTemplateAnnotations) || mutated
	mutated = applyment.MutateInt32(enforcerSpec.ReplicasCount, func() *int32 { return deployment.Spec.Replicas }, func(value int32) { deployment.Spec.Replicas = &value })
	//mutated = applyment.MutateString(enforcerSpec.ServiceAccountName, func() *string { return &template.Spec.ServiceAccountName }, func(value string) { template.Spec.ServiceAccountName = value })
	//mutated = applyment.MutateString(enforcerSpec.PriorityClassName, func() *string { return &template.Spec.PriorityClassName }, func(value string) { template.Spec.PriorityClassName = value })
	mutated = obj.mutateContainersList(&template.Spec)

	return mutated, nil
}

func (obj *EnforcerK8sObject) mutateContainersList(templatePodSpec *coreV1.PodSpec) bool {
	containersLengthChanged := false
	if len(templatePodSpec.Containers) != 1 {
		container := coreV1.Container{}
		obj.mutateContainer(&container)
		templatePodSpec.Containers = []coreV1.Container{container}
		containersLengthChanged = true
	}

	return obj.mutateContainer(&templatePodSpec.Containers[0]) || containersLengthChanged
}

func (obj *EnforcerK8sObject) mutateContainer(container *coreV1.Container) bool {
	mutated := false
	enforcerSpec := obj.cbContainersHardening.Spec.EnforcerSpec

	mutated = applyment.MutateString(EnforcerName, func() *string { return &container.Name }, func(value string) { container.Name = value })
	mutated = obj.mutateEnvVars(container, enforcerSpec.Env) || mutated
	mutated = obj.mutateImage(container, enforcerSpec.Image) || mutated
	mutated = obj.mutateSecurityContext(container, enforcerSpec.SecurityContext) || mutated
	mutated = obj.mutateContainerProbes(container, enforcerSpec.Probes) || mutated

	if reflect.DeepEqual(container.Ports, DesiredContainerPorts) {
		container.Ports = DesiredContainerPorts
		mutated = true
	}

	if reflect.DeepEqual(container.Resources, enforcerSpec.Resources) {
		container.Resources = enforcerSpec.Resources
		mutated = true
	}

	return mutated
}

func (obj *EnforcerK8sObject) mutateStringsMap(actualLabels map[string]string, desiredLabels map[string]string) bool {
	mutated := false
	for key, desiredValue := range desiredLabels {
		actualValue, ok := actualLabels[key]
		if !ok || actualValue != desiredValue {
			actualLabels[key] = desiredValue
			mutated = true
		}
	}

	return mutated
}

func (obj *EnforcerK8sObject) mutateEnvVars(container *coreV1.Container, desiredEnvsValues map[string]string) bool {
	mutated := false

	actualEnvs := make(map[string]*coreV1.EnvVar)
	for idx, envVar := range container.Env {
		actualEnvs[envVar.Name] = &(container.Env[idx])
	}

	for name, desiredEnvValue := range desiredEnvsValues {
		actualEnv, ok := actualEnvs[name]
		if ok && actualEnv.Value == desiredEnvValue {
			continue
		}

		mutated = true
		if !ok {
			container.Env = append(container.Env, coreV1.EnvVar{Name: name, Value: desiredEnvValue})
			continue
		}
		actualEnv.Value = desiredEnvValue
		actualEnv.ValueFrom = nil
	}

	return mutated
}

func (obj *EnforcerK8sObject) mutateImage(container *coreV1.Container, desiredImage cbcontainersv1.CBContainersHardeningEnforcerImageSpec) bool {
	mutated := false

	actualPullPolicy := string(container.ImagePullPolicy)
	desiredTag := desiredImage.Tag
	if desiredTag == "" {
		desiredTag = obj.cbContainersHardening.Spec.Version
	}
	desiredFullImage := fmt.Sprintf("%s:%s", desiredImage.Repository, desiredTag)

	mutated = applyment.MutateString(desiredFullImage, func() *string { return &container.Image }, func(value string) { container.Image = value }) || mutated
	mutated = applyment.MutateString(string(desiredImage.PullPolicy), func() *string { return &actualPullPolicy }, func(value string) { container.ImagePullPolicy = coreV1.PullPolicy(value) }) || mutated

	return mutated
}

func (obj *EnforcerK8sObject) mutateSecurityContext(container *coreV1.Container, desiredSecurityContext cbcontainersv1.CBContainersHardeningEnforcerSecurityContextSpec) bool {
	mutated := false

	if container.SecurityContext == nil {
		container.SecurityContext = &coreV1.SecurityContext{}
		mutated = true
	}

	mutated = applyment.MutateBool(desiredSecurityContext.AllowPrivilegeEscalation, func() *bool { return container.SecurityContext.AllowPrivilegeEscalation }, func(value bool) { container.SecurityContext.AllowPrivilegeEscalation = &value }) || mutated
	mutated = applyment.MutateBool(desiredSecurityContext.ReadOnlyRootFilesystem, func() *bool { return container.SecurityContext.ReadOnlyRootFilesystem }, func(value bool) { container.SecurityContext.ReadOnlyRootFilesystem = &value }) || mutated
	mutated = applyment.MutateInt64(desiredSecurityContext.RunAsUser, func() *int64 { return container.SecurityContext.RunAsUser }, func(value int64) { container.SecurityContext.RunAsUser = &value }) || mutated

	if reflect.DeepEqual(container.SecurityContext.Capabilities.Add, desiredSecurityContext.CapabilitiesToAdd) {
		container.SecurityContext.Capabilities.Add = desiredSecurityContext.CapabilitiesToAdd
		mutated = true
	}

	if reflect.DeepEqual(container.SecurityContext.Capabilities.Drop, desiredSecurityContext.CapabilitiesToDrop) {
		container.SecurityContext.Capabilities.Add = desiredSecurityContext.CapabilitiesToAdd
		mutated = true
	}

	return mutated
}

func (obj *EnforcerK8sObject) mutateContainerProbes(container *coreV1.Container, desiredProbes cbcontainersv1.CBContainersHardeningEnforcerProbesSpec) bool {
	mutated := false

	if container.ReadinessProbe == nil {
		container.ReadinessProbe = &coreV1.Probe{}
		mutated = true
	}

	if container.LivenessProbe == nil {
		container.LivenessProbe = &coreV1.Probe{}
		mutated = true
	}

	mutated = obj.mutateProbe(container.ReadinessProbe, desiredProbes.ReadinessPath, desiredProbes) || mutated
	mutated = obj.mutateProbe(container.LivenessProbe, desiredProbes.LivenessPath, desiredProbes) || mutated

	return mutated
}

func (obj *EnforcerK8sObject) mutateProbe(probe *coreV1.Probe, desiredPath string, desiredProbes cbcontainersv1.CBContainersHardeningEnforcerProbesSpec) bool {
	mutated := false

	if probe.Handler.HTTPGet == nil {
		probe.Handler = coreV1.Handler{
			HTTPGet: &coreV1.HTTPGetAction{},
		}
		mutated = true
	}
	actualHttpScheme := string(probe.HTTPGet.Scheme)

	if probe.HTTPGet.Port != desiredProbes.Port {
		probe.HTTPGet.Port = desiredProbes.Port
		mutated = true
	}

	mutated = applyment.MutateString(desiredPath, func() *string { return &probe.HTTPGet.Path }, func(value string) { probe.HTTPGet.Path = value }) || mutated
	mutated = applyment.MutateString(string(desiredProbes.Scheme), func() *string { return &actualHttpScheme }, func(value string) { probe.HTTPGet.Scheme = coreV1.URIScheme(value) }) || mutated
	mutated = applyment.MutateInt32(desiredProbes.InitialDelaySeconds, func() *int32 { return &probe.InitialDelaySeconds }, func(value int32) { probe.InitialDelaySeconds = value }) || mutated
	mutated = applyment.MutateInt32(desiredProbes.TimeoutSeconds, func() *int32 { return &probe.TimeoutSeconds }, func(value int32) { probe.TimeoutSeconds = value }) || mutated
	mutated = applyment.MutateInt32(desiredProbes.PeriodSeconds, func() *int32 { return &probe.PeriodSeconds }, func(value int32) { probe.PeriodSeconds = value }) || mutated
	mutated = applyment.MutateInt32(desiredProbes.SuccessThreshold, func() *int32 { return &probe.SuccessThreshold }, func(value int32) { probe.SuccessThreshold = value }) || mutated
	mutated = applyment.MutateInt32(desiredProbes.FailureThreshold, func() *int32 { return &probe.FailureThreshold }, func(value int32) { probe.FailureThreshold = value }) || mutated

	return mutated
}
