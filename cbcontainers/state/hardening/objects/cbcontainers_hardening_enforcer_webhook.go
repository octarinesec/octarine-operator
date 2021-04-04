package objects

import (
	"fmt"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/cbcontainers/utils"
	admissionsV1beta1 "k8s.io/api/admissionregistration/v1beta1"
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
	WebhookFailurePolicy = admissionsV1beta1.Ignore
	WebhookPath          = "/validate"

	ResourcesWebhookSideEffect  = admissionsV1beta1.SideEffectClassNoneOnDryRun
	NamespacesWebhookSideEffect = admissionsV1beta1.SideEffectClassNone
)

type EnforcerWebhookK8sObject struct {
	TlsSecretValues models.TlsSecretValues
}

func NewEnforcerWebhookK8sObject() *EnforcerWebhookK8sObject { return &EnforcerWebhookK8sObject{} }

func (obj *EnforcerWebhookK8sObject) EmptyK8sObject() client.Object {
	return &admissionsV1beta1.ValidatingWebhookConfiguration{}
}

func (obj *EnforcerWebhookK8sObject) HardeningChildNamespacedName(_ *cbcontainersv1.CBContainersHardening) types.NamespacedName {
	return types.NamespacedName{Name: EnforcerName, Namespace: ""}
}

func (obj *EnforcerWebhookK8sObject) MutateHardeningChildK8sObject(k8sObject client.Object, cbContainersHardening *cbcontainersv1.CBContainersHardening) error {
	webhookConfiguration, ok := k8sObject.(*admissionsV1beta1.ValidatingWebhookConfiguration)
	if !ok {
		return fmt.Errorf("expected Service K8s object")
	}
	enforcerSpec := cbContainersHardening.Spec.EnforcerSpec

	webhookConfiguration.Labels = enforcerSpec.DeploymentLabels
	obj.mutateWebhooks(webhookConfiguration, cbContainersHardening)

	return nil
}

func (obj *EnforcerWebhookK8sObject) mutateWebhooks(webhookConfiguration *admissionsV1beta1.ValidatingWebhookConfiguration, cbContainersHardening *cbcontainersv1.CBContainersHardening) {
	var resourcesWebhookObj *admissionsV1beta1.ValidatingWebhook
	var namespacesWebhookObj *admissionsV1beta1.ValidatingWebhook
	initializeWebhooks := false

	if webhookConfiguration.Webhooks == nil || len(webhookConfiguration.Webhooks) != 2 {
		initializeWebhooks = true
	} else {
		resourcesWebhook, resourcesWebhookFound := obj.findWebhookByName(webhookConfiguration.Webhooks, ResourcesWebhookName)
		resourcesWebhookObj = resourcesWebhook
		namespacesWebhook, namespacesWebhookFound := obj.findWebhookByName(webhookConfiguration.Webhooks, NamespacesWebhookName)
		namespacesWebhookObj = namespacesWebhook
		initializeWebhooks = !resourcesWebhookFound || !namespacesWebhookFound
	}

	if initializeWebhooks {
		webhookConfiguration.Webhooks = make([]admissionsV1beta1.ValidatingWebhook, 2)
		resourcesWebhookObj = &webhookConfiguration.Webhooks[0]
		namespacesWebhookObj = &webhookConfiguration.Webhooks[1]
	}

	obj.mutateResourcesWebhook(resourcesWebhookObj, cbContainersHardening.Namespace, cbContainersHardening.Spec.EnforcerSpec.WebhookTimeoutSeconds)
	obj.mutateNamespacesWebhook(namespacesWebhookObj, cbContainersHardening.Namespace, cbContainersHardening.Spec.EnforcerSpec.WebhookTimeoutSeconds)
}

func (obj *EnforcerWebhookK8sObject) findWebhookByName(webhooks []admissionsV1beta1.ValidatingWebhook, name string) (*admissionsV1beta1.ValidatingWebhook, bool) {
	for idx, webhook := range webhooks {
		if webhook.Name == name {
			return &webhooks[idx], true
		}
	}

	return nil, false
}

func (obj *EnforcerWebhookK8sObject) mutateResourcesWebhook(resourcesWebhook *admissionsV1beta1.ValidatingWebhook, namespace string, timeoutSeconds int32) {
	resourcesWebhook.Name = ResourcesWebhookName
	resourcesWebhook.FailurePolicy = &WebhookFailurePolicy
	resourcesWebhook.SideEffects = &ResourcesWebhookSideEffect
	resourcesWebhook.NamespaceSelector = obj.getResourcesNamespaceSelector(resourcesWebhook.NamespaceSelector, namespace)
	obj.mutateResourcesWebhooksRules(resourcesWebhook)
	resourcesWebhook.TimeoutSeconds = &timeoutSeconds
	resourcesWebhook.ClientConfig = admissionsV1beta1.WebhookClientConfig{
		Service: &admissionsV1beta1.ServiceReference{
			Namespace: namespace,
			Name:      EnforcerName,
			Path:      &WebhookPath,
		},
		CABundle: obj.TlsSecretValues.CaCert,
	}
}

func (obj *EnforcerWebhookK8sObject) getResourcesNamespaceSelector(selector *metav1.LabelSelector, namespace string) *metav1.LabelSelector {
	octarineIgnore := metav1.LabelSelectorRequirement{
		Key:      "octarine",
		Operator: metav1.LabelSelectorOpNotIn,
		Values:   []string{"ignore"},
	}

	cbContainersNamespace := metav1.LabelSelectorRequirement{
		Key:      "name",
		Operator: metav1.LabelSelectorOpNotIn,
		Values:   []string{namespace},
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

func (obj *EnforcerWebhookK8sObject) mutateResourcesWebhooksRules(webhook *admissionsV1beta1.ValidatingWebhook) {
	if webhook.Rules == nil || len(webhook.Rules) != 1 {
		webhook.Rules = make([]admissionsV1beta1.RuleWithOperations, 1)
	}

	webhook.Rules[0].Operations = []admissionsV1beta1.OperationType{admissionsV1beta1.OperationAll}
	webhook.Rules[0].Rule.APIGroups = []string{"*"}
	webhook.Rules[0].Rule.APIVersions = []string{"*"}

	expectedResourcesList := obj.getResourcesList()
	if !utils.StringsSlicesHaveSameItems(webhook.Rules[0].Rule.Resources, expectedResourcesList) {
		webhook.Rules[0].Rule.Resources = expectedResourcesList
	}
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

func (obj *EnforcerWebhookK8sObject) mutateNamespacesWebhook(namespacesWebhook *admissionsV1beta1.ValidatingWebhook, namespace string, timeoutSeconds int32) {
	namespacesWebhook.Name = NamespacesWebhookName
	namespacesWebhook.FailurePolicy = &WebhookFailurePolicy
	namespacesWebhook.SideEffects = &NamespacesWebhookSideEffect
	namespacesWebhook.NamespaceSelector = nil
	obj.mutateNamespacesWebhooksRules(namespacesWebhook)
	namespacesWebhook.TimeoutSeconds = &timeoutSeconds
	namespacesWebhook.ClientConfig = admissionsV1beta1.WebhookClientConfig{
		Service: &admissionsV1beta1.ServiceReference{
			Namespace: namespace,
			Name:      EnforcerName,
			Path:      &WebhookPath,
		},
		CABundle: obj.TlsSecretValues.CaCert,
	}
}

func (obj *EnforcerWebhookK8sObject) mutateNamespacesWebhooksRules(webhook *admissionsV1beta1.ValidatingWebhook) {
	if webhook.Rules == nil || len(webhook.Rules) != 1 {
		webhook.Rules = make([]admissionsV1beta1.RuleWithOperations, 1)
	}

	expectedOperations := []admissionsV1beta1.OperationType{admissionsV1beta1.Create, admissionsV1beta1.Update, admissionsV1beta1.Connect}
	if !utils.StringsSlicesHaveSameItems(obj.operationsToStrings(webhook.Rules[0].Operations), obj.operationsToStrings(expectedOperations)) {
		webhook.Rules[0].Operations = expectedOperations
	}
	webhook.Rules[0].Rule = admissionsV1beta1.Rule{
		APIGroups:   []string{"*"},
		APIVersions: []string{"*"},
		Resources:   []string{"namespaces"},
	}
}

func (obj *EnforcerWebhookK8sObject) operationsToStrings(operations []admissionsV1beta1.OperationType) []string {
	operationsStrings := make([]string, 0, len(operations))
	for _, operation := range operations {
		operationsStrings = append(operationsStrings, string(operation))
	}

	return operationsStrings
}
