package hardening

import operatorcontainerscarbonblackiov1 "github.com/vmware/cbcontainers-operator/api/v1"

type CBContainersHardeningChildK8sObject struct {
	cbContainersHardening *operatorcontainerscarbonblackiov1.CBContainersHardening
}

func (obj *CBContainersHardeningChildK8sObject) UpdateCbContainersHardening(cbContainersHardening *operatorcontainerscarbonblackiov1.CBContainersHardening) {
	obj.cbContainersHardening = cbContainersHardening
}
