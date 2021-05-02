package monitor

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/vmware/cbcontainers-operator/cbcontainers/monitor/models"
	hardeningObjects "github.com/vmware/cbcontainers-operator/cbcontainers/state/hardening/objects"
	admissionsV1 "k8s.io/api/admissionregistration/v1beta1"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"time"
)

const (
	HardeningFeature = "guardrails"
	RuntimeFeature   = "nodeguard"
)

type FeaturesStatusProvider interface {
	HardeningEnabled() (bool, error)
	RuntimeEnabled() (bool, error)
}

type HealthChecker interface {
	GetPods() (map[string]coreV1.Pod, error)
	GetReplicaSets() (map[string]appsV1.ReplicaSet, error)
	GetDeployments() (map[string]appsV1.Deployment, error)
	GetDaemonSets() (map[string]appsV1.DaemonSet, error)
	GetValidatingWebhookConfigurations() (map[string]admissionsV1.ValidatingWebhookConfiguration, error)
}

type MessageReporter interface {
	SendMonitorMessage(message models.HealthReportMessage) error
	Close() error
}

type MonitorAgent struct {
	account     string
	cluster     string
	accessToken string
	version     string

	healthChecker          HealthChecker
	featuresStatusProvider FeaturesStatusProvider
	messageReporter        MessageReporter

	// The interval for sending health reports to the backend
	interval time.Duration

	// Channel for stopping the agent
	stopChan chan struct{}

	log logr.Logger
}

func NewMonitorAgent(account, cluster, version string, healthChecker HealthChecker, featuresStatus FeaturesStatusProvider, messageReporter MessageReporter, interval time.Duration, log logr.Logger) *MonitorAgent {
	return &MonitorAgent{
		account:                account,
		cluster:                cluster,
		version:                version,
		healthChecker:          healthChecker,
		featuresStatusProvider: featuresStatus,
		messageReporter:        messageReporter,
		interval:               interval,
		stopChan:               make(chan struct{}),
		log:                    log,
	}
}

func (agent *MonitorAgent) Start() {
	go agent.run()
}

func (agent *MonitorAgent) Stop() {
	close(agent.stopChan)
}

func (agent *MonitorAgent) run() {
	for {
		select {
		case <-time.After(agent.interval):
			message, err := agent.buildHealthMessage()
			if err != nil {
				agent.log.Error(err, "error building health message")
			}

			if err = agent.messageReporter.SendMonitorMessage(message); err != nil {
				agent.log.Error(err, "error reporting message to backend")
			}
		case <-agent.stopChan:
			if err := agent.messageReporter.Close(); err != nil {
				agent.log.Error(err, "error closing reporter")
			}
			return
		}
	}
}

func (agent *MonitorAgent) buildHealthMessage() (models.HealthReportMessage, error) {
	hardeningEnabled, err := agent.featuresStatusProvider.HardeningEnabled()
	if err != nil {
		return models.HealthReportMessage{}, err
	}

	runtimeEnabled, err := agent.featuresStatusProvider.RuntimeEnabled()
	if err != nil {
		return models.HealthReportMessage{}, err
	}

	workloadsReports, err := agent.createWorkloadsHealthReports()
	if err != nil {
		return models.HealthReportMessage{}, err
	}

	webhooksReports, err := agent.createWebhooksHealthReport()
	if err != nil {
		return models.HealthReportMessage{}, err
	}

	enabledComponents := map[string]bool{
		HardeningFeature: hardeningEnabled,
		RuntimeFeature:   runtimeEnabled,
	}

	return models.NewHealthReportMessage(agent.account, agent.cluster, agent.version, enabledComponents, workloadsReports, webhooksReports), nil
}

func (agent *MonitorAgent) createWorkloadsHealthReports() (map[string]models.WorkloadHealthReport, error) {
	workloadsReports := make(map[string]models.WorkloadHealthReport)

	pods, err := agent.healthChecker.GetPods()
	if err != nil {
		return nil, err
	}

	replicasSets, err := agent.healthChecker.GetReplicaSets()
	if err != nil {
		return nil, err
	}

	deployments, err := agent.healthChecker.GetDeployments()
	if err != nil {
		return nil, err
	}

	daemonSets, err := agent.healthChecker.GetDaemonSets()
	if err != nil {
		return nil, err
	}

	agent.populateWithDeploymentsWorkloads(deployments, workloadsReports)
	agent.populateWithDaemonSetsWorkloads(daemonSets, workloadsReports)
	agent.updateWorkloadsReplicasWithPodsAndReplicaSets(pods, replicasSets, workloadsReports)

	return workloadsReports, nil
}

func (agent *MonitorAgent) populateWithDeploymentsWorkloads(deployments map[string]appsV1.Deployment, reports map[string]models.WorkloadHealthReport) {
	for _, deployment := range deployments {
		workloadMessage, err := agent.buildDeploymentMessage(deployment)
		if err != nil {
			agent.log.Error(err, "error building Deployment message")
			continue
		}

		if _, ok := reports[deployment.Name]; ok {
			agent.log.Info("duplicated workload name", "service", deployment.Name)
		}

		reports[deployment.Name] = workloadMessage
	}
}

func (agent *MonitorAgent) populateWithDaemonSetsWorkloads(daemonSets map[string]appsV1.DaemonSet, services map[string]models.WorkloadHealthReport) {
	for _, daemonSet := range daemonSets {
		workloadMessage, err := agent.buildDaemonSetMessage(daemonSet)
		if err != nil {
			agent.log.Error(err, "error building DaemonSet message")
			continue
		}

		if _, ok := services[daemonSet.Name]; ok {
			agent.log.Info("duplicate service name", "service", daemonSet.Name)
		}

		services[daemonSet.Name] = workloadMessage
	}
}

func (agent *MonitorAgent) buildDeploymentMessage(deployment appsV1.Deployment) (models.WorkloadHealthReport, error) {
	return agent.buildWorkloadMessage(models.WorkloadKindDeployment, deployment.Name, deployment.Spec.Replicas, deployment.Spec.Template.Spec.Containers, deployment.Status, deployment.Labels)
}

func (agent *MonitorAgent) buildDaemonSetMessage(daemon appsV1.DaemonSet) (models.WorkloadHealthReport, error) {
	return agent.buildWorkloadMessage(models.WorkloadKindDaemonSet, daemon.Name, nil, daemon.Spec.Template.Spec.Containers, daemon.Status, daemon.Labels)
}

func (agent *MonitorAgent) buildReplicaMessage(pod coreV1.Pod) (models.WorkloadReplicaHealthReport, error) {
	specContainers := make(map[string]models.ContainerHealthReport)
	for _, container := range pod.Spec.Containers {
		specContainers[container.Name] = models.ContainerHealthReport{
			Image: container.Image,
		}
	}

	status, err := json.Marshal(pod.Status)
	if err != nil {
		return models.WorkloadReplicaHealthReport{}, fmt.Errorf("error marshaling status for pod: %v", pod.Name)
	}

	spec := models.WorkloadReplicaSpecReport{
		Containers: specContainers,
	}

	return models.WorkloadReplicaHealthReport{
		Node:   pod.Spec.NodeName,
		Spec:   spec,
		Status: status,
	}, nil
}

func (agent *MonitorAgent) buildWorkloadMessage(workloadKind models.WorkloadKind, name string, replicas *int32, containers []coreV1.Container, statusObj interface{}, labels map[string]string) (models.WorkloadHealthReport, error) {
	containersReports := make(map[string]models.ContainerHealthReport)
	for _, container := range containers {
		containersReports[container.Name] = models.ContainerHealthReport{Image: container.Image}
	}

	status, err := json.Marshal(statusObj)
	if err != nil {
		return models.WorkloadHealthReport{}, fmt.Errorf("error marshaling status for %v: %v", workloadKind, name)
	}

	spec := models.WorkloadSpecReport{
		Containers: containersReports,
	}
	if replicas != nil {
		spec.Replicas = *replicas
	}

	return models.WorkloadHealthReport{
		Kind:            workloadKind,
		Spec:            spec,
		Status:          status,
		ReplicasReports: make(map[string]models.WorkloadReplicaHealthReport),
		Labels:          labels,
	}, nil

}

func (agent *MonitorAgent) updateWorkloadsReplicasWithPodsAndReplicaSets(pods map[string]coreV1.Pod, replicaSets map[string]appsV1.ReplicaSet, reports map[string]models.WorkloadHealthReport) {
	for _, pod := range pods {
		if len(pod.OwnerReferences) < 1 {
			agent.log.Info("found pod with no parent", "pod", pod.Name)
			continue
		}
		owner := pod.OwnerReferences[0]
		ownerName := owner.Name
		if owner.Kind == "ReplicaSet" {
			if rs, ok := replicaSets[owner.Name]; ok && len(rs.OwnerReferences) > 0 {
				ownerName = rs.OwnerReferences[0].Name
			} else {
				//logger.Info("couldn't determine pod parent", "pod", pod.Name)
				continue
			}
		}

		if workloadMessage, ok := reports[ownerName]; ok {
			replicaMsg, err := agent.buildReplicaMessage(pod)
			if err != nil {
				agent.log.Error(err, "error getting pod data", "pod", pod.Name)
				continue
			}

			workloadMessage.ReplicasReports[pod.Name] = replicaMsg
		}
	}
}

func (agent *MonitorAgent) createWebhooksHealthReport() (map[string]models.WebhookHealthReport, error) {
	webhooksReports := make(map[string]models.WebhookHealthReport)
	validatingWebhooks, err := agent.healthChecker.GetValidatingWebhookConfigurations()
	if err != nil {
		return nil, err
	}

	agent.populateWithValidatingWebhooks(validatingWebhooks, webhooksReports)
	return webhooksReports, nil
}

func (agent *MonitorAgent) populateWithValidatingWebhooks(webhooks map[string]admissionsV1.ValidatingWebhookConfiguration, reports map[string]models.WebhookHealthReport) {
	if webhook, ok := webhooks[hardeningObjects.EnforcerName]; ok {
		webhookMessage := agent.buildValidatingWebhookMessage(webhook)
		if _, ok := webhooks[webhook.Name]; ok {
			agent.log.Info("duplicated webhook name", "webhook", webhook.Name)
		}
		reports[webhook.Name] = webhookMessage
	} else {
		agent.log.Info("octarine validating webhook not found.", "webhook", webhook.Name)
	}
}

func (agent *MonitorAgent) buildValidatingWebhookMessage(webhook admissionsV1.ValidatingWebhookConfiguration) models.WebhookHealthReport {
	return models.WebhookHealthReport{
		Type: models.WebhookTypeValidating,
		Uid:  string(webhook.UID),
	}
}
