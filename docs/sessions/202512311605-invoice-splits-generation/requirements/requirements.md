# Requirements: Invoice Splits Generation

## Overview

Generate invoice split records automatically when a client invoice is marked as paid, using the existing worker queue for background processing.

## Functional Requirements

### FR1: Trigger on Mark Paid
- When `?inv paid <invoice_number>` command marks an invoice as paid
- Enqueue a background job to generate invoice splits
- Return response immediately (non-blocking)

### FR2: Generate Invoice Splits from Commission Data
- Query line items from the paid invoice
- For each line item with commission data:
  - Extract: % Sales, % Account Mgr, % Delivery Lead, % Hiring Referral
  - Extract persons: Sales Person, Account Manager, Delivery Lead, Hiring Referral
  - Extract amounts: Sales Amount, Account Amount, Delivery Lead Amount, Hiring Referral Amount
- Create Invoice Split record for each role with amount > 0

### FR3: Invoice Split Record Structure
- Name: `{Role} - {Project Code} - {Month Year}`
- Amount: From calculated commission amount
- Currency: From parent invoice
- Month: From invoice Issue Date
- Role: Sales | Account Manager | Delivery Lead | Hiring Referral
- Type: Commission
- Status: Pending
- Relations: Person (→Contractors), Deployment, Invoice Item, Client Invoices

## Invoice Split Schema

| Column | Type | Description |
|--------|------|-------------|
| Name | Title | `{Role} - {Project Code} - {Month Year}` |
| Amount | Number | Commission amount |
| Currency | Select | VND, USD, etc. |
| Month | Date | Invoice issue date |
| Role | Select | Sales, Account Manager, Delivery Lead, Hiring Referral |
| Type | Select | Commission |
| Status | Select | Pending, Paid |
| Person | Relation | → Contractors database |
| Deployment | Relation | → Deployment Tracker database |
| Invoice Item | Relation | → Client Invoice Line Items database |
| Client Invoices | Relation | → Client Invoices database |

### FR4: Mark Splits Generated
- After all splits created, update invoice `Splits Generated = true`
- Idempotency: Skip if `Splits Generated` already true

## Non-Functional Requirements

### NFR1: Background Processing
- Use existing Worker queue infrastructure
- Non-blocking to mark-paid flow

### NFR2: Idempotency
- Check `Splits Generated` flag before processing
- Skip invoices already processed

### NFR3: Logging
- DEBUG level logs for tracing

## Notion Database IDs

- Client Invoices: `2bf64b29-b84c-80e2-8cc7-000bfe534203`
- Invoice Splits: `2c364b29-b84c-804f-9856-000b58702dea`
- Contractors: `9d468753-ebb4-4977-a8dc-156428398a6b`
- Deployment Tracker: `2b864b29-b84c-8079-9568-dc17685f4f33`

## Key Property IDs (Client Invoices)

| Property | ID |
|----------|-----|
| % Sales | `gZX~` |
| % Account Mgr | `iF]]` |
| % Delivery Lead | `r@c{` |
| % Hiring Referral | `uyk[` |
| Sales Amount | `\LY?` |
| Account Amount | `zZE:` |
| Delivery Lead Amount | `[}gB` |
| Hiring Referral Amount | `XM:A` |
| Sales Person | `@[Qm` |
| Account Manager | `>Ow^` |
| Delivery Lead | `\|Rjc` |
| Hiring Referral | `nYwg` |
| Splits Generated | `}WM}` |
| Deployment Tracker | `OMj?` |
