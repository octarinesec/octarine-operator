package components

import (
	"fmt"
	"path"
	"strconv"

	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ConfigurationK8sObject struct {
	// Namespace is the Namespace in which the ConfigMap will be created.
	Namespace string
}

func NewConfigurationK8sObject(namespace string) *ConfigurationK8sObject {
	return &ConfigurationK8sObject{
		Namespace: namespace,
	}
}

func (obj *ConfigurationK8sObject) EmptyK8sObject() client.Object { return &v1.ConfigMap{} }

func (obj *ConfigurationK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: commonState.DataPlaneConfigmapName, Namespace: obj.Namespace}
}

func (obj *ConfigurationK8sObject) MutateK8sObject(k8sObject client.Object, agentSpec *cbcontainersv1.CBContainersAgentSpec) error {
	configMap, ok := k8sObject.(*v1.ConfigMap)
	if !ok {
		return fmt.Errorf("expected ConfigMap K8s object")
	}

	configMap.Data = map[string]string{
		commonState.DataPlaneConfigmapAccountKey:            agentSpec.Account,
		commonState.DataPlaneConfigmapClusterKey:            agentSpec.ClusterName,
		commonState.DataPlaneConfigmapAgentVersionKey:       agentSpec.Version,
		commonState.DataPlaneConfigmapDataplaneNamespaceKey: obj.Namespace,
		commonState.DataPlaneConfigmapApiSchemeKey:          agentSpec.Gateways.ApiGateway.Scheme,
		commonState.DataPlaneConfigmapApiHostKey:            agentSpec.Gateways.ApiGateway.Host,
		commonState.DataPlaneConfigmapApiPortKey:            strconv.Itoa(agentSpec.Gateways.ApiGateway.Port),
		commonState.DataPlaneConfigmapApiAdapterKey:         agentSpec.Gateways.ApiGateway.Adapter,
		commonState.DataPlaneConfigmapTlsSkipVerifyKey:      strconv.FormatBool(agentSpec.Gateways.GatewayTLS.InsecureSkipVerify),
		commonState.DataPlaneConfigmapTlsRootCAsPathKey:     path.Join(commonState.DataPlaneConfigmapTlsRootCAsDirPath, commonState.DataPlaneConfigmapTlsRootCAsFilePath),
		commonState.DataPlaneConfigmapTlsRootCAsFilePath:    string(agentSpec.Gateways.GatewayTLS.RootCAsBundle),
	}

	return nil
}
