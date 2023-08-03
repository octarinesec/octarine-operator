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

func NewMonitorDeploymentK8sObject(namespace string) *MonitorDeploymentK8sObject {
	return &MonitorDeploymentK8sObject{
		Namespace: namespace,
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

	deployment.Spec.Replicas = &MonitorReplicas
	deployment.ObjectMeta.Labels = desiredLabels
	deployment.Spec.Selector.MatchLabels = desiredLabels
	deployment.Spec.Template.ObjectMeta.Labels = desiredLabels
	deployment.Spec.Template.Spec.ServiceAccountName = commonState.MonitorServiceAccountName
	deployment.Spec.Template.Spec.PriorityClassName = commonState.DataPlanePriorityClassName
	applyment.EnforceMapContains(deployment.ObjectMeta.Annotations, monitor.DeploymentAnnotations)
	applyment.EnforceMapContains(deployment.Spec.Template.ObjectMeta.Annotations, monitor.PodTemplateAnnotations)
	desiredImagePullSecrets := getImagePullSecrets(agentSpec, agentSpec.Components.Basic.Monitor.Image.PullSecrets...)
	if objectsDiffer(desiredImagePullSecrets, deployment.Spec.Template.Spec.ImagePullSecrets) {
		deployment.Spec.Template.Spec.ImagePullSecrets = getImagePullSecrets(agentSpec, agentSpec.Components.Basic.Monitor.Image.PullSecrets...)
	}
	obj.mutateVolumes(&deployment.Spec.Template.Spec)
	obj.mutateAffinityAndNodeSelector(&deployment.Spec.Template.Spec, monitor)
	obj.mutateContainersList(&deployment.Spec.Template.Spec, agentSpec)
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

func (obj *MonitorDeploymentK8sObject) mutateContainersList(templatePodSpec *coreV1.PodSpec, agentSpec *cbcontainersv1.CBContainersAgentSpec) {
	if len(templatePodSpec.Containers) != 1 {
		container := coreV1.Container{}
		templatePodSpec.Containers = []coreV1.Container{container}
	}

	obj.mutateContainer(&templatePodSpec.Containers[0], agentSpec)
}

func (obj *MonitorDeploymentK8sObject) mutateContainer(container *coreV1.Container, agentSpec *cbcontainersv1.CBContainersAgentSpec) {
	monitorSpec := &agentSpec.Components.Basic.Monitor

	container.Name = MonitorName
	container.Resources = monitorSpec.Resources

	envVarBuilder := commonState.NewEnvVarBuilder().
		WithCommonDataPlane(agentSpec.AccessTokenSecretName).
		WithEventsGateway(&agentSpec.Gateways.CoreEventsGateway).
		WithEnvVarFromConfigmap(MonitorAgentVersionEnvVarKey, commonState.DataPlaneConfigmapAgentVersionKey).
		WithEnvVarFromConfigmap(MonitorDataplaneNamespaceEnvVarKey, commonState.DataPlaneConfigmapDataplaneNamespaceKey).
		WithSpec(monitorSpec.Env).
		WithProxySettings(agentSpec.Components.Settings.Proxy)
	commonState.MutateEnvVars(container, envVarBuilder)

	commonState.MutateImage(container, monitorSpec.Image, agentSpec.Version)
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
