package runtime

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	stateTypes "github.com/vmware/cbcontainers-operator/cbcontainers/state/types"
)

type RuntimeChildK8sObject interface {
	MutateRuntimeChildK8sObject(k8sObject client.Object, cbContainersRuntime *cbcontainersv1.CBContainersRuntime) error
	RuntimeChildNamespacedName(cbContainersRuntime *cbcontainersv1.CBContainersRuntime) types.NamespacedName
	stateTypes.DesiredK8sObjectInitializer
}

type DefaultRuntimeChildK8sObjectApplier struct{}

func NewDefaultRuntimeChildK8sObjectApplier() *DefaultRuntimeChildK8sObjectApplier {
	return &DefaultRuntimeChildK8sObjectApplier{}
}

func (applier *DefaultRuntimeChildK8sObjectApplier) ApplyRuntimeChildK8sObject(ctx context.Context, cbContainersRuntime *cbcontainersv1.CBContainersRuntime, client client.Client, runtimeChildK8sObject RuntimeChildK8sObject, applyOptionsList ...*applymentOptions.ApplyOptions) (bool, client.Object, error) {
	runtimeChildWrapper := NewCBContainersRuntimeChildK8sObject(cbContainersRuntime, runtimeChildK8sObject)
	return applyment.ApplyDesiredK8sObject(ctx, client, runtimeChildWrapper, applyOptionsList...)
}

func (applier *DefaultRuntimeChildK8sObjectApplier) DeleteK8sObjectIfExists(ctx context.Context, cbContainersRuntime *cbcontainersv1.CBContainersRuntime, client client.Client, runtimeChildK8sObject RuntimeChildK8sObject) (bool, error) {
	runtimeChildWrapper := NewCBContainersRuntimeChildK8sObject(cbContainersRuntime, runtimeChildK8sObject)
	return applyment.DeleteK8sObjectIfExists(ctx, client, runtimeChildWrapper)
}

type CBContainersRuntimeChildK8sObject struct {
	cbContainersRuntime *cbcontainersv1.CBContainersRuntime
	RuntimeChildK8sObject
}

func NewCBContainersRuntimeChildK8sObject(cbContainersRuntime *cbcontainersv1.CBContainersRuntime, runtimeChildK8sObject RuntimeChildK8sObject) *CBContainersRuntimeChildK8sObject {
	return &CBContainersRuntimeChildK8sObject{
		cbContainersRuntime:   cbContainersRuntime,
		RuntimeChildK8sObject: runtimeChildK8sObject,
	}
}

func (runtimeChildWrapper *CBContainersRuntimeChildK8sObject) NamespacedName() types.NamespacedName {
	return runtimeChildWrapper.RuntimeChildNamespacedName(runtimeChildWrapper.cbContainersRuntime)
}

func (runtimeChildWrapper *CBContainersRuntimeChildK8sObject) MutateK8sObject(k8sObject client.Object) error {
	return runtimeChildWrapper.MutateRuntimeChildK8sObject(k8sObject, runtimeChildWrapper.cbContainersRuntime)
}
