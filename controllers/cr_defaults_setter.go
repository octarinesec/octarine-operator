package controllers

import cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"

func (r *CBContainersAgentController) setDefaults(cbContainersAgent *cbcontainersv1.CBContainersAgent) error {
	if cbContainersAgent.Spec.ApiGatewaySpec.Scheme == "" {
		cbContainersAgent.Spec.ApiGatewaySpec.Scheme = "https"
	}

	if cbContainersAgent.Spec.ApiGatewaySpec.Port == 0 {
		cbContainersAgent.Spec.ApiGatewaySpec.Port = 443
	}

	if cbContainersAgent.Spec.ApiGatewaySpec.Adapter == "" {
		cbContainersAgent.Spec.ApiGatewaySpec.Adapter = "containers"
	}

	if cbContainersAgent.Spec.ApiGatewaySpec.AccessTokenSecretName == "" {
		cbContainersAgent.Spec.ApiGatewaySpec.AccessTokenSecretName = defaultAccessToken
	}

	if cbContainersAgent.Spec.GatewayTLS.RootCAsBundle == nil {
		cbContainersAgent.Spec.GatewayTLS.RootCAsBundle = make([]byte, 0)
	}

	if err := r.setClusterDefaults(&cbContainersAgent.Spec.CoreSpec); err != nil {
		return err
	}

	if err := r.setHardeningDefaults(&cbContainersAgent.Spec.HardeningSpec); err != nil {
		return err
	}

	if err := r.setRuntimeDefaults(&cbContainersAgent.Spec.RuntimeSpec); err != nil {
		return err
	}

	return nil
}

func (r *CBContainersAgentController) setClusterDefaults(cbContainersClusterSpec *cbcontainersv1.CBContainersCoreSpec) error {
	if cbContainersClusterSpec.EventsGatewaySpec.Port == 0 {
		cbContainersClusterSpec.EventsGatewaySpec.Port = 443
	}

	if cbContainersClusterSpec.MonitorSpec.Labels == nil {
		cbContainersClusterSpec.MonitorSpec.Labels = make(map[string]string)
	}

	if cbContainersClusterSpec.MonitorSpec.DeploymentAnnotations == nil {
		cbContainersClusterSpec.MonitorSpec.DeploymentAnnotations = make(map[string]string)
	}

	if cbContainersClusterSpec.MonitorSpec.PodTemplateAnnotations == nil {
		cbContainersClusterSpec.MonitorSpec.PodTemplateAnnotations = make(map[string]string)
	}

	if cbContainersClusterSpec.MonitorSpec.Env == nil {
		cbContainersClusterSpec.MonitorSpec.Env = make(map[string]string)
	}

	setDefaultImage(&cbContainersClusterSpec.MonitorSpec.Image, "cbartifactory/monitor")

	if err := setDefaultResourceRequirements(&cbContainersClusterSpec.MonitorSpec.Resources, "64Mi", "30m", "256Mi", "200m"); err != nil {
		return err
	}

	setDefaultHTTPProbes(&cbContainersClusterSpec.MonitorSpec.Probes)

	return nil
}

func (r *CBContainersAgentController) setHardeningDefaults(cbContainersHardeningSpec *cbcontainersv1.CBContainersHardeningSpec) error {
	if cbContainersHardeningSpec.EnforcerSpec.Labels == nil {
		cbContainersHardeningSpec.EnforcerSpec.Labels = make(map[string]string)
	}

	if cbContainersHardeningSpec.EnforcerSpec.DeploymentAnnotations == nil {
		cbContainersHardeningSpec.EnforcerSpec.DeploymentAnnotations = make(map[string]string)
	}

	if cbContainersHardeningSpec.EnforcerSpec.PodTemplateAnnotations == nil {
		cbContainersHardeningSpec.EnforcerSpec.PodTemplateAnnotations = make(map[string]string)
	}

	if cbContainersHardeningSpec.EnforcerSpec.ReplicasCount == nil {
		defaultReplicaCount := int32(1)
		cbContainersHardeningSpec.EnforcerSpec.ReplicasCount = &defaultReplicaCount
	}

	if cbContainersHardeningSpec.EnforcerSpec.Env == nil {
		cbContainersHardeningSpec.EnforcerSpec.Env = make(map[string]string)
	}

	setDefaultPrometheus(&cbContainersHardeningSpec.EnforcerSpec.Prometheus)

	setDefaultImage(&cbContainersHardeningSpec.EnforcerSpec.Image, "cbartifactory/guardrails-enforcer")

	if err := setDefaultResourceRequirements(&cbContainersHardeningSpec.EnforcerSpec.Resources, "64Mi", "30m", "256Mi", "200m"); err != nil {
		return err
	}

	setDefaultHTTPProbes(&cbContainersHardeningSpec.EnforcerSpec.Probes)

	if cbContainersHardeningSpec.EnforcerSpec.WebhookTimeoutSeconds == 0 {
		cbContainersHardeningSpec.EnforcerSpec.WebhookTimeoutSeconds = 5
	}

	if cbContainersHardeningSpec.StateReporterSpec.Labels == nil {
		cbContainersHardeningSpec.StateReporterSpec.Labels = make(map[string]string)
	}

	if cbContainersHardeningSpec.StateReporterSpec.DeploymentAnnotations == nil {
		cbContainersHardeningSpec.StateReporterSpec.DeploymentAnnotations = make(map[string]string)
	}

	if cbContainersHardeningSpec.StateReporterSpec.PodTemplateAnnotations == nil {
		cbContainersHardeningSpec.StateReporterSpec.PodTemplateAnnotations = make(map[string]string)
	}

	if cbContainersHardeningSpec.StateReporterSpec.Env == nil {
		cbContainersHardeningSpec.StateReporterSpec.Env = make(map[string]string)
	}

	setDefaultImage(&cbContainersHardeningSpec.StateReporterSpec.Image, "cbartifactory/guardrails-state-reporter")

	if err := setDefaultResourceRequirements(&cbContainersHardeningSpec.StateReporterSpec.Resources, "64Mi", "30m", "256Mi", "200m"); err != nil {
		return err
	}

	setDefaultHTTPProbes(&cbContainersHardeningSpec.StateReporterSpec.Probes)

	if cbContainersHardeningSpec.EventsGatewaySpec.Port == 0 {
		cbContainersHardeningSpec.EventsGatewaySpec.Port = 443
	}

	return nil
}

func (r *CBContainersAgentController) setRuntimeDefaults(cbContainersRuntimeSpec *cbcontainersv1.CBContainersRuntimeSpec) error {
	if cbContainersRuntimeSpec.ResolverSpec.Labels == nil {
		cbContainersRuntimeSpec.ResolverSpec.Labels = make(map[string]string)
	}

	if cbContainersRuntimeSpec.ResolverSpec.DeploymentAnnotations == nil {
		cbContainersRuntimeSpec.ResolverSpec.DeploymentAnnotations = make(map[string]string)
	}

	if cbContainersRuntimeSpec.ResolverSpec.PodTemplateAnnotations == nil {
		cbContainersRuntimeSpec.ResolverSpec.PodTemplateAnnotations = make(map[string]string)
	}

	if cbContainersRuntimeSpec.ResolverSpec.ReplicasCount == nil {
		defaultReplicaCount := int32(1)
		cbContainersRuntimeSpec.ResolverSpec.ReplicasCount = &defaultReplicaCount
	}

	if cbContainersRuntimeSpec.ResolverSpec.Env == nil {
		cbContainersRuntimeSpec.ResolverSpec.Env = make(map[string]string)
	}

	if cbContainersRuntimeSpec.ResolverSpec.EventsGatewaySpec.Port == 0 {
		cbContainersRuntimeSpec.ResolverSpec.EventsGatewaySpec.Port = 443
	}

	setDefaultPrometheus(&cbContainersRuntimeSpec.ResolverSpec.Prometheus)

	setDefaultImage(&cbContainersRuntimeSpec.ResolverSpec.Image, "cbartifactory/runtime-kubernetes-resolver")

	if err := setDefaultResourceRequirements(&cbContainersRuntimeSpec.ResolverSpec.Resources, "64Mi", "200m", "1024Mi", "900m"); err != nil {
		return err
	}

	setDefaultHTTPProbes(&cbContainersRuntimeSpec.ResolverSpec.Probes)

	if cbContainersRuntimeSpec.SensorSpec.Labels == nil {
		cbContainersRuntimeSpec.SensorSpec.Labels = make(map[string]string)
	}

	if cbContainersRuntimeSpec.SensorSpec.DaemonSetAnnotations == nil {
		cbContainersRuntimeSpec.SensorSpec.DaemonSetAnnotations = make(map[string]string)
	}

	if cbContainersRuntimeSpec.SensorSpec.PodTemplateAnnotations == nil {
		cbContainersRuntimeSpec.SensorSpec.PodTemplateAnnotations = make(map[string]string)
	}

	if cbContainersRuntimeSpec.SensorSpec.Env == nil {
		cbContainersRuntimeSpec.SensorSpec.Env = make(map[string]string)
	}

	setDefaultPrometheus(&cbContainersRuntimeSpec.SensorSpec.Prometheus)

	setDefaultImage(&cbContainersRuntimeSpec.SensorSpec.Image, "cbartifactory/runtime-kubernetes-sensor")

	if err := setDefaultResourceRequirements(&cbContainersRuntimeSpec.SensorSpec.Resources, "64Mi", "30m", "1024Mi", "500m"); err != nil {
		return err
	}

	setDefaultFileProbes(&cbContainersRuntimeSpec.SensorSpec.Probes)

	if cbContainersRuntimeSpec.SensorSpec.VerbosityLevel == nil {
		defaultVerbosity := 2
		cbContainersRuntimeSpec.SensorSpec.VerbosityLevel = &defaultVerbosity
	}

	if cbContainersRuntimeSpec.InternalGrpcPort == 0 {
		cbContainersRuntimeSpec.InternalGrpcPort = 443
	}

	return nil
}
