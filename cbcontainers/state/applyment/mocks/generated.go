package mocks

//go:generate mockgen -destination mock_desired_k8s_object.go -package mocks github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment DesiredK8sObject
