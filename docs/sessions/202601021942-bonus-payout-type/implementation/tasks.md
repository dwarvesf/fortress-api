# Bonus Payout Type Implementation

## Overview
Add support for `type=bonus` in the `/cronjobs/create-contractor-payouts` endpoint. The logic mirrors the existing `commission` type but filters for `Type=Bonus` in Invoice Splits.

## Tasks

### Task 1: Add QueryPendingBonusSplits method ✅
- **File(s)**: `pkg/service/notion/invoice_split.go`
- **Description**: Add `QueryPendingBonusSplits()` method that queries Invoice Splits with `Status=Pending` AND `Type=Bonus`. Copy the logic from `QueryPendingCommissionSplits` and change the Type filter from "Commission" to "Bonus".
- **Acceptance**: Method returns `[]PendingCommissionSplit` (reuse struct) for bonus type splits

### Task 2: Add CreateBonusPayout method ✅
- **File(s)**: `pkg/service/notion/contractor_payouts.go`
- **Description**: Add `CreateBonusPayoutInput` struct and `CreateBonusPayout()` method. Copy from `CreateCommissionPayout` and change the Type property from "Commission" to "Bonus".
- **Acceptance**: Method creates a payout record in Contractor Payouts with `Type=Bonus`

### Task 3: Add processBonusPayouts handler ✅
- **File(s)**: `pkg/handler/notion/contractor_payouts.go`
- **Description**: Add `processBonusPayouts()` method that:
  1. Calls `QueryPendingBonusSplits()` to get pending bonus splits
  2. For each split, checks idempotency via `CheckPayoutExistsByInvoiceSplit()`
  3. Creates payout via `CreateBonusPayout()`
  4. Returns summary response
- **Acceptance**: Method processes bonus splits and creates payouts with proper logging

### Task 4: Wire up bonus case in CreateContractorPayouts ✅
- **File(s)**: `pkg/handler/notion/contractor_payouts.go`
- **Description**: Replace the `case "bonus":` block (lines 74-77) that returns "not implemented" with a call to `h.processBonusPayouts(c, l, payoutType)`
- **Acceptance**: Endpoint accepts `?type=bonus` and processes bonus payouts

## Dependencies
- Task 1 → Task 3 (handler needs query method)
- Task 2 → Task 3 (handler needs create method)
- Task 3 → Task 4 (wiring needs handler method)

## Testing
After implementation, test with:
```bash
curl -X POST "http://localhost:8080/api/v1/cronjobs/create-contractor-payouts?type=bonus"
```
