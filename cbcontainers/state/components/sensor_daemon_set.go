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
	DaemonSetName       = "cbcontainers-node-agent"
	RuntimeSensorName   = "cbcontainers-runtime"
	ClusterScanningName = "cbcontainers-cluster-scanner"
	daemonSetLabelKey   = "app.kubernetes.io/name"

	runtimeSensorVerbosityFlag = "-v"
	runtimeSensorRunCommand    = "/run_sensor.sh"
	dnsPolicyDefault = coreV1.DNSDefault
	runtimeSensorDNSPolicy   = coreV1.DNSClusterFirstWithHostNet
	runtimeSensorHostNetwork = true
	runtimeSensorHostPID     = true

	desiredConnectionTimeoutSeconds = 60
	containerdRuntimeEndpoint       = "/var/run/containerd/containerd.sock"
	dockerRuntimeEndpoint           = "/var/run/dockershim.sock"
	crioRuntimeEndpoint             = "/var/run/crio/crio.sock"
)

var (
	sensorIsPrivileged       = true
	sensorRunAsUser    int64 = 0

	resolverAddress            = fmt.Sprintf("%s.%s.svc.cluster.local", ResolverName, commonState.DataPlaneNamespaceName)
	supportedContainerRuntimes = map[string]string{
		"containerd": containerdRuntimeEndpoint,
		"docker":     dockerRuntimeEndpoint,
		"crio":       crioRuntimeEndpoint,
	}
)

type SensorDaemonSetK8sObject struct{}

func NewSensorDaemonSetK8sObject() *SensorDaemonSetK8sObject {
	return &SensorDaemonSetK8sObject{}
}

func (obj *SensorDaemonSetK8sObject) EmptyK8sObject() client.Object {
	return &appsV1.DaemonSet{}
}

func (obj *SensorDaemonSetK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: DaemonSetName, Namespace: commonState.DataPlaneNamespaceName}
}

func (obj *SensorDaemonSetK8sObject) MutateK8sObject(k8sObject client.Object, agentSpec *cbContainersV1.CBContainersAgentSpec) error {
	daemonSet, ok := k8sObject.(*appsV1.DaemonSet)
	if !ok {
		return fmt.Errorf("expected DaemonSet K8s object")
	}

	runtimeProtection := &agentSpec.Components.RuntimeProtection

	obj.initiateDamonSet(daemonSet)

	if commonState.IsEnabled(runtimeProtection.Enabled) {
		daemonSet.Spec.Template.Spec.DNSPolicy = runtimeSensorDNSPolicy
		daemonSet.Spec.Template.Spec.HostNetwork = runtimeSensorHostNetwork
		daemonSet.Spec.Template.Spec.HostPID = runtimeSensorHostPID
	} else {
		// disable runtime special requirements that cluster-scanning doesn't need.
		// in case the cluster scanner was enabled after the runtime time was disabled (the values exists in the ds)
		daemonSet.Spec.Template.Spec.DNSPolicy = dnsPolicyDefault
		daemonSet.Spec.Template.Spec.HostNetwork = false
		daemonSet.Spec.Template.Spec.HostPID = false
	}

	obj.mutateLabels(daemonSet, agentSpec)
	obj.mutateAnnotations(daemonSet, agentSpec)
	obj.mutateVolumes(daemonSet, agentSpec)
	obj.mutateContainersList(&daemonSet.Spec.Template.Spec,
		agentSpec,
		agentSpec.Version,
		agentSpec.AccessTokenSecretName,
		runtimeProtection.InternalGrpcPort,
	)

	return nil
}

func (obj *SensorDaemonSetK8sObject) initiateDamonSet(daemonSet *appsV1.DaemonSet) {
	if daemonSet.Spec.Selector == nil {
		daemonSet.Spec.Selector = &metav1.LabelSelector{}
	}

	if daemonSet.ObjectMeta.Annotations == nil {
		daemonSet.ObjectMeta.Annotations = make(map[string]string)
	}

	if daemonSet.Spec.Template.ObjectMeta.Annotations == nil {
		daemonSet.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	daemonSet.Spec.Template.Spec.ServiceAccountName = commonState.DataPlaneServiceAccountName
	daemonSet.Spec.Template.Spec.PriorityClassName = commonState.DataPlanePriorityClassName
	daemonSet.Spec.Template.Spec.ImagePullSecrets = []coreV1.LocalObjectReference{{Name: commonState.RegistrySecretName}}
}

func (obj *SensorDaemonSetK8sObject) mutateLabels(daemonSet *appsV1.DaemonSet, agentSpec *cbContainersV1.CBContainersAgentSpec) {
	desiredLabels := make(map[string]string)
	desiredLabels[daemonSetLabelKey] = DaemonSetName
	if commonState.IsEnabled(agentSpec.Components.RuntimeProtection.Enabled) {
		applyment.EnforceMapContains(desiredLabels, agentSpec.Components.RuntimeProtection.Sensor.Labels)
	}

	if commonState.IsEnabled(agentSpec.Components.ClusterScanning.Enabled) {
		applyment.EnforceMapContains(desiredLabels, agentSpec.Components.ClusterScanning.ClusterScanningSensor.Labels)
	}

	daemonSet.ObjectMeta.Labels = desiredLabels
	daemonSet.Spec.Selector.MatchLabels = desiredLabels
	daemonSet.Spec.Template.ObjectMeta.Labels = desiredLabels
}

func (obj *SensorDaemonSetK8sObject) mutateAnnotations(daemonSet *appsV1.DaemonSet, agentSpec *cbContainersV1.CBContainersAgentSpec) {
	var prometheusEnabled bool
	var prometheusPort int

	if commonState.IsEnabled(agentSpec.Components.RuntimeProtection.Enabled) {
		runtimeSensor := agentSpec.Components.RuntimeProtection.Sensor
		applyment.EnforceMapContains(daemonSet.ObjectMeta.Annotations, runtimeSensor.DaemonSetAnnotations)
		applyment.EnforceMapContains(daemonSet.Spec.Template.ObjectMeta.Annotations, runtimeSensor.PodTemplateAnnotations)
		prometheusEnabled = *runtimeSensor.Prometheus.Enabled
		prometheusPort = runtimeSensor.Prometheus.Port
	}

	if commonState.IsEnabled(agentSpec.Components.ClusterScanning.Enabled) {
		clusterScanner := agentSpec.Components.ClusterScanning.ClusterScanningSensor
		applyment.EnforceMapContains(daemonSet.ObjectMeta.Annotations, clusterScanner.DaemonSetAnnotations)
		applyment.EnforceMapContains(daemonSet.Spec.Template.ObjectMeta.Annotations, clusterScanner.PodTemplateAnnotations)
		if commonState.IsDisabled(agentSpec.Components.RuntimeProtection.Enabled) {
			prometheusEnabled = *clusterScanner.Prometheus.Enabled
			prometheusPort = clusterScanner.Prometheus.Port
		}
	}


	applyment.EnforceMapContains(daemonSet.Spec.Template.ObjectMeta.Annotations, map[string]string{
		"prometheus.io/scrape": fmt.Sprint(prometheusEnabled),
		"prometheus.io/port":   fmt.Sprint(prometheusPort),
	})
}

func (obj *SensorDaemonSetK8sObject) mutateVolumes(daemonSet *appsV1.DaemonSet, agentSpec *cbContainersV1.CBContainersAgentSpec) {
	if commonState.IsEnabled(agentSpec.Components.ClusterScanning.Enabled) {
		obj.mutateClusterScannerVolumes(&daemonSet.Spec.Template.Spec)
	} else {
		// clean cluster-scanner volumes
		daemonSet.Spec.Template.Spec.Volumes = nil
	}
}

func (obj *SensorDaemonSetK8sObject) mutateContainersList(
	templatePodSpec *coreV1.PodSpec,
	agentSpec *cbContainersV1.CBContainersAgentSpec,
	version,
	accessTokenSecretName string,
	desiredGRPCPortValue int32) {
	containers := make([]coreV1.Container, 0, 2)

	if commonState.IsEnabled(agentSpec.Components.RuntimeProtection.Enabled) {
		containers = append(containers, coreV1.Container{Name: RuntimeSensorName})
	}

	if commonState.IsEnabled(agentSpec.Components.ClusterScanning.Enabled) {
		containers = append(containers, coreV1.Container{Name: ClusterScanningName})
	}

	if len(templatePodSpec.Containers) != len(containers) || obj.componentsWereSwitched(agentSpec, templatePodSpec, containers) {
		templatePodSpec.Containers = containers
	}

	if commonState.IsEnabled(agentSpec.Components.RuntimeProtection.Enabled) {
		obj.mutateRuntimeContainer(
			&templatePodSpec.Containers[obj.findContainerLocationByName(templatePodSpec.Containers, RuntimeSensorName)],
			&agentSpec.Components.RuntimeProtection.Sensor, version, desiredGRPCPortValue)
	}

	if commonState.IsEnabled(agentSpec.Components.ClusterScanning.Enabled) {
		obj.mutateClusterScannerContainer(
			&templatePodSpec.Containers[obj.findContainerLocationByName(templatePodSpec.Containers, ClusterScanningName)],
			&agentSpec.Components.ClusterScanning.ClusterScanningSensor, version, accessTokenSecretName,
			&agentSpec.Gateways.HardeningEventsGateway)
	}
}

// In case one enabled feature was off, and the other was on, and they were switched (first is now on, and second off),
// the containers count is the same (1), but the container name is wrong for the mutate containers methods, which will lead to
// index out of range error in finding the container by name.
// componentsWereSwitched checks if this scenario happened, and return boolean answer.
func (obj *SensorDaemonSetK8sObject) componentsWereSwitched(agentSpec *cbContainersV1.CBContainersAgentSpec, templatePodSpec *coreV1.PodSpec, containers []coreV1.Container) bool {
	if len(templatePodSpec.Containers) == 1 {
		if commonState.IsEnabled(agentSpec.Components.RuntimeProtection.Enabled) {
			return templatePodSpec.Containers[0].Name != RuntimeSensorName
		}

		if commonState.IsEnabled(agentSpec.Components.ClusterScanning.Enabled) {
			return templatePodSpec.Containers[0].Name != ClusterScanningName
		}
	}

	return false
}

func (obj *SensorDaemonSetK8sObject) findContainerLocationByName(containers []coreV1.Container, name string) int {
	for location, container := range containers {
		if container.Name == name {
			return location
		}
	}

	return -1
}

func (obj *SensorDaemonSetK8sObject) mutateRuntimeContainer(
	container *coreV1.Container,
	sensorSpec *cbContainersV1.CBContainersRuntimeSensorSpec,
	version string,
	desiredGRPCPortValue int32) {

	container.Name = RuntimeSensorName
	container.Resources = sensorSpec.Resources
	container.Args = []string{runtimeSensorVerbosityFlag, fmt.Sprintf("%d", *sensorSpec.VerbosityLevel)}
	container.Command = []string{runtimeSensorRunCommand}
	commonState.MutateImage(container, sensorSpec.Image, version)
	commonState.MutateContainerFileProbes(container, sensorSpec.Probes)
	obj.mutateRuntimeEnvVars(container, sensorSpec, desiredGRPCPortValue)
	obj.mutateSecurityContext(container)
}

func (obj *SensorDaemonSetK8sObject) mutateRuntimeEnvVars(
	container *coreV1.Container,
	sensorSpec *cbContainersV1.CBContainersRuntimeSensorSpec,
	desiredGRPCPortValue int32) {
	customEnvs := []coreV1.EnvVar{
		{Name: "RUNTIME_KUBERNETES_SENSOR_GRPC_PORT", Value: fmt.Sprintf("%d", desiredGRPCPortValue)},
		{Name: "RUNTIME_KUBERNETES_SENSOR_RESOLVER_ADDRESS", Value: resolverAddress},
		{Name: "RUNTIME_KUBERNETES_SENSOR_RESOLVER_CONNECTION_TIMEOUT_SECONDS", Value: fmt.Sprintf("%d", desiredConnectionTimeoutSeconds)},
		{Name: "RUNTIME_KUBERNETES_SENSOR_LIVENESS_PATH", Value: sensorSpec.Probes.LivenessPath},
		{Name: "RUNTIME_KUBERNETES_SENSOR_READINESS_PATH", Value: sensorSpec.Probes.ReadinessPath},
	}

	envVarBuilder := commonState.NewEnvVarBuilder().
		WithCustom(customEnvs...).
		WithSpec(sensorSpec.Env)
	commonState.MutateEnvVars(container, envVarBuilder)
}

func (obj *SensorDaemonSetK8sObject) mutateSecurityContext(container *coreV1.Container) {
	if container.SecurityContext == nil {
		container.SecurityContext = &coreV1.SecurityContext{}
	}

	container.SecurityContext.Privileged = &sensorIsPrivileged
	container.SecurityContext.RunAsUser = &sensorRunAsUser
}

func (obj *SensorDaemonSetK8sObject) mutateClusterScannerContainer(
	container *coreV1.Container,
	clusterScannerSpec *cbContainersV1.CBContainersClusterScannerAgentSpec,
	version string,
	accessTokenSecretName string,
	eventsGatewaySpec *cbContainersV1.CBContainersEventsGatewaySpec) {
	container.Name = ClusterScanningName
	container.Resources = clusterScannerSpec.Resources
	commonState.MutateImage(container, clusterScannerSpec.Image, version)
	// todo - mutate file probes
	commonState.MutateContainerHTTPProbes(container, clusterScannerSpec.Probes)
	obj.mutateClusterScannerEnvVars(container, clusterScannerSpec, accessTokenSecretName, eventsGatewaySpec)
	obj.mutateClusterScannerVolumesMounts(container)
	obj.mutateSecurityContext(container)
}

func (obj *SensorDaemonSetK8sObject) mutateClusterScannerEnvVars(container *coreV1.Container,
	clusterScannerSpec *cbContainersV1.CBContainersClusterScannerAgentSpec,
	accessTokenSecretName string, eventsGatewaySpec *cbContainersV1.CBContainersEventsGatewaySpec) {
	customEnvs := []coreV1.EnvVar{
		{Name: "CLUSTER_SCANNER_PROMETHEUS_PORT", Value: fmt.Sprintf("%d", clusterScannerSpec.Prometheus.Port)},
		{Name: "CLUSTER_SCANNER_IMAGE_SCANNING_REPORTER_HOST", Value: ImageScanningReporterName},
		{Name: "CLUSTER_SCANNER_IMAGE_SCANNING_REPORTER_PORT", Value: fmt.Sprintf("%d", ImageScanningReporterDesiredContainerPortValue)},
		{Name: "CLUSTER_SCANNER_IMAGE_SCANNING_REPORTER_SCHEME", Value: ImageScanningReporterDesiredContainerPortName},
	}

	envVarBuilder := commonState.NewEnvVarBuilder().
		WithCommonDataPlane(accessTokenSecretName).
		WithEventsGateway(eventsGatewaySpec).
		WithCustom(customEnvs...).
		WithSpec(clusterScannerSpec.Env).
		WithGatewayTLS()
	commonState.MutateEnvVars(container, envVarBuilder)
}

func (obj *SensorDaemonSetK8sObject) mutateClusterScannerVolumes(templatePodSpec *coreV1.PodSpec) {
	if templatePodSpec.Volumes == nil || len(templatePodSpec.Volumes) != len(supportedContainerRuntimes) + 1 {
		templatePodSpec.Volumes = make([]coreV1.Volume, 0)
	}

	// mutate root-cas volume, for https certificates
	commonState.MutateVolumesToIncludeRootCAsVolume(templatePodSpec)

	// mutate container-runtimes unix sockets files for the cluster-scanner CRI
	for name, path := range supportedContainerRuntimes {
		routeIndex := commonState.EnsureAndGetVolumeIndexForName(templatePodSpec, name)
		if templatePodSpec.Volumes[routeIndex].HostPath == nil {
			templatePodSpec.Volumes[routeIndex].HostPath = &coreV1.HostPathVolumeSource{Path: path}
		}
	}
}

func (obj *SensorDaemonSetK8sObject) mutateClusterScannerVolumesMounts(container *coreV1.Container) {
	if container.VolumeMounts == nil || len(container.VolumeMounts) != len(supportedContainerRuntimes)+1 {
		container.VolumeMounts = make([]coreV1.VolumeMount, 0)
	}

	// mutate mount for root-cas volume, for https server certificates
	commonState.MutateVolumeMountToIncludeRootCAsVolumeMount(container)

	// mutate mount for container-runtimes unix sockets files for the cluster-scanner CRI
	for name, mountPath := range supportedContainerRuntimes {
		index := commonState.EnsureAndGetVolumeMountIndexForName(container, name)
		commonState.MutateVolumeMount(container, index, mountPath, true)
	}
}
