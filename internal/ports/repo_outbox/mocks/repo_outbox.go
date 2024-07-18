// Code generated by MockGen. DO NOT EDIT.
// Source: repo_outbox.go

// Package mock_repo_outbox is a generated GoMock package.
package mock_repo_outbox

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	dto "github.com/lazylex/messaggio/internal/dto"
)

// MockInterface is a mock of Interface interface.
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface.
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance.
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockInterface) Add(id dto.MessageID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockInterfaceMockRecorder) Add(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockInterface)(nil).Add), id)
}

// Pop mocks base method.
func (m *MockInterface) Pop() dto.MessageID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Pop")
	ret0, _ := ret[0].(dto.MessageID)
	return ret0
}

// Pop indicates an expected call of Pop.
func (mr *MockInterfaceMockRecorder) Pop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pop", reflect.TypeOf((*MockInterface)(nil).Pop))
}
