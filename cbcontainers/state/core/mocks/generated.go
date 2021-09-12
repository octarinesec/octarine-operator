package mocks

//go:generate mockgen -destination mock_cluster_child_applier.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/state/cluster ClusterChildK8sObjectApplier
