# Requirements: Hourly Rate-Based Service Fee Display

**Session**: 202601072141-hourly-rate-service-fee
**Created**: 2026-01-07
**Status**: Confirmed with user

## Background

Currently, contractor invoices display all Service Fee line items with:
- Quantity = 1
- Unit Cost = Total Amount
- Description from Work Details or proof of works

This works for fixed-amount Service Fees, but for hourly-rate contractors, we need to display:
- Quantity = Total hours worked
- Unit Cost = Hourly rate
- Amount = Quantity × Unit Cost

## Requirements

### Functional Requirements

#### FR-1: Hourly Rate Detection
When processing Service Fee payouts for invoice generation:
- **IF** the payout has a `00 Service Rate` relation to Contractor Rates table
- **AND** the Contractor Rate's `Billing Type` = "Hourly Rate"
- **THEN** display as hourly-rate Service Fee

#### FR-2: Data Retrieval
For hourly-rate Service Fees, retrieve:
1. **From Contractor Payouts**:
   - `00 Service Rate` relation → ServiceRateID
   - `00 Task Order` relation → TaskOrderID
   - Amount, Currency

2. **From Contractor Rates** (via ServiceRateID):
   - Billing Type
   - Hourly Rate
   - Currency

3. **From Task Order Log** (via TaskOrderID):
   - Final Hours Worked (formula field)

#### FR-3: Line Item Display
For hourly-rate Service Fees:
- **Title**: "Service Fee (Development work from YYYY-MM-01 to YYYY-MM-DD)"
  - Date range = invoice month (1st to last day of month)
- **Description**: Aggregated proof of works from all hourly Service Fee items
- **Quantity**: Sum of all `Final Hours Worked` from Task Order Log entries
- **Unit Cost**: `Hourly Rate` from Contractor Rates (in original currency)
- **Amount**: Sum of all Service Fee payout amounts (already calculated in Notion)

#### FR-4: Aggregation
- **All** hourly-rate Service Fee items for the same invoice **MUST** be aggregated into a **single line item**
- Sum quantities (hours), use single hourly rate, total amount

#### FR-5: Currency Support
- Apply hourly rate logic to **both USD and VND** currencies
- Display in **original currency** (no forced conversion)
- Currency from payout determines display currency

#### FR-6: Fallback Behavior
Use default display (Qty=1, Unit Cost=Amount) when:
- `00 Service Rate` relation is missing/empty
- Contractor Rate fetch fails
- `Billing Type` is NOT "Hourly Rate" (e.g., "Monthly Fixed")
- Task Order Log fetch fails
- Any other data retrieval error

### Non-Functional Requirements

#### NFR-1: Backward Compatibility
- Existing invoices must continue to work unchanged
- Non-hourly Service Fees display as before
- Commission, Refund, Other payout types unchanged
- No breaking changes to existing functionality

#### NFR-2: Error Handling
- Graceful degradation on API failures
- Clear DEBUG logging at every decision point
- No invoice generation failures due to missing hourly rate data

#### NFR-3: Performance
- Acceptable: 2-6 additional Notion API calls per invoice
- Sequential fetching is acceptable (low volume: 1-3 Service Fees per invoice)
- No N+1 query issues

#### NFR-4: Code Quality
- Follow existing code patterns in contractor_invoice.go
- Reuse existing helper methods where possible
- Comprehensive DEBUG logging
- Unit test coverage for new logic

## Acceptance Criteria

### AC-1: Hourly Rate Detection
- [ ] Service Fee with valid ServiceRateID and BillingType="Hourly Rate" → detected as hourly
- [ ] Service Fee without ServiceRateID → fallback to default display
- [ ] Service Fee with BillingType="Monthly Fixed" → fallback to default display

### AC-2: Data Retrieval
- [ ] ServiceRateID extracted from Contractor Payouts `00 Service Rate` relation
- [ ] Contractor Rate data fetched by ServiceRateID (BillingType, HourlyRate, Currency)
- [ ] Task Order Log `Final Hours Worked` fetched by TaskOrderID

### AC-3: Single Invoice - Single Hourly Service Fee
- [ ] Title: "Service Fee (Development work from 2026-01-01 to 2026-01-31)" (example for Jan 2026)
- [ ] Quantity: Final Hours Worked from Task Order Log
- [ ] Unit Cost: Hourly Rate from Contractor Rates
- [ ] Amount: Matches payout Amount field
- [ ] Currency: Displays in original currency (USD or VND)

### AC-4: Single Invoice - Multiple Hourly Service Fees
- [ ] All hourly Service Fees aggregated into single line item
- [ ] Quantity: Sum of all Final Hours Worked values
- [ ] Unit Cost: Hourly Rate (assumes all use same rate)
- [ ] Amount: Sum of all payout amounts
- [ ] Description: Concatenated proof of works from all items

### AC-5: Mixed Invoice
- [ ] 1 hourly Service Fee + 1 non-hourly Service Fee → 2 separate line items
- [ ] 1 hourly Service Fee + 1 Commission → 2 separate line items
- [ ] Hourly Service Fee aggregated, other items displayed normally

### AC-6: Error Handling
- [ ] Missing ServiceRateID → fallback with DEBUG log
- [ ] Failed rate fetch → fallback with DEBUG log
- [ ] Failed hours fetch → use 0 hours with DEBUG log (still aggregate amount)
- [ ] Invalid BillingType → fallback with DEBUG log
- [ ] Invoice generation succeeds in all error scenarios

### AC-7: Multi-Currency
- [ ] USD hourly Service Fee → displayed in USD
- [ ] VND hourly Service Fee → displayed in VND
- [ ] Currency formatting follows existing formatCurrency rules

## Out of Scope

- Dynamic FX support fee calculation (remains hardcoded at $8)
- Section grouping for development work items (separate feature)
- Retroactive updates to existing invoices
- Support for billing types other than "Hourly Rate" and "Monthly Fixed"

## Data Sources

### Notion Tables
1. **Contractor Payouts**: Source of payout entries
   - Relation: `00 Service Rate` → Contractor Rates
   - Relation: `00 Task Order` → Task Order Log

2. **Contractor Rates**: Billing configuration
   - Fields: Billing Type (select), Hourly Rate (number), Currency (select)

3. **Task Order Log**: Hours worked
   - Field: Final Hours Worked (formula)

## Constraints

- Must maintain existing multi-currency support (from recent implementation)
- Must follow DEBUG logging pattern established in recent invoice changes
- Must reuse existing service layer methods where possible
- No changes to Notion database schema
- No changes to PDF template (uses existing fields)

## User Confirmations

**Date**: 2026-01-07

1. **Scope**: Apply to ALL Service Fees (USD and VND) where BillingType = "Hourly Rate" ✅
2. **Fallback**: Keep current behavior if data missing or BillingType not "Hourly Rate" ✅
3. **Date Range**: Use invoice month range (1st to last day) ✅
4. **Aggregation**: Sum ALL hourly-rate Service Fees into single line item ✅

## Related Documents

- Existing implementation plan: `/Users/quang/.claude/plans/curried-finding-peacock.md`
- Current session: `docs/sessions/202601071248-revise-contractor-invoice-pdf`
- Codebase patterns: `CLAUDE.md`
