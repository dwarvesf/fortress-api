# Implementation Tasks: Discord `?gen invoice` Command

## Session Information
**Session ID**: 202601062342-gen-invoice-discord-command
**Feature**: Discord command for contractor invoice generation via async webhook processing
**Date**: 2026-01-06

## Overview
This document provides a detailed, ordered task breakdown for implementing the `?gen invoice` Discord command. Tasks are organized by repository and include clear dependencies, complexity estimates, and implementation order.

## Implementation Strategy

### Phase 1: Foundation (fortress-api)
Build the webhook endpoint that receives requests from Discord bot

### Phase 2: Integration (fortress-discord)
Build the Discord command that sends requests to the webhook

### Phase 3: Testing & Deployment
End-to-end testing and deployment

---

## Repository 1: fortress-api

### Group A: Core Services (No Dependencies)

#### Task 1.1: Implement In-Memory Rate Limiter
**Description**: Create thread-safe in-memory rate limiter service for invoice generation requests

**Complexity**: Medium

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/ratelimit/invoice_rate_limiter.go`
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/ratelimit/invoice_rate_limiter_test.go`

**Implementation Details**:
- Implement `InvoiceRateLimiter` interface with `CheckLimit`, `GetRemainingAttempts`, `GetResetTime` methods
- Use `sync.RWMutex` for thread-safe access to counters map
- Implement daily reset at midnight UTC
- Add cleanup goroutine (hourly) to remove expired entries
- Maximum 3 generations per day per user
- Return descriptive errors with reset time information

**Unit Tests Required**:
- `TestCheckLimit_FirstRequest` - Allow first request
- `TestCheckLimit_ThirdRequest` - Allow third request
- `TestCheckLimit_FourthRequest` - Reject fourth request with error
- `TestCheckLimit_AfterReset` - Allow request after midnight reset
- `TestCheckLimit_Concurrent` - Thread-safety with 100+ concurrent requests
- `TestGetRemainingAttempts` - Correct remaining count calculation
- `TestCleanup` - Remove expired entries after 24 hours

**Dependencies**: None

**Acceptance Criteria**:
- Rate limiter correctly tracks requests per user
- Mutex prevents race conditions
- Cleanup goroutine prevents memory leaks
- All unit tests pass

---

#### Task 1.2: Add Google Drive File Sharing Method
**Description**: Extend Google Drive service to share files with email addresses

**Complexity**: Small

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/googledrive/google_drive.go`
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/googledrive/interface.go`
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/googledrive/google_drive_test.go`

**Implementation Details**:
- Add `ShareFileWithEmail(fileID, email string) error` method to interface
- Implement method using Drive API `Permissions.Create` with:
  - Type: "user"
  - Role: "reader"
  - EmailAddress: provided email
  - SendNotificationEmail: true
  - EmailMessage: "Your invoice has been generated and is ready for review."
- Validate fileID and email are non-empty
- Return descriptive errors

**Unit Tests Required**:
- `TestShareFileWithEmail_Success` - Successful file sharing (mocked Drive API)
- `TestShareFileWithEmail_EmptyFileID` - Validation error for empty fileID
- `TestShareFileWithEmail_EmptyEmail` - Validation error for empty email
- `TestShareFileWithEmail_DriveAPIError` - Handle Drive API errors gracefully

**Dependencies**: None

**Acceptance Criteria**:
- Method successfully calls Drive API with correct permissions
- Validation prevents empty parameters
- Error handling provides clear messages
- All unit tests pass

---

### Group B: Webhook Handler (Depends on Group A)

#### Task 1.3: Create Webhook Request/Response Models
**Description**: Define data structures for webhook communication with Discord bot

**Complexity**: Small

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/webhook/gen_invoice.go` (models section)

**Implementation Details**:
- Create `GenInvoiceRequest` struct with:
  - `DiscordUsername` (required)
  - `Month` in YYYY-MM format (required)
  - `DMChannelID` (required)
  - `DMMessageID` (required)
- Create `GenInvoiceResponse` struct with:
  - `Success` (bool)
  - `Message` (string)
- Add `Validate()` method to request:
  - Check all required fields are non-empty
  - Validate month format using `time.Parse("2006-01", month)`
  - Return descriptive validation errors

**Dependencies**: None

**Acceptance Criteria**:
- Structs follow Go naming conventions
- JSON tags match specification
- Validation catches all invalid inputs
- Error messages are user-friendly

---

#### Task 1.4: Implement Webhook Handler Logic
**Description**: Create webhook endpoint handler with async processing

**Complexity**: Large

**Files to Create/Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/webhook/gen_invoice.go` (handler implementation)
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/webhook/interface.go` (add method to interface)

**Implementation Details**:
- Add `GenInvoice(c *gin.Context)` method to `IWebhookHandler` interface
- Implement handler struct with dependencies:
  - `controller.IInvoiceController` (invoice generation)
  - `discord.IDiscordService` (DM message updates)
  - `notion.INotionService` (contractor email lookup)
  - `googledrive.GoogleDriveService` (file sharing)
  - `ratelimit.InvoiceRateLimiter` (rate limiting)
  - `*logrus.Logger` (structured logging)

**Handler Flow**:
1. Parse and validate JSON request body
2. Check rate limit for user
3. Return 200 OK immediately (< 1 second)
4. Start goroutine for async processing:
   - Parse month to get year and month integers
   - Call invoice controller to generate invoice
   - Get contractor's personal email from Notion
   - Share Drive file with email address
   - Update Discord DM with success/error embed

**Helper Methods**:
- `processInvoiceGeneration(req GenInvoiceRequest)` - Async processing logic
- `updateDiscordWithSuccess(channelID, messageID, fileURL, email, month)` - Success embed
- `updateDiscordWithError(channelID, messageID, title, message)` - Error embed
- `formatMonthDisplay(month string)` - Convert "2025-01" to "January 2025"

**Error Handling**:
- Rate limit exceeded: 429 response + Discord error message
- Invalid request: 400 response (no Discord update)
- Validation errors: 400 response (no Discord update)
- Async errors: Discord DM error embeds with specific messages
- Log all errors with structured fields (username, month, error)

**Dependencies**: Tasks 1.1, 1.2, 1.3

**Acceptance Criteria**:
- Handler returns within 1 second
- Async processing handles all error cases
- Discord DM updates for all error scenarios
- Structured logging at each processing step
- Rate limiting enforced correctly

---

#### Task 1.5: Add Webhook Handler Unit Tests
**Description**: Comprehensive tests for webhook handler behavior

**Complexity**: Medium

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/webhook/gen_invoice_test.go`

**Implementation Details**:
- Use table-driven tests with golden file pattern (if applicable)
- Mock all external dependencies (controller, services)
- Test synchronous response behavior
- Test async processing (may need to wait or use channels)

**Test Cases Required**:
- `TestGenInvoice_Success` - Happy path with valid request
- `TestGenInvoice_InvalidRequest` - Missing required fields
- `TestGenInvoice_InvalidMonthFormat` - Invalid YYYY-MM format
- `TestGenInvoice_RateLimited` - User exceeded daily limit
- `TestGenInvoice_GenerationFailed` - Invoice controller returns error
- `TestGenInvoice_EmailLookupFailed` - Notion service fails to find email
- `TestGenInvoice_FileSharingFailed` - Drive API returns error
- `TestGenInvoice_DiscordUpdateFailed` - Discord service fails (log error only)

**Dependencies**: Task 1.4

**Acceptance Criteria**:
- All test cases pass
- Code coverage > 80% for handler logic
- Mocks verify expected service calls
- Tests don't depend on external services

---

### Group C: Application Integration (Depends on Groups A & B)

#### Task 1.6: Register Webhook Route
**Description**: Add webhook endpoint to API routes

**Complexity**: Small

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/routes/v1.go` (or appropriate routes file)

**Implementation Details**:
- Add POST route at `/webhooks/discord/gen-invoice`
- Map route to `handler.Webhook.GenInvoice` method
- No authentication middleware (webhook is public but rate-limited)
- Place in webhooks route group

**Route Structure**:
```go
webhooks := router.Group("/webhooks")
{
    discord := webhooks.Group("/discord")
    {
        discord.POST("/gen-invoice", handler.Webhook.GenInvoice)
    }
}
```

**Dependencies**: Task 1.4

**Acceptance Criteria**:
- Route is registered correctly
- Endpoint is accessible at expected path
- Handler is called when POST request received
- No authentication required

---

#### Task 1.7: Initialize Services in Main Application
**Description**: Wire up new services and dependencies in main.go

**Complexity**: Small

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-api/cmd/server/main.go`

**Implementation Details**:
- Initialize rate limiter: `rateLimiter := ratelimit.NewInvoiceRateLimiter()`
- Pass rate limiter to webhook handler constructor
- Ensure webhook handler has all required dependencies:
  - Invoice controller (should already exist)
  - Discord service (should already exist)
  - Notion service (should already exist)
  - Google Drive service (should already exist)
  - Rate limiter (new)
  - Logger (should already exist)

**Dependencies**: Tasks 1.1, 1.4

**Acceptance Criteria**:
- Application compiles successfully
- Rate limiter starts cleanup goroutine on initialization
- Webhook handler has all dependencies wired correctly
- No circular dependency issues

---

#### Task 1.8: Verify Invoice Controller Response Format
**Description**: Ensure invoice controller returns FileID in response

**Complexity**: Small

**Files to Check/Modify** (if needed):
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go`

**Implementation Details**:
- Review `GenerateContractorInvoice` response structure
- Verify response includes:
  - `FileID` (Google Drive file ID)
  - `FileURL` (shareable URL)
- If missing, add fields to response model:
  ```go
  type GenerateContractorInvoiceResponse struct {
      Success bool   `json:"success"`
      Message string `json:"message"`
      FileID  string `json:"file_id"`   // Add if missing
      FileURL string `json:"file_url"`  // Add if missing
  }
  ```
- Ensure upload logic populates these fields after Drive upload

**Dependencies**: None

**Acceptance Criteria**:
- Response includes FileID and FileURL
- Values are populated correctly after Drive upload
- No breaking changes to existing consumers

---

### Group D: Notion Service Enhancement (Optional, depends on existing API)

#### Task 1.9: Add Contractor Email Lookup Method (If Not Exists)
**Description**: Ensure Notion service can retrieve contractor's personal email

**Complexity**: Small

**Files to Check/Modify** (if needed):
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/interface.go`
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/notion.go` (or equivalent)

**Implementation Details**:
- Check if method exists: `GetContractorPersonalEmail(discordUsername string) (string, error)`
- If missing, implement method:
  - Query Notion database for contractor by Discord username
  - Extract Personal Email field from contractor record
  - Return email or error if not found
- If exists, verify it returns personal email (not work email)

**Dependencies**: None

**Acceptance Criteria**:
- Method exists and returns personal email
- Error handling for contractor not found
- Error handling for missing email field
- Returns correct email for test contractors

---

## Repository 2: fortress-discord

### Group E: Foundation (fortress-discord)

#### Task 2.1: Create Model Layer
**Description**: Define data structures for invoice generation request/response

**Complexity**: Small

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/model/gen_invoice.go`

**Implementation Details**:
- Create `GenInvoiceRequest` struct:
  - `DiscordUsername string`
  - `Month string` (YYYY-MM format)
  - `DMChannelID string`
  - `DMMessageID string`
- Create `GenInvoiceResponse` struct:
  - `Success bool`
  - `Message string`
  - `FileURL string` (optional)
  - `Email string` (optional)
  - `Month string`
  - `Error string` (optional)
- Match fortress-api webhook request format exactly

**Dependencies**: None

**Acceptance Criteria**:
- Structs have correct JSON tags
- Fields match API specification
- Follows project naming conventions

---

#### Task 2.2: Create Adapter Layer
**Description**: HTTP client for posting webhooks to fortress-api

**Complexity**: Small

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/adapter/fortress/gen_invoice.go`

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/adapter/fortress/interface.go`

**Implementation Details**:
- Add `GenerateInvoice(ctx context.Context, req *model.GenInvoiceRequest) error` to interface
- Implement adapter struct with:
  - `*http.Client` (with timeout, reuse existing or create)
  - `webhookURL string` (from config)
- Implement method:
  - Marshal request to JSON
  - POST to webhook URL with context
  - Check response status code (200 = success, 400/429 = error)
  - Return error with response body if status >= 400
- Add timeout of 5 seconds to HTTP request

**Dependencies**: Task 2.1

**Acceptance Criteria**:
- Adapter posts correct JSON payload
- Handles HTTP errors gracefully
- Timeout prevents hanging requests
- Error messages include status code and response body

---

#### Task 2.3: Create View Layer
**Description**: Discord embed formatting for invoice generation messages

**Complexity**: Small

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/view/geninvoice/interface.go`
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/view/geninvoice/view.go`

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/view/interface.go`
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/view/view.go`

**Implementation Details**:
- Create `GenInvoiceView` interface with methods:
  - `ProcessingEmbed(month string) *discordgo.MessageEmbed`
  - `SuccessEmbed(month, fileURL, email string) *discordgo.MessageEmbed`
  - `ErrorEmbed(title, message string) *discordgo.MessageEmbed`
- Implement embeds with:
  - **Processing**: Blue color (0x3498db), "Processing..." message
  - **Success**: Green color (0x2ecc71), fields for Month, Email, File link
  - **Error**: Red color (0xe74c3c), error title and message
- Add helper: `formatMonthDisplay(month string)` - Convert "2025-01" to "January 2025"
- Include timestamps and footer text in all embeds

**Dependencies**: None

**Acceptance Criteria**:
- Embeds have correct colors and formatting
- Month displays as "January 2025" format
- File URL is a clickable link
- Embeds follow project style guidelines

---

### Group F: Business Logic (fortress-discord, depends on Group E)

#### Task 2.4: Create Service Layer
**Description**: Business logic for handling invoice generation command

**Complexity**: Medium

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/service/geninvoice/interface.go`
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/service/geninvoice/service.go`

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/service/interface.go`
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/service/service.go`

**Implementation Details**:
- Create `GenInvoiceService` interface:
  - `GenerateInvoice(ctx context.Context, userID, username, month string) error`
- Implement service struct with dependencies:
  - `*discordgo.Session` (Discord API client)
  - `fortress.GenInvoiceAdapter` (webhook client)
  - `view.GenInvoiceView` (embed formatting)
  - `logger` (logging)

**Service Flow**:
1. Validate month format (if provided) or default to current month
2. Create/get DM channel with user (by userID)
3. Send "Processing..." embed to DM
4. Build webhook request with username, month, dmChannelID, dmMessageID
5. POST webhook to fortress-api via adapter
6. If webhook fails, update DM with error embed
7. Return success (async processing continues in fortress-api)

**Helper Methods**:
- `isValidMonthFormat(month string) bool` - Validate YYYY-MM format
- `defaultMonth() string` - Return current month in YYYY-MM format

**Error Handling**:
- DM creation fails: Return error (user may have DMs disabled)
- Invalid month format: Return validation error
- Webhook POST fails: Update DM with error + return error

**Dependencies**: Tasks 2.1, 2.2, 2.3

**Acceptance Criteria**:
- Service creates DM and sends processing embed
- Webhook request posted with correct data
- Errors update DM before returning
- Month defaults to current if not provided
- Structured logging at each step

---

### Group G: Command Layer (fortress-discord, depends on Group F)

#### Task 2.5: Create Command Handler
**Description**: Discord command implementation for `?gen invoice`

**Complexity**: Medium

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/command/gen/interface.go`
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/command/gen/command.go`

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/command/command.go`

**Implementation Details**:
- Create `GenCommand` struct implementing command interface
- Implement required methods:
  - `Command() string` - Returns "gen"
  - `Help() string` - Returns help text
  - `Usage() string` - Returns usage instructions with examples
  - `Execute(ctx *base.CommandContext) error` - Command execution logic

**Command Structure**:
- Primary command: `?gen`
- Subcommands:
  - `invoice` or `inv` - Generate invoice
- Arguments:
  - Optional: `YYYY-MM` month format

**Usage Examples**:
```
?gen invoice          # Current month
?gen inv              # Short form
?gen invoice 2025-01  # Specific month
```

**Execute Flow**:
1. Parse args to get subcommand
2. If no subcommand, return usage error
3. Route to subcommand handler:
   - `invoice` or `inv` -> `handleInvoice()`
4. Extract month from args (if provided)
5. Get user ID and username from context
6. Call service.GenerateInvoice()
7. If DM creation fails, reply in channel with error
8. Otherwise, reply in channel: "Check your DMs for status"

**Error Handling**:
- No subcommand: Return usage instructions
- Unknown subcommand: Return "Unknown subcommand" error
- DM disabled: Reply in channel with instructions to enable DMs
- Service error: Reply in channel with generic error

**Dependencies**: Task 2.4

**Acceptance Criteria**:
- Command responds to `?gen invoice` and `?gen inv`
- Help and usage text are clear and complete
- Errors have user-friendly messages
- Channel replies guide user to check DMs
- Unknown subcommands handled gracefully

---

### Group H: Integration (fortress-discord, depends on Groups E, F, G)

#### Task 2.6: Register Command in Bot
**Description**: Wire up GenCommand in main bot initialization

**Complexity**: Small

**Files to Modify**:
- Main bot initialization file (likely `cmd/bot/main.go` or `pkg/discord/bot.go`)

**Implementation Details**:
- Initialize GenInvoiceView
- Initialize GenInvoiceAdapter with webhook URL from config
- Initialize GenInvoiceService with dependencies:
  - Discord session
  - Fortress adapter
  - View
  - Logger
- Initialize GenCommand with service
- Register GenCommand in command registry

**Initialization Order**:
```go
// Views
genInvoiceView := view.NewGenInvoiceView()

// Adapters
fortressAdapter := fortress.NewGenInvoiceAdapter(httpClient, config.FortressAPIWebhookURL)

// Services
genInvoiceService := service.NewGenInvoiceService(
    discordSession,
    fortressAdapter,
    genInvoiceView,
    logger,
)

// Commands
genCommand := gen.NewGenCommand(genInvoiceService)
commandRegistry.Register(genCommand)
```

**Dependencies**: Tasks 2.1, 2.2, 2.3, 2.4, 2.5

**Acceptance Criteria**:
- Command is registered and responds to `?gen`
- All dependencies are correctly wired
- Application compiles and starts successfully
- No circular dependencies

---

#### Task 2.7: Add Configuration
**Description**: Add fortress-api webhook URL to bot configuration

**Complexity**: Small

**Files to Modify**:
- Configuration file (e.g., `config.yaml`, `.env`, or equivalent)
- Configuration struct (if using typed config)

**Implementation Details**:
- Add configuration field: `FORTRESS_API_WEBHOOK_URL`
- Default value for local dev: `http://localhost:8080/webhooks/discord/gen-invoice`
- Production value: `https://fortress-api.production.com/webhooks/discord/gen-invoice`
- Optional: Add timeout configuration (default 5s)

**Example Config**:
```yaml
fortress_api:
  webhook_url: "http://localhost:8080/webhooks/discord/gen-invoice"
  timeout: 5s
```

**Dependencies**: None

**Acceptance Criteria**:
- Configuration is loaded correctly
- Different values for dev/staging/production
- URL validation (must be valid HTTP/HTTPS URL)
- Documentation for configuration values

---

## Repository 3: Testing & Deployment

### Group I: Integration Testing

#### Task 3.1: Manual Testing - fortress-api Webhook
**Description**: Test webhook endpoint in isolation before Discord integration

**Complexity**: Small

**Testing Steps**:
1. Start fortress-api locally: `make dev`
2. Use curl or Postman to POST webhook:
   ```bash
   curl -X POST http://localhost:8080/webhooks/discord/gen-invoice \
     -H "Content-Type: application/json" \
     -d '{
       "discord_username": "test_contractor",
       "month": "2025-01",
       "dm_channel_id": "test_channel_id",
       "dm_message_id": "test_message_id"
     }'
   ```
3. Verify:
   - Returns 200 OK immediately
   - Logs show async processing started
   - Rate limiter tracks request
   - Discord service called (may fail if test IDs invalid - that's ok)

**Test Cases**:
- Valid request (active contractor)
- Invalid month format (should return 400)
- Missing required fields (should return 400)
- Rate limiting (4th request should return 429)
- Non-existent contractor (async error, check logs)

**Dependencies**: All tasks in Group A, B, C

**Acceptance Criteria**:
- Webhook responds within 1 second
- Validation works correctly
- Rate limiting enforced
- Async processing starts
- Errors logged appropriately

---

#### Task 3.2: Manual Testing - fortress-discord Command
**Description**: Test Discord command end-to-end

**Complexity**: Medium

**Testing Steps**:
1. Start fortress-api: `make dev`
2. Start fortress-discord bot
3. In Discord, execute commands:
   - `?gen invoice` (current month)
   - `?gen inv 2025-01` (specific month)
   - `?gen invoice invalid-format` (should error)
   - `?gen unknown` (unknown subcommand)
   - `?gen` (no subcommand)
4. Verify DM behavior:
   - DM received with "Processing..." embed
   - DM updates with success/error after processing
5. Test rate limiting:
   - Execute command 4 times rapidly
   - 4th should show rate limit error in DM

**Test Cases**:
- Happy path (valid contractor, valid month)
- Default month (no month provided)
- Invalid month format
- Unknown subcommand
- DMs disabled (test with user who has DMs off)
- Rate limiting (4 requests)
- Invalid contractor (not in Notion)
- Missing personal email (contractor without email)

**Dependencies**: All tasks in Groups E, F, G, H + Task 3.1

**Acceptance Criteria**:
- Command executes successfully
- DM sent and updated correctly
- Errors shown in DM with clear messages
- Rate limiting works across requests
- Invoice generated and shared to email
- Google notification email received

---

#### Task 3.3: Automated Integration Tests (Optional)
**Description**: Write automated tests for end-to-end flow

**Complexity**: Large

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/webhook/integration_test.go` (optional)
- Integration test suite in fortress-discord (optional)

**Implementation Details**:
- Mock external services (Notion, Google Drive, Discord)
- Test full webhook flow from request to Discord update
- Test error scenarios
- Test rate limiting across multiple requests
- Use testcontainers or in-memory implementations

**Test Scenarios**:
- Full successful flow
- Rate limit exceeded
- Invalid contractor
- Missing email
- File sharing failed
- Discord update failed

**Dependencies**: Tasks 3.1, 3.2

**Acceptance Criteria**:
- Tests run in CI/CD pipeline
- No external service dependencies
- All scenarios covered
- Tests complete in < 30 seconds

---

### Group J: Documentation & Deployment

#### Task 3.4: Update API Documentation
**Description**: Document new webhook endpoint

**Complexity**: Small

**Files to Create/Modify**:
- API documentation (Swagger/OpenAPI if applicable)
- README updates for new endpoint

**Documentation to Add**:
- Endpoint: `POST /webhooks/discord/gen-invoice`
- Request schema with example
- Response schema with example
- Error codes (400, 429)
- Rate limiting rules
- Example curl commands

**Dependencies**: All implementation tasks

**Acceptance Criteria**:
- Endpoint documented in API docs
- Examples are accurate and tested
- Error codes and messages documented
- Rate limiting rules explained

---

#### Task 3.5: Update Discord Bot Documentation
**Description**: Document new command for users

**Complexity**: Small

**Files to Create/Modify**:
- Bot command documentation
- User guide or README

**Documentation to Add**:
- Command: `?gen invoice [YYYY-MM]`
- Alias: `?gen inv [YYYY-MM]`
- Description: Generate invoice for specified month
- Usage examples
- Rate limiting notice (3 per day)
- Troubleshooting (DMs disabled, not a contractor, etc.)

**Dependencies**: All implementation tasks

**Acceptance Criteria**:
- Command documented clearly
- Examples cover common use cases
- Troubleshooting section complete
- Help text matches actual command

---

#### Task 3.6: Deployment - fortress-api
**Description**: Deploy fortress-api changes to production

**Complexity**: Medium

**Deployment Steps**:
1. Merge PR to develop branch
2. Run full test suite: `make test`
3. Deploy to staging environment
4. Test webhook endpoint in staging
5. Monitor logs for errors
6. Deploy to production
7. Smoke test webhook endpoint

**Pre-Deployment Checklist**:
- All unit tests pass
- Integration tests pass (if implemented)
- Code review approved
- Rate limiter tested for memory leaks
- Google Drive permissions tested
- Discord service integration tested

**Rollback Plan**:
- Remove webhook route
- Redeploy previous version
- No database cleanup needed (in-memory state)

**Dependencies**: Tasks 3.1, 3.2, 3.4

**Acceptance Criteria**:
- Deployed to production successfully
- Webhook endpoint accessible
- No errors in production logs
- Rate limiter initializes correctly

---

#### Task 3.7: Deployment - fortress-discord
**Description**: Deploy fortress-discord changes to production

**Complexity**: Medium

**Deployment Steps**:
1. Ensure fortress-api deployed first (Task 3.6)
2. Merge PR to main/develop branch
3. Deploy to staging environment
4. Test command in staging Discord server
5. Monitor logs for errors
6. Deploy to production
7. Test command in production Discord server

**Pre-Deployment Checklist**:
- fortress-api webhook deployed and tested
- All tests pass
- Code review approved
- Configuration updated (webhook URL)
- Command registered correctly

**Rollback Plan**:
- Unregister GenCommand from bot
- Redeploy previous version
- Users get "Unknown command" error

**Dependencies**: Tasks 3.2, 3.5, 3.6

**Acceptance Criteria**:
- Bot deployed successfully
- Command responds in Discord
- End-to-end flow works
- No errors in production logs
- Users can generate invoices

---

#### Task 3.8: Post-Deployment Monitoring
**Description**: Monitor production usage and errors

**Complexity**: Small

**Monitoring Tasks**:
1. Monitor fortress-api logs:
   - Webhook requests per hour
   - Rate limiting hits
   - Async processing errors
   - Google Drive API errors
2. Monitor fortress-discord logs:
   - Command usage frequency
   - DM creation failures
   - Webhook POST failures
3. Monitor user feedback in Discord support channels
4. Track invoice generation success rate

**Metrics to Track**:
- Webhook requests/day
- Rate limit rejections/day
- Success rate (%)
- Average processing time
- Error types and frequency

**Dependencies**: Tasks 3.6, 3.7

**Acceptance Criteria**:
- Logging provides visibility into usage
- Errors are actionable
- Success rate > 95%
- No memory leaks from rate limiter

---

## Implementation Order

### Recommended Parallelization

**Phase 1 - Parallel** (Can start simultaneously):
- Task 1.1 (Rate Limiter)
- Task 1.2 (Google Drive)
- Task 2.1 (Discord Models)
- Task 2.3 (Discord Views)
- Task 2.7 (Discord Config)

**Phase 2 - Sequential**:
- Task 1.3 (Webhook Models)
- Task 1.4 (Webhook Handler) - requires 1.1, 1.2, 1.3
- Task 1.5 (Webhook Tests) - requires 1.4

**Phase 3 - Parallel**:
- Task 1.6 (Routes) - requires 1.4
- Task 1.7 (Main.go) - requires 1.1, 1.4
- Task 1.8 (Controller Check) - independent
- Task 1.9 (Notion Check) - independent
- Task 2.2 (Discord Adapter) - requires 2.1
- Task 2.4 (Discord Service) - requires 2.1, 2.2, 2.3

**Phase 4 - Sequential**:
- Task 2.5 (Discord Command) - requires 2.4
- Task 2.6 (Bot Registration) - requires 2.5

**Phase 5 - Testing**:
- Task 3.1 (API Testing) - requires Phase 2-3 complete
- Task 3.2 (Discord Testing) - requires Phase 4 complete

**Phase 6 - Deployment**:
- Task 3.4 (API Docs)
- Task 3.5 (Bot Docs)
- Task 3.6 (API Deploy) - requires 3.1
- Task 3.7 (Bot Deploy) - requires 3.2, 3.6
- Task 3.8 (Monitoring) - requires 3.6, 3.7

### Critical Path
Tasks on critical path (cannot be parallelized):
1. Task 1.3 → 1.4 → 1.5 → 1.6/1.7 → 3.1 → 3.6
2. Task 2.1 → 2.2 → 2.4 → 2.5 → 2.6 → 3.2 → 3.7

### Estimated Timeline
- **Phase 1**: 1 day (parallel)
- **Phase 2**: 1 day
- **Phase 3**: 1 day (parallel)
- **Phase 4**: 0.5 days
- **Phase 5**: 1 day
- **Phase 6**: 1 day

**Total**: 5-6 days (with parallelization)

---

## Task Summary by Complexity

### Small (< 2 hours each)
- Task 1.2: Google Drive sharing
- Task 1.3: Webhook models
- Task 1.6: Route registration
- Task 1.7: Main.go initialization
- Task 1.8: Controller check
- Task 1.9: Notion check
- Task 2.1: Discord models
- Task 2.2: Discord adapter
- Task 2.3: Discord views
- Task 2.6: Bot registration
- Task 2.7: Discord config
- Task 3.1: API manual testing
- Task 3.4: API documentation
- Task 3.5: Bot documentation
- Task 3.8: Monitoring setup

### Medium (2-4 hours each)
- Task 1.1: Rate limiter
- Task 1.5: Webhook tests
- Task 2.4: Discord service
- Task 2.5: Discord command
- Task 3.2: Discord manual testing
- Task 3.6: API deployment
- Task 3.7: Bot deployment

### Large (4-8 hours each)
- Task 1.4: Webhook handler
- Task 3.3: Automated integration tests (optional)

---

## Success Criteria

### For fortress-api
- Webhook endpoint responds < 1 second
- Rate limiter enforces 3/day limit correctly
- Invoice generation succeeds for valid contractors
- Google Drive files shared with correct permissions
- Discord DMs updated with success/error messages
- All unit tests pass
- No memory leaks from rate limiter

### For fortress-discord
- Command responds to `?gen invoice` and aliases
- DM sent with processing message immediately
- Webhook POST succeeds and returns quickly
- Error messages are user-friendly
- Help text is clear and accurate
- All unit tests pass

### For Integration
- End-to-end flow completes successfully
- Invoice generated and shared via email
- Google notification email received by contractor
- Rate limiting works across repositories
- Errors communicated clearly via Discord DM
- No breaking changes to existing functionality

---

## Risk Mitigation

### High Risk Areas
1. **Async Processing Errors**: DM updates may fail silently
   - **Mitigation**: Comprehensive error logging, manual monitoring initially

2. **Rate Limiter Memory Leaks**: Map grows indefinitely
   - **Mitigation**: Cleanup goroutine tested thoroughly, monitoring

3. **Google Drive Permissions**: File may not be accessible
   - **Mitigation**: Test with multiple contractors, verify notification emails

4. **Discord DMs Disabled**: Users can't receive notifications
   - **Mitigation**: Clear error message, fallback to channel reply

### Medium Risk Areas
1. **Notion Email Lookup**: Personal email may be missing
   - **Mitigation**: Error handling with actionable message

2. **Invoice Generation Failures**: Existing controller may have bugs
   - **Mitigation**: Reuse proven logic, comprehensive testing

3. **Webhook Timeout**: fortress-discord may timeout waiting
   - **Mitigation**: Return 200 OK immediately, async processing

---

## Next Steps After Implementation

1. **Monitor Usage**: Track command usage frequency
2. **Gather Feedback**: Ask contractors for feedback on UX
3. **Iterate on Error Messages**: Improve based on common errors
4. **Consider Enhancements**:
   - Batch generation (multiple months)
   - Email notifications in addition to Discord DM
   - Web dashboard for invoice history
   - Persistent rate limiting (Redis)

---

**Document Version**: 1.0
**Last Updated**: 2026-01-06
**Next Review**: After implementation completion
