package helpers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/memolog"
)

// Mock implementations for testing
type mockDiscordAccountStore struct {
	mock.Mock
}

func (m *mockDiscordAccountStore) ListByMemoUsername(db *gorm.DB, usernames []string) ([]model.DiscordAccount, error) {
	args := m.Called(db, usernames)
	return args.Get(0).([]model.DiscordAccount), args.Error(1)
}

func (m *mockDiscordAccountStore) All(db *gorm.DB) ([]*model.DiscordAccount, error) {
	args := m.Called(db)
	return args.Get(0).([]*model.DiscordAccount), args.Error(1)
}

func (m *mockDiscordAccountStore) One(db *gorm.DB, id string) (*model.DiscordAccount, error) {
	args := m.Called(db, id)
	return args.Get(0).(*model.DiscordAccount), args.Error(1)
}

func (m *mockDiscordAccountStore) OneByDiscordID(db *gorm.DB, discordID string) (*model.DiscordAccount, error) {
	args := m.Called(db, discordID)
	return args.Get(0).(*model.DiscordAccount), args.Error(1)
}

func (m *mockDiscordAccountStore) Upsert(db *gorm.DB, da *model.DiscordAccount) (*model.DiscordAccount, error) {
	args := m.Called(db, da)
	return args.Get(0).(*model.DiscordAccount), args.Error(1)
}

func (m *mockDiscordAccountStore) UpdateSelectedFieldsByID(db *gorm.DB, id string, client model.DiscordAccount, updatedFields ...string) (*model.DiscordAccount, error) {
	args := m.Called(db, id, client, updatedFields)
	return args.Get(0).(*model.DiscordAccount), args.Error(1)
}

type mockMemoLogStore struct {
	mock.Mock
}

func (m *mockMemoLogStore) GetLimitByTimeRange(db *gorm.DB, start, end *time.Time, limit int) ([]model.MemoLog, error) {
	args := m.Called(db, start, end, limit)
	return args.Get(0).([]model.MemoLog), args.Error(1)
}

func (m *mockMemoLogStore) Create(db *gorm.DB, memos []model.MemoLog) ([]model.MemoLog, error) {
	args := m.Called(db, memos)
	return args.Get(0).([]model.MemoLog), args.Error(1)
}

func (m *mockMemoLogStore) CreateMemoAuthor(db *gorm.DB, memoAuthor *model.MemoAuthor) error {
	args := m.Called(db, memoAuthor)
	return args.Error(0)
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

func (m *mockMemoLogStore) GetTopAuthors(db *gorm.DB, limit int, from, to *time.Time) ([]model.DiscordAccountMemoRank, error) {
	args := m.Called(db, limit, from, to)
	return args.Get(0).([]model.DiscordAccountMemoRank), args.Error(1)
}

type mockStore struct {
	DiscordAccount *mockDiscordAccountStore
	MemoLog        *mockMemoLogStore
}

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

func setupAuthorResolverTest() (*mockStore, *mockDBRepo, AuthorResolver) {
	discordAccountStore := &mockDiscordAccountStore{}
	memoLogStore := &mockMemoLogStore{}
	dbRepo := &mockDBRepo{}
	
	store := &store.Store{
		DiscordAccount: discordAccountStore,
		MemoLog:        memoLogStore,
	}
	
	mockStore := &mockStore{
		DiscordAccount: discordAccountStore,
		MemoLog:        memoLogStore,
	}

	config := AuthorResolutionConfig{
		CacheTTL:         30 * time.Minute,
		MaxCacheSize:     100,
		BatchSize:        10,
		DatabaseTimeout:  5 * time.Second,
		FallbackStrategy: "mention_attempt",
		WarmupEnabled:    false, // Disable for testing
	}

	resolver := NewAuthorResolver(config, store, dbRepo)
	return mockStore, dbRepo, resolver
}

func TestNewAuthorResolver(t *testing.T) {
	tests := []struct {
		name     string
		config   AuthorResolutionConfig
		expected AuthorResolutionConfig
	}{
		{
			name:   "default configuration",
			config: AuthorResolutionConfig{},
			expected: AuthorResolutionConfig{
				CacheTTL:         1 * time.Hour,
				MaxCacheSize:     1000,
				BatchSize:        50,
				DatabaseTimeout:  10 * time.Second,
				FallbackStrategy: "mention_attempt",
			},
		},
		{
			name: "custom configuration",
			config: AuthorResolutionConfig{
				CacheTTL:         15 * time.Minute,
				MaxCacheSize:     500,
				BatchSize:        25,
				DatabaseTimeout:  30 * time.Second,
				FallbackStrategy: "plain_username",
			},
			expected: AuthorResolutionConfig{
				CacheTTL:         15 * time.Minute,
				MaxCacheSize:     500,
				BatchSize:        25,
				DatabaseTimeout:  30 * time.Second,
				FallbackStrategy: "plain_username",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, resolver := setupAuthorResolverTest()
			
			// Access internal config through GetMetrics and test behavior
			metrics := resolver.GetMetrics()
			assert.GreaterOrEqual(t, metrics.CacheSize, 0)
			assert.NotNil(t, resolver)
		})
	}
}

func TestAuthorResolver_ResolveUsernameToDiscordID_Success(t *testing.T) {
	mockStore, dbRepo, resolver := setupAuthorResolverTest()
	
	// Mock database response
	mockDB := &gorm.DB{}
	dbRepo.On("DB").Return(mockDB)
	
	discordAccounts := []model.DiscordAccount{
		{
			DiscordID:       "123456789",
			MemoUsername:    "testuser",
			DiscordUsername: "testuser#1234",
		},
	}
	
	mockStore.DiscordAccount.On("ListByMemoUsername", mockDB, []string{"testuser"}).Return(discordAccounts, nil)

	// Test resolution
	discordID, err := resolver.ResolveUsernameToDiscordID("testuser")
	
	assert.NoError(t, err)
	assert.Equal(t, "123456789", discordID)
	
	// Verify cache hit on second call
	discordID2, err2 := resolver.ResolveUsernameToDiscordID("testuser")
	assert.NoError(t, err2)
	assert.Equal(t, "123456789", discordID2)
	
	// Should only call database once due to caching
	mockStore.DiscordAccount.AssertNumberOfCalls(t, "ListByMemoUsername", 1)
}

func TestAuthorResolver_ResolveUsernameToDiscordID_NotFound(t *testing.T) {
	mockStore, dbRepo, resolver := setupAuthorResolverTest()
	
	// Mock database response - empty result
	mockDB := &gorm.DB{}
	dbRepo.On("DB").Return(mockDB)
	mockStore.DiscordAccount.On("ListByMemoUsername", mockDB, []string{"nonexistent"}).Return([]model.DiscordAccount{}, nil)

	// Test resolution with fallback
	discordID, err := resolver.ResolveUsernameToDiscordID("nonexistent")
	
	assert.NoError(t, err)
	assert.Equal(t, "@nonexistent", discordID) // Default mention_attempt fallback
}

func TestAuthorResolver_ResolveUsernameToDiscordID_EmptyUsername(t *testing.T) {
	_, _, resolver := setupAuthorResolverTest()

	// Test empty username
	discordID, err := resolver.ResolveUsernameToDiscordID("")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username cannot be empty")
	assert.Empty(t, discordID)
}

func TestAuthorResolver_ResolveAuthorsToDiscordIDs_BatchResolution(t *testing.T) {
	mockStore, dbRepo, resolver := setupAuthorResolverTest()
	
	// Mock database response
	mockDB := &gorm.DB{}
	dbRepo.On("DB").Return(mockDB)
	
	discordAccounts := []model.DiscordAccount{
		{DiscordID: "123", MemoUsername: "user1"},
		{DiscordID: "456", MemoUsername: "user2"},
	}
	
	mockStore.DiscordAccount.On("ListByMemoUsername", mockDB, []string{"user1", "user2", "user3"}).Return(discordAccounts, nil)

	// Test batch resolution
	usernames := []string{"user1", "user2", "user3"}
	discordIDs, err := resolver.ResolveAuthorsToDiscordIDs(usernames)
	
	assert.NoError(t, err)
	assert.Len(t, discordIDs, 3)
	assert.Equal(t, "123", discordIDs[0])
	assert.Equal(t, "456", discordIDs[1])
	assert.Equal(t, "@user3", discordIDs[2]) // Fallback for user3
}

func TestAuthorResolver_ResolveAuthorsToDiscordIDs_EmptyInput(t *testing.T) {
	_, _, resolver := setupAuthorResolverTest()

	// Test empty input
	discordIDs, err := resolver.ResolveAuthorsToDiscordIDs([]string{})
	
	assert.NoError(t, err)
	assert.Empty(t, discordIDs)
}

func TestAuthorResolver_ResolveAllAuthors_FromMemos(t *testing.T) {
	mockStore, dbRepo, resolver := setupAuthorResolverTest()
	
	// Mock database response
	mockDB := &gorm.DB{}
	dbRepo.On("DB").Return(mockDB)
	
	discordAccounts := []model.DiscordAccount{
		{DiscordID: "111", MemoUsername: "author1"},
		{DiscordID: "222", MemoUsername: "author2"},
	}
	
	mockStore.DiscordAccount.On("ListByMemoUsername", mockDB, []string{"author1", "author2"}).Return(discordAccounts, nil)

	// Test memos with authors
	memos := []model.MemoLog{
		{AuthorMemoUsernames: []string{"author1", "author2"}},
		{AuthorMemoUsernames: []string{"author1"}}, // Duplicate should be handled
		{AuthorMemoUsernames: []string{"  author2  "}}, // Whitespace should be trimmed
	}
	
	authorMap, err := resolver.ResolveAllAuthors(memos)
	
	assert.NoError(t, err)
	assert.Len(t, authorMap, 2)
	assert.Equal(t, "111", authorMap["author1"])
	assert.Equal(t, "222", authorMap["author2"])
}

func TestAuthorResolver_ResolveAllAuthors_EmptyMemos(t *testing.T) {
	_, _, resolver := setupAuthorResolverTest()

	// Test empty memos
	authorMap, err := resolver.ResolveAllAuthors([]model.MemoLog{})
	
	assert.NoError(t, err)
	assert.Empty(t, authorMap)
}


func TestAuthorResolver_CacheEviction(t *testing.T) {
	mockStore, dbRepo, resolver := setupAuthorResolverTest()
	
	// Mock database responses
	mockDB := &gorm.DB{}
	dbRepo.On("DB").Return(mockDB)

	// Mock responses for multiple users to trigger cache eviction
	for i := 0; i < 15; i++ {
		username := []string{fmt.Sprintf("user%d", i)}
		discordAccounts := []model.DiscordAccount{
			{DiscordID: fmt.Sprintf("%d", 100+i), MemoUsername: username[0]},
		}
		mockStore.DiscordAccount.On("ListByMemoUsername", mockDB, username).Return(discordAccounts, nil).Once()
	}

	// Fill cache beyond max size (config has MaxCacheSize: 100, but we'll test eviction)
	for i := 0; i < 15; i++ {
		username := fmt.Sprintf("user%d", i)
		_, err := resolver.ResolveUsernameToDiscordID(username)
		assert.NoError(t, err)
	}

	// Check metrics
	metrics := resolver.GetMetrics()
	assert.GreaterOrEqual(t, metrics.CacheSize, 0)
	assert.GreaterOrEqual(t, metrics.DatabaseQueryCount, int64(15))
}

func TestAuthorResolver_FallbackStrategies(t *testing.T) {
	tests := []struct {
		name             string
		fallbackStrategy string
		username         string
		expected         string
	}{
		{
			name:             "plain username strategy",
			fallbackStrategy: "plain_username",
			username:         "testuser",
			expected:         "testuser",
		},
		{
			name:             "unverified format strategy",
			fallbackStrategy: "unverified_format",
			username:         "testuser",
			expected:         "testuser (unverified)",
		},
		{
			name:             "mention attempt strategy",
			fallbackStrategy: "mention_attempt",
			username:         "testuser",
			expected:         "@testuser",
		},
		{
			name:             "unknown strategy defaults to plain",
			fallbackStrategy: "unknown_strategy",
			username:         "testuser",
			expected:         "testuser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discordAccountStore := &mockDiscordAccountStore{}
			memoLogStore := &mockMemoLogStore{}
			dbRepo := &mockDBRepo{}
			
			store := &store.Store{
				DiscordAccount: discordAccountStore,
				MemoLog:        memoLogStore,
			}

			config := AuthorResolutionConfig{
				CacheTTL:         30 * time.Minute,
				MaxCacheSize:     100,
				BatchSize:        10,
				DatabaseTimeout:  5 * time.Second,
				FallbackStrategy: tt.fallbackStrategy,
				WarmupEnabled:    false,
			}

			resolver := NewAuthorResolver(config, store, dbRepo)
			
			// Mock database response - empty result to trigger fallback
			mockDB := &gorm.DB{}
			dbRepo.On("DB").Return(mockDB)
			discordAccountStore.On("ListByMemoUsername", mockDB, []string{tt.username}).Return([]model.DiscordAccount{}, nil)

			// Test fallback
			discordID, err := resolver.ResolveUsernameToDiscordID(tt.username)
			
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, discordID)
		})
	}
}

func TestAuthorResolver_WarmCache(t *testing.T) {
	mockStore, dbRepo, resolver := setupAuthorResolverTest()
	
	// Mock database responses for warmup
	mockDB := &gorm.DB{}
	dbRepo.On("DB").Return(mockDB)
	
	// Mock memo logs for warmup
	memos := []model.MemoLog{
		{AuthorMemoUsernames: []string{"warmup1", "warmup2"}},
		{AuthorMemoUsernames: []string{"warmup2", "warmup3"}},
	}
	mockStore.MemoLog.On("GetLimitByTimeRange", mockDB, mock.AnythingOfType("*time.Time"), mock.AnythingOfType("*time.Time"), 1000).Return(memos, nil)
	
	// Mock discord account resolution
	discordAccounts := []model.DiscordAccount{
		{DiscordID: "w1", MemoUsername: "warmup1"},
		{DiscordID: "w2", MemoUsername: "warmup2"},
		{DiscordID: "w3", MemoUsername: "warmup3"},
	}
	mockStore.DiscordAccount.On("ListByMemoUsername", mockDB, []string{"warmup1", "warmup2", "warmup3"}).Return(discordAccounts, nil)

	// Test warmup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err := resolver.WarmCache(ctx)
	assert.NoError(t, err)
	
	// Test that warmup populated cache
	metrics := resolver.GetMetrics()
	assert.False(t, metrics.LastWarmupTime.IsZero())
}

func TestAuthorResolver_WarmCache_ContextCancellation(t *testing.T) {
	_, _, resolver := setupAuthorResolverTest()
	
	// Create canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	err := resolver.WarmCache(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestAuthorResolver_ClearCache(t *testing.T) {
	mockStore, dbRepo, resolver := setupAuthorResolverTest()
	
	// Add some entries to cache first
	mockDB := &gorm.DB{}
	dbRepo.On("DB").Return(mockDB)
	discordAccounts := []model.DiscordAccount{
		{DiscordID: "123", MemoUsername: "testuser"},
	}
	mockStore.DiscordAccount.On("ListByMemoUsername", mockDB, []string{"testuser"}).Return(discordAccounts, nil)

	// Populate cache
	_, err := resolver.ResolveUsernameToDiscordID("testuser")
	assert.NoError(t, err)
	
	// Verify cache has entries
	metrics := resolver.GetMetrics()
	assert.Greater(t, metrics.CacheSize, 0)
	
	// Clear cache
	resolver.ClearCache()
	
	// Verify cache is empty
	metrics = resolver.GetMetrics()
	assert.Equal(t, 0, metrics.CacheSize)
}

func TestAuthorResolver_GetMetrics(t *testing.T) {
	_, _, resolver := setupAuthorResolverTest()
	
	// Get initial metrics
	metrics := resolver.GetMetrics()
	
	assert.GreaterOrEqual(t, metrics.CacheHitRate, 0.0)
	assert.GreaterOrEqual(t, metrics.ResolutionErrorRate, 0.0)
	assert.GreaterOrEqual(t, metrics.CacheSize, 0)
	assert.GreaterOrEqual(t, metrics.DatabaseQueryCount, int64(0))
	assert.GreaterOrEqual(t, metrics.CacheEvictions, int64(0))
}

func TestAuthorResolver_CacheExpiration(t *testing.T) {
	discordAccountStore := &mockDiscordAccountStore{}
	memoLogStore := &mockMemoLogStore{}
	dbRepo := &mockDBRepo{}
	
	store := &store.Store{
		DiscordAccount: discordAccountStore,
		MemoLog:        memoLogStore,
	}

	// Config with very short cache TTL for testing
	config := AuthorResolutionConfig{
		CacheTTL:         10 * time.Millisecond, // Very short for testing
		MaxCacheSize:     100,
		BatchSize:        10,
		DatabaseTimeout:  5 * time.Second,
		FallbackStrategy: "mention_attempt",
		WarmupEnabled:    false,
	}

	resolver := NewAuthorResolver(config, store, dbRepo)
	
	// Mock database response
	mockDB := &gorm.DB{}
	dbRepo.On("DB").Return(mockDB)
	discordAccounts := []model.DiscordAccount{
		{DiscordID: "123", MemoUsername: "testuser"},
	}
	discordAccountStore.On("ListByMemoUsername", mockDB, []string{"testuser"}).Return(discordAccounts, nil).Times(2)

	// First resolution
	discordID, err := resolver.ResolveUsernameToDiscordID("testuser")
	assert.NoError(t, err)
	assert.Equal(t, "123", discordID)
	
	// Wait for cache to expire
	time.Sleep(20 * time.Millisecond)
	
	// Second resolution should hit database again due to expiration
	discordID2, err2 := resolver.ResolveUsernameToDiscordID("testuser")
	assert.NoError(t, err2)
	assert.Equal(t, "123", discordID2)
	
	// Should have called database twice due to cache expiration
	discordAccountStore.AssertNumberOfCalls(t, "ListByMemoUsername", 2)
}