package monitor_test

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/monitor"
	"github.com/vmware/cbcontainers-operator/cbcontainers/test_utils/mocks"
	"testing"
)

type SetupAndAssertDefaultFeaturesProvider func(*mocks.MockClient, *monitor.DefaultFeaturesStatusProvider)

func testFeatures(t *testing.T, setupAndAssert SetupAndAssertDefaultFeaturesProvider) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	clientMock := mocks.NewMockClient(ctrl)
	setupAndAssert(clientMock, monitor.NewDefaultFeaturesStatusProvider(clientMock))
}

func TestHardeningEnabled(t *testing.T) {
	t.Run("When Client.List returns error, should return error", func(t *testing.T) {
		testFeatures(t, func(client *mocks.MockClient, provider *monitor.DefaultFeaturesStatusProvider) {
			client.EXPECT().List(gomock.Any(), &cbcontainersv1.CBContainersHardeningList{}).Return(fmt.Errorf(""))
			_, err := provider.HardeningEnabled()
			require.Error(t, err)
		})
	})

	t.Run("When Client.List returns no items, should return error", func(t *testing.T) {
		testFeatures(t, func(client *mocks.MockClient, provider *monitor.DefaultFeaturesStatusProvider) {
			client.EXPECT().List(gomock.Any(), &cbcontainersv1.CBContainersHardeningList{}).Return(nil)
			_, err := provider.HardeningEnabled()
			require.Error(t, err)
		})
	})

	t.Run("When Client.List returns 0 items, should return false", func(t *testing.T) {
		testFeatures(t, func(client *mocks.MockClient, provider *monitor.DefaultFeaturesStatusProvider) {
			client.EXPECT().List(gomock.Any(), &cbcontainersv1.CBContainersHardeningList{}).
				Do(func(ctx context.Context, list *cbcontainersv1.CBContainersHardeningList) {
					list.Items = make([]cbcontainersv1.CBContainersHardening, 0)
				}).
				Return(nil)
			enabled, err := provider.HardeningEnabled()
			require.NoError(t, err)
			require.False(t, enabled)
		})
	})

	t.Run("When Client.List returns 1 items, should return true", func(t *testing.T) {
		testFeatures(t, func(client *mocks.MockClient, provider *monitor.DefaultFeaturesStatusProvider) {
			client.EXPECT().List(gomock.Any(), &cbcontainersv1.CBContainersHardeningList{}).
				Do(func(ctx context.Context, list *cbcontainersv1.CBContainersHardeningList) {
					list.Items = make([]cbcontainersv1.CBContainersHardening, 1)
				}).
				Return(nil)
			enabled, err := provider.HardeningEnabled()
			require.NoError(t, err)
			require.True(t, enabled)
		})
	})
}

func TestRuntimeEnabled(t *testing.T) {
	t.Run("Always return false", func(t *testing.T) {
		testFeatures(t, func(client *mocks.MockClient, provider *monitor.DefaultFeaturesStatusProvider) {
			enabled, err := provider.RuntimeEnabled()
			require.NoError(t, err)
			require.False(t, enabled)
		})
	})
}
