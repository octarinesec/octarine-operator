package processors_test

import (
	"fmt"
	"testing"

	logrTesting "github.com/go-logr/logr/testing"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/cbcontainers/processors"
	"github.com/vmware/cbcontainers-operator/cbcontainers/processors/mocks"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/operator"
	"github.com/vmware/cbcontainers-operator/cbcontainers/test_utils"
)

type ClusterProcessorTestMocks struct {
	gatewayMock                 *mocks.MockAPIGateway
	gatewayCreatorMock          *mocks.MockAPIGatewayCreator
	operatorVersionProviderMock *mocks.MockOperatorVersionProvider
}

type SetupAndAssertClusterProcessorTest func(*ClusterProcessorTestMocks, *processors.AgentProcessor)

var (
	AccessToken = test_utils.RandomString()
)

func testClusterProcessor(t *testing.T, setupAndAssert SetupAndAssertClusterProcessorTest) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mocksObjects := &ClusterProcessorTestMocks{
		gatewayMock:                 mocks.NewMockAPIGateway(ctrl),
		gatewayCreatorMock:          mocks.NewMockAPIGatewayCreator(ctrl),
		operatorVersionProviderMock: mocks.NewMockOperatorVersionProvider(ctrl),
	}

	processor := processors.NewAgentProcessor(logrTesting.NewTestLogger(t), mocksObjects.gatewayCreatorMock, mocksObjects.operatorVersionProviderMock, "mockIdentifier")
	setupAndAssert(mocksObjects, processor)
}

func setupValidMocksCalls(testMocks *ClusterProcessorTestMocks, times int) {
	testMocks.gatewayCreatorMock.EXPECT().CreateGateway(gomock.Any(), AccessToken).Return(testMocks.gatewayMock, nil).Times(times)
	testMocks.gatewayMock.EXPECT().GetRegistrySecret().DoAndReturn(func() (*models.RegistrySecretValues, error) {
		return &models.RegistrySecretValues{Data: map[string][]byte{test_utils.RandomString(): {}}}, nil
	}).Times(times)
	testMocks.gatewayMock.EXPECT().RegisterCluster().Return(nil).Times(times)
	// this will skip the compatibility check
	// for all tests that do not explicitly test that
	testMocks.operatorVersionProviderMock.EXPECT().GetOperatorVersion().Return("", operator.ErrNotSemVer).AnyTimes()
}

func TestProcessorIsNotRecreatingComponentsForSameCR(t *testing.T) {
	testClusterProcessor(t, func(testMocks *ClusterProcessorTestMocks, processor *processors.AgentProcessor) {
		clusterCR := &cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Account: test_utils.RandomString(), ClusterName: test_utils.RandomString()}}
		setupValidMocksCalls(testMocks, 1)

		values1, err1 := processor.Process(clusterCR, AccessToken)
		values2, err2 := processor.Process(clusterCR, AccessToken)

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NotNil(t, values1)
		require.Equal(t, values1, values2)
	})
}

func TestProcessorIsReCreatingComponentsForDifferentCR(t *testing.T) {
	testClusterProcessor(t, func(testMocks *ClusterProcessorTestMocks, processor *processors.AgentProcessor) {
		clusterCR1 := &cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Account: test_utils.RandomString(), ClusterName: test_utils.RandomString()}}
		clusterCR2 := &cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Account: test_utils.RandomString(), ClusterName: test_utils.RandomString()}}
		setupValidMocksCalls(testMocks, 2)

		values1, err1 := processor.Process(clusterCR1, AccessToken)
		values2, err2 := processor.Process(clusterCR2, AccessToken)

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NotNil(t, values1)
		require.NotNil(t, values2)
		require.NotEqual(t, values1, values2)
	})
}

func TestProcessorReturnsErrorWhenCanNotGetRegisterySecret(t *testing.T) {
	testClusterProcessor(t, func(testMocks *ClusterProcessorTestMocks, processor *processors.AgentProcessor) {
		clusterCR := &cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Account: test_utils.RandomString(), ClusterName: test_utils.RandomString()}}
		testMocks.gatewayCreatorMock.EXPECT().CreateGateway(gomock.Any(), gomock.Any()).Return(testMocks.gatewayMock, nil)
		testMocks.gatewayMock.EXPECT().GetRegistrySecret().Return(nil, fmt.Errorf(""))
		_, err := processor.Process(clusterCR, AccessToken)
		require.Error(t, err)
	})
}

func TestProcessorReturnsErrorWhenCanNotRegisterCluster(t *testing.T) {
	testClusterProcessor(t, func(testMocks *ClusterProcessorTestMocks, processor *processors.AgentProcessor) {
		clusterCR := &cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Account: test_utils.RandomString(), ClusterName: test_utils.RandomString()}}
		testMocks.gatewayCreatorMock.EXPECT().CreateGateway(gomock.Any(), gomock.Any()).Return(testMocks.gatewayMock, nil)
		testMocks.gatewayMock.EXPECT().GetRegistrySecret().Return(&models.RegistrySecretValues{}, nil)
		testMocks.gatewayMock.EXPECT().RegisterCluster().Return(fmt.Errorf(""))
		_, err := processor.Process(clusterCR, AccessToken)
		require.Error(t, err)
	})
}

func TestProcessorReturnsErrorWhenOperatorVersionProviderReturnsUnknownError(t *testing.T) {
	testClusterProcessor(t, func(testMocks *ClusterProcessorTestMocks, processor *processors.AgentProcessor) {
		clusterCR := &cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Account: test_utils.RandomString(), ClusterName: test_utils.RandomString()}}
		testMocks.gatewayCreatorMock.EXPECT().CreateGateway(gomock.Any(), gomock.Any()).Return(testMocks.gatewayMock, nil)
		testMocks.gatewayMock.EXPECT().GetRegistrySecret().Return(&models.RegistrySecretValues{}, nil)
		testMocks.gatewayMock.EXPECT().RegisterCluster().Return(nil)
		testMocks.operatorVersionProviderMock.EXPECT().GetOperatorVersion().Return("", fmt.Errorf("intentional unknown error"))
		_, err := processor.Process(clusterCR, AccessToken)
		require.Error(t, err)
	})
}

func TestProcessorReturnsErrorWhenCanNotCreateGateway(t *testing.T) {
	testClusterProcessor(t, func(testMocks *ClusterProcessorTestMocks, processor *processors.AgentProcessor) {
		clusterCR := &cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Account: test_utils.RandomString(), ClusterName: test_utils.RandomString()}}
		testMocks.gatewayCreatorMock.EXPECT().CreateGateway(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf(""))
		_, err := processor.Process(clusterCR, AccessToken)
		require.Error(t, err)
	})
}

func TestCheckCompatibilityCompatibleVersions(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(*ClusterProcessorTestMocks)
	}{
		{
			name: "when GetOperatorVersion returns ErrNotSemVer",
			setup: func(testMocks *ClusterProcessorTestMocks) {
				testMocks.operatorVersionProviderMock.EXPECT().GetOperatorVersion().Return("", operator.ErrNotSemVer)
			},
		},
		{
			name: "when CreateGateway returns error",
			setup: func(testMocks *ClusterProcessorTestMocks) {
				testMocks.operatorVersionProviderMock.EXPECT().GetOperatorVersion().Return("", nil)
				testMocks.gatewayCreatorMock.EXPECT().CreateGateway(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("intentional error"))
			},
		},
		{
			name: "when GetCompatibilityMatrixEntryFor returns error",
			setup: func(testMocks *ClusterProcessorTestMocks) {
				testMocks.operatorVersionProviderMock.EXPECT().GetOperatorVersion().Return("", nil)
				testMocks.gatewayCreatorMock.EXPECT().CreateGateway(gomock.Any(), gomock.Any()).Return(testMocks.gatewayMock, nil)
				testMocks.gatewayMock.EXPECT().GetCompatibilityMatrixEntryFor(gomock.Any()).Return(nil, fmt.Errorf("intentional error"))
			},
		},
		{
			name: "when versions are compatible",
			setup: func(testMocks *ClusterProcessorTestMocks) {
				testMocks.operatorVersionProviderMock.EXPECT().GetOperatorVersion().Return("1.0.0", nil)
				testMocks.gatewayCreatorMock.EXPECT().CreateGateway(gomock.Any(), gomock.Any()).Return(testMocks.gatewayMock, nil)
				testMocks.gatewayMock.EXPECT().GetCompatibilityMatrixEntryFor(gomock.Any()).Return(&models.OperatorCompatibility{
					MinAgent: "0.9.0",
					MaxAgent: "1.1.0",
				}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testClusterProcessor(t, func(testMocks *ClusterProcessorTestMocks, processor *processors.AgentProcessor) {
				clusterCR := &cbcontainersv1.CBContainersAgent{Spec: cbcontainersv1.CBContainersAgentSpec{Version: "1.0.0", Account: test_utils.RandomString(), ClusterName: test_utils.RandomString()}}
				testMocks.gatewayCreatorMock.EXPECT().CreateGateway(gomock.Any(), gomock.Any()).Return(testMocks.gatewayMock, nil)
				testMocks.gatewayMock.EXPECT().GetRegistrySecret().Return(&models.RegistrySecretValues{}, nil)
				testMocks.gatewayMock.EXPECT().RegisterCluster().Return(nil)
				testCase.setup(testMocks)

				values, err := processor.Process(clusterCR, AccessToken)
				require.NoError(t, err)
				require.NotNil(t, values)
			})
		})
	}
}
