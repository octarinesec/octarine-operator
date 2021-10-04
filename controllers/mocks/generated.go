package mocks

//go:generate mockgen -destination mock_gateway.go -package mocks github.com/vmware/cbcontainers-operator/controllers Gateway
//go:generate mockgen -destination mock_operator_version_provider.go -package mocks github.com/vmware/cbcontainers-operator/controllers OperatorVersionProvider
//go:generate mockgen -destination mock_gateway_creator.go -package mocks github.com/vmware/cbcontainers-operator/controllers GatewayCreator
//go:generate mockgen -destination mock_api_gateway.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/processors APIGateway
//go:generate mockgen -destination mock_operator_version_provider.go -package mocks github.com/vmware/cbcontainers-operator/controllers OperatorVersionProvider
