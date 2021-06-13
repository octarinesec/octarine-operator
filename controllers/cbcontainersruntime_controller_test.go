package controllers_test

import (
	"context"
	"fmt"
	"testing"

	logrTesting "github.com/go-logr/logr/testing"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/test_utils"
	testUtilsMocks "github.com/vmware/cbcontainers-operator/cbcontainers/test_utils/mocks"
	"github.com/vmware/cbcontainers-operator/controllers"
	"github.com/vmware/cbcontainers-operator/controllers/mocks"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlRuntime "sigs.k8s.io/controller-runtime"
)

type SetupRuntimeControllerTest func(*RuntimeControllerTestMocks)

type RuntimeControllerTestMocks struct {
	client              *testUtilsMocks.MockClient
	RuntimeStateApplier *mocks.MockRuntimeStateApplier
	ctx                 context.Context
}

var (
	RuntimeVersion = test_utils.RandomString()

	RuntimeCustomResourceItems = []cbcontainersv1.CBContainersRuntime{
		{
			Spec: cbcontainersv1.CBContainersRuntimeSpec{Version: RuntimeVersion},
		},
	}
)

func testCBContainersRuntimeController(t *testing.T, setups ...SetupRuntimeControllerTest) (ctrlRuntime.Result, error) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mocksObjects := &RuntimeControllerTestMocks{
		ctx:                 context.TODO(),
		client:              testUtilsMocks.NewMockClient(ctrl),
		RuntimeStateApplier: mocks.NewMockRuntimeStateApplier(ctrl),
	}

	for _, setup := range setups {
		setup(mocksObjects)
	}

	controller := &controllers.CBContainersRuntimeReconciler{
		Client: mocksObjects.client,
		Log:    &logrTesting.TestLogger{T: t},
		Scheme: &runtime.Scheme{},

		RuntimeStateApplier: mocksObjects.RuntimeStateApplier,
	}

	return controller.Reconcile(mocksObjects.ctx, ctrlRuntime.Request{})
}

func setupRuntimeCustomResource(testMocks *RuntimeControllerTestMocks) {
	testMocks.client.EXPECT().List(testMocks.ctx, &cbcontainersv1.CBContainersRuntimeList{}).
		Do(func(ctx context.Context, list *cbcontainersv1.CBContainersRuntimeList, _ ...interface{}) {
			list.Items = RuntimeCustomResourceItems
		}).
		Return(nil)
}

func TestListRuntimeResourcesErrorShouldReturnError(t *testing.T) {
	_, err := testCBContainersRuntimeController(t, func(testMocks *RuntimeControllerTestMocks) {
		testMocks.client.EXPECT().List(testMocks.ctx, &cbcontainersv1.CBContainersRuntimeList{}).Return(fmt.Errorf(""))
	})

	require.Error(t, err)
}

func TestNotFindingAnyRuntimeResourceShouldReturnNil(t *testing.T) {
	result, err := testCBContainersRuntimeController(t, func(testMocks *RuntimeControllerTestMocks) {
		testMocks.client.EXPECT().List(testMocks.ctx, &cbcontainersv1.CBContainersRuntimeList{}).Return(nil)
	})

	require.NoError(t, err)
	require.Equal(t, result, ctrlRuntime.Result{})
}

func TestFindingMoreThanOneRuntimeResourceShouldReturnError(t *testing.T) {
	_, err := testCBContainersRuntimeController(t, func(testMocks *RuntimeControllerTestMocks) {
		testMocks.client.EXPECT().List(testMocks.ctx, &cbcontainersv1.CBContainersRuntimeList{}).
			Do(func(ctx context.Context, list *cbcontainersv1.CBContainersRuntimeList, _ ...interface{}) {
				list.Items = append(list.Items, cbcontainersv1.CBContainersRuntime{})
				list.Items = append(list.Items, cbcontainersv1.CBContainersRuntime{})
			}).
			Return(nil)
	})

	require.Error(t, err)
}

func TestRuntimeReconcile(t *testing.T) {
	t.Run("When state applier returns error, reconcile should return error", func(t *testing.T) {
		_, err := testCBContainersRuntimeController(t, setupRuntimeCustomResource, func(testMocks *RuntimeControllerTestMocks) {
			testMocks.RuntimeStateApplier.EXPECT().ApplyDesiredState(testMocks.ctx, &RuntimeCustomResourceItems[0], testMocks.client, gomock.Any()).Return(false, fmt.Errorf(""))
		})

		require.Error(t, err)
	})

	t.Run("When state applier returns state was changed, reconcile should return Requeue true", func(t *testing.T) {
		result, err := testCBContainersRuntimeController(t, setupRuntimeCustomResource, func(testMocks *RuntimeControllerTestMocks) {
			testMocks.RuntimeStateApplier.EXPECT().ApplyDesiredState(testMocks.ctx, &RuntimeCustomResourceItems[0], testMocks.client, gomock.Any()).Return(true, nil)
		})

		require.NoError(t, err)
		require.Equal(t, result, ctrlRuntime.Result{Requeue: true})
	})

	t.Run("When state applier returns state was not changed, reconcile should return default Requeue", func(t *testing.T) {
		result, err := testCBContainersRuntimeController(t, setupRuntimeCustomResource, func(testMocks *RuntimeControllerTestMocks) {
			testMocks.RuntimeStateApplier.EXPECT().ApplyDesiredState(testMocks.ctx, &RuntimeCustomResourceItems[0], testMocks.client, gomock.Any()).Return(false, nil)
		})

		require.NoError(t, err)
		require.Equal(t, result, ctrlRuntime.Result{})
	})
}
