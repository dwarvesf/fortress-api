package payroll

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/currency"
	commissionStore "github.com/dwarvesf/fortress-api/pkg/store/employeecommission"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

// calculatePayrolls return list of payrolls for all given users in batchDate
func (h *handler) calculatePayrolls(users []*model.Employee, batchDate time.Time, simplifyNotes bool) (res []*model.Payroll, err error) {
	batch := batchDate.Day()
	// HACK: sneak quang into this
	if batch == 1 {
		quang, err := h.store.Employee.OneByEmail(h.repo.DB(), "quang@d.foundation")
		if err != nil {
			h.logger.Error(err, "Can't insert quang into 1st payroll batch")
			return nil, err
		}
		users = append(users, quang)
	}
	if batch == 15 {
		inno, err := h.store.Employee.OneByEmail(h.repo.DB(), "mytx@d.foundation")
		if err != nil {
			h.logger.Error(err, "Can't insert inno into 15th payroll batch")
			return nil, err
		}
		users = append(users, inno)
	}
	dueDate := batchDate.AddDate(0, 1, 0)
	var expenses []bcModel.Todo
	woodlandID := consts.PlaygroundID
	expenseID := consts.PlaygroundExpenseTodoID
	opsID := consts.PlaygroundID
	opsExpenseID := consts.PlaygroundExpenseTodoID
	approver := consts.NamNguyenBasecampID

	if h.config.Env == "prod" {
		woodlandID = consts.WoodlandID
		expenseID = consts.ExpenseTodoID
		opsID = consts.OperationID
		opsExpenseID = consts.OpsExpenseTodoID
		approver = consts.HanBasecampID
	}

	// NocoDB stores all expenses in one table, Basecamp uses separate lists
	if h.service.Basecamp != nil {
		// Basecamp flow: fetch ops and team expenses separately
		opsTodoLists, err := h.service.PayrollExpenseProvider.GetAllInList(opsExpenseID, opsID)
		if err != nil {
			h.logger.Error(err, "can't get ops expense todo")
			return nil, err
		}

		for _, exps := range opsTodoLists {
			isApproved := false
			cmts, err := h.service.Basecamp.Comment.Gets(opsID, exps.ID)
			if err != nil {
				h.logger.Error(err, "can't get basecamp approved message")
				return nil, err
			}

			for _, cmt := range cmts {
				if cmt.Creator.ID == approver && strings.Contains(strings.ToLower(cmt.Content), "approve") {
					isApproved = true
					break
				}
			}

			if isApproved {
				expenses = append(expenses, exps)
			}
		}

		// get team expenses
		todolists, err := h.service.PayrollExpenseProvider.GetGroups(expenseID, woodlandID)
		if err != nil {
			h.logger.Error(err, "can't get groups expense")
			return nil, err
		}

		for i := range todolists {
			e, err := h.service.PayrollExpenseProvider.GetAllInList(todolists[i].ID, woodlandID)
			if err != nil {
				h.logger.Error(err, "can't get expense todo")
				return nil, err
			}

			for j := range e {
				isApproved := false
				cmts, err := h.service.Basecamp.Comment.Gets(woodlandID, e[j].ID)
				if err != nil {
					h.logger.Error(err, "can't get basecamp approved message")
					return nil, err
				}
				for k := range cmts {
					if cmts[k].Creator.ID == approver && strings.Contains(strings.ToLower(cmts[k].Content), "approve") {
						isApproved = true
						break
					}
				}
				if isApproved {
					expenses = append(expenses, e[j])
				}
			}
		}
	} else {
		// NocoDB flow: fetch from expense_submissions table
		h.logger.Debug("Fetching approved expense submissions from NocoDB (expense_submissions table)")
		allExpenses, err := h.service.PayrollExpenseProvider.GetAllInList(opsExpenseID, opsID)
		if err != nil {
			h.logger.Error(err, "can't get expense submissions from NocoDB")
			return nil, err
		}
		h.logger.Debug(fmt.Sprintf("Fetched %d expense submissions from NocoDB", len(allExpenses)))
		expenses = append(expenses, allExpenses...)

		// Also fetch accounting todos from accounting_todos table (only for NocoDB, not Notion)
		if h.service.PayrollAccountingTodoProvider != nil {
			h.logger.Debug("Fetching accounting todos from NocoDB (accounting_todos table)")
			accountingTodos, err := h.service.PayrollAccountingTodoProvider.GetAllInList(opsExpenseID, opsID)
			if err != nil {
				h.logger.Error(err, "can't get accounting todos from NocoDB")
				return nil, err
			}
			h.logger.Debug(fmt.Sprintf("Fetched %d accounting todos from NocoDB", len(accountingTodos)))
			expenses = append(expenses, accountingTodos...)
			h.logger.Debug(fmt.Sprintf("Total expenses (submissions + accounting todos): %d", len(expenses)))
		} else {
			h.logger.Debug("PayrollAccountingTodoProvider is nil (Notion flow), skipping accounting todos fetch")
		}
	}

	// Fetch accounting todos from Basecamp (NocoDB already fetched them above)
	if h.service.Basecamp != nil {
		h.logger.Debug("Fetching accounting todos from Basecamp")
		accountingExpenses, err := h.getAccountingExpense(batch)
		if err != nil {
			h.logger.Error(err, "can't get accounting todo")
			return nil, err
		}

		h.logger.Debug(fmt.Sprintf("Fetched %d accounting expenses from Basecamp", len(accountingExpenses)))
		expenses = append(expenses, accountingExpenses...)
		h.logger.Debug(fmt.Sprintf("Total expenses after appending accounting todos: %d", len(expenses)))
	}

	for i, u := range users {
		if users[i].BaseSalary.Currency == nil {
			continue
		}
		var total model.VietnamDong
		var baseSalary, contract int64
		if users[i].BaseSalary.Batch != batchDate.Day() {
			if users[i].TeamEmail != "quang@d.foundation" && users[i].TeamEmail != "mytx@d.foundation" {
				continue
			} else {
				users[i].BaseSalary.PersonalAccountAmount = 0
				users[i].BaseSalary.CompanyAccountAmount = 0
				users[i].BaseSalary.ContractAmount = 0
			}
		}

		// TODO...
		// try to calculate if user start/end after/before the payroll
		// fallback to default
		baseSalary, contract, _ = tryPartialCalculation(batchDate, dueDate, *u.JoinedDate, u.LeftDate, users[i].BaseSalary.PersonalAccountAmount, users[i].BaseSalary.CompanyAccountAmount)

		var bonus, commission, reimbursementAmount model.VietnamDong
		var bonusExplains, commissionExplains []model.CommissionExplain

		bonus, commission, reimbursementAmount, bonusExplains, commissionExplains = h.getBonus(*users[i], batchDate, expenses, simplifyNotes)

		commBytes, err := json.Marshal(&commissionExplains)
		if err != nil {
			return nil, err
		}

		bonusBytes, err := json.Marshal(&bonusExplains)
		if err != nil {
			return nil, err
		}

		if users[i].BaseSalary.Currency.Name != currency.VNDCurrency {
			c, _, err := h.service.Wise.Convert(float64(commission), currency.VNDCurrency, users[i].BaseSalary.Currency.Name)
			if err != nil {
				return nil, err
			}
			b, _, err := h.service.Wise.Convert(float64(bonus), currency.VNDCurrency, users[i].BaseSalary.Currency.Name)
			if err != nil {
				return nil, err
			}
			temp, _, err := h.service.Wise.Convert(float64(baseSalary+contract)+b+c, users[i].BaseSalary.Currency.Name, currency.VNDCurrency)
			if err != nil {
				return nil, err
			}
			total = model.NewVietnamDong(int64(temp))
		} else {
			temp := model.NewVietnamDong(baseSalary)
			baseSalary = int64(temp.Format())
			total = model.NewVietnamDong(baseSalary+contract) + bonus + commission
		}
		if total == 0 {
			continue
		}

		// get advance salary
		advanceAmountUSD := 0.0
		advanceSalaries, err := h.store.SalaryAdvance.ListNotPayBackByEmployeeID(h.repo.DB(), u.ID.String())
		if err != nil {
			return nil, err
		}

		for _, as := range advanceSalaries {
			advanceAmountUSD += as.AmountUSD
		}
		advanceAmountVND, _, err := h.service.Wise.Convert(advanceAmountUSD, currency.USDCurrency, currency.VNDCurrency)
		if err != nil {
			return nil, err
		}
		// calculate total minus advance salary

		total = model.NewVietnamDong(int64(float64(total) - advanceAmountVND))

		p := model.Payroll{
			EmployeeID:          users[i].ID,
			Total:               total.Format(),
			BaseSalaryAmount:    baseSalary,
			ConversionAmount:    total.Format() - reimbursementAmount,
			SalaryAdvanceAmount: advanceAmountVND,
			DueDate:             &dueDate,
			Month:               int64(batchDate.Month()),
			Year:                int64(batchDate.Year()),
			CommissionExplain:   commBytes,
			CommissionAmount:    commission,
			ProjectBonusAmount:  bonus,
			ProjectBonusExplain: bonusBytes,
			Employee:            *users[i],
			ContractAmount:      contract,
		}

		res = append(res, &p)
	}

	return res, nil
}

func (h *handler) getBonus(u model.Employee, batchDate time.Time, expenses []bcModel.Todo, simplifyNotes bool) (bonus, commission, reimbursementAmount model.VietnamDong, bonusExplain, commissionExplain []model.CommissionExplain) {
	h.logger.Info("get bonus")
	var explanation string
	bonusRecords, err := h.store.Bonus.GetByUserID(h.repo.DB(), u.ID)
	if err != nil {
		return
	}

	if explanation != "" {
		bonusExplain = append(bonusExplain,
			model.CommissionExplain{
				Amount: 0,
				Month:  int(batchDate.Month()),
				Year:   batchDate.Year(),
				Name:   explanation,
			})
	}
	for i := range bonusRecords {
		bonus += bonusRecords[i].Amount
		bonusExplain = append(bonusExplain,
			model.CommissionExplain{
				Amount: bonusRecords[i].Amount,
				Month:  int(batchDate.Month()),
				Year:   batchDate.Year(),
				Name:   bonusRecords[i].Name,
			})
	}
	for i := range expenses {
		hasReimbursement := false
		for j := range expenses[i].Assignees {
			if expenses[i].Assignees[j].ID == u.BasecampID {
				hasReimbursement = true
				break
			}
		}
		if hasReimbursement {
			name, amount, err := h.getReimbursement(expenses[i].Title)
			if err != nil {
				return
			}
			if amount == 0 {
				continue
			}

			bonus += amount
			reimbursementAmount += amount
			bonusExplain = append(bonusExplain,
				model.CommissionExplain{
					Amount:           amount,
					Month:            int(batchDate.Month()),
					Year:             batchDate.Year(),
					Name:             name,
					BasecampTodoID:   expenses[i].ID,
					BasecampBucketID: expenses[i].Bucket.ID,
					ExternalRef:      expenses[i].AppURL, // Store Notion page UUID for later status update
				})
		}
	}

	commissionQuery := commissionStore.Query{
		EmployeeID: u.ID.String(),
		IsPaid:     false,
	}
	userCommissions, err := h.store.EmployeeCommission.Get(h.repo.DB(), commissionQuery)
	if err != nil {
		return
	}

	for i := range userCommissions {
		commissionYear, commissionMonth, _ := userCommissions[i].CreatedAt.Date()
		commission += userCommissions[i].Amount
		if userCommissions[i].Invoice == nil {
			continue
		}
		// NOTE: Format commission notes based on simplifyNotes parameter
		// simplifyNotes=true: "InvoiceNumber - Bonus" for email template
		// simplifyNotes=false: "InvoiceNumber - DetailedNote" for payrolls/details API
		name := userCommissions[i].Invoice.Number
		if userCommissions[i].Note != "" {
			if simplifyNotes {
				name = fmt.Sprintf("%v - Bonus", name)
			} else {
				name = fmt.Sprintf("%v - %v", name, userCommissions[i].Note)
			}
		}
		commissionExplain = append(commissionExplain,
			model.CommissionExplain{
				ID:     userCommissions[i].ID,
				Amount: userCommissions[i].Amount,
				Month:  int(commissionMonth),
				Year:   commissionYear,
				Name:   name,
			})
	}
	return
}

func tryPartialCalculation(
	batchDate, dueDate, startDate time.Time,
	leftDate *time.Time,
	baseSalary, contract int64,
) (resBase, resContract int64, explanation string) {
	partialStart := batchDate
	partialEnd := dueDate
	isPartial := false
	if checkUserFirstBatch(startDate, batchDate) {
		partialStart = startDate
		isPartial = true
	}
	if leftDate != nil && leftDate.Before(dueDate) {
		partialEnd = *leftDate
		isPartial = true
	}
	if isPartial {
		var temp int64
		temp, explanation, _ = calculatePartialPayroll(partialStart, partialEnd, dueDate, batchDate, baseSalary+contract)
		baseSalary = temp - contract
		if baseSalary < 0 {
			contract += baseSalary
			baseSalary = 0
		}
	}

	return baseSalary, contract, explanation
}

func checkUserFirstBatch(startDate, batchDate time.Time) bool {
	if (startDate.Month() == batchDate.Month() && startDate.Year() == batchDate.Year()) ||
		(startDate.Month() == 12 && batchDate.Month() == 1 && startDate.Year() == batchDate.Year()-1) {
		return true
	}
	return false
}

func calculatePartialPayroll(startDate time.Time, endDate time.Time, dueDate time.Time, lastDueDate time.Time, totalSalary int64) (int64, string, error) {
	dayWorkOfMonth := int64(dueDate.Sub(lastDueDate).Hours() / 24)

	weekendsOfMonth := int64(timeutil.CountWeekendDays(lastDueDate, dueDate))

	// minus the weekends
	dayWorkOfMonth -= weekendsOfMonth

	// get day work in fist batch
	dayWorkOfFirstBatch := int64(endDate.Sub(startDate).Hours() / 24)

	// sum up with end date if left
	if !timeutil.IsSameDay(endDate, dueDate) {
		dayWorkOfFirstBatch++
	}

	// minus the weekends
	weekendsOfFirstBatch := int64(timeutil.CountWeekendDays(startDate, endDate))
	dayWorkOfFirstBatch -= weekendsOfFirstBatch

	if dayWorkOfFirstBatch == dayWorkOfMonth {
		return totalSalary, "", nil
	}

	// calculate salary per date
	// get the actual salary of first batch
	total := dayWorkOfFirstBatch * totalSalary / dayWorkOfMonth
	return total, fmt.Sprintf("Work from %s to %s", startDate.Format("2 Jan"), endDate.Format("2 Jan")), nil
}

func (h *handler) getReimbursement(expense string) (string, model.VietnamDong, error) {
	var amount model.VietnamDong
	splits := strings.Split(expense, "|")
	if len(splits) < 3 {
		return "", 0, nil
	}
	c := strings.TrimSpace(splits[2])

	// Default to VND if currency is empty
	if c == "" {
		c = currency.VNDCurrency
		h.logger.Debug(fmt.Sprintf("Currency empty in expense '%s', defaulting to VND", expense))
	}

	// Parse amount from expense title (format: "reason | amount | currency")
	amountStr := strings.TrimSpace(splits[1])
	bcAmount := h.extractExpenseAmount(amountStr)

	if c != currency.VNDCurrency {
		tempAmount, _, err := h.service.Wise.Convert(float64(bcAmount), c, currency.VNDCurrency)
		if err != nil {
			return "", 0, err
		}
		amount = model.NewVietnamDong(int64(tempAmount))
	} else {
		amount = model.NewVietnamDong(int64(bcAmount))
	}

	return strings.TrimSpace(splits[0]), amount.Format(), nil
}

// extractExpenseAmount parses amount from expense string (provider-agnostic)
func (h *handler) extractExpenseAmount(source string) int {
	if h.service.Basecamp != nil {
		return h.service.Basecamp.ExtractBasecampExpenseAmount(source)
	}
	// Fallback: parse as plain number (for NocoDB which provides clean numeric values)
	source = strings.Replace(source, ".", "", -1) // Remove thousand separators
	source = strings.TrimSpace(source)
	var val int
	_, _ = fmt.Sscanf(source, "%d", &val) // Explicitly ignore error - val defaults to 0 on parse failure
	return val
}

func (h *handler) getAccountingExpense(batch int) (res []bcModel.Todo, err error) {
	accountingID := consts.AccountingID
	accountingTodoID := consts.AccountingTodoID

	if h.config.Env != "prod" {
		accountingID = consts.PlaygroundID
		accountingTodoID = consts.PlaygroundTodoID
	}

	// NocoDB vs Basecamp flow
	if h.service.Basecamp != nil {
		// Basecamp flow: fetch from multiple lists and groups
		h.logger.Debug("Fetching accounting todos from Basecamp (multi-list/group)")

		// get accounting todo list
		lists, err := h.service.PayrollExpenseProvider.GetLists(accountingID, accountingTodoID)
		if err != nil {
			h.logger.Error(err, "can't get list of todo")
			return nil, err
		}

		var wg sync.WaitGroup
		var mu sync.Mutex
		// get all group in each list
		for i := range lists {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				groups, err := h.service.PayrollExpenseProvider.GetGroups(lists[i].ID, accountingID)
				if err != nil {
					h.logger.Error(err, "can't get groups in todo list")
					return
				}

				// filter out group only
				for j := range groups {
					if strings.ToLower(groups[j].Title) == "out" {
						// get all todo in out group
						todos, err := h.service.PayrollExpenseProvider.GetAllInList(groups[j].ID, accountingID)
						if err != nil {
							h.logger.Error(err, "can't get todo in out group")
							return
						}

						// find expense in list of todos
						for k := range todos {
							if len(todos[k].Assignees) == 1 && todos[k].Assignees[0].ID != consts.HanBasecampID {
								mu.Lock()
								res = append(res, todos[k])
								mu.Unlock()
							}
						}
					}
				}
			}(i)
		}
		wg.Wait()
	} else {
		// NocoDB flow: GetAllInList already filters by "out" group and excludes Han
		h.logger.Debug("Fetching accounting todos from NocoDB (single table, pre-filtered)")

		// accountingTodoID acts as the table identifier for NocoDB
		todos, err := h.service.PayrollExpenseProvider.GetAllInList(accountingTodoID, accountingID)
		if err != nil {
			h.logger.Error(err, "can't get accounting todos from NocoDB")
			return nil, err
		}

		h.logger.Debug(fmt.Sprintf("NocoDB returned %d accounting todos", len(todos)))
		for i := range todos {
			h.logger.Debug(fmt.Sprintf("NocoDB accounting todo %d: ID=%d, Title=%s, Assignees=%v", i, todos[i].ID, todos[i].Title, todos[i].Assignees))
		}

		res = todos
	}

	h.logger.Debug(fmt.Sprintf("getAccountingExpense returning %d todos", len(res)))
	return res, nil
}
