# Planning Phase Status

## Session
**Session ID**: 202601062342-gen-invoice-discord-command
**Feature**: Discord command `?gen invoice` for contractor invoice generation
**Phase**: Planning
**Status**: Complete
**Date**: 2026-01-06

## Overview
This planning phase has defined the architecture and implementation approach for a new Discord command that allows contractors to generate their own invoices via async processing with fortress-api.

## Deliverables

### Architecture Decision Records (ADRs)

#### ADR-001: Async Discord Command Pattern
**Location**: `planning/ADRs/ADR-001-async-discord-command-pattern.md`

**Decision**: Use DM message updates for async command processing
- Create DM with "Processing..." embed
- Pass dmChannelID and dmMessageID to webhook
- Update same DM message with result

**Key Benefits**:
- Avoids Discord 3-second timeout
- Private communication for financial data
- Single message maintains context
- Clean user experience

**Trade-offs**:
- Requires message tracking
- User must have DMs enabled

---

#### ADR-002: In-Memory Rate Limiting
**Location**: `planning/ADRs/ADR-002-in-memory-rate-limiting.md`

**Decision**: Use in-memory map with mutex for rate limiting
- Limit: 3 invoice generations per day per user
- Data structure: Go map with sync.RWMutex
- Reset: Daily at midnight UTC
- Cleanup: Hourly goroutine

**Key Benefits**:
- No database migrations (meets requirement)
- Extremely fast (nanosecond lookups)
- Zero infrastructure dependencies
- Simple implementation

**Trade-offs**:
- Lost on server restart (acceptable)
- Not shared across instances (single-instance deployment)
- No historical data

---

#### ADR-003: Google Drive File Sharing
**Location**: `planning/ADRs/ADR-003-google-drive-file-sharing.md`

**Decision**: Use Google Drive Share API instead of email attachments
- Share file with contractor's Personal Email via Drive API
- Google automatically sends notification email
- No SendGrid email needed

**Key Benefits**:
- Single source of truth (file in Drive)
- Automatic notification from Google
- No duplicate storage
- Better permission control
- Less code to maintain

**Trade-offs**:
- Cannot customize notification email template
- Relies on Google's notification system

---

### Technical Specifications

#### fortress-discord Specification
**Location**: `planning/specifications/fortress-discord-spec.md`

**Scope**: Complete implementation plan for Discord bot changes

**Components**:
1. **Model Layer**: Data structures for requests/responses
2. **Adapter Layer**: HTTP client for webhook POSTs to fortress-api
3. **Service Layer**: Business logic for command handling
4. **View Layer**: Discord embed formatting
5. **Command Layer**: Command registration and execution

**Files to Create** (5):
- `pkg/discord/model/gen_invoice.go`
- `pkg/adapter/fortress/gen_invoice.go`
- `pkg/discord/service/gen_invoice.go`
- `pkg/discord/view/gen_invoice.go`
- `pkg/discord/command/gen/gen.go`

**Files to Modify** (2):
- Configuration file (add webhook URL)
- Bot initialization (register command)

---

#### fortress-api Specification
**Location**: `planning/specifications/fortress-api-spec.md`

**Scope**: Complete implementation plan for API changes

**Components**:
1. **Rate Limiter**: In-memory rate limiting service
2. **Google Drive Enhancement**: File sharing method
3. **Webhook Handler**: Async invoice generation endpoint
4. **Routes**: Webhook route registration

**Files to Create** (4):
- `pkg/service/ratelimit/invoice_rate_limiter.go`
- `pkg/service/ratelimit/invoice_rate_limiter_test.go`
- `pkg/handler/webhook/gen_invoice.go`
- `pkg/handler/webhook/gen_invoice_test.go`

**Files to Modify** (7):
- `pkg/service/googledrive/google_drive.go` (add ShareFileWithEmail)
- `pkg/service/googledrive/interface.go` (update interface)
- `pkg/service/googledrive/google_drive_test.go` (add tests)
- `pkg/handler/webhook/interface.go` (add GenInvoice method)
- `pkg/routes/v1.go` (add webhook route)
- `cmd/server/main.go` (initialize rate limiter)
- `pkg/controller/invoice/contractor_invoice.go` (ensure FileID in response)

---

## Architecture Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ User Issues Command                                             │
│ ?gen invoice 2025-01                                            │
└────────────────┬────────────────────────────────────────────────┘
                 │
                 v
┌─────────────────────────────────────────────────────────────────┐
│ fortress-discord                                                │
│                                                                 │
│ 1. Validate month format                                        │
│ 2. Create/get DM channel with user                             │
│ 3. Send "Processing..." embed to DM                            │
│ 4. Store dmChannelID and dmMessageID                           │
│ 5. POST webhook to fortress-api with:                          │
│    - discord_username                                           │
│    - month (YYYY-MM)                                            │
│    - dm_channel_id                                              │
│    - dm_message_id                                              │
│ 6. Return immediately (< 3s)                                    │
└────────────────┬────────────────────────────────────────────────┘
                 │ HTTP POST
                 v
┌─────────────────────────────────────────────────────────────────┐
│ fortress-api Webhook                                            │
│ POST /webhooks/discord/gen-invoice                              │
│                                                                 │
│ 1. Parse and validate request                                   │
│ 2. Check rate limit (in-memory)                                 │
│ 3. Return 200 OK immediately                                    │
│ 4. Start async goroutine:                                       │
│    a. Generate invoice (existing controller)                    │
│    b. Get contractor email from Notion                          │
│    c. Upload PDF to Google Drive (existing)                     │
│    d. Share file with email (new method)                        │
│    e. Update Discord DM with result                             │
└────────────────┬────────────────────────────────────────────────┘
                 │ Async processing
                 v
┌─────────────────────────────────────────────────────────────────┐
│ External Services                                               │
│                                                                 │
│ - Notion: Fetch contractor data + personal email               │
│ - Google Drive: Upload PDF + Share with email                  │
│ - Discord: UpdateChannelMessage (edit DM)                       │
│ - Google: Send automatic notification email                     │
└─────────────────────────────────────────────────────────────────┘
                 │
                 v
┌─────────────────────────────────────────────────────────────────┐
│ User Result                                                     │
│                                                                 │
│ DM updated with:                                                │
│ - Success: File URL, email, month                               │
│ - Error: Error message and reason                               │
│                                                                 │
│ Email notification from Google with file access                │
└─────────────────────────────────────────────────────────────────┘
```

## Key Technical Decisions

### 1. Async Processing Pattern
- Immediate webhook response (< 3s) to avoid Discord timeout
- Background goroutine for long-running operations
- DM message update for result notification

### 2. Rate Limiting Strategy
- In-memory implementation (no database)
- Thread-safe with sync.RWMutex
- 3 attempts per day per user
- Resets at midnight UTC

### 3. File Delivery Method
- Google Drive API sharing
- Automatic email notification from Google
- Single source of truth for files

### 4. Error Handling
- All async errors communicated via Discord DM updates
- Graceful degradation (rate limit, validation errors)
- Comprehensive logging at each step

## Dependencies

### Existing Components (Reused)
- `pkg/controller/invoice/contractor_invoice.go` - Invoice generation
- `pkg/service/notion/task_order_log.go` - Email lookup
- `pkg/service/discord/discord.go` - Message updates
- `pkg/service/googledrive/google_drive.go` - File upload

### New Components
- Rate limiter service (in-memory)
- Webhook handler for Discord
- Google Drive file sharing method
- Discord command implementation

### External APIs
- Discord API (DM creation, message updates)
- Notion API (contractor data)
- Google Drive API (upload, sharing)
- Google Workspace (automatic email notifications)

## Testing Strategy

### Unit Tests
- Rate limiter: Thread-safety, reset logic, cleanup
- Google Drive service: File sharing, validation
- Webhook handler: Request validation, rate limiting
- Discord command: Subcommand routing, validation
- View layer: Embed formatting

### Integration Tests
- End-to-end webhook flow (mocked external services)
- Rate limit enforcement across requests
- Error scenarios with DM updates

### Manual Testing
- Command execution in Discord
- Rate limit enforcement (4th request fails)
- Invalid contractor handling
- Invalid month format
- DM message updates (success/error states)

## Constraints & Requirements Met

- Discord 3-second timeout: Async pattern with immediate response
- No database migrations: In-memory rate limiting
- Reuse existing logic: Leverages existing invoice controller
- Private communication: DM-based notifications
- File sharing: Google Drive API with automatic email
- Rate limiting: 3 per day per user
- Thread-safe: Mutex-protected map

## Risk Assessment

### Low Risk
- Existing invoice generation logic is proven
- Google Drive API already in use
- Discord service already handles message updates
- In-memory rate limiting is simple and fast

### Medium Risk
- User must have DMs enabled (mitigated with error message)
- Rate limiter lost on restart (acceptable trade-off)
- Google notification email cannot be customized (acceptable)

### Mitigation Strategies
- Comprehensive error handling with user-friendly messages
- Extensive logging for debugging
- Fallback to channel reply if DM fails
- Clear documentation for support team

## Next Steps

### 1. Requirements Review
- Review and approve ADRs with stakeholders
- Confirm technical approach
- Validate error handling strategy

### 2. Test Case Design
- Hand off to test-case-designer agent
- Define unit test cases for all components
- Define integration test scenarios
- Define manual test procedures

### 3. Implementation
- Implement fortress-api changes first (webhook endpoint)
- Test webhook endpoint in isolation
- Implement fortress-discord changes
- End-to-end testing in staging environment

### 4. Deployment
- Deploy fortress-api (webhook available)
- Deploy fortress-discord (command available)
- Monitor logs and error rates
- Gather user feedback

## Documentation References

### ADRs
- `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601062342-gen-invoice-discord-command/planning/ADRs/ADR-001-async-discord-command-pattern.md`
- `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601062342-gen-invoice-discord-command/planning/ADRs/ADR-002-in-memory-rate-limiting.md`
- `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601062342-gen-invoice-discord-command/planning/ADRs/ADR-003-google-drive-file-sharing.md`

### Specifications
- `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601062342-gen-invoice-discord-command/planning/specifications/fortress-discord-spec.md`
- `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601062342-gen-invoice-discord-command/planning/specifications/fortress-api-spec.md`

### Requirements
- `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601062342-gen-invoice-discord-command/requirements/requirements.md`

---

**Planning Phase Complete**: Ready for test case design and implementation
