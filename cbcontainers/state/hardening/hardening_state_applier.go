package hardening

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	hardeningObjects "github.com/vmware/cbcontainers-operator/cbcontainers/state/hardening/objects"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type HardeningChildK8sObjectApplier interface {
	ApplyHardeningChildK8sObject(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client, hardeningChildK8sObject HardeningChildK8sObject, applyOptionsList ...*applymentOptions.ApplyOptions) (bool, client.Object, error)
	DeleteK8sObjectIfExists(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client, hardeningChildK8sObject HardeningChildK8sObject) (bool, error)
}

type HardeningStateApplier struct {
	enforcerTlsSecret         *hardeningObjects.EnforcerTlsK8sObject
	enforcerDeployment        *hardeningObjects.EnforcerDeploymentK8sObject
	enforcerService           *hardeningObjects.EnforcerServiceK8sObject
	enforcerValidatingWebhook *hardeningObjects.EnforcerValidatingWebhookK8sObject
	enforcerMutatingWebhook   *hardeningObjects.EnforcerMutatingWebhookK8sObject
	stateReporterDeployment   *hardeningObjects.StateReporterDeploymentK8sObject
	childApplier              HardeningChildK8sObjectApplier
	log                       logr.Logger
}

func NewHardeningStateApplier(log logr.Logger, k8sVersion string, tlsSecretsValuesCreator hardeningObjects.TlsSecretsValuesCreator, childApplier HardeningChildK8sObjectApplier) *HardeningStateApplier {
	return &HardeningStateApplier{
		enforcerTlsSecret:         hardeningObjects.NewEnforcerTlsK8sObject(tlsSecretsValuesCreator),
		enforcerDeployment:        hardeningObjects.NewEnforcerDeploymentK8sObject(),
		enforcerService:           hardeningObjects.NewEnforcerServiceK8sObject(),
		enforcerValidatingWebhook: hardeningObjects.NewEnforcerValidatingWebhookK8sObject(k8sVersion),
		enforcerMutatingWebhook:   hardeningObjects.NewEnforcerMutatingWebhookK8sObject(k8sVersion),
		stateReporterDeployment:   hardeningObjects.NewStateReporterDeploymentK8sObject(),
		childApplier:              childApplier,
		log:                       log,
	}
}

func (c *HardeningStateApplier) ApplyDesiredState(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client, setOwner applymentOptions.OwnerSetter) (bool, error) {
	applyOptions := applymentOptions.NewApplyOptions().SetOwnerSetter(setOwner)

	mutatedEnforcer, err := c.applyEnforcer(ctx, cbContainersHardening, client, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied enforcer objects", "Mutated", mutatedEnforcer)

	mutatedStateReporter, err := c.applyStateReporter(ctx, cbContainersHardening, client, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied state reporter objects", "Mutated", mutatedStateReporter)

	return mutatedEnforcer || mutatedStateReporter, nil
}

func (c *HardeningStateApplier) applyEnforcerWebhooks(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client, tlsSecret *coreV1.Secret, applyOptions *applymentOptions.ApplyOptions) (bool, error) {
	tlsSecretValues := models.TlsSecretValuesFromSecretData(tlsSecret.Data)

	c.enforcerValidatingWebhook.UpdateTlsSecretValues(tlsSecretValues)
	mutatedValidatingWebhook, _, err := c.childApplier.ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.enforcerValidatingWebhook, applyOptions)
	if err != nil {
		return false, err
	}

	c.enforcerMutatingWebhook.UpdateTlsSecretValues(tlsSecretValues)
	mutatedMutatingWebhook, _, err := c.childApplier.ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.enforcerMutatingWebhook, applyOptions)
	if err != nil {
		return false, err
	}

	return mutatedValidatingWebhook || mutatedMutatingWebhook, nil
}

func (c *HardeningStateApplier) deleteEnforcerWebhooks(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client) (bool, error) {
	webhookDeleted := false

	if deleted, deleteErr := c.childApplier.DeleteK8sObjectIfExists(ctx, cbContainersHardening, client, c.enforcerValidatingWebhook); deleteErr != nil {
		return false, deleteErr
	} else if deleted {
		webhookDeleted = true
		c.log.Info("Deleted enforcer validating webhook")
	}

	if deleted, deleteErr := c.childApplier.DeleteK8sObjectIfExists(ctx, cbContainersHardening, client, c.enforcerMutatingWebhook); deleteErr != nil {
		return webhookDeleted, deleteErr
	} else if deleted {
		webhookDeleted = true
		c.log.Info("Deleted enforcer mutating webhook")
	}

	return webhookDeleted, nil
}

func (c *HardeningStateApplier) applyEnforcer(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client, applyOptions *applymentOptions.ApplyOptions) (bool, error) {
	mutatedSecret, secretK8sObject, err := c.childApplier.ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.enforcerTlsSecret, applyOptions, applymentOptions.NewApplyOptions().SetCreateOnly(true))
	if err != nil {
		return false, err
	}
	c.log.Info("Applied enforcer tls secret", "Mutated", mutatedSecret)

	tlsSecret, ok := secretK8sObject.(*coreV1.Secret)
	if !ok {
		return false, fmt.Errorf("expected Secret K8s object")
	}

	mutatedService, _, err := c.childApplier.ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.enforcerService, applyOptions)
	if err != nil {
		if _, deleteErr := c.deleteEnforcerWebhooks(ctx, cbContainersHardening, client); deleteErr != nil {
			return false, deleteErr
		}

		return false, err
	}
	c.log.Info("Applied enforcer service", "Mutated", mutatedService)

	mutatedDeployment, deploymentK8sObject, err := c.childApplier.ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.enforcerDeployment, applyOptions)
	if err != nil {
		if _, deleteErr := c.deleteEnforcerWebhooks(ctx, cbContainersHardening, client); deleteErr != nil {
			return false, deleteErr
		}

		return false, err
	}
	c.log.Info("Applied enforcer deployment", "Mutated", mutatedDeployment)

	enforcerDeployment, ok := deploymentK8sObject.(*appsV1.Deployment)
	if !ok {
		return false, fmt.Errorf("expected Deployment K8s object")
	}

	mutatedWebhooks := false
	if enforcerDeployment.Status.ReadyReplicas < 1 {
		if deleted, deleteErr := c.deleteEnforcerWebhooks(ctx, cbContainersHardening, client); deleteErr != nil {
			return false, deleteErr
		} else if deleted {
			mutatedWebhooks = true
		}
	} else {
		mutatedWebhooks, err = c.applyEnforcerWebhooks(ctx, cbContainersHardening, client, tlsSecret, applyOptions)
		if err != nil {
			return false, err
		}
		c.log.Info("Applied enforcer webhooks", "Mutated", mutatedWebhooks)
	}

	return mutatedSecret || mutatedDeployment || mutatedService || mutatedWebhooks, nil
}

func (c *HardeningStateApplier) applyStateReporter(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client, applyOptions *applymentOptions.ApplyOptions) (bool, error) {
	mutatedDeployment, _, err := c.childApplier.ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.stateReporterDeployment, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied state reporter deployment", "Mutated", mutatedDeployment)
	return mutatedDeployment, nil
}
