package agent_applyment

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AgentComponentBuilder interface {
	applyment.DesiredK8sObjectInitializer
	MutateK8sObject(object client.Object, agentSpec *cbcontainersv1.CBContainersAgentSpec) error
}

type DesiredAgentComponentWrapper struct {
	AgentComponentBuilder
	agentSpec *cbcontainersv1.CBContainersAgentSpec
}

func NewDesiredAgentComponentWrapper(desiredAgentComponent AgentComponentBuilder, agentSpec *cbcontainersv1.CBContainersAgentSpec) *DesiredAgentComponentWrapper {
	return &DesiredAgentComponentWrapper{
		AgentComponentBuilder: desiredAgentComponent,
		agentSpec:             agentSpec,
	}
}

func (d *DesiredAgentComponentWrapper) MutateK8sObject(object client.Object) error {
	return d.AgentComponentBuilder.MutateK8sObject(object, d.agentSpec)
}
