package hardening_test

import (
	"context"
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
	coreV1 "k8s.io/api/core/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

const (
	DefaultKubeletVersion = "v1.20.2"

	NumberOfExpectedAppliedObjects = 3
)

var (
	Version           = test_utils.RandomString()
	EventsGateWayHost = test_utils.RandomString()
)

type HardeningStateApplierTestMocks struct {
	log                   *logrTesting.TestLogger
	client                *testUtilsMocks.MockClient
	secretValuesCreator   *mocks.MockTlsSecretsValuesCreator
	childApplier          *mocks.MockHardeningChildK8sObjectApplier
	cbContainersHardening *cbcontainersv1.CBContainersHardening
	kubeletVersion        string
}

type HardeningStateApplierTestSetup func(*HardeningStateApplierTestMocks)

type K8sObjectDeatails struct {
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
			EventsGatewaySpec: cbcontainersv1.CBContainersHardeningEventsGatewaySpec{
				Host: EventsGateWayHost,
			},
		},
	}

	if k8sVersion == "" {
		k8sVersion = DefaultKubeletVersion
	}

	mockObjects := &HardeningStateApplierTestMocks{
		log:                   &logrTesting.TestLogger{},
		client:                testUtilsMocks.NewMockClient(ctrl),
		secretValuesCreator:   mocks.NewMockTlsSecretsValuesCreator(ctrl),
		childApplier:          mocks.NewMockHardeningChildK8sObjectApplier(ctrl),
		cbContainersHardening: cbContainersHardening,
		kubeletVersion:        k8sVersion,
	}

	mockObjects.client.EXPECT().List(gomock.Any(), gomock.Any()).Do(func(arg0 context.Context, arg1 *coreV1.NodeList, arg2 ...client.ListOption) {
		arg1.Items = append(arg1.Items, coreV1.Node{Status: coreV1.NodeStatus{NodeInfo: coreV1.NodeSystemInfo{KubeletVersion: mockObjects.kubeletVersion}}})
	}).Return(nil)

	setup(mockObjects)

	return hardening.NewHardeningStateApplier(mockObjects.log, mockObjects.secretValuesCreator, mockObjects.childApplier).ApplyDesiredState(context.Background(), cbContainersHardening, mockObjects.client, nil)
}

func getAppliedAndDeletedObjects(t *testing.T, k8sVersion string) ([]K8sObjectDeatails, []K8sObjectDeatails) {
	appliedObjects := make([]K8sObjectDeatails, 0)
	deletedObjects := make([]K8sObjectDeatails, 0)

	_, err := testHardeningStateApplier(t, func(mocks *HardeningStateApplierTestMocks) {
		mocks.childApplier.EXPECT().ApplyHardeningChildK8sObject(gomock.Any(), mocks.cbContainersHardening, mocks.client, gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, cr *cbcontainersv1.CBContainersHardening, client client.Client, obj hardening.HardeningChildK8sObject, options ...*options.ApplyOptions) (bool, client.Object, error) {
				namespacedName := obj.HardeningChildNamespacedName(cr)
				objType := reflect.TypeOf(obj.EmptyK8sObject())
				appliedObjects = append(appliedObjects, K8sObjectDeatails{Namespace: namespacedName.Namespace, Name: namespacedName.Name, ObjectType: objType})
				return true, obj.EmptyK8sObject(), nil
			}).AnyTimes()
		mocks.childApplier.EXPECT().DeleteK8sObjectIfExists(gomock.Any(), mocks.cbContainersHardening, mocks.client, gomock.Any()).
			DoAndReturn(func(ctx context.Context, cr *cbcontainersv1.CBContainersHardening, client client.Client, obj hardening.HardeningChildK8sObject) (bool, error) {
				namespacedName := obj.HardeningChildNamespacedName(cr)
				objType := reflect.TypeOf(obj.EmptyK8sObject())
				deletedObjects = append(deletedObjects, K8sObjectDeatails{Namespace: namespacedName.Namespace, Name: namespacedName.Name, ObjectType: objType})
				return true, nil
			}).AnyTimes()
	}, k8sVersion)

	require.NoError(t, err)
	//require.Len(t, appliedObjects, NumberOfExpectedAppliedObjects)
	return appliedObjects, deletedObjects
}

func TestTlsSecretIsApplied(t *testing.T) {
	appliedObjects, _ := getAppliedAndDeletedObjects(t, "")
	require.Contains(t, appliedObjects, K8sObjectDeatails{
		Namespace:  commonState.DataPlaneNamespaceName,
		Name:       objects.EnforcerTlsName,
		ObjectType: reflect.TypeOf(&coreV1.Secret{}),
	})
}
