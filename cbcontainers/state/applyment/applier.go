package applyment

import (
	"context"
	"encoding/json"
	"fmt"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	stateTypes "github.com/vmware/cbcontainers-operator/cbcontainers/state/types"
	"k8s.io/apimachinery/pkg/api/errors"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ApplyDesiredK8sObject(ctx context.Context, client client.Client, desiredK8sObject stateTypes.DesiredK8sObject, applyOptionsList ...*applymentOptions.ApplyOptions) (bool, error) {
	applyOptions := applymentOptions.MergeApplyOptions(applyOptionsList...)

	k8sObject := desiredK8sObject.EmptyK8sObject()
	namespacedName := desiredK8sObject.NamespacedName()
	foundErr := client.Get(ctx, namespacedName, k8sObject)

	if foundErr != nil && !errors.IsNotFound(foundErr) {
		return false, fmt.Errorf("failed getting K8s object: %v", foundErr)
	}

	beforeMutationRaw, _ := json.Marshal(k8sObject)
	mutateErr := desiredK8sObject.MutateK8sObject(k8sObject)
	if mutateErr != nil {
		return false, fmt.Errorf("failed mutating K8s object `%v`: %v", namespacedName, mutateErr)
	}

	if setOwner := applyOptions.OwnerSetter(); setOwner != nil {
		if ownerSetterErr := setOwner(k8sObject); ownerSetterErr != nil {
			return false, fmt.Errorf("failed setting owner to K8s object `%v`: %v", namespacedName, ownerSetterErr)
		}
	}

	// k8s object was not found should, need to create
	if foundErr != nil {
		k8sObject.SetNamespace(namespacedName.Namespace)
		k8sObject.SetName(namespacedName.Name)

		if creationErr := client.Create(ctx, k8sObject); creationErr != nil {
			return false, fmt.Errorf("failed creating K8s object `%v`: %v", namespacedName, creationErr)
		}
		return true, nil
	}

	if applyOptions.CreateOnly() {
		return false, nil
	}

	afterMutationRaw, _ := json.Marshal(k8sObject)
	k8sObjectWasChanged := !reflect.DeepEqual(beforeMutationRaw, afterMutationRaw)
	if !k8sObjectWasChanged {
		return false, nil
	}

	if updateErr := client.Update(ctx, k8sObject); updateErr != nil {
		return false, fmt.Errorf("failed updating exsiting K8s object `%v`: %v", namespacedName, updateErr)
	}

	return true, nil
}
