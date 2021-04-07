package hardening

import (
	"context"
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	hardeningObjects "github.com/vmware/cbcontainers-operator/cbcontainers/state/hardening/objects"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type HardeningStateApplier struct {
	enforcerTlsSecret       *hardeningObjects.EnforcerTlsK8sObject
	enforcerDeployment      *hardeningObjects.EnforcerDeploymentK8sObject
	enforcerService         *hardeningObjects.EnforcerServiceK8sObject
	enforcerWebhook         *hardeningObjects.EnforcerWebhookK8sObject
	stateReporterDeployment *hardeningObjects.StateReporterDeploymentK8sObject
}

func NewHardeningStateApplier(tlsSecretsValuesCreator hardeningObjects.TlsSecretsValuesCreator) *HardeningStateApplier {
	return &HardeningStateApplier{
		enforcerTlsSecret:       hardeningObjects.NewEnforcerTlsK8sObject(tlsSecretsValuesCreator),
		enforcerDeployment:      hardeningObjects.NewEnforcerDeploymentK8sObject(),
		enforcerService:         hardeningObjects.NewEnforcerServiceK8sObject(),
		enforcerWebhook:         hardeningObjects.NewEnforcerWebhookK8sObject(),
		stateReporterDeployment: hardeningObjects.NewStateReporterDeploymentK8sObject(),
	}
}

func (c *HardeningStateApplier) ApplyDesiredState(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client, setOwner applymentOptions.OwnerSetter) (bool, error) {
	applyOptions := applymentOptions.NewApplyOptions().SetOwnerSetter(setOwner)

	mutatedEnforcer, err := c.applyEnforcer(ctx, cbContainersHardening, client, applyOptions)
	if err != nil {
		return false, err
	}

	mutatedStateReporter, err := c.applyStateReporter(ctx, cbContainersHardening, client, applyOptions)
	if err != nil {
		return false, err
	}

	return mutatedEnforcer || mutatedStateReporter, nil
}

func (c *HardeningStateApplier) applyEnforcer(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client, applyOptions *applymentOptions.ApplyOptions) (bool, error) {
	mutatedSecret, secretK8sObject, err := ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.enforcerTlsSecret, applyOptions, applymentOptions.NewApplyOptions().SetCreateOnly(true))
	if err != nil {
		return false, err
	}

	tlsSecret, ok := secretK8sObject.(*coreV1.Secret)
	if !ok {
		return false, fmt.Errorf("expected Secret K8s object")
	}

	mutatedService, _, err := ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.enforcerService, applyOptions)
	if err != nil {
		if deleteErr := DeleteK8sObjectIfExists(ctx, cbContainersHardening, client, c.enforcerWebhook); deleteErr != nil {
			return false, deleteErr
		}
		return false, err
	}

	mutatedDeployment, deploymentK8sObject, err := ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.enforcerDeployment, applyOptions)
	if err != nil {
		if deleteErr := DeleteK8sObjectIfExists(ctx, cbContainersHardening, client, c.enforcerWebhook); deleteErr != nil {
			return false, deleteErr
		}
		return false, err
	}

	enforcerDeployment, ok := deploymentK8sObject.(*appsV1.Deployment)
	if !ok {
		return false, fmt.Errorf("expected Deployment K8s object")
	}

	mutatedWebhook := false
	if enforcerDeployment.Status.ReadyReplicas < 1 {
		if deleteErr := DeleteK8sObjectIfExists(ctx, cbContainersHardening, client, c.enforcerWebhook); deleteErr != nil {
			return false, deleteErr
		}
		mutatedWebhook = true
	} else {
		c.enforcerWebhook.TlsSecretValues = models.TlsSecretValuesFromSecretData(tlsSecret.Data)
		mutatedWebhook, _, err = ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.enforcerWebhook, applyOptions)
		if err != nil {
			return false, err
		}
	}

	return mutatedSecret || mutatedDeployment || mutatedService || mutatedWebhook, nil
}

func (c *HardeningStateApplier) applyStateReporter(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client, applyOptions *applymentOptions.ApplyOptions) (bool, error) {
	mutatedDeployment, _, err := ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.stateReporterDeployment, applyOptions)
	return mutatedDeployment, err
}
