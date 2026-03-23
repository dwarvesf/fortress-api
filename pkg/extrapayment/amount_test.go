package extrapayment_test

import (
	"errors"
	"testing"

	"github.com/dwarvesf/fortress-api/pkg/extrapayment"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type mockWiseService struct {
	convertedAmount float64
	rate            float64
	err             error
	convertCalls    int
}

func (m *mockWiseService) Convert(amount float64, source, target string) (float64, float64, error) {
	m.convertCalls++
	return m.convertedAmount, m.rate, m.err
}

func (m *mockWiseService) GetPayrollQuotes(sourceCurrency, targetCurrency string, targetAmount float64) (*model.TWQuote, error) {
	return nil, nil
}

func (m *mockWiseService) GetRate(source, target string) (float64, error) {
	return m.rate, m.err
}

func TestResolveAmountUSD(t *testing.T) {
	t.Run("returns USD amount without wise conversion", func(t *testing.T) {
		wiseSvc := &mockWiseService{}

		amountUSD, err := extrapayment.ResolveAmountUSD(logger.NewLogrusLogger("debug"), wiseSvc, "page-1", 125.5, "USD")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if amountUSD != 125.5 {
			t.Fatalf("expected 125.5 USD, got %v", amountUSD)
		}
		if wiseSvc.convertCalls != 0 {
			t.Fatalf("expected wise convert not to be called, got %d calls", wiseSvc.convertCalls)
		}
	})

	t.Run("converts non USD amount using wise and rounds to 2 decimals", func(t *testing.T) {
		wiseSvc := &mockWiseService{convertedAmount: 379.694, rate: 0.0000379694}

		amountUSD, err := extrapayment.ResolveAmountUSD(logger.NewLogrusLogger("debug"), wiseSvc, "page-2", 10000000, "VND")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if amountUSD != 379.69 {
			t.Fatalf("expected 379.69 USD, got %v", amountUSD)
		}
		if wiseSvc.convertCalls != 1 {
			t.Fatalf("expected wise convert to be called once, got %d calls", wiseSvc.convertCalls)
		}
	})

	t.Run("returns error when wise service is missing for non USD amount", func(t *testing.T) {
		_, err := extrapayment.ResolveAmountUSD(logger.NewLogrusLogger("debug"), nil, "page-3", 10000000, "VND")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error when wise conversion fails", func(t *testing.T) {
		wiseSvc := &mockWiseService{err: errors.New("wise unavailable")}

		_, err := extrapayment.ResolveAmountUSD(logger.NewLogrusLogger("debug"), wiseSvc, "page-4", 10000000, "VND")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
