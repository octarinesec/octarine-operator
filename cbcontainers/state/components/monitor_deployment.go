package components

import (
	"fmt"

	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	MonitorName     = "cbcontainers-monitor"
	MonitorLabelKey = "app.kubernetes.io/name"

	// MonitorAgentVersionEnvVarKey is the name of the monitor environment variable that holds the value of the agent version.
	MonitorAgentVersionEnvVarKey = "MONITOR_AGENT_VERSION"
	// MonitorDataplaneNamespaceEnvVarKey is the name of the monitor environment variable that holds the value of the dataplane namespace.
	MonitorDataplaneNamespaceEnvVarKey = "MONITOR_DATAPLANE_NAMESPACE"
)

var (
	MonitorReplicas                 int32 = 1
	MonitorAllowPrivilegeEscalation       = false
	MonitorReadOnlyRootFilesystem         = true
	MonitorRunAsUser                int64 = 1500
	MonitorRunAsNonRoot                   = true
	MonitorCapabilitiesToDrop             = []coreV1.Capability{"ALL"}
)

type MonitorDeploymentK8sObject struct {
	// Namespace is the Namespace in which the Deployment will be created.
	Namespace string
}

func NewMonitorDeploymentK8sObject() *MonitorDeploymentK8sObject {
	return &MonitorDeploymentK8sObject{
		Namespace: commonState.DataPlaneNamespaceName,
	}
}

func (obj *MonitorDeploymentK8sObject) EmptyK8sObject() client.Object {
	return &appsV1.Deployment{}
}

func (obj *MonitorDeploymentK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: MonitorName, Namespace: obj.Namespace}
}

func (obj *MonitorDeploymentK8sObject) MutateK8sObject(k8sObject client.Object, agentSpec *cbcontainersv1.CBContainersAgentSpec) error {
	monitor := &agentSpec.Components.Basic.Monitor
	deployment, ok := k8sObject.(*appsV1.Deployment)
	if !ok {
		return fmt.Errorf("expected Deployment K8s object")
	}

	desiredLabels := monitor.Labels
	if desiredLabels == nil {
		desiredLabels = make(map[string]string)
	}
	desiredLabels[MonitorLabelKey] = MonitorName

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
	deployment.Spec.Replicas = &MonitorReplicas
	deployment.ObjectMeta.Labels = desiredLabels
	deployment.Spec.Selector.MatchLabels = desiredLabels
	deployment.Spec.Template.ObjectMeta.Labels = desiredLabels
	deployment.Spec.Template.Spec.ServiceAccountName = commonState.MonitorServiceAccountName
	deployment.Spec.Template.Spec.PriorityClassName = commonState.DataPlanePriorityClassName
	applyment.EnforceMapContains(deployment.ObjectMeta.Annotations, monitor.DeploymentAnnotations)
	applyment.EnforceMapContains(deployment.Spec.Template.ObjectMeta.Annotations, monitor.PodTemplateAnnotations)
	if agentSpec.Components.Settings.CreateDefaultImagePullSecrets {
		deployment.Spec.Template.Spec.ImagePullSecrets = []coreV1.LocalObjectReference{{Name: commonState.RegistrySecretName}}
	}
	for _, secretName := range agentSpec.Components.Settings.ImagePullSecrets {
		deployment.Spec.Template.Spec.ImagePullSecrets = append(deployment.Spec.Template.Spec.ImagePullSecrets, coreV1.LocalObjectReference{Name: secretName})
	}
	for _, secretName := range agentSpec.Components.Basic.Monitor.ImagePullSecrets {
		deployment.Spec.Template.Spec.ImagePullSecrets = append(deployment.Spec.Template.Spec.ImagePullSecrets, coreV1.LocalObjectReference{Name: secretName})
	}
	obj.mutateVolumes(&deployment.Spec.Template.Spec)
	obj.mutateAffinityAndNodeSelector(&deployment.Spec.Template.Spec, monitor)
	obj.mutateContainersList(&deployment.Spec.Template.Spec, monitor, &agentSpec.Gateways.CoreEventsGateway, agentSpec.Version, agentSpec.AccessTokenSecretName)
	commonState.NewNodeTermsBuilder(&deployment.Spec.Template.Spec).Build()

	return nil
}

func (obj *MonitorDeploymentK8sObject) mutateVolumes(templatePodSpec *coreV1.PodSpec) {
	if templatePodSpec.Volumes == nil || len(templatePodSpec.Volumes) != 1 {
		templatePodSpec.Volumes = make([]coreV1.Volume, 0)
	}

	commonState.MutateVolumesToIncludeRootCAsVolume(templatePodSpec)
}

func (obj *MonitorDeploymentK8sObject) mutateAffinityAndNodeSelector(templatePodSpec *coreV1.PodSpec, monitorSpec *cbcontainersv1.CBContainersMonitorSpec) {
	templatePodSpec.Affinity = monitorSpec.Affinity
	templatePodSpec.NodeSelector = monitorSpec.NodeSelector
}

func (obj *MonitorDeploymentK8sObject) mutateContainersList(templatePodSpec *coreV1.PodSpec, monitorSpec *cbcontainersv1.CBContainersMonitorSpec, eventsGatewaySpec *cbcontainersv1.CBContainersEventsGatewaySpec, version, accessTokenSecretName string) {
	if len(templatePodSpec.Containers) != 1 {
		container := coreV1.Container{}
		templatePodSpec.Containers = []coreV1.Container{container}
	}

	obj.mutateContainer(&templatePodSpec.Containers[0], monitorSpec, eventsGatewaySpec, version, accessTokenSecretName)
}

func (obj *MonitorDeploymentK8sObject) mutateContainer(container *coreV1.Container, monitorSpec *cbcontainersv1.CBContainersMonitorSpec, eventsGatewaySpec *cbcontainersv1.CBContainersEventsGatewaySpec, version, accessTokenSecretName string) {
	container.Name = MonitorName
	container.Resources = monitorSpec.Resources

	envVarBuilder := commonState.NewEnvVarBuilder().
		WithCommonDataPlane(accessTokenSecretName).
		WithEventsGateway(eventsGatewaySpec).
		WithEnvVarFromConfigmap(MonitorAgentVersionEnvVarKey, commonState.DataPlaneConfigmapAgentVersionKey).
		WithEnvVarFromConfigmap(MonitorDataplaneNamespaceEnvVarKey, commonState.DataPlaneConfigmapDataplaneNamespaceKey).
		WithSpec(monitorSpec.Env)
	commonState.MutateEnvVars(container, envVarBuilder)

	commonState.MutateImage(container, monitorSpec.Image, version)
	commonState.MutateContainerHTTPProbes(container, monitorSpec.Probes)
	obj.mutateSecurityContext(container)
	obj.mutateVolumesMounts(container)
}

func (obj *MonitorDeploymentK8sObject) mutateSecurityContext(container *coreV1.Container) {
	if container.SecurityContext == nil {
		container.SecurityContext = &coreV1.SecurityContext{}
	}

	container.SecurityContext.AllowPrivilegeEscalation = &MonitorAllowPrivilegeEscalation
	container.SecurityContext.ReadOnlyRootFilesystem = &MonitorReadOnlyRootFilesystem
	container.SecurityContext.RunAsNonRoot = &MonitorRunAsNonRoot
	container.SecurityContext.RunAsUser = &MonitorRunAsUser
	container.SecurityContext.Capabilities = &coreV1.Capabilities{
		Drop: MonitorCapabilitiesToDrop,
	}
}

func (obj *MonitorDeploymentK8sObject) mutateVolumesMounts(container *coreV1.Container) {
	if container.VolumeMounts == nil || len(container.VolumeMounts) != 1 {
		container.VolumeMounts = make([]coreV1.VolumeMount, 0)
	}

	commonState.MutateVolumeMountToIncludeRootCAsVolumeMount(container)
}
