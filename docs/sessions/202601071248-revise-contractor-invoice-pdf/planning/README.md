# Planning Documentation - Contractor Invoice PDF Revision

This directory contains comprehensive planning documents for revising the contractor invoice PDF format to support multi-currency display (VND and USD).

## Document Structure

### Architecture Decision Records (ADRs)

ADRs document the key architectural decisions and their rationale.

1. **[ADR-001: Data Structure for Multi-Currency Support](ADRs/ADR-001-data-structure-multi-currency.md)**
   - Decision: Dual-amount storage approach
   - Adds `OriginalAmount` and `OriginalCurrency` to line items
   - Adds subtotal fields to invoice data
   - Pre-calculates all values in controller layer

2. **[ADR-002: Currency Formatting Approach](ADRs/ADR-002-currency-formatting-approach.md)**
   - Decision: Custom template functions using `text/template.FuncMap`
   - VND format: `45.000.000 ₫`
   - USD format: `$1,234.56`
   - Uses `golang.org/x/text/message` for locale-aware formatting

### Specifications

Specifications provide detailed implementation instructions.

1. **[spec-001: Data Structure Changes](specifications/spec-001-data-structure-changes.md)**
   - Exact struct modifications
   - Field-by-field definitions
   - Data population logic
   - Validation rules
   - Examples and edge cases

2. **[spec-002: Template Functions](specifications/spec-002-template-functions.md)**
   - `formatCurrency(amount, currency)` - Multi-currency formatter
   - `formatExchangeRate(rate)` - Exchange rate display
   - `isServiceFee(type)` - Section type detection
   - `isSectionHeader(type)` - Header detection
   - Helper functions and unit test requirements

3. **[spec-003: HTML Template Restructure](specifications/spec-003-html-template-restructure.md)**
   - Section-based rendering approach
   - Development Work: aggregated display with "Development work from [date] to [date]" title
   - Refund/Bonus: header + individual items
   - Multi-currency totals section
   - Exchange rate footnote
   - CSS style updates

4. **[spec-004: Calculation Logic](specifications/spec-004-calculation-logic.md)**
   - Step-by-step calculation algorithm
   - Subtotal calculation by currency
   - VND to USD conversion logic
   - Combined subtotal calculation
   - FX support fee addition
   - Examples and edge cases

### Status Document

**[STATUS.md](STATUS.md)** - Planning phase status and handoff information
- Overview of completed documents
- Key decisions summary
- Implementation order
- Acceptance criteria
- Test case requirements
- Handoff instructions for test case designer and implementation team

## Reading Order

### For Understanding Context
1. Start with `STATUS.md` for overview
2. Read ADRs to understand key decisions
3. Review specifications for implementation details

### For Implementation
1. `spec-001-data-structure-changes.md` - Implement structs and data flow
2. `spec-004-calculation-logic.md` - Implement calculation algorithms
3. `spec-002-template-functions.md` - Implement formatting functions
4. `spec-003-html-template-restructure.md` - Update template

### For Test Case Design
1. Review all specifications for test scenarios
2. See "Testing Requirements" sections in each spec
3. See "Acceptance Criteria" in STATUS.md
4. See "Edge Cases" sections for boundary conditions

## Quick Reference

### Files to Modify

**Backend (Go):**
- `pkg/controller/invoice/contractor_invoice.go` - Main changes
- `pkg/controller/invoice/contractor_invoice_formatters.go` - NEW file
- `pkg/controller/invoice/contractor_invoice_formatters_test.go` - NEW file
- `pkg/controller/invoice/contractor_invoice_calculations_test.go` - NEW file

**Template:**
- `pkg/templates/contractor-invoice-template.html` - Major restructure

### Key Technical Decisions

- **Data Flow:** Notion → Controller (preserve + convert) → Template (format + display)
- **Calculation Layer:** Controller (testable Go code)
- **Formatting Layer:** Template functions (presentation only)
- **Section Grouping:** Pre-processed in controller (cleaner template)
- **Error Handling:** Fail fast on API errors (no fallback rates)

### Currency Formatting Rules

**VND:**
- Symbol: `₫` (after amount, with space)
- Separator: `.` (period for thousands)
- Decimals: 0 (no minor units)
- Example: `45.000.000 ₫`

**USD:**
- Symbol: `$` (before amount, no space)
- Separator: `,` (comma for thousands)
- Decimals: 2 (required)
- Example: `$1,234.56`

**Exchange Rate:**
- Format: `1 USD = X VND`
- Example: `1 USD = 26,269 VND`

### Calculation Flow

```
Line Items (VND + USD)
    ↓
Group by Currency
    ↓
Subtotal VND → Convert → Subtotal USD from VND
Subtotal USD Items
    ↓
Combine → Subtotal USD
    ↓
Add FX Support ($8)
    ↓
Total USD
```

## Related Documentation

### Session Documents
- `../requirements/requirements.md` - User requirements
- `../research/STATUS.md` - Research findings
- `../research/go-html-template-best-practices.md`
- `../research/currency-formatting-best-practices.md`
- `../research/multi-currency-invoice-patterns.md`

### Codebase Documentation
- `/CLAUDE.md` - Project conventions
- `/docs/adr/` - Other architecture decisions
- `/pkg/controller/invoice/` - Current implementation

## Contact

For questions about planning decisions, refer to the ADRs for rationale and context.
