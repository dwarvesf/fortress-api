# Requirements: Contractor Payables Record Creation

## Overview

After successfully generating and uploading a contractor invoice PDF to Google Drive, automatically create a corresponding record in the Contractor Payables Notion database.

## Business Context

The Contractor Payables database aggregates payout items into payable records for each contractor. Currently, invoices are generated and uploaded but no tracking record is created in Notion, making it difficult to:
- Track invoice status (New → Pending → Paid)
- Link invoices to their source payout items
- Manage the payment lifecycle

## Functional Requirements

### FR-1: Create Payable Record After Upload
- **Trigger**: After successful PDF upload to Google Drive (not when `skipUpload=true`)
- **Action**: Create a new record in Contractor Payables Notion database
- **Database ID**: `2c264b29-b84c-8037-807c-000bf6d0792c`

### FR-2: Populate Required Fields
| Field | Source | Notes |
|-------|--------|-------|
| `Payable` | Empty or auto-gen | Auto Name formula will fill |
| `Total` | `invoiceData.TotalUSD` | Final invoice total |
| `Currency` | "USD" | Invoice currency is always USD |
| `Period` | `invoiceData.Month + "-01"` | Start of billing month |
| `Invoice Date` | Current date | When invoice was generated |
| `Invoice ID` | `invoiceData.InvoiceNumber` | e.g., CONTR-202512-A1B2 |
| `Payment Status` | "New" | Initial status |
| `Contractor` | Contractor page ID | Relation from rates query |
| `Payout Items` | Payout page IDs | Relation to aggregated payouts |
| `Attachments` | Google Drive URL | Link to uploaded PDF |

### FR-3: Non-Blocking Error Handling
- If Notion record creation fails, log the error but continue with the response
- Do not fail the entire invoice generation request

## Non-Functional Requirements

### NFR-1: Logging
- DEBUG level logs for all operations
- Log input parameters, Notion API calls, and results

### NFR-2: Configuration
- Database ID should be configurable via environment variable
- `NOTION_CONTRACTOR_PAYABLES_DB_ID`

## Out of Scope
- Updating existing payable records
- Deleting payable records
- Querying payable records
- Status transitions (handled separately)

## Dependencies
- Existing Contractor Invoice generation flow
- Notion API integration patterns (see `contractor_payouts.go`)
- Google Drive upload (already implemented)

## Acceptance Criteria
1. After successful invoice upload, a record appears in Contractor Payables database
2. All required fields are populated correctly
3. Relations (Contractor, Payout Items) link to correct records
4. PDF attachment URL is accessible
5. `skipUpload=true` does NOT create a payable record
6. Failed Notion creation does not break invoice generation
