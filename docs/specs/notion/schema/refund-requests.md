# Refund Requests Database Schema

## Overview

- **Database ID**: `2cc64b29-b84c-8066-adf2-cc56171cedf4`
- **Title**: Refund Requests
- **Created**: 2025-12-17
- **Last Edited**: 2025-12-19
- **Icon**: 🧾 (emoji)
- **URL**: https://www.notion.so/2cc64b29b84c8066adf2cc56171cedf4

## Purpose

The Refund Requests database tracks refund requests from contractors back to the company. These include advance returns, deduction reversals, bonus refunds, and overpayments.

## Properties

### Core Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Refund ID` | Title | `title` | Refund request identifier |
| `Amount` | Number | `]Auv` | Refund amount |
| `Currency` | Select | `hK_S` | VND, USD |
| `Description` | Rich Text | `xGUO` | Refund description |
| `Details` | Rich Text | | LLM-generated refund details (from `Description Formatted` formula) |
| `Notes` | Rich Text | `bLWY` | Additional notes |
| `Payout` | Rich Text | `zOMw` | Related payout reference |
| `Date Requested` | Date | `JJjq` | When request was submitted |
| `Date Approved` | Date | `=[<I` | When request was approved |
| `Proof / Attachment` | Files | `i]Pt` | Supporting documents |

### System Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `ID` | Unique ID | `h_mY` | Auto-incrementing ID |
| `Created by` | Created By | `U:{g` | Who created this request |

### Status Property

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Status` | Status | `RiU@` | Current refund status |

#### Status Options

| Status | Color | Group | Description |
|--------|-------|-------|-------------|
| Pending | Yellow | In progress | Awaiting approval |
| Approved | Blue | In progress | Approved, awaiting payment |
| Paid | Green | Complete | Refund processed |
| Rejected | Red | Complete | Request rejected |
| Cancelled | Gray | Complete | Request cancelled |

### Reason Property

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Reason` | Select | `VUIc` | Reason for refund |

#### Reason Options

| Reason | Color | Description |
|--------|-------|-------------|
| Advance Return | Blue | Returning advance payment |
| Deduction Reversal | Green | Reversing a previous deduction |
| Bonus Refund | Purple | Returning bonus payment |
| Overpayment | Red | Returning overpaid amount |
| Other | Gray | Other reason |

### Relations

| Property | Type | ID | Related Database | Description |
|----------|------|-----|------------------|-------------|
| `Contractor` | Relation | `>EMM` | Contractors (`ed2b9224-97d9-4dff-97f9-82598b61f65d`) | Contractor making the refund |
| `Project (optional)` | Relation | `={uT` | Projects (`2988f9de-9886-4c6f-a3ff-7f7ef74b3732`) | Related project |

### Rollup Properties

| Property | Type | ID | Source | Description |
|----------|------|-----|--------|-------------|
| `Discord` | Rollup | `T\`T[` | From `Contractor` → `Discord` | Contractor's Discord username |
| `Person` | Rollup | `{LsV` | From `Contractor` → `Person` | Contractor's person name |
| `Work Email` | Rollup | `FGBc` | From `Contractor` → `Team Email` | Contractor's work email |

### Formula Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Auto Name` | Formula | `~_b_` | Auto-generates name: `RFD-{YYYY}-{discord}-{obfuscatedID}` |
| `Description Formatted` | Formula | | Formatted description used as LLM input for generating `Details` |

## Sample Data Structure

```json
{
  "object": "page",
  "id": "2cc64b29-b84c-8030-example",
  "properties": {
    "Refund ID": {
      "type": "title",
      "title": [
        {
          "text": {
            "content": "RFD-2025-thanh.pham-A1B2"
          }
        }
      ]
    },
    "Amount": {
      "type": "number",
      "number": 500000
    },
    "Currency": {
      "type": "select",
      "select": {
        "name": "VND",
        "color": "yellow"
      }
    },
    "Status": {
      "type": "status",
      "status": {
        "name": "Pending",
        "color": "yellow"
      }
    },
    "Reason": {
      "type": "select",
      "select": {
        "name": "Advance Return",
        "color": "blue"
      }
    },
    "Contractor": {
      "type": "relation",
      "relation": [
        {"id": "contractor-page-id"}
      ]
    },
    "Description": {
      "type": "rich_text",
      "rich_text": [
        {
          "text": {
            "content": "Returning advance from December project"
          }
        }
      ]
    }
  }
}
```

## API Integration

### Query Refund Request by ID

```javascript
// Use page retrieve API with refund request page ID
```

### Query Pending Refund Requests

```javascript
{
  "filter": {
    "property": "Status",
    "status": {
      "equals": "Pending"
    }
  }
}
```

### Query by Contractor

```javascript
{
  "filter": {
    "property": "Contractor",
    "relation": {
      "contains": "<contractor_page_id>"
    }
  }
}
```

## Related Databases

- **Contractors** (`ed2b9224-97d9-4dff-97f9-82598b61f65d`): Contractor making the refund
- **Projects** (`2988f9de-9886-4c6f-a3ff-7f7ef74b3732`): Related project (optional)
- **Payouts** (`2c564b29-b84c-8045-80ee-000bee2e3669`): Links via Refund Request relation

## Notes

- Refund amounts are typically incoming payments (contractor pays company)
- Status flow: Pending → Approved → Paid (or Rejected/Cancelled)
- Linked to Payouts table via `Refund Request` relation for invoice generation
