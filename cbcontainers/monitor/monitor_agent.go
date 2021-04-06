package monitor

import (
	"encoding/json"
	"fmt"
	admissionsV1 "k8s.io/api/admissionregistration/v1"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"time"
)

type featuresStatus interface {
	GetEnabledFeatures() map[string]bool
}

type healthChecker interface {
	GetPods() (map[string]coreV1.Pod, error)
	GetReplicaSets() (map[string]appsV1.ReplicaSet, error)
	GetDeployments() (map[string]appsV1.Deployment, error)
	GetDaemonSets() (map[string]appsV1.DaemonSet, error)
	GetValidatingWebhookConfigurations() (map[string]admissionsV1.ValidatingWebhookConfiguration, error)
}

type messageReporter interface {
	SendMonitorMessage(message HealthReportMessage) error
}

type MonitorAgent struct {
	account     string
	cluster     string
	accessToken string
	version     string

	healthChecker   healthChecker
	featuresStatus  featuresStatus
	messageReporter messageReporter

	// The interval for sending health reports to the backend
	interval time.Duration

	// Channel for stopping the agent
	stopChan chan struct{}
}

func NewMonitorAgent(account, cluster, accessToken, version string, healthChecker healthChecker, featuresStatus featuresStatus, messageReporter messageReporter, interval time.Duration) *MonitorAgent {
	return &MonitorAgent{
		account:         account,
		cluster:         cluster,
		accessToken:     accessToken,
		version:         version,
		healthChecker:   healthChecker,
		featuresStatus:  featuresStatus,
		messageReporter: messageReporter,
		interval:        interval,
		stopChan:        make(chan struct{}),
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
				//logger.Error(err, "error building health message")
			}

			if err = agent.messageReporter.SendMonitorMessage(message); err != nil {
				//logger.Error(err, "error reporting message to backend")
			}
		case <-agent.stopChan:
			return
		}
	}
}

func (agent *MonitorAgent) buildHealthMessage() (HealthReportMessage, error) {
	workloadsReports, err := agent.createWorkloadsHealthReports()
	if err != nil {
		return HealthReportMessage{}, err
	}

	return HealthReportMessage{
		Account:           agent.account,
		Cluster:           agent.cluster,
		Version:           agent.version,
		EnabledComponents: agent.featuresStatus.GetEnabledFeatures(),
		Workloads:         workloadsReports,
		Webhooks:          nil,
	}, nil
}

func (agent *MonitorAgent) createWorkloadsHealthReports() (map[string]WorkloadHealthReport, error) {
	workloadsReports := make(map[string]WorkloadHealthReport)

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

func (agent *MonitorAgent) populateWithDeploymentsWorkloads(deployments map[string]appsV1.Deployment, reports map[string]WorkloadHealthReport) {
	for _, deployment := range deployments {
		workloadMessage, err := agent.buildDeploymentMessage(deployment)
		if err != nil {
			//logger.Error(err, "error building Deployment message")
			continue
		}

		if _, ok := reports[deployment.Name]; ok {
			//logger.Info("duplicated workload name", "service", deploymentName)
		}

		reports[deployment.Name] = workloadMessage
	}
}

func (agent *MonitorAgent) populateWithDaemonSetsWorkloads(daemonSets map[string]appsV1.DaemonSet, services map[string]WorkloadHealthReport) {
	for _, daemonSet := range daemonSets {
		workloadMessage, err := agent.buildDaemonSetMessage(daemonSet)
		if err != nil {
			//logger.Error(err, "error building DaemonSet message")
			continue
		}

		if _, ok := services[daemonSet.Name]; ok {
			//logger.Info("duplicate service name", "service", daemonSet.Name)
		}

		services[daemonSet.Name] = workloadMessage
	}
}

func (agent *MonitorAgent) buildDeploymentMessage(deployment appsV1.Deployment) (WorkloadHealthReport, error) {
	return agent.buildWorkloadMessage(WorkloadKindDeployment, deployment.Name, deployment.Spec.Replicas, deployment.Spec.Template.Spec.Containers, deployment.Status, deployment.Labels)
}

func (agent *MonitorAgent) buildDaemonSetMessage(daemon appsV1.DaemonSet) (WorkloadHealthReport, error) {
	return agent.buildWorkloadMessage(WorkloadKindDaemonSet, daemon.Name, nil, daemon.Spec.Template.Spec.Containers, daemon.Status, daemon.Labels)
}

func (agent *MonitorAgent) buildReplicaMessage(pod coreV1.Pod) (WorkloadReplicaHealthReport, error) {
	specContainers := make(map[string]ContainerHealthReport)
	for _, container := range pod.Spec.Containers {
		specContainers[container.Name] = ContainerHealthReport{
			Image: container.Image,
		}
	}

	status, err := json.Marshal(pod.Status)
	if err != nil {
		return WorkloadReplicaHealthReport{}, fmt.Errorf("error marshaling status for pod: %v", pod.Name)
	}

	spec := WorkloadReplicaSpecReport{
		Containers: specContainers,
	}

	return WorkloadReplicaHealthReport{
		Node:   pod.Spec.NodeName,
		Spec:   spec,
		Status: status,
	}, nil
}

func (agent *MonitorAgent) buildWorkloadMessage(workloadKind WorkloadKind, name string, replicas *int32, containers []coreV1.Container, statusObj interface{}, labels map[string]string) (WorkloadHealthReport, error) {
	containersReports := make(map[string]ContainerHealthReport)
	for _, container := range containers {
		containersReports[container.Name] = ContainerHealthReport{Image: container.Image}
	}

	status, err := json.Marshal(statusObj)
	if err != nil {
		return WorkloadHealthReport{}, fmt.Errorf("error marshaling status for %v: %v", workloadKind, name)
	}

	spec := WorkloadSpecReport{
		Containers: containersReports,
	}
	if replicas != nil {
		spec.Replicas = *replicas
	}

	return WorkloadHealthReport{
		Kind:            workloadKind,
		Spec:            spec,
		Status:          status,
		ReplicasReports: make(map[string]WorkloadReplicaHealthReport),
		Labels:          labels,
	}, nil

}

func (agent *MonitorAgent) updateWorkloadsReplicasWithPodsAndReplicaSets(pods map[string]coreV1.Pod, replicaSets map[string]appsV1.ReplicaSet, reports map[string]WorkloadHealthReport) {
	for _, pod := range pods {
		if len(pod.OwnerReferences) < 1 {
			//logger.Info("found pod with no parent", "pod", podName)
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
				//logger.Error(err, "error getting pod data", "pod", podName)
				continue
			}

			workloadMessage.ReplicasReports[pod.Name] = replicaMsg
		}
	}
}
