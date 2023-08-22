package config_applier_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	k8sMocks "github.com/vmware/cbcontainers-operator/cbcontainers/test_utils/mocks"
	"github.com/vmware/cbcontainers-operator/config_applier"
	mocksConfigApplier "github.com/vmware/cbcontainers-operator/config_applier/mocks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

// TODO: Compatibility checks
// TODO: Adding CNDR to the config options
// TODO: Properly handle version + custom image to override the custom image
// TODO: Check fields are applied to CR correctly

// TODO: Reads cluster, etc from CR correctly?
// TODO: Respects proxy
// TODO: Updates img with version
// TODO: Review gomock.any usages here
// TODO: version tests

var (
	trueV    = true
	truePtr  = &trueV
	falseV   = false
	falsePtr = &falseV
)

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

func TestCRFieldsAreChangedCorrectlyBasedOnRemoteChange(t *testing.T) {
	type appliedChangeTest struct {
		name          string
		change        config_applier.ConfigurationChange
		initialCR     cbcontainersv1.CBContainersAgent
		assertFinalCR func(*testing.T, *cbcontainersv1.CBContainersAgent)
	}

	// generateFeatureToggleTestCases produces a set of tests for a single feature toggle in the requested change
	// The tests validate if each toggle state (true, false, nil) is applied correctly or ignored when it's not needed against the CR's state (true, false, nil)
	generateFeatureToggleTestCases :=
		func(feature string,
			changeFieldSelector func(*config_applier.ConfigurationChange) **bool,
			crFieldSelector func(agent *cbcontainersv1.CBContainersAgent) **bool) []appliedChangeTest {

			var result []appliedChangeTest

			for _, crState := range []*bool{truePtr, falsePtr, nil} {
				cr := cbcontainersv1.CBContainersAgent{}
				crFieldPtr := crFieldSelector(&cr)
				*crFieldPtr = crState

				// Validate that each toggle state works (or doesn't do anything when it matches)
				for _, changeState := range []*bool{truePtr, falsePtr} {
					change := createPendingChange()
					changeFieldPtr := changeFieldSelector(&change)
					*changeFieldPtr = changeState

					expectedState := changeState // avoid closure issues
					result = append(result, appliedChangeTest{
						name:      fmt.Sprintf("toggle feature (%s) from (%v) to (%v)", feature, prettyPrintBoolPtr(crState), prettyPrintBoolPtr(changeState)),
						change:    change,
						initialCR: cr,
						assertFinalCR: func(t *testing.T, agent *cbcontainersv1.CBContainersAgent) {
							crFieldPostChangePtr := crFieldSelector(agent)
							assert.Equal(t, expectedState, *crFieldPostChangePtr)
						},
					})
				}

				// Validate that a change with the toggle unset does not modify the CR
				result = append(result, appliedChangeTest{
					name:      fmt.Sprintf("missing toggle feature (%s) with CR state (%v)", feature, prettyPrintBoolPtr(crState)),
					change:    createPendingChange(),
					initialCR: cr,
					assertFinalCR: func(t *testing.T, agent *cbcontainersv1.CBContainersAgent) {
						crFieldPostChangePtr := crFieldSelector(agent)
						assert.Equal(t, *crFieldPtr, *crFieldPostChangePtr)
					},
				})
			}

			return result
		}

	var testCases []appliedChangeTest

	clusterScannerToggleTestCases := generateFeatureToggleTestCases("cluster scanning",
		func(change *config_applier.ConfigurationChange) **bool {
			return &change.EnableClusterScanning
		}, func(agent *cbcontainersv1.CBContainersAgent) **bool {
			return &agent.Spec.Components.ClusterScanning.Enabled
		})

	runtimeToggleTestCases := generateFeatureToggleTestCases("runtime protection",
		func(change *config_applier.ConfigurationChange) **bool {
			return &change.EnableRuntime
		}, func(agent *cbcontainersv1.CBContainersAgent) **bool {
			return &agent.Spec.Components.RuntimeProtection.Enabled
		})

	cndrToggleTestCases := generateFeatureToggleTestCases("CNDR",
		func(change *config_applier.ConfigurationChange) **bool {
			return &change.EnableCNDR
		}, func(agent *cbcontainersv1.CBContainersAgent) **bool {
			if agent.Spec.Components.Cndr == nil {
				agent.Spec.Components.Cndr = &cbcontainersv1.CBContainersCndrSpec{}
			}
			return &agent.Spec.Components.Cndr.Enabled
		})

	testCases = append(testCases, clusterScannerToggleTestCases...)
	testCases = append(testCases, runtimeToggleTestCases...)
	testCases = append(testCases, cndrToggleTestCases...)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			applier, mocks := setupApplier(ctrl)

			changesList := []config_applier.ConfigurationChange{testCase.change}
			cr := testCase.initialCR

			mocks.api.EXPECT().GetConfigurationChanges(gomock.Any()).Return(changesList, nil)
			mocks.k8sClient.EXPECT().List(gomock.Any(), &cbcontainersv1.CBContainersAgentList{}).
				Do(func(ctx context.Context, list *cbcontainersv1.CBContainersAgentList, _ ...any) {
					list.Items = []cbcontainersv1.CBContainersAgent{cr}
				})

			mocks.k8sClient.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, item any, _ ...any) error {
					asCb, ok := item.(*cbcontainersv1.CBContainersAgent)
					require.True(t, ok)

					testCase.assertFinalCR(t, asCb)
					return nil
				})

			mocks.api.EXPECT().UpdateConfigurationChangeStatus(gomock.Any(), gomock.Any()).Return(nil)

			assert.NoError(t, applier.RunIteration(context.Background()))
		})
	}
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

func createPendingChange() config_applier.ConfigurationChange {
	return config_applier.ConfigurationChange{
		ID:     strconv.Itoa(rand.Int()),
		Status: "PENDING",
	}
}

func prettyPrintBoolPtr(v *bool) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%t", *v)
}
