package hardening

import (
	"context"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	hardeningObjects "github.com/vmware/cbcontainers-operator/cbcontainers/state/hardening/objects"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CBContainersClusterStateApplier struct {
	enforcerTlsSecret  *hardeningObjects.EnforcerTlsK8sObject
	enforcerDeployment *hardeningObjects.EnforcerK8sObject
	enforcerService    *hardeningObjects.EnforcerServiceK8sObject
}

func NewHardeningStateApplier(tlsSecretsValuesCreator hardeningObjects.TlsSecretsValuesCreator) *CBContainersClusterStateApplier {
	return &CBContainersClusterStateApplier{
		enforcerTlsSecret:  hardeningObjects.NewEnforcerTlsK8sObject(tlsSecretsValuesCreator),
		enforcerDeployment: hardeningObjects.NewEnforcerDeploymentK8sObject(),
		enforcerService:    hardeningObjects.NewEnforcerServiceK8sObject(),
	}
}

func (c *CBContainersClusterStateApplier) ApplyDesiredState(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client, setOwner applymentOptions.OwnerSetter) (bool, error) {
	applyOptions := applymentOptions.NewApplyOptions().SetOwnerSetter(setOwner)

	mutatedSecret, err := ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.enforcerTlsSecret, applyOptions, applymentOptions.NewApplyOptions().SetCreateOnly(true))
	if err != nil {
		return false, err
	}

	mutatedDeployment, err := ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.enforcerDeployment, applyOptions)
	if err != nil {
		return false, err
	}

	mutatedService, err := ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.enforcerService, applyOptions)
	if err != nil {
		return false, err
	}

	return mutatedSecret || mutatedDeployment || mutatedService, nil
}
