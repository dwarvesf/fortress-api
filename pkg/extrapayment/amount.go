package extrapayment

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
	"github.com/patrickmn/go-cache"
)

var amountCache = cache.New(30*24*time.Hour, 24*time.Hour)

func ResolveAmountUSD(l logger.Logger, wiseSvc wise.IService, pageID string, amount float64, currency string) (float64, error) {
	conversionLogger := l.Fields(logger.Fields{
		"component":       "extra_payment_amount_resolver",
		"page_id":         pageID,
		"source_currency": strings.ToUpper(currency),
		"original_amount": amount,
		"target_currency": "USD",
	})

	conversionLogger.Debug("resolving extra payment USD amount")

	if strings.EqualFold(currency, "USD") || currency == "" {
		conversionLogger.Debug("extra payment already in USD; using original amount")
		return amount, nil
	}

	if cachedAmount, found := amountCache.Get(pageID); found {
		conversionLogger.Debug("extra payment USD amount found in cache")
		return cachedAmount.(float64), nil
	}

	if wiseSvc == nil {
		err := fmt.Errorf("wise service is required to convert %s extra payment amount for entry %s", strings.ToUpper(currency), pageID)
		conversionLogger.Error(err, "failed to resolve extra payment USD amount")
		return 0, err
	}

	convertedAmount, rate, err := wiseSvc.Convert(amount, currency, "USD")
	if err != nil {
		conversionLogger.Error(err, "wise conversion failed for extra payment amount")
		return 0, fmt.Errorf("failed to convert %s to USD for entry %s: %w", strings.ToUpper(currency), pageID, err)
	}

	roundedAmount := math.Round(convertedAmount*100) / 100
	amountCache.Set(pageID, roundedAmount, cache.DefaultExpiration)
	
	conversionLogger.Fields(logger.Fields{
		"wise_rate":        rate,
		"converted_amount": roundedAmount,
	}).Debug("resolved extra payment USD amount via Wise")

	return roundedAmount, nil
}
