# ADR-001: Data Structure for Multi-Currency Support

**Status:** Proposed

**Date:** 2026-01-07

**Context:** Contractor invoices need to display line items in their original currency (VND or USD) while maintaining USD totals for payment processing. Currently, all amounts are converted to USD during invoice generation, losing the original currency information needed for proper display.

## Decision

We will adopt a **dual-amount storage approach** where each line item preserves both its original currency data and the USD-converted amount.

### Data Structure Changes

#### 1. ContractorInvoiceLineItem Enhancement

Add original currency fields to preserve source data:

```go
type ContractorInvoiceLineItem struct {
    // Existing fields
    Title       string
    Description string
    Hours       float64
    Rate        float64
    Amount      float64
    AmountUSD   float64
    Type        string
    CommissionRole    string
    CommissionProject string

    // NEW: Original currency fields
    OriginalAmount   float64 // Amount in original currency (VND or USD)
    OriginalCurrency string  // "VND" or "USD"
}
```

**Rationale:**
- Preserves source data for accurate display
- Maintains backward compatibility (existing fields unchanged)
- Enables template to choose which amount to display
- Supports audit trail of currency conversions

#### 2. ContractorInvoiceData Enhancement

Add subtotal calculation fields for multi-currency display:

```go
type ContractorInvoiceData struct {
    // Existing fields (unchanged)
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

    // NEW: Subtotal fields for display
    SubtotalVND        float64 // Sum of all VND-denominated items
    SubtotalUSDFromVND float64 // SubtotalVND converted to USD
    SubtotalUSDItems   float64 // Sum of all USD-denominated items
    SubtotalUSD        float64 // SubtotalUSDFromVND + SubtotalUSDItems
    FXSupport          float64 // Hardcoded $8 for now
}
```

**Rationale:**
- Pre-calculates all display values in business logic layer
- Avoids complex calculations in template
- Makes template easier to test
- Centralizes calculation logic in Go code (more testable)
- Supports clear audit trail: VND subtotal → USD conversion → USD items → FX → Total

### Data Flow

```
Notion PayoutEntry (Amount, Currency)
    ↓
Controller: Preserve original + convert to USD
    ↓
ContractorInvoiceLineItem (OriginalAmount, OriginalCurrency, AmountUSD)
    ↓
Controller: Calculate subtotals
    ↓
ContractorInvoiceData (SubtotalVND, SubtotalUSD, FXSupport, Total)
    ↓
Template: Format and display
```

## Alternatives Considered

### Alternative 1: Store Only USD with Currency Flag
**Rejected:** Loses precision of original VND amounts, makes it harder to verify against source data.

### Alternative 2: Calculate Subtotals in Template
**Rejected:** Moves business logic into presentation layer, harder to test, prone to calculation errors.

### Alternative 3: Separate VND and USD Line Item Arrays
**Rejected:** Complicates sorting and grouping logic, makes it harder to maintain item order.

## Consequences

### Positive
- Original currency data preserved for accurate display
- Clear separation between data storage (struct) and presentation (template)
- Subtotal calculations testable in Go code
- Supports future multi-currency scenarios beyond VND/USD
- Maintains backward compatibility with existing fields

### Negative
- Slight increase in struct size (2 fields per line item)
- Controller logic becomes more complex (subtotal calculations)
- Need to ensure OriginalAmount and AmountUSD stay in sync

### Neutral
- Requires migration of existing invoice generation logic
- Template must be updated to use OriginalAmount/OriginalCurrency

## Implementation Notes

1. **Backward Compatibility:** Existing `Amount` and `AmountUSD` fields remain unchanged for compatibility
2. **Validation:** Controller should validate that OriginalCurrency is either "VND" or "USD"
3. **Conversion Logic:** Existing Wise API integration handles USD conversion
4. **Rounding:** All USD amounts should be rounded to 2 decimal places, VND to 0 decimal places
5. **Exchange Rate:** Store the VND→USD exchange rate from Wise API in `ExchangeRate` field

## Related Documents

- Specification: spec-001-data-structure-changes.md
- Specification: spec-004-calculation-logic.md
- Requirements: requirements.md
