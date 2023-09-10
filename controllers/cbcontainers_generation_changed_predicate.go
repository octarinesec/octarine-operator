package controllers

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type StateGenerationChangedPredicate interface {
	ShouldProcessEvent(client.Object) bool
}

type CBContainersGenerationChangedPredicate struct {
	predicate.GenerationChangedPredicate
	statePredicate StateGenerationChangedPredicate
}

func NewCBContainersGenerationChangedPredicate(statePredicate StateGenerationChangedPredicate) CBContainersGenerationChangedPredicate {
	return CBContainersGenerationChangedPredicate{
		statePredicate: statePredicate,
	}
}

func (p CBContainersGenerationChangedPredicate) Create(e event.CreateEvent) bool {
	return p.statePredicate.ShouldProcessEvent(e.Object) || p.GenerationChangedPredicate.Create(e)
}

func (p CBContainersGenerationChangedPredicate) Update(e event.UpdateEvent) bool {
	return p.statePredicate.ShouldProcessEvent(e.ObjectNew) || p.GenerationChangedPredicate.Update(e)
}

func (p CBContainersGenerationChangedPredicate) Delete(e event.DeleteEvent) bool {
	return p.statePredicate.ShouldProcessEvent(e.Object) || p.GenerationChangedPredicate.Delete(e)
}
