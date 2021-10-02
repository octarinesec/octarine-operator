package v1

type CBContainersGatewaysSpec struct {
	// +kubebuilder:default:=<>
	GatewayTLS             CBContainersGatewayTLS        `json:"gatewayTLS,omitempty"`
	ApiGateway             CBContainersApiGatewaySpec    `json:"apiGateway,required"`
	CoreEventsGateway      CBContainersEventsGatewaySpec `json:"coreEventsGateway,required"`
	HardeningEventsGateway CBContainersEventsGatewaySpec `json:"hardeningEventsGateway,required"`
	RuntimeEventsGateway   CBContainersEventsGatewaySpec `json:"runtimeEventsGateway,required"`
}

type CBContainersGatewayTLS struct {
	// +kubebuilder:default:=false
	InsecureSkipVerify bool   `json:"insecureSkipVerify,omitempty"`
	RootCAsBundle      []byte `json:"rootCAsBundle,omitempty"`
}

type CBContainersEventsGatewaySpec struct {
	Host string `json:"host,required"`
	// +kubebuilder:default:=443
	Port int `json:"port,omitempty"`
}

type CBContainersApiGatewaySpec struct {
	Host string `json:"host,required"`
	// +kubebuilder:default:="https"
	Scheme string `json:"scheme,omitempty"`
	// +kubebuilder:default:=443
	Port int `json:"port,omitempty"`
	// +kubebuilder:default:="containers"
	Adapter string `json:"adapter,omitempty"`
}
