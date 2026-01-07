# Test Plan: Edge Cases and Boundary Conditions

**Version:** 1.0
**Date:** 2026-01-07
**Component:** Contractor Invoice PDF Generation
**Scope:** Edge cases, boundary conditions, error scenarios

## Purpose

This test plan documents edge cases and boundary conditions that must be tested to ensure robust invoice generation across all scenarios.

## Edge Case Categories

### 1. Numeric Boundary Conditions

#### EC-001: Very Large Amounts
**Scenario:** Invoice with amounts approaching float64 maximum
**Test Data:**
```go
VND: 999,999,999,999 (nearly 1 trillion)
USD: 999,999,999.99 (nearly 1 billion)
```
**Expected Behavior:**
- Formatting handles billions correctly
- No overflow errors
- Calculations remain accurate
- PDF renders without layout breaks

**Success Criteria:**
- Amounts display with correct separators
- All calculations accurate to 2 decimal places
- No scientific notation in output

#### EC-002: Very Small Amounts
**Scenario:** Invoice with sub-cent USD amounts
**Test Data:**
```go
USD: 0.001, 0.004, 0.009
VND: 0.1, 0.5, 0.9
```
**Expected Behavior:**
- Rounding to appropriate decimal places
- USD rounds to $0.00 for amounts < $0.005
- VND rounds to 0 ₫ for amounts < 0.5

**Success Criteria:**
- No negative zero values
- Consistent rounding behavior
- Subtotals and totals remain accurate

#### EC-003: Zero Amounts
**Scenario:** Invoice with all zero amounts
**Test Data:**
```go
lineItems := []ContractorInvoiceLineItem{
    {OriginalAmount: 0, OriginalCurrency: "VND"},
    {OriginalAmount: 0, OriginalCurrency: "USD"},
}
```
**Expected Behavior:**
- Invoice generates successfully
- All subtotals display as 0
- FX Support still added ($8.00)
- Total = $8.00

**Success Criteria:**
- No division by zero errors
- PDF renders correctly
- Exchange rate handling is graceful

#### EC-004: Single Penny/Dong
**Scenario:** Minimal non-zero amounts
**Test Data:**
```go
USD: 0.01 (one cent)
VND: 1 (one dong)
```
**Expected Behavior:**
- Amounts display correctly
- Calculations remain accurate
- No underflow errors

### 2. Currency Edge Cases

#### EC-005: All VND Invoice
**Scenario:** 100% VND-denominated items, no USD
**Test Data:**
```go
Multiple VND items totaling 100,000,000 VND
No USD items
```
**Expected Behavior:**
- SubtotalVND populated
- SubtotalUSDItems = 0
- Exchange rate fetched from Wise
- Conversion accurate
- Only VND amounts shown in line items

**Success Criteria:**
- Wise API called exactly once
- Exchange rate displayed in footer
- All VND amounts formatted correctly

#### EC-006: All USD Invoice
**Scenario:** 100% USD-denominated items, no VND
**Test Data:**
```go
Multiple USD items totaling $5,000
No VND items
```
**Expected Behavior:**
- SubtotalUSDItems populated
- SubtotalVND = 0
- SubtotalUSDFromVND = 0
- No Wise API call needed
- ExchangeRate = 1.0

**Success Criteria:**
- No unnecessary API calls
- Exchange rate note handled gracefully
- All USD amounts formatted correctly

#### EC-007: Single Item Invoices
**Scenario:** Invoice with only one line item (VND or USD)
**Test Data:**
```go
Case A: Single VND item (10,000,000 VND)
Case B: Single USD item ($500)
```
**Expected Behavior:**
- Invoice generates successfully
- Subtotals equal line item amount
- FX support added
- Total calculated correctly

**Success Criteria:**
- No array iteration errors
- Correct calculations
- PDF renders properly

#### EC-008: Mixed Currency with Imbalance
**Scenario:** 99% VND with tiny USD amount (or vice versa)
**Test Data:**
```go
VND: 99,000,000 (large)
USD: 1 (tiny)
```
**Expected Behavior:**
- Both amounts displayed correctly
- Subtotals accurate
- No precision loss in small amount

**Success Criteria:**
- Small amounts not lost in rounding
- Both currencies visible in invoice
- Totals accurate

### 3. Exchange Rate Edge Cases

#### EC-009: Exchange Rate API Failure
**Scenario:** Wise API returns error or timeout
**Test Data:** Mock Wise API to return error
**Expected Behavior:**
- Invoice generation fails gracefully
- Clear error message returned
- No invoice PDF generated
- Transaction not committed

**Success Criteria:**
- Error message: "failed to convert VND subtotal to USD"
- No partial invoice created
- Logs contain detailed error info

#### EC-010: Invalid Exchange Rate (Zero)
**Scenario:** Wise API returns rate = 0
**Test Data:** Mock Wise API to return rate 0
**Expected Behavior:**
- Validation catches zero rate
- Error returned: "invalid exchange rate: 0 (must be > 0)"
- Invoice generation fails

**Success Criteria:**
- No division by zero
- Clear validation error
- No invoice generated

#### EC-011: Invalid Exchange Rate (Negative)
**Scenario:** Wise API returns negative rate
**Test Data:** Mock Wise API to return rate -26269
**Expected Behavior:**
- Validation catches negative rate
- Error returned
- Invoice generation fails

**Success Criteria:**
- Validation error with rate value
- No invoice generated

#### EC-012: Extreme Exchange Rate
**Scenario:** Exchange rate outside normal range
**Test Data:**
```go
Very low: 1000 VND/USD (unrealistic)
Very high: 100000 VND/USD (unrealistic)
```
**Expected Behavior:**
- Calculations still work
- No overflow/underflow
- Amounts display correctly

**Success Criteria:**
- No calculation errors
- Formatting handles extreme values
- Warning logged for unusual rates (optional)

### 4. Data Validation Edge Cases

#### EC-013: Invalid Currency Code
**Scenario:** Payout with currency other than VND/USD
**Test Data:**
```go
payout := PayoutEntry{
    Amount: 1000,
    Currency: "EUR",
}
```
**Expected Behavior:**
- Validation error before invoice creation
- Error: "invalid currency: EUR (must be VND or USD)"
- No invoice generated

**Success Criteria:**
- Early validation prevents processing
- Clear error message
- No Wise API calls made

#### EC-014: Empty Currency Code
**Scenario:** Payout with empty currency string
**Test Data:**
```go
payout := PayoutEntry{
    Amount: 1000,
    Currency: "",
}
```
**Expected Behavior:**
- Validation error
- Invoice generation fails

**Success Criteria:**
- Error caught before calculations
- Clear validation message

#### EC-015: Negative Amounts
**Scenario:** Line item with negative amount
**Test Data:**
```go
lineItem := ContractorInvoiceLineItem{
    OriginalAmount: -500,
    OriginalCurrency: "USD",
}
```
**Expected Behavior:**
- Validation error
- Error: "negative amount for line item"
- Invoice not generated

**Success Criteria:**
- Validation prevents negative amounts
- Clear error message
- Note: Refunds should be positive amounts with Type="Refund"

#### EC-016: Null/Missing Data
**Scenario:** PayoutEntry with missing required fields
**Test Data:**
```go
payout := PayoutEntry{
    // Amount missing
    Currency: "USD",
}
```
**Expected Behavior:**
- Validation error or default handling
- Invoice generation fails safely

**Success Criteria:**
- No panics
- Clear error about missing field

### 5. Collection Edge Cases

#### EC-017: Empty Line Items Array
**Scenario:** Invoice with no line items
**Test Data:**
```go
lineItems := []ContractorInvoiceLineItem{}
```
**Expected Behavior:**
- Invoice generates (valid scenario for adjustment invoice)
- All subtotals = 0
- Total = FX Support ($8.00)
- ExchangeRate = 1.0

**Success Criteria:**
- No array iteration errors
- PDF renders with empty item section
- Calculations handle empty array

#### EC-018: Very Large Line Item Count
**Scenario:** Invoice with 100+ line items
**Test Data:** Generate 100+ line items
**Expected Behavior:**
- All items processed correctly
- Subtotals accurate
- PDF renders (may span multiple pages)
- Performance acceptable

**Success Criteria:**
- No performance degradation
- All items visible in PDF
- Page breaks handled correctly
- Calculations remain accurate

#### EC-019: Duplicate Line Items
**Scenario:** Multiple identical line items
**Test Data:**
```go
lineItems := []ContractorInvoiceLineItem{
    {OriginalAmount: 1000, OriginalCurrency: "USD", Title: "Service Fee"},
    {OriginalAmount: 1000, OriginalCurrency: "USD", Title: "Service Fee"},
}
```
**Expected Behavior:**
- Both items processed separately
- Subtotal = 2000
- No deduplication (unless business logic requires)

**Success Criteria:**
- All items appear in invoice
- Subtotals include all items

### 6. Formatting Edge Cases

#### EC-020: Numbers at Separator Boundaries
**Scenario:** Amounts that are exactly at separator boundaries
**Test Data:**
```go
VND: 1000, 1000000, 1000000000
USD: 1000.00, 1000000.00
```
**Expected Behavior:**
- Separators placed correctly
- No extra separators
- No missing separators

**Success Criteria:**
- VND: "1.000 ₫", "1.000.000 ₫", "1.000.000.000 ₫"
- USD: "$1,000.00", "$1,000,000.00"

#### EC-021: Rounding at Exact Half
**Scenario:** Amounts that are exactly at rounding boundary
**Test Data:**
```go
VND: 1234.5 → round to 1235 or 1234?
USD: 1.565 → round to 1.57 or 1.56?
```
**Expected Behavior:**
- Consistent rounding mode (banker's rounding or round half up)
- Documented behavior

**Success Criteria:**
- Consistent across all amounts
- Test documents expected behavior
- Implementation matches Go's math.Round

#### EC-022: Unicode Symbol Handling
**Scenario:** Currency symbols in various encodings
**Test Data:** VND dong symbol (₫) in UTF-8
**Expected Behavior:**
- Symbol displays correctly in PDF
- No encoding errors
- Symbol positioned correctly

**Success Criteria:**
- PDF displays ₫ symbol
- No replacement characters (�)
- Correct spacing

### 7. Calculation Edge Cases

#### EC-023: Precision Loss in Conversion
**Scenario:** VND amount that doesn't convert cleanly to USD
**Test Data:**
```go
VND: 45,678,123
Rate: 26,269.5
```
**Expected Behavior:**
- Conversion performed
- Result rounded to 2 decimals
- No cumulative rounding errors

**Success Criteria:**
- Result: 1738.37 USD (or accurate to 2 decimals)
- Rounding documented
- Total accurate

#### EC-024: Subtotal Summation Rounding
**Scenario:** Multiple items with rounding each
**Test Data:**
```go
USD items: 10.004, 10.005, 10.006
Each rounds to: 10.00, 10.01, 10.01
Sum: 30.02
```
**Expected Behavior:**
- Each item rounded independently
- Subtotal = sum of rounded values
- Or: Sum first, then round

**Success Criteria:**
- Documented rounding order
- Consistent behavior
- Test validates implementation choice

#### EC-025: FX Support Edge Case
**Scenario:** Invoice where FX Support > Subtotal
**Test Data:**
```go
Subtotal: $5.00
FX Support: $8.00
Total: $13.00
```
**Expected Behavior:**
- Invoice generates successfully
- FX Support not conditionally removed
- Both amounts visible

**Success Criteria:**
- All amounts displayed
- Total accurate
- No special handling needed

### 8. Template Rendering Edge Cases

#### EC-026: Missing Data in Template
**Scenario:** Template references field that is nil/empty
**Test Data:** InvoiceData with some fields unset
**Expected Behavior:**
- Template renders with default values
- Or: Error caught and logged
- No template execution panic

**Success Criteria:**
- No runtime errors
- PDF generates or fails gracefully

#### EC-027: Very Long Description Text
**Scenario:** Line item with 1000+ character description
**Test Data:** Description with very long text
**Expected Behavior:**
- Text rendered without breaking layout
- Text wrap or truncate as needed
- PDF remains readable

**Success Criteria:**
- No layout overflow
- Text visible in PDF

#### EC-028: Special Characters in Descriptions
**Scenario:** Descriptions with HTML, emoji, special chars
**Test Data:**
```
Description: "Fixed <bug> & added emoji 🎉"
```
**Expected Behavior:**
- Special chars escaped correctly
- No HTML injection
- PDF displays safely

**Success Criteria:**
- Text displays correctly
- No rendering errors
- No security issues

## Test Execution Strategy

### Priority Levels

**P0 (Critical):**
- EC-003: Zero amounts
- EC-005: All VND invoice
- EC-006: All USD invoice
- EC-009: Exchange rate API failure
- EC-013: Invalid currency code
- EC-017: Empty line items

**P1 (High):**
- EC-001: Very large amounts
- EC-002: Very small amounts
- EC-007: Single item invoices
- EC-010: Invalid exchange rate
- EC-015: Negative amounts

**P2 (Medium):**
- EC-004: Single penny/dong
- EC-008: Mixed currency imbalance
- EC-011: Negative exchange rate
- EC-020: Separator boundaries
- EC-023: Precision loss

**P3 (Low):**
- EC-012: Extreme exchange rates
- EC-018: Large line item count
- EC-021: Rounding at exact half
- EC-027: Very long descriptions

## Test Data Generation

### Mock Data Builders

Create helper functions to generate edge case test data:

```go
func makeEmptyInvoice() *ContractorInvoiceData
func makeAllVNDInvoice() *ContractorInvoiceData
func makeAllUSDInvoice() *ContractorInvoiceData
func makeLargeAmountInvoice() *ContractorInvoiceData
func makeZeroAmountInvoice() *ContractorInvoiceData
```

## Success Criteria

- All P0 test cases pass
- All P1 test cases pass
- P2 and P3 test cases documented and tracked
- No panics for any edge case
- All errors have clear messages
- Edge case behavior is documented

## Related Test Cases

- TC-001 through TC-007: Unit test specifications
- Integration test specifications (when created)
