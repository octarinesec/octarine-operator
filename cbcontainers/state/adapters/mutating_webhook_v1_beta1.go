package adapters

import (
	"fmt"
	admissionsV1Beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mutatingWebhookConfigurationV1Beta1 admissionsV1Beta1.MutatingWebhookConfiguration

func (webhookConfig *mutatingWebhookConfigurationV1Beta1) GetWebhooks() []WebhookAdapter {
	result := make([]WebhookAdapter, 0, len(webhookConfig.Webhooks))
	for i := range webhookConfig.Webhooks {
		var webhookAdapter WebhookAdapter = (*mutatingWebhookV1Beta1)(&webhookConfig.Webhooks[i])
		result = append(result, webhookAdapter)
	}
	return result
}

func (webhookConfig *mutatingWebhookConfigurationV1Beta1) SetWebhooks(webhooks []WebhookAdapter) ([]WebhookAdapter, error) {
	convertedWebhooks := make([]admissionsV1Beta1.MutatingWebhook, 0, len(webhooks))
	for _, webhookAdapter := range webhooks {
		convertedToV1Adapter, ok := webhookAdapter.(*mutatingWebhookV1Beta1)
		if !ok {
			return nil, fmt.Errorf("this is an adapter for v1beta1 but got a non-v1beta1 webhook: %v", webhookAdapter)
		}
		var mutatingWebhook = admissionsV1Beta1.MutatingWebhook(*convertedToV1Adapter)
		convertedWebhooks = append(convertedWebhooks, mutatingWebhook)
	}
	webhookConfig.Webhooks = convertedWebhooks
	// Return adapters for the new webhooks to enable direct modifications via the adapter methods
	return webhookConfig.GetWebhooks(), nil
}

func (webhookConfig *mutatingWebhookConfigurationV1Beta1) SetLabels(labels map[string]string) {
	webhookConfig.Labels = labels
}

type mutatingWebhookV1Beta1 admissionsV1Beta1.MutatingWebhook

func (w *mutatingWebhookV1Beta1) GetName() string { return w.Name }

func (w *mutatingWebhookV1Beta1) SetName(name string) {
	w.Name = name
}

func (w *mutatingWebhookV1Beta1) SetAdmissionReviewVersions(versions []string) {
	w.AdmissionReviewVersions = versions
}

func (w *mutatingWebhookV1Beta1) SetFailurePolicy(policy string) {
	temp := admissionsV1Beta1.FailurePolicyType(policy)
	w.FailurePolicy = &temp
}

func (w *mutatingWebhookV1Beta1) SetMatchPolicy(policy string) {
	temp := admissionsV1Beta1.MatchPolicyType(policy)
	w.MatchPolicy = &temp
}

func (w *mutatingWebhookV1Beta1) SetSideEffects(sideEffectsClass string) {
	temp := admissionsV1Beta1.SideEffectClass(sideEffectsClass)
	w.SideEffects = &temp
}

func (w *mutatingWebhookV1Beta1) GetNamespaceSelector() *metav1.LabelSelector {
	return w.NamespaceSelector
}

func (w *mutatingWebhookV1Beta1) SetNamespaceSelector(selector *metav1.LabelSelector) {
	w.NamespaceSelector = selector
}

func (w *mutatingWebhookV1Beta1) SetTimeoutSeconds(timeoutSeconds int32) {
	w.TimeoutSeconds = &timeoutSeconds
}

func (w *mutatingWebhookV1Beta1) SetCABundle(bundle []byte) {
	w.ClientConfig.CABundle = bundle
}

func (w *mutatingWebhookV1Beta1) SetServiceNamespace(namespace string) {
	w.InitializeServiceReference()
	w.ClientConfig.Service.Namespace = namespace
}

func (w *mutatingWebhookV1Beta1) SetServiceName(name string) {
	w.InitializeServiceReference()
	w.ClientConfig.Service.Name = name
}

func (w *mutatingWebhookV1Beta1) SetServicePath(path *string) {
	w.InitializeServiceReference()
	w.ClientConfig.Service.Path = path
}

func (w *mutatingWebhookV1Beta1) GetAdmissionRules() []AdmissionRuleAdapter {
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

func (w *mutatingWebhookV1Beta1) SetAdmissionRules(rules []AdmissionRuleAdapter) {
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
func (w *mutatingWebhookV1Beta1) InitializeServiceReference() {
	if w.ClientConfig.Service == nil {
		w.ClientConfig.Service = &admissionsV1Beta1.ServiceReference{}
	}
}
