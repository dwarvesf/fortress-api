# Requirements: Refactor Payout Functions for Updated Notion Schema

## Context

The Notion "Contractor Payouts" database schema has been updated. The Go codebase functions that interact with this database need to be refactored to match the new schema.

## Schema Changes Summary

### Removed Properties
| Property | Old Type | Reason |
|----------|----------|--------|
| `Type` | Select | Replaced by `Source Type` formula (auto-calculated) |
| `Direction` | Select | Removed from schema entirely |
| `Month` | Rich Text | Now a formula calculated from `Date` field |

### Renamed Relations
| Old Name | New Name |
|----------|----------|
| `Billing` | `00 Task Order` |
| `Refund` | `01 Refund` |
| `Invoice Split` | `02 Invoice Split` |

### Source Type Value Changes
| Old Value | New Value |
|-----------|-----------|
| `Contractor Payroll` | `Service Fee` |

### New Properties
| Property | Type | Description |
|----------|------|-------------|
| `Description` | Rich Text | For notes/description |

## Functional Requirements

### FR-1: Update Property Name Mappings
All Notion API queries and writes must use the new property names:
- Queries filtering by `Billing` must use `00 Task Order`
- Queries filtering by `Refund` must use `01 Refund`
- Queries filtering by `Invoice Split` must use `02 Invoice Split`

### FR-2: Remove Unsettable Property Writes
The following properties are now formulas and cannot be written:
- `Type` - Do not write this property
- `Month` - Do not write this property (set `Date` instead)
- `Direction` - Property no longer exists, do not write

### FR-3: Update Source Type Constants
Update the `PayoutSourceType` constant from `"Contractor Payroll"` to `"Service Fee"` to match the formula output.

### FR-4: Add Description Support
Add support for the new `Description` property in payout creation inputs.

## Affected Files

1. `pkg/service/notion/payout_types.go`
2. `pkg/service/notion/contractor_payouts.go`
3. `pkg/handler/notion/contractor_payouts.go`

## Non-Functional Requirements

### NFR-1: Backward Compatibility
- Existing payouts in Notion should continue to work
- No data migration required (Notion formulas auto-calculate)

### NFR-2: Logging
- Maintain existing DEBUG logging patterns
- Update log messages to reflect new property names

## Acceptance Criteria

1. All payout creation functions work with new schema
2. All payout query functions work with new property names
3. No writes to removed/formula properties
4. Existing tests pass (after updates)
