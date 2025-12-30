# Contractor Payout Data Preparation - Implementation Tasks

**Session**: 202512301554-contractor-payout-data
**Specification**: `docs/specs/notion/contractor-invoice.md`
**Status**: ✅ COMPLETE

---

## Task Overview

| Task | Status | Priority |
|------|--------|----------|
| 1. Configuration Updates | [x] | P0 |
| 2. ContractorPayoutsService | [x] | P0 |
| 3. ContractorFeesService | [x] | P0 |
| 4. InvoiceSplitService | [x] | P0 |
| 5. Data Models | [x] | P0 |

---

## Task 1: Configuration Updates

**File**: `pkg/config/config.go`
**Priority**: P0
**Status**: ✅ COMPLETE

### Subtasks
- [x] 1.1 Add `ContractorPayouts` field to `NotionDatabase` struct
- [x] 1.2 Add `ContractorFees` field to `NotionDatabase` struct
- [x] 1.3 Add `InvoiceSplit` field to `NotionDatabase` struct
- [x] 1.4 Add `RefundRequest` field to `NotionDatabase` struct
- [x] 1.5 Add env loading for all new database IDs
- [x] 1.6 Update `.env.sample` with new variables

### Code Changes
```go
// In NotionDatabase struct
ContractorPayouts string
ContractorFees    string
InvoiceSplit      string
RefundRequest     string
```

### Environment Variables
```bash
NOTION_CONTRACTOR_PAYOUTS_DB_ID=2c564b29-b84c-8045-80ee-000bee2e3669
NOTION_CONTRACTOR_FEES_DB_ID=2c264b29-b84c-8037-807c-000bf6d0792c
NOTION_INVOICE_SPLIT_DB_ID=2c364b29-b84c-804f-9856-000b58702dea
NOTION_REFUND_REQUEST_DB_ID=2cc64b29-b84c-8066-adf2-cc56171cedf4
```

---

## Task 2: ContractorPayoutsService

**File**: `pkg/service/notion/contractor_payouts.go` (NEW)
**Priority**: P0
**Status**: ✅ COMPLETE

### Subtasks
- [x] 2.1 Create `ContractorPayoutsService` struct
- [x] 2.2 Implement `NewContractorPayoutsService()` constructor
- [x] 2.3 Create `PayoutEntry` struct with fields:
  - PageID
  - PersonPageID (from Person relation)
  - SourceType (Contractor Payroll, Commission, Refund, Other)
  - Direction (Outgoing, Incoming)
  - Amount
  - Currency
  - Status
  - ContractorFeesID (from Contractor Fees relation)
  - InvoiceSplitID (from Invoice Split relation)
  - RefundRequestID (from Refund Request relation)
- [x] 2.4 Implement `QueryPendingPayoutsByContractor(ctx, contractorPageID)`:
  - Build filter: Person relation contains contractorPageID AND Status=Pending
  - Execute paginated query
  - Extract properties including relations
  - Determine SourceType based on which relation is set
- [x] 2.5 Add helper functions for property extraction
- [x] 2.6 Add comprehensive DEBUG logging

### Notion Query Filter
```javascript
{
  "filter": {
    "and": [
      {
        "property": "Person",
        "relation": { "contains": "<contractor_page_id>" }
      },
      {
        "property": "Status",
        "status": { "equals": "Pending" }
      }
    ]
  }
}
```

### Dependencies
- Task 1 (config with database ID)

---

## Task 3: ContractorFeesService

**File**: `pkg/service/notion/contractor_fees.go` (NEW)
**Priority**: P0
**Status**: ✅ COMPLETE

### Subtasks
- [x] 3.1 Create `ContractorFeesService` struct
- [x] 3.2 Implement `NewContractorFeesService()` constructor
- [x] 3.3 Create `ContractorFeesData` struct with fields:
  - PageID
  - TotalHoursWorked (rollup from Task Order Log)
  - HourlyRate (rollup from Contractor Rate)
  - FixedFee (rollup from Contractor Rate)
  - BillingType (rollup from Contractor Rate)
  - ProofOfWorks (rollup from Task Order Log)
  - TotalAmount (formula)
  - Currency
- [x] 3.4 Implement `GetContractorFeesByID(ctx, feesPageID)`:
  - Fetch page by ID
  - Extract rollup properties
  - Extract formula properties
- [x] 3.5 Add helper functions:
  - `extractRollupNumber()` - Extract number from rollup
  - `extractRollupRichText()` - Extract rich text from rollup
  - `extractFormulaNumber()` - Extract number from formula
- [x] 3.6 Add DEBUG logging

### Dependencies
- Task 1 (config with database ID)

---

## Task 4: InvoiceSplitService

**File**: `pkg/service/notion/invoice_split.go` (NEW)
**Priority**: P0
**Status**: ✅ COMPLETE

### Subtasks
- [x] 4.1 Create `InvoiceSplitService` struct
- [x] 4.2 Implement `NewInvoiceSplitService()` constructor
- [x] 4.3 Create `InvoiceSplitData` struct with fields:
  - PageID
  - Amount
  - Role (Sales, Account Manager, etc.)
  - Currency
- [x] 4.4 Implement `GetInvoiceSplitByID(ctx, splitPageID)`:
  - Fetch page by ID
  - Extract Amount (number)
  - Extract Role (select)
  - Extract Currency (select)
- [x] 4.5 Add DEBUG logging

### Dependencies
- Task 1 (config with database ID)

---

## Task 5: Data Models

**File**: `pkg/service/notion/payout_types.go` (NEW)
**Priority**: P0
**Status**: ✅ COMPLETE

### Subtasks
- [x] 5.1 Create `PayoutSourceType` enum:
  ```go
  const (
      PayoutSourceTypeContractorPayroll PayoutSourceType = "Contractor Payroll"
      PayoutSourceTypeCommission        PayoutSourceType = "Commission"
      PayoutSourceTypeRefund            PayoutSourceType = "Refund"
      PayoutSourceTypeOther             PayoutSourceType = "Other"
  )
  ```
- [x] 5.2 Create `PayoutDirection` enum:
  ```go
  const (
      PayoutDirectionOutgoing PayoutDirection = "Outgoing (you pay)"
      PayoutDirectionIncoming PayoutDirection = "Incoming (you receive)"
  )
  ```
- [x] 5.3 Create unified `PayoutLineItem` struct for invoice generation:
  ```go
  type PayoutLineItem struct {
      SourceType  PayoutSourceType
      Direction   PayoutDirection
      Title       string
      Description string
      Hours       float64 // Contractor Payroll only
      Rate        float64 // Contractor Payroll only
      Amount      float64
      AmountUSD   float64
      Currency    string
  }
  ```

### Dependencies
- None

---

## Execution Order

```
Task 1 (Config) ─────────────────────────────────────┐
                                                      │
Task 5 (Data Models) ────────────────────────────────┤
                                                      │
Task 2 (ContractorPayoutsService) ───────────────────┤
                                                      │
Task 3 (ContractorFeesService) ──────────────────────┤
                                                      │
Task 4 (InvoiceSplitService) ────────────────────────┘
```

---

## Notes

- **Database IDs** (from spec):
  - ContractorPayouts: `2c564b29-b84c-8045-80ee-000bee2e3669`
  - ContractorFees: `2c264b29-b84c-8037-807c-000bf6d0792c`
  - InvoiceSplit: `2c364b29-b84c-804f-9856-000b58702dea`
  - RefundRequest: `2cc64b29-b84c-8066-adf2-cc56171cedf4`

- **Source Type Determination**:
  - Contractor Fees relation set → `Contractor Payroll`
  - Invoice Split relation set → `Commission`
  - Refund Request relation set → `Refund`
  - None set → `Other`

- **Direction Handling**:
  - `Outgoing (you pay)` → Positive amount (company pays contractor)
  - `Incoming (you receive)` → Negative amount (deduction)

- **All code must include DEBUG logging**

---

## Completion Checklist

- [x] All 5 tasks completed
- [x] DEBUG logs added to all new code
