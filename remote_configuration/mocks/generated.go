package mocks

//go:generate mockgen -destination mock_configuration_api.go -package mocks github.com/vmware/cbcontainers-operator/remote_configuration ConfigurationChangesAPI
//go:generate mockgen -destination mock_change_validator.go -package mocks github.com/vmware/cbcontainers-operator/remote_configuration ChangeValidator
