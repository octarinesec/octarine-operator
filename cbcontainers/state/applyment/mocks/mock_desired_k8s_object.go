// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment (interfaces: DesiredK8sObject)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	types "k8s.io/apimachinery/pkg/types"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockDesiredK8sObject is a mock of DesiredK8sObject interface.
type MockDesiredK8sObject struct {
	ctrl     *gomock.Controller
	recorder *MockDesiredK8sObjectMockRecorder
}

// MockDesiredK8sObjectMockRecorder is the mock recorder for MockDesiredK8sObject.
type MockDesiredK8sObjectMockRecorder struct {
	mock *MockDesiredK8sObject
}

// NewMockDesiredK8sObject creates a new mock instance.
func NewMockDesiredK8sObject(ctrl *gomock.Controller) *MockDesiredK8sObject {
	mock := &MockDesiredK8sObject{ctrl: ctrl}
	mock.recorder = &MockDesiredK8sObjectMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDesiredK8sObject) EXPECT() *MockDesiredK8sObjectMockRecorder {
	return m.recorder
}

// EmptyK8sObject mocks base method.
func (m *MockDesiredK8sObject) EmptyK8sObject() client.Object {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EmptyK8sObject")
	ret0, _ := ret[0].(client.Object)
	return ret0
}

// EmptyK8sObject indicates an expected call of EmptyK8sObject.
func (mr *MockDesiredK8sObjectMockRecorder) EmptyK8sObject() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EmptyK8sObject", reflect.TypeOf((*MockDesiredK8sObject)(nil).EmptyK8sObject))
}

// MutateK8sObject mocks base method.
func (m *MockDesiredK8sObject) MutateK8sObject(arg0 client.Object) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MutateK8sObject", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// MutateK8sObject indicates an expected call of MutateK8sObject.
func (mr *MockDesiredK8sObjectMockRecorder) MutateK8sObject(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MutateK8sObject", reflect.TypeOf((*MockDesiredK8sObject)(nil).MutateK8sObject), arg0)
}

// NamespacedName mocks base method.
func (m *MockDesiredK8sObject) NamespacedName() types.NamespacedName {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NamespacedName")
	ret0, _ := ret[0].(types.NamespacedName)
	return ret0
}

// NamespacedName indicates an expected call of NamespacedName.
func (mr *MockDesiredK8sObjectMockRecorder) NamespacedName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NamespacedName", reflect.TypeOf((*MockDesiredK8sObject)(nil).NamespacedName))
}
