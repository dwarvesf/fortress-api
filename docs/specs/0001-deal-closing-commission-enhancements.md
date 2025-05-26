# Specification: Deal Closing Commission Enhancements

**Version:** 1.0

**Date:** 2024-07-02

## 1. Overview

This document outlines the technical specifications for implementing a new "deal-closing" commission type, integrating with Notion to identify relevant personnel, and updating the commission calculation and storage mechanisms.

## 2. Data Model Changes

### 2.1. `project_head_positions` ENUM (Database & Go Model)

*   **Database (Migration):**
    *   An `ALTER TYPE project_head_positions ADD VALUE IF NOT EXISTS 'deal-closing';` migration script will be created.
*   **Go Model (`pkg/model/project_head.go` or similar):**
    *   Add constant: `const ProjectHeadPositionDealClosing ProjectHeadPosition = "deal-closing"`

### 2.2. `commission_types` ENUM (Database & Go Model) - *Potentially*

*   **Database (Migration - if `commission_types` is an ENUM):**
    *   An `ALTER TYPE commission_types ADD VALUE IF NOT EXISTS 'deal_closing';` migration script will be created.
*   **Go Model (`pkg/model/commission.go` or similar):**
    *   Add constant: `const CommissionTypeDealClosing CommissionType = "deal_closing"`
    *   If `CommissionType` is a string type, this is a new constant definition. If it's a custom enum type, ensure it's compatible.

### 2.3. `ProjectHead` Model (`pkg/model/project_head.go`)

*   No new fields are strictly required for this model if it already supports soft deletes (e.g., via `gorm.DeletedAt`).
*   The `Position` field will now utilize the new `ProjectHeadPositionDealClosing` value.

### 2.4. `EmployeeCommission` Model (`pkg/model/commission.go`)

*   The `Type` field will now utilize the new `CommissionTypeDealClosing` value.
*   Ensure it supports soft deletes (e.g., via `gorm.DeletedAt`).

## 3. Service Layer Modifications (`pkg/service/notion/`)

### 3.1. `IService` Interface (`interface.go`)

*   Modify signature:
    ```go
    GetProjectHeadDisplayNames(pageID string) (salePersonName, techLeadName, accountManagerNames, dealClosingEmails string, err error)
    ```

### 3.2. `notionService` Struct (`notion.go`)

*   Add field: `db *gorm.DB`
*   Update constructor `New(secret, projectID string, l logger.Logger, db *gorm.DB) IService`

### 3.3. Helper Function: `extractEmailFromOptionName` (`notion.go`)

*   Signature: `func extractEmailFromOptionName(optionName string) string`
*   Purpose: Parses an email from a string like "Name (email@example.com)".
*   Logic: Uses regex `\(([^)]+)\)` to find text in parentheses. Checks if it contains "@". Returns trimmed email or empty string.

### 3.4. `GetProjectHeadDisplayNames` Method (`notion.go`)

*   **Return Signature:** `(salePersonName, techLeadName, accountManagerNames, dealClosingEmails string, err error)`
*   **Input:** `pageID string`
*   **Logic:**
    1.  Fetch Notion properties via `n.GetProjectInDB(pageID)`.
    2.  Extract `salePersonName`, `techLeadName`, `accountManagerNames` as currently done.
    3.  Access the "Deal Closing" property (type: `nt.DBPropTypeMultiSelect`).
    4.  Iterate options, use `extractEmailFromOptionName` to get emails. Store valid emails.
    5.  Set `dealClosingEmails` (comma-separated string of extracted emails).
    6.  **Database Sync (Transactioned):**
        *   Find `model.Project` by `pageID`.
        *   If project found in DB:
            *   Begin GORM transaction.
            *   Soft delete: `tx.Where("project_id = ? AND position = ?", project.ID, model.ProjectHeadPositionDealClosing).Delete(&model.ProjectHead{})`.
            *   For each extracted email:
                *   Find `model.Employee` by email (company or personal).
                *   If employee found, create and save `model.ProjectHead`:
                    *   `ProjectID`: `project.ID`
                    *   `EmployeeID`: `employee.ID`
                    *   `Position`: `model.ProjectHeadPositionDealClosing`
            *   Commit transaction. Handle rollback on errors.
        *   If no deal closing emails found in Notion, still perform the soft delete of existing `deal-closing` project heads in the DB for that project.

## 4. Controller Layer Modifications (`pkg/controller/invoice/`)

### 4.1. `calculateCommissionFromInvoice` Function (`commission.go`)

*   **Signature (ensure `db *gorm.DB` is available):** `func (c *controller) calculateCommissionFromInvoice(db *gorm.DB, l logger.Logger, invoice *model.Invoice) ([]model.EmployeeCommission, error)`
*   **New Logic for Deal Closing:**
    1.  Fetch active deal-closing heads: `db.Where("project_id = ? AND position = ? AND deleted_at IS NULL", invoice.ProjectID, model.ProjectHeadPositionDealClosing).Find(&dealClosingHeads)`.
    2.  If `len(dealClosingHeads) > 0`:
        *   `commissionRate := 2.0 / float64(len(dealClosingHeads))`.
        *   For each `head` in `dealClosingHeads`:
            *   Create `model.EmployeeCommission`:
                *   `EmployeeID`: `head.EmployeeID`
                *   `InvoiceID`: `invoice.ID`
                *   `ProjectID`: `invoice.ProjectID`
                *   `Amount`: `invoice.CommissionableValue * commissionRate`
                *   `Rate`: `commissionRate`
                *   `Type`: `model.CommissionTypeDealClosing`
                *   `Note`: e.g., "Deal Closing Commission"
            *   Append to the result slice.

## 5. API Endpoint (`pkg/handler/invoice/invoice.go` or similar)

### 5.1. Endpoint Definition

*   **Path:** `POST /invoices/{id}/calculate-commissions` (or `/invoices/{id}/commissions`)
*   **HTTP Method:** `POST`
*   **Path Parameter:** `id` (UUID of the invoice)
*   **Query Parameter:** `dry_run` (boolean, e.g., `?dry_run=true`)

### 5.2. Handler Function Logic

1.  Parse `invoiceID` from path and `dryRun` from query.
2.  Fetch `model.Invoice` by `invoiceID`. Handle not found.
3.  Call `controller.CalculateCommissionFromInvoice(db, logger, &invoice)`.
4.  If `dryRun` is `true`:
    *   Return calculated commissions as JSON (HTTP 200).
5.  If `dryRun` is `false`:
    *   Begin GORM transaction.
    *   Soft delete existing commissions: `tx.Where("invoice_id = ? AND deleted_at IS NULL", invoiceID).Delete(&model.EmployeeCommission{})`.
    *   Save newly calculated commissions: `tx.Create(&newCommissions)`.
    *   Commit transaction. Handle rollback on errors.
    *   Return saved commissions as JSON (HTTP 200 or 201).

## 6. Payroll Calculation Review

*   Locate functions responsible for payroll calculations that sum `EmployeeCommission` amounts.
*   Ensure these functions query for `EmployeeCommission` records with `deleted_at IS NULL` to only include active/current commissions.
    *   Example GORM query: `db.Where("employee_id = ? AND payroll_id = ? AND deleted_at IS NULL", empID, payrollID).Find(&commissions)`.

## 7. Assumptions

*   `ProjectHead` and `EmployeeCommission` GORM models are already configured for soft deletes (i.e., have a `DeletedAt gorm.DeletedAt` field).
*   `invoice.CommissionableValue` correctly represents the value upon which commissions are calculated.
*   The Notion page property for "Deal Closing" will be a multi-select field, and the selected options will follow the format "Name (email@domain.com)".
