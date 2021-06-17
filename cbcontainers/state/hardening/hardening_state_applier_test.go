package hardening_test

import (
	"context"
	"reflect"
	"testing"

	logrTesting "github.com/go-logr/logr/testing"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/hardening"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/hardening/mocks"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/hardening/objects"
	"github.com/vmware/cbcontainers-operator/cbcontainers/test_utils"
	testUtilsMocks "github.com/vmware/cbcontainers-operator/cbcontainers/test_utils/mocks"
	admissionsV1 "k8s.io/api/admissionregistration/v1beta1"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultKubeletVersion = "v1.20.2"

	NumberOfExpectedAppliedObjects = 5
)

type AppliedK8sObjectsChanger func(K8sObjectDetails, client.Object)

var (
	Version           = test_utils.RandomString()
	EventsGateWayHost = test_utils.RandomString()

	EnforcerDeploymentDetails = K8sObjectDetails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       objects.EnforcerName,
		ObjectType: reflect.TypeOf(&appsV1.Deployment{}),
	}

	EnforcerWebhookDetails = K8sObjectDetails{
		Namespace:  "",
		Name:       objects.EnforcerName,
		ObjectType: reflect.TypeOf(&admissionsV1.ValidatingWebhookConfiguration{}),
	}

	MutateDeploymentReadyReplicas = func(details K8sObjectDetails, object client.Object, readyReplicas int32) {
		if details != EnforcerDeploymentDetails {
			return
		}

		enforcerDeployment := object.(*appsV1.Deployment)
		enforcerDeployment.Status.ReadyReplicas = readyReplicas
	}

	MutateDeploymentToBeWithReadyReplica AppliedK8sObjectsChanger = func(details K8sObjectDetails, object client.Object) {
		MutateDeploymentReadyReplicas(details, object, 1)
	}

	MutateDeploymentToBeWithNoReadyReplica AppliedK8sObjectsChanger = func(details K8sObjectDetails, object client.Object) {
		MutateDeploymentReadyReplicas(details, object, 0)
	}
)

type HardeningStateApplierTestMocks struct {
	client                *testUtilsMocks.MockClient
	secretValuesCreator   *mocks.MockTlsSecretsValuesCreator
	childApplier          *mocks.MockHardeningChildK8sObjectApplier
	cbContainersHardening *cbcontainersv1.CBContainersHardening
	kubeletVersion        string
}

type HardeningStateApplierTestSetup func(*HardeningStateApplierTestMocks)

type K8sObjectDetails struct {
	Namespace  string
	Name       string
	ObjectType reflect.Type
}

func testHardeningStateApplier(t *testing.T, setup HardeningStateApplierTestSetup, k8sVersion string) (bool, error) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cbContainersHardening := &cbcontainersv1.CBContainersHardening{
		Spec: cbcontainersv1.CBContainersHardeningSpec{
			Version: Version,
			EventsGatewaySpec: cbcontainersv1.CBContainersEventsGatewaySpec{
				Host: EventsGateWayHost,
			},
		},
	}

	if k8sVersion == "" {
		k8sVersion = DefaultKubeletVersion
	}

	mockObjects := &HardeningStateApplierTestMocks{
		client:                testUtilsMocks.NewMockClient(ctrl),
		secretValuesCreator:   mocks.NewMockTlsSecretsValuesCreator(ctrl),
		childApplier:          mocks.NewMockHardeningChildK8sObjectApplier(ctrl),
		cbContainersHardening: cbContainersHardening,
		kubeletVersion:        k8sVersion,
	}

	setup(mockObjects)

	return hardening.NewHardeningStateApplier(&logrTesting.TestLogger{T: t}, k8sVersion, mockObjects.secretValuesCreator, mockObjects.childApplier).ApplyDesiredState(context.Background(), cbContainersHardening, mockObjects.client, nil)
}

func getAppliedAndDeletedObjects(t *testing.T, k8sVersion string, appliedK8sObjectsChangers ...AppliedK8sObjectsChanger) ([]K8sObjectDetails, []K8sObjectDetails, error) {
	appliedObjects := make([]K8sObjectDetails, 0)
	deletedObjects := make([]K8sObjectDetails, 0)

	_, err := testHardeningStateApplier(t, func(mocks *HardeningStateApplierTestMocks) {
		mocks.childApplier.EXPECT().ApplyHardeningChildK8sObject(gomock.Any(), mocks.cbContainersHardening, mocks.client, gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, cr *cbcontainersv1.CBContainersHardening, client client.Client, childObject hardening.HardeningChildK8sObject, options ...*options.ApplyOptions) (bool, client.Object, error) {
				namespacedName := childObject.HardeningChildNamespacedName(cr)
				k8sObject := childObject.EmptyK8sObject()
				objType := reflect.TypeOf(k8sObject)
				objectDetails := K8sObjectDetails{Namespace: namespacedName.Namespace, Name: namespacedName.Name, ObjectType: objType}
				appliedObjects = append(appliedObjects, objectDetails)

				for _, changeAppliedK8sObjects := range appliedK8sObjectsChangers {
					changeAppliedK8sObjects(objectDetails, k8sObject)
				}

				return true, k8sObject, nil
			}).AnyTimes()

		mocks.childApplier.EXPECT().DeleteK8sObjectIfExists(gomock.Any(), mocks.cbContainersHardening, mocks.client, gomock.Any()).
			DoAndReturn(func(ctx context.Context, cr *cbcontainersv1.CBContainersHardening, client client.Client, obj hardening.HardeningChildK8sObject) (bool, error) {
				namespacedName := obj.HardeningChildNamespacedName(cr)
				objType := reflect.TypeOf(obj.EmptyK8sObject())
				deletedObjects = append(deletedObjects, K8sObjectDetails{Namespace: namespacedName.Namespace, Name: namespacedName.Name, ObjectType: objType})
				return true, nil
			}).AnyTimes()
	}, k8sVersion)

	return appliedObjects, deletedObjects, err
}

func getAndAssertAppliedAndDeletedObjects(t *testing.T, k8sVersion string) ([]K8sObjectDetails, []K8sObjectDetails) {
	appliedObjects, deletedObjects, err := getAppliedAndDeletedObjects(t, k8sVersion, MutateDeploymentToBeWithReadyReplica)

	require.NoError(t, err)
	require.Len(t, appliedObjects, NumberOfExpectedAppliedObjects)
	return appliedObjects, deletedObjects
}

func TestTlsSecretIsApplied(t *testing.T) {
	appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t, "")
	require.Contains(t, appliedObjects, K8sObjectDetails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       objects.EnforcerTlsName,
		ObjectType: reflect.TypeOf(&coreV1.Secret{}),
	})
}

func TestEnforcerServiceIsApplied(t *testing.T) {
	appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t, "")
	require.Contains(t, appliedObjects, K8sObjectDetails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       objects.EnforcerName,
		ObjectType: reflect.TypeOf(&coreV1.Service{}),
	})
}

func TestEnforcerDeploymentIsApplied(t *testing.T) {
	appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t, "")
	require.Contains(t, appliedObjects, EnforcerDeploymentDetails)
}

func TestEnforcerWebhookIsApplied(t *testing.T) {
	appliedObjects, deletedObjects := getAndAssertAppliedAndDeletedObjects(t, "")
	require.Contains(t, appliedObjects, EnforcerWebhookDetails)
	require.NotContains(t, deletedObjects, EnforcerWebhookDetails)
}

func TestEnforcerWebhookIsDeleted(t *testing.T) {
	appliedObjects, deletedObjects, err := getAppliedAndDeletedObjects(t, "", MutateDeploymentToBeWithNoReadyReplica)
	require.NoError(t, err)
	require.Len(t, appliedObjects, NumberOfExpectedAppliedObjects-1)
	require.NotContains(t, appliedObjects, EnforcerWebhookDetails)
	require.Contains(t, deletedObjects, EnforcerWebhookDetails)
}

func TestStateReporterDeploymentIsApplied(t *testing.T) {
	appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t, "")
	require.Contains(t, appliedObjects, K8sObjectDetails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       objects.StateReporterName,
		ObjectType: reflect.TypeOf(&appsV1.Deployment{}),
	})
}
