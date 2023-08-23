package config_applier_test

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	k8sMocks "github.com/vmware/cbcontainers-operator/cbcontainers/test_utils/mocks"
	"github.com/vmware/cbcontainers-operator/config_applier"
	mocksConfigApplier "github.com/vmware/cbcontainers-operator/config_applier/mocks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

// TODO: Compatibility checks
// TODO: Adding CNDR to the config options
// TODO: Properly handle version + custom image to override the custom image
// TODO: Check fields are applied to CR correctly

// TODO: Reads cluster, etc from CR correctly?
// TODO: Respects proxy
// TODO: Review gomock.any usages here
// TODO: version tests
// TODO: Multiple changes are applied according to timestamp

type applierMocks struct {
	k8sClient *k8sMocks.MockClient
	api       *mocksConfigApplier.MockConfigurationChangesAPI
}

func setupApplier(ctrl *gomock.Controller) (*config_applier.Applier, applierMocks) {
	k8sClient := k8sMocks.NewMockClient(ctrl)
	api := mocksConfigApplier.NewMockConfigurationChangesAPI(ctrl)

	applier := config_applier.NewApplier(k8sClient, api, logr.Discard())
	mocksHolder := applierMocks{
		k8sClient: k8sClient,
		api:       api,
	}

	return applier, mocksHolder
}

func TestConfigChangeIsAppliedAndAcknowledgedCorrectly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	applier, mocks := setupApplier(ctrl)

	// TODO: Compatiblity check

	configChange := config_applier.RandomNonNilChange()
	mocks.api.EXPECT().GetConfigurationChanges(gomock.Any()).Return([]config_applier.ConfigurationChange{*configChange}, nil)

	mocks.k8sClient.EXPECT().List(gomock.Any(), &cbcontainersv1.CBContainersAgentList{}).
		Do(func(ctx context.Context, list *cbcontainersv1.CBContainersAgentList, _ ...any) {
			list.Items = []cbcontainersv1.CBContainersAgent{
				{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Generation: 1,
					},
					Spec:   cbcontainersv1.CBContainersAgentSpec{},
					Status: cbcontainersv1.CBContainersAgentStatus{},
				},
			}
		})

	mocks.k8sClient.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, item any, _ ...any) error {
			asCb, ok := item.(*cbcontainersv1.CBContainersAgent)
			require.True(t, ok)
			require.Equal(t, *configChange.AgentVersion, asCb.Spec.Version)
			asCb.ObjectMeta.Generation++
			return nil
		})

	mocks.api.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, update config_applier.ConfigurationChangeStatusUpdate) error {
		assert.Equal(t, configChange.ID, update.ID)
		assert.Equal(t, int64(2), update.AppliedGeneration)
		assert.Equal(t, "ACKNOWLEDGED", update.Status)
		assert.NotEmpty(t, update.AppliedTimestamp, "applied timestamp should be populated")

		parsedTime, err := time.Parse(time.RFC3339, update.AppliedTimestamp)
		assert.NoError(t, err)
		assert.True(t, time.Now().After(parsedTime))
		return nil
	})

	err := applier.RunIteration(context.Background())
	assert.NoError(t, err)
}

func TestWhenThereAreNoPendingChangesNothingHappens(t *testing.T) {
	testCases := []struct {
		name            string
		dataFromService []config_applier.ConfigurationChange
	}{
		{
			name:            "empty list",
			dataFromService: []config_applier.ConfigurationChange{},
		},
		{
			name: "list is not empty but there are no PENDING changes",
			dataFromService: []config_applier.ConfigurationChange{
				{ID: "123", Status: "non-existent"},
				{ID: "234", Status: "FAILED"},
				{ID: "345", Status: "ACKNOWLEDGED"},
				{ID: "456", Status: "SUCCEEDED"},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			applier, mocks := setupApplier(ctrl)

			mocks.api.EXPECT().GetConfigurationChanges(gomock.Any()).Return(tC.dataFromService, nil)
			mocks.api.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).Times(0)

			err := applier.RunIteration(context.Background())
			assert.NoError(t, err)
		})
	}
}

func TestWhenThereAreMultiplePendingChangesTheOldestIsSelected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	applier, mocks := setupApplier(ctrl)

	olderChange := config_applier.RandomNonNilChange()
	newerChange := config_applier.RandomNonNilChange()

	expectedVersion := "version-for-older-change"
	versionThatShouldNotBe := "version-for-newer-change"
	olderChange.AgentVersion = &expectedVersion
	newerChange.AgentVersion = &versionThatShouldNotBe
	olderChange.Timestamp = time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339)
	newerChange.Timestamp = time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)

	mocks.api.EXPECT().GetConfigurationChanges(gomock.Any()).Return([]config_applier.ConfigurationChange{*newerChange, *olderChange}, nil)

	mocks.k8sClient.EXPECT().List(gomock.Any(), &cbcontainersv1.CBContainersAgentList{}).
		Do(func(ctx context.Context, list *cbcontainersv1.CBContainersAgentList, _ ...any) {
			list.Items = []cbcontainersv1.CBContainersAgent{
				{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Generation: 1,
					},
					Spec:   cbcontainersv1.CBContainersAgentSpec{},
					Status: cbcontainersv1.CBContainersAgentStatus{},
				},
			}
		})

	mocks.k8sClient.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, item any, _ ...any) error {
			asCb, ok := item.(*cbcontainersv1.CBContainersAgent)
			require.True(t, ok)

			assert.Equal(t, expectedVersion, asCb.Spec.Version)
			return nil
		})

	mocks.api.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, update config_applier.ConfigurationChangeStatusUpdate) error {
		assert.Equal(t, olderChange.ID, update.ID)
		return nil
	})

	err := applier.RunIteration(context.Background())
	assert.NoError(t, err)
}

func TestWhenConfigurationAPIReturnsErrorForListShouldPropagateErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	applier, mocks := setupApplier(ctrl)

	errFromService := errors.New("some error")
	mocks.api.EXPECT().GetConfigurationChanges(gomock.Any()).Return(nil, errFromService)

	returnedErr := applier.RunIteration(context.Background())
	assert.Error(t, returnedErr)
	assert.ErrorIs(t, returnedErr, errFromService, "expected returned error to match or wrap error from service")
}

func TestWhenGettingCRFromAPIServerFailsChangeIsUpdatedAsFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	applier, mocks := setupApplier(ctrl)

	configChange := config_applier.RandomNonNilChange()
	mocks.api.EXPECT().GetConfigurationChanges(gomock.Any()).Return([]config_applier.ConfigurationChange{*configChange}, nil)

	errFromService := errors.New("some error")
	mocks.k8sClient.EXPECT().List(gomock.Any(), &cbcontainersv1.CBContainersAgentList{}).Return(errFromService)

	mocks.api.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, update config_applier.ConfigurationChangeStatusUpdate) error {
			assert.Equal(t, configChange.ID, update.ID)
			assert.Equal(t, "FAILED", update.Status)
			assert.NotEmpty(t, update.Reason)
			assert.Equal(t, int64(0), update.AppliedGeneration)
			assert.Empty(t, update.AppliedTimestamp)

			return nil
		})

	returnedErr := applier.RunIteration(context.Background())
	assert.Error(t, returnedErr)
	assert.ErrorIs(t, returnedErr, errFromService, "expected returned error to match or wrap error from service")
}

func TestWhenUpdatingCRFailsChangeIsUpdatedAsFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	applier, mocks := setupApplier(ctrl)

	configChange := config_applier.RandomNonNilChange()
	mocks.api.EXPECT().GetConfigurationChanges(gomock.Any()).Return([]config_applier.ConfigurationChange{*configChange}, nil)

	mocks.k8sClient.EXPECT().List(gomock.Any(), &cbcontainersv1.CBContainersAgentList{}).
		Do(func(ctx context.Context, list *cbcontainersv1.CBContainersAgentList, _ ...any) {
			list.Items = []cbcontainersv1.CBContainersAgent{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       cbcontainersv1.CBContainersAgentSpec{},
					Status:     cbcontainersv1.CBContainersAgentStatus{},
				},
			}
		})

	errFromService := errors.New("some error")
	mocks.k8sClient.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errFromService)

	mocks.api.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, update config_applier.ConfigurationChangeStatusUpdate) error {
			assert.Equal(t, configChange.ID, update.ID)
			assert.Equal(t, "FAILED", update.Status)
			assert.NotEmpty(t, update.Reason)
			assert.Equal(t, int64(0), update.AppliedGeneration)
			assert.Empty(t, update.AppliedTimestamp)

			return nil
		})

	returnedErr := applier.RunIteration(context.Background())
	assert.Error(t, returnedErr)
	assert.ErrorIs(t, returnedErr, errFromService, "expected returned error to match or wrap error from service")
}

func TestWhenUpdatingStatusToBackendFailsShouldReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	applier, mocks := setupApplier(ctrl)

	configChange := config_applier.RandomNonNilChange()
	mocks.api.EXPECT().GetConfigurationChanges(gomock.Any()).Return([]config_applier.ConfigurationChange{*configChange}, nil)

	mocks.k8sClient.EXPECT().List(gomock.Any(), &cbcontainersv1.CBContainersAgentList{}).
		Do(func(ctx context.Context, list *cbcontainersv1.CBContainersAgentList, _ ...any) {
			list.Items = []cbcontainersv1.CBContainersAgent{
				{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Generation: 1,
					},
					Spec:   cbcontainersv1.CBContainersAgentSpec{},
					Status: cbcontainersv1.CBContainersAgentStatus{},
				},
			}
		})

	mocks.k8sClient.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, item any, _ ...any) error {
			asCb, ok := item.(*cbcontainersv1.CBContainersAgent)
			require.True(t, ok)
			asCb.ObjectMeta.Generation++
			return nil
		})

	errFromService := errors.New("some error")
	mocks.api.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).Return(errFromService)

	returnedErr := applier.RunIteration(context.Background())
	assert.Error(t, returnedErr)
	assert.ErrorIs(t, errFromService, returnedErr, "expected returned error to match or wrap error from service")
}

func TestWhenThereIsNoCRInstalledChangeIsUpdatedAsFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	applier, mocks := setupApplier(ctrl)

	configChange := config_applier.RandomNonNilChange()
	mocks.api.EXPECT().GetConfigurationChanges(gomock.Any()).Return([]config_applier.ConfigurationChange{*configChange}, nil)

	mocks.k8sClient.EXPECT().List(gomock.Any(), &cbcontainersv1.CBContainersAgentList{}).
		Do(func(ctx context.Context, list *cbcontainersv1.CBContainersAgentList, _ ...any) {
			list.Items = []cbcontainersv1.CBContainersAgent{}
		})

	mocks.api.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, update config_applier.ConfigurationChangeStatusUpdate) error {
			assert.Equal(t, configChange.ID, update.ID)
			assert.Equal(t, "FAILED", update.Status)
			assert.NotEmpty(t, update.Reason)
			assert.Equal(t, int64(0), update.AppliedGeneration)
			assert.Empty(t, update.AppliedTimestamp)

			return nil
		})

	err := applier.RunIteration(context.Background())
	assert.Error(t, err)
	// TODO: Specific error exposed for this?
}
