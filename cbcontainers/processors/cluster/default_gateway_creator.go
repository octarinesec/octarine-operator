package cluster

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/communication/gateway"
)

type DefaultGatewayCreator struct {
}

func NewDefaultGatewayCreator() *DefaultGatewayCreator {
	return &DefaultGatewayCreator{}
}

func (creator *DefaultGatewayCreator) CreateGateway(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) *gateway.ApiGateway {
	spec := cbContainersCluster.Spec
	return gateway.NewApiGateway(spec.Account, spec.ClusterName, accessToken, spec.ApiGatewaySpec.Scheme, spec.ApiGatewaySpec.Host, spec.ApiGatewaySpec.Port, spec.ApiGatewaySpec.Adapter)
}
