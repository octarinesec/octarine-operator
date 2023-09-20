package mocks

//go:generate mockgen -destination mock_state_applier.go -package mocks github.com/vmware/cbcontainers-operator/controllers StateApplier
//go:generate mockgen -destination mock_agent_processor.go -package mocks github.com/vmware/cbcontainers-operator/controllers AgentProcessor
//go:generate mockgen -destination mock_access_token_provider.go -package mocks github.com/vmware/cbcontainers-operator/controllers AccessTokenProvider
