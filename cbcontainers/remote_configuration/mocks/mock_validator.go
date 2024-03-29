// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vmware/cbcontainers-operator/cbcontainers/remote_configuration (interfaces: ChangeValidator)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "github.com/vmware/cbcontainers-operator/api/v1"
	models "github.com/vmware/cbcontainers-operator/cbcontainers/models"
)

// MockChangeValidator is a mock of ChangeValidator interface.
type MockChangeValidator struct {
	ctrl     *gomock.Controller
	recorder *MockChangeValidatorMockRecorder
}

// MockChangeValidatorMockRecorder is the mock recorder for MockChangeValidator.
type MockChangeValidatorMockRecorder struct {
	mock *MockChangeValidator
}

// NewMockChangeValidator creates a new mock instance.
func NewMockChangeValidator(ctrl *gomock.Controller) *MockChangeValidator {
	mock := &MockChangeValidator{ctrl: ctrl}
	mock.recorder = &MockChangeValidatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockChangeValidator) EXPECT() *MockChangeValidatorMockRecorder {
	return m.recorder
}

// ValidateChange mocks base method.
func (m *MockChangeValidator) ValidateChange(arg0 models.ConfigurationChange, arg1 *v1.CBContainersAgent) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateChange", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateChange indicates an expected call of ValidateChange.
func (mr *MockChangeValidatorMockRecorder) ValidateChange(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateChange", reflect.TypeOf((*MockChangeValidator)(nil).ValidateChange), arg0, arg1)
}
