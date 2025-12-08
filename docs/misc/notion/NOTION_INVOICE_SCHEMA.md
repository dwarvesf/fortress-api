# Notion Invoice Database Schema

**Database ID:** `2bf64b29-b84c-8087-9a52-ed2f9d493096`
**URL:** https://www.notion.so/2bf64b29b84c80879a52ed2f9d493096

## Data Model

Self-referential table: both Invoices and Line Items in same database.

```
Invoice (Type="Invoice")
├── Line Item 1 (Type="Line Item", Parent item → Invoice)
├── Line Item 2 (Type="Line Item", Parent item → Invoice)
└── Line Item N (Type="Line Item", Parent item → Invoice)
```

## Properties

### Identity

| Property | ID | Type | Notes |
|----------|-----|------|-------|
| Invoice Number | `title` | title | Auto: `INV-YYYYMM-CODE-XXXX` |
| ID | `wX_b` | unique_id | Auto-increment |
| Type | `=BN:` | select | `Invoice`, `Line Item` |
| Auto Name | `IV{M` | formula | Display name |

### Status & Dates

| Property | ID | Type | Options |
|----------|-----|------|---------|
| Status | `nsb@` | status | Draft, Sent, Overdue, Paid, Cancelled |
| Issue Date | `U?Nt` | date | |
| Due Date | `GsHF` | date | |
| Paid Date | `wXqV` | date | |

### Financial - Line Item Level

| Property | ID | Type | Format |
|----------|-----|------|--------|
| Quantity | `ks~>` | number | number |
| Unit Price | `eQ^O` | number | number |
| Tax Rate | `}zed` | number | percent |
| Currency | `vVG<` | select | USD, EUR, GBP, JPY, CNY, SGD, CAD, AUD |

### Discounts

| Property | ID | Type | Options |
|----------|-----|------|---------|
| Discount Type | `]ZC:` | select | None, Percentage, Fixed Amount, Bulk Discount, Seasonal, Loyalty, Early Payment |
| Discount Value | `UE`Y` | number | |
| Discount Amount | `<TIG` | formula | Calculated |
| Discount Display | `cQRk` | formula | "10%" or "$100" |

### Commission Percentages

| Property | ID | Type |
|----------|-----|------|
| % Sales | `gZX~` | number (percent) |
| % Account Mgr | `iF]]` | number (percent) |
| % Delivery Lead | `r@c{` | number (percent) |
| % Hiring Referral | `uyk[` | number (percent) |

### Text Fields

| Property | ID | Type |
|----------|-----|------|
| Description | `_;tA` | rich_text |
| Role | `xbaN` | rich_text |
| Notes | `vL<E` | rich_text |
| Sent by | `<cgl` | rich_text |

### Payment

| Property | ID | Type | Options |
|----------|-----|------|---------|
| Payment Method | `kDRp` | select | Bank Transfer, Credit Card, Cash, Check, PayPal, Venmo, Crypto, Mobile Payment |

### Relations

| Property | ID | Type | Target DB |
|----------|-----|------|-----------|
| Project | `srY=` | relation | `0ddadba5-bbf2-440c-a286-9f607eca88db` |
| Parent item | `ES]D` | relation | Self (dual with Line Item) |
| Line Item | `Z_^m` | relation | Self (dual with Parent item) |
| Google Drive File | (connection) | relation | `2bf64b29-b84c-80e2-8cc7-000bfe534203` |
| Bank Account | `AAUf` | relation | Bank Accounts DB |
| Deployment Tracker | `OMj?` | relation | Deployments DB |
| Splits | `lV[i` | relation | Splits DB |

### Rollups

| Property | ID | Relation | Source Property | Function |
|----------|-----|----------|-----------------|----------|
| Total Amount | `=f;t` | Line Item | Line Total | sum |
| Code | `<D>S` | Project | Codename | show_original |
| Client | `sYY]` | Project | Client | show_original |
| Recipients | `OLc@` | Project | Recipient Emails | show_original |
| Redacted Codename | `psPe` | Project | Redacted Code | show_original |
| Bank Account Details | `Ag|z` | Bank Account | Details | show_original |
| Contractor | `v:KN` | Deployment Tracker | Contractor | show_original |
| Position | `V?gQ` | Deployment Tracker | Position | show_original |
| Sales Person | `@[Qm` | Project | Sales Person | show_original |
| All Sales Amounts | `EARk` | Line Item | Sales Amount | sum |
| All AM Amounts | `Mu]]` | Line Item | Account Amount | sum |
| All DL Amounts | `~fv[` | Line Item | Delivery Lead Amount | sum |
| All Hiring Ref Amounts | `GuqF` | Line Item | Hiring Referral Amount | sum |

### Formulas

| Property | ID | Description |
|----------|-----|-------------|
| Line Total | `Cw|k` | `Qty × Price × (1 + Tax) - Discount` (Line Item only) |
| Subtotal | `hDZr` | `Qty × Price` (Line Item only) |
| Final Total | `:<Ea` | `Total Amount - Discount` with currency symbol |
| Sales Amount | `\LY?` | `Line Total × % Sales / count(sales)` |
| Account Amount | `zZE:` | `Line Total × % AM / count(AMs)` |
| Delivery Lead Amount | `[}gB` | `Line Total × % DL / count(leads)` |
| Hiring Referral Amount | `XM:A` | `Line Total × % Hiring Referral` |
| Total Commission Paid | `FvNv` | Sum of all commission amounts |
| Account Manager | `>Ow^` | From Deployment Tracker |
| Delivery Lead | `|Rjc` | From Deployment Tracker |
| Hiring Referral | `nYwg` | From Deployment Tracker |

### Other

| Property | ID | Type |
|----------|-----|------|
| Splits Generated | `}WM}` | checkbox |
| Generate Splits | `kS;X` | button |

## Status Flow

```
Draft ──► Sent ──► Paid
            │
            ├──► Overdue
            │
            └──► Cancelled
```

## Query Examples

### Get Invoices Only (not Line Items)

```go
filter := &notion.DatabaseQueryFilter{
    Property: "Type",
    DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
        Select: &notion.SelectDatabaseQueryFilter{
            Equals: "Invoice",
        },
    },
}
```

### Get Line Items for Invoice

```go
filter := &notion.DatabaseQueryFilter{
    Property: "Parent item",
    DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
        Relation: &notion.RelationDatabaseQueryFilter{
            Contains: invoicePageID,
        },
    },
}
```

### Get Paid Invoices

```go
filter := &notion.DatabaseQueryFilter{
    And: []notion.DatabaseQueryFilter{
        {Property: "Type", DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
            Select: &notion.SelectDatabaseQueryFilter{Equals: "Invoice"},
        }},
        {Property: "Status", DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
            Status: &notion.StatusDatabaseQueryFilter{Equals: "Paid"},
        }},
    },
}
```
