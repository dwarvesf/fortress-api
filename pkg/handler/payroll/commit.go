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

	"github.com/dwarvesf/fortress-api/pkg/handler/payroll/errs"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/currency"
	"github.com/dwarvesf/fortress-api/pkg/store/payroll"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

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
		l.Errorf(err, "failed to process commit payroll for batch date", "date", batch)
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
	for _, b := range []model.Batch{model.FirstBatch, model.SecondBatch} {
		if batch != int(b) {
			continue
		}
		q := payroll.GetListPayrollInput{
			Month: month,
			Year:  year,
			Day:   int(b),
		}
		if email != "" {
			u, err := h.store.Employee.OneByEmail(h.repo.DB(), email)
			if err != nil {
				return err
			}
			q.UserID = u.ID.String()
		}

		batchDate := time.Date(year, time.Month(month), int(b), 0, 0, 0, 0, time.UTC)
		var payrolls []model.Payroll
		cPayroll, err := h.store.CachedPayroll.Get(h.repo.DB(), month, year, batch)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
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
			err = h.markBonusAsDone(&payrolls[i])
			if err != nil {
				return err
			}

			// hacky way to mark done commission
			if _, err := h.getCommissionExplains(&payrolls[i], true); err != nil {
				return err
			}

			// update is_pay_back to true if the employee has payback
			if payrolls[i].SalaryAdvanceAmount != 0 {
				salaryAdvances, err := h.store.SalaryAdvance.ListNotPayBackByEmployeeID(h.repo.DB(), payrolls[i].Employee.ID.String())
				if err != nil {
					return err
				}
				for _, salaryAdvance := range salaryAdvances {
					now := time.Now()
					salaryAdvance.IsPaidBack = true
					salaryAdvance.PaidAt = &now

					err = h.store.SalaryAdvance.Save(h.repo.DB(), &salaryAdvance)
					if err != nil {
						return err
					}
				}
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
		for _, pr := range payrolls {
			go func(p model.Payroll) {
				defer wg.Done()
				if h.config.Env == "prod" || p.Employee.TeamEmail == "quang@d.foundation" || p.Employee.TeamEmail == "huy@d.foundation" {
					c <- &p
				}
			}(pr)
		}
		wg.Wait()

		c <- nil
		go h.activateGmailQueue(c)

		err = h.storePayrollTransaction(payrolls, batchDate)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *handler) storePayrollTransaction(p []model.Payroll, batchDate time.Time) error {
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
		var organization = "Dwarves Foundation"
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
			cur, err := h.store.Currency.GetByName(h.repo.DB(), currency.VNDCurrency)
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
				Currency:         cur.Name,
				Date:             &now,
				CurrencyID:       &cur.ID,
				Organization:     organization,
				Metadata:         bonusBytes,
				ConversionRate:   1,
				Type:             t,
			}
			transactions = append(transactions, &bonusTransaction)
		}
	}

	return h.storeMultipleTransaction(transactions)
}

func (h *handler) storeMultipleTransaction(transactions []*model.AccountingTransaction) error {
	if err := h.store.Accounting.CreateMultipleTransaction(h.repo.DB(), transactions); err != nil {
		return err
	}
	return nil
}

func (h *handler) markBonusAsDone(p *model.Payroll) error {
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
				h.worker.Enqueue(bcModel.BasecampCommentMsg, cm)
			}
		}
	}
	return nil
}

func (h *handler) activateGmailQueue(p chan *model.Payroll) {
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
