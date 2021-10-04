package mocks

//go:generate mockgen -destination mock_state_applier.go -package mocks github.com/vmware/cbcontainers-operator/controllers StateApplier
//go:generate mockgen -destination mock_cluster_processor.go -package mocks github.com/vmware/cbcontainers-operator/controllers ClusterProcessor
//go:generate mockgen -destination mock_gateway_creator.go -package mocks github.com/vmware/cbcontainers-operator/controllers GatewayCreator
//go:generate mockgen -destination mock_gateway.go -package mocks github.com/vmware/cbcontainers-operator/controllers Gateway
//go:generate mockgen -destination mock_operator_version_provider.go -package mocks github.com/vmware/cbcontainers-operator/controllers OperatorVersionProvider