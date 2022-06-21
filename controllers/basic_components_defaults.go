package controllers

import cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"

func (r *CBContainersAgentController) setBasicComponentsDefaults(basic *cbcontainersv1.CBContainersBasicSpec) error {
	if err := r.setMonitorDefaults(&basic.Monitor); err != nil {
		return err
	}

	if err := r.setEnforcerDefaults(&basic.Enforcer); err != nil {
		return err
	}

	if err := r.setStateReporterDefaults(&basic.StateReporter); err != nil {
		return err
	}

	return nil
}

func (r *CBContainersAgentController) setMonitorDefaults(monitor *cbcontainersv1.CBContainersMonitorSpec) error {
	if monitor.Labels == nil {
		monitor.Labels = make(map[string]string)
	}

	if monitor.DeploymentAnnotations == nil {
		monitor.DeploymentAnnotations = make(map[string]string)
	}

	if monitor.PodTemplateAnnotations == nil {
		monitor.PodTemplateAnnotations = make(map[string]string)
	}

	if monitor.Env == nil {
		monitor.Env = make(map[string]string)
	}

	setDefaultImage(&monitor.Image, "cbartifactory/monitor")

	if err := setDefaultResourceRequirements(&monitor.Resources, "64Mi", "30m", "256Mi", "200m"); err != nil {
		return err
	}

	setDefaultHTTPProbes(&monitor.Probes)

	return nil
}

func (r *CBContainersAgentController) setEnforcerDefaults(enforcer *cbcontainersv1.CBContainersEnforcerSpec) error {
	if enforcer.Labels == nil {
		enforcer.Labels = make(map[string]string)
	}

	if enforcer.DeploymentAnnotations == nil {
		enforcer.DeploymentAnnotations = make(map[string]string)
	}

	if enforcer.PodTemplateAnnotations == nil {
		enforcer.PodTemplateAnnotations = make(map[string]string)
	}

	if enforcer.Env == nil {
		enforcer.Env = make(map[string]string)
	}

	if enforcer.ReplicasCount == nil {
		defaultReplicaCount := int32(1)
		enforcer.ReplicasCount = &defaultReplicaCount
	}

	setDefaultPrometheus(&enforcer.Prometheus)

	setDefaultImage(&enforcer.Image, "cbartifactory/guardrails-enforcer")

	if err := setDefaultResourceRequirements(&enforcer.Resources, "64Mi", "30m", "256Mi", "200m"); err != nil {
		return err
	}

	setDefaultHTTPProbes(&enforcer.Probes)

	if enforcer.WebhookTimeoutSeconds == 0 {
		enforcer.WebhookTimeoutSeconds = 5
	}

	if enforcer.FailurePolicy == "" {
		enforcer.FailurePolicy = "Ignore"
	}

	if enforcer.EnableEnforcementFeature == nil {
		enforcer.EnableEnforcementFeature = &trueRef
	}

	return nil
}

func (r *CBContainersAgentController) setStateReporterDefaults(stateReporter *cbcontainersv1.CBContainersStateReporterSpec) error {
	if stateReporter.Labels == nil {
		stateReporter.Labels = make(map[string]string)
	}

	if stateReporter.DeploymentAnnotations == nil {
		stateReporter.DeploymentAnnotations = make(map[string]string)
	}

	if stateReporter.PodTemplateAnnotations == nil {
		stateReporter.PodTemplateAnnotations = make(map[string]string)
	}

	if stateReporter.Env == nil {
		stateReporter.Env = make(map[string]string)
	}

	setDefaultImage(&stateReporter.Image, "cbartifactory/guardrails-state-reporter")

	if err := setDefaultResourceRequirements(&stateReporter.Resources, "64Mi", "30m", "256Mi", "200m"); err != nil {
		return err
	}

	setDefaultHTTPProbes(&stateReporter.Probes)

	return nil
}
