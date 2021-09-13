package agent_applyment

import (
	"context"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type componentApplier interface {
	Apply(ctx context.Context, desiredK8sObject applyment.DesiredK8sObject, applyOptionsList ...*applymentOptions.ApplyOptions) (bool, client.Object, error)
	Delete(ctx context.Context, desiredK8sObject applyment.DesiredK8sObject) (bool, error)
}

type AgentComponentApplier struct {
	applier componentApplier
}

func NewAgentComponent(componentApplier componentApplier) *AgentComponentApplier {
	return &AgentComponentApplier{
		applier: componentApplier,
	}
}

func (agentComponentApplier *AgentComponentApplier) Apply(ctx context.Context, builder AgentComponentBuilder, agentSpec *cbcontainersv1.CBContainersAgentSpec, applyOptionsList ...*applymentOptions.ApplyOptions) (bool, client.Object, error) {
	wrapper := NewDesiredAgentComponentWrapper(builder, agentSpec)
	return agentComponentApplier.applier.Apply(ctx, wrapper, applyOptionsList...)
}

func (agentComponentApplier *AgentComponentApplier) Delete(ctx context.Context, builder AgentComponentBuilder, agentSpec *cbcontainersv1.CBContainersAgentSpec) (bool, error) {
	wrapper := NewDesiredAgentComponentWrapper(builder, agentSpec)
	return agentComponentApplier.applier.Delete(ctx, wrapper)
}
