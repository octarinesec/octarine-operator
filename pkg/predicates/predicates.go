package predicates

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	crtpredicate "sigs.k8s.io/controller-runtime/pkg/predicate"
)

var log = logf.Log.WithName("predicates")

// DependentPredicateFuncs returns functions defined for filtering events.
// Copied from operator-sdk (internal/predicates/predicates.go) as we can't import internal package.
func DependentPredicateFuncs() crtpredicate.Funcs {

	dependentPredicate := crtpredicate.Funcs{
		// We don't need to reconcile dependent resource creation events
		// because dependent resources are only ever created during
		// reconciliation. Another reconcile would be redundant.
		CreateFunc: func(e event.CreateEvent) bool {
			log.V(2).Info("Skipping reconciliation for dependent resource creation",
				"name", e.Meta.GetName(), "namespace", e.Meta.GetNamespace(), "apiVersion",
				e.Object.GetObjectKind().GroupVersionKind().GroupVersion(), "kind",
				e.Object.GetObjectKind().GroupVersionKind().Kind)
			return false
		},

		// Reconcile when a dependent resource is deleted so that it can be
		// recreated.
		DeleteFunc: func(e event.DeleteEvent) bool {
			log.V(2).Info("Reconciling due to dependent resource deletion",
				"name", e.Meta.GetName(), "namespace", e.Meta.GetNamespace(), "apiVersion",
				e.Object.GetObjectKind().GroupVersionKind().GroupVersion(), "kind",
				e.Object.GetObjectKind().GroupVersionKind().Kind)
			return true
		},

		// Don't reconcile when a generic event is received for a dependent
		GenericFunc: func(e event.GenericEvent) bool {
			log.V(2).Info("Skipping reconcile due to generic event",
				"name", e.Meta.GetName(), "namespace", e.Meta.GetNamespace(), "apiVersion",
				e.Object.GetObjectKind().GroupVersionKind().GroupVersion(), "kind",
				e.Object.GetObjectKind().GroupVersionKind().Kind)
			return false
		},

		// Reconcile when a dependent resource is updated, so that it can
		// be patched back to the resource managed by the CR, if
		// necessary.
		UpdateFunc: func(e event.UpdateEvent) bool {
			log.V(2).Info("Reconciling due to dependent resource update",
				"name", e.MetaNew.GetName(), "namespace", e.MetaNew.GetNamespace(), "apiVersion",
				e.ObjectNew.GetObjectKind().GroupVersionKind().GroupVersion(), "kind",
				e.ObjectNew.GetObjectKind().GroupVersionKind().Kind)
			return true
		},
	}

	return dependentPredicate
}
