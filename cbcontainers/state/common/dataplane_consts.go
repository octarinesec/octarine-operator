package common

const (
	DataPlaneNamespaceName = "cbcontainers-dataplane"

	DataPlaneConfigmapName     = "cbcontainers-dataplane-config"
	RegistrySecretName         = "cbcontainers-registry-secret"
	DataPlanePriorityClassName = "cbcontainers-dataplane-priority-class"

	DataPlaneConfigmapAccountKey    = "Account"
	DataPlaneConfigmapClusterKey    = "Cluster"
	DataPlaneConfigmapApiSchemeKey  = "ApiScheme"
	DataPlaneConfigmapApiHostKey    = "ApiHost"
	DataPlaneConfigmapApiPortKey    = "ApiPort"
	DataPlaneConfigmapApiAdapterKey = "ApiAdapter"

	AccessTokenSecretKeyName = "accessToken"
)
