package worker

// GenerateInvoiceSplitsMsg is the worker message type for generating invoice splits
const GenerateInvoiceSplitsMsg = "generate_invoice_splits"

// GenerateInvoiceSplitsPayload contains the data needed to generate invoice splits
type GenerateInvoiceSplitsPayload struct {
	InvoicePageID string
}
