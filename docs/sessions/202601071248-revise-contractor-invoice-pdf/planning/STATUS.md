# Planning Status

**Session:** 202601071248-revise-contractor-invoice-pdf

**Status:** COMPLETED

**Last Updated:** 2026-01-07

## Overview

Planning phase for revising the contractor invoice PDF format to support multi-currency display (VND and USD) with section headers and proper totals calculation.

## Completed Planning Documents

### Architecture Decision Records (ADRs)

| Document | Description | Status |
|----------|-------------|--------|
| `ADR-001-data-structure-multi-currency.md` | Decision on dual-amount storage approach for preserving original currency data | Completed |
| `ADR-002-currency-formatting-approach.md` | Decision on custom template functions for VND/USD formatting | Completed |

### Specifications

| Document | Description | Status |
|----------|-------------|--------|
| `spec-001-data-structure-changes.md` | Detailed specification for struct modifications and field additions | Completed |
| `spec-002-template-functions.md` | Specification for new template helper functions (formatCurrency, formatExchangeRate, etc.) | Completed |
| `spec-003-html-template-restructure.md` | Specification for template layout changes with section grouping | Completed |
| `spec-004-calculation-logic.md` | Specification for subtotal and total calculation algorithms | Completed |

## Key Decisions Summary

### 1. Data Structure Approach (ADR-001)

**Decision:** Dual-amount storage with original currency preservation

**Key Points:**
- Add `OriginalAmount` and `OriginalCurrency` fields to `ContractorInvoiceLineItem`
- Add subtotal fields to `ContractorInvoiceData`: `SubtotalVND`, `SubtotalUSDFromVND`, `SubtotalUSDItems`, `SubtotalUSD`, `FXSupport`
- Pre-calculate all subtotals in controller (business logic layer)
- Template only handles formatting and display (presentation layer)

**Rationale:**
- Clear separation of concerns
- Testable calculations in Go code
- Preserves audit trail of currency conversions

### 2. Currency Formatting (ADR-002)

**Decision:** Custom template functions using Go's `text/template.FuncMap`

**Key Points:**
- VND format: `45.000.000 ₫` (period separator, symbol after, no decimals)
- USD format: `$1,234.56` (comma separator, symbol before, 2 decimals)
- Exchange rate format: `1 USD = 26,269 VND`
- Use `golang.org/x/text/message` for locale-aware number formatting

**Rationale:**
- Precise control over formatting
- No additional dependencies
- Simple template syntax
- Easy to test and maintain

### 3. Section Grouping

**Decision:** Pre-process line items into sections in controller

**Sections:**
1. **Development Work:** Aggregated row with total, title shows "Development work from [start_date] to [end_date]", descriptions below (no individual amounts)
2. **Refund:** Section header + individual item rows with amounts
3. **Bonus (Commission):** Section header + individual item rows with amounts

**Rationale:**
- Go html/template has limited programming capabilities
- Easier to test section grouping logic in Go
- Cleaner template code

### 4. Calculation Flow

**Flow:**
```
Line Items → Group by Currency → Calculate Subtotals →
Convert VND to USD → Combine USD Subtotals → Add FX Support → Total
```

**Key Calculations:**
- Subtotal VND = Sum of VND items (rounded to 0 decimals)
- Subtotal USD from VND = Subtotal VND converted via Wise API (rounded to 2 decimals)
- Subtotal USD Items = Sum of USD items (rounded to 2 decimals)
- Subtotal USD = Subtotal USD from VND + Subtotal USD Items
- FX Support = $8.00 (hardcoded)
- Total USD = Subtotal USD + FX Support

## Files to be Modified

### Backend (Go)

| File | Changes Required | Complexity |
|------|------------------|------------|
| `pkg/controller/invoice/contractor_invoice.go` | Add fields to structs, implement calculation logic, add helper functions | High |
| `pkg/controller/invoice/contractor_invoice_formatters.go` | NEW: Implement currency formatting helper functions | Medium |
| `pkg/controller/invoice/contractor_invoice_test.go` | Update tests for new fields and calculations | Medium |
| `pkg/controller/invoice/contractor_invoice_formatters_test.go` | NEW: Unit tests for formatting functions | Medium |
| `pkg/controller/invoice/contractor_invoice_calculations_test.go` | NEW: Unit tests for calculation logic | Medium |

### Frontend (HTML Template)

| File | Changes Required | Complexity |
|------|------------------|------------|
| `pkg/templates/contractor-invoice-template.html` | Restructure table body with sections, update totals footer, add exchange rate footnote | High |

## Implementation Order

### Phase 1: Data Structure Changes
1. Add new fields to `ContractorInvoiceLineItem` struct
2. Add new fields to `ContractorInvoiceData` struct
3. Update line item population to preserve original currency

### Phase 2: Calculation Logic
4. Implement subtotal calculation by currency
5. Implement VND to USD conversion with error handling
6. Implement combined subtotal and total calculation
7. Add validation for currency codes and amounts

### Phase 3: Formatting Functions
8. Implement `formatVND()` helper function
9. Implement `formatUSD()` helper function
10. Implement `formatVNDNumber()` and `formatUSDNumber()` helpers
11. Add `formatCurrency` template function
12. Add `formatExchangeRate` template function
13. Add `isServiceFee` and `isSectionHeader` helper functions

### Phase 4: Section Grouping
14. Add `ContractorInvoiceSection` struct
15. Implement `groupLineItemsIntoSections()` function
16. Update template data preparation

### Phase 5: Template Updates
17. Update template to use sections-based rendering
18. Update totals footer with multi-currency subtotals
19. Add exchange rate footnote
20. Update CSS styles for section headers and spacing

### Phase 6: Testing
21. Unit tests for formatting functions
22. Unit tests for calculation logic
23. Integration tests with template rendering
24. Visual testing of PDF output

## Acceptance Criteria

- [ ] PDF displays section headers (Development work from [date] to [date], Refund, Bonus)
- [ ] Development work items are aggregated with total amount, descriptions shown below
- [ ] Refund and Bonus items show individual rows with amounts
- [ ] All items display in their original currency (VND or USD)
- [ ] VND amounts formatted as: `45.000.000 ₫` (period separator, symbol after)
- [ ] USD amounts formatted as: `$1,234.56` (comma separator, symbol before)
- [ ] Totals section shows:
  - VND subtotal (if any VND items)
  - USD subtotal
  - FX support ($8)
  - Total
- [ ] Exchange rate footnote appears at bottom: `*FX Rate 1 USD = 26,269 VND`
- [ ] All calculations are accurate and properly rounded
- [ ] Existing invoice generation functionality remains intact

## Test Cases Required

### Unit Tests
- Currency formatting (VND, USD) with various amounts
- Exchange rate formatting
- Subtotal calculations (all VND, all USD, mixed)
- Rounding logic (VND to 0 decimals, USD to 2 decimals)
- Section grouping logic

### Integration Tests
- Template rendering with sections
- PDF generation with all section types
- Multi-currency invoice generation end-to-end

### Edge Cases
- Zero amounts
- Single item invoices
- Large amounts (billions in VND)
- Small amounts (cents in USD)
- Exchange rate API failures
- Invalid currency codes
- Negative amounts (should error)

## Dependencies

### External Services
- Wise API: Currency conversion and exchange rates (already implemented)

### Go Packages
- `golang.org/x/text/message`: Locale-aware number formatting (recommended)
- `golang.org/x/text/language`: Language/locale definitions (recommended)
- Existing: `github.com/Rhymond/go-money`, `math`, `fmt`, `strings`

### No Breaking Changes
- All new fields are additions (no removals)
- Existing template will continue to work (uses different fields)
- Can be deployed together (no migration needed)

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Exchange rate API failure during invoice generation | High | Return clear error, don't use fallback rate (ensures accurate invoices) |
| Incorrect rounding causes payment discrepancies | High | Comprehensive unit tests for rounding logic, use banker's rounding |
| Template rendering errors with complex sections | Medium | Pre-process sections in controller, keep template logic simple |
| Performance impact of number formatting | Low | Use efficient formatting library (golang.org/x/text), cache formatters |
| Missing edge cases in calculations | Medium | Extensive test coverage including edge cases |

## Handoff to Next Phase

### For Test Case Designer

**Location:** `docs/sessions/202601071248-revise-contractor-invoice-pdf/test-cases/`

**Required Test Plans:**
1. Unit test specifications for formatting functions
2. Unit test specifications for calculation logic
3. Integration test specifications for template rendering
4. Visual test specifications for PDF output
5. Edge case test specifications

**Key Acceptance Criteria:** See "Acceptance Criteria" section above

**Test Data Requirements:**
- Sample line items with VND amounts
- Sample line items with USD amounts
- Sample line items with mixed currencies
- Various exchange rates for testing
- Edge case data (zero, negative, very large amounts)

### For Implementation Team

**Start With:**
1. Review all specifications and ADRs
2. Implement data structure changes (spec-001)
3. Implement calculation logic (spec-004)
4. Implement formatting functions (spec-002)
5. Update template (spec-003)
6. Write tests per test case designer's specifications

**Critical Notes:**
- DO NOT use fallback exchange rates (return error if API fails)
- Ensure VND rounded to 0 decimals, USD to 2 decimals
- Pre-calculate all values in controller (not in template)
- Validate currency codes ("VND" or "USD" only)

## Related Documents

### Requirements
- `docs/sessions/202601071248-revise-contractor-invoice-pdf/requirements/requirements.md`

### Research
- `docs/sessions/202601071248-revise-contractor-invoice-pdf/research/go-html-template-best-practices.md`
- `docs/sessions/202601071248-revise-contractor-invoice-pdf/research/currency-formatting-best-practices.md`
- `docs/sessions/202601071248-revise-contractor-invoice-pdf/research/multi-currency-invoice-patterns.md`
- `docs/sessions/202601071248-revise-contractor-invoice-pdf/research/STATUS.md`

### Planning
- `ADRs/ADR-001-data-structure-multi-currency.md`
- `ADRs/ADR-002-currency-formatting-approach.md`
- `specifications/spec-001-data-structure-changes.md`
- `specifications/spec-002-template-functions.md`
- `specifications/spec-003-html-template-restructure.md`
- `specifications/spec-004-calculation-logic.md`

## Next Steps

1. **Test Case Designer** (@agent-test-case-designer): Create comprehensive test plans based on specifications
2. **Implementation Team**: Begin implementation following the phased approach outlined above
3. **QA**: Prepare visual testing environment for PDF output validation

## Notes

- All planning documents are complete and ready for implementation
- No ambiguities or open questions remaining
- All edge cases identified and documented
- Clear separation of concerns established
- Backward compatibility maintained
- FX support fee is hardcoded at $8 for now (future enhancement: dynamic calculation)
