package logger

import (
	"log"
	"os"
)

type Logger struct {
	level int
	*log.Logger
}

const (
	DEBUG = iota
	INFO
	WARN
	ERROR
)

var levelNames = map[string]int{
	"debug": DEBUG,
	"info":  INFO,
	"warn":  WARN,
	"error": ERROR,
}

func New(levelStr string) *Logger {
	level, ok := levelNames[levelStr]
	if !ok {
		level = INFO
	}

	return &Logger{
		level:  level,
		Logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.level <= DEBUG {
		l.Printf("[DEBUG] "+msg, args...)
	}
}

func (l *Logger) Info(msg string, args ...interface{}) {
	if l.level <= INFO {
		l.Printf("[INFO] "+msg, args...)
	}
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	if l.level <= WARN {
		l.Printf("[WARN] "+msg, args...)
	}
}

func (l *Logger) Error(msg string, args ...interface{}) {
	if l.level <= ERROR {
		l.Printf("[ERROR] "+msg, args...)
	}
}
