# Contractor Invoice Generation

## Overview

Contractor invoices aggregate pending payouts from three sources:
- **Contractor Billing** - Hourly/Fixed billing from Task Order Log
- **Invoice Split** - Commission payments
- **Refund Request** - Refund line items

## Current Flow (to be replaced)

```
1. Query Contractor Rates (by discord + month)
2. Query Task Order Log (by contractor + month)
3. Query Order Subitems
4. Build line items from subitems
5. Query Bank Account
6. Convert to USD
7. Generate invoice
```

## New Flow

```
1. Query Contractor Rates (by discord + month)
2. Query Pending Payouts (by contractor, status=Pending)
3. For each payout, build line item based on source:
   - Contractor Billing → get from payout Amount field
   - Invoice Split → get from payout Amount field
   - Refund Request → get from payout Amount field
4. Query Bank Account
5. Convert to USD
6. Generate invoice
```

## Data Flow

```
Query Pending Payouts (by Person + Status=Pending)
  │
  ├── Type = "Contractor Payroll"
  │     └── Check Contractor Billing (P?Ey) relation
  │           └── Retrieve from Contractor Billing:
  │                 - Total Hours Worked (rollup from Task Order Log)
  │                 - Hourly Rate (rollup from Contractor Rate)
  │                 - Fixed Fee (rollup from Contractor Rate)
  │                 - Billing Type (rollup from Contractor Rate)
  │                 - Proof of Works (rollup from Task Order Log)
  │                 - Total Amount (formula)
  │
  ├── Type = "Commission"
  │     └── Check Invoice Split (\CEP:) relation
  │           └── Retrieve from Invoice Split:
  │                 - Amount
  │                 - Role (Sales, Account Manager, etc.)
  │
  └── Type = "Refund"
        └── Check Refund Request (cS>|) relation
              └── Retrieve from Refund Request:
                    - Amount
                    - Refund details
```

## Query: Pending Payouts by Contractor

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
        "property": "Status",
        "status": {
          "equals": "Pending"
        }
      }
    ]
  }
}
```

## Source Type Determination

The `Source Type` formula in Payouts determines payout category:
- `Contractor Payroll` → Contractor Billing relation is set
- `Commission` → Invoice Split relation is set
- `Refund` → Refund Request relation is set
- `Other` → No source relation is set

## Direction Handling

- `Outgoing (you pay)` → Company pays contractor (positive amount)
- `Incoming (you receive)` → Contractor pays company (negative amount/deduction)

## Refactoring Plan

### Step 1: Create PayoutsService
- File: `pkg/service/notion/payouts.go`
- Method: `QueryPendingPayoutsByContractor(ctx, contractorPageID)`
- Returns: `[]*PayoutEntry` with Type, Direction, ContractorBillingID, InvoiceSplitID, RefundRequestID

### Step 2: Create ContractorBillingService
- File: `pkg/service/notion/contractor_billing.go`
- Method: `GetContractorBillingByID(ctx, billingPageID)`
- Returns: `*ContractorBillingData` with TotalHoursWorked, HourlyRate, FixedFee, BillingType, ProofOfWorks, TotalAmount

### Step 3: Create InvoiceSplitService
- File: `pkg/service/notion/invoice_split.go`
- Method: `GetInvoiceSplitByID(ctx, splitPageID)`
- Returns: `*InvoiceSplitData` with Amount, Role, Currency

### Step 4: Update ContractorInvoiceLineItem struct
- Add `SourceType` field (Contractor Payroll, Commission, Refund)
- Add `Direction` field (Outgoing, Incoming)

### Step 5: Update GenerateContractorInvoice
- Replace Task Order Log query with Payouts query
- For each payout:
  - If Type = "Contractor Payroll": call ContractorBillingService to get details
  - If Type = "Commission": call InvoiceSplitService to get details
  - If Type = "Refund": get amount from Payout directly
- Build line items from retrieved data

### Step 6: Update PDF template
- Handle different line item types (payroll vs commission vs refund)
- Show direction for incoming payments (deductions)

## Related Databases

| Database | ID | Purpose |
|----------|-----|---------|
| Payouts | `2c564b29-b84c-8045-80ee-000bee2e3669` | Aggregates all payout entries |
| Contractor Billing | `2c264b29-b84c-8037-807c-000bf6d0792c` | Links Task Order Log to rates |
| Invoice Split | `2c364b29-b84c-804f-9856-000b58702dea` | Commission payments |
| Refund Request | `2cc64b29-b84c-8066-adf2-cc56171cedf4` | Refund entries |
| Task Order Log | `2b964b29-b84c-801c-accb-dc8ca1e38a5f` | Hours and proof of work |
