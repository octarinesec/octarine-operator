// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vmware/cbcontainers-operator/cbcontainers/processors (interfaces: APIGatewayCreator)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "github.com/vmware/cbcontainers-operator/api/v1"
	processors "github.com/vmware/cbcontainers-operator/cbcontainers/processors"
)

// MockAPIGatewayCreator is a mock of APIGatewayCreator interface.
type MockAPIGatewayCreator struct {
	ctrl     *gomock.Controller
	recorder *MockAPIGatewayCreatorMockRecorder
}

// MockAPIGatewayCreatorMockRecorder is the mock recorder for MockAPIGatewayCreator.
type MockAPIGatewayCreatorMockRecorder struct {
	mock *MockAPIGatewayCreator
}

// NewMockAPIGatewayCreator creates a new mock instance.
func NewMockAPIGatewayCreator(ctrl *gomock.Controller) *MockAPIGatewayCreator {
	mock := &MockAPIGatewayCreator{ctrl: ctrl}
	mock.recorder = &MockAPIGatewayCreatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAPIGatewayCreator) EXPECT() *MockAPIGatewayCreatorMockRecorder {
	return m.recorder
}

// CreateGateway mocks base method.
func (m *MockAPIGatewayCreator) CreateGateway(arg0 *v1.CBContainersAgent, arg1 string) (processors.APIGateway, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateGateway", arg0, arg1)
	ret0, _ := ret[0].(processors.APIGateway)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateGateway indicates an expected call of CreateGateway.
func (mr *MockAPIGatewayCreatorMockRecorder) CreateGateway(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateGateway", reflect.TypeOf((*MockAPIGatewayCreator)(nil).CreateGateway), arg0, arg1)
}
