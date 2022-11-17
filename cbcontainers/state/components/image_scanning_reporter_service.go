package components

import (
	"fmt"

	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DesiredImageScanningReporterServicePortName  = "https"
	DesiredImageScanningReporterServicePortValue = 443
)

type ImageScanningReporterServiceK8sObject struct {
	// Namespace is the Namespace in which the Service will be created.
	Namespace string
}

func NewImageScanningReporterServiceK8sObject() *ImageScanningReporterServiceK8sObject {
	return &ImageScanningReporterServiceK8sObject{
		Namespace: commonState.DataPlaneNamespaceName,
	}
}

func (obj *ImageScanningReporterServiceK8sObject) EmptyK8sObject() client.Object {
	return &coreV1.Service{}
}

func (obj *ImageScanningReporterServiceK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: ImageScanningReporterName, Namespace: obj.Namespace}
}

func (obj *ImageScanningReporterServiceK8sObject) MutateK8sObject(k8sObject client.Object, agentSpec *cbcontainersv1.CBContainersAgentSpec) error {
	service, ok := k8sObject.(*coreV1.Service)
	if !ok {
		return fmt.Errorf("expected Service K8s object")
	}

	imageScanningReporter := &agentSpec.Components.ClusterScanning.ImageScanningReporter

	service.Namespace = agentSpec.Namespace
	service.Labels = imageScanningReporter.Labels
	service.Spec.Type = coreV1.ServiceTypeClusterIP
	service.Spec.Selector = map[string]string{
		ImageScanningReporterLabelKey: ImageScanningReporterName,
	}
	obj.mutatePorts(service)

	return nil
}

func (obj *ImageScanningReporterServiceK8sObject) mutatePorts(service *coreV1.Service) {
	if service.Spec.Ports == nil || len(service.Spec.Ports) != 1 {
		service.Spec.Ports = []coreV1.ServicePort{{}}
	}

	service.Spec.Ports[0].Name = DesiredImageScanningReporterServicePortName
	service.Spec.Ports[0].TargetPort = intstr.FromString(ImageScanningReporterDesiredContainerPortName)
	service.Spec.Ports[0].Port = DesiredImageScanningReporterServicePortValue
}
