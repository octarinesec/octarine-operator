package components

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/adapters"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	"github.com/vmware/cbcontainers-operator/cbcontainers/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ValidatingResourcesWebhookName  = "resources.validating-webhook.cbcontainers"
	ValidatingNamespacesWebhookName = "namespaces.validating-webhook.cbcontainers"
)

var (
	WebhookPath = "/validate"
	// WebhookMatchPolicy : This value's default changes across versions so we want to ensure consistency by setting it explicitly
	WebhookMatchPolicy = adapters.MatchPolicyEquivalent

	ResourcesWebhookSideEffect  = adapters.SideEffectClassNoneOnDryRun
	NamespacesWebhookSideEffect = adapters.SideEffectsClassNone
)

type EnforcerValidatingWebhookK8sObject struct {
	tlsSecretValues *models.TlsSecretValues
	kubeletVersion  string
}

func NewEnforcerValidatingWebhookK8sObject(kubeletVersion string) *EnforcerValidatingWebhookK8sObject {
	return &EnforcerValidatingWebhookK8sObject{
		kubeletVersion: kubeletVersion,
	}
}

func (obj *EnforcerValidatingWebhookK8sObject) UpdateTlsSecretValues(tlsSecretValues models.TlsSecretValues) {
	obj.tlsSecretValues = &tlsSecretValues
}

func (obj *EnforcerValidatingWebhookK8sObject) EmptyK8sObject() client.Object {
	return adapters.EmptyValidatingWebhookConfigForVersion(obj.kubeletVersion)
}

func (obj *EnforcerValidatingWebhookK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: EnforcerName, Namespace: ""}
}

func (obj *EnforcerValidatingWebhookK8sObject) MutateK8sObject(k8sObject client.Object, agentSpec *cbcontainersv1.CBContainersAgentSpec) error {
	webhookConfiguration, ok := adapters.TryGetValidatingWebhookConfigurationAdapter(k8sObject)
	if !ok {
		return fmt.Errorf("expected a valid instance of ValidatingWebhookConfiguration")
	}

	if obj.tlsSecretValues == nil {
		return fmt.Errorf("tls secret values weren't provided")
	}

	enforcer := &agentSpec.Components.Basic.Enforcer
	obj.mutateWebhookConfigurationLabels(webhookConfiguration, enforcer)
	return obj.mutateWebhooks(webhookConfiguration, enforcer)
}

func (obj *EnforcerValidatingWebhookK8sObject) mutateWebhooks(webhookConfiguration adapters.WebhookConfigurationAdapter, enforcer *cbcontainersv1.CBContainersEnforcerSpec) error {
	var resourcesWebhookObj adapters.WebhookAdapter
	var namespacesWebhookObj adapters.WebhookAdapter

	initializeWebhooks := false
	webhooks := webhookConfiguration.GetWebhooks()
	if webhooks == nil || len(webhooks) != 2 {
		initializeWebhooks = true
	} else {
		resourcesWebhook, resourcesWebhookFound := obj.findWebhookByName(webhooks, ValidatingResourcesWebhookName)
		resourcesWebhookObj = resourcesWebhook
		namespacesWebhook, namespacesWebhookFound := obj.findWebhookByName(webhooks, ValidatingNamespacesWebhookName)
		namespacesWebhookObj = namespacesWebhook
		initializeWebhooks = !resourcesWebhookFound || !namespacesWebhookFound
	}

	if initializeWebhooks {
		webhooks := []adapters.WebhookAdapter{
			adapters.EmptyValidatingWebhookAdapterForVersion(obj.kubeletVersion),
			adapters.EmptyValidatingWebhookAdapterForVersion(obj.kubeletVersion),
		}
		updatedWebhooks, err := webhookConfiguration.SetWebhooks(webhooks)
		if err != nil {
			return err
		}

		resourcesWebhookObj = updatedWebhooks[0]
		namespacesWebhookObj = updatedWebhooks[1]
	}

	obj.mutateResourcesWebhook(resourcesWebhookObj, enforcer.WebhookTimeoutSeconds, enforcer.FailurePolicy)
	obj.mutateNamespacesWebhook(namespacesWebhookObj, enforcer.WebhookTimeoutSeconds, enforcer.FailurePolicy)
	return nil
}

func (obj *EnforcerValidatingWebhookK8sObject) findWebhookByName(webhooks []adapters.WebhookAdapter, name string) (adapters.WebhookAdapter, bool) {
	for idx, webhook := range webhooks {
		if webhook.GetName() == name {
			return webhooks[idx], true
		}
	}

	return nil, false
}

func (obj *EnforcerValidatingWebhookK8sObject) mutateResourcesWebhook(resourcesWebhook adapters.WebhookAdapter, timeoutSeconds int32, failurePolicy string) {
	resourcesWebhook.SetName(ValidatingResourcesWebhookName)
	resourcesWebhook.SetAdmissionReviewVersions([]string{"v1beta1"})
	resourcesWebhook.SetFailurePolicy(failurePolicy)
	resourcesWebhook.SetSideEffects(ResourcesWebhookSideEffect)
	resourcesWebhook.SetMatchPolicy(WebhookMatchPolicy)
	namespaceSelector := obj.getResourcesNamespaceSelector(resourcesWebhook.GetNamespaceSelector())
	resourcesWebhook.SetNamespaceSelector(namespaceSelector)
	obj.mutateResourcesWebhooksRules(resourcesWebhook)
	if obj.kubeletVersion == "" || obj.kubeletVersion >= "v1.14" {
		resourcesWebhook.SetTimeoutSeconds(timeoutSeconds)
	}
	resourcesWebhook.SetCABundle(obj.tlsSecretValues.CaCert)
	resourcesWebhook.SetServiceName(EnforcerName)
	resourcesWebhook.SetServiceNamespace(commonState.DataPlaneNamespaceName)
	resourcesWebhook.SetServicePath(&WebhookPath)
}

func (obj *EnforcerValidatingWebhookK8sObject) getResourcesNamespaceSelector(selector *metav1.LabelSelector) *metav1.LabelSelector {
	octarineIgnore := metav1.LabelSelectorRequirement{
		Key:      "octarine",
		Operator: metav1.LabelSelectorOpNotIn,
		Values:   []string{"ignore"},
	}

	cbContainersNamespace := metav1.LabelSelectorRequirement{
		// See https://kubernetes.io/docs/reference/labels-annotations-taints/#kubernetes-io-metadata-name
		// This is the label that always matches the namespace name
		// We can't filter directly by namespace otherwise
		Key:      "kubernetes.io/metadata.name",
		Operator: metav1.LabelSelectorOpNotIn,
		Values:   []string{commonState.DataPlaneNamespaceName},
	}

	initializeLabelSelector := false
	if selector == nil || selector.MatchExpressions == nil || len(selector.MatchExpressions) != 2 {
		initializeLabelSelector = true
	} else {
		octarineIgnoreFound := false
		cbContainersNamespaceFound := false
		for _, requirement := range selector.MatchExpressions {
			if reflect.DeepEqual(requirement, octarineIgnore) {
				octarineIgnoreFound = true
			}
			if reflect.DeepEqual(requirement, cbContainersNamespace) {
				cbContainersNamespaceFound = true
			}
		}
		initializeLabelSelector = !octarineIgnoreFound || !cbContainersNamespaceFound
	}

	if initializeLabelSelector {
		return &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{octarineIgnore, cbContainersNamespace},
		}
	}

	return selector

}

func (obj *EnforcerValidatingWebhookK8sObject) mutateResourcesWebhooksRules(webhook adapters.WebhookAdapter) {
	rules := webhook.GetAdmissionRules()
	if rules == nil || len(rules) != 1 {
		rules = make([]adapters.AdmissionRuleAdapter, 1)
	}

	rules[0].Operations = []string{adapters.OperationAll}
	rules[0].APIVersions = []string{"*"}
	rules[0].APIGroups = []string{"*"}

	expectedResourcesList := obj.getResourcesList()
	if !utils.StringsSlicesHaveSameItems(rules[0].Resources, expectedResourcesList) {
		rules[0].Resources = expectedResourcesList
	}
	webhook.SetAdmissionRules(rules)
}

func (obj *EnforcerValidatingWebhookK8sObject) getResourcesList() []string {
	return []string{
		"pods/portforward",
		"pods/exec",
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
		"ingresses",
		"customresourcedefinitions",
	}
}

func (obj *EnforcerValidatingWebhookK8sObject) mutateNamespacesWebhook(namespacesWebhook adapters.WebhookAdapter, timeoutSeconds int32, failurePolicy string) {
	namespacesWebhook.SetName(ValidatingNamespacesWebhookName)
	namespacesWebhook.SetAdmissionReviewVersions([]string{"v1beta1"})
	namespacesWebhook.SetFailurePolicy(failurePolicy)
	namespacesWebhook.SetMatchPolicy(WebhookMatchPolicy)
	namespacesWebhook.SetSideEffects(NamespacesWebhookSideEffect)
	namespacesWebhook.SetNamespaceSelector(&metav1.LabelSelector{})
	if obj.kubeletVersion == "" || obj.kubeletVersion >= "v1.14" {
		namespacesWebhook.SetTimeoutSeconds(timeoutSeconds)
	}

	namespacesWebhook.SetCABundle(obj.tlsSecretValues.CaCert)
	namespacesWebhook.SetServiceNamespace(commonState.DataPlaneNamespaceName)
	namespacesWebhook.SetServiceName(EnforcerName)
	namespacesWebhook.SetServicePath(&WebhookPath)

	obj.mutateNamespacesWebhooksRules(namespacesWebhook)

}

func (obj *EnforcerValidatingWebhookK8sObject) mutateNamespacesWebhooksRules(webhook adapters.WebhookAdapter) {
	// This webhook exists to prevent someone from adding the "octarine:ignore" label on a namespace (except ours)
	// Which would essentially disable our product there
	// So we have a separate webhook call for all namespaces and the agent will forbid the request if the label is added
	rules := webhook.GetAdmissionRules()
	if rules == nil || len(rules) != 1 {
		rules = make([]adapters.AdmissionRuleAdapter, 1)
	}

	rules[0].Operations = []string{adapters.OperationAll}
	rules[0].APIVersions = []string{"*"}
	rules[0].APIGroups = []string{"*"}
	rules[0].Resources = []string{"namespaces"}
	webhook.SetAdmissionRules(rules)
}

func (obj *EnforcerValidatingWebhookK8sObject) mutateWebhookConfigurationLabels(webhook adapters.WebhookConfigurationAdapter, enforcerSpec *cbcontainersv1.CBContainersEnforcerSpec) {
	labels := enforcerSpec.Labels

	// AKS-specific label
	// AKS has a built-in Admission controller (Admission Enforcer) that tries to prevent affecting system namespaces (e.g. kube-system)
	// In practice, it modifies all validating/mutating webhooks and adds a matcher that ignores any namespaces with the `control-plane` label
	// AKS itself maintains which namespaces have the `control-plane` label but other products might use it as well by chance (e.g. kubeflow)
	// This would prevent us from doing our job as security product and open up system namespaces for attacks since we "silently" ignore anything in them without knowing it
	// Therefore we ask that our webhooks are excluded from this feature and can monitor all namespaces
	// References:
	//	https://docs.microsoft.com/en-us/azure/aks/faq#can-admission-controller-webhooks-impact-kube-system-and-internal-aks-namespaces
	//	https://github.com/Azure/AKS/issues/1771
	labels["admissions.enforcer/disabled"] = "true"

	webhook.SetLabels(labels)
}
