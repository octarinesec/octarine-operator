package adapters

import (
	admissionsV1 "k8s.io/api/admissionregistration/v1"
	admissionsV1Beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ValidatingWebhookConfigurationAdapter interface {
	// SetLabels will update the validating webhook configuration's labels
	SetLabels(labels map[string]string)
	// GetWebhooks will return the webhooks attached to this configuration instance wrapped in an adapter.
	// The returned adapters wrap pointers to the original values so any modifications are propagated on the objects
	// For adding/removing webhooks to the list, use SetWebhooks
	GetWebhooks() []ValidatingWebhookAdapter
	// SetWebhooks will replace the configuration's webhooks with the provided list
	// Any webhooks that don't match the configuration adapter's API version are ignored
	// E.g. passing a list of {v1, v1beta1, v1} wrapped webhooks to a v1 ValidatingWebhookAdapter will only set the 2 v1 webhooks and ignore the v1beta1
	SetWebhooks([]ValidatingWebhookAdapter)
}

type ValidatingWebhookAdapter interface {
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
	} else {
		return &admissionsV1.ValidatingWebhookConfiguration{}
	}
}

// EmptyValidatingWebhookAdapterForVersion creates an empty ValidatingWebhook instance for the given k8s version and returns an adapter that wraps it
func EmptyValidatingWebhookAdapterForVersion(k8sVersion string) ValidatingWebhookAdapter {
	if k8sVersion < "v1.16" {
		return (*validatingWebhookV1Beta1)(&admissionsV1Beta1.ValidatingWebhook{})
	} else {
		return (*validatingWebhookV1)(&admissionsV1.ValidatingWebhook{})
	}
}

func TryGetValidatingWebhookConfigurationAdapter(k8sObject client.Object) (ValidatingWebhookConfigurationAdapter, bool) {
	switch value := k8sObject.(type) {
	case *admissionsV1.ValidatingWebhookConfiguration:
		return (*validatingWebhookConfigurationV1)(value), true
	case *admissionsV1Beta1.ValidatingWebhookConfiguration:
		return (*validatingWebhookConfigurationV1Beta1)(value), true
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

type validatingWebhookConfigurationV1 admissionsV1.ValidatingWebhookConfiguration

func (webhookConfig *validatingWebhookConfigurationV1) GetWebhooks() []ValidatingWebhookAdapter {
	result := make([]ValidatingWebhookAdapter, 0, len(webhookConfig.Webhooks))
	for i := range webhookConfig.Webhooks {
		var webhookAdapter ValidatingWebhookAdapter = (*validatingWebhookV1)(&webhookConfig.Webhooks[i])
		result = append(result, webhookAdapter)
	}
	return result
}

func (webhookConfig *validatingWebhookConfigurationV1) SetWebhooks(webhooks []ValidatingWebhookAdapter) {
	convertedWebhooks := make([]admissionsV1.ValidatingWebhook, 0, len(webhooks))
	for _, webhookAdapter := range webhooks {
		convertedToV1Adapter, ok := webhookAdapter.(*validatingWebhookV1)
		if !ok {
			continue
		}
		var validatingWebhook = admissionsV1.ValidatingWebhook(*convertedToV1Adapter)
		convertedWebhooks = append(convertedWebhooks, validatingWebhook)
	}
	webhookConfig.Webhooks = convertedWebhooks
}

func (webhookConfig *validatingWebhookConfigurationV1) SetLabels(labels map[string]string) {
	webhookConfig.Labels = labels
}

type validatingWebhookConfigurationV1Beta1 admissionsV1Beta1.ValidatingWebhookConfiguration

func (webhookConfig *validatingWebhookConfigurationV1Beta1) GetWebhooks() []ValidatingWebhookAdapter {
	result := make([]ValidatingWebhookAdapter, 0, len(webhookConfig.Webhooks))
	for i := range webhookConfig.Webhooks {
		var webhookAdapter ValidatingWebhookAdapter = (*validatingWebhookV1Beta1)(&webhookConfig.Webhooks[i])
		result = append(result, webhookAdapter)
	}
	return result
}

func (webhookConfig *validatingWebhookConfigurationV1Beta1) SetWebhooks(webhooks []ValidatingWebhookAdapter) {
	convertedWebhooks := make([]admissionsV1Beta1.ValidatingWebhook, 0, len(webhooks))
	for _, webhookAdapter := range webhooks {
		convertedToV1Adapter, ok := webhookAdapter.(*validatingWebhookV1Beta1)
		if !ok {
			continue
		}
		var validatingWebhook = admissionsV1Beta1.ValidatingWebhook(*convertedToV1Adapter)
		convertedWebhooks = append(convertedWebhooks, validatingWebhook)
	}
	webhookConfig.Webhooks = convertedWebhooks
}

func (webhookConfig *validatingWebhookConfigurationV1Beta1) SetLabels(labels map[string]string) {
	webhookConfig.Labels = labels
}

type validatingWebhookV1 admissionsV1.ValidatingWebhook

func (w *validatingWebhookV1) SetCABundle(bundle []byte) {
	w.ClientConfig.CABundle = bundle
}

func (w *validatingWebhookV1) SetServiceNamespace(namespace string) {
	w.InitializeServiceReference()
	w.ClientConfig.Service.Namespace = namespace
}

func (w *validatingWebhookV1) SetServiceName(name string) {
	w.InitializeServiceReference()
	w.ClientConfig.Service.Name = name
}

func (w *validatingWebhookV1) SetServicePath(path *string) {
	w.InitializeServiceReference()
	w.ClientConfig.Service.Path = path
}

func (w *validatingWebhookV1) GetName() string { return w.Name }

func (w *validatingWebhookV1) SetName(name string) { w.Name = name }

func (w *validatingWebhookV1) SetAdmissionReviewVersions(versions []string) {
	w.AdmissionReviewVersions = versions
}

func (w *validatingWebhookV1) SetFailurePolicy(policy string) {
	temp := admissionsV1.FailurePolicyType(policy)
	w.FailurePolicy = &temp
}

func (w *validatingWebhookV1) SetMatchPolicy(policy string) {
	temp := admissionsV1.MatchPolicyType(policy)
	w.MatchPolicy = &temp
}

func (w *validatingWebhookV1) SetSideEffects(sideEffectsClass string) {
	temp := admissionsV1.SideEffectClass(sideEffectsClass)
	w.SideEffects = &temp
}

func (w *validatingWebhookV1) GetNamespaceSelector() *metav1.LabelSelector {
	return w.NamespaceSelector
}

func (w *validatingWebhookV1) SetNamespaceSelector(selector *metav1.LabelSelector) {
	w.NamespaceSelector = selector
}

func (w *validatingWebhookV1) SetTimeoutSeconds(timeoutSeconds int32) {
	w.TimeoutSeconds = &timeoutSeconds
}

func (w *validatingWebhookV1) SetAdmissionRules(rules []AdmissionRuleAdapter) {
	newRules := make([]admissionsV1.RuleWithOperations, 0, len(rules))
	for _, r := range rules {
		stringOperations := make([]admissionsV1.OperationType, 0, len(r.Operations))
		for _, op := range r.Operations {
			stringOperations = append(stringOperations, admissionsV1.OperationType(op))
		}
		newRules = append(newRules, admissionsV1.RuleWithOperations{
			Operations: stringOperations,
			Rule: admissionsV1.Rule{
				APIGroups:   r.APIGroups,
				APIVersions: r.APIVersions,
				Resources:   r.Resources,
				Scope:       (*admissionsV1.ScopeType)(r.Scope),
			},
		})
	}
	w.Rules = newRules
}

func (w *validatingWebhookV1) InitializeServiceReference() {
	if w.ClientConfig.Service == nil {
		w.ClientConfig.Service = &admissionsV1.ServiceReference{}
	}
}

type validatingWebhookV1Beta1 admissionsV1Beta1.ValidatingWebhook

func (w *validatingWebhookV1Beta1) GetName() string { return w.Name }

func (w *validatingWebhookV1Beta1) SetName(name string) {
	w.Name = name
}

func (w *validatingWebhookV1Beta1) SetAdmissionReviewVersions(versions []string) {
	w.AdmissionReviewVersions = versions
}

func (w *validatingWebhookV1Beta1) SetFailurePolicy(policy string) {
	temp := admissionsV1Beta1.FailurePolicyType(policy)
	w.FailurePolicy = &temp
}

func (w *validatingWebhookV1Beta1) SetMatchPolicy(policy string) {
	temp := admissionsV1Beta1.MatchPolicyType(policy)
	w.MatchPolicy = &temp
}

func (w *validatingWebhookV1Beta1) SetSideEffects(sideEffectsClass string) {
	temp := admissionsV1Beta1.SideEffectClass(sideEffectsClass)
	w.SideEffects = &temp
}

func (w *validatingWebhookV1Beta1) GetNamespaceSelector() *metav1.LabelSelector {
	return w.NamespaceSelector
}

func (w *validatingWebhookV1Beta1) SetNamespaceSelector(selector *metav1.LabelSelector) {
	w.NamespaceSelector = selector
}

func (w *validatingWebhookV1Beta1) SetTimeoutSeconds(timeoutSeconds int32) {
	w.TimeoutSeconds = &timeoutSeconds
}

func (w *validatingWebhookV1Beta1) SetCABundle(bundle []byte) {
	w.ClientConfig.CABundle = bundle
}

func (w *validatingWebhookV1Beta1) SetServiceNamespace(namespace string) {
	w.InitializeServiceReference()
	w.ClientConfig.Service.Namespace = namespace
}

func (w *validatingWebhookV1Beta1) SetServiceName(name string) {
	w.InitializeServiceReference()
	w.ClientConfig.Service.Name = name
}

func (w *validatingWebhookV1Beta1) SetServicePath(path *string) {
	w.InitializeServiceReference()
	w.ClientConfig.Service.Path = path
}

func (w *validatingWebhookV1Beta1) SetAdmissionRules(rules []AdmissionRuleAdapter) {
	newRules := make([]admissionsV1Beta1.RuleWithOperations, 0, len(rules))
	for _, r := range rules {
		stringOperations := make([]admissionsV1Beta1.OperationType, 0, len(r.Operations))
		for _, op := range r.Operations {
			stringOperations = append(stringOperations, admissionsV1Beta1.OperationType(op))
		}
		newRules = append(newRules, admissionsV1Beta1.RuleWithOperations{
			Operations: stringOperations,
			Rule: admissionsV1Beta1.Rule{
				APIGroups:   r.APIGroups,
				APIVersions: r.APIVersions,
				Resources:   r.Resources,
				Scope:       (*admissionsV1Beta1.ScopeType)(r.Scope),
			},
		})
	}
	w.Rules = newRules
}
func (w *validatingWebhookV1Beta1) InitializeServiceReference() {
	if w.ClientConfig.Service == nil {
		w.ClientConfig.Service = &admissionsV1Beta1.ServiceReference{}
	}
}
