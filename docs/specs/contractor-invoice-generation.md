# Contractor Invoice Generation System

## Status: Documented

## Overview

The Contractor Invoice Generation system is responsible for creating professional PDF invoices for independent contractors based on their work data in Notion. It integrates multiple services including Notion (for rates and payouts), Wise (for currency conversion), and Redis (for caching).

## Technical Architecture

### Core Data Models

The system relies on two primary data structures defined in `pkg/controller/invoice/contractor_invoice.go`:

#### `ContractorInvoiceData`
Represents the entire invoice, including metadata and calculated totals.
- `InvoiceNumber`: Unique identifier (e.g., `INVC-202512-QUANG-A1B2`).
- `TotalUSD`: The final amount to be paid, normalized to USD.
- `LineItems`: A slice of `ContractorInvoiceLineItem` objects.
- `ExchangeRate`: Captured at the time of generation for display purposes.
- `IsExtraPaymentOnly`: Flag indicating if the invoice contains only Commission and Extra Payment items.

#### `ContractorInvoiceLineItem`
Represents an individual charge or credit on the invoice.
- `Title`: Section header (e.g., "Software Development Services Rendered").
- `Description`: Detailed proof of work or task description.
- `AmountUSD`: The line item total in USD.
- `OriginalAmount` / `OriginalCurrency`: Preserves the source data (e.g., 45,000,000 VND).
- `IsHourlyRate`: Flag used to trigger hour-based aggregation logic.

---

## Core Workflow

### 1. Data Retrieval (Parallel & Concurrent)

The system maximizes efficiency by fetching data concurrently using `sync.WaitGroup`.

- **Contractor Rates**: Fetches billing type, base rate, and payday.
- **Payouts & Refunds**: 
    - Queries `Pending` payouts.
    - Queries approved refund requests and **auto-creates** missing payout entries using `payoutsService.CreateRefundPayout`.
- **Concurrency Management**: Services like `TaskOrderLogService` use a semaphore pattern (buffered channel) to limit Notion API calls (default: 10 concurrent requests).

### 2. Processing & Normalization

#### Currency Conversion (`extrapayment.ResolveAmountUSD`)
- If the original currency is not USD, the system converts it via the **Wise API**.
- **Caching**: Results are cached in **Redis** with a 30-day TTL. The cache stores the `displayRate` to ensure consistent recalculations even if the Notion amount is updated manually.

#### Description Cleaning (`stripDescriptionPrefixAndSuffix`)
- Strips Notion formula artifacts like `[PROJECT :: ...]` or `- $XX USD`.
- Extracts the project name and prepends it to the description (e.g., `FORTRESS - Feature Implementation`).

#### Service Fee Description Logic
Descriptions are generated based on contractor positions with the following priority:
1. **Design**: "Design Consulting Services Rendered (Date Range)"
2. **Operations**: "Operational Consulting Services Rendered (Date Range)"
3. **Default**: "Software Development Services Rendered (Date Range)"

### 3. Section Organization & Aggregation

#### Hourly Aggregation (`aggregateHourlyServiceFees`)
Multiple hourly work entries are consolidated into a single line item.
- Sums `TotalHours`, `TotalAmount`, and `TotalAmountUSD`.
- Validates consistency of rates across aggregated items, logging warnings if mismatches occur.

#### Visual Sections (`groupLineItemsIntoSections`)
Items are categorized into:
- **Development Work**: Aggregated Service Fees with `TaskOrderID`.
- **Fee**: Incentives (e.g., Management/Lead fees) from `InvoiceSplit`.
- **Expense Reimbursement**: Individual refunds.
- **Extra Payment**: Commissions and bonuses.

### 4. Proof of Work Formatting
The `FormatProofOfWorksByProject` method in `TaskOrderLogService` groups subitems by project and applies HTML formatting:
- **Project Headers**: Wrapped in `<b>` tags.
- **Deliverables**: Concatenated with line breaks and bullet points.
- **Spacers**: Injects `<div class="project-spacer"></div>` between different projects for layout control.

### 5. PDF Rendering
- **Engine**: Uses `wkhtmltopdf` (via `go-wkhtmltopdf`).
- **Template**: `contractor-invoice-template.html` uses Go's `html/template` with a custom `FuncMap` for money formatting, date manipulation, and HTML sanitization.

### 6. Notion Payable Creation
The system automatically creates or updates records in the **Contractor Payables** Notion database.
- **Payable Properties**: Includes total amount, currency, invoice ID, and period range.
- **Note Logic**: If the invoice contains **only** `Commission` and `Extra Payment` line items (e.g., no base service fee or refunds), the system automatically sets the **Note** field to `"Extra payment"` in Notion.
- **PDF Attachment**: The generated PDF is uploaded and attached to the record for record-keeping.

---

## Admin & Sync Features

### Force Sync Mode (`GenerateContractorInvoiceWithForceSync`)
Used by administrators to ensure data completeness before generation.
1. **Force Sync**: Queries Task Order Logs bypassing "Approved" status checks.
2. **Payout Overwrite**: If a payout already exists with status `New`, the system **deletes and recreates** it to reflect the latest timesheet data.
3. **Batch Processing**: Supports concurrent sync for all contractors in a specific batch.

## Related Files

| Component | File Path |
|-----------|-----------|
| Controller | `pkg/controller/invoice/contractor_invoice.go` |
| Payout Service | `pkg/service/notion/contractor_payouts.go` |
| Task Order Service | `pkg/service/notion/task_order_log.go` |
| Currency Logic | `pkg/extrapayment/amount.go` |
| HTML Template | `pkg/templates/contractor-invoice-template.html` |
