package cluster

import (
	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/monitor"
	"github.com/vmware/cbcontainers-operator/cbcontainers/monitor/reporters"
	"time"
)

const (
	MonitorInterval = 20 * time.Second
)

type DefaultMonitorCreator struct {
	healthChecker          monitor.HealthChecker
	featuresStatusProvider monitor.FeaturesStatusProvider
}

func NewDefaultMonitorCreator(healthChecker monitor.HealthChecker, featuresStatusProvider monitor.FeaturesStatusProvider) *DefaultMonitorCreator {
	return &DefaultMonitorCreator{
		healthChecker:          healthChecker,
		featuresStatusProvider: featuresStatusProvider,
	}
}

func (creator *DefaultMonitorCreator) CreateMonitor(cbContainersCluster *cbcontainersv1.CBContainersCluster, gateway Gateway) (Monitor, error) {
	spec := cbContainersCluster.Spec
	eventsSpec := cbContainersCluster.Spec.EventsGatewaySpec

	certPool, cert, err := gateway.GetCertificates("")
	if err != nil {
		return nil, err
	}

	reporter, err := reporters.NewGrpcMonitorReporter(eventsSpec.Host, eventsSpec.Port, certPool, cert)
	if err != nil {
		return nil, err
	}

	return monitor.NewMonitorAgent(spec.Account, spec.ClusterName, "", creator.healthChecker, creator.featuresStatusProvider, reporter, MonitorInterval), nil
}
