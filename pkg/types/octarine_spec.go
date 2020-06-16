package types

type octarine struct {
	Account           string
	Domain            string
	AccessToken       string
	AccessTokenSecret string
	Api               HostPort
	Messageproxy      HostPort
}

type global struct {
	Octarine octarine
}

type admissionController struct {
	TimeoutSeconds int
	AutoManage     bool
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
type OctarineSpec struct {
	Global     global
	Guardrails guardrailsSpec
	Nodeguard  nodeguardSpec
}

func NewOctarineSpec() *OctarineSpec {
	spec := new(OctarineSpec)

	// Set default values
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
