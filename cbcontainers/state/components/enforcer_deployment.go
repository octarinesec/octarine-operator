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
	EnforcerName = "cbcontainers-hardening-enforcer"

	DesiredContainerPortName  = "https"
	DesiredContainerPortValue = 8080

	DesiredTlsSecretVolumeName          = "cert"
	DesiredTlsSecretVolumeMountPath     = "/etc/octarine-certificates"
	DesiredTlsSecretVolumeMountReadOnly = true

	EnforcerLabelKey = "app.kubernetes.io/name"
)

var (
	DesiredTlsSecretVolumeDecimalDefaultMode int32 = 420 // 644 in octal
	DesiredTlsSecretVolumeOptionalValue            = true

	EnforcerAllowPrivilegeEscalation       = false
	EnforcerReadOnlyRootFilesystem         = true
	EnforcerRunAsUser                int64 = 1500
	EnforcerRunAsNonRoot                   = true
	EnforcerCapabilitiesToDrop             = []coreV1.Capability{"ALL"}
)

type EnforcerDeploymentK8sObject struct {
	// Namespace is the Namespace in which the Deployment will be created.
	Namespace string
}

func NewEnforcerDeploymentK8sObject(namespace string) *EnforcerDeploymentK8sObject {
	return &EnforcerDeploymentK8sObject{
		Namespace: namespace,
	}
}

func (obj *EnforcerDeploymentK8sObject) EmptyK8sObject() client.Object {
	return &appsV1.Deployment{}
}

func (obj *EnforcerDeploymentK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: EnforcerName, Namespace: obj.Namespace}
}

func (obj *EnforcerDeploymentK8sObject) MutateK8sObject(k8sObject client.Object, agentSpec *cbcontainersv1.CBContainersAgentSpec) error {
	enforcer := &agentSpec.Components.Basic.Enforcer

	deployment, ok := k8sObject.(*appsV1.Deployment)
	if !ok {
		return fmt.Errorf("expected Deployment K8s object")
	}

	desiredLabels := enforcer.Labels
	if desiredLabels == nil {
		desiredLabels = make(map[string]string)
	}
	desiredLabels[EnforcerLabelKey] = EnforcerName

	if deployment.Spec.Selector == nil {
		deployment.Spec.Selector = &metav1.LabelSelector{}
	}

	deployment.Spec.Replicas = enforcer.ReplicasCount
	deployment.ObjectMeta.Labels = desiredLabels
	deployment.Spec.Selector.MatchLabels = desiredLabels
	deployment.Spec.Template.ObjectMeta.Labels = desiredLabels
	deployment.Spec.Template.Spec.ServiceAccountName = commonState.EnforcerServiceAccountName
	deployment.Spec.Template.Spec.PriorityClassName = commonState.DataPlanePriorityClassName
	desiredImagePullSecrets := getImagePullSecrets(agentSpec, agentSpec.Components.Basic.Enforcer.Image.PullSecrets...)
	if objectsDiffer(deployment.Spec.Template.Spec.ImagePullSecrets, desiredImagePullSecrets) {
		deployment.Spec.Template.Spec.ImagePullSecrets = desiredImagePullSecrets
	}
	obj.mutateAnnotations(deployment, enforcer)
	obj.mutateVolumes(&deployment.Spec.Template.Spec)
	obj.mutateAffinityAndNodeSelector(&deployment.Spec.Template.Spec, enforcer)
	obj.mutateContainersList(&deployment.Spec.Template.Spec, agentSpec)
	commonState.NewNodeTermsBuilder(&deployment.Spec.Template.Spec).Build()

	return nil
}

func (obj *EnforcerDeploymentK8sObject) mutateAnnotations(deployment *appsV1.Deployment, enforcerSpec *cbcontainersv1.CBContainersEnforcerSpec) {
	if deployment.ObjectMeta.Annotations == nil {
		deployment.ObjectMeta.Annotations = make(map[string]string)
	}

	applyment.EnforceMapContains(deployment.ObjectMeta.Annotations, enforcerSpec.DeploymentAnnotations)

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	applyment.EnforceMapContains(deployment.Spec.Template.ObjectMeta.Annotations, map[string]string{
		"prometheus.io/scrape": fmt.Sprint(*enforcerSpec.Prometheus.Enabled),
		"prometheus.io/port":   fmt.Sprint(enforcerSpec.Prometheus.Port),
	})
	applyment.EnforceMapContains(deployment.Spec.Template.ObjectMeta.Annotations, enforcerSpec.PodTemplateAnnotations)
}

func (obj *EnforcerDeploymentK8sObject) mutateVolumes(templatePodSpec *coreV1.PodSpec) {
	if templatePodSpec.Volumes == nil || len(templatePodSpec.Volumes) != 2 {
		templatePodSpec.Volumes = make([]coreV1.Volume, 0)
	}

	tlsSecretVolumeIndex := commonState.EnsureAndGetVolumeIndexForName(templatePodSpec, DesiredTlsSecretVolumeName)
	if templatePodSpec.Volumes[tlsSecretVolumeIndex].Secret == nil {
		templatePodSpec.Volumes[tlsSecretVolumeIndex].Secret = &coreV1.SecretVolumeSource{}
	}
	templatePodSpec.Volumes[tlsSecretVolumeIndex].Secret.SecretName = EnforcerTlsName
	templatePodSpec.Volumes[tlsSecretVolumeIndex].Secret.DefaultMode = &DesiredTlsSecretVolumeDecimalDefaultMode
	templatePodSpec.Volumes[tlsSecretVolumeIndex].Secret.Optional = &DesiredTlsSecretVolumeOptionalValue

	commonState.MutateVolumesToIncludeRootCAsVolume(templatePodSpec)
}

func (obj *EnforcerDeploymentK8sObject) mutateAffinityAndNodeSelector(templatePodSpec *coreV1.PodSpec, enforcerSpec *cbcontainersv1.CBContainersEnforcerSpec) {
	templatePodSpec.Affinity = enforcerSpec.Affinity
	templatePodSpec.NodeSelector = enforcerSpec.NodeSelector
}

func (obj *EnforcerDeploymentK8sObject) mutateContainersList(templatePodSpec *coreV1.PodSpec, agentSpec *cbcontainersv1.CBContainersAgentSpec) {
	if len(templatePodSpec.Containers) != 1 {
		container := coreV1.Container{}
		templatePodSpec.Containers = []coreV1.Container{container}
	}

	obj.mutateContainer(&templatePodSpec.Containers[0], agentSpec)
}

func (obj *EnforcerDeploymentK8sObject) mutateContainer(container *coreV1.Container, agentSpec *cbcontainersv1.CBContainersAgentSpec) {
	enforcerSpec := &agentSpec.Components.Basic.Enforcer

	container.Name = EnforcerName
	container.Resources = enforcerSpec.Resources
	obj.mutateEnforcerEnvVars(container, agentSpec)
	commonState.MutateImage(container, enforcerSpec.Image, agentSpec.Version, agentSpec.Components.Settings.DefaultImagesRegistry)
	commonState.MutateContainerHTTPProbes(container, enforcerSpec.Probes)
	obj.mutateSecurityContext(container)
	obj.mutateContainerPorts(container)
	obj.mutateVolumesMounts(container)
}

func (obj *EnforcerDeploymentK8sObject) mutateEnforcerEnvVars(container *coreV1.Container, agentSpec *cbcontainersv1.CBContainersAgentSpec) {
	enforcerSpec := &agentSpec.Components.Basic.Enforcer

	customEnvs := []coreV1.EnvVar{
		{Name: "GUARDRAILS_ENFORCER_KEY_FILE_PATH", Value: fmt.Sprintf("%s/key", DesiredTlsSecretVolumeMountPath)},
		{Name: "GUARDRAILS_ENFORCER_CERT_FILE_PATH", Value: fmt.Sprintf("%s/signed_cert", DesiredTlsSecretVolumeMountPath)},
		{Name: "GUARDRAILS_ENFORCER_PROMETHEUS_PORT", Value: fmt.Sprintf("%d", enforcerSpec.Prometheus.Port)},
		{Name: "GUARDRAILS_ENFORCER_PORT", Value: fmt.Sprintf("%d", DesiredContainerPortValue)},
		{Name: "GIN_MODE", Value: "release"},
	}

	envVarBuilder := commonState.NewEnvVarBuilder().
		WithCommonDataPlane(agentSpec.AccessTokenSecretName).
		WithEventsGateway(&agentSpec.Gateways.HardeningEventsGateway).
		WithCustom(customEnvs...).
		WithSpec(enforcerSpec.Env).
		WithProxySettings(agentSpec.Components.Settings.Proxy)
	commonState.MutateEnvVars(container, envVarBuilder)
}

func (obj *EnforcerDeploymentK8sObject) mutateSecurityContext(container *coreV1.Container) {
	if container.SecurityContext == nil {
		container.SecurityContext = &coreV1.SecurityContext{}
	}

	container.SecurityContext.AllowPrivilegeEscalation = &EnforcerAllowPrivilegeEscalation
	container.SecurityContext.ReadOnlyRootFilesystem = &EnforcerReadOnlyRootFilesystem
	container.SecurityContext.RunAsNonRoot = &EnforcerRunAsNonRoot
	container.SecurityContext.RunAsUser = &EnforcerRunAsUser
	container.SecurityContext.Capabilities = &coreV1.Capabilities{
		Drop: EnforcerCapabilitiesToDrop,
	}
}

func (obj *EnforcerDeploymentK8sObject) mutateContainerPorts(container *coreV1.Container) {
	if container.Ports == nil || len(container.Ports) != 1 {
		container.Ports = []coreV1.ContainerPort{{}}
	}

	container.Ports[0].Name = DesiredContainerPortName
	container.Ports[0].ContainerPort = DesiredContainerPortValue
}

func (obj *EnforcerDeploymentK8sObject) mutateVolumesMounts(container *coreV1.Container) {
	if container.VolumeMounts == nil || len(container.VolumeMounts) != 2 {
		container.VolumeMounts = make([]coreV1.VolumeMount, 0)
	}

	tlsSecretVolumeMountIndex := commonState.EnsureAndGetVolumeMountIndexForName(container, DesiredTlsSecretVolumeName)
	commonState.MutateVolumeMount(container, tlsSecretVolumeMountIndex, DesiredTlsSecretVolumeMountPath, DesiredTlsSecretVolumeMountReadOnly)

	commonState.MutateVolumeMountToIncludeRootCAsVolumeMount(container)
}
