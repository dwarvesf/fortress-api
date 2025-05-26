# ADR-XXX: Introduction of Deal Closing Commissions and Soft Delete for Commission Records

**Status:** Proposed

**Date:** 2024-07-02

**Context:**

There is a requirement to calculate and assign commissions to employees who are involved in the "deal closing" phase of a project. The definitive source for identifying these individuals is expected to be a specific field within the project's Notion page. Additionally, commission calculations may need to be performed multiple times for an invoice (e.g., if project details change or corrections are needed). A mechanism is required to update commissions without losing the historical record of previous calculations, facilitating audits and potential rollbacks.

**Decision:**

1.  **New Commission Position & Type:**
    *   A new position value, `'deal-closing'`, will be added to the `project_head_positions` ENUM in the database.
    *   A corresponding constant `model.ProjectHeadPositionDealClosing` will be defined in the Go data models.
    *   A new commission type, `model.CommissionTypeDealClosing` (with value `'deal_closing'`), will be introduced for `EmployeeCommission` records.

2.  **Notion Integration for Deal Closing Personnel:**
    *   The `notionService.GetProjectHeadDisplayNames` function will be enhanced:
        *   It will fetch data from a "Deal Closing" multi-select field on the Notion project page. The expected format for each selection is "Employee Name (employee@email.com)".
        *   It will parse the email addresses from these selections.
        *   It will synchronize this information with the `project_heads` table in the application's database. For a given project:
            *   Existing `project_heads` entries with the `deal-closing` position will be soft-deleted.
            *   New `project_heads` entries will be created for each employee identified from Notion and found in the database via their email.

3.  **Commission Calculation Logic:**
    *   The `invoice.calculateCommissionFromInvoice` function will incorporate logic for the "deal-closing" commission:
        *   It will identify active (not soft-deleted) `project_heads` with the `deal-closing` position for the invoice's associated project.
        *   The total commission allocated for deal closing will be 2% of the invoice's commissionable value.
        *   This 2% will be divided equally among all identified deal-closing personnel for that project. For example, if there are two deal-closing individuals, each will receive a 1% commission rate on the commissionable value.

4.  **New API Endpoint for Commission Calculation:**
    *   A new API endpoint, `POST /invoices/{invoice_id}/calculate-commissions`, will be created.
    *   This endpoint will accept an `invoice_id` as a path parameter and an optional `dry_run=true` query parameter.
    *   **Dry Run Behavior (`dry_run=true`):** The endpoint will calculate all applicable commissions for the invoice (including the new "deal-closing" type and any existing types) and return them in the response body without persisting any changes to the database.
    *   **Standard Behavior (`dry_run=false` or omitted):**
        *   The endpoint will first soft-delete any existing `EmployeeCommission` records associated with the specified `invoice_id`.
        *   It will then calculate the new set of commissions.
        *   These newly calculated commissions will be saved as new records in the `employee_commissions` table.
        *   The saved commissions will be returned in the response body.

5.  **Payroll Calculation Adjustment:**
    *   The existing payroll calculation logic will be reviewed and, if necessary, updated to ensure it only includes active (not soft-deleted) `EmployeeCommission` records.

**Consequences:**

*   **Positive:**
    *   Provides a clear and automated way to calculate commissions for deal-closing activities.
    *   Leverages Notion as the source of truth for deal-closing personnel, aligning with existing workflows.
    *   The soft-delete mechanism for commission records preserves historical data, which is beneficial for auditing, tracking changes over time, and simplifies potential rollback procedures.
    *   The `dry_run` option offers flexibility for users to preview commission calculations before committing them.
    *   Ensures that payroll accurately reflects the most current commission data.
*   **Negative:**
    *   Introduces schema changes to the database (new ENUM values for `project_head_positions` and potentially `commission_types`).
    *   Increases the complexity of the `notionService` due to the new parsing and synchronization logic.
    *   The commission calculation logic in the `invoice` controller becomes more multifaceted.
    *   Requires careful implementation of database transactions to maintain data integrity during the sync and commission calculation/saving processes.
    *   The front-end or any systems consuming commission data might need adjustments if they previously did not account for soft-deleted records (though GORM typically handles this for reads if models are configured correctly).
