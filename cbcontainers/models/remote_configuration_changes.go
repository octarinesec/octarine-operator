package models

type ConfigurationChange struct {
	ID                                   string  `json:"id"`
	Status                               string  `json:"status"`
	AgentVersion                         *string `json:"agent_version"`
	EnableClusterScanning                *bool   `json:"enable_cluster_scanning"`
	EnableRuntime                        *bool   `json:"enable_runtime"`
	EnableCNDR                           *bool   `json:"enable_cndr"`
	EnableClusterScanningSecretDetection *bool   `json:"enable_cluster_scanning_secret_detection"`
	Timestamp                            string  `json:"timestamp"`
}

type ConfigurationChangeStatusUpdate struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Reason string `json:"reason"`
	// AppliedGeneration tracks the generation of the Custom resource where the change was applied
	AppliedGeneration int64 `json:"applied_generation"`
	// AppliedTimestamp records when the change was applied in RFC3339 format
	AppliedTimestamp  string `json:"applied_timestamp"`
	ClusterIdentifier string `json:"cluster_identifier"`
	ClusterGroup      string `json:"cluster_group"`
	ClusterName       string `json:"cluster_name"`
}
