# Implementation Tasks: Exchange Rate in Contractor Payables

## Overview

6 tasks to implement Exchange Rate field in Contractor Payables. Estimated total time: 20-25 minutes.

## Tasks

### Task 1: Update Contractor Payables Schema Documentation

- **File(s)**: `docs/specs/notion/schema/contractor-payables.md`
- **Description**: Add Exchange Rate property to Core Properties table (after Invoice ID, before Notes) and to sample data JSON (after Currency, before Payment Status). Use property ID `KnQx` and describe as "Exchange rate used for currency conversion".
- **Acceptance**: Exchange Rate listed in Core Properties with ID `KnQx`, included in sample JSON with example value like `1.0` or `25420.50`

### Task 2: Add Exchange Rate Field to CreatePayableInput Struct

- **File(s)**: `pkg/service/notion/contractor_payables.go` (line 33)
- **Description**: Add `ExchangeRate float64` field to CreatePayableInput struct after ContractorType, before PDFBytes. Include comment: "Exchange rate used for currency conversion (optional, only saved if > 0)"
- **Acceptance**: Field added with descriptive comment, struct compiles without errors

### Task 3: Update CreatePayable Method to Save Exchange Rate

- **File(s)**: `pkg/service/notion/contractor_payables.go` (after line 256)
- **Description**: After Contractor Type logic, add conditional to save Exchange Rate property using ID `KnQx` only when `input.ExchangeRate > 0`. Add debug log: `s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: set exchangeRate=%.4f", input.ExchangeRate))`
- **Acceptance**: Property saved with correct ID `KnQx`, conditional logic works (only saves when > 0), debug logging added

### Task 4: Update updatePayable Method to Save Exchange Rate

- **File(s)**: `pkg/service/notion/contractor_payables.go` (after line 362)
- **Description**: After Contractor Type logic, add conditional to save Exchange Rate property using ID `KnQx` only when `input.ExchangeRate > 0`. Mirror logic from CreatePayable method for consistency.
- **Acceptance**: Update logic identical to create logic, property saved correctly, conditional works

### Task 5: Update Webhook Handler to Pass Exchange Rate

- **File(s)**: `pkg/handler/webhook/gen_invoice.go` (line 154-164)
- **Description**: Add `ExchangeRate: invoiceData.ExchangeRate` to payableInput struct initialization, placed after ContractorType and before PDFBytes
- **Acceptance**: Exchange rate passed from invoice data to payable creation, field placement matches struct definition

### Task 6: Update Invoice Handler to Pass Exchange Rate

- **File(s)**: `pkg/handler/invoice/invoice.go` (line 447-457)
- **Description**: Add `ExchangeRate: invoiceData.ExchangeRate` to payableInput struct initialization, placed after ContractorType and before PDFBytes. Implementation should be identical to Task 5.
- **Acceptance**: Exchange rate passed from invoice data, implementation matches webhook handler

## Recommended Implementation Order

1. Task 2 (struct field) - Foundation
2. Task 3 (CreatePayable) - Core logic
3. Task 4 (updatePayable) - Core logic
4. Task 5 (webhook handler) - Handler layer
5. Task 6 (invoice handler) - Handler layer
6. Task 1 (documentation) - Final documentation update

## Key Details

- **Property ID**: `KnQx` (existing in Notion)
- **Conditional Logic**: Only save when `input.ExchangeRate > 0`
- **Debug Logging**: Use format `"[DEBUG] contractor_payables: set exchangeRate=%.4f"`
- **Consistency**: Both handlers use identical logic

## Testing Checklist

- [ ] Exchange Rate saved to Notion via Discord webhook
- [ ] Exchange Rate saved to Notion via API
- [ ] Exchange Rate NOT saved when value is 0
- [ ] Debug logs show exchange rate values
- [ ] Notion displays exchange rate in `KnQx` property
- [ ] Existing functionality works without exchange rate
