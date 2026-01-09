# Task Breakdown: Exchange Rate in Contractor Payables

## Overview

This document provides a detailed, dependency-ordered task breakdown for implementing the Exchange Rate feature in Contractor Payables. The Exchange Rate property already exists in Notion (ID: `KnQx`) and the `ContractorInvoiceData` struct already contains the `ExchangeRate` field populated during invoice generation.

## Implementation Tasks

### Task 1: Update Contractor Payables Schema Documentation

**File(s)**:
- `/Users/quang/workspace/dwarvesf/fortress-api/docs/specs/notion/schema/contractor-payables.md`

**Description**:
Add the Exchange Rate property to the schema documentation to reflect the new field being saved to Notion.

**Changes Required**:
1. Add Exchange Rate entry to the "Core Properties" table (line 18-30):
   ```markdown
   | Property | Type | ID | Description |
   |----------|------|-----|-------------|
   | `Payable` | Title | `title` | Main payable name/description |
   | `Total` | Number | `` `GxT`` | Total payable amount (formatted with commas) |
   | `Period` | Date | `>HDF` | Month this payable covers |
   | `Invoice Date` | Date | `QD]A` | Date of invoice |
   | `Payment Date` | Date | `{yf?` | Date payment was processed |
   | `Invoice ID` | Rich Text | `{ShZ` | Invoice number submitted by contractor |
   | `Exchange Rate` | Number | `KnQx` | Exchange rate used for currency conversion |
   | `Notes` | Rich Text | `jyfE` | Additional notes |
   | `ID` | Unique ID | `IW]b` | Auto-generated unique identifier |
   | `Attachments` | Files | `^LN`` | Attached invoice files |
   ```

2. Add Exchange Rate to sample data JSON (line 183-266):
   ```json
   "Exchange Rate": {
     "type": "number",
     "number": 1.0
   },
   ```
   Place this after the `Currency` property and before `Payment Status`.

**Acceptance Criteria**:
- Exchange Rate is listed in the Core Properties table with correct property ID (`KnQx`)
- Sample data JSON includes the Exchange Rate field with a realistic example value
- Documentation accurately describes the purpose (exchange rate used for currency conversion)

---

### Task 2: Add Exchange Rate Field to CreatePayableInput Struct

**File(s)**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payables.go`

**Description**:
Add the `ExchangeRate` field to the `CreatePayableInput` struct to accept exchange rate values when creating payable records.

**Changes Required**:
1. Add `ExchangeRate` field to `CreatePayableInput` struct (after line 33):
   ```go
   type CreatePayableInput struct {
       ContractorPageID string   // Relation to Contractor (required)
       Total            float64  // Total amount in USD (required)
       Currency         string   // "USD" or "VND" (required)
       Period           string   // YYYY-MM-DD start of month (required)
       InvoiceDate      string   // YYYY-MM-DD (required)
       InvoiceID        string   // Invoice number e.g., CONTR-202512-A1B2 (required)
       PayoutItemIDs    []string // Relation to Payout Items (required)
       ContractorType   string   // "Individual", "Sole Proprietor", "LLC", etc. (optional, defaults to "Individual")
       ExchangeRate     float64  // Exchange rate used for currency conversion (optional, only saved if > 0)
       PDFBytes         []byte   // PDF file bytes to upload to Notion (optional)
   }
   ```

**Acceptance Criteria**:
- `ExchangeRate float64` field is added to the struct
- Field includes a descriptive comment explaining it's optional and only saved when > 0
- Struct ordering is logical (added before `PDFBytes` since it's invoice-related data)

---

### Task 3: Update CreatePayable Method to Save Exchange Rate

**File(s)**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payables.go`

**Description**:
Modify the `CreatePayable` method to save the Exchange Rate property to Notion when creating a new payable record (only if value > 0).

**Changes Required**:
1. Add Exchange Rate property save logic after Contractor Type (after line 256):
   ```go
   // Add Contractor Type (default to "Individual" if not provided)
   contractorType := input.ContractorType
   if contractorType == "" {
       contractorType = "Individual"
   }
   props["Contractor Type"] = nt.DatabasePageProperty{
       Select: &nt.SelectOptions{
           Name: contractorType,
       },
   }
   s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: set contractorType=%s", contractorType))

   // Add Exchange Rate if provided (only save if > 0)
   if input.ExchangeRate > 0 {
       props["KnQx"] = nt.DatabasePageProperty{
           Number: &input.ExchangeRate,
       }
       s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: set exchangeRate=%.4f", input.ExchangeRate))
   }
   ```

**Acceptance Criteria**:
- Exchange Rate is only saved when `input.ExchangeRate > 0`
- Property ID `KnQx` is used (matching Notion database schema)
- Debug logging is added to track exchange rate values being set
- Placement is logical (after Contractor Type, before CreatePageParams)

---

### Task 4: Update updatePayable Method to Save Exchange Rate

**File(s)**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payables.go`

**Description**:
Modify the `updatePayable` method to save the Exchange Rate property when updating an existing payable record (only if value > 0).

**Changes Required**:
1. Add Exchange Rate property save logic after Contractor Type (after line 362):
   ```go
   // Add Contractor Type (default to "Individual" if not provided)
   contractorType := input.ContractorType
   if contractorType == "" {
       contractorType = "Individual"
   }
   props["Contractor Type"] = nt.DatabasePageProperty{
       Select: &nt.SelectOptions{
           Name: contractorType,
       },
   }

   // Add Exchange Rate if provided (only save if > 0)
   if input.ExchangeRate > 0 {
       props["KnQx"] = nt.DatabasePageProperty{
           Number: &input.ExchangeRate,
       }
   }
   ```

**Acceptance Criteria**:
- Exchange Rate is only saved when `input.ExchangeRate > 0`
- Property ID `KnQx` is used (matching Notion database schema)
- Update logic mirrors the create logic for consistency
- Placement is logical (after Contractor Type, before UpdatePage call)

---

### Task 5: Update Webhook Handler to Pass Exchange Rate

**File(s)**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/webhook/gen_invoice.go`

**Description**:
Update the Discord webhook handler (`processGenInvoice`) to include the exchange rate from invoice data when creating the contractor payable.

**Changes Required**:
1. Add `ExchangeRate` field to `payableInput` (line 154-164):
   ```go
   payableInput := notion.CreatePayableInput{
       ContractorPageID: invoiceData.ContractorPageID,
       Total:            invoiceData.TotalUSD,
       Currency:         "USD",
       Period:           invoiceData.Month + "-01",
       InvoiceDate:      time.Now().Format("2006-01-02"),
       InvoiceID:        invoiceData.InvoiceNumber,
       PayoutItemIDs:    invoiceData.PayoutPageIDs,
       ContractorType:   "Individual", // Default to Individual
       ExchangeRate:     invoiceData.ExchangeRate,
       PDFBytes:         pdfBytes,     // Upload PDF to Notion
   }
   ```

**Acceptance Criteria**:
- `ExchangeRate: invoiceData.ExchangeRate` is added to the input struct
- Field is placed logically (after `ContractorType`, before `PDFBytes`)
- No changes to existing fields or logic
- Exchange rate value comes directly from the generated invoice data

---

### Task 6: Update Invoice Handler to Pass Exchange Rate

**File(s)**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/invoice/invoice.go`

**Description**:
Update the invoice handler (`GenerateContractorInvoice`) to include the exchange rate from invoice data when creating the contractor payable.

**Changes Required**:
1. Add `ExchangeRate` field to `payableInput` (line 447-457):
   ```go
   payableInput := notion.CreatePayableInput{
       ContractorPageID: invoiceData.ContractorPageID,
       Total:            invoiceData.TotalUSD,
       Currency:         "USD",
       Period:           invoiceData.Month + "-01",
       InvoiceDate:      time.Now().Format("2006-01-02"),
       InvoiceID:        invoiceData.InvoiceNumber,
       PayoutItemIDs:    invoiceData.PayoutPageIDs,
       ContractorType:   "Individual", // Default to Individual
       ExchangeRate:     invoiceData.ExchangeRate,
       PDFBytes:         pdfBytes,     // Upload PDF to Notion
   }
   ```

**Acceptance Criteria**:
- `ExchangeRate: invoiceData.ExchangeRate` is added to the input struct
- Field is placed logically (after `ContractorType`, before `PDFBytes`)
- No changes to existing fields or logic
- Exchange rate value comes directly from the generated invoice data
- Implementation is identical to webhook handler for consistency

---

## Task Dependencies

```
Task 1 (Documentation)
  ├─ Independent (can be done first or in parallel)

Task 2 (Add struct field)
  ├─ Must complete before Task 3, 4, 5, 6

Task 3 (CreatePayable)
  ├─ Depends on Task 2
  ├─ Must complete before Task 5, 6

Task 4 (updatePayable)
  ├─ Depends on Task 2
  ├─ Must complete before Task 5, 6

Task 5 (Webhook handler)
  ├─ Depends on Task 2, 3

Task 6 (Invoice handler)
  ├─ Depends on Task 2, 3
```

## Recommended Implementation Order

1. **Task 2** - Add struct field (foundation for all other code changes)
2. **Task 3** - Update CreatePayable method (core service logic)
3. **Task 4** - Update updatePayable method (core service logic)
4. **Task 5** - Update webhook handler (handler layer)
5. **Task 6** - Update invoice handler (handler layer)
6. **Task 1** - Update documentation (can be done anytime, recommended last to ensure accuracy)

## Testing Notes

After implementation, verify:
1. Exchange Rate is saved to Notion when creating contractor payables via Discord webhook
2. Exchange Rate is saved to Notion when generating contractor invoices via API
3. Exchange Rate is NOT saved when value is 0 or negative
4. Debug logs show exchange rate values being set
5. Notion database displays the exchange rate correctly in the `KnQx` property
6. Existing functionality (payable creation without exchange rate) still works

## Implementation Time Estimates

- Task 1: 5 minutes (documentation update)
- Task 2: 2 minutes (add struct field)
- Task 3: 5 minutes (create method logic)
- Task 4: 5 minutes (update method logic)
- Task 5: 2 minutes (webhook handler)
- Task 6: 2 minutes (invoice handler)

**Total estimated time**: 20-25 minutes

## Additional Context

- The Exchange Rate property (`KnQx`) already exists in the Notion database schema
- The `ContractorInvoiceData.ExchangeRate` field is already populated during invoice generation
- This implementation simply passes the existing value to Notion when creating payable records
- The exchange rate represents the conversion rate from contractor currency to USD
- For USD contractors, the exchange rate will typically be 1.0
