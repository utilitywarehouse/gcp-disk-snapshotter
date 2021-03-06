// Code generated by MockGen. DO NOT EDIT.
// Source: metrics/prometheus.go

// Package metrics is a generated GoMock package.
package metrics

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockPrometheusInterface is a mock of PrometheusInterface interface
type MockPrometheusInterface struct {
	ctrl     *gomock.Controller
	recorder *MockPrometheusInterfaceMockRecorder
}

// MockPrometheusInterfaceMockRecorder is the mock recorder for MockPrometheusInterface
type MockPrometheusInterfaceMockRecorder struct {
	mock *MockPrometheusInterface
}

// NewMockPrometheusInterface creates a new mock instance
func NewMockPrometheusInterface(ctrl *gomock.Controller) *MockPrometheusInterface {
	mock := &MockPrometheusInterface{ctrl: ctrl}
	mock.recorder = &MockPrometheusInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPrometheusInterface) EXPECT() *MockPrometheusInterfaceMockRecorder {
	return m.recorder
}

// Init mocks base method
func (m *MockPrometheusInterface) Init() {
	m.ctrl.Call(m, "Init")
}

// Init indicates an expected call of Init
func (mr *MockPrometheusInterfaceMockRecorder) Init() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Init", reflect.TypeOf((*MockPrometheusInterface)(nil).Init))
}

// UpdateCreateSnapshotStatus mocks base method
func (m *MockPrometheusInterface) UpdateCreateSnapshotStatus(disk string, success bool) {
	m.ctrl.Call(m, "UpdateCreateSnapshotStatus", disk, success)
}

// UpdateCreateSnapshotStatus indicates an expected call of UpdateCreateSnapshotStatus
func (mr *MockPrometheusInterfaceMockRecorder) UpdateCreateSnapshotStatus(disk, success interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCreateSnapshotStatus", reflect.TypeOf((*MockPrometheusInterface)(nil).UpdateCreateSnapshotStatus), disk, success)
}

// UpdateDeleteSnapshotStatus mocks base method
func (m *MockPrometheusInterface) UpdateDeleteSnapshotStatus(disk string, success bool) {
	m.ctrl.Call(m, "UpdateDeleteSnapshotStatus", disk, success)
}

// UpdateDeleteSnapshotStatus indicates an expected call of UpdateDeleteSnapshotStatus
func (mr *MockPrometheusInterfaceMockRecorder) UpdateDeleteSnapshotStatus(disk, success interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateDeleteSnapshotStatus", reflect.TypeOf((*MockPrometheusInterface)(nil).UpdateDeleteSnapshotStatus), disk, success)
}

// UpdateOperationStatus mocks base method
func (m *MockPrometheusInterface) UpdateOperationStatus(operation_type string, success bool) {
	m.ctrl.Call(m, "UpdateOperationStatus", operation_type, success)
}

// UpdateOperationStatus indicates an expected call of UpdateOperationStatus
func (mr *MockPrometheusInterfaceMockRecorder) UpdateOperationStatus(operation_type, success interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOperationStatus", reflect.TypeOf((*MockPrometheusInterface)(nil).UpdateOperationStatus), operation_type, success)
}
