package models

type SensorMetadata struct {
	Version                 string `json:"version"`
	IsLatest                bool   `json:"is_latest" `
	SupportsRuntime         bool   `json:"supports_runtime"`
	SupportsClusterScanning bool   `json:"supports_cluster_scanning"`
	SupportsCndr            bool   `json:"supports_cndr"`
}
