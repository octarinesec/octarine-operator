package models

type RemoteChangeStatus string

var (
	ChangeStatusPending RemoteChangeStatus = "PENDING"
	ChangeStatusAcked   RemoteChangeStatus = "ACKNOWLEDGED"
	ChangeStatusFailed  RemoteChangeStatus = "FAILED"
)

type AdvancedSettings struct {
	ProxyVersion   *string `json:"proxy_version,omitempty"`
	RegistryServer *string `json:"registry_server,omitempty"`
}

type ConfigurationChange struct {
	ID               string             `json:"id"`
	Status           RemoteChangeStatus `json:"status"`
	AgentVersion     *string            `json:"agent_version"`
	AdvancedSettings *AdvancedSettings  `json:"advancedSettings,omitempty"`
	Timestamp        string             `json:"timestamp"`
}

type ConfigurationChangeStatusUpdate struct {
	ID                string             `json:"id"`
	ClusterIdentifier string             `json:"cluster_identifier"`
	ClusterGroup      string             `json:"cluster_group"`
	ClusterName       string             `json:"cluster_name"`
	Status            RemoteChangeStatus `json:"status"`

	// AppliedGeneration tracks the generation of the Custom resource where the change was applied
	AppliedGeneration int64 `json:"applied_generation"`
	// AppliedTimestamp records when the change was applied in RFC3339 format
	AppliedTimestamp string `json:"applied_timestamp"`

	// Error should hold information about encountered errors when the change application failed.
	// For system usage only, not meant for end-users.
	Error string `json:"encountered_error"`
	// ErrorReason should be populated if some additional information can be shown to the user (e.g. why a change was invalid)
	// It should not be used to store system information
	ErrorReason string `json:"error_reason"`
}
