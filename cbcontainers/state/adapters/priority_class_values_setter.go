package adapters

import (
	schedulingV1 "k8s.io/api/scheduling/v1"
	schedulingV1alpha1 "k8s.io/api/scheduling/v1alpha1"
	schedulingV1beta1 "k8s.io/api/scheduling/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func EmptyPriorityClassForVersion(k8sVersion string) client.Object {
	if k8sVersion == "" || k8sVersion >= "v1.14" {
		return &schedulingV1.PriorityClass{}
	} else if k8sVersion >= "v1.11" {
		return &schedulingV1beta1.PriorityClass{}
	}

	return &schedulingV1alpha1.PriorityClass{}
}

type PriorityClassValuesSetter interface {
	SetValue(value int32)
	SetGlobalDefault(globalDefault bool)
	SetDescription(description string)
}

func GetPriorityClassSetter(k8sObject client.Object) (PriorityClassValuesSetter, bool) {
	switch value := k8sObject.(type) {
	case *schedulingV1.PriorityClass:
		return (*PriorityClassV1Setter)(value), true
	case *schedulingV1beta1.PriorityClass:
		return (*PriorityClassV1beta1Setter)(value), true
	case *schedulingV1alpha1.PriorityClass:
		return (*PriorityClassV1alpha1Setter)(value), true
	}

	return nil, false
}

type PriorityClassV1Setter schedulingV1.PriorityClass

func (s *PriorityClassV1Setter) SetValue(value int32)              { s.Value = value }
func (s *PriorityClassV1Setter) SetGlobalDefault(global bool)      { s.GlobalDefault = global }
func (s *PriorityClassV1Setter) SetDescription(description string) { s.Description = description }

type PriorityClassV1beta1Setter schedulingV1beta1.PriorityClass

func (s *PriorityClassV1beta1Setter) SetValue(value int32)              { s.Value = value }
func (s *PriorityClassV1beta1Setter) SetGlobalDefault(global bool)      { s.GlobalDefault = global }
func (s *PriorityClassV1beta1Setter) SetDescription(description string) { s.Description = description }

type PriorityClassV1alpha1Setter schedulingV1alpha1.PriorityClass

func (s *PriorityClassV1alpha1Setter) SetValue(value int32)              { s.Value = value }
func (s *PriorityClassV1alpha1Setter) SetGlobalDefault(global bool)      { s.GlobalDefault = global }
func (s *PriorityClassV1alpha1Setter) SetDescription(description string) { s.Description = description }
