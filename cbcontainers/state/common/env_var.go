package common

import (
	"reflect"
	"strconv"
	"strings"

	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	eventGatewayHostVarName = "OCTARINE_MESSAGEPROXY_HOST"
	eventGatewayPortVarName = "OCTARINE_MESSAGEPROXY_PORT"
	accountVarName          = "OCTARINE_ACCOUNT"
	clusterVarName          = "OCTARINE_DOMAIN"
	clusterIDVarName        = "OCTARINE_CLUSTER_ID"
	accessTokenVarName      = "OCTARINE_ACCESS_TOKEN"
	apiSchemeVarName        = "OCTARINE_API_SCHEME"
	apiHostVarName          = "OCTARINE_API_HOST"
	apiPortVarName          = "OCTARINE_API_PORT"
	apiAdapterVarName       = "OCTARINE_API_ADAPTER_NAME"
	agentVersionVarName     = "OCTARINE_AGENT_VERSION"
	tlsSkipVerifyVarName    = "TLS_INSECURE_SKIP_VERIFY"
	tlsRootCAsPathVarName   = "TLS_ROOT_CAS_PATH"
	proxyNoProxyVarName     = "NO_PROXY"
	proxyHttpProxyVarName   = "HTTP_PROXY"
	proxyHttpsProxyVarName  = "HTTPS_PROXY"
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

func (b *EnvVarBuilder) WithGatewayTLS() *EnvVarBuilder {
	return b.WithEnvVarFromConfigmap(tlsSkipVerifyVarName, DataPlaneConfigmapTlsSkipVerifyKey).
		WithEnvVarFromConfigmap(tlsRootCAsPathVarName, DataPlaneConfigmapTlsRootCAsPathKey)
}

func (b *EnvVarBuilder) WithCustom(customEnvsToAdd ...coreV1.EnvVar) *EnvVarBuilder {
	for _, customEnvVar := range customEnvsToAdd {
		b.envVars[customEnvVar.Name] = customEnvVar
	}

	return b
}

func (b *EnvVarBuilder) WithProxySettings(proxySettings *cbcontainersv1.CBContainersProxySettings) *EnvVarBuilder {
	if proxySettings == nil || proxySettings.Enabled == nil || *proxySettings.Enabled == false {
		return b
	}

	var userNoProxy string
	if noProxyVar, ok := b.envVars[proxyNoProxyVarName]; ok {
		userNoProxy = noProxyVar.Value
	} else if proxySettings.NoProxy != nil {
		userNoProxy = *proxySettings.NoProxy
	}

	userNoProxy = strings.Trim(userNoProxy, ",")
	cbNoProxy := strings.Trim(*proxySettings.NoProxySuffix, ",")

	noProxyVal := strings.Join([]string{userNoProxy, cbNoProxy}, ",")
	// if one of userNoProxy or cbNoProxy is empty, we'll get a leading or trailing comma
	noProxyVal = strings.Trim(noProxyVal, ",") // thus, we strip it here
	b.envVars[proxyNoProxyVarName] = coreV1.EnvVar{Name: proxyNoProxyVarName, Value: noProxyVal}

	b.withDefaultValueOrNothing(proxyHttpProxyVarName, proxySettings.HttpProxy).
		withDefaultValueOrNothing(proxyHttpsProxyVarName, proxySettings.HttpsProxy)

	return b
}

func (b *EnvVarBuilder) withDefaultValueOrNothing(envName string, defValue *string) *EnvVarBuilder {
	var value string
	if envVar, ok := b.envVars[envName]; ok {
		value = envVar.Value
	} else if defValue != nil {
		value = *defValue
	}

	b.envVars[envName] = coreV1.EnvVar{Name: envName, Value: value}
	return b
}

func (b *EnvVarBuilder) WithEnvVarFromResource(envName, containerName, resourcePath string) *EnvVarBuilder {
	envVar := coreV1.EnvVar{
		Name: envName,
		ValueFrom: &coreV1.EnvVarSource{
			ResourceFieldRef: &coreV1.ResourceFieldSelector{
				ContainerName: containerName,
				Resource:      resourcePath,
				Divisor:       resource.Quantity{Format: resource.DecimalSI},
			},
		},
	}
	b.envVars[envName] = envVar

	return b
}

func (b *EnvVarBuilder) WithEnvVarFromField(envName, fieldPath, apiVersion string) *EnvVarBuilder {
	envVar := coreV1.EnvVar{
		Name: envName,
		ValueFrom: &coreV1.EnvVarSource{
			FieldRef: &coreV1.ObjectFieldSelector{
				FieldPath:  fieldPath,
				APIVersion: apiVersion,
			},
		},
	}
	b.envVars[envName] = envVar

	return b
}

func (b *EnvVarBuilder) WithEnvVarFromSecret(envName, accessKeySecretName, key string) *EnvVarBuilder {
	envVar := coreV1.EnvVar{
		Name: envName,
		ValueFrom: &coreV1.EnvVarSource{
			SecretKeyRef: &coreV1.SecretKeySelector{
				LocalObjectReference: coreV1.LocalObjectReference{Name: accessKeySecretName},
				Key:                  key,
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
	return b.WithEnvVarFromSecret(accessTokenVarName, accessKeySecretName, AccessTokenSecretKeyName).
		WithEnvVarFromConfigmap(accountVarName, DataPlaneConfigmapAccountKey).
		WithEnvVarFromConfigmap(clusterVarName, DataPlaneConfigmapClusterKey).
		WithEnvVarFromConfigmap(clusterIDVarName, DataPlaneConfigmapClusterIDKey).
		WithEnvVarFromConfigmap(apiSchemeVarName, DataPlaneConfigmapApiSchemeKey).
		WithEnvVarFromConfigmap(apiHostVarName, DataPlaneConfigmapApiHostKey).
		WithEnvVarFromConfigmap(apiPortVarName, DataPlaneConfigmapApiPortKey).
		WithEnvVarFromConfigmap(apiAdapterVarName, DataPlaneConfigmapApiAdapterKey).
		WithEnvVarFromConfigmap(agentVersionVarName, DataPlaneConfigmapAgentVersionKey).
		WithGatewayTLS()
}

func (b *EnvVarBuilder) IsEqual(actualEnv []coreV1.EnvVar) bool {
	if len(actualEnv) != len(b.envVars) {
		return false
	}

	for _, actualEnvVar := range actualEnv {
		desiredEnvVar, ok := b.envVars[actualEnvVar.Name]
		if !ok {
			return false
		}

		if desiredEnvVar.ValueFrom != nil && desiredEnvVar.ValueFrom.ResourceFieldRef != nil {
			if !b.isResourceFieldRefEquals(desiredEnvVar.ValueFrom.ResourceFieldRef, actualEnvVar.ValueFrom.ResourceFieldRef) {
				return false
			}
		} else if !reflect.DeepEqual(actualEnvVar, desiredEnvVar) {
			return false
		}
	}

	return true
}

func (b *EnvVarBuilder) isResourceFieldRefEquals(desiredResourceFieldRef, actualResourceFieldRef *coreV1.ResourceFieldSelector) bool {
	if desiredResourceFieldRef.ContainerName != actualResourceFieldRef.ContainerName {
		return false
	}
	if desiredResourceFieldRef.Resource != actualResourceFieldRef.Resource {
		return false
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
