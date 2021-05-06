package controllers_test

import (
	"context"
	"fmt"
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
	"testing"
)

type SetupHardeningControllerTest func(*HardeningControllerTestMocks)

type HardeningControllerTestMocks struct {
	client                *testUtilsMocks.MockClient
	HardeningStateApplier *mocks.MockHardeningStateApplier
	ctx                   context.Context
}

var (
	HardeningVersion = test_utils.RandomString()

	HardeningCustomResourceItems = []cbcontainersv1.CBContainersHardening{
		{
			Spec: cbcontainersv1.CBContainersHardeningSpec{Version: HardeningVersion},
		},
	}
)

func testCBContainersHardeningController(t *testing.T, setups ...SetupHardeningControllerTest) (ctrlRuntime.Result, error) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mocksObjects := &HardeningControllerTestMocks{
		ctx:                   context.TODO(),
		client:                testUtilsMocks.NewMockClient(ctrl),
		HardeningStateApplier: mocks.NewMockHardeningStateApplier(ctrl),
	}

	for _, setup := range setups {
		setup(mocksObjects)
	}

	controller := &controllers.CBContainersHardeningReconciler{
		Client: mocksObjects.client,
		Log:    &logrTesting.TestLogger{T: t},
		Scheme: &runtime.Scheme{},

		HardeningStateApplier: mocksObjects.HardeningStateApplier,
	}

	return controller.Reconcile(mocksObjects.ctx, ctrlRuntime.Request{})
}

func setupHardeningCustomResource(testMocks *HardeningControllerTestMocks) {
	testMocks.client.EXPECT().List(testMocks.ctx, &cbcontainersv1.CBContainersHardeningList{}).
		Do(func(ctx context.Context, list *cbcontainersv1.CBContainersHardeningList) {
			list.Items = HardeningCustomResourceItems
		}).
		Return(nil)
}

func TestListHardeningResourcesErrorShouldReturnError(t *testing.T) {
	_, err := testCBContainersHardeningController(t, func(testMocks *HardeningControllerTestMocks) {
		testMocks.client.EXPECT().List(testMocks.ctx, &cbcontainersv1.CBContainersHardeningList{}).Return(fmt.Errorf(""))
	})

	require.Error(t, err)
}

func TestNotFindingAnyHardeningResourceShouldReturnNil(t *testing.T) {
	result, err := testCBContainersHardeningController(t, func(testMocks *HardeningControllerTestMocks) {
		testMocks.client.EXPECT().List(testMocks.ctx, &cbcontainersv1.CBContainersHardeningList{}).Return(nil)
	})

	require.NoError(t, err)
	require.Equal(t, result, ctrlRuntime.Result{})
}

func TestFindingMoreThanOneHardeningResourceShouldReturnError(t *testing.T) {
	_, err := testCBContainersHardeningController(t, func(testMocks *HardeningControllerTestMocks) {
		testMocks.client.EXPECT().List(testMocks.ctx, &cbcontainersv1.CBContainersHardeningList{}).
			Do(func(ctx context.Context, list *cbcontainersv1.CBContainersHardeningList) {
				list.Items = append(list.Items, cbcontainersv1.CBContainersHardening{})
				list.Items = append(list.Items, cbcontainersv1.CBContainersHardening{})
			}).
			Return(nil)
	})

	require.Error(t, err)
}

func TestHardeningReconcile(t *testing.T) {
	t.Run("When state applier returns error, reconcile should return error", func(t *testing.T) {
		_, err := testCBContainersHardeningController(t, setupHardeningCustomResource, func(testMocks *HardeningControllerTestMocks) {
			testMocks.HardeningStateApplier.EXPECT().ApplyDesiredState(testMocks.ctx, &HardeningCustomResourceItems[0], testMocks.client, gomock.Any()).Return(false, fmt.Errorf(""))
		})

		require.Error(t, err)
	})

	t.Run("When state applier returns state was changed, reconcile should return Requeue true", func(t *testing.T) {
		result, err := testCBContainersHardeningController(t, setupHardeningCustomResource, func(testMocks *HardeningControllerTestMocks) {
			testMocks.HardeningStateApplier.EXPECT().ApplyDesiredState(testMocks.ctx, &HardeningCustomResourceItems[0], testMocks.client, gomock.Any()).Return(true, nil)
		})

		require.NoError(t, err)
		require.Equal(t, result, ctrlRuntime.Result{Requeue: true})
	})

	t.Run("When state applier returns state was not changed, reconcile should return default Requeue", func(t *testing.T) {
		result, err := testCBContainersHardeningController(t, setupHardeningCustomResource, func(testMocks *HardeningControllerTestMocks) {
			testMocks.HardeningStateApplier.EXPECT().ApplyDesiredState(testMocks.ctx, &HardeningCustomResourceItems[0], testMocks.client, gomock.Any()).Return(false, nil)
		})

		require.NoError(t, err)
		require.Equal(t, result, ctrlRuntime.Result{})
	})
}
