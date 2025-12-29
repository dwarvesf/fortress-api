# Payouts Database Schema

## Overview

- **Database ID**: `2c564b29-b84c-8045-80ee-000bee2e3669`
- **Title**: Payouts
- **Created**: 2025-12-10
- **Last Edited**: 2025-12-29
- **Icon**: Lock keyhole (gray)
- **URL**: https://www.notion.so/2c564b29b84c806c9d73def95eb7ff30

## Purpose

The Payouts database tracks all payment transactions for contractors including billing, refunds, and invoice splits. Each payout record aggregates into contractor invoices.

**Description**:
```
billing = deployment + task + timesheet + rate
refund
invoice split

sum ( ) payout ⇒ contractor invoice
```

## Properties

### Core Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Payout` | Title | `title` | Payout name/description |
| `Amount` | Number | `qxc^` | Payment amount in dollars |
| `Month` | Rich Text | `^ewn` | Month in YYYY-MM format |
| `Date` | Date | `}kwE` | Transaction date |
| `Paid Date` | Date | `U>Mb` | Date payment was processed |
| `Project` | Rich Text | `T\|;f` | Project name |
| `Notes` | Rich Text | `PwUo` | Additional notes |
| `Invoice / Ref #` | Rich Text | `aH:>` | Invoice or reference number |

### Status Property

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Status` | Status | `Q\|wl` | Current payment status |

#### Status Options

| Status | Color | Group | Description |
|--------|-------|-------|-------------|
| Pending | Yellow | In progress | Payment awaiting processing |
| Paid | Green | Complete | Payment completed |
| Cancelled | Red | Complete | Payment cancelled |

### Type Property

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Type` | Select | `s{\`O` | Type of payout |

#### Type Options

| Type | Color | Description |
|------|-------|-------------|
| Contractor Payroll | Blue | Regular contractor payment |
| Commission | Green | Commission-based payment |
| Refund | Orange | Refund to contractor |
| Bonus | Purple | Bonus payment |
| Penalty | Red | Penalty deduction |

### Direction Property

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Direction` | Select | `R{{m` | Payment direction |

#### Direction Options

| Direction | Color | Description |
|-----------|-------|-------------|
| Outgoing (you pay) | Red | Company pays contractor |
| Incoming (you receive) | Green | Contractor pays company |

### Relations

| Property | Type | ID | Related Database | Description |
|----------|------|-----|------------------|-------------|
| `Person` | Relation | `JiPB` | Contractors (`ed2b9224-97d9-4dff-97f9-82598b61f65d`) | Contractor receiving payment |
| `Contractor Billing` | Relation | `P?Ey` | Contractor Billing (`2c264b29-b84c-80e3-9bd8-c22e7612c1ea`) | Links to billing record with Task Order Log |
| `Invoice Split` | Relation | `\CEP:` | Invoice Split (`2c364b29-b84c-804f-9856-000b58702dea`) | Links to invoice split record |
| `Refund Request` | Relation | `cS>\|` | Refund Request (`2c564b29-b84c-8010-9bd8-000bc1384da3`) | Links to refund request |

### Rollup Properties

| Property | Type | ID | Source | Description |
|----------|------|-----|--------|-------------|
| `Rollup` | Rollup | `Aclb` | - | General rollup |
| `Rollup 1` | Rollup | `_cT^` | From `Refund Request` → `Contractor` | Contractor from refund request |

### Formula Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Source Type` | Formula | `MhMC` | Determines source type based on relations |

#### Source Type Formula

```javascript
if(
  not empty(Contractor Billing), "Contractor Payroll",
  if(not empty(Invoice Split),
    "Commission",
    if(not empty(Refund Request),
      "Refund",
      "Other")
  )
)
```

**Output**:
- `Contractor Payroll` - when Contractor Billing relation is set
- `Commission` - when Invoice Split relation is set
- `Refund` - when Refund Request relation is set
- `Other` - when no source relation is set

## Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│                     Payouts Table                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │ Contractor      │    │ Task Order Log  │                │
│  │ Billing         │───>│ (via relation)  │                │
│  │ (P?Ey)          │    │                 │                │
│  └─────────────────┘    └─────────────────┘                │
│          │                                                  │
│          ▼                                                  │
│  ┌─────────────────┐                                       │
│  │ Payout Record   │                                       │
│  │ - Amount        │                                       │
│  │ - Month         │                                       │
│  │ - Project       │                                       │
│  │ - Status        │                                       │
│  └─────────────────┘                                       │
│          │                                                  │
│          ▼                                                  │
│  ┌─────────────────┐                                       │
│  │ Contractor      │                                       │
│  │ Invoice         │                                       │
│  │ (sum of payouts)│                                       │
│  └─────────────────┘                                       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Sample Data Structure

### Payout Entry Example

```json
{
  "object": "page",
  "id": "2c564b29-b84c-8030-a95c-c67f06651b06",
  "properties": {
    "Payout": {
      "type": "title",
      "title": [
        {
          "text": {
            "content": "Monthly Website Maintenance"
          }
        }
      ]
    },
    "Amount": {
      "type": "number",
      "number": 1250
    },
    "Month": {
      "type": "rich_text",
      "rich_text": [
        {
          "text": {
            "content": "2025-12"
          }
        }
      ]
    },
    "Date": {
      "type": "date",
      "date": {
        "start": "2025-12-08"
      }
    },
    "Paid Date": {
      "type": "date",
      "date": {
        "start": "2025-12-15"
      }
    },
    "Project": {
      "type": "rich_text",
      "rich_text": [
        {
          "text": {
            "content": "Acme Corp"
          }
        }
      ]
    },
    "Status": {
      "type": "status",
      "status": {
        "name": "Pending",
        "color": "yellow"
      }
    },
    "Type": {
      "type": "select",
      "select": {
        "name": "Contractor Payroll",
        "color": "blue"
      }
    },
    "Direction": {
      "type": "select",
      "select": {
        "name": "Outgoing (you pay)",
        "color": "red"
      }
    },
    "Person": {
      "type": "relation",
      "relation": [
        {"id": "0f52bd39-42f8-4a5f-86f6-a816f1af4ced"}
      ]
    },
    "Contractor Billing": {
      "type": "relation",
      "relation": []
    },
    "Notes": {
      "type": "rich_text",
      "rich_text": [
        {
          "text": {
            "content": "December website maintenance and content updates"
          }
        }
      ]
    },
    "Invoice / Ref #": {
      "type": "rich_text",
      "rich_text": [
        {
          "text": {
            "content": "INV-2512"
          }
        }
      ]
    }
  }
}
```

## API Integration

### Query Payouts by Contractor and Month

```javascript
{
  "filter": {
    "and": [
      {
        "property": "Person",
        "relation": {
          "contains": "<contractor_page_id>"
        }
      },
      {
        "property": "Month",
        "rich_text": {
          "equals": "2025-12"
        }
      },
      {
        "property": "Type",
        "select": {
          "equals": "Contractor Payroll"
        }
      }
    ]
  }
}
```

### Query Pending Payouts

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

### Query Payouts with Contractor Billing

```javascript
{
  "filter": {
    "property": "Contractor Billing",
    "relation": {
      "is_not_empty": true
    }
  }
}
```

## Workflow

### Contractor Invoice Generation

1. **Query Payouts**: Filter by Person (contractor) and Month
2. **Sum Amounts**: Aggregate all payout amounts for the period
3. **Get Line Items**: Retrieve details from Contractor Billing → Task Order Log
4. **Generate Invoice**: Create invoice with aggregated data
5. **Update Status**: Mark payouts as Paid after invoice is paid

### Payout Types

| Type | Source | Description |
|------|--------|-------------|
| Contractor Payroll | Contractor Billing | Regular billing from timesheet/task orders |
| Commission | Invoice Split | Commission from client invoices |
| Refund | Refund Request | Refund back to contractor |
| Bonus | Manual | Ad-hoc bonus payments |
| Penalty | Manual | Deductions or penalties |

## Related Databases

- **Contractors** (`ed2b9224-97d9-4dff-97f9-82598b61f65d`): Contractor profiles
- **Contractor Billing** (`2c264b29-b84c-80e3-9bd8-c22e7612c1ea`): Billing records linking to Task Order Log
- **Invoice Split** (`2c364b29-b84c-804f-9856-000b58702dea`): Invoice split records for commissions
- **Refund Request** (`2cc64b29-b84c-8066-adf2-cc56171cedf4`): Refund request records

## Notes

- The `Contractor Billing` relation is the key link to Task Order Log for retrieving timesheet line items
- `Source Type` formula automatically determines the payout category based on which relation is populated
- Payouts are aggregated (`sum()`) to generate contractor invoices
- Month format is `YYYY-MM` (e.g., "2025-12")
