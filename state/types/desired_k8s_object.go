package types

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DesiredK8sObject interface {
	GetNamespacedName() types.NamespacedName
	GetK8sObject() client.Object
	GetEmptyK8sObject() client.Object
	VerifyK8sObject(client.Object) (bool, error)
}
