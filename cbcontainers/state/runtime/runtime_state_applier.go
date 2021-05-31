package runtime

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	runtimeObjects "github.com/vmware/cbcontainers-operator/cbcontainers/state/runtime/objects"
)

type RuntimeChildK8sObjectApplier interface {
	ApplyRuntimeChildK8sObject(ctx context.Context, cbContainersRuntime *cbcontainersv1.CBContainersRuntime, client client.Client, runtimeChildK8sObject RuntimeChildK8sObject, applyOptionsList ...*applymentOptions.ApplyOptions) (bool, client.Object, error)
	DeleteK8sObjectIfExists(ctx context.Context, cbContainersRuntime *cbcontainersv1.CBContainersRuntime, client client.Client, runtimeChildK8sObject RuntimeChildK8sObject) (bool, error)
}

type RuntimeStateApplier struct {
	resolverDeployment *runtimeObjects.ResolverDeploymentK8sObject
	resolverService    *runtimeObjects.ResolverServiceK8sObject
	sensorDaemonSet    *runtimeObjects.SensorDaemonSetK8sObject
	childApplier       RuntimeChildK8sObjectApplier
	log                logr.Logger
}

func NewRuntimeStateApplier(log logr.Logger, childApplier RuntimeChildK8sObjectApplier) *RuntimeStateApplier {
	return &RuntimeStateApplier{
		resolverDeployment: runtimeObjects.NewResolverDeploymentK8sObject(),
		resolverService:    runtimeObjects.NewResolverServiceK8sObject(),
		sensorDaemonSet:    runtimeObjects.NewSensorDaemonSetK8sObject(),
		childApplier:       childApplier,
		log:                log,
	}
}

func (c *RuntimeStateApplier) ApplyDesiredState(ctx context.Context, cbContainersRuntime *cbcontainersv1.CBContainersRuntime, client client.Client, setOwner applymentOptions.OwnerSetter) (bool, error) {
	applyOptions := applymentOptions.NewApplyOptions().SetOwnerSetter(setOwner)
	mutatedResolver, err := c.applyResolver(ctx, cbContainersRuntime, client, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied runtime kubernetes resolver objects", "Mutated", mutatedResolver)

	return mutatedResolver, nil
}

func (c *RuntimeStateApplier) applyResolver(ctx context.Context, cbContainersRuntime *cbcontainersv1.CBContainersRuntime, client client.Client, applyOptions *applymentOptions.ApplyOptions) (bool, error) {
	mutatedService, _, err := c.childApplier.ApplyRuntimeChildK8sObject(ctx, cbContainersRuntime, client, c.resolverService, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied kubernetes resolver service", "Mutated", mutatedService)

	mutatedDeployment, _, err := c.childApplier.ApplyRuntimeChildK8sObject(ctx, cbContainersRuntime, client, c.resolverDeployment, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied runtime kubernetes resolver deployment", "Mutated", mutatedDeployment)
	return mutatedService || mutatedDeployment, nil
}

func (c *RuntimeStateApplier) applySensor(ctx context.Context, cbContainersRuntime *cbcontainersv1.CBContainersRuntime, client client.Client, applyOptions *applymentOptions.ApplyOptions) (bool, error) {
	mutatedDaemonSet, _, err := c.childApplier.ApplyRuntimeChildK8sObject(ctx, cbContainersRuntime, client, c.sensorDaemonSet, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied runtime kubernetes sensor daemon set", "Mutated", mutatedDaemonSet)
	return mutatedDaemonSet, nil
}
