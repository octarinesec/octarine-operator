package mocks

//go:generate mockgen -destination mock_gateway.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/processors/cluster Gateway
//go:generate mockgen -destination mock_gateway_creator.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/processors/cluster GatewayCreator
//go:generate mockgen -destination mock_monitor.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/processors/cluster Monitor
//go:generate mockgen -destination mock_monitor_creator.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/processors/cluster MonitorCreator
