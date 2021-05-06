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
	enforcerTlsSecret       *hardeningObjects.EnforcerTlsK8sObject
	enforcerDeployment      *hardeningObjects.EnforcerDeploymentK8sObject
	enforcerService         *hardeningObjects.EnforcerServiceK8sObject
	enforcerWebhook         *hardeningObjects.EnforcerWebhookK8sObject
	stateReporterDeployment *hardeningObjects.StateReporterDeploymentK8sObject
	childApplier            HardeningChildK8sObjectApplier
	log                     logr.Logger
}

func NewHardeningStateApplier(log logr.Logger, tlsSecretsValuesCreator hardeningObjects.TlsSecretsValuesCreator, childApplier HardeningChildK8sObjectApplier) *HardeningStateApplier {
	return &HardeningStateApplier{
		enforcerTlsSecret:       hardeningObjects.NewEnforcerTlsK8sObject(tlsSecretsValuesCreator),
		enforcerDeployment:      hardeningObjects.NewEnforcerDeploymentK8sObject(),
		enforcerService:         hardeningObjects.NewEnforcerServiceK8sObject(),
		enforcerWebhook:         hardeningObjects.NewEnforcerWebhookK8sObject(),
		stateReporterDeployment: hardeningObjects.NewStateReporterDeploymentK8sObject(),
		childApplier:            childApplier,
		log:                     log,
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

func (c *HardeningStateApplier) applyEnforcer(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client, applyOptions *applymentOptions.ApplyOptions) (bool, error) {
	c.log.Info("Getting Nodes list")
	nodesList := &coreV1.NodeList{}
	if err := client.List(ctx, nodesList); err != nil || nodesList.Items == nil || len(nodesList.Items) < 1 {
		return false, fmt.Errorf("couldn't get nodes list")
	}

	mutatedSecret, secretK8sObject, err := c.childApplier.ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.enforcerTlsSecret, applyOptions, applymentOptions.NewApplyOptions().SetCreateOnly(true))
	if err != nil {
		return false, err
	}
	c.log.Info("Applied enfocer tls secret", "Mutated", mutatedSecret)

	tlsSecret, ok := secretK8sObject.(*coreV1.Secret)
	if !ok {
		return false, fmt.Errorf("expected Secret K8s object")
	}

	mutatedService, _, err := c.childApplier.ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.enforcerService, applyOptions)
	if err != nil {
		if deleted, deleteErr := c.childApplier.DeleteK8sObjectIfExists(ctx, cbContainersHardening, client, c.enforcerWebhook); deleteErr != nil {
			return false, deleteErr
		} else if deleted {
			c.log.Info("Deleted enforcer webhook")
		}
		return false, err
	}
	c.log.Info("Applied enforcer service", "Mutated", mutatedService)

	mutatedDeployment, deploymentK8sObject, err := c.childApplier.ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.enforcerDeployment, applyOptions)
	if err != nil {
		if deleted, deleteErr := c.childApplier.DeleteK8sObjectIfExists(ctx, cbContainersHardening, client, c.enforcerWebhook); deleteErr != nil {
			return false, deleteErr
		} else if deleted {
			c.log.Info("Deleted enforcer webhook")
		}
		return false, err
	}
	c.log.Info("Applied enforcer deployment", "Mutated", mutatedDeployment)

	enforcerDeployment, ok := deploymentK8sObject.(*appsV1.Deployment)
	if !ok {
		return false, fmt.Errorf("expected Deployment K8s object")
	}

	mutatedWebhook := false
	if enforcerDeployment.Status.ReadyReplicas < 1 {
		if deleted, deleteErr := c.childApplier.DeleteK8sObjectIfExists(ctx, cbContainersHardening, client, c.enforcerWebhook); deleteErr != nil {
			return false, deleteErr
		} else if deleted {
			mutatedWebhook = true
			c.log.Info("Deleted enforcer webhook")
		}
	} else {
		c.enforcerWebhook.UpdateTlsSecretValues(models.TlsSecretValuesFromSecretData(tlsSecret.Data))
		c.enforcerWebhook.UpdateKubeletVersion(nodesList.Items[0].Status.NodeInfo.KubeletVersion)
		mutatedWebhook, _, err = c.childApplier.ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.enforcerWebhook, applyOptions)
		if err != nil {
			return false, err
		}
		c.log.Info("Applied enforcer webhook", "Mutated", mutatedWebhook)
	}

	return mutatedSecret || mutatedDeployment || mutatedService || mutatedWebhook, nil
}

func (c *HardeningStateApplier) applyStateReporter(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client, applyOptions *applymentOptions.ApplyOptions) (bool, error) {
	mutatedDeployment, _, err := c.childApplier.ApplyHardeningChildK8sObject(ctx, cbContainersHardening, client, c.stateReporterDeployment, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied state reporter deployment", "Mutated", mutatedDeployment)
	return mutatedDeployment, nil
}
