package controllers_test

import (
	"context"
	"fmt"
	"testing"

	logrTesting "github.com/go-logr/logr/testing"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	"github.com/vmware/cbcontainers-operator/cbcontainers/test_utils"
	testUtilsMocks "github.com/vmware/cbcontainers-operator/cbcontainers/test_utils/mocks"
	"github.com/vmware/cbcontainers-operator/controllers"
	"github.com/vmware/cbcontainers-operator/controllers/mocks"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrlRuntime "sigs.k8s.io/controller-runtime"
)

type SetupClusterControllerTest func(*ClusterControllerTestMocks)

type ClusterControllerTestMocks struct {
	client           *testUtilsMocks.MockClient
	ClusterProcessor *mocks.MockClusterProcessor
	StateApplier     *mocks.MockStateApplier
	ctx              context.Context
}

const (
	MyClusterTokenValue = "my-token-value"
)

var (
	ClusterAccessTokenSecretName = test_utils.RandomString()

	ClusterCustomResourceItems = []cbcontainersv1.CBContainersAgent{
		{
			Spec: cbcontainersv1.CBContainersAgentSpec{
				ApiGatewaySpec: cbcontainersv1.CBContainersApiGatewaySpec{
					AccessTokenSecretName: ClusterAccessTokenSecretName,
				},
			},
		},
	}
)

func testCBContainersClusterController(t *testing.T, setups ...SetupClusterControllerTest) (ctrlRuntime.Result, error) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mocksObjects := &ClusterControllerTestMocks{
		ctx:              context.TODO(),
		client:           testUtilsMocks.NewMockClient(ctrl),
		ClusterProcessor: mocks.NewMockClusterProcessor(ctrl),
		StateApplier:     mocks.NewMockStateApplier(ctrl),
	}

	for _, setup := range setups {
		setup(mocksObjects)
	}

	controller := &controllers.CBContainersAgentController{
		Client: mocksObjects.client,
		Log:    &logrTesting.TestLogger{T: t},
		Scheme: &runtime.Scheme{},

		ClusterProcessor: mocksObjects.ClusterProcessor,
		StateApplier:     mocksObjects.StateApplier,
	}

	return controller.Reconcile(mocksObjects.ctx, ctrlRuntime.Request{})
}

func setupClusterCustomResource(testMocks *ClusterControllerTestMocks) {
	testMocks.client.EXPECT().List(testMocks.ctx, &cbcontainersv1.CBContainersAgentList{}).
		Do(func(ctx context.Context, list *cbcontainersv1.CBContainersAgentList, _ ...interface{}) {
			list.Items = ClusterCustomResourceItems
		}).
		Return(nil)
}

func setUpTokenSecretValues(testMocks *ClusterControllerTestMocks) {
	accessTokenSecretNamespacedName := types.NamespacedName{Name: ClusterAccessTokenSecretName, Namespace: commonState.DataPlaneNamespaceName}
	testMocks.client.EXPECT().Get(testMocks.ctx, accessTokenSecretNamespacedName, &corev1.Secret{}).
		Do(func(ctx context.Context, namespacedName types.NamespacedName, secret *corev1.Secret) {
			secret.Data = map[string][]byte{
				commonState.AccessTokenSecretKeyName: []byte(MyClusterTokenValue),
			}
		}).
		Return(nil)
}

func TestListClusterResourcesErrorShouldReturnError(t *testing.T) {
	_, err := testCBContainersClusterController(t, func(testMocks *ClusterControllerTestMocks) {
		testMocks.client.EXPECT().List(testMocks.ctx, &cbcontainersv1.CBContainersAgentList{}).Return(fmt.Errorf(""))
	})

	require.Error(t, err)
}

func TestNotFindingAnyClusterResourceShouldReturnNil(t *testing.T) {
	result, err := testCBContainersClusterController(t, func(testMocks *ClusterControllerTestMocks) {
		testMocks.client.EXPECT().List(testMocks.ctx, &cbcontainersv1.CBContainersAgentList{}).Return(nil)
	})

	require.NoError(t, err)
	require.Equal(t, result, ctrlRuntime.Result{})
}

func TestFindingMoreThanOneClusterResourceShouldReturnError(t *testing.T) {
	_, err := testCBContainersClusterController(t, func(testMocks *ClusterControllerTestMocks) {
		testMocks.client.EXPECT().List(testMocks.ctx, &cbcontainersv1.CBContainersAgentList{}).
			Do(func(ctx context.Context, list *cbcontainersv1.CBContainersAgentList, _ ...interface{}) {
				list.Items = append(list.Items, cbcontainersv1.CBContainersAgent{})
				list.Items = append(list.Items, cbcontainersv1.CBContainersAgent{})
			}).
			Return(nil)
	})

	require.Error(t, err)
}

func TestGetTokenSecretErrorShouldReturnError(t *testing.T) {
	_, err := testCBContainersClusterController(t, setupClusterCustomResource, func(testMocks *ClusterControllerTestMocks) {
		accessTokenSecretNamespacedName := types.NamespacedName{Name: ClusterAccessTokenSecretName, Namespace: commonState.DataPlaneNamespaceName}
		testMocks.client.EXPECT().Get(testMocks.ctx, accessTokenSecretNamespacedName, &corev1.Secret{}).Return(fmt.Errorf(""))
	})

	require.Error(t, err)
}

func TestTokenSecretWithoutTokenValueShouldReturnError(t *testing.T) {
	_, err := testCBContainersClusterController(t, setupClusterCustomResource, func(testMocks *ClusterControllerTestMocks) {
		accessTokenSecretNamespacedName := types.NamespacedName{Name: ClusterAccessTokenSecretName, Namespace: commonState.DataPlaneNamespaceName}
		testMocks.client.EXPECT().Get(testMocks.ctx, accessTokenSecretNamespacedName, &corev1.Secret{}).Return(nil)
	})

	require.Error(t, err)
}

func TestClusterReconcile(t *testing.T) {
	secretValues := &models.RegistrySecretValues{Data: map[string][]byte{test_utils.RandomString(): {}}}

	t.Run("When processor returns error, reconcile should return error", func(t *testing.T) {
		_, err := testCBContainersClusterController(t, setupClusterCustomResource, setUpTokenSecretValues, func(testMocks *ClusterControllerTestMocks) {
			testMocks.ClusterProcessor.EXPECT().Process(&ClusterCustomResourceItems[0], MyClusterTokenValue).Return(nil, fmt.Errorf(""))
		})

		require.Error(t, err)
	})

	t.Run("When state applier returns error, reconcile should return error", func(t *testing.T) {
		_, err := testCBContainersClusterController(t, setupClusterCustomResource, setUpTokenSecretValues, func(testMocks *ClusterControllerTestMocks) {
			testMocks.ClusterProcessor.EXPECT().Process(&ClusterCustomResourceItems[0], MyClusterTokenValue).Return(secretValues, nil)
			testMocks.StateApplier.EXPECT().ApplyDesiredState(testMocks.ctx, &ClusterCustomResourceItems[0].Spec, secretValues, gomock.Any()).Return(false, fmt.Errorf(""))
		})

		require.Error(t, err)
	})

	t.Run("When state applier returns state was changed, reconcile should return Requeue true", func(t *testing.T) {
		result, err := testCBContainersClusterController(t, setupClusterCustomResource, setUpTokenSecretValues, func(testMocks *ClusterControllerTestMocks) {
			testMocks.ClusterProcessor.EXPECT().Process(&ClusterCustomResourceItems[0], MyClusterTokenValue).Return(secretValues, nil)
			testMocks.StateApplier.EXPECT().ApplyDesiredState(testMocks.ctx, &ClusterCustomResourceItems[0].Spec, secretValues, gomock.Any()).Return(true, nil)
		})

		require.NoError(t, err)
		require.Equal(t, result, ctrlRuntime.Result{Requeue: true})
	})

	t.Run("When state applier returns state was not changed, reconcile should return default Requeue", func(t *testing.T) {
		result, err := testCBContainersClusterController(t, setupClusterCustomResource, setUpTokenSecretValues, func(testMocks *ClusterControllerTestMocks) {
			testMocks.ClusterProcessor.EXPECT().Process(&ClusterCustomResourceItems[0], MyClusterTokenValue).Return(secretValues, nil)
			testMocks.StateApplier.EXPECT().ApplyDesiredState(testMocks.ctx, &ClusterCustomResourceItems[0].Spec, secretValues, gomock.Any()).Return(false, nil)
		})

		require.NoError(t, err)
		require.Equal(t, result, ctrlRuntime.Result{})
	})
}
