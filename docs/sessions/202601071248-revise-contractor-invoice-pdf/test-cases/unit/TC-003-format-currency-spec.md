# Test Case TC-003: formatCurrency Function Specification

**Version:** 1.0
**Date:** 2026-01-07
**Component:** pkg/controller/invoice/contractor_invoice.go
**Function:** formatCurrency(amount float64, currency string) string

## Purpose

Verify that the currency dispatcher function correctly routes to the appropriate formatter based on currency code and handles unknown currencies gracefully.

## Function Signature

```go
func formatCurrency(amount float64, currency string) string
```

This function will be registered in the template FuncMap.

## Test Data Requirements

### Input Parameters
- `amount` (float64): Numeric amount to format
- `currency` (string): Currency code ("VND", "USD", or unknown)

### Expected Output
- For "VND": Delegates to `formatVND(amount)`
- For "USD": Delegates to `formatUSD(amount)`
- For unknown: Returns fallback format `"{amount} {currency}"`

## Test Cases

### TC-003-01: VND Currency (Uppercase)
**Input:** `amount=45000000, currency="VND"`
**Expected Output:** `"45.000.000 ₫"`
**Rationale:** Verify correct delegation to formatVND

### TC-003-02: USD Currency (Uppercase)
**Input:** `amount=1234.56, currency="USD"`
**Expected Output:** `"$1,234.56"`
**Rationale:** Verify correct delegation to formatUSD

### TC-003-03: VND Currency (Lowercase)
**Input:** `amount=500000, currency="vnd"`
**Expected Output:** `"500.000 ₫"`
**Rationale:** Verify case-insensitive handling (if implemented) or case-sensitive behavior

### TC-003-04: USD Currency (Lowercase)
**Input:** `amount=100, currency="usd"`
**Expected Output:** `"$100.00"`
**Rationale:** Verify case-insensitive handling (if implemented) or case-sensitive behavior

### TC-003-05: Unknown Currency (EUR)
**Input:** `amount=1000, currency="EUR"`
**Expected Output:** `"1000.00 EUR"`
**Rationale:** Verify fallback formatting for unsupported currencies

### TC-003-06: Unknown Currency (Empty String)
**Input:** `amount=100, currency=""`
**Expected Output:** `"100.00 "`
**Rationale:** Verify handling of empty currency string

### TC-003-07: Unknown Currency (Special Characters)
**Input:** `amount=500, currency="@#$"`
**Expected Output:** `"500.00 @#$"`
**Rationale:** Verify no panics with invalid currency codes

### TC-003-08: VND with Zero Amount
**Input:** `amount=0, currency="VND"`
**Expected Output:** `"0 ₫"`
**Rationale:** Verify delegation preserves zero handling

### TC-003-09: USD with Zero Amount
**Input:** `amount=0, currency="USD"`
**Expected Output:** `"$0.00"`
**Rationale:** Verify delegation preserves zero handling

### TC-003-10: VND with Negative Amount
**Input:** `amount=-1000000, currency="VND"`
**Expected Output:** `"-1.000.000 ₫"`
**Rationale:** Verify delegation preserves negative handling

### TC-003-11: USD with Negative Amount
**Input:** `amount=-500.50, currency="USD"`
**Expected Output:** `"$-500.50"`
**Rationale:** Verify delegation preserves negative handling

### TC-003-12: Mixed Case Currency (VnD)
**Input:** `amount=1000, currency="VnD"`
**Expected Output:** Depends on case sensitivity implementation
**Rationale:** Verify consistent case handling behavior

## Assertion Strategy

For each test case:
1. Call `formatCurrency(amount, currency)`
2. Compare result string to expected output using exact string match
3. Verify:
   - Correct delegation to formatVND for "VND"
   - Correct delegation to formatUSD for "USD"
   - Fallback format for unknown currencies
   - No panics or errors

## Error Conditions

- Function should never panic regardless of input
- Should handle null/empty currency strings gracefully
- Should not modify the amount value during dispatching

## Implementation Notes

### Case Sensitivity Decision Required

The implementation must decide on case sensitivity:

**Option A: Case-Insensitive (Recommended)**
```go
func formatCurrency(amount float64, currency string) string {
    switch strings.ToUpper(currency) {
    case "VND":
        return formatVND(amount)
    case "USD":
        return formatUSD(amount)
    default:
        return fmt.Sprintf("%.2f %s", amount, currency)
    }
}
```

**Option B: Case-Sensitive (Stricter)**
```go
func formatCurrency(amount float64, currency string) string {
    switch currency {
    case "VND":
        return formatVND(amount)
    case "USD":
        return formatUSD(amount)
    default:
        return fmt.Sprintf("%.2f %s", amount, currency)
    }
}
```

**Recommendation:** Use case-insensitive (Option A) for robustness, as Notion data might have inconsistent casing.

## Related Specifications

- spec-002-template-functions.md: formatCurrency implementation requirements
- ADR-002-currency-formatting-approach.md: Formatting rules
- TC-001: formatVND test specification
- TC-002: formatUSD test specification

## Dependencies

- `formatVND()` function
- `formatUSD()` function
- `strings.ToUpper()` (if case-insensitive)
- `fmt.Sprintf()` for fallback

## Test Implementation Notes

**Test File:** `pkg/controller/invoice/contractor_invoice_formatters_test.go`

**Test Structure:**
```go
func TestFormatCurrency(t *testing.T) {
    tests := []struct {
        name     string
        amount   float64
        currency string
        expected string
    }{
        // Test cases as specified above
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := formatCurrency(tt.amount, tt.currency)
            if result != tt.expected {
                t.Errorf("formatCurrency(%f, %s) = %s; want %s",
                    tt.amount, tt.currency, result, tt.expected)
            }
        })
    }
}
```

## Success Criteria

- All test cases pass with exact string match
- Correct delegation to currency-specific formatters
- Graceful fallback for unknown currencies
- No panics for any input combination
- Consistent behavior with case sensitivity decision
