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
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorcontainerscarbonblackiov1 "github.com/vmware/cbcontainers-operator/api/v1"
	clusterState "github.com/vmware/cbcontainers-operator/state/cluster"
)

// CBContainersClusterReconciler reconciles a CBContainersCluster object
type CBContainersClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	stateController *clusterState.CBContainersClusterStateController
}

// +kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainersclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainersclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainersclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

func (r *CBContainersClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("cbcontainerscluster", req.NamespacedName)
	result, err := Reconcile(ctx, req, r.Client, r.getStateController())
	if err != nil {
		r.Log.Error(err, err.Error())
	}

	return result, err
}

func (r *CBContainersClusterReconciler) getStateController() *clusterState.CBContainersClusterStateController {
	if r.stateController == nil {
		r.stateController = clusterState.NewStateController()
	}

	return r.stateController
}

func (r *CBContainersClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorcontainerscarbonblackiov1.CBContainersCluster{}).
		Owns(&v1.ConfigMap{}).
		Complete(r)
}
