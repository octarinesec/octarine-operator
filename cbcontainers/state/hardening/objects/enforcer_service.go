package objects

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DesiredServicePortName  = "https"
	DesiredServicePortValue = 443
)

type EnforcerServiceK8sObject struct {
	tlsSecretsValuesCreator TlsSecretsValuesCreator
}

func NewEnforcerServiceK8sObject() *EnforcerServiceK8sObject { return &EnforcerServiceK8sObject{} }

func (obj *EnforcerServiceK8sObject) EmptyK8sObject() client.Object {
	return &coreV1.Service{}
}

func (obj *EnforcerServiceK8sObject) HardeningChildNamespacedName(_ *cbcontainersv1.CBContainersHardening) types.NamespacedName {
	return types.NamespacedName{Name: EnforcerName, Namespace: commonState.DataPlaneNamespaceName}
}

func (obj *EnforcerServiceK8sObject) MutateHardeningChildK8sObject(k8sObject client.Object, cbContainersHardening *cbcontainersv1.CBContainersHardening) error {
	service, ok := k8sObject.(*coreV1.Service)
	if !ok {
		return fmt.Errorf("expected Service K8s object")
	}

	enforcerSpec := cbContainersHardening.Spec.EnforcerSpec

	service.Labels = enforcerSpec.DeploymentLabels
	service.Spec.Type = coreV1.ServiceTypeClusterIP
	service.Spec.Selector = enforcerSpec.PodTemplateLabels
	obj.mutatePorts(service)

	return nil
}

func (obj *EnforcerServiceK8sObject) mutatePorts(service *coreV1.Service) {
	if service.Spec.Ports == nil || len(service.Spec.Ports) != 1 {
		service.Spec.Ports = []coreV1.ServicePort{{}}
	}

	service.Spec.Ports[0].Name = DesiredServicePortName
	service.Spec.Ports[0].TargetPort = intstr.FromString(DesiredContainerPortName)
	service.Spec.Ports[0].Port = DesiredServicePortValue
}
