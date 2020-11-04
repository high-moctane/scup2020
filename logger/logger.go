package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
)

const (
	Info = iota
	Warn
	Error
	Fatal
	Panic
)

var once sync.Once
var logger *Logger

type Logger struct {
	level int
	l     *log.Logger
}

func Get() *Logger {
	once.Do(func() {
		level := Panic
		levelStr, ok := os.LookupEnv("SCUP_LOG_LEVEL")
		if ok {
			switch levelStr {
			case "INFO":
				level = Info
			case "WARN":
				level = Warn
			case "ERROR":
				level = Error
			case "FATAL":
				level = Fatal
			case "PANIC":
				level = Fatal
			default:
				panic(fmt.Sprintf("logger invalid log level: %s", levelStr))
			}
		}

		logger = &Logger{level: level, l: log.New(os.Stdout, "", log.LstdFlags)}
	})
	return logger
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= Info {
		l.l.Fatalf("%s: %s\n", "INFO:", fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= Warn {
		l.l.Fatalf("%s: %s\n", "WARN:", fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= Error {
		l.l.Fatalf("%s: %s\n", "ERROR:", fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	if l.level <= Fatal {
		l.l.Fatalf("%s: %s\n", "FATAL:", fmt.Sprintf(format, v...))
	}
}
