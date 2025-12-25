# Implementation Tasks: Discord Invoice Paid Command with Notion Integration

**Session:** 202512261508-discord-invoice-paid-notion
**Date:** 2025-12-26
**Status:** Ready for Implementation

## Overview

This task breakdown covers implementing the `?inv paid` Discord command rework to support both PostgreSQL and Notion invoice sources.

---

## Phase 1: fortress-api - Notion Service Layer

### Task 1.1: Add Notion Client Invoice Query Method
**File:** `pkg/service/notion/client_invoice.go` (new)
**Effort:** Medium

- [ ] Create `QueryClientInvoiceByNumber(invoiceNumber string) (*nt.Page, error)`
- [ ] Query Notion database `2bf64b29b84c80879a52ed2f9d493096`
- [ ] Filter by `(auto) Invoice Number` title contains
- [ ] Add DEBUG logs for request/response
- [ ] Handle not found case (return nil, nil)

### Task 1.2: Add Notion Invoice Status Update Method
**File:** `pkg/service/notion/client_invoice.go`
**Effort:** Medium

- [ ] Create `UpdateClientInvoiceStatus(pageID string, status string, paidDate time.Time) error`
- [ ] Update `Status` property to given status
- [ ] Set `Paid Date` property to given date
- [ ] Add DEBUG logs
- [ ] Handle Notion API errors

### Task 1.3: Add Notion Invoice Data Extraction Method
**File:** `pkg/service/notion/client_invoice.go`
**Effort:** Medium

- [ ] Create `ExtractClientInvoiceData(page *nt.Page) (*model.Invoice, error)`
- [ ] Extract invoice number from title
- [ ] Extract recipients from `Recipients` rollup
- [ ] Extract Google Drive file info from `Google Drive File` relation
- [ ] Map to `model.Invoice` structure for email/GDrive operations
- [ ] Add DEBUG logs

### Task 1.4: Update Notion Service Interface
**File:** `pkg/service/notion/interface.go`
**Effort:** Small

- [ ] Add new methods to `IService` interface:
  - `QueryClientInvoiceByNumber(invoiceNumber string) (*nt.Page, error)`
  - `UpdateClientInvoiceStatus(pageID string, status string, paidDate time.Time) error`
  - `ExtractClientInvoiceData(page *nt.Page) (*model.Invoice, error)`

---

## Phase 2: fortress-api - Controller Layer

### Task 2.1: Create Mark Paid Result Type
**File:** `pkg/controller/invoice/mark_paid.go` (new)
**Effort:** Small

- [ ] Create `MarkPaidResult` struct:
  ```go
  type MarkPaidResult struct {
      InvoiceNumber   string    `json:"invoice_number"`
      Source          string    `json:"source"` // "postgres", "notion", "both"
      PaidAt          time.Time `json:"paid_at"`
      PostgresUpdated bool      `json:"postgres_updated"`
      NotionUpdated   bool      `json:"notion_updated"`
  }
  ```
- [ ] Create `determineSource(pgFound, notionFound bool) string` helper

### Task 2.2: Implement MarkInvoiceAsPaidByNumber Controller Method
**File:** `pkg/controller/invoice/mark_paid.go`
**Effort:** Large

- [ ] Create `MarkInvoiceAsPaidByNumber(invoiceNumber string) (*MarkPaidResult, error)`
- [ ] Add DEBUG logs at entry point
- [ ] Search PostgreSQL by invoice number
- [ ] Search Notion by invoice number
- [ ] Return error if not found in either
- [ ] Validate PostgreSQL status if found
- [ ] Validate Notion status if found
- [ ] Call existing `MarkInvoiceAsPaidWithTaskRef()` for PostgreSQL
- [ ] Process Notion invoice (status, email, GDrive)
- [ ] Return result with source info

### Task 2.3: Implement Notion Invoice Post-Processing
**File:** `pkg/controller/invoice/mark_paid.go`
**Effort:** Medium

- [ ] Create `processNotionInvoicePaid(page *nt.Page) error`
- [ ] Update Notion status to "Paid" + set Paid Date
- [ ] Extract invoice data for email/GDrive
- [ ] Send thank you email
- [ ] Move PDF in GDrive (Sent → Paid)
- [ ] Add DEBUG logs for each step
- [ ] Handle partial failures gracefully

### Task 2.4: Update Invoice Controller Interface
**File:** `pkg/controller/invoice/interface.go`
**Effort:** Small

- [ ] Add `MarkInvoiceAsPaidByNumber(invoiceNumber string) (*MarkPaidResult, error)` to interface

---

## Phase 3: fortress-api - Handler Layer

### Task 3.1: Create Mark Paid Handler
**File:** `pkg/handler/invoice/mark_paid.go` (new)
**Effort:** Medium

- [ ] Create request struct:
  ```go
  type MarkPaidRequest struct {
      InvoiceNumber string `json:"invoice_number" binding:"required"`
  }
  ```
- [ ] Implement `MarkInvoiceAsPaid(c *gin.Context)` handler
- [ ] Parse and validate request
- [ ] Call controller method
- [ ] Return response with result or error
- [ ] Add DEBUG logs
- [ ] Add Swagger annotations

### Task 3.2: Register Route
**File:** `pkg/routes/v1.go`
**Effort:** Small

- [ ] Add route: `POST /api/v1/invoices/mark-paid`
- [ ] Apply appropriate middleware (auth, permissions)

---

## Phase 4: fortress-discord - Service Layer

### Task 4.1: Update Invoice Service Interface
**File:** `pkg/discord/service/invoice/interface.go`
**Effort:** Small

- [ ] Rename `MarkInvoicePaidViaAccountingTodo` to `MarkInvoicePaid`
- [ ] Update method signature if needed

### Task 4.2: Implement New MarkInvoicePaid Service Method
**File:** `pkg/discord/service/invoice/service.go`
**Effort:** Medium

- [ ] Remove NocoDB dependency
- [ ] Create `MarkInvoicePaid(ctx context.Context, invoiceNumber string) (*MarkPaidResult, error)`
- [ ] Call fortress-api `POST /api/v1/invoices/mark-paid`
- [ ] Parse response
- [ ] Add DEBUG logs

### Task 4.3: Add Fortress Adapter Method
**File:** `pkg/adapter/fortress/invoice.go`
**Effort:** Small

- [ ] Add `MarkInvoicePaid(invoiceNumber string) (*MarkPaidResult, error)`
- [ ] Make HTTP POST request to fortress-api
- [ ] Handle response/errors

---

## Phase 5: fortress-discord - Command Layer

### Task 5.1: Update ExecutePaidConfirmation Method
**File:** `pkg/discord/command/invoice/command.go`
**Effort:** Small

- [ ] Update to call new `MarkInvoicePaid` method
- [ ] Handle new response format (source info)
- [ ] Update success message to include source

### Task 5.2: Update View for Success Message
**File:** `pkg/discord/view/invoice/invoice.go`
**Effort:** Small

- [ ] Update `PaidSuccess` to show source info
- [ ] Format message to indicate where invoice was updated

---

## Phase 6: Cleanup

### Task 6.1: Remove NocoDB Dependencies from fortress-discord
**File:** `pkg/discord/service/invoice/service.go`
**Effort:** Small

- [ ] Remove `nocodb.QueryAccountingTodos()` calls
- [ ] Remove `nocodb.UpdateAccountingTodoStatus()` calls
- [ ] Remove NocoDB imports if no longer needed

---

## Dependency Graph

```
Phase 1 (Notion Service)
    ↓
Phase 2 (Controller)
    ↓
Phase 3 (Handler)
    ↓
Phase 4 (fortress-discord Service)
    ↓
Phase 5 (Command/View)
    ↓
Phase 6 (Cleanup)
```

---

## Estimated Effort

| Phase | Tasks | Effort |
|-------|-------|--------|
| Phase 1 | 4 | Medium |
| Phase 2 | 4 | Large |
| Phase 3 | 2 | Medium |
| Phase 4 | 3 | Medium |
| Phase 5 | 2 | Small |
| Phase 6 | 1 | Small |

---

## Testing Checklist

After implementation, verify:
- [ ] `?inv paid INV-XXX` works for PostgreSQL-only invoice
- [ ] `?inv paid INV-XXX` works for Notion-only invoice
- [ ] `?inv paid INV-XXX` works for invoice in both systems
- [ ] Error shown for non-existent invoice
- [ ] Error shown for invoice with invalid status
- [ ] Thank you email sent for both sources
- [ ] PDF moved in GDrive for both sources
- [ ] Commission/accounting created for PostgreSQL only
