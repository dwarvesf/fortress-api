package conversionrate

import (
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

var srcCurrencies = []string{"USD", "VND"}

func (r *controller) Sync(c *gin.Context) error {
	tx, done := r.repo.NewTransaction()

	// Get list conversion rate
	conversionRates, err := r.store.ConversionRate.GetList(tx.DB())
	if err != nil {
		return done(err)
	}

	currencyRateMap := make(map[string]float64)
	for _, conversionRate := range conversionRates {
		for _, srcCurrency := range srcCurrencies {
			if conversionRate.Currency.Name == srcCurrency {
				continue
			}

			// Get rate
			rate, err := r.service.Wise.GetRate(srcCurrency, conversionRate.Currency.Name)
			if err != nil {
				return done(err)
			}
			currencyRateMap[conversionRate.Currency.Name] = rate
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
