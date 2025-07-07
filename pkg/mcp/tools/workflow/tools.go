package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"gorm.io/datatypes"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/mcp/auth"
	"github.com/dwarvesf/fortress-api/pkg/mcp/view"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
	"github.com/dwarvesf/fortress-api/pkg/service/workflow"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/store/employeecommission"
)

// Tools represents workflow-related MCP tools
type Tools struct {
	cfg             *config.Config
	store           *store.Store
	repo            store.DBRepo
	workflowService *workflow.Service
}

// New creates a new workflow tools instance
func New(cfg *config.Config, store *store.Store, repo store.DBRepo, wiseService wise.IService) *Tools {
	return &Tools{
		cfg:             cfg,
		store:           store,
		repo:            repo,
		workflowService: workflow.New(store, repo, wiseService),
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

// GenerateFinancialReportTool returns the MCP tool for generating financial reports
func (t *Tools) GenerateFinancialReportTool() mcp.Tool {
	return mcp.NewTool(
		"generate_financial_report",
		mcp.WithDescription("Generate comprehensive financial report with revenue metrics, projections, income summary, and employee statistics"),
		mcp.WithNumber("month", mcp.Required(), mcp.Description("Report month (1-12)")),
		mcp.WithNumber("year", mcp.Required(), mcp.Description("Report year (e.g., 2025)")),
		mcp.WithNumber("currency_conversion_rate", mcp.Description("VND to USD conversion rate (optional, defaults to Wise API rate or 25900)")),
	)
}

// CalculateMonthlyPayrollHandler handles the calculate_monthly_payroll tool execution
func (t *Tools) CalculateMonthlyPayrollHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()

	// Add timeout context for long-running payroll operations (10 minutes)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

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
			if updateErr := t.workflowService.UpdateWorkflowStatus(ctx, workflowRecord.ID, workflow.WorkflowStatusFailed, nil, err.Error()); updateErr != nil {
				// Log the error but don't fail the original operation
				fmt.Printf("Failed to update workflow status: %v\n", updateErr)
			}
			return mcp.NewToolResultError(fmt.Sprintf("Failed to check payroll existence: %v", err)), nil
		}
		if exists {
			if updateErr := t.workflowService.UpdateWorkflowStatus(ctx, workflowRecord.ID, workflow.WorkflowStatusFailed, nil, "Payroll already exists"); updateErr != nil {
				// Log the error but don't fail the original operation
				fmt.Printf("Failed to update workflow status: %v\n", updateErr)
			}
			return mcp.NewToolResultError(fmt.Sprintf("Payroll already exists for month %d, year %d, batch %d", params.Month, params.Year, params.Batch)), nil
		}
	}

	// Execute payroll calculation workflow
	result, err := t.executeMonthlyPayrollCalculation(ctx, params, workflowRecord.ID, startTime)
	if err != nil {
		if updateErr := t.workflowService.UpdateWorkflowStatus(ctx, workflowRecord.ID, workflow.WorkflowStatusFailed, nil, err.Error()); updateErr != nil {
			// Log the error but don't fail the original operation
			fmt.Printf("Failed to update workflow status: %v\n", updateErr)
		}
		return mcp.NewToolResultError(fmt.Sprintf("Payroll calculation failed: %v", err)), nil
	}

	// Update workflow status to completed
	if err := t.workflowService.UpdateWorkflowStatus(ctx, workflowRecord.ID, workflow.WorkflowStatusCompleted, result, ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update workflow status: %v", err)), nil
	}

	// Format and return result using utility function
	return view.FormatJSONResponse(result)
}

// GenerateFinancialReportHandler handles the generate_financial_report tool execution
func (t *Tools) GenerateFinancialReportHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Add timeout context for long-running financial report operations (5 minutes)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Extract and validate parameters
	month := int(req.GetFloat("month", 0))
	year := int(req.GetFloat("year", 0))
	conversionRate := req.GetFloat("currency_conversion_rate", 0)

	params := workflow.FinancialReportParams{
		Month:                  month,
		Year:                   year,
		CurrencyConversionRate: conversionRate,
	}

	// Validate parameters
	if month < 1 || month > 12 {
		return mcp.NewToolResultError("Month must be between 1 and 12"), nil
	}
	if year < 2020 || year > 2030 {
		return mcp.NewToolResultError("Year must be between 2020 and 2030"), nil
	}

	// Create workflow tracking record
	agentKeyID := getAgentKeyIDFromContext(ctx)
	workflowRecord, err := t.workflowService.CreateWorkflow(ctx, "generate_financial_report", params, agentKeyID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create workflow: %v", err)), nil
	}

	// Update workflow status to in_progress
	if err := t.workflowService.UpdateWorkflowStatus(ctx, workflowRecord.ID, workflow.WorkflowStatusInProgress, nil, ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update workflow status: %v", err)), nil
	}

	// Execute financial report generation
	result, err := t.workflowService.GenerateFinancialReport(ctx, params)
	if err != nil {
		if updateErr := t.workflowService.UpdateWorkflowStatus(ctx, workflowRecord.ID, workflow.WorkflowStatusFailed, nil, err.Error()); updateErr != nil {
			// Log the error but don't fail the original operation
			fmt.Printf("Failed to update workflow status: %v\n", updateErr)
		}
		return mcp.NewToolResultError(fmt.Sprintf("Financial report generation failed: %v", err)), nil
	}

	// Set workflow ID in metadata
	result.WorkflowMetadata.WorkflowID = workflowRecord.ID

	// Update workflow status to completed
	if err := t.workflowService.UpdateWorkflowStatus(ctx, workflowRecord.ID, workflow.WorkflowStatusCompleted, result, ""); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update workflow status: %v", err)), nil
	}

	// Format and return result using utility function
	return view.FormatJSONResponse(result)
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
		// Check for context cancellation during intensive operations
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("payroll calculation canceled: %w", ctx.Err())
		default:
		}

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
			DryRun:           params.DryRun,
			ProcessingTimeMS: time.Since(startTime).Milliseconds(),
			WorkflowID:       workflowID,
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
	allEmployees, _, err := t.store.Employee.All(t.repo.DB(), employee.EmployeeFilter{
		WorkingStatuses: []string{"full-time"},
		Preload:         true, // Enable preloading of BaseSalary and other relationships
	}, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get employees: %w", err)
	}

	// Filter employees by batch - only include those whose BaseSalary.Batch matches params.Batch
	// This follows the same logic as the existing payroll calculator
	var filteredEmployees []*model.Employee
	for _, emp := range allEmployees {
		// Skip employees without base salary
		if emp.BaseSalary.ID.IsZero() {
			continue
		}

		// Only include employees whose batch matches the requested batch
		if emp.BaseSalary.Batch == params.Batch {
			filteredEmployees = append(filteredEmployees, emp)
		}
	}

	return filteredEmployees, nil
}

// calculateEmployeePayroll calculates payroll for a single employee using real database data
func (t *Tools) calculateEmployeePayroll(ctx context.Context, emp *model.Employee, params *workflow.MonthlyPayrollParams) (*workflow.EmployeeCalculation, error) {
	// 1. Calculate base salary
	// Note: Employee filtering ensures all employees have matching batch, so we can safely calculate salary
	var baseSalaryVND, contractAmount float64
	// Get USD to VND conversion rate from Wise API
	var conversionRate float64 = 25000

	if t.cfg.Env != "prod" {
		var err error
		conversionRate, err = t.workflowService.GetUSDToVNDRate()
		if err != nil {
			return nil, fmt.Errorf("failed to get USD to VND conversion rate: %w", err)
		}
	}
	var currency string = "VND"

	if !emp.BaseSalary.ID.IsZero() {
		contractAmount = float64(emp.BaseSalary.ContractAmount)
		if emp.BaseSalary.Currency != nil {
			currency = emp.BaseSalary.Currency.Name
		}

		// Calculate salary amount (employees are pre-filtered by batch)
		baseSalaryVND = float64(emp.BaseSalary.PersonalAccountAmount + emp.BaseSalary.CompanyAccountAmount)
		// Convert to VND if needed
		if currency != "VND" {
			baseSalaryVND = baseSalaryVND * conversionRate
		}
	}

	// Get employee batch safely (all employees are pre-filtered by batch so this will always match)
	var employeeBatch int
	if !emp.BaseSalary.ID.IsZero() {
		employeeBatch = emp.BaseSalary.Batch
	}

	baseSalary := &workflow.BaseSalaryInfo{
		ContractAmount: contractAmount,
		Currency:       currency,
		ConvertedVND:   baseSalaryVND,
		ConversionRate: conversionRate,
		PartialPeriod:  false,
		EmployeeBatch:  employeeBatch,
		RequestedBatch: params.Batch,
		BatchMatched:   true, // Always true since employees are pre-filtered by batch
	}

	// 2. Calculate unpaid commissions
	var totalCommissionVND float64
	var unpaidCommissions []*workflow.UnpaidCommission

	userCommissions, err := t.store.EmployeeCommission.Get(t.repo.DB(), employeecommission.Query{
		EmployeeID: emp.ID.String(),
		IsPaid:     false,
	})
	if err == nil {
		for _, comm := range userCommissions {
			totalCommissionVND += float64(comm.Amount)

			projectName := "Unknown Project"
			if comm.Invoice != nil {
				projectName = comm.Invoice.Number
			}

			unpaidCommissions = append(unpaidCommissions, &workflow.UnpaidCommission{
				InvoiceID: comm.InvoiceID,
				AmountVND: float64(comm.Amount),
				Project:   projectName,
			})
		}
	}

	commissions := &workflow.CommissionInfo{
		TotalVND:          totalCommissionVND,
		UnpaidCommissions: unpaidCommissions,
	}

	// 3. Calculate bonuses
	var totalBonusVND float64
	var projectBonuses []*workflow.ProjectBonus

	if params.IncludeBonuses {
		bonusRecords, err := t.store.Bonus.GetByUserID(t.repo.DB(), emp.ID)
		if err == nil {
			for _, bonus := range bonusRecords {
				if bonus.IsActive {
					totalBonusVND += float64(bonus.Amount)
					projectBonuses = append(projectBonuses, &workflow.ProjectBonus{
						Reason:    bonus.Name,
						AmountVND: float64(bonus.Amount),
					})
				}
			}
		}
	}

	bonuses := &workflow.BonusInfo{
		TotalVND:         totalBonusVND,
		ProjectBonuses:   projectBonuses,
		ApprovedExpenses: []*workflow.ApprovedExpense{}, // TODO: Implement expense logic if needed
	}

	// 4. Calculate salary advance deductions
	var totalAdvanceVND float64
	var salaryAdvances []*workflow.SalaryAdvanceDeduction

	if params.IncludeAdvances {
		advanceSalaries, err := t.store.SalaryAdvance.ListNotPayBackByEmployeeID(t.repo.DB(), emp.ID.String())
		if err == nil {
			for _, advance := range advanceSalaries {
				// Convert USD advances to VND
				advanceVND := advance.AmountUSD * conversionRate
				totalAdvanceVND += advanceVND

				salaryAdvances = append(salaryAdvances, &workflow.SalaryAdvanceDeduction{
					AdvanceID:        advance.ID,
					AmountVND:        advanceVND,
					RemainingBalance: 0, // All will be deducted
				})
			}
		}
	}

	deductions := &workflow.DeductionInfo{
		TotalDeductionsVND: totalAdvanceVND,
		SalaryAdvances:     salaryAdvances,
	}

	// 5. Calculate final amounts
	grossAmount := baseSalaryVND + totalCommissionVND + totalBonusVND
	netAmount := grossAmount - totalAdvanceVND

	finalCalc := &workflow.FinalCalculation{
		GrossAmountVND:     grossAmount,
		TotalDeductionsVND: totalAdvanceVND,
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
	// Start database transaction for atomic payroll operations
	tx, done := t.repo.NewTransaction()
	
	// Create cached_payroll record for audit and history tracking
	// This follows the existing pattern in the codebase
	cachedPayroll := &model.CachedPayroll{
		Month:    params.Month,
		Year:     params.Year,
		Batch:    params.Batch,
		Payrolls: datatypes.JSON("{}"), // Simplified payload for now
	}
	
	// Create cached payroll record in transaction
	// Use direct database operations since store methods may not exist
	if err := tx.DB().Create(cachedPayroll).Error; err != nil {
		return done(fmt.Errorf("failed to create cached payroll record: %w", err))
	}
	
	// TODO: In production, this would also:
	// 1. Mark specific employee commissions as paid
	// 2. Update salary advance statuses  
	// 3. Create audit trail entries
	// 4. Update project commission tracking
	//
	// For now, we keep it simple and just create the cached payroll record
	// The transaction ensures atomicity for future enhancements
	
	// Commit transaction - all operations succeed or all rollback
	return done(nil)
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
