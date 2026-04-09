package logger

import (
	"log"
	"os"
)

type Logger struct {
	info  *log.Logger
	error *log.Logger
}

func New() *Logger {
	return &Logger{
		info:  log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile),
		error: log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Logger) Info(msg string) {
	l.info.Println(msg)
}

func (l *Logger) Infof(format string, v ...any) {
	l.info.Printf(format, v...)
}

func (l *Logger) Error(msg string) {
	l.error.Println(msg)
}

func (l *Logger) Errorf(format string, v ...any) {
	l.error.Printf(format, v...)
}

func (l *Logger) Fatal(msg string) {
	l.error.Fatal(msg)
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.error.Fatalf(format, v...)
}