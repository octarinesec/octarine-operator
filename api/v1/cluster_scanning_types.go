package v1

import coreV1 "k8s.io/api/core/v1"

type CBContainersImageScanningReporterSpec struct {
	// +kubebuilder:default:=<>
	Labels map[string]string `json:"labels,omitempty"`
	// +kubebuilder:default:=<>
	DeploymentAnnotations map[string]string `json:"deploymentAnnotations,omitempty"`
	// +kubebuilder:default:=<>
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
	// +kubebuilder:default:={port: 7071}
	Prometheus CBContainersPrometheusSpec `json:"prometheus,omitempty"`
	// +kubebuilder:default:=<>
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// +kubebuilder:default:=<>
	Affinity *coreV1.Affinity `json:"affinity,omitempty"`
}

type CBContainersClusterScannerAgentSpec struct {
	// +kubebuilder:default:=<>
	Labels map[string]string `json:"labels,omitempty"`
	// +kubebuilder:default:=<>
	DaemonSetAnnotations map[string]string `json:"daemonSetAnnotations,omitempty"`
	// +kubebuilder:default:=<>
	PodTemplateAnnotations map[string]string `json:"podTemplateAnnotations,omitempty"`
	// +kubebuilder:default:=<>
	Env map[string]string `json:"env,omitempty"`
	// +kubebuilder:default:={repository:"cbartifactory/cluster-scanner"}
	Image CBContainersImageSpec `json:"image,omitempty"`
	// +kubebuilder:default:={requests: {memory: "64Mi", cpu: "30m"}, limits: {memory: "4Gi", cpu: "2000m"}}
	Resources coreV1.ResourceRequirements `json:"resources,omitempty"`
	// +kubebuilder:default:=<>
	Probes CBContainersFileProbesSpec `json:"probes,omitempty"`
	// +kubebuilder:default:={port: 7072}
	Prometheus CBContainersPrometheusSpec `json:"prometheus,omitempty"`
	// +kubebuilder:default:=<>
	K8sContainerEngine K8sContainerEngineSpec `json:"k8sContainerEngine,omitempty"`
	// +kubebuilder:default:=<>
	CLIFlags CLIFlags `json:"cliFlags,omitempty"`
}

type CLIFlags struct {
	// +kubebuilder:default:=false
	EnableSecretDetection bool `json:"enableSecretDetection,omitempty"`
	// +kubebuilder:default:={}
	SkipDirsOrFiles []string `json:"skipDirsOrFiles,omitempty"`
	// +kubebuilder:default:=false
	ScanBaseLayers bool `json:"scanBaseLayers,omitempty"`
	// +kubebuilder:default:=false
	IgnoreBuildInRegex bool `json:"ignoreBuildInRegex,omitempty"`
}

// CBContainersClusterScanningSpec defines the desired state of CBContainersClusterScanning
type CBContainersClusterScanningSpec struct {
	// +kubebuilder:default:=true
	Enabled *bool `json:"enabled,omitempty"`
	// +kubebuilder:default:=<>
	ImageScanningReporter CBContainersImageScanningReporterSpec `json:"imageScanningReporter,omitempty"`
	// +kubebuilder:default:=<>
	ClusterScannerAgent CBContainersClusterScannerAgentSpec `json:"clusterScanner,omitempty"`
}
