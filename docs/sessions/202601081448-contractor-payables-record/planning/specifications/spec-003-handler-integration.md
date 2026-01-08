# Specification: Handler Integration

## Overview
Integrate Contractor Payables record creation into the invoice generation handler.

## File Location
`pkg/handler/invoice/invoice.go`

## Integration Point
After successful Google Drive upload (line ~440), before building response.

## Current Flow
```
1. Parse request
2. Validate month format
3. Generate invoice data (controller)
4. Generate PDF (controller)
5. Upload to Google Drive ← INSERT AFTER THIS
6. Build response
7. Return success
```

## New Step: 5.5 Create Payables Record

### Condition
- Only execute when `!req.SkipUpload`
- Must have successful upload (fileURL is set)

### Implementation
```go
// 5.5 Create Contractor Payables record in Notion
if !req.SkipUpload {
    l.Debug("[DEBUG] creating contractor payables record in Notion")

    payableInput := notion.CreatePayableInput{
        ContractorPageID: invoiceData.ContractorPageID,
        Total:            invoiceData.TotalUSD,
        Currency:         "USD",
        Period:           invoiceData.Month + "-01",
        InvoiceDate:      time.Now().Format("2006-01-02"),
        InvoiceID:        invoiceData.InvoiceNumber,
        PayoutItemIDs:    invoiceData.PayoutPageIDs,
        AttachmentURL:    fileURL,
    }

    l.Debug(fmt.Sprintf("[DEBUG] payable input: contractor=%s total=%.2f payoutItems=%d",
        payableInput.ContractorPageID, payableInput.Total, len(payableInput.PayoutItemIDs)))

    payablePageID, err := h.service.Notion.ContractorPayables.CreatePayable(c.Request.Context(), payableInput)
    if err != nil {
        l.Error(err, "[DEBUG] failed to create contractor payables record")
        // Non-fatal: continue with response
    } else {
        l.Debug(fmt.Sprintf("[DEBUG] contractor payables record created: pageID=%s", payablePageID))
    }
}
```

## Service Access
- Access via `h.service.Notion.ContractorPayables`
- Requires service to be initialized in `pkg/service/service.go`

## Import Requirements
- `notion` package already imported via existing code
- No new imports needed
