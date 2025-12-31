# Requirements: Commission Payout Processing

## Overview

Add commission payout processing to the `create-contractor-payouts` cronjob endpoint. When called with `?type=commission`, the endpoint should query pending invoice splits and create corresponding payout entries in Contractor Payouts database.

## Context

The existing cronjob endpoint (`POST /cronjobs/create-contractor-payouts`) already supports:
- `contractor_payroll` - Processes Contractor Fees with Payment Status=New
- `refund` - Processes approved Refund Requests

The `commission` type is currently not implemented and returns HTTP 501.

## Functional Requirements

### FR1: Query Pending Commission Splits
- Query Invoice Splits database for records with:
  - Status = "Pending"
  - Type = "Commission"
- Extract data: PageID, Name, Amount, Currency, Person (relation), Role

### FR2: Idempotency Check
- Before creating a payout, check if one already exists for the invoice split
- Use "Invoice Split" relation on Contractor Payouts to detect duplicates
- Skip if payout already exists

### FR3: Create Commission Payout
- Create a new entry in Contractor Payouts database with:
  - Name: From invoice split Name
  - Person: From invoice split Person relation
  - Invoice Split: Relation to the invoice split page
  - Amount: From invoice split Amount
  - Currency: From invoice split Currency
  - Type: "Commission"
  - Direction: "Outgoing"
  - Status: "Pending"

### FR4: Response Format
- Return JSON response consistent with other payout types:
  - `payouts_created`: Number of payouts created
  - `splits_processed`: Total splits found
  - `splits_skipped`: Skipped due to existing payout or validation
  - `errors`: Number of errors
  - `details`: Array of per-split processing details
  - `type`: "Commission"

## Non-Functional Requirements

### NFR1: Logging
- DEBUG level logs for tracing each step
- INFO level for summary counts
- ERROR level for failures

### NFR2: Error Handling
- Continue processing other splits if one fails
- Log errors but don't fail entire batch

## Database Schema

### Invoice Splits (Source)
| Property | Type | Usage |
|----------|------|-------|
| Name | Title | Payout name |
| Amount | Number | Payout amount |
| Currency | Select | VND, USD |
| Status | Select | Filter: Pending |
| Type | Select | Filter: Commission |
| Person | Relation | → Contractors (for payout Person) |

### Contractor Payouts (Target)
| Property | Type | Usage |
|----------|------|-------|
| Name | Title | From split Name |
| Amount | Number | From split Amount |
| Currency | Select | From split Currency |
| Person | Relation | From split Person |
| Invoice Split | Relation | Link to source split |
| Type | Select | "Commission" |
| Direction | Select | "Outgoing" |
| Status | Status | "Pending" |

## API Endpoint

```
POST /api/v1/cronjobs/create-contractor-payouts?type=commission
```

## Dependencies

- Existing: `InvoiceSplitService`, `ContractorPayoutsService`
- Extends invoice splits generation feature (session 202512311605)
