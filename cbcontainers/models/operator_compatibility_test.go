package models_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
)

type testCase struct {
	min   models.AgentVersion
	max   models.AgentVersion
	agent string
}

func TestCheckCompatibilityCompatible(t *testing.T) {
	testCases := []testCase{
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8"), agent: "2.7"},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8"), agent: "2.8"},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8"), agent: "2.7.1"},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8"), agent: "2.7.5"},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8.1"), agent: "2.8.1"},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8.1"), agent: "2.7.1-alpha"},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8.1"), agent: "2.7.1-beta"},
		{min: models.AgentMinVersionNone, max: models.AgentVersion("2.8.1"), agent: "0.0.1"},
		{min: models.AgentVersion("2.7"), max: models.AgentMaxVersionLatest, agent: "3.0.0"},
		{min: models.AgentVersion("2.7"), max: models.AgentMaxVersionLatest, agent: "50.0.0"},
		{min: models.AgentVersion("1.0"), max: models.AgentVersion("2.8"), agent: "1.5"},
		{min: models.AgentVersion("1.0"), max: models.AgentVersion("2.8"), agent: "2.0"},
		{min: models.AgentVersion("0.0"), max: models.AgentVersion("2.8"), agent: "0.0"},
		{min: models.AgentVersion("0.0"), max: models.AgentVersion("2.8"), agent: "1.0"},
		{min: models.AgentVersion("0.0"), max: models.AgentVersion("2.8"), agent: "2.0"},
	}

	testCheckCompatibility(t, testCases, true)
}

func TestCheckCompatibilityIncompatible(t *testing.T) {
	testCases := []testCase{
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8"), agent: "2.8.1"},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8"), agent: "3"},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8"), agent: "2.6"},
		{min: models.AgentVersion("2.7"), max: models.AgentVersion("2.8"), agent: "2.6.9"},
		{min: models.AgentMinVersionNone, max: models.AgentVersion("2.8"), agent: "2.9"},
		{min: models.AgentVersion("2.7"), max: models.AgentMaxVersionLatest, agent: "2.6"},
	}

	testCheckCompatibility(t, testCases, false)
}

func testCheckCompatibility(t *testing.T, testCases []testCase, areCompatible bool) {
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("testCheckCompatibility %d", i), func(t *testing.T) {
			entry := models.OperatorCompatibility{MinAgent: testCase.min, MaxAgent: testCase.max}

			err := entry.CheckCompatibility(testCase.agent)
			if areCompatible {
				require.NoError(t, err, "CheckCompatibility should not return error when versions are compatible - min (%s), max (%s), agent (%s)", testCase.min, testCase.max, testCase.agent)
			} else {
				require.Error(t, err, "CheckCompatibility should return error when versions are not compatible - min (%s), max (%s), agent (%s)", testCase.min, testCase.max, testCase.agent)
			}
		})
	}
}
