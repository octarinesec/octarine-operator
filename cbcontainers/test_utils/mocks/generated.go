package mocks

//go:generate mockgen -destination mock_k8s_client.go -package mocks sigs.k8s.io/controller-runtime/pkg/client Client
