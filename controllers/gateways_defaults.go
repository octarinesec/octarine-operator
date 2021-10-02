package controllers

import cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"

func (r *CBContainersAgentController) setGatewaysDefaults(gateways *cbcontainersv1.CBContainersGatewaysSpec) {
	r.setGatewayTLSDefaults(&gateways.GatewayTLS)
	r.setAPIGatewayDefaults(&gateways.ApiGateway)
	r.setEventsGatewayDefaults(&gateways.CoreEventsGateway)
	r.setEventsGatewayDefaults(&gateways.HardeningEventsGateway)
	r.setEventsGatewayDefaults(&gateways.RuntimeEventsGateway)
}

func (r *CBContainersAgentController) setGatewayTLSDefaults(gatewayTLS *cbcontainersv1.CBContainersGatewayTLS) {
	if gatewayTLS.RootCAsBundle == nil {
		gatewayTLS.RootCAsBundle = make([]byte, 0)
	}
}

func (r *CBContainersAgentController) setAPIGatewayDefaults(apiGateway *cbcontainersv1.CBContainersApiGatewaySpec) {
	if apiGateway.Scheme == "" {
		apiGateway.Scheme = "https"
	}

	if apiGateway.Port == 0 {
		apiGateway.Port = 443
	}

	if apiGateway.Adapter == "" {
		apiGateway.Adapter = "containers"
	}
}

func (r *CBContainersAgentController) setEventsGatewayDefaults(eventsGateway *cbcontainersv1.CBContainersEventsGatewaySpec) {
	if eventsGateway.Port == 0 {
		eventsGateway.Port = 443
	}
}
