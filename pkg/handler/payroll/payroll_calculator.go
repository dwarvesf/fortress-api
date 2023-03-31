package payroll

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/consts"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/currency"
	commissionStore "github.com/dwarvesf/fortress-api/pkg/store/commission"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

// calculatePayrolls return list of payrolls for all given users in batchDate
func calculatePayrolls(h *handler, users []*model.Employee, batchDate time.Time) (res []*model.Payroll, err error) {
	isForecast := false
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
	dueDate := batchDate.AddDate(0, 1, 0)
	var expenses []bcModel.Todo
	woodlandID := consts.PlaygroundID
	expenseID := consts.PlaygroundExpenseTodoID
	opsID := consts.PlaygroundID
	opsExpenseID := consts.PlaygroundExpenseTodoID

	if h.config.Env == "prod" {
		woodlandID = consts.WoodlandID
		expenseID = consts.ExpenseTodoID
		opsID = consts.OperationID
		opsExpenseID = consts.OpsExpenseTodoID
	}

	opsTodolists, err := h.service.Basecamp.Todo.GetAllInList(opsExpenseID, opsID)
	if err != nil {
		h.logger.Error(err, "can't get ops expense todo")
		return nil, err
	}
	for _, exps := range opsTodolists {
		isApproved := false
		cmts, err := h.service.Basecamp.Comment.Gets(opsID, exps.ID)
		if err != nil {
			h.logger.Error(err, "can't get basecamp approved message")
			return nil, err
		}
		for _, cmt := range cmts {
			if cmt.Creator.ID == consts.HanBasecampID && strings.Contains(strings.ToLower(cmt.Content), "approve") {
				isApproved = true
				break
			}
		}
		if isApproved {
			expenses = append(expenses, exps)
		}
	}

	// get team expenses
	todolists, err := h.service.Basecamp.Todo.GetGroups(expenseID, woodlandID)
	if err != nil {
		h.logger.Error(err, "can't get groups expense")
		return nil, err
	}
	for i := range todolists {
		e, err := h.service.Basecamp.Todo.GetAllInList(todolists[i].ID, woodlandID)
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
				if cmts[k].Creator.ID == consts.HanBasecampID && strings.Contains(strings.ToLower(cmts[k].Content), "approve") {
					isApproved = true
					break
				}
			}
			if isApproved {
				expenses = append(expenses, e[j])
			}
		}
	}

	accountingExpenses, err := getAccountingExpense(h, batch)
	if err != nil {
		h.logger.Error(err, "can't get accounting todo")
		return nil, err
	}

	expenses = append(expenses, accountingExpenses...)

	for i, u := range users {
		if users[i].BaseSalary.Currency == nil {
			continue
		}
		var total model.VietnamDong
		var baseSalary, contract int64
		if users[i].BaseSalary.Batch != batchDate.Day() {
			if users[i].TeamEmail != "quang@d.foundation" {
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
		baseSalary, contract, _ = tryParticalCalculation(batchDate, dueDate, *u.JoinedDate, u.LeftDate, users[i].BaseSalary.PersonalAccountAmount, users[i].BaseSalary.CompanyAccountAmount)

		var bonus, commission, reimbursementAmount model.VietnamDong
		var bonusExplains, commissionExplains []model.CommissionExplain

		// TODO...
		// get bonus
		if !isForecast {
			bonus, commission, reimbursementAmount, bonusExplains, commissionExplains = getBonus(h, *users[i], batchDate, expenses)
		}

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

		p := model.Payroll{
			EmployeeID:          users[i].ID,
			Total:               total.Format(),
			BaseSalaryAmount:    baseSalary,
			ConversionAmount:    total.Format() - reimbursementAmount,
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

	if isForecast {
		candidates, err := h.store.Recruitment.GetOffered(h.repo.DB(), batchDate, dueDate)
		if err != nil {
			return nil, err
		}
		for i := range candidates {
			if dueDate.Sub(*candidates[i].OfferStartDate).Hours()/24 < 15 {
				continue
			}
			baseSalary, contract, _ := tryParticalCalculation(batchDate, dueDate, *candidates[i].OfferStartDate, nil, int64(candidates[i].OfferSalary), 0)
			total := model.NewVietnamDong(baseSalary)

			p := model.Payroll{
				Total:            total.Format(),
				BaseSalaryAmount: int64(total.Format()),
				ConversionAmount: total.Format(),
				DueDate:          &dueDate,
				Month:            int64(batchDate.Month()),
				Year:             int64(batchDate.Year()),
				Employee: model.Employee{
					FullName: candidates[i].Name,
					BaseSalary: model.BaseSalary{
						Batch: batchDate.Day(),
						Currency: &model.Currency{
							Name: "VND", // assume the stored offer salary is always VND
						},
					},
				},
				ContractAmount: contract,
			}

			res = append(res, &p)
		}
	}

	return res, nil
}

func getBonus(
	h *handler,
	u model.Employee,
	batchDate time.Time,
	expenses []bcModel.Todo,
) (bonus, commission, reimbursementAmount model.VietnamDong, bonusExplain, commissionExplain []model.CommissionExplain) {
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
			name, amount, err := getReimbursement(h, expenses[i].Title)
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
				})
		}
	}

	commissionQuery := commissionStore.Query{
		EmployeeID: u.ID.String(),
		IsPaid:     false,
	}
	userCommissions, err := h.store.Commission.Get(h.repo.DB(), commissionQuery)
	if err != nil {
		return
	}

	for i := range userCommissions {
		commissionYear, commissionMonth, _ := userCommissions[i].CreatedAt.Date()
		commission += userCommissions[i].Amount
		if userCommissions[i].Invoice == nil {
			continue
		}
		name := userCommissions[i].Invoice.Number
		if userCommissions[i].Note != "" {
			name = fmt.Sprintf("%v - %v", name, userCommissions[i].Note)
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

func tryParticalCalculation(
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

func getReimbursement(h *handler, expense string) (string, model.VietnamDong, error) {
	var amount model.VietnamDong
	splits := strings.Split(expense, "|")
	if len(splits) < 3 {
		return "", 0, nil
	}
	c := strings.TrimSpace(splits[2])
	bcAmount := basecamp.ExtractBasecampExpenseAmount(strings.TrimSpace(splits[1]))
	if c != currency.VNDCurrency {
		tempAmount, _, err := h.service.Wise.Convert(float64(bcAmount), c, currency.VNDCurrency)
		if err != nil {
			return "", 0, err
		}
		amount = model.NewVietnamDong(int64(tempAmount))
	} else {
		amount = model.NewVietnamDong(int64(basecamp.ExtractBasecampExpenseAmount(strings.TrimSpace(splits[1]))))
	}

	return strings.TrimSpace(splits[0]), amount.Format(), nil
}

func getAccountingExpense(h *handler, batch int) (res []bcModel.Todo, err error) {
	accountingID := consts.AccountingID
	accountingTodoID := consts.AccountingTodoID

	if h.config.Env != "prod" {
		accountingID = consts.PlaygroundID
		accountingTodoID = consts.PlaygroundTodoID
	}

	// get accounting todo list
	lists, err := h.service.Basecamp.Todo.GetLists(accountingID, accountingTodoID)
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
			groups, err := h.service.Basecamp.Todo.GetGroups(lists[i].ID, accountingID)
			if err != nil {
				h.logger.Error(err, "can't get groups in todo list")
				return
			}

			// filter out group only
			for j := range groups {
				if strings.ToLower(groups[j].Title) == "out" {
					// get all todo in out grou
					todos, err := h.service.Basecamp.Todo.GetAllInList(groups[j].ID, accountingID)
					if err != nil {
						h.logger.Error(err, "can't get todo in out group")
						return
					}

					// find expense in list of todos
					for k := range todos {
						if len(todos[k].Assignees) == 1 && todos[k].Assignees[0].ID != consts.HanBasecampID {
							// HACK: some todo is specific batch only
							if strings.Contains(todos[k].Title, "Tiền điện") {
								if batch != 1 {
									continue
								}
							} else if strings.Contains(todos[k].Title, "Office Rental") || strings.Contains(todos[k].Title, "CBRE") {
								if batch != 15 {
									continue
								}
							}
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

	return res, nil
}
