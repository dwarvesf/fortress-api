# Specification: Handler Refactor

## Overview
Refactor `processContractorPayrollPayouts` to use Task Order Log instead of Contractor Fees.

## File
`pkg/handler/notion/contractor_payouts.go`

## Changes

### 1. Update Services Used

**Current:**
```go
contractorFeesService := h.service.Notion.ContractorFees
```

**New:**
```go
taskOrderLogService := h.service.Notion.TaskOrderLog
contractorRatesService := h.service.Notion.ContractorRates
```

### 2. Replace Query

**Current:**
```go
newFees, err := contractorFeesService.QueryNewFees(ctx)
```

**New:**
```go
approvedOrders, err := taskOrderLogService.QueryApprovedOrders(ctx)
```

### 3. Process Each Order

For each `order` in `approvedOrders`:

```go
// 1. Extract month from date
month := order.Date.Format("2006-01")

// 2. Get contractor rate
rate, err := contractorRatesService.QueryRatesByContractorPageID(ctx, order.ContractorPageID, month)
if err != nil {
    // Log error and skip
    continue
}

// 3. Calculate amount based on billing type
var amount float64
if rate.BillingType == "Monthly Fixed" {
    amount = rate.MonthlyFixed
} else {
    amount = rate.HourlyRate * order.FinalHoursWorked
}

// 4. Check if payout exists (idempotency)
exists, existingPayoutID, err := contractorPayoutsService.CheckPayoutExistsByContractorFee(ctx, order.PageID)

// 5. Create payout
payoutName := fmt.Sprintf("Development work on %s", formatMonthYear(month))
payoutInput := notionsvc.CreatePayoutInput{
    Name:             payoutName,
    ContractorPageID: order.ContractorPageID,
    TaskOrderID:      order.PageID,
    Amount:           amount,
    Currency:         rate.Currency,
    Date:             order.Date.Format("2006-01-02"),
    Description:      order.ProofOfWorks,
}

// 6. Update Task Order Log status
taskOrderLogService.UpdateOrderStatus(ctx, order.PageID, "Pending")
```

### 4. Update PayoutType Map (Optional)

```go
var PayoutType = map[string]string{
    "contractor_payroll": "Service Fee",  // Changed from "Contractor Payroll"
    ...
}
```

## Response Structure
Keep existing response structure with updated field names if needed.
