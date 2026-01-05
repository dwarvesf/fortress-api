# ADR-001: Initialize Task Order Logs via Active Deployments

## Status
Accepted

## Context
Business needs to pre-create Task Order Log entries at the beginning of each month for all active contractors. The current schema establishes contractor relationships through:
- Order → Sub-item (Line Items) → Deployment → Contractor

An empty Order has no contractor link. The contractor relationship must come through Line Items.

## Decision
Initialize Task Order Logs by:
1. Querying all active Deployments (using existing `QueryActiveDeploymentsByMonth`)
2. Grouping deployments by contractor
3. For each contractor:
   - Create one Order (if not exists)
   - Create one Line Item per deployment, linking to that Order and Deployment
4. Line Items are created with empty Timesheet relation (to be filled by sync endpoint later)

## Consequences

### Positive
- Establishes contractor link via Deployment
- Compatible with existing sync endpoint upsert logic
- One Order per contractor per month (not per deployment)
- Idempotent - safe to run multiple times

### Negative
- Creates "empty" Line Items that will be updated later
- Requires sync endpoint to handle upsert of existing Line Items

## Implementation Notes
- Reuse `CheckOrderExistsByContractor` to find existing Orders
- Reuse `CheckLineItemExists` to avoid duplicate Line Items
- Modify `CreateTimesheetLineItem` to allow empty Timesheet relation
