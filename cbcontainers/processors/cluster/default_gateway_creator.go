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

func (creator *DefaultGatewayCreator) CreateGateway(cbContainersCluster *cbcontainersv1.CBContainersAgent, accessToken string) (Gateway, error) {
	spec := cbContainersCluster.Spec
	return gateway.NewBuilder(spec.Account, spec.ClusterName, accessToken, spec.ApiGatewaySpec.Host).
		SetURLComponents(spec.ApiGatewaySpec.Scheme, spec.ApiGatewaySpec.Port, spec.ApiGatewaySpec.Adapter).
		SetTLSInsecureSkipVerify(spec.GatewayTLS.InsecureSkipVerify).
		SetTLSRootCAsBundle(spec.GatewayTLS.RootCAsBundle).
		Build()
}
