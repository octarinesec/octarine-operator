package types

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DesiredK8sObject interface {
	NamespacedName() types.NamespacedName
	EmptyK8sObject() client.Object
	MutateK8sObject(client.Object) (bool, error)
}
