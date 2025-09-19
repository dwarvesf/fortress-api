package duckdb

import (
	"context"
)

// QueryFilter represents a filter condition for querying parquet data
type QueryFilter struct {
	Column   string      `json:"column"`
	Operator string      `json:"operator"` // =, !=, >, <, >=, <=, LIKE, IN
	Value    interface{} `json:"value"`
}

// QueryOptions represents options for querying parquet data
type QueryOptions struct {
	Columns []string      `json:"columns,omitempty"`  // specific columns to select, empty means all
	Filters []QueryFilter `json:"filters,omitempty"`  // WHERE conditions
	OrderBy []string      `json:"order_by,omitempty"` // ORDER BY columns
	Limit   int           `json:"limit,omitempty"`    // LIMIT rows
	Offset  int           `json:"offset,omitempty"`   // OFFSET rows
}

// IService defines the interface for DuckDB service
type IService interface {
	// ReadParquetFromURL reads parquet data from a URL and executes a SQL query on it
	ReadParquetFromURL(ctx context.Context, parquetURL, query string) ([]map[string]interface{}, error)

	// QueryParquetWithFilters queries parquet data from URL with structured filters
	QueryParquetWithFilters(ctx context.Context, parquetURL string, options QueryOptions) ([]map[string]interface{}, error)

	// ExecuteQuery executes a raw SQL query on the DuckDB connection
	ExecuteQuery(ctx context.Context, query string) ([]map[string]interface{}, error)

	// GetParquetSchema returns the schema/structure of a parquet file
	GetParquetSchema(ctx context.Context, parquetURL string) ([]map[string]interface{}, error)

	// Close closes the database connection
	Close() error
}
