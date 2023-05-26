// Package payroll please edit this file only with approval from hnh
package payroll

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/consts"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/handler/payroll/errs"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/currency"
	"github.com/dwarvesf/fortress-api/pkg/store"
	commissionStore "github.com/dwarvesf/fortress-api/pkg/store/commission"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/store/payroll"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/dwarvesf/fortress-api/pkg/worker"
)

type handler struct {
	controller *controller.Controller
	store      *store.Store
	service    *service.Service
	logger     logger.Logger
	worker     *worker.Worker
	repo       store.DBRepo
	config     *config.Config
}

// New returns a handler
func New(controller *controller.Controller, store *store.Store, repo store.DBRepo, service *service.Service, worker *worker.Worker, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		controller: controller,
		store:      store,
		repo:       repo,
		service:    service,
		worker:     worker,
		logger:     logger,
		config:     cfg,
	}
}

// GetPayrollsByMonth godoc
// @Summary  Get payrolls by month
// @Description Get payrolls by month
// @Tags payrolls
// @Accept  json
// @Produce  json
// @Success 200 {object} []model.Payroll
// @Failure 400 {object} view.ErrorResponse
func (h *handler) GetPayrollsByMonth(c *gin.Context) {
	q := c.Request.URL.Query()

	email := q.Get("email")

	if q.Get("next") == "true" {
		res, err := GetPayrollDetailHandler(h, 0, 0, 0, email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}

		c.JSON(http.StatusOK, view.CreateResponse(res, nil, nil, nil, ""))
		return
	}

	batch, err := strconv.ParseInt(q.Get("date"), 0, 64)
	if err != nil {
		batch = 0
	}

	year, err := strconv.ParseInt(q.Get("year"), 0, 64)
	if err != nil || year <= 0 {
		year = int64(time.Now().Year())
	}

	if q.Get("month") == "" {
		if year != int64(time.Now().Year()) {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidYear, nil, ""))
			return
		}
		year = int64(time.Now().Year())
	}
	month, err := strconv.ParseInt(q.Get("month"), 0, 64)
	if err != nil {
		month = int64(time.Now().Month())
	}

	res, err := GetPayrollDetailHandler(h, int(month), int(year), int(batch), email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(res, nil, nil, nil, ""))
}

type payrollResponse struct {
	DisplayName          string            `json:"display_name"`
	BaseSalary           int64             `json:"base_salary"`
	Bonus                float64           `json:"bonus"`
	TotalWithoutContract float64           `json:"total_without_contract"`
	TotalWithContract    model.VietnamDong `json:"total_with_contract"`
	Notes                []string          `json:"notes"`
	Date                 int               `json:"date"`
	Month                int               `json:"month"`
	Year                 int               `json:"year"`
	BankAccountNumber    string            `json:"bank_account_number"`
	TWRecipientID        string            `json:"tw_recipient_id"` // will be removed
	TWRecipientName      string            `json:"tw_recipient_name"`
	TWAccountNumber      string            `json:"tw_account_number"`
	Bank                 string            `json:"bank"`
	HasContract          bool              `json:"has_contract"`
	PayrollID            string            `json:"payroll_id"`
	IsCommit             bool              `json:"is_commit"`
	IsPaid               bool              `json:"is_paid"`
	TWGBP                float64           `json:"tw_gbp"` // will be removed
	TWAmount             float64           `json:"tw_amount"`
	TWFee                float64           `json:"tw_fee"`
	TWEmail              string            `json:"tw_email"`
	TWCurrency           string            `json:"tw_currency"`
	Currency             string            `json:"currency"`
}

func GetPayrollBHXHHandler(h *handler) (interface{}, error) {
	type payrollBHXHResponse struct {
		DisplayName   string `json:"display_name"`
		BHXH          int64  `json:"bhxh"`
		Batch         int    `json:"batch"`
		AccountNumber string `json:"account_number"`
		Bank          string `json:"bank"`
	}
	var res []payrollBHXHResponse

	isLeft := false
	for _, b := range []int{int(model.FirstBatch), int(model.SecondBatch)} {
		date := time.Date(time.Now().Year(), time.Now().Month(), b, 0, 0, 0, 0, time.Now().Location())
		us, _, err := h.store.Employee.All(h.repo.DB(), employee.EmployeeFilter{IsLeft: &isLeft, BatchDate: &date, Preload: true}, model.Pagination{Page: 0, Size: 500})
		if err != nil {
			return nil, err
		}
		for i := range us {
			if us[i].BaseSalary.CompanyAccountAmount == 0 || us[i].BaseSalary.Batch != b {
				continue
			}
			res = append(res, payrollBHXHResponse{
				DisplayName:   us[i].FullName,
				BHXH:          us[i].BaseSalary.CompanyAccountAmount,
				Batch:         b,
				AccountNumber: us[i].LocalBankNumber,
				Bank:          us[i].LocalBranchName,
			})
		}
	}

	return res, nil
}

func GetPayrollDetailHandler(h *handler, month, year, batch int, email string) (interface{}, error) {
	tx, done := h.repo.NewTransaction()
	if month == 0 && year == 0 && batch == 0 {
		date, err := h.store.Payroll.GetLatestCommitTime(tx.DB())
		if err != nil {
			h.logger.Error(err, "can't get payroll latest commit time")
			date = time.Now()
		}
		month = int(date.Month())
		year = date.Year()
		batch = int(model.FirstBatch)
		if date.Day() == int(model.FirstBatch) {
			batch = int(model.SecondBatch)
			month, year = timeutil.LastMonthYear(month, year)
		}
	} else {
		month, year = timeutil.LastMonthYear(month, year)
	}

	res := []payrollResponse{}
	isPaid := true
	var subTotal int64
	var bonusTotal model.VietnamDong

	for _, b := range []int{int(model.FirstBatch), int(model.SecondBatch)} {
		if batch != 0 && batch != b {
			continue
		}
		q := payroll.GetListPayrollInput{
			Month: month,
			Year:  year,
			Day:   b,
		}
		if email != "" {
			u, err := h.store.Employee.OneByEmail(tx.DB(), email)
			if err != nil {
				h.logger.Error(err, "can't get employee by email")
				return nil, err
			}
			q.UserID = u.ID.String()
		}

		batchDate := time.Date(year, time.Month(month), b, 0, 0, 0, 0, time.UTC)
		payrolls, err := h.store.Payroll.GetList(tx.DB(), q)
		if err != nil {
			h.logger.Error(err, "can't get payroll list")
			return nil, done(err)
		}

		if len(payrolls) == 0 {
			isPaid = false

			// TODO : get all user payroll
			isLeft := false
			us, _, err := h.store.Employee.All(tx.DB(), employee.EmployeeFilter{IsLeft: &isLeft, BatchDate: &batchDate, Preload: true}, model.Pagination{Page: 0, Size: 500})
			if err != nil {
				h.logger.Error(err, "can't get list active employee")
				return nil, err
			}

			var tempPayrolls []model.Payroll
			newPayrolls, err := calculatePayrolls(h, us, batchDate)
			if err != nil {
				h.logger.Error(err, "can't calculate payroll")
				return nil, err
			}
			for i := range newPayrolls {
				tempPayrolls = append(tempPayrolls, *newPayrolls[i])
			}
			payrolls = tempPayrolls
		}

		for i := range payrolls {
			var notes []string

			// get tw quotes (default -> GBP)
			c := payrolls[i].Employee.WiseCurrency
			if c == "" {
				c = "GBP"
			}

			// !isForecast
			if batchDate.Month() <= time.Now().Month() || (batchDate.Month() == 12 && time.Now().Month() == 1) {
				toDate := batchDate.AddDate(0, 1, 0)
				commissionQuery := commissionStore.Query{
					EmployeeID: payrolls[i].Employee.ID.String(),
					IsPaid:     isPaid,
					FromDate:   &batchDate,
					ToDate:     &toDate,
				}
				userCommissions, err := h.store.Commission.Get(h.repo.DB(), commissionQuery)
				if err != nil {
					return nil, err
				}
				duplicate := map[string]int{}
				for j := range userCommissions {
					if userCommissions[j].Amount != 0 {
						name := userCommissions[j].Invoice.Number
						if userCommissions[j].Note != "" {
							name = fmt.Sprintf("%v - %v", name, userCommissions[j].Note)
						}
						duplicate[name] += int(userCommissions[j].Amount)
					}
				}
				for j := range duplicate {
					notes = append(notes, fmt.Sprintf("%v (%v)", j, utils.FormatCurrencyAmount(duplicate[j])))
				}

				bonusExplain := []model.CommissionExplain{}
				if len(payrolls[i].ProjectBonusExplain) > 0 {
					if err := json.Unmarshal(payrolls[i].ProjectBonusExplain, &bonusExplain); err != nil {
						return nil, errs.ErrCannotReadProjectBonusExplain
					}
				}
				for j := range bonusExplain {
					notes = append(notes, fmt.Sprintf("%v (%v)", bonusExplain[j].Name, utils.FormatCurrencyAmount(int(bonusExplain[j].Amount))))
				}
			}

			bonus, sub, err := preparePayroll(h, c, &payrolls[i], false)
			if err != nil {
				return nil, err
			}
			subTotal += sub

			for j := range payrolls[i].CommissionExplains {
				isContained := false
				for n := range notes {
					if strings.Contains(notes[n], payrolls[i].CommissionExplains[j].Name) {
						isContained = true
					}
				}
				if isContained {
					continue
				}
				notes = append(notes, fmt.Sprintf("%v (%v)", payrolls[i].CommissionExplains[j].Name, utils.FormatCurrencyAmount(int(payrolls[i].CommissionExplains[j].Amount))))
			}
			for j := range payrolls[i].ProjectBonusExplains {
				isContained := false
				for n := range notes {
					if strings.Contains(notes[n], payrolls[i].ProjectBonusExplains[j].Name) {
						isContained = true
					}
				}
				if isContained {
					continue
				}
				notes = append(notes, fmt.Sprintf("%v (%v)", payrolls[i].ProjectBonusExplains[j].Name, utils.FormatCurrencyAmount(int(payrolls[i].ProjectBonusExplains[j].Amount))))
			}

			r := payrollResponse{
				DisplayName:          payrolls[i].Employee.FullName,
				BaseSalary:           payrolls[i].BaseSalaryAmount,
				Bonus:                bonus,
				TotalWithContract:    payrolls[i].Total,
				TotalWithoutContract: payrolls[i].TotalAllowance,
				Notes:                notes,
				Date:                 payrolls[i].Employee.BaseSalary.Batch,
				Month:                month,
				Year:                 year,
				BankAccountNumber:    payrolls[i].Employee.LocalBankNumber,
				Bank:                 payrolls[i].Employee.LocalBranchName,
				HasContract:          payrolls[i].ContractAmount != 0,
				PayrollID:            "",
				TWRecipientID:        payrolls[i].Employee.WiseRecipientID,
				TWRecipientName:      payrolls[i].Employee.WiseRecipientName,
				TWAccountNumber:      payrolls[i].Employee.WiseAccountNumber,
				TWEmail:              payrolls[i].Employee.WiseRecipientEmail,
				TWGBP:                payrolls[i].TWAmount,
				TWAmount:             payrolls[i].TWAmount,
				TWFee:                payrolls[i].TWFee,
				TWCurrency:           c,
				IsCommit:             !payrolls[i].ID.IsZero(),
				IsPaid:               payrolls[i].IsPaid,
				Currency:             payrolls[i].Employee.BaseSalary.Currency.Name,
			}
			if r.IsCommit {
				r.PayrollID = payrolls[i].ID.String()
			}
			res = append(res, r)
		}
		if email == "" {
			err = cachePayroll(h, month, year, batch, payrolls)
			if err != nil {
				return nil, err
			}
		}
	}

	// return res, done(nil)
	sub := model.NewVietnamDong(subTotal)
	return map[string]interface{}{
		"sub_total":   sub.Format(),
		"bonus_total": bonusTotal,
		"payrolls":    res,
	}, nil
}

func cachePayroll(h *handler, month, year, batch int, payrolls []model.Payroll) error {
	payrollsBytes, err := json.Marshal(&payrolls)
	if err != nil {
		return err
	}
	cPayroll, err := h.store.CachedPayroll.Get(h.repo.DB(), month, year, batch)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	cPayroll.Month = month
	cPayroll.Year = year
	cPayroll.Batch = batch
	cPayroll.Payrolls = payrollsBytes
	return h.store.CachedPayroll.Set(h.repo.DB(), cPayroll)
}

func preparePayroll(h *handler, c string, p *model.Payroll, markPaid bool) (float64, int64, error) {
	var bonus float64
	var subTotal int64
	if p.Employee.BaseSalary.Currency.Name != currency.VNDCurrency {
		c, _, err := h.service.Wise.Convert(float64(p.CommissionAmount), currency.VNDCurrency, p.Employee.BaseSalary.Currency.Name)
		if err != nil {
			return 0, 0, err
		}
		b, _, err := h.service.Wise.Convert(float64(p.ProjectBonusAmount), currency.VNDCurrency, p.Employee.BaseSalary.Currency.Name)
		if err != nil {
			return 0, 0, err
		}
		p.TotalAllowance = float64(p.BaseSalaryAmount) + c + b
		bonus = b + c
		base, _, err := h.service.Wise.Convert(float64(p.BaseSalaryAmount), p.Employee.BaseSalary.Currency.Name, currency.VNDCurrency)
		if err != nil {
			return 0, 0, err
		}
		subTotal += int64(base)
	} else {
		p.TotalAllowance = float64(p.BaseSalaryAmount) + float64(p.CommissionAmount+p.ProjectBonusAmount)
		bonus = float64(p.CommissionAmount + p.ProjectBonusAmount)
		subTotal += p.BaseSalaryAmount
	}

	projectBonusExplains, err := getProjectBonusExplains(p)
	if err != nil {
		return 0, 0, err
	}
	p.ProjectBonusExplains = projectBonusExplains

	commissionExplains, err := getCommissionExplains(h, p, markPaid)
	if err != nil {
		return 0, 0, err
	}
	p.CommissionExplains = commissionExplains

	for i, v := range p.ProjectBonusExplains {
		formattedAmount, err := getFormattedAmount(h, p, v.Amount)
		if err != nil {
			return 0, 0, err
		}
		p.ProjectBonusExplains[i].FormattedAmount = formattedAmount
	}
	for i, v := range p.CommissionExplains {
		formattedAmount, err := getFormattedAmount(h, p, v.Amount)
		if err != nil {
			return 0, 0, err
		}
		p.CommissionExplains[i].FormattedAmount = formattedAmount
	}

	quote, err := h.service.Wise.GetPayrollQuotes(c, p.Employee.BaseSalary.Currency.Name, p.TotalAllowance)
	if err != nil {
		h.logger.Error(err, "cannot use wise to get quote")
		return 0, 0, err
	}
	p.TWAmount = quote.SourceAmount
	p.TWRate = quote.Rate
	p.TWFee = quote.Fee

	return bonus, subTotal, nil
}

func getProjectBonusExplains(p *model.Payroll) ([]model.ProjectBonusExplain, error) {
	projectBonusExplains := []model.ProjectBonusExplain{}
	err := json.Unmarshal(
		p.ProjectBonusExplain,
		&projectBonusExplains,
	)
	if err != nil {
		return nil, err
	}

	tempBonus := map[string]*model.ProjectBonusExplain{}
	for i := range projectBonusExplains {
		if tempBonus[projectBonusExplains[i].Name] == nil {
			tempBonus[projectBonusExplains[i].Name] = &projectBonusExplains[i]
		} else {
			tempBonus[projectBonusExplains[i].Name].Amount += projectBonusExplains[i].Amount
		}
	}
	var tempBonusExplains []model.ProjectBonusExplain
	for i := range tempBonus {
		tempBonusExplains = append(tempBonusExplains, *tempBonus[i])
	}
	return tempBonusExplains, nil
}

func getCommissionExplains(h *handler, p *model.Payroll, markPaid bool) ([]model.CommissionExplain, error) {
	commissionExplains := []model.CommissionExplain{}
	err := json.Unmarshal(
		p.CommissionExplain,
		&commissionExplains,
	)
	if err != nil {
		return nil, err
	}
	if markPaid {
		for i := range commissionExplains {
			err := h.store.Commission.MarkPaid(h.repo.DB(), commissionExplains[i].ID)
			if err != nil {
				return nil, err
			}
		}
	}

	tempCommission := map[string]*model.CommissionExplain{}
	for i := range commissionExplains {
		if commissionExplains[i].Amount == 0 {
			continue
		}
		if tempCommission[commissionExplains[i].Name] == nil {
			tempCommission[commissionExplains[i].Name] = &commissionExplains[i]
		} else {
			tempCommission[commissionExplains[i].Name].Amount += commissionExplains[i].Amount
		}
	}
	var tempCommissionExplains []model.CommissionExplain
	for i := range tempCommission {
		tempCommissionExplains = append(tempCommissionExplains, *tempCommission[i])
	}
	return tempCommissionExplains, nil
}

func getFormattedAmount(h *handler, p *model.Payroll, amount model.VietnamDong) (string, error) {
	if p.Employee.BaseSalary.Currency.Name != currency.VNDCurrency {
		temp, _, err := h.service.Wise.Convert(float64(amount), currency.VNDCurrency, p.Employee.BaseSalary.Currency.Name)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%.02f", temp), nil
	}
	return amount.String(), nil
}

func storePayrollTransaction(h *handler, p []model.Payroll, batchDate time.Time) error {
	var transactions []*model.AccountingTransaction
	for i := range p {
		m := model.AccountingMetadata{
			Source: "payroll",
			ID:     p[i].ID.String(),
		}
		bonusBytes, err := json.Marshal(&m)
		if err != nil {
			return err
		}
		var organization string = "Dwarves Foundation"
		now := time.Now()

		// baseSalaryConversionAmount = total - bonus and reimbursement (commission + projectBonus) - contract, cannot use origin baseSalary becauuse it is not converted to VND
		baseSalaryConversionAmount := p[i].Total - model.NewVietnamDong(p[i].ContractAmount) - p[i].CommissionAmount - p[i].ProjectBonusAmount
		payrollTransaction := model.AccountingTransaction{
			Amount:           float64(p[i].BaseSalaryAmount),
			ConversionAmount: baseSalaryConversionAmount,
			Name:             fmt.Sprintf("Payroll TW - %v %v", p[i].Employee.FullName, batchDate.Format("02 January 2006")),
			Category:         p[i].Employee.BaseSalary.Category,
			Currency:         p[i].Employee.BaseSalary.Currency.Name,
			Date:             &now,
			CurrencyID:       &p[i].Employee.BaseSalary.Currency.ID,
			Organization:     organization,
			Metadata:         bonusBytes,
			ConversionRate:   float64(baseSalaryConversionAmount) / float64(p[i].BaseSalaryAmount),
			Type:             p[i].Employee.BaseSalary.Type,
		}
		transactions = append(transactions, &payrollTransaction)

		// separate BHXH and transferwise
		if p[i].ContractAmount != 0 {
			bhxhTransaction := model.AccountingTransaction{
				Amount:           float64(p[i].ContractAmount),
				ConversionAmount: model.NewVietnamDong(p[i].ContractAmount),
				Name:             fmt.Sprintf("Payroll BHXH - %v %v", p[i].Employee.FullName, batchDate.Format("02 January 2006")),
				Category:         p[i].Employee.BaseSalary.Category,
				Currency:         p[i].Employee.BaseSalary.Currency.Name,
				Date:             &now,
				CurrencyID:       &p[i].Employee.BaseSalary.Currency.ID,
				Organization:     organization,
				Metadata:         bonusBytes,
				ConversionRate:   1,
				Type:             p[i].Employee.BaseSalary.Type,
			}
			transactions = append(transactions, &bhxhTransaction)
		}

		// bonusConversionAmount = total bonus (commission + projectBonus) - reimbursement  (total - conversionAmount)
		bonusConversionAmount := p[i].CommissionAmount + p[i].ProjectBonusAmount - (p[i].Total - p[i].ConversionAmount)
		if bonusConversionAmount != 0 {
			currency, err := h.store.Currency.GetByName(h.repo.DB(), currency.VNDCurrency)
			if err != nil {
				return err
			}
			category := p[i].Employee.BaseSalary.Category
			t := model.AccountingSE
			if p[i].Employee.BaseSalary.Category == model.AccountingRec {
				category = model.AccountingCommHiring
				t = p[i].Employee.BaseSalary.Type
			}
			var commissionName string
			for j := range p[i].CommissionExplains {
				commissionName += p[i].CommissionExplains[j].Name + " "
			}
			if strings.Contains(commissionName, "Account") {
				category = model.AccountingCommAccount
			}
			if strings.Contains(commissionName, "Sales") {
				category = model.AccountingCommSales
			}
			if strings.Contains(commissionName, "Lead") {
				category = model.AccountingCommLead
			}
			bonusTransaction := model.AccountingTransaction{
				Amount:           float64(bonusConversionAmount),
				ConversionAmount: bonusConversionAmount,
				Name:             fmt.Sprintf("Bonus - %v %v", p[i].Employee.FullName, batchDate.Format("02 January 2006")),
				Category:         category,
				Currency:         currency.Name,
				Date:             &now,
				CurrencyID:       &currency.ID,
				Organization:     organization,
				Metadata:         bonusBytes,
				ConversionRate:   1,
				Type:             t,
			}
			transactions = append(transactions, &bonusTransaction)
		}
	}

	return StoreMultipleTransaction(h, transactions)
}

func StoreMultipleTransaction(h *handler, transactions []*model.AccountingTransaction) error {
	if err := h.store.Accounting.CreateMultipleTransaction(h.repo.DB(), transactions); err != nil {
		return err
	}
	return nil
}

func (h *handler) GetPayrollsBHXH(c *gin.Context) {
	res, err := GetPayrollBHXHHandler(h)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(res, nil, nil, nil, ""))
}

func (h *handler) CommitPayroll(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "payroll",
		"method":  "CommitPayroll",
	})

	l.Info("Start commit payroll")

	year, err := strconv.ParseInt(c.Query("year"), 0, 64)
	if err != nil || year <= 0 {
		year = int64(time.Now().Year())
	}

	if c.Query("month") == "" {
		if year != int64(time.Now().Year()) {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrBadRequest, nil, ""))
			return
		}
		year = int64(time.Now().Year())
	}
	month, err := strconv.ParseInt(c.Query("month"), 0, 64)
	if err != nil {
		l.Errorf(err, "failed to parse month", "month", month)
		month = int64(time.Now().Month())
	}

	batch, err := strconv.ParseInt(c.Query("date"), 0, 64)
	if err != nil {
		l.Errorf(err, "failed to parse date", "date", batch)
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrBadRequest, nil, ""))
		return
	}

	email := c.Query("email")

	err = h.commitPayrollHandler(int(month), int(year), int(batch), email)
	if err != nil {
		l.Errorf(err, "failed to parse date", "date", batch)
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	err = h.controller.Discord.Log(model.LogDiscordInput{
		Type: "payroll_commit",
		Data: map[string]interface{}{
			"batch_number": strconv.Itoa(int(batch)),
			"month":        time.Month(month).String(),
			"year":         year,
		},
	})
	if err != nil {
		l.Error(err, "failed to logs to discord")
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
}

func (h *handler) commitPayrollHandler(month, year, batch int, email string) error {
	month, year = timeutil.LastMonthYear(month, year)
	if month > int(time.Now().Month()) {
		if month != 12 && time.Now().Month() != 1 {
			return errors.New("cannot commit payroll too far away")
		}
	}
	for _, b := range []int{int(model.FirstBatch), int(model.SecondBatch)} {
		if batch != b {
			continue
		}
		q := payroll.GetListPayrollInput{
			Month: month,
			Year:  year,
			Day:   b,
		}
		if email != "" {
			u, err := h.store.Employee.OneByEmail(h.repo.DB(), email)
			if err != nil {
				return err
			}
			q.UserID = u.ID.String()
		}

		batchDate := time.Date(year, time.Month(month), b, 0, 0, 0, 0, time.UTC)
		var payrolls []model.Payroll
		cPayroll, err := h.store.CachedPayroll.Get(h.repo.DB(), month, year, batch)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errs.ErrPayrollNotSnapshotted
			}
			return err
		}
		err = json.Unmarshal(cPayroll.Payrolls, &payrolls)
		if err != nil {
			return err
		}

		// this is for testing
		// filteredPayrolls := []model.Payroll{}
		// for _, p := range payrolls {
		// 	if p.Employee.TeamEmail == "huy@d.foundation" {
		// 		filteredPayrolls = append(filteredPayrolls, p)
		// 	}
		// }
		// payrolls = filteredPayrolls

		for i := range payrolls {
			payrolls[i].IsPaid = true
			err = markBonusAsDone(h, &payrolls[i])
			if err != nil {
				return err
			}
			// hacky way to mark done commission
			if _, err := getCommissionExplains(h, &payrolls[i], true); err != nil {
				return err
			}
		}

		// Batch insert payrolls
		err = h.store.Payroll.InsertList(h.repo.DB(), payrolls)
		if err != nil {
			return err
		}

		// Using WaitGroup go routines to SendPayrollPaidEmail
		var wg sync.WaitGroup
		wg.Add(len(payrolls))
		c := make(chan *model.Payroll, len(payrolls)+1)
		for _, payroll := range payrolls {
			go func(p model.Payroll) {
				defer wg.Done()
				if h.config.Env == "prod" || p.Employee.TeamEmail == "quang@d.foundation" || p.Employee.TeamEmail == "huy@d.foundation" {
					c <- &p
				}
			}(payroll)
		}
		wg.Wait()
		c <- nil
		go activateGmailQueue(h, c)

		err = storePayrollTransaction(h, payrolls, batchDate)
		if err != nil {
			return err
		}
	}

	return nil
}

func markBonusAsDone(h *handler, p *model.Payroll) error {
	var projectBonusExplains []model.ProjectBonusExplain
	err := json.Unmarshal(
		p.ProjectBonusExplain,
		&projectBonusExplains,
	)
	if err != nil {
		return err
	}

	for i := range projectBonusExplains {
		// handle the bonus from basecamp todo (expense reimbursement and accounting)
		if projectBonusExplains[i].BasecampBucketID != 0 && projectBonusExplains[i].BasecampTodoID != 0 {
			woodlandID := consts.PlaygroundID
			accountingID := consts.PlaygroundID
			if h.config.Env == "prod" {
				woodlandID = consts.WoodlandID
				accountingID = consts.AccountingID
			}

			// expense reimbursement -> from woodland -> mark done
			// accounting -> comment confirm message
			switch projectBonusExplains[i].BasecampBucketID {
			case woodlandID:
				err := h.service.Basecamp.Todo.Complete(projectBonusExplains[i].BasecampBucketID,
					projectBonusExplains[i].BasecampTodoID)
				if err != nil {
					return err
				}

			case accountingID:
				mention, err := h.service.Basecamp.BasecampMention(p.Employee.BasecampID)
				if err != nil {
					return err
				}
				msg := fmt.Sprintf("Amount has been deposited in your payroll %v", mention)
				cm := h.service.Basecamp.BuildCommentMessage(projectBonusExplains[i].BasecampBucketID, projectBonusExplains[i].BasecampTodoID, msg, "")
				h.worker.Enqueue(model.BasecampCommentMsg, cm)
			}
		}
	}
	return nil
}

func (h *handler) MarkPayrollAsPaid(c *gin.Context) {
	ids := []string{}
	if err := c.Bind(&ids); err != nil {
		return
	}

	err := markPayrollAsPaid(h, ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
	}
	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "marking the payroll as paid was successful"))
}

// MarkPayrollAsPaid from selected payroll row in database
// update the is_paid and is_sent_mail in payroll table
// send email into the users that marked as paid
func markPayrollAsPaid(h *handler, ids []string) error {
	for _, id := range ids {
		q := payroll.GetListPayrollInput{
			ID: id,
		}
		ps, err := h.store.Payroll.GetList(h.repo.DB(), q)
		if err != nil {
			return err
		}
		if len(ps) == 0 {
			continue
		}

		if err := h.service.GoogleMail.SendPayrollPaidMail(&ps[0]); err != nil {
			return err
		}

		fields := map[string]interface{}{
			"is_paid": true,
		}
		if err := h.store.Payroll.UpdateSpecificFields(h.repo.DB(), id, fields); err != nil {
			return err
		}
	}
	return nil
}

func activateGmailQueue(h *handler, p chan *model.Payroll) {
	h.logger.Info("gmail queue activated")
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	emailCh := make(chan *model.Payroll)

	go func() {
		for c := range emailCh {
			h.logger.Info(fmt.Sprintf("sending email %v", c.Employee.TeamEmail))
			err := h.service.GoogleMail.SendPayrollPaidMail(c)
			if err != nil {
				h.logger.Error(err, "error when sending email")
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			// Do something every second
		case c := <-p:
			if c == nil {
				return
			}
			emailCh <- c
		}
	}
}
