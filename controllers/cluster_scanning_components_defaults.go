package controllers

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
)

const (
	// Source: https://github.com/containers/storage/blob/main/docs/containers-storage.conf.5.md

	crioDefaultStoragePathCRD = "/var/lib/containers/storage"
	crioDefaultConfigPathCRD  = "/etc/containers/storage.conf"
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

	if err := setDefaultResourceRequirements(&clusterScannerAgent.Resources, "64Mi", "30m", "4Gi", "2000m"); err != nil {
		return err
	}

	setDefaultFileProbes(&clusterScannerAgent.Probes)

	emptyK8sContainerEngineSpec := cbcontainersv1.K8sContainerEngineSpec{}
	if clusterScannerAgent.K8sContainerEngine != emptyK8sContainerEngineSpec {
		if err := validateClusterScannerK8sContainerEngineSpec(clusterScannerAgent.K8sContainerEngine); err != nil {
			return err
		}
	}

	if clusterScannerAgent.K8sContainerEngine.CRIO.ConfigPath == "" {
		clusterScannerAgent.K8sContainerEngine.CRIO.ConfigPath = crioDefaultConfigPathCRD
	}
	if clusterScannerAgent.K8sContainerEngine.CRIO.StoragePath == "" {
		clusterScannerAgent.K8sContainerEngine.CRIO.StoragePath = crioDefaultStoragePathCRD
	}

	return nil
}

func validateClusterScannerK8sContainerEngineSpec(spec cbcontainersv1.K8sContainerEngineSpec) error {
	if spec.Endpoint == "" {
		return fmt.Errorf("k8s container engine endpoint must be provided if configuring k8s container engine option")
	}

	if spec.EngineType == "" {
		return fmt.Errorf("k8s container engine type must be provided if configuring k8s container engine option")
	}

	for _, engineType := range cbcontainersv1.SupportedK8sEngineTypes {
		if engineType == spec.EngineType {
			return nil
		}
	}

	return fmt.Errorf("invalid engine type %v provided", spec.EngineType)
}
