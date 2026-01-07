# Test Case TC-006: Data Structure Population Specification

**Version:** 1.0
**Date:** 2026-01-07
**Component:** pkg/controller/invoice/contractor_invoice.go
**Structs:** ContractorInvoiceLineItem, ContractorInvoiceData

## Purpose

Verify that new data structure fields (OriginalAmount, OriginalCurrency, SubtotalVND, etc.) are correctly populated from Notion PayoutEntry data and calculation results.

## Data Structures

### ContractorInvoiceLineItem (New Fields)
```go
type ContractorInvoiceLineItem struct {
    // ... existing fields ...

    // NEW fields to verify:
    OriginalAmount   float64 // From payout.Amount
    OriginalCurrency string  // From payout.Currency
}
```

### ContractorInvoiceData (New Fields)
```go
type ContractorInvoiceData struct {
    // ... existing fields ...

    // NEW fields to verify:
    SubtotalVND        float64 // Calculated sum of VND items
    SubtotalUSDFromVND float64 // SubtotalVND converted to USD
    SubtotalUSDItems   float64 // Calculated sum of USD items
    SubtotalUSD        float64 // SubtotalUSDFromVND + SubtotalUSDItems
    FXSupport          float64 // Hardcoded 8.0
}
```

## Test Data Requirements

### Mock Notion PayoutEntry
```go
type PayoutEntry struct {
    Amount   float64
    Currency string
    // ... other fields
}
```

## Test Cases

### TC-006-01: LineItem Population from VND Payout
**Input:**
```go
payout := notion.PayoutEntry{
    Amount:        45000000,
    Currency:      "VND",
    SourceType:    "Service Fee",
    // ... other fields
}
```
**Expected LineItem:**
```go
ContractorInvoiceLineItem{
    OriginalAmount:   45000000.0,
    OriginalCurrency: "VND",
    AmountUSD:        1713.41,  // Converted
    Type:             "Service Fee",
    // ... other fields populated correctly
}
```
**Rationale:** Verify VND payout data preservation

### TC-006-02: LineItem Population from USD Payout
**Input:**
```go
payout := notion.PayoutEntry{
    Amount:        1500.00,
    Currency:      "USD",
    SourceType:    "Service Fee",
    // ... other fields
}
```
**Expected LineItem:**
```go
ContractorInvoiceLineItem{
    OriginalAmount:   1500.0,
    OriginalCurrency: "USD",
    AmountUSD:        1500.0,  // No conversion
    Type:             "Service Fee",
    // ... other fields populated correctly
}
```
**Rationale:** Verify USD payout data preservation

### TC-006-03: LineItem Population with Commission Fields
**Input:**
```go
payout := notion.PayoutEntry{
    Amount:            100,
    Currency:          "USD",
    SourceType:        "Commission",
    CommissionRole:    "Backend Developer",
    CommissionProject: "Project Alpha",
    // ... other fields
}
```
**Expected LineItem:**
```go
ContractorInvoiceLineItem{
    OriginalAmount:    100.0,
    OriginalCurrency:  "USD",
    AmountUSD:         100.0,
    Type:              "Commission",
    CommissionRole:    "Backend Developer",
    CommissionProject: "Project Alpha",
    // ... other fields populated correctly
}
```
**Rationale:** Verify commission-specific field population

### TC-006-04: InvoiceData Population (All VND)
**Input:**
- Multiple VND line items
- Calculated subtotals
**Expected InvoiceData:**
```go
ContractorInvoiceData{
    // Existing fields populated...

    // NEW fields:
    SubtotalVND:        45500000.0,
    SubtotalUSDFromVND: 1731.89,
    SubtotalUSDItems:   0.0,
    SubtotalUSD:        1731.89,
    FXSupport:          8.0,
    Total:              1739.89,
    TotalUSD:           1739.89,
    ExchangeRate:       26269.0,
}
```
**Rationale:** Verify invoice-level field population for VND scenario

### TC-006-05: InvoiceData Population (All USD)
**Input:**
- Multiple USD line items
- Calculated subtotals
**Expected InvoiceData:**
```go
ContractorInvoiceData{
    // Existing fields populated...

    // NEW fields:
    SubtotalVND:        0.0,
    SubtotalUSDFromVND: 0.0,
    SubtotalUSDItems:   1600.0,
    SubtotalUSD:        1600.0,
    FXSupport:          8.0,
    Total:              1608.0,
    TotalUSD:           1608.0,
    ExchangeRate:       1.0,
}
```
**Rationale:** Verify invoice-level field population for USD scenario

### TC-006-06: InvoiceData Population (Mixed)
**Input:**
- Mixed VND and USD line items
- Calculated subtotals
**Expected InvoiceData:**
```go
ContractorInvoiceData{
    // Existing fields populated...

    // NEW fields:
    SubtotalVND:        45500000.0,
    SubtotalUSDFromVND: 1731.89,
    SubtotalUSDItems:   100.0,
    SubtotalUSD:        1831.89,
    FXSupport:          8.0,
    Total:              1839.89,
    TotalUSD:           1839.89,
    ExchangeRate:       26269.0,
}
```
**Rationale:** Verify invoice-level field population for mixed scenario

### TC-006-07: Backward Compatibility - Existing Fields Unchanged
**Input:** Any payout data
**Expected Behavior:**
- All existing fields (Title, Description, Hours, Rate, Amount, AmountUSD) remain populated
- Existing field values are not modified by new field additions
- Currency field remains "USD" for payment processing
**Rationale:** Verify backward compatibility with existing code

### TC-006-08: Empty OriginalCurrency Handling
**Input:**
```go
payout := notion.PayoutEntry{
    Amount:   1000,
    Currency: "",  // Empty
}
```
**Expected Behavior:** Return validation error (invalid currency)
**Rationale:** Verify validation of currency field

### TC-006-09: Null/Missing Currency Handling
**Input:**
```go
payout := notion.PayoutEntry{
    Amount:   1000,
    // Currency field not set
}
```
**Expected Behavior:** Return validation error or use default
**Rationale:** Verify handling of missing data

### TC-006-10: Zero Amount Preservation
**Input:**
```go
payout := notion.PayoutEntry{
    Amount:   0,
    Currency: "USD",
}
```
**Expected LineItem:**
```go
ContractorInvoiceLineItem{
    OriginalAmount:   0.0,
    OriginalCurrency: "USD",
    AmountUSD:        0.0,
}
```
**Rationale:** Verify zero amounts are preserved correctly

### TC-006-11: High Precision Amount Preservation
**Input:**
```go
payout := notion.PayoutEntry{
    Amount:   1234.567890,
    Currency: "USD",
}
```
**Expected LineItem:**
```go
ContractorInvoiceLineItem{
    OriginalAmount:   1234.567890,  // Full precision preserved
    OriginalCurrency: "USD",
    AmountUSD:        1234.57,      // Rounded for display
}
```
**Rationale:** Verify precision handling in OriginalAmount vs display fields

### TC-006-12: Multiple Payouts to Multiple LineItems
**Input:**
```go
payouts := []notion.PayoutEntry{
    {Amount: 10000000, Currency: "VND", SourceType: "Service Fee"},
    {Amount: 5000000, Currency: "VND", SourceType: "Refund"},
    {Amount: 100, Currency: "USD", SourceType: "Commission"},
}
```
**Expected:** 3 LineItems created with correct field population
**Rationale:** Verify one-to-one mapping preservation

## Assertion Strategy

For each test case:
1. Create mock Notion PayoutEntry data
2. Call line item creation logic
3. Assert all new fields are populated:
   - `OriginalAmount` == `payout.Amount`
   - `OriginalCurrency` == `payout.Currency`
4. Call invoice data creation logic
5. Assert all new invoice-level fields are populated correctly:
   - SubtotalVND, SubtotalUSDFromVND, SubtotalUSDItems
   - SubtotalUSD, FXSupport, Total, TotalUSD, ExchangeRate
6. Verify existing fields remain unchanged

## Data Integrity Checks

### Invariants to Verify

1. **Currency Consistency:**
   - `OriginalCurrency` must match `payout.Currency`

2. **Amount Consistency:**
   - `OriginalAmount` must match `payout.Amount`
   - For USD: `AmountUSD` should equal `OriginalAmount`
   - For VND: `AmountUSD` should equal `OriginalAmount / ExchangeRate`

3. **Subtotal Consistency:**
   - `SubtotalVND` == sum of all `OriginalAmount` where `OriginalCurrency == "VND"`
   - `SubtotalUSDItems` == sum of all `AmountUSD` where `OriginalCurrency == "USD"`
   - `SubtotalUSD` == `SubtotalUSDFromVND + SubtotalUSDItems`

4. **Total Consistency:**
   - `TotalUSD` == `SubtotalUSD + FXSupport`
   - `Total` == `TotalUSD` (for backward compatibility)

5. **FX Support:**
   - `FXSupport` must always be 8.0 (current hardcoded value)

## Error Conditions

- Missing Currency field → validation error
- Invalid Currency code (not VND/USD) → validation error
- Negative Amount → validation error
- Mismatched AmountUSD vs OriginalAmount+Currency → data integrity error

## Related Specifications

- spec-001-data-structure-changes.md: Field definitions
- ADR-001-data-structure-multi-currency.md: Design rationale
- TC-005: Calculation logic specification

## Dependencies

- Notion PayoutEntry data structure
- Wise API for currency conversion
- Validation logic

## Test Implementation Notes

**Test File:** `pkg/controller/invoice/contractor_invoice_data_test.go`

**Test Structure:**
```go
func TestLineItemPopulation(t *testing.T) {
    tests := []struct {
        name             string
        payout           notion.PayoutEntry
        expectedLineItem ContractorInvoiceLineItem
        expectError      bool
    }{
        // Test cases as specified above
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Create line item from payout
            // Assert field values match expected
        })
    }
}

func TestInvoiceDataPopulation(t *testing.T) {
    tests := []struct {
        name              string
        lineItems         []ContractorInvoiceLineItem
        exchangeRate      float64
        expectedInvoiceData ContractorInvoiceData
    }{
        // Test cases as specified above
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Create invoice data
            // Assert all new fields populated correctly
        })
    }
}
```

## Success Criteria

- All new fields populated correctly from source data
- No data loss during transformation
- Existing fields remain unchanged (backward compatibility)
- Data integrity invariants hold
- Proper validation of invalid data
- No panics for edge cases
