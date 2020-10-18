package octarine_api

import (
	"crypto/tls"
	"fmt"
	"github.com/go-resty/resty/v2"
	. "github.com/octarinesec/octarine-operator/pkg/types"
	"net/http"
)

type OctarineApiClient struct {
	account     	string
	accessToken 	string
	api         	ApiSpec
	client      	*resty.Client
}

func NewOctarineApiClient(account, accessToken string, api ApiSpec) *OctarineApiClient {
	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	return &OctarineApiClient{account, accessToken, api,client}
}

func (o *OctarineApiClient) baseUrl() string {
	return fmt.Sprintf("https://%s:%d/%v/v1/orgs/%v", o.api.Host, o.api.Port, o.api.AdapterName, o.account)
}

func (o *OctarineApiClient) baseRequest() *resty.Request {
	r := o.client.R()
	r.SetHeader("X-Auth-Token", o.accessToken)
	return r
}

func (o *OctarineApiClient) baseRequestWithRetries() *resty.Request {
	r := o.client.
		SetRetryCount(3).
		SetRetryWaitTime(5).
		R()
	r.SetHeader("X-Auth-Token", o.accessToken)
	return r
}

func (o *OctarineApiClient) RegisterDomain(domain string) error {
	url := fmt.Sprintf("%s/account/%s/domains", o.baseUrl(), o.account)

	resp, err := o.baseRequest().
		SetBody(map[string]interface{}{
			"name":           domain,
			"labels":         map[string]string{},
			"inbounddefault": "allow",
		}).
		Post(url)

	if err != nil {
		return err
	} else if !resp.IsSuccess() && resp.StatusCode() != http.StatusConflict { // ignore conflict (409) response, which means the domain already exists
		return fmt.Errorf("failed registering domain %s (%d): %s", domain, resp.StatusCode(), resp)
	}

	return nil
}

func (o *OctarineApiClient) RegisterAccountFeatures(features ...AccountFeature) error {
	url := fmt.Sprintf("%s/accounts/%s", o.baseUrl(), o.account)

	resp, err := o.baseRequest().
		SetBody(map[string]interface{}{
			"features": features,
		}).
		Patch(url)

	if err != nil {
		return err
	} else if !resp.IsSuccess() {
		return fmt.Errorf("failed registering account features (%d): %s", resp.StatusCode(), resp)
	}

	return nil
}

func (o *OctarineApiClient) GetRegistrySecret() (*RegistrySecret, error) {
	url := fmt.Sprintf("%s/account/%s/registrySecret", o.baseUrl(), o.account)

	resp, err := o.baseRequest().
		SetResult(&RegistrySecret{}).
		Get(url)

	if err != nil {
		return nil, err
	} else if !resp.IsSuccess() {
		return nil, fmt.Errorf("failed retrieving registry secret (%d): %s", resp.StatusCode(), resp)
	}

	return resp.Result().(*RegistrySecret), nil
}
