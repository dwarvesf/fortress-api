# ADR-003: Reset Approval Status on Line Item Update

## Status

Accepted

## Context

When a line item is updated (hours changed, new timesheets added), the existing approval status may no longer be valid. The approver approved specific hours/data that has now changed.

## Decision

When updating a line item, set both Order and Line Item status to "Pending Approval".

## Rationale

- **Data integrity** - Approval was for original data, not updated data
- **Audit trail** - Re-approval ensures someone reviews the changes
- **Workflow consistency** - Order status should reflect line item states
- **Prevents auto-approval** - Changed data doesn't slip through

## Consequences

### Positive
- Ensures updated data gets reviewed
- Maintains approval workflow integrity
- Clear audit trail of changes

### Negative
- Previously approved orders may need re-approval
- Additional notification may be needed to alert approvers

## Implementation

1. `UpdateTimesheetLineItem` sets Line Item status to "Pending Approval"
2. `UpdateTimesheetLineItem` also updates parent Order status to "Pending Approval"
3. Add `UpdateOrderStatus` method if not exists
