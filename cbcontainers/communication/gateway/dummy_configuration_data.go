package gateway

import (
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"math/rand"
	"strconv"
)

// TODO: This will be removed once real APIs are implemented for this but it helps try the feature while in development
// API task - CNS-2790

var (
	dummyAgentVersions = []string{"2.12.1", "2.10.0", "2.12.0", "2.11.0", "3.0.0"}
)

func randomRemoteConfigChange() *models.ConfigurationChange {
	versionRand, nilRand := rand.Intn(len(dummyAgentVersions)), rand.Int()

	if nilRand%5 == 1 {
		return nil
	}

	changeVersion := &dummyAgentVersions[versionRand]

	return &models.ConfigurationChange{
		ID:           strconv.Itoa(rand.Int()),
		AgentVersion: changeVersion,
		Status:       models.ChangeStatusPending,
	}
}
