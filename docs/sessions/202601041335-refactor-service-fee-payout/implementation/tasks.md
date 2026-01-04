# Implementation Tasks: Refactor Service Fee Payout

## Overview
Refactor `processContractorPayrollPayouts` to use Task Order Log as data source instead of Contractor Fees.

---

## Task 1: Add UpdateOrderStatus to TaskOrderLogService

**Status:** ✅ COMPLETE (Already exists)

**File:** `pkg/service/notion/task_order_log.go:1008-1031`

Method `UpdateOrderStatus` already exists in the codebase.

---

## Task 2: Add QueryRatesByContractorPageID to ContractorRatesService

**Status:** ✅ COMPLETE (Using existing method)

**File:** `pkg/service/notion/contractor_rates.go:287-407`

Using existing `FindActiveRateByContractor(ctx, contractorPageID, orderDate)` method which provides the same functionality.

---

## Task 3: Refactor processContractorPayrollPayouts Handler

**Status:** ✅ COMPLETE

**File:** `pkg/handler/notion/contractor_payouts.go:87-276`

### 3.1 Update service dependencies
- [x] Add `taskOrderLogService := h.service.Notion.TaskOrderLog`
- [x] Add `contractorRatesService := h.service.Notion.ContractorRates`
- [x] Keep `contractorPayoutsService` (for CheckPayoutExists and CreatePayout)

### 3.2 Replace data query
- [x] Remove: `newFees, err := contractorFeesService.QueryNewFees(ctx)`
- [x] Add: `approvedOrders, err := taskOrderLogService.QueryApprovedOrders(ctx)`

### 3.3 Update processing loop
- [x] Change loop from `for _, fee := range newFees` to `for _, order := range approvedOrders`
- [x] Extract month: `month := order.Date.Format("2006-01")`
- [x] Get rate: `rate, err := contractorRatesService.FindActiveRateByContractor(ctx, order.ContractorPageID, order.Date)`
- [x] Calculate amount based on billing type
- [x] Update CreatePayoutInput fields
- [x] Replace `contractorFeesService.UpdatePaymentStatus` with `taskOrderLogService.UpdateOrderStatus`

### 3.4 Update response details
- [x] Change `fee_page_id` to `order_page_id` in detail map
- [x] Update log messages to reference "order" instead of "fee"

---

## Task 4: Update PayoutType Map

**Status:** ✅ COMPLETE

**File:** `pkg/handler/notion/contractor_payouts.go:28-34`

- [x] Changed `"contractor_payroll": "Contractor Payroll"` to `"contractor_payroll": "Service Fee"`

---

## Task 5: Build and Verify

**Status:** ✅ COMPLETE

### 5.1 Build verification
- [x] Run `go build ./...` - ✅ Passed
- [x] Run `golangci-lint run ./...` - ✅ 0 issues

### 5.2 Manual testing
- [ ] Create test Task Order Log entry with Status=Approved
- [ ] Call endpoint: `cronjobs/create-contractor-payouts?type=contractor_payroll`
- [ ] Verify payout created with correct amount
- [ ] Verify Task Order Log status updated to Completed

---

## Execution Order

1. Task 1 (UpdateOrderStatus method)
2. Task 2 (QueryRatesByContractorPageID method)
3. Task 3 (Handler refactor)
4. Task 4 (Optional - PayoutType map)
5. Task 5 (Build verification)

## Estimated Scope
- 2 new service methods
- 1 handler refactor (~50 lines changed)
- 3 files modified
