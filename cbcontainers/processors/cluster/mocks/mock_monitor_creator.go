// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vmware/cbcontainers-operator/cbcontainers/processors/cluster (interfaces: MonitorCreator)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "github.com/vmware/cbcontainers-operator/api/v1"
	cluster "github.com/vmware/cbcontainers-operator/cbcontainers/processors/cluster"
)

// MockMonitorCreator is a mock of MonitorCreator interface.
type MockMonitorCreator struct {
	ctrl     *gomock.Controller
	recorder *MockMonitorCreatorMockRecorder
}

// MockMonitorCreatorMockRecorder is the mock recorder for MockMonitorCreator.
type MockMonitorCreatorMockRecorder struct {
	mock *MockMonitorCreator
}

// NewMockMonitorCreator creates a new mock instance.
func NewMockMonitorCreator(ctrl *gomock.Controller) *MockMonitorCreator {
	mock := &MockMonitorCreator{ctrl: ctrl}
	mock.recorder = &MockMonitorCreatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMonitorCreator) EXPECT() *MockMonitorCreatorMockRecorder {
	return m.recorder
}

// CreateMonitor mocks base method.
func (m *MockMonitorCreator) CreateMonitor(arg0 *v1.CBContainersCluster, arg1 cluster.Gateway) (cluster.Monitor, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMonitor", arg0, arg1)
	ret0, _ := ret[0].(cluster.Monitor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateMonitor indicates an expected call of CreateMonitor.
func (mr *MockMonitorCreatorMockRecorder) CreateMonitor(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMonitor", reflect.TypeOf((*MockMonitorCreator)(nil).CreateMonitor), arg0, arg1)
}
