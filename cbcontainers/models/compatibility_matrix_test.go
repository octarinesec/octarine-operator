package models_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
)

func TestCheckCompatibility(t *testing.T) {
	testCases := []struct {
		min           models.AgentVersion
		max           models.AgentVersion
		agent         string
		areCompatible bool
	}{
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8"), agent: "2.7", areCompatible: true},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8"), agent: "2.8", areCompatible: true},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8"), agent: "2.7.1", areCompatible: true},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8"), agent: "2.7.5", areCompatible: true},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8.1"), agent: "2.8.1", areCompatible: true},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8.1"), agent: "2.7.1-alpha", areCompatible: true},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8.1"), agent: "2.7.1-beta", areCompatible: true},
		{min: models.AgentMinVersionNone, max: models.AgentVersion("2.8.1"), agent: "0.0.1", areCompatible: true},
		{min: models.AgentVersion("2.7"), max: models.AgentMaxVersionLatest, agent: "3.0.0", areCompatible: true},
		{min: models.AgentVersion("2.7"), max: models.AgentMaxVersionLatest, agent: "50.0.0", areCompatible: true},
		{min: models.AgentVersion("1.0"), max: models.AgentVersion("2.8"), agent: "1.5", areCompatible: true},
		{min: models.AgentVersion("1.0"), max: models.AgentVersion("2.8"), agent: "2.0", areCompatible: true},
		{min: models.AgentVersion("0.0"), max: models.AgentVersion("2.8"), agent: "0.0", areCompatible: true},
		{min: models.AgentVersion("0.0"), max: models.AgentVersion("2.8"), agent: "1.0", areCompatible: true},
		{min: models.AgentVersion("0.0"), max: models.AgentVersion("2.8"), agent: "2.0", areCompatible: true},

		{min: models.AgentMinVersionNone, max: models.AgentMaxVersionLatest, agent: "0.0", areCompatible: true},
		{min: models.AgentMinVersionNone, max: models.AgentMaxVersionLatest, agent: "0.0.1", areCompatible: true},
		{min: models.AgentMinVersionNone, max: models.AgentMaxVersionLatest, agent: "0.0.0-beta", areCompatible: true},
		{min: models.AgentMinVersionNone, max: models.AgentMaxVersionLatest, agent: "30.0.0-beta", areCompatible: true},

		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8"), agent: "2.8.1", areCompatible: false},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8"), agent: "3", areCompatible: false},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8"), agent: "2.6", areCompatible: false},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8"), agent: "2.6.9", areCompatible: false},
		{min: models.AgentMinVersionNone, max: models.AgentVersion("2.8"), agent: "2.9", areCompatible: false},
		{min: models.AgentVersion("2.7"), max: models.AgentMaxVersionLatest, agent: "2.6", areCompatible: false},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("TestCheckCompatibility %d", i), func(t *testing.T) {
			entry := models.CompatibilityMatrixEntry{Min: testCase.min, Max: testCase.max}

			err := entry.CheckCompatibility(testCase.agent)
			if testCase.areCompatible {
				require.NoError(t, err, "CheckCompatibility should not return error when versions are compatible - min (%s), max (%s), agent (%s)", testCase.min, testCase.max, testCase.agent)
			} else {
				require.Error(t, err, "CheckCompatibility should return error when versions are not compatible - min (%s), max (%s), agent (%s)", testCase.min, testCase.max, testCase.agent)
			}
		})
	}
}
