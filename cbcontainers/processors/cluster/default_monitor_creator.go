package cluster

import (
	"crypto/rand"
	"crypto/rsa"
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
	privateKey             *rsa.PrivateKey
}

func NewDefaultMonitorCreator(healthChecker monitor.HealthChecker, featuresStatusProvider monitor.FeaturesStatusProvider) (*DefaultMonitorCreator, error) {
	privateKey, err := makePrivateKey()
	if err != nil {
		return nil, err
	}

	return &DefaultMonitorCreator{
		healthChecker:          healthChecker,
		featuresStatusProvider: featuresStatusProvider,
		privateKey:             privateKey,
	}, nil
}

func makePrivateKey() (*rsa.PrivateKey, error) {
	reader := rand.Reader
	privateKey, err := rsa.GenerateKey(reader, 4096)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func (creator *DefaultMonitorCreator) CreateMonitor(cbContainersCluster *cbcontainersv1.CBContainersCluster, gateway Gateway) (Monitor, error) {
	spec := cbContainersCluster.Spec
	eventsSpec := cbContainersCluster.Spec.EventsGatewaySpec

	certPool, cert, err := gateway.GetCertificates("monitor-agent", creator.privateKey)
	if err != nil {
		return nil, err
	}

	reporter, err := reporters.NewGrpcMonitorReporter(eventsSpec.Host, eventsSpec.Port, certPool, cert)
	if err != nil {
		return nil, err
	}

	return monitor.NewMonitorAgent(spec.Account, spec.ClusterName, "", creator.healthChecker, creator.featuresStatusProvider, reporter, MonitorInterval), nil
}
