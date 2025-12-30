# Integrate Payout Services into Invoice Generation

**Session**: 202512301700-integrate-payout-services
**Specification**: `docs/specs/notion/contractor-invoice.md`
**Status**: Complete

---

## Tasks

### Task 1: Add Name field to PayoutEntry struct

- **File(s)**: `pkg/service/notion/contractor_payouts.go`
- **Description**: Extract `Name` property from Payout page
- **Changes**:
  - Add `Name string` field to `PayoutEntry` struct
  - Extract Name (title) property in query
- **Acceptance**: PayoutEntry includes Name field

---

### Task 2: Update GenerateContractorInvoice to use PayoutsService

- **File(s)**: `pkg/controller/invoice/contractor_invoice.go`
- **Description**: Replace Task Order Log query with Contractor Payouts query
- **Changes**:
  1. Initialize `ContractorPayoutsService`, `ContractorFeesService`
  2. Get contractor page ID from `ContractorRatesService`
  3. Call `QueryPendingPayoutsByContractor(ctx, contractorPageID)`
  4. For each payout entry:
     - Use `Name` as line item Title
     - Use `Amount` converted to USD as line item Amount
     - If `Contractor Payroll` type: fetch `ProofOfWorks` from `ContractorFeesService` and append to Description
  5. Convert all amounts to USD using Wise service
- **Acceptance**: Invoice generates with data from Payouts database

---

### Task 3: Add DEBUG logging

- **File(s)**: `pkg/controller/invoice/contractor_invoice.go`
- **Description**: Ensure comprehensive DEBUG logging for new payout flow
- **Changes**:
  - Log payout query results
  - Log each payout processing
  - Log USD conversion
- **Acceptance**: DEBUG logs trace full invoice generation flow

---

## Execution Order

```
Task 1 (Add Name field) → Task 2 (Update controller) → Task 3 (Add logging)
```

---

## Notes

- Use `Name` from Payout as line item description
- Convert all amounts to USD
- For `Contractor Payroll`: append `ProofOfWorks` from ContractorFees to description
- No need to display source type separately in invoice
- Filter: `Status = Pending` AND `Direction = Outgoing`
