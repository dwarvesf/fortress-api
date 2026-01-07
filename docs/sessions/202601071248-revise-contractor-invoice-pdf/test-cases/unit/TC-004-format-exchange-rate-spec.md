# Test Case TC-004: formatExchangeRate Function Specification

**Version:** 1.0
**Date:** 2026-01-07
**Component:** pkg/controller/invoice/contractor_invoice.go
**Function:** formatExchangeRate(rate float64) string

## Purpose

Verify that exchange rates are formatted correctly in the standard format "1 USD = X VND" with proper VND number formatting (period separators, no decimals).

## Function Signature

```go
func formatExchangeRate(rate float64) string
```

This function will be registered in the template FuncMap.

## Test Data Requirements

### Input Parameters
- `rate` (float64): VND per 1 USD exchange rate

### Expected Output
- String formatted as: `"1 USD = {vnd_amount} VND"`
- VND amount with period separators
- No decimal places (rounded to nearest whole number)

## Test Cases

### TC-004-01: Current Rate (Typical)
**Input:** `26269.5`
**Expected Output:** `"1 USD = 26.270 VND"`
**Rationale:** Verify current real-world exchange rate formatting

### TC-004-02: Round Number
**Input:** `25000`
**Expected Output:** `"1 USD = 25.000 VND"`
**Rationale:** Verify formatting without rounding needed

### TC-004-03: Rounding Up
**Input:** `26269.7`
**Expected Output:** `"1 USD = 26.270 VND"`
**Rationale:** Verify rounding up to nearest whole number

### TC-004-04: Rounding Down
**Input:** `26269.4`
**Expected Output:** `"1 USD = 26.269 VND"`
**Rationale:** Verify rounding down to nearest whole number

### TC-004-05: Exact Half (Rounding)
**Input:** `26269.5`
**Expected Output:** `"1 USD = 26.270 VND"` (or `"1 USD = 26.269 VND"` depending on rounding mode)
**Rationale:** Verify rounding behavior at exact halfway point

### TC-004-06: Low Rate (Historical)
**Input:** `15000`
**Expected Output:** `"1 USD = 15.000 VND"`
**Rationale:** Verify formatting with smaller numbers

### TC-004-07: High Rate (Future)
**Input:** `35000.99`
**Expected Output:** `"1 USD = 35.001 VND"`
**Rationale:** Verify formatting with larger numbers

### TC-004-08: Rate with Many Decimals
**Input:** `26269.456789`
**Expected Output:** `"1 USD = 26.269 VND"`
**Rationale:** Verify proper rounding with high precision input

### TC-004-09: Very Low Rate (Edge Case)
**Input:** `100`
**Expected Output:** `"1 USD = 100 VND"`
**Rationale:** Verify formatting without thousands separator

### TC-004-10: Rate = 1 (Edge Case)
**Input:** `1.0`
**Expected Output:** `"1 USD = 1 VND"`
**Rationale:** Verify minimal rate handling

### TC-004-11: Zero Rate (Invalid but Defensive)
**Input:** `0`
**Expected Output:** `"1 USD = 0 VND"`
**Rationale:** Verify no panic with zero (should be caught in validation)

### TC-004-12: Negative Rate (Invalid but Defensive)
**Input:** `-26269`
**Expected Output:** `"1 USD = -26.269 VND"`
**Rationale:** Verify no panic with negative (should be caught in validation)

### TC-004-13: Large Rate (Millions)
**Input:** `1000000`
**Expected Output:** `"1 USD = 1.000.000 VND"`
**Rationale:** Verify separator handling with millions

### TC-004-14: Fractional Rate < 1
**Input:** `0.5`
**Expected Output:** `"1 USD = 1 VND"` (rounds up) or `"1 USD = 0 VND"` (rounds down)
**Rationale:** Verify rounding for sub-unit rates

## Assertion Strategy

For each test case:
1. Call `formatExchangeRate(rate)`
2. Compare result string to expected output using exact string match
3. Verify:
   - Format starts with "1 USD = "
   - VND amount has period separators
   - VND amount has no decimal places
   - Format ends with " VND"
   - Proper rounding applied

## Error Conditions

- Function should handle all float64 values without panicking
- Should not return empty string for any valid float64
- Should maintain consistent spacing: "1 USD = X VND"
- Should not introduce extra whitespace

## Related Specifications

- spec-002-template-functions.md: formatExchangeRate implementation requirements
- ADR-002-currency-formatting-approach.md: Exchange rate display rules
- TC-001: formatVND test specification (related number formatting)

## Dependencies

- `math.Round()` for rounding to whole number
- `formatVNDNumber()` helper for VND number formatting
- `fmt.Sprintf()` for string composition

## Test Implementation Notes

**Test File:** `pkg/controller/invoice/contractor_invoice_formatters_test.go`

**Test Structure:**
```go
func TestFormatExchangeRate(t *testing.T) {
    tests := []struct {
        name     string
        rate     float64
        expected string
    }{
        // Test cases as specified above
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := formatExchangeRate(tt.rate)
            if result != tt.expected {
                t.Errorf("formatExchangeRate(%f) = %s; want %s",
                    tt.rate, result, tt.expected)
            }
        })
    }
}
```

## Usage in Template

```html
<!-- Footer exchange rate footnote -->
<p class="exchange-rate-note">
    *FX Rate {{formatExchangeRate .Invoice.ExchangeRate}}
</p>
```

## Success Criteria

- All test cases pass with exact string match
- No panics or errors for any valid float64 input
- Consistent format: "1 USD = X VND"
- VND amount properly formatted with periods
- Proper rounding behavior
