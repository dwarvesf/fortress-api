# Contractor Fees Database Schema

## Overview

**Database Name**: Contractor Fees
**Data Source ID**: `2c264b29-b84c-8037-807c-000bf6d0792c`
**Block ID**: `2c264b29-b84c-80e3-9bd8-c22e7612c1ea`
**Description**: Instance of contractor rate ~ task order (work)

## Properties

### Core Fields

| Property | ID | Type | Description |
|----------|-----|------|-------------|
| Billing Name | `title` | title | Primary identifier for the billing entry |
| Payment Status | `BkF\` | status | `New`, `Pending`, `Paid` |
| Invoice Date | `QD]A` | date | Date of invoice |
| Invoice Number | `{ShZ` | rich_text | Invoice reference number |
| Payment Date | `{yf?` | date | Date payment was made |
| Notes | `jyfE` | rich_text | Additional notes |
| Attachments | `^LN\`` | files | Supporting documents |
| Contractor Type | `oE_j` | select | Individual, Sole Proprietor, LLC, Corporation, Partnership, Non-Profit, Government Entity |

### Relations

| Property | ID | Type | Related Database |
|----------|-----|------|------------------|
| Task Order Log | `MyT\` | relation | `2b964b29-b84c-801e-ab9e-000b0662b987` |
| Contractor Rate | `a_@z` | relation | `2c464b29-b84c-80cf-bef6-000b42bce15e` |

### Rollups (from Contractor Rate)

| Property | ID | Source Property |
|----------|-----|-----------------|
| Hourly Rate | `{VlD` | Hourly Rate |
| Fixed Fee | `cQ~E` | Monthly Fixed |
| Currency | `EJ\|p` | Currency |
| Billing Type | `T\|LK` | Billing Type |

### Rollups (from Task Order Log)

| Property | ID | Source Property |
|----------|-----|-----------------|
| Total Hours Worked | `fmna` | Final Hours Worked |
| Proof of Works | `UWWr` | Proof of Works |

### Formulas

| Property | ID | Description |
|----------|-----|-------------|
| Total Amount | `=;nd` | `hourlyRate * workedHours` or `hourlyRate * workedHours + fixedFee` |
| Contractor | `ArKD` | Derived from Task Order Log → Contractor |
| Auto Name | `_JzS` | Derived from Task Order Log |

## Payment Status Values

| Status | Color | Group |
|--------|-------|-------|
| New | blue | To-do |
| Pending | yellow | In progress |
| Paid | green | Complete |

## Contractor Type Values

| Option | Color |
|--------|-------|
| Individual | blue |
| Sole Proprietor | green |
| LLC | purple |
| Corporation | red |
| Partnership | orange |
| Non-Profit | pink |
| Government Entity | gray |

## Related Databases

| Database | ID | Relationship |
|----------|-----|--------------|
| Task Order Log | `2b964b29-b84c-801e-ab9e-000b0662b987` | Source of hours worked, proof of works |
| Contractor Rates | `2c464b29-b84c-80cf-bef6-000b42bce15e` | Source of rates, billing type, currency |
| Contractor Payouts | `2c564b29-b84c-8045-80ee-000bee2e3669` | References this via Billing relation |

## Query Examples

### Query by Payment Status

```javascript
{
  "filter": {
    "property": "Payment Status",
    "status": {
      "equals": "Pending"
    }
  }
}
```

### Query by Contractor Rate

```javascript
{
  "filter": {
    "property": "Contractor Rate",
    "relation": {
      "contains": "<contractor_rate_page_id>"
    }
  }
}
```

## Data Flow

```
Task Order Log (hours, proof of work)
        │
        ▼
Contractor Fees ◄── Contractor Rate (hourly rate, fixed fee, currency)
        │
        ▼
Contractor Payouts (aggregation for invoice)
```

## Notes

- Each Contractor Fee entry represents a billing instance for a specific task order
- Total Amount is calculated from hourly rate * hours worked, plus fixed fee if applicable
- Proof of Works is rolled up from the linked Task Order Log
- Payment Status tracks the lifecycle: New → Pending → Paid
