package mocks

//go:generate mockgen -destination mock_features_status_provider_generated.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/monitor FeaturesStatusProvider
//go:generate mockgen -destination mock_health_checker_generated.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/monitor HealthChecker
//go:generate mockgen -destination mock_message_reporter_generated.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/monitor MessageReporter
