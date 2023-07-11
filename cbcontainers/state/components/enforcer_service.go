package components

import (
	"fmt"

	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
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
	// Namespace is the Namespace in which the Service will be created.
	Namespace string
}

func NewEnforcerServiceK8sObject(namespace string) *EnforcerServiceK8sObject {
	return &EnforcerServiceK8sObject{
		Namespace: namespace,
	}
}

func (obj *EnforcerServiceK8sObject) EmptyK8sObject() client.Object {
	return &coreV1.Service{}
}

func (obj *EnforcerServiceK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: EnforcerName, Namespace: obj.Namespace}
}

func (obj *EnforcerServiceK8sObject) MutateK8sObject(k8sObject client.Object, agentSpec *cbcontainersv1.CBContainersAgentSpec) error {
	service, ok := k8sObject.(*coreV1.Service)
	if !ok {
		return fmt.Errorf("expected Service K8s object")
	}

	enforcer := &agentSpec.Components.Basic.Enforcer

	service.Labels = enforcer.Labels
	service.Spec.Type = coreV1.ServiceTypeClusterIP
	service.Spec.Selector = map[string]string{
		EnforcerLabelKey: EnforcerName,
	}
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
