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
	"github.com/vmware/cbcontainers-operator/cbcontainers/remote_configuration"
	"github.com/vmware/cbcontainers-operator/cbcontainers/remote_configuration/mocks"
	k8sMocks "github.com/vmware/cbcontainers-operator/cbcontainers/test_utils/mocks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

type configuratorMocks struct {
	k8sClient           *k8sMocks.MockClient
	apiGateway          *mocks.MockApiGateway
	accessTokenProvider *mocks.MockAccessTokenProvider
	validator           *mocks.MockChangeValidator

	stubAccessToken string
	stubNamespace   string
	stubClusterID   string
}

// setupConfigurator sets up mocks and creates a Configurator instance with those mocks and some dummy data
func setupConfigurator(ctrl *gomock.Controller) (*remote_configuration.Configurator, configuratorMocks) {
	k8sClient := k8sMocks.NewMockClient(ctrl)
	apiGateway := mocks.NewMockApiGateway(ctrl)
	accessTokenProvider := mocks.NewMockAccessTokenProvider(ctrl)
	validator := mocks.NewMockChangeValidator(ctrl)

	var mockAPIProvider remote_configuration.ApiCreator = func(
		cbContainersCluster *cbcontainersv1.CBContainersAgent,
		accessToken string,
	) (remote_configuration.ApiGateway, error) {
		return apiGateway, nil
	}
	var mockValidatorProvider remote_configuration.ValidatorCreator = func(gateway remote_configuration.ApiGateway) (remote_configuration.ChangeValidator, error) {
		return validator, nil
	}

	namespace := "namespace-name"
	accessToken := "access-token"
	clusterID := "1234567"
	accessTokenProvider.EXPECT().GetCBAccessToken(gomock.Any(), gomock.Any(), namespace).Return(accessToken, nil).AnyTimes()

	configurator := remote_configuration.NewConfigurator(
		k8sClient,
		nil,
		mockAPIProvider,
		logr.Discard(),
		accessTokenProvider,
		mockValidatorProvider,
		namespace,
		clusterID,
	)

	mocksHolder := configuratorMocks{
		k8sClient:           k8sClient,
		apiGateway:          apiGateway,
		accessTokenProvider: accessTokenProvider,
		validator:           validator,
		stubAccessToken:     accessToken,
		stubNamespace:       namespace,
		stubClusterID:       clusterID,
	}

	return configurator, mocksHolder
}

func TestConfigChangeIsAppliedAndAcknowledgedCorrectly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configurator, mocks := setupConfigurator(ctrl)

	// Setup stub data
	var initialGeneration, finalGeneration int64 = 1, 2
	expectedAgentVersion := "3.0.0"
	cr := &cbcontainersv1.CBContainersAgent{ObjectMeta: metav1.ObjectMeta{Generation: initialGeneration}}
	configChange := randomPendingConfigChange()
	configChange.AgentVersion = &expectedAgentVersion

	setupCRInK8S(mocks.k8sClient, cr)
	mocks.validator.EXPECT().ValidateChange(configChange, cr).Return(nil)
	mocks.apiGateway.EXPECT().GetSensorMetadata().Return([]models.SensorMetadata{{Version: expectedAgentVersion}}, nil)
	mocks.apiGateway.EXPECT().GetConfigurationChanges(gomock.Any(), mocks.stubClusterID).Return([]models.ConfigurationChange{configChange}, nil)

	// Setup mock assertions
	mocks.apiGateway.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, update models.ConfigurationChangeStatusUpdate) error {
		assert.Equal(t, configChange.ID, update.ID)
		assert.Equal(t, finalGeneration, update.AppliedGeneration)
		assert.Equal(t, models.ChangeStatusAcked, update.Status)
		assert.NotEmpty(t, update.AppliedTimestamp, "applied timestamp should be populated")
		assert.Equal(t, mocks.stubClusterID, update.ClusterIdentifier)

		parsedTime, err := time.Parse(time.RFC3339, update.AppliedTimestamp)
		assert.NoError(t, err)
		assert.True(t, time.Now().After(parsedTime))
		return nil
	})

	setupUpdateCRMock(t, mocks.k8sClient, func(agent *cbcontainersv1.CBContainersAgent) {
		assert.Equal(t, expectedAgentVersion, agent.Spec.Version)
		agent.ObjectMeta.Generation = finalGeneration
	})

	err := configurator.RunIteration(context.Background())
	assert.NoError(t, err)
}

func TestWhenSensorMetadataIsAvailableItIsUsed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configurator, mocks := setupConfigurator(ctrl)

	expectedAgentVersion := "3.0.0"
	cr := &cbcontainersv1.CBContainersAgent{}
	configChange := randomPendingConfigChange()
	configChange.AgentVersion = &expectedAgentVersion

	setupCRInK8S(mocks.k8sClient, cr)
	setupValidatorAcceptAll(mocks.validator)
	mocks.apiGateway.EXPECT().GetSensorMetadata().Return([]models.SensorMetadata{{
		Version:                        expectedAgentVersion,
		SupportsRuntime:                true,
		SupportsClusterScanning:        true,
		SupportsClusterScanningSecrets: true,
		SupportsCndr:                   true,
	}}, nil)
	mocks.apiGateway.EXPECT().GetConfigurationChanges(gomock.Any(), mocks.stubClusterID).Return([]models.ConfigurationChange{configChange}, nil)

	// Setup mock assertions
	mocks.apiGateway.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).Return(nil)

	setupUpdateCRMock(t, mocks.k8sClient, func(agent *cbcontainersv1.CBContainersAgent) {
		assert.True(t, *agent.Spec.Components.ClusterScanning.Enabled)
		assert.True(t, agent.Spec.Components.ClusterScanning.ClusterScannerAgent.CLIFlags.EnableSecretDetection)
		assert.True(t, *agent.Spec.Components.RuntimeProtection.Enabled)
		assert.True(t, *agent.Spec.Components.Cndr.Enabled)
	})

	err := configurator.RunIteration(context.Background())
	assert.NoError(t, err)
}

func TestWhenSensorMetadataFailsShouldPropagateErr(t *testing.T) {

}

func TestWhenChangeIsNotApplicableShouldReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configurator, mocks := setupConfigurator(ctrl)

	cr := &cbcontainersv1.CBContainersAgent{}
	agentVersion := "5.0.0"
	configChange := randomPendingConfigChange()
	configChange.AgentVersion = &agentVersion

	setupCRInK8S(mocks.k8sClient, cr)
	mocks.apiGateway.EXPECT().GetConfigurationChanges(gomock.Any(), mocks.stubClusterID).Return([]models.ConfigurationChange{configChange}, nil)

	mocks.validator.EXPECT().ValidateChange(configChange, cr).Return(errors.New("some error"))

	// Setup mock assertions
	mocks.apiGateway.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, update models.ConfigurationChangeStatusUpdate) error {
			assert.Equal(t, configChange.ID, update.ID)
			assert.Equal(t, models.ChangeStatusFailed, update.Status)
			assert.NotEmpty(t, update.Error)
			assert.NotEmpty(t, update.ErrorReason) // Should be populated for validation errors
			assert.Equal(t, int64(0), update.AppliedGeneration)
			assert.Empty(t, update.AppliedTimestamp)
			assert.Equal(t, mocks.stubClusterID, update.ClusterIdentifier)

			return nil
		})

	err := configurator.RunIteration(context.Background())
	assert.Error(t, err)
}

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

			setupCRInK8S(mocks.k8sClient, nil)
			mocks.apiGateway.EXPECT().GetConfigurationChanges(gomock.Any(), mocks.stubClusterID).Return(tC.dataFromService, nil)
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

	olderChange := randomPendingConfigChange()
	newerChange := randomPendingConfigChange()

	expectedVersion := "version-for-older-change"
	versionThatShouldNotBe := "version-for-newer-change"
	olderChange.AgentVersion = &expectedVersion
	newerChange.AgentVersion = &versionThatShouldNotBe
	olderChange.Timestamp = time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339)
	newerChange.Timestamp = time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)

	setupCRInK8S(mocks.k8sClient, nil)
	setupValidatorAcceptAll(mocks.validator)
	setupEmptySensorMetadata(mocks.apiGateway)
	mocks.apiGateway.EXPECT().GetConfigurationChanges(gomock.Any(), mocks.stubClusterID).Return([]models.ConfigurationChange{newerChange, olderChange}, nil)

	setupUpdateCRMock(t, mocks.k8sClient, func(agent *cbcontainersv1.CBContainersAgent) {
		assert.Equal(t, expectedVersion, agent.Spec.Version)
	})

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

	setupCRInK8S(mocks.k8sClient, nil)

	errFromService := errors.New("some error")
	mocks.apiGateway.EXPECT().GetConfigurationChanges(gomock.Any(), mocks.stubClusterID).Return(nil, errFromService)

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

	configChange := randomPendingConfigChange()

	setupCRInK8S(mocks.k8sClient, nil)
	setupValidatorAcceptAll(mocks.validator)
	setupEmptySensorMetadata(mocks.apiGateway)
	mocks.apiGateway.EXPECT().GetConfigurationChanges(gomock.Any(), mocks.stubClusterID).Return([]models.ConfigurationChange{configChange}, nil)

	errFromService := errors.New("some error")
	mocks.k8sClient.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errFromService)

	mocks.apiGateway.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, update models.ConfigurationChangeStatusUpdate) error {
			assert.Equal(t, configChange.ID, update.ID)
			assert.Equal(t, models.ChangeStatusFailed, update.Status)
			assert.NotEmpty(t, update.Error)
			assert.Empty(t, update.ErrorReason)
			assert.Equal(t, int64(0), update.AppliedGeneration)
			assert.Empty(t, update.AppliedTimestamp)
			assert.Equal(t, mocks.stubClusterID, update.ClusterIdentifier)

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

	configChange := randomPendingConfigChange()

	setupCRInK8S(mocks.k8sClient, nil)
	setupValidatorAcceptAll(mocks.validator)
	setupEmptySensorMetadata(mocks.apiGateway)
	mocks.apiGateway.EXPECT().GetConfigurationChanges(gomock.Any(), mocks.stubClusterID).Return([]models.ConfigurationChange{configChange}, nil)
	mocks.k8sClient.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	errFromService := errors.New("some error")
	mocks.apiGateway.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).Return(errFromService)

	returnedErr := configurator.RunIteration(context.Background())
	assert.Error(t, returnedErr)
	assert.ErrorIs(t, errFromService, returnedErr, "expected returned error to match or wrap error from service")
}

func TestWhenThereIsNoCRInstalledNothingHappens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configurator, mocks := setupConfigurator(ctrl)

	mocks.k8sClient.EXPECT().List(gomock.Any(), &cbcontainersv1.CBContainersAgentList{}).
		Do(func(ctx context.Context, list *cbcontainersv1.CBContainersAgentList, _ ...any) {
			list.Items = []cbcontainersv1.CBContainersAgent{}
		})

	// No other mock calls should happen without a CR
	assert.NoError(t, configurator.RunIteration(context.Background()))
}

func TestWhenFeatureIsDisabledInCRNothingHappens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configurator, mocks := setupConfigurator(ctrl)

	fal := false
	cr := &cbcontainersv1.CBContainersAgent{
		Spec: cbcontainersv1.CBContainersAgentSpec{
			Components: cbcontainersv1.CBContainersComponentsSpec{
				Settings: cbcontainersv1.CBContainersComponentsSettings{
					RemoteConfiguration: &cbcontainersv1.CBContainersRemoteConfigurationSettings{
						EnabledForAgent: &fal,
					},
				}},
		},
	}
	setupCRInK8S(mocks.k8sClient, cr)

	// No other mock calls should happen once we know the feature is disabled
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

func setupUpdateCRMock(t *testing.T, mock *k8sMocks.MockClient, assert func(*cbcontainersv1.CBContainersAgent)) {
	mock.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, item any, _ ...any) error {
			asCb, ok := item.(*cbcontainersv1.CBContainersAgent)
			require.True(t, ok)

			assert(asCb)
			return nil
		})
}

func setupValidatorAcceptAll(validator *mocks.MockChangeValidator) {
	validator.EXPECT().ValidateChange(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
}

// setupEmptySensorMetadata simulates the case where no sensor metadata is available; hence no feature toggles are enabled
func setupEmptySensorMetadata(api *mocks.MockApiGateway) {
	api.EXPECT().GetSensorMetadata().Return([]models.SensorMetadata{}, nil)
}
