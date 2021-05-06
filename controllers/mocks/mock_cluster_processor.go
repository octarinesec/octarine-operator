// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vmware/cbcontainers-operator/controllers (interfaces: ClusterProcessor)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "github.com/vmware/cbcontainers-operator/api/v1"
	models "github.com/vmware/cbcontainers-operator/cbcontainers/models"
)

// MockClusterProcessor is a mock of ClusterProcessor interface.
type MockClusterProcessor struct {
	ctrl     *gomock.Controller
	recorder *MockClusterProcessorMockRecorder
}

// MockClusterProcessorMockRecorder is the mock recorder for MockClusterProcessor.
type MockClusterProcessorMockRecorder struct {
	mock *MockClusterProcessor
}

// NewMockClusterProcessor creates a new mock instance.
func NewMockClusterProcessor(ctrl *gomock.Controller) *MockClusterProcessor {
	mock := &MockClusterProcessor{ctrl: ctrl}
	mock.recorder = &MockClusterProcessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClusterProcessor) EXPECT() *MockClusterProcessorMockRecorder {
	return m.recorder
}

// Process mocks base method.
func (m *MockClusterProcessor) Process(arg0 *v1.CBContainersCluster, arg1 string) (*models.RegistrySecretValues, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Process", arg0, arg1)
	ret0, _ := ret[0].(*models.RegistrySecretValues)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Process indicates an expected call of Process.
func (mr *MockClusterProcessorMockRecorder) Process(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Process", reflect.TypeOf((*MockClusterProcessor)(nil).Process), arg0, arg1)
}
