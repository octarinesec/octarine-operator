package monitor_test

import (
	"encoding/json"
	"fmt"
	logrTesting "github.com/go-logr/logr/testing"
	"github.com/golang/mock/gomock"
	"github.com/vmware/cbcontainers-operator/cbcontainers/monitor"
	"github.com/vmware/cbcontainers-operator/cbcontainers/monitor/mocks"
	"github.com/vmware/cbcontainers-operator/cbcontainers/monitor/models"
	hardeningObjects "github.com/vmware/cbcontainers-operator/cbcontainers/state/hardening/objects"
	"github.com/vmware/cbcontainers-operator/cbcontainers/test_utils"
	admissionsV1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

type AgentTestSetupFunc func(*monitor.MonitorAgent, *TestMonitorObjects) (models.HealthReportMessage, error)

func testMonitorAgent(t *testing.T, setup AgentTestSetupFunc) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testObjects := &TestMonitorObjects{
		healthCheckerMock:   mocks.NewMockHealthChecker(ctrl),
		featuresMock:        mocks.NewMockFeaturesStatusProvider(ctrl),
		messageReporterMock: mocks.NewMockMessageReporter(ctrl),
	}
	agent := monitor.NewMonitorAgent(Account, Cluster, Version, testObjects.healthCheckerMock, testObjects.featuresMock, testObjects.messageReporterMock, SendInterval, &logrTesting.TestLogger{T: t})

	expectedHealthReportMessage, err := setup(agent, testObjects)
	if err == nil {
		testObjects.messageReporterMock.EXPECT().SendMonitorMessage(expectedHealthReportMessage).Return(nil)
	}

	testObjects.messageReporterMock.EXPECT().Close().Return(nil)

	agent.Start()
	time.Sleep(SendInterval)
	agent.Stop()
	time.Sleep(SendInterval)
}

func TestWorkloadsReports(t *testing.T) {
	testWorkloadsReports := func(t *testing.T, expectedDeployments map[string]appsV1.Deployment, expectedDaemonSets map[string]appsV1.DaemonSet, expectedReplicaSets map[string]appsV1.ReplicaSet, expectedPods map[string]coreV1.Pod, expectedWorkloadHealthReports map[string]models.WorkloadHealthReport) {
		testMonitorAgent(t, func(agent *monitor.MonitorAgent, testObjects *TestMonitorObjects) (models.HealthReportMessage, error) {
			testObjects.featuresMock.EXPECT().HardeningEnabled().Return(true, nil)
			testObjects.featuresMock.EXPECT().RuntimeEnabled().Return(true, nil)
			testObjects.healthCheckerMock.EXPECT().GetPods().Return(expectedPods, nil)
			testObjects.healthCheckerMock.EXPECT().GetDaemonSets().Return(expectedDaemonSets, nil)
			testObjects.healthCheckerMock.EXPECT().GetDeployments().Return(expectedDeployments, nil)
			testObjects.healthCheckerMock.EXPECT().GetReplicaSets().Return(expectedReplicaSets, nil)
			testObjects.healthCheckerMock.EXPECT().GetValidatingWebhookConfigurations().Return(map[string]admissionsV1beta1.ValidatingWebhookConfiguration{}, nil)

			return models.NewHealthReportMessage(Account, Cluster, Version, map[string]bool{monitor.HardeningFeature: true, monitor.RuntimeFeature: true}, expectedWorkloadHealthReports, map[string]models.WebhookHealthReport{}), nil
		})
	}

	workloadName := test_utils.RandomString()
	workloadLabels := test_utils.RandomLabels()
	workloadMeta := metav1.ObjectMeta{Name: workloadName, Labels: workloadLabels}
	replicaSetName := fmt.Sprintf("%v-%v", workloadName, test_utils.RandomString())
	podName := fmt.Sprintf("%v-%v", replicaSetName, test_utils.RandomString())
	nodeName := test_utils.RandomString()

	replicaCount := int32(2)
	readyReplicaCount := int32(1)
	deploymentStatus := appsV1.DeploymentStatus{Replicas: replicaCount, ReadyReplicas: readyReplicaCount}
	deploymentStatusBytes, _ := json.Marshal(deploymentStatus)
	daemonSetStatus := appsV1.DaemonSetStatus{DesiredNumberScheduled: replicaCount, NumberReady: readyReplicaCount}
	daemonSetStatusBytes, _ := json.Marshal(daemonSetStatus)
	podStatus := coreV1.PodStatus{Message: test_utils.RandomString()}
	podStatusBytes, _ := json.Marshal(podStatus)

	firstContainer := test_utils.RandomString()
	firstImage := test_utils.RandomString()
	secondContainer := test_utils.RandomString()
	secondImage := test_utils.RandomString()
	podTemplateSpec := coreV1.PodTemplateSpec{
		Spec: coreV1.PodSpec{
			Containers: []coreV1.Container{
				{Name: firstContainer, Image: firstImage},
				{Name: secondContainer, Image: secondImage},
			},
		},
	}

	createDeployments := func() map[string]appsV1.Deployment {
		return map[string]appsV1.Deployment{
			workloadName: {
				ObjectMeta: workloadMeta, Status: deploymentStatus,
				Spec: appsV1.DeploymentSpec{Replicas: &replicaCount, Template: podTemplateSpec},
			},
		}
	}

	createDaemonSets := func() map[string]appsV1.DaemonSet {
		return map[string]appsV1.DaemonSet{
			workloadName: {
				ObjectMeta: workloadMeta, Status: daemonSetStatus,
				Spec: appsV1.DaemonSetSpec{Template: podTemplateSpec},
			},
		}
	}

	createReplicaSets := func(ownerKind string) map[string]appsV1.ReplicaSet {
		return map[string]appsV1.ReplicaSet{
			replicaSetName: {
				ObjectMeta: metav1.ObjectMeta{
					Name:            replicaSetName,
					OwnerReferences: []metav1.OwnerReference{{Kind: ownerKind, Name: workloadName}}},
			},
		}
	}

	createPods := func() map[string]coreV1.Pod {
		expectedPodSpec := coreV1.PodSpec{
			NodeName:   nodeName,
			Containers: []coreV1.Container{{Name: secondContainer, Image: secondImage}},
		}
		return map[string]coreV1.Pod{
			podName: {
				ObjectMeta: metav1.ObjectMeta{
					Name:            podName,
					OwnerReferences: []metav1.OwnerReference{{Kind: "ReplicaSet", Name: replicaSetName}},
				},
				Status: podStatus,
				Spec:   expectedPodSpec,
			},
		}
	}

	createExpectedWorkloadHealthReports := func(kind models.WorkloadKind, replicasReports map[string]models.WorkloadReplicaHealthReport) map[string]models.WorkloadHealthReport {
		spec := models.WorkloadSpecReport{
			Containers: map[string]models.ContainerHealthReport{
				firstContainer:  {Image: firstImage},
				secondContainer: {Image: secondImage},
			},
		}
		var statusBytes []byte

		if kind == models.WorkloadKindDeployment {
			spec.Replicas = replicaCount
			statusBytes = deploymentStatusBytes
		} else if kind == models.WorkloadKindDaemonSet {
			spec.Replicas = 0
			statusBytes = daemonSetStatusBytes
		} else {
			panic("Unsupported workload kind on test")
		}

		return map[string]models.WorkloadHealthReport{
			workloadName: {
				Kind:            kind,
				Labels:          workloadLabels,
				Status:          statusBytes,
				Spec:            spec,
				ReplicasReports: replicasReports,
			},
		}
	}

	createExpectedWorkLoadReplicaReports := func() map[string]models.WorkloadReplicaHealthReport {
		return map[string]models.WorkloadReplicaHealthReport{
			podName: {
				Node:   nodeName,
				Status: podStatusBytes,
				Spec:   models.WorkloadReplicaSpecReport{Containers: map[string]models.ContainerHealthReport{secondContainer: {Image: secondImage}}},
			},
		}
	}

	t.Run("When there are no workloads, there should be no workloads reports", func(t *testing.T) {
		testWorkloadsReports(t, map[string]appsV1.Deployment{}, map[string]appsV1.DaemonSet{}, map[string]appsV1.ReplicaSet{}, map[string]coreV1.Pod{}, map[string]models.WorkloadHealthReport{})
	})

	t.Run("When there is a deployment with no replicas, it should be reported with no replicas", func(t *testing.T) {
		expectedWorkloadReport := createExpectedWorkloadHealthReports(models.WorkloadKindDeployment, map[string]models.WorkloadReplicaHealthReport{})
		testWorkloadsReports(t, createDeployments(), map[string]appsV1.DaemonSet{}, map[string]appsV1.ReplicaSet{}, map[string]coreV1.Pod{}, expectedWorkloadReport)
	})

	t.Run("When there is a daemonSet with no replicas, it should be reported with no replicas", func(t *testing.T) {
		expectedWorkloadReport := createExpectedWorkloadHealthReports(models.WorkloadKindDaemonSet, map[string]models.WorkloadReplicaHealthReport{})
		testWorkloadsReports(t, map[string]appsV1.Deployment{}, createDaemonSets(), map[string]appsV1.ReplicaSet{}, map[string]coreV1.Pod{}, expectedWorkloadReport)
	})

	t.Run("When there is a deployment with replicaset but no pods, it should be reported with no replicas", func(t *testing.T) {
		expectedWorkloadReport := createExpectedWorkloadHealthReports(models.WorkloadKindDeployment, map[string]models.WorkloadReplicaHealthReport{})
		testWorkloadsReports(t, createDeployments(), map[string]appsV1.DaemonSet{}, createReplicaSets("Deployment"), map[string]coreV1.Pod{}, expectedWorkloadReport)
	})

	t.Run("When there is a daemonSet with replicaset but no pods, it should be reported with no replicas", func(t *testing.T) {
		expectedWorkloadReport := createExpectedWorkloadHealthReports(models.WorkloadKindDaemonSet, map[string]models.WorkloadReplicaHealthReport{})
		testWorkloadsReports(t, map[string]appsV1.Deployment{}, createDaemonSets(), createReplicaSets("DaemonSet"), map[string]coreV1.Pod{}, expectedWorkloadReport)
	})

	t.Run("When there is a deployment with replicas, it should be reported with replicas", func(t *testing.T) {
		expectedWorkloadReport := createExpectedWorkloadHealthReports(models.WorkloadKindDeployment, createExpectedWorkLoadReplicaReports())
		testWorkloadsReports(t, createDeployments(), map[string]appsV1.DaemonSet{}, createReplicaSets("Deployment"), createPods(), expectedWorkloadReport)
	})

	t.Run("When there is a daemonSet with replicas, it should be reported with replicas", func(t *testing.T) {
		expectedWorkloadReport := createExpectedWorkloadHealthReports(models.WorkloadKindDaemonSet, createExpectedWorkLoadReplicaReports())
		testWorkloadsReports(t, map[string]appsV1.Deployment{}, createDaemonSets(), createReplicaSets("DaemonSet"), createPods(), expectedWorkloadReport)
	})
}

func TestWebhookReports(t *testing.T) {
	testWebhookReports := func(t *testing.T, expectedValidatingWebhooks map[string]admissionsV1beta1.ValidatingWebhookConfiguration, expectedWebhookReports map[string]models.WebhookHealthReport) {
		testMonitorAgent(t, func(agent *monitor.MonitorAgent, testObjects *TestMonitorObjects) (models.HealthReportMessage, error) {
			testObjects.featuresMock.EXPECT().HardeningEnabled().Return(true, nil)
			testObjects.featuresMock.EXPECT().RuntimeEnabled().Return(true, nil)
			testObjects.healthCheckerMock.EXPECT().GetPods().Return(map[string]coreV1.Pod{}, nil)
			testObjects.healthCheckerMock.EXPECT().GetDaemonSets().Return(map[string]appsV1.DaemonSet{}, nil)
			testObjects.healthCheckerMock.EXPECT().GetDeployments().Return(map[string]appsV1.Deployment{}, nil)
			testObjects.healthCheckerMock.EXPECT().GetReplicaSets().Return(map[string]appsV1.ReplicaSet{}, nil)
			testObjects.healthCheckerMock.EXPECT().GetValidatingWebhookConfigurations().Return(expectedValidatingWebhooks, nil)

			return models.NewHealthReportMessage(Account, Cluster, Version, map[string]bool{monitor.HardeningFeature: true, monitor.RuntimeFeature: true}, map[string]models.WorkloadHealthReport{}, expectedWebhookReports), nil
		})
	}

	t.Run("When there are no validating web hooks, there should be no reports", func(t *testing.T) {
		testWebhookReports(t, map[string]admissionsV1beta1.ValidatingWebhookConfiguration{}, map[string]models.WebhookHealthReport{})
	})

	t.Run("When there are validating web hooks but no enforcer web hook, there should be no reports", func(t *testing.T) {
		testWebhookReports(t, map[string]admissionsV1beta1.ValidatingWebhookConfiguration{
			test_utils.RandomString(): {
				Webhooks: []admissionsV1beta1.ValidatingWebhook{{Name: test_utils.RandomString()}},
			},
		}, map[string]models.WebhookHealthReport{})
	})

	t.Run("When there are validating web hooks including the enforcer web hook, there should be one report with the enforcer webhook", func(t *testing.T) {
		uid := test_utils.RandomString()
		testWebhookReports(t, map[string]admissionsV1beta1.ValidatingWebhookConfiguration{
			test_utils.RandomString(): {
				Webhooks: []admissionsV1beta1.ValidatingWebhook{{Name: test_utils.RandomString()}},
			},
			hardeningObjects.EnforcerName: {
				ObjectMeta: metav1.ObjectMeta{
					Name: hardeningObjects.EnforcerName,
					UID:  types.UID(uid),
				},
				Webhooks: []admissionsV1beta1.ValidatingWebhook{{Name: test_utils.RandomString()}},
			},
		}, map[string]models.WebhookHealthReport{
			hardeningObjects.EnforcerName: {
				Type: models.WebhookTypeValidating,
				Uid:  uid,
			},
		})
	})
}

func TestEnabledComponents(t *testing.T) {
	testEnabledComponents := func(t *testing.T, hardeningEnabled, runtimeEnabled bool) {
		testMonitorAgent(t, func(agent *monitor.MonitorAgent, testObjects *TestMonitorObjects) (models.HealthReportMessage, error) {
			testObjects.featuresMock.EXPECT().HardeningEnabled().Return(hardeningEnabled, nil)
			testObjects.featuresMock.EXPECT().RuntimeEnabled().Return(runtimeEnabled, nil)
			testObjects.healthCheckerMock.EXPECT().GetPods().Return(map[string]coreV1.Pod{}, nil)
			testObjects.healthCheckerMock.EXPECT().GetDaemonSets().Return(map[string]appsV1.DaemonSet{}, nil)
			testObjects.healthCheckerMock.EXPECT().GetDeployments().Return(map[string]appsV1.Deployment{}, nil)
			testObjects.healthCheckerMock.EXPECT().GetReplicaSets().Return(map[string]appsV1.ReplicaSet{}, nil)
			testObjects.healthCheckerMock.EXPECT().GetValidatingWebhookConfigurations().Return(map[string]admissionsV1beta1.ValidatingWebhookConfiguration{}, nil)

			return models.NewHealthReportMessage(Account, Cluster, Version, map[string]bool{monitor.HardeningFeature: hardeningEnabled, monitor.RuntimeFeature: runtimeEnabled}, map[string]models.WorkloadHealthReport{}, map[string]models.WebhookHealthReport{}), nil
		})
	}

	t.Run("When can't get hardening enabled status, log error", func(t *testing.T) {
		testMonitorAgent(t, func(agent *monitor.MonitorAgent, testObjects *TestMonitorObjects) (models.HealthReportMessage, error) {
			err := fmt.Errorf("mock-error-for-getting-hardening")
			testObjects.featuresMock.EXPECT().HardeningEnabled().Return(false, err)
			return models.HealthReportMessage{}, err
		})
	})

	t.Run("When can't get runtime enabled status, log error", func(t *testing.T) {
		testMonitorAgent(t, func(agent *monitor.MonitorAgent, testObjects *TestMonitorObjects) (models.HealthReportMessage, error) {
			err := fmt.Errorf("mock-error-for-getting-runtime")
			testObjects.featuresMock.EXPECT().HardeningEnabled().Return(true, nil)
			testObjects.featuresMock.EXPECT().RuntimeEnabled().Return(false, err)
			return models.HealthReportMessage{}, err
		})
	})

	possibleFlags := []bool{true, false}
	for _, hardeningEnabled := range possibleFlags {
		for _, runtimeEnabled := range possibleFlags {
			t.Run(fmt.Sprintf("When Hardening is set to %v and Runtime is set to %v, message should be built properly", hardeningEnabled, runtimeEnabled), func(t *testing.T) {
				testEnabledComponents(t, hardeningEnabled, runtimeEnabled)
			})
		}
	}
}
