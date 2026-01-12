# Implementation Status

**Session**: 202601121659-fetch-refund-commission-before-payday
**Status**: ✅ COMPLETED

## Summary

When generating contractor invoice with a `month` param, the system now also fetches pending Refund/Commission payouts where `Date` < Payday of the given month, and merges them with existing payouts.

## Changes Made

### 1. `pkg/service/notion/contractor_payouts.go`

Added new method `QueryPendingRefundCommissionBeforeDate`:
- Queries pending payouts where Date < beforeDate
- Filters to only include Refund or Commission source types
- Includes parallel FetchInvoiceSplitInfo for commission payouts

### 2. `pkg/controller/invoice/contractor_invoice.go`

Modified `GenerateContractorInvoice`:
- Added parallel query for Refund/Commission payouts before Payday cutoff date
- Cutoff date calculated as: `{month}-{payDay}` (e.g., `2025-01-15`)
- Merged Refund/Commission payouts with existing payouts (deduplicated by PageID)
- Non-fatal error handling for Refund/Commission query

## Flow

```
1. Get PayDay from Contractor Rates
2. Calculate cutoff date: month + PayDay (e.g., 2025-01-15)
3. Query in parallel:
   - Regular payouts (by month)
   - Refund/Commission payouts (Date < cutoff date)
   - Bank account
4. Merge payouts (deduplicate by PageID)
5. Continue with invoice generation
```

## Verification

- [x] `go build ./...` passes
- [x] All tasks completed
