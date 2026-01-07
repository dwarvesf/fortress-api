# Test Case TC-005: Subtotal Calculation Logic Specification

**Version:** 1.0
**Date:** 2026-01-07
**Component:** pkg/controller/invoice/contractor_invoice.go
**Function:** Subtotal calculation in GenerateContractorInvoice

## Purpose

Verify that multi-currency subtotal calculations are correct for all scenarios including VND-only, USD-only, mixed currencies, and edge cases.

## Calculation Algorithm

```
1. Sum line items by currency:
   - SubtotalVND = sum of all items where OriginalCurrency == "VND"
   - SubtotalUSDItems = sum of all items where OriginalCurrency == "USD"

2. Round subtotals:
   - SubtotalVND = Round(SubtotalVND) to 0 decimals
   - SubtotalUSDItems = Round(SubtotalUSDItems * 100) / 100 to 2 decimals

3. Convert VND to USD (if SubtotalVND > 0):
   - SubtotalUSDFromVND = Wise.Convert(SubtotalVND, "VND", "USD")
   - ExchangeRate = rate from Wise API

4. Calculate combined subtotal:
   - SubtotalUSD = SubtotalUSDFromVND + SubtotalUSDItems
   - Round to 2 decimals

5. Add FX Support:
   - FXSupport = 8.00 (hardcoded)
   - TotalUSD = SubtotalUSD + FXSupport
   - Round to 2 decimals
```

## Test Data Requirements

### Mock Data Structures

**ContractorInvoiceLineItem:**
```go
type ContractorInvoiceLineItem struct {
    OriginalAmount   float64
    OriginalCurrency string
    AmountUSD        float64
    // ... other fields
}
```

**Expected Outputs:**
```go
type CalculationResult struct {
    SubtotalVND        float64
    SubtotalUSDFromVND float64
    SubtotalUSDItems   float64
    SubtotalUSD        float64
    FXSupport          float64
    TotalUSD           float64
    ExchangeRate       float64
}
```

## Test Cases

### TC-005-01: All VND Items
**Input:**
```go
lineItems := []ContractorInvoiceLineItem{
    {OriginalAmount: 45000000, OriginalCurrency: "VND", AmountUSD: 1713.41},
    {OriginalAmount: 500000, OriginalCurrency: "VND", AmountUSD: 19.03},
}
exchangeRate := 26269.0
```
**Expected Output:**
```go
SubtotalVND:        45500000.0
SubtotalUSDFromVND: 1731.89  // 45500000 / 26269 = 1731.89
SubtotalUSDItems:   0.0
SubtotalUSD:        1731.89
FXSupport:          8.0
TotalUSD:           1739.89
ExchangeRate:       26269.0
```
**Rationale:** Verify VND-only calculation path

### TC-005-02: All USD Items
**Input:**
```go
lineItems := []ContractorInvoiceLineItem{
    {OriginalAmount: 1500, OriginalCurrency: "USD", AmountUSD: 1500},
    {OriginalAmount: 100, OriginalCurrency: "USD", AmountUSD: 100},
}
exchangeRate := 1.0 // No conversion needed
```
**Expected Output:**
```go
SubtotalVND:        0.0
SubtotalUSDFromVND: 0.0
SubtotalUSDItems:   1600.0
SubtotalUSD:        1600.0
FXSupport:          8.0
TotalUSD:           1608.0
ExchangeRate:       1.0
```
**Rationale:** Verify USD-only calculation path (no conversion)

### TC-005-03: Mixed Currencies
**Input:**
```go
lineItems := []ContractorInvoiceLineItem{
    {OriginalAmount: 45000000, OriginalCurrency: "VND", AmountUSD: 1713.41},
    {OriginalAmount: 500000, OriginalCurrency: "VND", AmountUSD: 19.03},
    {OriginalAmount: 100, OriginalCurrency: "USD", AmountUSD: 100},
}
exchangeRate := 26269.0
```
**Expected Output:**
```go
SubtotalVND:        45500000.0
SubtotalUSDFromVND: 1731.89
SubtotalUSDItems:   100.0
SubtotalUSD:        1831.89
FXSupport:          8.0
TotalUSD:           1839.89
ExchangeRate:       26269.0
```
**Rationale:** Verify mixed currency calculation

### TC-005-04: Single Item (VND)
**Input:**
```go
lineItems := []ContractorInvoiceLineItem{
    {OriginalAmount: 10000000, OriginalCurrency: "VND", AmountUSD: 380.64},
}
exchangeRate := 26269.0
```
**Expected Output:**
```go
SubtotalVND:        10000000.0
SubtotalUSDFromVND: 380.64
SubtotalUSDItems:   0.0
SubtotalUSD:        380.64
FXSupport:          8.0
TotalUSD:           388.64
ExchangeRate:       26269.0
```
**Rationale:** Verify minimal VND scenario

### TC-005-05: Single Item (USD)
**Input:**
```go
lineItems := []ContractorInvoiceLineItem{
    {OriginalAmount: 500, OriginalCurrency: "USD", AmountUSD: 500},
}
exchangeRate := 1.0
```
**Expected Output:**
```go
SubtotalVND:        0.0
SubtotalUSDFromVND: 0.0
SubtotalUSDItems:   500.0
SubtotalUSD:        500.0
FXSupport:          8.0
TotalUSD:           508.0
ExchangeRate:       1.0
```
**Rationale:** Verify minimal USD scenario

### TC-005-06: Empty Line Items
**Input:**
```go
lineItems := []ContractorInvoiceLineItem{}
```
**Expected Output:**
```go
SubtotalVND:        0.0
SubtotalUSDFromVND: 0.0
SubtotalUSDItems:   0.0
SubtotalUSD:        0.0
FXSupport:          8.0
TotalUSD:           8.0
ExchangeRate:       1.0
```
**Rationale:** Verify handling of empty invoice (edge case)

### TC-005-07: Zero Amount Items
**Input:**
```go
lineItems := []ContractorInvoiceLineItem{
    {OriginalAmount: 0, OriginalCurrency: "VND", AmountUSD: 0},
    {OriginalAmount: 0, OriginalCurrency: "USD", AmountUSD: 0},
}
```
**Expected Output:**
```go
SubtotalVND:        0.0
SubtotalUSDFromVND: 0.0
SubtotalUSDItems:   0.0
SubtotalUSD:        0.0
FXSupport:          8.0
TotalUSD:           8.0
ExchangeRate:       1.0
```
**Rationale:** Verify zero amounts don't break calculations

### TC-005-08: VND Fractional Rounding
**Input:**
```go
lineItems := []ContractorInvoiceLineItem{
    {OriginalAmount: 1234567.89, OriginalCurrency: "VND", AmountUSD: 47.0},
}
exchangeRate := 26269.0
```
**Expected Output:**
```go
SubtotalVND:        1234568.0  // Rounded to whole number
SubtotalUSDFromVND: 47.0       // 1234568 / 26269
SubtotalUSDItems:   0.0
SubtotalUSD:        47.0
FXSupport:          8.0
TotalUSD:           55.0
ExchangeRate:       26269.0
```
**Rationale:** Verify VND rounding to nearest whole number

### TC-005-09: USD Fractional Rounding
**Input:**
```go
lineItems := []ContractorInvoiceLineItem{
    {OriginalAmount: 123.456789, OriginalCurrency: "USD", AmountUSD: 123.456789},
}
```
**Expected Output:**
```go
SubtotalVND:        0.0
SubtotalUSDFromVND: 0.0
SubtotalUSDItems:   123.46  // Rounded to 2 decimals
SubtotalUSD:        123.46
FXSupport:          8.0
TotalUSD:           131.46
ExchangeRate:       1.0
```
**Rationale:** Verify USD rounding to 2 decimal places

### TC-005-10: Large VND Amount (Billions)
**Input:**
```go
lineItems := []ContractorInvoiceLineItem{
    {OriginalAmount: 1000000000, OriginalCurrency: "VND", AmountUSD: 38063.61},
}
exchangeRate := 26269.0
```
**Expected Output:**
```go
SubtotalVND:        1000000000.0
SubtotalUSDFromVND: 38063.61  // 1000000000 / 26269
SubtotalUSDItems:   0.0
SubtotalUSD:        38063.61
FXSupport:          8.0
TotalUSD:           38071.61
ExchangeRate:       26269.0
```
**Rationale:** Verify handling of very large amounts

### TC-005-11: Multiple Items Same Currency
**Input:**
```go
lineItems := []ContractorInvoiceLineItem{
    {OriginalAmount: 10000000, OriginalCurrency: "VND", AmountUSD: 380.64},
    {OriginalAmount: 5000000, OriginalCurrency: "VND", AmountUSD: 190.32},
    {OriginalAmount: 2000000, OriginalCurrency: "VND", AmountUSD: 76.13},
}
exchangeRate := 26269.0
```
**Expected Output:**
```go
SubtotalVND:        17000000.0
SubtotalUSDFromVND: 647.09  // 17000000 / 26269
SubtotalUSDItems:   0.0
SubtotalUSD:        647.09
FXSupport:          8.0
TotalUSD:           655.09
ExchangeRate:       26269.0
```
**Rationale:** Verify correct summation of multiple items

### TC-005-12: Mixed with Small Amounts
**Input:**
```go
lineItems := []ContractorInvoiceLineItem{
    {OriginalAmount: 100000, OriginalCurrency: "VND", AmountUSD: 3.81},
    {OriginalAmount: 0.50, OriginalCurrency: "USD", AmountUSD: 0.50},
}
exchangeRate := 26269.0
```
**Expected Output:**
```go
SubtotalVND:        100000.0
SubtotalUSDFromVND: 3.81
SubtotalUSDItems:   0.50
SubtotalUSD:        4.31
FXSupport:          8.0
TotalUSD:           12.31
ExchangeRate:       26269.0
```
**Rationale:** Verify handling of small amounts

## Validation Test Cases

### TC-005-V1: Invalid Currency Code
**Input:**
```go
lineItems := []ContractorInvoiceLineItem{
    {OriginalAmount: 1000, OriginalCurrency: "EUR", AmountUSD: 1000},
}
```
**Expected Behavior:** Return error "invalid currency: EUR (must be VND or USD)"
**Rationale:** Verify currency validation

### TC-005-V2: Negative Amount
**Input:**
```go
lineItems := []ContractorInvoiceLineItem{
    {OriginalAmount: -500, OriginalCurrency: "USD", AmountUSD: -500},
}
```
**Expected Behavior:** Return error "negative amount for line item"
**Rationale:** Verify amount validation (negatives should be Refund type)

### TC-005-V3: Exchange Rate API Failure
**Input:** VND items exist, but Wise API returns error
**Expected Behavior:** Return error "failed to convert VND subtotal to USD"
**Rationale:** Verify error propagation from Wise API

### TC-005-V4: Zero Exchange Rate
**Input:** Wise API returns rate = 0
**Expected Behavior:** Return error "invalid exchange rate: 0 (must be > 0)"
**Rationale:** Verify exchange rate validation

### TC-005-V5: Negative Exchange Rate
**Input:** Wise API returns rate = -26269
**Expected Behavior:** Return error "invalid exchange rate: -26269 (must be > 0)"
**Rationale:** Verify exchange rate validation

## Assertion Strategy

For each test case:
1. Create line items with specified values
2. Mock Wise API response (if applicable)
3. Call calculation logic
4. Assert each output field matches expected value:
   - SubtotalVND (exact match for VND)
   - SubtotalUSDFromVND (2 decimal precision)
   - SubtotalUSDItems (2 decimal precision)
   - SubtotalUSD (2 decimal precision)
   - FXSupport (exact match: 8.0)
   - TotalUSD (2 decimal precision)
   - ExchangeRate (4 decimal precision)

## Error Conditions

- Invalid currency code → return error
- Negative amounts → return error
- Exchange rate <= 0 → return error
- Wise API failure → return error with context
- Should never panic

## Related Specifications

- spec-001-data-structure-changes.md: Data structures
- spec-004-calculation-logic.md: Calculation algorithm
- ADR-001-data-structure-multi-currency.md: Design rationale

## Dependencies

- Wise API for currency conversion
- `math.Round()` for rounding
- Validation logic for currency codes and amounts

## Test Implementation Notes

**Test File:** `pkg/controller/invoice/contractor_invoice_calculations_test.go`

**Test Structure:**
```go
func TestCalculateSubtotals(t *testing.T) {
    tests := []struct {
        name                   string
        lineItems              []ContractorInvoiceLineItem
        mockExchangeRate       float64
        mockWiseError          error
        expectedSubtotalVND    float64
        expectedSubtotalUSD    float64
        expectedFXSupport      float64
        expectedTotal          float64
        expectedError          string
    }{
        // Test cases as specified above
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Mock Wise API
            // Call calculation function
            // Assert results or error
        })
    }
}
```

## Success Criteria

- All calculation test cases produce correct results within 0.01 precision
- All validation test cases return appropriate errors
- No panics for any input combination
- Consistent rounding behavior
- Correct handling of edge cases (zero, empty, large amounts)
