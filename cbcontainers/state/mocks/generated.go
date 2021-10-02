package mocks

//go:generate mockgen -destination agent_component_applier.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/state AgentComponentApplier
//go:generate mockgen -destination mock_secret_values_creator.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/state/components TlsSecretsValuesCreator
