package wise

import "github.com/dwarvesf/fortress-api/pkg/model"

type IService interface {
	Convert(amount float64, source, target string) (convertedAmount float64, rate float64, error error)
	GetPayrollQuotes(sourceCurrency, targetCurrency string, targetAmount float64) (*model.TWQuote, error)
	GetRate(source, target string) (rate float64, err error)
}
