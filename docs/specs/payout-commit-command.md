# Payout Commit Discord Command Specification

## Overview

Discord command to commit Contractor Payables from `Pending` to `Paid` status for a given month and batch (PayDay).

**Command:** `?payout commit <month> <batch>`

| Parameter | Format | Example | Description |
|-----------|--------|---------|-------------|
| `month` | YYYY-MM | 2025-01 | The billing period |
| `batch` | 1 or 15 | 15 | Maps to contractor's PayDay from Service Rate |

## User Flow

```
User: ?payout commit 2025-01 15
         │
         ▼
Discord validates args (month format, batch 1/15)
         │
         ▼
API queries Contractor Payables:
  - Payment Status = "Pending"
  - Period = 2025-01-01
         │
         ▼
For each payable, get contractor's PayDay from Service Rate
Filter where PayDay == batch (15)
         │
         ▼
Discord shows preview embed:
  ┌─────────────────────────────────────┐
  │ Confirm Payout Commit               │
  │                                     │
  │ Month: 2025-01                      │
  │ Batch: 15                           │
  │ Count: 3 payables                   │
  │ Total: $15,000.00                   │
  │                                     │
  │ Contractors:                        │
  │ • John Doe - $5,000.00              │
  │ • Jane Smith - $7,500.00            │
  │ • Bob Wilson - $2,500.00            │
  │                                     │
  │ [Confirm] [Cancel]                  │
  └─────────────────────────────────────┘
         │
         ▼
User clicks Confirm
         │
         ▼
API updates each payable:
  - Payment Status = "Paid"
  - Payment Date = now
         │
         ▼
Discord shows result:
  "Successfully committed 3 payables for 2025-01 batch 15"
```

## Database Schema References

### Contractor Payables (`2c264b29-b84c-8037-807c-000bf6d0792c`)

| Property | Type | Description |
|----------|------|-------------|
| `Payment Status` | Status | New, Pending, Paid, Cancelled |
| `Period` | Date | Month this payable covers (YYYY-MM-01) |
| `Payment Date` | Date | Date payment was processed |
| `Total` | Number | Total payable amount |
| `Currency` | Select | USD or VND |
| `Contractor` | Relation | Link to Contractors database |
| `Payout Items` | Relation | Links to Contractor Payouts |

### Contractor Payouts (`2c564b29-b84c-8045-80ee-000bee2e3669`)

| Property | Type | Description |
|----------|------|-------------|
| `Status` | Status | Pending, Paid, Cancelled |
| `01 Refund` | Relation | Links to Refund Request (if source is refund) |
| `02 Invoice Split` | Relation | Links to Invoice Split (if source is commission) |
| `Source Type` | Formula | "Service Fee", "Commission", "Refund", "Other" |

### Invoice Split (`2c364b29-b84c-804f-9856-000b58702dea`)

| Property | Type | Description |
|----------|------|-------------|
| `Status` | Select | Pending, Paid, Cancelled |
| `Amount` | Number | Split amount |
| `Currency` | Select | USD, VND, GBP, SGD |

### Refund Request (`2cc64b29-b84c-8066-adf2-cc56171cedf4`)

| Property | Type | Description |
|----------|------|-------------|
| `Status` | Status | Pending, Approved, Paid, Rejected, Cancelled |
| `Amount` | Number | Refund amount |
| `Currency` | Select | VND, USD |

### Service Rate (`2c464b29-b84c-80cf-bef6-000b42bce15e`)

| Property | Type | Description |
|----------|------|-------------|
| `PayDay` | Select | Day of month for payment (1, 15, etc.) |
| `Contractor` | Relation | Link to Contractors database |

---

## Cascade Status Update Logic

When committing payables, the system must update statuses across multiple related tables:

```
Contractor Payable (Pending → Paid)
    │
    └── Payout Items (Pending → Paid)
            │
            ├── Invoice Split (Pending → Paid)    ← if linked via "02 Invoice Split"
            │
            └── Refund Request (Approved → Paid)  ← if linked via "01 Refund"
```

### Update Sequence

For each Contractor Payable being committed:

1. **Get Payout Items** - Retrieve all linked payout records from `Payout Items` relation

2. **For each Payout Item:**
   - Check `02 Invoice Split` relation:
     - If not empty → Update Invoice Split `Status` to "Paid"
   - Check `01 Refund` relation:
     - If not empty → Update Refund Request `Status` to "Paid"
   - Update Payout `Status` to "Paid"

3. **Update Contractor Payable:**
   - Set `Payment Status` to "Paid"
   - Set `Payment Date` to current date

### Important Notes

- **Invoice Split**: Uses `Select` property type for Status
- **Refund Request**: Uses `Status` property type (different from Select)
- **Contractor Payouts**: Uses `Status` property type
- **Contractor Payables**: Uses `Status` property type

### Notion API Update Patterns

**Update Invoice Split Status (Select):**
```go
params := nt.UpdatePageParams{
    DatabasePageProperties: nt.DatabasePageProperties{
        "Status": nt.DatabasePageProperty{
            Select: &nt.SelectOptions{
                Name: "Paid",
            },
        },
    },
}
_, err := client.UpdatePage(ctx, invoiceSplitPageID, params)
```

**Update Refund Request Status (Status):**
```go
params := nt.UpdatePageParams{
    DatabasePageProperties: nt.DatabasePageProperties{
        "Status": nt.DatabasePageProperty{
            Status: &nt.SelectOptions{
                Name: "Paid",
            },
        },
    },
}
_, err := client.UpdatePage(ctx, refundPageID, params)
```

**Update Contractor Payout Status (Status):**
```go
params := nt.UpdatePageParams{
    DatabasePageProperties: nt.DatabasePageProperties{
        "Status": nt.DatabasePageProperty{
            Status: &nt.SelectOptions{
                Name: "Paid",
            },
        },
    },
}
_, err := client.UpdatePage(ctx, payoutPageID, params)
```

**Update Contractor Payable (Status + Payment Date):**
```go
params := nt.UpdatePageParams{
    DatabasePageProperties: nt.DatabasePageProperties{
        "Payment Status": nt.DatabasePageProperty{
            Status: &nt.SelectOptions{
                Name: "Paid",
            },
        },
        "Payment Date": nt.DatabasePageProperty{
            Date: &nt.Date{
                Start: nt.NewDateTime(time.Now(), false),
            },
        },
    },
}
_, err := client.UpdatePage(ctx, payablePageID, params)
```

## Design Decisions

### Batch Filtering

**Decision:** Use PayDay from Contractor's Service Rate relation

**Rationale:**
- Each contractor has an assigned PayDay (1 or 15) in their Service Rate
- This determines which bi-monthly batch they belong to
- Filtering by PayDay ensures correct contractors are included

### Confirmation Flow

**Decision:** Require user confirmation before committing

**Rationale:**
- Prevents accidental mass updates
- Shows preview with count, total, and contractor names
- User can verify before proceeding

### Empty Result Handling

**Decision:** Show info message with count=0

**Rationale:**
- Not an error condition
- User should know no pending payables exist for criteria
- Friendly UX

## Implementation Files

### fortress-discord

| File | Action | Description |
|------|--------|-------------|
| `pkg/discord/command/payout/command.go` | Create | Command handler |
| `pkg/discord/service/payout/service.go` | Create | Service layer |
| `pkg/discord/view/payout/payout.go` | Create | Discord embeds/views |
| `pkg/adapter/fortress/payout.go` | Create | API adapter |
| `pkg/model/payout.go` | Create | Data models |
| `pkg/discord/command/command.go` | Modify | Register command |

### fortress-api

| File | Action | Description |
|------|--------|-------------|
| `pkg/service/notion/contractor_payables.go` | Modify | Add query/update methods |
| `pkg/service/notion/contractor_payouts.go` | Modify | Add GetPayoutWithRelations, UpdatePayoutStatus |
| `pkg/service/notion/invoice_split.go` | Modify | Add UpdateInvoiceSplitStatus |
| `pkg/service/notion/refund_request.go` | Create | Add UpdateRefundRequestStatus |
| `pkg/handler/payout/payout.go` | Create | HTTP handlers |
| `pkg/controller/payout/payout.go` | Create | Business logic with cascade updates |
| `pkg/routes/v1.go` | Modify | Register routes |
| `pkg/handler/webhook/discord_interaction.go` | Modify | Handle confirm button |

## API Endpoints

### Preview Commit

```
GET /api/v1/contractor-payables/preview-commit?month=2025-01&batch=15
```

**Response:**
```json
{
  "month": "2025-01",
  "batch": 15,
  "count": 3,
  "total_amount": 15000.00,
  "contractors": [
    {"name": "John Doe", "amount": 5000.00},
    {"name": "Jane Smith", "amount": 7500.00},
    {"name": "Bob Wilson", "amount": 2500.00}
  ]
}
```

### Execute Commit

```
POST /api/v1/contractor-payables/commit
Content-Type: application/json

{
  "month": "2025-01",
  "batch": 15
}
```

**Response:**
```json
{
  "month": "2025-01",
  "batch": 15,
  "updated": 3,
  "failed": 0
}
```

## Error Handling

| Scenario | Response |
|----------|----------|
| Invalid month format | "Invalid month format. Use YYYY-MM (e.g., 2025-01)" |
| Invalid batch | "Batch must be 1 or 15" |
| No pending payables | Info message: "No pending payables found for 2025-01 batch 15" |
| Notion API error | Log error, return partial results if possible |
| Partial failure | Return success count + failed count |

## Permission Requirements

- Command requires admin/ops role
- API endpoints require `PayrollsRead` and `PayrollsCreate` permissions
