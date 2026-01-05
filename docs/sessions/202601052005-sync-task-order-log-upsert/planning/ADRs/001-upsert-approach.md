# ADR-001: Upsert Approach for Line Item Updates

## Status

Accepted

## Context

The `SyncTaskOrderLogs` endpoint creates Task Order Log entries from approved timesheets. Currently, if timesheets are updated after initial sync, the changes are not reflected in existing line items.

Three approaches were considered:
1. **Upsert approach** - Fetch existing line item, compare hours/timesheets, update if different
2. **Force refresh flag** - Add `?force=true` param to delete and recreate line items
3. **Version tracking** - Store timesheet `last_edited_time` and compare

## Decision

Use the **Upsert approach** (Option 1).

## Rationale

- **Accurate change detection** - Compares actual data (hours, timesheet count)
- **No data loss** - Keeps same Notion page ID, preserves history
- **Efficient** - Only updates when changes detected
- **Maintains Notion page history** - No delete/recreate cycles

## Consequences

### Positive
- Existing line items retain their page IDs
- Notion revision history preserved
- Minimal API calls (only update when needed)

### Negative
- More complex logic to fetch and compare existing data
- Additional API call to get line item details before comparing

## Implementation

1. Add `GetLineItemDetails` method to fetch existing line item data
2. Add `UpdateTimesheetLineItem` method to update changed fields
3. Modify handler to compare and decide create vs update
