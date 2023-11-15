package models_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
)

type testCase struct {
	min   models.Version
	max   models.Version
	agent models.Version
}

func TestCheckCompatibilityCompatible(t *testing.T) {
	testCases := []testCase{
		{min: models.Version("2.7"), max: models.Version("2.8"), agent: "2.7"},
		{min: models.Version("2.7"), max: models.Version("2.8"), agent: "2.8"},
		{min: models.Version("2.7"), max: models.Version("2.8"), agent: "2.7.1"},
		{min: models.Version("2.7"), max: models.Version("2.8"), agent: "2.7.5"},
		{min: models.Version("2.7"), max: models.Version("2.8.1"), agent: "2.8.1"},
		{min: models.Version("2.7"), max: models.Version("2.8.1"), agent: "2.7.1-alpha"},
		{min: models.Version("2.7"), max: models.Version("2.8.1"), agent: "2.7.1-beta"},
		{min: models.MinVersionNone, max: models.Version("2.8.1"), agent: "0.0.1"},
		{min: models.Version("2.7"), max: models.MaxVersionLatest, agent: "3.0.0"},
		{min: models.Version("2.7"), max: models.MaxVersionLatest, agent: "50.0.0"},
		{min: models.Version("1.0"), max: models.Version("2.8"), agent: "1.5"},
		{min: models.Version("1.0"), max: models.Version("2.8"), agent: "2.0"},
		{min: models.Version("0.0"), max: models.Version("2.8"), agent: "0.0"},
		{min: models.Version("0.0"), max: models.Version("2.8"), agent: "1.0"},
		{min: models.Version("0.0"), max: models.Version("2.8"), agent: "2.0"},
		{min: models.Version("1.2"), max: models.Version("1.30"), agent: "1.10"},
		{min: models.Version("1.1.2"), max: models.Version("1.1.30"), agent: "1.1.10"},
		{min: models.Version("1.2.105"), max: models.Version("1.30.0"), agent: "1.10.40"},
	}

	testCheckCompatibility(t, testCases, true)
}

func TestCheckCompatibilityIncompatible(t *testing.T) {
	testCases := []testCase{
		{min: models.Version("2.7"), max: models.Version("2.8"), agent: "2.8.1"},
		{min: models.Version("2.7"), max: models.Version("2.8"), agent: "3"},
		{min: models.Version("2.7"), max: models.Version("2.8"), agent: "2.6"},
		{min: models.Version("2.7"), max: models.Version("2.8"), agent: "2.6.9"},
		{min: models.MinVersionNone, max: models.Version("2.8"), agent: "2.9"},
		{min: models.Version("2.7"), max: models.MaxVersionLatest, agent: "2.6"},
		{min: models.Version("1.10"), max: models.Version("1.30"), agent: "1.2"},
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
