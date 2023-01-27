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
	StateReporterName     = "cbcontainers-hardening-state-reporter"
	StateReporterLabelKey = "app.kubernetes.io/name"
)

var (
	StateReporterReplicas                 int32 = 1
	StateReporterAllowPrivilegeEscalation       = false
	StateReporterReadOnlyRootFilesystem         = true
	StateReporterRunAsUser                int64 = 1500
	StateReporterRunAsNonRoot                   = true
	StateReporterCapabilitiesToDrop             = []coreV1.Capability{"ALL"}
)

type StateReporterDeploymentK8sObject struct {
	// Namespace is the Namespace in which the Deployment will be created.
	Namespace string
}

func NewStateReporterDeploymentK8sObject() *StateReporterDeploymentK8sObject {
	return &StateReporterDeploymentK8sObject{
		Namespace: commonState.DataPlaneNamespaceName,
	}
}

func (obj *StateReporterDeploymentK8sObject) EmptyK8sObject() client.Object {
	return &appsV1.Deployment{}
}

func (obj *StateReporterDeploymentK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: StateReporterName, Namespace: obj.Namespace}
}

func (obj *StateReporterDeploymentK8sObject) MutateK8sObject(k8sObject client.Object, agentSpec *cbContainersV1.CBContainersAgentSpec) error {
	deployment, ok := k8sObject.(*appsV1.Deployment)
	if !ok {
		return fmt.Errorf("expected Deployment K8s object")
	}

	stateReporter := &agentSpec.Components.Basic.StateReporter

	desiredLabels := stateReporter.Labels
	if desiredLabels == nil {
		desiredLabels = make(map[string]string)
	}
	desiredLabels[StateReporterLabelKey] = StateReporterName

	if deployment.Spec.Selector == nil {
		deployment.Spec.Selector = &metav1.LabelSelector{}
	}

	if deployment.ObjectMeta.Annotations == nil {
		deployment.ObjectMeta.Annotations = make(map[string]string)
	}

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	deployment.Namespace = agentSpec.Namespace
	deployment.Spec.Replicas = &StateReporterReplicas
	deployment.ObjectMeta.Labels = desiredLabels
	deployment.Spec.Selector.MatchLabels = desiredLabels
	deployment.Spec.Template.ObjectMeta.Labels = desiredLabels
	deployment.Spec.Template.Spec.ServiceAccountName = commonState.StateReporterServiceAccountName
	deployment.Spec.Template.Spec.PriorityClassName = commonState.DataPlanePriorityClassName
	applyment.EnforceMapContains(deployment.ObjectMeta.Annotations, stateReporter.DeploymentAnnotations)
	applyment.EnforceMapContains(deployment.Spec.Template.ObjectMeta.Annotations, stateReporter.PodTemplateAnnotations)
	if agentSpec.Components.Settings.CreateDefaultImagePullSecrets {
		deployment.Spec.Template.Spec.ImagePullSecrets = []coreV1.LocalObjectReference{{Name: commonState.RegistrySecretName}}
	}
	for _, secretName := range agentSpec.Components.Settings.ImagePullSecrets {
		deployment.Spec.Template.Spec.ImagePullSecrets = append(deployment.Spec.Template.Spec.ImagePullSecrets, coreV1.LocalObjectReference{Name: secretName})
	}
	for _, secretName := range agentSpec.Components.Basic.StateReporter.Image.PullSecrets {
		deployment.Spec.Template.Spec.ImagePullSecrets = append(deployment.Spec.Template.Spec.ImagePullSecrets, coreV1.LocalObjectReference{Name: secretName})
	}
	obj.mutateVolumes(&deployment.Spec.Template.Spec)
	obj.mutateAffinityAndNodeSelector(&deployment.Spec.Template.Spec, stateReporter)
	obj.mutateContainersList(&deployment.Spec.Template.Spec, stateReporter, &agentSpec.Gateways.HardeningEventsGateway, agentSpec.Version, agentSpec.AccessTokenSecretName)
	commonState.NewNodeTermsBuilder(&deployment.Spec.Template.Spec).Build()

	return nil
}

func (obj *StateReporterDeploymentK8sObject) mutateVolumes(templatePodSpec *coreV1.PodSpec) {
	if templatePodSpec.Volumes == nil || len(templatePodSpec.Volumes) != 1 {
		templatePodSpec.Volumes = make([]coreV1.Volume, 0)
	}

	commonState.MutateVolumesToIncludeRootCAsVolume(templatePodSpec)
}

func (obj *StateReporterDeploymentK8sObject) mutateAffinityAndNodeSelector(templatePodSpec *coreV1.PodSpec, stateReporterSpec *cbContainersV1.CBContainersStateReporterSpec) {
	templatePodSpec.Affinity = stateReporterSpec.Affinity
	templatePodSpec.NodeSelector = stateReporterSpec.NodeSelector
}

func (obj *StateReporterDeploymentK8sObject) mutateContainersList(templatePodSpec *coreV1.PodSpec, stateReporterSpec *cbContainersV1.CBContainersStateReporterSpec, eventsGatewaySpec *cbContainersV1.CBContainersEventsGatewaySpec, version, accessTokenSecretName string) {
	if len(templatePodSpec.Containers) != 1 {
		container := coreV1.Container{}
		templatePodSpec.Containers = []coreV1.Container{container}
	}

	obj.mutateContainer(&templatePodSpec.Containers[0], stateReporterSpec, eventsGatewaySpec, version, accessTokenSecretName)
}

func (obj *StateReporterDeploymentK8sObject) mutateContainer(container *coreV1.Container, stateReporterSpec *cbContainersV1.CBContainersStateReporterSpec, eventsGatewaySpec *cbContainersV1.CBContainersEventsGatewaySpec, version, accessTokenSecretName string) {
	container.Name = StateReporterName
	container.Resources = stateReporterSpec.Resources

	envVarBuilder := commonState.NewEnvVarBuilder().
		WithCommonDataPlane(accessTokenSecretName).
		WithEventsGateway(eventsGatewaySpec).
		WithSpec(stateReporterSpec.Env)
	commonState.MutateEnvVars(container, envVarBuilder)

	commonState.MutateImage(container, stateReporterSpec.Image, version)
	commonState.MutateContainerHTTPProbes(container, stateReporterSpec.Probes)
	obj.mutateSecurityContext(container)
	obj.mutateVolumesMounts(container)
}

func (obj *StateReporterDeploymentK8sObject) mutateSecurityContext(container *coreV1.Container) {
	if container.SecurityContext == nil {
		container.SecurityContext = &coreV1.SecurityContext{}
	}

	container.SecurityContext.AllowPrivilegeEscalation = &StateReporterAllowPrivilegeEscalation
	container.SecurityContext.ReadOnlyRootFilesystem = &StateReporterReadOnlyRootFilesystem
	container.SecurityContext.RunAsNonRoot = &StateReporterRunAsNonRoot
	container.SecurityContext.RunAsUser = &StateReporterRunAsUser
	container.SecurityContext.Capabilities = &coreV1.Capabilities{
		Drop: StateReporterCapabilitiesToDrop,
	}
}

func (obj *StateReporterDeploymentK8sObject) mutateVolumesMounts(container *coreV1.Container) {
	if container.VolumeMounts == nil || len(container.VolumeMounts) != 1 {
		container.VolumeMounts = make([]coreV1.VolumeMount, 0)
	}

	commonState.MutateVolumeMountToIncludeRootCAsVolumeMount(container)
}
