package types

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type octarine struct {
	Account           string
	Domain            string
	AccessToken       string
	AccessTokenSecret string
	ApiAdapterName    string
	Api               ApiSpec
	Messageproxy      ApiSpec
}

type global struct {
	Octarine octarine
}

type admissionController struct {
	TimeoutSeconds    int
	AutoManage        bool // operator flag only (not from the chart) - default value should be set
	NamespaceSelector *metav1.LabelSelector
}

type enforcer struct {
	AdmissionController admissionController
}

type guardrailsSpec struct {
	Enabled  bool
	Enforcer enforcer
}

type nodeguardSpec struct {
	Enabled bool
}

// The spec of the Octarine CR. This is used by the helm operator as well, thus the spec corresponds to the helm
// chart's values.yaml.
// Default values are loaded from the chart's values.yaml.
type OctarineSpec struct {
	Global     global
	Guardrails guardrailsSpec
	Nodeguard  nodeguardSpec
}

func NewOctarineSpec() *OctarineSpec {
	spec := new(OctarineSpec)

	// Set default values for parameters not loaded from the chart
	spec.Guardrails.Enforcer.AdmissionController.AutoManage = true

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
