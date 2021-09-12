package runtime_test

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
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/runtime"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/runtime/mocks"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/runtime/objects"
	"github.com/vmware/cbcontainers-operator/cbcontainers/test_utils"
	testUtilsMocks "github.com/vmware/cbcontainers-operator/cbcontainers/test_utils/mocks"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	NumberOfExpectedAppliedObjects = 3
)

type AppliedK8sObjectsChanger func(K8sObjectDetails, client.Object)

var (
	Version = test_utils.RandomString()

	ResolverDeploymentDetails = K8sObjectDetails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       objects.ResolverName,
		ObjectType: reflect.TypeOf(&appsV1.Deployment{}),
	}

	MutateDeploymentReadyReplicas = func(details K8sObjectDetails, object client.Object, readyReplicas int32) {
		if details != ResolverDeploymentDetails {
			return
		}

		enforcerDeployment := object.(*appsV1.Deployment)
		enforcerDeployment.Status.ReadyReplicas = readyReplicas
	}

	MutateDeploymentToBeWithReadyReplica AppliedK8sObjectsChanger = func(details K8sObjectDetails, object client.Object) {
		MutateDeploymentReadyReplicas(details, object, 1)
	}
)

type RuntimeStateApplierTestMocks struct {
	client              *testUtilsMocks.MockClient
	childApplier        *mocks.MockRuntimeChildK8sObjectApplier
	cbContainersRuntime *cbcontainersv1.CBContainersRuntimeSpec
}

type RuntimeStateApplierTestSetup func(*RuntimeStateApplierTestMocks)

type K8sObjectDetails struct {
	Namespace  string
	Name       string
	ObjectType reflect.Type
}

func testRuntimeStateApplier(t *testing.T, setup RuntimeStateApplierTestSetup) (bool, error) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cbContainersRuntime := &cbcontainersv1.CBContainersRuntimeSpec{}

	mockObjects := &RuntimeStateApplierTestMocks{
		client:              testUtilsMocks.NewMockClient(ctrl),
		childApplier:        mocks.NewMockRuntimeChildK8sObjectApplier(ctrl),
		cbContainersRuntime: cbContainersRuntime,
	}

	setup(mockObjects)

	return runtime.NewRuntimeStateApplier(&logrTesting.TestLogger{T: t}, mockObjects.childApplier).ApplyDesiredState(context.Background(), cbContainersRuntime, Version, "", mockObjects.client, nil)
}

func getAppliedAndDeletedObjects(t *testing.T, appliedK8sObjectsChangers ...AppliedK8sObjectsChanger) ([]K8sObjectDetails, []K8sObjectDetails, error) {
	appliedObjects := make([]K8sObjectDetails, 0)
	deletedObjects := make([]K8sObjectDetails, 0)

	_, err := testRuntimeStateApplier(t, func(mocks *RuntimeStateApplierTestMocks) {
		mocks.childApplier.EXPECT().ApplyRuntimeChildK8sObject(gomock.Any(), mocks.cbContainersRuntime, Version, gomock.Any(), mocks.client, gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, cr *cbcontainersv1.CBContainersRuntimeSpec, client client.Client, childObject runtime.RuntimeChildK8sObject, options ...*options.ApplyOptions) (bool, client.Object, error) {
				namespacedName := childObject.RuntimeChildNamespacedName(cr)
				k8sObject := childObject.EmptyK8sObject()
				objType := reflect.TypeOf(k8sObject)
				objectDetails := K8sObjectDetails{Namespace: namespacedName.Namespace, Name: namespacedName.Name, ObjectType: objType}
				appliedObjects = append(appliedObjects, objectDetails)

				for _, changeAppliedK8sObjects := range appliedK8sObjectsChangers {
					changeAppliedK8sObjects(objectDetails, k8sObject)
				}

				return true, k8sObject, nil
			}).AnyTimes()

		mocks.childApplier.EXPECT().DeleteK8sObjectIfExists(gomock.Any(), mocks.cbContainersRuntime, mocks.client, gomock.Any()).
			DoAndReturn(func(ctx context.Context, cr *cbcontainersv1.CBContainersRuntimeSpec, client client.Client, obj runtime.RuntimeChildK8sObject) (bool, error) {
				namespacedName := obj.RuntimeChildNamespacedName(cr)
				objType := reflect.TypeOf(obj.EmptyK8sObject())
				deletedObjects = append(deletedObjects, K8sObjectDetails{Namespace: namespacedName.Namespace, Name: namespacedName.Name, ObjectType: objType})
				return true, nil
			}).AnyTimes()
	})

	return appliedObjects, deletedObjects, err
}

func getAndAssertAppliedAndDeletedObjects(t *testing.T) ([]K8sObjectDetails, []K8sObjectDetails) {
	appliedObjects, deletedObjects, err := getAppliedAndDeletedObjects(t, MutateDeploymentToBeWithReadyReplica)

	require.NoError(t, err)
	require.Len(t, appliedObjects, NumberOfExpectedAppliedObjects)
	return appliedObjects, deletedObjects
}

func TestResolverServiceIsApplied(t *testing.T) {
	appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t)
	require.Contains(t, appliedObjects, K8sObjectDetails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       objects.ResolverName,
		ObjectType: reflect.TypeOf(&coreV1.Service{}),
	})
}

func TestEnforcerDeploymentIsApplied(t *testing.T) {
	appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t)
	require.Contains(t, appliedObjects, ResolverDeploymentDetails)
}

func TestSensorDaemonsetIsApplied(t *testing.T) {
	appliedObjects, _ := getAndAssertAppliedAndDeletedObjects(t)
	require.Contains(t, appliedObjects, K8sObjectDetails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       objects.SensorName,
		ObjectType: reflect.TypeOf(&appsV1.DaemonSet{}),
	})
}
