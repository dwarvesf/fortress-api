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
			name:     "Commission prefix with invoice ID - keeps project name and invoice ref",
			input:    "[RENAISS :: INV-DO5S8] Account Management Incentive for Invoice INV-DO5S8",
			expected: "RENAISS - Account Management Incentive for Invoice INV-DO5S8",
		},
		{
			name:     "Sales commission prefix - keeps project name and invoice ref",
			input:    "[PLOT :: INV-OBI5D] Sales Commission for Invoice INV-OBI5D",
			expected: "PLOT - Sales Commission for Invoice INV-OBI5D",
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
			name:     "Amount suffix with larger amount - keeps invoice ref",
			input:    "Sales Commission for Invoice INV-OBI5D - $182.85 USD",
			expected: "Sales Commission for Invoice INV-OBI5D",
		},
		{
			name:     "Amount suffix with whole number - keeps invoice ref",
			input:    "Sales Commission for Invoice INV-HD567 - $240 USD",
			expected: "Sales Commission for Invoice INV-HD567",
		},
		// Combined prefix and suffix tests - project name kept, suffix stripped
		{
			name:     "Both prefix and suffix - keeps project name",
			input:    "[RENAISS :: INV-DO5S8] Account Management Incentive (Jan 2026 Client Retention) - $43.64 USD",
			expected: "RENAISS - Account Management Incentive (Jan 2026 Client Retention)",
		},
		{
			name:     "Both prefix and suffix - sales commission keeps project and invoice ref",
			input:    "[PLOT :: INV-OBI5D] Sales Commission for Invoice INV-OBI5D (Dec 2025 Services) - $182.85 USD",
			expected: "PLOT - Sales Commission for Invoice INV-OBI5D (Dec 2025 Services)",
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

func TestGroupItemsByDescription(t *testing.T) {
	tests := []struct {
		name     string
		input    []ContractorInvoiceLineItem
		expected []ContractorInvoiceLineItem
	}{
		{
			name:     "empty slice",
			input:    nil,
			expected: nil,
		},
		{
			name: "single item - returned as-is (early return)",
			input: []ContractorInvoiceLineItem{
				{Description: "Delivery Lead", AmountUSD: 100, OriginalAmount: 2500000, OriginalCurrency: "VND", Type: "ServiceFee", PayoutPageIDs: []string{"p1"}},
			},
			expected: []ContractorInvoiceLineItem{
				{Description: "Delivery Lead", AmountUSD: 100, OriginalAmount: 2500000, OriginalCurrency: "VND", Type: "ServiceFee", PayoutPageIDs: []string{"p1"}},
			},
		},
		{
			name: "all unique descriptions - no grouping",
			input: []ContractorInvoiceLineItem{
				{Description: "Delivery Lead", AmountUSD: 100, OriginalAmount: 100, PayoutPageIDs: []string{"p1"}},
				{Description: "Account Management", AmountUSD: 200, OriginalAmount: 200, PayoutPageIDs: []string{"p2"}},
			},
			expected: []ContractorInvoiceLineItem{
				{Description: "Delivery Lead", Hours: 1, Rate: 100, AmountUSD: 100, OriginalAmount: 100, PayoutPageIDs: []string{"p1"}},
				{Description: "Account Management", Hours: 1, Rate: 200, AmountUSD: 200, OriginalAmount: 200, PayoutPageIDs: []string{"p2"}},
			},
		},
		{
			name: "two items with same description - grouped with summed amounts",
			input: []ContractorInvoiceLineItem{
				{Description: "Delivery Lead", AmountUSD: 100, OriginalAmount: 2500000, OriginalCurrency: "VND", Type: "ServiceFee", PayoutPageIDs: []string{"p1"}},
				{Description: "Delivery Lead", AmountUSD: 150, OriginalAmount: 3750000, OriginalCurrency: "VND", Type: "ServiceFee", PayoutPageIDs: []string{"p2"}},
			},
			expected: []ContractorInvoiceLineItem{
				{Description: "Delivery Lead", Hours: 1, Rate: 250, AmountUSD: 250, OriginalAmount: 6250000, OriginalCurrency: "VND", Type: "ServiceFee", PayoutPageIDs: []string{"p1", "p2"}},
			},
		},
		{
			name: "three items same description - all merged",
			input: []ContractorInvoiceLineItem{
				{Description: "Sales Commission", AmountUSD: 50, OriginalAmount: 50, Type: "Commission", PayoutPageIDs: []string{"p1"}},
				{Description: "Sales Commission", AmountUSD: 75, OriginalAmount: 75, Type: "Commission", PayoutPageIDs: []string{"p2"}},
				{Description: "Sales Commission", AmountUSD: 25, OriginalAmount: 25, Type: "Commission", PayoutPageIDs: []string{"p3"}},
			},
			expected: []ContractorInvoiceLineItem{
				{Description: "Sales Commission", Hours: 1, Rate: 150, AmountUSD: 150, OriginalAmount: 150, Type: "Commission", PayoutPageIDs: []string{"p1", "p2", "p3"}},
			},
		},
		{
			name: "mixed unique and duplicate - preserves first-occurrence order",
			input: []ContractorInvoiceLineItem{
				{Description: "Delivery Lead", AmountUSD: 100, OriginalAmount: 100, Type: "ServiceFee", PayoutPageIDs: []string{"p1"}},
				{Description: "Account Management", AmountUSD: 200, OriginalAmount: 200, Type: "Commission", PayoutPageIDs: []string{"p2"}},
				{Description: "Delivery Lead", AmountUSD: 50, OriginalAmount: 50, Type: "ServiceFee", PayoutPageIDs: []string{"p3"}},
				{Description: "Referral Bonus", AmountUSD: 80, OriginalAmount: 80, Type: "ExtraPayment", PayoutPageIDs: []string{"p4"}},
				{Description: "Account Management", AmountUSD: 100, OriginalAmount: 100, Type: "Commission", PayoutPageIDs: []string{"p5"}},
			},
			expected: []ContractorInvoiceLineItem{
				{Description: "Delivery Lead", Hours: 1, Rate: 150, AmountUSD: 150, OriginalAmount: 150, Type: "ServiceFee", PayoutPageIDs: []string{"p1", "p3"}},
				{Description: "Account Management", Hours: 1, Rate: 300, AmountUSD: 300, OriginalAmount: 300, Type: "Commission", PayoutPageIDs: []string{"p2", "p5"}},
				{Description: "Referral Bonus", Hours: 1, Rate: 80, AmountUSD: 80, OriginalAmount: 80, Type: "ExtraPayment", PayoutPageIDs: []string{"p4"}},
			},
		},
		{
			name: "preserves Type and OriginalCurrency from first item",
			input: []ContractorInvoiceLineItem{
				{Description: "Bonus", AmountUSD: 100, OriginalAmount: 2500000, OriginalCurrency: "VND", Type: "Commission", PayoutPageIDs: []string{"p1"}},
				{Description: "Bonus", AmountUSD: 200, OriginalAmount: 5000000, OriginalCurrency: "USD", Type: "ExtraPayment", PayoutPageIDs: []string{"p2"}},
			},
			expected: []ContractorInvoiceLineItem{
				{Description: "Bonus", Hours: 1, Rate: 300, AmountUSD: 300, OriginalAmount: 7500000, OriginalCurrency: "VND", Type: "Commission", PayoutPageIDs: []string{"p1", "p2"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := groupItemsByDescription(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGroupLineItemsIntoSections_GroupsDuplicateDescriptions tests that the groupLineItemsIntoSections
// function correctly groups Fee and Extra Payment items with duplicate descriptions.
// Test data mirrors real Notion payout entries where the same contractor has multiple payouts
// with identical descriptions across different invoices/months.
func TestGroupLineItemsIntoSections_GroupsDuplicateDescriptions(t *testing.T) {
	l := logger.NewLogrusLogger("debug")

	tests := []struct {
		name                   string
		items                  []ContractorInvoiceLineItem
		invoiceType            string
		expectedFeeCount       int    // expected number of items in Fee section
		expectedExtraCount     int    // expected number of items in Extra Payment section
		checkSection           string // which section to inspect in detail
		expectedAmounts        []float64
		expectedDescriptions   []string
		expectedPayoutIDCounts []int // expected number of payout page IDs per grouped item
	}{
		{
			name: "Fee section - duplicate Delivery Lead entries from multiple Notion payouts",
			items: []ContractorInvoiceLineItem{
				// Real pattern: same contractor has "Delivery Lead" fee across multiple invoice splits
				{
					Description:      "RENAISS - Delivery Lead",
					Hours:            1,
					Rate:             43.64,
					Amount:           43.64,
					AmountUSD:        43.64,
					Type:             string(notion.PayoutSourceTypeServiceFee),
					OriginalAmount:   1100000,
					OriginalCurrency: "VND",
					PayoutPageIDs:    []string{"payout-abc-001"},
				},
				{
					Description:      "RENAISS - Delivery Lead",
					Hours:            1,
					Rate:             43.64,
					Amount:           43.64,
					AmountUSD:        43.64,
					Type:             string(notion.PayoutSourceTypeServiceFee),
					OriginalAmount:   1100000,
					OriginalCurrency: "VND",
					PayoutPageIDs:    []string{"payout-abc-002"},
				},
				{
					Description:      "PLOT - Account Management",
					Hours:            1,
					Rate:             87.28,
					Amount:           87.28,
					AmountUSD:        87.28,
					Type:             string(notion.PayoutSourceTypeServiceFee),
					OriginalAmount:   2200000,
					OriginalCurrency: "VND",
					PayoutPageIDs:    []string{"payout-abc-003"},
				},
			},
			invoiceType:            "",
			expectedFeeCount:       2, // "RENAISS - Delivery Lead" grouped + "PLOT - Account Management"
			checkSection:           "Fee",
			expectedAmounts:        []float64{87.28, 87.28}, // 43.64+43.64, 87.28
			expectedDescriptions:   []string{"RENAISS - Delivery Lead", "PLOT - Account Management"},
			expectedPayoutIDCounts: []int{2, 1},
		},
		{
			name: "Extra Payment section - duplicate Sales Commission entries from multiple invoice splits",
			items: []ContractorInvoiceLineItem{
				// Real pattern: same commission description appears for multiple invoice splits
				{
					Description:      "PLOT - Sales Commission for Invoice INV-OBI5D",
					Hours:            1,
					Rate:             182.85,
					Amount:           182.85,
					AmountUSD:        182.85,
					Type:             string(notion.PayoutSourceTypeCommission),
					OriginalAmount:   182.85,
					OriginalCurrency: "USD",
					PayoutPageIDs:    []string{"payout-comm-001"},
				},
				{
					Description:      "PLOT - Sales Commission for Invoice INV-OBI5D",
					Hours:            1,
					Rate:             91.42,
					Amount:           91.42,
					AmountUSD:        91.42,
					Type:             string(notion.PayoutSourceTypeCommission),
					OriginalAmount:   91.42,
					OriginalCurrency: "USD",
					PayoutPageIDs:    []string{"payout-comm-002"},
				},
				{
					Description:      "Referral Fee",
					Hours:            1,
					Rate:             50.00,
					Amount:           50.00,
					AmountUSD:        50.00,
					Type:             string(notion.PayoutSourceTypeExtraPayment),
					OriginalAmount:   50.00,
					OriginalCurrency: "USD",
					PayoutPageIDs:    []string{"payout-extra-001"},
				},
			},
			invoiceType:            "",
			expectedExtraCount:     2, // "PLOT - Sales Commission..." grouped + "Referral Fee"
			checkSection:           "Extra Payment",
			expectedAmounts:        []float64{274.27, 50.00}, // 182.85+91.42, 50.00
			expectedDescriptions:   []string{"PLOT - Sales Commission for Invoice INV-OBI5D", "Referral Fee"},
			expectedPayoutIDCounts: []int{2, 1},
		},
		{
			name: "Mixed sections - Fee and Extra Payment both have duplicates",
			items: []ContractorInvoiceLineItem{
				// Fee items (ServiceFee without TaskOrderID)
				{
					Description:      "Delivery Lead",
					Hours:            1,
					Rate:             43.64,
					Amount:           43.64,
					AmountUSD:        43.64,
					Type:             string(notion.PayoutSourceTypeServiceFee),
					OriginalAmount:   1100000,
					OriginalCurrency: "VND",
					PayoutPageIDs:    []string{"fee-001"},
				},
				{
					Description:      "Delivery Lead",
					Hours:            1,
					Rate:             43.64,
					Amount:           43.64,
					AmountUSD:        43.64,
					Type:             string(notion.PayoutSourceTypeServiceFee),
					OriginalAmount:   1100000,
					OriginalCurrency: "VND",
					PayoutPageIDs:    []string{"fee-002"},
				},
				{
					Description:      "Delivery Lead",
					Hours:            1,
					Rate:             43.64,
					Amount:           43.64,
					AmountUSD:        43.64,
					Type:             string(notion.PayoutSourceTypeServiceFee),
					OriginalAmount:   1100000,
					OriginalCurrency: "VND",
					PayoutPageIDs:    []string{"fee-003"},
				},
				// Extra Payment items
				{
					Description:      "Account Management Incentive",
					Hours:            1,
					Rate:             100.00,
					Amount:           100.00,
					AmountUSD:        100.00,
					Type:             string(notion.PayoutSourceTypeCommission),
					OriginalAmount:   2500000,
					OriginalCurrency: "VND",
					PayoutPageIDs:    []string{"comm-001"},
				},
				{
					Description:      "Account Management Incentive",
					Hours:            1,
					Rate:             75.00,
					Amount:           75.00,
					AmountUSD:        75.00,
					Type:             string(notion.PayoutSourceTypeCommission),
					OriginalAmount:   1875000,
					OriginalCurrency: "VND",
					PayoutPageIDs:    []string{"comm-002"},
				},
			},
			invoiceType:      "",
			expectedFeeCount: 1, // 3x "Delivery Lead" -> 1 grouped item
			expectedExtraCount: 1, // 2x "Account Management Incentive" -> 1 grouped item
		},
		{
			name: "invoiceType=extra_payment filters to Extra Payment section only",
			items: []ContractorInvoiceLineItem{
				// Fee item - should be excluded
				{
					Description:   "Delivery Lead",
					AmountUSD:     43.64,
					Type:          string(notion.PayoutSourceTypeServiceFee),
					PayoutPageIDs: []string{"fee-001"},
				},
				// Extra Payment items with duplicates
				{
					Description:      "Sales Commission",
					Hours:            1,
					Rate:             100.00,
					AmountUSD:        100.00,
					Type:             string(notion.PayoutSourceTypeCommission),
					OriginalAmount:   100.00,
					OriginalCurrency: "USD",
					PayoutPageIDs:    []string{"comm-001"},
				},
				{
					Description:      "Sales Commission",
					Hours:            1,
					Rate:             50.00,
					AmountUSD:        50.00,
					Type:             string(notion.PayoutSourceTypeCommission),
					OriginalAmount:   50.00,
					OriginalCurrency: "USD",
					PayoutPageIDs:    []string{"comm-002"},
				},
			},
			invoiceType:            "extra_payment",
			expectedFeeCount:       -1, // Fee section should not exist
			expectedExtraCount:     1,  // 2x "Sales Commission" -> 1 grouped item
			checkSection:           "Extra Payment",
			expectedAmounts:        []float64{150.00},
			expectedDescriptions:   []string{"Sales Commission"},
			expectedPayoutIDCounts: []int{2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections := groupLineItemsIntoSections(tt.items, tt.invoiceType, l)

			// Find sections by name
			var feeSection, extraSection *ContractorInvoiceSection
			for i := range sections {
				switch sections[i].Name {
				case "Fee":
					feeSection = &sections[i]
				case "Extra Payment":
					extraSection = &sections[i]
				}
			}

			// Verify Fee section item count
			if tt.expectedFeeCount >= 0 {
				if tt.expectedFeeCount == 0 {
					assert.Nil(t, feeSection, "Fee section should not exist")
				} else {
					if assert.NotNil(t, feeSection, "Fee section should exist") {
						assert.Len(t, feeSection.Items, tt.expectedFeeCount, "Fee section item count")
					}
				}
			}

			// Verify Extra Payment section item count
			if tt.expectedExtraCount >= 0 {
				if tt.expectedExtraCount == 0 {
					assert.Nil(t, extraSection, "Extra Payment section should not exist")
				} else {
					if assert.NotNil(t, extraSection, "Extra Payment section should exist") {
						assert.Len(t, extraSection.Items, tt.expectedExtraCount, "Extra Payment section item count")
					}
				}
			}

			// Detailed checks on specified section
			if tt.checkSection != "" {
				var targetSection *ContractorInvoiceSection
				switch tt.checkSection {
				case "Fee":
					targetSection = feeSection
				case "Extra Payment":
					targetSection = extraSection
				}

				if assert.NotNil(t, targetSection, "target section %q should exist", tt.checkSection) {
					for i, item := range targetSection.Items {
						if i < len(tt.expectedDescriptions) {
							assert.Equal(t, tt.expectedDescriptions[i], item.Description, "item[%d] description", i)
						}
						if i < len(tt.expectedAmounts) {
							assert.InDelta(t, tt.expectedAmounts[i], item.AmountUSD, 0.01, "item[%d] AmountUSD", i)
							assert.InDelta(t, item.AmountUSD, item.Rate, 0.01, "item[%d] Rate should equal AmountUSD", i)
							assert.Equal(t, float64(1), item.Hours, "item[%d] Hours should be 1", i)
						}
						if i < len(tt.expectedPayoutIDCounts) {
							assert.Len(t, item.PayoutPageIDs, tt.expectedPayoutIDCounts[i], "item[%d] PayoutPageIDs count", i)
						}
					}
				}
			}
		})
	}
}
