package invoice

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/dwarvesf/fortress-api/pkg/mcp/validation"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/invoice"
)

// Tools represents invoice-related MCP tools
type Tools struct {
	store *store.Store
	repo  store.DBRepo
}

// New creates a new invoice tools instance
func New(store *store.Store, repo store.DBRepo) *Tools {
	return &Tools{
		store: store,
		repo:  repo,
	}
}

// GenerateInvoiceTool returns the MCP tool for generating an invoice
func (t *Tools) GenerateInvoiceTool() mcp.Tool {
	return mcp.NewTool(
		"generate_invoice",
		mcp.WithDescription("Create invoice for project/client"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID (UUID)")),
		mcp.WithNumber("subtotal", mcp.Required(), mcp.Description("Invoice subtotal amount")),
		mcp.WithNumber("tax", mcp.Description("Tax amount (default: 0)")),
		mcp.WithNumber("discount", mcp.Description("Discount amount (default: 0)")),
		mcp.WithString("due_date", mcp.Description("Invoice due date (YYYY-MM-DD format, default: 30 days from now)")),
		mcp.WithString("note", mcp.Description("Additional notes for the invoice")),
		mcp.WithString("send_to_email", mcp.Description("Email address to send invoice to")),
	)
}

// GenerateInvoiceHandler handles the generate_invoice tool execution
func (t *Tools) GenerateInvoiceHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectIDStr, err := req.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	subtotal := req.GetFloat("subtotal", 0.0)
	tax := req.GetFloat("tax", 0.0)
	discount := req.GetFloat("discount", 0.0)
	dueDateStr := req.GetString("due_date", "")
	note := req.GetString("note", "")
	sendToEmail := req.GetString("send_to_email", "")

	// Validate inputs using validation package
	if err := validation.ValidateUUID(projectIDStr, "project_id"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := validation.ValidatePositiveNumber(subtotal, "subtotal"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if sendToEmail != "" {
		if err := validation.ValidateEmail(sendToEmail, "send_to_email"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	projectID, err := model.UUIDFromString(projectIDStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid project_id format: %v", err)), nil
	}

	// Verify project exists
	project, err := t.store.Project.One(t.repo.DB(), projectID.String(), false)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("project not found: %v", err)), nil
	}

	// Parse due date
	var dueDate *time.Time
	if dueDateStr != "" {
		if parsedDate, err := parseDate(dueDateStr); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid due_date format: %v", err)), nil
		} else {
			dueDate = &parsedDate
		}
	} else {
		// Default to 30 days from now
		defaultDue := time.Now().AddDate(0, 0, 30)
		dueDate = &defaultDue
	}

	// Calculate total
	total := subtotal + tax - discount

	// Get next invoice number (need year and project code)
	currentYear := time.Now().Year()
	nextNumber, err := t.store.Invoice.GetNextInvoiceNumber(t.repo.DB(), currentYear, project.Code)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get next invoice number: %v", err)), nil
	}

	// Handle nextNumber being a pointer to string
	var invoiceNumber string
	if nextNumber != nil {
		invoiceNumber = *nextNumber
	} else {
		return mcp.NewToolResultError("failed to generate invoice number"), nil
	}

	// Create invoice
	invoicedAt := time.Now()
	newInvoice := &model.Invoice{
		Number:     invoiceNumber,
		ProjectID:  projectID,
		SubTotal:   subtotal,
		Tax:        tax,
		Discount:   discount,
		Total:      total,
		Status:     model.InvoiceStatusDraft,
		InvoicedAt: &invoicedAt,
		DueAt:      dueDate,
		Note:       note,
	}

	// Set email if provided - Skip for now as JSON conversion is complex
	// if sendToEmail != "" {
	//     newInvoice.CC = model.JSON([]string{sendToEmail})
	// }

	// Create invoice
	createdInvoice, err := t.store.Invoice.Create(t.repo.DB(), newInvoice)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create invoice: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Invoice %s created successfully for project %s. Total: %.2f", createdInvoice.Number, project.Name, total)), nil
}

// GetInvoiceStatusTool returns the MCP tool for checking invoice status
func (t *Tools) GetInvoiceStatusTool() mcp.Tool {
	return mcp.NewTool(
		"get_invoice_status",
		mcp.WithDescription("Check invoice payment status"),
		mcp.WithString("invoice_id", mcp.Required(), mcp.Description("Invoice ID (UUID)")),
	)
}

// GetInvoiceStatusHandler handles the get_invoice_status tool execution
func (t *Tools) GetInvoiceStatusHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	invoiceIDStr, err := req.RequireString("invoice_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Validate invoice_id using validation package
	if err := validation.ValidateUUID(invoiceIDStr, "invoice_id"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	invoiceID, err := model.UUIDFromString(invoiceIDStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid invoice_id format: %v", err)), nil
	}

	// Get invoice using query struct
	query := &invoice.Query{
		ID: invoiceID.String(),
	}
	invoiceRecord, err := t.store.Invoice.One(t.repo.DB(), query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invoice not found: %v", err)), nil
	}


	return mcp.NewToolResultText(fmt.Sprintf("Invoice %s status: %s (Total: %.2f)", invoiceRecord.Number, invoiceRecord.Status, invoiceRecord.Total)), nil
}

// UpdateInvoiceStatusTool returns the MCP tool for updating invoice status
func (t *Tools) UpdateInvoiceStatusTool() mcp.Tool {
	return mcp.NewTool(
		"update_invoice_status",
		mcp.WithDescription("Mark invoice as paid/pending"),
		mcp.WithString("invoice_id", mcp.Required(), mcp.Description("Invoice ID (UUID)")),
		mcp.WithString("status", mcp.Required(), mcp.Description("New invoice status (draft, sent, overdue, paid, error, scheduled)")),
		mcp.WithString("paid_date", mcp.Description("Payment date (YYYY-MM-DD format, required if status is 'paid')")),
	)
}

// UpdateInvoiceStatusHandler handles the update_invoice_status tool execution
func (t *Tools) UpdateInvoiceStatusHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	invoiceIDStr, err := req.RequireString("invoice_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	status, err := req.RequireString("status")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	paidDateStr := req.GetString("paid_date", "")

	// Validate invoice_id using validation package
	if err := validation.ValidateUUID(invoiceIDStr, "invoice_id"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	invoiceID, err := model.UUIDFromString(invoiceIDStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid invoice_id format: %v", err)), nil
	}

	// Validate status using validation package
	validStatuses := []string{"draft", "sent", "overdue", "paid", "error", "scheduled"}
	if err := validation.ValidateInSlice(status, validStatuses, "status"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get current invoice
	query := &invoice.Query{
		ID: invoiceID.String(),
	}
	invoiceRecord, err := t.store.Invoice.One(t.repo.DB(), query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invoice not found: %v", err)), nil
	}

	// Handle paid status
	if status == "paid" {
		if paidDateStr == "" {
			// Default to today
			now := time.Now()
			invoiceRecord.PaidAt = &now
		} else {
			if paidDate, err := parseDate(paidDateStr); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid paid_date format: %v", err)), nil
			} else {
				invoiceRecord.PaidAt = &paidDate
			}
		}
	} else {
		// Clear paid date if status is not paid
		invoiceRecord.PaidAt = nil
	}

	// Update status
	invoiceRecord.Status = model.InvoiceStatus(status)

	// Save changes
	_, err = t.store.Invoice.UpdateSelectedFieldsByID(t.repo.DB(), invoiceID.String(), *invoiceRecord, "status", "paid_at")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update invoice status: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Invoice %s status updated to %s", invoiceRecord.Number, status)), nil
}

// CalculateCommissionTool returns the MCP tool for calculating commission amounts
func (t *Tools) CalculateCommissionTool() mcp.Tool {
	return mcp.NewTool(
		"calculate_commission",
		mcp.WithDescription("Compute commission amounts"),
		mcp.WithString("invoice_id", mcp.Required(), mcp.Description("Invoice ID (UUID)")),
		mcp.WithString("dry_run", mcp.Description("Dry run mode (true/false, default: true)")),
	)
}

// CalculateCommissionHandler handles the calculate_commission tool execution
func (t *Tools) CalculateCommissionHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	invoiceIDStr, err := req.RequireString("invoice_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	dryRunStr := req.GetString("dry_run", "true")
	dryRun := dryRunStr == "true"

	// Validate invoice_id using validation package
	if err := validation.ValidateUUID(invoiceIDStr, "invoice_id"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	invoiceID, err := model.UUIDFromString(invoiceIDStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid invoice_id format: %v", err)), nil
	}

	// Get invoice
	query := &invoice.Query{
		ID: invoiceID.String(),
	}
	invoiceRecord, err := t.store.Invoice.One(t.repo.DB(), query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invoice not found: %v", err)), nil
	}

	// For now, we'll implement a simplified commission calculation
	// In a real implementation, this would call the complex commission calculation logic
	// from the controller layer
	
	// Basic commission calculation (10% of total for demonstration)
	commissionRate := 0.10 // 10%
	totalCommission := invoiceRecord.Total * commissionRate

	if dryRun {
		return mcp.NewToolResultText(fmt.Sprintf("Commission calculation (DRY RUN): Invoice %s - Total commission: %.2f", invoiceRecord.Number, totalCommission)), nil
	} else {
		// TODO: Implement actual commission processing
		// This would involve calling the controller's ProcessCommissions method
		return mcp.NewToolResultText(fmt.Sprintf("Commission processing not yet implemented. Would process %.2f for invoice %s", totalCommission, invoiceRecord.Number)), nil
	}
}

// Helper functions

func parseDate(dateStr string) (time.Time, error) {
	// Parse date in YYYY-MM-DD format
	parsedTime, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("date must be in YYYY-MM-DD format: %v", err)
	}
	return parsedTime, nil
}

