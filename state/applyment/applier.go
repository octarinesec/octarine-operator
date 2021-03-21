package applyment

import (
	"context"
	"fmt"
	stateTypes "github.com/vmware/cbcontainers-operator/state/types"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//func applyState(ctx context.Context, client client.Client, desiredState []stateTypes.DesiredK8sObject) (bool, error) {
//	applyOccurred := false
//
//	for _, desiredK8sObject := range desiredState {
//		if ok, err := applyDesiredK8sObject(ctx, client, desiredK8sObject); err != nil {
//			logrus.Error(err)
//			return false, err
//		} else if ok {
//			applyOccurred = true
//		}
//	}
//
//	return applyOccurred, nil
//}

func ApplyDesiredK8sObject(ctx context.Context, client client.Client, desiredK8sObject stateTypes.DesiredK8sObject) (bool, error) {
	k8sObject := desiredK8sObject.EmptyK8sObject()
	foundErr := client.Get(ctx, desiredK8sObject.NamespacedName(), k8sObject)

	if foundErr != nil && !errors.IsNotFound(foundErr) {
		return false, fmt.Errorf("failed getting K8s object: %v", foundErr)
	}

	mutated, mutateErr := desiredK8sObject.MutateK8sObject(k8sObject)
	if mutateErr != nil {
		return false, fmt.Errorf("failed mutating K8s object: %v", foundErr)
	}

	// k8s object was not found should, need to create
	if foundErr != nil {
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
