package common

const (
	// DataPlaneNamespaceName is the name of the default namespace, where CBContainers Dataplane is installed.
	//
	// This is configurable by the user, so it should not be assumed that this will always be the name where the dataplane is deployed.
	DataPlaneNamespaceName  = "cbcontainers-dataplane"
	KubeSystemNamespaceName = "kube-system"

	DataPlaneConfigmapName = "cbcontainers-dataplane-config"
	// RegistrySecretName is the name of the secret that contains the image pull secret for the default registry of the agent images.
	//
	// The creation of this secret is optional, as the users may override the agent images and not use the registry we provide.
	RegistrySecretName                = "cbcontainers-registry-secret"
	DataPlaneServiceAccountName       = "cbcontainers-operator"
	AgentNodeServiceAccountName       = "cbcontainers-agent-node"
	StateReporterServiceAccountName   = "cbcontainers-state-reporter"
	EnforcerServiceAccountName        = "cbcontainers-enforcer"
	MonitorServiceAccountName         = "cbcontainers-monitor"
	ImageScanningServiceAccountName   = "cbcontainers-image-scanning"
	RuntimeResolverServiceAccountName = "cbcontainers-runtime-resolver"
	DataPlanePriorityClassName        = "cbcontainers-dataplane-priority-class"

	DataPlaneConfigmapAccountKey        = "Account"
	DataPlaneConfigmapClusterKey        = "Cluster"
	DataPlaneConfigmapAgentVersionKey   = "AgentVersion"
	DataPlaneConfigmapApiSchemeKey      = "ApiScheme"
	DataPlaneConfigmapApiHostKey        = "ApiHost"
	DataPlaneConfigmapApiPortKey        = "ApiPort"
	DataPlaneConfigmapApiAdapterKey     = "ApiAdapter"
	DataPlaneConfigmapTlsSkipVerifyKey  = "TLS.SkipVerify"
	DataPlaneConfigmapTlsRootCAsPathKey = "TLS.RootCAsPath"
	// DataPlaneConfigmapDataplaneNamespaceKey is the config map key that points to the value of the dataplane namespace
	DataPlaneConfigmapDataplaneNamespaceKey = "DataplaneNamespace"

	DataPlaneConfigmapTlsRootCAsDirPath  = "/etc/gateway-certs"
	DataPlaneConfigmapTlsRootCAsFilePath = "root.pem"

	AccessTokenSecretKeyName = "accessToken"

	RootCAVolumeDefaultMode int32 = 420 // 644 in octal
)
