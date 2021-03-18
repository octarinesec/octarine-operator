package guardrails

import (
	operatorcontainerscarbonblackiov1 "github.com/vmware/cbcontainers-operator/api/v1"
	stateTypes "github.com/vmware/cbcontainers-operator/state/types"
)

func Build(guardrails *operatorcontainerscarbonblackiov1.CBContainersGuardrails) []stateTypes.DesiredK8sObject {
	return []stateTypes.DesiredK8sObject{}
}
