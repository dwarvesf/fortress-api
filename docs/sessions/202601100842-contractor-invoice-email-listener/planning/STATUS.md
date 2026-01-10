# Planning Status - Contractor Invoice Email Listener

## Session Information

- **Session ID**: 202601100842-contractor-invoice-email-listener
- **Status**: Planning Complete
- **Date**: 2026-01-10
- **Agent**: Project Manager

## Overview

Completed comprehensive planning for the Contractor Invoice Email Listener feature. This feature will monitor a Gmail inbox (`bill@d.foundation`) for contractor invoice submissions and automatically update Notion Contractor Payables status from "New" to "Pending".

## Documents Created

### 1. Architecture Decision Record (ADR)

**File**: `planning/ADRs/001-email-listener-architecture.md`

**Key Decisions**:
- **Email Monitoring**: Gmail API polling (5-minute interval) vs Push Notifications
  - **Decision**: Polling approach for simplicity and existing infrastructure alignment
  - **Rationale**: No new infrastructure needed, acceptable latency, easier maintenance

- **PDF Parsing Library**: pdfcpu vs unidoc/unipdf vs go-pdf
  - **Decision**: Use `github.com/pdfcpu/pdfcpu`
  - **Rationale**: Pure Go, Apache 2.0 license, active maintenance, good text extraction

- **Service Architecture**: Layered approach with separation of concerns
  - Gmail operations in `pkg/service/googlemail/` (extended)
  - Business logic in `pkg/service/invoiceemail/` (new)
  - Worker integration via existing `pkg/worker/worker.go` pattern

- **Error Handling**: Best-effort with continue-on-error strategy
  - Individual email failures don't block batch processing
  - Detailed DEBUG logging for debugging
  - Idempotent operations allow safe retries

- **Email Tracking**: Gmail labels (`fortress-api/processed`)
  - Prevents duplicate processing
  - No external state needed
  - Allows manual inspection

### 2. Gmail Inbox Service Specification

**File**: `planning/specifications/gmail-inbox-service.md`

**Scope**: Extend existing `pkg/service/googlemail/` with inbox reading capabilities

**New Methods**:
- `ListInboxMessages(query, maxResults)` - List unread messages
- `GetMessage(messageID)` - Fetch full message with headers/attachments
- `GetAttachment(messageID, attachmentID)` - Download attachment bytes
- `AddLabel(messageID, labelName)` - Mark message as processed
- `GetOrCreateLabel(labelName)` - Label management

**Helper Functions**:
- `ParseEmailHeaders(message)` - Extract Subject, From, To, Date
- `FindPDFAttachment(message)` - Find first PDF in message parts

**API Quotas**:
- 250 quota units per user per second
- Estimated usage: ~36,000 units/day (0.0036% of limit)
- Exponential backoff for rate limit handling

**Testing**:
- 13 unit test cases defined
- Integration tests with test Gmail account
- Mock Gmail API responses for unit tests

### 3. Invoice Email Processor Specification

**File**: `planning/specifications/invoice-email-processor.md`

**Scope**: New service `pkg/service/invoiceemail/` for invoice email processing logic

**Core Methods**:
- `ProcessInvoiceEmails(ctx)` - Batch processing orchestration
- `ProcessSingleEmail(ctx, messageID)` - Single email end-to-end flow
- `ExtractInvoiceID(ctx, message)` - Invoice ID extraction from subject/PDF

**Processing Flow**:
1. List unread messages without processed label
2. For each message:
   - Fetch full message
   - Extract Invoice ID (subject first, PDF fallback)
   - Find matching Notion payable record
   - Check status is "New"
   - Update status to "Pending"
   - Mark email as processed

**Invoice ID Extraction**:
- **Pattern**: `CONTR-YYYYMM-XXXX` (e.g., `CONTR-202501-A1B2`)
- **Regex**: `CONTR-\d{6}-[A-Z0-9]+`
- **Strategy**: Subject line first (fast), PDF parsing fallback (slow)
- **PDF**: Extract text from first page only using pdfcpu

**Error Handling**:
- Recoverable errors: Log and skip (no Invoice ID, already Pending)
- Non-recoverable errors: Log and retry next run (API auth failures)
- Partial failures: Log but accept (Notion updated, label failed)

**Testing**:
- 7 unit tests for processor
- 7 unit tests for extractor
- Test data: sample subjects and PDF content

### 4. Configuration Specification

**File**: `planning/specifications/configuration.md`

**Config Structure**:
```go
type InvoiceEmailListener struct {
    Enabled        bool          // Feature toggle
    EmailAddress   string        // Email to monitor
    RefreshToken   string        // OAuth refresh token
    PollInterval   time.Duration // Poll frequency
    ProcessedLabel string        // Gmail label
    MaxMessages    int64         // Batch size
    PDFMaxSizeMB   int           // PDF download limit
}
```

**Environment Variables**:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `INVOICE_LISTENER_ENABLED` | Yes | `false` | Feature toggle |
| `INVOICE_LISTENER_EMAIL` | Yes | - | Email address to monitor |
| `INVOICE_LISTENER_REFRESH_TOKEN` | Yes | - | OAuth refresh token |
| `INVOICE_LISTENER_POLL_INTERVAL` | No | `5m` | Poll interval |
| `INVOICE_LISTENER_LABEL` | No | `fortress-api/processed` | Gmail label |
| `INVOICE_LISTENER_MAX_MESSAGES` | No | `50` | Max per batch |
| `INVOICE_LISTENER_PDF_MAX_SIZE_MB` | No | `5` | Max PDF size |

**Configuration Validation**:
- Validates required fields when feature enabled
- Validates poll interval >= 1 minute
- Validates max messages between 1 and 500
- Validates PDF size between 1 and 50 MB
- Logs configuration status on startup

**Security**:
- Refresh token stored in environment variables (not code)
- Use Vault or secret manager in production
- Mask sensitive values in logs
- Use dedicated Gmail account for billing

**Feature Toggle**:
- Default disabled (`Enabled=false`)
- Graceful degradation when disabled
- No API calls if feature disabled
- Easy rollback via environment variable

## Technical Architecture

### Service Layers

```
pkg/service/googlemail/     (Extended)
  ├── google_mail.go        (Existing + new inbox methods)
  ├── interface.go          (Updated interface)
  └── google_mail_inbox_test.go (New tests)

pkg/service/invoiceemail/   (New)
  ├── interface.go          (Service interface)
  ├── processor.go          (Business logic)
  ├── extractor.go          (Invoice ID extraction)
  ├── processor_test.go     (Unit tests)
  └── extractor_test.go     (Extraction tests)

pkg/worker/
  └── worker.go             (Add InvoiceEmailCheckMsg handler)

pkg/config/
  └── config.go             (Add InvoiceEmailListener struct)
```

### Data Flow

```
Cron (Every 5 min)
    ↓
Worker.handleInvoiceEmailCheck()
    ↓
InvoiceEmailProcessor.ProcessInvoiceEmails()
    ↓
GmailService.ListInboxMessages("is:unread -label:processed has:attachment")
    ↓
For each message:
    ├→ GmailService.GetMessage(messageID)
    ├→ ExtractInvoiceID(subject or PDF)
    ├→ NotionService.QueryPayable(invoiceID)
    ├→ NotionService.UpdateStatus(pageID, "Pending")
    └→ GmailService.AddLabel(messageID, "processed")
```

### Integration Points

1. **Gmail API** (existing OAuth2 credentials)
   - Reuse `GOOGLE_API_CLIENT_ID` and `GOOGLE_API_CLIENT_SECRET`
   - New refresh token: `INVOICE_LISTENER_REFRESH_TOKEN`
   - Scopes already configured: `gmail.readonly`, `gmail.modify`, `gmail.labels`

2. **Notion API** (existing service)
   - Use existing `ContractorPayablesService`
   - Query by Invoice ID (rich text property)
   - Update Payment Status (Status property type)

3. **Worker System** (existing infrastructure)
   - Add new message type: `InvoiceEmailCheckMsg`
   - Trigger via cron (every 5 minutes)
   - Follow existing error handling patterns

## Dependencies

### New Dependencies

```go
// go.mod
require (
    github.com/pdfcpu/pdfcpu v0.5.0  // PDF text extraction
)
```

### Existing Dependencies (Reused)

- `google.golang.org/api/gmail/v1` - Gmail API client
- `golang.org/x/oauth2` - OAuth2 token management
- `github.com/dstotijn/go-notion` - Notion API client

## Risk Assessment

### Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Gmail API rate limits | Low | Medium | Exponential backoff, batch limits |
| PDF parsing failures | Medium | Low | Robust error handling, manual fallback |
| Refresh token expiration | Low | High | Validation on startup, monitoring alerts |
| Notion schema changes | Low | Medium | Property existence checks, version pinning |
| Non-standard Invoice ID formats | Medium | Low | Detailed logging, manual review process |

### Operational Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Email inbox overload | Low | Low | Batch size limit (50-100), poll interval tuning |
| Duplicate processing | Low | Medium | Gmail labels, idempotent updates |
| Processing delays | Low | Low | Acceptable per requirements (5-min latency) |

## Success Criteria

### Functional Requirements

- ✓ Poll Gmail inbox every 5 minutes (configurable)
- ✓ Extract Invoice ID from subject or PDF
- ✓ Match Invoice ID with Notion Contractor Payables
- ✓ Update status from "New" to "Pending" only
- ✓ Mark emails as processed to prevent duplicates
- ✓ Handle failures gracefully (continue processing)

### Non-Functional Requirements

- ✓ Fully configurable via environment variables
- ✓ Follows existing codebase patterns (service layers, worker integration)
- ✓ Observable with DEBUG logging for all operations
- ✓ Resilient to individual email failures
- ✓ Idempotent operations (safe to re-run)
- ✓ Testable with unit and integration tests

## Next Steps

### Handoff to Test Case Designer

The planning phase is complete. The next phase is test case design, which should:

1. **Review Planning Documents**:
   - Read all specifications and ADR
   - Understand data flow and error handling
   - Identify edge cases and failure scenarios

2. **Design Test Cases**:
   - **Unit Tests**: Gmail service methods, invoice extraction, processor logic
   - **Integration Tests**: End-to-end with test Gmail account and Notion workspace
   - **Test Data**: Sample emails with various Invoice ID formats, malformed PDFs

3. **Test Plan Structure**:
   - Test case descriptions with inputs/outputs
   - Mock data requirements
   - Test fixtures (sample PDFs, email responses)
   - Acceptance criteria for each test

4. **Edge Cases to Cover**:
   - No PDF attachment
   - PDF with no Invoice ID
   - Multiple Invoice IDs in one email (first match)
   - Payable not found in Notion
   - Payable already "Pending" or "Paid"
   - Gmail API rate limit hit
   - Notion API temporarily unavailable
   - Malformed PDF (corrupted file)
   - Very large PDF (exceeds size limit)

### Implementation Considerations

When implementation begins, developers should:

1. Start with Gmail inbox service extension (minimal risk)
2. Add PDF parsing library and test extraction
3. Implement invoice email processor service
4. Add configuration and validation
5. Integrate with worker/cron system
6. Add comprehensive logging
7. Test end-to-end in development environment
8. Deploy with feature disabled initially
9. Enable gradually in staging, then production

## Documentation Status

| Document | Status | Completeness | Review Required |
|----------|--------|--------------|-----------------|
| ADR-001: Email Listener Architecture | ✓ Complete | 100% | Ready |
| Spec: Gmail Inbox Service | ✓ Complete | 100% | Ready |
| Spec: Invoice Email Processor | ✓ Complete | 100% | Ready |
| Spec: Configuration | ✓ Complete | 100% | Ready |
| STATUS.md (this file) | ✓ Complete | 100% | Ready |

## Approval

This planning phase is complete and ready for handoff to the test case designer. All architectural decisions are documented, specifications are detailed enough for implementation, and configuration is fully defined.

**Prepared by**: Project Manager Agent
**Date**: 2026-01-10
**Session**: 202601100842-contractor-invoice-email-listener
**Status**: Planning Phase Complete ✓
