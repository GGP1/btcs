package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

var logger = New(false, true, os.Stderr)

const (
	// info designates informational messages that highlight the progress of the application at coarse-grained level
	info level = iota
	// debug designates fine-grained informational events useful to debug an application
	debug
	// err designates err events
	err
	// fatal shows an error and exits
	fatal
)

// level represents the logging level used.
type level uint8

// Logger contains the logging options.
type Logger struct {
	out           io.Writer
	development   bool
	disable       bool
	showTimestamp bool
}

// New creates a new logger.
func New(development, showTimestamp bool, out ...io.Writer) *Logger {
	return &Logger{
		development:   development,
		showTimestamp: showTimestamp,
		out:           io.MultiWriter(out...),
	}
}

func (l *Logger) log(level level, message string) {
	if l.disable || (level == debug && !l.development) {
		return
	}

	var lvl string
	switch level {
	case info:
		lvl = "INF"
	case debug:
		lvl = "DBG"
	case err:
		lvl = "ERR"
	case fatal:
		lvl = "FTL"
	}

	timestamp := ""
	if l.showTimestamp {
		timestamp = time.Now().Local().Format("2006-01-02T15:04:05") + " - "
	}

	log := fmt.Sprintf("[%s] %s%s", lvl, timestamp, message)

	fmt.Fprintln(l.out, log)
}

// AddOut adds n writers.
func AddOut(out ...io.Writer) {
	if len(out) == 0 {
		return
	}
	out = append(out, logger.out)
	logger.out = io.MultiWriter(out...)
}

// Disable turns off the logger.
func Disable() {
	logger.disable = true
}

// Info provides useful information about the server.
func Info(args ...interface{}) {
	logger.log(info, fmt.Sprint(args...))
}

// Infof is like Info but takes a formatted message.
func Infof(format string, args ...interface{}) {
	logger.log(info, fmt.Sprintf(format, args...))
}

// Debug provides useful information for debugging.
func Debug(args ...interface{}) {
	logger.log(debug, fmt.Sprint(args...))
}

// Debugf is like Debug but takes a formatted message.
func Debugf(format string, args ...interface{}) {
	logger.log(debug, fmt.Sprintf(format, args...))
}

// Error reports the application errors.
func Error(args ...interface{}) {
	logger.log(err, fmt.Sprint(args...))
}

// Errorf is like Error but takes a formatted message.
func Errorf(format string, args ...interface{}) {
	logger.log(err, fmt.Sprintf(format, args...))
}

// Fatal reports the application errors and exists.
func Fatal(args ...interface{}) {
	logger.log(fatal, fmt.Sprint(args...))
	os.Exit(1)
}

// Fatalf is like Fatal but takes a formatted message.
func Fatalf(format string, args ...interface{}) {
	logger.log(fatal, fmt.Sprintf(format, args...))
	os.Exit(1)
}

// SetDevelopment enables/disables the development mode.
func SetDevelopment(dev bool) {
	logger.development = dev
}

// SetOut sets the writers where the information will be logged.
func SetOut(w ...io.Writer) {
	logger.out = io.MultiWriter(w...)
}

// ShowTimestamp determines where the timestamp will be logged or not.
func ShowTimestamp(show bool) {
	logger.showTimestamp = show
}
