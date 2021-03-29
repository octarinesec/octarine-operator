package hardening

import (
	"context"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CBContainersClusterStateApplier struct {
	enforcerDeployment *EnforcerK8sObject
}

func NewHardeningStateApplier() *CBContainersClusterStateApplier {
	return &CBContainersClusterStateApplier{
		enforcerDeployment: NewEnforcerDeploymentK8sObject(),
	}
}

func (c *CBContainersClusterStateApplier) ApplyDesiredState(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client, setOwner applyment.OwnerSetter) (bool, error) {
	c.enforcerDeployment.UpdateCbContainersHardening(cbContainersHardening)
	ok, err := applyment.ApplyDesiredK8sObject(ctx, client, c.enforcerDeployment, setOwner)
	if err != nil {
		return false, err
	}

	return ok, nil
}
