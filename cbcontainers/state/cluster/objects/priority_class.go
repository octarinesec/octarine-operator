package objects

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	clusterStateUtils "github.com/vmware/cbcontainers-operator/cbcontainers/state/cluster/utils"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	schedulingV1 "k8s.io/api/scheduling/v1"
	schedulingV1alpha1 "k8s.io/api/scheduling/v1alpha1"
	schedulingV1beta1 "k8s.io/api/scheduling/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PriorityClassValue         = 999999999
	PriorityClassGlobalDefault = false
	PriorityClassDescription   = "This priority class should be used only for CBContainers pods."
)

type PriorityClassK8sObject struct {
	kubeletVersion string
}

func NewPriorityClassK8sObject() *PriorityClassK8sObject { return &PriorityClassK8sObject{} }

func (obj *PriorityClassK8sObject) UpdateKubeletVersion(kubeletVersion string) {
	obj.kubeletVersion = kubeletVersion
}

func (obj *PriorityClassK8sObject) EmptyK8sObject() client.Object {
	if obj.kubeletVersion == "" || obj.kubeletVersion >= "v1.14" {
		return &schedulingV1.PriorityClass{}
	} else if obj.kubeletVersion >= "v1.11" {
		return &schedulingV1beta1.PriorityClass{}
	}

	return &schedulingV1alpha1.PriorityClass{}
}

func (obj *PriorityClassK8sObject) ClusterChildNamespacedName(_ *cbcontainersv1.CBContainersCluster) types.NamespacedName {
	return types.NamespacedName{Name: commonState.DataPlanePriorityClassName, Namespace: ""}
}

func (obj *PriorityClassK8sObject) MutateClusterChildK8sObject(k8sObject client.Object, _ *cbcontainersv1.CBContainersCluster) error {
	priorityClassSetter, ok := clusterStateUtils.GetPriorityClassSetter(k8sObject)
	if !ok {
		return fmt.Errorf("expected PriorityClass setter K8s object")
	}

	priorityClassSetter.SetValue(PriorityClassValue)
	priorityClassSetter.SetGlobalDefault(PriorityClassGlobalDefault)
	priorityClassSetter.SetDescription(PriorityClassDescription)

	return nil
}
