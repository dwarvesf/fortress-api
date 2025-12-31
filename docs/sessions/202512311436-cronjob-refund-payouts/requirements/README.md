# Requirements: Cronjob Refund Payouts

## Overview

Add refund payout type support to the existing `POST /cronjobs/contractor-payouts?type=refund` endpoint.

## User Clarifications

1. **Direction**: Outgoing (company pays contractor for refund)
2. **Status Update**: Do NOT update Refund Request status to "Paid" after payout creation

## Functional Requirements

### FR-1: Query Approved Refund Requests
- Query Refund Requests database with `Status=Approved`
- Extract: PageID, Amount, Currency, Contractor relation, Reason, Description

### FR-2: Create Refund Payout
- Create Payout entry with:
  - Type: "Refund"
  - Direction: "Outgoing"
  - Refund Request relation: link to source request
  - Person: Contractor from refund request
  - Amount, Currency: from refund request
  - Status: "Pending"

### FR-3: Idempotency
- Check if payout already exists for Refund Request before creating
- Skip if payout already exists

### FR-4: No Status Update
- Do NOT update Refund Request status after payout creation
- Leave status as "Approved"

## Technical Context

### Source Database
- **Refund Requests**: `2cc64b29-b84c-8066-adf2-cc56171cedf4`
- Key properties: Amount, Currency, Status, Reason, Contractor relation

### Target Database
- **Payouts**: `2c564b29-b84c-8045-80ee-000bee2e3669`
- Refund Request relation ID: `cS>|`

### Existing Pattern
- Follow `processContractorPayrollPayouts` pattern in `pkg/handler/notion/contractor_payouts.go`
