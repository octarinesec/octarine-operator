package controllers

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
)

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

	if runtimeResolver.NodesToReplicasRatio == 0 {
		runtimeResolver.NodesToReplicasRatio = 5
	}

	setDefaultPrometheus(&runtimeResolver.Prometheus)

	setDefaultImage(&runtimeResolver.Image, "cbartifactory/runtime-kubernetes-resolver")

	if err := setDefaultResourceRequirements(&runtimeResolver.Resources, "512Mi", "200m", "2Gi", "900m"); err != nil {
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
