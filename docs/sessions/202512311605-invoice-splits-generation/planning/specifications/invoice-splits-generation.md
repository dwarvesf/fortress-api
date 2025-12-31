# Specification: Invoice Splits Generation

## Overview

Generate commission split records when client invoice is marked as paid.

## Components

### 1. Worker Message Type

```go
// pkg/model/worker_message.go or pkg/worker/message_types.go
const GenerateInvoiceSplitsMsg = "generate_invoice_splits"

type GenerateInvoiceSplitsPayload struct {
    InvoicePageID string
}
```

### 2. ClientInvoiceService Extensions

**File:** `pkg/service/notion/invoice.go`

#### 2.1 QueryInvoiceLineItemsWithCommissions
Query line items with commission data for a given invoice.

**Input:** `invoicePageID string`
**Output:** `[]LineItemCommissionData, error`

```go
type LineItemCommissionData struct {
    PageID              string
    DeploymentPageID    string

    // Commission percentages
    SalesPercent        float64
    AccountMgrPercent   float64
    DeliveryLeadPercent float64
    HiringRefPercent    float64

    // Calculated amounts
    SalesAmount         float64
    AccountMgrAmount    float64
    DeliveryLeadAmount  float64
    HiringRefAmount     float64

    // Person page IDs (from rollups)
    SalesPersonIDs      []string
    AccountMgrIDs       []string
    DeliveryLeadIDs     []string
    HiringRefIDs        []string

    // Metadata
    Currency            string
    Month               time.Time
    ProjectCode         string
}
```

#### 2.2 MarkSplitsGenerated
Update invoice `Splits Generated` checkbox to true.

**Input:** `invoicePageID string`
**Output:** `error`

#### 2.3 IsSplitsGenerated
Check if splits already generated (idempotency).

**Input:** `invoicePageID string`
**Output:** `bool, error`

### 3. InvoiceSplitService Extensions

**File:** `pkg/service/notion/invoice_split.go`

#### 3.1 CreateCommissionSplit
Create a single invoice split record.

**Input:**
```go
type CreateCommissionSplitInput struct {
    Name              string
    Amount            float64
    Currency          string
    Month             time.Time
    Role              string    // Sales, Account Manager, Delivery Lead, Hiring Referral
    Type              string    // Commission
    Status            string    // Pending
    ContractorPageID  string    // Maps to "Person" relation (→ Contractors)
    DeploymentPageID  string
    InvoiceItemPageID string
    InvoicePageID     string
}
```

**Notion Property Mapping:**
| Input Field | Notion Property | Type |
|-------------|-----------------|------|
| Name | Name | Title |
| Amount | Amount | Number |
| Currency | Currency | Select |
| Month | Month | Date |
| Role | Role | Select |
| Type | Type | Select |
| Status | Status | Select |
| ContractorPageID | Person | Relation → Contractors |
| DeploymentPageID | Deployment | Relation |
| InvoiceItemPageID | Invoice Item | Relation |
| InvoicePageID | Client Invoices | Relation |

**Output:** `*InvoiceSplitData, error`

### 4. Worker Handler

**File:** `pkg/worker/worker.go`

Add case for `GenerateInvoiceSplitsMsg` in `ProcessMessage()` switch.

#### 4.1 handleGenerateInvoiceSplits

```go
func (w *Worker) handleGenerateInvoiceSplits(l logger.Logger, payload interface{}) error {
    // 1. Extract payload
    // 2. Check if splits already generated (idempotency)
    // 3. Query line items with commission data
    // 4. For each line item, for each role with amount > 0:
    //    - Create invoice split record
    // 5. Mark splits generated on invoice
}
```

### 5. Integration Point

**File:** `pkg/controller/invoice/mark_paid.go`

In `processNotionInvoicePaid()`, after status update, enqueue splits generation:

```go
// After line 161 (status update)
c.worker.Enqueue(GenerateInvoiceSplitsMsg, GenerateInvoiceSplitsPayload{
    InvoicePageID: page.ID,
})
```

## Data Flow

```
mark-paid request
       │
       ▼
processNotionInvoicePaid()
       │
       ├─► Update status to "Paid"
       │
       ├─► worker.Enqueue(GenerateInvoiceSplitsMsg)
       │
       └─► Return (send email, move PDF in parallel)

Worker (async)
       │
       ▼
handleGenerateInvoiceSplits()
       │
       ├─► Check Splits Generated flag
       │
       ├─► Query line items with commissions
       │
       ├─► For each commission (Sales, AM, DL, HR):
       │       │
       │       └─► Create Invoice Split record
       │
       └─► Mark Splits Generated = true
```

## Error Handling

- Log errors but don't fail the entire flow
- Invoice marked as paid even if splits fail
- Can manually re-trigger via Notion "Generate Splits" button
