package controllers

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
)

func (r *CBContainersAgentController) setClusterScanningComponentsDefaults(clusterScanning *cbcontainersv1.CBContainersClusterScanningSpec) error {
	if clusterScanning.Enabled == nil {
		clusterScanning.Enabled = &trueRef
	}

	if !(*clusterScanning.Enabled) {
		return nil
	}

	if err := r.setImageScanningReporterDefaults(&clusterScanning.ImageScanningReporter); err != nil {
		return err
	}

	if err := r.setClusterScannerAgentDefaults(&clusterScanning.ClusterScannerAgent); err != nil {
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

func (r *CBContainersAgentController) setClusterScannerAgentDefaults(clusterScannerAgent *cbcontainersv1.CBContainersClusterScannerAgentSpec) error {
	if clusterScannerAgent.Labels == nil {
		clusterScannerAgent.Labels = make(map[string]string)
	}

	if clusterScannerAgent.DaemonSetAnnotations == nil {
		clusterScannerAgent.DaemonSetAnnotations = make(map[string]string)
	}

	if clusterScannerAgent.PodTemplateAnnotations == nil {
		clusterScannerAgent.PodTemplateAnnotations = make(map[string]string)
	}

	if clusterScannerAgent.Env == nil {
		clusterScannerAgent.Env = make(map[string]string)
	}

	setDefaultPrometheusWithPort(&clusterScannerAgent.Prometheus, 7072)

	setDefaultImage(&clusterScannerAgent.Image, "cbartifactory/cluster-scanner")

	if err := setDefaultResourceRequirements(&clusterScannerAgent.Resources, "64Mi", "30m", "6Gi", "2000m"); err != nil {
		return err
	}

	setDefaultFileProbes(&clusterScannerAgent.Probes)

	if err := validateClusterScannerK8sContainerEngineSpec(clusterScannerAgent.K8sContainerEngine); err != nil {
		return err
	}

	return nil
}

func validateClusterScannerK8sContainerEngineSpec(spec cbcontainersv1.K8sContainerEngineSpec) error {
	if spec.Endpoint == "" && spec.EngineType != "" {
		return fmt.Errorf("k8s container engine endpoint must be provided if the k8s container engine type has been set")
	}

	if spec.Endpoint != "" && spec.EngineType == "" {
		return fmt.Errorf("k8s container engine type must be provided if the container endpoint has been set")
	}

	if spec.EngineType != "" {
		found := false
		for _, engineType := range cbcontainersv1.SupportedK8sEngineTypes {
			if engineType == spec.EngineType {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid engine type %v provided", spec.EngineType)
		}
	}

	return nil
}
