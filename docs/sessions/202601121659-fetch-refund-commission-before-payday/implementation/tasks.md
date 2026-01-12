# Task Breakdown: Fetch Refund/Commission Payouts Before Payday

## Summary

When generating contractor invoice with a `month` param, also fetch pending Refund/Commission payouts where `Date` < Payday of the given month, and combine with existing payouts.

## Prerequisites

- Issue date already uses Payday from Contractor Rates (implemented)
- Due date = issue date + 10 days (implemented)

---

## Tasks

### Task 1: Add QueryPendingRefundCommissionBeforeDate Method ✅ DONE

- **File(s)**: `pkg/service/notion/contractor_payouts.go`
- **Description**: Add new method to query pending Refund/Commission payouts before a cutoff date
  - Filter: Person = contractorPageID
  - Filter: Status = Pending
  - Filter: Date < beforeDate (YYYY-MM-DD format)
  - Post-filter: Only include SourceType = Refund OR Commission
- **Acceptance**: Method compiles and returns []PayoutEntry filtered correctly

### Task 2: Integrate New Query in GenerateContractorInvoice ✅ DONE

- **File(s)**: `pkg/controller/invoice/contractor_invoice.go`
- **Description**: After fetching regular payouts (around line 141):
  1. Build cutoff date: `fmt.Sprintf("%s-%02d", month, payDay)` (e.g., `2025-01-15`)
  2. Call `QueryPendingRefundCommissionBeforeDate` with cutoff date
  3. Merge results with existing payouts (deduplicate by PageID)
- **Acceptance**: Invoice generation includes Refund/Commission payouts dated before Payday

### Task 3: Add Debug Logging ✅ DONE

- **File(s)**: Both files above
- **Description**: Add debug logs for:
  - Cutoff date calculation
  - Number of Refund/Commission payouts found
  - Merged payout count
- **Acceptance**: Logs visible during invoice generation

### Task 4: Build Verification ✅ DONE

- **File(s)**: N/A
- **Description**: Run `go build ./...` to verify compilation
- **Acceptance**: No compilation errors

---

## Execution Order

1. Task 1 (new method)
2. Task 2 (integration)
3. Task 3 (logging - can be done with Task 1 & 2)
4. Task 4 (verification)
