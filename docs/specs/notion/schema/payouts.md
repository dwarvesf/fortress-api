# Contractor Payouts Database Schema

## Overview

- **Database ID**: `2c564b29-b84c-8045-80ee-000bee2e3669`
- **Title**: Contractor Payouts
- **Created**: 2025-12-10
- **Last Edited**: 2026-01-04
- **Icon**: Lock keyhole (gray)
- **URL**: https://www.notion.so/2c564b29b84c806c9d73def95eb7ff30

## Purpose

The Contractor Payouts database tracks all payment transactions for contractors including service fees, refunds, and invoice splits. Each payout record aggregates into contractor invoices.

## Properties

### Core Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Name` | Title | `title` | Payout name/description |
| `Amount` | Number | `qxc^` | Payment amount (formatted with commas) |
| `Date` | Date | `}kwE` | Transaction date |
| `Paid Date` | Date | `U>Mb` | Date payment was processed |
| `Description` | Rich Text | `M?bq` | Additional notes/description |
| `ID` | Unique ID | `v[qI` | Auto-generated unique identifier |

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

### Currency Property

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Currency` | Select | `jofx` | Payment currency |

#### Currency Options

| Currency | Color | Description |
|----------|-------|-------------|
| VND | Green | Vietnamese Dong |
| USD | Blue | US Dollar |

### Relations

| Property | Type | ID | Related Database | Description |
|----------|------|-----|------------------|-------------|
| `Person` | Relation | `JiPB` | Contractors (`ed2b9224-97d9-4dff-97f9-82598b61f65d`) | Contractor receiving payment |
| `00 Task Order` | Relation | `fG\s` | Task Order (`2b964b29-b84c-801e-ab9e-000b0662b987`) | Links to task order record |
| `00 Service Rate` | Relation | `h@y@` | Service Rate (`2c464b29-b84c-80cf-bef6-000b42bce15e`) | Links to service rate record |
| `01 Refund` | Relation | `lBpb` | Refund Request (`2cc64b29-b84c-8065-b90d-000b31aed3e1`) | Links to refund request (dual property) |
| `02 Invoice Split` | Relation | `\EP:` | Invoice Split (`2c364b29-b84c-804f-9856-000b58702dea`) | Links to invoice split record (dual property) |

### Rollup Properties

| Property | Type | ID | Source | Description |
|----------|------|-----|--------|-------------|
| `Discord` | Rollup | `Y[S:` | From `Person` → `Discord` | Discord handle from contractor |
| `00 Contractor` | Rollup | `MuHD` | From `00 Service Rate` → `Contractor` | Contractor from service rate |
| `00 Payday` | Rollup | `p@:o` | From `00 Service Rate` → `Payday` | Payday from service rate |
| `01 Refund Contractor` | Rollup | `_cT^` | From `01 Refund` → `Contractor` | Contractor from refund request |
| `02 Split Person` | Rollup | `Fr@j` | From `02 Invoice Split` → `Person` | Person from invoice split |

### Formula Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Month` | Formula | `^ewn` | Formats date as YYYY-MM |
| `Source Type` | Formula | `MhMC` | Determines source type based on relations |
| `Auto Name` | Formula | `{\|:f` | Generates formatted payout name |

#### Month Formula

```javascript
if(empty(Date),
  /* Return empty if no date exists */
  empty(),
  /* Format as YYYY-MM */
  formatDate(Date, "YYYY") + "-" +
    /* Pad month with leading zero if needed */
    if(month(Date) < 10,
      "0" + format(month(Date)),
      format(month(Date))
    )
)
```

#### Source Type Formula

```javascript
if(
  not empty(00 Task Order), "Service Fee",
  if(not empty(02 Invoice Split),
    "Commission",
    if(not empty(01 Refund),
      "Refund",
      "Extra Payment")
  )
)
```

**Output**:
- `Service Fee` - when 00 Task Order relation is set
- `Commission` - when 02 Invoice Split relation is set
- `Refund` - when 01 Refund relation is set
- `Extra Payment` - when no source relation is set

#### Auto Name Formula

Generates a formatted payout name with structure:
```
PYT :: YYYYMM :: [TYPE :: @discord] :: XXXXX
```

Where:
- `PYT` - Payout prefix
- `YYYYMM` - Month formatted from Date
- `TYPE` - Abbreviated source type (FEE, RFD, SPL)
- `@discord` - Discord handle from Person relation
- `XXXXX` - 5-character obfuscated ID

## Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│                 Contractor Payouts Table                     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │ 00 Task Order   │    │ 00 Service Rate │                │
│  │ (fG\s)          │    │ (h@y@)          │                │
│  └─────────────────┘    └─────────────────┘                │
│          │                      │                           │
│          ▼                      ▼                           │
│  ┌─────────────────────────────────────────┐               │
│  │ Payout Record                           │               │
│  │ - Amount                                │               │
│  │ - Date → Month (formula)                │               │
│  │ - Currency                              │               │
│  │ - Status                                │               │
│  │ - Source Type (formula)                 │               │
│  │ - Auto Name (formula)                   │               │
│  └─────────────────────────────────────────┘               │
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
    "Name": {
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
    "Currency": {
      "type": "select",
      "select": {
        "name": "USD",
        "color": "blue"
      }
    },
    "Status": {
      "type": "status",
      "status": {
        "name": "Pending",
        "color": "yellow"
      }
    },
    "Person": {
      "type": "relation",
      "relation": [
        {"id": "0f52bd39-42f8-4a5f-86f6-a816f1af4ced"}
      ]
    },
    "00 Task Order": {
      "type": "relation",
      "relation": []
    },
    "00 Service Rate": {
      "type": "relation",
      "relation": []
    },
    "Description": {
      "type": "rich_text",
      "rich_text": [
        {
          "text": {
            "content": "December website maintenance and content updates"
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
        "formula": {
          "string": {
            "equals": "2025-12"
          }
        }
      },
      {
        "property": "Source Type",
        "formula": {
          "string": {
            "equals": "Service Fee"
          }
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

### Query Payouts with Task Order

```javascript
{
  "filter": {
    "property": "00 Task Order",
    "relation": {
      "is_not_empty": true
    }
  }
}
```

### Query by Currency

```javascript
{
  "filter": {
    "property": "Currency",
    "select": {
      "equals": "USD"
    }
  }
}
```

## Workflow

### Contractor Invoice Generation

1. **Query Payouts**: Filter by Person (contractor) and Month formula
2. **Sum Amounts**: Aggregate all payout amounts for the period (by Currency)
3. **Get Line Items**: Retrieve details from 00 Task Order or 00 Service Rate
4. **Generate Invoice**: Create invoice with aggregated data
5. **Update Status**: Mark payouts as Paid after invoice is paid

### Source Types

| Source Type | Source Relation | Description |
|-------------|-----------------|-------------|
| Service Fee | 00 Task Order | Regular service fee from task orders |
| Commission | 02 Invoice Split | Commission from client invoices |
| Refund | 01 Refund | Refund back to contractor |
| Extra Payment | None | Manual entries without linked source |

## Related Databases

- **Contractors** (`ed2b9224-97d9-4dff-97f9-82598b61f65d`): Contractor profiles
- **Task Order** (`2b964b29-b84c-801e-ab9e-000b0662b987`): Task order records
- **Service Rate** (`2c464b29-b84c-80cf-bef6-000b42bce15e`): Service rate records with payday info
- **Invoice Split** (`2c364b29-b84c-804f-9856-000b58702dea`): Invoice split records for commissions
- **Refund Request** (`2cc64b29-b84c-8065-b90d-000b31aed3e1`): Refund request records

## Notes

- `Month` is now a **formula** that auto-calculates from `Date` field (YYYY-MM format)
- `Source Type` formula automatically determines the payout category based on which relation is populated
- `Auto Name` formula generates a standardized payout identifier with obfuscated ID
- `Currency` property supports VND and USD
- Payouts are aggregated (`sum()`) to generate contractor invoices
- The `00 Task Order` relation replaces the old `Contractor Billing` relation
- The `00 Service Rate` relation provides contractor and payday information via rollups
