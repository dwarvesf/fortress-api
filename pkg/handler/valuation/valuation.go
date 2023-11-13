// please edit this file only with approval from hnh
package valuation

import (
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"
)

type handler struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
}

// New returns a handler
func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		store:   store,
		repo:    repo,
		service: service,
		logger:  logger,
		config:  cfg,
	}
}

// One godoc
// @Summary Get valuation by year
// @Description Get valuation
// @Tags valuation
// @Accept  json
// @Produce  json
// @Param year path int true "Year"
// @Success 200 {object} model.Valuation
// @Failure 400 {object} ErrorResponse
func (h *handler) One(c *gin.Context) {
	// parse params & prepare logger
	year := c.Param("year")
	if !utils.IsNumber(year) {
		c.JSON(400, gin.H{"message": "year must be number"})
		return
	}

	// we convert all number to usd
	convertTo := "USD"

	l := h.logger.Fields(logger.Fields{
		"handler": "valuation",
		"method":  "Details",
		"year":    year,
	})

	// account receivable is a number of money that company has to receive from customer
	// in our case, receivable is an amount of unpaid invoice
	receivable, err := h.store.Valuation.GetAccountReceivable(h.repo.DB(), year)
	if err != nil {
		l.Error(err, "can't get account receivable from this year")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, year, "can't get account receivable from this year"))
		return
	}

	// revenue equal all paid invoice, yield from investment & bank interest
	revenue, err := h.store.Valuation.GetRevenue(h.repo.DB(), year)
	if err != nil {
		l.Error(err, "can't get revenue from this year")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, year, "can't get revenue from this year"))
		return
	}

	// liabilities is a money that company has to pay in the future
	liabilities, amount, err := h.store.Valuation.GetLiabilities(h.repo.DB(), year)
	if err != nil {
		l.Error(err, "can't get liabilities from this year")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, year, "can't get liabilities from this year"))
		return
	}

	items := make([]model.AccountingItem, 0)
	for i := range liabilities {
		items = append(items, model.AccountingItem{
			Name:   liabilities[i].Name,
			Amount: liabilities[i].Total,
		})
	}

	// get investment
	investment, err := h.store.Valuation.GetInvestment(h.repo.DB(), year)
	if err != nil {
		l.Error(err, "can't get investment from this year")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, year, "can't get investment from this year"))
		return
	}

	// assets is a total worth of assets that company holds
	assets, err := h.store.Valuation.GetAssetAmount(h.repo.DB(), year)
	if err != nil {
		l.Error(err, "can't get liabilities from this year")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, year, "can't get liabilities from this year"))
		return
	}

	// expenses is a total amount we spent for operation expenses
	expenses, err := h.store.Valuation.GetExpense(h.repo.DB(), year)
	if err != nil {
		l.Error(err, "can't get expense from this year")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, year, "can't get expense from this year"))
		return
	}

	// payroll is a total amount we spent for salary
	payroll, err := h.store.Valuation.GetPayroll(h.repo.DB(), year)
	if err != nil {
		l.Error(err, "can't get expense from this year")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, year, "can't get expense from this year"))
		return
	}

	// build up response
	var valuation model.Valuation

	valuation.Year = year
	valuation.Currency = convertTo

	valuation.Assets = assets
	// we temporary doesn't need to return detail item for this rn
	valuation.AccountReceivable.Total = h.convertCurrency(receivable, convertTo)
	valuation.Liabilities.Total = h.convertCurrency(amount, convertTo)
	valuation.Liabilities.Items = items

	valuation.Income.Total = h.convertCurrency(revenue, convertTo)

	valuation.Outcome.Detail.Expense = h.convertCurrency(expenses, convertTo)
	valuation.Outcome.Detail.Payroll = h.convertCurrency(payroll, convertTo)
	valuation.Outcome.Detail.Investment = h.convertCurrency(investment, convertTo)
	valuation.Outcome.Total = valuation.Outcome.Detail.Expense + valuation.Outcome.Detail.Payroll + valuation.Outcome.Detail.Investment

	// return
	c.JSON(http.StatusOK, view.CreateResponse[any](valuation, nil, nil, year, "get valuation successfully"))
}

func (h *handler) convertCurrency(currency *model.CurrencyView, convertTo string) (convertedAmount float64) {
	if currency == nil {
		h.logger.Warn("currency struct is nil")
		return 0
	}

	// convert currency
	vndSource, _, _ := h.service.Wise.Convert(currency.VND, "VND", convertTo)
	usdSource, _, _ := h.service.Wise.Convert(currency.USD, "USD", convertTo)
	gbpSource, _, _ := h.service.Wise.Convert(currency.GBP, "GBP", convertTo)
	eurSource, _, _ := h.service.Wise.Convert(currency.EUR, "EUR", convertTo)
	sgdSource, _, _ := h.service.Wise.Convert(currency.SGD, "SGD", convertTo)

	return vndSource + usdSource + gbpSource + eurSource + sgdSource
}
