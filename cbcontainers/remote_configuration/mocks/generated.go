package mocks

//go:generate mockgen -destination mock_api_gateway.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/remote_configuration ApiGateway
//go:generate mockgen -destination mock_access_token_provider.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/remote_configuration AccessTokenProvider
//go:generate mockgen -destination mock_validator.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/remote_configuration ChangeValidator
