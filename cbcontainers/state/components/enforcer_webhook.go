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
	ResourcesWebhookName  = "resources.validating-webhook.cbcontainers"
	NamespacesWebhookName = "namespaces.validating-webhook.cbcontainers"
)

var (
	WebhookFailurePolicy = adapters.FailurePolicyIgnore
	WebhookPath          = "/validate"
	// WebhookMatchPolicy : This value's default changes across versions so we want to ensure consistency by setting it explicitly
	WebhookMatchPolicy = adapters.MatchPolicyEquivalent

	ResourcesWebhookSideEffect  = adapters.SideEffectClassNoneOnDryRun
	NamespacesWebhookSideEffect = adapters.SideEffectsClassNone
)

type EnforcerWebhookK8sObject struct {
	tlsSecretValues *models.TlsSecretValues
	kubeletVersion  string
}

func NewEnforcerWebhookK8sObject(kubeletVersion string) *EnforcerWebhookK8sObject {
	return &EnforcerWebhookK8sObject{
		kubeletVersion: kubeletVersion,
	}
}

func (obj *EnforcerWebhookK8sObject) UpdateTlsSecretValues(tlsSecretValues models.TlsSecretValues) {
	obj.tlsSecretValues = &tlsSecretValues
}

func (obj *EnforcerWebhookK8sObject) EmptyK8sObject() client.Object {
	return adapters.EmptyValidatingWebhookConfigForVersion(obj.kubeletVersion)
}

func (obj *EnforcerWebhookK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: EnforcerName, Namespace: ""}
}

func (obj *EnforcerWebhookK8sObject) MutateK8sObject(k8sObject client.Object, agentSpec *cbcontainersv1.CBContainersAgentSpec) error {
	webhookConfiguration, ok := adapters.TryGetValidatingWebhookConfigurationAdapter(k8sObject)
	if !ok {
		return fmt.Errorf("expected a valid instance of ValidatingWebhookConfiguration")
	}

	if obj.tlsSecretValues == nil {
		return fmt.Errorf("tls secret values weren't provided")
	}

	hardeningSpec := &agentSpec.HardeningSpec
	enforcerSpec := &hardeningSpec.EnforcerSpec

	webhookConfiguration.SetLabels(enforcerSpec.Labels)
	return obj.mutateWebhooks(webhookConfiguration, hardeningSpec)
}

func (obj *EnforcerWebhookK8sObject) mutateWebhooks(webhookConfiguration adapters.ValidatingWebhookConfigurationAdapter, cbContainersHardeningSpec *cbcontainersv1.CBContainersHardeningSpec) error {
	var resourcesWebhookObj adapters.ValidatingWebhookAdapter
	var namespacesWebhookObj adapters.ValidatingWebhookAdapter

	initializeWebhooks := false
	webhooks := webhookConfiguration.GetWebhooks()
	if webhooks == nil || len(webhooks) != 2 {
		initializeWebhooks = true
	} else {
		resourcesWebhook, resourcesWebhookFound := obj.findWebhookByName(webhooks, ResourcesWebhookName)
		resourcesWebhookObj = resourcesWebhook
		namespacesWebhook, namespacesWebhookFound := obj.findWebhookByName(webhooks, NamespacesWebhookName)
		namespacesWebhookObj = namespacesWebhook
		initializeWebhooks = !resourcesWebhookFound || !namespacesWebhookFound
	}

	if initializeWebhooks {
		webhooks := []adapters.ValidatingWebhookAdapter{
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

	obj.mutateResourcesWebhook(resourcesWebhookObj, cbContainersHardeningSpec.EnforcerSpec.WebhookTimeoutSeconds)
	obj.mutateNamespacesWebhook(namespacesWebhookObj, cbContainersHardeningSpec.EnforcerSpec.WebhookTimeoutSeconds)
	return nil
}

func (obj *EnforcerWebhookK8sObject) findWebhookByName(webhooks []adapters.ValidatingWebhookAdapter, name string) (adapters.ValidatingWebhookAdapter, bool) {
	for idx, webhook := range webhooks {
		if webhook.GetName() == name {
			return webhooks[idx], true
		}
	}

	return nil, false
}

func (obj *EnforcerWebhookK8sObject) mutateResourcesWebhook(resourcesWebhook adapters.ValidatingWebhookAdapter, timeoutSeconds int32) {
	resourcesWebhook.SetName(ResourcesWebhookName)
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

func (obj *EnforcerWebhookK8sObject) getResourcesNamespaceSelector(selector *metav1.LabelSelector) *metav1.LabelSelector {
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

func (obj *EnforcerWebhookK8sObject) mutateResourcesWebhooksRules(webhook adapters.ValidatingWebhookAdapter) {
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

func (obj *EnforcerWebhookK8sObject) getResourcesList() []string {
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

func (obj *EnforcerWebhookK8sObject) mutateNamespacesWebhook(namespacesWebhook adapters.ValidatingWebhookAdapter, timeoutSeconds int32) {
	namespacesWebhook.SetName(NamespacesWebhookName)
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

func (obj *EnforcerWebhookK8sObject) mutateNamespacesWebhooksRules(webhook adapters.ValidatingWebhookAdapter) {
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
