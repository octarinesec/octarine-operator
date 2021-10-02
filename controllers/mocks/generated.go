package mocks

//go:generate mockgen -destination mock_state_applier.go -package mocks github.com/vmware/cbcontainers-operator/controllers StateApplier
//go:generate mockgen -destination mock_cluster_processor.go -package mocks github.com/vmware/cbcontainers-operator/controllers ClusterProcessor
