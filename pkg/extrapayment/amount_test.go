package extrapayment_test

import (
	"context"
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
	ctx := context.Background()

	t.Run("returns USD amount without wise conversion", func(t *testing.T) {
		wiseSvc := &mockWiseService{}

		amountUSD, rate, err := extrapayment.ResolveAmountUSD(ctx, logger.NewLogrusLogger("debug"), wiseSvc, nil, "page-1", 125.5, "USD")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if amountUSD != 125.5 {
			t.Fatalf("expected 125.5 USD, got %v", amountUSD)
		}
		if rate != 1.0 {
			t.Fatalf("expected rate 1.0, got %v", rate)
		}
		if wiseSvc.convertCalls != 0 {
			t.Fatalf("expected wise convert not to be called, got %d calls", wiseSvc.convertCalls)
		}
	})

	t.Run("converts non USD amount using wise and rounds to 2 decimals", func(t *testing.T) {
		wiseSvc := &mockWiseService{convertedAmount: 379.694, rate: 0.0000379694}

		amountUSD, rate, err := extrapayment.ResolveAmountUSD(ctx, logger.NewLogrusLogger("debug"), wiseSvc, nil, "page-2", 10000000, "VND")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if amountUSD != 379.69 {
			t.Fatalf("expected 379.69 USD, got %v", amountUSD)
		}
		if rate != 0.0000379694 {
			t.Fatalf("expected rate 0.0000379694, got %v", rate)
		}
		if wiseSvc.convertCalls != 1 {
			t.Fatalf("expected wise convert to be called once, got %d calls", wiseSvc.convertCalls)
		}
	})

	t.Run("returns error when wise service is missing for non USD amount", func(t *testing.T) {
		_, _, err := extrapayment.ResolveAmountUSD(ctx, logger.NewLogrusLogger("debug"), nil, nil, "page-3", 10000000, "VND")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error when wise conversion fails", func(t *testing.T) {
		wiseSvc := &mockWiseService{err: errors.New("wise unavailable")}

		_, _, err := extrapayment.ResolveAmountUSD(ctx, logger.NewLogrusLogger("debug"), wiseSvc, nil, "page-4", 10000000, "VND")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("skips cache when redis is nil", func(t *testing.T) {
		wiseSvc := &mockWiseService{convertedAmount: 400.00, rate: 0.00004}

		// First call should hit the mock Wise service
		amountUSD1, rate1, err1 := extrapayment.ResolveAmountUSD(ctx, logger.NewLogrusLogger("debug"), wiseSvc, nil, "page-5-cache-test", 10000000, "VND")
		if err1 != nil {
			t.Fatalf("expected no error, got %v", err1)
		}
		if amountUSD1 != 400.00 {
			t.Fatalf("expected 400.00 USD, got %v", amountUSD1)
		}
		if rate1 != 0.00004 {
			t.Fatalf("expected rate 0.00004, got %v", rate1)
		}
		if wiseSvc.convertCalls != 1 {
			t.Fatalf("expected wise convert to be called once, got %d calls", wiseSvc.convertCalls)
		}

		// Second call should ALSO call Wise since redis is nil (no caching)
		amountUSD2, rate2, err2 := extrapayment.ResolveAmountUSD(ctx, logger.NewLogrusLogger("debug"), wiseSvc, nil, "page-5-cache-test", 10000000, "VND")
		if err2 != nil {
			t.Fatalf("expected no error, got %v", err2)
		}
		if amountUSD2 != 400.00 {
			t.Fatalf("expected 400.00 USD, got %v", amountUSD2)
		}
		if rate2 != 0.00004 {
			t.Fatalf("expected rate 0.00004, got %v", rate2)
		}
		if wiseSvc.convertCalls != 2 {
			t.Fatalf("expected wise convert to be called twice (no cache), got %d calls", wiseSvc.convertCalls)
		}
	})
}
