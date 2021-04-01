package cluster

import (
	"fmt"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

type ConfigurationK8sObject struct {
	CBContainersClusterChildK8sObject
}

func NewConfigurationK8sObject() *ConfigurationK8sObject { return &ConfigurationK8sObject{} }

func (obj *ConfigurationK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: commonState.DataPlaneConfigmapName, Namespace: obj.cbContainersCluster.Namespace}
}

func (obj *ConfigurationK8sObject) EmptyK8sObject() client.Object { return &v1.ConfigMap{} }

func (obj *ConfigurationK8sObject) MutateK8sObject(k8sObject client.Object) error {
	configMap, ok := k8sObject.(*v1.ConfigMap)
	if !ok {
		return fmt.Errorf("expected ConfigMap K8s object")
	}

	configMap.Data = map[string]string{
		commonState.DataPlaneConfigmapAccountKey:    obj.cbContainersCluster.Spec.Account,
		commonState.DataPlaneConfigmapClusterKey:    obj.cbContainersCluster.Spec.ClusterName,
		commonState.DataPlaneConfigmapApiSchemeKey:  obj.cbContainersCluster.Spec.ApiGatewaySpec.Scheme,
		commonState.DataPlaneConfigmapApiHostKey:    obj.cbContainersCluster.Spec.ApiGatewaySpec.Host,
		commonState.DataPlaneConfigmapApiPortKey:    strconv.Itoa(obj.cbContainersCluster.Spec.ApiGatewaySpec.Port),
		commonState.DataPlaneConfigmapApiAdapterKey: obj.cbContainersCluster.Spec.ApiGatewaySpec.Adapter,
	}

	return nil
}
