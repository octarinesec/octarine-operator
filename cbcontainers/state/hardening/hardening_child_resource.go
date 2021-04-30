package hardening

import (
	"context"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	stateTypes "github.com/vmware/cbcontainers-operator/cbcontainers/state/types"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type hardeningChildK8sObject interface {
	MutateHardeningChildK8sObject(k8sObject client.Object, cbContainersHardening *cbcontainersv1.CBContainersHardening) error
	HardeningChildNamespacedName(cbContainersHardening *cbcontainersv1.CBContainersHardening) types.NamespacedName
	stateTypes.DesiredK8sObjectInitializer
}

func ApplyHardeningChildK8sObject(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client, hardeningChildK8sObject hardeningChildK8sObject, applyOptionsList ...*applymentOptions.ApplyOptions) (bool, client.Object, error) {
	hardeningChildWrapper := NewCBContainersHardeningChildK8sObject(cbContainersHardening, hardeningChildK8sObject)
	return applyment.ApplyDesiredK8sObject(ctx, client, hardeningChildWrapper, applyOptionsList...)
}

func DeleteK8sObjectIfExists(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client, hardeningChildK8sObject hardeningChildK8sObject) (bool, error) {
	hardeningChildWrapper := NewCBContainersHardeningChildK8sObject(cbContainersHardening, hardeningChildK8sObject)
	return applyment.DeleteK8sObjectIfExists(ctx, client, hardeningChildWrapper)
}

type CBContainersHardeningChildK8sObject struct {
	cbContainersHardening *cbcontainersv1.CBContainersHardening
	hardeningChildK8sObject
}

func NewCBContainersHardeningChildK8sObject(cbContainersHardening *cbcontainersv1.CBContainersHardening, hardeningChildK8sObject hardeningChildK8sObject) *CBContainersHardeningChildK8sObject {
	return &CBContainersHardeningChildK8sObject{
		cbContainersHardening:   cbContainersHardening,
		hardeningChildK8sObject: hardeningChildK8sObject,
	}
}

func (hardeningChildWrapper *CBContainersHardeningChildK8sObject) NamespacedName() types.NamespacedName {
	return hardeningChildWrapper.HardeningChildNamespacedName(hardeningChildWrapper.cbContainersHardening)
}

func (hardeningChildWrapper *CBContainersHardeningChildK8sObject) MutateK8sObject(k8sObject client.Object) error {
	return hardeningChildWrapper.MutateHardeningChildK8sObject(k8sObject, hardeningChildWrapper.cbContainersHardening)
}
