package monitor

import (
	"context"
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DefaultFeaturesStatusProvider struct {
	client client.Client
}

func NewDefaultFeaturesStatusProvider(client client.Client) *DefaultFeaturesStatusProvider {
	return &DefaultFeaturesStatusProvider{
		client: client,
	}
}

func (provider DefaultFeaturesStatusProvider) HardeningEnabled() (bool, error) {
	cbContainersHardeningList := &cbcontainersv1.CBContainersHardeningList{}
	if err := provider.client.List(context.Background(), cbContainersHardeningList); err != nil {
		return false, fmt.Errorf("couldn't find CBContainersHardening k8s object: %v", err)
	}

	if cbContainersHardeningList.Items == nil {
		return false, fmt.Errorf("couldn't find CBContainersHardening k8s object")
	}

	return len(cbContainersHardeningList.Items) > 0, nil
}

func (provider DefaultFeaturesStatusProvider) RuntimeEnabled() (bool, error) {
	return false, nil
}
