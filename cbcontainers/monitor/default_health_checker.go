package monitor

import (
	"context"
	"fmt"
	admissionsV1 "k8s.io/api/admissionregistration/v1"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DefaultHealthChecker struct {
	client    client.Client
	namespace string
}

func NewDefaultHealthChecker(client client.Client, namespace string) *DefaultHealthChecker {
	return &DefaultHealthChecker{
		client:    client,
		namespace: namespace,
	}
}

func (healthChecker *DefaultHealthChecker) listClusterObjects(list client.ObjectList) error {
	return healthChecker.list(list, &client.ListOptions{Namespace: ""})
}

func (healthChecker *DefaultHealthChecker) listNamespaceObjects(list client.ObjectList) error {
	return healthChecker.list(list, &client.ListOptions{Namespace: healthChecker.namespace})
}

func (healthChecker *DefaultHealthChecker) list(list client.ObjectList, option client.ListOption) error {
	return healthChecker.client.List(context.Background(), list, option)
}

func (healthChecker *DefaultHealthChecker) GetPods() (map[string]coreV1.Pod, error) {
	podsList := &coreV1.PodList{}
	if err := healthChecker.listNamespaceObjects(podsList); err != nil {
		return nil, err
	}

	if podsList.Items == nil {
		return nil, fmt.Errorf("got nil for Pods list")
	}

	pods := make(map[string]coreV1.Pod)
	for _, pod := range podsList.Items {
		pods[pod.Name] = pod
	}

	return pods, nil
}

func (healthChecker *DefaultHealthChecker) GetReplicaSets() (map[string]appsV1.ReplicaSet, error) {
	replicaSetList := &appsV1.ReplicaSetList{}
	if err := healthChecker.listNamespaceObjects(replicaSetList); err != nil {
		return nil, err
	}

	if replicaSetList.Items == nil {
		return nil, fmt.Errorf("got nil for ReplicaSets list")
	}

	replicaSets := make(map[string]appsV1.ReplicaSet)
	for _, replicaSet := range replicaSetList.Items {
		replicaSets[replicaSet.Name] = replicaSet
	}

	return replicaSets, nil
}

func (healthChecker *DefaultHealthChecker) GetDeployments() (map[string]appsV1.Deployment, error) {
	deploymentList := &appsV1.DeploymentList{}
	if err := healthChecker.listNamespaceObjects(deploymentList); err != nil {
		return nil, err
	}

	if deploymentList.Items == nil {
		return nil, fmt.Errorf("got nil for Deployments list")
	}

	deployments := make(map[string]appsV1.Deployment)
	for _, deployment := range deploymentList.Items {
		deployments[deployment.Name] = deployment
	}

	return deployments, nil
}

func (healthChecker *DefaultHealthChecker) GetDaemonSets() (map[string]appsV1.DaemonSet, error) {
	daemonSetList := &appsV1.DaemonSetList{}
	if err := healthChecker.listNamespaceObjects(daemonSetList); err != nil {
		return nil, err
	}

	if daemonSetList.Items == nil {
		return nil, fmt.Errorf("got nil for DaemonSets list")
	}

	daemonSets := make(map[string]appsV1.DaemonSet)
	for _, daemonSet := range daemonSetList.Items {
		daemonSets[daemonSet.Name] = daemonSet
	}

	return daemonSets, nil
}

func (healthChecker *DefaultHealthChecker) GetValidatingWebhookConfigurations() (map[string]admissionsV1.ValidatingWebhookConfiguration, error) {
	validatingWebhookConfigurationList := &admissionsV1.ValidatingWebhookConfigurationList{}
	if err := healthChecker.listClusterObjects(validatingWebhookConfigurationList); err != nil {
		return nil, err
	}

	if validatingWebhookConfigurationList.Items == nil {
		return nil, fmt.Errorf("got nil for ValidatingWebhookConfigurations list")
	}

	validatingWebhookConfigurations := make(map[string]admissionsV1.ValidatingWebhookConfiguration)
	for _, validatingWebhookConfiguration := range validatingWebhookConfigurationList.Items {
		validatingWebhookConfigurations[validatingWebhookConfiguration.Name] = validatingWebhookConfiguration
	}

	return validatingWebhookConfigurations, nil
}
