package monitor

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/vmware/cbcontainers-operator/cbcontainers/monitor/mocks"
	"github.com/vmware/cbcontainers-operator/cbcontainers/monitor/models"
	admissionsV1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"testing"
	"time"
)

const (
	Account      = "account_name"
	Cluster      = "cluster_name"
	Version      = "version"
	SendInterval = 2 * time.Second
)

type TestMonitorObjects struct {
	healthCheckerMock   *mocks.MockHealthChecker
	featuresMock        *mocks.MockFeaturesStatusProvider
	messageReporterMock *mocks.MockMessageReporter
}

type SetupFunc func(*MonitorAgent, *TestMonitorObjects) models.HealthReportMessage

func testMonitorAgent(t *testing.T, setup SetupFunc) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testObjects := &TestMonitorObjects{
		healthCheckerMock:   mocks.NewMockHealthChecker(ctrl),
		featuresMock:        mocks.NewMockFeaturesStatusProvider(ctrl),
		messageReporterMock: mocks.NewMockMessageReporter(ctrl),
	}
	agent := NewMonitorAgent(Account, Cluster, Version, testObjects.healthCheckerMock, testObjects.featuresMock, testObjects.messageReporterMock, SendInterval)

	expectedHealthReportMessage := setup(agent, testObjects)
	testObjects.messageReporterMock.EXPECT().SendMonitorMessage(expectedHealthReportMessage).Return(nil)
	testObjects.messageReporterMock.EXPECT().Close().Return(nil)

	agent.Start()
	time.Sleep(SendInterval)
	agent.Stop()
	time.Sleep(SendInterval)
}

func TestEnabledComponents(t *testing.T) {
	testEnabledComponents := func(t *testing.T, hardeningEnabled, runtimeEnabled bool) {
		testMonitorAgent(t, func(agent *MonitorAgent, testObjects *TestMonitorObjects) models.HealthReportMessage {
			testObjects.featuresMock.EXPECT().HardeningEnabled().Return(hardeningEnabled, nil)
			testObjects.featuresMock.EXPECT().RuntimeEnabled().Return(runtimeEnabled, nil)
			testObjects.healthCheckerMock.EXPECT().GetPods().Return(map[string]coreV1.Pod{}, nil)
			testObjects.healthCheckerMock.EXPECT().GetDaemonSets().Return(map[string]appsV1.DaemonSet{}, nil)
			testObjects.healthCheckerMock.EXPECT().GetDeployments().Return(map[string]appsV1.Deployment{}, nil)
			testObjects.healthCheckerMock.EXPECT().GetReplicaSets().Return(map[string]appsV1.ReplicaSet{}, nil)
			testObjects.healthCheckerMock.EXPECT().GetValidatingWebhookConfigurations().Return(map[string]admissionsV1beta1.ValidatingWebhookConfiguration{}, nil)

			return models.NewHealthReportMessage(Account, Cluster, Version, map[string]bool{HardeningFeature: hardeningEnabled, RuntimeFeature: runtimeEnabled}, map[string]models.WorkloadHealthReport{}, map[string]models.WebhookHealthReport{})
		})
	}

	possibleFlags := []bool{true, false}
	for _, hardeningEnabled := range possibleFlags {
		for _, runtimeEnabled := range possibleFlags {
			t.Run(fmt.Sprintf("When Hardening is set to %v and Runtime is set to %v, message should be built properly", hardeningEnabled, runtimeEnabled), func(t *testing.T) {
				testEnabledComponents(t, hardeningEnabled, runtimeEnabled)
			})
		}
	}
}
