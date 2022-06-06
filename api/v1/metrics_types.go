package v1

type CBContainersPrometheusSpec struct {
	// +kubebuilder:default:=false
	Enabled *bool `json:"enabled,omitempty"`
	Port    int   `json:"port,omitempty"`
}
