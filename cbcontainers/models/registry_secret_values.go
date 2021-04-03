package models

import v1 "k8s.io/api/core/v1"

type RegistrySecretValues struct {
	Type v1.SecretType     `json:"type"`
	Data map[string][]byte `json:"data"`
}
