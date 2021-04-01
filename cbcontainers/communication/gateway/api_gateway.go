package gateway

import (
	"crypto/tls"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"net/http"
)

type ApiGateway struct {
	account     string
	cluster     string
	accessToken string
	scheme      string
	host        string
	port        int
	adapter     string
	client      *resty.Client
}

func NewApiGateway(account, cluster string, accessToken string, scheme, host string, port int, adapter string) *ApiGateway {
	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	return &ApiGateway{
		account:     account,
		cluster:     cluster,
		accessToken: accessToken,
		scheme:      scheme,
		host:        host,
		port:        port,
		adapter:     adapter,
		client:      client,
	}
}

func (gateway *ApiGateway) baseUrl() string {
	return fmt.Sprintf("%v://%s:%d/%v/v1/orgs/%v", gateway.scheme, gateway.host, gateway.port, gateway.adapter, gateway.account)
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
	return fmt.Sprintf("%s/account/%s/%s", gateway.baseUrl(), gateway.account, resourceName)
}

func (gateway *ApiGateway) RegisterCluster() error {
	url := gateway.getResourcePathWithAccountPath("clusters")
	resp, err := gateway.baseRequest().
		SetBody(map[string]interface{}{
			"name":           gateway.cluster,
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
