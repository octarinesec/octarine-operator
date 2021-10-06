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
	ImageScanningReporterName     = "cbcontainers-image-scanning-reporter"
	ImageScanningReporterLabelKey = "app.kubernetes.io/name"

	ImageScanningReporterDesiredContainerPortName  = "https"
	ImageScanningReporterDesiredContainerPortValue = 443

	desiredTlsSecretVolumeName          = "cert"
	desiredTlsSecretVolumeMountPath     = "/etc/octarine-certificates"
	desiredTlsSecretVolumeMountReadOnly = true
)

var (
	desiredTlsSecretVolumeDecimalDefaultMode int32 = 420 // 644 in octal
	desiredTlsSecretVolumeOptionalValue            = true

	ImageScanningReporterAllowPrivilegeEscalation       = false
	ImageScanningReporterReadOnlyRootFilesystem         = true
	ImageScanningReporterRunAsUser                int64 = 0
	ImageScanningReporterCapabilitiesToAdd              = []coreV1.Capability{"NET_BIND_SERVICE"}
	ImageScanningReporterCapabilitiesToDrop             = []coreV1.Capability{"ALL"}
)

type ImageScanningReporterDeploymentK8sObject struct{}

func NewImageScanningReporterDeploymentK8sObject() *ImageScanningReporterDeploymentK8sObject {
	return &ImageScanningReporterDeploymentK8sObject{}
}

func (obj *ImageScanningReporterDeploymentK8sObject) EmptyK8sObject() client.Object {
	return &appsV1.Deployment{}
}

func (obj *ImageScanningReporterDeploymentK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: ImageScanningReporterName, Namespace: commonState.DataPlaneNamespaceName}
}

func (obj *ImageScanningReporterDeploymentK8sObject) MutateK8sObject(k8sObject client.Object, agentSpec *cbcontainersv1.CBContainersAgentSpec) error {
	deployment, ok := k8sObject.(*appsV1.Deployment)
	if !ok {
		return fmt.Errorf("expected Deployment K8s object")
	}

	clusterScanning := &agentSpec.Components.ClusterScanning
	imageScanningReporter := &clusterScanning.ImageScanningReporter

	obj.initiateDeployment(deployment, imageScanningReporter)
	obj.mutateLabels(deployment, imageScanningReporter)
	obj.mutateAnnotations(deployment, imageScanningReporter)
	obj.mutateVolumes(&deployment.Spec.Template.Spec)
	obj.mutateContainersList(&deployment.Spec.Template.Spec, imageScanningReporter, &agentSpec.Gateways.HardeningEventsGateway, agentSpec.Version, agentSpec.AccessTokenSecretName)

	return nil
}

// initiateDeployment initiate the deployment attributes with empty or default values.
func (obj *ImageScanningReporterDeploymentK8sObject) initiateDeployment(deployment *appsV1.Deployment, imageScanningReporter *cbcontainersv1.CBContainersImageScanningReporterSpec) {
	if deployment.Spec.Selector == nil {
		deployment.Spec.Selector = &metav1.LabelSelector{}
	}

	if deployment.ObjectMeta.Annotations == nil {
		deployment.ObjectMeta.Annotations = make(map[string]string)
	}

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	deployment.Spec.Replicas = imageScanningReporter.ReplicasCount
	deployment.Spec.Template.Spec.ServiceAccountName = commonState.DataPlaneServiceAccountName
	deployment.Spec.Template.Spec.PriorityClassName = commonState.DataPlanePriorityClassName
	deployment.Spec.Template.Spec.ImagePullSecrets = []coreV1.LocalObjectReference{{Name: commonState.RegistrySecretName}}
}

func (obj *ImageScanningReporterDeploymentK8sObject) mutateLabels(deployment *appsV1.Deployment, imageScanningReporterSpec *cbcontainersv1.CBContainersImageScanningReporterSpec) {
	desiredLabels := imageScanningReporterSpec.Labels
	if desiredLabels == nil {
		desiredLabels = make(map[string]string)
	}

	desiredLabels[ImageScanningReporterLabelKey] = ImageScanningReporterName
	deployment.ObjectMeta.Labels = desiredLabels
	deployment.Spec.Selector.MatchLabels = desiredLabels
	deployment.Spec.Template.ObjectMeta.Labels = desiredLabels
}

func (obj *ImageScanningReporterDeploymentK8sObject) mutateAnnotations(deployment *appsV1.Deployment, imageScanningReporterSpec *cbcontainersv1.CBContainersImageScanningReporterSpec) {
	applyment.EnforceMapContains(deployment.ObjectMeta.Annotations, imageScanningReporterSpec.DeploymentAnnotations)

	applyment.EnforceMapContains(deployment.Spec.Template.ObjectMeta.Annotations, map[string]string{
		"prometheus.io/scrape": fmt.Sprint(*imageScanningReporterSpec.Prometheus.Enabled),
		"prometheus.io/port":   fmt.Sprint(imageScanningReporterSpec.Prometheus.Port),
	})
	applyment.EnforceMapContains(deployment.Spec.Template.ObjectMeta.Annotations, imageScanningReporterSpec.PodTemplateAnnotations)
}

func (obj *ImageScanningReporterDeploymentK8sObject) mutateVolumes(templatePodSpec *coreV1.PodSpec) {
	if templatePodSpec.Volumes == nil || len(templatePodSpec.Volumes) != 2 {
		templatePodSpec.Volumes = make([]coreV1.Volume, 0)
	}

	// mutate tls secret volume, for cp backend gateway
	tlsSecretVolumeIndex := commonState.EnsureAndGetVolumeIndexForName(templatePodSpec, desiredTlsSecretVolumeName)
	if templatePodSpec.Volumes[tlsSecretVolumeIndex].Secret == nil {
		templatePodSpec.Volumes[tlsSecretVolumeIndex].Secret = &coreV1.SecretVolumeSource{}
	}
	templatePodSpec.Volumes[tlsSecretVolumeIndex].Secret.SecretName = ImageScanningReporterTlsName
	templatePodSpec.Volumes[tlsSecretVolumeIndex].Secret.DefaultMode = &desiredTlsSecretVolumeDecimalDefaultMode
	templatePodSpec.Volumes[tlsSecretVolumeIndex].Secret.Optional = &desiredTlsSecretVolumeOptionalValue

	// mutate root-cas volume, for https server certificates
	commonState.MutateVolumesToIncludeRootCAsVolume(templatePodSpec)
}

func (obj *ImageScanningReporterDeploymentK8sObject) mutateContainersList(templatePodSpec *coreV1.PodSpec, imageScanningReporterSpec *cbcontainersv1.CBContainersImageScanningReporterSpec, eventsGatewaySpec *cbcontainersv1.CBContainersEventsGatewaySpec, version, accessTokenSecretName string) {
	if len(templatePodSpec.Containers) != 1 {
		container := coreV1.Container{}
		templatePodSpec.Containers = []coreV1.Container{container}
	}

	obj.mutateContainer(&templatePodSpec.Containers[0], imageScanningReporterSpec, eventsGatewaySpec, version, accessTokenSecretName)
}

func (obj *ImageScanningReporterDeploymentK8sObject) mutateContainer(container *coreV1.Container, imageScanningReporterSpec *cbcontainersv1.CBContainersImageScanningReporterSpec, eventsGatewaySpec *cbcontainersv1.CBContainersEventsGatewaySpec, version, accessTokenSecretName string) {
	container.Name = ImageScanningReporterName
	container.Resources = imageScanningReporterSpec.Resources
	obj.mutateImageScanningReporterEnvVars(container, imageScanningReporterSpec, accessTokenSecretName, eventsGatewaySpec)
	commonState.MutateImage(container, imageScanningReporterSpec.Image, version)
	// TODO - check if the http probes should be insert from the default values
	//commonState.MutateContainerHTTPProbes(container, imageScanningReporterSpec.Probes)
	obj.mutateSecurityContext(container)
	obj.mutateContainerPorts(container)
	obj.mutateVolumesMounts(container)
}

func (obj *ImageScanningReporterDeploymentK8sObject) mutateImageScanningReporterEnvVars(container *coreV1.Container, imageScanningReporterSpec *cbcontainersv1.CBContainersImageScanningReporterSpec, accessTokenSecretName string, eventsGatewaySpec *cbcontainersv1.CBContainersEventsGatewaySpec) {
	customEnvs := []coreV1.EnvVar{
		// TODO - check if this env vars are required for cluster-scanner-gateway
		//{Name: "IMAGE_SCANNING_REPORTER_KEY_FILE_PATH", Value: fmt.Sprintf("%s/key", desiredTlsSecretVolumeMountPath)},
		//{Name: "IMAGE_SCANNING_REPORTER_FILE_PATH", Value: fmt.Sprintf("%s/signed_cert", desiredTlsSecretVolumeMountPath)},
		{Name: "IMAGE_SCANNING_REPORTER_PROMETHEUS_PORT", Value: fmt.Sprintf("%d", imageScanningReporterSpec.Prometheus.Port)},
		{Name: "GIN_MODE", Value: "release"},
	}

	envVarBuilder := commonState.NewEnvVarBuilder().
		WithCommonDataPlane(accessTokenSecretName).
		WithEventsGateway(eventsGatewaySpec).
		WithCustom(customEnvs...).
		WithSpec(imageScanningReporterSpec.Env).
		WithGatewayTLS() // to run https server with octarine certificates
	commonState.MutateEnvVars(container, envVarBuilder)
}

func (obj *ImageScanningReporterDeploymentK8sObject) mutateSecurityContext(container *coreV1.Container) {
	if container.SecurityContext == nil {
		container.SecurityContext = &coreV1.SecurityContext{}
	}

	container.SecurityContext.AllowPrivilegeEscalation = &ImageScanningReporterAllowPrivilegeEscalation
	container.SecurityContext.ReadOnlyRootFilesystem = &ImageScanningReporterReadOnlyRootFilesystem
	container.SecurityContext.RunAsUser = &ImageScanningReporterRunAsUser
	container.SecurityContext.Capabilities = &coreV1.Capabilities{
		Add:  ImageScanningReporterCapabilitiesToAdd,
		Drop: ImageScanningReporterCapabilitiesToDrop,
	}
}

func (obj *ImageScanningReporterDeploymentK8sObject) mutateContainerPorts(container *coreV1.Container) {
	if container.Ports == nil || len(container.Ports) != 1 {
		container.Ports = []coreV1.ContainerPort{{}}
	}

	container.Ports[0].Name = ImageScanningReporterDesiredContainerPortName
	container.Ports[0].ContainerPort = ImageScanningReporterDesiredContainerPortValue
}

func (obj *ImageScanningReporterDeploymentK8sObject) mutateVolumesMounts(container *coreV1.Container) {
	if container.VolumeMounts == nil || len(container.VolumeMounts) != 2 {
		container.VolumeMounts = make([]coreV1.VolumeMount, 0)
	}

	// mutate mount for tls secret volume, for cp backend gateway
	tlsSecretVolumeMountIndex := commonState.EnsureAndGetVolumeMountIndexForName(container, desiredTlsSecretVolumeName)
	commonState.MutateVolumeMount(container, tlsSecretVolumeMountIndex, desiredTlsSecretVolumeMountPath, desiredTlsSecretVolumeMountReadOnly)

	// mutate mount for root-cas volume, for https server certificates
	commonState.MutateVolumeMountToIncludeRootCAsVolumeMount(container)
}
