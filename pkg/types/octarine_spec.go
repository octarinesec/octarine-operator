package types

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type octarine struct {
	Account           string
	Domain            string
	AccessToken       string
	AccessTokenSecret string
	Api               HostPort
	Messageproxy      HostPort
	Version           interface{}
}

type global struct {
	Octarine octarine
}

type metadata struct {
	Name string
}

type admissionController struct {
	TimeoutSeconds    int
	AutoManage        bool // operator flag only (not from the chart) - default value should be set
	NamespaceSelector *metav1.LabelSelector
}

type guardrailsSpec struct {
	Enabled             bool
	AdmissionController admissionController
}

type nodeguardSpec struct {
	Enabled bool
}

// The spec of the Octarine CR. This is used by the helm operator as well, thus the spec corresponds to the helm
// chart's values.yaml.
// Default values are loaded from the chart's values.yaml.
type OctarineSpec struct {
	Global     global
	Metadata   metadata
	Guardrails guardrailsSpec
	Nodeguard  nodeguardSpec
}

func NewOctarineSpec() *OctarineSpec {
	spec := new(OctarineSpec)

	// Set default values for parameters not loaded from the chart
	spec.Guardrails.AdmissionController.AutoManage = true

	return spec
}

func (s *OctarineSpec) GetAccountFeatures() (features []AccountFeature) {
	if s.Guardrails.Enabled {
		features = append(features, Guardrails)
	}
	if s.Nodeguard.Enabled {
		features = append(features, Nodeaguard)
	}
	return
}
