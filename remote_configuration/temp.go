package remote_configuration

import (
	"context"
	"math/rand"
	"strconv"
)

var versions = []string{"2.12.1", "2.10.0", "2.12.0", "2.11.0", "3.0.0"}

var (
	tr  = true
	fal = false
)

type DummyAPI struct {
}

func (d DummyAPI) GetConfigurationChanges(ctx context.Context) ([]ConfigurationChange, error) {
	c := RandomChange()
	if c != nil {
		return []ConfigurationChange{*c}, nil

	}
	return nil, nil
}

func (d DummyAPI) UpdateConfigurationChangeStatus(ctx context.Context, update ConfigurationChangeStatusUpdate) error {
	return nil
}

func RandomNonNilChange() *ConfigurationChange {
	for {
		c := RandomChange()
		if c != nil {
			return c
		}
	}
}

func RandomChange() *ConfigurationChange {
	csRand, runtimeRand, cndrRand, versionRand := rand.Int(), rand.Int(), rand.Int(), rand.Intn(len(versions)+1)

	//csRand, runtimeRand, versionRand = 1, 2, 3
	if versionRand == len(versions) {
		return nil
	}

	changeVersion := &versions[versionRand]

	var changeClusterScanning *bool
	var changeRuntime *bool
	var changeCNDR *bool

	switch csRand % 5 {
	case 1, 3:
		changeClusterScanning = &tr
	case 2, 4:
		changeClusterScanning = &fal
	default:
		changeClusterScanning = nil
	}

	switch runtimeRand % 5 {
	case 1, 3:
		changeRuntime = &tr
	case 2, 4:
		changeRuntime = &fal
	default:
		changeRuntime = nil
	}

	if changeVersion != nil && *changeVersion == "3.0.0" && cndrRand%2 == 0 {
		changeCNDR = &tr
	} else {
		changeCNDR = &fal
	}

	return &ConfigurationChange{
		ID:                    strconv.Itoa(rand.Int()),
		AgentVersion:          changeVersion,
		EnableClusterScanning: changeClusterScanning,
		EnableRuntime:         changeRuntime,
		EnableCNDR:            changeCNDR,
		Status:                string(statusPending),
	}
}

type ConfigurationChange struct {
	ID                    string  `json:"id"`
	Status                string  `json:"status"`
	AgentVersion          *string `json:"agent_version"`
	EnableClusterScanning *bool   `json:"enable_cluster_scanning"`
	EnableRuntime         *bool   `json:"enable_runtime"`
	EnableCNDR            *bool   `json:"enable_cndr"`
	Timestamp             string  `json:"timestamp"`
}

type ConfigurationChangeStatusUpdate struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Reason string `json:"reason"`
	// AppliedGeneration tracks the generation of the Custom resource where the change was applied
	AppliedGeneration int64 `json:"applied_generation"`
	// AppliedTimestamp records when the change was applied in RFC3339 format
	AppliedTimestamp string `json:"applied_timestamp"`

	// TODO: CLuster and group. Cluster identifier?
}

type changeStatus string

var (
	statusPending      changeStatus = "PENDING"
	statusAcknowledged changeStatus = "ACKNOWLEDGED" // TODO: Acknowledged or applied?
	statusFailed       changeStatus = "FAILED"
)
