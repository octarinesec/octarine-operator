package objects

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
)

var (
	MonitorReplicas                 int32 = 1
	MonitorAllowPrivilegeEscalation       = false
	MonitorReadOnlyRootFilesystem         = true
	MonitorRunAsUser                int64 = 1500
	MonitorCapabilitiesToDrop             = []coreV1.Capability{"ALL"}
)

type MonitorDeploymentK8sObject struct{}

func NewMonitorDeploymentK8sObject() *MonitorDeploymentK8sObject {
	return &MonitorDeploymentK8sObject{}
}

func (obj *MonitorDeploymentK8sObject) EmptyK8sObject() client.Object {
	return &appsV1.Deployment{}
}

func (obj *MonitorDeploymentK8sObject) ClusterChildNamespacedName(_ *cbcontainersv1.CBContainersAgent) types.NamespacedName {
	return types.NamespacedName{Name: MonitorName, Namespace: commonState.DataPlaneNamespaceName}
}

func (obj *MonitorDeploymentK8sObject) MutateClusterChildK8sObject(k8sObject client.Object, cbContainersCluster *cbcontainersv1.CBContainersAgent) error {
	monitorSpec := cbContainersCluster.Spec.ClusterSpec.MonitorSpec
	deployment, ok := k8sObject.(*appsV1.Deployment)
	if !ok {
		return fmt.Errorf("expected Deployment K8s object")
	}

	desiredLabels := monitorSpec.Labels
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
	deployment.Spec.Template.Spec.ServiceAccountName = commonState.DataPlaneServiceAccountName
	deployment.Spec.Template.Spec.PriorityClassName = commonState.DataPlanePriorityClassName
	applyment.EnforceMapContains(deployment.ObjectMeta.Annotations, monitorSpec.DeploymentAnnotations)
	applyment.EnforceMapContains(deployment.Spec.Template.ObjectMeta.Annotations, monitorSpec.PodTemplateAnnotations)
	deployment.Spec.Template.Spec.ImagePullSecrets = []coreV1.LocalObjectReference{{Name: commonState.RegistrySecretName}}
	obj.mutateContainersList(&deployment.Spec.Template.Spec, &cbContainersCluster.Spec.ClusterSpec.MonitorSpec, &cbContainersCluster.Spec.ClusterSpec.EventsGatewaySpec, cbContainersCluster.Spec.Version, cbContainersCluster.Spec.ApiGatewaySpec.AccessTokenSecretName)

	return nil
}

func (obj *MonitorDeploymentK8sObject) mutateContainersList(templatePodSpec *coreV1.PodSpec, monitorSpec *cbcontainersv1.CBContainersClusterMonitorSpec, eventsGatewaySpec *cbcontainersv1.CBContainersEventsGatewaySpec, version, accessTokenSecretName string) {
	if len(templatePodSpec.Containers) != 1 {
		container := coreV1.Container{}
		templatePodSpec.Containers = []coreV1.Container{container}
	}

	obj.mutateContainer(&templatePodSpec.Containers[0], monitorSpec, eventsGatewaySpec, version, accessTokenSecretName)
}

func (obj *MonitorDeploymentK8sObject) mutateContainer(container *coreV1.Container, monitorSpec *cbcontainersv1.CBContainersClusterMonitorSpec, eventsGatewaySpec *cbcontainersv1.CBContainersEventsGatewaySpec, version, accessTokenSecretName string) {
	container.Name = MonitorName
	container.Resources = monitorSpec.Resources

	envVarBuilder := commonState.NewEnvVarBuilder().
		WithCommonDataPlane(accessTokenSecretName).
		WithEventsGateway(eventsGatewaySpec).
		WithSpec(monitorSpec.Env)
	commonState.MutateEnvVars(container, envVarBuilder)

	commonState.MutateImage(container, monitorSpec.Image, version)
	commonState.MutateContainerHTTPProbes(container, monitorSpec.Probes)
	obj.mutateSecurityContext(container)
}

func (obj *MonitorDeploymentK8sObject) mutateSecurityContext(container *coreV1.Container) {
	if container.SecurityContext == nil {
		container.SecurityContext = &coreV1.SecurityContext{}
	}

	container.SecurityContext.AllowPrivilegeEscalation = &MonitorAllowPrivilegeEscalation
	container.SecurityContext.ReadOnlyRootFilesystem = &MonitorReadOnlyRootFilesystem
	container.SecurityContext.RunAsUser = &MonitorRunAsUser
	container.SecurityContext.Capabilities = &coreV1.Capabilities{
		Drop: MonitorCapabilitiesToDrop,
	}
}
