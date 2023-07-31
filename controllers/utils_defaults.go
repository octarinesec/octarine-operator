package controllers

import (
	"net"
	"os"
	"strings"

	"github.com/vmware/cbcontainers-operator/api/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	defaultAccessToken = "cbcontainers-access-token"
	defaultCompanyCode = "cbcontainers-company-code"
)

var trueRef = true
var falseRef = false

func setDefaultProbesCommon(probesSpec *v1.CBContainersCommonProbesSpec) {
	if probesSpec.InitialDelaySeconds == 0 {
		probesSpec.InitialDelaySeconds = 3
	}

	if probesSpec.TimeoutSeconds == 0 {
		probesSpec.TimeoutSeconds = 1
	}

	if probesSpec.PeriodSeconds == 0 {
		probesSpec.PeriodSeconds = 30
	}

	if probesSpec.SuccessThreshold == 0 {
		probesSpec.SuccessThreshold = 1
	}

	if probesSpec.FailureThreshold == 0 {
		probesSpec.FailureThreshold = 3
	}
}

func setDefaultHTTPProbes(probesSpec *v1.CBContainersHTTPProbesSpec) {
	if probesSpec.ReadinessPath == "" {
		probesSpec.ReadinessPath = "/ready"
	}

	if probesSpec.LivenessPath == "" {
		probesSpec.LivenessPath = "/alive"
	}

	if probesSpec.Port == 0 {
		probesSpec.Port = 8181
	}

	if probesSpec.Scheme == "" {
		probesSpec.Scheme = coreV1.URISchemeHTTP
	}

	setDefaultProbesCommon(&probesSpec.CBContainersCommonProbesSpec)
}

func setDefaultFileProbes(probesSpec *v1.CBContainersFileProbesSpec) {
	if probesSpec.ReadinessPath == "" {
		probesSpec.ReadinessPath = "/tmp/ready"
	}

	if probesSpec.LivenessPath == "" {
		probesSpec.LivenessPath = "/tmp/alive"
	}

	setDefaultProbesCommon(&probesSpec.CBContainersCommonProbesSpec)
}

func setDefaultPrometheus(prometheusSpec *v1.CBContainersPrometheusSpec) {
	setDefaultPrometheusWithPort(prometheusSpec, 7071)
}

func setDefaultPrometheusWithPort(prometheusSpec *v1.CBContainersPrometheusSpec, port int) {
	if prometheusSpec.Enabled == nil {
		prometheusSpec.Enabled = &falseRef
	}

	if prometheusSpec.Port == 0 {
		prometheusSpec.Port = port
	}
}

func setDefaultImage(imageSpec *v1.CBContainersImageSpec, imageName string) {
	if imageSpec.Repository == "" {
		imageSpec.Repository = imageName
	}

	if imageSpec.PullPolicy == "" {
		imageSpec.PullPolicy = "IfNotPresent"
	}
}

func setDefaultResourceRequirements(resources *coreV1.ResourceRequirements, requestMemory, requestCpu, limitMemory, limitCpu string) error {
	if resources.Requests == nil {
		resources.Requests = make(coreV1.ResourceList)
	}

	if err := setDefaultsResourcesList(resources.Requests, requestMemory, requestCpu); err != nil {
		return err
	}

	if resources.Limits == nil {
		resources.Limits = make(coreV1.ResourceList)
	}

	if err := setDefaultsResourcesList(resources.Limits, limitMemory, limitCpu); err != nil {
		return err
	}

	return nil
}

func setDefaultsResourcesList(list coreV1.ResourceList, memory, cpu string) error {
	if err := setDefaultResource(list, coreV1.ResourceMemory, memory); err != nil {
		return err
	}

	if err := setDefaultResource(list, coreV1.ResourceCPU, cpu); err != nil {
		return err
	}

	return nil
}

func setDefaultResource(list coreV1.ResourceList, resourceName coreV1.ResourceName, value string) error {
	if _, ok := list[resourceName]; !ok {
		quantity, err := resource.ParseQuantity(value)
		if err != nil {
			return err
		}

		list[resourceName] = quantity
	}

	return nil
}

var netLookupHost = initNetLookupHost()

func initNetLookupHost() func(_ string) ([]string, error) {
	// If we are running in a unit test, we need a more
	// predictable implementation of netLookupHost
	if strings.HasSuffix(os.Args[0], ".test") {
		return func(_ string) ([]string, error) {
			return []string{"10.96.0.1"}, nil
		}
	}
	return net.LookupHost
}
