# Sync Task Order Log - Upsert Enhancement Requirements

## Overview

Enhance the `SyncTaskOrderLogs` endpoint to support updating existing line items when timesheet data changes.

## Problem Statement

Current behavior: If a timesheet is updated (hours changed) or new timesheets are added after initial sync, existing line items are NOT updated. The sync is idempotent but only for creation - no updates occur.

## Requirements

### Functional Requirements

1. **Detect changes in timesheets**
   - Compare stored hours vs newly calculated hours
   - Compare stored timesheet relations vs new timesheet IDs

2. **Update line items when changes detected**
   - Update `Line Item Hours` field
   - Update `Timesheet` relations
   - Update `Proof of Works` (re-summarize via LLM)

3. **Reset approval status on update**
   - Set updated Line Item status to "Pending Approval"
   - Set parent Order status to "Pending Approval"

4. **Remove Deployment from Order type**
   - Order records should not set Deployment relation
   - Only Line Items (Type=Timesheet) link to Deployment

### Non-Functional Requirements

1. Maintain idempotency - running multiple times should be safe
2. Add DEBUG logging for change detection
3. Track `line_items_updated` count in response

## Decisions Made

### Decision 1: Remove Deployment from Order type

- Order is a parent container for line items
- Each line item already links to its specific Deployment
- Setting one arbitrary Deployment on Order provides no value

### Decision 2: Reset status on line item update

- Updated data requires re-approval
- Order status reflects aggregate state of its line items
- Both Order and Line Item reset to "Pending Approval" on any change

## Source

Based on discussion documented in: `docs/specs/sync-task-order-log-upsert.md`
