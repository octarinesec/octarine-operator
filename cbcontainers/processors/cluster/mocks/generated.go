package mocks

//go:generate mockgen -destination mock_gateway.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/processors/cluster Gateway
//go:generate mockgen -destination mock_gateway_creator.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/processors/cluster GatewayCreator
