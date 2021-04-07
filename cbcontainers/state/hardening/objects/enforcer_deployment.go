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
	EnforcerName = "cbcontainers-hardening-enforcer"

	DesiredContainerPortName  = "https"
	DesiredContainerPortValue = 443

	DesiredTlsSecretVolumeName      = "cert"
	DesiredTlsSecretVolumeMountPath = "/etc/octarine-certificates"
)

var (
	DesiredTlsSecretVolumeDecimalDefaultMode int32 = 420 // 644 in octal
	DesiredTlsSecretVolumeOptionalValue            = true

	EnforcerAllowPrivilegeEscalation       = false
	EnforcerReadOnlyRootFilesystem         = true
	EnforcerRunAsUser                int64 = 0
	EnforcerCapabilitiesToAdd              = []coreV1.Capability{"NET_BIND_SERVICE"}
	EnforcerCapabilitiesToDrop             = []coreV1.Capability{"ALL"}

	EnforcerEnvVars = []coreV1.EnvVar{
		{Name: "GUARDRAILS_ENFORCER_KEY_FILE_PATH", Value: fmt.Sprintf("%s/key", DesiredTlsSecretVolumeMountPath)},
		{Name: "GUARDRAILS_ENFORCER_CERT_FILE_PATH", Value: fmt.Sprintf("%s/signed_cert", DesiredTlsSecretVolumeMountPath)},
		{Name: "GIN_MODE", Value: "release"},
	}
)

type EnforcerDeploymentK8sObject struct{}

func NewEnforcerDeploymentK8sObject() *EnforcerDeploymentK8sObject {
	return &EnforcerDeploymentK8sObject{}
}

func (obj *EnforcerDeploymentK8sObject) EmptyK8sObject() client.Object {
	return &appsV1.Deployment{}
}

func (obj *EnforcerDeploymentK8sObject) HardeningChildNamespacedName(_ *cbcontainersv1.CBContainersHardening) types.NamespacedName {
	return types.NamespacedName{Name: EnforcerName, Namespace: commonState.DataPlaneNamespaceName}
}

func (obj *EnforcerDeploymentK8sObject) MutateHardeningChildK8sObject(k8sObject client.Object, cbContainersHardening *cbcontainersv1.CBContainersHardening) error {
	enforcerSpec := cbContainersHardening.Spec.EnforcerSpec

	deployment, ok := k8sObject.(*appsV1.Deployment)
	if !ok {
		return fmt.Errorf("expected Deployment K8s object")
	}

	if deployment.Spec.Selector == nil {
		deployment.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: enforcerSpec.PodTemplateLabels,
		}
	}

	if deployment.ObjectMeta.Annotations == nil {
		deployment.ObjectMeta.Annotations = make(map[string]string)
	}

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	deployment.Spec.Replicas = &enforcerSpec.ReplicasCount
	deployment.ObjectMeta.Labels = enforcerSpec.DeploymentLabels
	deployment.Spec.Selector.MatchLabels = enforcerSpec.PodTemplateLabels
	deployment.Spec.Template.ObjectMeta.Labels = enforcerSpec.PodTemplateLabels
	applyment.EnforceMapContains(deployment.ObjectMeta.Annotations, enforcerSpec.DeploymentAnnotations)
	applyment.EnforceMapContains(deployment.Spec.Template.ObjectMeta.Annotations, enforcerSpec.PodTemplateAnnotations)
	deployment.Spec.Template.Spec.ImagePullSecrets = []coreV1.LocalObjectReference{{Name: commonState.RegistrySecretName}}
	obj.mutateVolumes(&deployment.Spec.Template.Spec)
	obj.mutateContainersList(&deployment.Spec.Template.Spec, &cbContainersHardening.Spec.EnforcerSpec, &cbContainersHardening.Spec.EventsGatewaySpec, cbContainersHardening.Spec.Version, cbContainersHardening.Spec.AccessTokenSecretName)
	//applyment.MutateString(enforcerSpec.ServiceAccountName, func() *string { return &template.Spec.ServiceAccountName }, func(value string) { template.Spec.ServiceAccountName = value })
	//applyment.MutateString(enforcerSpec.PriorityClassName, func() *string { return &template.Spec.PriorityClassName }, func(value string) { template.Spec.PriorityClassName = value })

	return nil
}

func (obj *EnforcerDeploymentK8sObject) mutateVolumes(templatePodSpec *coreV1.PodSpec) {
	if templatePodSpec.Volumes == nil || len(templatePodSpec.Volumes) != 1 || templatePodSpec.Volumes[0].Secret == nil {
		templatePodSpec.Volumes = []coreV1.Volume{
			{
				VolumeSource: coreV1.VolumeSource{
					Secret: &coreV1.SecretVolumeSource{},
				},
			},
		}
	}

	templatePodSpec.Volumes[0].Name = DesiredTlsSecretVolumeName
	templatePodSpec.Volumes[0].Secret.SecretName = EnforcerTlsName
	templatePodSpec.Volumes[0].Secret.DefaultMode = &DesiredTlsSecretVolumeDecimalDefaultMode
	templatePodSpec.Volumes[0].Secret.Optional = &DesiredTlsSecretVolumeOptionalValue
}

func (obj *EnforcerDeploymentK8sObject) mutateContainersList(templatePodSpec *coreV1.PodSpec, enforcerSpec *cbcontainersv1.CBContainersHardeningEnforcerSpec, eventsGatewaySpec *cbcontainersv1.CBContainersHardeningEventsGatewaySpec, version, accessTokenSecretName string) {
	if len(templatePodSpec.Containers) != 1 {
		container := coreV1.Container{}
		templatePodSpec.Containers = []coreV1.Container{container}
	}

	obj.mutateContainer(&templatePodSpec.Containers[0], enforcerSpec, eventsGatewaySpec, version, accessTokenSecretName)
}

func (obj *EnforcerDeploymentK8sObject) mutateContainer(container *coreV1.Container, enforcerSpec *cbcontainersv1.CBContainersHardeningEnforcerSpec, eventsGatewaySpec *cbcontainersv1.CBContainersHardeningEventsGatewaySpec, version, accessTokenSecretName string) {
	container.Name = EnforcerName
	container.Resources = enforcerSpec.Resources
	mutateEnvVars(container, enforcerSpec.Env, accessTokenSecretName, eventsGatewaySpec, EnforcerEnvVars...)
	mutateImage(container, enforcerSpec.Image, version)
	mutateContainerProbes(container, enforcerSpec.Probes)
	obj.mutateSecurityContext(container)
	obj.mutateContainerPorts(container)
	obj.mutateVolumesMounts(container)
}

func (obj *EnforcerDeploymentK8sObject) mutateSecurityContext(container *coreV1.Container) {
	if container.SecurityContext == nil {
		container.SecurityContext = &coreV1.SecurityContext{}
	}

	container.SecurityContext.AllowPrivilegeEscalation = &EnforcerAllowPrivilegeEscalation
	container.SecurityContext.ReadOnlyRootFilesystem = &EnforcerReadOnlyRootFilesystem
	container.SecurityContext.RunAsUser = &EnforcerRunAsUser
	container.SecurityContext.Capabilities = &coreV1.Capabilities{
		Add:  EnforcerCapabilitiesToAdd,
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
	if container.VolumeMounts == nil || len(container.VolumeMounts) != 1 {
		container.VolumeMounts = []coreV1.VolumeMount{{}}
	}

	container.VolumeMounts[0].Name = DesiredTlsSecretVolumeName
	container.VolumeMounts[0].MountPath = DesiredTlsSecretVolumeMountPath
}