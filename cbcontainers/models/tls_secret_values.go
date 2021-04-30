package models

const (
	CaCertKey     = "ca.crt"
	CaKeyKey      = "ca.key"
	SignedCertKey = "signed_cert"
	KeyKey        = "key"
)

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

func TlsSecretValuesFromSecretData(data map[string][]byte) TlsSecretValues {
	return NewTlsSecretValues(data[CaCertKey], data[CaKeyKey], data[SignedCertKey], data[KeyKey])
}

func (tlsSecretValues TlsSecretValues) ToDataMap() map[string][]byte {
	return map[string][]byte{
		CaCertKey:     tlsSecretValues.CaCert,
		CaKeyKey:      tlsSecretValues.CaKey,
		SignedCertKey: tlsSecretValues.SignedCert,
		KeyKey:        tlsSecretValues.Key,
	}
}
