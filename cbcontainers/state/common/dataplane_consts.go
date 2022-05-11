package common

const (
	DataPlaneNamespaceName = "cbcontainers-dataplane"

	DataPlaneConfigmapName      = "cbcontainers-dataplane-config"
	RegistrySecretName          = "cbcontainers-registry-secret"
	DataPlaneServiceAccountName = "cbcontainers-operator"
	DataPlanePriorityClassName  = "cbcontainers-dataplane-priority-class"

	DataPlaneConfigmapAccountKey        = "Account"
	DataPlaneConfigmapClusterKey        = "Cluster"
	DataPlaneConfigmapAgentVersionKey   = "AgentVersion"
	DataPlaneConfigmapApiSchemeKey      = "ApiScheme"
	DataPlaneConfigmapApiHostKey        = "ApiHost"
	DataPlaneConfigmapApiPortKey        = "ApiPort"
	DataPlaneConfigmapApiAdapterKey     = "ApiAdapter"
	DataPlaneConfigmapTlsSkipVerifyKey  = "TLS.SkipVerify"
	DataPlaneConfigmapTlsRootCAsPathKey = "TLS.RootCAsPath"

	DataPlaneConfigmapTlsRootCAsDirPath  = "/etc/gateway-certs"
	DataPlaneConfigmapTlsRootCAsFilePath = "root.pem"

	AccessTokenSecretKeyName = "accessToken"

	RootCAVolumeDefaultMode int32 = 420 // 644 in octal
)
