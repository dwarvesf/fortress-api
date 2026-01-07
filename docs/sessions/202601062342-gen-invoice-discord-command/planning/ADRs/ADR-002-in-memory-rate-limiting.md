# ADR-002: In-Memory Rate Limiting for Invoice Generation

## Status
Proposed

## Context
The `?gen invoice` command triggers resource-intensive operations:
- Notion API calls to fetch contractor data
- Invoice PDF generation
- Google Drive API uploads
- File sharing operations
- Discord API message updates

Without rate limiting, a user could:
- Spam the command and overload external APIs
- Generate duplicate invoices accidentally
- Exhaust API quotas (Google Drive, Notion)
- Create excessive load on the system

We need to limit users to **3 invoice generations per day** to prevent abuse while allowing legitimate use cases (regeneration due to errors, month changes, etc.).

### Constraints
- Requirement: No database migrations allowed
- Must be thread-safe for concurrent requests
- Should reset daily
- Performance: Sub-millisecond lookup time
- Memory footprint should be minimal (hundreds of users max)

### Options Considered

#### Option 1: Database-Backed Rate Limiting (Rejected)
Store rate limit counters in PostgreSQL with daily reset logic.

**Pros:**
- Persistent across server restarts
- Centralized state for multi-instance deployments
- Can query rate limit history
- Accurate daily reset with scheduled jobs

**Cons:**
- **Violates requirement**: No database migrations allowed
- Database query overhead on every request
- Requires migration for new table
- Adds complexity (schema, queries, indexes)
- Overkill for simple rate limiting

#### Option 2: Redis-Backed Rate Limiting (Rejected)
Use Redis with TTL for rate limit counters.

**Pros:**
- Fast in-memory performance
- Persistent across server restarts
- Native TTL support for daily reset
- Scales to multiple instances

**Cons:**
- Requires new infrastructure dependency
- Additional operational overhead
- Not available in current stack
- Overkill for current scale

#### Option 3: In-Memory Map with Mutex (Selected)
Use Go map with sync.RWMutex for thread-safe access.

**Pros:**
- **Meets requirement**: No database changes needed
- Extremely fast (nanosecond lookups)
- Simple implementation
- Zero infrastructure dependencies
- Sufficient for single-instance deployment

**Cons:**
- Lost on server restart (acceptable trade-off)
- Not shared across multiple instances (current deployment is single-instance)
- No historical data
- Manual daily reset logic needed

## Decision
We will use **Option 3: In-Memory Map with Mutex**.

### Data Structure
```go
type RateLimiter struct {
    mu       sync.RWMutex
    counters map[string]*UserLimit
    maxDaily int
}

type UserLimit struct {
    Count     int
    ResetAt   time.Time
}
```

### Implementation Pattern

#### Rate Limit Check
```go
func (rl *RateLimiter) CheckLimit(discordUsername string) error {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    now := time.Now()
    limit, exists := rl.counters[discordUsername]

    // First request or expired limit
    if !exists || now.After(limit.ResetAt) {
        rl.counters[discordUsername] = &UserLimit{
            Count:   1,
            ResetAt: getNextMidnight(now),
        }
        return nil
    }

    // Check if limit exceeded
    if limit.Count >= rl.maxDaily {
        return fmt.Errorf("rate limit exceeded: %d/%d requests today, resets at %s",
            limit.Count, rl.maxDaily, limit.ResetAt.Format("15:04"))
    }

    // Increment counter
    limit.Count++
    return nil
}
```

#### Daily Reset Logic
```go
func getNextMidnight(now time.Time) time.Time {
    // Reset at midnight in server timezone (UTC or configured TZ)
    return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
}
```

#### Cleanup (Optional)
```go
func (rl *RateLimiter) CleanupExpired() {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    now := time.Now()
    for username, limit := range rl.counters {
        if now.After(limit.ResetAt.Add(24 * time.Hour)) {
            delete(rl.counters, username)
        }
    }
}

// Run cleanup every hour
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    for range ticker.C {
        rateLimiter.CleanupExpired()
    }
}()
```

### Configuration
```go
const (
    MaxInvoiceGenerationsPerDay = 3
)

// Initialize at startup
var rateLimiter = &RateLimiter{
    counters: make(map[string]*UserLimit),
    maxDaily: MaxInvoiceGenerationsPerDay,
}
```

### Error Response
When rate limit is exceeded, return HTTP 429 with body:
```json
{
  "error": "Rate limit exceeded",
  "message": "You have generated 3 invoices today. Limit resets at midnight UTC.",
  "limit": 3,
  "reset_at": "2025-01-07T00:00:00Z"
}
```

## Consequences

### Positive
- Meets requirement: No database migrations
- Extremely fast: O(1) map lookup with minimal lock contention
- Simple implementation: ~50 lines of code
- Zero infrastructure dependencies
- Easy to test and reason about
- Sufficient for current scale (single instance, hundreds of users)

### Negative
- Lost on server restart (users get "free" resets)
- Not shared across multiple instances (if we scale horizontally)
- No historical rate limit data for analysis
- Memory grows with unique users (mitigated by cleanup)

### Acceptable Trade-offs

#### Server Restart = Reset
This is acceptable because:
- Server restarts are infrequent (deployments, crashes)
- Impact is limited to current day's counters
- Worst case: User gets extra attempts on restart day
- Not a security concern (just a courtesy limit)

#### Single Instance Limitation
This is acceptable because:
- Current deployment is single-instance
- If we scale to multiple instances, we can:
  - Upgrade to Redis (infrastructure exists by then)
  - Use sticky sessions (route same user to same instance)
  - Accept that limits are per-instance (3x per instance = 6x total with 2 instances)

#### No Historical Data
This is acceptable because:
- Rate limiting is protective, not analytical
- We can add logging if usage analysis is needed later
- Logs would show all generation attempts anyway

### Testing Strategy
```go
func TestRateLimiter(t *testing.T) {
    // Test cases:
    // 1. First request succeeds
    // 2. Third request succeeds
    // 3. Fourth request fails (rate limited)
    // 4. After reset time, request succeeds
    // 5. Concurrent requests are thread-safe
}
```

### Migration Path (Future)
If we need persistent rate limiting later:
1. Add Redis to infrastructure
2. Replace in-memory map with Redis client
3. Use Redis INCR + EXPIRE for counters
4. Zero code changes to webhook handler (same interface)

## References
- Go sync.RWMutex: https://pkg.go.dev/sync#RWMutex
- HTTP 429 Status: https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/429
- Rate limiting patterns: https://blog.logrocket.com/rate-limiting-go-application/
