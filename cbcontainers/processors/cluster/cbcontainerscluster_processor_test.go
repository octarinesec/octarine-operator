package cluster_test

import (
	"fmt"
	logrTesting "github.com/go-logr/logr/testing"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	"github.com/vmware/cbcontainers-operator/cbcontainers/processors/cluster"
	"github.com/vmware/cbcontainers-operator/cbcontainers/processors/cluster/mocks"
	"github.com/vmware/cbcontainers-operator/cbcontainers/test_utils"
	"testing"
)

type ClusterProcessorTestMocks struct {
	gatewayMock        *mocks.MockGateway
	gatewayCreatorMock *mocks.MockGatewayCreator
	monitorMock        *mocks.MockMonitor
	monitorCreatorMock *mocks.MockMonitorCreator
}

type SetupAndAssertClusterProcessorTest func(*ClusterProcessorTestMocks, *cluster.CBContainerClusterProcessor)

var (
	AccessToken = test_utils.RandomString()
)

func testClusterProcessor(t *testing.T, setupAndAssert SetupAndAssertClusterProcessorTest) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mocksObjects := &ClusterProcessorTestMocks{
		gatewayMock:        mocks.NewMockGateway(ctrl),
		gatewayCreatorMock: mocks.NewMockGatewayCreator(ctrl),
		monitorMock:        mocks.NewMockMonitor(ctrl),
		monitorCreatorMock: mocks.NewMockMonitorCreator(ctrl),
	}

	processor := cluster.NewCBContainerClusterProcessor(&logrTesting.TestLogger{T: t}, mocksObjects.gatewayCreatorMock, mocksObjects.monitorCreatorMock)
	setupAndAssert(mocksObjects, processor)
}

func setupValidMocksCalls(testMocks *ClusterProcessorTestMocks, times int) {
	testMocks.gatewayCreatorMock.EXPECT().CreateGateway(gomock.Any(), AccessToken).Return(testMocks.gatewayMock).Times(times)
	testMocks.monitorCreatorMock.EXPECT().CreateMonitor(gomock.Any(), testMocks.gatewayMock).Return(testMocks.monitorMock, nil).Times(times)
	testMocks.gatewayMock.EXPECT().GetRegistrySecret().DoAndReturn(func() (*models.RegistrySecretValues, error) {
		return &models.RegistrySecretValues{Data: map[string][]byte{test_utils.RandomString(): {}}}, nil
	}).Times(times)
	testMocks.gatewayMock.EXPECT().RegisterCluster().Return(nil).Times(times)
	testMocks.monitorMock.EXPECT().Start().Times(times)
}

func TestProcessorIsNotRecreatingComponentsForSameCR(t *testing.T) {
	testClusterProcessor(t, func(testMocks *ClusterProcessorTestMocks, processor *cluster.CBContainerClusterProcessor) {
		clusterCR := &cbcontainersv1.CBContainersCluster{Spec: cbcontainersv1.CBContainersClusterSpec{Account: test_utils.RandomString(), ClusterName: test_utils.RandomString()}}
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
	testClusterProcessor(t, func(testMocks *ClusterProcessorTestMocks, processor *cluster.CBContainerClusterProcessor) {
		clusterCR1 := &cbcontainersv1.CBContainersCluster{Spec: cbcontainersv1.CBContainersClusterSpec{Account: test_utils.RandomString(), ClusterName: test_utils.RandomString()}}
		clusterCR2 := &cbcontainersv1.CBContainersCluster{Spec: cbcontainersv1.CBContainersClusterSpec{Account: test_utils.RandomString(), ClusterName: test_utils.RandomString()}}
		setupValidMocksCalls(testMocks, 2)
		testMocks.monitorMock.EXPECT().Stop().Times(1)

		values1, err1 := processor.Process(clusterCR1, AccessToken)
		values2, err2 := processor.Process(clusterCR2, AccessToken)

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NotNil(t, values1)
		require.NotNil(t, values2)
		require.NotEqual(t, values1, values2)
	})
}

func TestProcessorReturnsErrorWhenCanNotCreateMonitor(t *testing.T) {
	testClusterProcessor(t, func(testMocks *ClusterProcessorTestMocks, processor *cluster.CBContainerClusterProcessor) {
		clusterCR := &cbcontainersv1.CBContainersCluster{Spec: cbcontainersv1.CBContainersClusterSpec{Account: test_utils.RandomString(), ClusterName: test_utils.RandomString()}}
		testMocks.gatewayCreatorMock.EXPECT().CreateGateway(gomock.Any(), gomock.Any()).Return(testMocks.gatewayMock)
		testMocks.monitorCreatorMock.EXPECT().CreateMonitor(gomock.Any(), gomock.Any()).Return(testMocks.monitorMock, fmt.Errorf(""))
		_, err := processor.Process(clusterCR, AccessToken)
		require.Error(t, err)
	})
}

func TestProcessorReturnsErrorWhenCanNotGetRegisterySecret(t *testing.T) {
	testClusterProcessor(t, func(testMocks *ClusterProcessorTestMocks, processor *cluster.CBContainerClusterProcessor) {
		clusterCR := &cbcontainersv1.CBContainersCluster{Spec: cbcontainersv1.CBContainersClusterSpec{Account: test_utils.RandomString(), ClusterName: test_utils.RandomString()}}
		testMocks.gatewayCreatorMock.EXPECT().CreateGateway(gomock.Any(), gomock.Any()).Return(testMocks.gatewayMock)
		testMocks.monitorCreatorMock.EXPECT().CreateMonitor(gomock.Any(), gomock.Any()).Return(testMocks.monitorMock, nil)
		testMocks.gatewayMock.EXPECT().GetRegistrySecret().Return(nil, fmt.Errorf(""))
		_, err := processor.Process(clusterCR, AccessToken)
		require.Error(t, err)
	})
}

func TestProcessorReturnsErrorWhenCanNotRegisterCluster(t *testing.T) {
	testClusterProcessor(t, func(testMocks *ClusterProcessorTestMocks, processor *cluster.CBContainerClusterProcessor) {
		clusterCR := &cbcontainersv1.CBContainersCluster{Spec: cbcontainersv1.CBContainersClusterSpec{Account: test_utils.RandomString(), ClusterName: test_utils.RandomString()}}
		testMocks.gatewayCreatorMock.EXPECT().CreateGateway(gomock.Any(), gomock.Any()).Return(testMocks.gatewayMock)
		testMocks.monitorCreatorMock.EXPECT().CreateMonitor(gomock.Any(), gomock.Any()).Return(testMocks.monitorMock, nil)
		testMocks.gatewayMock.EXPECT().GetRegistrySecret().Return(&models.RegistrySecretValues{}, nil)
		testMocks.gatewayMock.EXPECT().RegisterCluster().Return(fmt.Errorf(""))
		_, err := processor.Process(clusterCR, AccessToken)
		require.Error(t, err)
	})
}
