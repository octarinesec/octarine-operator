package objects

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

type ConfigurationK8sObject struct{}

func NewConfigurationK8sObject() *ConfigurationK8sObject { return &ConfigurationK8sObject{} }

func (obj *ConfigurationK8sObject) EmptyK8sObject() client.Object { return &v1.ConfigMap{} }

func (obj *ConfigurationK8sObject) ClusterChildNamespacedName(_ *cbcontainersv1.CBContainersAgent) types.NamespacedName {
	return types.NamespacedName{Name: commonState.DataPlaneConfigmapName, Namespace: commonState.DataPlaneNamespaceName}
}

func (obj *ConfigurationK8sObject) MutateClusterChildK8sObject(k8sObject client.Object, cbContainersCluster *cbcontainersv1.CBContainersAgent) error {
	configMap, ok := k8sObject.(*v1.ConfigMap)
	if !ok {
		return fmt.Errorf("expected ConfigMap K8s object")
	}

	configMap.Data = map[string]string{
		commonState.DataPlaneConfigmapAccountKey:    cbContainersCluster.Spec.Account,
		commonState.DataPlaneConfigmapClusterKey:    cbContainersCluster.Spec.ClusterName,
		commonState.DataPlaneConfigmapApiSchemeKey:  cbContainersCluster.Spec.ApiGatewaySpec.Scheme,
		commonState.DataPlaneConfigmapApiHostKey:    cbContainersCluster.Spec.ApiGatewaySpec.Host,
		commonState.DataPlaneConfigmapApiPortKey:    strconv.Itoa(cbContainersCluster.Spec.ApiGatewaySpec.Port),
		commonState.DataPlaneConfigmapApiAdapterKey: cbContainersCluster.Spec.ApiGatewaySpec.Adapter,
		commonState.DataPlaneConfigmapTlsSkipVerifyKey: strconv.FormatBool(cbContainersCluster.Spec.GatewayTLS.InsecureSkipVerify),
		commonState.DataPlaneConfigmapTlsRootCAsPathKey: path.Join(commonState.DataPlaneConfigmapTlsRootCAsDirPath, commonState.DataPlaneConfigmapTlsRootCAsFilePath),
		commonState.DataPlaneConfigmapTlsRootCAsFilePath: string(cbContainersCluster.Spec.GatewayTLS.RootCAsBundle),
	}

	return nil
}
