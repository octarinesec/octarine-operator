package components

import (
	"fmt"

	cbcontainersv1 "github.com/vmware/cbcontainers-operator/api/v1"
	"github.com/vmware/cbcontainers-operator/cbcontainers/models"
	commonState "github.com/vmware/cbcontainers-operator/cbcontainers/state/common"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ImageScanningReporterTlsName = "cbcontainers-hardening-image-scanning-reporter-tls"
)

type tlsSecretsValuesCreator interface {
	CreateTlsSecretsValues(resourceNamespacedName types.NamespacedName) (models.TlsSecretValues, error)
}

type ImageScanningReporterTlsK8sObject struct {
	tlsSecretsValuesCreator tlsSecretsValuesCreator
}

func NewImageScanningReporterTlsK8sObject(tlsSecretsValuesCreator tlsSecretsValuesCreator) *ImageScanningReporterTlsK8sObject {
	return &ImageScanningReporterTlsK8sObject{
		tlsSecretsValuesCreator: tlsSecretsValuesCreator,
	}
}

func (obj *ImageScanningReporterTlsK8sObject) EmptyK8sObject() client.Object {
	return &coreV1.Secret{}
}

func (obj *ImageScanningReporterTlsK8sObject) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: ImageScanningReporterTlsName, Namespace: commonState.DataPlaneNamespaceName}
}

func (obj *ImageScanningReporterTlsK8sObject) MutateK8sObject(k8sObject client.Object, _ *cbcontainersv1.CBContainersAgentSpec) error {
	secret, ok := k8sObject.(*coreV1.Secret)
	if !ok {
		return fmt.Errorf("expected Secret K8s object")
	}

	tlsSecretValues, err := obj.tlsSecretsValuesCreator.CreateTlsSecretsValues(types.NamespacedName{Name: ImageScanningReporterName, Namespace: commonState.DataPlaneNamespaceName})
	if err != nil {
		return err
	}

	secret.Data = tlsSecretValues.ToDataMap()

	return nil
}
