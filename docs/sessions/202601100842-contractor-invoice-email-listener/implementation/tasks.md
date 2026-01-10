# Implementation Tasks - Contractor Invoice Email Listener

## Overview
This document breaks down the implementation into ordered, actionable tasks.

**Session**: `202601100842-contractor-invoice-email-listener`
**Status**: Ready for Implementation

---

## Task 1: Add PDF Parser Dependency

**Priority**: High (Blocker for other tasks)
**Estimated Effort**: Small

### Description
Add the pdfcpu library to go.mod for PDF text extraction.

### Files to Modify
- `go.mod`

### Steps
1. Run `go get github.com/pdfcpu/pdfcpu@v0.5.0`
2. Run `go mod tidy`
3. Verify no conflicts with existing dependencies

### Acceptance Criteria
- [ ] pdfcpu added to go.mod
- [ ] No dependency conflicts
- [ ] `go build` succeeds

---

## Task 2: Add Configuration Struct

**Priority**: High (Blocker for other tasks)
**Estimated Effort**: Small

### Description
Add `InvoiceListener` configuration struct and environment variable loading.

### Files to Modify
- `pkg/config/config.go`

### Steps
1. Add `InvoiceListener` struct with fields:
   - `Enabled` (bool)
   - `Email` (string)
   - `RefreshToken` (string)
   - `PollInterval` (time.Duration)
   - `ProcessedLabel` (string)
   - `MaxMessages` (int)
   - `PDFMaxSizeMB` (int)
2. Add environment variable loading in `generateConfigFromViper()`
3. Add validation for required fields when enabled

### Acceptance Criteria
- [ ] Config struct defined
- [ ] Environment variables loaded
- [ ] Defaults applied for optional fields
- [ ] Validation errors on missing required fields (when enabled)

---

## Task 3: Extend Gmail Service Interface

**Priority**: High
**Estimated Effort**: Medium

### Description
Add inbox reading methods to the Gmail service interface.

### Files to Modify
- `pkg/service/googlemail/interface.go`

### Steps
1. Add new interface methods:
   - `ListInboxMessages(ctx, query, maxResults) ([]Message, error)`
   - `GetMessage(ctx, messageID) (*MessageDetail, error)`
   - `GetAttachment(ctx, messageID, attachmentID) ([]byte, error)`
   - `AddLabel(ctx, messageID, labelID) error`
   - `GetOrCreateLabel(ctx, labelName) (string, error)`
2. Define `Message` and `MessageDetail` types

### Acceptance Criteria
- [ ] Interface methods defined
- [ ] Types for message data defined
- [ ] Existing code still compiles

---

## Task 4: Implement Gmail Inbox Service

**Priority**: High
**Estimated Effort**: Large

### Description
Implement inbox reading functionality in the Gmail service.

### Files to Create
- `pkg/service/googlemail/inbox.go`

### Steps
1. Implement `ListInboxMessages()`:
   - Use Gmail API `Users.Messages.List`
   - Filter by query (e.g., `to:bill@d.foundation is:unread`)
   - Return list of message IDs and metadata
2. Implement `GetMessage()`:
   - Use Gmail API `Users.Messages.Get`
   - Parse headers (From, Subject, Date)
   - Extract attachment metadata
3. Implement `GetAttachment()`:
   - Use Gmail API `Users.Messages.Attachments.Get`
   - Decode base64 content
   - Return raw bytes
4. Implement `AddLabel()`:
   - Use Gmail API `Users.Messages.Modify`
5. Implement `GetOrCreateLabel()`:
   - Use Gmail API `Users.Labels.List` and `Users.Labels.Create`
6. Add DEBUG logging for all operations

### Acceptance Criteria
- [ ] All methods implemented
- [ ] DEBUG logs at every step
- [ ] Error handling with wrapped errors
- [ ] Rate limit handling

---

## Task 5: Create PDF Parser Service

**Priority**: High
**Estimated Effort**: Medium

### Description
Create a service for PDF text extraction.

### Files to Create
- `pkg/service/pdfparser/interface.go`
- `pkg/service/pdfparser/parser.go`

### Steps
1. Define interface:
   - `ExtractText(pdfBytes []byte) (string, error)`
2. Implement using pdfcpu:
   - Parse PDF from bytes
   - Extract text content from all pages
   - Handle errors gracefully (corrupted, encrypted PDFs)
3. Add DEBUG logging

### Acceptance Criteria
- [ ] Interface and implementation created
- [ ] Handles valid PDFs correctly
- [ ] Returns meaningful errors for invalid PDFs
- [ ] DEBUG logs for text extraction

---

## Task 6: Create Invoice ID Extractor

**Priority**: High
**Estimated Effort**: Medium

### Description
Create utility functions for Invoice ID extraction.

### Files to Create
- `pkg/service/invoiceemail/extractor.go`

### Steps
1. Implement `ExtractInvoiceIDFromSubject(subject string) (string, error)`:
   - Use regex pattern `CONTR-\d{6}-[A-Z0-9]+`
   - Return first match or error if not found
2. Implement `ExtractInvoiceIDFromPDF(pdfBytes []byte) (string, error)`:
   - Extract text using PDF parser
   - Search for Invoice ID pattern
   - Return first match or error
3. Implement `ExtractInvoiceID(subject string, pdfBytes []byte) (string, error)`:
   - Try subject first
   - Fallback to PDF if subject fails
4. Add DEBUG logging

### Acceptance Criteria
- [ ] Regex correctly matches Invoice ID pattern
- [ ] Subject extraction works
- [ ] PDF fallback works
- [ ] Clear error messages when not found

---

## Task 7: Add Notion Payable Query Method

**Priority**: High
**Estimated Effort**: Small

### Description
Add method to find payable by Invoice ID.

### Files to Modify
- `pkg/service/notion/contractor_payables.go`

### Steps
1. Add `FindPayableByInvoiceID(ctx, invoiceID) (*ExistingPayable, error)`:
   - Query Contractor Payables database
   - Filter by Invoice ID = invoiceID AND Payment Status = "New"
   - Return page ID and status
2. Add DEBUG logging

### Acceptance Criteria
- [ ] Query correctly filters by Invoice ID
- [ ] Only returns "New" status payables
- [ ] Returns nil if not found (not error)
- [ ] DEBUG logs for query

---

## Task 8: Create Invoice Email Processor Service

**Priority**: High
**Estimated Effort**: Large

### Description
Create the main processor service that orchestrates the workflow.

### Files to Create
- `pkg/service/invoiceemail/interface.go`
- `pkg/service/invoiceemail/processor.go`

### Steps
1. Define interface:
   - `ProcessIncomingInvoices(ctx) (*ProcessResult, error)`
2. Implement processor:
   - Query Gmail for unread emails to target address
   - For each email:
     a. Extract Invoice ID (subject → PDF fallback)
     b. Find matching payable in Notion
     c. Update status to "Pending" if found
     d. Add processed label to email
   - Return summary (processed, failed, skipped counts)
3. Add comprehensive DEBUG logging
4. Handle errors gracefully (continue on individual failures)

### Acceptance Criteria
- [ ] Processes multiple emails in batch
- [ ] Extracts Invoice ID correctly
- [ ] Updates Notion status
- [ ] Marks emails as processed
- [ ] Continues on individual email failures
- [ ] Returns accurate processing summary

---

## Task 9: Create Cron Handler

**Priority**: High
**Estimated Effort**: Small

### Description
Create handler for cron job execution.

### Files to Create
- `pkg/handler/invoiceemail/interface.go`
- `pkg/handler/invoiceemail/handler.go`

### Steps
1. Create handler that wraps the processor service
2. Add entry-point method for cron:
   - `ProcessInvoiceEmails()`
3. Add logging for cron execution start/end
4. Handle panics gracefully

### Acceptance Criteria
- [ ] Handler wraps processor correctly
- [ ] Logs cron execution
- [ ] Panics don't crash the worker

---

## Task 10: Register Handler and Cron Job

**Priority**: High
**Estimated Effort**: Small

### Description
Wire up the handler and register the cron job.

### Files to Modify
- `pkg/handler/handler.go`
- `pkg/worker/worker.go`

### Steps
1. Add `InvoiceEmail` handler to `Handler` struct
2. Initialize handler in handler initialization
3. Register cron job in worker:
   - Schedule based on config poll interval
   - Only register if feature is enabled

### Acceptance Criteria
- [ ] Handler registered in handler struct
- [ ] Cron job registered when enabled
- [ ] Cron job NOT registered when disabled
- [ ] Interval configurable

---

## Task 11: Write Unit Tests

**Priority**: High
**Estimated Effort**: Large

### Description
Implement unit tests as specified in test-cases.

### Files to Create
- `pkg/service/googlemail/inbox_test.go`
- `pkg/service/pdfparser/parser_test.go`
- `pkg/service/invoiceemail/extractor_test.go`
- `pkg/service/invoiceemail/processor_test.go`

### Test Data to Create
- `pkg/service/invoiceemail/testdata/invoice_with_id.pdf`
- `pkg/service/invoiceemail/testdata/invoice_no_id.pdf`

### Steps
1. Follow test specifications in `test-cases/unit/`
2. Create mock implementations for dependencies
3. Create test PDF files
4. Aim for 90%+ coverage

### Acceptance Criteria
- [ ] All specified test cases implemented
- [ ] Tests pass
- [ ] 90%+ code coverage
- [ ] Mock dependencies properly

---

## Task 12: Integration Testing

**Priority**: Medium
**Estimated Effort**: Medium

### Description
Manual integration testing with real Gmail account.

### Steps
1. Set up test environment with valid credentials
2. Create test payable in Notion with known Invoice ID
3. Send test email to target inbox
4. Trigger cron manually or wait for execution
5. Verify:
   - Email processed
   - Payable status updated to "Pending"
   - Email marked with processed label

### Acceptance Criteria
- [ ] End-to-end flow works
- [ ] Status updated correctly
- [ ] Email labeled correctly
- [ ] No duplicate processing on re-run

---

## Task Order Summary

| Order | Task | Dependencies | Priority |
|-------|------|--------------|----------|
| 1 | Add PDF Parser Dependency | None | High |
| 2 | Add Configuration Struct | None | High |
| 3 | Extend Gmail Service Interface | None | High |
| 4 | Implement Gmail Inbox Service | Task 3 | High |
| 5 | Create PDF Parser Service | Task 1 | High |
| 6 | Create Invoice ID Extractor | Task 5 | High |
| 7 | Add Notion Payable Query Method | None | High |
| 8 | Create Invoice Email Processor | Tasks 4, 6, 7 | High |
| 9 | Create Cron Handler | Task 8 | High |
| 10 | Register Handler and Cron Job | Tasks 2, 9 | High |
| 11 | Write Unit Tests | All above | High |
| 12 | Integration Testing | Task 11 | Medium |

---

## Estimated Total Effort

- **Small tasks**: 4 (Tasks 1, 2, 7, 9, 10)
- **Medium tasks**: 3 (Tasks 3, 5, 6, 12)
- **Large tasks**: 3 (Tasks 4, 8, 11)

---

## Next Steps

Use the `proceed` command to begin implementation following this task breakdown.
