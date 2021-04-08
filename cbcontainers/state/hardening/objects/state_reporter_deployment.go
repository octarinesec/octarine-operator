package objects

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
	StateReporterName = "cbcontainers-hardening-state-reporter"
)

var (
	StateReporterReplicas                 int32 = 1
	StateReporterAllowPrivilegeEscalation       = false
	StateReporterReadOnlyRootFilesystem         = true
	StateReporterRunAsUser                int64 = 1500
	StateReporterCapabilitiesToDrop             = []coreV1.Capability{"ALL"}
)

type StateReporterDeploymentK8sObject struct{}

func NewStateReporterDeploymentK8sObject() *StateReporterDeploymentK8sObject {
	return &StateReporterDeploymentK8sObject{}
}

func (obj *StateReporterDeploymentK8sObject) EmptyK8sObject() client.Object {
	return &appsV1.Deployment{}
}

func (obj *StateReporterDeploymentK8sObject) HardeningChildNamespacedName(_ *cbContainersV1.CBContainersHardening) types.NamespacedName {
	return types.NamespacedName{Name: StateReporterName, Namespace: commonState.DataPlaneNamespaceName}
}

func (obj *StateReporterDeploymentK8sObject) MutateHardeningChildK8sObject(k8sObject client.Object, cbContainersHardening *cbContainersV1.CBContainersHardening) error {
	stateReporterSpec := cbContainersHardening.Spec.StateReporterSpec
	deployment, ok := k8sObject.(*appsV1.Deployment)
	if !ok {
		return fmt.Errorf("expected Deployment K8s object")
	}

	if deployment.Spec.Selector == nil {
		deployment.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: stateReporterSpec.PodTemplateLabels,
		}
	}

	if deployment.ObjectMeta.Annotations == nil {
		deployment.ObjectMeta.Annotations = make(map[string]string)
	}

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	deployment.Spec.Replicas = &StateReporterReplicas
	deployment.ObjectMeta.Labels = stateReporterSpec.DeploymentLabels
	deployment.Spec.Selector.MatchLabels = stateReporterSpec.PodTemplateLabels
	deployment.Spec.Template.ObjectMeta.Labels = stateReporterSpec.PodTemplateLabels
	deployment.Spec.Template.Spec.ServiceAccountName = commonState.DataPlaneServiceAccountName
	deployment.Spec.Template.Spec.PriorityClassName = commonState.DataPlanePriorityClassName
	applyment.EnforceMapContains(deployment.ObjectMeta.Annotations, stateReporterSpec.DeploymentAnnotations)
	applyment.EnforceMapContains(deployment.Spec.Template.ObjectMeta.Annotations, stateReporterSpec.PodTemplateAnnotations)
	deployment.Spec.Template.Spec.ImagePullSecrets = []coreV1.LocalObjectReference{{Name: commonState.RegistrySecretName}}
	obj.mutateContainersList(&deployment.Spec.Template.Spec, &cbContainersHardening.Spec.StateReporterSpec, &cbContainersHardening.Spec.EventsGatewaySpec, cbContainersHardening.Spec.Version, cbContainersHardening.Spec.AccessTokenSecretName)

	return nil
}

func (obj *StateReporterDeploymentK8sObject) mutateContainersList(templatePodSpec *coreV1.PodSpec, stateReporterSpec *cbContainersV1.CBContainersHardeningStateReporterSpec, eventsGatewaySpec *cbContainersV1.CBContainersHardeningEventsGatewaySpec, version, accessTokenSecretName string) {
	if len(templatePodSpec.Containers) != 1 {
		container := coreV1.Container{}
		templatePodSpec.Containers = []coreV1.Container{container}
	}

	obj.mutateContainer(&templatePodSpec.Containers[0], stateReporterSpec, eventsGatewaySpec, version, accessTokenSecretName)
}

func (obj *StateReporterDeploymentK8sObject) mutateContainer(container *coreV1.Container, stateReporterSpec *cbContainersV1.CBContainersHardeningStateReporterSpec, eventsGatewaySpec *cbContainersV1.CBContainersHardeningEventsGatewaySpec, version, accessTokenSecretName string) {
	container.Name = StateReporterName
	container.Resources = stateReporterSpec.Resources
	mutateEnvVars(container, stateReporterSpec.Env, accessTokenSecretName, eventsGatewaySpec)
	mutateImage(container, stateReporterSpec.Image, version)
	obj.mutateSecurityContext(container)
}

func (obj *StateReporterDeploymentK8sObject) mutateSecurityContext(container *coreV1.Container) {
	if container.SecurityContext == nil {
		container.SecurityContext = &coreV1.SecurityContext{}
	}

	container.SecurityContext.AllowPrivilegeEscalation = &StateReporterAllowPrivilegeEscalation
	container.SecurityContext.ReadOnlyRootFilesystem = &StateReporterReadOnlyRootFilesystem
	container.SecurityContext.RunAsUser = &StateReporterRunAsUser
	container.SecurityContext.Capabilities = &coreV1.Capabilities{
		Drop: StateReporterCapabilitiesToDrop,
	}
}
