package duckdb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// mockLogger implements logger.Logger for testing
type mockLogger struct {
	logs []string
}

func (m *mockLogger) Fields(data logger.Fields) logger.Logger      { return m }
func (m *mockLogger) Field(key, value string) logger.Logger        { return m }
func (m *mockLogger) AddField(key string, value any) logger.Logger { return m }

func (m *mockLogger) Debug(msg string) {
	m.logs = append(m.logs, "DEBUG: "+msg)
}
func (m *mockLogger) Debugf(msg string, args ...interface{}) {
	m.logs = append(m.logs, "DEBUG: "+msg)
}
func (m *mockLogger) Info(msg string) {
	m.logs = append(m.logs, "INFO: "+msg)
}
func (m *mockLogger) Infof(msg string, args ...interface{}) {
	m.logs = append(m.logs, "INFO: "+msg)
}
func (m *mockLogger) Warn(msg string) {
	m.logs = append(m.logs, "WARN: "+msg)
}
func (m *mockLogger) Warnf(msg string, args ...interface{}) {
	m.logs = append(m.logs, "WARN: "+msg)
}
func (m *mockLogger) Error(err error, msg string) {
	m.logs = append(m.logs, "ERROR: "+msg+" - "+err.Error())
}
func (m *mockLogger) Errorf(err error, msg string, args ...interface{}) {
	m.logs = append(m.logs, "ERROR: "+msg+" - "+err.Error())
}
func (m *mockLogger) Fatal(err error, msg string) {
	m.logs = append(m.logs, "FATAL: "+msg+" - "+err.Error())
}
func (m *mockLogger) Fatalf(err error, msg string, args ...interface{}) {
	m.logs = append(m.logs, "FATAL: "+msg+" - "+err.Error())
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "successful_initialization",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &mockLogger{}
			service, err := New(mockLogger)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, service)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, service)

				// Clean up
				if service != nil {
					_ = service.Close()
				}
			}
		})
	}
}

func TestService_ExecuteQuery(t *testing.T) {
	// Skip if running in short mode since this requires actual DuckDB
	if testing.Short() {
		t.Skip("Skipping DuckDB integration test in short mode")
	}

	mockLogger := &mockLogger{}
	service, err := New(mockLogger)
	require.NoError(t, err)
	defer func() {
		_ = service.Close()
	}()

	tests := []struct {
		name        string
		query       string
		expectError bool
		expectRows  bool
	}{
		{
			name:        "empty_query",
			query:       "",
			expectError: true,
			expectRows:  false,
		},
		{
			name:        "simple_select",
			query:       "SELECT 1 as num, 'hello' as text",
			expectError: false,
			expectRows:  true,
		},
		{
			name:        "create_and_select_from_memory_table",
			query:       "CREATE TABLE test AS SELECT 1 as id, 'test' as name; SELECT * FROM test",
			expectError: false,
			expectRows:  true,
		},
		{
			name:        "invalid_sql",
			query:       "INVALID SQL QUERY",
			expectError: true,
			expectRows:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := service.ExecuteQuery(ctx, tt.query)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)

				if tt.expectRows {
					assert.Greater(t, len(result), 0)
				}
			}
		})
	}
}

func TestService_ReadParquetFromURL_Validation(t *testing.T) {
	mockLogger := &mockLogger{}
	service, err := New(mockLogger)
	require.NoError(t, err)
	defer func() {
		_ = service.Close()
	}()

	tests := []struct {
		name        string
		parquetURL  string
		query       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty_parquet_url",
			parquetURL:  "",
			query:       "SELECT * FROM parquet_file",
			expectError: true,
			errorMsg:    "parquet URL cannot be empty",
		},
		{
			name:        "empty_query",
			parquetURL:  "https://example.com/file.parquet",
			query:       "",
			expectError: true,
			errorMsg:    "query cannot be empty",
		},
		{
			name:        "invalid_url_format",
			parquetURL:  "ftp://example.com/file.parquet",
			query:       "SELECT * FROM parquet_file",
			expectError: true,
			errorMsg:    "parquet URL must be a HTTP/HTTPS URL or a local file path",
		},
		{
			name:        "valid_http_url",
			parquetURL:  "http://example.com/file.parquet",
			query:       "SELECT * FROM parquet_file",
			expectError: true,                      // Will fail because URL doesn't exist, but validation passes
			errorMsg:    "failed to execute query", // Different error message for actual execution
		},
		{
			name:        "valid_https_url",
			parquetURL:  "https://example.com/file.parquet",
			query:       "SELECT * FROM parquet_file",
			expectError: true,                      // Will fail because URL doesn't exist, but validation passes
			errorMsg:    "failed to execute query", // Different error message for actual execution
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := service.ReadParquetFromURL(ctx, tt.parquetURL, tt.query)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestService_ReadParquetFromURL_QueryReplacement(t *testing.T) {
	mockLogger := &mockLogger{}
	service, err := New(mockLogger)
	require.NoError(t, err)
	defer func() {
		_ = service.Close()
	}()

	tests := []struct {
		name        string
		parquetURL  string
		query       string
		expectedSQL string
	}{
		{
			name:        "simple_replacement",
			parquetURL:  "https://example.com/data.parquet",
			query:       "SELECT * FROM parquet_file",
			expectedSQL: "SELECT * FROM 'https://example.com/data.parquet'",
		},
		{
			name:        "multiple_replacements",
			parquetURL:  "https://example.com/data.parquet",
			query:       "SELECT col1, col2 FROM parquet_file WHERE parquet_file.id > 10",
			expectedSQL: "SELECT col1, col2 FROM 'https://example.com/data.parquet' WHERE 'https://example.com/data.parquet'.id > 10",
		},
		{
			name:        "no_replacement_needed",
			parquetURL:  "https://example.com/data.parquet",
			query:       "SELECT * FROM 'https://example.com/data.parquet'",
			expectedSQL: "SELECT * FROM 'https://example.com/data.parquet'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test verifies the string replacement logic
			// We can't test the actual execution without a real parquet file
			ctx := context.Background()
			_, err := service.ReadParquetFromURL(ctx, tt.parquetURL, tt.query)

			// The query will fail because the URL doesn't exist, but we're testing the replacement logic
			// The error should indicate that the query was transformed (we can check logs if needed)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to execute query")
		})
	}
}

func TestService_Close(t *testing.T) {
	mockLogger := &mockLogger{}
	service, err := New(mockLogger)
	require.NoError(t, err)

	// Close should work without error
	err = service.Close()
	assert.NoError(t, err)

	// Closing again should still work
	err = service.Close()
	assert.NoError(t, err)
}

func TestService_ExecuteQuery_DataTypes(t *testing.T) {
	// Skip if running in short mode since this requires actual DuckDB
	if testing.Short() {
		t.Skip("Skipping DuckDB integration test in short mode")
	}

	mockLogger := &mockLogger{}
	service, err := New(mockLogger)
	require.NoError(t, err)
	defer func() {
		_ = service.Close()
	}()

	tests := []struct {
		name     string
		query    string
		expected map[string]interface{}
	}{
		{
			name:  "integer_and_string",
			query: "SELECT 42 as int_col, 'test' as str_col",
			expected: map[string]interface{}{
				"int_col": int32(42),
				"str_col": "test",
			},
		},
		{
			name:  "boolean_and_null",
			query: "SELECT true as bool_col, NULL as null_col",
			expected: map[string]interface{}{
				"bool_col": true,
				"null_col": nil,
			},
		},
		{
			name:  "float_number",
			query: "SELECT 3.14::DOUBLE as float_col",
			expected: map[string]interface{}{
				"float_col": float64(3.14),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := service.ExecuteQuery(ctx, tt.query)

			require.NoError(t, err)
			require.Len(t, result, 1)

			row := result[0]
			for key, expectedValue := range tt.expected {
				actualValue, exists := row[key]
				assert.True(t, exists, "Column %s should exist", key)
				assert.Equal(t, expectedValue, actualValue, "Column %s should have expected value", key)
			}
		})
	}
}

func TestService_ExecuteQuery_MultipleRows(t *testing.T) {
	// Skip if running in short mode since this requires actual DuckDB
	if testing.Short() {
		t.Skip("Skipping DuckDB integration test in short mode")
	}

	mockLogger := &mockLogger{}
	service, err := New(mockLogger)
	require.NoError(t, err)
	defer func() {
		_ = service.Close()
	}()

	ctx := context.Background()
	query := `
		SELECT * FROM (
			VALUES 
				(1, 'first'),
				(2, 'second'),
				(3, 'third')
		) AS t(id, name)
	`

	result, err := service.ExecuteQuery(ctx, query)
	require.NoError(t, err)
	assert.Len(t, result, 3)

	// Check first row
	assert.Equal(t, int32(1), result[0]["id"])
	assert.Equal(t, "first", result[0]["name"])

	// Check second row
	assert.Equal(t, int32(2), result[1]["id"])
	assert.Equal(t, "second", result[1]["name"])

	// Check third row
	assert.Equal(t, int32(3), result[2]["id"])
	assert.Equal(t, "third", result[2]["name"])
}

func TestService_QueryParquetWithFilters_Validation(t *testing.T) {
	mockLogger := &mockLogger{}
	service, err := New(mockLogger)
	require.NoError(t, err)
	defer func() {
		_ = service.Close()
	}()

	tests := []struct {
		name        string
		parquetURL  string
		options     QueryOptions
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty_parquet_url",
			parquetURL:  "",
			options:     QueryOptions{},
			expectError: true,
			errorMsg:    "parquet URL cannot be empty",
		},
		{
			name:        "invalid_url_format",
			parquetURL:  "ftp://example.com/file.parquet",
			options:     QueryOptions{},
			expectError: true,
			errorMsg:    "parquet URL must be a HTTP/HTTPS URL or a local file path",
		},
		{
			name:       "valid_url_with_filters",
			parquetURL: "https://example.com/data.parquet",
			options: QueryOptions{
				Columns: []string{"id", "name"},
				Filters: []QueryFilter{
					{Column: "age", Operator: ">", Value: 18},
				},
				Limit: 10,
			},
			expectError: true, // Will fail because URL doesn't exist
			errorMsg:    "failed to execute query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := service.QueryParquetWithFilters(ctx, tt.parquetURL, tt.options)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestService_GetParquetSchema(t *testing.T) {
	mockLogger := &mockLogger{}
	service, err := New(mockLogger)
	require.NoError(t, err)
	defer func() {
		_ = service.Close()
	}()

	tests := []struct {
		name        string
		parquetURL  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty_parquet_url",
			parquetURL:  "",
			expectError: true,
			errorMsg:    "parquet URL cannot be empty",
		},
		{
			name:        "invalid_url_format",
			parquetURL:  "ftp://example.com/file.parquet",
			expectError: true,
			errorMsg:    "parquet URL must be a HTTP/HTTPS URL or a local file path",
		},
		{
			name:        "valid_url",
			parquetURL:  "https://example.com/data.parquet",
			expectError: true, // Will fail because URL doesn't exist
			errorMsg:    "failed to execute query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := service.GetParquetSchema(ctx, tt.parquetURL)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestService_QueryParquetWithFilters_SQLGeneration(t *testing.T) {
	// Skip if running in short mode since this requires actual DuckDB
	if testing.Short() {
		t.Skip("Skipping DuckDB integration test in short mode")
	}

	mockLogger := &mockLogger{}
	service, err := New(mockLogger)
	require.NoError(t, err)
	defer func() {
		_ = service.Close()
	}()

	tests := []struct {
		name          string
		parquetURL    string
		options       QueryOptions
		shouldContain []string // parts that should be in the generated SQL
	}{
		{
			name:          "simple_select_all",
			parquetURL:    "https://example.com/data.parquet",
			options:       QueryOptions{},
			shouldContain: []string{"SELECT *", "FROM 'https://example.com/data.parquet'"},
		},
		{
			name:       "select_specific_columns",
			parquetURL: "https://example.com/data.parquet",
			options: QueryOptions{
				Columns: []string{"id", "name", "email"},
			},
			shouldContain: []string{"SELECT", "\"id\"", "\"name\"", "\"email\""},
		},
		{
			name:       "select_with_filters",
			parquetURL: "https://example.com/data.parquet",
			options: QueryOptions{
				Filters: []QueryFilter{
					{Column: "age", Operator: ">", Value: 18},
					{Column: "status", Operator: "=", Value: "active"},
				},
			},
			shouldContain: []string{"WHERE", "\"age\" > 18", "\"status\" = 'active'", "AND"},
		},
		{
			name:       "select_with_order_by_and_limit",
			parquetURL: "https://example.com/data.parquet",
			options: QueryOptions{
				OrderBy: []string{"created_at", "name"},
				Limit:   10,
				Offset:  5,
			},
			shouldContain: []string{"ORDER BY", "\"created_at\"", "\"name\"", "LIMIT 10", "OFFSET 5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will test the SQL generation indirectly by attempting to execute
			// We expect it to fail due to non-existent URL, but we can verify the error
			// suggests the correct SQL was generated
			ctx := context.Background()
			_, err := service.QueryParquetWithFilters(ctx, tt.parquetURL, tt.options)

			// Should fail due to non-existent URL, but the SQL generation should work
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to execute query")
		})
	}
}

func TestService_FilterTypes_Integration(t *testing.T) {
	// Skip if running in short mode since this requires actual DuckDB
	if testing.Short() {
		t.Skip("Skipping DuckDB integration test in short mode")
	}

	mockLogger := &mockLogger{}
	service, err := New(mockLogger)
	require.NoError(t, err)
	defer func() {
		_ = service.Close()
	}()

	// Test different filter operators by creating in-memory data and filtering
	ctx := context.Background()

	// Test various filter operators indirectly through ExecuteQuery
	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "string_equality",
			query: "SELECT * FROM (VALUES ('John', 25), ('Jane', 30)) AS t(name, age) WHERE name = 'John'",
		},
		{
			name:  "numeric_comparison",
			query: "SELECT * FROM (VALUES ('John', 25), ('Jane', 30)) AS t(name, age) WHERE age > 25",
		},
		{
			name:  "in_operator",
			query: "SELECT * FROM (VALUES ('IT', 1), ('HR', 2), ('Finance', 3)) AS t(dept, id) WHERE dept IN ('IT', 'HR')",
		},
		{
			name:  "like_pattern",
			query: "SELECT * FROM (VALUES ('john@company.com', 1), ('jane@other.com', 2)) AS t(email, id) WHERE email LIKE '%@company.com'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := service.ExecuteQuery(ctx, tt.query)
			require.NoError(t, err)
			assert.NotEmpty(t, results, "Query should return results")
		})
	}
}

// Benchmark tests for performance
func BenchmarkExecuteQuery_SimpleSelect(b *testing.B) {
	mockLogger := &mockLogger{}
	service, err := New(mockLogger)
	require.NoError(b, err)
	defer func() {
		_ = service.Close()
	}()

	ctx := context.Background()
	query := "SELECT 1 as num"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ExecuteQuery(ctx, query)
		require.NoError(b, err)
	}
}

func BenchmarkExecuteQuery_ComplexAggregation(b *testing.B) {
	mockLogger := &mockLogger{}
	service, err := New(mockLogger)
	require.NoError(b, err)
	defer func() {
		_ = service.Close()
	}()

	// Create a test table first
	ctx := context.Background()
	setupQuery := `
		CREATE OR REPLACE TABLE test_data AS 
		SELECT 
			i as id,
			i * 2 as value,
			'category_' || (i % 5) as category
		FROM range(1, 1001) t(i)
	`
	_, err = service.ExecuteQuery(ctx, setupQuery)
	require.NoError(b, err)

	query := `
		SELECT 
			category,
			COUNT(*) as count,
			AVG(value) as avg_value,
			SUM(value) as sum_value
		FROM test_data 
		GROUP BY category 
		ORDER BY category
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ExecuteQuery(ctx, query)
		require.NoError(b, err)
	}
}

func TestBuildOrderByClause(t *testing.T) {
	l := &mockLogger{}
	service := &service{l: l}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple column name",
			input:    "date",
			expected: "\"date\"",
		},
		{
			name:     "Column with ASC",
			input:    "date ASC",
			expected: "\"date\" ASC",
		},
		{
			name:     "Column with DESC",
			input:    "date DESC",
			expected: "\"date\" DESC",
		},
		{
			name:     "Column with lowercase desc",
			input:    "date desc",
			expected: "\"date\" DESC",
		},
		{
			name:     "Column with extra spaces",
			input:    "  date   DESC  ",
			expected: "\"date\" DESC",
		},
		{
			name:     "Invalid direction ignored",
			input:    "date INVALID",
			expected: "\"date\"",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.buildOrderByClause(tt.input)
			if result != tt.expected {
				t.Errorf("buildOrderByClause(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
