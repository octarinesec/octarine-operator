package runtime

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
)

type RuntimeStateApplier struct {

}

func NewRuntimeStateApplier() *RuntimeStateApplier {
	return &RuntimeStateApplier{

	}
}

func (c *RuntimeStateApplier) ApplyDesiredState(ctx context.Context, cbContainersRuntime *cbcontainersv1.CBContainersRuntime, client client.Client, setOwner applymentOptions.OwnerSetter) (bool, error) {
	return false, nil
}
