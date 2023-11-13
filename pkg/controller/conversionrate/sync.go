package conversionrate

import (
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

var destCurrencies = []string{"USD", "VND"}

func (r *controller) Sync(c *gin.Context) error {
	tx, done := r.repo.NewTransaction()

	// Get list conversion rate
	conversionRates, err := r.store.ConversionRate.GetList(tx.DB())
	if err != nil {
		return done(err)
	}

	currencyRateMap := make(map[string]float64)
	for _, conversionRate := range conversionRates {
		srcCurrency := conversionRate.Currency.Name
		for _, destCurrency := range destCurrencies {
			rate, err := r.service.Wise.GetRate(srcCurrency, destCurrency)
			if err != nil {
				return done(err)
			}
			currencyRateMap[destCurrency] = rate
		}

		for k, v := range currencyRateMap {
			switch k {
			case "USD":
				conversionRate.ToUSD = decimal.NewFromFloat(v)
			case "VND":
				conversionRate.ToVND = decimal.NewFromFloat(v)
			}
		}

		if err := r.store.ConversionRate.Update(tx.DB(), &conversionRate); err != nil {
			return done(err)
		}
	}

	return done(nil)
}
