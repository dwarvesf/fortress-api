# ADR-002: Remove Deployment Field from Order Type

## Status

Accepted

## Context

Currently, when creating an Order record (Type=Order), the system sets a Deployment relation using the first project's deployment found. This is arbitrary because:

- A contractor may have multiple projects/deployments in the same month
- The "first" project is determined by Go map iteration order (random)
- The Deployment field on Order provides no meaningful information

## Decision

Do not set the `Deployment` field for Type=Order records.

## Rationale

- **Order is a container** - It groups line items for a contractor+month
- **Line items have Deployment** - Each Timesheet line item already links to its specific Deployment
- **No value added** - One arbitrary Deployment on Order is misleading
- **Cleaner data model** - Order identifies contractor via its child line items

## Consequences

### Positive
- Cleaner, more accurate data model
- No misleading arbitrary deployment link
- Simpler Order creation logic

### Negative
- Must traverse to line items to find deployments (if needed)
- Existing Orders may have Deployment set (historical data)

## Implementation

1. Modify `CreateOrder` to not set Deployment relation
2. Remove `deploymentID` parameter from `CreateOrder` function
3. Update handler to not search for first deployment for Order creation
