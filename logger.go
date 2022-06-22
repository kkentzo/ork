package main

import (
	"fmt"
	"os"

	"github.com/apsdehal/go-logger"
)

const (
	LOG_LEVEL_INFO  = "info"
	LOG_LEVEL_ERROR = "error"
	LOG_LEVEL_DEBUG = "debug"
)

var (
	logLevels = map[string]logger.LogLevel{
		LOG_LEVEL_INFO:  logger.InfoLevel,
		LOG_LEVEL_DEBUG: logger.DebugLevel,
		LOG_LEVEL_ERROR: logger.ErrorLevel,
	}
)

type Logger interface {
	Fatal(string)
	Fatalf(string, ...interface{})

	Error(string)
	Errorf(string, ...interface{})

	Info(string)
	Infof(string, ...interface{})

	Debug(string)
	Debugf(string, ...interface{})

	Output(string)

	// implements the io.Writer interface
	Write(p []byte) (n int, err error)

	SetLogLevel(string) error
	GetLogLevel() logger.LogLevel
}

type OrkLogger struct {
	level logger.LogLevel
	*logger.Logger
}

func NewLogger() (Logger, error) {
	l, err := logger.New("ork", 1, os.Stdout)
	if err != nil {
		return nil, err
	}
	l.SetFormat("[%{level}] %{message}")
	l.SetLogLevel(logger.InfoLevel)
	return &OrkLogger{Logger: l, level: logger.InfoLevel}, nil
}

func (l *OrkLogger) SetLogLevel(level string) error {
	lvl, ok := logLevels[level]
	if !ok {
		return fmt.Errorf("unknown log level: %s", level)
	}
	l.Logger.SetLogLevel(lvl)
	l.level = lvl
	return nil
}

func (l *OrkLogger) GetLogLevel() logger.LogLevel {
	return l.level
}

func (l *OrkLogger) Write(p []byte) (n int, err error) {
	l.Output(string(p))
	return len(p), nil
}

func (l *OrkLogger) Output(message string) {
	fmt.Print(message)
}
