# Implementation Tasks: Invoice Splits Generation

## Task 1: Add Worker Message Type
**File**: `pkg/worker/message_types.go` (NEW)

- [ ] Create `GenerateInvoiceSplitsMsg` constant
- [ ] Create `GenerateInvoiceSplitsPayload` struct

---

## Task 2: Extend ClientInvoiceService - Query Line Items with Commissions
**File**: `pkg/service/notion/invoice.go`

- [ ] Create `LineItemCommissionData` struct
- [ ] Implement `QueryLineItemsWithCommissions(invoicePageID)` method
  - Query line items with Parent item = invoicePageID
  - Extract commission %: Sales, Account Mgr, Delivery Lead, Hiring Referral
  - Extract calculated amounts from formulas
  - Extract person IDs from rollups
  - Extract Deployment relation
  - Extract Currency, Issue Date (Month), Project Code

---

## Task 3: Extend ClientInvoiceService - Mark Splits Generated
**File**: `pkg/service/notion/invoice.go`

- [ ] Implement `MarkSplitsGenerated(invoicePageID)` method
  - Update "Splits Generated" checkbox to true via Notion API
- [ ] Implement `IsSplitsGenerated(invoicePageID)` method
  - Query invoice page, check "Splits Generated" property

---

## Task 4: Extend InvoiceSplitService - Create Commission Split
**File**: `pkg/service/notion/invoice_split.go`

- [ ] Create `CreateCommissionSplitInput` struct
- [ ] Implement `CreateCommissionSplit(input)` method
  - Create page in Invoice Splits database
  - Set properties: Name, Amount, Currency, Month, Role, Type, Status
  - Set relations: Contractor, Deployment, Invoice Item, Client Invoices

---

## Task 5: Add Worker Handler
**File**: `pkg/worker/worker.go`

- [ ] Import message types
- [ ] Add case for `GenerateInvoiceSplitsMsg` in `ProcessMessage()` switch
- [ ] Implement `handleGenerateInvoiceSplits(l, payload)` method:
  1. Extract `GenerateInvoiceSplitsPayload`
  2. Check `IsSplitsGenerated()` - skip if true
  3. Call `QueryLineItemsWithCommissions()`
  4. For each line item:
     - For each role (Sales, AM, DL, HR) with amount > 0:
       - Build split name
       - Call `CreateCommissionSplit()`
  5. Call `MarkSplitsGenerated()`
  6. Log summary

---

## Task 6: Integrate with Mark Paid Flow
**File**: `pkg/controller/invoice/mark_paid.go`

- [ ] In `processNotionInvoicePaid()`, after line 161 (status update):
  - Enqueue `GenerateInvoiceSplitsMsg` with invoice page ID

---

## Task 7: Update Service Initialization
**Files**: `pkg/service/notion/notion_services.go`, `pkg/service/service.go`

- [ ] Ensure InvoiceSplitService is initialized and accessible
- [ ] Ensure Worker has access to Notion services

---

## Task 8: Build & Verify
- [ ] Run `go build ./...`
- [ ] Fix any compilation errors

---

## Dependencies

```
Task 1 (message types)
    │
    ├── Task 2 (query line items) ──┐
    │                               │
    ├── Task 3 (mark splits)        ├── Task 5 (worker handler) ── Task 6 (integration)
    │                               │
    └── Task 4 (create split) ──────┘
                                    │
                                    └── Task 7 (service init)
                                            │
                                            └── Task 8 (build)
```

## Estimated Order
1. Task 1 (message types)
2. Task 2, 3, 4 (parallel - service methods)
3. Task 5 (worker handler)
4. Task 7 (service init)
5. Task 6 (integration)
6. Task 8 (build & verify)
