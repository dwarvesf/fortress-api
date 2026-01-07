package ratelimit

import (
	"sync"
	"testing"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/logger"
)

func TestCheckLimit_FirstRequest(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	rl := NewInvoiceRateLimiter(l)
	defer rl.Stop()

	err := rl.CheckLimit("test_user")
	if err != nil {
		t.Errorf("first request should be allowed, got error: %v", err)
	}
}

func TestCheckLimit_ThirdRequest(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	rl := NewInvoiceRateLimiter(l)
	defer rl.Stop()

	username := "test_user_3rd"

	// First 3 requests should be allowed
	for i := 1; i <= 3; i++ {
		err := rl.CheckLimit(username)
		if err != nil {
			t.Errorf("request %d should be allowed, got error: %v", i, err)
		}
	}
}

func TestCheckLimit_FourthRequest(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	rl := NewInvoiceRateLimiter(l)
	defer rl.Stop()

	username := "test_user_4th"

	// First 3 requests should be allowed
	for i := 1; i <= 3; i++ {
		err := rl.CheckLimit(username)
		if err != nil {
			t.Errorf("request %d should be allowed, got error: %v", i, err)
		}
	}

	// Fourth request should be rejected
	err := rl.CheckLimit(username)
	if err == nil {
		t.Error("fourth request should be rejected")
	}
}

func TestCheckLimit_DifferentUsers(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	rl := NewInvoiceRateLimiter(l)
	defer rl.Stop()

	// Each user should have their own limit
	for i := 0; i < 5; i++ {
		err := rl.CheckLimit("user_a")
		if i < 3 && err != nil {
			t.Errorf("user_a request %d should be allowed, got error: %v", i+1, err)
		}
		if i >= 3 && err == nil {
			t.Errorf("user_a request %d should be rejected", i+1)
		}
	}

	// user_b should have their own limit, completely independent
	for i := 0; i < 3; i++ {
		err := rl.CheckLimit("user_b")
		if err != nil {
			t.Errorf("user_b request %d should be allowed, got error: %v", i+1, err)
		}
	}
}

func TestCheckLimit_Concurrent(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	rl := NewInvoiceRateLimiter(l)
	defer rl.Stop()

	username := "concurrent_user"
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// Run 100 concurrent requests
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := rl.CheckLimit(username)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Only 3 should succeed (the rate limit)
	if successCount != MaxInvoiceGenerationsPerDay {
		t.Errorf("expected %d successful requests, got %d", MaxInvoiceGenerationsPerDay, successCount)
	}
}

func TestGetRemainingAttempts(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	rl := NewInvoiceRateLimiter(l)
	defer rl.Stop()

	username := "remaining_test_user"

	// Before any requests
	remaining := rl.GetRemainingAttempts(username)
	if remaining != MaxInvoiceGenerationsPerDay {
		t.Errorf("expected %d remaining attempts before any requests, got %d", MaxInvoiceGenerationsPerDay, remaining)
	}

	// After one request
	_ = rl.CheckLimit(username)
	remaining = rl.GetRemainingAttempts(username)
	if remaining != MaxInvoiceGenerationsPerDay-1 {
		t.Errorf("expected %d remaining attempts after 1 request, got %d", MaxInvoiceGenerationsPerDay-1, remaining)
	}

	// After two more requests
	_ = rl.CheckLimit(username)
	_ = rl.CheckLimit(username)
	remaining = rl.GetRemainingAttempts(username)
	if remaining != 0 {
		t.Errorf("expected 0 remaining attempts after 3 requests, got %d", remaining)
	}
}

func TestGetResetTime(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	rl := NewInvoiceRateLimiter(l)
	defer rl.Stop()

	username := "reset_time_user"

	// Before any requests - should return next midnight UTC
	resetTime := rl.GetResetTime(username)
	expectedReset := getNextMidnightUTC(time.Now())
	if !resetTime.Equal(expectedReset) {
		t.Errorf("reset time before requests should be next midnight UTC, got %v, expected %v", resetTime, expectedReset)
	}

	// After a request
	rl.CheckLimit(username)
	resetTime = rl.GetResetTime(username)
	if resetTime.Before(time.Now()) {
		t.Error("reset time should be in the future")
	}
}

func TestCleanup(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	rl := &invoiceRateLimiter{
		counters: make(map[string]*userLimit),
		maxDaily: MaxInvoiceGenerationsPerDay,
		logger:   l,
		stopCh:   make(chan struct{}),
	}

	// Add an expired entry (26 hours ago)
	expiredTime := time.Now().Add(-26 * time.Hour)
	rl.counters["expired_user"] = &userLimit{
		Count:   1,
		ResetAt: expiredTime,
	}

	// Add a valid entry
	rl.counters["valid_user"] = &userLimit{
		Count:   1,
		ResetAt: getNextMidnightUTC(time.Now()),
	}

	// Run cleanup
	rl.cleanup()

	// Expired user should be removed
	if _, exists := rl.counters["expired_user"]; exists {
		t.Error("expired user should have been removed")
	}

	// Valid user should remain
	if _, exists := rl.counters["valid_user"]; !exists {
		t.Error("valid user should not have been removed")
	}
}

func TestGetNextMidnightUTC(t *testing.T) {
	// Test with a known time
	now := time.Date(2026, 1, 7, 10, 30, 0, 0, time.UTC)
	midnight := getNextMidnightUTC(now)

	expected := time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC)
	if !midnight.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, midnight)
	}

	// Test with a time near midnight
	nearMidnight := time.Date(2026, 1, 7, 23, 59, 0, 0, time.UTC)
	midnightNear := getNextMidnightUTC(nearMidnight)

	expectedNear := time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC)
	if !midnightNear.Equal(expectedNear) {
		t.Errorf("expected %v, got %v", expectedNear, midnightNear)
	}
}
