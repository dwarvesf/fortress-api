package workflow

import (
	"context"
	"fmt"
	"time"

	"encoding/json"

	"gorm.io/gorm"
	"gorm.io/datatypes"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
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
	ContractAmount  float64 `json:"contract_amount"`
	Currency        string  `json:"currency"`
	ConvertedVND    float64 `json:"converted_vnd"`
	ConversionRate  float64 `json:"conversion_rate"`
	PartialPeriod   bool    `json:"partial_period"`
	EmployeeBatch   int     `json:"employee_batch"`   // Employee's assigned batch (1 or 15)
	RequestedBatch  int     `json:"requested_batch"`  // Requested payroll batch
	BatchMatched    bool    `json:"batch_matched"`    // Whether batches match
}

// CommissionInfo represents commission calculation details
type CommissionInfo struct {
	TotalVND           float64                `json:"total_vnd"`
	UnpaidCommissions  []*UnpaidCommission    `json:"unpaid_commissions"`
}

// UnpaidCommission represents an unpaid commission
type UnpaidCommission struct {
	InvoiceID  model.UUID `json:"invoice_id"`
	AmountVND  float64    `json:"amount_vnd"`
	Project    string     `json:"project"`
}

// BonusInfo represents bonus calculation details
type BonusInfo struct {
	TotalVND         float64           `json:"total_vnd"`
	ProjectBonuses   []*ProjectBonus   `json:"project_bonuses"`
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
	GrossAmountVND      float64 `json:"gross_amount_vnd"`
	TotalDeductionsVND  float64 `json:"total_deductions_vnd"`
	NetAmountVND        float64 `json:"net_amount_vnd"`
}

// WorkflowMetadata represents metadata about the workflow execution
type WorkflowMetadata struct {
	CalculationDate  time.Time `json:"calculation_date"`
	DryRun          bool      `json:"dry_run"`
	ProcessingTimeMS int64     `json:"processing_time_ms"`
	WorkflowID      model.UUID `json:"workflow_id,omitempty"`
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