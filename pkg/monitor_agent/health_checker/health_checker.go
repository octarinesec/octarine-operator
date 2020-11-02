package health_checker

import (
	"context"
	"github.com/go-logr/logr"
	admissions "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type HealthChecker struct {
	logger    logr.Logger
	namespace string
	k8sClient client.Client
}

// Creates and returns list options for listing k8s resources
func (hc *HealthChecker) geK8sListOptions() []client.ListOption {
	return []client.ListOption{
		client.InNamespace(hc.namespace),
	}
}

func (hc *HealthChecker) GetPods() (map[string]corev1.Pod, error) {
	results := make(map[string]corev1.Pod)
	foundPods := &corev1.PodList{}
	if err := hc.k8sClient.List(context.TODO(), foundPods, hc.geK8sListOptions()...); err != nil && !k8serr.IsNotFound(err) {
		hc.logger.Error(err, "Error getting Pods")
		return nil, err
	} else if err == nil {
		for _, pod := range foundPods.Items {
			results[pod.Name] = pod
		}
	}
	return results, nil
}

func (hc *HealthChecker) GetReplicaSets() (map[string]appsv1.ReplicaSet, error) {
	results := make(map[string]appsv1.ReplicaSet)
	foundRS := &appsv1.ReplicaSetList{}
	if err := hc.k8sClient.List(context.TODO(), foundRS, hc.geK8sListOptions()...); err != nil && !k8serr.IsNotFound(err) {
		hc.logger.Error(err, "Error getting ReplicaSets")
		return nil, err
	} else if err == nil {
		for _, rs := range foundRS.Items {
			results[rs.Name] = rs
		}
	}
	return results, nil
}

func (hc *HealthChecker) GetDeployments() (map[string]appsv1.Deployment, error) {
	results := make(map[string]appsv1.Deployment)
	foundDeps := &appsv1.DeploymentList{}
	if err := hc.k8sClient.List(context.TODO(), foundDeps, hc.geK8sListOptions()...); err != nil && !k8serr.IsNotFound(err) {
		hc.logger.Error(err, "Error getting Deployments")
		return nil, err
	} else if err == nil {
		for _, dep := range foundDeps.Items {
			results[dep.Name] = dep
		}
	}
	return results, nil
}

func (hc *HealthChecker) GetDaemonSets() (map[string]appsv1.DaemonSet, error) {
	results := make(map[string]appsv1.DaemonSet)
	foundDS := &appsv1.DaemonSetList{}
	if err := hc.k8sClient.List(context.TODO(), foundDS, hc.geK8sListOptions()...); err != nil && !k8serr.IsNotFound(err) {
		hc.logger.Error(err, "Error getting DaemosSets")
		return nil, err
	} else if err == nil {
		for _, ds := range foundDS.Items {
			results[ds.Name] = ds
		}
	}
	return results, nil
}

func (hc *HealthChecker) GetValidatingWebhookConfigurations() (map[string]admissions.ValidatingWebhookConfiguration, error) {
	results := make(map[string]admissions.ValidatingWebhookConfiguration)
	foundWenhooks := &admissions.ValidatingWebhookConfigurationList{}
	if err := hc.k8sClient.List(context.TODO(), foundWenhooks); err != nil && !k8serr.IsNotFound(err) {
		hc.logger.Error(err, "Error getting ValidatingWebhookConfiguration")
		return nil, err
	} else if err == nil {
		for _, webhook := range foundWenhooks.Items {
			results[webhook.Name] = webhook
		}
	}
	return results, nil
}

func NewHealthChecker(logger logr.Logger, namespace string, k8sClient client.Client) *HealthChecker {
	return &HealthChecker{
		logger:    logger,
		namespace: namespace,
		k8sClient: k8sClient,
	}
}
