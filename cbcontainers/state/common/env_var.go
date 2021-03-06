package common

import (
	"reflect"
	"strconv"

	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	coreV1 "k8s.io/api/core/v1"
)

const (
	eventGatewayHostVarName = "OCTARINE_MESSAGEPROXY_HOST"
	eventGatewayPortVarName = "OCTARINE_MESSAGEPROXY_PORT"
	accountVarName          = "OCTARINE_ACCOUNT"
	clusterVarName          = "OCTARINE_DOMAIN"
	accessTokenVarName      = "OCTARINE_ACCESS_TOKEN"
	apiSchemeVarName        = "OCTARINE_API_SCHEME"
	apiHostVarName          = "OCTARINE_API_HOST"
	apiPortVarName          = "OCTARINE_API_PORT"
	apiAdapterVarName       = "OCTARINE_API_ADAPTER_NAME"
)

type EnvVarBuilder struct {
	envVars map[string]coreV1.EnvVar
}

func NewEnvVarBuilder() *EnvVarBuilder {
	return &EnvVarBuilder{envVars: make(map[string]coreV1.EnvVar, 0)}
}

// WithSpec Must be the last builder - to override all the predefined env vars
func (b *EnvVarBuilder) WithSpec(desiredEnvsValues map[string]string) *EnvVarBuilder {
	for desiredEnvVarName, desiredEnvVarValue := range desiredEnvsValues {
		b.envVars[desiredEnvVarName] = coreV1.EnvVar{Name: desiredEnvVarName, Value: desiredEnvVarValue}
	}

	return b
}

func (b *EnvVarBuilder) WithEventsGateway(eventsGatewaySpec *cbcontainersv1.CBContainersEventsGatewaySpec) *EnvVarBuilder {
	b.envVars[eventGatewayHostVarName] = coreV1.EnvVar{Name: eventGatewayHostVarName, Value: eventsGatewaySpec.Host}
	b.envVars[eventGatewayPortVarName] = coreV1.EnvVar{Name: eventGatewayPortVarName, Value: strconv.Itoa(eventsGatewaySpec.Port)}

	return b
}

func (b *EnvVarBuilder) WithCustom(customEnvsToAdd ...coreV1.EnvVar) *EnvVarBuilder {
	for _, customEnvVar := range customEnvsToAdd {
		b.envVars[customEnvVar.Name] = customEnvVar
	}

	return b
}

func (b *EnvVarBuilder) WithEnvVarFromSecret(envName, accessKeySecretName string) *EnvVarBuilder {
	envVar := coreV1.EnvVar{
		Name: envName,
		ValueFrom: &coreV1.EnvVarSource{
			SecretKeyRef: &coreV1.SecretKeySelector{
				LocalObjectReference: coreV1.LocalObjectReference{Name: accessKeySecretName},
				Key:                  AccessTokenSecretKeyName,
			},
		},
	}
	b.envVars[envName] = envVar

	return b
}

func (b *EnvVarBuilder) WithEnvVarFromConfigmap(envName, configKeyName string) *EnvVarBuilder {
	envVar := coreV1.EnvVar{
		Name: envName,
		ValueFrom: &coreV1.EnvVarSource{
			ConfigMapKeyRef: &coreV1.ConfigMapKeySelector{
				LocalObjectReference: coreV1.LocalObjectReference{Name: DataPlaneConfigmapName},
				Key:                  configKeyName,
			},
		},
	}
	b.envVars[envName] = envVar

	return b
}

func (b *EnvVarBuilder) WithCommonDataPlane(accessKeySecretName string) *EnvVarBuilder {
	return b.WithEnvVarFromSecret(accessTokenVarName, accessKeySecretName).
		WithEnvVarFromConfigmap(accountVarName, DataPlaneConfigmapAccountKey).
		WithEnvVarFromConfigmap(clusterVarName, DataPlaneConfigmapClusterKey).
		WithEnvVarFromConfigmap(apiSchemeVarName, DataPlaneConfigmapApiSchemeKey).
		WithEnvVarFromConfigmap(apiHostVarName, DataPlaneConfigmapApiHostKey).
		WithEnvVarFromConfigmap(apiPortVarName, DataPlaneConfigmapApiPortKey).
		WithEnvVarFromConfigmap(apiAdapterVarName, DataPlaneConfigmapApiAdapterKey)
}

func (b *EnvVarBuilder) IsEqual(actualEnv []coreV1.EnvVar) bool {
	if len(actualEnv) != len(b.envVars) {
		return false
	}

	for _, actualEnvVar := range actualEnv {
		desiredEnvVar, ok := b.envVars[actualEnvVar.Name]
		if !ok || !reflect.DeepEqual(actualEnvVar, desiredEnvVar) {
			return false
		}
	}

	return true
}

func (b *EnvVarBuilder) Build() []coreV1.EnvVar {
	envVarsToReturn := make([]coreV1.EnvVar, 0, len(b.envVars))
	for _, desiredEnvVar := range b.envVars {
		envVarsToReturn = append(envVarsToReturn, desiredEnvVar)
	}

	return envVarsToReturn
}

func MutateEnvVars(container *coreV1.Container, builder *EnvVarBuilder) {
	if builder.IsEqual(container.Env) {
		return
	}

	container.Env = builder.Build()
}
