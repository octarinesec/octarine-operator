package gateway

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
)

type SignRequest struct {
	KeyBits int    `json:"key-bits" binding:"required"`
	CertPem []byte `json:"cert-pem" binding:"required"`
}

type SignResponse struct {
	Certificate string `json:"certificate" binding:"required"`
}

type TrustChainResponse struct {
	Chain string `json:"chain" binding:"required"`
}

// Makes a CSR
func makeCertificateRequest(privateKey interface{},
	commonName string, organizationalUnit []string, organization []string) ([]byte, error) {

	random := rand.Reader

	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:         commonName,
			OrganizationalUnit: organizationalUnit,
			Organization:       organization,
		},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	derBytes, err := x509.CreateCertificateRequest(random, &template, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate request: %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: derBytes}), nil
}

func (gateway *ApiGateway) getCertificateSigned(pemCsr []byte, asCA bool) (*x509.Certificate, error) {
	urlPath := "certificates/sign"
	if asCA {
		urlPath += "_ca"
	}
	url := gateway.baseUrl(urlPath)
	resp, err := gateway.baseRequest().
		SetResult(&SignResponse{}).
		SetBody(SignRequest{KeyBits: 1024, CertPem: pemCsr}).
		Post(url)

	if err != nil {
		return nil, err
	} else if !resp.IsSuccess() {
		return nil, fmt.Errorf("request to sign certificate has failed (%d): %s", resp.StatusCode(), resp)
	}

	signResp := resp.Result().(*SignResponse)
	if signResp.Certificate == "" {
		return nil, fmt.Errorf("got empty certificate for certificate signing request")
	}
	block, _ := pem.Decode([]byte(signResp.Certificate))
	if block == nil {
		return nil, fmt.Errorf("failed to PEM decode certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return cert, nil
}

// Makes a CSR and gets it signed at the backend
func (gateway *ApiGateway) getSignedCertificate(privateKey interface{}, commonName string, organizationalUnit []string, organization []string, asCA bool) (*x509.Certificate, error) {
	csr, err := makeCertificateRequest(privateKey, commonName, organizationalUnit, organization)
	if err != nil {
		return nil, fmt.Errorf("failed to make certificate request: %v", err)
	}

	return gateway.getCertificateSigned(csr, asCA)
}

func (gateway *ApiGateway) getTrustChain() ([]byte, error) {
	url := gateway.baseUrl("certificates/trustchain")
	resp, err := gateway.baseRequest().
		SetResult(&TrustChainResponse{}).
		Get(url)

	if err != nil {
		return nil, err
	} else if !resp.IsSuccess() {
		return nil, fmt.Errorf("trust chain request failed (%d): %s", resp.StatusCode(), resp)
	}

	tcResp := resp.Result().(*TrustChainResponse)
	return []byte(tcResp.Chain), nil
}

func (gateway *ApiGateway) getCertificates(commonName string, organizationalUnit []string, organization []string, privateKey *rsa.PrivateKey) (*x509.CertPool, *tls.Certificate, error) {
	cert, err := gateway.getSignedCertificate(privateKey, commonName, organizationalUnit, organization, false)
	if err != nil {
		return nil, nil, err
	}

	certPoolPEM, err := gateway.getTrustChain()
	if err != nil {
		return nil, nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certPoolPEM)

	certPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		},
	)
	certPEM = append(certPEM, certPoolPEM...)

	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)

	tlsCert, err := tls.X509KeyPair(certPEM, privateKeyPEM)
	if err != nil {
		return nil, nil, err
	}

	return certPool, &tlsCert, nil
}
