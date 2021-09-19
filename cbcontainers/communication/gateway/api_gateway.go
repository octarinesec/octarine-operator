package gateway

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"net/http"
)

type ApiGateway struct {
	account       string
	cluster       string
	accessToken   string
	agentFeatures []string
	scheme        string
	host          string
	port          int
	adapter       string
	client        *resty.Client
}

func createClient(tlsInsecureSkipVerify bool, rootCAsBundle []byte) (*resty.Client, error) {
	client := resty.New()
	tlsConfig := tls.Config{
		InsecureSkipVerify: tlsInsecureSkipVerify,
	}

	if len(rootCAsBundle) > 0 {
		rootCAs := x509.NewCertPool()
		ok := rootCAs.AppendCertsFromPEM(rootCAsBundle)
		if !ok {
			return nil, fmt.Errorf("failed to parse root certificates")
		}

		tlsConfig.RootCAs = rootCAs
	}

	client.SetTLSClientConfig(&tlsConfig)
	return client, nil
}

func NewApiGateway(account, cluster string, accessToken string, agentFeatures []string, scheme, host string, port int, adapter string,
	tlsInsecureSkipVerify bool, rootCAsBundle []byte) (*ApiGateway, error) {

	client, err := createClient(tlsInsecureSkipVerify, rootCAsBundle)
	if err != nil {
		return nil, err
	}

	return &ApiGateway{
		account:       account,
		cluster:       cluster,
		accessToken:   accessToken,
		agentFeatures: agentFeatures,
		scheme:        scheme,
		host:          host,
		port:          port,
		adapter:       adapter,
		client:        client,
	}, nil
}

func (gateway *ApiGateway) baseUrl(postFix string) string {
	return fmt.Sprintf("%v://%s:%d/%v/v1/orgs/%v/%s", gateway.scheme, gateway.host, gateway.port, gateway.adapter, gateway.account, postFix)
}

func (gateway *ApiGateway) baseRequest() *resty.Request {
	r := gateway.client.R()
	r.SetHeader("X-Auth-Token", gateway.accessToken)
	return r
}

func (gateway *ApiGateway) baseRequestWithRetries() *resty.Request {
	r := gateway.client.
		SetRetryCount(3).
		SetRetryWaitTime(5).
		R()
	r.SetHeader("X-Auth-Token", gateway.accessToken)
	return r
}

func (gateway *ApiGateway) getResourcePathWithAccountPath(resourceName string) string {
	return gateway.baseUrl(fmt.Sprintf("account/%s/%s", gateway.account, resourceName))
}

func (gateway *ApiGateway) RegisterCluster() error {
	url := gateway.getResourcePathWithAccountPath("clusters")
	resp, err := gateway.baseRequest().
		SetBody(map[string]interface{}{
			"name":           gateway.cluster,
			"agent_features": gateway.agentFeatures,
			"labels":         map[string]string{},
			"inbounddefault": "allow",
		}).
		Post(url)

	if err != nil {
		return err
	} else if !resp.IsSuccess() && resp.StatusCode() != http.StatusConflict { // ignore conflict (409) response, which means the domain already exists
		return fmt.Errorf("failed creating cluster %s (%d): %s", gateway.cluster, resp.StatusCode(), resp)
	}

	return nil
}

func (gateway *ApiGateway) GetRegistrySecret() (*models.RegistrySecretValues, error) {
	url := gateway.getResourcePathWithAccountPath("registrySecret")

	resp, err := gateway.baseRequest().
		SetResult(&models.RegistrySecretValues{}).
		Get(url)

	if err != nil {
		return nil, err
	} else if !resp.IsSuccess() {
		return nil, fmt.Errorf("failed retrieving registry secret (%d): %s", resp.StatusCode(), resp)
	}

	return resp.Result().(*models.RegistrySecretValues), nil
}
