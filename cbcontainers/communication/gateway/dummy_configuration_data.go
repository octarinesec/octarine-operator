package gateway

import (
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"math/rand"
	"strconv"
)

// TODO: This will be removed once real APIs are implemented for this but it helps try the feature while in development
// API task - CNS-2790

var (
	tr                 = true
	fal                = false
	dummyAgentVersions = []string{"2.12.1", "2.10.0", "2.12.0", "2.11.0", "3.0.0"}
)

func randomRemoteConfigChange() *models.ConfigurationChange {
	csRand, runtimeRand, cndrRand, versionRand, nilRand := rand.Int(), rand.Int(), rand.Int(), rand.Intn(len(dummyAgentVersions)), rand.Int()

	if nilRand%5 == 1 {
		return nil
	}

	changeVersion := &dummyAgentVersions[versionRand]

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
		Status:                models.ChangeStatusPending,
	}
}
