package components

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
	DesiredServiceGRPCPortName = "https"
)

type ResolverServiceK8sObject struct{}

func NewResolverServiceK8sObject() *ResolverServiceK8sObject { return &ResolverServiceK8sObject{} }

func (obj *ResolverServiceK8sObject) EmptyK8sObject() client.Object {
	return &coreV1.Service{}
}

func (obj *ResolverServiceK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: ResolverName, Namespace: commonState.DataPlaneNamespaceName}
}

func (obj *ResolverServiceK8sObject) MutateK8sObject(k8sObject client.Object, agentSpec *cbcontainersv1.CBContainersAgentSpec) error {
	service, ok := k8sObject.(*coreV1.Service)
	if !ok {
		return fmt.Errorf("expected Service K8s object")
	}

	runtimeProtection := &agentSpec.Components.RuntimeProtection
	resolver := &runtimeProtection.Resolver
	service.Labels = resolver.Labels
	service.Spec.Type = coreV1.ServiceTypeClusterIP
	service.Spec.ClusterIP = coreV1.ClusterIPNone

	service.Spec.Selector = map[string]string{
		resolverLabelKey: ResolverName,
	}
	obj.mutatePorts(service, runtimeProtection.InternalGrpcPort)

	return nil
}

func (obj *ResolverServiceK8sObject) mutatePorts(service *coreV1.Service, desiredGRPCPortValue int32) {
	if service.Spec.Ports == nil || len(service.Spec.Ports) != 1 {
		service.Spec.Ports = []coreV1.ServicePort{{}}
	}

	service.Spec.Ports[0].Name = DesiredServiceGRPCPortName
	service.Spec.Ports[0].TargetPort = intstr.FromString(desiredDeploymentGRPCPortName)
	service.Spec.Ports[0].Port = desiredGRPCPortValue
}
