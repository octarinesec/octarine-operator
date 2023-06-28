package components

import (
	"context"
	"fmt"
	cbContainersV1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"math"
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

type ResolverDeploymentK8sObject struct {
	// Namespace is the Namespace in which the Deployment will be created.
	Namespace string
	APIReader client.Reader
}

func NewResolverDeploymentK8sObject(apiReader client.Reader) *ResolverDeploymentK8sObject {
	return &ResolverDeploymentK8sObject{
		Namespace: commonState.DataPlaneNamespaceName,
		APIReader: apiReader,
	}
}

func (obj *ResolverDeploymentK8sObject) EmptyK8sObject() client.Object {
	return &appsV1.Deployment{}
}

func (obj *ResolverDeploymentK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: ResolverName, Namespace: obj.Namespace}
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

	defaultReplicasCount := int32(1)
	replicasCount := &defaultReplicasCount

	if resolver.ReplicasCount != nil {
		replicasCount = resolver.ReplicasCount
	} else {
		if dynamicReplicasCount, err := obj.getDynamicReplicasCount(resolver.NodesToReplicasRatio); err == nil {
			replicasCount = dynamicReplicasCount
		}
	}

	deployment.Namespace = obj.Namespace
	deployment.Spec.Replicas = replicasCount
	deployment.ObjectMeta.Labels = desiredLabels
	deployment.Spec.Selector.MatchLabels = desiredLabels
	deployment.Spec.Template.ObjectMeta.Labels = desiredLabels
	deployment.Spec.Template.Spec.ServiceAccountName = commonState.RuntimeResolverServiceAccountName
	deployment.Spec.Template.Spec.PriorityClassName = commonState.DataPlanePriorityClassName
	desiredImagePullSecrets := getImagePullSecrets(agentSpec, agentSpec.Components.RuntimeProtection.Resolver.Image.PullSecrets...)
	if objectsDiffer(desiredImagePullSecrets, deployment.Spec.Template.Spec.ImagePullSecrets) {
		deployment.Spec.Template.Spec.ImagePullSecrets = getImagePullSecrets(agentSpec, agentSpec.Components.RuntimeProtection.Resolver.Image.PullSecrets...)
	}
	obj.mutateAnnotations(deployment, agentSpec)
	obj.mutateVolumes(deployment, agentSpec)
	obj.mutateAffinityAndNodeSelector(deployment, agentSpec)
	obj.mutateContainersList(deployment, agentSpec)
	commonState.NewNodeTermsBuilder(&deployment.Spec.Template.Spec).Build()

	return nil
}

func (obj *ResolverDeploymentK8sObject) mutateVolumes(deployment *appsV1.Deployment, agentSpec *cbContainersV1.CBContainersAgentSpec) {
	templatePodSpec := &deployment.Spec.Template.Spec
	if templatePodSpec.Volumes == nil || len(templatePodSpec.Volumes) != 1 {
		templatePodSpec.Volumes = make([]coreV1.Volume, 0)
	}

	commonState.MutateVolumesToIncludeRootCAsVolume(templatePodSpec)
}

func (obj *ResolverDeploymentK8sObject) mutateAffinityAndNodeSelector(deployment *appsV1.Deployment, agentSpec *cbContainersV1.CBContainersAgentSpec) {
	resolverSpec := &agentSpec.Components.RuntimeProtection.Resolver

	templatePodSpec := &deployment.Spec.Template.Spec
	templatePodSpec.Affinity = resolverSpec.Affinity
	templatePodSpec.NodeSelector = resolverSpec.NodeSelector
}

func (obj *ResolverDeploymentK8sObject) mutateAnnotations(deployment *appsV1.Deployment, agentSpec *cbContainersV1.CBContainersAgentSpec) {
	resolverSpec := &agentSpec.Components.RuntimeProtection.Resolver

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
	deployment *appsV1.Deployment,
	agentSpec *cbContainersV1.CBContainersAgentSpec) {

	templatePodSpec := &deployment.Spec.Template.Spec
	if len(templatePodSpec.Containers) != 1 {
		container := coreV1.Container{}
		templatePodSpec.Containers = []coreV1.Container{container}
	}

	obj.mutateContainer(&templatePodSpec.Containers[0], agentSpec)
}

func (obj *ResolverDeploymentK8sObject) mutateContainer(
	container *coreV1.Container,
	agentSpec *cbContainersV1.CBContainersAgentSpec) {

	resolverSpec := &agentSpec.Components.RuntimeProtection.Resolver

	container.Name = ResolverName
	container.Resources = resolverSpec.Resources
	commonState.MutateImage(container, resolverSpec.Image, agentSpec.Version)
	commonState.MutateContainerHTTPProbes(container, resolverSpec.Probes)
	obj.mutateEnvVars(container, agentSpec)
	obj.mutateContainerPorts(container, agentSpec)
	obj.mutateSecurityContext(container)
	obj.mutateVolumesMounts(container)
}

func (obj *ResolverDeploymentK8sObject) mutateContainerPorts(container *coreV1.Container, agentSpec *cbContainersV1.CBContainersAgentSpec) {
	if container.Ports == nil || len(container.Ports) != 1 {
		container.Ports = []coreV1.ContainerPort{{}}
	}

	container.Ports[0].Name = desiredDeploymentGRPCPortName
	container.Ports[0].ContainerPort = agentSpec.Components.RuntimeProtection.InternalGrpcPort
}

func (obj *ResolverDeploymentK8sObject) mutateEnvVars(
	container *coreV1.Container, agentSpec *cbContainersV1.CBContainersAgentSpec) {

	runtimeProtection := &agentSpec.Components.RuntimeProtection
	resolverSpec := &runtimeProtection.Resolver
	desiredGRPCPortValue := runtimeProtection.InternalGrpcPort
	eventsGatewaySpec := &agentSpec.Gateways.RuntimeEventsGateway
	accessTokenSecretName := agentSpec.AccessTokenSecretName

	customEnvs := []coreV1.EnvVar{
		{Name: "RUNTIME_KUBERNETES_RESOLVER_GRPC_PORT", Value: fmt.Sprintf("%d", desiredGRPCPortValue)},
		{Name: "RUNTIME_KUBERNETES_RESOLVER_LOG_LEVEL", Value: runtimeProtection.Resolver.LogLevel},
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

func (obj *ResolverDeploymentK8sObject) getDynamicReplicasCount(nodesToReplicasRatio int32) (*int32, error) {
	nodesList := &coreV1.NodeList{}
	if err := obj.APIReader.List(context.Background(), nodesList); err != nil || nodesList.Items == nil || len(nodesList.Items) < 1 {
		return nil, fmt.Errorf("error getting list of nodes: %v", err)
	}
	replicasCount := int32(math.Ceil(float64(len(nodesList.Items)) / float64(nodesToReplicasRatio)))

	return &replicasCount, nil
}
