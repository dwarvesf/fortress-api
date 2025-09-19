package discord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/discord/helpers"
	"github.com/dwarvesf/fortress-api/pkg/service/duckdb"
	"github.com/dwarvesf/fortress-api/pkg/service/parquet"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/memolog"
)

// Mock DuckDB service for testing
type mockDuckDBService struct {
	mock.Mock
}

func (m *mockDuckDBService) QueryParquetWithFilters(ctx context.Context, parquetURL string, options duckdb.QueryOptions) ([]map[string]interface{}, error) {
	args := m.Called(ctx, parquetURL, options)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *mockDuckDBService) ReadParquetFromURL(ctx context.Context, parquetURL, query string) ([]map[string]interface{}, error) {
	args := m.Called(ctx, parquetURL, query)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *mockDuckDBService) ExecuteQuery(ctx context.Context, query string) ([]map[string]interface{}, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *mockDuckDBService) GetParquetSchema(ctx context.Context, parquetURL string) ([]map[string]interface{}, error) {
	args := m.Called(ctx, parquetURL)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *mockDuckDBService) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Mock MemoLog store for testing
type mockMemoLogStore struct {
	mock.Mock
}

func (m *mockMemoLogStore) Create(db *gorm.DB, b []model.MemoLog) ([]model.MemoLog, error) {
	args := m.Called(db, b)
	return args.Get(0).([]model.MemoLog), args.Error(1)
}

func (m *mockMemoLogStore) GetLimitByTimeRange(db *gorm.DB, start, end *time.Time, limit int) ([]model.MemoLog, error) {
	args := m.Called(db, start, end, limit)
	return args.Get(0).([]model.MemoLog), args.Error(1)
}

func (m *mockMemoLogStore) List(db *gorm.DB, filter memolog.ListFilter) ([]model.MemoLog, error) {
	args := m.Called(db, filter)
	return args.Get(0).([]model.MemoLog), args.Error(1)
}

func (m *mockMemoLogStore) GetRankByDiscordID(db *gorm.DB, discordID string) (*model.DiscordAccountMemoRank, error) {
	args := m.Called(db, discordID)
	return args.Get(0).(*model.DiscordAccountMemoRank), args.Error(1)
}

func (m *mockMemoLogStore) ListNonAuthor(db *gorm.DB) ([]model.MemoLog, error) {
	args := m.Called(db)
	return args.Get(0).([]model.MemoLog), args.Error(1)
}

func (m *mockMemoLogStore) CreateMemoAuthor(db *gorm.DB, memoAuthor *model.MemoAuthor) error {
	args := m.Called(db, memoAuthor)
	return args.Error(0)
}

func (m *mockMemoLogStore) GetTopAuthors(db *gorm.DB, limit int, from, to *time.Time) ([]model.DiscordAccountMemoRank, error) {
	args := m.Called(db, limit, from, to)
	return args.Get(0).([]model.DiscordAccountMemoRank), args.Error(1)
}

// Mock database repository for testing
type mockDBRepo struct {
	mock.Mock
}

func (m *mockDBRepo) DB() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func (m *mockDBRepo) NewTransaction() (store.DBRepo, store.FinallyFunc) {
	args := m.Called()
	return args.Get(0).(store.DBRepo), args.Get(1).(store.FinallyFunc)
}

func (m *mockDBRepo) SetNewDB(db *gorm.DB) {
	m.Called(db)
}

// Mock ParquetSync service for testing
type mockParquetSyncService struct {
	mock.Mock
}

func (m *mockParquetSyncService) StartBackgroundSync(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockParquetSyncService) StopBackgroundSync() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockParquetSyncService) SyncNow(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockParquetSyncService) GetLocalFilePath() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockParquetSyncService) IsLocalFileReady() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockParquetSyncService) GetSyncStatus() parquet.SyncStatus {
	args := m.Called()
	return args.Get(0).(parquet.SyncStatus)
}

func (m *mockParquetSyncService) GetRemoteURL() string {
	args := m.Called()
	return args.String(0)
}

// setupBasicParquetSyncExpectations sets up common ParquetSync mock expectations
func setupBasicParquetSyncExpectations(mockParquetSync *mockParquetSyncService) {
	mockParquetSync.On("IsLocalFileReady").Return(false)
	mockParquetSync.On("GetRemoteURL").Return("https://github.com/dwarvesf/memo.d.foundation/raw/refs/heads/main/db/vault.parquet")
}

// Setup test handler with mocked dependencies
func setupDuckDBIntegrationTest() (*handler, *mockDuckDBService, *mockParquetSyncService, *mockMemoLogStore, *mockDBRepo) {
	mockDuckDB := &mockDuckDBService{}
	mockMemoLogStore := &mockMemoLogStore{}
	mockDBRepo := &mockDBRepo{}
	mockParquetSync := &mockParquetSyncService{}

	// Mock stores
	mockStore := &store.Store{
		MemoLog: mockMemoLogStore,
	}

	// Mock services
	mockService := &service.Service{
		DuckDB:      mockDuckDB,
		ParquetSync: mockParquetSync,
	}

	// Create handler
	h := New(
		&controller.Controller{},
		mockStore,
		mockDBRepo,
		mockService,
		logger.NewLogrusLogger(),
		&config.Config{},
	)

	return h.(*handler), mockDuckDB, mockParquetSync, mockMemoLogStore, mockDBRepo
}

func TestHandler_getMemosFromParquet_Success(t *testing.T) {
	h, mockDuckDB, mockParquetSync, _, _ := setupDuckDBIntegrationTest()

	// Test parameters
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	// Mock ParquetSync expectations
	mockParquetSync.On("IsLocalFileReady").Return(false)
	mockParquetSync.On("GetRemoteURL").Return("https://github.com/dwarvesf/memo.d.foundation/raw/refs/heads/main/db/vault.parquet")

	// Mock successful parquet query response
	parquetData := []map[string]interface{}{
		{
			"date":      "2024-01-03",
			"title":     "Test Memo 1",
			"authors":   []interface{}{"author1", "author2"},
			"tags":      []interface{}{"tag1", "tag2"},
			"file_path": "test/memo-1.md",
			"content":   "This is test content 1",
		},
		{
			"date":      "2024-01-05",
			"title":     "Test Memo 2",
			"authors":   []interface{}{"author3"},
			"tags":      []interface{}{"tag3"},
			"file_path": "test/memo-2.md",
			"content":   "This is test content 2",
		},
	}

	// Set up expectations
	expectedURL := "https://github.com/dwarvesf/memo.d.foundation/raw/refs/heads/main/db/vault.parquet"
	expectedOptions := duckdb.QueryOptions{
		Filters: []duckdb.QueryFilter{
			{Column: "date", Operator: ">=", Value: "2024-01-01"},
			{Column: "date", Operator: "<=", Value: "2024-01-07"},
		},
		OrderBy: []string{"date DESC"},
		Limit:   100,
		Offset:  0,
	}

	mockDuckDB.On("QueryParquetWithFilters", mock.AnythingOfType("*context.timerCtx"), expectedURL, expectedOptions).Return(parquetData, nil)

	// Execute test
	result, err := h.getMemosFromParquet(start, end, "weekly")

	// Verify results
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	
	// Verify first memo
	assert.Equal(t, "Test Memo 1", result[0].Title)
	assert.Equal(t, []string{"author1", "author2"}, result[0].AuthorMemoUsernames)
	assert.Contains(t, []string(result[0].Tags), "tag1")
	assert.Contains(t, []string(result[0].Tags), "tag2")
	assert.Equal(t, "https://memo.d.foundation/test/memo-1/", result[0].URL)
	assert.Equal(t, "This is test content 1", result[0].Description)

	// Verify second memo
	assert.Equal(t, "Test Memo 2", result[1].Title)
	assert.Equal(t, []string{"author3"}, result[1].AuthorMemoUsernames)
	assert.Contains(t, []string(result[1].Tags), "tag3")
	assert.Equal(t, "https://memo.d.foundation/test/memo-2/", result[1].URL)
	assert.Equal(t, "This is test content 2", result[1].Description)

	// Verify mock was called
	mockDuckDB.AssertExpectations(t)
}

func TestHandler_getMemosFromParquet_DuckDBServiceUnavailable(t *testing.T) {
	// Setup handler without DuckDB service
	mockMemoLogStore := &mockMemoLogStore{}
	mockDBRepo := &mockDBRepo{}

	mockStore := &store.Store{
		MemoLog: mockMemoLogStore,
	}

	// Service without DuckDB
	mockService := &service.Service{
		DuckDB: nil,
	}

	h := New(
		&controller.Controller{},
		mockStore,
		mockDBRepo,
		mockService,
		logger.NewLogrusLogger(),
		&config.Config{},
	).(*handler)

	// Test parameters
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	// Mock database fallback
	mockDB := &gorm.DB{}
	expectedMemos := []model.MemoLog{
		{
			Title:                "Database Memo",
			AuthorMemoUsernames:  []string{"db_author"},
		},
	}

	mockDBRepo.On("DB").Return(mockDB)
	mockMemoLogStore.On("GetLimitByTimeRange", mockDB, &start, &end, 1000).Return(expectedMemos, nil)

	// Execute test
	result, err := h.getMemosFromParquet(start, end, "weekly")

	// Verify results
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Database Memo", result[0].Title)
	assert.Equal(t, []string{"db_author"}, result[0].AuthorMemoUsernames)

	// Verify mocks
	mockDBRepo.AssertExpectations(t)
	mockMemoLogStore.AssertExpectations(t)
}

func TestHandler_getMemosFromParquet_DuckDBQueryError(t *testing.T) {
	h, mockDuckDB, mockParquetSync, mockMemoLogStore, mockDBRepo := setupDuckDBIntegrationTest()

	// Test parameters
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	// Setup ParquetSync expectations
	setupBasicParquetSyncExpectations(mockParquetSync)

	// Mock DuckDB failure
	mockDuckDB.On("QueryParquetWithFilters", mock.Anything, mock.Anything, mock.Anything).Return([]map[string]interface{}{}, errors.New("DuckDB connection failed"))

	// Mock database fallback
	mockDB := &gorm.DB{}
	expectedMemos := []model.MemoLog{
		{
			Title:               "Fallback Memo",
			AuthorMemoUsernames: []string{"fallback_author"},
		},
	}

	mockDBRepo.On("DB").Return(mockDB)
	mockMemoLogStore.On("GetLimitByTimeRange", mockDB, &start, &end, 1000).Return(expectedMemos, nil)

	// Execute test
	result, err := h.getMemosFromParquet(start, end, "weekly")

	// Verify results - should fall back to database
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Fallback Memo", result[0].Title)
	assert.Equal(t, []string{"fallback_author"}, result[0].AuthorMemoUsernames)

	// Verify mocks
	mockDuckDB.AssertExpectations(t)
	mockDBRepo.AssertExpectations(t)
	mockMemoLogStore.AssertExpectations(t)
}

func TestHandler_getMemosFromParquet_DataTransformationError(t *testing.T) {
	h, mockDuckDB, mockParquetSync, mockMemoLogStore, mockDBRepo := setupDuckDBIntegrationTest()

	// Test parameters
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)
	
	// Setup ParquetSync expectations
	setupBasicParquetSyncExpectations(mockParquetSync)

	// Mock parquet data with invalid format that will cause transformation error
	invalidParquetData := []map[string]interface{}{
		{
			"date":    "invalid-date-format", // This will cause transformation error
			"title":   "Test Memo",
			"authors": []interface{}{"author1"},
			"tags":    []interface{}{"tag1"},
			"content": "Content",
		},
	}

	mockDuckDB.On("QueryParquetWithFilters", mock.Anything, mock.Anything, mock.Anything).Return(invalidParquetData, nil)

	// Mock database fallback
	mockDB := &gorm.DB{}
	expectedMemos := []model.MemoLog{
		{
			Title:               "Database Fallback Memo",
			AuthorMemoUsernames: []string{"db_author"},
		},
	}

	mockDBRepo.On("DB").Return(mockDB)
	mockMemoLogStore.On("GetLimitByTimeRange", mockDB, &start, &end, 1000).Return(expectedMemos, nil)

	// Execute test
	result, err := h.getMemosFromParquet(start, end, "weekly")

	// Verify results - should fall back to database due to transformation error
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Database Fallback Memo", result[0].Title)

	// Verify mocks
	mockDuckDB.AssertExpectations(t)
	mockDBRepo.AssertExpectations(t)
	mockMemoLogStore.AssertExpectations(t)
}

func TestHandler_getMemosFromParquet_EmptyParquetData(t *testing.T) {
	h, mockDuckDB, mockParquetSync, _, _ := setupDuckDBIntegrationTest()

	// Setup ParquetSync expectations
	setupBasicParquetSyncExpectations(mockParquetSync)

	// Test parameters
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	// Mock empty parquet response
	emptyData := []map[string]interface{}{}
	mockDuckDB.On("QueryParquetWithFilters", mock.Anything, mock.Anything, mock.Anything).Return(emptyData, nil)

	// Execute test
	result, err := h.getMemosFromParquet(start, end, "weekly")

	// Verify results
	assert.NoError(t, err)
	assert.Empty(t, result)

	// Verify mock
	mockDuckDB.AssertExpectations(t)
}

func TestHandler_getMemosFromParquet_SingleAuthorAndTag(t *testing.T) {
	h, mockDuckDB, mockParquetSync, _, _ := setupDuckDBIntegrationTest()

	// Setup ParquetSync expectations
	setupBasicParquetSyncExpectations(mockParquetSync)

	// Test parameters
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	// Mock parquet data with single author/tag as string instead of array
	parquetData := []map[string]interface{}{
		{
			"date":    "2024-01-03",
			"title":   "Single Author Memo",
			"authors": "single_author", // String instead of array
			"tags":    "single_tag",    // String instead of array
			"url":     "https://example.com/memo",
			"content": "Content with single author",
		},
	}

	mockDuckDB.On("QueryParquetWithFilters", mock.Anything, mock.Anything, mock.Anything).Return(parquetData, nil)

	// Execute test
	result, err := h.getMemosFromParquet(start, end, "weekly")

	// Verify results
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Single Author Memo", result[0].Title)
	assert.Equal(t, []string{"single_author"}, result[0].AuthorMemoUsernames)
	assert.Contains(t, []string(result[0].Tags), "single_tag")

	// Verify mock
	mockDuckDB.AssertExpectations(t)
}

func TestHandler_getMemosFromParquet_DateFiltering(t *testing.T) {
	h, mockDuckDB, mockParquetSync, _, _ := setupDuckDBIntegrationTest()

	// Setup ParquetSync expectations
	setupBasicParquetSyncExpectations(mockParquetSync)

	// Test parameters - narrow date range
	start := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

	// Mock parquet data with dates both inside and outside range
	parquetData := []map[string]interface{}{
		{
			"date":    "2024-01-02", // Before range
			"title":   "Before Range",
			"authors": []interface{}{"author1"},
			"tags":    []interface{}{"tag1"},
			"content": "Before range content",
		},
		{
			"date":    "2024-01-04", // Within range
			"title":   "Within Range",
			"authors": []interface{}{"author2"},
			"tags":    []interface{}{"tag2"},
			"content": "Within range content",
		},
		{
			"date":    "2024-01-06", // After range
			"title":   "After Range",
			"authors": []interface{}{"author3"},
			"tags":    []interface{}{"tag3"},
			"content": "After range content",
		},
	}

	mockDuckDB.On("QueryParquetWithFilters", mock.Anything, mock.Anything, mock.Anything).Return(parquetData, nil)

	// Execute test
	result, err := h.getMemosFromParquet(start, end, "weekly")

	// Verify results - only memo within range should be returned
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Within Range", result[0].Title)
	assert.Equal(t, []string{"author2"}, result[0].AuthorMemoUsernames)

	// Verify mock
	mockDuckDB.AssertExpectations(t)
}

func TestHandler_getMemosFromParquet_ComplexDataTypes(t *testing.T) {
	h, mockDuckDB, mockParquetSync, _, _ := setupDuckDBIntegrationTest()

	// Setup ParquetSync expectations
	setupBasicParquetSyncExpectations(mockParquetSync)

	// Test parameters
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	// Mock parquet data with complex nested structures and missing fields
	parquetData := []map[string]interface{}{
		{
			"date":  "2024-01-03",
			"title": "Complex Memo",
			"authors": []interface{}{
				"author1",
				123,      // Non-string author (should be filtered out)
				"author2",
			},
			"tags": []interface{}{
				"tag1",
				nil,     // Nil tag (should be filtered out) 
				"tag2",
			},
			"url":     "",              // Empty URL
			"content": "Complex content",
			// Missing some fields to test robustness
		},
		{
			"date":    "2024-01-04",
			"title":   "Minimal Memo",
			"authors": []interface{}{"minimal_author"},
			// Missing tags, url, content
		},
	}

	mockDuckDB.On("QueryParquetWithFilters", mock.Anything, mock.Anything, mock.Anything).Return(parquetData, nil)

	// Execute test
	result, err := h.getMemosFromParquet(start, end, "weekly")

	// Verify results
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// First memo
	assert.Equal(t, "Complex Memo", result[0].Title)
	assert.Equal(t, []string{"author1", "author2"}, result[0].AuthorMemoUsernames) // Non-string filtered out
	assert.Equal(t, "", result[0].URL) // Empty URL preserved
	assert.Equal(t, "Complex content", result[0].Description)

	// Second memo
	assert.Equal(t, "Minimal Memo", result[1].Title)
	assert.Equal(t, []string{"minimal_author"}, result[1].AuthorMemoUsernames)
	assert.Empty(t, result[1].URL)
	assert.Empty(t, result[1].Description)

	// Verify mock
	mockDuckDB.AssertExpectations(t)
}

func TestHandler_getMemosFromParquet_TimeoutHandling(t *testing.T) {
	h, mockDuckDB, mockParquetSync, mockMemoLogStore, mockDBRepo := setupDuckDBIntegrationTest()

	// Setup ParquetSync expectations
	setupBasicParquetSyncExpectations(mockParquetSync)

	// Test parameters
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	// Mock DuckDB timeout (context deadline exceeded)
	mockDuckDB.On("QueryParquetWithFilters", mock.Anything, mock.Anything, mock.Anything).Return(
		[]map[string]interface{}{}, 
		context.DeadlineExceeded,
	)

	// Mock database fallback
	mockDB := &gorm.DB{}
	expectedMemos := []model.MemoLog{
		{
			Title:               "Timeout Fallback Memo",
			AuthorMemoUsernames: []string{"timeout_author"},
		},
	}

	mockDBRepo.On("DB").Return(mockDB)
	mockMemoLogStore.On("GetLimitByTimeRange", mockDB, &start, &end, 1000).Return(expectedMemos, nil)

	// Execute test
	result, err := h.getMemosFromParquet(start, end, "weekly")

	// Verify results - should fall back to database due to timeout
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Timeout Fallback Memo", result[0].Title)

	// Verify mocks
	mockDuckDB.AssertExpectations(t)
	mockDBRepo.AssertExpectations(t)
	mockMemoLogStore.AssertExpectations(t)
}

func TestHandler_getMemosFromParquet_DataTransformerConfiguration(t *testing.T) {
	h, mockDuckDB, mockParquetSync, _, _ := setupDuckDBIntegrationTest()

	// Setup ParquetSync expectations
	setupBasicParquetSyncExpectations(mockParquetSync)

	// Verify that data transformer is configured correctly
	transformer := h.dataTransformer
	assert.NotNil(t, transformer)

	// Test data transformer configuration through behavior
	testRecord := helpers.ParquetMemoRecord{
		Date:    "2024-01-03",
		Title:   "Test Config Memo",
		Authors: []string{"test_author"},
		Tags:    []string{"test_tag"},
		Content: "Test content for config verification",
	}

	result, err := transformer.TransformParquetRecord(testRecord)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "25", result.Reward.String()) // Default reward from config
	assert.Contains(t, []string(result.Category), "others") // Default category from config

	// Test parameters for integration
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	// Mock parquet data
	parquetData := []map[string]interface{}{
		{
			"date":    "2024-01-03",
			"title":   "Config Test Memo",
			"authors": []interface{}{"config_author"},
			"tags":    []interface{}{"config_tag"},
			"content": "Configuration test content",
		},
	}

	mockDuckDB.On("QueryParquetWithFilters", mock.Anything, mock.Anything, mock.Anything).Return(parquetData, nil)

	// Execute integration test
	integrationResult, integrationErr := h.getMemosFromParquet(start, end, "weekly")

	// Verify integration results
	assert.NoError(t, integrationErr)
	assert.Len(t, integrationResult, 1)
	assert.Equal(t, "25", integrationResult[0].Reward.String()) // Verify transformer config applied
	assert.Contains(t, []string(integrationResult[0].Category), "others")

	// Verify mock
	mockDuckDB.AssertExpectations(t)
}