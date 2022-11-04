package main

import (
	"github.com/stretchr/testify/mock"
)

type MockLogger struct {
	mock.Mock
}

func (ml *MockLogger) Infof(format string, v ...interface{}) {
	ml.Called(format, v)
}

func (ml *MockLogger) Warnf(format string, v ...interface{}) {
	ml.Called(format, v)
}

func (ml *MockLogger) Printf(format string, v ...interface{}) {
	ml.Called(format, v)
}

func (ml *MockLogger) Donef(format string, v ...interface{}) {
	ml.Called(format, v)
}

func (ml *MockLogger) Debugf(format string, v ...interface{}) {
	ml.Called(format, v)
}

func (ml *MockLogger) Errorf(format string, v ...interface{}) {
	ml.Called(format, v)
}

func (ml *MockLogger) TInfof(format string, v ...interface{}) {
	ml.Called(format, v)
}

func (ml *MockLogger) TWarnf(format string, v ...interface{}) {
	ml.Called(format, v)
}

func (ml *MockLogger) TPrintf(format string, v ...interface{}) {
	ml.Called(format, v)
}

func (ml *MockLogger) TDonef(format string, v ...interface{}) {
	ml.Called(format, v)
}

func (ml *MockLogger) TDebugf(format string, v ...interface{}) {
	ml.Called(format, v)
}

func (ml *MockLogger) TErrorf(format string, v ...interface{}) {
	ml.Called(format, v)
}

func (ml *MockLogger) Println() {
	ml.Called()
}

func (ml *MockLogger) EnableDebugLog(enable bool) {
	ml.Called(enable)
}
