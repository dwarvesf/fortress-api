# Specification: Data Structure Changes

**Version:** 1.0
**Date:** 2026-01-07
**Status:** Ready for Implementation

## Overview

This specification details the exact changes required to support multi-currency display in contractor invoices, including struct modifications, field additions, and data population logic.

## File: pkg/controller/invoice/contractor_invoice.go

### 1. ContractorInvoiceLineItem Struct Changes

**Location:** Lines 48-61

**Current Structure:**
```go
type ContractorInvoiceLineItem struct {
    Title       string
    Description string
    Hours       float64
    Rate        float64
    Amount      float64
    AmountUSD   float64
    Type        string

    CommissionRole    string
    CommissionProject string
}
```

**Updated Structure:**
```go
type ContractorInvoiceLineItem struct {
    Title       string
    Description string
    Hours       float64 // Quantity for display (always 1 for current use case)
    Rate        float64 // Unit cost in USD (for backward compatibility)
    Amount      float64 // Total amount in USD (for backward compatibility)
    AmountUSD   float64 // Amount converted to USD
    Type        string  // Payout source type

    // Commission-specific fields
    CommissionRole    string
    CommissionProject string

    // NEW: Original currency fields
    OriginalAmount   float64 // Amount in original currency (VND or USD)
    OriginalCurrency string  // "VND" or "USD"
}
```

**New Fields:**
- `OriginalAmount` (float64): The amount in its original currency from Notion PayoutEntry
- `OriginalCurrency` (string): The currency code from Notion PayoutEntry ("VND" or "USD")

**Field Population Logic:**
```go
// In GenerateContractorInvoice function, when building line items:
lineItem := ContractorInvoiceLineItem{
    Title:             title,
    Description:       description,
    Hours:             1,
    Rate:              amountUSD,
    Amount:            amountUSD,
    AmountUSD:         amountUSD,
    Type:              string(payout.SourceType),
    CommissionRole:    payout.CommissionRole,
    CommissionProject: payout.CommissionProject,

    // NEW: Preserve original currency
    OriginalAmount:   payout.Amount,        // From Notion
    OriginalCurrency: payout.Currency,      // From Notion
}
```

### 2. ContractorInvoiceData Struct Changes

**Location:** Lines 25-46

**Current Structure:**
```go
type ContractorInvoiceData struct {
    InvoiceNumber     string
    ContractorName    string
    Month             string
    Date              time.Time
    DueDate           time.Time
    Description       string
    BillingType       string
    Currency          string
    LineItems         []ContractorInvoiceLineItem
    MonthlyFixed      float64
    MonthlyFixedUSD   float64
    Total             float64
    TotalUSD          float64
    ExchangeRate      float64
    BankAccountHolder string
    BankName          string
    BankAccountNumber string
    BankSwiftBIC      string
    BankBranch        string
}
```

**Updated Structure:**
```go
type ContractorInvoiceData struct {
    InvoiceNumber     string
    ContractorName    string
    Month             string
    Date              time.Time
    DueDate           time.Time
    Description       string
    BillingType       string
    Currency          string // Keep as "USD" for payment processing
    LineItems         []ContractorInvoiceLineItem
    MonthlyFixed      float64
    MonthlyFixedUSD   float64
    Total             float64
    TotalUSD          float64
    ExchangeRate      float64 // VND to USD exchange rate from Wise
    BankAccountHolder string
    BankName          string
    BankAccountNumber string
    BankSwiftBIC      string
    BankBranch        string

    // NEW: Multi-currency subtotal fields
    SubtotalVND        float64 // Sum of all VND-denominated items
    SubtotalUSDFromVND float64 // SubtotalVND converted to USD at ExchangeRate
    SubtotalUSDItems   float64 // Sum of all USD-denominated items
    SubtotalUSD        float64 // SubtotalUSDFromVND + SubtotalUSDItems
    FXSupport          float64 // FX support fee (hardcoded $8 for now)
}
```

**New Fields:**
- `SubtotalVND` (float64): Sum of all line items with `OriginalCurrency == "VND"`
- `SubtotalUSDFromVND` (float64): `SubtotalVND` converted to USD at the `ExchangeRate`
- `SubtotalUSDItems` (float64): Sum of all line items with `OriginalCurrency == "USD"`
- `SubtotalUSD` (float64): `SubtotalUSDFromVND + SubtotalUSDItems`
- `FXSupport` (float64): FX support fee (hardcoded to 8.0 for now, TODO: implement calculation)

## Calculation Logic

### 3. Subtotal Calculation Implementation

**Location:** After line item grouping, before creating invoiceData (around line 310)

**New Code Block:**
```go
// Calculate subtotals for multi-currency display
var subtotalVND float64
var subtotalUSDItems float64

for _, item := range lineItems {
    switch item.OriginalCurrency {
    case "VND":
        subtotalVND += item.OriginalAmount
    case "USD":
        subtotalUSDItems += item.AmountUSD
    default:
        l.Debug(fmt.Sprintf("Unknown currency: %s, treating as USD", item.OriginalCurrency))
        subtotalUSDItems += item.AmountUSD
    }
}

// Round subtotals
subtotalVND = math.Round(subtotalVND) // VND has no minor units
subtotalUSDItems = math.Round(subtotalUSDItems*100) / 100

l.Debug(fmt.Sprintf("subtotals - VND: %.0f, USD items: %.2f", subtotalVND, subtotalUSDItems))

// Get exchange rate for VND → USD conversion
// Query Wise API for current VND to USD rate
exchangeRate := 1.0
var subtotalUSDFromVND float64

if subtotalVND > 0 {
    // Convert VND subtotal to USD
    amountUSD, rate, err := c.service.Wise.Convert(subtotalVND, "VND", "USD")
    if err != nil {
        l.Error(err, "failed to convert VND subtotal to USD")
        return nil, fmt.Errorf("failed to convert VND subtotal to USD: %w", err)
    }
    exchangeRate = rate
    subtotalUSDFromVND = math.Round(amountUSD*100) / 100

    l.Debug(fmt.Sprintf("converted VND %.0f to USD %.2f at rate %.4f", subtotalVND, subtotalUSDFromVND, exchangeRate))
}

// Calculate total USD
subtotalUSD := subtotalUSDFromVND + subtotalUSDItems
fxSupport := 8.0 // Hardcoded for now, TODO: implement calculation
totalUSD := subtotalUSD + fxSupport

l.Debug(fmt.Sprintf("calculated totals - subtotalUSD: %.2f, fxSupport: %.2f, totalUSD: %.2f", subtotalUSD, fxSupport, totalUSD))
```

### 4. Update invoiceData Creation

**Location:** Lines 349-369

**Changes:**
```go
invoiceData := &ContractorInvoiceData{
    InvoiceNumber:     invoiceNumber,
    ContractorName:    rateData.ContractorName,
    Month:             month,
    Date:              now,
    DueDate:           dueDate,
    Description:       description,
    BillingType:       rateData.BillingType,
    Currency:          "USD", // Keep as USD for payment
    LineItems:         lineItems,
    MonthlyFixed:      0,
    MonthlyFixedUSD:   0,
    Total:             totalUSD,      // CHANGED: Use calculated total
    TotalUSD:          totalUSD,      // CHANGED: Use calculated total
    ExchangeRate:      exchangeRate,  // CHANGED: Use actual rate from Wise
    BankAccountHolder: bankAccount.AccountHolderName,
    BankName:          bankAccount.BankName,
    BankAccountNumber: bankAccount.AccountNumber,
    BankSwiftBIC:      bankAccount.SwiftBIC,
    BankBranch:        bankAccount.BranchAddress,

    // NEW: Populate subtotal fields
    SubtotalVND:        subtotalVND,
    SubtotalUSDFromVND: subtotalUSDFromVND,
    SubtotalUSDItems:   subtotalUSDItems,
    SubtotalUSD:        subtotalUSD,
    FXSupport:          fxSupport,
}
```

## Data Validation

### 5. Validation Rules

Add validation before creating invoiceData:

```go
// Validate currency codes
for i, item := range lineItems {
    if item.OriginalCurrency != "VND" && item.OriginalCurrency != "USD" {
        l.Error(nil, fmt.Sprintf("invalid currency for line item %d: %s", i, item.OriginalCurrency))
        return nil, fmt.Errorf("invalid currency for line item %d: %s (must be VND or USD)", i, item.OriginalCurrency)
    }

    if item.OriginalAmount < 0 {
        l.Error(nil, fmt.Sprintf("negative amount for line item %d: %.2f %s", i, item.OriginalAmount, item.OriginalCurrency))
        return nil, fmt.Errorf("negative amount for line item %d", i)
    }
}
```

## Testing Considerations

### Test Cases Required

1. **All VND items:** SubtotalVND populated, SubtotalUSDItems = 0
2. **All USD items:** SubtotalUSDItems populated, SubtotalVND = 0
3. **Mixed currencies:** Both subtotals populated correctly
4. **Zero amounts:** Handle gracefully (no division by zero)
5. **Invalid currency:** Return error
6. **Negative amounts:** Return error
7. **Exchange rate fetch failure:** Return error with clear message
8. **Rounding:** VND rounded to 0 decimals, USD to 2 decimals

### Edge Cases

- Empty line items array
- Single line item
- Large amounts (billions in VND)
- Very small USD amounts (< $0.01)
- Exchange rate = 0 (should never happen, but handle)

## Migration Notes

### Backward Compatibility

- Existing `Amount`, `AmountUSD`, `Rate`, `Hours` fields remain unchanged
- Templates using old fields will continue to work
- New templates should prefer `OriginalAmount` and `OriginalCurrency`

### Rollout Strategy

1. Update structs with new fields
2. Update data population logic to fill new fields
3. Update template to use new fields
4. Deploy together (no interim state needed)

## Related Documents

- ADR-001: Data Structure for Multi-Currency Support
- spec-004-calculation-logic.md: Detailed calculation algorithms
- spec-002-template-functions.md: Template function usage
