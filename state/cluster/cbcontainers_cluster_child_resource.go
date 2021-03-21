package cluster

import operatorcontainerscarbonblackiov1 "github.com/vmware/cbcontainers-operator/api/v1"

type CBContainersClusterChildK8sObject struct {
	cbContainersCluster *operatorcontainerscarbonblackiov1.CBContainersCluster
}

func (obj *CBContainersClusterChildK8sObject) UpdateCbContainersCluster(cbContainersCluster *operatorcontainerscarbonblackiov1.CBContainersCluster) {
	obj.cbContainersCluster = cbContainersCluster
}
