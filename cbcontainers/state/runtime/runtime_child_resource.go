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
	MutateRuntimeChildK8sObject(k8sObject client.Object, cbContainersRuntimeSpec *cbcontainersv1.CBContainersRuntimeSpec, agentVersion, accessTokenSecretName string) error
	RuntimeChildNamespacedName(cbContainersRuntime *cbcontainersv1.CBContainersRuntimeSpec) types.NamespacedName
	stateTypes.DesiredK8sObjectInitializer
}

type DefaultRuntimeChildK8sObjectApplier struct{}

func NewDefaultRuntimeChildK8sObjectApplier() *DefaultRuntimeChildK8sObjectApplier {
	return &DefaultRuntimeChildK8sObjectApplier{}
}

func (applier *DefaultRuntimeChildK8sObjectApplier) ApplyRuntimeChildK8sObject(ctx context.Context, cbContainersRuntime *cbcontainersv1.CBContainersRuntimeSpec, agentVersion, accessTokenSecretName string, client client.Client, runtimeChildK8sObject RuntimeChildK8sObject, applyOptionsList ...*applymentOptions.ApplyOptions) (bool, client.Object, error) {
	runtimeChildWrapper := NewCBContainersRuntimeChildK8sObject(cbContainersRuntime, runtimeChildK8sObject, agentVersion, accessTokenSecretName)
	return applyment.ApplyDesiredK8sObject(ctx, client, runtimeChildWrapper, applyOptionsList...)
}

func (applier *DefaultRuntimeChildK8sObjectApplier) DeleteK8sObjectIfExists(ctx context.Context, cbContainersRuntime *cbcontainersv1.CBContainersRuntimeSpec, client client.Client, runtimeChildK8sObject RuntimeChildK8sObject) (bool, error) {
	runtimeChildWrapper := NewCBContainersRuntimeChildK8sObject(cbContainersRuntime, runtimeChildK8sObject, "", "")
	return applyment.DeleteK8sObjectIfExists(ctx, client, runtimeChildWrapper)
}

type CBContainersRuntimeChildK8sObject struct {
	cbContainersRuntimeSpec *cbcontainersv1.CBContainersRuntimeSpec
	RuntimeChildK8sObject
	AgentVersion          string
	AccessTokenSecretName string
}

func NewCBContainersRuntimeChildK8sObject(cbContainersRuntimeSpec *cbcontainersv1.CBContainersRuntimeSpec, runtimeChildK8sObject RuntimeChildK8sObject, agentVersion, accessTokenSecretName string) *CBContainersRuntimeChildK8sObject {
	return &CBContainersRuntimeChildK8sObject{
		cbContainersRuntimeSpec: cbContainersRuntimeSpec,
		RuntimeChildK8sObject:   runtimeChildK8sObject,
		AgentVersion:            agentVersion,
		AccessTokenSecretName:   accessTokenSecretName,
	}
}

func (runtimeChildWrapper *CBContainersRuntimeChildK8sObject) NamespacedName() types.NamespacedName {
	return runtimeChildWrapper.RuntimeChildNamespacedName(runtimeChildWrapper.cbContainersRuntimeSpec)
}

func (runtimeChildWrapper *CBContainersRuntimeChildK8sObject) MutateK8sObject(k8sObject client.Object) error {
	return runtimeChildWrapper.MutateRuntimeChildK8sObject(k8sObject, runtimeChildWrapper.cbContainersRuntimeSpec, runtimeChildWrapper.AgentVersion, runtimeChildWrapper.AccessTokenSecretName)
}
