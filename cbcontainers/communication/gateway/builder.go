package gateway

import "github.com/vmware/cbcontainers-operator/cbcontainers/models"

const (
	DefaultScheme  = "https"
	DefaultPort    = 443
	DefaultAdapter = "containers"
)

type Builder struct {
	account               string
	cluster               string
	accessToken           string
	agentFeatures         []string
	scheme                string
	host                  string
	port                  int
	adapter               string
	tlsInsecureSkipVerify bool
	tlsRootCAsBundle      []byte
}

func NewBuilder(account, cluster, accessToken, host string) *Builder {
	return &Builder{
		account:               account,
		cluster:               cluster,
		accessToken:           accessToken,
		scheme:                DefaultScheme,
		host:                  host,
		port:                  DefaultPort,
		adapter:               DefaultAdapter,
		tlsInsecureSkipVerify: false,
		tlsRootCAsBundle:      nil,
	}
}

func (builder *Builder) SetURLComponents(scheme string, port int, adapter string) *Builder {
	builder.scheme = scheme
	builder.port = port
	builder.adapter = adapter
	return builder
}

func (builder *Builder) SetTLSInsecureSkipVerify(insecureSkipVerify bool) *Builder {
	builder.tlsInsecureSkipVerify = insecureSkipVerify
	return builder
}

func (builder *Builder) SetTLSRootCAsBundle(rootCAsBundle []byte) *Builder {
	builder.tlsRootCAsBundle = rootCAsBundle
	return builder
}

func (builder *Builder) WithRuntimeProtection() *Builder {
	builder.agentFeatures = append(builder.agentFeatures, models.AgentComponentRuntimeProtection)
	return builder
}

func (builder *Builder) Build() (*ApiGateway, error) {
	return NewApiGateway(
		builder.account,
		builder.cluster,
		builder.accessToken,
		builder.agentFeatures,
		builder.scheme,
		builder.host,
		builder.port,
		builder.adapter,
		builder.tlsInsecureSkipVerify,
		builder.tlsRootCAsBundle)
}
