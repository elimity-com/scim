package logging

import (
	"io"
	"log"
	"os"
)

type Logger interface {
	Debugf(format string, args ...interface{})    // Debug message
	Infof(format string, args ...interface{})     // Information message
	Warningf(format string, args ...interface{})  // Warning message
	Errorf(format string, args ...interface{})    // Error message
	Criticalf(format string, args ...interface{}) // Critical message
}

// NullLogger implements a no-op Logger
type NullLogger struct{}

func (n NullLogger) Debugf(format string, args ...interface{})    {}
func (n NullLogger) Infof(format string, args ...interface{})     {}
func (n NullLogger) Warningf(format string, args ...interface{})  {}
func (n NullLogger) Errorf(format string, args ...interface{})    {}
func (n NullLogger) Criticalf(format string, args ...interface{}) {}

//SimpleLogger implements a Logger that directs output to an io.Writer
type SimpleLogger struct {
	DebugLogger    *log.Logger
	InfoLogger     *log.Logger
	WarningLogger  *log.Logger
	ErrorLogger    *log.Logger
	CriticalLogger *log.Logger
}

func NewSimpleLogger(out io.Writer) Logger {
	if out == nil {
		out = os.Stderr
	}
	return SimpleLogger{
		DebugLogger:    log.New(out, "DEBUG: ", log.LstdFlags|log.Lshortfile|log.LUTC),
		InfoLogger:     log.New(out, "INFO: ", log.LstdFlags|log.Lshortfile|log.LUTC),
		WarningLogger:  log.New(out, "WARNING: ", log.LstdFlags|log.Lshortfile|log.LUTC),
		ErrorLogger:    log.New(out, "ERROR: ", log.LstdFlags|log.Lshortfile|log.LUTC),
		CriticalLogger: log.New(out, "CRITICAL: ", log.LstdFlags|log.Lshortfile|log.LUTC),
	}
}

func (l SimpleLogger) Debugf(format string, args ...interface{}) {
	l.DebugLogger.Printf(format, args...)
}

func (l SimpleLogger) Infof(format string, args ...interface{}) {
	l.InfoLogger.Printf(format, args...)
}

func (l SimpleLogger) Warningf(format string, args ...interface{}) {
	l.WarningLogger.Printf(format, args...)
}

func (l SimpleLogger) Errorf(format string, args ...interface{}) {
	l.ErrorLogger.Printf(format, args...)
}

func (l SimpleLogger) Criticalf(format string, args ...interface{}) {
	l.CriticalLogger.Printf(format, args...)
}
