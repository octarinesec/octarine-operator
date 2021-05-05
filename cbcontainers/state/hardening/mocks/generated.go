package mocks

//go:generate mockgen -destination mock_hardening_child_applier.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/state/hardening HardeningChildK8sObjectApplier
//go:generate mockgen -destination mock_secret_values_creator.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/state/hardening/objects TlsSecretsValuesCreator
