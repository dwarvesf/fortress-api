package invoice

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
)

// Mock services
type mockContractorRatesService struct {
	mock.Mock
}

func (m *mockContractorRatesService) FetchContractorRateByPageID(ctx context.Context, pageID string) (*notion.ContractorRateData, error) {
	args := m.Called(ctx, pageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*notion.ContractorRateData), args.Error(1)
}

type mockTaskOrderLogService struct {
	mock.Mock
}

func (m *mockTaskOrderLogService) FetchTaskOrderHoursByPageID(ctx context.Context, pageID string) (float64, error) {
	args := m.Called(ctx, pageID)
	return args.Get(0).(float64), args.Error(1)
}

func TestHelper_FetchHourlyRateData(t *testing.T) {
	l := logger.NewLogrusLogger("debug")

	t.Run("Success Flow", func(t *testing.T) {
		ratesMock := new(mockContractorRatesService)
		taskMock := new(mockTaskOrderLogService)

		payout := notion.PayoutEntry{
			PageID:        "payout-1",
			ServiceRateID: "rate-1",
			TaskOrderID:   "task-1",
		}

		ratesMock.On("FetchContractorRateByPageID", mock.Anything, "rate-1").Return(&notion.ContractorRateData{
			BillingType: "Hourly Rate",
			HourlyRate:  50.0,
			Currency:    "USD",
		}, nil)

		taskMock.On("FetchTaskOrderHoursByPageID", mock.Anything, "task-1").Return(10.5, nil)

		result := fetchHourlyRateData(context.Background(), payout, ratesMock, taskMock, l)

		assert.NotNil(t, result)
		assert.Equal(t, 50.0, result.HourlyRate)
		assert.Equal(t, 10.5, result.Hours)
		assert.Equal(t, "Hourly Rate", result.BillingType)
		assert.Equal(t, "USD", result.Currency)
	})

	t.Run("Missing ServiceRateID", func(t *testing.T) {
		payout := notion.PayoutEntry{ServiceRateID: ""}
		result := fetchHourlyRateData(context.Background(), payout, nil, nil, l)
		assert.Nil(t, result)
	})

	t.Run("Rate Fetch Failure", func(t *testing.T) {
		ratesMock := new(mockContractorRatesService)
		payout := notion.PayoutEntry{ServiceRateID: "rate-fail"}

		ratesMock.On("FetchContractorRateByPageID", mock.Anything, "rate-fail").Return((*notion.ContractorRateData)(nil), errors.New("fail"))

		result := fetchHourlyRateData(context.Background(), payout, ratesMock, nil, l)
		assert.Nil(t, result)
	})

	t.Run("Non-Hourly Billing Type", func(t *testing.T) {
		ratesMock := new(mockContractorRatesService)
		payout := notion.PayoutEntry{ServiceRateID: "rate-monthly"}

		ratesMock.On("FetchContractorRateByPageID", mock.Anything, "rate-monthly").Return(&notion.ContractorRateData{
			BillingType: "Monthly Fixed",
		}, nil)

		result := fetchHourlyRateData(context.Background(), payout, ratesMock, nil, l)
		assert.Nil(t, result)
	})

	t.Run("Hours Fetch Failure (Graceful Degradation)", func(t *testing.T) {
		ratesMock := new(mockContractorRatesService)
		taskMock := new(mockTaskOrderLogService)

		payout := notion.PayoutEntry{
			ServiceRateID: "rate-1",
			TaskOrderID:   "task-fail",
		}

		ratesMock.On("FetchContractorRateByPageID", mock.Anything, "rate-1").Return(&notion.ContractorRateData{
			BillingType: "Hourly Rate",
			HourlyRate:  50.0,
		}, nil)

		taskMock.On("FetchTaskOrderHoursByPageID", mock.Anything, "task-fail").Return(0.0, errors.New("fail"))

		result := fetchHourlyRateData(context.Background(), payout, ratesMock, taskMock, l)

		assert.NotNil(t, result)
		assert.Equal(t, 0.0, result.Hours)
		assert.Equal(t, 50.0, result.HourlyRate)
	})
}

func TestHelper_AggregateHourlyServiceFees(t *testing.T) {
	l := logger.NewLogrusLogger("debug")

	t.Run("Multiple Hourly Items - Preserves First Item Title", func(t *testing.T) {
		items := []ContractorInvoiceLineItem{
			{IsHourlyRate: true, Hours: 10, Rate: 50, Amount: 500, AmountUSD: 500, Description: "Desc A", OriginalCurrency: "USD", Title: "Design Consulting Services Rendered (January 1-31, 2026)", TaskOrderID: "task-order-1"},
			{IsHourlyRate: true, Hours: 5, Rate: 50, Amount: 250, AmountUSD: 250, Description: "Desc B", OriginalCurrency: "USD", Title: "Design Consulting Services Rendered (January 1-31, 2026)", TaskOrderID: "task-order-2"},
		}

		result := aggregateHourlyServiceFees(items, "2026-01", l)

		assert.Len(t, result, 1)
		agg := result[0]
		assert.Equal(t, 15.0, agg.Hours)
		assert.Equal(t, 750.0, agg.Amount)
		assert.Equal(t, 750.0, agg.AmountUSD)
		assert.Equal(t, 50.0, agg.Rate)
		assert.Equal(t, "USD", agg.OriginalCurrency)
		assert.Contains(t, agg.Description, "Desc A")
		assert.Contains(t, agg.Description, "Desc B")
		// Title should be preserved from first item (payout description)
		assert.Equal(t, "Design Consulting Services Rendered (January 1-31, 2026)", agg.Title)
		// TaskOrderID should be preserved from first item
		assert.Equal(t, "task-order-1", agg.TaskOrderID)
	})

	t.Run("Multiple Hourly Items - Fallback Title When Empty", func(t *testing.T) {
		items := []ContractorInvoiceLineItem{
			{IsHourlyRate: true, Hours: 10, Rate: 50, Amount: 500, AmountUSD: 500, Description: "Desc A", OriginalCurrency: "USD"},
			{IsHourlyRate: true, Hours: 5, Rate: 50, Amount: 250, AmountUSD: 250, Description: "Desc B", OriginalCurrency: "USD"},
		}

		result := aggregateHourlyServiceFees(items, "2026-01", l)

		assert.Len(t, result, 1)
		agg := result[0]
		assert.Equal(t, 15.0, agg.Hours)
		assert.Equal(t, 750.0, agg.Amount)
		// When first item has no Title, fallback to generated title
		assert.Contains(t, agg.Title, "Service Fee")
	})

	t.Run("Mixed Items", func(t *testing.T) {
		items := []ContractorInvoiceLineItem{
			{IsHourlyRate: true, Hours: 10, Rate: 50, Amount: 500},
			{IsHourlyRate: false, Type: "Commission", Amount: 100},
		}

		result := aggregateHourlyServiceFees(items, "2026-01", l)

		assert.Len(t, result, 2)
		// Check types
		assert.Equal(t, "Commission", result[0].Type)
		assert.Equal(t, string(notion.PayoutSourceTypeServiceFee), result[1].Type)
	})
}

func TestHelper_GenerateServiceFeeTitle(t *testing.T) {
	assert.Equal(t, "Service Fee (Development work from 2026-01-01 to 2026-01-31)", generateServiceFeeTitle("2026-01"))
	assert.Equal(t, "Service Fee", generateServiceFeeTitle("invalid"))
}

func TestHelper_GenerateServiceFeeDescription(t *testing.T) {
	t.Run("Design Position", func(t *testing.T) {
		// Position contains "design" (case-insensitive)
		positions := []string{"Product Designer"}
		result := generateServiceFeeDescription("2025-12", positions)
		assert.Equal(t, "Design Consulting Services Rendered (December 1-31, 2025)", result)
	})

	t.Run("Design Position - Multiple positions including design", func(t *testing.T) {
		positions := []string{"Frontend", "UI Designer"}
		result := generateServiceFeeDescription("2025-12", positions)
		// Should match "design" in "UI Designer"
		assert.Equal(t, "Design Consulting Services Rendered (December 1-31, 2025)", result)
	})

	t.Run("Operation Executive Position", func(t *testing.T) {
		positions := []string{"Operation Executive"}
		result := generateServiceFeeDescription("2025-12", positions)
		assert.Equal(t, "Operational Consulting Services Rendered (December 1-31, 2025)", result)
	})

	t.Run("Operation Executive Position - Case Insensitive", func(t *testing.T) {
		positions := []string{"OPERATION EXECUTIVE"}
		result := generateServiceFeeDescription("2025-12", positions)
		assert.Equal(t, "Operational Consulting Services Rendered (December 1-31, 2025)", result)
	})

	t.Run("Software Development - Default", func(t *testing.T) {
		positions := []string{"Backend", "Frontend"}
		result := generateServiceFeeDescription("2025-12", positions)
		assert.Equal(t, "Software Development Services Rendered (December 1-31, 2025)", result)
	})

	t.Run("Empty Positions", func(t *testing.T) {
		positions := []string{}
		result := generateServiceFeeDescription("2025-12", positions)
		assert.Equal(t, "Software Development Services Rendered (December 1-31, 2025)", result)
	})

	t.Run("Nil Positions", func(t *testing.T) {
		result := generateServiceFeeDescription("2025-12", nil)
		assert.Equal(t, "Software Development Services Rendered (December 1-31, 2025)", result)
	})

	t.Run("Invalid Month Format", func(t *testing.T) {
		result := generateServiceFeeDescription("invalid", nil)
		assert.Equal(t, "Software Development Services Rendered", result)
	})

	t.Run("Different Month", func(t *testing.T) {
		positions := []string{"Backend"}
		result := generateServiceFeeDescription("2026-02", positions)
		assert.Equal(t, "Software Development Services Rendered (February 1-28, 2026)", result)
	})

	t.Run("Design Priority Over Operation Executive", func(t *testing.T) {
		// If contractor has both design and operation executive, design takes priority
		positions := []string{"Operation Executive", "Product Designer"}
		result := generateServiceFeeDescription("2025-12", positions)
		// Since "Product Designer" contains "design", it should match design first
		assert.Equal(t, "Design Consulting Services Rendered (December 1-31, 2025)", result)
	})
}

func TestHelper_ConcatenateDescriptions(t *testing.T) {
	input := []string{"A", "", "  ", "B"}
	assert.Equal(t, "A\n\nB", concatenateDescriptions(input))
}

func TestHelper_StripDescriptionPrefixAndSuffix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Prefix tests - project name is kept, invoice reference stripped
		{
			name:     "Commission prefix with invoice ID - keeps project name, strips invoice ref",
			input:    "[RENAISS :: INV-DO5S8] Account Management Incentive for Invoice INV-DO5S8",
			expected: "RENAISS - Account Management Incentive",
		},
		{
			name:     "Sales commission prefix - keeps project name, strips invoice ref",
			input:    "[PLOT :: INV-OBI5D] Sales Commission for Invoice INV-OBI5D",
			expected: "PLOT - Sales Commission",
		},
		{
			name:     "Fee prefix with contractor - FEE is stripped (not a project)",
			input:    "[FEE :: SCOUTQA :: nikkingtr] :: 9JHY6",
			expected: ":: 9JHY6",
		},
		{
			name:     "PYT prefix - stripped (not a project)",
			input:    "[PYT :: 202512] Some payment description",
			expected: "Some payment description",
		},
		{
			name:     "No prefix - returns unchanged",
			input:    "Regular description without prefix",
			expected: "Regular description without prefix",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Open bracket but no close bracket",
			input:    "[incomplete prefix",
			expected: "[incomplete prefix",
		},
		{
			name:     "Prefix without separator - no project extracted",
			input:    "[PREFIX]NoSpace",
			expected: "NoSpace",
		},
		{
			name:     "Prefix with multiple spaces",
			input:    "[PROJECT :: ID]   Multiple spaces after",
			expected: "PROJECT - Multiple spaces after",
		},
		{
			name:     "Just prefix, no content after",
			input:    "[PROJECT :: ID] ",
			expected: "",
		},
		// Suffix tests
		{
			name:     "Amount suffix only",
			input:    "Account Management Incentive - $43.64 USD",
			expected: "Account Management Incentive",
		},
		{
			name:     "Amount suffix with larger amount - strips invoice ref",
			input:    "Sales Commission for Invoice INV-OBI5D - $182.85 USD",
			expected: "Sales Commission",
		},
		{
			name:     "Amount suffix with whole number - strips invoice ref",
			input:    "Sales Commission for Invoice INV-HD567 - $240 USD",
			expected: "Sales Commission",
		},
		// Combined prefix and suffix tests - project name kept, suffix stripped
		{
			name:     "Both prefix and suffix - keeps project name",
			input:    "[RENAISS :: INV-DO5S8] Account Management Incentive (Jan 2026 Client Retention) - $43.64 USD",
			expected: "RENAISS - Account Management Incentive (Jan 2026 Client Retention)",
		},
		{
			name:     "Both prefix and suffix - sales commission keeps project, strips invoice ref",
			input:    "[PLOT :: INV-OBI5D] Sales Commission for Invoice INV-OBI5D (Dec 2025 Services) - $182.85 USD",
			expected: "PLOT - Sales Commission (Dec 2025 Services)",
		},
		// Edge cases for suffix
		{
			name:     "Suffix without USD - not stripped",
			input:    "Some description - $100",
			expected: "Some description - $100",
		},
		{
			name:     "Suffix with different currency - not stripped",
			input:    "Some description - $100 EUR",
			expected: "Some description - $100 EUR",
		},
		{
			name:     "Dollar sign in middle - not stripped",
			input:    "Item costs $50 USD for processing",
			expected: "Item costs $50 USD for processing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripDescriptionPrefixAndSuffix(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
