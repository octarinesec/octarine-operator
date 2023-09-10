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
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"time"

	"github.com/vmware/cbcontainers-operator/cbcontainers/state/adapters"
	appsV1 "k8s.io/api/apps/v1"

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

const (
	// conflictRetryTime should be used when hitting 409 status from the API Server on CR updates
	// for this case it is better to use a fixed requeue duration instead of the default exponential backoff to prevent multiple concurrent changes from holding back the reconcile queue without reason
	conflictRetryTime = 3 * time.Second
)

type StateApplier interface {
	ApplyDesiredState(ctx context.Context, agentSpec *cbcontainersv1.CBContainersAgentSpec, secret *models.RegistrySecretValues, setOwner applymentOptions.OwnerSetter) (bool, error)
	ShouldProcessEvent(client.Object) bool
}

type AgentProcessor interface {
	Process(cbContainersAgent *cbcontainersv1.CBContainersAgent, accessToken string) (*models.RegistrySecretValues, error)
}

type CBContainersAgentController struct {
	client.Client
	Log              logr.Logger
	Scheme           *runtime.Scheme
	ClusterProcessor AgentProcessor
	StateApplier     StateApplier
	K8sVersion       string
	// Namespace is the kubernetes namespace for all agent components
	Namespace string
}

func (r *CBContainersAgentController) getContainersAgentObject(ctx context.Context) (*cbcontainersv1.CBContainersAgent, error) {
	cbContainersAgentsList := &cbcontainersv1.CBContainersAgentList{}
	if err := r.List(ctx, cbContainersAgentsList); err != nil {
		return nil, fmt.Errorf("couldn't list CBContainersAgent k8s objects: %v", err)
	}

	if cbContainersAgentsList.Items == nil || len(cbContainersAgentsList.Items) == 0 {
		return nil, nil
	}

	if len(cbContainersAgentsList.Items) > 1 {
		return nil, fmt.Errorf("there is more than 1 CBContainersAgent k8s object, please delete unwanted resources")
	}

	return &cbContainersAgentsList.Items[0], nil
}

// The following values must be kept in-sync with constants for generated RBAC to work properly:
// - default dataplane namespace (see common.DataPlaneNamespaceName)
// - cluster-wide dataplane priority class  (see common.DataPlanePriorityClassName) - avoids access to all priority classes on the cluster
// - the cluster-wide webhooks (see components.EnforcerName) - avoids access to all webhooks on the cluster

// +kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainersagents,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainersagents/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.containers.carbonblack.io,resources=cbcontainersagents/finalizers,verbs=update
// +kubebuilder:rbac:groups=scheduling.k8s.io,resources=priorityclasses,verbs=delete;get;patch;update,resourceNames=cbcontainers-dataplane-priority-class
// +kubebuilder:rbac:groups=scheduling.k8s.io,resources=priorityclasses,verbs=create;list;watch
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources={validatingwebhookconfigurations,mutatingwebhookconfigurations},verbs=delete;get;patch;update,resourceNames=cbcontainers-hardening-enforcer
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources={validatingwebhookconfigurations,mutatingwebhookconfigurations},verbs=create;list;watch
// +kubebuilder:rbac:groups={core},resources={nodes},verbs=list
// +kubebuilder:rbac:groups={core},resources={namespaces},verbs=get
// +kubebuilder:rbac:groups={policy},resources={podsecuritypolicies},verbs=use,resourceNames={cbcontainers-manager-psp}
// +kubebuilder:rbac:groups={apps,core},resources={deployments,services,daemonsets},namespace=cbcontainers-dataplane,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=core,resources={configmaps,secrets},namespace=cbcontainers-dataplane,verbs=get;list;watch;create;update;patch;delete;deletecollection

func (r *CBContainersAgentController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("\n\n")
	r.Log.Info("Got reconcile request", "namespaced name", req.NamespacedName)
	r.Log.Info("Starting reconciling")

	r.Log.Info("Getting CBContainersAgent object")
	cbContainersAgent, err := r.getContainersAgentObject(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	if cbContainersAgent == nil {
		r.Log.Info("No CBContainersAgent object found")
		return ctrl.Result{}, nil
	}

	if err := r.setAgentDefaults(&cbContainersAgent.Spec); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set defaults to cluster CR: %v", err)
	}

	setOwner := func(controlledResource metav1.Object) error {
		return ctrl.SetControllerReference(cbContainersAgent, controlledResource, r.Scheme)
	}

	accessToken, err := r.getAccessToken(context.Background(), cbContainersAgent)
	if err != nil {
		return ctrl.Result{}, err
	}

	var registrySecret *models.RegistrySecretValues
	if cbContainersAgent.Spec.Components.Settings.ShouldCreateDefaultImagePullSecrets() {
		r.Log.Info("Getting registry secret values")
		registrySecret, err = r.getRegistrySecretValues(ctx, cbContainersAgent, accessToken)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		r.Log.Info(`Skipping default image pull secrets creation, because "spec.components.basic.createImagePullSecrets" is set to "false"`)
	}

	r.Log.Info("Applying desired state")
	stateWasChanged, err := r.StateApplier.ApplyDesiredState(ctx, &cbContainersAgent.Spec, registrySecret, setOwner)
	if err != nil {
		return ctrl.Result{}, err
	}

	r.Log.Info("Finished reconciling", "Requiring", stateWasChanged)

	if err = r.updateCRStatus(ctx, cbContainersAgent, stateWasChanged); err != nil {
		if k8sErrors.IsConflict(err) {
			r.Log.Info("Custom resource was changed during reconciliation, scheduling another iteration to fully update status")
			// Something changed in the CR while we were doing updates, requeue in a bit to get fresh data
			// Note: this is the recommended approach for operator-sdk instead of retrying the Get->Update cycle within a reconciliation
			return ctrl.Result{RequeueAfter: conflictRetryTime}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to update CBContainersAgent status: %w", err)
	}

	r.Log.Info("\n\n")
	return ctrl.Result{Requeue: stateWasChanged}, nil
}

func (r *CBContainersAgentController) getRegistrySecretValues(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersAgent, accessToken string) (*models.RegistrySecretValues, error) {
	return r.ClusterProcessor.Process(cbContainersCluster, accessToken)
}

func (r *CBContainersAgentController) getAccessToken(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersAgent) (string, error) {
	accessTokenSecretNamespacedName := types.NamespacedName{Name: cbContainersCluster.Spec.AccessTokenSecretName, Namespace: r.Namespace}
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

func (r *CBContainersAgentController) updateCRStatus(ctx context.Context, cbContainersCluster *cbcontainersv1.CBContainersAgent, agentStateWasChanged bool) error {
	// If we don't expect more changes (i.e. nothing changed in reality) and we haven't updated the status, we do so now.
	if !agentStateWasChanged && cbContainersCluster.Status.ObservedGeneration < cbContainersCluster.ObjectMeta.Generation {
		r.Log.Info("Updating CBContainersAgent status")
		cbContainersCluster.Status.ObservedGeneration = cbContainersCluster.ObjectMeta.Generation
		return r.Client.Status().Update(ctx, cbContainersCluster)
	}
	return nil
}

func (r *CBContainersAgentController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cbcontainersv1.CBContainersAgent{}).
		WithEventFilter(NewCBContainersGenerationChangedPredicate(r.StateApplier)).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(adapters.EmptyPriorityClassForVersion(r.K8sVersion)).
		Owns(&appsV1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&appsV1.DaemonSet{}).
		Owns(adapters.EmptyValidatingWebhookConfigForVersion(r.K8sVersion)).
		Owns(adapters.EmptyMutatingWebhookConfigForVersion(r.K8sVersion)).
		Complete(r)
}
