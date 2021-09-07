package models

import "fmt"

type AgentVersion string

const (
	AgentMinVersionNone   AgentVersion = "none"
	AgentMaxVersionLatest AgentVersion = "latest"
	AgentVersionUnknown   AgentVersion = ""
)

func (v AgentVersion) IsBiggerThan(version string) bool {
	if v == AgentMinVersionNone || v == AgentVersionUnknown {
		return false
	}
	return string(v) > version
}

func (v AgentVersion) IsLessThan(version string) bool {
	if v == AgentMaxVersionLatest || v == AgentVersionUnknown {
		return false
	}
	return string(v) < version
}

// CompatibilityMatrixResponse is the response returned by the GET Compatibility Matrix API.
type CompatibilityMatrixResponse struct {
	Operators CompatibilityMatrix `json:"operators"`
}

type CompatibilityMatrix map[string]*CompatibilityMatrixEntry

// CompatibilityMatrixEntry shows the min and max supported agent versions.
type CompatibilityMatrixEntry struct {
	Min AgentVersion `json:"min_agent"`
	Max AgentVersion `json:"max_agent"`
}

// CheckCompatibility uses the given entry to check whether the given agentVersion are compatible.
//
// It returns an error if the versions are not compatible, nil if they are.
// If the min and max agent versions for this operator version are unknown it will
// skip the check and return true.
func (entry CompatibilityMatrixEntry) CheckCompatibility(agentVersion string) error {
	if entry.Max.IsLessThan(agentVersion) {
		return fmt.Errorf("agent version too high, upgrade the operator to use that agent version")
	}
	if entry.Min.IsBiggerThan(agentVersion) {
		return fmt.Errorf("agent version too low, downgrade the operator to use that agent version")
	}

	// if we are here it means the operator and the agent version are compatibile
	return nil
}
