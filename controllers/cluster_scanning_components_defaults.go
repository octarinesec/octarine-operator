package controllers

import cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"

func (r *CBContainersAgentController) setClusterScanningComponentsDefaults(clusterScanning *cbcontainersv1.CBContainersClusterScanningSpec) error {
	if clusterScanning.Enabled == nil {
		clusterScanning.Enabled = &falseRef
	}

	if !(*clusterScanning.Enabled) {
		return nil
	}

	if err := r.setImageScanningReporterDefaults(&clusterScanning.ImageScanningReporter); err != nil {
		return err
	}

	if err := r.setClusterScanningSensorDefaults(&clusterScanning.ClusterScanningSensor); err != nil {
		return err
	}

	return nil
}

func (r *CBContainersAgentController) setImageScanningReporterDefaults(imageScanningReporter *cbcontainersv1.CBContainersImageScanningReporterSpec) error {
	if imageScanningReporter.Labels == nil {
		imageScanningReporter.Labels = make(map[string]string)
	}

	if imageScanningReporter.DeploymentAnnotations == nil {
		imageScanningReporter.DeploymentAnnotations = make(map[string]string)
	}

	if imageScanningReporter.PodTemplateAnnotations == nil {
		imageScanningReporter.PodTemplateAnnotations = make(map[string]string)
	}

	if imageScanningReporter.Env == nil {
		imageScanningReporter.Env = make(map[string]string)
	}

	if imageScanningReporter.ReplicasCount == nil {
		defaultReplicaCount := int32(1)
		imageScanningReporter.ReplicasCount = &defaultReplicaCount
	}

	setDefaultPrometheus(&imageScanningReporter.Prometheus)

	setDefaultImage(&imageScanningReporter.Image, "cbartifactory/image-scanning-reporter")

	if err := setDefaultResourceRequirements(&imageScanningReporter.Resources, "64Mi", "200m", "1024Mi", "900m"); err != nil {
		return err
	}

	setDefaultHTTPProbes(&imageScanningReporter.Probes)

	return nil
}

func (r *CBContainersAgentController) setClusterScanningSensorDefaults(clusterScanningSensor *cbcontainersv1.CBContainersClusterScanningSensorSpec) error {
	if clusterScanningSensor.Labels == nil {
		clusterScanningSensor.Labels = make(map[string]string)
	}

	if clusterScanningSensor.DaemonSetAnnotations == nil {
		clusterScanningSensor.DaemonSetAnnotations = make(map[string]string)
	}

	if clusterScanningSensor.PodTemplateAnnotations == nil {
		clusterScanningSensor.PodTemplateAnnotations = make(map[string]string)
	}

	if clusterScanningSensor.Env == nil {
		clusterScanningSensor.Env = make(map[string]string)
	}

	setDefaultPrometheus(&clusterScanningSensor.Prometheus)

	setDefaultImage(&clusterScanningSensor.Image, "cbartifactory/cluster-scanner")

	if err := setDefaultResourceRequirements(&clusterScanningSensor.Resources, "64Mi", "30m", "1024Mi", "500m"); err != nil {
		return err
	}

	setDefaultFileProbes(&clusterScanningSensor.Probes)

	if clusterScanningSensor.VerbosityLevel == nil {
		defaultVerbosity := 2
		clusterScanningSensor.VerbosityLevel = &defaultVerbosity
	}

	return nil
}
