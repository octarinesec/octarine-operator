package hardening

import (
	"context"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type tlsSecretsValuesCreator interface {
	CreateTlsSecretsValues(resourceNamespacedName types.NamespacedName) (models.TlsSecretValues, error)
}

type CBContainersClusterStateApplier struct {
	enforcerTlsSecret  *EnforcerTlsK8sObject
	enforcerDeployment *EnforcerK8sObject
}

func NewHardeningStateApplier(tlsSecretsValuesCreator tlsSecretsValuesCreator) *CBContainersClusterStateApplier {
	return &CBContainersClusterStateApplier{
		enforcerTlsSecret:  NewEnforcerTlsK8sObject(tlsSecretsValuesCreator),
		enforcerDeployment: NewEnforcerDeploymentK8sObject(),
	}
}

func (c *CBContainersClusterStateApplier) ApplyDesiredState(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client, setOwner applymentOptions.OwnerSetter) (bool, error) {
	applyOptions := applymentOptions.NewApplyOptions().SetOwnerSetter(setOwner)

	c.enforcerTlsSecret.UpdateCbContainersHardening(cbContainersHardening)
	mutatedSecret, err := applyment.ApplyDesiredK8sObject(ctx, client, c.enforcerTlsSecret, applyOptions, applymentOptions.NewApplyOptions().SetCreateOnly(true))
	if err != nil {
		return false, err
	}

	c.enforcerDeployment.UpdateCbContainersHardening(cbContainersHardening)
	mutatedDeployment, err := applyment.ApplyDesiredK8sObject(ctx, client, c.enforcerDeployment, applyOptions)
	if err != nil {
		return false, err
	}

	return mutatedSecret || mutatedDeployment, nil
}
