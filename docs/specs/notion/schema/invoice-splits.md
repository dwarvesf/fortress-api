# Invoice Splits Database Schema

## Overview

- **Database ID**: `2c364b29-b84c-804f-9856-000b58702dea`
- **Title**: Invoice Splits
- **Created**: 2025-12-08
- **Last Edited**: 2025-12-26
- **Icon**: Lock keyhole (gray)
- **URL**: https://www.notion.so/2c364b29b84c80498a8df7befd22f7fc

## Purpose

The Invoice Splits database tracks commission payments and other revenue splits from client invoices. Each split record represents a portion of an invoice allocated to a contractor based on their role (Sales, Account Manager, Delivery Lead, etc.).

## Properties

### Core Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Name` | Title | `title` | Split entry name/description |
| `Amount` | Number | `\CdP` | Split amount in dollars |
| `Month` | Date | `=Puv` | Month of the split |
| `Invoice` | Rich Text | `@V\`g` | Invoice reference number |
| `Person` | Rich Text | `TdYN` | Person name receiving the split |
| `Auto Name` | Rich Text | `Qp@c` | Auto-generated name |
| `Notes` | Rich Text | `HVvA` | Additional notes |

### System Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `ID` | Unique ID | `:^Ua` | Auto-incrementing unique identifier |

### Status Property

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Status` | Select | `HBp:` | Current payment status |

#### Status Options

| Status | Color | Description |
|--------|-------|-------------|
| Pending | Yellow | Split awaiting payment |
| Paid | Green | Split paid out |
| Cancelled | Red | Split cancelled |

### Currency Property

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Currency` | Select | `YDMP` | Payment currency |

#### Currency Options

| Currency | Color |
|----------|-------|
| USD | Gray |
| VND | Yellow |
| GBP | Purple |
| SGD | Blue |

### Type Property

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Type` | Select | `KxqL` | Type of split |

#### Type Options

| Type | Color | Description |
|------|-------|-------------|
| Commission | Green | Commission-based payment |
| Reimbursement | Blue | Expense reimbursement |
| Bonus | Purple | Bonus payment |
| Salary | Orange | Salary component |
| Tax | Red | Tax-related |
| Fee | Yellow | Service fee |
| Other | Gray | Other type |

### Role Property

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Role` | Select | `OQz<` | Role receiving the split |

#### Role Options

| Role | Color | Description |
|------|-------|-------------|
| Sales | Green | Sales commission |
| Account Manager | Blue | Account management commission |
| Delivery Lead | Purple | Delivery lead commission |
| Hiring Referral | Yellow | Hiring referral bonus |
| Bonus | Orange | General bonus |
| Fee | Gray | Fee allocation |
| Penalty | Red | Penalty deduction |

### Relations

| Property | Type | ID | Related Database | Description |
|----------|------|-----|------------------|-------------|
| `Contractor` | Relation | `m\pE` | Contractors (`9d468753-ebb4-4977-a8dc-156428398a6b`) | Contractor receiving the split |
| `Deployment` | Relation | `fcIT` | Deployments (`2b864b29-b84c-8079-9568-dc17685f4f33`) | Related deployment |
| `Invoice Item` | Relation | `VywX` | Client Invoices (`2bf64b29-b84c-80e2-8cc7-000bfe534203`) | Linked invoice line item (dual property syncs to `Splits`) |
| `Client Invoices` | Relation | `z=>S` | Client Invoices (`2bf64b29-b84c-80e2-8cc7-000bfe534203`) | Parent invoice record |

### Rollup Properties

| Property | Type | ID | Source | Description |
|----------|------|-----|--------|-------------|
| `Project` | Rollup | `:KYY` | From `Deployment` → `Project` | Project name |
| `Charge` | Rollup | `mcyq` | From `Deployment` → `Contractor` | Contractor from deployment (charge person) |

## Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│                   Invoice Splits Table                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐         ┌─────────────────┐           │
│  │ Invoice Item    │         │ Deployment      │           │
│  │ (VywX)          │         │ (fcIT)          │           │
│  └────────┬────────┘         └────────┬────────┘           │
│           │                           │                     │
│           ▼                           ▼                     │
│  ┌─────────────────────────────────────────────┐           │
│  │ Invoice Split Record                         │           │
│  │ - Amount                                     │           │
│  │ - Currency                                   │           │
│  │ - Role (Sales, Account Manager, etc.)        │           │
│  │ - Type (Commission, Bonus, etc.)             │           │
│  │ - Status (Pending, Paid, Cancelled)          │           │
│  └────────┬────────────────────────────────────┘           │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐                                       │
│  │ Contractor      │                                       │
│  │ (i<ks)          │                                       │
│  └────────┬────────┘                                       │
│           │                                                 │
│           ▼                                                 │
│  ┌─────────────────┐                                       │
│  │ Payouts Table   │                                       │
│  │ (aggregation)   │                                       │
│  └─────────────────┘                                       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Sample Data Structure

### Invoice Split Entry Example

```json
{
  "object": "page",
  "id": "2c364b29-b84c-8030-a95c-example",
  "properties": {
    "Name": {
      "type": "title",
      "title": [
        {
          "text": {
            "content": "Sales Commission - Acme Corp Dec 2025"
          }
        }
      ]
    },
    "Amount": {
      "type": "number",
      "number": 500
    },
    "Currency": {
      "type": "select",
      "select": {
        "name": "USD",
        "color": "gray"
      }
    },
    "Month": {
      "type": "date",
      "date": {
        "start": "2025-12-01"
      }
    },
    "Invoice": {
      "type": "rich_text",
      "rich_text": [
        {
          "text": {
            "content": "INV-2025-001"
          }
        }
      ]
    },
    "Status": {
      "type": "select",
      "select": {
        "name": "Pending",
        "color": "yellow"
      }
    },
    "Type": {
      "type": "select",
      "select": {
        "name": "Commission",
        "color": "green"
      }
    },
    "Role": {
      "type": "select",
      "select": {
        "name": "Sales",
        "color": "green"
      }
    },
    "Contractor": {
      "type": "relation",
      "relation": [
        {"id": "contractor-page-id"}
      ]
    },
    "Deployment": {
      "type": "relation",
      "relation": [
        {"id": "deployment-page-id"}
      ]
    }
  }
}
```

## API Integration

### Query Invoice Splits by Contractor

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

### Query Pending Splits

```javascript
{
  "filter": {
    "property": "Status",
    "select": {
      "equals": "Pending"
    }
  }
}
```

### Query Splits by Month and Type

```javascript
{
  "filter": {
    "and": [
      {
        "property": "Month",
        "date": {
          "equals": "2025-12-01"
        }
      },
      {
        "property": "Type",
        "select": {
          "equals": "Commission"
        }
      }
    ]
  }
}
```

## Workflow

### Commission Split Creation

1. **Invoice Creation**: Client invoice created with line items
2. **Split Calculation**: Calculate commission splits based on roles
3. **Split Creation**: Create Invoice Split entries for each recipient
4. **Payout Linkage**: Link splits to Payouts table for aggregation
5. **Payment Processing**: Mark splits as Paid when processed

### Role-Based Commission Structure

| Role | Typical Use Case |
|------|------------------|
| Sales | Initial client acquisition commission |
| Account Manager | Ongoing account management commission |
| Delivery Lead | Project delivery management commission |
| Hiring Referral | Bonus for successful referrals |

## Related Databases

- **Contractors** (`9d468753-ebb4-4977-a8dc-156428398a6b`): Recipients of splits
- **Deployments** (`2b864b29-b84c-8079-9568-dc17685f4f33`): Related project deployments
- **Client Invoices** (`2bf64b29-b84c-80e2-8cc7-000bfe534203`): Source invoice/line items
- **Contractor Payouts**: Aggregated payout entries (env: NOTION_CONTRACTOR_PAYOUTS_DB_ID)

## Notes

- Invoice Splits are linked to Payouts via the `Invoice Split` relation in Payouts table
- The `Type` property distinguishes between Commission, Bonus, Reimbursement, etc.
- The `Role` property determines the commission structure (Sales, Account Manager, etc.)
- Splits can be in different currencies (USD, VND, GBP, SGD)
- The dual property relation with Invoice Items enables bidirectional sync
