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
	ControllerName     = "cbcontainers-runtime-controller"
	ControllerLabelKey = "app.kubernetes.io/name"
)

var (
	ControllerAllowPrivilegeEscalation       = false
	ControllerReadOnlyRootFilesystem         = true
	ControllerRunAsUser                int64 = 1500
	ControllerCapabilitiesToDrop             = []coreV1.Capability{"ALL"}
)

type ControllerDeploymentK8sObject struct{}

func NewControllerDeploymentK8sObject() *ControllerDeploymentK8sObject {
	return &ControllerDeploymentK8sObject{}
}

func (obj *ControllerDeploymentK8sObject) EmptyK8sObject() client.Object {
	return &appsV1.Deployment{}
}

func (obj *ControllerDeploymentK8sObject) RuntimeChildNamespacedName(_ *cbContainersV1.CBContainersRuntime) types.NamespacedName {
	return types.NamespacedName{Name: ControllerName, Namespace: commonState.DataPlaneNamespaceName}
}

func (obj *ControllerDeploymentK8sObject) MutateRuntimeChildK8sObject(k8sObject client.Object, cbContainersRuntime *cbContainersV1.CBContainersRuntime) error {
	controllerSpec := cbContainersRuntime.Spec.ControllerSpec
	deployment, ok := k8sObject.(*appsV1.Deployment)
	if !ok {
		return fmt.Errorf("expected Deployment K8s object")
	}

	desiredLabels := controllerSpec.Labels
	if desiredLabels == nil {
		desiredLabels = make(map[string]string)
	}
	desiredLabels[ControllerLabelKey] = ControllerName

	if deployment.Spec.Selector == nil {
		deployment.Spec.Selector = &metav1.LabelSelector{}
	}

	if deployment.ObjectMeta.Annotations == nil {
		deployment.ObjectMeta.Annotations = make(map[string]string)
	}

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	deployment.Spec.Replicas = controllerSpec.ReplicasCount
	deployment.ObjectMeta.Labels = desiredLabels
	deployment.Spec.Selector.MatchLabels = desiredLabels
	deployment.Spec.Template.ObjectMeta.Labels = desiredLabels
	deployment.Spec.Template.Spec.ServiceAccountName = commonState.DataPlaneServiceAccountName
	deployment.Spec.Template.Spec.PriorityClassName = commonState.DataPlanePriorityClassName
	applyment.EnforceMapContains(deployment.ObjectMeta.Annotations, controllerSpec.DeploymentAnnotations)
	applyment.EnforceMapContains(deployment.Spec.Template.ObjectMeta.Annotations, controllerSpec.PodTemplateAnnotations)
	deployment.Spec.Template.Spec.ImagePullSecrets = []coreV1.LocalObjectReference{{Name: commonState.RegistrySecretName}}
	obj.mutateContainersList(&deployment.Spec.Template.Spec, &cbContainersRuntime.Spec.ControllerSpec,
		&cbContainersRuntime.Spec.ControllerSpec.EventsGatewaySpec,cbContainersRuntime.Spec.Version, cbContainersRuntime.Spec.AccessTokenSecretName)

	return nil
}

func (obj *ControllerDeploymentK8sObject) mutateContainersList(templatePodSpec *coreV1.PodSpec, controllerSpec *cbContainersV1.CBContainersRuntimeControllerSpec, eventsGatewaySpec *common_specs.CBContainersEventsGatewaySpec, version, accessTokenSecretName string) {
	if len(templatePodSpec.Containers) != 1 {
		container := coreV1.Container{}
		templatePodSpec.Containers = []coreV1.Container{container}
	}

	obj.mutateContainer(&templatePodSpec.Containers[0], controllerSpec, eventsGatewaySpec, version, accessTokenSecretName)
}

func (obj *ControllerDeploymentK8sObject) mutateContainer(container *coreV1.Container, controllerSpec *cbContainersV1.CBContainersRuntimeControllerSpec, eventsGatewaySpec *common_specs.CBContainersEventsGatewaySpec, version, accessTokenSecretName string) {
	container.Name = ControllerName
	container.Resources = controllerSpec.Resources
	commonState.MutateEnvVars(container, controllerSpec.Env, accessTokenSecretName, eventsGatewaySpec)
	commonState.MutateImage(container, controllerSpec.Image, version)
	obj.mutateSecurityContext(container)
}

func (obj *ControllerDeploymentK8sObject) mutateSecurityContext(container *coreV1.Container) {
	if container.SecurityContext == nil {
		container.SecurityContext = &coreV1.SecurityContext{}
	}

	container.SecurityContext.AllowPrivilegeEscalation = &ControllerAllowPrivilegeEscalation
	container.SecurityContext.ReadOnlyRootFilesystem = &ControllerReadOnlyRootFilesystem
	container.SecurityContext.RunAsUser = &ControllerRunAsUser
	container.SecurityContext.Capabilities = &coreV1.Capabilities{
		Drop: ControllerCapabilitiesToDrop,
	}
}
