// Package gormtx is a generated GoMock package.
// mockgen -source=db_tx_manager.go -destination=db_tx_manager_mock.go --package=gormtx --build_flags=--mod=mod
package gormtx

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	gorm "gorm.io/gorm"
)

// MockDBTxManager is a mock of DBTxManager interface.
type MockDBTxManager struct {
	ctrl     *gomock.Controller
	recorder *MockDBTxManagerMockRecorder
}

// MockDBTxManagerMockRecorder is the mock recorder for MockDBTxManager.
type MockDBTxManagerMockRecorder struct {
	mock *MockDBTxManager
}

// NewMockDBTxManager creates a new mock instance.
func NewMockDBTxManager(ctrl *gomock.Controller) *MockDBTxManager {
	mock := &MockDBTxManager{ctrl: ctrl}
	mock.recorder = &MockDBTxManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDBTxManager) EXPECT() *MockDBTxManagerMockRecorder {
	return m.recorder
}

// AutoDB mocks base method.
func (m *MockDBTxManager) AutoDB(ctx context.Context) *gorm.DB {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AutoDB", ctx)
	ret0, _ := ret[0].(*gorm.DB)
	return ret0
}

// AutoDB indicates an expected call of AutoDB.
func (mr *MockDBTxManagerMockRecorder) AutoDB(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AutoDB", reflect.TypeOf((*MockDBTxManager)(nil).AutoDB), ctx)
}

// BackupDB mocks base method.
func (m *MockDBTxManager) BackupDB(ctx context.Context) *gorm.DB {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BackupDB", ctx)
	ret0, _ := ret[0].(*gorm.DB)
	return ret0
}

// BackupDB indicates an expected call of BackupDB.
func (mr *MockDBTxManagerMockRecorder) BackupDB(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BackupDB", reflect.TypeOf((*MockDBTxManager)(nil).BackupDB), ctx)
}

// CloseTx mocks base method.
func (m *MockDBTxManager) CloseTx(ctx context.Context, txid uint64, err *error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CloseTx", ctx, txid, err)
}

// CloseTx indicates an expected call of CloseTx.
func (mr *MockDBTxManagerMockRecorder) CloseTx(ctx, txid, err interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloseTx", reflect.TypeOf((*MockDBTxManager)(nil).CloseTx), ctx, txid, err)
}

// MainDB mocks base method.
func (m *MockDBTxManager) MainDB(ctx context.Context) *gorm.DB {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MainDB", ctx)
	ret0, _ := ret[0].(*gorm.DB)
	return ret0
}

// MainDB indicates an expected call of MainDB.
func (mr *MockDBTxManagerMockRecorder) MainDB(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MainDB", reflect.TypeOf((*MockDBTxManager)(nil).MainDB), ctx)
}

// MustMainTx mocks base method.
func (m *MockDBTxManager) MustMainTx(ctx context.Context) *gorm.DB {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MustMainTx", ctx)
	ret0, _ := ret[0].(*gorm.DB)
	return ret0
}

// MustMainTx indicates an expected call of MustMainTx.
func (mr *MockDBTxManagerMockRecorder) MustMainTx(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MustMainTx", reflect.TypeOf((*MockDBTxManager)(nil).MustMainTx), ctx)
}

// NonTx mocks base method.
func (m *MockDBTxManager) NonTx(ctx context.Context) context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NonTx", ctx)
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// NonTx indicates an expected call of NonTx.
func (mr *MockDBTxManagerMockRecorder) NonTx(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NonTx", reflect.TypeOf((*MockDBTxManager)(nil).NonTx), ctx)
}

// OpenTx mocks base method.
func (m *MockDBTxManager) OpenTx(ctx context.Context, opts ...Option) (context.Context, uint64) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "OpenTx", varargs...)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(uint64)
	return ret0, ret1
}

// OpenTx indicates an expected call of OpenTx.
func (mr *MockDBTxManagerMockRecorder) OpenTx(ctx interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OpenTx", reflect.TypeOf((*MockDBTxManager)(nil).OpenTx), varargs...)
}
