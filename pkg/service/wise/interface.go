package wise

type WiseService interface {
	GetPayrollQuotes(sourceCurrency, targetCurrency string, targetAmount float64) (*TWQuote, error)
	Convert(amount float64, sourceCurrency, targetCurrency string) (float64, float64, error)
	GetRate(sourceCurrency, targetCurrency string) (float64, error)
}
