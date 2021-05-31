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
	sensorName     = "cbcontainers-runtime-sensor"
	sensorLabelKey = "app.kubernetes.io/name"

	sensorVerbosityFlag = "-v"
	sensorRunCommand    = "/run_worker.sh"

	desiredConnectionTimeoutSeconds = 60
)

var (
	sensorIsPrivileged       = true
	sensorRunAsUser    int64 = 0

	resolverAddress = fmt.Sprintf("%s.%s.svc.cluster.local", resolverName, commonState.DataPlaneNamespaceName)
)

type SensorDaemonSetK8sObject struct{}

func NewSensorDaemonSetK8sObject() *SensorDaemonSetK8sObject {
	return &SensorDaemonSetK8sObject{}
}

func (obj *SensorDaemonSetK8sObject) EmptyK8sObject() client.Object {
	return &appsV1.DaemonSet{}
}

func (obj *SensorDaemonSetK8sObject) RuntimeChildNamespacedName(_ *cbContainersV1.CBContainersRuntime) types.NamespacedName {
	return types.NamespacedName{Name: sensorName, Namespace: commonState.DataPlaneNamespaceName}
}

func (obj *SensorDaemonSetK8sObject) MutateRuntimeChildK8sObject(k8sObject client.Object, cbContainersRuntime *cbContainersV1.CBContainersRuntime) error {
	sensorSpec := &cbContainersRuntime.Spec.SensorSpec
	daemonSet, ok := k8sObject.(*appsV1.DaemonSet)
	if !ok {
		return fmt.Errorf("expected DaemonSet K8s object")
	}

	desiredLabels := sensorSpec.Labels
	if desiredLabels == nil {
		desiredLabels = make(map[string]string)
	}
	desiredLabels[sensorLabelKey] = sensorName

	if daemonSet.Spec.Selector == nil {
		daemonSet.Spec.Selector = &metav1.LabelSelector{}
	}

	if daemonSet.ObjectMeta.Annotations == nil {
		daemonSet.ObjectMeta.Annotations = make(map[string]string)
	}

	if daemonSet.Spec.Template.ObjectMeta.Annotations == nil {
		daemonSet.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	daemonSet.ObjectMeta.Labels = desiredLabels
	daemonSet.Spec.Selector.MatchLabels = desiredLabels
	daemonSet.Spec.Template.ObjectMeta.Labels = desiredLabels
	daemonSet.Spec.Template.Spec.ServiceAccountName = commonState.DataPlaneServiceAccountName
	daemonSet.Spec.Template.Spec.PriorityClassName = commonState.DataPlanePriorityClassName
	daemonSet.Spec.Template.Spec.ImagePullSecrets = []coreV1.LocalObjectReference{{Name: commonState.RegistrySecretName}}

	daemonSet.Spec.Template.Spec.DNSPolicy = coreV1.DNSClusterFirstWithHostNet
	daemonSet.Spec.Template.Spec.HostNetwork = true
	daemonSet.Spec.Template.Spec.HostPID = true

	obj.mutateAnnotations(daemonSet, sensorSpec)
	obj.mutateContainersList(&daemonSet.Spec.Template.Spec,
		sensorSpec,
		cbContainersRuntime.Spec.Version,
		cbContainersRuntime.Spec.AccessTokenSecretName,
		cbContainersRuntime.Spec.InternalGrpcPort,
	)

	return nil
}

func (obj *SensorDaemonSetK8sObject) mutateAnnotations(daemonSet *appsV1.DaemonSet, sensorSpec *cbContainersV1.CBContainersRuntimeSensorSpec) {
	if daemonSet.ObjectMeta.Annotations == nil {
		daemonSet.ObjectMeta.Annotations = make(map[string]string)
	}

	applyment.EnforceMapContains(daemonSet.ObjectMeta.Annotations, sensorSpec.DeploymentAnnotations)

	if daemonSet.Spec.Template.ObjectMeta.Annotations == nil {
		daemonSet.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	applyment.EnforceMapContains(daemonSet.Spec.Template.ObjectMeta.Annotations, map[string]string{
		"prometheus.io/scrape": fmt.Sprint(*sensorSpec.Prometheus.Enabled),
		"prometheus.io/port":   fmt.Sprint(sensorSpec.Prometheus.Port),
	})
	applyment.EnforceMapContains(daemonSet.Spec.Template.ObjectMeta.Annotations, sensorSpec.PodTemplateAnnotations)
}

func (obj *SensorDaemonSetK8sObject) mutateContainersList(
	templatePodSpec *coreV1.PodSpec,
	sensorSpec *cbContainersV1.CBContainersRuntimeSensorSpec,
	version,
	accessTokenSecretName string,
	desiredGRPCPortValue int32) {

	if len(templatePodSpec.Containers) != 1 {
		container := coreV1.Container{}
		templatePodSpec.Containers = []coreV1.Container{container}
	}

	obj.mutateContainer(&templatePodSpec.Containers[0], sensorSpec, version, accessTokenSecretName, desiredGRPCPortValue)
}

func (obj *SensorDaemonSetK8sObject) mutateContainer(
	container *coreV1.Container,
	sensorSpec *cbContainersV1.CBContainersRuntimeSensorSpec,
	version,
	accessTokenSecretName string,
	desiredGRPCPortValue int32) {

	container.Name = sensorName
	container.Resources = sensorSpec.Resources
	container.Args = []string{sensorVerbosityFlag, fmt.Sprintf("%d", sensorSpec.VerbosityLevel)}
	container.Command = []string{sensorRunCommand}
	commonState.MutateImage(container, sensorSpec.Image, version)
	commonState.MutateContainerFileProbes(container, sensorSpec.Probes)
	obj.mutateEnvVars(container, sensorSpec, accessTokenSecretName, desiredGRPCPortValue)
	obj.mutateContainerPorts(container, desiredGRPCPortValue)
	obj.mutateSecurityContext(container)
}

func (obj *SensorDaemonSetK8sObject) mutateContainerPorts(container *coreV1.Container, desiredGRPCPortValue int32) {
	if container.Ports == nil || len(container.Ports) != 1 {
		container.Ports = []coreV1.ContainerPort{{}}
	}

	container.Ports[0].Name = desiredDeploymentGRPCPortName
	container.Ports[0].ContainerPort = desiredGRPCPortValue
	container.Ports[0].HostPort = desiredGRPCPortValue

}

func (obj *SensorDaemonSetK8sObject) mutateEnvVars(
	container *coreV1.Container,
	sensorSpec *cbContainersV1.CBContainersRuntimeSensorSpec,
	accessTokenSecretName string,
	desiredGRPCPortValue int32) {

	commonState.MutateEnvVars(container, sensorSpec.Env, accessTokenSecretName, nil,
		coreV1.EnvVar{Name: "RUNTIME_KUBERNETES_SENSOR_GRPC_PORT", Value: fmt.Sprintf("%d", desiredGRPCPortValue)},
		coreV1.EnvVar{Name: "RUNTIME_KUBERNETES_SENSOR_RESOLVER_ADDRESS", Value: resolverAddress},
		coreV1.EnvVar{Name: "RUNTIME_KUBERNETES_SENSOR_RESOLVER_CONNECTION_TIMEOUT_SECONDS", Value: fmt.Sprintf("%d", desiredConnectionTimeoutSeconds)},
		coreV1.EnvVar{Name: "RUNTIME_KUBERNETES_SENSOR_LIVENESS_PATH", Value: sensorSpec.Probes.LivenessPath},
		coreV1.EnvVar{Name: "RUNTIME_KUBERNETES_SENSOR_READINESS_PATH", Value: sensorSpec.Probes.ReadinessPath},
	)
}

func (obj *SensorDaemonSetK8sObject) mutateSecurityContext(container *coreV1.Container) {
	if container.SecurityContext == nil {
		container.SecurityContext = &coreV1.SecurityContext{}
	}

	container.SecurityContext.Privileged = &sensorIsPrivileged
	container.SecurityContext.RunAsUser = &sensorRunAsUser
}
