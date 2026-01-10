# ADR-001: Email Listener Architecture for Contractor Invoice Processing

## Status

Proposed

## Context

The fortress-api needs to monitor a Gmail inbox (`bill@d.foundation`) for contractor invoice submissions. When a contractor sends an invoice email with a PDF attachment, the system should:

1. Extract the Invoice ID from the email (subject line or PDF content)
2. Match the Invoice ID with a record in the Notion Contractor Payables database
3. Update the payable status from "New" to "Pending"
4. Mark the email as processed to prevent duplicate handling

### Requirements

- **Monitoring Method**: Gmail API-based (not SendGrid Inbound Parse)
- **Poll Interval**: Every 5 minutes (configurable)
- **Invoice ID Extraction**: Subject line first, fallback to PDF parsing
- **Status Update**: Only update records with status "New"
- **Email Tracking**: Use Gmail labels to mark processed emails
- **Configuration**: Fully configurable email address, refresh token, poll interval

### Existing Infrastructure

- Gmail API integration exists in `pkg/service/googlemail/` for **sending** emails only
- OAuth2 configured with full Gmail access scopes (`gmail.readonly`, `gmail.modify`)
- Contractor Payables service exists: `pkg/service/notion/contractor_payables.go`
- Cron/worker infrastructure: `pkg/worker/worker.go`
- No PDF parsing library currently in use

## Decision

We will implement a **polling-based Gmail inbox listener** with the following architecture:

### 1. Email Monitoring Approach: Gmail API Polling

**Choice**: Use Gmail API `Users.Messages.List` with periodic polling via cron job

**Rationale**:
- **Simplicity**: No additional infrastructure (Pub/Sub, webhooks) required
- **Existing Patterns**: Matches current cron-based worker architecture
- **Credential Reuse**: Leverages existing OAuth2 Google credentials
- **Acceptable Latency**: 5-minute polling interval meets business requirements
- **Lower Complexity**: Easier to debug, test, and maintain

**Rejected Alternative: Push Notifications (Gmail Watch API + Pub/Sub)**
- Requires Google Cloud Pub/Sub setup and maintenance
- Needs webhook endpoint exposed to internet
- Watch expires every 7 days (requires renewal logic)
- Over-engineered for current scale and latency requirements
- Higher operational complexity

### 2. PDF Parsing Library: pdfcpu

**Choice**: Use `github.com/pdfcpu/pdfcpu` for PDF text extraction

**Evaluation Criteria**:
- Pure Go implementation (no CGO dependencies)
- Active maintenance and community support
- Text extraction capabilities
- License compatibility (Apache 2.0)
- Performance for single-page invoice PDFs

**Library Comparison**:

| Library | Pros | Cons | Decision |
|---------|------|------|----------|
| **pdfcpu** | Pure Go, active, Apache 2.0, good text extraction | Larger API surface | **CHOSEN** |
| unidoc/unipdf | Commercial support, feature-rich | Dual license (AGPL/Commercial), cost | Rejected - licensing |
| go-pdf/gopdf | Simple, MIT license | Write-only (no reading) | Rejected - wrong use case |
| ledongthuc/pdf | Simple reader, MIT | Less maintained, basic features | Rejected - limited |

**Rationale for pdfcpu**:
- No CGO means easier builds and deployments
- Apache 2.0 license compatible with commercial use
- Active development (last commit within months)
- Good documentation and examples for text extraction
- Used by other enterprise Go projects

### 3. Service Architecture: Layered Approach

**Structure**:
```
pkg/service/googlemail/     (extend existing)
  - ListInboxMessages()     (new: fetch unread messages)
  - GetMessage()            (new: fetch full message with attachments)
  - AddLabel()              (new: mark as processed)

pkg/service/invoiceemail/   (new service)
  - ProcessInvoiceEmail()   (orchestration)
  - ExtractInvoiceID()      (subject + PDF parsing)
  - UpdatePayableStatus()   (calls Notion service)

pkg/worker/
  - handleInvoiceEmailCheck() (new cron handler)
```

**Rationale**:
- **Separation of Concerns**: Gmail operations separate from business logic
- **Reusability**: Gmail inbox methods can be used for other features
- **Testability**: Each layer can be mocked and tested independently
- **Existing Patterns**: Matches current service architecture (e.g., `notion`, `basecamp`)

### 4. Cron Integration: Existing Worker Pattern

**Choice**: Add new message type to existing worker queue system

**Implementation**:
```go
// pkg/worker/worker.go
const InvoiceEmailCheckMsg = "invoice_email_check"

func (w *Worker) ProcessMessage() error {
    // ... existing switch cases
    case InvoiceEmailCheckMsg:
        _ = w.handleInvoiceEmailCheck(w.logger, message.Payload)
}
```

**Rationale**:
- Consistent with existing worker pattern (Basecamp todos, invoice splits)
- Leverages existing graceful shutdown and error handling
- No new infrastructure required
- Easy to trigger manually for testing

### 5. Error Handling Strategy: Best-Effort with Logging

**Approach**: Continue processing on individual email failures

**Behavior**:
- If one email fails to process, log error and continue with next email
- Track success/failure counts
- Log every operation at DEBUG level with email ID
- Non-fatal errors: invalid Invoice ID, PDF parse failure, already processed
- Fatal errors: Gmail API auth failure, Notion API unavailable

**Rationale**:
- One malformed email shouldn't block processing of valid emails
- Detailed logging enables debugging and manual intervention
- Idempotent operations make retries safe
- Aligns with existing Notion update patterns (best-effort, eventual consistency)

### 6. Email Processing Tracking: Gmail Labels

**Choice**: Use Gmail label "fortress-api/processed" to mark processed emails

**Label Management**:
- Create label on first run if not exists
- Apply label after successful Notion update
- Query excludes labeled messages: `-label:fortress-api/processed`

**Rationale**:
- Native Gmail feature, no external state needed
- Prevents duplicate processing
- Allows manual inspection of processed emails
- Can be reset for reprocessing if needed
- Gmail API supports label creation and application

### 7. Invoice ID Extraction Strategy: Subject-First with PDF Fallback

**Extraction Flow**:
1. **Step 1**: Parse email subject line with regex: `CONTR-\d{6}-[A-Z0-9]+`
2. **Step 2**: If not found, download PDF attachment
3. **Step 3**: Extract text from first page of PDF
4. **Step 4**: Search extracted text with same regex
5. **Step 5**: If still not found, log error and skip email

**Rationale**:
- Subject line parsing is fast and requires no PDF processing
- PDF parsing only when necessary (reduces API quota usage)
- Regex pattern is specific enough to avoid false positives
- First page only (invoices have ID on first page)
- Fail gracefully with clear logging for manual review

## Consequences

### Positive

- **Simple Implementation**: Extends existing services, minimal new infrastructure
- **Maintainable**: Follows established codebase patterns
- **Testable**: Clear separation of concerns enables unit testing
- **Observable**: DEBUG logging provides audit trail
- **Resilient**: Continue-on-error prevents cascading failures
- **Flexible**: Fully configurable via environment variables
- **Idempotent**: Safe to re-run on same emails

### Negative

- **Latency**: 5-minute polling interval (vs real-time push notifications)
- **API Quota**: Higher Gmail API usage than push notifications
- **PDF Dependency**: New external library to maintain
- **Manual Intervention**: Some edge cases may require manual processing
- **No Transactions**: Eventual consistency model (Notion update may fail after label applied)

### Mitigation Strategies

1. **Latency**: 5 minutes is acceptable per business requirements; can reduce interval if needed
2. **API Quota**: Batch requests, implement exponential backoff on rate limits
3. **PDF Parsing**: Robust error handling, detailed logging for parse failures
4. **Edge Cases**: Clear error messages in logs, documented manual process for unusual formats
5. **Consistency**: Idempotent updates allow reprocessing if Notion update fails

## Implementation Notes

### Gmail API Query

```go
// Query unread messages without processed label
query := "is:unread -label:fortress-api/processed has:attachment"
messages, err := service.Users.Messages.List("me").Q(query).Do()
```

### Regex Pattern

```go
// Invoice ID pattern: CONTR-YYYYMM-XXXX
invoiceIDRegex := regexp.MustCompile(`CONTR-\d{6}-[A-Z0-9]+`)
```

### Configuration

```go
type InvoiceEmailListener struct {
    EmailAddress    string        // INVOICE_LISTENER_EMAIL
    RefreshToken    string        // INVOICE_LISTENER_REFRESH_TOKEN
    PollInterval    time.Duration // INVOICE_LISTENER_POLL_INTERVAL
    ProcessedLabel  string        // INVOICE_LISTENER_LABEL
}
```

## Alternatives Considered

### Alternative 1: SendGrid Inbound Parse

**Approach**: Configure MX records to route emails to SendGrid webhook endpoint

**Rejected Because**:
- Requires DNS changes (MX records)
- Different email provider (adds vendor dependency)
- User decision confirmed Gmail API approach
- Less control over email routing

### Alternative 2: IMAP Polling

**Approach**: Use IMAP protocol to poll Gmail inbox

**Rejected Because**:
- Gmail recommends OAuth2 + API over IMAP for apps
- IMAP doesn't support labels as elegantly
- Gmail API provides better error messages and rate limiting
- OAuth2 already configured for Gmail API

### Alternative 3: Store Processing State in Database

**Approach**: Track processed message IDs in fortress-api PostgreSQL database

**Rejected Because**:
- Adds database schema and migrations
- Gmail labels are native and sufficient
- No need for persistent state beyond Gmail
- Increases complexity without clear benefit

## References

- Requirements: `/docs/sessions/202601100842-contractor-invoice-email-listener/requirements/requirements.md`
- Gmail API Docs: https://developers.google.com/gmail/api
- pdfcpu GitHub: https://github.com/pdfcpu/pdfcpu
- Existing Gmail Service: `/pkg/service/googlemail/google_mail.go`
- Contractor Payables Service: `/pkg/service/notion/contractor_payables.go`
- Worker Pattern: `/pkg/worker/worker.go`
