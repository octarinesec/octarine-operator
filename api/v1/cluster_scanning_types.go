package v1

import coreV1 "k8s.io/api/core/v1"

type CBContainersImageScanningReporterSpec struct {
	// +kubebuilder:default:=<>
	Labels map[string]string `json:"labels,omitempty"`
	// +kubebuilder:default:=<>
	DeploymentAnnotations map[string]string `json:"deploymentAnnotations,omitempty"`
	// +kubebuilder:default:={prometheus.io/scrape: "false", prometheus.io/port: "7071"}
	PodTemplateAnnotations map[string]string `json:"podTemplateAnnotations,omitempty"`
	// +kubebuilder:default:=1
	ReplicasCount *int32 `json:"replicasCount,omitempty"`
	// +kubebuilder:default:=<>
	Env map[string]string `json:"env,omitempty"`
	// +kubebuilder:default:={repository:"cbartifactory/image-scanning-reporter"}
	Image CBContainersImageSpec `json:"image,omitempty"`
	// +kubebuilder:default:={requests: {memory: "64Mi", cpu: "200m"}, limits: {memory: "1024Mi", cpu: "900m"}}
	Resources coreV1.ResourceRequirements `json:"resources,omitempty"`
	// +kubebuilder:default:=<>
	Probes CBContainersHTTPProbesSpec `json:"probes,omitempty"`
	// +kubebuilder:default:=<>
	Prometheus CBContainersPrometheusSpec `json:"prometheus,omitempty"`
}

type CBContainersClusterScannerAgentSpec struct {
	// +kubebuilder:default:=<>
	Labels map[string]string `json:"labels,omitempty"`
	// +kubebuilder:default:=<>
	DaemonSetAnnotations map[string]string `json:"daemonSetAnnotations,omitempty"`
	// +kubebuilder:default:={prometheus.io/scrape: "false", prometheus.io/port: "7071"}
	PodTemplateAnnotations map[string]string `json:"podTemplateAnnotations,omitempty"`
	// +kubebuilder:default:=<>
	Env map[string]string `json:"env,omitempty"`
	// +kubebuilder:default:={repository:"cbartifactory/cluster-scanner"}
	Image CBContainersImageSpec `json:"image,omitempty"`
	// +kubebuilder:default:={requests: {memory: "256Mi", cpu: "30m"}, limits: {memory: "4Gi", cpu: "500m"}}
	Resources coreV1.ResourceRequirements `json:"resources,omitempty"`
	// +kubebuilder:default:=<>
	Probes CBContainersFileProbesSpec `json:"probes,omitempty"`
	// +kubebuilder:default:=<>
	Prometheus CBContainersPrometheusSpec `json:"prometheus,omitempty"`
}

// CBContainersClusterScanningSpec defines the desired state of CBContainersClusterScanning
type CBContainersClusterScanningSpec struct {
	// +kubebuilder:default:=false
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:default:=<>
	ImageScanningReporter CBContainersImageScanningReporterSpec `json:"imageScanningReporter,omitempty"`
	// +kubebuilder:default:=<>
	ClusterScannerAgent CBContainersClusterScannerAgentSpec `json:"clusterScanner,omitempty"`
}
