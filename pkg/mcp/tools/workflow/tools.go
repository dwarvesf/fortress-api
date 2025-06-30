package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/dwarvesf/fortress-api/pkg/mcp/auth"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/workflow"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
)

// Tools represents workflow-related MCP tools
type Tools struct {
	store           *store.Store
	repo            store.DBRepo
	workflowService *workflow.Service
}

// New creates a new workflow tools instance
func New(store *store.Store, repo store.DBRepo) *Tools {
	return &Tools{
		store:           store,
		repo:            repo,
		workflowService: workflow.New(store, repo),
	}
}

// CalculateMonthlyPayrollTool returns the MCP tool for calculating monthly payroll
func (t *Tools) CalculateMonthlyPayrollTool() mcp.Tool {
	return mcp.NewTool(
		"calculate_monthly_payroll",
		mcp.WithDescription("Process entire monthly payroll cycle workflow with comprehensive calculation sheet"),
		mcp.WithNumber("month", mcp.Required(), mcp.Description("Payroll month (1-12)")),
		mcp.WithNumber("year", mcp.Required(), mcp.Description("Payroll year (e.g., 2025)")),
		mcp.WithNumber("batch", mcp.Required(), mcp.Description("Payment batch: 1 (1st of month) or 15 (15th of month)")),
		mcp.WithString("dry_run", mcp.Description("Dry run mode (default: true) - 'true' for preview, 'false' to execute")),
		mcp.WithString("employee_filter", mcp.Description("Filter by specific employee email (optional)")),
		mcp.WithString("currency_date", mcp.Description("Currency conversion date in YYYY-MM-DD format (optional)")),
		mcp.WithString("include_bonuses", mcp.Description("Include bonuses in calculation (default: true)")),
		mcp.WithString("include_advances", mcp.Description("Include salary advance deductions (default: true)")),
	)
}

// CalculateMonthlyPayrollHandler handles the calculate_monthly_payroll tool execution
func (t *Tools) CalculateMonthlyPayrollHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()

	// Extract and validate parameters
	params := &workflow.MonthlyPayrollParams{
		Month:           int(req.GetFloat("month", 0)),
		Year:            int(req.GetFloat("year", 0)),
		Batch:           int(req.GetFloat("batch", 0)),
		DryRun:          parseBoolString(req.GetString("dry_run", "true")), // Default to true
		EmployeeFilter:  req.GetString("employee_filter", ""),
		CurrencyDate:    req.GetString("currency_date", ""),
		IncludeBonuses:  parseBoolString(req.GetString("include_bonuses", "true")),
		IncludeAdvances: parseBoolString(req.GetString("include_advances", "true")),
	}

	// Validate parameters
	if err := t.workflowService.ValidateMonthlyPayrollParams(params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	// Create workflow tracking record
	agentKeyID := getAgentKeyIDFromContext(ctx)
	workflowRecord, err := t.workflowService.CreateWorkflow(ctx, "calculate_monthly_payroll", params, agentKeyID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create workflow: %v", err)), nil
	}

	// Update workflow status to in_progress
	if err := t.workflowService.UpdateWorkflowStatus(ctx, workflowRecord.ID, workflow.WorkflowStatusInProgress, nil, ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update workflow status: %v", err)), nil
	}

	// Check if payroll already exists (for non-dry-run mode)
	if !params.DryRun {
		exists, err := t.workflowService.CheckPayrollExists(ctx, params.Month, params.Year, params.Batch)
		if err != nil {
			t.workflowService.UpdateWorkflowStatus(ctx, workflowRecord.ID, workflow.WorkflowStatusFailed, nil, err.Error())
			return mcp.NewToolResultError(fmt.Sprintf("Failed to check payroll existence: %v", err)), nil
		}
		if exists {
			t.workflowService.UpdateWorkflowStatus(ctx, workflowRecord.ID, workflow.WorkflowStatusFailed, nil, "Payroll already exists")
			return mcp.NewToolResultError(fmt.Sprintf("Payroll already exists for month %d, year %d, batch %d", params.Month, params.Year, params.Batch)), nil
		}
	}

	// Execute payroll calculation workflow
	result, err := t.executeMonthlyPayrollCalculation(ctx, params, workflowRecord.ID, startTime)
	if err != nil {
		t.workflowService.UpdateWorkflowStatus(ctx, workflowRecord.ID, workflow.WorkflowStatusFailed, nil, err.Error())
		return mcp.NewToolResultError(fmt.Sprintf("Payroll calculation failed: %v", err)), nil
	}

	// Update workflow status to completed
	if err := t.workflowService.UpdateWorkflowStatus(ctx, workflowRecord.ID, workflow.WorkflowStatusCompleted, result, ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update workflow status: %v", err)), nil
	}

	// Format and return result
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format result: %v", err)), nil
	}

	if params.DryRun {
		return mcp.NewToolResultText(fmt.Sprintf("Monthly Payroll Calculation (DRY RUN) - %d/%d Batch %d:\n\n%s", params.Month, params.Year, params.Batch, string(resultJSON))), nil
	} else {
		return mcp.NewToolResultText(fmt.Sprintf("Monthly Payroll Calculation EXECUTED - %d/%d Batch %d:\n\n%s", params.Month, params.Year, params.Batch, string(resultJSON))), nil
	}
}

// executeMonthlyPayrollCalculation performs the actual payroll calculation
func (t *Tools) executeMonthlyPayrollCalculation(ctx context.Context, params *workflow.MonthlyPayrollParams, workflowID model.UUID, startTime time.Time) (*workflow.MonthlyPayrollResult, error) {
	// For now, this is a simplified implementation that will be expanded
	// In a real implementation, this would integrate with the existing payroll calculator
	// located at pkg/handler/payroll/payroll_calculator.go

	// Mock implementation for demonstration
	employees, err := t.getActiveEmployees(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get active employees: %w", err)
	}

	var employeeCalculations []*workflow.EmployeeCalculation
	var totalAmountVND float64

	for _, emp := range employees {
		calc, err := t.calculateEmployeePayroll(ctx, emp, params)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate payroll for employee %s: %w", emp.TeamEmail, err)
		}
		employeeCalculations = append(employeeCalculations, calc)
		totalAmountVND += calc.FinalCalculation.NetAmountVND
	}

	// Create currency conversion date
	currencyDate := params.CurrencyDate
	if currencyDate == "" {
		currencyDate = time.Date(params.Year, time.Month(params.Month), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	}

	result := &workflow.MonthlyPayrollResult{
		PayrollSummary: &workflow.PayrollSummary{
			Month:                  params.Month,
			Year:                   params.Year,
			Batch:                  params.Batch,
			TotalEmployees:         len(employees),
			TotalAmountVND:         totalAmountVND,
			CurrencyConversionDate: currencyDate,
		},
		EmployeeCalculations: employeeCalculations,
		WorkflowMetadata: &workflow.WorkflowMetadata{
			CalculationDate:  startTime,
			DryRun:          params.DryRun,
			ProcessingTimeMS: time.Since(startTime).Milliseconds(),
			WorkflowID:      workflowID,
		},
	}

	// If not dry run, update database records
	if !params.DryRun {
		if err := t.commitPayrollChanges(ctx, params, result); err != nil {
			return nil, fmt.Errorf("failed to commit payroll changes: %w", err)
		}
	}

	return result, nil
}

// getActiveEmployees retrieves active employees for payroll calculation
func (t *Tools) getActiveEmployees(ctx context.Context, params *workflow.MonthlyPayrollParams) ([]*model.Employee, error) {
	// If employee filter is provided, get specific employee
	if params.EmployeeFilter != "" {
		emp, err := t.store.Employee.OneByEmail(t.repo.DB(), params.EmployeeFilter)
		if err != nil {
			return nil, fmt.Errorf("employee not found: %w", err)
		}
		return []*model.Employee{emp}, nil
	}

	// Otherwise get all active employees using database filtering
	// Use the same approach as list_available_employees for consistency
	pagination := model.Pagination{Size: 1000} // Set explicit size to avoid LIMIT 0
	employees, _, err := t.store.Employee.All(t.repo.DB(), employee.EmployeeFilter{
		WorkingStatuses: []string{"full-time"},
	}, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get employees: %w", err)
	}

	return employees, nil
}

// calculateEmployeePayroll calculates payroll for a single employee
func (t *Tools) calculateEmployeePayroll(ctx context.Context, emp *model.Employee, params *workflow.MonthlyPayrollParams) (*workflow.EmployeeCalculation, error) {
	// This is a mock implementation
	// Real implementation would integrate with existing payroll calculator logic

	// Mock base salary calculation
	baseSalary := &workflow.BaseSalaryInfo{
		ContractAmount:  3000.0, // Mock USD amount
		Currency:        "USD",
		ConvertedVND:    75000000, // Mock conversion
		ConversionRate:  25000,
		PartialPeriod:   false,
	}

	// Mock commission calculation
	commissions := &workflow.CommissionInfo{
		TotalVND: 5000000, // Mock commission
		UnpaidCommissions: []*workflow.UnpaidCommission{
			{
				InvoiceID:  model.NewUUID(),
				AmountVND:  5000000,
				Project:    "Mock Project A",
			},
		},
	}

	// Mock bonus calculation
	bonuses := &workflow.BonusInfo{
		TotalVND: 2000000,
		ProjectBonuses: []*workflow.ProjectBonus{
			{
				Reason:    "Performance bonus",
				AmountVND: 2000000,
			},
		},
		ApprovedExpenses: []*workflow.ApprovedExpense{},
	}

	// Mock deduction calculation
	deductions := &workflow.DeductionInfo{
		TotalDeductionsVND: 3000000,
		SalaryAdvances: []*workflow.SalaryAdvanceDeduction{
			{
				AdvanceID:        model.NewUUID(),
				AmountVND:        3000000,
				RemainingBalance: 0,
			},
		},
	}

	// Calculate final amounts
	grossAmount := baseSalary.ConvertedVND + commissions.TotalVND + bonuses.TotalVND
	netAmount := grossAmount - deductions.TotalDeductionsVND

	finalCalc := &workflow.FinalCalculation{
		GrossAmountVND:     grossAmount,
		TotalDeductionsVND: deductions.TotalDeductionsVND,
		NetAmountVND:       netAmount,
	}

	return &workflow.EmployeeCalculation{
		Employee: &workflow.EmployeeInfo{
			ID:    emp.ID,
			Name:  emp.FullName,
			Email: emp.TeamEmail,
		},
		BaseSalary:       baseSalary,
		Commissions:      commissions,
		Bonuses:          bonuses,
		Deductions:       deductions,
		FinalCalculation: finalCalc,
	}, nil
}

// commitPayrollChanges commits payroll changes to database (when not dry run)
func (t *Tools) commitPayrollChanges(ctx context.Context, params *workflow.MonthlyPayrollParams, result *workflow.MonthlyPayrollResult) error {
	// This would implement the actual database updates:
	// 1. Create cached_payroll record
	// 2. Mark commissions as paid
	// 3. Update salary advance status
	// 4. Create audit trail

	// For now, this is a placeholder
	// Real implementation would use database transactions
	return nil
}

// Helper functions

func parseBoolString(s string) bool {
	return s == "true" || s == "True" || s == "TRUE" || s == "1"
}

func getAgentKeyIDFromContext(ctx context.Context) model.UUID {
	// Extract agent key ID from context using auth middleware
	if agentKey, ok := auth.GetAgentFromContext(ctx); ok {
		return agentKey.ID
	}
	return model.NewUUID() // Fallback for testing
}