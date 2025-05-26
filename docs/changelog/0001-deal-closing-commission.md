# Changelog: Deal Closing Commission and Enhanced Calculation

**Date:** 2024-07-02

## ✨ New Features

*   **Deal Closing Commission:** Introduced a new commission type for employees involved in the "deal closing" phase of projects.
    *   The system now identifies deal-closing personnel based on a designated "Deal Closing" multi-select field in the project's Notion page.
    *   A commission of 2% of the invoice's commissionable value is allocated for deal closing and is distributed equally among the identified deal-closing personnel for that project.
*   **Commission Calculation Endpoint (`POST /invoices/{invoice_id}/calculate-commissions`):**
    *   Added a new API endpoint to calculate (and re-calculate) all commissions for a given invoice.
    *   This endpoint supports a `dry_run=true` query parameter, allowing users to preview commission calculations without saving them to the database.
    *   When not in dry run mode, this endpoint will update existing commissions for an invoice by soft-deleting the old records and creating new ones. This preserves historical commission data for auditing and tracking.

## 🛠️ Improvements

*   **Notion Integration:** Enhanced the Notion service to synchronize deal-closing personnel from Notion to the Fortress database, ensuring `ProjectHead` records are kept up-to-date.
*   **Data Integrity:** Commission updates now use a soft-delete mechanism. This means previous commission records for an invoice are marked as deleted but retained in the database, rather than being permanently removed. This improves auditability and allows for easier tracking of commission history.

## ⚙️ Technical Changes (for Developers)

*   Added `'deal-closing'` to the `project_head_positions` database ENUM.
*   Added `'deal_closing'` to the `commission_types` database ENUM (if applicable, or as a new string constant in models).
*   The `notionService.GetProjectHeadDisplayNames` function now handles the new "Deal Closing" Notion field and syncs data to `ProjectHead` records.
*   The `invoice.calculateCommissionFromInvoice` function now includes logic for the new deal-closing commission type.
*   Payroll calculation logic should be verified to ensure it only processes active (not soft-deleted) commission records.

## ⚠️ Important Notes

*   Ensure the "Deal Closing" field in your project Notion pages is a multi-select type, and each selected option is formatted as `Employee Name (employee@email.com)` for correct parsing.
*   Existing payroll systems or reports consuming commission data should be checked to ensure they correctly handle the soft-delete status of commission records (i.e., only process records where `deleted_at` is null).
