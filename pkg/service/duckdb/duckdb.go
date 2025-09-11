package duckdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/marcboeker/go-duckdb/v2"

	"github.com/dwarvesf/fortress-api/pkg/logger"
)

type service struct {
	db *sql.DB
	l  logger.Logger
}

// New creates a new DuckDB service instance
func New(l logger.Logger) (IService, error) {
	// Create in-memory DuckDB database
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("failed to open DuckDB connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping DuckDB: %w", err)
	}

	// Install and load httpfs extension for URL support
	if _, err := db.Exec("INSTALL httpfs"); err != nil {
		l.Warnf("failed to install httpfs extension: %v", err)
	}
	if _, err := db.Exec("LOAD httpfs"); err != nil {
		l.Warnf("failed to load httpfs extension: %v", err)
	}

	return &service{
		db: db,
		l:  l,
	}, nil
}

// ReadParquetFromURL reads parquet data from a URL and executes a SQL query on it
func (s *service) ReadParquetFromURL(ctx context.Context, parquetURL, query string) ([]map[string]interface{}, error) {
	if parquetURL == "" {
		return nil, fmt.Errorf("parquet URL cannot be empty")
	}

	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	// Validate URL format - allow HTTP/HTTPS URLs and local file paths
	isHTTPURL := strings.HasPrefix(parquetURL, "http://") || strings.HasPrefix(parquetURL, "https://")
	isLocalPath := strings.HasPrefix(parquetURL, "/") || strings.Contains(parquetURL, ".parquet")
	if !isHTTPURL && !isLocalPath {
		return nil, fmt.Errorf("parquet URL must be a HTTP/HTTPS URL or a local file path")
	}

	// Replace the table reference in the query with the parquet URL
	// This allows flexible queries like "SELECT * FROM parquet_file WHERE condition"
	finalQuery := strings.ReplaceAll(query, "parquet_file", fmt.Sprintf("'%s'", parquetURL))

	s.l.Infof("executing DuckDB query on parquet URL: %s", parquetURL)
	return s.ExecuteQuery(ctx, finalQuery)
}

// ExecuteQuery executes a raw SQL query on the DuckDB connection
func (s *service) ExecuteQuery(ctx context.Context, query string) ([]map[string]interface{}, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		s.l.Errorf(err, "failed to execute query")
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create slice to hold column values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan row values
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create result map
		row := make(map[string]interface{})
		for i, col := range columns {
			// Handle different data types
			val := values[i]
			if val != nil {
				switch v := val.(type) {
				case []byte:
					// Convert byte arrays to strings
					row[col] = string(v)
				default:
					row[col] = v
				}
			} else {
				row[col] = nil
			}
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	s.l.Infof("query executed successfully, returned %d rows", len(results))
	return results, nil
}

// QueryParquetWithFilters queries parquet data from URL with structured filters
func (s *service) QueryParquetWithFilters(ctx context.Context, parquetURL string, options QueryOptions) ([]map[string]interface{}, error) {
	if parquetURL == "" {
		return nil, fmt.Errorf("parquet URL cannot be empty")
	}

	// Validate URL format - allow HTTP/HTTPS URLs and local file paths
	isHTTPURL := strings.HasPrefix(parquetURL, "http://") || strings.HasPrefix(parquetURL, "https://")
	isLocalPath := strings.HasPrefix(parquetURL, "/") || strings.Contains(parquetURL, ".parquet")
	if !isHTTPURL && !isLocalPath {
		return nil, fmt.Errorf("parquet URL must be a HTTP/HTTPS URL or a local file path")
	}

	query := s.buildQuery(parquetURL, options)
	s.l.Infof("executing structured query on parquet URL: %s", parquetURL)

	return s.ExecuteQuery(ctx, query)
}

// GetParquetSchema returns the schema/structure of a parquet file
func (s *service) GetParquetSchema(ctx context.Context, parquetURL string) ([]map[string]interface{}, error) {
	if parquetURL == "" {
		return nil, fmt.Errorf("parquet URL cannot be empty")
	}

	// Validate URL format - allow HTTP/HTTPS URLs and local file paths
	isHTTPURL := strings.HasPrefix(parquetURL, "http://") || strings.HasPrefix(parquetURL, "https://")
	isLocalPath := strings.HasPrefix(parquetURL, "/") || strings.Contains(parquetURL, ".parquet")
	if !isHTTPURL && !isLocalPath {
		return nil, fmt.Errorf("parquet URL must be a HTTP/HTTPS URL or a local file path")
	}

	// Use DESCRIBE to get schema information
	query := fmt.Sprintf("DESCRIBE SELECT * FROM '%s' LIMIT 0", parquetURL)
	s.l.Infof("getting schema for parquet URL: %s", parquetURL)

	return s.ExecuteQuery(ctx, query)
}

// buildQuery constructs a SQL query from QueryOptions
func (s *service) buildQuery(parquetURL string, options QueryOptions) string {
	var query strings.Builder

	// SELECT clause
	query.WriteString("SELECT ")
	if len(options.Columns) > 0 {
		for i, col := range options.Columns {
			if i > 0 {
				query.WriteString(", ")
			}
			query.WriteString(s.escapeColumnName(col))
		}
	} else {
		query.WriteString("*")
	}

	// FROM clause
	query.WriteString(fmt.Sprintf(" FROM '%s'", parquetURL))

	// WHERE clause
	if len(options.Filters) > 0 {
		query.WriteString(" WHERE ")
		for i, filter := range options.Filters {
			if i > 0 {
				query.WriteString(" AND ")
			}
			query.WriteString(s.buildFilterCondition(filter))
		}
	}

	// ORDER BY clause
	if len(options.OrderBy) > 0 {
		query.WriteString(" ORDER BY ")
		for i, col := range options.OrderBy {
			if i > 0 {
				query.WriteString(", ")
			}
			query.WriteString(s.buildOrderByClause(col))
		}
	}

	// LIMIT clause
	if options.Limit > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", options.Limit))
	}

	// OFFSET clause
	if options.Offset > 0 {
		query.WriteString(fmt.Sprintf(" OFFSET %d", options.Offset))
	}

	return query.String()
}

// buildFilterCondition builds a single filter condition
func (s *service) buildFilterCondition(filter QueryFilter) string {
	column := s.escapeColumnName(filter.Column)

	switch strings.ToUpper(filter.Operator) {
	case "=", "!=", ">", "<", ">=", "<=":
		return fmt.Sprintf("%s %s %s", column, filter.Operator, s.escapeValue(filter.Value))
	case "LIKE":
		return fmt.Sprintf("%s LIKE %s", column, s.escapeValue(filter.Value))
	case "IN":
		if slice, ok := filter.Value.([]interface{}); ok {
			var values []string
			for _, v := range slice {
				values = append(values, s.escapeValue(v))
			}
			return fmt.Sprintf("%s IN (%s)", column, strings.Join(values, ", "))
		}
		return fmt.Sprintf("%s IN (%s)", column, s.escapeValue(filter.Value))
	case "IS NULL":
		return fmt.Sprintf("%s IS NULL", column)
	case "IS NOT NULL":
		return fmt.Sprintf("%s IS NOT NULL", column)
	default:
		// Default to equality
		return fmt.Sprintf("%s = %s", column, s.escapeValue(filter.Value))
	}
}

// buildOrderByClause builds ORDER BY clause handling column names and ASC/DESC
func (s *service) buildOrderByClause(orderBy string) string {
	parts := strings.Fields(strings.TrimSpace(orderBy))
	if len(parts) == 0 {
		return ""
	}

	// First part is always the column name
	column := s.escapeColumnName(parts[0])

	// Check if there's a direction (ASC/DESC)
	if len(parts) > 1 {
		direction := strings.ToUpper(parts[1])
		if direction == "ASC" || direction == "DESC" {
			return fmt.Sprintf("%s %s", column, direction)
		}
	}

	return column
}

// escapeColumnName escapes column names to prevent SQL injection
func (s *service) escapeColumnName(column string) string {
	// Remove any quotes and wrap in double quotes
	cleaned := strings.ReplaceAll(column, "\"", "")
	return fmt.Sprintf("\"%s\"", cleaned)
}

// escapeValue escapes values based on their type
func (s *service) escapeValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Escape single quotes in strings
		escaped := strings.ReplaceAll(v, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case nil:
		return "NULL"
	default:
		// Convert to string and treat as string
		escaped := strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	}
}

// Close closes the database connection
func (s *service) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
