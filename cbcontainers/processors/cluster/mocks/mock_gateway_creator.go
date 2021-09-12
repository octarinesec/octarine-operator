// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vmware/cbcontainers-operator/cbcontainers/processors/cluster (interfaces: GatewayCreator)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "github.com/vmware/cbcontainers-operator/api/v1"
	cluster "github.com/vmware/cbcontainers-operator/cbcontainers/processors/cluster"
)

// MockGatewayCreator is a mock of GatewayCreator interface.
type MockGatewayCreator struct {
	ctrl     *gomock.Controller
	recorder *MockGatewayCreatorMockRecorder
}

// MockGatewayCreatorMockRecorder is the mock recorder for MockGatewayCreator.
type MockGatewayCreatorMockRecorder struct {
	mock *MockGatewayCreator
}

// NewMockGatewayCreator creates a new mock instance.
func NewMockGatewayCreator(ctrl *gomock.Controller) *MockGatewayCreator {
	mock := &MockGatewayCreator{ctrl: ctrl}
	mock.recorder = &MockGatewayCreatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGatewayCreator) EXPECT() *MockGatewayCreatorMockRecorder {
	return m.recorder
}

// CreateGateway mocks base method.
func (m *MockGatewayCreator) CreateGateway(arg0 *v1.CBContainersAgent, arg1 string) (cluster.Gateway, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateGateway", arg0, arg1)
	ret0, _ := ret[0].(cluster.Gateway)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateGateway indicates an expected call of CreateGateway.
func (mr *MockGatewayCreatorMockRecorder) CreateGateway(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateGateway", reflect.TypeOf((*MockGatewayCreator)(nil).CreateGateway), arg0, arg1)
}
