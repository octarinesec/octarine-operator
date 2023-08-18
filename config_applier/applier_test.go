package config_applier_test

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/test_utils/mocks"
	"github.com/vmware/cbcontainers-operator/config_applier"
	mocksConfigApplier "github.com/vmware/cbcontainers-operator/config_applier/mocks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestConfigChangeIsAppliedAndAcknowledgedCorrectly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockK8sClient := mocks.NewMockClient(ctrl)
	mockAPI := mocksConfigApplier.NewMockConfigurationAPI(ctrl)

	applier := config_applier.Applier{
		K8sClient: mockK8sClient,
		Logger:    logr.Discard(),
		Api:       mockAPI,
	}

	// Once a change appears
	// It should find our CR
	// It should be applied to the CR
	// It should be ACKed with proper CR generation and ID
	// TODO: Compatiblity check

	configChange := config_applier.RandomChange()
	mockAPI.EXPECT().GetConfigurationChanges(gomock.Any()).Return([]config_applier.ConfigurationChange{*configChange}, nil)

	mockK8sClient.EXPECT().List(gomock.Any(), &cbcontainersv1.CBContainersAgentList{}).
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

	mockK8sClient.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, item any, _ ...any) error {
			asCb, ok := item.(*cbcontainersv1.CBContainersAgent)
			require.True(t, ok)
			asCb.ObjectMeta.Generation++
			return nil
		})

	mockAPI.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, update config_applier.ConfigurationChangeStatusUpdate) error {
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
			mockAPI := mocksConfigApplier.NewMockConfigurationAPI(ctrl)
			applier := config_applier.Applier{
				K8sClient: nil,
				Logger:    logr.Discard(),
				Api:       mockAPI,
			}

			mockAPI.EXPECT().GetConfigurationChanges(gomock.Any()).Return(tC.dataFromService, nil)
			mockAPI.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).Times(0)

			err := applier.RunIteration(context.Background())
			assert.NoError(t, err)
		})
	}
}

func TestWhenConfigurationAPIReturnsErrorForListShouldPropagateErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAPI := mocksConfigApplier.NewMockConfigurationAPI(ctrl)
	applier := config_applier.Applier{
		K8sClient: nil,
		Logger:    logr.Discard(),
		Api:       mockAPI,
	}

	errFromService := errors.New("some error")
	mockAPI.EXPECT().GetConfigurationChanges(gomock.Any()).Return(nil, errFromService)

	returnedErr := applier.RunIteration(context.Background())
	assert.Error(t, returnedErr)
	assert.ErrorIs(t, returnedErr, errFromService, "expected returned error to match or wrap error from service")
}
