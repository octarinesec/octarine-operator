package components

import (
	"fmt"
	"strconv"
	"strings"

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
	DaemonSetName                = "cbcontainers-node-agent"
	RuntimeContainerName         = "cbcontainers-runtime"
	ClusterScanningContainerName = "cbcontainers-cluster-scanner"
	CndrContainerName            = "cbcontainers-cndr"
	daemonSetLabelKey            = "app.kubernetes.io/name"

	runtimeSensorRunCommand  = "/run_sensor.sh"
	defaultDnsPolicy         = coreV1.DNSClusterFirst
	runtimeSensorDNSPolicy   = coreV1.DNSClusterFirstWithHostNet
	runtimeSensorHostNetwork = true
	runtimeSensorHostPID     = true

	desiredConnectionTimeoutSeconds = 60

	// k8s container runtime/container engine endpoints
	containerdRuntimeEndpoint         = "/var/run/containerd/containerd.sock"
	microk8sContainerdRuntimeEndpoint = "/var/snap/microk8s/common/run/containerd.sock"
	k3sContainerdRuntimeEndpoint      = "/run/k3s/containerd/containerd.sock"
	dockerRuntimeEndpoint             = "/var/run/dockershim.sock"
	dockerSock                        = "/var/run/docker.sock"
	crioRuntimeEndpoint               = "/var/run/crio/crio.sock"
	hostRootPath                      = "/var/opt/root"

	// configuredContainerRuntimeVolumeName is used when the customer has specified a non-standard runtime endpoint in the CRD
	// as this means we need a special volume+mount for this endpoint
	configuredContainerRuntimeVolumeName = "configured-container-runtime-endpoint"

	// CRI-O specific volumes since the socket is not enough to read image blobs from the host

	crioStorageVolumeName       = "crio-storage"
	crioStorageConfigVolumeName = "crio-storage-config"
	crioConfigVolumeName        = "crio-config"

	// Source: https://github.com/containers/storage/blob/main/docs/containers-storage.conf.5.md and https://github.com/cri-o/cri-o/blob/main/docs/crio.conf.5.md

	crioStorageDefaultPath       = "/var/lib/containers"
	crioStorageConfigDefaultPath = "/etc/containers/storage.conf"
	crioConfigDefaultPath        = "/etc/crio/crio.conf"

	cndrCompanyCodeVarName = "CB_COMPANY_CODES"
	cndrCompanyCodeKeyName = "companyCode"
)

var (
	sensorIsPrivileged       = true
	sensorRunAsUser    int64 = 0

	supportedContainerRuntimes = map[string]string{
		"containerd":          containerdRuntimeEndpoint,
		"microk8s-containerd": microk8sContainerdRuntimeEndpoint,
		"k3s-containerd":      k3sContainerdRuntimeEndpoint,
		"docker":              dockerRuntimeEndpoint,
		"crio":                crioRuntimeEndpoint,
		"dockersock":          dockerSock,
	}

	hostPathDirectory         = coreV1.HostPathDirectory
	hostPathDirectoryOrCreate = coreV1.HostPathDirectoryOrCreate
	hostPathFile              = coreV1.HostPathFile
	cndrHostPaths             = map[string]*coreV1.HostPathVolumeSource{
		"boot":        {Path: "/boot", Type: &hostPathDirectory},
		"cb-data-dir": {Path: "/var/opt/carbonblack", Type: &hostPathDirectoryOrCreate},
		"os-release":  {Path: "/etc/os-release", Type: &hostPathFile},
		"root":        {Path: "/", Type: &hostPathDirectory},
	}
	// Optional to have a different mount volume that the host path. If not exits the host path will be used.
	cndrVolumeMounts = map[string]string{
		"root": hostRootPath,
	}
	cndrReadOnlyMounts = map[string]struct{}{
		"root":       {},
		"boot":       {},
		"os-release": {},
	}
)

type SensorDaemonSetK8sObject struct {
	// Namespace is the Namespace in which the DaemonSet will be created.
	Namespace string
}

func NewSensorDaemonSetK8sObject(namespace string) *SensorDaemonSetK8sObject {
	return &SensorDaemonSetK8sObject{
		Namespace: namespace,
	}
}

func (obj *SensorDaemonSetK8sObject) EmptyK8sObject() client.Object {
	return &appsV1.DaemonSet{}
}

func (obj *SensorDaemonSetK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: DaemonSetName, Namespace: obj.Namespace}
}

func (obj *SensorDaemonSetK8sObject) MutateK8sObject(k8sObject client.Object, agentSpec *cbContainersV1.CBContainersAgentSpec) error {
	daemonSet, ok := k8sObject.(*appsV1.DaemonSet)
	if !ok {
		return fmt.Errorf("expected DaemonSet K8s object")
	}

	runtimeProtection := &agentSpec.Components.RuntimeProtection

	obj.initiateDaemonSet(daemonSet, agentSpec)

	if commonState.IsEnabled(runtimeProtection.Enabled) || isCndrEnbaled(agentSpec.Components.Cndr) {
		daemonSet.Spec.Template.Spec.DNSPolicy = runtimeSensorDNSPolicy
		daemonSet.Spec.Template.Spec.HostNetwork = runtimeSensorHostNetwork
		daemonSet.Spec.Template.Spec.HostPID = runtimeSensorHostPID
	} else {
		// disable runtime special requirements that cluster-scanning doesn't need.
		// in case the cluster scanner was enabled after the runtime time was disabled (the values exists in the ds)
		daemonSet.Spec.Template.Spec.DNSPolicy = defaultDnsPolicy
		daemonSet.Spec.Template.Spec.HostNetwork = false
		daemonSet.Spec.Template.Spec.HostPID = false
	}

	obj.mutateLabels(daemonSet, agentSpec)
	obj.mutateAnnotations(daemonSet, agentSpec)
	obj.mutateVolumes(daemonSet, agentSpec)
	obj.mutateTolerations(daemonSet, agentSpec)
	obj.mutateContainersList(daemonSet, agentSpec)
	commonState.NewNodeTermsBuilder(&daemonSet.Spec.Template.Spec).Build()

	return nil
}

func (obj *SensorDaemonSetK8sObject) initiateDaemonSet(daemonSet *appsV1.DaemonSet, agentSpec *cbContainersV1.CBContainersAgentSpec) {
	if daemonSet.Spec.Selector == nil {
		daemonSet.Spec.Selector = &metav1.LabelSelector{}
	}

	if daemonSet.ObjectMeta.Annotations == nil {
		daemonSet.ObjectMeta.Annotations = make(map[string]string)
	}

	if daemonSet.Spec.Template.ObjectMeta.Annotations == nil {
		daemonSet.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	daemonSet.Spec.Template.Spec.ServiceAccountName = commonState.AgentNodeServiceAccountName
	daemonSet.Spec.Template.Spec.PriorityClassName = commonState.DataPlanePriorityClassName
	desiredImagePullSecrets := getImagePullSecrets(agentSpec, agentSpec.Components.RuntimeProtection.Sensor.Image.PullSecrets...)
	if objectsDiffer(desiredImagePullSecrets, daemonSet.Spec.Template.Spec.ImagePullSecrets) {
		daemonSet.Spec.Template.Spec.ImagePullSecrets = getImagePullSecrets(agentSpec, agentSpec.Components.RuntimeProtection.Sensor.Image.PullSecrets...)
	}
}

func (obj *SensorDaemonSetK8sObject) mutateLabels(daemonSet *appsV1.DaemonSet, agentSpec *cbContainersV1.CBContainersAgentSpec) {
	desiredLabels := make(map[string]string)
	desiredLabels[daemonSetLabelKey] = DaemonSetName
	if commonState.IsEnabled(agentSpec.Components.RuntimeProtection.Enabled) {
		applyment.EnforceMapContains(desiredLabels, agentSpec.Components.RuntimeProtection.Sensor.Labels)
	}

	if commonState.IsEnabled(agentSpec.Components.ClusterScanning.Enabled) {
		applyment.EnforceMapContains(desiredLabels, agentSpec.Components.ClusterScanning.ClusterScannerAgent.Labels)
	}

	daemonSet.ObjectMeta.Labels = desiredLabels
	daemonSet.Spec.Selector.MatchLabels = desiredLabels
	daemonSet.Spec.Template.ObjectMeta.Labels = desiredLabels
}

func (obj *SensorDaemonSetK8sObject) mutateAnnotations(daemonSet *appsV1.DaemonSet, agentSpec *cbContainersV1.CBContainersAgentSpec) {
	prometheusEnabled := false

	if commonState.IsEnabled(agentSpec.Components.RuntimeProtection.Enabled) {
		runtimeSensor := agentSpec.Components.RuntimeProtection.Sensor
		applyment.EnforceMapContains(daemonSet.ObjectMeta.Annotations, runtimeSensor.DaemonSetAnnotations)
		applyment.EnforceMapContains(daemonSet.Spec.Template.ObjectMeta.Annotations, runtimeSensor.PodTemplateAnnotations)
		if *runtimeSensor.Prometheus.Enabled {
			prometheusEnabled = true
		}

	}

	if commonState.IsEnabled(agentSpec.Components.ClusterScanning.Enabled) {
		clusterScanner := agentSpec.Components.ClusterScanning.ClusterScannerAgent
		applyment.EnforceMapContains(daemonSet.ObjectMeta.Annotations, clusterScanner.DaemonSetAnnotations)
		applyment.EnforceMapContains(daemonSet.Spec.Template.ObjectMeta.Annotations, clusterScanner.PodTemplateAnnotations)
		if *clusterScanner.Prometheus.Enabled {
			prometheusEnabled = true
		}
	}

	if prometheusEnabled {
		applyment.EnforceMapContains(daemonSet.Spec.Template.ObjectMeta.Annotations, map[string]string{
			"prometheus.io/scrape": fmt.Sprint(true),
		})
	}
}

func isCndrEnbaled(cndrSpec *cbContainersV1.CBContainersCndrSpec) bool {
	return cndrSpec != nil && commonState.IsEnabled(cndrSpec.Enabled)
}

func (obj *SensorDaemonSetK8sObject) getExpectedVolumeCount(agentSpec *cbContainersV1.CBContainersAgentSpec) int {
	expectedVolumesCount := 0

	if commonState.IsEnabled(agentSpec.Components.ClusterScanning.Enabled) || isCndrEnbaled(agentSpec.Components.Cndr) {
		expectedVolumesCount += len(supportedContainerRuntimes)
	}

	if commonState.IsEnabled(agentSpec.Components.ClusterScanning.Enabled) {
		clusterScannerSpec := &agentSpec.Components.ClusterScanning.ClusterScannerAgent
		expectedVolumesCount += 1 // RootCA
		if clusterScannerSpec.K8sContainerEngine.Endpoint != "" {
			expectedVolumesCount += 1
		}

		// cri-o specific mounts to use the containers/storage library with the host's storage area
		// one mount is for the image store on the host, one is for the storage.conf and one is for crio.conf => we expect 3 more
		expectedVolumesCount += 3
	}

	if isCndrEnbaled(agentSpec.Components.Cndr) {
		expectedVolumesCount += len(cndrHostPaths)
	}

	return expectedVolumesCount
}

func (obj *SensorDaemonSetK8sObject) mutateVolumes(daemonSet *appsV1.DaemonSet, agentSpec *cbContainersV1.CBContainersAgentSpec) {
	templatePodSpec := &daemonSet.Spec.Template.Spec

	expectedVolumeCount := obj.getExpectedVolumeCount(agentSpec)
	if templatePodSpec.Volumes == nil || len(templatePodSpec.Volumes) != expectedVolumeCount {
		// clean cluster-scanner & cndr volumes
		templatePodSpec.Volumes = make([]coreV1.Volume, 0, expectedVolumeCount)
	}

	if commonState.IsEnabled(agentSpec.Components.ClusterScanning.Enabled) || isCndrEnbaled(agentSpec.Components.Cndr) {
		obj.mutateContainerRuntimesVolumes(&daemonSet.Spec.Template.Spec)
	}

	if commonState.IsEnabled(agentSpec.Components.ClusterScanning.Enabled) {
		obj.mutateClusterScannerVolumes(&daemonSet.Spec.Template.Spec, &agentSpec.Components.ClusterScanning.ClusterScannerAgent)
	}

	if isCndrEnbaled(agentSpec.Components.Cndr) {
		obj.mutateCndrVolumes(&daemonSet.Spec.Template.Spec, &agentSpec.Components.Cndr.Sensor)
	}
}

func (obj *SensorDaemonSetK8sObject) mutateTolerations(daemonSet *appsV1.DaemonSet, agentSpec *cbContainersV1.CBContainersAgentSpec) {
	daemonSet.Spec.Template.Spec.Tolerations = agentSpec.Components.Settings.DaemonSetsTolerations
}

func (obj *SensorDaemonSetK8sObject) mutateContainersList(daemonSet *appsV1.DaemonSet, agentSpec *cbContainersV1.CBContainersAgentSpec) {

	var runtimeContainer coreV1.Container
	var clusterScannerContainer coreV1.Container
	var cndrContainer coreV1.Container

	templatePodSpec := &daemonSet.Spec.Template.Spec

	desiredContainers := make([]coreV1.Container, 0, 2)
	runtimeEnabled := false
	clusterScannerEnabled := false
	cndrEnabled := false
	runtimeMissing := false
	clusterScannerMissing := false
	cndrMissing := false

	if commonState.IsEnabled(agentSpec.Components.RuntimeProtection.Enabled) {
		runtimeEnabled = true
		if runtimeContainerLocation := obj.findContainerLocationByName(templatePodSpec.Containers, RuntimeContainerName); runtimeContainerLocation == -1 {
			runtimeMissing = true
			runtimeContainer = coreV1.Container{Name: RuntimeContainerName}
		} else {
			runtimeContainer = templatePodSpec.Containers[runtimeContainerLocation]
		}

		desiredContainers = append(desiredContainers, runtimeContainer)
	}

	if commonState.IsEnabled(agentSpec.Components.ClusterScanning.Enabled) {
		clusterScannerEnabled = true
		if clusterScannerContainerLocation := obj.findContainerLocationByName(templatePodSpec.Containers, ClusterScanningContainerName); clusterScannerContainerLocation == -1 {
			clusterScannerMissing = true
			clusterScannerContainer = coreV1.Container{Name: ClusterScanningContainerName}
		} else {
			clusterScannerContainer = templatePodSpec.Containers[clusterScannerContainerLocation]
		}

		desiredContainers = append(desiredContainers, clusterScannerContainer)
	}

	if isCndrEnbaled(agentSpec.Components.Cndr) {
		cndrEnabled = true
		if cndrContainerLocation := obj.findContainerLocationByName(templatePodSpec.Containers, CndrContainerName); cndrContainerLocation == -1 {
			cndrMissing = true
			cndrContainer = coreV1.Container{Name: CndrContainerName}
		} else {
			cndrContainer = templatePodSpec.Containers[cndrContainerLocation]
		}

		desiredContainers = append(desiredContainers, cndrContainer)
	}

	if obj.isStateChanged(len(templatePodSpec.Containers), len(desiredContainers), runtimeEnabled, clusterScannerEnabled, runtimeMissing, clusterScannerMissing, cndrEnabled, cndrMissing) {
		templatePodSpec.Containers = desiredContainers
	}

	if commonState.IsEnabled(agentSpec.Components.RuntimeProtection.Enabled) {
		obj.mutateRuntimeContainer(
			&templatePodSpec.Containers[obj.findContainerLocationByName(templatePodSpec.Containers, RuntimeContainerName)],
			agentSpec)
	}

	if commonState.IsEnabled(agentSpec.Components.ClusterScanning.Enabled) {
		obj.mutateClusterScannerContainer(
			&templatePodSpec.Containers[obj.findContainerLocationByName(templatePodSpec.Containers, ClusterScanningContainerName)],
			agentSpec)
	}

	if isCndrEnbaled(agentSpec.Components.Cndr) {
		obj.mutateCndrContainer(
			&templatePodSpec.Containers[obj.findContainerLocationByName(templatePodSpec.Containers, CndrContainerName)],
			agentSpec)
	}
}

func (obj *SensorDaemonSetK8sObject) isStateChanged(actualContainersLength, desiredContainersLength int, runtimeEnabled, clusterScannerEnabled, runtimeMissing, clusterScannerMissing, cndrEnabled, cndrMissing bool) bool {
	// actual containers' length is different from desired containers' length.
	// test cases
	// there are more containers than the 2 allowed
	// there 0 containers when at least one component should be enabled
	// there are different components amount then the desired count.
	if actualContainersLength != desiredContainersLength {
		return true
	}

	// runtime enabled and container missing is actual state
	if runtimeEnabled && runtimeMissing {
		return true
	}

	// cluster scanner enabled and container is missing in actual state
	if clusterScannerEnabled && clusterScannerMissing {
		return true
	}

	// cluster scanner enabled and container is missing in actual state
	if cndrEnabled && cndrMissing {
		return true
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

func (obj *SensorDaemonSetK8sObject) mutateRuntimeContainer(container *coreV1.Container, agentSpec *cbContainersV1.CBContainersAgentSpec) {
	sensorSpec := &agentSpec.Components.RuntimeProtection.Sensor

	container.Name = RuntimeContainerName
	container.Resources = sensorSpec.Resources
	container.Command = []string{runtimeSensorRunCommand}
	commonState.MutateImage(container, sensorSpec.Image, agentSpec.Version, agentSpec.Components.Settings.DefaultImagesRegistry)
	commonState.MutateContainerFileProbes(container, sensorSpec.Probes)
	if commonState.IsEnabled(sensorSpec.Prometheus.Enabled) {
		container.Ports = []coreV1.ContainerPort{{Name: "metrics", ContainerPort: int32(sensorSpec.Prometheus.Port)}}
	}
	obj.mutateRuntimeEnvVars(container, agentSpec)
	obj.mutateSecurityContext(container)
}

func (obj *SensorDaemonSetK8sObject) mutateRuntimeEnvVars(container *coreV1.Container, agentSpec *cbContainersV1.CBContainersAgentSpec) {

	sensorSpec := &agentSpec.Components.RuntimeProtection.Sensor
	runtimeProtection := &agentSpec.Components.RuntimeProtection
	desiredGRPCPortValue := runtimeProtection.InternalGrpcPort

	customEnvs := []coreV1.EnvVar{
		{Name: "RUNTIME_KUBERNETES_SENSOR_GRPC_PORT", Value: fmt.Sprintf("%d", desiredGRPCPortValue)},
		{Name: "RUNTIME_KUBERNETES_SENSOR_RESOLVER_ADDRESS", Value: obj.resolverAddress()},
		{Name: "RUNTIME_KUBERNETES_SENSOR_RESOLVER_CONNECTION_TIMEOUT_SECONDS", Value: fmt.Sprintf("%d", desiredConnectionTimeoutSeconds)},
		{Name: "RUNTIME_KUBERNETES_SENSOR_LIVENESS_PATH", Value: sensorSpec.Probes.LivenessPath},
		{Name: "RUNTIME_KUBERNETES_SENSOR_READINESS_PATH", Value: sensorSpec.Probes.ReadinessPath},
		{Name: "RUNTIME_KUBERNETES_SENSOR_LOG_LEVEL", Value: runtimeProtection.Sensor.LogLevel},
	}

	envVarBuilder := commonState.NewEnvVarBuilder().
		WithCustom(customEnvs...).
		WithSpec(sensorSpec.Env).
		WithProxySettings(agentSpec.Components.Settings.Proxy)
	commonState.MutateEnvVars(container, envVarBuilder)
}

func (obj *SensorDaemonSetK8sObject) resolverAddress() string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", ResolverName, obj.Namespace)
}

func (obj *SensorDaemonSetK8sObject) mutateSecurityContext(container *coreV1.Container) {
	if container.SecurityContext == nil {
		container.SecurityContext = &coreV1.SecurityContext{}
	}

	container.SecurityContext.Privileged = &sensorIsPrivileged
	container.SecurityContext.RunAsUser = &sensorRunAsUser
}

func (obj *SensorDaemonSetK8sObject) mutateClusterScannerContainer(container *coreV1.Container, agentSpec *cbContainersV1.CBContainersAgentSpec) {

	clusterScannerSpec := &agentSpec.Components.ClusterScanning.ClusterScannerAgent

	container.Name = ClusterScanningContainerName
	container.Resources = clusterScannerSpec.Resources
	commonState.MutateImage(container, clusterScannerSpec.Image, agentSpec.Version, agentSpec.Components.Settings.DefaultImagesRegistry)
	commonState.MutateContainerFileProbes(container, clusterScannerSpec.Probes)
	if commonState.IsEnabled(clusterScannerSpec.Prometheus.Enabled) {
		container.Ports = []coreV1.ContainerPort{{Name: "metrics", ContainerPort: int32(clusterScannerSpec.Prometheus.Port)}}
	}
	obj.mutateClusterScannerEnvVars(container, agentSpec)
	obj.mutateClusterScannerVolumesMounts(container, agentSpec)
	obj.mutateSecurityContext(container)
}

func (obj *SensorDaemonSetK8sObject) mutateCndrContainer(container *coreV1.Container, agentSpec *cbContainersV1.CBContainersAgentSpec) {

	cndrSpec := &agentSpec.Components.Cndr.Sensor

	container.Name = CndrContainerName
	container.Resources = cndrSpec.Resources
	commonState.MutateImage(container, cndrSpec.Image, agentSpec.Version, agentSpec.Components.Settings.DefaultImagesRegistry)
	if commonState.IsEnabled(cndrSpec.Prometheus.Enabled) {
		container.Ports = []coreV1.ContainerPort{{Name: "metrics", ContainerPort: int32(cndrSpec.Prometheus.Port)}}
	}
	obj.mutateCndrEnvVars(container, agentSpec)
	obj.mutateCndrVolumesMounts(container, agentSpec)
	obj.mutateSecurityContext(container)
}

func (obj *SensorDaemonSetK8sObject) mutateCndrEnvVars(container *coreV1.Container, agentSpec *cbContainersV1.CBContainersAgentSpec) {
	cndrSpec := agentSpec.Components.Cndr

	customEnvs := []coreV1.EnvVar{
		{Name: "HOST_ROOT_PATH", Value: hostRootPath},
	}

	envVarBuilder := commonState.NewEnvVarBuilder().
		WithCommonDataPlane(agentSpec.AccessTokenSecretName).
		WithEventsGateway(&agentSpec.Gateways.HardeningEventsGateway).
		WithCustom(customEnvs...).
		WithEnvVarFromSecret(cndrCompanyCodeVarName, cndrSpec.CompanyCodeSecretName, cndrCompanyCodeKeyName).
		WithSpec(cndrSpec.Sensor.Env).
		WithProxySettings(agentSpec.Components.Settings.Proxy)

	commonState.MutateEnvVars(container, envVarBuilder)
}

func (obj *SensorDaemonSetK8sObject) mutateCndrVolumes(templatePodSpec *coreV1.PodSpec, cndrSendorSpec *cbContainersV1.CBContainersCndrSensorSpec) {
	// mutate host dirs required by the linux sensor
	for name, hostPath := range cndrHostPaths {
		routeIndex := commonState.EnsureAndGetVolumeIndexForName(templatePodSpec, name)
		templatePodSpec.Volumes[routeIndex].HostPath = hostPath
	}
	commonState.MutateVolumesToIncludeRootCAsVolume(templatePodSpec)
}

func (obj *SensorDaemonSetK8sObject) mutateCndrVolumesMounts(container *coreV1.Container, agentSpec *cbContainersV1.CBContainersAgentSpec) {
	expectedLength := len(cndrHostPaths) +
		len(supportedContainerRuntimes) +
		1 //root-ca

	if container.VolumeMounts == nil || len(container.VolumeMounts) != expectedLength {
		container.VolumeMounts = make([]coreV1.VolumeMount, 0)
	}

	// mutate mount for required host dirs by the linux sensor
	for name, hostPath := range cndrHostPaths {
		index := commonState.EnsureAndGetVolumeMountIndexForName(container, name)
		_, readOnly := cndrReadOnlyMounts[name]
		mountPath, ok := cndrVolumeMounts[name]
		if !ok {
			mountPath = hostPath.Path
		}
		commonState.MutateVolumeMount(container, index, mountPath, readOnly)
	}

	// mutate mount for container-runtimes unix sockets files for the container tracking processor
	for name, mountPath := range supportedContainerRuntimes {
		index := commonState.EnsureAndGetVolumeMountIndexForName(container, name)
		commonState.MutateVolumeMount(container, index, mountPath, true)
	}

	// mutate mount for root-cas volume, for https server certificates
	commonState.MutateVolumeMountToIncludeRootCAsVolumeMount(container)
}

func (obj *SensorDaemonSetK8sObject) mutateClusterScannerEnvVars(container *coreV1.Container, agentSpec *cbContainersV1.CBContainersAgentSpec) {
	clusterScannerSpec := &agentSpec.Components.ClusterScanning.ClusterScannerAgent

	customEnvs := []coreV1.EnvVar{
		{Name: "CLUSTER_SCANNER_PROMETHEUS_PORT", Value: fmt.Sprintf("%d", clusterScannerSpec.Prometheus.Port)},
		{Name: "CLUSTER_SCANNER_IMAGE_SCANNING_REPORTER_HOST", Value: obj.imageScanningReporterAddress()},
		{Name: "CLUSTER_SCANNER_IMAGE_SCANNING_REPORTER_PORT", Value: fmt.Sprintf("%d", ImageScanningReporterDesiredContainerPortValue)},
		{Name: "CLUSTER_SCANNER_IMAGE_SCANNING_REPORTER_SCHEME", Value: ImageScanningReporterDesiredContainerPortName},
		{Name: "CLUSTER_SCANNER_LIVENESS_PATH", Value: clusterScannerSpec.Probes.LivenessPath},
		{Name: "CLUSTER_SCANNER_READINESS_PATH", Value: clusterScannerSpec.Probes.ReadinessPath},
		{Name: "CLUSTER_SCANNER_CLI_FLAGS_SKIP_SECRETS_DETECTION", Value: strconv.FormatBool(clusterScannerSpec.CLIFlags.SkipSecretsDetection)},
		{Name: "CLUSTER_SCANNER_CLI_FLAGS_SKIP_DIRS_OR_FILES", Value: strings.Join(clusterScannerSpec.CLIFlags.SkipDirsOrFiles, ",")},
		{Name: "CLUSTER_SCANNER_CLI_FLAGS_SCAN_BASE_LAYER", Value: strconv.FormatBool(clusterScannerSpec.CLIFlags.ScanBaseLayer)},
		{Name: "CLUSTER_SCANNER_CLI_FLAGS_IGNORE_BUILT_IN_REGEX", Value: strconv.FormatBool(clusterScannerSpec.CLIFlags.IgnoreBuiltInRegex)},
	}

	if clusterScannerSpec.K8sContainerEngine.Endpoint != "" && clusterScannerSpec.K8sContainerEngine.EngineType != "" {
		customEnvs = append(customEnvs, coreV1.EnvVar{Name: "CLUSTER_SCANNER_ENDPOINT", Value: clusterScannerSpec.K8sContainerEngine.Endpoint})
		customEnvs = append(customEnvs, coreV1.EnvVar{Name: "CLUSTER_SCANNER_CONTAINER_RUNTIME", Value: clusterScannerSpec.K8sContainerEngine.EngineType.String()})
	}

	envVarBuilder := commonState.NewEnvVarBuilder().
		WithCommonDataPlane(agentSpec.AccessTokenSecretName).
		WithEventsGateway(&agentSpec.Gateways.HardeningEventsGateway).
		WithCustom(customEnvs...).
		WithEnvVarFromResource("CLUSTER_SCANNER_LIMITS_MEMORY", ClusterScanningContainerName, "limits.memory").
		WithEnvVarFromResource("CLUSTER_SCANNER_REQUESTS_MEMORY", ClusterScanningContainerName, "requests.memory").
		WithEnvVarFromField("CLUSTER_SCANNER_NODE_NAME", "spec.nodeName", "v1").
		WithSpec(clusterScannerSpec.Env).
		WithGatewayTLS().
		WithProxySettings(agentSpec.Components.Settings.Proxy)

	commonState.MutateEnvVars(container, envVarBuilder)
}

func (obj *SensorDaemonSetK8sObject) imageScanningReporterAddress() string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", ImageScanningReporterName, obj.Namespace)
}

func (obj *SensorDaemonSetK8sObject) mutateContainerRuntimesVolumes(templatePodSpec *coreV1.PodSpec) {
	// mutate container-runtimes unix sockets files for the cluster-scanner CRI
	for name, path := range supportedContainerRuntimes {
		routeIndex := commonState.EnsureAndGetVolumeIndexForName(templatePodSpec, name)
		if templatePodSpec.Volumes[routeIndex].HostPath == nil {
			templatePodSpec.Volumes[routeIndex].HostPath = &coreV1.HostPathVolumeSource{}
		}

		templatePodSpec.Volumes[routeIndex].HostPath.Path = path
	}
}

func (obj *SensorDaemonSetK8sObject) mutateClusterScannerVolumes(templatePodSpec *coreV1.PodSpec, clusterScannerSpec *cbContainersV1.CBContainersClusterScannerAgentSpec) {
	if clusterScannerSpec.K8sContainerEngine.Endpoint != "" {
		routeIndex := commonState.EnsureAndGetVolumeIndexForName(templatePodSpec, configuredContainerRuntimeVolumeName)
		if templatePodSpec.Volumes[routeIndex].HostPath == nil {
			templatePodSpec.Volumes[routeIndex].HostPath = &coreV1.HostPathVolumeSource{}
		}
		templatePodSpec.Volumes[routeIndex].HostPath.Path = clusterScannerSpec.K8sContainerEngine.Endpoint
	}

	// Ensure we have volumes for CRI-O
	crioStorageIx := commonState.EnsureAndGetVolumeIndexForName(templatePodSpec, crioStorageVolumeName)
	storagePath := clusterScannerSpec.K8sContainerEngine.CRIO.StoragePath
	if storagePath == "" {
		storagePath = crioStorageDefaultPath
	}
	if templatePodSpec.Volumes[crioStorageIx].HostPath == nil {
		templatePodSpec.Volumes[crioStorageIx].HostPath = &coreV1.HostPathVolumeSource{}
	}
	templatePodSpec.Volumes[crioStorageIx].HostPath.Path = storagePath

	crioStorageConfigIx := commonState.EnsureAndGetVolumeIndexForName(templatePodSpec, crioStorageConfigVolumeName)
	storageConfigPath := clusterScannerSpec.K8sContainerEngine.CRIO.StorageConfigPath
	if storageConfigPath == "" {
		storageConfigPath = crioStorageConfigDefaultPath
	}
	if templatePodSpec.Volumes[crioStorageConfigIx].HostPath == nil {
		templatePodSpec.Volumes[crioStorageConfigIx].HostPath = &coreV1.HostPathVolumeSource{}
	}
	templatePodSpec.Volumes[crioStorageConfigIx].HostPath.Path = storageConfigPath

	crioConfigIx := commonState.EnsureAndGetVolumeIndexForName(templatePodSpec, crioConfigVolumeName)
	configPath := clusterScannerSpec.K8sContainerEngine.CRIO.ConfigPath
	if configPath == "" {
		configPath = crioConfigDefaultPath
	}
	if templatePodSpec.Volumes[crioConfigIx].HostPath == nil {
		templatePodSpec.Volumes[crioConfigIx].HostPath = &coreV1.HostPathVolumeSource{}
	}
	templatePodSpec.Volumes[crioConfigIx].HostPath.Path = configPath
	// End CRI-O config

	// mutate root-cas volume, for https certificates
	commonState.MutateVolumesToIncludeRootCAsVolume(templatePodSpec)
}

func (obj *SensorDaemonSetK8sObject) mutateClusterScannerVolumesMounts(container *coreV1.Container, agentSpec *cbContainersV1.CBContainersAgentSpec) {
	containerRuntimes := getClusterScannerContainerRuntimes(&agentSpec.Components.ClusterScanning.ClusterScannerAgent)

	// We expect to see
	// container runtimes mounts + root CA mount + 3 mounts for CRI-O
	if container.VolumeMounts == nil || len(container.VolumeMounts) != (len(containerRuntimes)+1+3) {
		container.VolumeMounts = make([]coreV1.VolumeMount, 0)
	}

	// mutate mount for root-cas volume, for https server certificates
	commonState.MutateVolumeMountToIncludeRootCAsVolumeMount(container)

	// mutate mount for container-runtimes unix sockets files for the cluster-scanner CRI
	for name, mountPath := range containerRuntimes {
		index := commonState.EnsureAndGetVolumeMountIndexForName(container, name)
		commonState.MutateVolumeMount(container, index, mountPath, true)
	}

	// mutate mounts for the CRI-O engine storage
	crioStorageIx := commonState.EnsureAndGetVolumeMountIndexForName(container, crioStorageVolumeName)
	// The storage MUST be R/W as file-based locks are used to enable concurrent access to the same storage from both CRI-O and our agent
	commonState.MutateVolumeMount(container, crioStorageIx, crioStorageDefaultPath, false)
	crioStorageConfigIx := commonState.EnsureAndGetVolumeMountIndexForName(container, crioStorageConfigVolumeName)
	commonState.MutateVolumeMount(container, crioStorageConfigIx, crioStorageConfigDefaultPath, true)
	crioConfigIx := commonState.EnsureAndGetVolumeMountIndexForName(container, crioConfigVolumeName)
	commonState.MutateVolumeMount(container, crioConfigIx, crioConfigDefaultPath, true)
}

// The cluster scanner uses the container runtime unix socket to fetch the running containers to scan.
// getContainerRuntimes returns the unix paths to mount to the daemon set, so the cluster scanner container could access them.
// Returns supported container runtimes paths, and customer custom endpoint path if provided.
func getClusterScannerContainerRuntimes(clusterScannerSpec *cbContainersV1.CBContainersClusterScannerAgentSpec) map[string]string {
	containerRuntimes := make(map[string]string)
	for name, endpoint := range supportedContainerRuntimes {
		containerRuntimes[name] = endpoint
	}

	if clusterScannerSpec.K8sContainerEngine.Endpoint != "" {
		containerRuntimes[configuredContainerRuntimeVolumeName] = clusterScannerSpec.K8sContainerEngine.Endpoint
	}

	return containerRuntimes
}
