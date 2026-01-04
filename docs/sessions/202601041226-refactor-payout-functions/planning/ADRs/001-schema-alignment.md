# ADR-001: Align Code with Updated Notion Payouts Schema

## Status
Accepted

## Context
The Notion "Contractor Payouts" database schema was updated on 2026-01-04 with:
- Removal of `Type`, `Direction`, `Month` as settable properties (now formulas or removed)
- Renaming of relation properties with prefix conventions (`00 Task Order`, `01 Refund`, `02 Invoice Split`)
- New `Description` property
- Source Type formula now outputs `Service Fee` instead of `Contractor Payroll`

## Decision
Refactor the Go code to align with the new schema by:

1. **Update constants**: Change `PayoutSourceTypeContractorPayroll` value to `"Service Fee"`
2. **Remove Direction**: Remove `PayoutDirection` type usage from `PayoutEntry` struct
3. **Update property names**: Change all Notion API property name strings to new names
4. **Remove formula writes**: Remove attempts to write `Type`, `Month`, `Direction` properties
5. **Add Description**: Add optional `Description` field to create inputs

## Consequences

### Positive
- Code matches current Notion schema
- Simpler code (fewer properties to manage)
- Formula-based Source Type is more reliable

### Negative
- Breaking change for any external callers (if any)
- Need to update all affected functions

### Risks
- None - Notion formulas handle backward compatibility
