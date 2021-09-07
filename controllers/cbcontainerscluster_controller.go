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
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
)

type ClusterStateApplier interface {
	GetPriorityClassEmptyK8sObject() client.Object
	ApplyDesiredState(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersCluster, secret *models.RegistrySecretValues, client client.Client, setOwner applymentOptions.OwnerSetter) (bool, error)
}

type ClusterProcessor interface {
	Process(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) (*models.RegistrySecretValues, error)
}

type Gateway interface {
	GetCompatibilityMatrixEntryFor(operatorVersion string) (*models.CompatibilityMatrixEntry, error)
}

type GatewayCreator interface {
	CreateGateway(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) Gateway
}

type OperatorVersionProvider interface {
	GetOperatorVersion() (string, error)
}

type CBContainersClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	// GatewayCreator will be used to create a gateway.ApiGateway
	GatewayCreator
	// _apiGateway is a singleton gateway.ApiGateway struct.
	// It is initialized the first time it is used and reused after that.
	// It should not be accessed directly, but through the apiGateway method.
	_apiGateway Gateway

	OperatorVersionProvider

	ClusterProcessor    ClusterProcessor
	ClusterStateApplier ClusterStateApplier
}

func (r *CBContainersClusterReconciler) getContainersClusterObject(ctx context.Context) (*cbcontainersv1.CBContainersCluster, error) {
	cbContainersClusterList := &cbcontainersv1.CBContainersClusterList{}
	if err := r.List(ctx, cbContainersClusterList); err != nil {
		return nil, fmt.Errorf("couldn't list CBContainersCluster k8s objects: %v", err)
	}

	if cbContainersClusterList.Items == nil || len(cbContainersClusterList.Items) == 0 {
		return nil, nil
	}

	if len(cbContainersClusterList.Items) > 1 {
		return nil, fmt.Errorf("there is more than 1 CBContainersCluster k8s object, please delete unwanted resources")
	}

	return &cbContainersClusterList.Items[0], nil
}

// +kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainersclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainersclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainersclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources={configmaps,secrets},verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=scheduling.k8s.io,resources=priorityclasses,verbs=*

func (r *CBContainersClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("\n\n")
	r.Log.Info("Got reconcile request", "namespaced name", req.NamespacedName)
	r.Log.Info("Starting reconciling")

	r.Log.Info("Getting CBContainersCluster object")
	cbContainersCluster, err := r.getContainersClusterObject(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	if cbContainersCluster == nil {
		return ctrl.Result{}, nil
	}

	if err := r.setDefaults(cbContainersCluster); err != nil {
		return ctrl.Result{}, fmt.Errorf("faild to set defaults to cluster CR: %v", err)
	}

	setOwner := func(controlledResource metav1.Object) error {
		return ctrl.SetControllerReference(cbContainersCluster, controlledResource, r.Scheme)
	}

	accessToken, err := r.getAccessToken(context.Background(), cbContainersCluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	operatorVersion, err := r.OperatorVersionProvider.GetOperatorVersion()
	if err != nil {
		return ctrl.Result{}, err
	}

	m, err := r.apiGateway(cbContainersCluster, accessToken).GetCompatibilityMatrixEntryFor(operatorVersion)
	if err == nil {
		// if there is no error check the compatibility
		// if there is an error skip the check

		err = m.CheckCompatibility(cbContainersCluster.Spec.Version)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		r.Log.Error(err, "error while getting compatibility matrix from backend")
	}

	r.Log.Info("Getting registry secret values")
	registrySecret, err := r.getRegistrySecretValues(ctx, cbContainersCluster, accessToken)
	if err != nil {
		return ctrl.Result{}, err
	}

	r.Log.Info("Applying desired state")
	stateWasChanged, err := r.ClusterStateApplier.ApplyDesiredState(ctx, cbContainersCluster, registrySecret, r.Client, setOwner)
	if err != nil {
		return ctrl.Result{}, err
	}

	r.Log.Info("Finished reconciling", "Requiring", stateWasChanged)
	r.Log.Info("\n\n")
	return ctrl.Result{Requeue: stateWasChanged}, nil
}

func (r *CBContainersClusterReconciler) apiGateway(cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) Gateway {
	if r._apiGateway == nil {
		r._apiGateway = r.GatewayCreator.CreateGateway(cbContainersCluster, accessToken)
	}
	return r._apiGateway
}

func (r *CBContainersClusterReconciler) setDefaults(cbContainersCluster *cbcontainersv1.CBContainersCluster) error {
	if cbContainersCluster.Spec.ApiGatewaySpec.Scheme == "" {
		cbContainersCluster.Spec.ApiGatewaySpec.Scheme = "https"
	}

	if cbContainersCluster.Spec.ApiGatewaySpec.Port == 0 {
		cbContainersCluster.Spec.ApiGatewaySpec.Port = 443
	}

	if cbContainersCluster.Spec.ApiGatewaySpec.Adapter == "" {
		cbContainersCluster.Spec.ApiGatewaySpec.Adapter = "containers"
	}

	if cbContainersCluster.Spec.ApiGatewaySpec.AccessTokenSecretName == "" {
		cbContainersCluster.Spec.ApiGatewaySpec.AccessTokenSecretName = defaultAccessToken
	}

	if cbContainersCluster.Spec.EventsGatewaySpec.Port == 0 {
		cbContainersCluster.Spec.EventsGatewaySpec.Port = 443
	}

	if cbContainersCluster.Spec.MonitorSpec.Labels == nil {
		cbContainersCluster.Spec.MonitorSpec.Labels = make(map[string]string)
	}

	if cbContainersCluster.Spec.MonitorSpec.DeploymentAnnotations == nil {
		cbContainersCluster.Spec.MonitorSpec.DeploymentAnnotations = make(map[string]string)
	}

	if cbContainersCluster.Spec.MonitorSpec.PodTemplateAnnotations == nil {
		cbContainersCluster.Spec.MonitorSpec.PodTemplateAnnotations = make(map[string]string)
	}

	if cbContainersCluster.Spec.MonitorSpec.Env == nil {
		cbContainersCluster.Spec.MonitorSpec.Env = make(map[string]string)
	}

	setDefaultImage(&cbContainersCluster.Spec.MonitorSpec.Image, "cbartifactory/monitor")

	if err := setDefaultResourceRequirements(&cbContainersCluster.Spec.MonitorSpec.Resources, "64Mi", "30m", "256Mi", "200m"); err != nil {
		return err
	}

	setDefaultHTTPProbes(&cbContainersCluster.Spec.MonitorSpec.Probes)

	return nil
}

func (r *CBContainersClusterReconciler) getRegistrySecretValues(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersCluster, accessToken string) (*models.RegistrySecretValues, error) {
	return r.ClusterProcessor.Process(cbContainersCluster, accessToken)
}

func (r *CBContainersClusterReconciler) getAccessToken(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersCluster) (string, error) {
	accessTokenSecretNamespacedName := types.NamespacedName{Name: cbContainersCluster.Spec.ApiGatewaySpec.AccessTokenSecretName, Namespace: commonState.DataPlaneNamespaceName}
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
		Owns(&corev1.Secret{}).
		Owns(r.ClusterStateApplier.GetPriorityClassEmptyK8sObject()).
		Complete(r)
}
