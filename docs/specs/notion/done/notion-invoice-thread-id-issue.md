# Notion Invoice Thread ID Issue

## Problem Summary

When marking a Notion-only invoice as paid via the Discord `?inv paid` button, the thank you email fails to send because the `ThreadID` is missing.

---

## Root Cause Analysis

### Flow: Sending Invoice from Notion

1. User clicks "Send Invoice" button in Notion
2. `HandleNotionInvoiceSend` webhook is triggered
3. Invoice email is sent via `SendInvoiceMail()` which returns a `threadID`
4. Status is updated to "Sent" in Notion
5. **Problem**: `threadID` was NOT stored in Notion

### Flow: Marking Invoice as Paid via Discord

1. User clicks `?inv paid` button in Discord
2. `processInvoicePaidConfirm` looks up invoice in PostgreSQL
3. If not found in PostgreSQL, fallback to Notion query
4. Extract invoice data from Notion (including `ThreadID`)
5. Call `SendInvoiceThankYouMail(invoice)`
6. **Problem**: `invoice.ThreadID` is empty, causing `ErrMissingThreadID` error

### Why ThreadID is Required

The thank you email must be sent as a **reply** to the original invoice email thread. Gmail uses `ThreadID` to link related emails together. Without it:
- The thank you email would be sent as a new standalone email
- Client loses the email conversation context
- `SendInvoiceThankYouMail` returns `ErrMissingThreadID` and fails

---

## Solution

### 1. Store Thread ID in Notion After Sending Invoice

**File**: `pkg/handler/webhook/notion_invoice.go`

When invoice is sent, store the `threadID` in Notion:

```go
updatePayload := map[string]interface{}{
    "properties": map[string]interface{}{
        "Status": map[string]interface{}{
            "status": map[string]string{
                "name": "Sent",
            },
        },
        "Thread ID": map[string]interface{}{
            "rich_text": []map[string]interface{}{
                {
                    "type": "text",
                    "text": map[string]string{
                        "content": threadID,
                    },
                },
            },
        },
    },
}
```

### 2. Extract Thread ID When Querying Invoice from Notion

**File**: `pkg/handler/webhook/notion_invoice_helpers.go`

In `extractInvoiceDataFromNotion()`:

```go
// Extract Thread ID (stored after sending invoice email)
threadID := ""
if threadIDProp, ok := props["Thread ID"]; ok && len(threadIDProp.RichText) > 0 {
    threadID = threadIDProp.RichText[0].PlainText
    l.Debug(fmt.Sprintf("extracted Thread ID: %s", threadID))
} else {
    l.Debug("Thread ID not found in Notion properties")
}

// Include in invoice model
invoice := &model.Invoice{
    // ... other fields
    ThreadID: threadID,
}
```

### 3. Use Thread ID When Sending Thank You Email

**File**: `pkg/handler/webhook/notion_invoice_helpers.go`

In `MarkNotionInvoiceAsPaid()`:

```go
// invoice already has ThreadID from extractInvoiceDataFromNotion()
err := h.service.GoogleMail.SendInvoiceThankYouMail(invoice)
```

---

## Notion Database Configuration

Add a new property to the **Client Invoice** database in Notion:

| Property Name | Type | Description |
|--------------|------|-------------|
| Thread ID | Rich Text | Gmail thread ID for email threading |

This field is automatically populated when the invoice is sent and should not be manually edited.

---

## Environment Variables

```bash
# Required for querying invoices from Notion
NOTION_CLIENT_INVOICE_DB_ID=<your-client-invoice-database-id>
```

---

## Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                    SEND INVOICE FLOW                                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Notion "Send Invoice" Button                                        │
│           │                                                          │
│           ▼                                                          │
│  HandleNotionInvoiceSend()                                           │
│           │                                                          │
│           ▼                                                          │
│  SendInvoiceMail() ──────► Returns threadID                          │
│           │                                                          │
│           ▼                                                          │
│  Update Notion Page:                                                 │
│    - Status = "Sent"                                                 │
│    - Thread ID = threadID  ◄──── NEW: Store threadID                 │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                    MARK AS PAID FLOW                                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Discord "?inv paid" Button                                          │
│           │                                                          │
│           ▼                                                          │
│  processInvoicePaidConfirm()                                         │
│           │                                                          │
│           ├──► Try PostgreSQL ──► Found? ──► Use existing flow       │
│           │                                                          │
│           └──► Not Found ──► QueryInvoiceFromNotionByNumber()        │
│                                    │                                 │
│                                    ▼                                 │
│                    extractInvoiceDataFromNotion()                    │
│                              │                                       │
│                              ├── Extract Month, Year                 │
│                              ├── Extract Recipients                  │
│                              └── Extract Thread ID  ◄──── NEW        │
│                                    │                                 │
│                                    ▼                                 │
│                    MarkNotionInvoiceAsPaid()                         │
│                              │                                       │
│                              ├── Update Status = "Paid"              │
│                              └── SendInvoiceThankYouMail(invoice)    │
│                                         │                            │
│                                         ▼                            │
│                              Uses invoice.ThreadID to reply          │
│                              to original email thread                │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Related Files

| File | Changes |
|------|---------|
| `pkg/config/config.go` | Added `ClientInvoice` to `NotionDatabase` struct |
| `pkg/handler/webhook/notion_invoice.go` | Store Thread ID after sending invoice |
| `pkg/handler/webhook/notion_invoice_helpers.go` | Extract Thread ID, query invoice by number, mark as paid |
| `pkg/handler/webhook/discord_interaction.go` | Notion fallback in `processInvoicePaidConfirm` |

---

## Testing

1. Send an invoice from Notion
2. Verify `Thread ID` property is populated in Notion
3. Click `?inv paid` button in Discord
4. Verify:
   - Notion status updated to "Paid"
   - Thank you email sent as reply to original thread
