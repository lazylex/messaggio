// Code generated by MockGen. DO NOT EDIT.
// Source: repository.go

// Package mock_repository is a generated GoMock package.
package mock_repository

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	uuid "github.com/google/uuid"
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

// ProcessedCount mocks base method.
func (m *MockInterface) ProcessedCount(ctx context.Context) (dto.Processed, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessedCount", ctx)
	ret0, _ := ret[0].(dto.Processed)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProcessedCount indicates an expected call of ProcessedCount.
func (mr *MockInterfaceMockRecorder) ProcessedCount(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessedCount", reflect.TypeOf((*MockInterface)(nil).ProcessedCount), ctx)
}

// SaveMessage mocks base method.
func (m *MockInterface) SaveMessage(ctx context.Context, data dto.MessageID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveMessage", ctx, data)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveMessage indicates an expected call of SaveMessage.
func (mr *MockInterfaceMockRecorder) SaveMessage(ctx, data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveMessage", reflect.TypeOf((*MockInterface)(nil).SaveMessage), ctx, data)
}

// UpdateStatus mocks base method.
func (m *MockInterface) UpdateStatus(ctx context.Context, id uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateStatus", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateStatus indicates an expected call of UpdateStatus.
func (mr *MockInterfaceMockRecorder) UpdateStatus(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateStatus", reflect.TypeOf((*MockInterface)(nil).UpdateStatus), ctx, id)
}
