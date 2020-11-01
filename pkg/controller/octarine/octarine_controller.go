package octarine

import (
	"github.com/octarinesec/octarine-operator/pkg/predicates"
	"github.com/operator-framework/operator-sdk/pkg/handler"
	hcontroller "github.com/operator-framework/operator-sdk/pkg/helm/controller"
	"github.com/operator-framework/operator-sdk/pkg/helm/flags"
	"github.com/operator-framework/operator-sdk/pkg/helm/release"
	"github.com/operator-framework/operator-sdk/pkg/helm/watches"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	crthandler "sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const controllerName = "octarine-go-controller"

var (
	log = logf.Log.WithName(controllerName)

	HelmFlags *flags.HelmOperatorFlags
)

// Add creates new Octarine and Helm Controllers and adds them to the Manager. The Manager will set fields on the
// Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	helmWatches, err := addHelmController(mgr)
	if err != nil {
		return err
	}
	return addOctarineController(mgr, helmWatches)
}

// Adds a new Octarine Controller to mgr, using Octarine Reconciler
func addOctarineController(mgr manager.Manager, helmWatches []watches.Watch) error {

	for _, watch := range helmWatches {

		// Create a new controller
		reconciler, err := newReconciler(mgr, watch)
		if err != nil {
			return err
		}
		c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: reconciler})
		if err != nil {
			return err
		}

		// using predefined functions for filtering events
		dependentPredicate := predicates.DependentPredicateFuncs()

		// Object to watch - unstructured as it's a helm operator spec
		o := &unstructured.Unstructured{}
		o.SetGroupVersionKind(watch.GroupVersionKind)

		// Watch for changes to primary resource
		if err := c.Watch(&source.Kind{Type: o}, &crthandler.EnqueueRequestForObject{}); err != nil {
			return err
		}

		// Watch for changes to secondary resource Deployments and requeue the owner Octarine (for Guardrails webhook reconcilement according to deployment status)
		err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &crthandler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    o,
		}, dependentPredicate)
		if err != nil {
			return err
		}

		// Watch for changes to secondary resource validating webhook and requeue the owner Octarine (for Guardrails webhook)
		err = c.Watch(&source.Kind{Type: &admissionregistrationv1beta1.ValidatingWebhookConfiguration{}},
			&handler.EnqueueRequestForAnnotation{Type: watch.GroupKind().String()}, dependentPredicate)
		if err != nil {
			return err
		}

		// Watch for changes to secondary resource secret and requeue the owner Octarine (for Guardrails tls secret)
		err = c.Watch(&source.Kind{Type: &v1.Secret{}}, &crthandler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    o,
		}, dependentPredicate)
		if err != nil {
			return err
		}
	}

	return nil
}

// Adds a new Helm controller to mgr, to manage Octarine Helm release
func addHelmController(mgr manager.Manager) ([]watches.Watch, error) {
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		return nil, err
	}

	// Load helm operator watches
	helmWatches, err := watches.Load(HelmFlags.WatchesFile)
	if err != nil {
		return nil, err
	}

	for _, w := range helmWatches {
		// Register the controller with the factory.
		err := hcontroller.Add(mgr, hcontroller.WatchOptions{
			Namespace:               namespace,
			GVK:                     w.GroupVersionKind,
			ManagerFactory:          release.NewManagerFactory(mgr, w.ChartDir),
			ReconcilePeriod:         HelmFlags.ReconcilePeriod,
			WatchDependentResources: *w.WatchDependentResources,
			OverrideValues:          w.OverrideValues,
			MaxWorkers:              HelmFlags.MaxWorkers,
		})
		if err != nil {
			return nil, err
		}
	}

	return helmWatches, nil
}
