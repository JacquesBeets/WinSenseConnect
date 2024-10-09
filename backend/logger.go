package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/svc/eventlog"
)

type LogLevel int

const (
	LogOff LogLevel = iota
	LogErrors
	LogDebug
)

type Logger struct {
	filePath string
	config   *Config
	elog     *eventlog.Log
}

func NewLogger(filename string, config *Config, serviceName string) (*Logger, error) {
	logPath, err := getLogFilePath(filename)
	if err != nil {
		return nil, err
	}

	elog, err := eventlog.Open(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open event log: %v", err)
	}

	return &Logger{
		filePath: logPath,
		config:   config,
		elog:     elog,
	}, nil
}

func (l *Logger) Log(message string, level LogLevel) {
	var configLevel LogLevel
	if l.config == nil {
		configLevel = LogDebug // Default to debug level if config is nil
	} else {
		configLevel = getLogLevel(l.config.LogLevel)
	}

	if level > configLevel {
		return
	}

	// File logging
	f, err := os.OpenFile(l.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	} else {
		defer f.Close()
		logger := log.New(f, "", log.LstdFlags)
		logger.Println(message)
		f.Sync() // Force write to disk
	}

	// Windows Event logging
	switch level {
	case LogDebug:
		l.elog.Info(1, message)
	case LogErrors:
		l.elog.Error(1, message)
	}
}

func (l *Logger) Debug(message string) {
	l.Log(message, LogDebug)
}

func (l *Logger) Error(message string) {
	l.Log(message, LogErrors)
}

func (l *Logger) Close() {
	if l.elog != nil {
		l.elog.Close()
	}
}

func getLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return LogDebug
	case "errors":
		return LogErrors
	default:
		return LogOff
	}
}

func getLogFilePath(filename string) (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %v", err)
	}
	return filepath.Join(filepath.Dir(exePath), filename), nil
}
