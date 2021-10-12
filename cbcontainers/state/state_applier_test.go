package state_test

import (
	"context"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/agent_applyment"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/components"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/mocks"
	admissionsV1 "k8s.io/api/admissionregistration/v1"
	admissionsV1Beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsV1 "k8s.io/api/apps/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"

	logrTesting "github.com/go-logr/logr/testing"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	"github.com/vmware/cbcontainers-operator/cbcontainers/test_utils"
	testUtilsMocks "github.com/vmware/cbcontainers-operator/cbcontainers/test_utils/mocks"
	coreV1 "k8s.io/api/core/v1"
	schedulingV1 "k8s.io/api/scheduling/v1"
	schedulingV1alpha1 "k8s.io/api/scheduling/v1alpha1"
	schedulingV1beta1 "k8s.io/api/scheduling/v1beta1"
)

const (
	DefaultKubeletVersion = "v1.20.2"

	NumberOfExpectedAppliedObjects = 12
)

type AppliedK8sObjectsChanger func(K8sObjectDetails, client.Object)

var (
	trueRef bool = true

	Account                    = test_utils.RandomString()
	Cluster                    = test_utils.RandomString()
	ApiGateWayScheme           = test_utils.RandomString()
	ApiGateWayHost             = test_utils.RandomString()
	ApiGateWayPort             = 4206
	ApiGateWayAdapter          = test_utils.RandomString()
	CoreEventsGateWayHost      = test_utils.RandomString()
	HardeningEventsGateWayHost = test_utils.RandomString()
	RuntimeEventsGateWayHost   = test_utils.RandomString()

	EnforcerDeploymentDetails = K8sObjectDetails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       components.EnforcerName,
		ObjectType: reflect.TypeOf(&appsV1.Deployment{}),
	}

	EnforcerValidatingWebhookDetails = K8sObjectDetails{
		Namespace:  "",
		Name:       components.EnforcerName,
		ObjectType: reflect.TypeOf(&admissionsV1.ValidatingWebhookConfiguration{}),
	}

	EnforcerMutatingWebhookDetails = K8sObjectDetails{
		Namespace:  "",
		Name:       components.EnforcerName,
		ObjectType: reflect.TypeOf(&admissionsV1.MutatingWebhookConfiguration{}),
	}

	ResolverDeploymentDetails = K8sObjectDetails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       components.ResolverName,
		ObjectType: reflect.TypeOf(&appsV1.Deployment{}),
	}

	MutateDeploymentsReadyReplicas = func(details K8sObjectDetails, deploymentsToMutate []K8sObjectDetails, object client.Object, readyReplicas int32) {
		foundDeploymentThatNeedsToBeMutated := false
		for _, dep := range deploymentsToMutate {
			if details != dep {
				continue
			}

			foundDeploymentThatNeedsToBeMutated = true
			break
		}

		if !foundDeploymentThatNeedsToBeMutated {
			return
		}

		enforcerDeployment := object.(*appsV1.Deployment)
		enforcerDeployment.Status.ReadyReplicas = readyReplicas
	}

	MutateDeploymentsToBeWithReadyReplica = func(deploymentToMutate ...K8sObjectDetails) AppliedK8sObjectsChanger {
		return func(details K8sObjectDetails, object client.Object) {
			MutateDeploymentsReadyReplicas(details, deploymentToMutate, object, 1)
		}
	}

	MutateDeploymentsToBeWithNoReadyReplica = func(deploymentToMutate ...K8sObjectDetails) AppliedK8sObjectsChanger {
		return func(details K8sObjectDetails, object client.Object) {
			MutateDeploymentsReadyReplicas(details, deploymentToMutate, object, 0)
		}
	}
)

type StateApplierTestMocks struct {
	client              *testUtilsMocks.MockClient
	secretValuesCreator *mocks.MockTlsSecretsValuesCreator
	componentApplier    *mocks.MockAgentComponentApplier
	agentSpec           *cbcontainersv1.CBContainersAgentSpec
	kubeletVersion      string
}

type StateApplierTestSetup func(*StateApplierTestMocks)

type K8sObjectDetails struct {
	Namespace  string
	Name       string
	ObjectType reflect.Type
}

func testStateApplier(t *testing.T, setup StateApplierTestSetup, k8sVersion string) (bool, error) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	agentSpec := &cbcontainersv1.CBContainersAgentSpec{
		Account:     Account,
		ClusterName: Cluster,
		Gateways: cbcontainersv1.CBContainersGatewaysSpec{
			ApiGateway: cbcontainersv1.CBContainersApiGatewaySpec{
				Scheme:  ApiGateWayScheme,
				Host:    ApiGateWayHost,
				Port:    ApiGateWayPort,
				Adapter: ApiGateWayAdapter,
			},
			CoreEventsGateway: cbcontainersv1.CBContainersEventsGatewaySpec{
				Host: CoreEventsGateWayHost,
			},
			HardeningEventsGateway: cbcontainersv1.CBContainersEventsGatewaySpec{
				Host: HardeningEventsGateWayHost,
			},
			RuntimeEventsGateway: cbcontainersv1.CBContainersEventsGatewaySpec{
				Host: RuntimeEventsGateWayHost,
			},
		},
		Components: cbcontainersv1.CBContainersComponentsSpec{
			RuntimeProtection: cbcontainersv1.CBContainersRuntimeProtectionSpec{
				Enabled: &trueRef,
			},
		},
	}

	if k8sVersion == "" {
		k8sVersion = DefaultKubeletVersion
	}

	mockObjects := &StateApplierTestMocks{
		client:              testUtilsMocks.NewMockClient(ctrl),
		secretValuesCreator: mocks.NewMockTlsSecretsValuesCreator(ctrl),
		componentApplier:    mocks.NewMockAgentComponentApplier(ctrl),
		agentSpec:           agentSpec,
		kubeletVersion:      k8sVersion,
	}

	setup(mockObjects)

	stateApplier := state.NewStateApplier(mockObjects.componentApplier, k8sVersion, mockObjects.secretValuesCreator, &logrTesting.TestLogger{T: t})
	return stateApplier.ApplyDesiredState(context.Background(), agentSpec, nil, nil)
}

func getAppliedAndDeletedObjects(t *testing.T, k8sVersion string, setup StateApplierTestSetup, appliedK8sObjectsChangers ...AppliedK8sObjectsChanger) ([]K8sObjectDetails, []K8sObjectDetails, error) {
	appliedObjects := make([]K8sObjectDetails, 0)
	deletedObjects := make([]K8sObjectDetails, 0)

	_, err := testStateApplier(t, func(mocks *StateApplierTestMocks) {
		if setup != nil {
			setup(mocks)
		}
		mocks.componentApplier.EXPECT().Apply(gomock.Any(), gomock.Any(), mocks.agentSpec, gomock.Any()).
			DoAndReturn(func(ctx context.Context, obj agent_applyment.AgentComponentBuilder, cr *cbcontainersv1.CBContainersAgentSpec, options ...*options.ApplyOptions) (bool, client.Object, error) {
				namespacedName := obj.NamespacedName()
				k8sObject := obj.EmptyK8sObject()
				objType := reflect.TypeOf(k8sObject)
				objectDetails := K8sObjectDetails{Namespace: namespacedName.Namespace, Name: namespacedName.Name, ObjectType: objType}

				appliedObjects = append(appliedObjects, objectDetails)

				for _, changeAppliedK8sObjects := range appliedK8sObjectsChangers {
					changeAppliedK8sObjects(objectDetails, k8sObject)
				}

				return true, k8sObject, nil
			}).AnyTimes()

		mocks.componentApplier.EXPECT().Delete(gomock.Any(), gomock.Any(), mocks.agentSpec).
			DoAndReturn(func(ctx context.Context, obj agent_applyment.AgentComponentBuilder, cr *cbcontainersv1.CBContainersAgentSpec) (bool, error) {
				namespacedName := obj.NamespacedName()
				objType := reflect.TypeOf(obj.EmptyK8sObject())
				deletedObjects = append(deletedObjects, K8sObjectDetails{Namespace: namespacedName.Namespace, Name: namespacedName.Name, ObjectType: objType})
				return true, nil
			}).AnyTimes()
	}, k8sVersion)

	return appliedObjects, deletedObjects, err
}

func getAndAssertAppliedAndDeletedObjects(t *testing.T, k8sVersion string, setup StateApplierTestSetup) ([]K8sObjectDetails, []K8sObjectDetails) {
	appliedObjects, deletedObjects, err := getAppliedAndDeletedObjects(t, k8sVersion, setup, MutateDeploymentsToBeWithReadyReplica(EnforcerDeploymentDetails, ResolverDeploymentDetails))

	require.NoError(t, err)
	require.Len(t, appliedObjects, NumberOfExpectedAppliedObjects)
	return appliedObjects, deletedObjects
}

func TestConfigMapIsApplied(t *testing.T) {
	appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t, "", nil)
	require.Contains(t, appliedObjects, K8sObjectDetails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       commonState.DataPlaneConfigmapName,
		ObjectType: reflect.TypeOf(&coreV1.ConfigMap{}),
	})
}

func TestSecretIsApplied(t *testing.T) {
	appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t, "", nil)
	require.Contains(t, appliedObjects, K8sObjectDetails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       commonState.RegistrySecretName,
		ObjectType: reflect.TypeOf(&coreV1.Secret{}),
	})
}

func TestPriorityClassIsApplied(t *testing.T) {
	testPriorityClassIsApplied := func(t *testing.T, objType reflect.Type, k8sVersion string) {
		appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t, k8sVersion, nil)
		require.Contains(t, appliedObjects, K8sObjectDetails{
			Namespace:  "",
			Name:       commonState.DataPlanePriorityClassName,
			ObjectType: objType,
		})
	}

	t.Run("With K8s version v1.14 or higher, should use `schedulingV1`", func(t *testing.T) {
		testPriorityClassIsApplied(t, reflect.TypeOf(&schedulingV1.PriorityClass{}), DefaultKubeletVersion)
	})

	t.Run("With K8s version lower then v1.14 but higher or equal to v1.11, should use `schedulingV1beta1`", func(t *testing.T) {
		testPriorityClassIsApplied(t, reflect.TypeOf(&schedulingV1beta1.PriorityClass{}), "v1.13.2")
	})

	t.Run("With K8s version lower then v1.11, should use `schedulingV1alpha1`", func(t *testing.T) {
		testPriorityClassIsApplied(t, reflect.TypeOf(&schedulingV1alpha1.PriorityClass{}), "v1.08")
	})

}

func TestMonitorDeploymentIsApplied(t *testing.T) {
	appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t, "", nil)
	require.Contains(t, appliedObjects, K8sObjectDetails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       components.MonitorName,
		ObjectType: reflect.TypeOf(&appsV1.Deployment{}),
	})
}

func TestTlsSecretIsApplied(t *testing.T) {
	appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t, "", nil)
	require.Contains(t, appliedObjects, K8sObjectDetails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       components.EnforcerTlsName,
		ObjectType: reflect.TypeOf(&coreV1.Secret{}),
	})
}

func TestEnforcerServiceIsApplied(t *testing.T) {
	appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t, "", nil)
	require.Contains(t, appliedObjects, K8sObjectDetails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       components.EnforcerName,
		ObjectType: reflect.TypeOf(&coreV1.Service{}),
	})
}

func TestEnforcerDeploymentIsApplied(t *testing.T) {
	appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t, "", nil)
	require.Contains(t, appliedObjects, EnforcerDeploymentDetails)
}

func TestEnforcerWebhooksAreApplied(t *testing.T) {
	webhooksAreAppliedTestsForVersion := func(t *testing.T, k8sVersion string, validatingWebhook K8sObjectDetails, mutatingWebhook K8sObjectDetails) {
		t.Helper()

		withMutatingWebhook := func(mocks *StateApplierTestMocks) {
			mocks.agentSpec.Components.Basic.Enforcer.EnableEnforcingWebhook = true
		}

		t.Run("With default spec, should apply the validating webhook and delete the mutating webhook", func(t *testing.T) {
			appliedObjects, deletedObjects, err := getAppliedAndDeletedObjects(t, k8sVersion, nil, MutateDeploymentsToBeWithReadyReplica(EnforcerDeploymentDetails, ResolverDeploymentDetails))
			require.NoError(t, err)
			require.Contains(t, appliedObjects, validatingWebhook)
			require.NotContains(t, appliedObjects, mutatingWebhook)
			require.NotContains(t, deletedObjects, validatingWebhook)
			require.Contains(t, deletedObjects, mutatingWebhook)
		})

		t.Run("With enforcing webhook enabled, should apply both webhooks", func(t *testing.T) {
			appliedObjects, deletedObjects, err := getAppliedAndDeletedObjects(t, k8sVersion, withMutatingWebhook, MutateDeploymentsToBeWithReadyReplica(EnforcerDeploymentDetails, ResolverDeploymentDetails))
			require.NoError(t, err)
			require.Contains(t, appliedObjects, validatingWebhook)
			require.Contains(t, appliedObjects, mutatingWebhook)
			require.NotContains(t, deletedObjects, validatingWebhook)
			require.NotContains(t, deletedObjects, mutatingWebhook)
		})
	}

	t.Run("With empty K8S version, should use `v1` by default", func(t *testing.T) {
		k8sVersion := ""
		webhooksAreAppliedTestsForVersion(t, k8sVersion, EnforcerValidatingWebhookDetails, EnforcerMutatingWebhookDetails)
	})

	t.Run("With K8S version 1.15 or lower, should use `v1beta1` version of webhook", func(t *testing.T) {
		k8sVersion := "v1.15"
		legacyValidatingWebhook := EnforcerValidatingWebhookDetails
		legacyValidatingWebhook.ObjectType = reflect.TypeOf(&admissionsV1Beta1.ValidatingWebhookConfiguration{})
		legacyMutatingWebhook := EnforcerMutatingWebhookDetails
		legacyMutatingWebhook.ObjectType = reflect.TypeOf(&admissionsV1Beta1.MutatingWebhookConfiguration{})

		webhooksAreAppliedTestsForVersion(t, k8sVersion, legacyValidatingWebhook, legacyMutatingWebhook)
	})

	t.Run("With K8S version 1.16 or higher, should use `v1` version of webhook", func(t *testing.T) {
		k8sVersion := "v1.16"
		webhooksAreAppliedTestsForVersion(t, k8sVersion, EnforcerValidatingWebhookDetails, EnforcerMutatingWebhookDetails)
	})
}

func TestEnforcerWebhooksAreDeleted(t *testing.T) {
	assertWebhooksAreDeleted := func(t *testing.T, appliedObjects, deletedObjects []K8sObjectDetails, webhookObjects ...K8sObjectDetails) {
		require.NotSubset(t, appliedObjects, webhookObjects)
		require.Subset(t, deletedObjects, webhookObjects)
	}

	t.Run("With empty K8S version, should use `v1` by default", func(t *testing.T) {
		appliedObjects, deletedObjects, err := getAppliedAndDeletedObjects(t, "", nil)
		require.NoError(t, err)
		assertWebhooksAreDeleted(t, appliedObjects, deletedObjects, EnforcerValidatingWebhookDetails, EnforcerMutatingWebhookDetails)
	})

	t.Run("With K8S version 1.15 or lower, should use `v1beta1` version of webhook", func(t *testing.T) {
		appliedObjects, deletedObjects, err := getAppliedAndDeletedObjects(t, "v1.15", nil)
		require.NoError(t, err)
		legacyValidatingWebhook := EnforcerValidatingWebhookDetails
		legacyValidatingWebhook.ObjectType = reflect.TypeOf(&admissionsV1Beta1.ValidatingWebhookConfiguration{})
		legacyMutatingWebhook := EnforcerMutatingWebhookDetails
		legacyMutatingWebhook.ObjectType = reflect.TypeOf(&admissionsV1Beta1.MutatingWebhookConfiguration{})
		assertWebhooksAreDeleted(t, appliedObjects, deletedObjects, legacyValidatingWebhook, legacyMutatingWebhook)
	})

	t.Run("With K8S version 1.16 or higher, should use `v1` version of webhook", func(t *testing.T) {
		appliedObjects, deletedObjects, err := getAppliedAndDeletedObjects(t, "v1.16", nil)
		require.NoError(t, err)
		assertWebhooksAreDeleted(t, appliedObjects, deletedObjects, EnforcerValidatingWebhookDetails, EnforcerMutatingWebhookDetails)
	})
}

func TestStateReporterDeploymentIsApplied(t *testing.T) {
	appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t, "", nil)
	require.Contains(t, appliedObjects, K8sObjectDetails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       components.StateReporterName,
		ObjectType: reflect.TypeOf(&appsV1.Deployment{}),
	})
}

func TestResolverServiceIsApplied(t *testing.T) {
	appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t, "", nil)
	require.Contains(t, appliedObjects, K8sObjectDetails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       components.ResolverName,
		ObjectType: reflect.TypeOf(&coreV1.Service{}),
	})
}

func TestResolverDeploymentIsApplied(t *testing.T) {
	appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t, "", nil)
	require.Contains(t, appliedObjects, ResolverDeploymentDetails)
}

func TestSensorDaemonsetIsApplied(t *testing.T) {
	appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t, "", nil)
	require.Contains(t, appliedObjects, K8sObjectDetails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       components.SensorName,
		ObjectType: reflect.TypeOf(&appsV1.DaemonSet{}),
	})
}
