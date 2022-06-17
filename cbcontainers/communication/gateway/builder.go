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
	agentComponents       []string
	clusterLabels         map[string]string
	scheme                string
	host                  string
	port                  int
	adapter               string
	tlsInsecureSkipVerify bool
	tlsRootCAsBundle      []byte
}

func NewBuilder(account, cluster, accessToken, host string, clusterLabels map[string]string) *Builder {
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
		clusterLabels:         clusterLabels,
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
	builder.agentComponents = append(builder.agentComponents, models.AgentComponentRuntimeProtection)
	return builder
}

func (builder *Builder) WithClusterScanning() *Builder {
	builder.agentComponents = append(builder.agentComponents, models.AgentComponentClusterScanning)
	return builder
}

func (builder *Builder) WithGuardrailsEnforce() *Builder {
	builder.agentComponents = append(builder.agentComponents, models.AgentComponentGuardrailsEnforce)
	return builder
}

func (builder *Builder) Build() (*ApiGateway, error) {
	return NewApiGateway(
		builder.account,
		builder.cluster,
		builder.accessToken,
		builder.agentComponents,
		builder.clusterLabels,
		builder.scheme,
		builder.host,
		builder.port,
		builder.adapter,
		builder.tlsInsecureSkipVerify,
		builder.tlsRootCAsBundle)
}
