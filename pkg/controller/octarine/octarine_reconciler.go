package octarine

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/mitchellh/mapstructure"
	"github.com/octarinesec/octarine-operator/pkg/helm_utils"
	. "github.com/octarinesec/octarine-operator/pkg/octarine_api"
	. "github.com/octarinesec/octarine-operator/pkg/types"
	"github.com/operator-framework/operator-sdk/pkg/helm/watches"
	"github.com/redhat-cop/operator-utils/pkg/util"
	"helm.sh/helm/v3/pkg/chart/loader"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	k8serr "k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strconv"
)

// Annotation to set on the Octarine CR instance, indicates whether it has been initialized
const annotationInitialized = "operator.octarinesec.com/initialized"

// blank assignment to verify that ReconcileOctarine implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileOctarine{}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, helmWatch watches.Watch) (reconcile.Reconciler, error) {
	// Load the chart
	chart, err := loader.LoadDir(helmWatch.ChartDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart dir: %w", err)
	}

	return &ReconcileOctarine{
		ReconcilerBase:    util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor(controllerName)),
		gvk:               helmWatch.GroupVersionKind,
		helmDefaultValues: helm_utils.GetDefaultValues(chart),
	}, nil
}

// ReconcileOctarine reconciles a Octarine object.
// Octarine object is also reconciled (in parallel) by a helm operator, thus it's an unstructured object (values of the
// helm chart).
type ReconcileOctarine struct {
	// Provides operator utils
	util.ReconcilerBase

	// The GVK this reconciler watches & handles
	gvk schema.GroupVersionKind

	// The default Helm values for the kind managed by the corresponding helm operator
	helmDefaultValues map[string]interface{}
}

// Return logger for the given request, using its namespace and name
func reqLogger(request reconcile.Request) logr.Logger {
	return log.WithValues(
		"namespace", request.Namespace,
		"name", request.Name,
	)
}

// Reconcile reads that state of the cluster for a Octarine object and makes changes based on the state read
// and what is in the Octarine.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileOctarine) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := reqLogger(request)
	reqLogger.Info("Reconciling Octarine Dataplane")
	octarine := r.octarineCR(request)

	if err := r.GetClient().Get(context.TODO(), request.NamespacedName, octarine); err != nil {
		if k8serr.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get resource.")
		return reconcile.Result{}, err
	}

	octarineSpec, err := r.parseSpec(octarine)
	if err != nil {
		reqLogger.Error(err, "Failed parsing the CR spec as an OctarineSpec")
		return reconcile.Result{}, err
	}

	if err := r.populateAccessToken(octarineSpec, request.Namespace); err != nil {
		reqLogger.Error(err, "Failed populating Octarine access token")
		return reconcile.Result{}, err
	}

	if !r.isInitialized(octarine) {
		if err := r.initialize(reqLogger, octarine, octarineSpec); err != nil {
			return reconcile.Result{}, err
		}
	}

	if util.IsBeingDeleted(octarine) {
		if err := r.cleanup(reqLogger, octarine); err != nil {
			return reconcile.Result{}, err
		}
	}

	if err := r.reconcileMonitor(reqLogger, octarine, octarineSpec); err != nil {
		reqLogger.Error(err, "Failed reconciling monitor agent")
		return reconcile.Result{}, err
	}

	if octarineSpec.Guardrails.Enabled {
		if err := r.reconcileGuardrails(reqLogger, octarine, octarineSpec); err != nil {
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("Reconciled Octarine Dataplane")
	return reconcile.Result{}, nil
}

// Returns Octarine CR (unstructured as it's watched by this operator and the helm operator)
func (r *ReconcileOctarine) octarineCR(request reconcile.Request) *unstructured.Unstructured {
	octarine := &unstructured.Unstructured{}
	octarine.SetGroupVersionKind(r.gvk)
	octarine.SetNamespace(request.Namespace)
	octarine.SetName(request.Name)
	return octarine
}

// Returns true if the given octarine CR instance has already been initialized
func (r *ReconcileOctarine) isInitialized(octarine *unstructured.Unstructured) bool {
	val, ok := octarine.GetAnnotations()[annotationInitialized]
	if !ok {
		return false
	}

	initialized, _ := strconv.ParseBool(val)
	return initialized
}

// Initializes the octarine installation
func (r *ReconcileOctarine) initialize(reqLogger logr.Logger, octarine *unstructured.Unstructured, octarineSpec *OctarineSpec) error {
	reqLogger.V(1).Info("initializing octarine CR instance", "octarine CR", octarine)

	if err := r.syncOctarineAccountDetails(reqLogger, octarine, octarineSpec); err != nil {
		reqLogger.Error(err, "Failed reconciling octarine account")
		return err
	}

	if err := r.labelNs(reqLogger, octarine.GetNamespace()); err != nil {
		reqLogger.Error(err, "Failed labeling the namespace")
		return err
	}

	// Add this controller as a finalizer to the octarine CR instance
	util.AddFinalizer(octarine, controllerName)

	// Set the octarine CR instance as initialized
	annotations := octarine.GetAnnotations()
	annotations[annotationInitialized] = strconv.FormatBool(true)
	octarine.SetAnnotations(annotations)

	// Update the octarine CR instance to set the finalizer
	err := r.updateResource(octarine)
	if err != nil {
		reqLogger.Error(err, "Failed updating Octarine CR instance")
		return err
	}

	reqLogger.V(1).Info("Finished Octarine initialization")

	return nil
}

// Cleans up the octarine installation.
// This is based on the finalizer that was added to the instance - if it doesn't exist, then the instance has already
// been cleaned up.
func (r *ReconcileOctarine) cleanup(reqLogger logr.Logger, octarine *unstructured.Unstructured) error {
	if !util.HasFinalizer(octarine, controllerName) {
		// Resource is already terminated (and has already been cleaned up)
		reqLogger.V(1).Info("Resource is terminated, skipping cleanup")
		return nil
	}
	reqLogger.V(1).Info("Cleaning up", "octarine CR", octarine)

	// add cleanup logic here

	// Cleanup done - remove finalizer
	util.RemoveFinalizer(octarine, controllerName)

	// Update the octarine CR instance to remove the finalizer
	err := r.updateResource(octarine)
	if err != nil {
		reqLogger.Error(err, "Failed updating Octarine CR instance")
		return err
	}

	reqLogger.V(1).Info("Finished cleaning up")

	return nil
}

func (r *ReconcileOctarine) updateResource(o runtime.Object) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.GetClient().Update(context.TODO(), o)
	})
}

// Labels the octarine namespace, which is important for Guardrails' webhook namespace selector
func (r *ReconcileOctarine) labelNs(reqLogger logr.Logger, namespace string) error {
	// The labels to set for the namespace
	labels := map[string]string{
		"name":     namespace,
		"octarine": "ignore",
	}
	updated := false
	ns := &corev1.Namespace{}

	// Fetch current namespace labels
	if err := r.GetClient().Get(context.TODO(), types.NamespacedName{Name: namespace, Namespace: ""}, ns); err != nil {
		return err
	}
	nsLabels := ns.GetLabels()
	if nsLabels == nil {
		nsLabels = make(map[string]string)
	}

	// For each label, if it doesn't exist or it has a different value - set it's new value
	for key, val := range labels {
		if actualVal, ok := nsLabels[key]; !ok || actualVal != val {
			nsLabels[key] = val
			updated = true
		}
	}

	if updated {
		reqLogger.V(1).Info("Labeling namespace")
		if err := r.GetClient().Update(context.TODO(), ns); err != nil {
			return err
		}
	}

	return nil
}

// Syncs Octarine account details with the backend - registers the domain and the account features
func (r *ReconcileOctarine) syncOctarineAccountDetails(reqLogger logr.Logger, octarine *unstructured.Unstructured, octarineSpec *OctarineSpec) error {
	apiClient := NewOctarineApiClient(octarineSpec.Global.Octarine.Account, octarineSpec.Global.Octarine.AccessToken,
		octarineSpec.Global.Octarine.Api)

	if err := r.createRegistrySecret(reqLogger, apiClient, octarine); err != nil {
		reqLogger.Error(err, "Failed creating registry secret")
		return err
	}

	if err := r.registerDomain(reqLogger, apiClient, octarineSpec); err != nil {
		reqLogger.Error(err, "Failed registering domain")
		return err
	}

	if err := r.registerAccountFeatures(reqLogger, apiClient, octarineSpec); err != nil {
		reqLogger.Error(err, "Failed registering account features")
		return err
	}

	return nil
}

func (r *ReconcileOctarine) createRegistrySecret(reqLogger logr.Logger, apiClient *OctarineApiClient, octarine *unstructured.Unstructured) error {
	reqLogger.V(1).Info("reconciling registry secret")

	secretName, err := registrySecretName()
	if err != nil {
		return err
	}

	// Find secret
	found := &corev1.Secret{}
	err = r.GetClient().Get(context.TODO(), secretName, found)
	if err != nil && !k8serr.IsNotFound(err) {
		return err
	} else if k8serr.IsNotFound(err) {
		// Secret doesn't exist - get it from the backend
		registrySecret, err := apiClient.GetRegistrySecret()
		if err != nil {
			return err
		}

		// Create the secret
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName.Name,
				Namespace: secretName.Namespace,
			},
			Type: registrySecret.Type,
			Data: registrySecret.Data,
		}

		// Create secret in k8s
		reqLogger.V(1).Info("creating registry secret")
		if err := r.CreateOrUpdateResource(octarine, "", secret); err != nil {
			return err
		}
	}

	return nil
}

// Registers the domain in Octarine's control plane
func (r *ReconcileOctarine) registerDomain(reqLogger logr.Logger, apiClient *OctarineApiClient, octarineSpec *OctarineSpec) error {
	domain := octarineSpec.Global.Octarine.Domain
	reqLogger.V(1).Info("Registering domain", "domain", domain)
	if err := apiClient.RegisterDomain(domain); err != nil {
		return err
	}
	return nil
}

// Registers the account features in Octarine's control plane, according to the features which are installed (guardrails, nodeguard)
func (r *ReconcileOctarine) registerAccountFeatures(reqLogger logr.Logger, apiClient *OctarineApiClient, octarineSpec *OctarineSpec) error {
	features := octarineSpec.GetAccountFeatures()
	reqLogger.V(1).Info("Registering account features", "account features", features)
	if err := apiClient.RegisterAccountFeatures(features...); err != nil {
		return err
	}
	return nil
}

// Parses the given unstructured object's spec as an OctarineSpec. This will also add default helm chart values to the
// OctarineSpec from the chart's values.yaml.
func (r *ReconcileOctarine) parseSpec(o *unstructured.Unstructured) (*OctarineSpec, error) {
	spec, ok := o.Object["spec"].(map[string]interface{})
	if !ok {
		return nil, errors.New("failed to get spec: expected map[string]interface{}")
	}

	octarineSpec := NewOctarineSpec()

	// First load the charts default values
	if err := mapstructure.Decode(r.helmDefaultValues, octarineSpec); err != nil {
		return nil, fmt.Errorf("failed decoding Helm chart default values into OctarineSpec")
	}

	// Load values from the CR (they override the chart values)
	if err := mapstructure.Decode(spec, octarineSpec); err != nil {
		return nil, fmt.Errorf("failed decoding CR spec into OctarineSpec")
	}

	return octarineSpec, nil
}

// Populates the access token from the Octarine spec - if it wasn't explicitly set, read it from the secret
func (r *ReconcileOctarine) populateAccessToken(spec *OctarineSpec, namespace string) error {
	if spec.Global.Octarine.AccessToken == "" {
		accessTokenSecret := spec.Global.Octarine.AccessTokenSecret
		if accessTokenSecret == "" {
			return errors.New("either global.octarine.accessToken or global.octarine.accessTokenSecret must be specified")
		}

		// Get the secret which contains the access token
		secret := &corev1.Secret{}
		if err := r.GetClient().Get(context.TODO(), types.NamespacedName{Name: accessTokenSecret, Namespace: namespace}, secret); err != nil {
			return fmt.Errorf("failed to get secret %s: %v", accessTokenSecret, err)
		}

		// Get the access token from the secret
		accessToken := string(secret.Data["accessToken"])
		if accessToken == "" {
			return fmt.Errorf("access token wasn't found in secret %s", accessTokenSecret)
		}

		// Set the token in Octarine spec
		spec.Global.Octarine.AccessToken = accessToken
	}

	return nil
}
