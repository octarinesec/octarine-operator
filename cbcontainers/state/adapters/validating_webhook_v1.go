package adapters

import (
	"fmt"
	admissionsV1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type validatingWebhookConfigurationV1 admissionsV1.ValidatingWebhookConfiguration

func (webhookConfig *validatingWebhookConfigurationV1) GetWebhooks() []WebhookAdapter {
	result := make([]WebhookAdapter, 0, len(webhookConfig.Webhooks))
	for i := range webhookConfig.Webhooks {
		var webhookAdapter WebhookAdapter = (*validatingWebhookV1)(&webhookConfig.Webhooks[i])
		result = append(result, webhookAdapter)
	}
	return result
}

func (webhookConfig *validatingWebhookConfigurationV1) SetWebhooks(webhooks []WebhookAdapter) ([]WebhookAdapter, error) {
	convertedWebhooks := make([]admissionsV1.ValidatingWebhook, 0, len(webhooks))
	for _, webhookAdapter := range webhooks {
		convertedToV1Adapter, ok := webhookAdapter.(*validatingWebhookV1)
		if !ok {
			return nil, fmt.Errorf("this is an adapter for v1 but got a non-v1 webhook: %v", webhookAdapter)
		}
		var validatingWebhook = admissionsV1.ValidatingWebhook(*convertedToV1Adapter)
		convertedWebhooks = append(convertedWebhooks, validatingWebhook)
	}
	webhookConfig.Webhooks = convertedWebhooks
	// Return adapters for the new webhooks to enable direct modifications via the adapter methods
	return webhookConfig.GetWebhooks(), nil
}

func (webhookConfig *validatingWebhookConfigurationV1) SetLabels(labels map[string]string) {
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

func (w *validatingWebhookV1) GetAdmissionRules() []AdmissionRuleAdapter {
	result := make([]AdmissionRuleAdapter, 0, len(w.Rules))
	for i := range w.Rules {
		var ruleAdapter = AdmissionRuleAdapter{
			Operations:  make([]string, 0, len(w.Rules[i].Operations)),
			APIGroups:   w.Rules[i].APIGroups,
			APIVersions: w.Rules[i].APIVersions,
			Resources:   w.Rules[i].Resources,
			Scope:       (*string)(w.Rules[i].Scope),
		}
		for _, op := range w.Rules[i].Operations {
			ruleAdapter.Operations = append(ruleAdapter.Operations, string(op))
		}
		result = append(result, ruleAdapter)
	}
	return result
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
