# Test Case TC-001: formatVND Function Specification

**Version:** 1.0
**Date:** 2026-01-07
**Component:** pkg/controller/invoice/contractor_invoice.go
**Function:** formatVND(amount float64) string

## Purpose

Verify that VND amounts are formatted correctly with period separators, no decimal places, and the dong symbol positioned after the amount.

## Function Signature

```go
func formatVND(amount float64) string
```

## Test Data Requirements

### Input Parameters
- `amount` (float64): Numeric amount to format (can be positive, negative, zero, or fractional)

### Expected Output
- String formatted as: `{integer_part} ₫`
- Thousands separator: `.` (period)
- No decimal places
- Symbol `₫` after amount with space

## Test Cases

### TC-001-01: Zero Amount
**Input:** `0`
**Expected Output:** `"0 ₫"`
**Rationale:** Baseline case for minimal value

### TC-001-02: Small Amount (No Separators)
**Input:** `100`
**Expected Output:** `"100 ₫"`
**Rationale:** Verify formatting without thousands separators

### TC-001-03: Thousands (One Separator)
**Input:** `500000`
**Expected Output:** `"500.000 ₫"`
**Rationale:** Verify single period separator at correct position

### TC-001-04: Millions (Two Separators)
**Input:** `45000000`
**Expected Output:** `"45.000.000 ₫"`
**Rationale:** Verify multiple period separators for large amounts

### TC-001-05: Irregular Amount (Multiple Groups)
**Input:** `1234567`
**Expected Output:** `"1.234.567 ₫"`
**Rationale:** Verify separator placement with non-uniform grouping

### TC-001-06: Billions (Three Separators)
**Input:** `1000000000`
**Expected Output:** `"1.000.000.000 ₫"`
**Rationale:** Verify handling of very large amounts

### TC-001-07: Fractional Amount (Rounding Required)
**Input:** `1234.56`
**Expected Output:** `"1.235 ₫"`
**Rationale:** Verify rounding to nearest whole number (VND has no minor units)

### TC-001-08: Half Rounding Down
**Input:** `999.4`
**Expected Output:** `"999 ₫"`
**Rationale:** Verify rounding behavior for < 0.5

### TC-001-09: Half Rounding Up
**Input:** `999.5`
**Expected Output:** `"1.000 ₫"`
**Rationale:** Verify rounding behavior for >= 0.5

### TC-001-10: Negative Amount (Edge Case)
**Input:** `-1000`
**Expected Output:** `"-1.000 ₫"`
**Rationale:** Verify handling of negative values (edge case for refunds)

### TC-001-11: Negative Large Amount
**Input:** `-45000000`
**Expected Output:** `"-45.000.000 ₫"`
**Rationale:** Verify negative sign with multiple separators

### TC-001-12: Very Small Fractional (Rounds to Zero)
**Input:** `0.4`
**Expected Output:** `"0 ₫"`
**Rationale:** Verify amounts < 0.5 round to zero

### TC-001-13: Maximum Float64 (Boundary)
**Input:** `1.7976931348623157e+308`
**Expected Output:** Should handle without overflow
**Rationale:** Verify boundary condition for float64 max value

### TC-001-14: Minimum Positive (Near Zero)
**Input:** `0.1`
**Expected Output:** `"0 ₫"`
**Rationale:** Verify minimal positive values round to zero

## Assertion Strategy

For each test case:
1. Call `formatVND(input)`
2. Compare result string to expected output using exact string match
3. Verify:
   - Correct symbol placement (after amount with space)
   - Period separators every 3 digits from right
   - No decimal places
   - Proper rounding behavior

## Error Conditions

- Function should handle all float64 values without panicking
- Should not return empty string for any valid float64
- Should not introduce extra whitespace

## Related Specifications

- spec-002-template-functions.md: formatVND implementation requirements
- ADR-002-currency-formatting-approach.md: Formatting rules

## Dependencies

- `math.Round()` for rounding to whole number
- `formatVNDNumber()` helper for separator insertion
- `fmt.Sprintf()` for string formatting

## Test Implementation Notes

**Test File:** `pkg/controller/invoice/contractor_invoice_formatters_test.go`

**Test Structure:**
```go
func TestFormatVND(t *testing.T) {
    tests := []struct {
        name     string
        amount   float64
        expected string
    }{
        // Test cases as specified above
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := formatVND(tt.amount)
            if result != tt.expected {
                t.Errorf("formatVND(%f) = %s; want %s", tt.amount, result, tt.expected)
            }
        })
    }
}
```

## Success Criteria

- All test cases pass with exact string match
- No panics or errors for any valid float64 input
- Consistent formatting across all amount ranges
- Rounding follows standard Go math.Round behavior
