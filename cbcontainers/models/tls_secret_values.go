package models

type TlsSecretValues struct {
	CaCert     []byte `json:"caCert"`
	CaKey      []byte `json:"caKey"`
	SignedCert []byte `json:"signedCert"`
	Key        []byte `json:"key"`
}

func NewTlsSecretValues(caCert, caKey, signedCert, key []byte) TlsSecretValues {
	return TlsSecretValues{
		CaCert:     caCert,
		CaKey:      caKey,
		SignedCert: signedCert,
		Key:        key,
	}
}
