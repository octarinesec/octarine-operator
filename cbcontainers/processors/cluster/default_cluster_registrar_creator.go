package cluster

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/communication/gateway"
)

type DefaultClusterRegistrarCreator struct {
}

func NewDefaultClusterRegistrarCreator() *DefaultClusterRegistrarCreator {
	return &DefaultClusterRegistrarCreator{}
}

func (registrarCreator *DefaultClusterRegistrarCreator) CreateClusterRegistrar(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) ClusterRegistrar {
	spec := cbContainersCluster.Spec
	return gateway.NewApiGateway(spec.Account, spec.ClusterName, accessToken, spec.ApiGatewaySpec.Scheme, spec.ApiGatewaySpec.Host, spec.ApiGatewaySpec.Port, spec.ApiGatewaySpec.Adapter)
}
