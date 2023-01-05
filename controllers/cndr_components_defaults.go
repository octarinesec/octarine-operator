package controllers

import cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"

func (r *CBContainersAgentController) setCndrComponentsDefaults(cndr *cbcontainersv1.CBContainersCndrSpec) error {
	if cndr == nil {
		return nil
	}

	if cndr.Enabled == nil {
		cndr.Enabled = &falseRef
	}

	if cndr.CompanyCodeSecretName == "" {
		cndr.CompanyCodeSecretName = defaultCompanyCode
	}

	if !(*cndr.Enabled) {
		return nil
	}

	if err := r.setCndrSensorDefaults(&cndr.Sensor); err != nil {
		return err
	}

	return nil
}

func (r *CBContainersAgentController) setCndrSensorDefaults(cndrSensor *cbcontainersv1.CBContainersCndrSensorSpec) error {
	if cndrSensor.Labels == nil {
		cndrSensor.Labels = make(map[string]string)
	}

	if cndrSensor.DaemonSetAnnotations == nil {
		cndrSensor.DaemonSetAnnotations = make(map[string]string)
	}

	if cndrSensor.PodTemplateAnnotations == nil {
		cndrSensor.PodTemplateAnnotations = make(map[string]string)
	}

	if cndrSensor.Env == nil {
		cndrSensor.Env = make(map[string]string)
	}

	setDefaultPrometheusWithPort(&cndrSensor.Prometheus, 7071)

	setDefaultImage(&cndrSensor.Image, "cbartifactory/cndr")

	if err := setDefaultResourceRequirements(&cndrSensor.Resources, "64Mi", "30m", "1024Mi", "500m"); err != nil {
		return err
	}

	if cndrSensor.LogLevel == "" {
		cndrSensor.LogLevel = "info"
	}

	return nil
}
