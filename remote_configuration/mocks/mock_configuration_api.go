// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vmware/cbcontainers-operator/remote_configuration (interfaces: ConfigurationChangesAPI)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	remote_configuration "github.com/vmware/cbcontainers-operator/remote_configuration"
)

// MockConfigurationChangesAPI is a mock of ConfigurationChangesAPI interface.
type MockConfigurationChangesAPI struct {
	ctrl     *gomock.Controller
	recorder *MockConfigurationChangesAPIMockRecorder
}

// MockConfigurationChangesAPIMockRecorder is the mock recorder for MockConfigurationChangesAPI.
type MockConfigurationChangesAPIMockRecorder struct {
	mock *MockConfigurationChangesAPI
}

// NewMockConfigurationChangesAPI creates a new mock instance.
func NewMockConfigurationChangesAPI(ctrl *gomock.Controller) *MockConfigurationChangesAPI {
	mock := &MockConfigurationChangesAPI{ctrl: ctrl}
	mock.recorder = &MockConfigurationChangesAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConfigurationChangesAPI) EXPECT() *MockConfigurationChangesAPIMockRecorder {
	return m.recorder
}

// GetConfigurationChanges mocks base method.
func (m *MockConfigurationChangesAPI) GetConfigurationChanges(arg0 context.Context) ([]remote_configuration.ConfigurationChange, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConfigurationChanges", arg0)
	ret0, _ := ret[0].([]remote_configuration.ConfigurationChange)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetConfigurationChanges indicates an expected call of GetConfigurationChanges.
func (mr *MockConfigurationChangesAPIMockRecorder) GetConfigurationChanges(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConfigurationChanges", reflect.TypeOf((*MockConfigurationChangesAPI)(nil).GetConfigurationChanges), arg0)
}

// UpdateConfigurationChangeStatus mocks base method.
func (m *MockConfigurationChangesAPI) UpdateConfigurationChangeStatus(arg0 context.Context, arg1 remote_configuration.ConfigurationChangeStatusUpdate) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateConfigurationChangeStatus", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateConfigurationChangeStatus indicates an expected call of UpdateConfigurationChangeStatus.
func (mr *MockConfigurationChangesAPIMockRecorder) UpdateConfigurationChangeStatus(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateConfigurationChangeStatus", reflect.TypeOf((*MockConfigurationChangesAPI)(nil).UpdateConfigurationChangeStatus), arg0, arg1)
}
