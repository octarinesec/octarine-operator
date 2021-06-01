/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorcontainerscarbonblackiov1 "github.com/vmware/cbcontainers-operator/api/v1"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
)

type runtimeStateApplier interface {
	ApplyDesiredState(ctx context.Context, cbContainersRuntime *operatorcontainerscarbonblackiov1.CBContainersRuntime, client client.Client, setOwner applymentOptions.OwnerSetter) (bool, error)
}

// CBContainersRuntimeReconciler reconciles a CBContainersRuntime object
type CBContainersRuntimeReconciler struct {
	client.Client
	Log                 logr.Logger
	Scheme              *runtime.Scheme
	RuntimeStateApplier runtimeStateApplier
}

func (r *CBContainersRuntimeReconciler) getContainersRuntimeObject(ctx context.Context) (*operatorcontainerscarbonblackiov1.CBContainersRuntime, error) {
	cbContainersRuntimeList := &operatorcontainerscarbonblackiov1.CBContainersRuntimeList{}
	if err := r.List(ctx, cbContainersRuntimeList); err != nil {
		return nil, fmt.Errorf("couldn't list CBContainersRuntime k8s objects: %v", err)
	}

	if cbContainersRuntimeList.Items == nil || len(cbContainersRuntimeList.Items) == 0 {
		return nil, nil
	}

	if len(cbContainersRuntimeList.Items) > 1 {
		return nil, fmt.Errorf("there is more than 1 CBContainersRuntime k8s object, please delete unwanted resources")
	}

	return &cbContainersRuntimeList.Items[0], nil
}

//+kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainersruntimes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainersruntimes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainersruntimes/finalizers,verbs=update

func (r *CBContainersRuntimeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("\n\n")
	r.Log.Info("Got reconcile request", "namespaced name", req.NamespacedName)
	r.Log.Info("Starting reconciling")

	r.Log.Info("Getting CBContainersRuntime object")
	cbContainersRuntime, err := r.getContainersRuntimeObject(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	if cbContainersRuntime == nil {
		return ctrl.Result{}, nil
	}

	r.setDefaults(cbContainersRuntime)

	setOwner := func(controlledResource metav1.Object) error {
		return ctrl.SetControllerReference(cbContainersRuntime, controlledResource, r.Scheme)
	}

	r.Log.Info("Applying desired state")
	stateWasChanged, err := r.RuntimeStateApplier.ApplyDesiredState(ctx, cbContainersRuntime, r.Client, setOwner)
	if err != nil {
		return ctrl.Result{}, err
	}

	r.Log.Info("Finished reconciling", "Requiring", stateWasChanged)
	r.Log.Info("\n\n")
	return ctrl.Result{Requeue: stateWasChanged}, nil
}

func (r *CBContainersRuntimeReconciler) setDefaults(cbContainersRuntime *operatorcontainerscarbonblackiov1.CBContainersRuntime) {
	// TODO
}

// SetupWithManager sets up the controller with the Manager.
func (r *CBContainersRuntimeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorcontainerscarbonblackiov1.CBContainersRuntime{}).
		Owns(&appsV1.Deployment{}).
		Owns(&coreV1.Service{}).
		Owns(&appsV1.DaemonSet{}).
		Complete(r)
}
