package common_specs

type CBContainersPrometheusSpec struct {
	// +kubebuilder:default:=false
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:default:=7071
	Port int `json:"port,omitempty"`
}
