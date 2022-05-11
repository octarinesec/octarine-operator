package components

import (
	"fmt"

	cbContainersV1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ResolverName     = "cbcontainers-runtime-resolver"
	resolverLabelKey = "app.kubernetes.io/name"

	desiredDeploymentGRPCPortName       = "grpc"
	desiredInitializationTimeoutMinutes = 3
)

var (
	resolverAllowPrivilegeEscalation       = false
	resolverReadOnlyRootFilesystem         = true
	resolverRunAsUser                int64 = 1500
	resolverRunAsNonRoot                   = true
	resolverCapabilitiesToDrop             = []coreV1.Capability{"ALL"}
)

type ResolverDeploymentK8sObject struct{}

func NewResolverDeploymentK8sObject() *ResolverDeploymentK8sObject {
	return &ResolverDeploymentK8sObject{}
}

func (obj *ResolverDeploymentK8sObject) EmptyK8sObject() client.Object {
	return &appsV1.Deployment{}
}

func (obj *ResolverDeploymentK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: ResolverName, Namespace: commonState.DataPlaneNamespaceName}
}

func (obj *ResolverDeploymentK8sObject) MutateK8sObject(k8sObject client.Object, agentSpec *cbContainersV1.CBContainersAgentSpec) error {
	deployment, ok := k8sObject.(*appsV1.Deployment)
	if !ok {
		return fmt.Errorf("expected Deployment K8s object")
	}

	runtimeProtection := &agentSpec.Components.RuntimeProtection
	resolver := &runtimeProtection.Resolver

	desiredLabels := resolver.Labels
	if desiredLabels == nil {
		desiredLabels = make(map[string]string)
	}
	desiredLabels[resolverLabelKey] = ResolverName

	if deployment.Spec.Selector == nil {
		deployment.Spec.Selector = &metav1.LabelSelector{}
	}

	if deployment.ObjectMeta.Annotations == nil {
		deployment.ObjectMeta.Annotations = make(map[string]string)
	}

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	deployment.Spec.Replicas = resolver.ReplicasCount
	deployment.ObjectMeta.Labels = desiredLabels
	deployment.Spec.Selector.MatchLabels = desiredLabels
	deployment.Spec.Template.ObjectMeta.Labels = desiredLabels
	deployment.Spec.Template.Spec.ServiceAccountName = commonState.RuntimeResolverServiceAccountName
	deployment.Spec.Template.Spec.PriorityClassName = commonState.DataPlanePriorityClassName
	deployment.Spec.Template.Spec.ImagePullSecrets = []coreV1.LocalObjectReference{{Name: commonState.RegistrySecretName}}
	obj.mutateAnnotations(deployment, resolver)
	obj.mutateVolumes(&deployment.Spec.Template.Spec)
	obj.mutateAffinityAndNodeSelector(&deployment.Spec.Template.Spec, resolver)
	obj.mutateContainersList(&deployment.Spec.Template.Spec,
		resolver,
		&agentSpec.Gateways.RuntimeEventsGateway,
		agentSpec.Version,
		agentSpec.AccessTokenSecretName,
		runtimeProtection.InternalGrpcPort,
	)

	return nil
}

func (obj *ResolverDeploymentK8sObject) mutateVolumes(templatePodSpec *coreV1.PodSpec) {
	if templatePodSpec.Volumes == nil || len(templatePodSpec.Volumes) != 1 {
		templatePodSpec.Volumes = make([]coreV1.Volume, 0)
	}

	commonState.MutateVolumesToIncludeRootCAsVolume(templatePodSpec)
}

func (obj *ResolverDeploymentK8sObject) mutateAffinityAndNodeSelector(templatePodSpec *coreV1.PodSpec, resolverSpec *cbContainersV1.CBContainersRuntimeResolverSpec) {
	templatePodSpec.Affinity = resolverSpec.Affinity
	templatePodSpec.NodeSelector = resolverSpec.NodeSelector
}

func (obj *ResolverDeploymentK8sObject) mutateAnnotations(deployment *appsV1.Deployment, resolverSpec *cbContainersV1.CBContainersRuntimeResolverSpec) {
	if deployment.ObjectMeta.Annotations == nil {
		deployment.ObjectMeta.Annotations = make(map[string]string)
	}

	applyment.EnforceMapContains(deployment.ObjectMeta.Annotations, resolverSpec.DeploymentAnnotations)

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	applyment.EnforceMapContains(deployment.Spec.Template.ObjectMeta.Annotations, map[string]string{
		"prometheus.io/scrape": fmt.Sprint(*resolverSpec.Prometheus.Enabled),
		"prometheus.io/port":   fmt.Sprint(resolverSpec.Prometheus.Port),
	})
	applyment.EnforceMapContains(deployment.Spec.Template.ObjectMeta.Annotations, resolverSpec.PodTemplateAnnotations)
}

func (obj *ResolverDeploymentK8sObject) mutateContainersList(
	templatePodSpec *coreV1.PodSpec,
	resolverSpec *cbContainersV1.CBContainersRuntimeResolverSpec,
	eventsGatewaySpec *cbContainersV1.CBContainersEventsGatewaySpec,
	version,
	accessTokenSecretName string,
	desiredGRPCPortValue int32) {

	if len(templatePodSpec.Containers) != 1 {
		container := coreV1.Container{}
		templatePodSpec.Containers = []coreV1.Container{container}
	}

	obj.mutateContainer(&templatePodSpec.Containers[0], resolverSpec, eventsGatewaySpec,
		version, accessTokenSecretName, desiredGRPCPortValue)
}

func (obj *ResolverDeploymentK8sObject) mutateContainer(
	container *coreV1.Container,
	resolverSpec *cbContainersV1.CBContainersRuntimeResolverSpec,
	eventsGatewaySpec *cbContainersV1.CBContainersEventsGatewaySpec,
	version,
	accessTokenSecretName string,
	desiredGRPCPortValue int32) {

	container.Name = ResolverName
	container.Resources = resolverSpec.Resources
	commonState.MutateImage(container, resolverSpec.Image, version)
	commonState.MutateContainerHTTPProbes(container, resolverSpec.Probes)
	obj.mutateEnvVars(container, resolverSpec, eventsGatewaySpec, accessTokenSecretName, desiredGRPCPortValue)
	obj.mutateContainerPorts(container, desiredGRPCPortValue)
	obj.mutateSecurityContext(container)
	obj.mutateVolumesMounts(container)
}

func (obj *ResolverDeploymentK8sObject) mutateContainerPorts(container *coreV1.Container, desiredGRPCPortValue int32) {
	if container.Ports == nil || len(container.Ports) != 1 {
		container.Ports = []coreV1.ContainerPort{{}}
	}

	container.Ports[0].Name = desiredDeploymentGRPCPortName
	container.Ports[0].ContainerPort = desiredGRPCPortValue
}

func (obj *ResolverDeploymentK8sObject) mutateEnvVars(
	container *coreV1.Container,
	resolverSpec *cbContainersV1.CBContainersRuntimeResolverSpec,
	eventsGatewaySpec *cbContainersV1.CBContainersEventsGatewaySpec,
	accessTokenSecretName string,
	desiredGRPCPortValue int32) {

	customEnvs := []coreV1.EnvVar{
		{Name: "RUNTIME_KUBERNETES_RESOLVER_GRPC_PORT", Value: fmt.Sprintf("%d", desiredGRPCPortValue)},
		{Name: "RUNTIME_KUBERNETES_RESOLVER_PROMETHEUS_PORT", Value: fmt.Sprintf("%d", resolverSpec.Prometheus.Port)},
		{Name: "RUNTIME_KUBERNETES_RESOLVER_PROBES_PORT", Value: fmt.Sprintf("%d", resolverSpec.Probes.Port)},
		{Name: "RUNTIME_KUBERNETES_RESOLVER_INITIALIZATION_TIMEOUT_MINUTES", Value: fmt.Sprintf("%d", desiredInitializationTimeoutMinutes)},
		{Name: "GIN_MODE", Value: "release"},
	}

	envVarBuilder := commonState.NewEnvVarBuilder().
		WithCommonDataPlane(accessTokenSecretName).
		WithEventsGateway(eventsGatewaySpec).
		WithCustom(customEnvs...).
		WithSpec(resolverSpec.Env)
	commonState.MutateEnvVars(container, envVarBuilder)
}

func (obj *ResolverDeploymentK8sObject) mutateSecurityContext(container *coreV1.Container) {
	if container.SecurityContext == nil {
		container.SecurityContext = &coreV1.SecurityContext{}
	}

	container.SecurityContext.AllowPrivilegeEscalation = &resolverAllowPrivilegeEscalation
	container.SecurityContext.ReadOnlyRootFilesystem = &resolverReadOnlyRootFilesystem
	container.SecurityContext.RunAsNonRoot = &resolverRunAsNonRoot
	container.SecurityContext.RunAsUser = &resolverRunAsUser
	container.SecurityContext.Capabilities = &coreV1.Capabilities{
		Drop: resolverCapabilitiesToDrop,
	}
}

func (obj *ResolverDeploymentK8sObject) mutateVolumesMounts(container *coreV1.Container) {
	if container.VolumeMounts == nil || len(container.VolumeMounts) != 1 {
		container.VolumeMounts = make([]coreV1.VolumeMount, 0)
	}

	commonState.MutateVolumeMountToIncludeRootCAsVolumeMount(container)
}