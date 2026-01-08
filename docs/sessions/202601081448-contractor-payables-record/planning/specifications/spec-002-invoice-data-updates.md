# Specification: ContractorInvoiceData Updates

## Overview
Update the `ContractorInvoiceData` struct to include data needed for creating the Payables record.

## File Location
`pkg/controller/invoice/contractor_invoice.go`

## Current Struct (partial)
```go
type ContractorInvoiceData struct {
    InvoiceNumber     string
    ContractorName    string
    Month             string
    // ... existing fields
}
```

## Required Additions

### New Fields
```go
type ContractorInvoiceData struct {
    // ... existing fields ...

    // New fields for Payables record creation
    ContractorPageID string   // Contractor Notion page ID (from rates query)
    PayoutPageIDs    []string // Payout item page IDs for relation
}
```

## Data Population

### ContractorPageID
- Source: `rateData.ContractorPageID` from ContractorRatesService query
- Location: Line ~110 in GenerateContractorInvoice after rates query
- Already available in `rateData` - just needs to be passed through

### PayoutPageIDs
- Source: Collect from `payouts` slice during iteration
- Location: Line ~213-324 loop where payouts are processed
- Each `payout.PageID` should be collected

## Implementation Location
`pkg/controller/invoice/contractor_invoice.go:GenerateContractorInvoice`

```go
// After line 117 (rateData query)
// rateData.ContractorPageID is already available

// During payout processing loop (line ~213)
var payoutPageIDs []string
for i, payout := range payouts {
    payoutPageIDs = append(payoutPageIDs, payout.PageID)
    // ... existing processing
}

// When building invoiceData (line ~521)
invoiceData := &ContractorInvoiceData{
    // ... existing fields
    ContractorPageID: rateData.ContractorPageID,
    PayoutPageIDs:    payoutPageIDs,
}
```
