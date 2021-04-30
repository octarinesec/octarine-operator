package types

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DesiredK8sObject interface {
	DesiredK8sObjectInitializer
	MutatableK8sObject
}

type DesiredK8sObjectInitializer interface {
	EmptyK8sObject() client.Object
}

type MutatableK8sObject interface {
	NamespacedName() types.NamespacedName
	MutateK8sObject(client.Object) error
}
