# Contractor Billing Database Schema

## Overview

- **Database ID**: `2c264b29-b84c-8037-807c-000bf6d0792c`
- **Title**: Contractor Billing
- **Created**: 2025-12-07
- **Last Edited**: 2025-12-29
- **Icon**: Receipt (gray)
- **URL**: https://www.notion.so/2c264b29b84c80e39bd8c22e7612c1ea

## Purpose

The Contractor Billing database links Task Order Logs to Contractor Rates, aggregating billing information for contractor payouts. Each billing record connects a specific task order to the contractor's rate structure, enabling automatic calculation of amounts based on hourly or fixed billing types.

## Properties

### Core Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Billing Name` | Title | `title` | Billing entry name/description |
| `Invoice Number` | Rich Text | `{ShZ` | Invoice reference number |
| `Invoice Date` | Date | `QD]A` | Date of invoice |
| `Payment Date` | Date | `{yf?` | Date payment was processed |
| `Notes` | Rich Text | `jyfE` | Additional notes |
| `Attachments` | Files | `^LN\`` | Attached documents |

### Status Property

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Payment Status` | Status | `BkF\` | Current payment status |

#### Status Options

| Status | Color | Group | Description |
|--------|-------|-------|-------------|
| New | Blue | To-do | New billing entry |
| Pending | Yellow | In progress | Payment awaiting processing |
| Paid | Green | Complete | Payment completed |

### Contractor Type Property

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Contractor Type` | Select | `oE_j` | Type of contractor entity |

#### Contractor Type Options

| Type | Color | Description |
|------|-------|-------------|
| Individual | Blue | Individual contractor |
| Sole Proprietor | Green | Sole proprietorship |
| LLC | Purple | Limited Liability Company |
| Corporation | Red | Corporation |
| Partnership | Orange | Partnership |
| Non-Profit | Pink | Non-profit organization |
| Government Entity | Gray | Government entity |

### Relations

| Property | Type | ID | Related Database | Description |
|----------|------|-----|------------------|-------------|
| `Task Order Log` | Relation | `MyT\` | Task Order Log (`2b964b29-b84c-801c-accb-dc8ca1e38a5f`) | Links to task order entry with hours and proof of work |
| `Contractor Rate` | Relation | `a_@z` | Contractor Rates (`2c464b29-b84c-805b-bcde-dc052e613f4d`) | Links to contractor's rate configuration |

### Rollup Properties

| Property | Type | ID | Source | Description |
|----------|------|-----|--------|-------------|
| `Total Hours Worked` | Rollup | `fmna` | From `Task Order Log` → `Final Hours Worked` | Total hours from task order |
| `Proof of Works` | Rollup | `UWWr` | From `Task Order Log` → `Proof of Works` | Work description from task order |
| `Hourly Rate` | Rollup | `{VlD` | From `Contractor Rate` → `Hourly Rate` | Contractor's hourly rate |
| `Fixed Fee` | Rollup | `cQ~E` | From `Contractor Rate` → `Monthly Fixed` | Contractor's monthly fixed fee |
| `Billing Type` | Rollup | `T\|LK` | From `Contractor Rate` → `Billing Type` | Hourly Rate or Monthly Fixed |
| `Currency` | Rollup | `EJ\|p` | From `Contractor Rate` → `Currency` | Payment currency |

### Formula Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Total Amount` | Formula | `=;nd` | Calculates total based on billing type |
| `Contractor` | Formula | `ArKD` | Extracts contractor from Task Order Log |
| `Auto Name` | Formula | `_JzS` | Auto-generates name from Task Order Log |

#### Total Amount Formula

```javascript
let (
  hourlyRate, {{Hourly Rate}}.first(),
  workedHours, {{Total Hours Worked}}.first(),
  fixedFee, {{Fixed Fee}}.first(),

  if (
    hourlyRate > 0,
    hourlyRate * workedHours,
    hourlyRate * workedHours + fixedFee
  )
)
```

**Logic**:
- If `Hourly Rate > 0`: Calculate as `hourlyRate * workedHours`
- Otherwise: Use `fixedFee` (Monthly Fixed billing)

#### Contractor Formula

```javascript
{{Task Order Log}}.first().{{Contractor}}
```

Extracts contractor name from the linked Task Order Log entry.

## Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│                  Contractor Billing Table                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐         ┌─────────────────┐           │
│  │ Task Order Log  │         │ Contractor Rate │           │
│  │ (MyT\)          │         │ (a_@z)          │           │
│  └────────┬────────┘         └────────┬────────┘           │
│           │                           │                     │
│           ▼                           ▼                     │
│  ┌─────────────────┐         ┌─────────────────┐           │
│  │ Rollups:        │         │ Rollups:        │           │
│  │ - Hours Worked  │         │ - Hourly Rate   │           │
│  │ - Proof of Work │         │ - Fixed Fee     │           │
│  │ - Contractor    │         │ - Billing Type  │           │
│  └────────┬────────┘         │ - Currency      │           │
│           │                  └────────┬────────┘           │
│           │                           │                     │
│           └───────────┬───────────────┘                     │
│                       ▼                                     │
│              ┌─────────────────┐                           │
│              │ Formulas:       │                           │
│              │ - Total Amount  │                           │
│              │ - Contractor    │                           │
│              │ - Auto Name     │                           │
│              └────────┬────────┘                           │
│                       │                                     │
│                       ▼                                     │
│              ┌─────────────────┐                           │
│              │ Payouts Table   │                           │
│              │ (aggregation)   │                           │
│              └─────────────────┘                           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Sample Data Structure

### Contractor Billing Entry Example

```json
{
  "object": "page",
  "id": "2c264b29-b84c-8030-a95c-example",
  "properties": {
    "Billing Name": {
      "type": "title",
      "title": [
        {
          "text": {
            "content": "Order thanh.pham :: 2025 Dec"
          }
        }
      ]
    },
    "Invoice Number": {
      "type": "rich_text",
      "rich_text": [
        {
          "text": {
            "content": "CONTR-202512-A1B2"
          }
        }
      ]
    },
    "Invoice Date": {
      "type": "date",
      "date": {
        "start": "2025-12-15"
      }
    },
    "Payment Status": {
      "type": "status",
      "status": {
        "name": "Pending",
        "color": "yellow"
      }
    },
    "Contractor Type": {
      "type": "select",
      "select": {
        "name": "Individual",
        "color": "blue"
      }
    },
    "Task Order Log": {
      "type": "relation",
      "relation": [
        {"id": "2b964b29-b84c-800c-bb63-fe28a4546f23"}
      ]
    },
    "Contractor Rate": {
      "type": "relation",
      "relation": [
        {"id": "2c464b29-b84c-805b-bcde-example"}
      ]
    },
    "Total Hours Worked": {
      "type": "rollup",
      "rollup": {
        "number": 40,
        "function": "show_original"
      }
    },
    "Hourly Rate": {
      "type": "rollup",
      "rollup": {
        "number": 25,
        "function": "show_original"
      }
    },
    "Total Amount": {
      "type": "formula",
      "formula": {
        "number": 1000
      }
    },
    "Currency": {
      "type": "rollup",
      "rollup": {
        "string": "USD",
        "function": "show_original"
      }
    },
    "Billing Type": {
      "type": "rollup",
      "rollup": {
        "string": "Hourly Rate",
        "function": "show_original"
      }
    }
  }
}
```

## API Integration

### Query Contractor Billing by Task Order Log

```javascript
{
  "filter": {
    "property": "Task Order Log",
    "relation": {
      "contains": "<task_order_log_page_id>"
    }
  }
}
```

### Query Pending Billing Entries

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

## Workflow

### Billing Entry Creation

1. **Task Order Completion**: Task Order Log entry marked as Completed
2. **Billing Creation**: Create Contractor Billing entry linked to Task Order Log
3. **Rate Linking**: Link to appropriate Contractor Rate for the contractor
4. **Auto-Calculation**: Rollups and formulas automatically calculate amounts
5. **Invoice Generation**: Generate invoice from billing data
6. **Payout Creation**: Create Payout entry linked to this Contractor Billing

### Payment Processing

1. **New Status**: Billing entry created with "New" status
2. **Pending Status**: Invoice generated and sent, status updated to "Pending"
3. **Paid Status**: Payment processed, status updated to "Paid" with Payment Date

## Relationship to Other Databases

```
Contractor Rates  ──────┐
                        │
                        ▼
Task Order Log  ──────► Contractor Billing ──────► Payouts
                              │
                              │
                              ▼
                        Contractor Invoice
```

## Related Databases

- **Task Order Log** (`2b964b29-b84c-801c-accb-dc8ca1e38a5f`): Source of hours and proof of work
- **Contractor Rates** (`2c464b29-b84c-805b-bcde-dc052e613f4d`): Rate configuration for billing calculation
- **Payouts** (`2c564b29-b84c-8045-80ee-000bee2e3669`): Aggregated payout entries linking to this billing

## Notes

- The `Task Order Log` relation is the key link to hours worked and proof of work
- The `Contractor Rate` relation provides billing configuration (hourly rate vs fixed fee)
- `Total Amount` formula automatically calculates based on billing type:
  - Hourly Rate: `hourlyRate * workedHours`
  - Monthly Fixed: Uses `fixedFee` directly
- This database serves as the bridge between Task Order Log and Payouts tables
