package extrapayment

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
)

type CachedConversion struct {
	AmountUSD float64 `json:"amount_usd"`
	Rate      float64 `json:"rate"`
}

// ResolveAmountUSD returns the converted USD amount and the exchange rate used (target per source)
func ResolveAmountUSD(ctx context.Context, l logger.Logger, wiseSvc wise.IService, rdb *redis.Client, pageID string, amount float64, currency string) (float64, float64, error) {
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
		return amount, 1.0, nil
	}

	cacheKey := fmt.Sprintf("extra_payment_conversion:%s", pageID)
	if rdb != nil {
		cachedValue, err := rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			var conversion CachedConversion
			if err := json.Unmarshal([]byte(cachedValue), &conversion); err == nil {
				conversionLogger.Debug("extra payment USD amount and rate found in Redis")
				return conversion.AmountUSD, conversion.Rate, nil
			}
		}
	}

	if wiseSvc == nil {
		err := fmt.Errorf("wise service is required to convert %s extra payment amount for entry %s", strings.ToUpper(currency), pageID)
		conversionLogger.Error(err, "failed to resolve extra payment USD amount")
		return 0, 0, err
	}

	convertedAmount, rate, err := wiseSvc.Convert(amount, currency, "USD")
	if err != nil {
		conversionLogger.Error(err, "wise conversion failed for extra payment amount")
		return 0, 0, fmt.Errorf("failed to convert %s to USD for entry %s: %w", strings.ToUpper(currency), pageID, err)
	}

	roundedAmount := math.Round(convertedAmount*100) / 100
	if rdb != nil {
		val, _ := json.Marshal(CachedConversion{
			AmountUSD: roundedAmount,
			Rate:      rate,
		})
		rdb.Set(ctx, cacheKey, val, 30*24*time.Hour)
	}
	
	conversionLogger.Fields(logger.Fields{
		"wise_rate":        rate,
		"converted_amount": roundedAmount,
	}).Debug("resolved extra payment USD amount via Wise")

	return roundedAmount, rate, nil
}
