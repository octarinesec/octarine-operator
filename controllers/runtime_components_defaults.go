package controllers

import (
	"context"
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"math"
)

func getScaledReplicasCount() (*int32, error) {
	// Get the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting in-cluster config: %v", err)
	}

	// Create a Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating Kubernetes client: %v", err)
	}

	// Get the list of nodes in the cluster
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting list of nodes: %v", err)
	}

	nodesCount := int32(math.Ceil(float64(len(nodes.Items)) / 3))

	return &nodesCount, nil
}

func (r *CBContainersAgentController) setRuntimeProtectionComponentsDefaults(runtime *cbcontainersv1.CBContainersRuntimeProtectionSpec) error {
	if runtime.Enabled == nil {
		runtime.Enabled = &trueRef
	}

	if !(*runtime.Enabled) {
		return nil
	}

	if err := r.setRuntimeResolverDefaults(&runtime.Resolver); err != nil {
		return err
	}

	if err := r.setRuntimeSensorDefaults(&runtime.Sensor); err != nil {
		return err
	}

	if runtime.InternalGrpcPort == 0 {
		runtime.InternalGrpcPort = 8080
	}

	return nil
}

func (r *CBContainersAgentController) setRuntimeResolverDefaults(runtimeResolver *cbcontainersv1.CBContainersRuntimeResolverSpec) error {
	if runtimeResolver.Labels == nil {
		runtimeResolver.Labels = make(map[string]string)
	}

	if runtimeResolver.DeploymentAnnotations == nil {
		runtimeResolver.DeploymentAnnotations = make(map[string]string)
	}

	if runtimeResolver.PodTemplateAnnotations == nil {
		runtimeResolver.PodTemplateAnnotations = make(map[string]string)
	}

	if runtimeResolver.Env == nil {
		runtimeResolver.Env = make(map[string]string)
	}

	defaultReplicaCount := int32(1)
	replicasCount := &defaultReplicaCount

	nodesCount, err := getScaledReplicasCount()
	if err != nil {
		r.Log.Error(err, "failed to determine nodes count: %v, using replicas count defaults")
	} else {
		replicasCount = nodesCount
	}

	runtimeResolver.ReplicasCount = replicasCount

	setDefaultPrometheus(&runtimeResolver.Prometheus)

	setDefaultImage(&runtimeResolver.Image, "cbartifactory/runtime-kubernetes-resolver")

	if err := setDefaultResourceRequirements(&runtimeResolver.Resources, "64Mi", "200m", "1024Mi", "900m"); err != nil {
		return err
	}

	setDefaultHTTPProbes(&runtimeResolver.Probes)

	if runtimeResolver.LogLevel == "" {
		runtimeResolver.LogLevel = "info"
	}

	return nil
}

func (r *CBContainersAgentController) setRuntimeSensorDefaults(runtimeSensor *cbcontainersv1.CBContainersRuntimeSensorSpec) error {
	if runtimeSensor.Labels == nil {
		runtimeSensor.Labels = make(map[string]string)
	}

	if runtimeSensor.DaemonSetAnnotations == nil {
		runtimeSensor.DaemonSetAnnotations = make(map[string]string)
	}

	if runtimeSensor.PodTemplateAnnotations == nil {
		runtimeSensor.PodTemplateAnnotations = make(map[string]string)
	}

	if runtimeSensor.Env == nil {
		runtimeSensor.Env = make(map[string]string)
	}

	setDefaultPrometheusWithPort(&runtimeSensor.Prometheus, 7071)

	setDefaultImage(&runtimeSensor.Image, "cbartifactory/runtime-kubernetes-sensor")

	if err := setDefaultResourceRequirements(&runtimeSensor.Resources, "64Mi", "30m", "1024Mi", "500m"); err != nil {
		return err
	}

	setDefaultFileProbes(&runtimeSensor.Probes)

	if runtimeSensor.LogLevel == "" {
		runtimeSensor.LogLevel = "info"
	}

	return nil
}
