package config

import (
	"testing"
)

type mockENV struct {
	values map[string]string
	bools  map[string]bool
}

func (m *mockENV) GetString(key string) string {
	return m.values[key]
}

func (m *mockENV) GetBool(key string) bool {
	return m.bools[key]
}

func TestGenerate_LogLevel(t *testing.T) {
	tests := []struct {
		name          string
		envLogLevel   string
		expectedLevel string
	}{
		{
			name:          "debug level from env",
			envLogLevel:   "debug",
			expectedLevel: "debug",
		},
		{
			name:          "info level from env",
			envLogLevel:   "info",
			expectedLevel: "info",
		},
		{
			name:          "warn level from env",
			envLogLevel:   "warn",
			expectedLevel: "warn",
		},
		{
			name:          "error level from env",
			envLogLevel:   "error",
			expectedLevel: "error",
		},
		{
			name:          "fatal level from env",
			envLogLevel:   "fatal",
			expectedLevel: "fatal",
		},
		{
			name:          "default to info when empty",
			envLogLevel:   "",
			expectedLevel: "info",
		},
		{
			name:          "invalid level defaults to info",
			envLogLevel:   "invalid",
			expectedLevel: "info",
		},
		{
			name:          "case insensitive - uppercase",
			envLogLevel:   "DEBUG",
			expectedLevel: "debug",
		},
		{
			name:          "case insensitive - mixed case",
			envLogLevel:   "WaRn",
			expectedLevel: "warn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &mockENV{
				values: map[string]string{
					"LOG_LEVEL": tt.envLogLevel,
				},
				bools: map[string]bool{},
			}

			cfg := Generate(env)

			if cfg.LogLevel != tt.expectedLevel {
				t.Errorf("Generate() LogLevel = %v, want %v", cfg.LogLevel, tt.expectedLevel)
			}
		})
	}
}

func TestGenerate_LogLevel_WithOtherConfig(t *testing.T) {
	env := &mockENV{
		values: map[string]string{
			"LOG_LEVEL":     "debug",
			"ENV":           "production",
			"PORT":          "9090",
			"API_KEY":       "test-key",
			"JWT_SECRET_KEY": "secret",
		},
		bools: map[string]bool{
			"DEBUG": true,
		},
	}

	cfg := Generate(env)

	if cfg.LogLevel != "debug" {
		t.Errorf("Generate() LogLevel = %v, want debug", cfg.LogLevel)
	}

	// Ensure other fields are still populated correctly
	if cfg.Debug != true {
		t.Errorf("Generate() Debug = %v, want true", cfg.Debug)
	}
	if cfg.Env != "production" {
		t.Errorf("Generate() Env = %v, want production", cfg.Env)
	}
	if cfg.ApiServer.Port != "9090" {
		t.Errorf("Generate() ApiServer.Port = %v, want 9090", cfg.ApiServer.Port)
	}
}
