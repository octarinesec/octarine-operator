package remote_configuration

import (
	"context"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
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

func (d DummyAPI) GetConfigurationChanges(ctx context.Context) ([]models.ConfigurationChange, error) {
	c := RandomChange()
	if c != nil {
		return []models.ConfigurationChange{*c}, nil

	}
	return nil, nil
}

func (d DummyAPI) UpdateConfigurationChangeStatus(ctx context.Context, update models.ConfigurationChangeStatusUpdate) error {
	return nil
}

func RandomNonNilChange() models.ConfigurationChange {
	for {
		c := RandomChange()
		if c != nil {
			return *c
		}
	}
}

func RandomChange() *models.ConfigurationChange {
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

	return &models.ConfigurationChange{
		ID:                    strconv.Itoa(rand.Int()),
		AgentVersion:          changeVersion,
		EnableClusterScanning: changeClusterScanning,
		EnableRuntime:         changeRuntime,
		EnableCNDR:            changeCNDR,
		Status:                string(statusPending),
	}
}

type changeStatus string

var (
	statusPending      changeStatus = "PENDING"
	statusAcknowledged changeStatus = "ACKNOWLEDGED" // TODO: Acknowledged or applied?
	statusFailed       changeStatus = "FAILED"
)
