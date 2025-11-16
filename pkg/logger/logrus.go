package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type LogrusLogger struct {
	logger *logrus.Logger
	fields Fields
}

var L Logger

// parseLogLevel converts a log level string to logrus.Level
// Returns InfoLevel as default for invalid or empty levels
func parseLogLevel(levelStr string) logrus.Level {
	// Normalize to lowercase and trim spaces
	normalized := strings.ToLower(strings.TrimSpace(levelStr))

	switch normalized {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn", "warning":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

func NewLogrusLogger(logLevel string) Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})

	// Set the log level from the parameter
	level := parseLogLevel(logLevel)
	l.SetLevel(level)

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	defaultFields := Fields{
		"service":  "fortress-api",
		"hostname": hostname,
	}

	L = &LogrusLogger{
		logger: l,
		fields: defaultFields,
	}

	return L
}

func (l *LogrusLogger) Fields(data Fields) Logger {
	return &LogrusLogger{
		logger: l.logger,
		fields: data,
	}
}

func (l *LogrusLogger) Field(key, value string) Logger {
	return &LogrusLogger{
		logger: l.logger,
		fields: Fields{key: value},
	}
}

func (l LogrusLogger) AddField(key string, value any) Logger {
	l.fields[key] = value
	return &LogrusLogger{
		logger: l.logger,
		fields: l.fields,
	}
}

func (l *LogrusLogger) Debug(msg string) {
	l.logger.WithFields(logrus.Fields(l.fields)).Debug(msg)
}

func (l *LogrusLogger) Debugf(msg string, args ...interface{}) {
	l.logger.WithFields(logrus.Fields(l.fields)).Debugf(msg, args...)
}

func (l *LogrusLogger) Info(msg string) {
	l.logger.WithFields(logrus.Fields(l.fields)).Info(msg)
}

func (l *LogrusLogger) Infof(msg string, args ...interface{}) {
	l.logger.WithFields(logrus.Fields(l.fields)).Infof(msg, args...)
}

func (l *LogrusLogger) Warn(msg string) {
	l.logger.WithFields(logrus.Fields(l.fields)).Warn(msg)
}

func (l *LogrusLogger) Warnf(msg string, args ...interface{}) {
	l.logger.WithFields(logrus.Fields(l.fields)).Warnf(msg, args...)
}

func (l *LogrusLogger) Error(err error, msg string) {
	l.logger.WithFields(logrus.Fields(l.fields)).WithError(err).Error(msg)
}

func (l *LogrusLogger) Errorf(err error, msg string, args ...interface{}) {
	l.logger.WithFields(logrus.Fields(l.fields)).WithError(err).Errorf(msg, args...)
}

func (l *LogrusLogger) Fatal(err error, msg string) {
	l.logger.WithFields(logrus.Fields(l.fields)).WithError(err).Fatal(msg)
}

func (l *LogrusLogger) Fatalf(err error, msg string, args ...interface{}) {
	l.logger.WithFields(logrus.Fields(l.fields)).WithError(err).Fatalf(msg, args...)
}
