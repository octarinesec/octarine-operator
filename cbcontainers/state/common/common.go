package common

import coreV1 "k8s.io/api/core/v1"

func GetCommonDataPlaneEnvVars(accessKeySecretName string) []coreV1.EnvVar {
	return []coreV1.EnvVar{
		getEnvVarFromConfigmap("OCTARINE_ACCOUNT", DataPlaneConfigmapAccountKey),
		getEnvVarFromConfigmap("OCTARINE_DOMAIN", DataPlaneConfigmapClusterKey),
		getEnvVarFromSecret("OCTARINE_ACCESS_TOKEN", accessKeySecretName),
		getEnvVarFromConfigmap("OCTARINE_API_SCHEME", DataPlaneConfigmapApiSchemeKey),
		getEnvVarFromConfigmap("OCTARINE_API_HOST", DataPlaneConfigmapApiHostKey),
		getEnvVarFromConfigmap("OCTARINE_API_PORT", DataPlaneConfigmapApiPortKey),
		getEnvVarFromConfigmap("OCTARINE_API_ADAPTER_NAME", DataPlaneConfigmapApiAdapterKey),
	}
}

func getEnvVarFromSecret(envName, accessKeySecretName string) coreV1.EnvVar {
	return coreV1.EnvVar{
		Name: envName,
		ValueFrom: &coreV1.EnvVarSource{
			SecretKeyRef: &coreV1.SecretKeySelector{
				LocalObjectReference: coreV1.LocalObjectReference{Name: accessKeySecretName},
				Key:                  AccessTokenSecretKeyName,
			},
		},
	}
}

func getEnvVarFromConfigmap(envName, configKeyName string) coreV1.EnvVar {
	return coreV1.EnvVar{
		Name: envName,
		ValueFrom: &coreV1.EnvVarSource{
			ConfigMapKeyRef: &coreV1.ConfigMapKeySelector{
				LocalObjectReference: coreV1.LocalObjectReference{Name: DataPlaneConfigmapName},
				Key:                  configKeyName,
			},
		},
	}
}
