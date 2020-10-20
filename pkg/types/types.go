package types

import (
	v1 "k8s.io/api/core/v1"
)

type ApiSpec struct {
	Host 		string
	Port 		int
	AdapterName string
}

// Octarine account features to indicate which features are installed on the cluster
type AccountFeature string

const (
	Guardrails = AccountFeature("guardrail")
	Nodeaguard = AccountFeature("nodeguard")
)

type RegistrySecret struct {
	Type v1.SecretType
	Data map[string][]byte
}
