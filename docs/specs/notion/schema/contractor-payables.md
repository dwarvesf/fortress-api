# Contractor Payables Database Schema

## Overview

- **Database ID**: `2c264b29-b84c-8037-807c-000bf6d0792c`
- **Title**: Contractor Payables
- **Created**: 2025-12-07T18:47:00.000Z
- **Last Edited**: 2026-01-15T03:55:00.000Z
- **Icon**: 💸
- **URL**: https://www.notion.so/2c264b29b84c80e39bd8c22e7612c1ea
- **Parent Database**: `2c264b29-b84c-80e3-9bd8-c22e7612c1ea`

## Purpose

The Contractor Payables database aggregates payout items into payable records for each contractor. It represents the sum of all payouts for a given period.

### Workflow
1. Sum payout amounts for a contractor
2. Use Discord `/gen invoice` command to generate invoice file
3. System returns link to invoice for preview
4. Send invoice via email service to contractor

## Properties

### Core Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Payable` | Title | `title` | Main payable name/description |
| `Total` | Number | `` `GxT`` | Total payable amount (formatted with commas) |
| `Period` | Date | `>HDF` | Month this payable covers |
| `Invoice Date` | Date | `QD]A` | Date of invoice |
| `Payment Date` | Date | `{yf?` | Date payment was processed |
| `Invoice ID` | Rich Text | `{ShZ` | Invoice number submitted by contractor |
| `Notes` | Rich Text | `jyfE` | Additional notes |
| `ID` | Unique ID | `IW]b` | Auto-generated unique identifier |
| `Attachments` | Files | `^LN`` | Attached invoice files |
| `Exchange Rate` | Number | `KnQx` | Exchange rate for currency conversion |
| `Created time` | Created Time | `nf{Z` | Timestamp when payable record was created |

### Payment Status Property

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Payment Status` | Status | `BkF\` | Current payment status |

#### Status Options

| Status | Color | Group | Description |
|--------|-------|-------|-------------|
| New | Blue | To-do | Newly created payable |
| Pending | Yellow | In progress | Payment awaiting processing |
| Paid | Green | Complete | Payment completed |
| Cancelled | Red | Complete | Payment cancelled |

### Currency Property

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Currency` | Select | `EJ|p` | Payment currency |

#### Currency Options

| Currency | Color | Description |
|----------|-------|-------------|
| USD | Blue | US Dollar |
| VND | Green | Vietnamese Dong |

### Contractor Type Property

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Contractor Type` | Select | `oE_j` | Legal entity type of contractor |

#### Contractor Type Options

| Type | Color |
|------|-------|
| Individual | Blue |
| Sole Proprietor | Green |
| LLC | Purple |
| Corporation | Red |
| Partnership | Orange |
| Non-Profit | Pink |
| Government Entity | Gray |

### Relations

| Property | Type | ID | Related Database | Description |
|----------|------|-----|------------------|-------------|
| `Contractor` | Relation | `ZVms` | Contractors (DB: `9d468753-ebb4-4977-a8dc-156428398a6b`, DS: `ed2b9224-97d9-4dff-97f9-82598b61f65d`) | The contractor this payable belongs to |
| `Payout Items` | Relation | `sMWI` | Contractor Payouts (DB: `2c564b29-b84c-806c-9d73-def95eb7ff30`, DS: `2c564b29-b84c-8045-80ee-000bee2e3669`) | Individual payout line items aggregated into this payable |

### Rollup Properties

| Property | Type | ID | Source | Description |
|----------|------|-----|--------|-------------|
| `Discord` | Rollup | `x]OV` | From `Contractor` → `Discord` | Discord handle from contractor |

### Formula Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Auto Name` | Formula | `_l\V` | Generates formatted payable name |

#### Auto Name Formula

Generates a formatted payable name with structure:
```
PAY :: [discord] :: YYYYMM :: [Status] :: Currency :: XXXX
```

Where:
- `PAY` - Payable prefix
- `[discord]` - Discord handle from Contractor relation
- `YYYYMM` - Period formatted from Invoice Date or Period
- `[Status]` - Payment status for quick reference
- `Currency` - USD or VND
- `XXXX` - 4-character obfuscated ID

```javascript
lets(
  /* Get contractor information */
  contractor, Contractor.first(),
  discord, if(not empty(contractor), contractor.Discord, ""),

  /* Format the date */
  invoiceDate, if(not empty(Invoice Date), Invoice Date,
               if(not empty(Period), Period, now())),
  monthYear, formatDate(invoiceDate, "YYYYMM"),

  /* Generate pseudo-random ID for security */
  chars, split("ABCDEFGHJKLMNPQRSTUVWXYZ23456789", ""),
  realID, toNumber(ID) * 2654435761 % 4294967295,
  obfuscatedID, chars.at(realID % 32) + chars.at(floor(realID / 32) % 32) +
               chars.at(floor(realID / 1024) % 32) + chars.at(floor(realID / 32768) % 32),

  /* Create the formatted name */
  style(
    "PAY :: " +
    if(empty(discord), "", "[" + discord + "] :: ") +
    monthYear + " :: " +
    if(not empty(Payment Status), "[" + Payment Status + "] :: ", "") +
    if(not empty(Currency), Currency + " :: ", "") +
    style(obfuscatedID, "red"),
    "b"
  )
)
```

## Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│                 Contractor Payables Table                    │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────┐                                        │
│  │ Contractor      │ (relation)                             │
│  │ (ZVms)          │────────────────┐                       │
│  └─────────────────┘                │                       │
│                                     ▼                       │
│  ┌─────────────────────────────────────────┐               │
│  │ Payable Record                          │               │
│  │ - Total (sum of payout items)           │               │
│  │ - Period → Month (YYYY-MM)              │               │
│  │ - Currency                              │               │
│  │ - Payment Status                        │               │
│  │ - Auto Name (formula)                   │               │
│  │ - Contractor Type                       │               │
│  └─────────────────────────────────────────┘               │
│          │                                                  │
│          │ aggregates                                       │
│          ▼                                                  │
│  ┌─────────────────┐                                        │
│  │ Payout Items    │ (relation)                             │
│  │ (sMWI)          │                                        │
│  │ - Service Fee   │                                        │
│  │ - Commission    │                                        │
│  │ - Refund        │                                        │
│  └─────────────────┘                                        │
│          │                                                  │
│          ▼                                                  │
│  ┌─────────────────┐                                        │
│  │ Contractor      │                                        │
│  │ Invoice (PDF)   │                                        │
│  └─────────────────┘                                        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Sample Data Structure

### Payable Entry Example

```json
{
  "object": "page",
  "id": "2c264b29-b84c-8030-a95c-example",
  "properties": {
    "Payable": {
      "type": "title",
      "title": [
        {
          "text": {
            "content": "PAY :: [@john.doe] :: 202512 :: [Pending] :: USD :: A1B2"
          }
        }
      ]
    },
    "Total": {
      "type": "number",
      "number": 2500.00
    },
    "Period": {
      "type": "date",
      "date": {
        "start": "2025-12-01",
        "end": "2025-12-31"
      }
    },
    "Invoice Date": {
      "type": "date",
      "date": {
        "start": "2025-12-28"
      }
    },
    "Currency": {
      "type": "select",
      "select": {
        "name": "USD",
        "color": "blue"
      }
    },
    "Payment Status": {
      "type": "status",
      "status": {
        "name": "Pending",
        "color": "yellow"
      }
    },
    "Contractor": {
      "type": "relation",
      "relation": [
        {"id": "contractor-page-id-here"}
      ]
    },
    "Payout Items": {
      "type": "relation",
      "relation": [
        {"id": "payout-1-id"},
        {"id": "payout-2-id"},
        {"id": "payout-3-id"}
      ]
    },
    "Contractor Type": {
      "type": "select",
      "select": {
        "name": "Individual",
        "color": "blue"
      }
    },
    "Invoice ID": {
      "type": "rich_text",
      "rich_text": [
        {
          "text": {
            "content": "CONTR-202512-A1B2"
          }
        }
      ]
    }
  }
}
```

## API Integration

### Query Payables by Contractor

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

### Query Pending Payables

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

### Query Payables by Period

```javascript
{
  "filter": {
    "property": "Period",
    "date": {
      "equals": "2025-12-01"
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

### Invoice Generation Flow

1. **Aggregate Payouts**: Sum all `Payout Items` for a contractor and period
2. **Create Payable**: Generate payable record with total and metadata
3. **Generate Invoice**: Use `/gen invoice` Discord command to create PDF
4. **Upload Attachment**: Store generated invoice PDF in `Attachments`
5. **Send Invoice**: Email invoice to contractor
6. **Update Status**: Mark as Paid after payment received

### Status Transitions

```
New → Pending → Paid
         ↓
     Cancelled
```

## Related Databases

- **Contractors**
  - Database ID: `9d468753-ebb4-4977-a8dc-156428398a6b`
  - Data Source ID: `ed2b9224-97d9-4dff-97f9-82598b61f65d`
  - Description: Contractor profiles with Discord handles and payment information

- **Contractor Payouts**
  - Database ID: `2c564b29-b84c-806c-9d73-def95eb7ff30`
  - Data Source ID: `2c564b29-b84c-8045-80ee-000bee2e3669`
  - Description: Individual payout line items (Service Fee, Commission, Refund, Other)

## Notes

- `Total` represents the sum of all linked `Payout Items`
- `Period` defines the billing month for the payable
- `Auto Name` formula generates a standardized identifier for easy reference
- `Contractor Type` helps with tax and compliance requirements
- `Payout Items` relation links to individual payout entries from Contractor Payouts database
- Discord `/gen invoice` command uses this database to generate contractor invoices
