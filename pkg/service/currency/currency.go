package currency

import (
	"encoding/json"
	"math"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type service struct {
	cacheMap *cache.Cache
	token    string
	cfg      *config.Config
}

func New(cfg *config.Config) IService {
	cacheMap := cache.New(24*time.Hour, 24*time.Hour)
	return &service{
		token:    cfg.CurrencyLayer.APIKey,
		cacheMap: cacheMap,
		cfg:      cfg,
	}
}

// TODO: need test
func (s *service) Convert(value float64, src, target string) (float64, float64, error) {
	rate, err := s.getRateForTwoCurrency(src, target)
	if err != nil {
		return 0, 0, err
	}
	return math.Ceil(value * rate), rate, nil
}

// here come the magic, since we can convert to USD only
// 1 USD = X_src
// 1 USD = Y_target
// <=> X_src = Y_target
// <=> src = Y/X target
func (s *service) getRateForTwoCurrency(src, target string) (float64, error) {
	x, err := s.GetRate(src)
	if err != nil {
		return 0, err
	}
	y, err := s.GetRate(target)
	if err != nil {
		return 0, err
	}
	return y / x, nil
}

func (s *service) USDToVND(usd float64) (float64, error) {
	res, err := s.usdCentToVND(int64(usd) * 100)
	if err != nil {
		return 0, err
	}
	return float64(res / 100), nil
}

func (s *service) VNDToUSD(vnd float64) (float64, error) {
	res, err := s.vndToUSDCent(int64(vnd))
	if err != nil {
		return 0, err
	}
	return float64(res / 100), nil
}

func (s *service) usdCentToVND(usd int64) (int64, error) {
	rate, err := s.getRate("VND")
	if err != nil {
		return 0, err
	}
	return usd * int64(rate), nil
}

func (s *service) vndToUSDCent(vnd int64) (int64, error) {
	rate, err := s.getRate("VND")
	if err != nil {
		return 0, err
	}
	float := float64(vnd) / rate
	return int64(math.Ceil(float) * 100), nil
}

func (s *service) GetByName(db *gorm.DB, name string) (*model.Currency, error) {
	c := model.Currency{}
	return &c, db.Where("name = ?", name).First(&c).Error
}

func (s *service) GetByID(db *gorm.DB, id model.UUID) (*model.Currency, error) {
	c := model.Currency{}
	return &c, db.Where("id = ?", id).First(&c).Error
}

// getRate will return the conversation rate between USD and target currency
func (s *service) GetRate(target string) (float64, error) {
	return s.getRate(target)
}

// TODO: clean this up
func (s *service) getRate(target string) (float64, error) {
	// first, we hit the cache
	t, _ := s.cacheMap.Get(target)
	if target, ok := t.(float64); ok {
		if target != 0 {
			return target, nil
		}
	}

	// move on, if on prod we do a real query, otherwise do fixed rate for reduce the cost
	if s.cfg.Env != "prod" {
		switch target {
		case "USD":
			return 1, nil
		case "CAD":
			return 1.34275, nil
		case "GBP":
			return 0.79185, nil
		case "EUR":
			return 0.89795, nil
		case "VND":
			return 25900, nil
		}
	}

	var client http.Client
	endpoint := "http://apilayer.net/api/live?currencies=USD,CAD,GBP,EUR,VND&access_key=" + s.token

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	type out struct {
		Quotes struct {
			USDVND float64 `json:"USDVND" desc:"US Dollar to Vietnamese Dong"`
			USDCAD float64 `json:"USDCAD" desc:"US Dollar to Canadian Dollar"`
			USDGBP float64 `json:"USDGBP" desc:"US Dollar to British Pound Sterling"`
			USDEUR float64 `json:"USDEUR" desc:"US Dollar to Euro"`
			USDUSD float64 `json:"USDUSD" desc:"US Dollar to US Dollar (always 1.0)"`
		} `json:"quotes"`
	}

	o := out{}
	err = json.NewDecoder(resp.Body).Decode(&o)
	if err != nil {
		return 0, err
	}

	s.cacheMap.Set("USD", o.Quotes.USDUSD, time.Hour*24)
	s.cacheMap.Set("GBP", o.Quotes.USDGBP, time.Hour*24)
	s.cacheMap.Set("EUR", o.Quotes.USDEUR, time.Hour*24)
	s.cacheMap.Set("VND", o.Quotes.USDVND, time.Hour*24)

	switch target {
	case "USD":
		return o.Quotes.USDUSD, nil
	case "GBP":
		return o.Quotes.USDGBP, nil
	case "EUR":
		return o.Quotes.USDEUR, nil
	case "VND":
		return o.Quotes.USDVND, nil
	case "CAD":
		return o.Quotes.USDCAD, nil
	}

	return 0, nil
}

func (s *service) GetCurrencyOption(db *gorm.DB) ([]model.Currency, error) {
	res := []model.Currency{}
	return res, db.Find(&res).Error
}
