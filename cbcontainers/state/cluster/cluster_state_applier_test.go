package cluster_test

import (
	"context"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/cluster/objects"
	appsV1 "k8s.io/api/apps/v1"
	"reflect"
	"testing"

	logrTesting "github.com/go-logr/logr/testing"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/cluster"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/cluster/mocks"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	"github.com/vmware/cbcontainers-operator/cbcontainers/test_utils"
	testUtilsMocks "github.com/vmware/cbcontainers-operator/cbcontainers/test_utils/mocks"
	coreV1 "k8s.io/api/core/v1"
	schedulingV1 "k8s.io/api/scheduling/v1"
	schedulingV1alpha1 "k8s.io/api/scheduling/v1alpha1"
	schedulingV1beta1 "k8s.io/api/scheduling/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultKubeletVersion = "v1.20.2"

	NumberOfExpectedAppliedObjects = 4
)

var (
	Account           = test_utils.RandomString()
	Cluster           = test_utils.RandomString()
	ApiGateWayScheme  = test_utils.RandomString()
	ApiGateWayHost    = test_utils.RandomString()
	ApiGateWayPort    = 4206
	ApiGateWayAdapter = test_utils.RandomString()
)

type ClusterStateApplierTestMocks struct {
	client              *testUtilsMocks.MockClient
	childApplier        *mocks.MockClusterChildK8sObjectApplier
	cbContainersCluster *cbcontainersv1.CBContainersAgent
	kubeletVersion      string
}

type ClusterStateApplierTestSetup func(*ClusterStateApplierTestMocks)

type AppliedObject struct {
	Namespace  string
	Name       string
	ObjectType reflect.Type
}

func testClusterStateApplier(t *testing.T, setup ClusterStateApplierTestSetup, k8sVersion string) (bool, error) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cbContainersCluster := &cbcontainersv1.CBContainersAgent{
		Spec: cbcontainersv1.CBContainersAgentSpec{
			Account:     Account,
			ClusterName: Cluster,
			ApiGatewaySpec: cbcontainersv1.CBContainersApiGatewaySpec{
				Scheme:  ApiGateWayScheme,
				Host:    ApiGateWayHost,
				Port:    ApiGateWayPort,
				Adapter: ApiGateWayAdapter,
			},
		},
	}

	if k8sVersion == "" {
		k8sVersion = DefaultKubeletVersion
	}

	mockObjects := &ClusterStateApplierTestMocks{
		client:              testUtilsMocks.NewMockClient(ctrl),
		childApplier:        mocks.NewMockClusterChildK8sObjectApplier(ctrl),
		cbContainersCluster: cbContainersCluster,
		kubeletVersion:      k8sVersion,
	}

	setup(mockObjects)

	return cluster.NewClusterStateApplier(&logrTesting.TestLogger{T: t}, k8sVersion, mockObjects.childApplier).ApplyDesiredState(context.Background(), cbContainersCluster, nil, mockObjects.client, nil)
}

func getAppliedObjects(t *testing.T, k8sVersion string) []AppliedObject {
	actualAppliedObjects := make([]AppliedObject, 0)

	_, err := testClusterStateApplier(t, func(mocks *ClusterStateApplierTestMocks) {
		mocks.childApplier.EXPECT().ApplyClusterChildK8sObject(gomock.Any(), mocks.cbContainersCluster, mocks.client, gomock.Any(), gomock.Any()).
			Do(func(ctx context.Context, cr *cbcontainersv1.CBContainersAgent, client client.Client, obj cluster.ClusterChildK8sObject, options ...*options.ApplyOptions) {
				namespacedName := obj.ClusterChildNamespacedName(cr)
				objType := reflect.TypeOf(obj.EmptyK8sObject())
				actualAppliedObjects = append(actualAppliedObjects, AppliedObject{Namespace: namespacedName.Namespace, Name: namespacedName.Name, ObjectType: objType})
			}).
			Return(true, nil, nil).AnyTimes()
	}, k8sVersion)

	require.NoError(t, err)
	require.Len(t, actualAppliedObjects, NumberOfExpectedAppliedObjects)
	return actualAppliedObjects
}

func TestConfigMapIsApplied(t *testing.T) {
	appliedObjects := getAppliedObjects(t, "")
	require.Contains(t, appliedObjects, AppliedObject{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       commonState.DataPlaneConfigmapName,
		ObjectType: reflect.TypeOf(&coreV1.ConfigMap{}),
	})
}

func TestSecretIsApplied(t *testing.T) {
	appliedObjects := getAppliedObjects(t, "")
	require.Contains(t, appliedObjects, AppliedObject{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       commonState.RegistrySecretName,
		ObjectType: reflect.TypeOf(&coreV1.Secret{}),
	})
}

func TestPriorityClassIsApplied(t *testing.T) {
	testPriorityClassIsApplied := func(t *testing.T, objType reflect.Type, k8sVersion string) {
		appliedObjects := getAppliedObjects(t, k8sVersion)
		require.Contains(t, appliedObjects, AppliedObject{
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
	appliedObjects := getAppliedObjects(t, "")
	require.Contains(t, appliedObjects, AppliedObject{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       objects.MonitorName,
		ObjectType: reflect.TypeOf(&appsV1.Deployment{}),
	})
}
