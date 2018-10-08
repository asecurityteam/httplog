// Automatically generated by MockGen. DO NOT EDIT!
// Source: bitbucket.org/atlassian/logevent (interfaces: Logger)

// nolint
package httplog

import (
	logevent "bitbucket.org/atlassian/logevent"
	gomock "github.com/golang/mock/gomock"
)

// Mock of Logger interface
type MockLogger struct {
	ctrl     *gomock.Controller
	recorder *_MockLoggerRecorder
}

// Recorder for MockLogger (not exported)
type _MockLoggerRecorder struct {
	mock *MockLogger
}

func NewMockLogger(ctrl *gomock.Controller) *MockLogger {
	mock := &MockLogger{ctrl: ctrl}
	mock.recorder = &_MockLoggerRecorder{mock}
	return mock
}

func (_m *MockLogger) EXPECT() *_MockLoggerRecorder {
	return _m.recorder
}

func (_m *MockLogger) Copy() logevent.Logger {
	ret := _m.ctrl.Call(_m, "Copy")
	ret0, _ := ret[0].(logevent.Logger)
	return ret0
}

func (_mr *_MockLoggerRecorder) Copy() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Copy")
}

func (_m *MockLogger) Debug(_param0 interface{}) {
	_m.ctrl.Call(_m, "Debug", _param0)
}

func (_mr *_MockLoggerRecorder) Debug(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Debug", arg0)
}

func (_m *MockLogger) Error(_param0 interface{}) {
	_m.ctrl.Call(_m, "Error", _param0)
}

func (_mr *_MockLoggerRecorder) Error(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Error", arg0)
}

func (_m *MockLogger) Info(_param0 interface{}) {
	_m.ctrl.Call(_m, "Info", _param0)
}

func (_mr *_MockLoggerRecorder) Info(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Info", arg0)
}

func (_m *MockLogger) SetField(_param0 string, _param1 interface{}) {
	_m.ctrl.Call(_m, "SetField", _param0, _param1)
}

func (_mr *_MockLoggerRecorder) SetField(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetField", arg0, arg1)
}

func (_m *MockLogger) Warn(_param0 interface{}) {
	_m.ctrl.Call(_m, "Warn", _param0)
}

func (_mr *_MockLoggerRecorder) Warn(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Warn", arg0)
}
