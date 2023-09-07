package mocks

//go:generate mockgen -destination mock_api_gateway.go -package mocks github.com/vmware/cbcontainers-operator/remote_configuration ApiGateway
//go:generate mockgen -destination mock_access_token_provider.go -package mocks github.com/vmware/cbcontainers-operator/remote_configuration AccessTokenProvider
//go:generate mockgen -destination mock_resource_syncer.go -package mocks github.com/vmware/cbcontainers-operator/remote_configuration CustomResourceSyncer
