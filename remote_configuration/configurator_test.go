package remote_configuration_test

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	k8sMocks "github.com/vmware/cbcontainers-operator/cbcontainers/test_utils/mocks"
	"github.com/vmware/cbcontainers-operator/remote_configuration"
	mocksConfigurator "github.com/vmware/cbcontainers-operator/remote_configuration/mocks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

// TODO: What error data to show and what not?

// TODO: Reads cluster, etc from CR correctly?
// TODO: Respects proxy
// TODO: Review gomock.any usages here

// TODO: error on compatiblity calls

type configuratorMocks struct {
	k8sClient        *k8sMocks.MockClient
	configChangesAPI *mocksConfigurator.MockConfigurationChangesAPI
	changeValidator  *mocksConfigurator.MockChangeValidator
}

func setupConfigurator(ctrl *gomock.Controller) (*remote_configuration.Configurator, configuratorMocks) {
	k8sClient := k8sMocks.NewMockClient(ctrl)
	configChangesAPI := mocksConfigurator.NewMockConfigurationChangesAPI(ctrl)
	changeValidator := mocksConfigurator.NewMockChangeValidator(ctrl)

	configurator := remote_configuration.NewConfigurator(k8sClient, configChangesAPI, changeValidator, logr.Discard())

	mocksHolder := configuratorMocks{
		k8sClient:        k8sClient,
		configChangesAPI: configChangesAPI,
		changeValidator:  changeValidator,
	}

	return configurator, mocksHolder
}

func TestConfigChangeIsAppliedAndAcknowledgedCorrectly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configurator, mocks := setupConfigurator(ctrl)
	var initialGeneration, finalGeneration int64 = 1, 2

	setupCRInK8S(mocks.k8sClient, &cbcontainersv1.CBContainersAgent{ObjectMeta: metav1.ObjectMeta{Generation: initialGeneration}})
	setupChangeValidatorToAcceptAll(mocks.changeValidator)

	configChange := remote_configuration.RandomNonNilChange()
	mocks.configChangesAPI.EXPECT().GetConfigurationChanges(gomock.Any()).Return([]remote_configuration.ConfigurationChange{configChange}, nil)

	mocks.configChangesAPI.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, update remote_configuration.ConfigurationChangeStatusUpdate) error {
		assert.Equal(t, configChange.ID, update.ID)
		assert.Equal(t, finalGeneration, update.AppliedGeneration)
		assert.Equal(t, "ACKNOWLEDGED", update.Status)
		assert.NotEmpty(t, update.AppliedTimestamp, "applied timestamp should be populated")

		parsedTime, err := time.Parse(time.RFC3339, update.AppliedTimestamp)
		assert.NoError(t, err)
		assert.True(t, time.Now().After(parsedTime))
		return nil
	})

	assertUpdateCR(t, mocks.k8sClient, func(agent *cbcontainersv1.CBContainersAgent) {
		assert.Equal(t, *configChange.AgentVersion, agent.Spec.Version)
		agent.ObjectMeta.Generation = finalGeneration
	})

	err := configurator.RunIteration(context.Background())
	assert.NoError(t, err)
}

func TestWhenChangeIsNotApplicableShouldReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)

	configurator, mocks := setupConfigurator(ctrl)

	cr := setupCRInK8S(mocks.k8sClient, nil)

	configChange := remote_configuration.RandomNonNilChange()
	mocks.configChangesAPI.EXPECT().GetConfigurationChanges(gomock.Any()).Return([]remote_configuration.ConfigurationChange{configChange}, nil)

	mocks.changeValidator.EXPECT().ValidateChange(configChange, cr).Return(false, "your data is wrong pal")

	mocks.configChangesAPI.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, update remote_configuration.ConfigurationChangeStatusUpdate) error {
			assert.Equal(t, configChange.ID, update.ID)
			assert.Equal(t, "FAILED", update.Status)
			assert.NotEmpty(t, update.Reason)
			assert.Equal(t, int64(0), update.AppliedGeneration)
			assert.Empty(t, update.AppliedTimestamp)

			return nil
		})

	err := configurator.RunIteration(context.Background())
	assert.Error(t, err)
}

func TestWhenThereAreNoPendingChangesNothingHappens(t *testing.T) {
	testCases := []struct {
		name            string
		dataFromService []remote_configuration.ConfigurationChange
	}{
		{
			name:            "empty list",
			dataFromService: []remote_configuration.ConfigurationChange{},
		},
		{
			name: "list is not empty but there are no PENDING changes",
			dataFromService: []remote_configuration.ConfigurationChange{
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

			configurator, mocks := setupConfigurator(ctrl)

			setupCRInK8S(mocks.k8sClient, nil)
			mocks.configChangesAPI.EXPECT().GetConfigurationChanges(gomock.Any()).Return(tC.dataFromService, nil)
			mocks.configChangesAPI.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).Times(0)

			err := configurator.RunIteration(context.Background())
			assert.NoError(t, err)
		})
	}
}

func TestWhenThereAreMultiplePendingChangesTheOldestIsSelected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configurator, mocks := setupConfigurator(ctrl)

	olderChange := remote_configuration.RandomNonNilChange()
	newerChange := remote_configuration.RandomNonNilChange()

	expectedVersion := "version-for-older-change"
	versionThatShouldNotBe := "version-for-newer-change"
	olderChange.AgentVersion = &expectedVersion
	newerChange.AgentVersion = &versionThatShouldNotBe
	olderChange.Timestamp = time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339)
	newerChange.Timestamp = time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)

	setupCRInK8S(mocks.k8sClient, nil)
	setupChangeValidatorToAcceptAll(mocks.changeValidator)

	mocks.configChangesAPI.EXPECT().GetConfigurationChanges(gomock.Any()).Return([]remote_configuration.ConfigurationChange{newerChange, olderChange}, nil)

	assertUpdateCR(t, mocks.k8sClient, func(agent *cbcontainersv1.CBContainersAgent) {
		assert.Equal(t, expectedVersion, agent.Spec.Version)
	})

	mocks.configChangesAPI.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, update remote_configuration.ConfigurationChangeStatusUpdate) error {
			assert.Equal(t, olderChange.ID, update.ID)
			return nil
		})

	err := configurator.RunIteration(context.Background())
	assert.NoError(t, err)
}

func TestWhenConfigurationAPIReturnsErrorForListShouldPropagateErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configurator, mocks := setupConfigurator(ctrl)

	setupCRInK8S(mocks.k8sClient, nil)

	errFromService := errors.New("some error")
	mocks.configChangesAPI.EXPECT().GetConfigurationChanges(gomock.Any()).Return(nil, errFromService)

	returnedErr := configurator.RunIteration(context.Background())

	assert.Error(t, returnedErr)
	assert.ErrorIs(t, returnedErr, errFromService, "expected returned error to match or wrap error from service")
}

func TestWhenGettingCRFromAPIServerFailsAnErrorIsReturned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configurator, mocks := setupConfigurator(ctrl)

	errFromService := errors.New("some error")
	mocks.k8sClient.EXPECT().List(gomock.Any(), &cbcontainersv1.CBContainersAgentList{}).Return(errFromService)

	returnedErr := configurator.RunIteration(context.Background())
	assert.Error(t, returnedErr)
	assert.ErrorIs(t, returnedErr, errFromService, "expected returned error to match or wrap error from service")
}

func TestWhenUpdatingCRFailsChangeIsUpdatedAsFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configurator, mocks := setupConfigurator(ctrl)

	setupCRInK8S(mocks.k8sClient, nil)
	setupChangeValidatorToAcceptAll(mocks.changeValidator)

	configChange := remote_configuration.RandomNonNilChange()
	mocks.configChangesAPI.EXPECT().GetConfigurationChanges(gomock.Any()).Return([]remote_configuration.ConfigurationChange{configChange}, nil)

	errFromService := errors.New("some error")
	mocks.k8sClient.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errFromService)

	mocks.configChangesAPI.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, update remote_configuration.ConfigurationChangeStatusUpdate) error {
			assert.Equal(t, configChange.ID, update.ID)
			assert.Equal(t, "FAILED", update.Status)
			assert.NotEmpty(t, update.Reason)
			assert.Equal(t, int64(0), update.AppliedGeneration)
			assert.Empty(t, update.AppliedTimestamp)

			return nil
		})

	returnedErr := configurator.RunIteration(context.Background())
	assert.Error(t, returnedErr)
	assert.ErrorIs(t, returnedErr, errFromService, "expected returned error to match or wrap error from service")
}

func TestWhenUpdatingStatusToBackendFailsShouldReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // tODO: Remove all; redundant since 1.14

	configurator, mocks := setupConfigurator(ctrl)

	setupCRInK8S(mocks.k8sClient, nil)
	setupChangeValidatorToAcceptAll(mocks.changeValidator)

	configChange := remote_configuration.RandomNonNilChange()
	mocks.configChangesAPI.EXPECT().GetConfigurationChanges(gomock.Any()).Return([]remote_configuration.ConfigurationChange{configChange}, nil)

	mocks.k8sClient.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	errFromService := errors.New("some error")
	mocks.configChangesAPI.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).Return(errFromService)

	returnedErr := configurator.RunIteration(context.Background())
	assert.Error(t, returnedErr)
	assert.ErrorIs(t, errFromService, returnedErr, "expected returned error to match or wrap error from service")
}

func TestWhenThereIsNoCRInstalledNothingHappens(t *testing.T) {
	ctrl := gomock.NewController(t)

	configurator, mocks := setupConfigurator(ctrl)

	mocks.k8sClient.EXPECT().List(gomock.Any(), &cbcontainersv1.CBContainersAgentList{}).
		Do(func(ctx context.Context, list *cbcontainersv1.CBContainersAgentList, _ ...any) {
			list.Items = []cbcontainersv1.CBContainersAgent{}
		})

	assert.NoError(t, configurator.RunIteration(context.Background()))
}

// setupCRInK8S ensures the mock client will return 1 agent item for List calls - either the provided one or an empty CR otherwise
func setupCRInK8S(mock *k8sMocks.MockClient, item *cbcontainersv1.CBContainersAgent) *cbcontainersv1.CBContainersAgent {
	if item == nil {
		item = &cbcontainersv1.CBContainersAgent{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       cbcontainersv1.CBContainersAgentSpec{},
			Status:     cbcontainersv1.CBContainersAgentStatus{},
		}
	}
	mock.EXPECT().List(gomock.Any(), &cbcontainersv1.CBContainersAgentList{}).
		Do(func(ctx context.Context, list *cbcontainersv1.CBContainersAgentList, _ ...any) {
			list.Items = []cbcontainersv1.CBContainersAgent{*item}
		})

	return item
}

func assertUpdateCR(t *testing.T, mock *k8sMocks.MockClient, assert func(*cbcontainersv1.CBContainersAgent)) {
	mock.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, item any, _ ...any) error {
			asCb, ok := item.(*cbcontainersv1.CBContainersAgent)
			require.True(t, ok)

			assert(asCb)
			return nil
		})
}

func setupChangeValidatorToAcceptAll(mock *mocksConfigurator.MockChangeValidator) {
	mock.EXPECT().ValidateChange(gomock.Any(), gomock.Any()).Return(true, "")
}
