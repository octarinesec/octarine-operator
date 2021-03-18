package controllers

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	stateTypes "github.com/vmware/cbcontainers-operator/state/types"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func applyState(ctx context.Context, client client.Client, desiredState []stateTypes.DesiredK8sObject) error {
	for _, desiredK8sObject := range desiredState {
		if err := handleDesiredK8sObject(ctx, client, desiredK8sObject); err != nil {
			logrus.Error(err)
			return err
		}
	}

	return nil
}

func handleDesiredK8sObject(ctx context.Context, client client.Client, desiredK8sObject stateTypes.DesiredK8sObject) error {
	existingK8sObjectToFind := desiredK8sObject.GetEmptyK8sObject()
	err := client.Get(ctx, desiredK8sObject.GetNamespacedName(), existingK8sObjectToFind)

	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed getting K8s object: %v", err)
	}

	if err != nil && errors.IsNotFound(err) {
		if creationErr := client.Create(ctx, desiredK8sObject.GetK8sObject()); creationErr != nil {
			return fmt.Errorf("failed creating K8s object: %v", creationErr)
		}
		return nil
	}

	ok, verifyErr := desiredK8sObject.VerifyK8sObject(existingK8sObjectToFind)
	if verifyErr != nil {
		return fmt.Errorf("failed verifing exsiting K8s object: %v", verifyErr)
	}

	if !ok {
		if updateErr := client.Update(ctx, existingK8sObjectToFind); updateErr != nil {
			return fmt.Errorf("failed updating exsiting K8s object: %v", updateErr)
		}
	}

	return nil
}
