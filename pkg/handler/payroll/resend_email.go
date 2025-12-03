package payroll

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/handler/payroll/errs"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store/payroll"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

func (h *handler) resendPayrollEmailHandler(month, year, batch int, email string) error {
	l := h.logger.Fields(logger.Fields{
		"handler": "payroll",
		"method":  "resendPayrollEmailHandler",
		"month":   month,
		"year":    year,
		"batch":   batch,
		"email":   email,
	})

	l.Info("start resend payroll email handler")

	month, year = timeutil.LastMonthYear(month, year)
	if month > int(time.Now().Month()) {
		if month != 12 && time.Now().Month() != 1 {
			l.Error(errors.New("cannot resend email for future payroll"), "invalid month/year")
			return errors.New("cannot resend email for future payroll")
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
				l.Errorf(err, "failed to find employee by email", "email", email)
				return err
			}
			q.UserID = u.ID.String()
		}

		var payrolls []model.Payroll
		cPayroll, err := h.store.CachedPayroll.Get(h.repo.DB(), month, year, batch)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				l.Error(errs.ErrPayrollNotSnapshotted, "payroll not snapshotted")
				return errs.ErrPayrollNotSnapshotted
			}
			l.Errorf(err, "failed to get cached payroll")
			return err
		}

		l.Info(fmt.Sprintf("found cached payroll for month=%d, year=%d, batch=%d", month, year, batch))

		err = json.Unmarshal(cPayroll.Payrolls, &payrolls)
		if err != nil {
			l.Errorf(err, "failed to unmarshal payrolls")
			return err
		}

		l.Info(fmt.Sprintf("found %d payrolls to resend emails", len(payrolls)))

		// Populate FormattedAmount for commission and bonus explains
		for i := range payrolls {
			// CommissionExplains and ProjectBonusExplains should already be populated from JSON unmarshaling
			// If not, they will remain empty arrays

			l.Debug(fmt.Sprintf("processing payroll for employee %s, commissions: %d, bonuses: %d",
				payrolls[i].Employee.TeamEmail,
				len(payrolls[i].CommissionExplains),
				len(payrolls[i].ProjectBonusExplains)))

			for j, v := range payrolls[i].ProjectBonusExplains {
				formattedAmount, err := h.getFormattedAmount(&payrolls[i], v.Amount)
				if err != nil {
					l.Errorf(err, "failed to format project bonus amount", "employee", payrolls[i].Employee.TeamEmail)
					continue
				}
				payrolls[i].ProjectBonusExplains[j].FormattedAmount = formattedAmount
			}
			for j, v := range payrolls[i].CommissionExplains {
				// Simplify commission notes for email payslip
				// NOTE: Convert detailed notes (e.g., "2025104-KAFI-009 - Hiring - Nguyễn Hoàng Anh")
				// to simplified format (e.g., "2025104-KAFI-009 - Bonus") for email template
				name := v.Name
				if strings.Contains(name, " - ") {
					parts := strings.SplitN(name, " - ", 2)
					payrolls[i].CommissionExplains[j].Name = parts[0] + " - Bonus"
				}

				formattedAmount, err := h.getFormattedAmount(&payrolls[i], v.Amount)
				if err != nil {
					l.Errorf(err, "failed to format commission amount", "employee", payrolls[i].Employee.TeamEmail)
					continue
				}
				l.Debug(fmt.Sprintf("formatted commission %s: amount=%d, formatted=%s", payrolls[i].CommissionExplains[j].Name, v.Amount, formattedAmount))
				payrolls[i].CommissionExplains[j].FormattedAmount = formattedAmount
			}
		}

		// Filter by specific email if provided
		if email != "" {
			var filteredPayrolls []model.Payroll
			for _, p := range payrolls {
				if p.Employee.TeamEmail == email {
					filteredPayrolls = append(filteredPayrolls, p)
				}
			}
			payrolls = filteredPayrolls
			l.Info(fmt.Sprintf("filtered to %d payroll(s) for email %s", len(payrolls), email))
		}

		if len(payrolls) == 0 {
			l.Error(errors.New("no payrolls found"), "no payrolls to send emails for")
			return errors.New("no payrolls found for the specified criteria")
		}

		// Using WaitGroup go routines to SendPayrollPaidEmail
		var wg sync.WaitGroup
		wg.Add(len(payrolls))
		c := make(chan *model.Payroll, len(payrolls)+1)
		for _, pr := range payrolls {
			go func(p model.Payroll) {
				defer wg.Done()
				l.Info(fmt.Sprintf("queuing email for %s", p.Employee.TeamEmail))
				c <- &p
			}(pr)
		}
		wg.Wait()

		c <- nil
		go h.activateGmailQueue(c)

		l.Info("payroll emails queued successfully")
	}

	return nil
}
