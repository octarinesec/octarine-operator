package hardening

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

const (
	EnforcerName = "cbcontainers-hardening-enforcer"

	DesiredContainerPortName  = "https"
	DesiredContainerPortValue = 443
)

type EnforcerK8sObject struct {
	CBContainersHardeningChildK8sObject
}

func NewEnforcerDeploymentK8sObject() *EnforcerK8sObject { return &EnforcerK8sObject{} }

func (obj *EnforcerK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: EnforcerName, Namespace: obj.cbContainersHardening.Namespace}
}

func (obj *EnforcerK8sObject) EmptyK8sObject() client.Object {
	return &appsV1.Deployment{}
}

func (obj *EnforcerK8sObject) MutateK8sObject(k8sObject client.Object) error {
	enforcerSpec := obj.cbContainersHardening.Spec.EnforcerSpec

	deployment, ok := k8sObject.(*appsV1.Deployment)
	if !ok {
		return fmt.Errorf("expected Deployment K8s object")
	}

	if deployment.Spec.Selector == nil {
		deployment.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: enforcerSpec.PodTemplateLabels,
		}
	}

	if deployment.ObjectMeta.Annotations == nil {
		deployment.ObjectMeta.Annotations = make(map[string]string)
	}

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	deployment.ObjectMeta.Labels = enforcerSpec.DeploymentLabels
	deployment.Spec.Selector.MatchLabels = enforcerSpec.PodTemplateLabels
	deployment.Spec.Template.ObjectMeta.Labels = enforcerSpec.PodTemplateLabels
	applyment.EnforceMapContains(deployment.ObjectMeta.Annotations, enforcerSpec.DeploymentAnnotations)
	applyment.EnforceMapContains(deployment.Spec.Template.ObjectMeta.Annotations, enforcerSpec.PodTemplateAnnotations)
	deployment.Spec.Replicas = &enforcerSpec.ReplicasCount
	deployment.Spec.Template.Spec.ImagePullSecrets = []coreV1.LocalObjectReference{{Name: commonState.RegistrySecretName}}
	obj.mutateContainersList(&deployment.Spec.Template.Spec)
	//applyment.MutateString(enforcerSpec.ServiceAccountName, func() *string { return &template.Spec.ServiceAccountName }, func(value string) { template.Spec.ServiceAccountName = value })
	//applyment.MutateString(enforcerSpec.PriorityClassName, func() *string { return &template.Spec.PriorityClassName }, func(value string) { template.Spec.PriorityClassName = value })

	return nil
}

func (obj *EnforcerK8sObject) mutateContainersList(templatePodSpec *coreV1.PodSpec) {
	if len(templatePodSpec.Containers) != 1 {
		container := coreV1.Container{}
		templatePodSpec.Containers = []coreV1.Container{container}
	}

	obj.mutateContainer(&templatePodSpec.Containers[0])
}

func (obj *EnforcerK8sObject) mutateContainer(container *coreV1.Container) {
	enforcerSpec := obj.cbContainersHardening.Spec.EnforcerSpec

	container.Name = EnforcerName
	obj.mutateEnvVars(container, enforcerSpec.Env)
	obj.mutateImage(container, enforcerSpec.Image)
	obj.mutateSecurityContext(container, enforcerSpec.SecurityContext)
	obj.mutateContainerProbes(container, enforcerSpec.Probes)
	obj.mutateContainerPorts(container)
	container.Resources = enforcerSpec.Resources
}

func (obj *EnforcerK8sObject) mutateEnvVars(container *coreV1.Container, desiredEnvsValues map[string]string) {
	desiredEnvVars := obj.getDesiredEnvVars(desiredEnvsValues)

	if !obj.shouldChangeEnvVars(container, desiredEnvVars) {
		return
	}

	container.Env = make([]coreV1.EnvVar, 0, len(desiredEnvVars))
	for _, desiredEnvVar := range desiredEnvVars {
		container.Env = append(container.Env, desiredEnvVar)
	}
}

func (obj *EnforcerK8sObject) shouldChangeEnvVars(container *coreV1.Container, desiredEnvVars map[string]coreV1.EnvVar) bool {
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

func (obj *EnforcerK8sObject) getDesiredEnvVars(desiredEnvsValues map[string]string) map[string]coreV1.EnvVar {
	desiredEnvVars := make(map[string]coreV1.EnvVar)
	for desiredEnvVarName, desiredEnvVarValue := range desiredEnvsValues {
		desiredEnvVars[desiredEnvVarName] = coreV1.EnvVar{Name: desiredEnvVarName, Value: desiredEnvVarValue}
	}
	envsToAdd := commonState.GetCommonDataPlaneEnvVars(obj.cbContainersHardening.Spec.AccessTokenSecretName)
	envsToAdd = append(envsToAdd, obj.getEventsGateWayEnvVars()...)

	for _, dataPlaneEnvVar := range envsToAdd {
		if _, ok := desiredEnvVars[dataPlaneEnvVar.Name]; ok {
			continue
		}
		desiredEnvVars[dataPlaneEnvVar.Name] = dataPlaneEnvVar
	}
	return desiredEnvVars
}

func (obj *EnforcerK8sObject) getEventsGateWayEnvVars() []coreV1.EnvVar {
	eventsGatewaySpec := obj.cbContainersHardening.Spec.EventsGatewaySpec
	return []coreV1.EnvVar{
		{Name: "OCTARINE_MESSAGEPROXY_HOST", Value: eventsGatewaySpec.Host},
		{Name: "OCTARINE_MESSAGEPROXY_PORT", Value: strconv.Itoa(eventsGatewaySpec.Port)},
	}
}

func (obj *EnforcerK8sObject) mutateImage(container *coreV1.Container, desiredImage cbcontainersv1.CBContainersHardeningEnforcerImageSpec) {
	desiredTag := desiredImage.Tag
	if desiredTag == "" {
		desiredTag = obj.cbContainersHardening.Spec.Version
	}
	desiredFullImage := fmt.Sprintf("%s:%s", desiredImage.Repository, desiredTag)

	container.Image = desiredFullImage
	container.ImagePullPolicy = desiredImage.PullPolicy
}

func (obj *EnforcerK8sObject) mutateSecurityContext(container *coreV1.Container, desiredSecurityContext cbcontainersv1.CBContainersHardeningEnforcerSecurityContextSpec) {
	if container.SecurityContext == nil {
		container.SecurityContext = &coreV1.SecurityContext{}
	}
	container.SecurityContext.AllowPrivilegeEscalation = &desiredSecurityContext.AllowPrivilegeEscalation
	container.SecurityContext.ReadOnlyRootFilesystem = &desiredSecurityContext.ReadOnlyRootFilesystem
	container.SecurityContext.RunAsUser = &desiredSecurityContext.RunAsUser
	container.SecurityContext.Capabilities = &coreV1.Capabilities{
		Add:  desiredSecurityContext.CapabilitiesToAdd,
		Drop: desiredSecurityContext.CapabilitiesToDrop,
	}
}

func (obj *EnforcerK8sObject) mutateContainerProbes(container *coreV1.Container, desiredProbes cbcontainersv1.CBContainersHardeningEnforcerProbesSpec) {
	if container.ReadinessProbe == nil {
		container.ReadinessProbe = &coreV1.Probe{}
	}

	if container.LivenessProbe == nil {
		container.LivenessProbe = &coreV1.Probe{}
	}

	obj.mutateProbe(container.ReadinessProbe, desiredProbes.ReadinessPath, desiredProbes)
	obj.mutateProbe(container.LivenessProbe, desiredProbes.LivenessPath, desiredProbes)
}

func (obj *EnforcerK8sObject) mutateProbe(probe *coreV1.Probe, desiredPath string, desiredProbes cbcontainersv1.CBContainersHardeningEnforcerProbesSpec) {
	if probe.Handler.HTTPGet == nil {
		probe.Handler = coreV1.Handler{
			HTTPGet: &coreV1.HTTPGetAction{},
		}
	}

	probe.HTTPGet.Path = desiredPath
	probe.HTTPGet.Port = desiredProbes.Port
	probe.HTTPGet.Scheme = desiredProbes.Scheme
	probe.InitialDelaySeconds = desiredProbes.InitialDelaySeconds
	probe.TimeoutSeconds = desiredProbes.TimeoutSeconds
	probe.PeriodSeconds = desiredProbes.PeriodSeconds
	probe.SuccessThreshold = desiredProbes.SuccessThreshold
	probe.FailureThreshold = desiredProbes.FailureThreshold
}

func (obj *EnforcerK8sObject) mutateContainerPorts(container *coreV1.Container) {
	if container.Ports == nil || len(container.Ports) != 1 {
		container.Ports = []coreV1.ContainerPort{{}}
	}

	container.Ports[0].Name = DesiredContainerPortName
	container.Ports[0].ContainerPort = DesiredContainerPortValue
}
