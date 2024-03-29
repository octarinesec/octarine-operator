package gateway

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
)

type DefaultGatewayCreator struct {
}

func NewDefaultGatewayCreator() *DefaultGatewayCreator {
	return &DefaultGatewayCreator{}
}

func (creator *DefaultGatewayCreator) CreateGateway(cbContainersAgent *cbcontainersv1.CBContainersAgent, accessToken string) (*ApiGateway, error) {
	spec := cbContainersAgent.Spec
	builder := NewBuilder(spec.Account, spec.ClusterName, accessToken, spec.Gateways.ApiGateway.Host, cbContainersAgent.ObjectMeta.Labels).
		SetURLComponents(spec.Gateways.ApiGateway.Scheme, spec.Gateways.ApiGateway.Port, spec.Gateways.ApiGateway.Adapter).
		SetTLSInsecureSkipVerify(spec.Gateways.GatewayTLS.InsecureSkipVerify).
		SetTLSRootCAsBundle(spec.Gateways.GatewayTLS.RootCAsBundle)

	if spec.Components.RuntimeProtection.Enabled != nil && *spec.Components.RuntimeProtection.Enabled {
		builder.WithRuntimeProtection()
	}

	if spec.Components.ClusterScanning.Enabled != nil && *spec.Components.ClusterScanning.Enabled {
		builder.WithClusterScanning()
	}

	if spec.Components.Basic.Enforcer.EnableEnforcementFeature != nil && *spec.Components.Basic.Enforcer.EnableEnforcementFeature {
		builder.WithGuardrailsEnforce()
	}

	if spec.Components.Cndr != nil && spec.Components.Cndr.Enabled != nil && *spec.Components.ClusterScanning.Enabled {
		builder.WithCndr()
	}

	return builder.Build()
}
