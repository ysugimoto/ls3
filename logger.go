package main

import (
	"os"
	"time"
)

// Write log interface
type logWritable interface {
	Close()
	log(message string)
}

// Append file logger
type Logger struct {
	fp *os.File

	logWritable
}

// Close file pointer
func (l *Logger) Close() {
	l.fp.Close()
}

// logWritable::log implementation
func (l *Logger) log(message string) {
	now := time.Now().Format("2006-01-02 15:03:04")
	l.fp.WriteString(now + " " + message + "\n")
}

// Empty logger
type EmptyLogger struct {
	logWritable
}

func (e *EmptyLogger) Close() {
	// noop
}

// logWritable::log implementation
func (e *EmptyLogger) log(message string) {
	// noop
}

var logger logWritable

func init() {
	if os.Getenv("DEBUG") == "" {
		logger = &EmptyLogger{}
		return
	}

	cwd, _ := os.Getwd()
	fp, err := os.OpenFile(cwd+"/ls3.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	_, err = fp.WriteString("Logger started\n")
	if err != nil {
		panic(err)
	}
	logger = &Logger{
		fp: fp,
	}
}
