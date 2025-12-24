# Contractor Invoice Generation - Implementation Tasks

**Session**: 202512241732-contractor-invoice-generation
**Specification**: `docs/specs/notion/contractor-invoice-generation.md`
**Status**: Not Started

---

## Task Overview

| Task | Status | Priority |
|------|--------|----------|
| 1. Configuration Updates | [x] | P0 |
| 2. Contractor Rates Service | [x] | P0 |
| 3. Task Order Log Extensions | [x] | P0 |
| 4. Request/Response Models | [x] | P0 |
| 5. Controller Logic | [x] | P0 |
| 6. HTML Template Updates | [x] | P1 |
| 7. Handler Implementation | [ ] | P0 |
| 8. Route Registration | [ ] | P0 |
| 9. Manual Testing | [ ] | P1 |

---

## Task 1: Configuration Updates

**File**: `pkg/config/config.go`
**Priority**: P0
**Estimated Time**: 15 min

### Subtasks
- [x] 1.1 Add `ContractorRates` field to `NotionDatabase` struct
- [x] 1.2 Add env loading for `NOTION_CONTRACTOR_RATES_DB_ID`
- [x] 1.3 Add `ContractorInvoiceDirID` field to `Invoice` struct
- [x] 1.4 Add env loading for `CONTRACTOR_INVOICE_DIR_ID`
- [x] 1.5 Update `.env.sample` with new variables

### Code Changes
```go
// In NotionDatabase struct
ContractorRates string

// In Invoice struct
ContractorInvoiceDirID string
```

### Environment Variables
```bash
NOTION_CONTRACTOR_RATES_DB_ID=2c464b29b84c805bbcdedc052e613f4d
CONTRACTOR_INVOICE_DIR_ID=1_9Ai9erlvc39vDMoYCswroQxWzajKUX2
```

---

## Task 2: Contractor Rates Service

**File**: `pkg/service/notion/contractor_rates.go` (NEW)
**Priority**: P0
**Estimated Time**: 2 hours

### Subtasks
- [ ] 2.1 Create `ContractorRatesService` struct
- [ ] 2.2 Implement `NewContractorRatesService()` constructor
- [ ] 2.3 Create `ContractorRateData` struct with fields:
  - ContractorPageID, Discord, BillingType
  - MonthlyFixed, HourlyRate, GrossFixed
  - Currency, StartDate, EndDate
- [ ] 2.4 Implement `QueryRatesByDiscordAndMonth()` method:
  - Build Notion query filters (Discord rollup, Status=Active, date range)
  - Execute paginated query
  - Extract properties from response
- [ ] 2.5 Add helper functions:
  - `extractRollupString()` - Extract string from rollup array
  - `extractFormulaNumber()` - Extract number from formula property
  - `extractNumber()` - Extract number property
  - `extractSelect()` - Extract select option name
  - `extractDate()` - Extract date property
- [ ] 2.6 Add comprehensive DEBUG logging

### Dependencies
- Task 1 (config with database ID)

---

## Task 3: Task Order Log Extensions

**File**: `pkg/service/notion/task_order_log.go`
**Priority**: P0
**Estimated Time**: 1 hour

### Subtasks
- [ ] 3.1 Create `OrderSubitem` struct:
  - PageID, ProjectName, ProjectID
  - Hours, ProofOfWork
- [ ] 3.2 Implement `QueryOrderSubitems()` method:
  - Filter by Type="Timesheet" and Parent item=orderPageID
  - Extract Line Item Hours, Proof of Works, Project
  - Handle pagination
- [ ] 3.3 Add helper functions:
  - `extractRichText()` - Extract from rich_text property
  - `extractRollupRelation()` - Extract relation from rollup
  - `extractTitle()` - Extract title property
- [ ] 3.4 Add DEBUG logging

### Dependencies
- None (extends existing service)

---

## Task 4: Request/Response Models

**Files**:
- `pkg/handler/invoice/request/contractor_invoice.go` (NEW)
- `pkg/view/contractor_invoice.go` (NEW)
- `pkg/handler/invoice/errs/errors.go` (NEW or MODIFY)

**Priority**: P0
**Estimated Time**: 30 min

### Subtasks
- [ ] 4.1 Create request model:
  ```go
  type GenerateContractorInvoiceRequest struct {
      ContractorDiscord string `json:"contractorDiscord" binding:"required"`
      Month             string `json:"month" binding:"required"`
  }
  ```
- [ ] 4.2 Create response models:
  ```go
  type ContractorInvoiceResponse struct {
      InvoiceNumber  string  `json:"invoiceNumber"`
      ContractorName string  `json:"contractorName"`
      Month          string  `json:"month"`
      BillingType    string  `json:"billingType"`
      Currency       string  `json:"currency"`
      Total          float64 `json:"total"`
      PDFFileURL     string  `json:"pdfFileUrl"`
      GeneratedAt    string  `json:"generatedAt"`
  }

  type ContractorInvoiceLineItem struct {
      ProjectName string  `json:"projectName"`
      Description string  `json:"description"`
      Hours       float64 `json:"hours,omitempty"`
      Rate        float64 `json:"rate,omitempty"`
      Amount      float64 `json:"amount,omitempty"`
  }
  ```
- [ ] 4.3 Create error definitions:
  - `ErrContractorRatesNotFound`
  - `ErrTaskOrderLogNotFound`
  - `ErrInvalidMonthFormat`
  - `ErrUnsupportedBillingType`
  - `ErrNoSubitemsFound`

### Dependencies
- None

---

## Task 5: Controller Logic

**File**: `pkg/controller/invoice/contractor_invoice.go` (NEW)
**Priority**: P0
**Estimated Time**: 3 hours

### Subtasks
- [ ] 5.1 Create `ContractorInvoiceData` internal struct
- [ ] 5.2 Implement `GenerateContractorInvoice()` method:
  - Query Contractor Rates from Notion
  - Query Task Order Log for month
  - Query Order Subitems
  - Build line items based on billing type (Monthly Fixed vs Hourly Rate)
  - Generate invoice number (CONTR-{YYYYMM}-{random-4-chars})
  - Calculate dates (invoice date, due date)
- [ ] 5.3 Implement `generateRandomAlphanumeric()` helper
- [ ] 5.4 Implement `GenerateContractorInvoicePDF()` method:
  - Setup currency formatter (go-money)
  - Create template FuncMap (formatMoney, formatDate, isMonthlyFixed, isHourlyRate, add)
  - Parse contractor-invoice-template.html
  - Execute template with data
  - Convert HTML to PDF using wkhtmltopdf
- [ ] 5.5 Add comprehensive DEBUG logging at each step
- [ ] 5.6 Add interface method to `IController`

### Dependencies
- Task 2 (Contractor Rates Service)
- Task 3 (Task Order Log Extensions)
- Task 4 (Models)

---

## Task 6: HTML Template Updates

**File**: `pkg/templates/contractor-invoice-template.html`
**Priority**: P1
**Estimated Time**: 1 hour

### Subtasks
- [ ] 6.1 Move template from project root to `pkg/templates/`
- [ ] 6.2 Add Go template variables:
  - `{{.Invoice.InvoiceNumber}}`
  - `{{.Invoice.ContractorName}}`
  - `{{.Invoice.Date | formatDate}}`
  - `{{.Invoice.DueDate | formatDate}}`
  - `{{.Invoice.Total | formatMoney}}`
  - `{{.Invoice.Currency}}`
- [ ] 6.3 Add conditional table headers:
  ```html
  {{if isHourlyRate}}
  <th>HOURS</th>
  <th>RATE</th>
  <th>TOTAL</th>
  {{end}}
  ```
- [ ] 6.4 Add conditional table rows:
  ```html
  {{range $index, $item := .LineItems}}
  <tr>
      <td>{{add $index 1}}</td>
      <td>{{$item.ProjectName}}</td>
      <td>{{$item.Description}}</td>
      {{if isHourlyRate}}
      <td>{{$item.Hours}}</td>
      <td>{{formatMoney $item.Rate}}</td>
      <td>{{formatMoney $item.Amount}}</td>
      {{end}}
  </tr>
  {{end}}
  ```
- [ ] 6.5 Test template rendering locally

### Dependencies
- Task 5 (Controller needs template)

---

## Task 7: Handler Implementation

**File**: `pkg/handler/invoice/contractor_invoice.go` (NEW)
**Priority**: P0
**Estimated Time**: 1.5 hours

### Subtasks
- [ ] 7.1 Implement `GenerateContractorInvoice()` handler:
  - Parse JSON request body
  - Validate month format (regex: `^\d{4}-\d{2}$`)
  - Call controller.GenerateContractorInvoice()
  - Call controller.GenerateContractorInvoicePDF()
  - Upload PDF to Google Drive (use existing service)
  - Build and return response
- [ ] 7.2 Add `isValidMonthFormat()` helper function
- [ ] 7.3 Add DEBUG logging for request/response
- [ ] 7.4 Add interface method to `IHandler`
- [ ] 7.5 Handle all error cases with proper HTTP status codes:
  - 400: Invalid month format, missing params
  - 404: Contractor rates not found, task order log not found
  - 500: Notion query failed, PDF generation failed, upload failed

### Dependencies
- Task 5 (Controller Logic)
- Task 4 (Request/Response Models)

---

## Task 8: Route Registration

**File**: `pkg/routes/v1.go`
**Priority**: P0
**Estimated Time**: 15 min

### Subtasks
- [ ] 8.1 Add route to invoice group:
  ```go
  invoiceRoute.POST("/contractor/generate",
      conditionalAuthMW,
      conditionalPermMW(model.PermissionInvoicesCreate),
      h.Invoice.GenerateContractorInvoice)
  ```
- [ ] 8.2 Verify middleware configuration

### Dependencies
- Task 7 (Handler Implementation)

---

## Task 9: Manual Testing

**Priority**: P1
**Estimated Time**: 1 hour

### Subtasks
- [ ] 9.1 Test Monthly Fixed billing (adeki_, 2025-12)
  ```bash
  curl -X POST http://localhost:8080/api/v1/invoices/contractor/generate \
    -H "Content-Type: application/json" \
    -d '{"contractorDiscord": "adeki_", "month": "2025-12"}'
  ```
- [ ] 9.2 Test Hourly Rate billing (nhidesign9, 2025-11)
- [ ] 9.3 Test error cases:
  - Invalid month format
  - Non-existent contractor
  - Missing required fields
- [ ] 9.4 Verify PDF content and formatting
- [ ] 9.5 Verify Google Drive upload

### Dependencies
- All previous tasks

---

## Execution Order

```
Task 1 (Config) ─────────────────────────────────────┐
                                                      │
Task 4 (Models) ─────────────────────────────────────┤
                                                      │
Task 2 (Contractor Rates Service) ───────────────────┼──► Task 5 (Controller)
                                                      │         │
Task 3 (Task Order Log Extensions) ──────────────────┤         │
                                                      │         ▼
                                         Task 6 (Template) ──► Task 7 (Handler)
                                                                    │
                                                                    ▼
                                                              Task 8 (Routes)
                                                                    │
                                                                    ▼
                                                              Task 9 (Testing)
```

---

## Notes

- **Reference Spec**: `docs/specs/notion/contractor-invoice-generation.md`
- **Invoice Number Format**: `CONTR-{YYYYMM}-{random-4-chars}` (e.g., CONTR-202512-A7K9)
- **Supported Billing Types**: Monthly Fixed, Hourly Rate
- **Currencies**: VND (no decimals), USD (2 decimals)
- **PDF Tool**: wkhtmltopdf
- **Google Drive Storage**:
  - Parent folder ID: Read from env `CONTRACTOR_INVOICE_DIR_ID`
  - Subfolder: `{contractor_full_name}/`
  - File naming: `{invoice_number}.pdf` (e.g., `CONTR-202512-A7K9.pdf`)
- **All code must include DEBUG logging**

---

## Completion Checklist

- [ ] All 9 tasks completed
- [ ] DEBUG logs added to all new code
- [ ] Manual testing passed
- [ ] PDF generated correctly for both billing types
- [ ] Google Drive upload working
- [ ] Error handling verified
