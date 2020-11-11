package types

import (
	v1 "k8s.io/api/core/v1"
)

type ApiSpec struct {
	Host 		string
	Port 		int
	AdapterName string
}

type RegistrySecret struct {
	Type v1.SecretType
	Data map[string][]byte
}
