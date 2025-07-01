package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

// Service represents the workflow service
type Service struct {
	store       *store.Store
	repo        store.DBRepo
	wiseService wise.IService
}

// New creates a new workflow service
func New(store *store.Store, repo store.DBRepo, wiseService wise.IService) *Service {
	return &Service{
		store:       store,
		repo:        repo,
		wiseService: wiseService,
	}
}

// WorkflowStatus represents the status of a workflow
type WorkflowStatus string

const (
	WorkflowStatusPending    WorkflowStatus = "pending"
	WorkflowStatusInProgress WorkflowStatus = "in_progress"
	WorkflowStatusCompleted  WorkflowStatus = "completed"
	WorkflowStatusFailed     WorkflowStatus = "failed"
	WorkflowStatusRolledBack WorkflowStatus = "rolled_back"
)

// MonthlyPayrollParams represents parameters for monthly payroll calculation
type MonthlyPayrollParams struct {
	Month           int    `json:"month" binding:"required,min=1,max=12"`
	Year            int    `json:"year" binding:"required,min=2020,max=2030"`
	Batch           int    `json:"batch" binding:"required,oneof=1 15"`
	DryRun          bool   `json:"dry_run"`
	EmployeeFilter  string `json:"employee_filter,omitempty"`
	CurrencyDate    string `json:"currency_date,omitempty"`
	IncludeBonuses  bool   `json:"include_bonuses"`
	IncludeAdvances bool   `json:"include_advances"`
}

// MonthlyPayrollResult represents the result of monthly payroll calculation
type MonthlyPayrollResult struct {
	PayrollSummary       *PayrollSummary        `json:"payroll_summary"`
	EmployeeCalculations []*EmployeeCalculation `json:"employee_calculations"`
	WorkflowMetadata     *WorkflowMetadata      `json:"workflow_metadata"`
}

// PayrollSummary represents summary information for the payroll
type PayrollSummary struct {
	Month                  int     `json:"month"`
	Year                   int     `json:"year"`
	Batch                  int     `json:"batch"`
	TotalEmployees         int     `json:"total_employees"`
	TotalAmountVND         float64 `json:"total_amount_vnd"`
	CurrencyConversionDate string  `json:"currency_conversion_date"`
}

// EmployeeCalculation represents detailed calculation for an employee
type EmployeeCalculation struct {
	Employee         *EmployeeInfo     `json:"employee"`
	BaseSalary       *BaseSalaryInfo   `json:"base_salary"`
	Commissions      *CommissionInfo   `json:"commissions"`
	Bonuses          *BonusInfo        `json:"bonuses"`
	Deductions       *DeductionInfo    `json:"deductions"`
	FinalCalculation *FinalCalculation `json:"final_calculation"`
}

// EmployeeInfo represents basic employee information
type EmployeeInfo struct {
	ID    model.UUID `json:"id"`
	Name  string     `json:"name"`
	Email string     `json:"email"`
}

// BaseSalaryInfo represents base salary calculation details
type BaseSalaryInfo struct {
	ContractAmount float64 `json:"contract_amount"`
	Currency       string  `json:"currency"`
	ConvertedVND   float64 `json:"converted_vnd"`
	ConversionRate float64 `json:"conversion_rate"`
	PartialPeriod  bool    `json:"partial_period"`
	EmployeeBatch  int     `json:"employee_batch"`  // Employee's assigned batch (1 or 15)
	RequestedBatch int     `json:"requested_batch"` // Requested payroll batch
	BatchMatched   bool    `json:"batch_matched"`   // Whether batches match
}

// CommissionInfo represents commission calculation details
type CommissionInfo struct {
	TotalVND          float64             `json:"total_vnd"`
	UnpaidCommissions []*UnpaidCommission `json:"unpaid_commissions"`
}

// UnpaidCommission represents an unpaid commission
type UnpaidCommission struct {
	InvoiceID model.UUID `json:"invoice_id"`
	AmountVND float64    `json:"amount_vnd"`
	Project   string     `json:"project"`
}

// BonusInfo represents bonus calculation details
type BonusInfo struct {
	TotalVND         float64            `json:"total_vnd"`
	ProjectBonuses   []*ProjectBonus    `json:"project_bonuses"`
	ApprovedExpenses []*ApprovedExpense `json:"approved_expenses"`
}

// ProjectBonus represents a project-based bonus
type ProjectBonus struct {
	Reason    string  `json:"reason"`
	AmountVND float64 `json:"amount_vnd"`
}

// ApprovedExpense represents an approved expense
type ApprovedExpense struct {
	ExpenseID   string  `json:"expense_id"`
	Description string  `json:"description"`
	AmountVND   float64 `json:"amount_vnd"`
}

// DeductionInfo represents deduction calculation details
type DeductionInfo struct {
	SalaryAdvances     []*SalaryAdvanceDeduction `json:"salary_advances"`
	TotalDeductionsVND float64                   `json:"total_deductions_vnd"`
}

// SalaryAdvanceDeduction represents a salary advance deduction
type SalaryAdvanceDeduction struct {
	AdvanceID        model.UUID `json:"advance_id"`
	AmountVND        float64    `json:"amount_vnd"`
	RemainingBalance float64    `json:"remaining_balance"`
}

// FinalCalculation represents the final payroll calculation
type FinalCalculation struct {
	GrossAmountVND     float64 `json:"gross_amount_vnd"`
	TotalDeductionsVND float64 `json:"total_deductions_vnd"`
	NetAmountVND       float64 `json:"net_amount_vnd"`
}

// WorkflowMetadata represents metadata about the workflow execution
type WorkflowMetadata struct {
	CalculationDate  time.Time  `json:"calculation_date"`
	DryRun           bool       `json:"dry_run"`
	ProcessingTimeMS int64      `json:"processing_time_ms"`
	WorkflowID       model.UUID `json:"workflow_id,omitempty"`
}

// CreateWorkflow creates a new workflow record
func (s *Service) CreateWorkflow(ctx context.Context, workflowType string, inputData interface{}, agentKeyID model.UUID) (*model.AgentWorkflow, error) {
	// Convert inputData to JSON
	inputJSON, err := convertToJSON(inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert input data to JSON: %w", err)
	}

	totalSteps := 1 // For monthly payroll, this is a single-step workflow
	workflow := &model.AgentWorkflow{
		BaseModel: model.BaseModel{
			ID: model.NewUUID(),
		},
		WorkflowType: workflowType,
		Status:       model.AgentWorkflowStatusPending,
		InputData:    inputJSON,
		AgentKeyID:   agentKeyID,
		TotalSteps:   &totalSteps,
	}

	if err := s.store.AgentWorkflow.Create(s.repo.DB(), workflow); err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	return workflow, nil
}

// UpdateWorkflowStatus updates the status of a workflow
func (s *Service) UpdateWorkflowStatus(ctx context.Context, workflowID model.UUID, status WorkflowStatus, outputData interface{}, errorMessage string) error {
	updates := map[string]interface{}{
		"status": string(status),
	}

	if outputData != nil {
		outputJSON, err := convertToJSON(outputData)
		if err != nil {
			return fmt.Errorf("failed to convert output data to JSON: %w", err)
		}
		updates["output_data"] = outputJSON
	}

	if errorMessage != "" {
		updates["error_message"] = errorMessage
	}

	if status == WorkflowStatusCompleted || status == WorkflowStatusFailed {
		updates["steps_completed"] = 1
	}

	return s.store.AgentWorkflow.UpdateSelectedFieldsByID(s.repo.DB(), workflowID.String(), updates)
}

// ValidateMonthlyPayrollParams validates the parameters for monthly payroll calculation
func (s *Service) ValidateMonthlyPayrollParams(params *MonthlyPayrollParams) error {
	if params.Month < 1 || params.Month > 12 {
		return fmt.Errorf("month must be between 1 and 12")
	}

	if params.Year < 2020 || params.Year > 2030 {
		return fmt.Errorf("year must be between 2020 and 2030")
	}

	if params.Batch != 1 && params.Batch != 15 {
		return fmt.Errorf("batch must be either 1 or 15")
	}

	// Parse currency date if provided
	if params.CurrencyDate != "" {
		if _, err := time.Parse("2006-01-02", params.CurrencyDate); err != nil {
			return fmt.Errorf("currency_date must be in YYYY-MM-DD format: %v", err)
		}
	}

	// Validate employee email if provided
	if params.EmployeeFilter != "" {
		if _, err := s.store.Employee.OneByEmail(s.repo.DB(), params.EmployeeFilter); err != nil {
			return fmt.Errorf("employee not found: %v", err)
		}
	}

	return nil
}

// CheckPayrollExists checks if payroll already exists for the given month/year/batch
func (s *Service) CheckPayrollExists(ctx context.Context, month, year, batch int) (bool, error) {
	// Use the correct method signature for CachedPayroll store
	_, err := s.store.CachedPayroll.Get(s.repo.DB(), month, year, batch)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check payroll existence: %w", err)
	}
	return true, nil
}

// GetUSDToVNDRate gets the current USD to VND exchange rate using the Wise API
func (s *Service) GetUSDToVNDRate() (float64, error) {
	rate, err := s.wiseService.GetRate("USD", "VND")
	if err != nil {
		return 0, fmt.Errorf("failed to get USD to VND exchange rate from Wise API: %w", err)
	}
	return rate, nil
}

// FinancialReportParams represents parameters for financial report generation
type FinancialReportParams struct {
	Month                  int     `json:"month" binding:"required,min=1,max=12"`
	Year                   int     `json:"year" binding:"required,min=2020,max=2030"`
	CurrencyConversionRate float64 `json:"currency_conversion_rate,omitempty"`
}

// FinancialReportResult represents the result of financial report generation
type FinancialReportResult struct {
	RevenueReport    *RevenueReport    `json:"revenue_report"`
	ProjectionReport *ProjectionReport `json:"projection_report"`
	IncomeSummary    *IncomeSummary    `json:"income_summary"`
	EmployeeStats    *EmployeeStats    `json:"employee_stats"`
	WorkflowMetadata *WorkflowMetadata `json:"workflow_metadata"`
}

// RevenueReport represents current month metrics based on PAID invoices
type RevenueReport struct {
	AvgRevenue       float64 `json:"avg_revenue_usd"`
	AvgCostPerHead   float64 `json:"avg_cost_per_head_usd"`
	AvgProfitPerHead float64 `json:"avg_profit_per_head_usd"`
	AvgMarginPerHead float64 `json:"avg_margin_per_head_percent"`
}

// ProjectionReport represents current month metrics based on ALL invoices
type ProjectionReport struct {
	ProjectedAvgRevenue    float64 `json:"projected_avg_revenue_usd"`
	ProjectedProfitPerHead float64 `json:"projected_profit_per_head_usd"`
	ProjectedMarginPerHead float64 `json:"projected_margin_per_head_percent"`
	ProjectedRevenue       float64 `json:"projected_revenue_usd"`
}

// IncomeSummary represents income and receivables information
type IncomeSummary struct {
	IncomeLastMonth   float64 `json:"income_last_month_usd"`
	IncomeThisYear    float64 `json:"income_this_year_usd"`
	AccountReceivable float64 `json:"account_receivable_usd"`
}

// EmployeeStats represents employee utilization information
type EmployeeStats struct {
	TotalEmployees    int `json:"total_employees"`
	BillableEmployees int `json:"billable_employees"`
}

// GenerateFinancialReport generates comprehensive financial report for a given month/year
func (s *Service) GenerateFinancialReport(ctx context.Context, params FinancialReportParams) (*FinancialReportResult, error) {
	startTime := time.Now()

	// Default currency conversion rate if not provided
	conversionRate := params.CurrencyConversionRate
	if conversionRate == 0 {
		var err error
		conversionRate, err = s.GetUSDToVNDRate()
		if err != nil {
			// Fallback to default rate if Wise API fails
			conversionRate = 25900
		}
	}

	// 1. Get current month paid invoice revenue (Revenue Report)
	currentMonthRevenue, err := s.getCurrentMonthPaidRevenue(ctx, params.Month, params.Year)
	if err != nil {
		return nil, fmt.Errorf("failed to get current month paid revenue: %w", err)
	}

	// Debug: Log conversion rate and revenue for troubleshooting
	fmt.Printf("DEBUG: ConversionRate=%.2f, CurrentMonthRevenue=%.2f VND (%.2f USD)\n", conversionRate, currentMonthRevenue, currentMonthRevenue/conversionRate)

	// 2. Get current month total expenses
	currentMonthExpenses, err := s.getCurrentMonthExpenses(ctx, params.Month, params.Year)
	if err != nil {
		return nil, fmt.Errorf("failed to get current month expenses: %w", err)
	}

	// Debug: Log expenses for troubleshooting
	fmt.Printf("DEBUG: CurrentMonthExpenses=%.2f VND (%.2f USD)\n", currentMonthExpenses, currentMonthExpenses/conversionRate)

	// 2.5. Get current month payroll costs 
	currentMonthPayrolls, err := s.getCurrentMonthPayrolls(ctx, params.Month, params.Year)
	if err != nil {
		return nil, fmt.Errorf("failed to get current month payrolls: %w", err)
	}

	// Debug: Log payrolls for troubleshooting
	fmt.Printf("DEBUG: CurrentMonthPayrolls=%.2f VND (%.2f USD)\n", currentMonthPayrolls, currentMonthPayrolls/conversionRate)

	// Calculate total outcome (expenses + payrolls)
	totalOutcome := currentMonthExpenses + currentMonthPayrolls
	fmt.Printf("DEBUG: TotalOutcome=%.2f VND (%.2f USD)\n", totalOutcome, totalOutcome/conversionRate)

	// 3. Get current month ALL invoices (Projection Report)
	currentMonthAllInvoices, err := s.getCurrentMonthAllInvoices(ctx, params.Month, params.Year)
	if err != nil {
		return nil, fmt.Errorf("failed to get current month all invoices: %w", err)
	}

	// 4. Get YTD income
	ytdIncome, err := s.getYTDIncome(ctx, params.Year)
	if err != nil {
		return nil, fmt.Errorf("failed to get YTD income: %w", err)
	}

	// 5. Get outstanding receivables
	outstandingReceivables, err := s.getOutstandingReceivables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get outstanding receivables: %w", err)
	}

	// 6. Get employee statistics
	employeeStats, err := s.getEmployeeStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee stats: %w", err)
	}

	// 7. Get previous month revenue for Income Summary
	previousMonth := params.Month - 1
	previousYear := params.Year
	if previousMonth < 1 {
		previousMonth = 12
		previousYear = params.Year - 1
	}

	previousMonthRevenue, err := s.getCurrentMonthPaidRevenue(ctx, previousMonth, previousYear)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous month revenue: %w", err)
	}

	// Debug: Log previous month revenue for troubleshooting
	fmt.Printf("DEBUG: PreviousMonthRevenue (%d/%d)=%.2f VND (%.2f USD)\n", previousMonth, previousYear, previousMonthRevenue, previousMonthRevenue/conversionRate)

	// Calculate metrics (all in VND, then convert to USD for response)
	totalEmployees := float64(employeeStats.TotalEmployees)

	// Revenue Report calculations (based on PAID invoices) - calculate in VND then convert to USD
	// Handle zero division protection
	var avgRevenueVND, avgCostPerHeadVND, avgProfitPerHeadVND float64
	if totalEmployees > 0 {
		avgRevenueVND = currentMonthRevenue / totalEmployees
		avgCostPerHeadVND = totalOutcome / totalEmployees
		avgProfitPerHeadVND = (currentMonthRevenue - totalOutcome) / totalEmployees
	}

	revenueReport := &RevenueReport{
		AvgRevenue:       avgRevenueVND / conversionRate,       // Convert to USD
		AvgCostPerHead:   avgCostPerHeadVND / conversionRate,   // Convert to USD
		AvgProfitPerHead: avgProfitPerHeadVND / conversionRate, // Convert to USD
	}

	// Calculate margin per head (avoid division by zero) - percentage stays the same
	if currentMonthRevenue > 0 {
		revenueReport.AvgMarginPerHead = ((currentMonthRevenue - totalOutcome) / currentMonthRevenue) * 100
	}

	// Projection Report calculations (based on ALL invoices) - calculate in VND then convert to USD
	// Handle zero division protection
	var projectedAvgRevenueVND, projectedProfitPerHeadVND float64
	if totalEmployees > 0 {
		projectedAvgRevenueVND = currentMonthAllInvoices / totalEmployees
		projectedProfitPerHeadVND = (currentMonthAllInvoices - totalOutcome) / totalEmployees
	}

	projectionReport := &ProjectionReport{
		ProjectedAvgRevenue:    projectedAvgRevenueVND / conversionRate,    // Convert to USD
		ProjectedProfitPerHead: projectedProfitPerHeadVND / conversionRate, // Convert to USD
		ProjectedRevenue:       currentMonthAllInvoices / conversionRate,   // Convert to USD
	}

	// Calculate projected margin per head (avoid division by zero) - percentage stays the same
	if currentMonthAllInvoices > 0 {
		projectionReport.ProjectedMarginPerHead = ((currentMonthAllInvoices - totalOutcome) / currentMonthAllInvoices) * 100
	}

	// Income Summary - convert VND to USD
	incomeSummary := &IncomeSummary{
		IncomeLastMonth:   previousMonthRevenue / conversionRate,   // Convert to USD (Fixed: now uses previous month)
		IncomeThisYear:    ytdIncome / conversionRate,              // Convert to USD
		AccountReceivable: outstandingReceivables / conversionRate, // Convert to USD
	}

	// Workflow metadata
	workflowMetadata := &WorkflowMetadata{
		CalculationDate:  time.Now(),
		DryRun:           false, // Financial reports are always read-only
		ProcessingTimeMS: time.Since(startTime).Milliseconds(),
	}

	return &FinancialReportResult{
		RevenueReport:    revenueReport,
		ProjectionReport: projectionReport,
		IncomeSummary:    incomeSummary,
		EmployeeStats:    employeeStats,
		WorkflowMetadata: workflowMetadata,
	}, nil
}

// getCurrentMonthPaidRevenue gets revenue from paid invoices for the specified month/year (returns VND)
func (s *Service) getCurrentMonthPaidRevenue(ctx context.Context, month, year int) (float64, error) {
	// Query invoices that are PAID in the specified month/year
	var invoices []model.Invoice
	db := s.repo.DB()

	// Get paid invoices for the month/year
	err := db.Where("paid_at IS NOT NULL AND EXTRACT(MONTH FROM paid_at) = ? AND EXTRACT(YEAR FROM paid_at) = ?",
		month, year).Find(&invoices).Error
	if err != nil {
		return 0, fmt.Errorf("failed to query paid invoices: %w", err)
	}

	var totalRevenueVND float64
	for _, invoice := range invoices {
		// Use ConversionAmount which should be in VND
		if invoice.ConversionAmount > 0 {
			totalRevenueVND += invoice.ConversionAmount
		} else {
			// Fallback: if no ConversionAmount, assume Total is in VND
			totalRevenueVND += invoice.Total
		}
	}

	return totalRevenueVND, nil
}

// getCurrentMonthExpenses gets total expenses for the specified month/year (returns VND)
func (s *Service) getCurrentMonthExpenses(ctx context.Context, month, year int) (float64, error) {
	// Query accounting transactions for expense types (SE, OP, OV, CA)
	var transactions []model.AccountingTransaction
	db := s.repo.DB()

	// Get expense transactions for the month/year
	err := db.Where("type IN (?, ?, ?, ?) AND EXTRACT(MONTH FROM date) = ? AND EXTRACT(YEAR FROM date) = ?",
		model.AccountingSE, model.AccountingOP, model.AccountingOV, model.AccountingCA, month, year).Find(&transactions).Error
	if err != nil {
		return 0, fmt.Errorf("failed to query expense transactions: %w", err)
	}

	var totalExpensesVND float64
	for _, transaction := range transactions {
		// ConversionAmount is in VND as VietnamDong type
		// VietnamDong is stored as int64, so convert to float64
		vndAmount := float64(transaction.ConversionAmount)
		totalExpensesVND += vndAmount
	}

	return totalExpensesVND, nil
}

// getCurrentMonthAllInvoices gets revenue from ALL invoices for the specified month/year (returns VND)
func (s *Service) getCurrentMonthAllInvoices(ctx context.Context, month, year int) (float64, error) {
	// Query ALL invoices (regardless of status) created in the specified month/year
	var invoices []model.Invoice
	db := s.repo.DB()

	// Get all invoices for the month/year
	err := db.Where("EXTRACT(MONTH FROM created_at) = ? AND EXTRACT(YEAR FROM created_at) = ?",
		month, year).Find(&invoices).Error
	if err != nil {
		return 0, fmt.Errorf("failed to query all invoices: %w", err)
	}

	var totalRevenueVND float64
	for _, invoice := range invoices {
		// Use ConversionAmount which should be in VND
		if invoice.ConversionAmount > 0 {
			totalRevenueVND += invoice.ConversionAmount
		} else {
			// Fallback: if no ConversionAmount, assume Total is in VND
			totalRevenueVND += invoice.Total
		}
	}

	return totalRevenueVND, nil
}

// getYTDIncome gets year-to-date income for the specified year (returns VND)
func (s *Service) getYTDIncome(ctx context.Context, year int) (float64, error) {
	// Query income transactions (type = "In") for the year
	var transactions []model.AccountingTransaction
	db := s.repo.DB()

	// Get income transactions for the year
	err := db.Where("type = ? AND EXTRACT(YEAR FROM date) = ?",
		model.AccountingIncome, year).Find(&transactions).Error
	if err != nil {
		return 0, fmt.Errorf("failed to query income transactions: %w", err)
	}

	var totalIncomeVND float64
	for _, transaction := range transactions {
		// ConversionAmount is in VND as VietnamDong type
		// VietnamDong is stored as int64, so convert to float64
		vndAmount := float64(transaction.ConversionAmount)
		totalIncomeVND += vndAmount
	}

	return totalIncomeVND, nil
}

// getCurrentMonthPayrolls gets total payroll costs for given month/year (returns VND)
func (s *Service) getCurrentMonthPayrolls(ctx context.Context, month, year int) (float64, error) {
	var payrolls []model.Payroll
	db := s.repo.DB()
	
	err := db.Where("month = ? AND year = ? AND is_paid = ?", month, year, true).Find(&payrolls).Error
	if err != nil {
		return 0, fmt.Errorf("failed to query paid payrolls: %w", err)
	}
	
	var totalPayrollsVND float64
	for _, payroll := range payrolls {
		// Use ConversionAmount which should be in VND
		totalPayrollsVND += float64(payroll.ConversionAmount)
	}
	
	return totalPayrollsVND, nil
}

// getOutstandingReceivables gets total outstanding receivables (returns VND)
func (s *Service) getOutstandingReceivables(ctx context.Context) (float64, error) {
	// Query unpaid invoices (status != paid)
	var invoices []model.Invoice
	db := s.repo.DB()

	// Get unpaid invoices, just count from February 2025 onwards
	err := db.Where(`status != ? AND (year > 2025 OR (year = 2025 AND month >= 2))`, model.InvoiceStatusPaid).Find(&invoices).Error
	if err != nil {
		return 0, fmt.Errorf("failed to query unpaid invoices: %w", err)
	}

	var totalReceivablesVND float64
	for _, invoice := range invoices {
		// Use ConversionAmount which should be in VND
		if invoice.ConversionAmount > 0 {
			totalReceivablesVND += invoice.ConversionAmount
		} else {
			// Fallback: if no ConversionAmount, assume Total is in VND
			totalReceivablesVND += invoice.Total
		}
	}

	return totalReceivablesVND, nil
}

// getEmployeeStats gets employee statistics
func (s *Service) getEmployeeStats(ctx context.Context) (*EmployeeStats, error) {
	db := s.repo.DB()

	// Get total employees (active)
	var totalEmployees int64
	err := db.Model(&model.Employee{}).Where("deleted_at IS NULL AND working_status = ?",
		model.WorkingStatusFullTime).Count(&totalEmployees).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total employees: %w", err)
	}

	// Get billable employees (those with active projects)
	var billableEmployees int64
	err = db.Table("employees").
		Joins("JOIN project_members ON employees.id = project_members.employee_id").
		Joins("JOIN projects ON project_members.project_id = projects.id").
		Where("employees.deleted_at IS NULL AND employees.working_status = ? AND projects.status = ?",
			model.WorkingStatusFullTime, model.ProjectStatusActive).
		Distinct("employees.id").
		Count(&billableEmployees).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count billable employees: %w", err)
	}

	return &EmployeeStats{
		TotalEmployees:    int(totalEmployees),
		BillableEmployees: int(billableEmployees),
	}, nil
}

// Helper function to convert data to JSON
func convertToJSON(data interface{}) (datatypes.JSON, error) {
	if data == nil {
		return datatypes.JSON{}, nil
	}

	// Convert to JSON bytes and then to datatypes.JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return datatypes.JSON{}, err
	}

	var result datatypes.JSON
	if err := result.UnmarshalJSON(jsonBytes); err != nil {
		return datatypes.JSON{}, err
	}

	return result, nil
}
