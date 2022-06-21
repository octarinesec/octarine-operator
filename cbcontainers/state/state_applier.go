package state

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/agent_applyment"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/components"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AgentComponentApplier interface {
	Apply(ctx context.Context, builder agent_applyment.AgentComponentBuilder, agentSpec *cbcontainersv1.CBContainersAgentSpec, applyOptionsList ...*applymentOptions.ApplyOptions) (bool, client.Object, error)
	Delete(ctx context.Context, builder agent_applyment.AgentComponentBuilder, agentSpec *cbcontainersv1.CBContainersAgentSpec) (bool, error)
}

type StateApplier struct {
	desiredConfigMap                *components.ConfigurationK8sObject
	desiredRegistrySecret           *components.RegistrySecretK8sObject
	desiredPriorityClass            *components.PriorityClassK8sObject
	desiredMonitorDeployment        *components.MonitorDeploymentK8sObject
	enforcerTlsSecret               *components.EnforcerTlsK8sObject
	enforcerDeployment              *components.EnforcerDeploymentK8sObject
	enforcerService                 *components.EnforcerServiceK8sObject
	enforcerValidatingWebhook       *components.EnforcerValidatingWebhookK8sObject
	enforcerMutatingWebhook         *components.EnforcerMutatingWebhookK8sObject
	stateReporterDeployment         *components.StateReporterDeploymentK8sObject
	resolverDeployment              *components.ResolverDeploymentK8sObject
	resolverService                 *components.ResolverServiceK8sObject
	sensorDaemonSet                 *components.SensorDaemonSetK8sObject
	imageScanningReporterDeployment *components.ImageScanningReporterDeploymentK8sObject
	imageScanningReporterService    *components.ImageScanningReporterServiceK8sObject
	applier                         AgentComponentApplier
	log                             logr.Logger
}

func NewStateApplier(agentComponentApplier AgentComponentApplier, k8sVersion string, tlsSecretsValuesCreator components.TlsSecretsValuesCreator, log logr.Logger) *StateApplier {
	return &StateApplier{
		desiredConfigMap:                components.NewConfigurationK8sObject(),
		desiredRegistrySecret:           components.NewRegistrySecretK8sObject(),
		desiredPriorityClass:            components.NewPriorityClassK8sObject(k8sVersion),
		desiredMonitorDeployment:        components.NewMonitorDeploymentK8sObject(),
		enforcerTlsSecret:               components.NewEnforcerTlsK8sObject(tlsSecretsValuesCreator),
		enforcerDeployment:              components.NewEnforcerDeploymentK8sObject(),
		enforcerService:                 components.NewEnforcerServiceK8sObject(),
		enforcerValidatingWebhook:       components.NewEnforcerValidatingWebhookK8sObject(k8sVersion),
		enforcerMutatingWebhook:         components.NewEnforcerMutatingWebhookK8sObject(k8sVersion),
		stateReporterDeployment:         components.NewStateReporterDeploymentK8sObject(),
		resolverDeployment:              components.NewResolverDeploymentK8sObject(),
		resolverService:                 components.NewResolverServiceK8sObject(),
		sensorDaemonSet:                 components.NewSensorDaemonSetK8sObject(),
		imageScanningReporterDeployment: components.NewImageScanningReporterDeploymentK8sObject(),
		imageScanningReporterService:    components.NewImageScanningReporterServiceK8sObject(),
		applier:                         agentComponentApplier,
		log:                             log,
	}
}

func (c *StateApplier) GetPriorityClassEmptyK8sObject() client.Object {
	return c.desiredPriorityClass.EmptyK8sObject()
}

func (c *StateApplier) ApplyDesiredState(ctx context.Context, agentSpec *cbcontainersv1.CBContainersAgentSpec, registrySecret *models.RegistrySecretValues, setOwner applymentOptions.OwnerSetter) (bool, error) {
	applyOptions := applymentOptions.NewApplyOptions().SetOwnerSetter(setOwner)

	coreMutated, err := c.applyCoreComponents(ctx, agentSpec, registrySecret, applyOptions)
	if err != nil {
		return false, err
	}

	mutatedEnforcer, err := c.applyEnforcer(ctx, agentSpec, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied enforcer objects", "Mutated", mutatedEnforcer)

	mutatedStateReporter, err := c.applyStateReporter(ctx, agentSpec, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied state reporter objects", "Mutated", mutatedStateReporter)

	mutatedRuntimeResolver, runtimeResolverDeleted := false, false
	mutatedComponentsDaemonSet, componentsDamonSetDeleted := false, false
	var deleteErr error = nil

	if common.IsEnabled(agentSpec.Components.RuntimeProtection.Enabled) {
		mutatedRuntimeResolver, err = c.applyResolver(ctx, agentSpec, applyOptions)
		if err != nil {
			return false, err
		}
		c.log.Info("Applied runtime kubernetes resolver objects", "Mutated", mutatedRuntimeResolver)

	} else {
		runtimeResolverDeleted, deleteErr = c.deleteResolver(ctx, agentSpec)
		if deleteErr != nil {
			return false, deleteErr
		}
	}

	mutatedImageScanningReporter, imageScanningReporterDeleted := false, false
	if common.IsEnabled(agentSpec.Components.ClusterScanning.Enabled) {
		mutatedImageScanningReporter, err = c.applyImageScanningReporter(ctx, agentSpec, applyOptions)
		if err != nil {
			return false, err
		}

		c.log.Info("Applied image scanning reporter objects", "Mutated", mutatedImageScanningReporter)
	} else {
		imageScanningReporterDeleted, err = c.deleteImageScanningReporter(ctx, agentSpec)
		if err != nil {
			return false, err
		}
	}

	if common.IsEnabled(agentSpec.Components.ClusterScanning.Enabled) || common.IsEnabled(agentSpec.Components.RuntimeProtection.Enabled) {
		mutatedComponentsDaemonSet, err = c.applyComponentsDamonSet(ctx, agentSpec, applyOptions)
		if err != nil {
			return false, err
		}
		c.log.Info("Applied featured components daemon set objects", "Mutated", mutatedComponentsDaemonSet)
	} else {
		componentsDamonSetDeleted, err = c.deleteComponentsDamonSet(ctx, agentSpec)
		if err != nil {
			return false, err
		}
	}

	return coreMutated || mutatedEnforcer || mutatedStateReporter || mutatedRuntimeResolver || mutatedComponentsDaemonSet || runtimeResolverDeleted || mutatedImageScanningReporter || imageScanningReporterDeleted || componentsDamonSetDeleted, nil
}

func (c *StateApplier) applyCoreComponents(ctx context.Context, agentSpec *cbcontainersv1.CBContainersAgentSpec, registrySecret *models.RegistrySecretValues, applyOptions *applymentOptions.ApplyOptions) (bool, error) {
	mutatedConfigmap, _, err := c.applier.Apply(ctx, c.desiredConfigMap, agentSpec, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied config map", "Mutated", mutatedConfigmap)

	c.desiredRegistrySecret.UpdateRegistrySecretValues(registrySecret)
	mutatedRegistrySecret, _, err := c.applier.Apply(ctx, c.desiredRegistrySecret, agentSpec, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied registry secret", "Mutated", mutatedRegistrySecret)

	mutatedPriorityClass, _, err := c.applier.Apply(ctx, c.desiredPriorityClass, agentSpec, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied priority class", "Mutated", mutatedPriorityClass)

	mutatedMonitor, _, err := c.applier.Apply(ctx, c.desiredMonitorDeployment, agentSpec, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied Monitor", "Mutated", mutatedMonitor)

	return mutatedConfigmap || mutatedRegistrySecret || mutatedPriorityClass || mutatedMonitor, nil
}

func (c *StateApplier) applyEnforcer(ctx context.Context, agentSpec *cbcontainersv1.CBContainersAgentSpec, applyOptions *applymentOptions.ApplyOptions) (bool, error) {
	mutatedSecret, secretK8sObject, err := c.applier.Apply(ctx, c.enforcerTlsSecret, agentSpec, applyOptions, applymentOptions.NewApplyOptions().SetCreateOnly(true))
	if err != nil {
		return false, err
	}
	c.log.Info("Applied enforcer tls secret", "Mutated", mutatedSecret)

	tlsSecret, ok := secretK8sObject.(*coreV1.Secret)
	if !ok {
		return false, fmt.Errorf("expected Secret K8s object")
	}

	mutatedService, _, err := c.applier.Apply(ctx, c.enforcerService, agentSpec, applyOptions)
	if err != nil {
		deleted, deleteErr := c.deleteAllEnforcerWebhooks(ctx, agentSpec)
		c.log.Info("Deleted enforcer webhooks because of an error while applying enforcer service", "deleted", deleted, "deletion-error", deleteErr)
		if deleteErr != nil {
			return deleted, deleteErr
		}

		return deleted, err
	}
	c.log.Info("Applied enforcer service", "Mutated", mutatedService)

	mutatedDeployment, deploymentK8sObject, err := c.applier.Apply(ctx, c.enforcerDeployment, agentSpec, applyOptions)
	if err != nil {
		deleted, deleteErr := c.deleteAllEnforcerWebhooks(ctx, agentSpec)
		c.log.Info("Deleted enforcer webhooks because of an error while applying enforcer deployment", "deleted", deleted, "deletion-error", deleteErr)
		if deleteErr != nil {
			return deleted, deleteErr
		}

		return deleted, err
	}
	c.log.Info("Applied enforcer deployment", "Mutated", mutatedDeployment)

	enforcerDeployment, ok := deploymentK8sObject.(*appsV1.Deployment)
	if !ok {
		return false, fmt.Errorf("expected Deployment K8s object")
	}

	mutatedWebhooks := false
	if enforcerDeployment.Status.ReadyReplicas < 1 {
		if deleted, deleteErr := c.deleteAllEnforcerWebhooks(ctx, agentSpec); deleteErr != nil {
			return false, deleteErr
		} else if deleted {
			mutatedWebhooks = true
		}
	} else {
		mutatedWebhooks, err = c.applyEnforcerWebhooks(ctx, agentSpec, tlsSecret, applyOptions)
		if err != nil {
			return false, err
		}
		c.log.Info("Applied enforcer webhooks", "Mutated", mutatedWebhooks)
	}

	return mutatedSecret || mutatedDeployment || mutatedService || mutatedWebhooks, nil
}

func (c *StateApplier) applyStateReporter(ctx context.Context, agentSpec *cbcontainersv1.CBContainersAgentSpec, applyOptions *applymentOptions.ApplyOptions) (bool, error) {
	mutatedDeployment, _, err := c.applier.Apply(ctx, c.stateReporterDeployment, agentSpec, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied state reporter deployment", "Mutated", mutatedDeployment)
	return mutatedDeployment, nil
}

func (c *StateApplier) applyResolver(ctx context.Context, agentSpec *cbcontainersv1.CBContainersAgentSpec, applyOptions *applymentOptions.ApplyOptions) (bool, error) {
	mutatedService, _, err := c.applier.Apply(ctx, c.resolverService, agentSpec, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied kubernetes resolver service", "Mutated", mutatedService)

	mutatedDeployment, _, err := c.applier.Apply(ctx, c.resolverDeployment, agentSpec, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied runtime kubernetes resolver deployment", "Mutated", mutatedDeployment)
	return mutatedService || mutatedDeployment, nil
}

// applyComponentsDamonSet applies the daemon set that stores the runtime sensor and/or the cluster-scanning scanner containers.
// the daemon set is set to be applied if either of the featured components are enabled.
func (c *StateApplier) applyComponentsDamonSet(ctx context.Context, agentSpec *cbcontainersv1.CBContainersAgentSpec, applyOptions *applymentOptions.ApplyOptions) (bool, error) {
	mutatedDaemonSet, _, err := c.applier.Apply(ctx, c.sensorDaemonSet, agentSpec, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied daemon set featured components", "Mutated", mutatedDaemonSet)
	return mutatedDaemonSet, nil
}

func (c *StateApplier) deleteResolver(ctx context.Context, agentSpec *cbcontainersv1.CBContainersAgentSpec) (bool, error) {
	resolverServiceDeleted, deleteErr := c.applier.Delete(ctx, c.resolverService, agentSpec)
	if deleteErr != nil {
		return false, deleteErr
	} else if resolverServiceDeleted {
		c.log.Info("Deleted resolver service")
	}

	resolverDeploymentDeleted, deleteErr := c.applier.Delete(ctx, c.resolverDeployment, agentSpec)
	if deleteErr != nil {
		return false, deleteErr
	} else if resolverDeploymentDeleted {
		c.log.Info("Deleted resolver deployment")
	}

	return resolverServiceDeleted || resolverDeploymentDeleted, nil
}

func (c *StateApplier) applyImageScanningReporter(ctx context.Context, agentSpec *cbcontainersv1.CBContainersAgentSpec, applyOptions *applymentOptions.ApplyOptions) (bool, error) {
	mutatedService, _, err := c.applier.Apply(ctx, c.imageScanningReporterService, agentSpec, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied image scanning reporter service", "Mutated", mutatedService)

	mutatedDeployment, _, err := c.applier.Apply(ctx, c.imageScanningReporterDeployment, agentSpec, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied image scanning reporter deployment", "Mutated", mutatedDeployment)

	return mutatedService || mutatedDeployment, nil
}

func (c *StateApplier) deleteImageScanningReporter(ctx context.Context, agentSpec *cbcontainersv1.CBContainersAgentSpec) (bool, error) {
	imageScanningReporterServiceDeleted, deleteErr := c.applier.Delete(ctx, c.imageScanningReporterService, agentSpec)
	if deleteErr != nil {
		return false, deleteErr
	} else if imageScanningReporterServiceDeleted {
		c.log.Info("Deleted image scanning reporter service")
	}

	imageScanningReporterDeploymentDeleted, deleteErr := c.applier.Delete(ctx, c.imageScanningReporterDeployment, agentSpec)
	if deleteErr != nil {
		return false, deleteErr
	} else if imageScanningReporterDeploymentDeleted {
		c.log.Info("Deleted image scanning reporter deployment")
	}

	return imageScanningReporterServiceDeleted || imageScanningReporterDeploymentDeleted, nil
}

// deleteComponentsDamonSet deletes the daemonset that runs the runtime sensor and/or the cluster-scanning scanner containers.
// the daemon set is being deleted only if all the feature components are disabled (runtime & cluster-scanner)
func (c *StateApplier) deleteComponentsDamonSet(ctx context.Context, agentSpec *cbcontainersv1.CBContainersAgentSpec) (bool, error) {
	sensorDaemonSetDeleted, deleteErr := c.applier.Delete(ctx, c.sensorDaemonSet, agentSpec)
	if deleteErr != nil {
		return false, deleteErr
	} else if sensorDaemonSetDeleted {
		c.log.Info("Deleted featured components daemonset")
	}

	return sensorDaemonSetDeleted, nil
}

func (c *StateApplier) applyEnforcerWebhooks(ctx context.Context, agentSpec *cbcontainersv1.CBContainersAgentSpec, tlsSecret *coreV1.Secret, applyOptions *applymentOptions.ApplyOptions) (bool, error) {
	tlsSecretValues := models.TlsSecretValuesFromSecretData(tlsSecret.Data)

	c.enforcerValidatingWebhook.UpdateTlsSecretValues(tlsSecretValues)
	mutatedValidatingWebhook, _, err := c.applier.Apply(ctx, c.enforcerValidatingWebhook, agentSpec, applyOptions)
	if err != nil {
		return false, err
	}

	mutatedMutatingWebhook := false
	if agentSpec.Components.Basic.Enforcer.EnableEnforcementFeature != nil && *agentSpec.Components.Basic.Enforcer.EnableEnforcementFeature {
		c.enforcerMutatingWebhook.UpdateTlsSecretValues(tlsSecretValues)
		mutatedMutatingWebhook, _, err = c.applier.Apply(ctx, c.enforcerMutatingWebhook, agentSpec, applyOptions)
		if err != nil {
			return false, err
		}
	} else {
		// Note that if someone changes the flag in sequence false -> true -> false, we should delete the webhook again
		deletedWebhookDueToDisabledFlag, deleteErr := c.deleteEnforcerWebhook(ctx, agentSpec, c.enforcerMutatingWebhook)
		if deleteErr != nil {
			return false, deleteErr
		}
		mutatedMutatingWebhook = deletedWebhookDueToDisabledFlag
	}

	c.log.Info("Applied enforcer webhooks", "Mutated validated webhook", mutatedValidatingWebhook, "Mutated mutating webhook", mutatedMutatingWebhook)
	return mutatedValidatingWebhook || mutatedMutatingWebhook, nil
}

func (c *StateApplier) deleteAllEnforcerWebhooks(ctx context.Context, agentSpec *cbcontainersv1.CBContainersAgentSpec) (bool, error) {
	return c.deleteEnforcerWebhook(ctx, agentSpec, c.enforcerValidatingWebhook, c.enforcerMutatingWebhook)
}

func (c *StateApplier) deleteEnforcerWebhook(ctx context.Context, agentSpec *cbcontainersv1.CBContainersAgentSpec, webhooks ...agent_applyment.AgentComponentBuilder) (bool, error) {
	deletedAnything := false

	for _, webhook := range webhooks {
		if deleted, deleteErr := c.applier.Delete(ctx, webhook, agentSpec); deleteErr != nil {
			return false, deleteErr
		} else if deleted {
			c.log.Info("Deleted webhook", "webhook-name", webhook.NamespacedName())
			deletedAnything = true
		}
	}

	return deletedAnything, nil
}
