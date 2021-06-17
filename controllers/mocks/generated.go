package mocks

//go:generate mockgen -destination mock_cluster_state_applier.go -package mocks github.com/vmware/cbcontainers-operator/controllers ClusterStateApplier
//go:generate mockgen -destination mock_cluster_processor.go -package mocks github.com/vmware/cbcontainers-operator/controllers ClusterProcessor
//go:generate mockgen -destination mock_hardening_state_applier.go -package mocks github.com/vmware/cbcontainers-operator/controllers HardeningStateApplier
//go:generate mockgen -destination mock_runtime_state_applier.go -package mocks github.com/vmware/cbcontainers-operator/controllers RuntimeStateApplier
