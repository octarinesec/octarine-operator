package mocks

//go:generate mockgen -destination mock_runtime_child_applier.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/state/runtime RuntimeChildK8sObjectApplier
