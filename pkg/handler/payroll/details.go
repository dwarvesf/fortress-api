package payroll

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/handler/payroll/errs"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/currency"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/store/employeecommission"
	"github.com/dwarvesf/fortress-api/pkg/store/payroll"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

func (h *handler) GetPayrollsByMonth(c *gin.Context) {
	q := c.Request.URL.Query()

	email := q.Get("email")

	if q.Get("next") == "true" {
		res, err := h.getPayrollDetailHandler(0, 0, 0, email)
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

	res, err := h.getPayrollDetailHandler(int(month), int(year), int(batch), email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(res, nil, nil, nil, ""))
}

func (h *handler) getPayrollDetailHandler(month, year, batch int, email string) (interface{}, error) {
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

	var res []payrollResponse
	var subTotal int64
	var bonusTotal model.VietnamDong

	isPaid := true

	for _, b := range []model.Batch{model.FirstBatch, model.SecondBatch} {
		if batch != 0 && batch != int(b) {
			continue
		}
		q := payroll.GetListPayrollInput{
			Month: month,
			Year:  year,
			Day:   int(b),
		}
		if email != "" {
			u, err := h.store.Employee.OneByEmail(tx.DB(), email)
			if err != nil {
				h.logger.Error(err, "can't get employee by email")
				return nil, err
			}
			q.UserID = u.ID.String()
		}

		batchDate := time.Date(year, time.Month(month), int(b), 0, 0, 0, 0, time.UTC)
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
			newPayrolls, err := h.calculatePayrolls(us, batchDate)
			if err != nil {
				h.logger.Error(err, "can't calculate payroll")
				return nil, err
			}

			for i := range newPayrolls {
				tempPayrolls = append(tempPayrolls, *newPayrolls[i])
			}
			payrolls = tempPayrolls
		}

		standardizedPayrolls := make([]model.Payroll, 0)
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
				commissionQuery := employeecommission.Query{
					EmployeeID: payrolls[i].Employee.ID.String(),
					IsPaid:     isPaid,
					FromDate:   &batchDate,
					ToDate:     &toDate,
				}
				userCommissions, err := h.store.EmployeeCommission.Get(h.repo.DB(), commissionQuery)
				if err != nil {
					return nil, err
				}
				duplicate := map[string]int{}
				for j := range userCommissions {
					if userCommissions[j].Amount != 0 {
						name := userCommissions[j].Invoice.Number
						if userCommissions[j].Note != "" {
							name = fmt.Sprintf("%v - Bonus", name)
						}
						duplicate[name] += int(userCommissions[j].Amount)
					}
				}
				for j := range duplicate {
					notes = append(notes, fmt.Sprintf("%v (%v)", j, utils.FormatCurrencyAmount(duplicate[j])))
				}

				var bonusExplain []model.CommissionExplain
				if len(payrolls[i].ProjectBonusExplain) > 0 {
					if err := json.Unmarshal(payrolls[i].ProjectBonusExplain, &bonusExplain); err != nil {
						return nil, errs.ErrCannotReadProjectBonusExplain
					}
				}
				for j := range bonusExplain {
					notes = append(notes, fmt.Sprintf("%v (%v)", bonusExplain[j].Name, utils.FormatCurrencyAmount(int(bonusExplain[j].Amount))))
				}
			}

			standardizedPayroll, bonus, sub, err := h.preparePayroll(c, payrolls[i], false)
			if err != nil {
				return nil, err
			}
			subTotal += sub

			for j := range standardizedPayroll.CommissionExplains {
				isContained := false
				for n := range notes {
					if strings.Contains(notes[n], standardizedPayroll.CommissionExplains[j].Name) {
						isContained = true
					}
				}
				if isContained {
					continue
				}
				notes = append(notes, fmt.Sprintf("%v (%v)", standardizedPayroll.CommissionExplains[j].Name, utils.FormatCurrencyAmount(int(standardizedPayroll.CommissionExplains[j].Amount))))
			}
			for j := range standardizedPayroll.ProjectBonusExplains {
				isContained := false
				for n := range notes {
					if strings.Contains(notes[n], standardizedPayroll.ProjectBonusExplains[j].Name) {
						isContained = true
					}
				}
				if isContained {
					continue
				}
				notes = append(notes, fmt.Sprintf("%v (%v)", standardizedPayroll.ProjectBonusExplains[j].Name, utils.FormatCurrencyAmount(int(standardizedPayroll.ProjectBonusExplains[j].Amount))))
			}

			r := payrollResponse{
				DisplayName:          standardizedPayroll.Employee.FullName,
				BaseSalary:           standardizedPayroll.BaseSalaryAmount,
				SalaryAdvanceAmount:  standardizedPayroll.SalaryAdvanceAmount,
				Bonus:                bonus,
				TotalWithContract:    standardizedPayroll.Total,
				TotalWithoutContract: standardizedPayroll.TotalAllowance,
				Notes:                notes,
				Date:                 standardizedPayroll.Employee.BaseSalary.Batch,
				Month:                month,
				Year:                 year,
				BankAccountNumber:    standardizedPayroll.Employee.LocalBankNumber,
				Bank:                 standardizedPayroll.Employee.LocalBranchName,
				HasContract:          standardizedPayroll.ContractAmount != 0,
				PayrollID:            "",
				TWRecipientID:        standardizedPayroll.Employee.WiseRecipientID,
				TWRecipientName:      standardizedPayroll.Employee.WiseRecipientName,
				TWAccountNumber:      standardizedPayroll.Employee.WiseAccountNumber,
				TWEmail:              standardizedPayroll.Employee.WiseRecipientEmail,
				TWGBP:                standardizedPayroll.TWAmount,
				TWAmount:             standardizedPayroll.TWAmount,
				TWFee:                standardizedPayroll.TWFee,
				TWCurrency:           c,
				IsCommit:             !standardizedPayroll.ID.IsZero(),
				IsPaid:               standardizedPayroll.IsPaid,
				Currency:             standardizedPayroll.Employee.BaseSalary.Currency.Name,
			}
			if r.IsCommit {
				r.PayrollID = standardizedPayroll.ID.String()
			}
			res = append(res, r)
			standardizedPayrolls = append(standardizedPayrolls, *standardizedPayroll)
		}

		if email == "" {
			err = h.cachePayroll(month, year, batch, standardizedPayrolls)
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

func (h *handler) preparePayroll(c string, p model.Payroll, markPaid bool) (*model.Payroll, float64, int64, error) {
	var bonus float64
	var subTotal int64

	if p.Employee.BaseSalary.Currency.Name != currency.VNDCurrency {
		c, _, err := h.service.Wise.Convert(float64(p.CommissionAmount), currency.VNDCurrency, p.Employee.BaseSalary.Currency.Name)
		if err != nil {
			return nil, 0, 0, err
		}

		b, _, err := h.service.Wise.Convert(float64(p.ProjectBonusAmount), currency.VNDCurrency, p.Employee.BaseSalary.Currency.Name)
		if err != nil {
			return nil, 0, 0, err
		}

		bonus = b + c

		p.TotalAllowance = float64(p.BaseSalaryAmount) + bonus - p.SalaryAdvanceAmount

		base, _, err := h.service.Wise.Convert(float64(p.BaseSalaryAmount), p.Employee.BaseSalary.Currency.Name, currency.VNDCurrency)
		if err != nil {
			return nil, 0, 0, err
		}
		subTotal += int64(base)
	} else {
		bonus = float64(p.CommissionAmount + p.ProjectBonusAmount)
		p.TotalAllowance = float64(p.BaseSalaryAmount) + bonus - p.SalaryAdvanceAmount
		subTotal += p.BaseSalaryAmount
	}

	projectBonusExplains, err := getProjectBonusExplains(&p)
	if err != nil {
		return nil, 0, 0, err
	}
	p.ProjectBonusExplains = projectBonusExplains

	commissionExplains, err := h.getCommissionExplains(&p, markPaid)
	if err != nil {
		return nil, 0, 0, err
	}
	p.CommissionExplains = commissionExplains

	for i, v := range p.ProjectBonusExplains {
		formattedAmount, err := h.getFormattedAmount(&p, v.Amount)
		if err != nil {
			return nil, 0, 0, err
		}
		p.ProjectBonusExplains[i].FormattedAmount = formattedAmount
	}
	for i, v := range p.CommissionExplains {
		formattedAmount, err := h.getFormattedAmount(&p, v.Amount)
		if err != nil {
			return nil, 0, 0, err
		}
		p.CommissionExplains[i].FormattedAmount = formattedAmount
	}

	quote, err := h.service.Wise.GetPayrollQuotes(c, p.Employee.BaseSalary.Currency.Name, p.TotalAllowance)
	if err != nil {
		h.logger.Error(err, "cannot use wise to get quote")
		return nil, 0, 0, err
	}
	p.TWAmount = quote.SourceAmount
	p.TWRate = quote.Rate
	p.TWFee = quote.Fee

	return &p, bonus, subTotal, nil
}

func (h *handler) getFormattedAmount(p *model.Payroll, amount model.VietnamDong) (string, error) {
	if p.Employee.BaseSalary.Currency.Name != currency.VNDCurrency {
		temp, _, err := h.service.Wise.Convert(float64(amount), currency.VNDCurrency, p.Employee.BaseSalary.Currency.Name)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%.02f", temp), nil
	}
	return amount.String(), nil
}

func getProjectBonusExplains(p *model.Payroll) ([]model.ProjectBonusExplain, error) {
	projectBonusExplains := make([]model.ProjectBonusExplain, 0)
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

func (h *handler) getCommissionExplains(p *model.Payroll, markPaid bool) ([]model.CommissionExplain, error) {
	commissionExplains := make([]model.CommissionExplain, 0)
	err := json.Unmarshal(
		p.CommissionExplain,
		&commissionExplains,
	)
	if err != nil {
		return nil, err
	}
	if markPaid {
		for i := range commissionExplains {
			err := h.store.EmployeeCommission.MarkPaid(h.repo.DB(), commissionExplains[i].ID)
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

func (h *handler) cachePayroll(month, year, batch int, payrolls []model.Payroll) error {
	payrollsBytes, err := json.Marshal(&payrolls)
	if err != nil {
		return err
	}

	cPayroll, err := h.store.CachedPayroll.Get(h.repo.DB(), month, year, batch)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	cPayroll.Month = month
	cPayroll.Year = year
	cPayroll.Batch = batch
	cPayroll.Payrolls = payrollsBytes

	return h.store.CachedPayroll.Set(h.repo.DB(), cPayroll)
}
