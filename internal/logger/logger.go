package logging

import "log"

type LoggerMode int
const(
	LoggerDebug LoggerMode = iota
	LoggerInfo
)

type Logger struct {
	mode LoggerMode
}

func NewLogger(loggerMode LoggerMode) Logger {
	return Logger{
		mode: loggerMode,
	}
}

func (logger Logger) Debug(msg string) {
	if logger.mode == LoggerDebug {
		log.Println("DEBUG " + msg)
	}
}

func (logger Logger) Info(msg string) {
	log.Println("INFO " + msg)
}

func (logger Logger) Error(msg string, err error) {
	log.Println("ERROR " + msg)
}