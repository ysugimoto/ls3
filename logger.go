package main

import (
	"os"
)

type logWritable interface {
	Close()
	log(message string)
}

type Logger struct {
	fp *os.File

	logWritable
}

func (l *Logger) Close() {
	l.fp.Close()
}

func (l *Logger) log(message string) {
	l.fp.WriteString(message + "\n")
}

type EmptyLogger struct {
	logWritable
}

func (e *EmptyLogger) Close() {
	// noop
}

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
