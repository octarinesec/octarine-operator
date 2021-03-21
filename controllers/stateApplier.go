package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StateApplier interface {
	ApplyState(ctx context.Context, namespacedName types.NamespacedName, client client.Client) (bool, error)
}

func Reconcile(ctx context.Context, req ctrl.Request, client client.Client, stateApplier StateApplier) (ctrl.Result, error) {
	ok, err := stateApplier.ApplyState(ctx, req.NamespacedName, client)
	if err != nil {
		return ctrl.Result{}, err
	}

	if ok {
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}
