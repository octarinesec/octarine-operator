package objects

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/hardening/adapters"
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
	// This value's default changes across versions so we want to ensure consistency by setting it explicitly
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

func (obj *EnforcerWebhookK8sObject) HardeningChildNamespacedName(_ *cbcontainersv1.CBContainersHardening) types.NamespacedName {
	return types.NamespacedName{Name: EnforcerName, Namespace: ""}
}

func (obj *EnforcerWebhookK8sObject) MutateHardeningChildK8sObject(k8sObject client.Object, cbContainersHardening *cbcontainersv1.CBContainersHardening) error {
	webhookConfiguration, ok := adapters.TryGetValidatingWebhookConfigurationAdapter(k8sObject)
	if !ok {
		return fmt.Errorf("expected a valid instance of ValidatingWebhookConfiguration")
	}

	if obj.tlsSecretValues == nil {
		return fmt.Errorf("tls secret values weren't provided")
	}

	enforcerSpec := cbContainersHardening.Spec.EnforcerSpec

	webhookConfiguration.SetLabels(enforcerSpec.Labels)
	obj.mutateWebhooks(webhookConfiguration, cbContainersHardening)

	return nil
}

func (obj *EnforcerWebhookK8sObject) mutateWebhooks(webhookConfiguration adapters.ValidatingWebhookConfigurationAdapter, cbContainersHardening *cbcontainersv1.CBContainersHardening) {
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
		resourcesWebhookObj = webhooks[0]
		namespacesWebhookObj = webhooks[1]
		webhookConfiguration.SetWebhooks(webhooks)
	}

	obj.mutateResourcesWebhook(resourcesWebhookObj, cbContainersHardening.Spec.EnforcerSpec.WebhookTimeoutSeconds)
	obj.mutateNamespacesWebhook(namespacesWebhookObj, cbContainersHardening.Spec.EnforcerSpec.WebhookTimeoutSeconds)
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
	rules := []adapters.AdmissionRuleAdapter{
		{
			Operations:  []string{adapters.OperationAll},
			APIGroups:   []string{"*"},
			APIVersions: []string{"*"},
			Resources:   obj.getResourcesList(),
		},
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
	rules := []adapters.AdmissionRuleAdapter{
		{
			Operations:  []string{adapters.OperationCreate, adapters.OperationUpdate, adapters.OperationConnect},
			APIGroups:   []string{"*"},
			APIVersions: []string{"*"},
			Resources:   []string{"namespaces"},
		},
	}
	webhook.SetAdmissionRules(rules)
}
