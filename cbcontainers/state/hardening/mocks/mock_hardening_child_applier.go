// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vmware/cbcontainers-operator/cbcontainers/state/hardening (interfaces: HardeningChildK8sObjectApplier)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "github.com/vmware/cbcontainers-operator/api/v1"
	options "github.com/vmware/cbcontainers-operator/cbcontainers/state/applyment/options"
	hardening "github.com/vmware/cbcontainers-operator/cbcontainers/state/hardening"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockHardeningChildK8sObjectApplier is a mock of HardeningChildK8sObjectApplier interface.
type MockHardeningChildK8sObjectApplier struct {
	ctrl     *gomock.Controller
	recorder *MockHardeningChildK8sObjectApplierMockRecorder
}

// MockHardeningChildK8sObjectApplierMockRecorder is the mock recorder for MockHardeningChildK8sObjectApplier.
type MockHardeningChildK8sObjectApplierMockRecorder struct {
	mock *MockHardeningChildK8sObjectApplier
}

// NewMockHardeningChildK8sObjectApplier creates a new mock instance.
func NewMockHardeningChildK8sObjectApplier(ctrl *gomock.Controller) *MockHardeningChildK8sObjectApplier {
	mock := &MockHardeningChildK8sObjectApplier{ctrl: ctrl}
	mock.recorder = &MockHardeningChildK8sObjectApplierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHardeningChildK8sObjectApplier) EXPECT() *MockHardeningChildK8sObjectApplierMockRecorder {
	return m.recorder
}

// ApplyHardeningChildK8sObject mocks base method.
func (m *MockHardeningChildK8sObjectApplier) ApplyHardeningChildK8sObject(arg0 context.Context, arg1 *v1.CBContainersHardeningSpec, arg2 client.Client, arg3 hardening.HardeningChildK8sObject, arg4, arg5 string, arg6 ...*options.ApplyOptions) (bool, client.Object, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2, arg3, arg4, arg5}
	for _, a := range arg6 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ApplyHardeningChildK8sObject", varargs...)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(client.Object)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ApplyHardeningChildK8sObject indicates an expected call of ApplyHardeningChildK8sObject.
func (mr *MockHardeningChildK8sObjectApplierMockRecorder) ApplyHardeningChildK8sObject(arg0, arg1, arg2, arg3, arg4, arg5 interface{}, arg6 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2, arg3, arg4, arg5}, arg6...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyHardeningChildK8sObject", reflect.TypeOf((*MockHardeningChildK8sObjectApplier)(nil).ApplyHardeningChildK8sObject), varargs...)
}

// DeleteK8sObjectIfExists mocks base method.
func (m *MockHardeningChildK8sObjectApplier) DeleteK8sObjectIfExists(arg0 context.Context, arg1 *v1.CBContainersHardeningSpec, arg2 client.Client, arg3 hardening.HardeningChildK8sObject) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteK8sObjectIfExists", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteK8sObjectIfExists indicates an expected call of DeleteK8sObjectIfExists.
func (mr *MockHardeningChildK8sObjectApplierMockRecorder) DeleteK8sObjectIfExists(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteK8sObjectIfExists", reflect.TypeOf((*MockHardeningChildK8sObjectApplier)(nil).DeleteK8sObjectIfExists), arg0, arg1, arg2, arg3)
}
