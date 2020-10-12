package octarine

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/octarinesec/octarine-operator/pkg/tls_utils"
	"github.com/octarinesec/octarine-operator/pkg/types"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8stypes "k8s.io/apimachinery/pkg/types"
	k8serr "k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciles Guardrails - if its deployment is available the webhook will be configured, otherwise the webhook will be
// deleted (if it's configured)
func (r *ReconcileOctarine) reconcileGuardrails(reqLogger logr.Logger, octarine *unstructured.Unstructured, octarineSpec *types.OctarineSpec) error {
	reqLogger.V(1).Info("reconciling guardrails webhook")
	if !octarineSpec.Guardrails.Enforcer.AdmissionController.AutoManage {
		reqLogger.V(2).Info("Guardrails.Enforcer.AdmissionController.AutoManage is disabled")
		if err := r.reconcileGuardrailsSecret(reqLogger, octarine); err != nil {
			reqLogger.Error(err, "error reconciling guardrails secret")
			return err
		}

		if err := r.reconcileGuardrailsWebhook(reqLogger, octarine, octarineSpec); err != nil {
			reqLogger.Error(err, "error reconciling guardrails webhook")
			return err
		}

		return nil
	}

	reqLogger.V(2).Info("Guardrails.AdmissionController.AutoManage is enabled")
	available, err := r.guardrailsDeploymentAvailable(reqLogger, octarine)
	if err != nil {
		reqLogger.Error(err, "error determining guardrails deployment availability")
		return err
	}

	if available {
		reqLogger.V(1).Info("Guardrails deployment available")

		if err := r.reconcileGuardrailsSecret(reqLogger, octarine); err != nil {
			reqLogger.Error(err, "error reconciling guardrails secret")
			return err
		}

		if err := r.reconcileGuardrailsWebhook(reqLogger, octarine, octarineSpec); err != nil {
			reqLogger.Error(err, "error reconciling guardrails webhook")
			return err
		}
	} else {
		reqLogger.V(1).Info("Guardrails deployment not available")

		if err := r.deleteGuardrailsWebhook(reqLogger, octarine); err != nil {
			reqLogger.Error(err, "error deleting guardrails webhook")
			return err
		}
	}

	return nil
}

// Returns true if Guardrails deployment is available (determined by the 'Available' condition of the deployment)
func (r *ReconcileOctarine) guardrailsDeploymentAvailable(reqLogger logr.Logger, octarine *unstructured.Unstructured) (bool, error) {
	// Matchers for listing the guardrails deployment(s) - matching by app name label (set by helm) and the namespace
	matchers := []client.ListOption{
		client.MatchingLabels{"app.kubernetes.io/name": "guardrails"},
		client.InNamespace(octarine.GetNamespace()),
	}

	// Get matching deployments
	foundDeps := &v1.DeploymentList{}
	if err := r.GetClient().List(context.TODO(), foundDeps, matchers...); err != nil && k8serr.IsNotFound(err) {
		// No deployments found
		reqLogger.V(2).Info("no guardrails deployment(s) found")
		return false, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed getting deployments")
		return false, err
	}

	// Check the status of the found deployment
	if len(foundDeps.Items) > 0 {
		// Assuming only one guardrails deployment can be deployed
		dep := foundDeps.Items[0]
		for _, condition := range dep.Status.Conditions {
			if condition.Type == v1.DeploymentAvailable && condition.Status == corev1.ConditionTrue {
				return true, nil
			}
		}
	}

	return false, nil
}

// Deletes guardrails webhook config
func (r *ReconcileOctarine) deleteGuardrailsWebhook(reqLogger logr.Logger, octarine *unstructured.Unstructured) error {
	webhookName := guardrailsWebhookName(octarine)
	found := &admissionregistrationv1beta1.ValidatingWebhookConfiguration{}
	err := r.GetClient().Get(context.TODO(), k8stypes.NamespacedName{Name: webhookName, Namespace: ""}, found)
	if err != nil && !k8serr.IsNotFound(err) {
		return err
	} else if err == nil {
		reqLogger.V(1).Info("deleting guardrails webhook")
		if err := r.GetClient().Delete(context.TODO(), found); err != nil {
			return err
		}
	}

	return nil
}

// Reconciles guardrails webhook TLS secret
func (r *ReconcileOctarine) reconcileGuardrailsSecret(reqLogger logr.Logger, octarine *unstructured.Unstructured) error {
	reqLogger.V(1).Info("reconciling guardrails secret")

	secretName, err := guardrailsSecretName(octarine)
	if err != nil {
		return err
	}
	serviceName, err := guardrailsServiceName(octarine)
	if err != nil {
		return err
	}

	// Find secret
	found := &corev1.Secret{}
	err = r.GetClient().Get(context.TODO(), secretName, found)
	if err != nil && !k8serr.IsNotFound(err) {
		return err
	} else if err != nil {
		// Secret doesn't exist - create it
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName.Name,
				Namespace: secretName.Namespace,
			},
			Data: map[string][]byte{},
		}

		// Create CA
		caCert, caKey, err := tls_utils.CreateCertificateAuthority(reqLogger)
		if err != nil {
			return err
		}
		secret.Data["ca.crt"] = caCert
		secret.Data["ca.key"] = caKey

		// Create Cert
		cert, key, err := tls_utils.CreateCertFromCA(reqLogger, serviceName, caCert, caKey)
		if err != nil {
			return err
		}
		secret.Data["signed_cert"] = cert
		secret.Data["key"] = key

		// Create secret in k8s
		reqLogger.V(1).Info("creating/updating guardrails webhook tls secret")
		if err := r.CreateOrUpdateResource(octarine, "", secret); err != nil {
			return err
		}
	}

	return nil
}

// Reconciles guardrails webhook
func (r *ReconcileOctarine) reconcileGuardrailsWebhook(reqLogger logr.Logger, octarine *unstructured.Unstructured, octarineSpec *types.OctarineSpec) error {
	reqLogger.V(1).Info("reconciling guardrails validating webhook")

	secretName, err := guardrailsSecretName(octarine)
	if err != nil {
		return err
	}
	serviceName, err := guardrailsServiceName(octarine)
	if err != nil {
		return err
	}
	webhookName := guardrailsWebhookName(octarine)
	policy := admissionregistrationv1beta1.Ignore
	sideEffectsNoneOnDryRun := admissionregistrationv1beta1.SideEffectClassNoneOnDryRun
	sideEffectsNone := admissionregistrationv1beta1.SideEffectClassNone
	timeoutSeconds := int32(octarineSpec.Guardrails.Enforcer.AdmissionController.TimeoutSeconds)
	path := "/validate"

	// Read the CA bundle from the secret
	tlsSecret := &corev1.Secret{}
	err = r.GetClient().Get(context.TODO(), secretName, tlsSecret)
	if err != nil {
		return err
	}

	// Create namespace selectors
	var resourcesWebhookSelector, nsWebhookSelector *metav1.LabelSelector
	userSelector := octarineSpec.Guardrails.Enforcer.AdmissionController.NamespaceSelector
	if userSelector != nil {
		resourcesWebhookSelector = userSelector
		nsWebhookSelector = userSelector
	} else {
		resourcesWebhookSelector = &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "octarine",
					Operator: metav1.LabelSelectorOpNotIn,
					Values:   []string{"ignore"},
				},
				{
					Key:      "name",
					Operator: metav1.LabelSelectorOpNotIn,
					Values:   []string{serviceName.Namespace},
				},
			},
		}
	}

	// Create the validating webhook config
	webhookConfig := &admissionregistrationv1beta1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: webhookName,
		},
		Webhooks: []admissionregistrationv1beta1.ValidatingWebhook{
			{
				Name:              "resources.validating-webhook.octarine",
				FailurePolicy:     &policy,
				SideEffects:       &sideEffectsNoneOnDryRun,
				NamespaceSelector: resourcesWebhookSelector,
				Rules: []admissionregistrationv1beta1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1beta1.OperationType{admissionregistrationv1beta1.OperationAll},
						Rule: admissionregistrationv1beta1.Rule{
							APIGroups:   []string{"*"},
							APIVersions: []string{"*"},
							Resources: []string{"pods/portforward",
								"pods/exec",
								"namespaces",
								"pods",
								"replicasets",
								"services",
								"roles",
								"rolebindings",
								"clusterroles",
								"clusterrolebindings",
								"networkpolicies",
								"deployments",
								"replicationcontrollers",
								"daemonsets",
								"statefulsets",
								"jobs",
								"cronjobs",
								"ingresses"},
						},
					},
				},
				TimeoutSeconds: &timeoutSeconds,
				ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
					Service: &admissionregistrationv1beta1.ServiceReference{
						Namespace: serviceName.Namespace,
						Name:      serviceName.Name,
						Path:      &path,
					},
					CABundle: tlsSecret.Data["ca.crt"],
				},
			},
			{
				Name:              "namespaces.validating-webhook.octarine",
				FailurePolicy:     &policy,
				SideEffects:       &sideEffectsNone,
				NamespaceSelector: nsWebhookSelector,
				Rules: []admissionregistrationv1beta1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1beta1.OperationType{admissionregistrationv1beta1.Create, admissionregistrationv1beta1.Update},
						Rule: admissionregistrationv1beta1.Rule{
							APIGroups:   []string{"*"},
							APIVersions: []string{"*"},
							Resources:   []string{"namespaces"},
						},
					},
				},
				TimeoutSeconds: &timeoutSeconds,
				ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
					Service: &admissionregistrationv1beta1.ServiceReference{
						Namespace: serviceName.Namespace,
						Name:      serviceName.Name,
						Path:      &path,
					},
					CABundle: tlsSecret.Data["ca.crt"],
				},
			},
		},
	}

	reqLogger.V(1).Info("creating/updating guardrails webhook")
	if err := r.CreateOrUpdateResource(octarine, "", webhookConfig); err != nil {
		return err
	}

	return nil
}
