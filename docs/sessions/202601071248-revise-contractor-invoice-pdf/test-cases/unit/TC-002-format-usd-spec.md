# Test Case TC-002: formatUSD Function Specification

**Version:** 1.0
**Date:** 2026-01-07
**Component:** pkg/controller/invoice/contractor_invoice.go
**Function:** formatUSD(amount float64) string

## Purpose

Verify that USD amounts are formatted correctly with comma separators, 2 decimal places, and the dollar symbol positioned before the amount.

## Function Signature

```go
func formatUSD(amount float64) string
```

## Test Data Requirements

### Input Parameters
- `amount` (float64): Numeric amount to format (can be positive, negative, zero, or fractional)

### Expected Output
- String formatted as: `${integer_part}.{decimal_part}`
- Thousands separator: `,` (comma)
- Always 2 decimal places
- Symbol `$` before amount with no space

## Test Cases

### TC-002-01: Zero Amount
**Input:** `0`
**Expected Output:** `"$0.00"`
**Rationale:** Baseline case with proper zero formatting

### TC-002-02: Cents Only
**Input:** `0.99`
**Expected Output:** `"$0.99"`
**Rationale:** Verify decimal formatting without thousands separator

### TC-002-03: Whole Dollars (No Cents)
**Input:** `100`
**Expected Output:** `"$100.00"`
**Rationale:** Verify .00 suffix for whole dollar amounts

### TC-002-04: Thousands (One Separator)
**Input:** `1234.56`
**Expected Output:** `"$1,234.56"`
**Rationale:** Verify single comma separator with decimals

### TC-002-05: Millions (Two Separators)
**Input:** `1000000`
**Expected Output:** `"$1,000,000.00"`
**Rationale:** Verify multiple comma separators

### TC-002-06: Irregular Amount
**Input:** `12345.67`
**Expected Output:** `"$12,345.67"`
**Rationale:** Verify separator placement with non-uniform grouping

### TC-002-07: Billions (Three Separators)
**Input:** `1000000000`
**Expected Output:** `"$1,000,000,000.00"`
**Rationale:** Verify handling of very large amounts

### TC-002-08: Rounding Up (Third Decimal >= 5)
**Input:** `1234.567`
**Expected Output:** `"$1,234.57"`
**Rationale:** Verify rounding up to 2 decimal places

### TC-002-09: Rounding Down (Third Decimal < 5)
**Input:** `1234.564`
**Expected Output:** `"$1,234.56"`
**Rationale:** Verify rounding down to 2 decimal places

### TC-002-10: Exact Half Rounding
**Input:** `1234.565`
**Expected Output:** `"$1,234.57"` (banker's rounding) or `"$1,234.56"` (depends on implementation)
**Rationale:** Verify rounding behavior at exact halfway point

### TC-002-11: Negative Amount (Edge Case)
**Input:** `-100.50`
**Expected Output:** `"$-100.50"`
**Rationale:** Verify handling of negative values

### TC-002-12: Negative Large Amount
**Input:** `-1000000.99`
**Expected Output:** `"$-1,000,000.99"`
**Rationale:** Verify negative sign with separators and decimals

### TC-002-13: Very Small Amount
**Input:** `0.01`
**Expected Output:** `"$0.01"`
**Rationale:** Verify minimum positive monetary unit (penny)

### TC-002-14: Submicro Amount (Rounds to Zero)
**Input:** `0.004`
**Expected Output:** `"$0.00"`
**Rationale:** Verify amounts < $0.005 round to zero

### TC-002-15: Large Decimal Precision
**Input:** `123.456789`
**Expected Output:** `"$123.46"`
**Rationale:** Verify proper rounding with many decimal places

### TC-002-16: Typical Invoice Amount
**Input:** `1731.89`
**Expected Output:** `"$1,731.89"`
**Rationale:** Real-world example from requirements

### TC-002-17: Maximum Float64 (Boundary)
**Input:** `1.7976931348623157e+308`
**Expected Output:** Should handle without overflow
**Rationale:** Verify boundary condition for float64 max value

## Assertion Strategy

For each test case:
1. Call `formatUSD(input)`
2. Compare result string to expected output using exact string match
3. Verify:
   - Correct symbol placement (before amount with no space)
   - Comma separators every 3 digits from right
   - Exactly 2 decimal places
   - Proper rounding behavior

## Error Conditions

- Function should handle all float64 values without panicking
- Should not return empty string for any valid float64
- Should not introduce extra whitespace
- Decimal separator must always be period (.)

## Related Specifications

- spec-002-template-functions.md: formatUSD implementation requirements
- ADR-002-currency-formatting-approach.md: Formatting rules

## Dependencies

- `math.Round()` for rounding to 2 decimal places
- `formatUSDNumber()` helper for separator and decimal formatting
- `fmt.Sprintf()` for string formatting

## Test Implementation Notes

**Test File:** `pkg/controller/invoice/contractor_invoice_formatters_test.go`

**Test Structure:**
```go
func TestFormatUSD(t *testing.T) {
    tests := []struct {
        name     string
        amount   float64
        expected string
    }{
        // Test cases as specified above
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := formatUSD(tt.amount)
            if result != tt.expected {
                t.Errorf("formatUSD(%f) = %s; want %s", tt.amount, result, tt.expected)
            }
        })
    }
}
```

## Success Criteria

- All test cases pass with exact string match
- No panics or errors for any valid float64 input
- Consistent formatting across all amount ranges
- Rounding follows Go's standard rounding behavior
- Always displays exactly 2 decimal places
