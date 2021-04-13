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
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	admissionsV1 "k8s.io/api/admissionregistration/v1beta1"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
)

type hardeningStateApplier interface {
	ApplyDesiredState(ctx context.Context, cbContainersHardening *cbcontainersv1.CBContainersHardening, client client.Client, setOwner applymentOptions.OwnerSetter) (bool, error)
}

type CBContainersHardeningReconciler struct {
	client.Client
	Log                   logr.Logger
	Scheme                *runtime.Scheme
	HardeningStateApplier hardeningStateApplier
}

func (r *CBContainersHardeningReconciler) getContainersHardeningObject(ctx context.Context) (*cbcontainersv1.CBContainersHardening, error) {
	cbContainersHardeningList := &cbcontainersv1.CBContainersHardeningList{}
	if err := r.List(ctx, cbContainersHardeningList); err != nil {
		return nil, fmt.Errorf("couldn't list CBContainersCluster k8s objects: %v", err)
	}

	if cbContainersHardeningList.Items == nil || len(cbContainersHardeningList.Items) == 0 {
		return nil, nil
	}

	if len(cbContainersHardeningList.Items) > 1 {
		return nil, fmt.Errorf("there is more than 1 CBContainersCluster k8s object, please delete unwanted resources")
	}

	return &cbContainersHardeningList.Items[0], nil
}

// +kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainershardenings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainershardenings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainershardenings/finalizers,verbs=update
// +kubebuilder:rbac:groups={apps,core},resources={deployments,services},verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=*
// +kubebuilder:rbac:groups={rbac.authorization.k8s.io,networking.k8s.io,apiextensions.k8s.io,extensions,rbac,batch,apps,core},resources={namespaces,clusterrolebindings,services,networkpolicies,ingresses,rolebindings,cronjobs,jobs,replicationcontrollers,statefulsets,daemonsets,deployments,replicasets,pods,nodes,customresourcedefinitions},verbs=get;list;watch

func (r *CBContainersHardeningReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("\n\nGot reconcile request", "namespaced name", req.NamespacedName)
	r.Log.Info("Starting reconciling")

	r.Log.Info("Getting CBContainersHardening object")
	cbContainersHardening, err := r.getContainersHardeningObject(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	if cbContainersHardening == nil {
		return ctrl.Result{}, nil
	}

	setOwner := func(controlledResource metav1.Object) error {
		return ctrl.SetControllerReference(cbContainersHardening, controlledResource, r.Scheme)
	}

	r.Log.Info("Applying desired state")
	stateWasChanged, err := r.HardeningStateApplier.ApplyDesiredState(ctx, cbContainersHardening, r.Client, setOwner)
	if err != nil {
		return ctrl.Result{}, err
	}

	r.Log.Info("Finished reconciling\n\n", "Requiring", stateWasChanged)
	return ctrl.Result{Requeue: stateWasChanged}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CBContainersHardeningReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cbcontainersv1.CBContainersHardening{}).
		Owns(&appsV1.Deployment{}).
		Owns(&coreV1.Service{}).
		Owns(&admissionsV1.ValidatingWebhookConfiguration{}).
		Complete(r)
}
