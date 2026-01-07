# Specification: fortress-api Changes for Discord Invoice Generation

## Overview
Implement a webhook endpoint to handle async invoice generation requests from fortress-discord, with in-memory rate limiting and Google Drive file sharing.

## Architecture Components

### 1. Rate Limiter (New Component)

#### Location
`pkg/service/ratelimit/invoice_rate_limiter.go`

#### Interface

```go
package ratelimit

import (
    "fmt"
    "sync"
    "time"
)

type InvoiceRateLimiter interface {
    // CheckLimit returns error if user exceeded rate limit
    CheckLimit(discordUsername string) error

    // GetRemainingAttempts returns remaining attempts for today
    GetRemainingAttempts(discordUsername string) int

    // GetResetTime returns when the limit resets for user
    GetResetTime(discordUsername string) time.Time
}
```

#### Implementation

```go
const MaxInvoiceGenerationsPerDay = 3

type rateLimiter struct {
    mu       sync.RWMutex
    counters map[string]*userLimit
    maxDaily int
}

type userLimit struct {
    Count   int
    ResetAt time.Time
}

func NewInvoiceRateLimiter() InvoiceRateLimiter {
    rl := &rateLimiter{
        counters: make(map[string]*userLimit),
        maxDaily: MaxInvoiceGenerationsPerDay,
    }

    // Start cleanup goroutine
    go rl.cleanupLoop()

    return rl
}

func (rl *rateLimiter) CheckLimit(discordUsername string) error {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    now := time.Now()
    limit, exists := rl.counters[discordUsername]

    // First request or expired limit
    if !exists || now.After(limit.ResetAt) {
        rl.counters[discordUsername] = &userLimit{
            Count:   1,
            ResetAt: getNextMidnight(now),
        }
        return nil
    }

    // Check if limit exceeded
    if limit.Count >= rl.maxDaily {
        return fmt.Errorf("rate limit exceeded: %d/%d requests today, resets at %s",
            limit.Count, rl.maxDaily, limit.ResetAt.Format("15:04 MST"))
    }

    // Increment counter
    limit.Count++
    return nil
}

func (rl *rateLimiter) GetRemainingAttempts(discordUsername string) int {
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

func (rl *rateLimiter) GetResetTime(discordUsername string) time.Time {
    rl.mu.RLock()
    defer rl.mu.RUnlock()

    limit, exists := rl.counters[discordUsername]
    if !exists {
        return getNextMidnight(time.Now())
    }
    return limit.ResetAt
}

func (rl *rateLimiter) cleanupLoop() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()

    for range ticker.C {
        rl.cleanup()
    }
}

func (rl *rateLimiter) cleanup() {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    now := time.Now()
    for username, limit := range rl.counters {
        // Remove entries older than 24 hours past reset
        if now.After(limit.ResetAt.Add(24 * time.Hour)) {
            delete(rl.counters, username)
        }
    }
}

func getNextMidnight(now time.Time) time.Time {
    // Reset at midnight UTC
    return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
}
```

#### Unit Tests
`pkg/service/ratelimit/invoice_rate_limiter_test.go`

```go
func TestCheckLimit_FirstRequest(t *testing.T) {
    // Should allow first request
}

func TestCheckLimit_ThirdRequest(t *testing.T) {
    // Should allow third request
}

func TestCheckLimit_FourthRequest(t *testing.T) {
    // Should reject fourth request
}

func TestCheckLimit_AfterReset(t *testing.T) {
    // Should allow request after reset time
}

func TestCheckLimit_Concurrent(t *testing.T) {
    // Should be thread-safe with concurrent requests
}

func TestGetRemainingAttempts(t *testing.T) {
    // Should return correct remaining count
}

func TestCleanup(t *testing.T) {
    // Should remove expired entries
}
```

### 2. Google Drive Service Enhancement

#### Location
`pkg/service/googledrive/google_drive.go`

#### Add Method to Existing Service

```go
// ShareFileWithEmail grants read access to a file and sends notification email
func (s *service) ShareFileWithEmail(fileID, email string) error {
    if fileID == "" {
        return fmt.Errorf("fileID is required")
    }
    if email == "" {
        return fmt.Errorf("email is required")
    }

    permission := &drive.Permission{
        Type:         "user",
        Role:         "reader",
        EmailAddress: email,
    }

    _, err := s.client.Permissions.Create(fileID, permission).
        SendNotificationEmail(true).
        EmailMessage("Your invoice has been generated and is ready for review.").
        Do()

    if err != nil {
        return fmt.Errorf("failed to share file: %w", err)
    }

    return nil
}
```

#### Update Interface
`pkg/service/googledrive/interface.go`

```go
type GoogleDriveService interface {
    // ... existing methods ...

    // ShareFileWithEmail shares a file with an email address
    ShareFileWithEmail(fileID, email string) error
}
```

#### Unit Tests
`pkg/service/googledrive/google_drive_test.go`

```go
func TestShareFileWithEmail_Success(t *testing.T) {
    // Test successful file sharing
}

func TestShareFileWithEmail_InvalidEmail(t *testing.T) {
    // Test error handling for invalid email
}

func TestShareFileWithEmail_EmptyFileID(t *testing.T) {
    // Test validation for empty fileID
}

func TestShareFileWithEmail_EmptyEmail(t *testing.T) {
    // Test validation for empty email
}
```

### 3. Webhook Handler

#### Location
`pkg/handler/webhook/gen_invoice.go`

#### Request/Response Models

```go
package webhook

import "time"

// GenInvoiceRequest represents Discord webhook payload
type GenInvoiceRequest struct {
    DiscordUsername string `json:"discord_username" binding:"required"`
    Month           string `json:"month" binding:"required"`
    DMChannelID     string `json:"dm_channel_id" binding:"required"`
    DMMessageID     string `json:"dm_message_id" binding:"required"`
}

// GenInvoiceResponse represents the response to Discord webhook
type GenInvoiceResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}

// Validate validates the request
func (r *GenInvoiceRequest) Validate() error {
    if r.DiscordUsername == "" {
        return fmt.Errorf("discord_username is required")
    }
    if r.Month == "" {
        return fmt.Errorf("month is required")
    }
    if r.DMChannelID == "" {
        return fmt.Errorf("dm_channel_id is required")
    }
    if r.DMMessageID == "" {
        return fmt.Errorf("dm_message_id is required")
    }

    // Validate month format
    _, err := time.Parse("2006-01", r.Month)
    if err != nil {
        return fmt.Errorf("invalid month format, expected YYYY-MM: %w", err)
    }

    return nil
}
```

#### Handler Interface

```go
package webhook

type IWebhookHandler interface {
    // ... existing methods ...

    // GenInvoice handles invoice generation webhook from Discord
    GenInvoice(c *gin.Context)
}
```

#### Handler Implementation

```go
type handler struct {
    controller       controller.IInvoiceController
    discordService   discord.IDiscordService
    notionService    notion.INotionService
    googleDrive      googledrive.GoogleDriveService
    rateLimiter      ratelimit.InvoiceRateLimiter
    logger           *logrus.Logger
}

func (h *handler) GenInvoice(c *gin.Context) {
    // 1. Parse and validate request
    var req GenInvoiceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, GenInvoiceResponse{
            Success: false,
            Message: fmt.Sprintf("Invalid request: %v", err),
        })
        return
    }

    if err := req.Validate(); err != nil {
        c.JSON(http.StatusBadRequest, GenInvoiceResponse{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    // 2. Check rate limit
    if err := h.rateLimiter.CheckLimit(req.DiscordUsername); err != nil {
        h.logger.WithFields(logrus.Fields{
            "username": req.DiscordUsername,
            "month":    req.Month,
        }).Warn("Rate limit exceeded")

        // Update Discord DM with rate limit error
        go h.updateDiscordWithError(req.DMChannelID, req.DMMessageID,
            "Rate Limit Exceeded",
            err.Error())

        c.JSON(http.StatusTooManyRequests, GenInvoiceResponse{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    // 3. Return 200 immediately (async processing starts)
    c.JSON(http.StatusOK, GenInvoiceResponse{
        Success: true,
        Message: "Invoice generation started",
    })

    // 4. Process async in goroutine
    go h.processInvoiceGeneration(req)
}

func (h *handler) processInvoiceGeneration(req GenInvoiceRequest) {
    ctx := context.Background()
    logger := h.logger.WithFields(logrus.Fields{
        "username": req.DiscordUsername,
        "month":    req.Month,
    })

    logger.Info("Starting invoice generation")

    // Parse month for controller
    date, err := time.Parse("2006-01", req.Month)
    if err != nil {
        logger.WithError(err).Error("Failed to parse month")
        h.updateDiscordWithError(req.DMChannelID, req.DMMessageID,
            "Invalid Month Format",
            "Failed to parse month. Please use YYYY-MM format.")
        return
    }

    // Generate invoice using existing controller
    invoiceReq := controller.GenerateContractorInvoiceRequest{
        DiscordUsername: req.DiscordUsername,
        Month:           int(date.Month()),
        Year:            date.Year(),
    }

    resp, err := h.controller.GenerateContractorInvoice(ctx, invoiceReq)
    if err != nil {
        logger.WithError(err).Error("Failed to generate invoice")
        h.updateDiscordWithError(req.DMChannelID, req.DMMessageID,
            "Invoice Generation Failed",
            fmt.Sprintf("Failed to generate invoice: %v", err))
        return
    }

    // Get contractor's personal email from Notion
    personalEmail, err := h.notionService.GetContractorPersonalEmail(req.DiscordUsername)
    if err != nil || personalEmail == "" {
        logger.WithError(err).Error("Failed to get contractor email")
        h.updateDiscordWithError(req.DMChannelID, req.DMMessageID,
            "Email Lookup Failed",
            "Failed to find your email address. Please contact HR.")
        return
    }

    // Share Google Drive file with email
    if resp.FileID != "" {
        err = h.googleDrive.ShareFileWithEmail(resp.FileID, personalEmail)
        if err != nil {
            logger.WithError(err).Error("Failed to share file")
            h.updateDiscordWithError(req.DMChannelID, req.DMMessageID,
                "File Sharing Failed",
                "Invoice generated but failed to share. Please contact support.")
            return
        }
    }

    // Update Discord DM with success
    h.updateDiscordWithSuccess(req.DMChannelID, req.DMMessageID, resp.FileURL, personalEmail, req.Month)

    logger.Info("Invoice generation completed successfully")
}

func (h *handler) updateDiscordWithSuccess(channelID, messageID, fileURL, email, month string) {
    embed := &model.Embed{
        Title:       "Invoice Generated Successfully",
        Description: fmt.Sprintf("Your invoice for %s has been generated and shared to your email.", formatMonthDisplay(month)),
        Color:       "2ecc71", // Green
        Fields: []model.EmbedField{
            {Name: "Month", Value: formatMonthDisplay(month), Inline: true},
            {Name: "Email", Value: email, Inline: true},
            {Name: "File", Value: fmt.Sprintf("[View in Google Drive](%s)", fileURL), Inline: false},
        },
        Footer: &model.EmbedFooter{
            Text: "Check your email for file access notification",
        },
    }

    err := h.discordService.UpdateChannelMessage(channelID, messageID, "", []*model.Embed{embed})
    if err != nil {
        h.logger.WithError(err).WithFields(logrus.Fields{
            "channelID": channelID,
            "messageID": messageID,
        }).Error("Failed to update Discord message with success")
    }
}

func (h *handler) updateDiscordWithError(channelID, messageID, title, message string) {
    embed := &model.Embed{
        Title:       title,
        Description: message,
        Color:       "e74c3c", // Red
        Footer: &model.EmbedFooter{
            Text: "Contact support if this persists",
        },
    }

    err := h.discordService.UpdateChannelMessage(channelID, messageID, "", []*model.Embed{embed})
    if err != nil {
        h.logger.WithError(err).WithFields(logrus.Fields{
            "channelID": channelID,
            "messageID": messageID,
        }).Error("Failed to update Discord message with error")
    }
}

func formatMonthDisplay(month string) string {
    t, err := time.Parse("2006-01", month)
    if err != nil {
        return month
    }
    return t.Format("January 2006")
}
```

#### Unit Tests
`pkg/handler/webhook/gen_invoice_test.go`

```go
func TestGenInvoice_Success(t *testing.T) {
    // Test successful invoice generation flow
}

func TestGenInvoice_InvalidRequest(t *testing.T) {
    // Test validation errors
}

func TestGenInvoice_RateLimited(t *testing.T) {
    // Test rate limit enforcement
}

func TestGenInvoice_GenerationFailed(t *testing.T) {
    // Test error handling when invoice generation fails
}

func TestGenInvoice_EmailLookupFailed(t *testing.T) {
    // Test error handling when email lookup fails
}

func TestGenInvoice_FileSharingFailed(t *testing.T) {
    // Test error handling when file sharing fails
}
```

### 4. Routes

#### Location
`pkg/routes/v1.go`

#### Add Webhook Route

```go
func RegisterWebhookRoutes(router *gin.RouterGroup, handler webhook.IWebhookHandler) {
    webhooks := router.Group("/webhooks")
    {
        // ... existing webhook routes ...

        // Discord invoice generation webhook
        webhooks.POST("/discord/gen-invoice", handler.GenInvoice)
    }
}
```

Or if webhooks are in a separate route file:

`pkg/routes/webhook.go`

```go
func (r *Router) registerWebhookRoutes() {
    webhook := r.router.Group("/webhooks")
    {
        // ... existing routes ...

        discord := webhook.Group("/discord")
        {
            discord.POST("/gen-invoice", r.handler.Webhook.GenInvoice)
        }
    }
}
```

## Controller Enhancement (Optional)

The existing `GenerateContractorInvoice` controller may need to return additional fields:

#### Location
`pkg/controller/invoice/contractor_invoice.go`

#### Response Model Enhancement

```go
type GenerateContractorInvoiceResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    FileID  string `json:"file_id"`   // Google Drive file ID (add if not present)
    FileURL string `json:"file_url"`  // Public or shareable URL
}
```

If the controller doesn't return `FileID`, add it to the invoice generation logic:

```go
// After uploading to Google Drive
fileID := uploadedFile.Id
fileURL := fmt.Sprintf("https://drive.google.com/file/d/%s/view", fileID)

return &GenerateContractorInvoiceResponse{
    Success: true,
    Message: "Invoice generated successfully",
    FileID:  fileID,
    FileURL: fileURL,
}, nil
```

## Initialization (Main Application)

#### Location
`cmd/server/main.go`

#### Add Rate Limiter to Dependencies

```go
// Initialize services
rateLimiter := ratelimit.NewInvoiceRateLimiter()

// Pass to handler
webhookHandler := webhook.NewHandler(
    invoiceController,
    discordService,
    notionService,
    googleDriveService,
    rateLimiter,
    logger,
)
```

## Configuration

No new configuration needed - uses existing:
- Discord service configuration
- Google Drive API credentials
- Notion API configuration

## Database Schema

No database changes needed (requirement constraint - in-memory rate limiting).

## Error Handling

### HTTP Status Codes

| Status | Scenario |
|--------|----------|
| 200 OK | Request accepted, processing started |
| 400 Bad Request | Invalid request payload or validation error |
| 429 Too Many Requests | Rate limit exceeded |
| 500 Internal Server Error | Unexpected error (should not happen after 200) |

### Discord Message Updates

All async errors are communicated via Discord DM message updates:

1. Rate limit exceeded
2. Invalid contractor status
3. Invoice generation failed
4. Email lookup failed
5. File sharing failed

## Logging

Log at key points:

```go
// Request received
logger.Info("Invoice generation webhook received")

// Rate limit check
logger.Warn("Rate limit exceeded")

// Processing started
logger.Info("Starting invoice generation")

// Invoice generated
logger.Info("Invoice generated successfully")

// File shared
logger.Info("File shared with contractor")

// Errors
logger.WithError(err).Error("Failed to generate invoice")
```

## Testing Strategy

### Unit Tests
- Rate limiter: Thread-safety, reset logic, cleanup
- Google Drive service: File sharing, error handling
- Webhook handler: Request validation, rate limiting, async processing
- Helper functions: Month formatting, date parsing

### Integration Tests
- End-to-end webhook flow (mock Discord/Google/Notion)
- Rate limit enforcement across multiple requests
- Error scenarios with Discord message updates

### Manual Testing
1. Test with valid request from fortress-discord
2. Test rate limiting (4th request should fail)
3. Test with invalid contractor
4. Test with invalid month format
5. Test DM message updates

## Files to Create

1. `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/ratelimit/invoice_rate_limiter.go`
2. `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/ratelimit/invoice_rate_limiter_test.go`
3. `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/webhook/gen_invoice.go`
4. `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/webhook/gen_invoice_test.go`

## Files to Modify

1. `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/googledrive/google_drive.go`
   - Add `ShareFileWithEmail` method

2. `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/googledrive/interface.go`
   - Add method to interface

3. `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/googledrive/google_drive_test.go`
   - Add tests for new method

4. `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/webhook/interface.go`
   - Add `GenInvoice` method

5. `/Users/quang/workspace/dwarvesf/fortress-api/pkg/routes/v1.go` (or webhook routes file)
   - Add webhook route

6. `/Users/quang/workspace/dwarvesf/fortress-api/cmd/server/main.go`
   - Initialize rate limiter
   - Pass to webhook handler

7. `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go` (if needed)
   - Ensure response includes FileID and FileURL

## Deployment Considerations

### Environment Variables
No new environment variables needed - uses existing Google Drive and Discord credentials.

### Rollout Strategy
1. Deploy fortress-api changes first
2. Test webhook endpoint manually (curl or Postman)
3. Deploy fortress-discord
4. Test end-to-end flow
5. Monitor logs for errors

### Monitoring
- Rate limiter memory usage
- Webhook response times
- Error rates for async processing
- Google Drive API quota usage

### Rollback Plan
- Remove webhook route
- Redeploy previous version
- No database cleanup needed
