package cluster

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DataAccountKey = "Account"
	DataClusterKey = "Cluster"
)

type ConfigurationK8sObject struct {
	CBContainersClusterChildK8sObject
}

func NewConfigurationK8sObject() *ConfigurationK8sObject { return &ConfigurationK8sObject{} }

func (obj *ConfigurationK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: obj.cbContainersCluster.Name, Namespace: obj.cbContainersCluster.Namespace}
}

func (obj *ConfigurationK8sObject) EmptyK8sObject() client.Object { return &v1.ConfigMap{} }

func (obj *ConfigurationK8sObject) MutateK8sObject(k8sObject client.Object) error {
	configMap, ok := k8sObject.(*v1.ConfigMap)
	if !ok {
		return fmt.Errorf("expected ConfigMap K8s object")
	}

	configMap.Data = map[string]string{
		DataAccountKey: obj.cbContainersCluster.Spec.Account,
		DataClusterKey: obj.cbContainersCluster.Spec.ClusterName,
	}

	return nil
}
