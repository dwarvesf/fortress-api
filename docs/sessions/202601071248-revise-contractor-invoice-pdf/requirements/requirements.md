# Requirements: Revise Contractor Invoice PDF Format

## Overview
Update the contractor invoice PDF to match the expected format with section headers, original currency display, and proper totals section with FX support.

## Clarified Requirements (from user discussion)

### Currency Display Rules
1. **All items display in their original currency** (VND shows đ, USD shows $)
2. **Exception**: Hourly rate Service Fee items → always display in USD
3. **Subtotal VND**: Sum of all VND-denominated items
4. **Subtotal USD**: VND subtotal converted to USD + USD-denominated items
5. **FX Support**: $8 (hardcoded, TODO: implement calculation later)
6. **Total**: Subtotal USD + FX Support

### Line Item Display Logic
1. **Development Work Section**: ONE aggregated row with total amount, title shows "Development work from [start_date] to [end_date]", description items shown below (no individual amounts)
2. **Refund**: Section header row (bold, no amounts) + each refund as separate row with Qty/Cost/Total
3. **Bonus**: Section header row (bold, no amounts) + each bonus as separate row with Qty/Cost/Total

### Expected PDF Format
```
| DESCRIPTION                              | QUANTITY | UNIT COST      | TOTAL         |
|------------------------------------------|----------|----------------|---------------|
| **Development work from Dec 1 to Dec 31**|    1     | đ 45,000,000   | đ 45,000,000  |
|   Ascenda: Project mgmt...               |          |                |               |
|   Dwarves: Client onb...                 |          |                |               |
|                           |          |                |               |
| **Refund**                |          |                |               |  <- section header
|   Advance return xyz...   |    1     | đ 500,000      | đ 500,000     |  <- individual item
|   Advance return xyz...   |    1     | đ 650,000      | đ 650,000     |  <- individual item
|                           |          |                |               |
| **Bonus**                 |          |                |               |  <- section header
|   Ascenda Bonus           |    1     | $100.00        | $100.00       |  <- individual item
|   Inloop Bonus            |    1     | $115.00        | $115.00       |  <- individual item
|---------------------------|----------|----------------|---------------|
|                           |          | Subtotal       | đ 56,394,910  |
|                           |          |                | $2,146.82     |
|                           |          | FX support     | $8            |
|                           |          | Total          | $2,154.82     |

*FX Rate USD 1 ~ 26269 VND
```

## Technical Constraints
- Bonus items can be in VND or USD - display in original currency
- FX support fee is hardcoded to $8 for now (calculation to be implemented later)
- Exchange rate should be captured from Wise API and displayed as "1 USD = X VND"
- Must maintain existing GORM/database compatibility

## Files Involved
- `pkg/controller/invoice/contractor_invoice.go` - Data structures and generation logic
- `pkg/templates/contractor-invoice-template.html` - HTML template for PDF

## Acceptance Criteria
1. PDF displays section headers (Development work from [date] to [date], Refund, Bonus)
2. Development work items are aggregated with descriptions below (one total row)
3. Refund and Bonus items show individual rows with amounts
4. All items display in their original currency (VND or USD)
5. Totals section shows VND subtotal, USD subtotal, FX support, and final total
6. FX rate footnote appears at bottom
