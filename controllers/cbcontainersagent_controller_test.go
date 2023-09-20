package controllers_test

import (
	"context"
	"fmt"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
	"testing"
	"time"

	logrTesting "github.com/go-logr/logr/testr"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/cbcontainers/test_utils"
	testUtilsMocks "github.com/vmware/cbcontainers-operator/cbcontainers/test_utils/mocks"
	"github.com/vmware/cbcontainers-operator/controllers"
	"github.com/vmware/cbcontainers-operator/controllers/mocks"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlRuntime "sigs.k8s.io/controller-runtime"
)

type SetupClusterControllerTest func(*ClusterControllerTestMocks)

type ClusterControllerTestMocks struct {
	client              *testUtilsMocks.MockClient
	statusWriter        *testUtilsMocks.MockStatusWriter
	accessTokenProvider *mocks.MockAccessTokenProvider
	mockAgentProcessor  *mocks.MockAgentProcessor
	stateApplier        *mocks.MockStateApplier
	ctx                 context.Context
}

const (
	MyClusterTokenValue = "my-token-value"
	// agentNamespace helps validate we don't depend on a hardcoded namespace anywhere
	agentNamespace = "dummy-namespace"
)

var (
	ClusterAccessTokenSecretName = test_utils.RandomString()

	true_ = true

	ClusterCustomResourceItems = []cbcontainersv1.CBContainersAgent{
		{
			Spec: cbcontainersv1.CBContainersAgentSpec{
				Version:               "21.7.0",
				AccessTokenSecretName: ClusterAccessTokenSecretName,
				Components: cbcontainersv1.CBContainersComponentsSpec{
					Settings: cbcontainersv1.CBContainersComponentsSettings{
						CreateDefaultImagePullSecrets: &true_,
					},
				},
			},
		},
	}
)

func testCBContainersClusterController(t *testing.T, setups ...SetupClusterControllerTest) (ctrlRuntime.Result, error) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStatusWriter := testUtilsMocks.NewMockStatusWriter(ctrl)
	mockK8SClient := testUtilsMocks.NewMockClient(ctrl)
	mockK8SClient.EXPECT().Status().Return(mockStatusWriter).AnyTimes()

	mocksObjects := &ClusterControllerTestMocks{
		ctx:                 context.TODO(),
		client:              mockK8SClient,
		statusWriter:        mockStatusWriter,
		accessTokenProvider: mocks.NewMockAccessTokenProvider(ctrl),
		mockAgentProcessor:  mocks.NewMockAgentProcessor(ctrl),
		stateApplier:        mocks.NewMockStateApplier(ctrl),
	}

	for _, setup := range setups {
		setup(mocksObjects)
	}

	controller := &controllers.CBContainersAgentController{
		Client:    mocksObjects.client,
		Log:       logrTesting.New(t),
		Scheme:    &runtime.Scheme{},
		Namespace: agentNamespace,

		AccessTokenProvider: mocksObjects.accessTokenProvider,
		ClusterProcessor:    mocksObjects.mockAgentProcessor,
		StateApplier:        mocksObjects.stateApplier,
	}

	return controller.Reconcile(mocksObjects.ctx, ctrlRuntime.Request{})
}

func setupClusterCustomResource(items ...cbcontainersv1.CBContainersAgent) SetupClusterControllerTest {
	if len(items) == 0 {
		items = make([]cbcontainersv1.CBContainersAgent, len(ClusterCustomResourceItems))
		copy(items, ClusterCustomResourceItems)
	}

	return func(testMocks *ClusterControllerTestMocks) {
		testMocks.client.EXPECT().List(testMocks.ctx, &cbcontainersv1.CBContainersAgentList{}).
			Do(func(ctx context.Context, list *cbcontainersv1.CBContainersAgentList, _ ...interface{}) {
				list.Items = items
			}).
			Return(nil)
	}
}

func setUpAccessToken(testMocks *ClusterControllerTestMocks) {
	testMocks.accessTokenProvider.
		EXPECT().
		GetCBAccessToken(testMocks.ctx, gomock.AssignableToTypeOf(&cbcontainersv1.CBContainersAgent{}), agentNamespace).
		Return(MyClusterTokenValue, nil)
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
	_, err := testCBContainersClusterController(t, setupClusterCustomResource(), func(testMocks *ClusterControllerTestMocks) {
		testMocks.accessTokenProvider.
			EXPECT().
			GetCBAccessToken(testMocks.ctx, gomock.AssignableToTypeOf(&cbcontainersv1.CBContainersAgent{}), agentNamespace).
			Return("", fmt.Errorf("some error"))
	})

	require.Error(t, err)
}

func TestTokenSecretWithoutTokenValueShouldReturnError(t *testing.T) {
	_, err := testCBContainersClusterController(t, setupClusterCustomResource(), func(testMocks *ClusterControllerTestMocks) {
		testMocks.accessTokenProvider.
			EXPECT().
			GetCBAccessToken(testMocks.ctx, gomock.AssignableToTypeOf(&cbcontainersv1.CBContainersAgent{}), agentNamespace).
			Return("", nil)
	})

	require.Error(t, err)
}

func TestClusterReconcile(t *testing.T) {
	secretValues := &models.RegistrySecretValues{Data: map[string][]byte{test_utils.RandomString(): {}}}

	t.Run("When processor returns error, reconcile should return error", func(t *testing.T) {
		_, err := testCBContainersClusterController(t, setupClusterCustomResource(), setUpAccessToken, func(testMocks *ClusterControllerTestMocks) {
			testMocks.mockAgentProcessor.EXPECT().Process(MatchAgentResource(&ClusterCustomResourceItems[0]), MyClusterTokenValue).Return(nil, fmt.Errorf(""))
		})

		require.Error(t, err)
	})

	t.Run("When state applier returns error, reconcile should return error", func(t *testing.T) {
		_, err := testCBContainersClusterController(t, setupClusterCustomResource(), setUpAccessToken, func(testMocks *ClusterControllerTestMocks) {
			testMocks.mockAgentProcessor.EXPECT().Process(MatchAgentResource(&ClusterCustomResourceItems[0]), MyClusterTokenValue).Return(secretValues, nil)
			testMocks.stateApplier.EXPECT().ApplyDesiredState(testMocks.ctx, MatchAgentSpec(&ClusterCustomResourceItems[0].Spec), secretValues, gomock.Any()).Return(false, fmt.Errorf(""))
		})

		require.Error(t, err)
	})

	t.Run("When state applier returns state was changed, reconcile should return Requeue true", func(t *testing.T) {
		result, err := testCBContainersClusterController(t, setupClusterCustomResource(), setUpAccessToken, func(testMocks *ClusterControllerTestMocks) {
			testMocks.mockAgentProcessor.EXPECT().Process(MatchAgentResource(&ClusterCustomResourceItems[0]), MyClusterTokenValue).Return(secretValues, nil)
			testMocks.stateApplier.EXPECT().ApplyDesiredState(testMocks.ctx, MatchAgentSpec(&ClusterCustomResourceItems[0].Spec), secretValues, gomock.Any()).Return(true, nil)
		})

		require.NoError(t, err)
		require.Equal(t, result, ctrlRuntime.Result{Requeue: true})
	})

	t.Run("When state applier returns state was not changed, reconcile should return default Requeue", func(t *testing.T) {
		result, err := testCBContainersClusterController(t, setupClusterCustomResource(), setUpAccessToken, func(testMocks *ClusterControllerTestMocks) {
			testMocks.mockAgentProcessor.EXPECT().Process(MatchAgentResource(&ClusterCustomResourceItems[0]), MyClusterTokenValue).Return(secretValues, nil)
			testMocks.stateApplier.EXPECT().ApplyDesiredState(testMocks.ctx, MatchAgentSpec(&ClusterCustomResourceItems[0].Spec), secretValues, gomock.Any()).Return(false, nil)
		})

		require.NoError(t, err)
		require.Equal(t, result, ctrlRuntime.Result{})
	})
}

func TestStatusUpdates(t *testing.T) {
	secretValues := &models.RegistrySecretValues{Data: map[string][]byte{test_utils.RandomString(): {}}}

	t.Run("When state has changed, the ObservedGeneration should not be updated", func(t *testing.T) {
		resourceWithStatus := ClusterCustomResourceItems[0]
		resourceWithStatus.ObjectMeta.Generation = 2
		resourceWithStatus.Status.ObservedGeneration = 1

		result, err := testCBContainersClusterController(t, setupClusterCustomResource(resourceWithStatus), setUpAccessToken, func(testMocks *ClusterControllerTestMocks) {
			testMocks.mockAgentProcessor.EXPECT().Process(MatchAgentResource(&resourceWithStatus), MyClusterTokenValue).Return(secretValues, nil)
			testMocks.stateApplier.EXPECT().ApplyDesiredState(testMocks.ctx, MatchAgentSpec(&resourceWithStatus.Spec), secretValues, gomock.Any()).Return(true, nil)
			testMocks.statusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
		})

		require.NoError(t, err)
		require.Equal(t, result, ctrlRuntime.Result{Requeue: true})
	})

	t.Run("when state has not changed but the CR and status generations are the same, ObservedGeneration should not be updated", func(t *testing.T) {
		resourceWithStatus := ClusterCustomResourceItems[0]
		resourceWithStatus.ObjectMeta.Generation = 1
		resourceWithStatus.Status.ObservedGeneration = 1

		result, err := testCBContainersClusterController(t, setupClusterCustomResource(resourceWithStatus), setUpAccessToken, func(testMocks *ClusterControllerTestMocks) {
			testMocks.mockAgentProcessor.EXPECT().Process(MatchAgentResource(&resourceWithStatus), MyClusterTokenValue).Return(secretValues, nil)
			testMocks.stateApplier.EXPECT().ApplyDesiredState(testMocks.ctx, MatchAgentSpec(&resourceWithStatus.Spec), secretValues, gomock.Any()).Return(false, nil)
			testMocks.statusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
		})

		require.NoError(t, err)
		require.Equal(t, result, ctrlRuntime.Result{})
	})

	t.Run("When state has not changed and the CR and status generations differ, ObservedGeneration should be updated", func(t *testing.T) {
		resourceBeforeReconcile := ClusterCustomResourceItems[0]
		resourceBeforeReconcile.ObjectMeta.Generation = 2
		resourceBeforeReconcile.Status.ObservedGeneration = 1

		expectedResourceWithUpdatedStatus := resourceBeforeReconcile
		expectedResourceWithUpdatedStatus.Status.ObservedGeneration = expectedResourceWithUpdatedStatus.Generation

		result, err := testCBContainersClusterController(t, setupClusterCustomResource(resourceBeforeReconcile), setUpAccessToken, func(testMocks *ClusterControllerTestMocks) {
			testMocks.mockAgentProcessor.EXPECT().Process(MatchAgentResource(&resourceBeforeReconcile), MyClusterTokenValue).Return(secretValues, nil)
			testMocks.stateApplier.EXPECT().ApplyDesiredState(testMocks.ctx, MatchAgentSpec(&resourceBeforeReconcile.Spec), secretValues, gomock.Any()).Return(false, nil)
			testMocks.statusWriter.EXPECT().Update(testMocks.ctx, MatchAgentResource(&expectedResourceWithUpdatedStatus), gomock.Any()).Times(1).Return(nil)
		})

		require.NoError(t, err)
		require.Equal(t, result, ctrlRuntime.Result{})
	})

	t.Run("When updating status and getting a conflict, a requeue should be scheduled", func(t *testing.T) {
		resourceBeforeReconcile := ClusterCustomResourceItems[0]
		resourceBeforeReconcile.ObjectMeta.Generation = 2
		resourceBeforeReconcile.Status.ObservedGeneration = 1

		expectedResourceWithUpdatedStatus := resourceBeforeReconcile
		expectedResourceWithUpdatedStatus.Status.ObservedGeneration = expectedResourceWithUpdatedStatus.Generation

		result, err := testCBContainersClusterController(t, setupClusterCustomResource(resourceBeforeReconcile), setUpAccessToken, func(testMocks *ClusterControllerTestMocks) {
			testMocks.mockAgentProcessor.EXPECT().Process(MatchAgentResource(&resourceBeforeReconcile), MyClusterTokenValue).Return(secretValues, nil)
			testMocks.stateApplier.EXPECT().ApplyDesiredState(testMocks.ctx, MatchAgentSpec(&resourceBeforeReconcile.Spec), secretValues, gomock.Any()).Return(false, nil)
			testMocks.statusWriter.EXPECT().Update(testMocks.ctx, MatchAgentResource(&expectedResourceWithUpdatedStatus), gomock.Any()).Return(k8sErrors.NewConflict(schema.GroupResource{}, "conflict", nil))
		})

		require.NoError(t, err)
		require.Greater(t, result.RequeueAfter, time.Duration(0))
	})

	t.Run("When updating status and getting error that is not conflict, a error should be returned", func(t *testing.T) {
		resourceBeforeReconcile := ClusterCustomResourceItems[0]
		resourceBeforeReconcile.ObjectMeta.Generation = 2
		resourceBeforeReconcile.Status.ObservedGeneration = 1

		expectedResourceWithUpdatedStatus := resourceBeforeReconcile
		expectedResourceWithUpdatedStatus.Status.ObservedGeneration = expectedResourceWithUpdatedStatus.Generation

		result, err := testCBContainersClusterController(t, setupClusterCustomResource(resourceBeforeReconcile), setUpAccessToken, func(testMocks *ClusterControllerTestMocks) {
			testMocks.mockAgentProcessor.EXPECT().Process(MatchAgentResource(&resourceBeforeReconcile), MyClusterTokenValue).Return(secretValues, nil)
			testMocks.stateApplier.EXPECT().ApplyDesiredState(testMocks.ctx, MatchAgentSpec(&resourceBeforeReconcile.Spec), secretValues, gomock.Any()).Return(false, nil)
			testMocks.statusWriter.EXPECT().Update(testMocks.ctx, MatchAgentResource(&expectedResourceWithUpdatedStatus), gomock.Any()).Return(fmt.Errorf("some error"))
		})

		require.Error(t, err)
		require.Equal(t, result, ctrlRuntime.Result{})
	})
}

// partialCBContainersAgentMatcher matches a given cbcontainersv1.CBContainersAgent parameter based on some fields only
// this should be used when the object returned to the controller and the one passed to the controller's processor differ due to default values being set
// so any fields that have defaults should _not_ be compared in this matcher, the rest can be added if it makes sense
type partialCBContainersAgentMatcher struct {
	expected *cbcontainersv1.CBContainersAgent
}

func (p *partialCBContainersAgentMatcher) Matches(x interface{}) bool {
	actual, ok := x.(*cbcontainersv1.CBContainersAgent)
	if !ok {
		return false
	}

	return p.expected.Spec.Version == actual.Spec.Version &&
		p.expected.Spec.ClusterName == actual.Spec.ClusterName &&
		p.expected.Spec.Account == actual.Spec.Account &&
		reflect.DeepEqual(p.expected.ObjectMeta, actual.ObjectMeta) &&
		p.expected.Status == actual.Status
}

func (p *partialCBContainersAgentMatcher) String() string {
	return fmt.Sprintf("matches a CBContainers CR to (%v) but only looks at interesting fields", p.expected)
}

// partialCBContainersSpecMatcher is the same as partialCBContainersSpecMatcher but for the spec only
type partialCBContainersSpecMatcher struct {
	expected *cbcontainersv1.CBContainersAgentSpec
}

func (p *partialCBContainersSpecMatcher) Matches(x interface{}) bool {
	actual, ok := x.(*cbcontainersv1.CBContainersAgentSpec)
	if !ok {
		return false
	}

	return p.expected.Version == actual.Version &&
		p.expected.ClusterName == actual.ClusterName &&
		p.expected.Account == actual.Account
}

func (p *partialCBContainersSpecMatcher) String() string {
	return fmt.Sprintf("matches a CBContainers spec to (%v) but only looks at interesting fields", p.expected)
}

func MatchAgentResource(expected *cbcontainersv1.CBContainersAgent) gomock.Matcher {
	return &partialCBContainersAgentMatcher{expected: expected}
}

func MatchAgentSpec(expected *cbcontainersv1.CBContainersAgentSpec) gomock.Matcher {
	return &partialCBContainersSpecMatcher{expected: expected}
}
