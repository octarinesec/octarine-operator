package remote_configuration_test

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	k8sMocks "github.com/vmware/cbcontainers-operator/cbcontainers/test_utils/mocks"
	"github.com/vmware/cbcontainers-operator/remote_configuration"
	mocksConfigurator "github.com/vmware/cbcontainers-operator/remote_configuration/mocks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

// TODO: Add back the .finish for older mockgens
// TODO: What error data to show and what not?

// TODO: Reads cluster, etc from CR correctly?
// TODO: Respects proxy
// TODO: Review gomock.any usages here

// TODO: error on compatiblity calls

type configuratorMocks struct {
	k8sClient           *k8sMocks.MockClient
	apiGateway          *mocksConfigurator.MockApiGateway
	accessTokenProvider *mocksConfigurator.MockAccessTokenProvider
	syncer              *mocksConfigurator.MockCustomResourceSyncer

	stubAccessToken     string
	stubOperatorVersion string
	stubNamespace       string

	testCtx context.Context
}

// setupConfigurator TODO
func setupConfigurator(ctrl *gomock.Controller) (*remote_configuration.Configurator, configuratorMocks) {
	k8sClient := k8sMocks.NewMockClient(ctrl)
	apiGateway := mocksConfigurator.NewMockApiGateway(ctrl)
	accessTokenProvider := mocksConfigurator.NewMockAccessTokenProvider(ctrl)
	syncer := mocksConfigurator.NewMockCustomResourceSyncer(ctrl)

	var mockAPIProvider remote_configuration.ApiCreator = func(
		cbContainersCluster *cbcontainersv1.CBContainersAgent,
		accessToken string,
	) (remote_configuration.ApiGateway, error) {
		return apiGateway, nil
	}

	namespace := "namespace-name"
	accessToken := "access-token"
	operatorVersion := "1.2.3"
	accessTokenProvider.EXPECT().GetCBAccessToken(gomock.Any(), gomock.Any(), namespace).Return(accessToken, nil).AnyTimes()

	configurator := remote_configuration.NewConfigurator(
		mockAPIProvider,
		logr.Discard(),
		accessTokenProvider,
		syncer,
		operatorVersion,
		namespace,
	)

	mocksHolder := configuratorMocks{
		k8sClient:           k8sClient,
		apiGateway:          apiGateway,
		accessTokenProvider: accessTokenProvider,
		syncer:              syncer,
		stubAccessToken:     accessToken,
		stubOperatorVersion: operatorVersion,
		stubNamespace:       namespace,
		testCtx:             context.Background(), // TODO: Remove?
	}

	return configurator, mocksHolder
}

func TestConfigChangeIsAppliedAndAcknowledgedCorrectly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configurator, mocks := setupConfigurator(ctrl)
	var initialGeneration, finalGeneration int64 = 1, 2

	cr := &cbcontainersv1.CBContainersAgent{ObjectMeta: metav1.ObjectMeta{Generation: initialGeneration}}
	mocks.syncer.EXPECT().GetCR(gomock.Any()).Return(cr, nil)

	// TODO: Generation

	configChange := remote_configuration.RandomNonNilChange()
	mocks.apiGateway.EXPECT().GetConfigurationChanges(gomock.Any()).Return([]models.ConfigurationChange{configChange}, nil)

	mocks.apiGateway.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, update models.ConfigurationChangeStatusUpdate) error {
		assert.Equal(t, configChange.ID, update.ID)
		assert.Equal(t, finalGeneration, update.AppliedGeneration)
		assert.Equal(t, "ACKNOWLEDGED", update.Status)
		assert.NotEmpty(t, update.AppliedTimestamp, "applied timestamp should be populated")

		parsedTime, err := time.Parse(time.RFC3339, update.AppliedTimestamp)
		assert.NoError(t, err)
		assert.True(t, time.Now().After(parsedTime))
		return nil
	})

	mocks.apiGateway.EXPECT().GetSensorMetadata().Return(nil, nil)
	mocks.apiGateway.EXPECT().GetCompatibilityMatrixEntryFor(mocks.stubOperatorVersion).Return(&models.OperatorCompatibility{}, nil)

	mocks.syncer.
		EXPECT().
		ApplyChangeToCR(gomock.Any(), configChange, cr, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ models.ConfigurationChange, cr *cbcontainersv1.CBContainersAgent, _ any) error {
			// We simulate the generation bump that we expect k8s to do
			cr.Generation = finalGeneration
			return nil
		})

	err := configurator.RunIteration(context.Background())
	assert.NoError(t, err)
}

// TODO: reintroduce

//func TestWhenChangeIsNotApplicableShouldReturnError(t *testing.T) {
//	ctrl := gomock.NewController(t)
//
//	configurator, mocks := setupConfigurator(ctrl)
//
//	cr := setupCRInK8S(mocks.k8sClient, nil)
//	if cr != nil {
//		// Placeholder
//	}
//
//	configChange := remote_configuration.RandomNonNilChange()
//	mocks.apiGateway.EXPECT().GetConfigurationChanges(gomock.Any()).Return([]models.ConfigurationChange{configChange}, nil)
//
//	//mocks.changeValidator.EXPECT().ValidateChange(configChange, cr).Return(errors.New("your data is wrong pal"))
//
//	mocks.apiGateway.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).
//		DoAndReturn(func(_ context.Context, update models.ConfigurationChangeStatusUpdate) error {
//			assert.Equal(t, configChange.ID, update.ID)
//			assert.Equal(t, "FAILED", update.Status)
//			assert.NotEmpty(t, update.Reason)
//			assert.Equal(t, int64(0), update.AppliedGeneration)
//			assert.Empty(t, update.AppliedTimestamp)
//
//			return nil
//		})
//
//	err := configurator.RunIteration(context.Background())
//	assert.Error(t, err)
//}

func TestWhenThereAreNoPendingChangesNothingHappens(t *testing.T) {
	testCases := []struct {
		name            string
		dataFromService []models.ConfigurationChange
	}{
		{
			name:            "empty list",
			dataFromService: []models.ConfigurationChange{},
		},
		{
			name: "list is not empty but there are no PENDING changes",
			dataFromService: []models.ConfigurationChange{
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

			mocks.syncer.EXPECT().GetCR(gomock.Any()).Return(&cbcontainersv1.CBContainersAgent{}, nil)
			mocks.apiGateway.EXPECT().GetConfigurationChanges(gomock.Any()).Return(tC.dataFromService, nil)
			mocks.apiGateway.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).Times(0)

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

	mocks.syncer.EXPECT().GetCR(gomock.Any()).Return(&cbcontainersv1.CBContainersAgent{}, nil)
	mocks.apiGateway.EXPECT().GetConfigurationChanges(gomock.Any()).Return([]models.ConfigurationChange{newerChange, olderChange}, nil)
	mocks.apiGateway.EXPECT().GetSensorMetadata().Return(nil, nil)
	mocks.apiGateway.EXPECT().GetCompatibilityMatrixEntryFor(mocks.stubOperatorVersion).Return(&models.OperatorCompatibility{}, nil)

	mocks.syncer.EXPECT().ApplyChangeToCR(gomock.Any(), olderChange, gomock.Any(), gomock.Any()).Return(nil).Times(1)

	mocks.apiGateway.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, update models.ConfigurationChangeStatusUpdate) error {
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

	mocks.syncer.EXPECT().GetCR(gomock.Any()).Return(&cbcontainersv1.CBContainersAgent{}, nil)

	errFromService := errors.New("some error")
	mocks.apiGateway.EXPECT().GetConfigurationChanges(gomock.Any()).Return(nil, errFromService)

	returnedErr := configurator.RunIteration(context.Background())

	assert.Error(t, returnedErr)
	assert.ErrorIs(t, returnedErr, errFromService, "expected returned error to match or wrap error from service")
}

func TestWhenGettingCRFromAPIServerFailsAnErrorIsReturned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configurator, mocks := setupConfigurator(ctrl)

	errFromK8S := errors.New("some error")
	mocks.syncer.EXPECT().GetCR(gomock.Any()).Return(nil, errFromK8S)

	returnedErr := configurator.RunIteration(context.Background())
	assert.Error(t, returnedErr)
	assert.ErrorIs(t, returnedErr, errFromK8S, "expected returned error to match or wrap error from service")
}

func TestWhenUpdatingCRFailsChangeIsUpdatedAsFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configurator, mocks := setupConfigurator(ctrl)

	configChange := remote_configuration.RandomNonNilChange()

	mocks.apiGateway.EXPECT().GetConfigurationChanges(gomock.Any()).Return([]models.ConfigurationChange{configChange}, nil)
	mocks.syncer.EXPECT().GetCR(gomock.Any()).Return(&cbcontainersv1.CBContainersAgent{}, nil)
	mocks.apiGateway.EXPECT().GetSensorMetadata().Return(nil, nil)
	mocks.apiGateway.EXPECT().GetCompatibilityMatrixEntryFor(mocks.stubOperatorVersion).Return(&models.OperatorCompatibility{}, nil)

	errFromService := errors.New("some error")
	mocks.syncer.EXPECT().ApplyChangeToCR(gomock.Any(), configChange, gomock.Any(), gomock.Any()).Return(errFromService)

	mocks.apiGateway.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, update models.ConfigurationChangeStatusUpdate) error {
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
	defer ctrl.Finish()

	configurator, mocks := setupConfigurator(ctrl)

	configChange := remote_configuration.RandomNonNilChange()

	mocks.apiGateway.EXPECT().GetConfigurationChanges(gomock.Any()).Return([]models.ConfigurationChange{configChange}, nil)
	mocks.syncer.EXPECT().GetCR(gomock.Any()).Return(&cbcontainersv1.CBContainersAgent{}, nil)
	mocks.apiGateway.EXPECT().GetSensorMetadata().Return(nil, nil)
	mocks.apiGateway.EXPECT().GetCompatibilityMatrixEntryFor(mocks.stubOperatorVersion).Return(&models.OperatorCompatibility{}, nil)
	mocks.syncer.EXPECT().ApplyChangeToCR(gomock.Any(), configChange, gomock.Any(), gomock.Any()).Return(nil)

	errFromService := errors.New("some error")
	mocks.apiGateway.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).Return(errFromService)

	returnedErr := configurator.RunIteration(context.Background())
	assert.Error(t, returnedErr)
	assert.ErrorIs(t, errFromService, returnedErr, "expected returned error to match or wrap error from service")
}

func TestWhenThereIsNoCRInstalledNothingHappens(t *testing.T) {
	ctrl := gomock.NewController(t)

	configurator, mocks := setupConfigurator(ctrl)

	mocks.syncer.EXPECT().GetCR(gomock.Any()).Return(nil, nil)

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
