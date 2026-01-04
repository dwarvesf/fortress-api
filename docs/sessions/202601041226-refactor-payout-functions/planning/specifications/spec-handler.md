# Specification: Handler Changes

## File
`pkg/handler/notion/contractor_payouts.go`

## Changes Required

### 1. PayoutType Map (line 27-34)

**Current:**
```go
PayoutType = map[string]string{
    "contractor_payroll": "Contractor Payroll",
    "bonus":              "Bonus",
    "commission":         "Commission",
    "refund":             "Refund",
}
```

**Assessment:** This map is used for the `type` query parameter mapping, but the values are no longer written to Notion (Type is now a formula).

**Decision:** Keep as-is for API backward compatibility. The string values are not used in Notion writes anymore.

### 2. processContractorPayrollPayouts (line 89-242)

**Update CreatePayoutInput usage:**
```go
// OLD:
payoutInput := notionsvc.CreatePayoutInput{
    Name:             payoutName,
    ContractorPageID: fee.ContractorPageID,
    ContractorFeeID:  fee.PageID,
    Amount:           fee.TotalAmount,
    Currency:         fee.Currency,
    Month:            fee.Month,      // REMOVE
    Date:             fee.Date,
    Type:             payoutType,     // REMOVE
}

// NEW:
payoutInput := notionsvc.CreatePayoutInput{
    Name:             payoutName,
    ContractorPageID: fee.ContractorPageID,
    TaskOrderID:      fee.PageID,
    Amount:           fee.TotalAmount,
    Currency:         fee.Currency,
    Date:             fee.Date,
    Description:      "",  // Optional, leave empty
}
```

### 3. processRefundPayouts (line 244-405)

**Update CreateRefundPayoutInput usage:**
```go
// OLD:
payoutInput := notionsvc.CreateRefundPayoutInput{
    Name:             payoutName,
    ContractorPageID: refund.ContractorPageID,
    RefundRequestID:  refund.PageID,
    Amount:           refund.Amount,
    Currency:         refund.Currency,
    Month:            month,           // REMOVE
    Date:             refund.DateRequested,
}

// NEW:
payoutInput := notionsvc.CreateRefundPayoutInput{
    Name:             payoutName,
    ContractorPageID: refund.ContractorPageID,
    RefundRequestID:  refund.PageID,
    Amount:           refund.Amount,
    Currency:         refund.Currency,
    Date:             refund.DateRequested,
}
```

### 4. processCommissionPayouts & processBonusPayouts

No changes needed - these already don't set Month or Type in their input structs.
