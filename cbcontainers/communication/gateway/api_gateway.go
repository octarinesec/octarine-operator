package gateway

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
)

var (
	ErrGettingOperatorCompatibility = errors.New("error while getting the operator compatibility")
)

// TODO: Extract the cluster group + name + ID as separate struct identifying a cluster and used together

type ApiGateway struct {
	account         string
	cluster         string
	accessToken     string
	agentComponents []string
	clusterLabels   map[string]string
	scheme          string
	host            string
	port            int
	adapter         string
	client          *resty.Client
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

func NewApiGateway(account, cluster string, accessToken string, agentComponents []string, clusterLabels map[string]string, scheme, host string, port int, adapter string,
	tlsInsecureSkipVerify bool, rootCAsBundle []byte) (*ApiGateway, error) {

	client, err := createClient(tlsInsecureSkipVerify, rootCAsBundle)
	if err != nil {
		return nil, err
	}

	return &ApiGateway{
		account:         account,
		cluster:         cluster,
		accessToken:     accessToken,
		agentComponents: agentComponents,
		clusterLabels:   clusterLabels,
		scheme:          scheme,
		host:            host,
		port:            port,
		adapter:         adapter,
		client:          client,
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
	return gateway.baseUrl(fmt.Sprintf("management/%v", resourceName))
}

func (gateway *ApiGateway) SplitToGroupAndMember() (string, string, error) {
	parts := strings.Split(gateway.cluster, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("cluster name '%v' is not in group:member format with two parts", gateway.cluster)
	}

	return parts[0], parts[1], nil
}
func (gateway *ApiGateway) RegisterCluster(clusterIdentifier string) error {
	url := gateway.getResourcePathWithAccountPath("clusters")

	group, member, err := gateway.SplitToGroupAndMember()
	if err != nil {
		return err
	}

	resp, err := gateway.baseRequest().
		SetBody(map[string]interface{}{
			"group":          group,
			"member":         member,
			"components":     gateway.agentComponents,
			"labels":         gateway.clusterLabels,
			"inbounddefault": "allow",
			"identifier":     clusterIdentifier,
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
	url := gateway.getResourcePathWithAccountPath("registry_secret")

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

func (gateway *ApiGateway) GetCompatibilityMatrixEntryFor(operatorVersion string) (*models.OperatorCompatibility, error) {
	url := gateway.baseUrl("setup/compatibility/{operatorVersion}")
	resp, err := gateway.baseRequest().
		SetResult(&models.OperatorCompatibility{}).
		SetPathParam("operatorVersion", operatorVersion).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGettingOperatorCompatibility, err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("%w: status code (%d)", ErrGettingOperatorCompatibility, resp.StatusCode())
	}
	r, ok := resp.Result().(*models.OperatorCompatibility)
	if !ok {
		return nil, fmt.Errorf("malformed response from the backend")
	}

	return r, nil
}

func (gateway *ApiGateway) GetSensorMetadata() ([]models.SensorMetadata, error) {
	type getSensorsResponse struct {
		Sensors []models.SensorMetadata `json:"sensors"`
	}

	url := gateway.baseUrl("/setup/sensors")
	resp, err := gateway.baseRequest().
		SetResult(getSensorsResponse{}).
		Get(url)

	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("failed to get sensor metadata with status code (%d)", resp.StatusCode())
	}

	r, ok := resp.Result().(getSensorsResponse)
	if !ok {
		return nil, fmt.Errorf("malformed sensor metadata response")
	}
	return r.Sensors, nil
}

func (gateway *ApiGateway) GetConfigurationChanges(ctx context.Context, clusterIdentifier string) ([]models.ConfigurationChange, error) {
	// TODO: Real implementation with CNS-2790
	c := randomRemoteConfigChange()
	if c != nil {
		return []models.ConfigurationChange{*c}, nil

	}
	return nil, nil
}

func (gateway *ApiGateway) UpdateConfigurationChangeStatus(context.Context, models.ConfigurationChangeStatusUpdate) error {
	// TODO: Real implementation with CNS-2790

	return nil
}
