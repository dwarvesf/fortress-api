package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewLogrusLogger_WithLogLevel(t *testing.T) {
	tests := []struct {
		name            string
		logLevel        string
		testMessage     string
		logFunc         func(Logger, string)
		shouldAppear    bool
		expectedLevel   logrus.Level
	}{
		{
			name:          "debug level shows debug messages",
			logLevel:      "debug",
			testMessage:   "debug message",
			logFunc:       func(l Logger, msg string) { l.Debug(msg) },
			shouldAppear:  true,
			expectedLevel: logrus.DebugLevel,
		},
		{
			name:          "info level hides debug messages",
			logLevel:      "info",
			testMessage:   "debug message",
			logFunc:       func(l Logger, msg string) { l.Debug(msg) },
			shouldAppear:  false,
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "info level shows info messages",
			logLevel:      "info",
			testMessage:   "info message",
			logFunc:       func(l Logger, msg string) { l.Info(msg) },
			shouldAppear:  true,
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "warn level hides info messages",
			logLevel:      "warn",
			testMessage:   "info message",
			logFunc:       func(l Logger, msg string) { l.Info(msg) },
			shouldAppear:  false,
			expectedLevel: logrus.WarnLevel,
		},
		{
			name:          "warn level shows warn messages",
			logLevel:      "warn",
			testMessage:   "warning message",
			logFunc:       func(l Logger, msg string) { l.Warn(msg) },
			shouldAppear:  true,
			expectedLevel: logrus.WarnLevel,
		},
		{
			name:          "error level shows error messages",
			logLevel:      "error",
			testMessage:   "error message",
			logFunc:       func(l Logger, msg string) { l.Error(nil, msg) },
			shouldAppear:  true,
			expectedLevel: logrus.ErrorLevel,
		},
		{
			name:          "error level hides warn messages",
			logLevel:      "error",
			testMessage:   "warning message",
			logFunc:       func(l Logger, msg string) { l.Warn(msg) },
			shouldAppear:  false,
			expectedLevel: logrus.ErrorLevel,
		},
		{
			name:          "empty level defaults to info",
			logLevel:      "",
			testMessage:   "info message",
			logFunc:       func(l Logger, msg string) { l.Info(msg) },
			shouldAppear:  true,
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "invalid level defaults to info",
			logLevel:      "invalid",
			testMessage:   "info message",
			logFunc:       func(l Logger, msg string) { l.Info(msg) },
			shouldAppear:  true,
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "case insensitive - uppercase DEBUG",
			logLevel:      "DEBUG",
			testMessage:   "debug message",
			logFunc:       func(l Logger, msg string) { l.Debug(msg) },
			shouldAppear:  true,
			expectedLevel: logrus.DebugLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			logger := NewLogrusLogger(tt.logLevel)

			// Get the underlying logrus logger to configure output
			logrusLogger, ok := logger.(*LogrusLogger)
			if !ok {
				t.Fatal("Logger is not a LogrusLogger")
			}
			logrusLogger.logger.SetOutput(&buf)

			// Verify the log level was set correctly
			if logrusLogger.logger.Level != tt.expectedLevel {
				t.Errorf("Logger level = %v, want %v", logrusLogger.logger.Level, tt.expectedLevel)
			}

			// Test the actual logging
			tt.logFunc(logger, tt.testMessage)

			output := buf.String()

			if tt.shouldAppear {
				if !strings.Contains(output, tt.testMessage) {
					t.Errorf("Expected message %q to appear in logs, but got: %q", tt.testMessage, output)
				}
			} else {
				if strings.Contains(output, tt.testMessage) {
					t.Errorf("Expected message %q NOT to appear in logs, but got: %q", tt.testMessage, output)
				}
			}
		})
	}
}

func TestNewLogrusLogger_DebugfFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogrusLogger("debug")

	logrusLogger := logger.(*LogrusLogger)
	logrusLogger.logger.SetOutput(&buf)

	logger.Debugf("formatted %s message with number %d", "debug", 42)

	output := buf.String()
	if !strings.Contains(output, "formatted debug message with number 42") {
		t.Errorf("Expected formatted message in output, got: %q", output)
	}
}

func TestNewLogrusLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogrusLogger("info")

	logrusLogger := logger.(*LogrusLogger)
	logrusLogger.logger.SetOutput(&buf)

	logger.Info("test message")

	output := buf.String()

	// Verify it's valid JSON
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Log output is not valid JSON: %v, output: %q", err, output)
	}

	// Verify expected fields
	if logEntry["msg"] != "test message" {
		t.Errorf("Expected msg field to be 'test message', got: %v", logEntry["msg"])
	}

	if logEntry["level"] != "info" {
		t.Errorf("Expected level field to be 'info', got: %v", logEntry["level"])
	}

	// Verify default fields
	if _, ok := logEntry["service"]; !ok {
		t.Error("Expected 'service' field in log output")
	}
	if _, ok := logEntry["hostname"]; !ok {
		t.Error("Expected 'hostname' field in log output")
	}
}

func TestNewLogrusLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogrusLogger("info")

	logrusLogger := logger.(*LogrusLogger)
	logrusLogger.logger.SetOutput(&buf)

	loggerWithFields := logger.Fields(Fields{
		"request_id": "123",
		"user":       "test-user",
	})

	loggerWithFields.Info("test message")

	output := buf.String()

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Log output is not valid JSON: %v", err)
	}

	if logEntry["request_id"] != "123" {
		t.Errorf("Expected request_id to be '123', got: %v", logEntry["request_id"])
	}

	if logEntry["user"] != "test-user" {
		t.Errorf("Expected user to be 'test-user', got: %v", logEntry["user"])
	}
}
