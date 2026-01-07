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

	t.Run("Multiple Hourly Items", func(t *testing.T) {
		items := []ContractorInvoiceLineItem{
			{IsHourlyRate: true, Hours: 10, Rate: 50, Amount: 500, AmountUSD: 500, Description: "Desc A", OriginalCurrency: "USD"},
			{IsHourlyRate: true, Hours: 5, Rate: 50, Amount: 250, AmountUSD: 250, Description: "Desc B", OriginalCurrency: "USD"},
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

func TestHelper_ConcatenateDescriptions(t *testing.T) {
	input := []string{"A", "", "  ", "B"}
	assert.Equal(t, "A\n\nB", concatenateDescriptions(input))
}
