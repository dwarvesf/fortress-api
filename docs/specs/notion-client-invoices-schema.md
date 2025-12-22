# Notion Client Invoices Database Schema

**Database ID:** `2bf64b29b84c80879a52ed2f9d493096`
**Database Name:** Client Invoices
**Last Updated:** 2025-12-22
**URL:** https://www.notion.so/2bf64b29b84c80879a52ed2f9d493096

## Overview

The Client Invoices database manages invoice generation, tracking, and commission calculations for client projects. It uses a hierarchical model with parent Invoice records and child Line Item records.

## Record Types

### Invoice (Type = "Invoice")
Parent records representing complete invoices sent to clients.

### Line Item (Type = "Line Item")
Child records representing individual billable items within an invoice.

## Database Statistics

- **Total Records:** 25 (as of 2025-12-22)
- **Active Invoices:** 11
- **Line Items:** 14
- **Properties:** 63

## Core Properties

### Identification

| Property | Type | Description |
|----------|------|-------------|
| `(auto) Invoice Number` | title | Auto-generated format: `INV-YYYYMM-[PROJECT]-[ID]` |
| `ID` | unique_id | Sequential unique identifier |
| `Type` | select | "Invoice" or "Line Item" |
| `Auto Name` | formula | Context-aware display name for line items |

### Dates

| Property | Type | Description |
|----------|------|-------------|
| `Issue Date` | date | Date invoice was issued |
| `Due Date` | formula | Calculated as Issue Date + 7 days |
| `Paid Date` | date | Date payment was received |

### Status & Workflow

| Property | Type | Description |
|----------|------|-------------|
| `Status` | status | Draft / Sent / Overdue / Paid / Cancelled |
| `Generate Invoice` | button | Webhook trigger to generate invoice PDF |
| `Preview Invoice` | button | Webhook trigger to preview invoice |
| `⚠️ Send invoice` | button | Webhook trigger to send invoice via email |
| `Sent by` | rich_text | User who sent the invoice |

### Financial Core

| Property | Type | Description |
|----------|------|-------------|
| `Currency` | select | USD / VND / EUR / GBP / JPY / CNY / SGD / CAD / AUD |
| `Billing Type` | select | Resource / Milestone / Mixed |
| `Subtotal` | formula | Base amount before discounts/tax |
| `Tax Rate` | number | Tax percentage (e.g., 0.1 = 10%) |
| `Final Total` | formula | `(Subtotal - Discount) * (1 + Tax Rate)` |

### Line Item Details

| Property | Type | Description |
|----------|------|-------------|
| `Quantity` | number | Number of units |
| `Unit Price` | number | Price per unit |
| `Fixed Amount` | number | Alternative to unit pricing |
| `Line Total` | formula | Calculated line total with discount/tax |
| `Description` | rich_text | Line item description |

### Discount Management

| Property | Type | Description |
|----------|------|-------------|
| `Discount Type` | select | None / Percentage / Fixed Amount / Bulk Discount / Seasonal / Loyalty / Early Payment |
| `Discount Value` | number | Percentage or fixed amount |
| `Discount Amount` | formula | Calculated discount amount |
| `Discount Display` | formula | Human-readable discount (e.g., "10%" or "$500") |

### Commission Tracking

| Property | Type | Description |
|----------|------|-------------|
| `% Sales` | number | Sales commission percentage |
| `% Account Mgr` | number | Account manager commission percentage |
| `% Delivery Lead` | number | Delivery lead commission percentage |
| `% Hiring Referral` | number | Hiring referral commission percentage |
| `Sales Amount` | formula | Calculated sales commission |
| `Account Amount` | formula | Calculated account manager commission |
| `Delivery Lead Amount` | formula | Calculated delivery lead commission |
| `Hiring Referral Amount` | formula | Calculated hiring referral commission |
| `Total Commission Paid` | formula | Sum of all commissions (Invoice level) |

### Commission Rollups (Invoice Level)

| Property | Type | Description |
|----------|------|-------------|
| `All Sales Amounts` | rollup | Sum of sales commissions from line items |
| `All AM Amounts` | rollup | Sum of account manager commissions |
| `All DL Amounts` | rollup | Sum of delivery lead commissions |
| `All Hiring Ref Amounts` | rollup | Sum of hiring referral commissions |

### Relations

| Property | Type | Target Database | Description |
|----------|------|-----------------|-------------|
| `Project` | relation | Projects | Associated project |
| `Deployment Tracker` | relation | Deployment Tracker | Resource deployment details |
| `Line Item` | relation | Client Invoices (self) | Child line items (from Invoice) |
| `Parent item` | relation | Client Invoices (self) | Parent invoice (from Line Item) |
| `Splits` | relation | Commission Splits | Commission split records |
| `Google Drive File` | relation | Google Drive | Generated invoice PDFs |

### Rollup Properties

| Property | Type | Source | Description |
|----------|------|--------|-------------|
| `Client` | rollup | Project → Client | Client name |
| `Code` | rollup | Project → Codename | Project code |
| `Redacted Codename` | rollup | Project → Redacted Code | Redacted project name |
| `Recipients` | rollup | Project → Recipient Emails | Invoice recipient emails |
| `Resource` | rollup | Deployment Tracker → Contractor | Contractor name |
| `Contractor` | rollup | Deployment Tracker → Contractor | Contractor profile |
| `Position` | rollup | Deployment Tracker → Position | Contractor position |
| `Account Manager` | formula | Deployment Tracker → Account Manager | Account manager name |
| `Sales Person` | rollup | Deployment Tracker → Final Sales Credit | Sales person |
| `Delivery Lead` | formula | Deployment Tracker → Delivery Lead | Delivery lead |
| `Hiring Referral` | formula | Deployment Tracker → Hiring Referral | Hiring referral person |
| `Total Amount` | rollup | Line Item → Line Total | Sum of line item totals |

### Payment & Banking

| Property | Type | Description |
|----------|------|-------------|
| `Payment Method` | select | Bank Transfer / Credit Card / Cash / Check / PayPal / Venmo / Crypto / Mobile Payment |
| `Bank Account` | relation | Bank accounts database |
| `Bank Account Details` | rollup | Bank account information |

### Metadata

| Property | Type | Description |
|----------|------|-------------|
| `Notes` | rich_text | Internal notes |
| `Attachment` | files | File attachments |
| `Splits Generated` | checkbox | Whether commission splits have been generated |
| `Generate Splits` | button | Webhook trigger to generate commission splits |

## Key Formula Logic

### Auto Name Formula

Generates context-aware names for invoices and line items:

**Invoice Format:** `INV-YYYYMM-[PROJECT_CODE]-[OBFUSCATED_ID]`

**Line Item Format:** `[Project] :: [Contractor] / [Month Year]`

Uses obfuscated ID generation for invoice numbers:
```
realID = (ID * 2654435761) % 4294967295
obfuscatedID = base32-encoded realID
```

### Discount Amount Calculation

```javascript
if (Discount Type == "Percentage") {
  base * (Discount Value / 100)
} else if (Discount Type == "Fixed Amount") {
  Discount Value
} else {
  0
}
```

### Final Total Calculation

**Invoice Level:**
```javascript
preDiscount = sum(Line Item totals)
invoiceDiscount = Discount Amount
finalNumber = (preDiscount - invoiceDiscount) * (Tax Rate + 1)
round(finalNumber * 100) / 100
```

**Line Item Level:**
```javascript
base = Subtotal (Quantity * Unit Price or Fixed Amount)
discountAmount = calculated discount
subtotalAfterDiscount = base - discountAmount
Line Total = subtotalAfterDiscount * (1 + Tax Rate)
```

### Commission Splits

All commission calculations handle multiple recipients by dividing the base amount:

```javascript
baseAmount = Line Total * Commission Percentage
count = number of recipients
perPerson = baseAmount / count
round(perPerson * 100) / 100
```

## Workflow Integration

### Webhook Buttons

The database includes three webhook buttons that trigger external processes:

1. **Generate Invoice** - Creates PDF invoice in Google Drive
2. **Preview Invoice** - Generates preview without saving
3. **⚠️ Send invoice** - Sends invoice via email to recipients

### Commission Splits Workflow

1. User clicks "Generate Splits" button
2. System creates records in Commission Splits database
3. Each split links back to invoice via "Splits" relation
4. "Splits Generated" checkbox is marked true

## Active Invoices (December 2025)

| Invoice Number | Project | Issue Date | Total | Currency | Billing Type | Status |
|----------------|---------|------------|-------|----------|--------------|--------|
| INV-202512-NGHENHAN- | nghenhan.trade | 2025-12-03 | 8,000 | USD | Resource | Draft |
| INV-202512-ASCENDA- | Ascenda | 2025-12-04 | 19,550 | USD | Resource | Draft |
| INV-202512-MUDAH- | Mudah | 2025-12-02 | 8,000 | USD | Resource | Draft |
| INV-202512-PLOT- | Plot | 2025-12-15 | 8,982 | USD | Mixed | Draft |
| INV-202512-DAIICHI- | Daiichi | 2025-12-16 | 121,500,000 | VND | Milestone | Draft |

## API Integration Considerations

### Reading Invoices

To query invoices with full details:

```javascript
{
  "database_id": "2bf64b29b84c80879a52ed2f9d493096",
  "filter": {
    "property": "Type",
    "select": {
      "equals": "Invoice"
    }
  }
}
```

### Key Properties for API

When working with this database via API:

**Required for Invoice Creation:**
- `Type` = "Invoice"
- `Project` (relation)
- `Issue Date`
- `Currency`
- `Billing Type`

**Auto-calculated (read-only):**
- `(auto) Invoice Number`
- `Due Date`
- `Final Total`
- All commission amounts
- All rollup properties

### Property ID Mapping

Critical property IDs for API access:

| Property | ID |
|----------|-----|
| Type | `=DBN:` |
| Issue Date | `U?Nt` |
| Status | `nsb@` |
| Final Total | `:CEa` |
| Currency | `vVG<` |
| Project | `srY=` |
| Deployment Tracker | `OMj?` |

## Notes

- Invoice numbers use obfuscated IDs to prevent sequential guessing
- Due dates automatically calculated as Issue Date + 7 days
- Commission splits support multiple recipients with automatic division
- Tax calculations apply after discounts
- Line items inherit tax rate from parent invoice
- Empty invoices (N/A invoice numbers) are likely templates or incomplete drafts
