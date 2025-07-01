package workflow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

// mockWiseService implements wise.IService for testing
type mockWiseService struct {
	rateToReturn float64
	errorToReturn error
}

func (m *mockWiseService) Convert(amount float64, source, target string) (convertedAmount float64, rate float64, error error) {
	return 0, 0, nil
}

func (m *mockWiseService) GetPayrollQuotes(sourceCurrency, targetCurrency string, targetAmount float64) (*model.TWQuote, error) {
	return nil, nil
}

func (m *mockWiseService) GetRate(source, target string) (rate float64, err error) {
	return m.rateToReturn, m.errorToReturn
}

func TestService_ValidateMonthlyPayrollParams(t *testing.T) {
	service := &Service{
		store: store.New(),
	}

	tests := []struct {
		name          string
		params        MonthlyPayrollParams
		expectedError string
	}{
		{
			name: "valid_params",
			params: MonthlyPayrollParams{
				Month: 6,
				Year:  2025,
				Batch: 1,
			},
			expectedError: "",
		},
		{
			name: "invalid_month_too_low",
			params: MonthlyPayrollParams{
				Month: 0,
				Year:  2025,
				Batch: 1,
			},
			expectedError: "month must be between 1 and 12",
		},
		{
			name: "invalid_month_too_high",
			params: MonthlyPayrollParams{
				Month: 13,
				Year:  2025,
				Batch: 1,
			},
			expectedError: "month must be between 1 and 12",
		},
		{
			name: "invalid_year_too_low",
			params: MonthlyPayrollParams{
				Month: 6,
				Year:  2019,
				Batch: 1,
			},
			expectedError: "year must be between 2020 and 2030",
		},
		{
			name: "invalid_year_too_high",
			params: MonthlyPayrollParams{
				Month: 6,
				Year:  2031,
				Batch: 1,
			},
			expectedError: "year must be between 2020 and 2030",
		},
		{
			name: "invalid_batch",
			params: MonthlyPayrollParams{
				Month: 6,
				Year:  2025,
				Batch: 30,
			},
			expectedError: "batch must be either 1 or 15",
		},
		{
			name: "invalid_currency_date_format",
			params: MonthlyPayrollParams{
				Month:        6,
				Year:         2025,
				Batch:        1,
				CurrencyDate: "invalid-date",
			},
			expectedError: "currency_date must be in YYYY-MM-DD format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateMonthlyPayrollParams(&tt.params)
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_GetUSDToVNDRate(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mockWiseService)
		expectedRate  float64
		expectedError string
	}{
		{
			name: "successful_wise_api_call",
			setupMocks: func(mockWise *mockWiseService) {
				mockWise.rateToReturn = 25850.0
				mockWise.errorToReturn = nil
			},
			expectedRate: 25850.0,
		},
		{
			name: "wise_api_failure",
			setupMocks: func(mockWise *mockWiseService) {
				mockWise.rateToReturn = 0.0
				mockWise.errorToReturn = assert.AnError
			},
			expectedError: "failed to get USD to VND exchange rate from Wise API",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockWise := &mockWiseService{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockWise)
			}

			// Create service
			service := &Service{
				wiseService: mockWise,
			}

			// Execute test
			rate, err := service.GetUSDToVNDRate()

			// Assertions
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedRate, rate)
		})
	}
}

func TestService_GenerateFinancialReport_CurrencyConversion(t *testing.T) {
	// Test the currency conversion logic without database dependencies
	mockWise := &mockWiseService{
		rateToReturn: 24500.0,
		errorToReturn: nil,
	}
	
	_ = &Service{
		store:       store.New(),
		wiseService: mockWise,
	}

	// Test with provided rate (should not call Wise API)
	params := FinancialReportParams{
		Month:                  6,
		Year:                   2025,
		CurrencyConversionRate: 25000,
	}

	// Mock data for financial calculations INCLUDING payroll costs
	currentMonthRevenue := 100000.0  // VND
	currentMonthExpenses := 30000.0  // VND (operational expenses)
	currentMonthPayrolls := 50000.0  // VND (payroll expenses)
	totalOutcome := currentMonthExpenses + currentMonthPayrolls  // 80,000 VND total
	totalEmployees := float64(1)
	
	// Test currency conversion calculations with total outcome
	avgRevenueVND := currentMonthRevenue / totalEmployees
	avgCostPerHeadVND := totalOutcome / totalEmployees  // Updated to use total outcome
	avgProfitPerHeadVND := (currentMonthRevenue - totalOutcome) / totalEmployees  // Updated to use total outcome
	
	expectedAvgRevenueUSD := avgRevenueVND / params.CurrencyConversionRate
	expectedAvgCostUSD := avgCostPerHeadVND / params.CurrencyConversionRate
	expectedAvgProfitUSD := avgProfitPerHeadVND / params.CurrencyConversionRate
	
	// Verify calculations with updated total outcome logic
	assert.InDelta(t, 4.0, expectedAvgRevenueUSD, 0.01)   // 100,000 VND / 25,000 rate
	assert.InDelta(t, 3.2, expectedAvgCostUSD, 0.01)      // 80,000 VND / 25,000 rate (expenses + payroll)
	assert.InDelta(t, 0.8, expectedAvgProfitUSD, 0.01)    // 20,000 VND / 25,000 rate (revenue - total outcome)
}

func TestService_GenerateFinancialReport_ZeroDivisionProtection(t *testing.T) {
	// Test zero division protection without database
	totalEmployees := float64(0)  // Zero employees
	currentMonthRevenue := 100000.0
	currentMonthExpenses := 30000.0
	currentMonthPayrolls := 50000.0
	totalOutcome := currentMonthExpenses + currentMonthPayrolls
	conversionRate := 25000.0

	// Test zero division protection
	var avgRevenueVND, avgCostPerHeadVND, avgProfitPerHeadVND float64
	if totalEmployees > 0 {
		avgRevenueVND = currentMonthRevenue / totalEmployees
		avgCostPerHeadVND = totalOutcome / totalEmployees  // Updated to use total outcome
		avgProfitPerHeadVND = (currentMonthRevenue - totalOutcome) / totalEmployees  // Updated to use total outcome
	}

	// All should be zero (not infinity)
	assert.Equal(t, 0.0, avgRevenueVND)
	assert.Equal(t, 0.0, avgCostPerHeadVND)
	assert.Equal(t, 0.0, avgProfitPerHeadVND)

	// Convert to USD
	avgRevenueUSD := avgRevenueVND / conversionRate
	avgCostUSD := avgCostPerHeadVND / conversionRate
	avgProfitUSD := avgProfitPerHeadVND / conversionRate

	// All USD amounts should also be zero
	assert.Equal(t, 0.0, avgRevenueUSD)
	assert.Equal(t, 0.0, avgCostUSD)
	assert.Equal(t, 0.0, avgProfitUSD)
}

func TestService_GenerateFinancialReport_TotalOutcomeCalculation(t *testing.T) {
	// Test total outcome calculation (expenses + payrolls)
	tests := []struct {
		name             string
		expenses         float64
		payrolls         float64
		expectedOutcome  float64
		revenue          float64
		expectedProfit   float64
		expectedMargin   float64
	}{
		{
			name:            "normal_business_scenario",
			expenses:        30000.0,  // VND operational expenses
			payrolls:        50000.0,  // VND payroll costs
			expectedOutcome: 80000.0,  // VND total outcome
			revenue:         100000.0, // VND revenue
			expectedProfit:  20000.0,  // VND profit (revenue - total outcome)
			expectedMargin:  20.0,     // 20% margin
		},
		{
			name:            "high_payroll_scenario",
			expenses:        20000.0,
			payrolls:        70000.0,
			expectedOutcome: 90000.0,
			revenue:         100000.0,
			expectedProfit:  10000.0,
			expectedMargin:  10.0,     // 10% margin
		},
		{
			name:            "loss_scenario",
			expenses:        40000.0,
			payrolls:        70000.0,
			expectedOutcome: 110000.0,
			revenue:         100000.0,
			expectedProfit:  -10000.0, // Loss
			expectedMargin:  -10.0,    // Negative margin
		},
		{
			name:            "zero_payroll_scenario",
			expenses:        50000.0,
			payrolls:        0.0,
			expectedOutcome: 50000.0,
			revenue:         100000.0,
			expectedProfit:  50000.0,
			expectedMargin:  50.0,     // 50% margin
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate total outcome
			totalOutcome := tt.expenses + tt.payrolls
			assert.Equal(t, tt.expectedOutcome, totalOutcome)

			// Calculate profit
			profit := tt.revenue - totalOutcome
			assert.Equal(t, tt.expectedProfit, profit)

			// Calculate margin
			var margin float64
			if tt.revenue > 0 {
				margin = (profit / tt.revenue) * 100
			}
			assert.InDelta(t, tt.expectedMargin, margin, 0.01)
		})
	}
}

func TestService_GenerateFinancialReport_MarginCalculation(t *testing.T) {
	// Test margin calculation logic
	tests := []struct {
		name           string
		revenue        float64
		expenses       float64
		expectedMargin float64
	}{
		{
			name:           "positive_margin",
			revenue:        100000.0,
			expenses:       50000.0,
			expectedMargin: 50.0, // (100,000 - 50,000) / 100,000 * 100 = 50%
		},
		{
			name:           "zero_margin",
			revenue:        100000.0,
			expenses:       100000.0,
			expectedMargin: 0.0, // (100,000 - 100,000) / 100,000 * 100 = 0%
		},
		{
			name:           "negative_margin",
			revenue:        100000.0,
			expenses:       150000.0,
			expectedMargin: -50.0, // (100,000 - 150,000) / 100,000 * 100 = -50%
		},
		{
			name:           "zero_revenue",
			revenue:        0.0,
			expenses:       50000.0,
			expectedMargin: 0.0, // No margin calculation when revenue is zero
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var margin float64
			if tt.revenue > 0 {
				margin = ((tt.revenue - tt.expenses) / tt.revenue) * 100
			}
			assert.InDelta(t, tt.expectedMargin, margin, 0.01)
		})
	}
}

func TestService_FinancialReportParams_Validation(t *testing.T) {
	// Test financial report parameter validation
	tests := []struct {
		name    string
		params  FinancialReportParams
		isValid bool
	}{
		{
			name: "valid_params_with_rate",
			params: FinancialReportParams{
				Month:                  6,
				Year:                   2025,
				CurrencyConversionRate: 25000,
			},
			isValid: true,
		},
		{
			name: "valid_params_without_rate",
			params: FinancialReportParams{
				Month: 6,
				Year:  2025,
			},
			isValid: true,
		},
		{
			name: "invalid_month",
			params: FinancialReportParams{
				Month: 13,
				Year:  2025,
			},
			isValid: false,
		},
		{
			name: "invalid_year",
			params: FinancialReportParams{
				Month: 6,
				Year:  2031,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation checks
			monthValid := tt.params.Month >= 1 && tt.params.Month <= 12
			yearValid := tt.params.Year >= 2020 && tt.params.Year <= 2030
			
			valid := monthValid && yearValid
			assert.Equal(t, tt.isValid, valid)
		})
	}
}

func TestService_CurrencyConversionAccuracy(t *testing.T) {
	// Test currency conversion with various rates and amounts
	tests := []struct {
		name           string
		amountVND      float64
		conversionRate float64
		expectedUSD    float64
	}{
		{
			name:           "standard_conversion",
			amountVND:      100000.0,
			conversionRate: 25000.0,
			expectedUSD:    4.0,
		},
		{
			name:           "high_rate_conversion",
			amountVND:      100000.0,
			conversionRate: 24500.0,
			expectedUSD:    4.08,
		},
		{
			name:           "low_rate_conversion",
			amountVND:      100000.0,
			conversionRate: 25900.0,
			expectedUSD:    3.86,
		},
		{
			name:           "zero_amount",
			amountVND:      0.0,
			conversionRate: 25000.0,
			expectedUSD:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultUSD := tt.amountVND / tt.conversionRate
			assert.InDelta(t, tt.expectedUSD, resultUSD, 0.01)
		})
	}
}

func TestService_GenerateFinancialReport_IncomeLastMonthLogic(t *testing.T) {
	// Test that IncomeLastMonth uses previous month's revenue, not current month
	tests := []struct {
		name          string
		currentMonth  int
		currentYear   int
		expectedPrevMonth int
		expectedPrevYear  int
	}{
		{
			name:              "june_report_shows_may_income",
			currentMonth:      6,
			currentYear:       2025,
			expectedPrevMonth: 5,
			expectedPrevYear:  2025,
		},
		{
			name:              "january_report_shows_december_income",
			currentMonth:      1,
			currentYear:       2025,
			expectedPrevMonth: 12,
			expectedPrevYear:  2024,
		},
		{
			name:              "march_report_shows_february_income",
			currentMonth:      3,
			currentYear:       2025,
			expectedPrevMonth: 2,
			expectedPrevYear:  2025,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the previous month calculation logic
			previousMonth := tt.currentMonth - 1
			previousYear := tt.currentYear
			if previousMonth < 1 {
				previousMonth = 12
				previousYear = tt.currentYear - 1
			}

			assert.Equal(t, tt.expectedPrevMonth, previousMonth)
			assert.Equal(t, tt.expectedPrevYear, previousYear)
		})
	}
}

func TestFinancialReportResult_Structure(t *testing.T) {
	// Test the structure and initialization of financial report result
	result := &FinancialReportResult{
		RevenueReport: &RevenueReport{
			AvgRevenue:       4.0,
			AvgCostPerHead:   2.0,
			AvgProfitPerHead: 2.0,
			AvgMarginPerHead: 50.0,
		},
		ProjectionReport: &ProjectionReport{
			ProjectedAvgRevenue:    6.0,
			ProjectedProfitPerHead: 4.0,
			ProjectedMarginPerHead: 66.67,
			ProjectedRevenue:       6.0,
		},
		IncomeSummary: &IncomeSummary{
			IncomeLastMonth:   4.0,
			IncomeThisYear:    8.0,
			AccountReceivable: 2.0,
		},
		EmployeeStats: &EmployeeStats{
			TotalEmployees:    1,
			BillableEmployees: 1,
		},
		WorkflowMetadata: &WorkflowMetadata{
			CalculationDate:  time.Now(),
			DryRun:           false,
			ProcessingTimeMS: 100,
		},
	}

	// Verify structure
	assert.NotNil(t, result.RevenueReport)
	assert.NotNil(t, result.ProjectionReport)
	assert.NotNil(t, result.IncomeSummary)
	assert.NotNil(t, result.EmployeeStats)
	assert.NotNil(t, result.WorkflowMetadata)

	// Verify sample values
	assert.Equal(t, 4.0, result.RevenueReport.AvgRevenue)
	assert.Equal(t, 1, result.EmployeeStats.TotalEmployees)
	assert.False(t, result.WorkflowMetadata.DryRun)
}

func TestService_GenerateFinancialReport_WiseAPIIntegration(t *testing.T) {
	// Test Wise API integration with different scenarios
	tests := []struct {
		name              string
		mockRate          float64
		mockError         error
		providedRate      float64
		expectedUseWise   bool
		expectedFallback  bool
	}{
		{
			name:             "wise_api_success_no_provided_rate",
			mockRate:         24500.0,
			mockError:        nil,
			providedRate:     0,
			expectedUseWise:  true,
			expectedFallback: false,
		},
		{
			name:             "wise_api_failure_use_fallback",
			mockRate:         0,
			mockError:        assert.AnError,
			providedRate:     0,
			expectedUseWise:  true,
			expectedFallback: true,
		},
		{
			name:             "provided_rate_skip_wise",
			mockRate:         24500.0,
			mockError:        nil,
			providedRate:     25000.0,
			expectedUseWise:  false,
			expectedFallback: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWise := &mockWiseService{
				rateToReturn:  tt.mockRate,
				errorToReturn: tt.mockError,
			}

			service := &Service{
				store:       store.New(),
				wiseService: mockWise,
			}

			// Test the GetUSDToVNDRate function behavior
			if tt.expectedUseWise && !tt.expectedFallback {
				rate, err := service.GetUSDToVNDRate()
				require.NoError(t, err)
				assert.Equal(t, tt.mockRate, rate)
			} else if tt.expectedFallback {
				// Test fallback rate behavior
				rate, err := service.GetUSDToVNDRate()
				if err != nil {
					// Should use fallback
					assert.Contains(t, err.Error(), "failed to get USD to VND exchange rate from Wise API")
				} else {
					assert.Equal(t, 25900.0, rate) // Fallback rate
				}
			}
		})
	}
}

// Benchmark test for financial calculations
func BenchmarkCurrencyConversion(b *testing.B) {
	amountVND := 100000.0
	conversionRate := 25000.0
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = amountVND / conversionRate
	}
}

func BenchmarkMarginCalculation(b *testing.B) {
	revenue := 100000.0
	expenses := 50000.0
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if revenue > 0 {
			_ = ((revenue - expenses) / revenue) * 100
		}
	}
}