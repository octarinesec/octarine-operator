package components

import (
	"fmt"

	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/adapters"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
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

func NewPriorityClassK8sObject(kubeletVersion string) *PriorityClassK8sObject {
	return &PriorityClassK8sObject{
		kubeletVersion: kubeletVersion,
	}
}

func (obj *PriorityClassK8sObject) EmptyK8sObject() client.Object {
	return adapters.EmptyPriorityClassForVersion(obj.kubeletVersion)
}

func (obj *PriorityClassK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: commonState.DataPlanePriorityClassName, Namespace: ""}
}

func (obj *PriorityClassK8sObject) MutateK8sObject(k8sObject client.Object, _ *cbcontainersv1.CBContainersAgentSpec) error {
	priorityClassSetter, ok := adapters.GetPriorityClassSetter(k8sObject)
	if !ok {
		return fmt.Errorf("expected PriorityClass setter K8s object")
	}

	priorityClassSetter.SetValue(PriorityClassValue)
	priorityClassSetter.SetGlobalDefault(PriorityClassGlobalDefault)
	priorityClassSetter.SetDescription(PriorityClassDescription)

	return nil
}
