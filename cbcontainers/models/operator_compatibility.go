package models

import "fmt"

// OperatorCompatibility shows the min and max supported agent versions.
type OperatorCompatibility struct {
	MinAgent AgentVersion `json:"min_agent"`
	MaxAgent AgentVersion `json:"max_agent"`
}

// CheckCompatibility uses the given entry to check whether the given agentVersion are compatible.
//
// It returns an error if the versions are not compatible, nil if they are.
// If the min and max agent versions for this operator version are unknown it will
// skip the check and return true.
func (c OperatorCompatibility) CheckCompatibility(agentVersion string) error {
	if c.MaxAgent.IsLessThan(agentVersion) {
		return fmt.Errorf("agent version too high, upgrade the operator to use that agent version: max is [%s], desired is [%s]", c.MaxAgent, agentVersion)
	}
	if c.MinAgent.IsLargerThan(agentVersion) {
		return fmt.Errorf("agent version too low, downgrade the operator to use that agent version: min is [%s], desired is [%s]", c.MinAgent, agentVersion)
	}

	// if we are here it means the operator and the agent version are compatibile
	return nil
}