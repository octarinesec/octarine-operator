package applyment

import (
	"context"
	"fmt"
	stateTypes "github.com/vmware/cbcontainers-operator/cbcontainers/state/types"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ApplyDesiredK8sObject(ctx context.Context, client client.Client, desiredK8sObject stateTypes.DesiredK8sObject) (bool, error) {
	k8sObject := desiredK8sObject.EmptyK8sObject()
	namespacedName := desiredK8sObject.NamespacedName()
	foundErr := client.Get(ctx, namespacedName, k8sObject)

	if foundErr != nil && !errors.IsNotFound(foundErr) {
		return false, fmt.Errorf("failed getting K8s object: %v", foundErr)
	}

	mutated, mutateErr := desiredK8sObject.MutateK8sObject(k8sObject)
	if mutateErr != nil {
		return false, fmt.Errorf("failed mutating K8s object: %v", foundErr)
	}

	// k8s object was not found should, need to create
	if foundErr != nil {
		k8sObject.SetNamespace(namespacedName.Namespace)
		k8sObject.SetName(namespacedName.Name)
		if creationErr := client.Create(ctx, k8sObject); creationErr != nil {
			return false, fmt.Errorf("failed creating K8s object: %v", creationErr)
		}
		return true, nil
	}

	if !mutated {
		return false, nil
	}

	if updateErr := client.Update(ctx, k8sObject); updateErr != nil {
		return false, fmt.Errorf("failed updating exsiting K8s object: %v", updateErr)
	}

	return true, nil
}
