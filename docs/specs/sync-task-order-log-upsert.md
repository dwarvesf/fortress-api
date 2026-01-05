# Sync Task Order Log - Upsert Enhancement

## Status: Draft

## Context

Current `SyncTaskOrderLogs` endpoint (`POST /cronjobs/sync-task-order-logs`) creates Task Order Log entries from approved timesheets but has a limitation:

**Problem**: If a timesheet is updated (hours changed) or new timesheets are added after initial sync, existing line items are NOT updated.

### Current Behavior

```
CheckLineItemExists(orderID, deploymentID)
  -> If exists: SKIP (no update)
  -> If not exists: CREATE
```

- Only checks existence by `orderID + deploymentID`
- No comparison of hours or timesheet count
- Running twice is idempotent but doesn't reflect changes

## Proposed Solution: Upsert Approach

Fetch existing line item, compare hours/timesheets, update if different.

### Detection Logic

Compare:
1. **Hours** - stored `Line Item Hours` vs newly calculated hours
2. **Timesheet count** - stored `Timesheet` relation count vs new timesheet IDs count
3. **Timesheet IDs** - stored timesheet page IDs vs new timesheet page IDs

### Benefits

- Accurate change detection
- No data loss (keeps same Notion page ID)
- Efficient (only updates when changed)
- Maintains Notion page history

## Files to Modify

| File | Changes |
|------|---------|
| `pkg/service/notion/task_order_log.go` | Add `GetLineItemDetails`, `UpdateTimesheetLineItem` methods |
| `pkg/handler/notion/task_order_log.go` | Update logic to fetch/compare/update |

## Decisions

### Decision 1: Remove Deployment from Order type

**Date**: 2025-01-05

**Context**: Order records currently link to a single Deployment (first project's deployment found). This is arbitrary and misleading since a contractor may have multiple projects/deployments.

**Decision**: Do not set `Deployment` field for Type=Order records.

**Rationale**:
- Order is a parent container for line items
- Each line item (Type=Timesheet) already links to its specific Deployment
- Setting one arbitrary Deployment on Order provides no value

**Impact**:
- Modify `CreateOrder` to not set Deployment relation
- Order identifies contractor via line items' deployments

### Decision 2: Reset status on line item update

**Date**: 2025-01-05

**Context**: When a line item is updated (hours changed, new timesheets added), the approval workflow needs to restart.

**Decision**: When updating a line item, set both Order and Line Item status to "Pending Approval".

**Rationale**:
- Updated data requires re-approval
- Order status reflects aggregate state of its line items
- Prevents auto-approval of changed data

**Impact**:
- `UpdateTimesheetLineItem` must also update parent Order status
- Both records reset to "Pending Approval" on any change

## Open Questions

- [ ] Should we also update `Proof of Works` (re-summarize)?
- [ ] Should we track `line_items_updated` count in response?
- [ ] Should we add a `force` param to always update regardless of changes?

## Related Code

- Handler: `pkg/handler/notion/task_order_log.go:32`
- Service: `pkg/service/notion/task_order_log.go:453` (CheckLineItemExists)
- Service: `pkg/service/notion/task_order_log.go:510` (CreateTimesheetLineItem)
