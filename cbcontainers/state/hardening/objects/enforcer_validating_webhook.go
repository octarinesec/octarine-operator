package objects

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/hardening/adapters"
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
	WebhookFailurePolicy = adapters.FailurePolicyIgnore
	WebhookPath          = "/validate"
	// This value's default changes across versions so we want to ensure consistency by setting it explicitly
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

func (obj *EnforcerValidatingWebhookK8sObject) HardeningChildNamespacedName(_ *cbcontainersv1.CBContainersHardening) types.NamespacedName {
	return types.NamespacedName{Name: EnforcerName, Namespace: ""}
}

func (obj *EnforcerValidatingWebhookK8sObject) MutateHardeningChildK8sObject(k8sObject client.Object, cbContainersHardening *cbcontainersv1.CBContainersHardening) error {
	webhookConfiguration, ok := adapters.TryGetValidatingWebhookConfigurationAdapter(k8sObject)
	if !ok {
		return fmt.Errorf("expected a valid instance of ValidatingWebhookConfiguration")
	}

	if obj.tlsSecretValues == nil {
		return fmt.Errorf("tls secret values weren't provided")
	}

	enforcerSpec := cbContainersHardening.Spec.EnforcerSpec

	webhookConfiguration.SetLabels(enforcerSpec.Labels)
	return obj.mutateWebhooks(webhookConfiguration, cbContainersHardening)
}

func (obj *EnforcerValidatingWebhookK8sObject) mutateWebhooks(webhookConfiguration adapters.WebhookConfigurationAdapter, cbContainersHardening *cbcontainersv1.CBContainersHardening) error {
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

	obj.mutateResourcesWebhook(resourcesWebhookObj, cbContainersHardening.Spec.EnforcerSpec.WebhookTimeoutSeconds)
	obj.mutateNamespacesWebhook(namespacesWebhookObj, cbContainersHardening.Spec.EnforcerSpec.WebhookTimeoutSeconds)
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

func (obj *EnforcerValidatingWebhookK8sObject) mutateResourcesWebhook(resourcesWebhook adapters.WebhookAdapter, timeoutSeconds int32) {
	resourcesWebhook.SetName(ValidatingResourcesWebhookName)
	resourcesWebhook.SetAdmissionReviewVersions([]string{"v1beta1"})
	resourcesWebhook.SetFailurePolicy(WebhookFailurePolicy)
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
		Key:      "name",
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
		"ingresses",
		"customresourcedefinitions",
	}
}

func (obj *EnforcerValidatingWebhookK8sObject) mutateNamespacesWebhook(namespacesWebhook adapters.WebhookAdapter, timeoutSeconds int32) {
	namespacesWebhook.SetName(ValidatingNamespacesWebhookName)
	namespacesWebhook.SetAdmissionReviewVersions([]string{"v1beta1"})
	namespacesWebhook.SetFailurePolicy(WebhookFailurePolicy)
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
	rules := webhook.GetAdmissionRules()
	if rules == nil || len(rules) != 1 {
		rules = make([]adapters.AdmissionRuleAdapter, 1)
	}

	rules[0].Operations = []string{adapters.OperationCreate, adapters.OperationUpdate, adapters.OperationConnect}
	rules[0].APIVersions = []string{"*"}
	rules[0].APIGroups = []string{"*"}
	rules[0].Resources = []string{"namespaces"}
	webhook.SetAdmissionRules(rules)
}
