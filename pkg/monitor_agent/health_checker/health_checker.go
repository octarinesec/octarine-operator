package health_checker

import (
	"context"
	"github.com/go-logr/logr"
	admissions "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type HealthChecker struct {
	logger       logr.Logger
	namespace    string
	k8sClientset *kubernetes.Clientset
}

// Creates and returns list options for listing k8s resources
func (hc *HealthChecker) geK8sListOptions() []client.ListOption {
	return []client.ListOption{
		client.InNamespace(hc.namespace),
	}
}

func (hc *HealthChecker) GetPods() (map[string]corev1.Pod, error) {
	results := make(map[string]corev1.Pod)

	pods, err := hc.k8sClientset.CoreV1().Pods(hc.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		hc.logger.Error(err, "Error getting Pods")
		return nil, err
	}

	for _, pod := range pods.Items {
		results[pod.Name] = pod
	}

	return results, nil
}

func (hc *HealthChecker) GetReplicaSets() (map[string]appsv1.ReplicaSet, error) {
	results := make(map[string]appsv1.ReplicaSet)

	replicaSets, err := hc.k8sClientset.AppsV1().ReplicaSets(hc.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		hc.logger.Error(err, "Error getting ReplicaSets")
		return nil, err
	}

	for _, rs := range replicaSets.Items {
		results[rs.Name] = rs
	}

	return results, nil
}

func (hc *HealthChecker) GetDeployments() (map[string]appsv1.Deployment, error) {
	results := make(map[string]appsv1.Deployment)

	deps, err := hc.k8sClientset.AppsV1().Deployments(hc.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		hc.logger.Error(err, "Error getting Deployments")
		return nil, err
	}

	for _, dep := range deps.Items {
		results[dep.Name] = dep
	}

	return results, nil
}

func (hc *HealthChecker) GetDaemonSets() (map[string]appsv1.DaemonSet, error) {
	results := make(map[string]appsv1.DaemonSet)

	daemons, err := hc.k8sClientset.AppsV1().DaemonSets(hc.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		hc.logger.Error(err, "Error getting DaemosSets")
		return nil, err
	}

	for _, ds := range daemons.Items {
		results[ds.Name] = ds
	}

	return results, nil
}

func (hc *HealthChecker) GetValidatingWebhookConfigurations() (map[string]admissions.ValidatingWebhookConfiguration, error) {
	results := make(map[string]admissions.ValidatingWebhookConfiguration)

	webhooks, err := hc.k8sClientset.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		hc.logger.Error(err, "Error getting ValidatingWebhookConfiguration")
		return nil, err
	}

	for _, webhook := range webhooks.Items {
		results[webhook.Name] = webhook
	}

	return results, nil
}

func NewHealthChecker(logger logr.Logger, namespace string, k8sClientset *kubernetes.Clientset) *HealthChecker {
	return &HealthChecker{
		logger:       logger,
		namespace:    namespace,
		k8sClientset: k8sClientset,
	}
}
