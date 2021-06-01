package v1

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
	// +kubebuilder:default:="cbcontainers-access-token"
	AccessTokenSecretName string `json:"accessTokenSecretName,omitempty"`
}
