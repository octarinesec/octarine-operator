package applyment

import (
	"context"
	"encoding/json"
	"fmt"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	stateTypes "github.com/vmware/cbcontainers-operator/cbcontainers/state/types"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DeleteK8sObjectIfExists(ctx context.Context, client client.Client, desiredK8sObject stateTypes.DesiredK8sObject) error {
	k8sObject, objectExists, err := getK8sObject(ctx, client, desiredK8sObject, desiredK8sObject.NamespacedName())
	if err != nil {
		return err
	}

	if !objectExists {
		return nil
	}

	return client.Delete(ctx, k8sObject)
}

func ApplyDesiredK8sObject(ctx context.Context, client client.Client, desiredK8sObject stateTypes.DesiredK8sObject, applyOptionsList ...*applymentOptions.ApplyOptions) (bool, client.Object, error) {
	applyOptions := applymentOptions.MergeApplyOptions(applyOptionsList...)
	namespacedName := desiredK8sObject.NamespacedName()

	k8sObject, objectExists, err := getK8sObject(ctx, client, desiredK8sObject, namespacedName)
	if err != nil {
		return false, nil, err
	}

	if objectExists && applyOptions.CreateOnly() {
		return false, k8sObject, nil
	}

	beforeMutationRaw, _ := json.Marshal(k8sObject)
	if err := desiredK8sObject.MutateK8sObject(k8sObject); err != nil {
		return false, nil, fmt.Errorf("failed mutating K8s object `%v`: %v", namespacedName, err)
	}

	if !objectExists {
		if err := createK8sObject(ctx, client, k8sObject, namespacedName, applyOptions); err != nil {
			return false, nil, err
		}

		return true, k8sObject, nil
	}

	k8sObjectWasChanged, err := updateK8sObject(ctx, client, applyOptions, k8sObject, namespacedName, beforeMutationRaw)
	if err != nil {
		return false, nil, err
	}

	return k8sObjectWasChanged, k8sObject, nil
}

func getK8sObject(ctx context.Context, client client.Client, desiredK8sObject stateTypes.DesiredK8sObject, namespacedName types.NamespacedName) (client.Object, bool, error) {
	k8sObject := desiredK8sObject.EmptyK8sObject()

	err := client.Get(ctx, namespacedName, k8sObject)
	if err != nil && !errors.IsNotFound(err) {
		return nil, false, fmt.Errorf("failed getting K8s object: %v", err)
	}

	objectsExists := err == nil || !errors.IsNotFound(err)

	return k8sObject, objectsExists, nil
}

func createK8sObject(ctx context.Context, client client.Client, k8sObject client.Object, namespacedName types.NamespacedName, applyOptions *applymentOptions.ApplyOptions) error {
	k8sObject.SetNamespace(namespacedName.Namespace)
	k8sObject.SetName(namespacedName.Name)

	if err := setOwner(applyOptions, k8sObject, namespacedName); err != nil {
		return err
	}

	if err := client.Create(ctx, k8sObject); err != nil {
		return fmt.Errorf("failed creating K8s object `%v`: %v", namespacedName, err)
	}

	return nil
}

func updateK8sObject(ctx context.Context, client client.Client, applyOptions *applymentOptions.ApplyOptions, k8sObject client.Object, namespacedName types.NamespacedName, beforeMutationRaw []byte) (bool, error) {
	if err := setOwner(applyOptions, k8sObject, namespacedName); err != nil {
		return false, err
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

func setOwner(applyOptions *applymentOptions.ApplyOptions, k8sObject client.Object, namespacedName types.NamespacedName) error {
	setOwner := applyOptions.OwnerSetter()
	if setOwner == nil {
		return nil
	}

	if err := setOwner(k8sObject); err != nil {
		return fmt.Errorf("failed setting owner to K8s object `%v`: %v", namespacedName, err)
	}

	return nil
}
