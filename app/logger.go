package app

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var _ logrus.Formatter = (*utcFormatter)(nil)

// utcFormatter is a wrapper to logrus formatter by formatting all timestamps in utc.
type utcFormatter struct {
	formatter logrus.Formatter
}

func (tf *utcFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Normalize the time to utc.
	entry.Time = entry.Time.UTC()
	return tf.formatter.Format(entry)
}

func NewLogger(level string, format string) *logrus.Logger {
	// For now, limit the log levels to keep it simple.
	var lvl logrus.Level
	switch strings.ToLower(level) {
	case "", "info", "debug", "trace":
		lvl = logrus.InfoLevel
	case "warning", "warn":
		lvl = logrus.WarnLevel
	case "error":
		lvl = logrus.ErrorLevel
	default:
		lvl = logrus.InfoLevel
	}

	formatter := logrus.TextFormatter{
		DisableColors:   false,
		FullTimestamp:   true,
		TimestampFormat: "",
	}

	var tf utcFormatter
	switch strings.ToLower(format) {
	case "", "text":
		tf.formatter = &formatter
	case "json":
		tf.formatter = &logrus.JSONFormatter{}
	default:
		tf.formatter = &formatter
	}

	logger := logrus.Logger{
		Out:       os.Stderr,
		Formatter: &tf,
		Level:     lvl,
		ExitFunc:  exitHandler,
	}

	return &logger
}

func exitHandler(code int) {
	// For now, just wraps os.Exit.
	os.Exit(code)
}
