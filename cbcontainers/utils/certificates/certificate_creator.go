package certificates

import (
	"fmt"
	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/helpers"
	"github.com/cloudflare/cfssl/initca"
	csigner "github.com/cloudflare/cfssl/signer"
	"github.com/cloudflare/cfssl/signer/local"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"k8s.io/apimachinery/pkg/types"
)

type CertificateCreator struct{}

func NewCertificateCreator() *CertificateCreator { return &CertificateCreator{} }

func (creator *CertificateCreator) CreateTlsSecretsValues(resourceNamespacedName types.NamespacedName) (models.TlsSecretValues, error) {
	caCert, caKey, err := creator.createCertificateAuthority()
	if err != nil {
		return models.TlsSecretValues{}, err
	}

	cert, key, err := creator.createCertFromCA(resourceNamespacedName, caCert, caKey)
	if err != nil {
		return models.TlsSecretValues{}, err
	}

	return models.NewTlsSecretValues(caCert, caKey, cert, key), nil
}

// CreateCertificateAuthority will create a new CA and return the
// pem encoded cert, key and error
func (creator *CertificateCreator) createCertificateAuthority() ([]byte, []byte, error) {
	req := csr.CertificateRequest{
		KeyRequest: &csr.KeyRequest{
			A: "rsa",
			S: 2048,
		},
		CN: "octarine_ca",
		Hosts: []string{
			"octarine_ca",
		},
		CA: &csr.CAConfig{
			Expiry: "8760h",
		},
	}

	//logger.V(1).Info("creating CA")
	cert, _, key, err := initca.New(&req)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

// CreateCertFromCA takes a certificate authority cert and key and generates a
// new cert, and uses the CA to sign it
func (creator *CertificateCreator) createCertFromCA(namespacedName types.NamespacedName, caCert []byte, caKey []byte) ([]byte, []byte, error) {
	// Parse the ca into an x509 object
	parsedCaCert, err := helpers.ParseCertificatePEM(caCert)
	if err != nil {
		return nil, nil, err
	}
	parsedCaKey, err := helpers.ParsePrivateKeyPEM(caKey)
	if err != nil {
		return nil, nil, err
	}

	svcFullname := fmt.Sprintf("%s.%s.svc", namespacedName.Name, namespacedName.Namespace)
	req := csr.CertificateRequest{
		KeyRequest: &csr.KeyRequest{
			A: "rsa",
			S: 2048,
		},
		CN: svcFullname,
		Hosts: []string{
			svcFullname,
		},
	}
	certReq, key, err := csr.ParseRequest(&req)
	if err != nil {
		return nil, nil, err
	}

	signer, err := local.NewSigner(parsedCaKey, parsedCaCert, csigner.DefaultSigAlgo(parsedCaKey), nil)
	if err != nil {
		return nil, nil, err
	}

	//logger.V(1).Info("signing certficiate")
	signedCert, err := signer.Sign(csigner.SignRequest{
		Hosts:   []string{svcFullname},
		Request: string(certReq),
		Subject: &csigner.Subject{
			CN: svcFullname,
		},
		Profile: svcFullname,
	})
	if err != nil {
		return nil, nil, err
	}

	return signedCert, key, nil
}
