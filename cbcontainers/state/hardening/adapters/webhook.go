package adapters

import (
	admissionsV1 "k8s.io/api/admissionregistration/v1"
	admissionsV1Beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WebhookConfigurationAdapter interface {
	// SetLabels will update the validating webhook configuration's labels
	SetLabels(labels map[string]string)
	// GetWebhooks will return the webhooks attached to this configuration instance wrapped in an adapter.
	// The returned adapters wrap pointers to the original values so any modifications are propagated on the objects
	// For adding/removing webhooks to the list, use SetWebhooks
	GetWebhooks() []WebhookAdapter
	// SetWebhook§s will replace the configuration's webhooks with the provided list
	//
	// This method creates copied values of the provided webhooks so any pointers to the passed values will _not_ modify them directly
	// This method returns adapters for the inner webhooks after replacement to support this use-case
	// Calling SetXXX() to the returned adapters _will_ reflect the changes into the configuration's webhooks
	//
	// If any of the provided webhooks do not match the version of the adapter, an error is returned
	// E.g. passing a list of {v1, v1beta1, v1} wrapped webhooks to a v1 WebhookAdapter will produce an error
	SetWebhooks([]WebhookAdapter) ([]WebhookAdapter, error)
}

type WebhookAdapter interface {
	GetName() string
	SetName(name string)
	SetAdmissionReviewVersions(versions []string)
	SetFailurePolicy(policy string)
	SetSideEffects(sideEffectsClass string)
	SetMatchPolicy(policy string)
	GetNamespaceSelector() *metav1.LabelSelector
	SetNamespaceSelector(selector *metav1.LabelSelector)
	SetTimeoutSeconds(timeoutSeconds int32)
	SetCABundle(bundle []byte)
	SetServiceNamespace(namespace string)
	SetServiceName(name string)
	SetServicePath(path *string)
	GetAdmissionRules() []AdmissionRuleAdapter
	SetAdmissionRules([]AdmissionRuleAdapter)
}

// These are the same between v1 and v1beta1 - if they diverge; the adapters should handle internal conversion.
const (
	FailurePolicyFail   = string(admissionsV1.Fail)
	FailurePolicyIgnore = string(admissionsV1.Ignore)

	OperationAll     = string(admissionsV1.OperationAll)
	OperationCreate  = string(admissionsV1.Create)
	OperationUpdate  = string(admissionsV1.Update)
	OperationDelete  = string(admissionsV1.Delete)
	OperationConnect = string(admissionsV1.Connect)

	MatchPolicyExact      = string(admissionsV1.Exact)
	MatchPolicyEquivalent = string(admissionsV1.Equivalent)

	SideEffectsClassNone        = string(admissionsV1.SideEffectClassNone)
	SideEffectClassNoneOnDryRun = string(admissionsV1.SideEffectClassNoneOnDryRun)
)

// EmptyValidatingWebhookConfigForVersion returns an empty ValidatingWebhookConfiguration instance that is suitable for the provided k8s version
func EmptyValidatingWebhookConfigForVersion(k8sVersion string) client.Object {
	if k8sVersion < "v1.16" {
		return &admissionsV1Beta1.ValidatingWebhookConfiguration{}
	}

	return &admissionsV1.ValidatingWebhookConfiguration{}
}

// EmptyValidatingWebhookAdapterForVersion creates an empty ValidatingWebhook instance for the given k8s version and returns an adapter that wraps it
func EmptyValidatingWebhookAdapterForVersion(k8sVersion string) WebhookAdapter {
	if k8sVersion < "v1.16" {
		return (*validatingWebhookV1Beta1)(&admissionsV1Beta1.ValidatingWebhook{})
	} else {
		return (*validatingWebhookV1)(&admissionsV1.ValidatingWebhook{})
	}
}

func TryGetValidatingWebhookConfigurationAdapter(k8sObject client.Object) (WebhookConfigurationAdapter, bool) {
	switch value := k8sObject.(type) {
	case *admissionsV1.ValidatingWebhookConfiguration:
		return (*validatingWebhookConfigurationV1)(value), true
	case *admissionsV1Beta1.ValidatingWebhookConfiguration:
		return (*validatingWebhookConfigurationV1Beta1)(value), true
	}

	return nil, false
}

// EmptyMutatingWebhookConfigForVersion returns an empty MutatingWebhookConfiguration instance that is suitable for the provided k8s version
func EmptyMutatingWebhookConfigForVersion(k8sVersion string) client.Object {
	if k8sVersion < "v1.16" {
		return &admissionsV1Beta1.MutatingWebhookConfiguration{}
	}

	return &admissionsV1.MutatingWebhookConfiguration{}
}

// EmptyMutatingWebhookAdapterForVersion creates an empty MutatingWebhook instance for the given k8s version and returns an adapter that wraps it
func EmptyMutatingWebhookAdapterForVersion(k8sVersion string) WebhookAdapter {
	if k8sVersion < "v1.16" {
		return (*mutatingWebhookV1Beta1)(&admissionsV1Beta1.MutatingWebhook{})
	} else {
		return (*mutatingWebhookV1)(&admissionsV1.MutatingWebhook{})
	}
}

func TryGetMutatingWebhookConfigurationAdapter(k8sObject client.Object) (WebhookConfigurationAdapter, bool) {
	switch value := k8sObject.(type) {
	case *admissionsV1.MutatingWebhookConfiguration:
		return (*mutatingWebhookConfigurationV1)(value), true
	case *admissionsV1Beta1.MutatingWebhookConfiguration:
		return (*mutatingWebhookConfigurationV1Beta1)(value), true
	}

	return nil, false
}

// AdmissionRuleAdapter is a simple struct that mimics the admission.Rule struct
type AdmissionRuleAdapter struct {
	Operations  []string
	APIGroups   []string
	APIVersions []string
	Resources   []string
	Scope       *string
}