# Requirements: Initialize Task Order Logs

## Business Context

At the beginning of each month, business needs to have Task Order Log entries (Orders with Line Items) pre-created for all active contractors. This allows:
1. Visibility into expected billing for the month
2. Tracking of contractor work before timesheets are approved
3. Upsert capability when timesheets are later approved via sync endpoint

## Functional Requirements

### FR-001: Initialize Task Order Logs Endpoint

Create an endpoint that initializes Task Order Log entries for all active deployments.

**Input:**
- `month` (required): Target month in YYYY-MM format

**Process:**
1. Query all active Deployments from Notion
2. For each deployment:
   - Check if an Order + Line Item already exists for that deployment and month
   - If not, create:
     - Order entry (Type="Order", Status="Pending Approval", Date=last day of month)
     - Line Item entry (Type="Timesheet", linked to Deployment, Parent item=Order, empty Timesheet relation)
3. Return counts of Orders and Line Items created

**Output:**
- `orders_created`: Count of new Orders created
- `line_items_created`: Count of new Line Items created
- `deployments_processed`: Total deployments processed
- `skipped`: Count of deployments skipped (already have Order/Line Item)

### FR-002: Schema Compatibility

The Line Item entries created must be compatible with the existing sync endpoint so that:
- When timesheets are approved later, the sync endpoint can upsert (update) these existing Line Items
- The Contractor relationship is established via: Line Item → Deployment → Contractor

## Non-Functional Requirements

- Debug logging for traceability
- Idempotent operation (safe to run multiple times)
- Error handling for individual deployment failures (continue processing others)

## Existing Code Context

- `CreateOrder`: Creates Order entry (Type="Order")
- `CreateTimesheetLineItem`: Creates Line Item entry (Type="Timesheet")
- `CheckOrderExistsByContractor`: Checks if Order exists for contractor+month
- `CheckLineItemExists`: Checks if Line Item exists for order+deployment

## Open Questions

None - requirements are clear.
