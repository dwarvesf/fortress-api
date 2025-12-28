# Requirements: Discord Invoice Paid Command with Notion Integration

**Session:** 202512261508-discord-invoice-paid-notion
**Date:** 2025-12-26
**Status:** Clarified

## Summary

Rework the `?inv paid` Discord command to support marking invoices as paid from both PostgreSQL (Fortress) and Notion Client Invoices database.

## Background

- Previously used NocoDB accounting todos to trigger invoice payment via webhook
- NocoDB is being deprecated in favor of Notion
- Discord has a 3-second interaction timeout constraint

## Functional Requirements

### FR-1: Discord Command Flow
- User runs `?inv paid <invoice_number>`
- Validate args only (no API calls - must be fast for Discord timeout)
- Show confirmation dialog immediately
- After user confirms → show "Processing..." message
- Call fortress-api endpoint
- Display success/error result

### FR-2: fortress-api Endpoint
- New endpoint: `POST /api/v1/invoices/mark-paid`
- Request body: `{ "invoice_number": "INV-XXX" }`
- Search both PostgreSQL and Notion for the invoice
- Return result with source info (`postgres`, `notion`, or `both`)

### FR-3: PostgreSQL Invoice Processing
When invoice found in PostgreSQL:
- Validate status (must be `sent` or `overdue`)
- Update status to `paid`
- Set `paid_at` timestamp
- Store commission records
- Create accounting transaction
- Send thank you email
- Move PDF in GDrive (Sent → Paid folder)

### FR-4: Notion Invoice Processing
When invoice found in Notion Client Invoices database:
- Validate status (must be `Sent` or `Overdue`)
- Update `Status` property to `Paid`
- Set `Paid Date` property to today
- Send thank you email (extract recipients from Notion)
- Move PDF in GDrive (Sent → Paid folder)

### FR-5: Source Handling
- If invoice found in both → update both systems
- If found in only one → update that system
- If not found in either → return error

## Non-Functional Requirements

### NFR-1: Performance
- Discord command must respond within 3 seconds (show confirmation)
- Actual processing can be async after confirmation

### NFR-2: Logging
- Debug logs at each step for traceability
- Error logging for all failure cases

## Data Sources

### PostgreSQL
- Table: `invoices`
- Query by: `number` field

### Notion
- Database: Client Invoices (`2bf64b29b84c80879a52ed2f9d493096`)
- Query by: `(auto) Invoice Number` title property
- Key properties:
  - `Status` (status): Draft / Sent / Overdue / Paid / Cancelled
  - `Paid Date` (date)
  - `Recipients` (rollup): for thank you email
  - `Google Drive File` (relation): for PDF location

## Reference Documents

- Spec: `docs/specs/notion/discord-invoice-paid-command.md`
- Notion Schema: `docs/specs/notion/schema/client-invoices.md`
