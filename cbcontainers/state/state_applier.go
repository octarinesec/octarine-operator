package state

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/agent_applyment"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
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
	desiredConfigMap         *components.ConfigurationK8sObject
	desiredRegistrySecret    *components.RegistrySecretK8sObject
	desiredPriorityClass     *components.PriorityClassK8sObject
	desiredMonitorDeployment *components.MonitorDeploymentK8sObject
	enforcerTlsSecret        *components.EnforcerTlsK8sObject
	enforcerDeployment       *components.EnforcerDeploymentK8sObject
	enforcerService          *components.EnforcerServiceK8sObject
	enforcerWebhook          *components.EnforcerWebhookK8sObject
	stateReporterDeployment  *components.StateReporterDeploymentK8sObject
	resolverDeployment       *components.ResolverDeploymentK8sObject
	resolverService          *components.ResolverServiceK8sObject
	sensorDaemonSet          *components.SensorDaemonSetK8sObject
	applier                  AgentComponentApplier
	log                      logr.Logger
}

func NewStateApplier(agentComponentApplier AgentComponentApplier, k8sVersion string, tlsSecretsValuesCreator components.TlsSecretsValuesCreator, log logr.Logger) *StateApplier {
	return &StateApplier{
		desiredConfigMap:         components.NewConfigurationK8sObject(),
		desiredRegistrySecret:    components.NewRegistrySecretK8sObject(),
		desiredPriorityClass:     components.NewPriorityClassK8sObject(k8sVersion),
		desiredMonitorDeployment: components.NewMonitorDeploymentK8sObject(),
		enforcerTlsSecret:        components.NewEnforcerTlsK8sObject(tlsSecretsValuesCreator),
		enforcerDeployment:       components.NewEnforcerDeploymentK8sObject(),
		enforcerService:          components.NewEnforcerServiceK8sObject(),
		enforcerWebhook:          components.NewEnforcerWebhookK8sObject(k8sVersion),
		stateReporterDeployment:  components.NewStateReporterDeploymentK8sObject(),
		resolverDeployment:       components.NewResolverDeploymentK8sObject(),
		resolverService:          components.NewResolverServiceK8sObject(),
		sensorDaemonSet:          components.NewSensorDaemonSetK8sObject(),
		applier:                  agentComponentApplier,
		log:                      log,
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

	mutatedResolver, err := c.applyResolver(ctx, agentSpec, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied runtime kubernetes resolver objects", "Mutated", mutatedResolver)

	mutatedSensor, err := c.applySensor(ctx, agentSpec, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied runtime kubernetes sensor objects", "Mutated", mutatedSensor)

	return coreMutated || mutatedEnforcer || mutatedStateReporter || mutatedResolver || mutatedSensor, nil
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
		if deleted, deleteErr := c.applier.Delete(ctx, c.enforcerWebhook, agentSpec); deleteErr != nil {
			return false, deleteErr
		} else if deleted {
			c.log.Info("Deleted enforcer webhook")
		}
		return false, err
	}
	c.log.Info("Applied enforcer service", "Mutated", mutatedService)

	mutatedDeployment, deploymentK8sObject, err := c.applier.Apply(ctx, c.enforcerDeployment, agentSpec, applyOptions)
	if err != nil {
		if deleted, deleteErr := c.applier.Delete(ctx, c.enforcerWebhook, agentSpec); deleteErr != nil {
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
		if deleted, deleteErr := c.applier.Delete(ctx, c.enforcerWebhook, agentSpec); deleteErr != nil {
			return false, deleteErr
		} else if deleted {
			mutatedWebhook = true
			c.log.Info("Deleted enforcer webhook")
		}
	} else {
		c.enforcerWebhook.UpdateTlsSecretValues(models.TlsSecretValuesFromSecretData(tlsSecret.Data))
		mutatedWebhook, _, err = c.applier.Apply(ctx, c.enforcerWebhook, agentSpec, applyOptions)
		if err != nil {
			return false, err
		}
		c.log.Info("Applied enforcer webhook", "Mutated", mutatedWebhook)
	}

	return mutatedSecret || mutatedDeployment || mutatedService || mutatedWebhook, nil
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

func (c *StateApplier) applySensor(ctx context.Context, agentSpec *cbcontainersv1.CBContainersAgentSpec, applyOptions *applymentOptions.ApplyOptions) (bool, error) {
	mutatedDaemonSet, _, err := c.applier.Apply(ctx, c.sensorDaemonSet, agentSpec, applyOptions)
	if err != nil {
		return false, err
	}
	c.log.Info("Applied runtime kubernetes sensor daemon set", "Mutated", mutatedDaemonSet)
	return mutatedDaemonSet, nil
}