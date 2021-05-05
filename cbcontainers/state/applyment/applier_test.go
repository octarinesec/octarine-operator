package applyment

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/mocks"
	applymentOptions "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	"github.com/vmware/cbcontainers-operator/cbcontainers/test_utils"
	testUtilsMocks "github.com/vmware/cbcontainers-operator/cbcontainers/test_utils/mocks"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

var (
	NamespacedName = types.NamespacedName{Namespace: test_utils.RandomString(), Name: test_utils.RandomString()}
	K8sObject      = &FakeTypeK8sObject{Foo: "bar"}
)

type FakeTypeK8sObject struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Foo string
}

func (*FakeTypeK8sObject) GetObjectKind() schema.ObjectKind { return nil }
func (*FakeTypeK8sObject) DeepCopyObject() runtime.Object   { return nil }

type ApplierTestMocks struct {
	client           *testUtilsMocks.MockClient
	desiredK8sObject *mocks.MockDesiredK8sObject
}

type ApplierTestSetup func(*ApplierTestMocks)

func createMocks(ctrl *gomock.Controller, setup ApplierTestSetup) *ApplierTestMocks {
	mockObjects := &ApplierTestMocks{
		client:           testUtilsMocks.NewMockClient(ctrl),
		desiredK8sObject: mocks.NewMockDesiredK8sObject(ctrl),
	}

	mockObjects.desiredK8sObject.EXPECT().NamespacedName().Return(NamespacedName)
	mockObjects.desiredK8sObject.EXPECT().EmptyK8sObject().Return(K8sObject)
	setup(mockObjects)
	return mockObjects
}

func testApplyDesiredK8sObject(t *testing.T, setup ApplierTestSetup, applyOptionsList ...*applymentOptions.ApplyOptions) (bool, client.Object, error) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockObjects := createMocks(ctrl, setup)

	return ApplyDesiredK8sObject(context.Background(), mockObjects.client, mockObjects.desiredK8sObject, applyOptionsList...)
}

func testDeleteK8sObjectIfExists(t *testing.T, setup ApplierTestSetup) (bool, error) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockObjects := createMocks(ctrl, setup)

	return DeleteK8sObjectIfExists(context.Background(), mockObjects.client, mockObjects.desiredK8sObject)
}

func TestApplyFailsWhenGetFails(t *testing.T) {
	_, _, err := testApplyDesiredK8sObject(t, func(mocks *ApplierTestMocks) {
		mocks.client.EXPECT().Get(gomock.Any(), NamespacedName, K8sObject).Return(fmt.Errorf(""))
	})

	require.Error(t, err)
}

func TestObjectIsCreatedWhenNotExisting(t *testing.T) {
	var controlledResourceToSetOwner metav1.Object
	changed, _, err := testApplyDesiredK8sObject(t, func(mocks *ApplierTestMocks) {
		notFoundError := errors.NewNotFound(schema.GroupResource{}, "")
		mocks.client.EXPECT().Get(gomock.Any(), NamespacedName, K8sObject).Return(notFoundError)
		mocks.desiredK8sObject.EXPECT().MutateK8sObject(K8sObject).Return(nil)
		mocks.client.EXPECT().Create(gomock.Any(), K8sObject).Return(nil)
	}, applymentOptions.NewApplyOptions().SetOwnerSetter(func(controlledResource metav1.Object) error {
		controlledResourceToSetOwner = controlledResource
		return nil
	}))

	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, controlledResourceToSetOwner, K8sObject)
	require.Equal(t, K8sObject.Namespace, NamespacedName.Namespace)
	require.Equal(t, K8sObject.Name, NamespacedName.Name)
}

func TestApplyFailsWhenNotExistingAndCreationFails(t *testing.T) {
	_, _, err := testApplyDesiredK8sObject(t, func(mocks *ApplierTestMocks) {
		notFoundError := errors.NewNotFound(schema.GroupResource{}, "")
		mocks.client.EXPECT().Get(gomock.Any(), NamespacedName, K8sObject).Return(notFoundError)
		mocks.desiredK8sObject.EXPECT().MutateK8sObject(K8sObject).Return(nil)
		mocks.client.EXPECT().Create(gomock.Any(), K8sObject).Return(fmt.Errorf(""))
	})

	require.Error(t, err)
}

func TestObjectIsUpdatedWhenExistingAndChanged(t *testing.T) {
	changed, _, err := testApplyDesiredK8sObject(t, func(mocks *ApplierTestMocks) {
		mocks.client.EXPECT().Get(gomock.Any(), NamespacedName, K8sObject).Return(nil)
		mocks.desiredK8sObject.EXPECT().MutateK8sObject(K8sObject).Do(func(object *FakeTypeK8sObject) {
			object.Foo += object.Foo
		}).Return(nil)
		mocks.client.EXPECT().Update(gomock.Any(), K8sObject).Return(nil)
	})

	require.NoError(t, err)
	require.True(t, changed)
}

func TestObjectIsNotUpdatedWhenExistingButNotChanged(t *testing.T) {
	changed, _, err := testApplyDesiredK8sObject(t, func(mocks *ApplierTestMocks) {
		mocks.client.EXPECT().Get(gomock.Any(), NamespacedName, K8sObject).Return(nil)
		mocks.desiredK8sObject.EXPECT().MutateK8sObject(K8sObject).Return(nil)
	})

	require.NoError(t, err)
	require.False(t, changed)
}

func TestObjectIsNotUpdatedWhenExistingAndChangedButWithCreationOnlyFlag(t *testing.T) {
	changed, _, err := testApplyDesiredK8sObject(t, func(mocks *ApplierTestMocks) {
		mocks.client.EXPECT().Get(gomock.Any(), NamespacedName, K8sObject).Return(nil)
	}, applymentOptions.NewApplyOptions().SetCreateOnly(true))

	require.NoError(t, err)
	require.False(t, changed)
}

func TestApplyFailsWhenExistingAndUpdateFails(t *testing.T) {
	_, _, err := testApplyDesiredK8sObject(t, func(mocks *ApplierTestMocks) {
		mocks.client.EXPECT().Get(gomock.Any(), NamespacedName, K8sObject).Return(nil)
		mocks.desiredK8sObject.EXPECT().MutateK8sObject(K8sObject).Do(func(object *FakeTypeK8sObject) {
			object.Foo += object.Foo
		}).Return(nil)
		mocks.client.EXPECT().Update(gomock.Any(), K8sObject).Return(fmt.Errorf(""))
	})

	require.Error(t, err)
}

func TestDeleteFailsWhenGetFails(t *testing.T) {
	_, err := testDeleteK8sObjectIfExists(t, func(mocks *ApplierTestMocks) {
		mocks.client.EXPECT().Get(gomock.Any(), NamespacedName, K8sObject).Return(fmt.Errorf(""))
	})

	require.Error(t, err)
}

func TestDeleteFailsWhenDeleteCallFails(t *testing.T) {
	_, err := testDeleteK8sObjectIfExists(t, func(mocks *ApplierTestMocks) {
		mocks.client.EXPECT().Get(gomock.Any(), NamespacedName, K8sObject).Return(nil)
		mocks.client.EXPECT().Delete(gomock.Any(), K8sObject).Return(fmt.Errorf(""))
	})

	require.Error(t, err)
}

func TestObjectIsDeletedWhenExisting(t *testing.T) {
	deleted, err := testDeleteK8sObjectIfExists(t, func(mocks *ApplierTestMocks) {
		mocks.client.EXPECT().Get(gomock.Any(), NamespacedName, K8sObject).Return(nil)
		mocks.client.EXPECT().Delete(gomock.Any(), K8sObject).Return(nil)
	})

	require.NoError(t, err)
	require.True(t, deleted)
}

func TestObjectIsNotDeletedWhenNotExisting(t *testing.T) {
	deleted, err := testDeleteK8sObjectIfExists(t, func(mocks *ApplierTestMocks) {
		notFoundError := errors.NewNotFound(schema.GroupResource{}, "")
		mocks.client.EXPECT().Get(gomock.Any(), NamespacedName, K8sObject).Return(notFoundError)
	})

	require.NoError(t, err)
	require.False(t, deleted)
}
