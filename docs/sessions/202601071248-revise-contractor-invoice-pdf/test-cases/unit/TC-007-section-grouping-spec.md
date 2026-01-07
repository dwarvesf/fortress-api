# Test Case TC-007: Section Grouping Helper Functions Specification

**Version:** 1.0
**Date:** 2026-01-07
**Component:** pkg/controller/invoice/contractor_invoice.go
**Functions:** isSectionHeader, isServiceFee

## Purpose

Verify that template helper functions correctly identify line item types for proper section rendering in the invoice PDF.

## Function Signatures

```go
func isSectionHeader(itemType string) bool
func isServiceFee(itemType string) bool
```

These functions will be registered in the template FuncMap.

## Business Logic

### Section Types

1. **Service Fee (Aggregated):**
   - Display as single row with total
   - Show individual descriptions below
   - Grouped by role/project

2. **Refund (Section Header):**
   - Display as section header
   - Itemize each refund below

3. **Commission/Bonus (Section Header):**
   - Display as section header
   - Itemize each bonus below

## Test Cases for isSectionHeader

### TC-007-01: Refund Type
**Input:** `"Refund"`
**Expected Output:** `true`
**Rationale:** Refund should be rendered as section header

### TC-007-02: Commission Type
**Input:** `"Commission"`
**Expected Output:** `true`
**Rationale:** Commission should be rendered as section header

### TC-007-03: Service Fee Type
**Input:** `"Service Fee"`
**Expected Output:** `false`
**Rationale:** Service Fee has special aggregated rendering (not section header)

### TC-007-04: Contractor Payroll Type
**Input:** `"Contractor Payroll"`
**Expected Output:** `false`
**Rationale:** Regular line item, not section header

### TC-007-05: Empty String
**Input:** `""`
**Expected Output:** `false`
**Rationale:** Empty type should not be section header

### TC-007-06: Unknown Type
**Input:** `"Unknown"`
**Expected Output:** `false`
**Rationale:** Unknown types should not be section headers

### TC-007-07: Case Sensitivity - Lowercase
**Input:** `"refund"`
**Expected Output:** `false` (if case-sensitive) or `true` (if case-insensitive)
**Rationale:** Verify case handling behavior
**Note:** Recommend case-sensitive to match Notion data consistency

### TC-007-08: Case Sensitivity - Mixed Case
**Input:** `"REFUND"`
**Expected Output:** `false` (if case-sensitive) or `true` (if case-insensitive)
**Rationale:** Verify case handling behavior

### TC-007-09: Whitespace Handling
**Input:** `" Refund "`
**Expected Output:** `false` (if strict) or `true` (if trimmed)
**Rationale:** Verify whitespace handling
**Note:** Recommend trimming for robustness

### TC-007-10: Special Characters
**Input:** `"Refund@"`
**Expected Output:** `false`
**Rationale:** Verify exact match requirement

## Test Cases for isServiceFee

### TC-007-11: Service Fee Type
**Input:** `"Service Fee"`
**Expected Output:** `true`
**Rationale:** Service Fee should be identified correctly

### TC-007-12: Refund Type
**Input:** `"Refund"`
**Expected Output:** `false`
**Rationale:** Refund is not a service fee

### TC-007-13: Commission Type
**Input:** `"Commission"`
**Expected Output:** `false`
**Rationale:** Commission is not a service fee

### TC-007-14: Contractor Payroll Type
**Input:** `"Contractor Payroll"`
**Expected Output:** `false`
**Rationale:** Contractor Payroll is not a service fee

### TC-007-15: Empty String
**Input:** `""`
**Expected Output:** `false`
**Rationale:** Empty type is not a service fee

### TC-007-16: Case Sensitivity - Lowercase
**Input:** `"service fee"`
**Expected Output:** `false` (if case-sensitive) or `true` (if case-insensitive)
**Rationale:** Verify case handling behavior

### TC-007-17: Case Sensitivity - Mixed Case
**Input:** `"SERVICE FEE"`
**Expected Output:** `false` (if case-sensitive) or `true` (if case-insensitive)
**Rationale:** Verify case handling behavior

### TC-007-18: Whitespace Handling
**Input:** `" Service Fee "`
**Expected Output:** `false` (if strict) or `true` (if trimmed)
**Rationale:** Verify whitespace handling

### TC-007-19: Partial Match
**Input:** `"Service"`
**Expected Output:** `false`
**Rationale:** Verify exact match requirement

### TC-007-20: Substring Match
**Input:** `"Service Fee Extra"`
**Expected Output:** `false`
**Rationale:** Verify exact match requirement

## Assertion Strategy

For each test case:
1. Call function with input string
2. Assert boolean result matches expected value
3. Verify no panics or errors
4. Verify consistent behavior across calls

## Implementation Recommendations

### Recommended Implementation (Case-Sensitive, No Trim)

```go
func isSectionHeader(itemType string) bool {
    return itemType == "Refund" || itemType == "Commission"
}

func isServiceFee(itemType string) bool {
    return itemType == "Service Fee"
}
```

**Rationale:**
- Notion data is consistent (always proper case)
- Simple exact match is fastest
- No ambiguity in matching logic
- Fails fast on unexpected data

### Alternative Implementation (Case-Insensitive, Trimmed)

```go
func isSectionHeader(itemType string) bool {
    t := strings.TrimSpace(strings.ToLower(itemType))
    return t == "refund" || t == "commission"
}

func isServiceFee(itemType string) bool {
    return strings.TrimSpace(strings.ToLower(itemType)) == "service fee"
}
```

**Rationale:**
- More robust to data variations
- Handles whitespace and case issues
- Slightly slower due to string operations

## Template Usage Examples

```html
<!-- In template -->
{{range .LineItems}}
    {{if isSectionHeader .Type}}
        <!-- Render as section header row -->
        <tr class="section-header">
            <td colspan="3"><strong>{{.Title}}</strong></td>
        </tr>
    {{else if isServiceFee .Type}}
        <!-- Render as aggregated service fee -->
        <tr class="service-fee-total">
            <td>Service Fee</td>
            <td>{{formatCurrency .OriginalAmount .OriginalCurrency}}</td>
        </tr>
        <tr class="service-fee-details">
            <td colspan="2">{{.Description}}</td>
        </tr>
    {{else}}
        <!-- Render as regular line item -->
        <tr>
            <td>{{.Title}}</td>
            <td>{{formatCurrency .OriginalAmount .OriginalCurrency}}</td>
        </tr>
    {{end}}
{{end}}
```

## Error Conditions

- Function should never panic regardless of input
- Should handle null/empty strings gracefully
- Should not modify input string

## Related Specifications

- spec-002-template-functions.md: Template function definitions
- spec-003-html-template-restructure.md: Template usage patterns

## Dependencies

- `strings` package (if using case-insensitive/trim approach)

## Test Implementation Notes

**Test File:** `pkg/controller/invoice/contractor_invoice_helpers_test.go`

**Test Structure:**
```go
func TestIsSectionHeader(t *testing.T) {
    tests := []struct {
        name     string
        itemType string
        expected bool
    }{
        // Test cases as specified above
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := isSectionHeader(tt.itemType)
            if result != tt.expected {
                t.Errorf("isSectionHeader(%q) = %v; want %v",
                    tt.itemType, result, tt.expected)
            }
        })
    }
}

func TestIsServiceFee(t *testing.T) {
    tests := []struct {
        name     string
        itemType string
        expected bool
    }{
        // Test cases as specified above
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := isServiceFee(tt.itemType)
            if result != tt.expected {
                t.Errorf("isServiceFee(%q) = %v; want %v",
                    tt.itemType, result, tt.expected)
            }
        })
    }
}
```

## Success Criteria

- All test cases pass
- No panics for any input string
- Consistent behavior across multiple calls
- Clear documentation of case sensitivity behavior
- Template rendering works correctly based on function results
