package main

import (
	"fmt"
	"os"

	"github.com/apsdehal/go-logger"
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
}

type OrkLogger struct {
	*logger.Logger
}

func NewLogger() (Logger, error) {
	l, err := logger.New("ork", 1, os.Stdout)
	if err != nil {
		return nil, err
	}
	l.SetFormat("[%{level}] %{message}")
	return &OrkLogger{Logger: l}, nil
}

func (l *OrkLogger) Output(message string) {
	fmt.Print(message)
}
