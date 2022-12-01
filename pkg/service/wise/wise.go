package wise

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	// default for now
	client = http.Client{
		Timeout: 5 * time.Second,
	}
)

const (
	url = "https://api.transferwise.com/"
	// sandboxUrl = "https://api.sandbox.transferwise.tech/"
	sandboxUrl = "https://api.transferwise.com/"

	// api version
	apiv1 = "v1/"

	// get quote url
	quotes = "quotes"

	// get transfer rates
	rates = "rates"
)

type WiseClient struct {
	sync.Mutex

	url      string
	key      string
	profile  string
	cachemap map[string]float64
	isProd   bool
}

func New(key, profile, env string) WiseService {
	client := &WiseClient{
		url:      sandboxUrl,
		key:      key,
		profile:  profile,
		cachemap: map[string]float64{},
		isProd:   env == "prod",
	}

	if client.isProd {
		client.url = url
	}

	go client.janitor()
	return client
}

func (tw *WiseClient) janitor() {
	t := time.NewTicker(5 * time.Minute)
	for {
		<-t.C
		tw.Lock()
		tw.cachemap = map[string]float64{}
		tw.Unlock()
	}
}

func (tw *WiseClient) getcache(key string) float64 {
	if rate, ok := tw.cachemap[key]; ok {
		return rate
	}
	return 0
}

func (tw *WiseClient) setcache(key string, val float64) {
	tw.Lock()
	defer tw.Unlock()
	tw.cachemap[key] = val
}

func (tw *WiseClient) getUrl(api string) string {
	return tw.url + apiv1 + api
}

func (tw *WiseClient) getAuthHeader() string {
	return "Bearer " + tw.key

}

func (tw *WiseClient) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	req.Header.Set("Authorization", tw.getAuthHeader())
	req.Header.Set("Content-Type", "application/json")
	return req, err
}

func (tw *WiseClient) GetPayrollQuotes(sourceCurrency, targetCurrency string, targetAmount float64) (*TWQuote, error) {
	var q *TWQuote
	if !tw.isProd {
		return &TWQuote{
			SourceAmount: 0,
			Fee:          0,
			Rate:         0,
		}, nil
	}

	payload := strings.NewReader(fmt.Sprintf("{\n\t\"profile\": %v,\n\t\"source\": \"%s\",\n\t\"target\": \"%s\",\n\t\"rateType\": \"FIXED\",\n\t\"targetAmount\": %v,\n\t\"type\": \"BALANCE_PAYOUT\"\n}", tw.profile, sourceCurrency, targetCurrency, targetAmount))

	req, _ := tw.newRequest("POST", tw.getUrl(quotes), payload)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body := resp.Body

	res, _ := ioutil.ReadAll(body)

	return q, json.Unmarshal(res, &q)
}

func (tw *WiseClient) Convert(amount float64, sourceCurrency, targetCurrency string) (float64, float64, error) {
	if sourceCurrency == targetCurrency {
		return amount, 1, nil
	}
	rate, err := tw.GetRate(sourceCurrency, targetCurrency)
	if err != nil {
		return 0, 0, err
	}

	return amount * rate, rate, nil
}

func (tw *WiseClient) GetRate(sourceCurrency, targetCurrency string) (float64, error) {
	if sourceCurrency == targetCurrency {
		return 1, nil
	}
	if !tw.isProd {
		x := getLocalRate(sourceCurrency)
		y := getLocalRate(targetCurrency)
		return y / x, nil
	}

	return tw.getTWRate(sourceCurrency, targetCurrency)
}

func (tw *WiseClient) getTWRate(sourceCurrency, targetCurrency string) (float64, error) {
	var q []TWRate

	rate := tw.getcache(sourceCurrency + targetCurrency)
	if rate != 0 {
		return rate, nil
	}

	url := fmt.Sprintf("%v?source=%v&target=%v", tw.getUrl(rates), sourceCurrency, targetCurrency)

	req, _ := tw.newRequest("GET", url, nil)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body := resp.Body

	res, _ := ioutil.ReadAll(body)
	err = json.Unmarshal(res, &q)
	if err != nil {
		return 0, err
	}
	if len(q) == 0 {
		return 0, errors.New("cannot get exchange rates")
	}

	tw.setcache(sourceCurrency+targetCurrency, q[0].Rate)

	return q[0].Rate, nil
}

func getLocalRate(target string) float64 {
	switch target {
	case "USD":
		return 1
	case "CAD":
		return 1.34275
	case "GBP":
		return 0.79185
	case "EUR":
		return 0.89795
	case "VND":
		return 23416
	}
	return 1
}
