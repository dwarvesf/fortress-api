# Implementation Tasks

## Overview
Tasks for implementing Contractor Payables record creation after invoice upload.

## Task List

### Task 1: Add Config Field
**File:** `pkg/config/config.go`
**Effort:** Small
**Status:** ✅ COMPLETED

- [x] Add `ContractorPayables string` to `NotionDBs` struct (line ~281)
- [x] Add env var loading `ContractorPayables: v.GetString("NOTION_CONTRACTOR_PAYABLES_DB_ID")` (line ~511)

### Task 2: Create ContractorPayablesService
**File:** `pkg/service/notion/contractor_payables.go` (NEW)
**Effort:** Medium
**Status:** ✅ COMPLETED

- [x] Create new file `contractor_payables.go`
- [x] Define `ContractorPayablesService` struct
- [x] Define `CreatePayableInput` struct
- [x] Implement `NewContractorPayablesService` constructor
- [x] Implement `CreatePayable` method with:
  - Notion property mapping for all fields
  - Date parsing for Period and InvoiceDate
  - Relation handling for Contractor and Payout Items
  - External file attachment for PDF URL
  - DEBUG logging throughout

### Task 3: Update Services Struct
**File:** `pkg/service/notion/notion_services.go`
**Effort:** Small
**Status:** ✅ COMPLETED

- [x] Add `ContractorPayables *ContractorPayablesService` field

### Task 4: Initialize Service
**File:** `pkg/service/service.go`
**Effort:** Small
**Status:** ✅ COMPLETED

- [x] Add `ContractorPayables: notion.NewContractorPayablesService(cfg, logger.L)` (line ~309)

### Task 5: Update ContractorInvoiceData
**File:** `pkg/controller/invoice/contractor_invoice.go`
**Effort:** Small
**Status:** ✅ COMPLETED

- [x] Add `ContractorPageID string` field to struct (line ~55)
- [x] Add `PayoutPageIDs []string` field to struct (line ~56)
- [x] Collect payout page IDs during processing loop (line ~557-562)
- [x] Set `ContractorPageID` from `rateData.ContractorPageID` (line ~554)
- [x] Set `PayoutPageIDs` in invoiceData

### Task 6: Integrate in Handler
**File:** `pkg/handler/invoice/invoice.go`
**Effort:** Medium
**Status:** ✅ COMPLETED

- [x] Add step 5.5 after successful upload (line ~442)
- [x] Create `CreatePayableInput` from invoiceData
- [x] Call `h.service.Notion.ContractorPayables.CreatePayable`
- [x] Handle error non-blocking (log and continue)
- [x] Add DEBUG logging

### Task 7: Verification
**Effort:** Medium
**Status:** ✅ COMPLETED

- [x] Run `go build ./cmd/server` to verify compilation
- [ ] Test with `skipUpload=true` (should NOT create record) - Manual test required
- [ ] Test with actual upload (should create record) - Manual test required
- [ ] Verify Notion record has correct data and relations - Manual test required

## Execution Order
```
Task 1 (Config)
    ↓
Task 2 (Service) ← Main implementation
    ↓
Task 3 (Services struct)
    ↓
Task 4 (Initialize)
    ↓
Task 5 (InvoiceData)
    ↓
Task 6 (Handler)
    ↓
Task 7 (Verify)
```

## Dependencies
- Task 2 depends on Task 1 (needs config)
- Task 4 depends on Tasks 2, 3 (needs service and struct)
- Task 6 depends on Tasks 4, 5 (needs service and data)

## Environment Setup
Ensure `.env` has:
```
NOTION_CONTRACTOR_PAYABLES_DB_ID=2c264b29-b84c-8037-807c-000bf6d0792c
```
