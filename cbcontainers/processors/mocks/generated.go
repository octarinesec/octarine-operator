package mocks

//go:generate mockgen -destination mock_api_gateway.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/processors APIGateway
//go:generate mockgen -destination mock_api_gateway_creator.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/processors APIGatewayCreator
