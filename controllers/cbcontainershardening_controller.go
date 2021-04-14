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
	"k8s.io/apimachinery/pkg/api/resource"
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
	r.Log.Info("\n\n")
	r.Log.Info("Got reconcile request", "namespaced name", req.NamespacedName)
	r.Log.Info("Starting reconciling")

	r.Log.Info("Getting CBContainersHardening object")
	cbContainersHardening, err := r.getContainersHardeningObject(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	if cbContainersHardening == nil {
		return ctrl.Result{}, nil
	}

	if err := r.setDefaults(cbContainersHardening); err != nil {
		return ctrl.Result{}, fmt.Errorf("faild to set defaults to CR: %v", err)
	}

	setOwner := func(controlledResource metav1.Object) error {
		return ctrl.SetControllerReference(cbContainersHardening, controlledResource, r.Scheme)
	}

	r.Log.Info("Applying desired state")
	stateWasChanged, err := r.HardeningStateApplier.ApplyDesiredState(ctx, cbContainersHardening, r.Client, setOwner)
	if err != nil {
		return ctrl.Result{}, err
	}

	r.Log.Info("Finished reconciling", "Requiring", stateWasChanged)
	r.Log.Info("\n\n")
	return ctrl.Result{Requeue: stateWasChanged}, nil
}

func (r *CBContainersHardeningReconciler) setDefaults(cbContainersHardening *cbcontainersv1.CBContainersHardening) error {
	if cbContainersHardening.Spec.AccessTokenSecretName == "" {
		cbContainersHardening.Spec.AccessTokenSecretName = "cbcontainers-access-token"
	}

	if cbContainersHardening.Spec.AccessTokenSecretName == "" {
		cbContainersHardening.Spec.AccessTokenSecretName = "cbcontainers-access-token"
	}

	if cbContainersHardening.Spec.EnforcerSpec.Labels == nil {
		cbContainersHardening.Spec.EnforcerSpec.Labels = make(map[string]string)
	}

	if cbContainersHardening.Spec.EnforcerSpec.DeploymentAnnotations == nil {
		cbContainersHardening.Spec.EnforcerSpec.DeploymentAnnotations = make(map[string]string)
	}

	if cbContainersHardening.Spec.EnforcerSpec.PodTemplateAnnotations == nil {
		cbContainersHardening.Spec.EnforcerSpec.PodTemplateAnnotations = map[string]string{
			"prometheus.io/scrape": "false",
			"prometheus.io/port":   "7071",
		}
	}

	if cbContainersHardening.Spec.EnforcerSpec.ReplicasCount == nil {
		defaultReplicaCount := int32(1)
		cbContainersHardening.Spec.EnforcerSpec.ReplicasCount = &defaultReplicaCount
	}

	if cbContainersHardening.Spec.EnforcerSpec.Env == nil {
		cbContainersHardening.Spec.EnforcerSpec.Env = make(map[string]string)
	}

	r.setDefaultImage(&cbContainersHardening.Spec.EnforcerSpec.Image, "cbartifactory/guardrails-enforcer")

	if err := r.setDefaultResourceRequirements(cbContainersHardening.Spec.EnforcerSpec.Resources, "64Mi", "30m", "256Mi", "200m"); err != nil {
		return err
	}

	r.setDefaultProbes(&cbContainersHardening.Spec.EnforcerSpec.Probes)

	if cbContainersHardening.Spec.EnforcerSpec.WebhookTimeoutSeconds == 0 {
		cbContainersHardening.Spec.EnforcerSpec.WebhookTimeoutSeconds = 5
	}

	if cbContainersHardening.Spec.StateReporterSpec.Labels == nil {
		cbContainersHardening.Spec.StateReporterSpec.Labels = make(map[string]string)
	}

	if cbContainersHardening.Spec.StateReporterSpec.DeploymentAnnotations == nil {
		cbContainersHardening.Spec.StateReporterSpec.DeploymentAnnotations = make(map[string]string)
	}

	if cbContainersHardening.Spec.StateReporterSpec.PodTemplateAnnotations == nil {
		cbContainersHardening.Spec.StateReporterSpec.PodTemplateAnnotations = map[string]string{
			"prometheus.io/scrape": "false",
			"prometheus.io/port":   "7071",
		}
	}

	if cbContainersHardening.Spec.StateReporterSpec.Env == nil {
		cbContainersHardening.Spec.StateReporterSpec.Env = make(map[string]string)
	}

	r.setDefaultImage(&cbContainersHardening.Spec.StateReporterSpec.Image, "cbartifactory/guardrails-state-reporter")

	if err := r.setDefaultResourceRequirements(cbContainersHardening.Spec.StateReporterSpec.Resources, "64Mi", "30m", "256Mi", "200m"); err != nil {
		return err
	}

	r.setDefaultProbes(&cbContainersHardening.Spec.StateReporterSpec.Probes)

	if cbContainersHardening.Spec.EventsGatewaySpec.Port == 0 {
		cbContainersHardening.Spec.EventsGatewaySpec.Port = 443
	}

	return nil
}

func (r *CBContainersHardeningReconciler) setDefaultProbes(probesSpec *cbcontainersv1.CBContainersHardeningProbesSpec) {
	if probesSpec.ReadinessPath == "" {
		probesSpec.ReadinessPath = "/ready"
	}

	if probesSpec.LivenessPath == "" {
		probesSpec.LivenessPath = "/alive"
	}

	if probesSpec.Port == 0 {
		probesSpec.Port = 8181
	}

	if probesSpec.Scheme == "" {
		probesSpec.Scheme = coreV1.URISchemeHTTP
	}

	if probesSpec.InitialDelaySeconds == 0 {
		probesSpec.InitialDelaySeconds = 3
	}

	if probesSpec.TimeoutSeconds == 0 {
		probesSpec.TimeoutSeconds = 1
	}

	if probesSpec.PeriodSeconds == 0 {
		probesSpec.PeriodSeconds = 30
	}

	if probesSpec.SuccessThreshold == 0 {
		probesSpec.SuccessThreshold = 1
	}

	if probesSpec.FailureThreshold == 0 {
		probesSpec.FailureThreshold = 3
	}
}

func (r *CBContainersHardeningReconciler) setDefaultImage(imageSpec *cbcontainersv1.CBContainersHardeningImageSpec, imageName string) {
	if imageSpec.Repository == "" {
		imageSpec.Repository = imageName
	}

	if imageSpec.PullPolicy == "" {
		imageSpec.PullPolicy = "Always"
	}
}

func (r *CBContainersHardeningReconciler) setDefaultResourceRequirements(resources coreV1.ResourceRequirements, requestMemory, requestCpu, limitMemory, limitCpu string) error {
	if resources.Requests == nil {
		resources.Requests = make(coreV1.ResourceList)
	}

	if err := r.setDefaultsResourcesList(resources.Requests, requestMemory, requestCpu); err != nil {
		return err
	}

	if resources.Limits == nil {
		resources.Limits = make(coreV1.ResourceList)
	}

	if err := r.setDefaultsResourcesList(resources.Limits, limitMemory, limitCpu); err != nil {
		return err
	}

	return nil
}

func (r *CBContainersHardeningReconciler) setDefaultsResourcesList(list coreV1.ResourceList, memory, cpu string) error {
	if err := r.setDefaultResource(list, coreV1.ResourceMemory, memory); err != nil {
		return err
	}

	if err := r.setDefaultResource(list, coreV1.ResourceCPU, cpu); err != nil {
		return err
	}

	return nil
}

func (r *CBContainersHardeningReconciler) setDefaultResource(list coreV1.ResourceList, resourceName coreV1.ResourceName, value string) error {
	if _, ok := list[resourceName]; !ok {
		quantity, err := resource.ParseQuantity(value)
		if err != nil {
			return err
		}

		list[resourceName] = quantity
	}

	return nil
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
