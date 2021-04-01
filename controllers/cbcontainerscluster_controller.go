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
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
)

type clusterStateApplier interface {
	ApplyDesiredState(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersCluster, secret *models.RegistrySecretValues, client client.Client, setOwner applyment.OwnerSetter) (bool, error)
}

type clusterProcessor interface {
	Process(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) (*models.RegistrySecretValues, error)
}

type CBContainersClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	ClusterProcessor    clusterProcessor
	ClusterStateApplier clusterStateApplier
}

// +kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainersclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainersclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainersclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

func (r *CBContainersClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("cbcontainerscluster", req.NamespacedName)
	cbContainersCluster := &cbcontainersv1.CBContainersCluster{}
	if err := r.Get(ctx, req.NamespacedName, cbContainersCluster); err != nil {
		return ctrl.Result{}, fmt.Errorf("couldn't find CBContainersCluster k8s object: %v", err)
	}

	setOwner := func(controlledResource metav1.Object) error {
		return ctrl.SetControllerReference(cbContainersCluster, controlledResource, r.Scheme)
	}

	registrySecret, err := r.runProcessor(ctx, cbContainersCluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	stateWasChanged, err := r.ClusterStateApplier.ApplyDesiredState(ctx, cbContainersCluster, registrySecret, r.Client, setOwner)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: stateWasChanged}, nil
}

func (r *CBContainersClusterReconciler) runProcessor(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersCluster) (*models.RegistrySecretValues, error) {
	accessToken, err := r.getAccessToken(ctx, cbContainersCluster)
	if err != nil {
		return nil, err
	}

	return r.ClusterProcessor.Process(cbContainersCluster, accessToken)
}

func (r *CBContainersClusterReconciler) getAccessToken(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersCluster) (string, error) {
	accessTokenSecretNamespacedName := types.NamespacedName{Name: cbContainersCluster.Spec.ApiGatewaySpec.AccessTokenSecretName, Namespace: cbContainersCluster.Namespace}
	accessTokenSecret := &corev1.Secret{}
	if err := r.Get(ctx, accessTokenSecretNamespacedName, accessTokenSecret); err != nil {
		return "", fmt.Errorf("couldn't find access token secret k8s object: %v", err)
	}

	accessToken := string(accessTokenSecret.Data[commonState.AccessTokenSecretKeyName])
	if accessToken == "" {
		return "", fmt.Errorf("the k8s secret %v is missing the key %v", accessTokenSecretNamespacedName, commonState.AccessTokenSecretKeyName)
	}

	return accessToken, nil
}

func (r *CBContainersClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cbcontainersv1.CBContainersCluster{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
