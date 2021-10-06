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

	// TODO - check if the clusterScanning should have InternalGrpcPort as param
	//if clusterScanning.InternalGrpcPort == 0 {
	//	clusterScanning.InternalGrpcPort = 443
	//}

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

	// TODO - check if the sevice uses the same default liveness and readiness
	//setDefaultHTTPProbes(&imageScanningReporter.Probes)

	return nil
}