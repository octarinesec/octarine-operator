package objects

import (
	"fmt"

	cbContainersV1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/api/v1/common_specs"
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
	ResolverLabelKey = "app.kubernetes.io/name"

	ResolverDesiredGRPCPortName = "grpc"
)

var (
	ResolverAllowPrivilegeEscalation       = false
	ResolverReadOnlyRootFilesystem         = true
	ResolverRunAsUser                int64 = 1500
	ResolverCapabilitiesToDrop             = []coreV1.Capability{"ALL"}
)

type ResolverDeploymentK8sObject struct{}

func NewResolverDeploymentK8sObject() *ResolverDeploymentK8sObject {
	return &ResolverDeploymentK8sObject{}
}

func (obj *ResolverDeploymentK8sObject) EmptyK8sObject() client.Object {
	return &appsV1.Deployment{}
}

func (obj *ResolverDeploymentK8sObject) RuntimeChildNamespacedName(_ *cbContainersV1.CBContainersRuntime) types.NamespacedName {
	return types.NamespacedName{Name: ResolverName, Namespace: commonState.DataPlaneNamespaceName}
}

func (obj *ResolverDeploymentK8sObject) MutateRuntimeChildK8sObject(k8sObject client.Object, cbContainersRuntime *cbContainersV1.CBContainersRuntime) error {
	resolverSpec := &cbContainersRuntime.Spec.ResolverSpec
	deployment, ok := k8sObject.(*appsV1.Deployment)
	if !ok {
		return fmt.Errorf("expected Deployment K8s object")
	}

	desiredLabels := resolverSpec.Labels
	if desiredLabels == nil {
		desiredLabels = make(map[string]string)
	}
	desiredLabels[ResolverLabelKey] = ResolverName

	if deployment.Spec.Selector == nil {
		deployment.Spec.Selector = &metav1.LabelSelector{}
	}

	if deployment.ObjectMeta.Annotations == nil {
		deployment.ObjectMeta.Annotations = make(map[string]string)
	}

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	deployment.Spec.Replicas = resolverSpec.ReplicasCount
	deployment.ObjectMeta.Labels = desiredLabels
	deployment.Spec.Selector.MatchLabels = desiredLabels
	deployment.Spec.Template.ObjectMeta.Labels = desiredLabels
	deployment.Spec.Template.Spec.ServiceAccountName = commonState.DataPlaneServiceAccountName
	deployment.Spec.Template.Spec.PriorityClassName = commonState.DataPlanePriorityClassName
	deployment.Spec.Template.Spec.ImagePullSecrets = []coreV1.LocalObjectReference{{Name: commonState.RegistrySecretName}}
	obj.mutateAnnotations(deployment, resolverSpec)
	obj.mutateContainersList(&deployment.Spec.Template.Spec, resolverSpec,
		&cbContainersRuntime.Spec.ResolverSpec.EventsGatewaySpec, cbContainersRuntime.Spec.Version,
		cbContainersRuntime.Spec.AccessTokenSecretName, cbContainersRuntime.Spec.InternalGrpcPort)

	return nil
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
	eventsGatewaySpec *common_specs.CBContainersEventsGatewaySpec,
	version,
	accessTokenSecretName string,
	desiredGRPCPort int32) {

	if len(templatePodSpec.Containers) != 1 {
		container := coreV1.Container{}
		templatePodSpec.Containers = []coreV1.Container{container}
	}

	obj.mutateContainer(&templatePodSpec.Containers[0], resolverSpec, eventsGatewaySpec,
		version, accessTokenSecretName, desiredGRPCPort)
}

func (obj *ResolverDeploymentK8sObject) mutateContainer(
	container *coreV1.Container,
	resolverSpec *cbContainersV1.CBContainersRuntimeResolverSpec,
	eventsGatewaySpec *common_specs.CBContainersEventsGatewaySpec,
	version,
	accessTokenSecretName string,
	desiredGRPCPort int32) {

	container.Name = ResolverName
	container.Resources = resolverSpec.Resources
	commonState.MutateImage(container, resolverSpec.Image, version)
	commonState.MutateContainerHTTPProbes(container, resolverSpec.Probes)
	obj.mutateEnvVars(container, resolverSpec, eventsGatewaySpec, accessTokenSecretName)
	obj.mutateContainerPorts(container, desiredGRPCPort)
	obj.mutateSecurityContext(container)
}

func (obj *ResolverDeploymentK8sObject) mutateContainerPorts(container *coreV1.Container, desiredGRPCPort int32) {
	if container.Ports == nil || len(container.Ports) != 1 {
		container.Ports = []coreV1.ContainerPort{{}}
	}

	container.Ports[0].Name = ResolverDesiredGRPCPortName
	container.Ports[0].ContainerPort = desiredGRPCPort
}

func (obj *ResolverDeploymentK8sObject) mutateEnvVars(
	container *coreV1.Container,
	resolverSpec *cbContainersV1.CBContainersRuntimeResolverSpec,
	eventsGatewaySpec *common_specs.CBContainersEventsGatewaySpec,
	accessTokenSecretName string,) {

	commonState.MutateEnvVars(container, resolverSpec.Env, accessTokenSecretName, eventsGatewaySpec,
		coreV1.EnvVar{Name: "RUNTIME_KUBERNETES_RESOLVER_PROMETHEUS_PORT", Value: fmt.Sprintf("%d", resolverSpec.Prometheus.Port)},
		coreV1.EnvVar{Name: "GIN_MODE", Value: "release"},
	)
}

func (obj *ResolverDeploymentK8sObject) mutateSecurityContext(container *coreV1.Container) {
	if container.SecurityContext == nil {
		container.SecurityContext = &coreV1.SecurityContext{}
	}

	container.SecurityContext.AllowPrivilegeEscalation = &ResolverAllowPrivilegeEscalation
	container.SecurityContext.ReadOnlyRootFilesystem = &ResolverReadOnlyRootFilesystem
	container.SecurityContext.RunAsUser = &ResolverRunAsUser
	container.SecurityContext.Capabilities = &coreV1.Capabilities{
		Drop: ResolverCapabilitiesToDrop,
	}
}
