# Invoice Email Processor Specification

## Overview

The Invoice Email Processor is a new service that orchestrates the processing of contractor invoice emails. It extracts Invoice IDs from emails/PDFs, matches them with Notion Contractor Payables records, and updates the status from "New" to "Pending".

## Service Architecture

### Package: `pkg/service/invoiceemail/`

### Files Structure

```
pkg/service/invoiceemail/
├── interface.go          (IService interface)
├── processor.go          (ProcessorService implementation)
├── extractor.go          (Invoice ID extraction logic)
├── processor_test.go     (unit tests)
└── extractor_test.go     (extraction tests)
```

## Interface Definition

### File: `pkg/service/invoiceemail/interface.go`

```go
package invoiceemail

import (
    "context"
    "google.golang.org/api/gmail/v1"
)

// IService defines the invoice email processing service
type IService interface {
    // ProcessInvoiceEmails fetches and processes all unread invoice emails
    ProcessInvoiceEmails(ctx context.Context) (*ProcessingResult, error)

    // ProcessSingleEmail processes one email by message ID
    ProcessSingleEmail(ctx context.Context, messageID string) error

    // ExtractInvoiceID extracts Invoice ID from subject or PDF
    ExtractInvoiceID(ctx context.Context, message *gmail.Message) (string, error)
}

// ProcessingResult summarizes batch processing results
type ProcessingResult struct {
    TotalEmails      int       // Total emails processed
    SuccessfulUpdates int      // Payables successfully updated
    SkippedEmails    int       // Emails skipped (no Invoice ID, already processed, etc.)
    FailedUpdates    int       // Failed to update Notion
    Errors           []EmailProcessingError
}

// EmailProcessingError tracks individual email processing errors
type EmailProcessingError struct {
    MessageID   string // Gmail message ID
    Subject     string // Email subject
    InvoiceID   string // Extracted Invoice ID (if found)
    Error       string // Error message
    Step        string // Processing step where error occurred
}
```

## Service Implementation

### File: `pkg/service/invoiceemail/processor.go`

```go
package invoiceemail

import (
    "context"
    "fmt"
    "regexp"

    "github.com/dwarvesf/fortress-api/pkg/config"
    "github.com/dwarvesf/fortress-api/pkg/logger"
    "github.com/dwarvesf/fortress-api/pkg/service/googlemail"
    "github.com/dwarvesf/fortress-api/pkg/service/notion"
)

type ProcessorService struct {
    cfg           *config.Config
    logger        logger.Logger
    gmailService  googlemail.IService
    notionService *notion.ContractorPayablesService
}

// NewProcessorService creates a new invoice email processor
func NewProcessorService(
    cfg *config.Config,
    logger logger.Logger,
    gmailService googlemail.IService,
    notionService *notion.ContractorPayablesService,
) IService {
    return &ProcessorService{
        cfg:           cfg,
        logger:        logger,
        gmailService:  gmailService,
        notionService: notionService,
    }
}
```

### Method: ProcessInvoiceEmails

**Purpose**: Main orchestration method for batch processing

**Implementation**:

```go
func (p *ProcessorService) ProcessInvoiceEmails(ctx context.Context) (*ProcessingResult, error) {
    p.logger.Debug("[DEBUG] invoice_email: starting batch processing")

    result := &ProcessingResult{
        Errors: make([]EmailProcessingError, 0),
    }

    // Step 1: List unread emails without processed label
    query := fmt.Sprintf("is:unread -label:%s has:attachment", p.cfg.InvoiceEmailListener.ProcessedLabel)
    messages, err := p.gmailService.ListInboxMessages(query, 50)
    if err != nil {
        p.logger.Error(err, "[ERROR] invoice_email: failed to list messages")
        return nil, fmt.Errorf("failed to list messages: %w", err)
    }

    p.logger.Debug(fmt.Sprintf("[DEBUG] invoice_email: found %d unread messages", len(messages)))
    result.TotalEmails = len(messages)

    // Step 2: Process each message
    for _, msg := range messages {
        err := p.ProcessSingleEmail(ctx, msg.Id)
        if err != nil {
            result.FailedUpdates++
            // Continue processing other emails
            p.logger.Error(err, fmt.Sprintf("[ERROR] invoice_email: failed to process message=%s", msg.Id))
        } else {
            result.SuccessfulUpdates++
        }
    }

    result.SkippedEmails = result.TotalEmails - result.SuccessfulUpdates - result.FailedUpdates

    p.logger.Info(fmt.Sprintf("[INFO] invoice_email: batch complete total=%d success=%d failed=%d skipped=%d",
        result.TotalEmails, result.SuccessfulUpdates, result.FailedUpdates, result.SkippedEmails))

    return result, nil
}
```

### Method: ProcessSingleEmail

**Purpose**: Process one email end-to-end

**Flow**:
1. Fetch full message
2. Extract Invoice ID
3. Find Notion payable record
4. Check status is "New"
5. Update status to "Pending"
6. Mark email as processed

**Implementation**:

```go
func (p *ProcessorService) ProcessSingleEmail(ctx context.Context, messageID string) error {
    p.logger.Debug(fmt.Sprintf("[DEBUG] invoice_email: processing message=%s", messageID))

    // Step 1: Get full message
    message, err := p.gmailService.GetMessage(messageID)
    if err != nil {
        return fmt.Errorf("failed to get message: %w", err)
    }

    // Step 2: Parse headers
    headers := googlemail.ParseEmailHeaders(message)
    subject := headers["subject"]
    from := headers["from"]
    p.logger.Debug(fmt.Sprintf("[DEBUG] invoice_email: message=%s subject='%s' from='%s'", messageID, subject, from))

    // Step 3: Extract Invoice ID
    invoiceID, err := p.ExtractInvoiceID(ctx, message)
    if err != nil {
        p.logger.Error(err, fmt.Sprintf("[ERROR] invoice_email: failed to extract Invoice ID from message=%s", messageID))
        // Mark as processed to avoid reprocessing
        _ = p.gmailService.AddLabel(messageID, p.cfg.InvoiceEmailListener.ProcessedLabel)
        return fmt.Errorf("failed to extract Invoice ID: %w", err)
    }

    p.logger.Debug(fmt.Sprintf("[DEBUG] invoice_email: extracted invoiceID=%s from message=%s", invoiceID, messageID))

    // Step 4: Find payable in Notion
    payable, err := p.findPayableByInvoiceID(ctx, invoiceID)
    if err != nil {
        p.logger.Error(err, fmt.Sprintf("[ERROR] invoice_email: failed to find payable invoiceID=%s", invoiceID))
        _ = p.gmailService.AddLabel(messageID, p.cfg.InvoiceEmailListener.ProcessedLabel)
        return fmt.Errorf("failed to find payable: %w", err)
    }

    if payable == nil {
        p.logger.Debug(fmt.Sprintf("[DEBUG] invoice_email: no payable found for invoiceID=%s, skipping", invoiceID))
        _ = p.gmailService.AddLabel(messageID, p.cfg.InvoiceEmailListener.ProcessedLabel)
        return nil
    }

    // Step 5: Check status is "New"
    if payable.Status != "New" {
        p.logger.Debug(fmt.Sprintf("[DEBUG] invoice_email: payable status=%s (not New), skipping invoiceID=%s", payable.Status, invoiceID))
        _ = p.gmailService.AddLabel(messageID, p.cfg.InvoiceEmailListener.ProcessedLabel)
        return nil
    }

    // Step 6: Update status to "Pending"
    err = p.notionService.UpdatePayableStatus(ctx, payable.PageID, "Pending", "")
    if err != nil {
        p.logger.Error(err, fmt.Sprintf("[ERROR] invoice_email: failed to update payable pageID=%s invoiceID=%s", payable.PageID, invoiceID))
        return fmt.Errorf("failed to update payable status: %w", err)
    }

    p.logger.Info(fmt.Sprintf("[INFO] invoice_email: updated payable pageID=%s invoiceID=%s status=New→Pending", payable.PageID, invoiceID))

    // Step 7: Mark email as processed
    err = p.gmailService.AddLabel(messageID, p.cfg.InvoiceEmailListener.ProcessedLabel)
    if err != nil {
        p.logger.Error(err, fmt.Sprintf("[ERROR] invoice_email: failed to add label to message=%s", messageID))
        // Non-fatal: status already updated, just log error
    }

    return nil
}
```

### Method: findPayableByInvoiceID

**Purpose**: Query Notion for payable record by Invoice ID

**Implementation**:

```go
func (p *ProcessorService) findPayableByInvoiceID(ctx context.Context, invoiceID string) (*PayableRecord, error) {
    // Query Notion Contractor Payables database
    // Filter: Invoice ID (rich text) equals invoiceID
    filter := &nt.DatabaseQueryFilter{
        Property: "Invoice ID",
        DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
            RichText: &nt.TextPropertyFilter{
                Equals: invoiceID,
            },
        },
    }

    resp, err := p.notionService.client.QueryDatabase(ctx, p.cfg.Notion.Databases.ContractorPayables, &nt.DatabaseQuery{
        Filter:   filter,
        PageSize: 1,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to query Notion: %w", err)
    }

    if len(resp.Results) == 0 {
        return nil, nil // Not found
    }

    page := resp.Results[0]
    props, ok := page.Properties.(nt.DatabasePageProperties)
    if !ok {
        return nil, errors.New("failed to parse page properties")
    }

    // Extract status
    status := ""
    if statusProp, exists := props["Payment Status"]; exists && statusProp.Status != nil {
        status = statusProp.Status.Name
    }

    return &PayableRecord{
        PageID:    page.ID,
        InvoiceID: invoiceID,
        Status:    status,
    }, nil
}

// PayableRecord represents a Contractor Payable from Notion
type PayableRecord struct {
    PageID    string
    InvoiceID string
    Status    string
}
```

## Invoice ID Extraction

### File: `pkg/service/invoiceemail/extractor.go`

### Method: ExtractInvoiceID

**Purpose**: Extract Invoice ID from subject or PDF

**Strategy**:
1. Try subject line first (fast path)
2. Fallback to PDF parsing (slow path)
3. Return error if not found in either

**Implementation**:

```go
package invoiceemail

import (
    "context"
    "errors"
    "fmt"
    "regexp"

    "github.com/pdfcpu/pdfcpu/pkg/api"
    "google.golang.org/api/gmail/v1"
)

var invoiceIDRegex = regexp.MustCompile(`CONTR-\d{6}-[A-Z0-9]+`)

func (p *ProcessorService) ExtractInvoiceID(ctx context.Context, message *gmail.Message) (string, error) {
    // Step 1: Try subject line
    headers := googlemail.ParseEmailHeaders(message)
    subject := headers["subject"]

    invoiceID := extractFromText(subject)
    if invoiceID != "" {
        p.logger.Debug(fmt.Sprintf("[DEBUG] invoice_email: found invoiceID=%s in subject", invoiceID))
        return invoiceID, nil
    }

    p.logger.Debug("[DEBUG] invoice_email: Invoice ID not in subject, trying PDF")

    // Step 2: Find PDF attachment
    attachmentID, filename, found := googlemail.FindPDFAttachment(message)
    if !found {
        return "", errors.New("no PDF attachment found")
    }

    p.logger.Debug(fmt.Sprintf("[DEBUG] invoice_email: found PDF attachment file=%s", filename))

    // Step 3: Download PDF
    pdfBytes, err := p.gmailService.GetAttachment(message.Id, attachmentID)
    if err != nil {
        return "", fmt.Errorf("failed to download PDF: %w", err)
    }

    p.logger.Debug(fmt.Sprintf("[DEBUG] invoice_email: downloaded PDF size=%d bytes", len(pdfBytes)))

    // Step 4: Extract text from PDF
    text, err := extractTextFromPDF(pdfBytes)
    if err != nil {
        return "", fmt.Errorf("failed to extract text from PDF: %w", err)
    }

    // Step 5: Search text for Invoice ID
    invoiceID = extractFromText(text)
    if invoiceID == "" {
        return "", errors.New("Invoice ID not found in subject or PDF")
    }

    p.logger.Debug(fmt.Sprintf("[DEBUG] invoice_email: found invoiceID=%s in PDF", invoiceID))
    return invoiceID, nil
}

// extractFromText extracts Invoice ID from text using regex
func extractFromText(text string) string {
    matches := invoiceIDRegex.FindStringSubmatch(text)
    if len(matches) > 0 {
        return matches[0]
    }
    return ""
}

// extractTextFromPDF extracts text from PDF bytes (first page only)
func extractTextFromPDF(pdfBytes []byte) (string, error) {
    // Write bytes to temp file (pdfcpu works with files)
    tmpFile, err := os.CreateTemp("", "invoice-*.pdf")
    if err != nil {
        return "", fmt.Errorf("failed to create temp file: %w", err)
    }
    defer os.Remove(tmpFile.Name())
    defer tmpFile.Close()

    if _, err := tmpFile.Write(pdfBytes); err != nil {
        return "", fmt.Errorf("failed to write PDF to temp file: %w", err)
    }

    // Extract text from first page
    text, err := api.ExtractText(tmpFile.Name(), []string{"1"}, nil)
    if err != nil {
        return "", fmt.Errorf("pdfcpu extraction failed: %w", err)
    }

    return text, nil
}
```

### Regex Pattern

**Invoice ID Format**: `CONTR-YYYYMM-XXXX`

**Examples**:
- `CONTR-202501-A1B2`
- `CONTR-202412-XYZ9`
- `CONTR-202503-1234`

**Regex**: `CONTR-\d{6}-[A-Z0-9]+`

**Explanation**:
- `CONTR-`: Literal prefix
- `\d{6}`: Exactly 6 digits (YYYYMM)
- `-`: Literal dash
- `[A-Z0-9]+`: One or more uppercase letters or digits

## Error Handling Strategy

### Error Categories

1. **Recoverable Errors (Log and Skip)**:
   - No PDF attachment found
   - Invoice ID not found in subject or PDF
   - Payable record not found in Notion
   - Payable status is not "New" (already processed)
   - **Action**: Log warning, mark email as processed, continue

2. **Non-Recoverable Errors (Log and Retry)**:
   - Gmail API authentication failure
   - Notion API authentication failure
   - Network timeout
   - **Action**: Log error, don't mark as processed, retry next run

3. **Partial Failures (Log and Continue)**:
   - Notion update succeeds but Gmail label fails
   - **Action**: Log error, status update succeeded (acceptable)

### Logging Standards

**Entry Logging**:
```go
logger.Debug(fmt.Sprintf("[DEBUG] invoice_email: processing message=%s", messageID))
```

**Success Logging**:
```go
logger.Info(fmt.Sprintf("[INFO] invoice_email: updated payable pageID=%s invoiceID=%s status=New→Pending", pageID, invoiceID))
```

**Error Logging**:
```go
logger.Error(err, fmt.Sprintf("[ERROR] invoice_email: failed to extract Invoice ID from message=%s", messageID))
```

**Summary Logging**:
```go
logger.Info(fmt.Sprintf("[INFO] invoice_email: batch complete total=%d success=%d failed=%d", total, success, failed))
```

## Service Integration

### File: `pkg/service/service.go`

Add to Service struct:

```go
type Service struct {
    // ... existing services
    InvoiceEmail invoiceemail.IService
}

func New(cfg *config.Config, logger logger.Logger) (*Service, error) {
    // ... existing initializations

    // Initialize Invoice Email Processor
    invoiceEmailService := invoiceemail.NewProcessorService(
        cfg,
        logger,
        googleMailService,
        contractorPayablesService,
    )

    return &Service{
        // ... existing services
        InvoiceEmail: invoiceEmailService,
    }, nil
}
```

## Worker Integration

### File: `pkg/worker/worker.go`

Add message type and handler:

```go
const InvoiceEmailCheckMsg = "invoice_email_check"

func (w *Worker) ProcessMessage() error {
    // ... existing switch cases
    case InvoiceEmailCheckMsg:
        _ = w.handleInvoiceEmailCheck(w.logger, message.Payload)
}

func (w *Worker) handleInvoiceEmailCheck(l logger.Logger, payload interface{}) error {
    l = l.Fields(logger.Fields{"worker": "invoiceEmailCheck"})
    l.Info("processing invoice email check")

    result, err := w.service.InvoiceEmail.ProcessInvoiceEmails(w.ctx)
    if err != nil {
        l.Error(err, "failed to process invoice emails")
        return err
    }

    l.Info(fmt.Sprintf("invoice email check complete: total=%d success=%d failed=%d skipped=%d",
        result.TotalEmails, result.SuccessfulUpdates, result.FailedUpdates, result.SkippedEmails))

    return nil
}
```

## Cron Job Setup

### Trigger via Cron

Add cron job in deployment environment:

```bash
# Every 5 minutes
*/5 * * * * /usr/local/bin/trigger-worker-job invoice_email_check
```

Or use built-in Go cron library (e.g., `robfig/cron`):

```go
// pkg/worker/cron.go
func (w *Worker) StartCron() {
    c := cron.New()
    c.AddFunc("*/5 * * * *", func() {
        w.Enqueue(InvoiceEmailCheckMsg, nil)
    })
    c.Start()
}
```

## Testing Strategy

### Unit Tests

**File**: `pkg/service/invoiceemail/processor_test.go`

**Test Cases**:
1. `TestProcessSingleEmail_Success` - Full flow succeeds
2. `TestProcessSingleEmail_NoInvoiceID` - Invoice ID not found
3. `TestProcessSingleEmail_PayableNotFound` - No matching Notion record
4. `TestProcessSingleEmail_AlreadyPending` - Status not "New"
5. `TestProcessSingleEmail_NotionUpdateFails` - Notion API error
6. `TestProcessInvoiceEmails_BatchSuccess` - Process multiple emails
7. `TestProcessInvoiceEmails_PartialFailure` - Some emails fail

**File**: `pkg/service/invoiceemail/extractor_test.go`

**Test Cases**:
1. `TestExtractInvoiceID_FromSubject` - Found in subject
2. `TestExtractInvoiceID_FromPDF` - Found in PDF
3. `TestExtractInvoiceID_NotFound` - Not in subject or PDF
4. `TestExtractFromText_ValidPattern` - Regex matches
5. `TestExtractFromText_InvalidPattern` - Regex doesn't match
6. `TestExtractTextFromPDF_Success` - PDF parsing works
7. `TestExtractTextFromPDF_InvalidPDF` - Malformed PDF

### Test Data

**Sample Subject Lines**:
- `Invoice CONTR-202501-A1B2 for January 2025`
- `Contractor Invoice - CONTR-202412-XYZ9`
- `December Invoice` (no Invoice ID - should fallback to PDF)

**Sample PDF Content**:
```
CONTRACTOR INVOICE

Invoice ID: CONTR-202501-A1B2
Date: January 15, 2025
Contractor: John Doe

Total: $5,000.00
```

## Dependencies

### New Dependencies

Add to `go.mod`:

```
github.com/pdfcpu/pdfcpu v0.5.0
```

### Existing Dependencies

- `google.golang.org/api/gmail/v1` - Gmail API
- `github.com/dstotijn/go-notion` - Notion API

## Performance Considerations

1. **Batch Size**: Limit to 50 messages per run (configurable)
2. **Timeout**: Set context timeout for Notion API calls (30 seconds)
3. **PDF Size**: Limit PDF download to first 5MB (configurable)
4. **Text Extraction**: Only extract first page of PDF (Invoice ID always on page 1)
5. **Concurrent Processing**: Sequential for now, can parallelize later if needed

## Security Considerations

1. **Input Validation**: Validate messageID format before API calls
2. **PDF Safety**: Use pdfcpu (pure Go, no system calls) to avoid shell injection
3. **Temp Files**: Clean up temp PDF files after processing
4. **Error Messages**: Don't log full email content (privacy)
5. **Rate Limiting**: Respect Gmail API quotas with exponential backoff

## References

- Requirements: `/docs/sessions/202601100842-contractor-invoice-email-listener/requirements/requirements.md`
- ADR: `/docs/sessions/202601100842-contractor-invoice-email-listener/planning/ADRs/001-email-listener-architecture.md`
- Gmail Service Spec: `/docs/sessions/202601100842-contractor-invoice-email-listener/planning/specifications/gmail-inbox-service.md`
- Contractor Payables Service: `/pkg/service/notion/contractor_payables.go`
- pdfcpu Documentation: https://pdfcpu.io/
