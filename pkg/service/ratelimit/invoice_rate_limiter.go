package ratelimit

import (
	"fmt"
	"sync"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// MaxInvoiceGenerationsPerDay is the maximum number of invoice generations allowed per day per user
const MaxInvoiceGenerationsPerDay = 3

// IInvoiceRateLimiter defines the interface for invoice rate limiting
type IInvoiceRateLimiter interface {
	// CheckLimit checks if the user has exceeded their daily rate limit
	// Returns error if limit exceeded, nil if request is allowed
	CheckLimit(discordUsername string) error

	// GetRemainingAttempts returns the number of remaining attempts for today
	GetRemainingAttempts(discordUsername string) int

	// GetResetTime returns when the limit resets for the user
	GetResetTime(discordUsername string) time.Time

	// Stop gracefully stops the rate limiter cleanup goroutine
	Stop()
}

// userLimit tracks rate limit data for a single user
type userLimit struct {
	Count   int
	ResetAt time.Time
}

// invoiceRateLimiter implements in-memory rate limiting for invoice generation
type invoiceRateLimiter struct {
	mu       sync.RWMutex
	counters map[string]*userLimit
	maxDaily int
	logger   logger.Logger
	stopCh   chan struct{}
}

// NewInvoiceRateLimiter creates a new rate limiter for invoice generation
func NewInvoiceRateLimiter(l logger.Logger) IInvoiceRateLimiter {
	rl := &invoiceRateLimiter{
		counters: make(map[string]*userLimit),
		maxDaily: MaxInvoiceGenerationsPerDay,
		logger:   l,
		stopCh:   make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	l.Debugf("[DEBUG] invoice_rate_limiter: initialized with maxDaily=%d", MaxInvoiceGenerationsPerDay)

	return rl
}

// CheckLimit checks if the user has exceeded their daily rate limit
func (rl *invoiceRateLimiter) CheckLimit(discordUsername string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limit, exists := rl.counters[discordUsername]

	rl.logger.Debug(fmt.Sprintf("[DEBUG] invoice_rate_limiter: CheckLimit username=%s exists=%v", discordUsername, exists))

	// First request or expired limit
	if !exists || now.After(limit.ResetAt) {
		rl.counters[discordUsername] = &userLimit{
			Count:   1,
			ResetAt: getNextMidnightUTC(now),
		}
		rl.logger.Debug(fmt.Sprintf("[DEBUG] invoice_rate_limiter: new/reset limit for username=%s count=1 resetAt=%s",
			discordUsername, rl.counters[discordUsername].ResetAt.Format(time.RFC3339)))
		return nil
	}

	// Check if limit exceeded
	if limit.Count >= rl.maxDaily {
		rl.logger.Debug(fmt.Sprintf("[DEBUG] invoice_rate_limiter: rate limit exceeded username=%s count=%d max=%d resetAt=%s",
			discordUsername, limit.Count, rl.maxDaily, limit.ResetAt.Format(time.RFC3339)))
		return fmt.Errorf("rate limit exceeded: %d/%d requests today, resets at %s",
			limit.Count, rl.maxDaily, limit.ResetAt.Format("15:04 MST"))
	}

	// Increment counter
	limit.Count++
	rl.logger.Debug(fmt.Sprintf("[DEBUG] invoice_rate_limiter: incremented count username=%s count=%d",
		discordUsername, limit.Count))
	return nil
}

// GetRemainingAttempts returns the number of remaining attempts for today
func (rl *invoiceRateLimiter) GetRemainingAttempts(discordUsername string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	limit, exists := rl.counters[discordUsername]
	if !exists || time.Now().After(limit.ResetAt) {
		return rl.maxDaily
	}

	remaining := rl.maxDaily - limit.Count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetResetTime returns when the limit resets for the user
func (rl *invoiceRateLimiter) GetResetTime(discordUsername string) time.Time {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	limit, exists := rl.counters[discordUsername]
	if !exists {
		return getNextMidnightUTC(time.Now())
	}
	return limit.ResetAt
}

// Stop gracefully stops the rate limiter cleanup goroutine
func (rl *invoiceRateLimiter) Stop() {
	close(rl.stopCh)
	rl.logger.Debug("[DEBUG] invoice_rate_limiter: stopped")
}

// cleanupLoop runs hourly to remove expired entries
func (rl *invoiceRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCh:
			return
		}
	}
}

// cleanup removes entries older than 24 hours past their reset time
func (rl *invoiceRateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	removedCount := 0
	for username, limit := range rl.counters {
		// Remove entries older than 24 hours past reset
		if now.After(limit.ResetAt.Add(24 * time.Hour)) {
			delete(rl.counters, username)
			removedCount++
		}
	}

	if removedCount > 0 {
		rl.logger.Debug(fmt.Sprintf("[DEBUG] invoice_rate_limiter: cleanup removed %d expired entries", removedCount))
	}
}

// getNextMidnightUTC returns the next midnight in UTC
func getNextMidnightUTC(now time.Time) time.Time {
	utcNow := now.UTC()
	return time.Date(utcNow.Year(), utcNow.Month(), utcNow.Day()+1, 0, 0, 0, 0, time.UTC)
}
