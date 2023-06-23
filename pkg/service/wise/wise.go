package wise

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

var (
	// default for now
	client = http.Client{
		Timeout: 5 * time.Second,
	}
)

const (
	// api version
	apiV1 = "v1/"

	// get quote url
	quotes = "quotes"

	// get transfer rates
	rates = "rates"
)

type wiseService struct {
	sync.Mutex
	cacheMap map[string]float64
	profile  string
	cfg      *config.Config
	l        logger.Logger
}

func New(cfg *config.Config, l logger.Logger) IService {
	client := &wiseService{
		cfg:      cfg,
		l:        l,
		profile:  cfg.Wise.Profile,
		cacheMap: make(map[string]float64),
	}
	go client.janitor()
	return client
}

func (w *wiseService) janitor() {
	t := time.NewTicker(5 * time.Minute)
	for {
		<-t.C
		w.Lock()
		w.cacheMap = map[string]float64{}
		w.Unlock()
	}
}

func (w *wiseService) Convert(amount float64, sourceCurrency, targetCurrency string) (float64, float64, error) {
	rate, err := w.GetRate(sourceCurrency, targetCurrency)
	if err != nil {
		return 0, 0, err
	}

	return amount * rate, rate, nil
}

func (w *wiseService) GetRate(sourceCurrency, targetCurrency string) (float64, error) {
	if sourceCurrency == targetCurrency {
		return 1, nil
	}

	return w.getTWRate(sourceCurrency, targetCurrency)
}

func (w *wiseService) getTWRate(sourceCurrency, targetCurrency string) (float64, error) {
	// if run_mode is non-prod, we use mock data
	if w.cfg.Env != "prod" {
		return getLocalRate(targetCurrency)
	}

	var conversionRate []model.WiseConversionRate

	l := w.l.Fields(logger.Fields{
		"handler": "wise",
		"method":  "getTWRate",
	})

	// try to get from cache to reduce api call
	rate := w.getCache(sourceCurrency + targetCurrency)
	if rate != 0 {
		return rate, nil
	}

	// build up request
	url := fmt.Sprintf("%v?source=%v&target=%v", w.getUrl(rates), sourceCurrency, targetCurrency)
	req, err := w.newRequest("GET", url, nil)
	if err != nil {
		l.Error(err, "can't build request")
		return 0, err
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	// read response
	resp, err := client.Do(req)
	if err != nil {
		l.Error(err, "can't get response")
		return 0, err
	}
	defer resp.Body.Close()
	body := resp.Body
	res, err := io.ReadAll(body)
	if err != nil {
		l.Error(err, "can't read response")
		return 0, err
	}

	err = json.Unmarshal(res, &conversionRate)
	if len(conversionRate) == 0 {
		l.Fields(logger.Fields{"msg": string(res)}).Error(err, "can't unmarshal response")
		return 0, errors.New("cannot get exchange rates")
	}

	// save to cache for further request within 5 minutes
	w.setCache(sourceCurrency+targetCurrency, conversionRate[0].Rate)

	return conversionRate[0].Rate, nil
}

// ///////////////////
// INTERNAL FUNCTIONS
// ///////////////////
func (w *wiseService) getCache(key string) float64 {
	if rate, ok := w.cacheMap[key]; ok {
		return rate
	}
	return 0
}

func (w *wiseService) setCache(key string, val float64) {
	w.Lock()
	defer w.Unlock()
	w.cacheMap[key] = val
}

// getLocalRate get conversion rate without making api call, for non-prod env
func getLocalRate(target string) (float64, error) {
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
		return 23416, nil
	case "SGD":
		return 1.3845, nil
	}
	return 1, nil
}

func (w *wiseService) getUrl(api string) string {
	return w.cfg.Wise.Url + apiV1 + api
}

func (w *wiseService) getAuthHeader() string {
	return "Bearer " + w.cfg.Wise.APIKey
}

func (w *wiseService) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	req.Header.Set("Authorization", w.getAuthHeader())
	req.Header.Set("Content-Type", "application/json")
	return req, err
}

func (w *wiseService) GetPayrollQuotes(sourceCurrency, targetCurrency string, targetAmount float64) (*model.TWQuote, error) {
	var q *model.TWQuote
	if w.cfg.Env != "prod" {
		return &model.TWQuote{
			SourceAmount: 0,
			Fee:          0,
			Rate:         0,
		}, nil
	}

	// Todo: (hnh)
	payload := strings.NewReader(fmt.Sprintf("{\n\t\"profile\": %v,\n\t\"source\": \"%s\",\n\t\"target\": \"%s\",\n\t\"rateType\": \"FIXED\",\n\t\"targetAmount\": %v,\n\t\"type\": \"BALANCE_PAYOUT\"\n}", w.profile, sourceCurrency, targetCurrency, targetAmount))

	req, _ := w.newRequest("POST", w.getUrl(quotes), payload)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body := resp.Body

	res, _ := io.ReadAll(body)

	return q, json.Unmarshal(res, &q)
}
