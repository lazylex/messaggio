// Code generated by MockGen. DO NOT EDIT.
// Source: service.go

// Package mock_service is a generated GoMock package.
package mock_service

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockMetricsInterface is a mock of MetricsInterface interface.
type MockMetricsInterface struct {
	ctrl     *gomock.Controller
	recorder *MockMetricsInterfaceMockRecorder
}

// MockMetricsInterfaceMockRecorder is the mock recorder for MockMetricsInterface.
type MockMetricsInterfaceMockRecorder struct {
	mock *MockMetricsInterface
}

// NewMockMetricsInterface creates a new mock instance.
func NewMockMetricsInterface(ctrl *gomock.Controller) *MockMetricsInterface {
	mock := &MockMetricsInterface{ctrl: ctrl}
	mock.recorder = &MockMetricsInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetricsInterface) EXPECT() *MockMetricsInterfaceMockRecorder {
	return m.recorder
}

// IncomingMsgInc mocks base method.
func (m *MockMetricsInterface) IncomingMsgInc() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "IncomingMsgInc")
}

// IncomingMsgInc indicates an expected call of IncomingMsgInc.
func (mr *MockMetricsInterfaceMockRecorder) IncomingMsgInc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IncomingMsgInc", reflect.TypeOf((*MockMetricsInterface)(nil).IncomingMsgInc))
}

// ProblemsSavingInDB mocks base method.
func (m *MockMetricsInterface) ProblemsSavingInDB() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ProblemsSavingInDB")
}

// ProblemsSavingInDB indicates an expected call of ProblemsSavingInDB.
func (mr *MockMetricsInterfaceMockRecorder) ProblemsSavingInDB() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProblemsSavingInDB", reflect.TypeOf((*MockMetricsInterface)(nil).ProblemsSavingInDB))
}

// ProcessedMsgInc mocks base method.
func (m *MockMetricsInterface) ProcessedMsgInc() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ProcessedMsgInc")
}

// ProcessedMsgInc indicates an expected call of ProcessedMsgInc.
func (mr *MockMetricsInterfaceMockRecorder) ProcessedMsgInc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessedMsgInc", reflect.TypeOf((*MockMetricsInterface)(nil).ProcessedMsgInc))
}
