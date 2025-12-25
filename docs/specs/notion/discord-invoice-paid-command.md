# Discord Invoice Paid Command (`?inv paid`)

**Last Updated:** 2025-12-26
**Status:** Draft

## Overview

Rework the `?inv paid` Discord command to support marking invoices as paid from both PostgreSQL (Fortress) and Notion Client Invoices database.

## Background

Previously, the command used NocoDB accounting todos to trigger invoice payment via webhook. With the migration to Notion, we need to update the flow to:
1. Support Notion Client Invoices database
2. Maintain backward compatibility with PostgreSQL invoices
3. Handle Discord's 3-second interaction timeout

## Flow

```
User: ?inv paid <invoice_number>
           │
           ▼
┌─────────────────────────────────┐
│ 1. Validate args only           │
│    (no API calls - fast)        │
└─────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────┐
│ 2. Show confirmation dialog     │
│    immediately (<3s timeout)    │
└─────────────────────────────────┘
           │
           ▼ (User confirms)
┌─────────────────────────────────┐
│ 3. Show "Processing..." msg     │
└─────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────┐
│ 4. Call fortress-api endpoint   │
│    POST /invoices/mark-paid     │
└─────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────┐
│ 5. fortress-api:                │
│    - Check PostgreSQL           │
│    - Check Notion               │
│    - Update where found         │
│    - Return result              │
└─────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────┐
│ 6. Show success/error message   │
└─────────────────────────────────┘
```

## Components

### fortress-discord

**Command:** `pkg/discord/command/invoice/command.go`

| Method | Responsibility |
|--------|----------------|
| `paid()` | Validate args, show confirmation dialog |
| `ExecutePaidConfirmation()` | Show processing, call fortress-api, display result |

**Service:** `pkg/discord/service/invoice/service.go`

| Method | Responsibility |
|--------|----------------|
| `MarkInvoicePaid()` | Call fortress-api `POST /invoices/mark-paid` |

### fortress-api

**Endpoint:** `POST /api/v1/invoices/mark-paid`

**Request:**
```json
{
  "invoice_number": "INV-202512-PROJECT-ABC123"
}
```

**Response (Success):**
```json
{
  "data": {
    "invoice_number": "INV-202512-PROJECT-ABC123",
    "source": "notion",  // or "postgres" or "both"
    "paid_at": "2025-12-26T10:30:00Z"
  }
}
```

**Response (Error):**
```json
{
  "error": "Invoice not found in PostgreSQL or Notion"
}
```

**Handler:** `pkg/handler/invoice/mark_paid.go` (new)

## Step 5: fortress-api Logic (Detailed)

### Current Logic (NocoDB-based)

**Entry Points:**
- `pkg/handler/webhook/nocodb_invoice.go:MarkInvoiceAsPaidViaNoco` - NocoDB webhook
- `pkg/handler/webhook/basecamp.go:MarkInvoiceAsPaidViaBasecamp` - Basecamp webhook

**Flow:**
```
1. Webhook receives event (NocoDB status change to "paid")
2. Verify signature
3. Find invoice in PostgreSQL by invoice_number or fortress_invoice_id
4. Call controller.MarkInvoiceAsPaidWithTaskRef()
   ├── Validate status (must be "sent" or "overdue")
   ├── Set invoice.Status = "paid"
   └── Call processPaidInvoice() which runs 3 goroutines:
       ├── processPaidInvoiceData()
       │   ├── Set invoice.PaidAt = now
       │   ├── Update invoice in PostgreSQL
       │   ├── Store commission records
       │   └── Create accounting transaction
       ├── sendThankYouEmail()
       │   └── Send thank you email via Gmail
       └── movePaidInvoiceGDrive()
           └── Move PDF from "Sent" to "Paid" folder
5. Log to Discord audit channel
```

**Problems with current approach:**
- Requires webhook trigger (NocoDB/Basecamp)
- No direct API endpoint for Discord command
- NocoDB is being deprecated

---

### New Logic (Notion + PostgreSQL)

**Entry Point:** `pkg/handler/invoice/mark_paid.go:MarkInvoiceAsPaidByNumber` (new)

**Endpoint:** `POST /api/v1/invoices/mark-paid`

**Flow:**
```
1. Receive request with invoice_number
2. Search in parallel:
   ├── PostgreSQL: store.Invoice.One(db, &Query{Number: invoiceNumber})
   └── Notion: QueryDatabase(clientInvoicesDB, filter by title)
3. Determine source:
   ├── Not found in either → return error
   ├── Found in PostgreSQL only → source = "postgres"
   ├── Found in Notion only → source = "notion"
   └── Found in both → source = "both"
4. Validate status:
   ├── PostgreSQL: must be "sent" or "overdue"
   └── Notion: must be "Sent" or "Overdue"
5. Update based on source:
   ├── PostgreSQL (if exists):
   │   ├── Call existing MarkInvoiceAsPaidWithTaskRef()
   │   │   ├── Update status to "paid"
   │   │   ├── Set paid_at
   │   │   ├── Store commission
   │   │   ├── Create accounting transaction
   │   │   ├── Send thank you email
   │   │   └── Move PDF in GDrive
   │   └── Skip Basecamp/NocoDB todo management
   └── Notion (if exists):
       ├── Update Status to "Paid"
       ├── Set Paid Date to today
       ├── Send thank you email (extract recipients from Notion)
       └── Move PDF in GDrive (Sent → Paid folder)
6. Return success with source info
```

**Key Differences:**

| Aspect | Current (NocoDB) | New (Notion) |
|--------|------------------|--------------|
| Trigger | Webhook from NocoDB | Direct API call |
| Invoice Lookup | PostgreSQL only | PostgreSQL + Notion |
| External Task System | NocoDB accounting_todos | Notion Client Invoices |
| Update Flow | NocoDB → webhook → fortress-api | fortress-discord → API → fortress-api |
| Thank You Email | Always sent | Sent for both PostgreSQL and Notion |
| Commission | PostgreSQL invoices only | PostgreSQL invoices only |
| GDrive Move | PostgreSQL invoices only | Both PostgreSQL and Notion |

**New Controller Method:**

```go
// MarkInvoiceAsPaidByNumber handles marking invoice as paid from Discord command
// It searches both PostgreSQL and Notion, updates where found
func (c *controller) MarkInvoiceAsPaidByNumber(invoiceNumber string) (*MarkPaidResult, error) {
    l := c.logger.Fields(logger.Fields{
        "controller":    "invoice",
        "method":        "MarkInvoiceAsPaidByNumber",
        "invoiceNumber": invoiceNumber,
    })

    l.Debug("starting mark invoice as paid by number")

    // 1. Search PostgreSQL
    l.Debug("searching invoice in PostgreSQL")
    pgInvoice, pgErr := c.store.Invoice.One(c.repo.DB(), &invoice.Query{Number: invoiceNumber})
    if pgErr != nil {
        l.Debugf("invoice not found in PostgreSQL: %v", pgErr)
    } else {
        l.Debugf("found invoice in PostgreSQL: id=%s status=%s", pgInvoice.ID, pgInvoice.Status)
    }

    // 2. Search Notion
    l.Debug("searching invoice in Notion Client Invoices")
    notionPage, notionErr := c.service.Notion.QueryClientInvoiceByNumber(invoiceNumber)
    if notionErr != nil {
        l.Debugf("invoice not found in Notion: %v", notionErr)
    } else {
        l.Debugf("found invoice in Notion: pageID=%s", notionPage.ID)
    }

    // 3. Check if found anywhere
    if pgErr != nil && notionErr != nil {
        l.Debug("invoice not found in either PostgreSQL or Notion")
        return nil, errors.New("invoice not found")
    }

    result := &MarkPaidResult{
        InvoiceNumber: invoiceNumber,
        PaidAt:        time.Now(),
    }

    // 4. Update PostgreSQL if exists
    if pgInvoice != nil {
        l.Debug("processing PostgreSQL invoice")
        // Validate status
        if pgInvoice.Status != model.InvoiceStatusSent &&
           pgInvoice.Status != model.InvoiceStatusOverdue {
            return nil, fmt.Errorf("cannot mark as paid: status is %s", pgInvoice.Status)
        }
        // Use existing logic (includes commission, accounting, email, GDrive)
        c.MarkInvoiceAsPaidWithTaskRef(pgInvoice, nil, true)
        result.PostgresUpdated = true
        l.Debug("PostgreSQL invoice marked as paid")
    }

    // 5. Update Notion if exists
    if notionPage != nil {
        l.Debug("processing Notion invoice")
        // Validate status
        status := extractNotionStatus(notionPage)
        l.Debugf("Notion invoice status: %s", status)
        if status != "Sent" && status != "Overdue" {
            return nil, fmt.Errorf("cannot mark as paid: Notion status is %s", status)
        }

        // 5a. Update Notion page (Status + Paid Date)
        l.Debug("updating Notion invoice status to Paid")
        if err := c.service.Notion.UpdateInvoiceStatus(notionPage.ID, "Paid", time.Now()); err != nil {
            l.Errorf(err, "failed to update Notion invoice status")
            return nil, fmt.Errorf("failed to update Notion: %w", err)
        }

        // 5b. Extract invoice data for email and GDrive
        l.Debug("extracting Notion invoice data for post-processing")
        notionInvoice, err := c.extractNotionInvoiceData(notionPage)
        if err != nil {
            l.Errorf(err, "failed to extract Notion invoice data")
            // Continue - status already updated
        } else {
            // 5c. Send thank you email
            l.Debug("sending thank you email for Notion invoice")
            if err := c.service.GoogleMail.SendInvoiceThankYouMail(notionInvoice); err != nil {
                l.Errorf(err, "failed to send thank you email for Notion invoice")
            }

            // 5d. Move PDF in GDrive (Sent → Paid)
            l.Debug("moving Notion invoice PDF to Paid folder in GDrive")
            if err := c.service.GoogleDrive.MoveInvoicePDF(notionInvoice, "Sent", "Paid"); err != nil {
                l.Errorf(err, "failed to move Notion invoice PDF in GDrive")
            }
        }

        result.NotionUpdated = true
        l.Debug("Notion invoice marked as paid")
    }

    // 6. Determine source
    result.Source = determineSource(pgInvoice != nil, notionPage != nil)
    l.Debugf("mark invoice as paid completed: source=%s", result.Source)

    return result, nil
}

// MarkPaidResult contains the result of marking an invoice as paid
type MarkPaidResult struct {
    InvoiceNumber   string    `json:"invoice_number"`
    Source          string    `json:"source"` // "postgres", "notion", or "both"
    PaidAt          time.Time `json:"paid_at"`
    PostgresUpdated bool      `json:"postgres_updated"`
    NotionUpdated   bool      `json:"notion_updated"`
}

func determineSource(pgFound, notionFound bool) string {
    if pgFound && notionFound {
        return "both"
    }
    if pgFound {
        return "postgres"
    }
    return "notion"
}
```

## Notion Integration

**Database:** Client Invoices (`2bf64b29b84c80879a52ed2f9d493096`)

**Query (find by invoice number):**
```http
POST https://api.notion.com/v1/databases/2bf64b29b84c80879a52ed2f9d493096/query
Content-Type: application/json
Authorization: Bearer {NOTION_SECRET}
Notion-Version: 2022-06-28

{
  "filter": {
    "property": "(auto) Invoice Number",
    "title": {
      "contains": "INV-202512-PROJECT"
    }
  }
}
```

**Update (mark as paid):**
```http
PATCH https://api.notion.com/v1/pages/{page_id}
Content-Type: application/json
Authorization: Bearer {NOTION_SECRET}
Notion-Version: 2022-06-28

{
  "properties": {
    "Status": {
      "status": {
        "name": "Paid"
      }
    },
    "Paid Date": {
      "date": {
        "start": "2025-12-26"
      }
    }
  }
}
```

## Validation Rules

| Rule | Description |
|------|-------------|
| Invoice must exist | In PostgreSQL OR Notion (at least one) |
| Status must be valid | Only `Sent` or `Overdue` can be marked as `Paid` |
| Idempotent | If already `Paid`, return success without error |

## Error Cases

| Scenario | Error Message |
|----------|---------------|
| No invoice number provided | "Invoice number is required" |
| Invoice not found | "Invoice {number} not found" |
| Invalid status | "Cannot mark invoice as paid: current status is {status}" |
| Notion API error | "Failed to update Notion: {error}" |
| Fortress API error | "Failed to update invoice: {error}" |

## Migration Notes

- Remove NocoDB dependency from fortress-discord
- Remove `nocodb.QueryAccountingTodos()` and `nocodb.UpdateAccountingTodoStatus()` calls
- Fortress-api handles all business logic for invoice lookup and update
