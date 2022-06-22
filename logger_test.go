package main

import (
	"fmt"
	"os"

	"github.com/apsdehal/go-logger"
)

type MockLogger struct {
	logs    map[logger.LogLevel][]string
	outputs []string
	*OrkLogger
}

func NewMockLogger() *MockLogger {
	logs := map[logger.LogLevel][]string{}
	levels := []logger.LogLevel{
		logger.CriticalLevel,
		logger.ErrorLevel,
		logger.WarningLevel,
		logger.NoticeLevel,
		logger.InfoLevel,
		logger.DebugLevel,
	}
	for _, lvl := range levels {
		logs[lvl] = []string{}
	}
	actual, err := NewLogger()
	if err != nil {
		fmt.Println("failed to initialize MockLogger!")
		os.Exit(1)
	}
	return &MockLogger{logs: logs, outputs: []string{}, OrkLogger: actual.(*OrkLogger)}
}

func (l *MockLogger) SetLogLevel(lvl string) error {
	return l.OrkLogger.SetLogLevel(lvl)
}

func (l *MockLogger) GetLogLevel() logger.LogLevel {
	return l.OrkLogger.level
}

func (l *MockLogger) Logs(lvl logger.LogLevel) []string {
	return l.logs[lvl]
}

func (l *MockLogger) Outputs() []string {
	return l.outputs
}

func (l *MockLogger) Fatal(msg string) {
	l.logs[logger.CriticalLevel] = append(l.logs[logger.CriticalLevel], msg)
}

func (l *MockLogger) Fatalf(msg string, a ...interface{}) {
	l.Fatal(fmt.Sprintf(msg, a...))
}

func (l *MockLogger) Error(msg string) {
	if logger.ErrorLevel <= l.GetLogLevel() {
		l.logs[logger.ErrorLevel] = append(l.logs[logger.ErrorLevel], msg)
	}
}

func (l *MockLogger) Errorf(msg string, a ...interface{}) {
	l.Error(fmt.Sprintf(msg, a...))
}

func (l *MockLogger) Info(msg string) {
	if logger.InfoLevel <= l.GetLogLevel() {
		l.logs[logger.InfoLevel] = append(l.logs[logger.InfoLevel], msg)
	}
}

func (l *MockLogger) Infof(msg string, a ...interface{}) {
	l.Info(fmt.Sprintf(msg, a...))
}

func (l *MockLogger) Debug(msg string) {
	if logger.DebugLevel <= l.GetLogLevel() {
		l.logs[logger.DebugLevel] = append(l.logs[logger.DebugLevel], msg)
	}
}

func (l *MockLogger) Debugf(msg string, a ...interface{}) {
	l.Debug(fmt.Sprintf(msg, a...))
}

func (l *MockLogger) Write(p []byte) (n int, err error) {
	l.Output(string(p))
	return len(p), nil
}

func (l *MockLogger) Output(msg string) {
	l.outputs = append(l.outputs, msg)
}
