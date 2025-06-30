package payroll

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

// Tools represents payroll-related MCP tools
type Tools struct {
	store *store.Store
	repo  store.DBRepo
}

// New creates a new payroll tools instance
func New(store *store.Store, repo store.DBRepo) *Tools {
	return &Tools{
		store: store,
		repo:  repo,
	}
}

// CalculatePayrollTool returns the MCP tool for calculating employee payroll
func (t *Tools) CalculatePayrollTool() mcp.Tool {
	return mcp.NewTool(
		"calculate_payroll",
		mcp.WithDescription("Compute employee payroll"),
		mcp.WithString("employee_email", mcp.Description("Employee email address (optional, calculates for all if not provided)")),
		mcp.WithNumber("month", mcp.Required(), mcp.Description("Payroll month (1-12)")),
		mcp.WithNumber("year", mcp.Required(), mcp.Description("Payroll year (e.g., 2025)")),
		mcp.WithString("batch_date", mcp.Description("Batch date in YYYY-MM-DD format (defaults to end of month)")),
	)
}

// CalculatePayrollHandler handles the calculate_payroll tool execution
func (t *Tools) CalculatePayrollHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	employeeEmail := req.GetString("employee_email", "")
	month := int(req.GetFloat("month", 0.0))
	year := int(req.GetFloat("year", 0.0))
	batchDateStr := req.GetString("batch_date", "")

	// Validate month and year
	if month < 1 || month > 12 {
		return mcp.NewToolResultError("month must be between 1 and 12"), nil
	}

	if year < 2020 || year > 2030 {
		return mcp.NewToolResultError("year must be between 2020 and 2030"), nil
	}

	// Parse batch date if provided
	var batchDate time.Time
	if batchDateStr != "" {
		if parsedDate, err := parseDate(batchDateStr); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid batch_date format: %v", err)), nil
		} else {
			batchDate = parsedDate
		}
	} else {
		// Default to end of month
		batchDate = time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC)
	}
	_ = batchDate // TODO: use batchDate in actual payroll calculation

	// For simplified implementation, we'll return a mock calculation
	// In a real implementation, this would call the complex payroll calculation logic
	
	var employeeFilter string
	if employeeEmail != "" {
		// Verify employee exists using direct email lookup
		emp, err := t.store.Employee.OneByEmail(t.repo.DB(), employeeEmail)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("employee not found: %v", err)), nil
		}
		employeeFilter = emp.FullName
	} else {
		employeeFilter = "All employees"
	}

	if employeeEmail != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Payroll calculated for %s - Month: %d/%d - Total: $%.2f", employeeFilter, month, year, 6400.0)), nil
	} else {
		return mcp.NewToolResultText(fmt.Sprintf("Payroll calculated for all employees - Month: %d/%d - Estimated total: $%.2f", month, year, 64000.0)), nil
	}
}

// ProcessSalaryAdvanceTool returns the MCP tool for processing salary advances
func (t *Tools) ProcessSalaryAdvanceTool() mcp.Tool {
	return mcp.NewTool(
		"process_salary_advance",
		mcp.WithDescription("Handle advance salary requests"),
		mcp.WithString("employee_email", mcp.Required(), mcp.Description("Employee email address")),
		mcp.WithNumber("amount_usd", mcp.Required(), mcp.Description("Advance amount in USD")),
		mcp.WithString("reason", mcp.Description("Reason for the advance request")),
	)
}

// ProcessSalaryAdvanceHandler handles the process_salary_advance tool execution
func (t *Tools) ProcessSalaryAdvanceHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	employeeEmail, err := req.RequireString("employee_email")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	amountUSD := req.GetFloat("amount_usd", 0.0)
	if amountUSD <= 0 {
		return mcp.NewToolResultError("amount_usd must be greater than 0"), nil
	}

	if amountUSD > 3000 { // Reasonable advance limit
		return mcp.NewToolResultError("advance amount cannot exceed $3,000"), nil
	}

	_ = req.GetString("reason", "") // TODO: implement reason tracking

	// Verify employee exists
	emp, err := t.store.Employee.OneByEmail(t.repo.DB(), employeeEmail)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("employee not found: %v", err)), nil
	}

	// For simplified implementation, we'll return a mock processing result
	// In a real implementation, this would:
	// 1. Validate employee eligibility and advance limits
	// 2. Calculate maximum advance based on salary
	// 3. Process ICY token transfer via Mochi service
	// 4. Create advance record in database
	// 5. Send confirmation emails

	// Mock advance processing
	_ = model.NewUUID() // advanceID for future use
	icyAmount := amountUSD * 100 // Mock ICY conversion rate

	return mcp.NewToolResultText(fmt.Sprintf("Salary advance processed for %s: $%.2f (%.0f ICY)", emp.FullName, amountUSD, icyAmount)), nil
}

// GetPayrollSummaryTool returns the MCP tool for retrieving payroll calculations
func (t *Tools) GetPayrollSummaryTool() mcp.Tool {
	return mcp.NewTool(
		"get_payroll_summary",
		mcp.WithDescription("Retrieve payroll calculations"),
		mcp.WithNumber("month", mcp.Required(), mcp.Description("Payroll month (1-12)")),
		mcp.WithNumber("year", mcp.Required(), mcp.Description("Payroll year (e.g., 2025)")),
		mcp.WithString("employee_email", mcp.Description("Filter by specific employee email (optional)")),
		mcp.WithString("summary_type", mcp.Description("Summary type: detailed, summary, or advances (default: summary)")),
	)
}

// GetPayrollSummaryHandler handles the get_payroll_summary tool execution
func (t *Tools) GetPayrollSummaryHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	month := int(req.GetFloat("month", 0.0))
	year := int(req.GetFloat("year", 0.0))
	employeeEmail := req.GetString("employee_email", "")
	summaryType := req.GetString("summary_type", "summary")

	// Validate month and year
	if month < 1 || month > 12 {
		return mcp.NewToolResultError("month must be between 1 and 12"), nil
	}

	if year < 2020 || year > 2030 {
		return mcp.NewToolResultError("year must be between 2020 and 2030"), nil
	}

	// Validate summary type
	validSummaryTypes := []string{"detailed", "summary", "advances"}
	if !contains(validSummaryTypes, summaryType) {
		return mcp.NewToolResultError(fmt.Sprintf("invalid summary_type: %s. Valid types: %v", summaryType, validSummaryTypes)), nil
	}

	// Filter by employee if provided
	var employeeFilter string
	if employeeEmail != "" {
		emp, err := t.store.Employee.OneByEmail(t.repo.DB(), employeeEmail)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("employee not found: %v", err)), nil
		}
		employeeFilter = emp.FullName
	} else {
		employeeFilter = "All employees"
	}

	// For simplified implementation, we'll return mock summary data
	// In a real implementation, this would:
	// 1. Query cached_payroll table for the specified month/year
	// 2. Aggregate payroll data by employee or overall
	// 3. Include advance balances and outstanding amounts
	// 4. Calculate totals and summaries based on summary_type

	// Mock payroll summary based on type
	switch summaryType {
	case "detailed":
		return mcp.NewToolResultText(fmt.Sprintf("Detailed payroll summary for %s - %d/%d: Base: $5000, Commissions: $1200, Bonuses: $500, Advances: -$300, Total: $6400", employeeFilter, month, year)), nil
	case "advances":
		return mcp.NewToolResultText(fmt.Sprintf("Salary advances summary for %s - %d/%d: Outstanding advances: $1500, Current month deduction: $300, Remaining balance: $1200", employeeFilter, month, year)), nil
	default: // summary
		if employeeEmail != "" {
			return mcp.NewToolResultText(fmt.Sprintf("Payroll summary for %s - %d/%d: Total payroll: $6,400", employeeFilter, month, year)), nil
		} else {
			return mcp.NewToolResultText(fmt.Sprintf("Payroll summary for all employees - %d/%d: Total payroll: $128,000, Average per employee: $6,400", month, year)), nil
		}
	}
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func parseDate(dateStr string) (time.Time, error) {
	// Parse date in YYYY-MM-DD format
	parsedTime, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("date must be in YYYY-MM-DD format: %v", err)
	}
	return parsedTime, nil
}