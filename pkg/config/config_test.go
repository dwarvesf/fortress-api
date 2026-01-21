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
			"LOG_LEVEL":      "debug",
			"ENV":            "production",
			"PORT":           "9090",
			"API_KEY":        "test-key",
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

func Test_parseKeyValuePairs(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want map[string]string
	}{
		{
			name: "empty string",
			raw:  "",
			want: nil,
		},
		{
			name: "single pair",
			raw:  "han@example.com:123",
			want: map[string]string{"han@example.com": "123"},
		},
		{
			name: "multiple pairs with spaces",
			raw:  " han@example.com : 123 , ops@example.com:456 ",
			want: map[string]string{
				"han@example.com": "123",
				"ops@example.com": "456",
			},
		},
		{
			name: "invalid entries skipped",
			raw:  "invalid,foo:bar,baz:",
			want: map[string]string{"foo": "bar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseKeyValuePairs(tt.raw)
			if len(tt.want) == 0 && got != nil {
				t.Fatalf("expected nil map, got %v", got)
			}
			if len(tt.want) != len(got) {
				t.Fatalf("size mismatch: got %d, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Fatalf("expected %s=%s, got %s", k, v, got[k])
				}
			}
		})
	}
}

func TestGenerate_ExpenseIntegration(t *testing.T) {
		env := &mockENV{
			values: map[string]string{
				"NOCO_EXPENSE_WORKSPACE_ID":      "ws_1",
				"NOCO_EXPENSE_TABLE_ID":          "tbl_1",
				"NOCO_EXPENSE_WEBHOOK_SECRET":    "secret",
				"NOCO_EXPENSE_APPROVER_MAPPING":  "han@example.com:123, ops@example.com:456",
			"NOCO_BASE_URL":                  "https://nocodb",
			"NOCO_TOKEN":                     "token",
			"NOCO_WORKSPACE_ID":              "nw",
			"NOCO_BASE_ID":                   "nb",
			"NOCO_INVOICE_TABLE_ID":          "invoice",
			"NOCO_INVOICE_COMMENTS_TABLE_ID": "comments",
		},
		bools: map[string]bool{},
	}

	cfg := Generate(env)

	exp := cfg.ExpenseIntegration
	if exp.Noco.WorkspaceID != "ws_1" {
		t.Fatalf("expected workspace ws_1, got %s", exp.Noco.WorkspaceID)
	}
	if exp.Noco.TableID != "tbl_1" {
		t.Fatalf("expected table tbl_1, got %s", exp.Noco.TableID)
	}
	if exp.Noco.WebhookSecret != "secret" {
		t.Fatalf("expected secret, got %s", exp.Noco.WebhookSecret)
	}
	if exp.ApproverMapping["han@example.com"] != "123" {
		t.Fatalf("expected mapping for han@example.com=123, got %s", exp.ApproverMapping["han@example.com"])
	}
	if exp.ApproverMapping["ops@example.com"] != "456" {
		t.Fatalf("expected mapping for ops@example.com=456, got %s", exp.ApproverMapping["ops@example.com"])
	}
}

func TestGenerate_TaskOrderLogWorkerPoolSize(t *testing.T) {
	tests := []struct {
		name          string
		envValue      string
		expectedSize  int
		description   string
	}{
		{
			name:          "default value when not set",
			envValue:      "",
			expectedSize:  5,
			description:   "should default to 5 when env var not set",
		},
		{
			name:          "custom valid value",
			envValue:      "10",
			expectedSize:  10,
			description:   "should use custom value within valid range",
		},
		{
			name:          "minimum boundary - valid 1",
			envValue:      "1",
			expectedSize:  1,
			description:   "should accept minimum value of 1",
		},
		{
			name:          "maximum boundary - valid 20",
			envValue:      "20",
			expectedSize:  20,
			description:   "should accept maximum value of 20",
		},
		{
			name:          "below minimum - clamp to 1",
			envValue:      "0",
			expectedSize:  1,
			description:   "should clamp 0 to minimum of 1",
		},
		{
			name:          "negative value - clamp to 1",
			envValue:      "-5",
			expectedSize:  1,
			description:   "should clamp negative values to minimum of 1",
		},
		{
			name:          "above maximum - clamp to 20",
			envValue:      "25",
			expectedSize:  20,
			description:   "should clamp values above 20 to maximum of 20",
		},
		{
			name:          "very large value - clamp to 20",
			envValue:      "1000",
			expectedSize:  20,
			description:   "should clamp very large values to maximum of 20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &mockENV{
				values: map[string]string{},
				bools:  map[string]bool{},
			}

			if tt.envValue != "" {
				env.values["TASK_ORDER_LOG_WORKER_POOL_SIZE"] = tt.envValue
			}

			cfg := Generate(env)

			if cfg.TaskOrderLogWorkerPoolSize != tt.expectedSize {
				t.Errorf("%s: Generate() TaskOrderLogWorkerPoolSize = %d, want %d",
					tt.description, cfg.TaskOrderLogWorkerPoolSize, tt.expectedSize)
			}
		})
	}
}

func TestGenerate_TaskOrderLogWorkerPoolSize_WithOtherConfig(t *testing.T) {
	env := &mockENV{
		values: map[string]string{
			"TASK_ORDER_LOG_WORKER_POOL_SIZE": "15",
			"LOG_LEVEL":                       "debug",
			"ENV":                             "production",
		},
		bools: map[string]bool{
			"DEBUG": true,
		},
	}

	cfg := Generate(env)

	// Verify TaskOrderLogWorkerPoolSize
	if cfg.TaskOrderLogWorkerPoolSize != 15 {
		t.Errorf("Generate() TaskOrderLogWorkerPoolSize = %d, want 15", cfg.TaskOrderLogWorkerPoolSize)
	}

	// Ensure other fields are still populated correctly
	if cfg.Debug != true {
		t.Errorf("Generate() Debug = %v, want true", cfg.Debug)
	}
	if cfg.Env != "production" {
		t.Errorf("Generate() Env = %v, want production", cfg.Env)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("Generate() LogLevel = %v, want debug", cfg.LogLevel)
	}
}
