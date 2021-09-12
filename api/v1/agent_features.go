package v1

type AgentFeature string
type AgentFeaturesList []AgentFeature

var (
	AgentFeatureHardeningBasic    = AgentFeature("HARDENING_BASIC")
	AgentFeatureHardeningAdvanced = AgentFeature("HARDENING_ADVANCED")
	AgentFeatureRuntime           = AgentFeature("RUNTIME")
)
