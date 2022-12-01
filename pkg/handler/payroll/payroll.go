package payroll

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/store/payroll"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/view"
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

// ListPayrollsMonthly godoc
// @Summary List employees' payroll.
// @Description Get employees' payroll. We can get the next batch payroll list with query next=true or specified batches.
// @Tags Payroll
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param next query string false "get next batch payroll flag (true)"
// @Param batch query int false "date of the specific batch (1/15). If not specified, both batch payroll lists are returned"
// @Param month query int false " month of the specific batch"
// @Param year query int false "year of the specific batch. If not specified, get current year."
// @Success 200 {object} view.PayrollList
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /payrolls [get]
func (h *handler) ListPayrollsMonthly(c *gin.Context) {
	// 1.1 parse year, month, batch from param
	query := GetListPayrollInput{}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "List",
		"query":   query,
	})

	// 1.2 validate query
	if err := query.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// 2.1 query next month payroll
	if query.Next == "true" {
		res, err := h.listPayrollsDetails(0, 0, 0)
		if err != nil {
			l.Error(err, "list next payroll failed")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, query, ""))
			return
		}
		c.JSON(http.StatusOK, view.CreateResponse(res, nil, nil, nil, ""))
		return
	}

	// 2.2 or query specified month payroll
	res, err := h.listPayrollsDetails(query.Year, query.Month, query.Batch)
	if err != nil {
		l.Error(err, "list payroll failed")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}
	c.JSON(http.StatusOK, view.CreateResponse(res, nil, nil, nil, ""))
}

// listPayrollsDetails list the payrolls of the provided month (year, time, batch) and email (optional)
func (h *handler) listPayrollsDetails(year int, month time.Month, batch int) (*view.PayrollList, error) {
	n := time.Now()
	if year == 0 && month == 0 && batch == 0 {
		date, err := h.store.Payroll.GetLatestCommitTime(h.repo.DB())
		if err != nil {
			date = &n
		}

		year, month, batch = date.Date()
		if batch == 1 {
			batch = 15
			year, month = utils.LastYearMonth(year, month)
		} else {
			batch = 1
		}
	} else {
		year, month = utils.LastYearMonth(year, month)
	}
	var payrolls []model.Payroll
	var err error
	for _, b := range []int{1, 15} {
		if batch != 0 && batch != b {
			continue
		}
		input := payroll.PayrollInput{
			Year:  year,
			Month: month,
			Day:   b,
		}
		batchDate := time.Date(year, month, b, 0, 0, 0, 0, time.UTC)
		payrolls, err = h.store.Payroll.List(h.repo.DB(), input)
		if err != nil {
			return nil, err
		}
		// specified batch
		if batch != 0 {
			payrollMap := map[string]model.BaseModel{}
			if len(payrolls) != 0 {
				for i := range payrolls {
					payrollMap[payrolls[i].EmployeeID.String()] = payrolls[i].BaseModel
				}
			}

			sob := time.Date(year, month, b, 0, 0, 0, 0, n.Location())
			input := employee.GetAllInput{
				StartOfBatch: &sob,
				Preload:      true,
			}
			employees, _, err := h.store.Employee.All(h.repo.DB(), input, model.Pagination{})
			if err != nil {
				return nil, err
			}
			payrolls, err = h.calculatePayroll(employees, batchDate)
			if err != nil {
				return nil, err
			}
			if len(payrollMap) != 0 {
				for i := range payrolls {
					payrolls[i].BaseModel = payrollMap[payrolls[i].EmployeeID.String()]
				}
			}
		}

		for i := range payrolls {
			if !payrolls[i].IsPaid {
				quote, err := h.service.Wise.GetPayrollQuotes(payrolls[i].Employee.WiseCurrency, payrolls[i].Employee.EmployeeBaseSalary.Currency.Name, float64(payrolls[i].Total))
				if err != nil {
					return nil, err
				}
				payrolls[i].WiseAmount = quote.SourceAmount
				payrolls[i].WiseFee = quote.Fee
				payrolls[i].WiseRate = quote.Rate
			}
		}
		err = h.store.Payroll.Save(h.repo.DB(), payrolls)
		if err != nil {
			return nil, err
		}
	}

	return view.ToPayrollList(payrolls, month, year), nil
}

// calculatePayroll calculate payrolls of each user from base salary, contract, commission, bonus and reimbursement
func (h *handler) calculatePayroll(employees []*model.Employee, batchDate time.Time) ([]model.Payroll, error) {
	// HACK: sneak quang into this
	if batchDate.Day() == 1 {
		quang, err := h.store.Employee.OneByTeamEmail(h.repo.DB(), "quang@d.foundation")
		if err != nil {
			return nil, err
		}
		employees = append(employees, quang)
	}
	var res []model.Payroll
	var expenses []bcModel.Todo
	woodlandID := bcModel.BasecampPlaygroundID
	expenseID := bcModel.BasecampPlaygroundExpenseTodoID
	opsID := bcModel.BasecampPlaygroundID
	opsExpenseID := bcModel.BasecampPlaygroundExpenseTodoID
	if h.config.Env == "prod" {
		woodlandID = bcModel.BasecampWoodlandID
		expenseID = bcModel.BasecampExpenseTodoID
		opsID = bcModel.BasecampOpsID
		opsExpenseID = bcModel.BasecampOpsExpenseTodoID
	}
	// get team expenses
	todolists, err := h.service.Bc3.Todo.GetGroups(expenseID, woodlandID)
	if err != nil {
		return nil, err
	}
	for i := range todolists {
		e, err := h.service.Bc3.Todo.GetAllInList(todolists[i].ID, woodlandID)
		if err != nil {
			return nil, err
		}
		for j := range e {
			isApproved := false
			cmts, err := h.service.Bc3.Comment.Gets(e[j].ID, woodlandID)
			if err != nil {
				return nil, err
			}
			for k := range cmts {
				if cmts[k].Creator.ID == bcModel.HanBasecampID && strings.Contains(strings.ToLower(cmts[k].Content), "approve") {
					isApproved = true
					break
				}
			}
			if isApproved {
				expenses = append(expenses, e[j])
			}
		}
	}


	opsTodolists, err := h.service.Bc3.Todo.GetAllInList(opsExpenseID, opsID)
	if err != nil {
		return nil, err
	}
	for _, exps := range opsTodolists {
		isApproved := false
		cmts, err := h.service.Bc3.Comment.Gets(opsID, exps.ID)
		if err != nil {
			return nil, err
		}
		for _, cmt := range cmts {
			if cmt.Creator.ID == bcModel.HanBasecampID && strings.Contains(strings.ToLower(cmt.Content), "approve") {
				isApproved = true
				break
			}
		}
		if isApproved {
			expenses = append(expenses, exps)
		}
	}
	accountingExpenses, err := h.getAccountingExpense(batchDate.Day())
	if err != nil {
		return nil, err
	}
	expenses = append(expenses, accountingExpenses...)

	for _, emp := range employees {
		if emp.EmployeeBaseSalary.PayrollBatch != batchDate.Day() {
			if emp.TeamEmail != "quang@d.foundation" {
				continue
			} else {
				emp.EmployeeBaseSalary.PersonalAmount = 0
				emp.EmployeeBaseSalary.ContractAmount = 0
			}
		}
		dueDate := batchDate.AddDate(0, 1, 0)
		personal, contract, exp := tryParticalCalculation(batchDate, dueDate, *emp.JoinedDate, emp.LeftDate, emp.EmployeeBaseSalary.PersonalAmount, emp.EmployeeBaseSalary.ContractAmount)
		bonus, reimbursement, commission, bonusExp, commExp, err := h.calculateBonus(*emp, batchDate, expenses)
		if err != nil {
			return nil, err
		}
		exps := []string{}
		if exp != "" {
			exps = append(exps, exp)
		}
		for _, ex := range bonusExp {
			exps = append(exps, ex.Name)
		}
		for _, ex := range commExp {
			exps = append(exps, ex.Name)
		}
		exp = strings.Join(exps, "\n")
		bonusBytes, err := json.Marshal(&bonusExp)
		if err != nil {
			return nil, err
		}
		commBytes, err := json.Marshal(&commExp)
		if err != nil {
			return nil, err
		}
		total := model.NewVietnamDong(personal + contract)
		total = total.Format() + bonus + reimbursement + commission
		res = append(res, model.Payroll{
			EmployeeID:        emp.ID,
			Total:             total,
			AccountedAmount:   total - reimbursement,
			DueDate:           dueDate,
			Month:             batchDate.Month(),
			Year:              batchDate.Year(),
			PersonalAmount:    personal,
			ContractAmount:    contract,
			Bonus:             bonus + reimbursement,
			BonusExplain:      model.JSON(bonusBytes),
			Commission:        commission,
			CommissionExplain: model.JSON(commBytes),
			Notes:             exp,
			IsPaid:            false,
			Employee:          *emp,
		})
	}
	return res, nil
}

func (h *handler) getAccountingExpense(batch int) (res []bcModel.Todo, err error) {
	accountingID := bcModel.BasecampAccountingID
	accountingTodoID := bcModel.BasecampAccountingTodoID
	if h.config.Env != "prod" {
		accountingID = bcModel.BasecampPlaygroundID
		accountingTodoID = bcModel.BasecampPlaygroundTodoID
	}

	// get accounting todo list
	lists, err := h.service.Bc3.Todo.GetLists(accountingTodoID, accountingID)
	if err != nil {
		return nil, err
	}
	// get all group in each list
	for i := range lists {
		groups, err := h.service.Bc3.Todo.GetGroups(lists[i].ID, accountingID)
		if err != nil {
			return nil, err
		}

		// filter out group only
		for j := range groups {
			if strings.ToLower(groups[j].Title) == "out" {
				// get all todo in out group
				todos, err := h.service.Bc3.Todo.GetAllInList(groups[j].ID, accountingID)
				if err != nil {
					return nil, err
				}

				// find expense in list of todos
				for k := range todos {
					if len(todos[k].Assignees) == 1 && todos[k].Assignees[0].ID != bcModel.HanBasecampID {
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

						res = append(res, todos[k])
					}
				}
			}
		}
	}

	return res, nil
}

// calculateBonus get bonus from db, reimbursement from Basecamp an commission (?). Reimbursement explain is grouped into bonus explain
func (h *handler) calculateBonus(emp model.Employee, batchDate time.Time, expenses []bcModel.Todo) (model.VietnamDong, model.VietnamDong, model.VietnamDong, []model.PayrollExplain, []model.PayrollExplain, error) {
	var bonus, reimbursement, commission model.VietnamDong
	var bonusExps, commExps []model.PayrollExplain
	bonuses, err := h.store.EmployeeBonus.GetByEmployeeID(h.repo.DB(), emp.ID)
	if err != nil {
		return 0, 0, 0, nil, nil, err
	}
	for _, b := range bonuses {
		bonus += b.Amount
		bonusExps = append(bonusExps, model.PayrollExplain{
			Amount: b.Amount,
			Month:  batchDate.Month(),
			Year:   batchDate.Year(),
			Name:   b.Name,
		})
	}
	for _, e := range expenses {
		hasReimbursement := false
		for _, id := range e.AssigneeIDs {
			if fmt.Sprint(id) == emp.BasecampID {
				hasReimbursement = true
				break
			}
		}
		if hasReimbursement {
			name, amount, err := h.getReimbursementAmount(e.Title)
			if err != nil {
				return 0, 0, 0, nil, nil, err
			}
			if amount == 0 {
				continue
			}
			reimbursement += amount
			bonusExps = append(bonusExps, model.PayrollExplain{
				Amount:           amount,
				Month:            batchDate.Month(),
				Year:             batchDate.Year(),
				Name:             name,
				BasecampTodoID:   e.ID,
				BasecampBucketID: e.Bucket.ID,
			})
		}
	}
	comm, err := h.store.Commission.ListEmployeeCommissions(h.repo.DB(), emp.ID, false)
	if err != nil {
		return 0, 0, 0, nil, nil, err
	}
	for _, c := range comm {
		temp_ammount := model.NewVietnamDong(c.Amount)
		commission += temp_ammount.Format()
		commExps = append(commExps, model.PayrollExplain{
			ID:     c.ID,
			Amount: temp_ammount.Format(),
			Month:  batchDate.Month(),
			Year:   batchDate.Year(),
			Name:   c.Note,
		})
	}
	return bonus, reimbursement, commission, bonusExps, commExps, nil
}

// getReimbursementAmount extract amount from expense title `Something | 100.000 | VND`
func (h *handler) getReimbursementAmount(title string) (string, model.VietnamDong, error) {
	var amount model.VietnamDong
	splits := strings.Split(title, "|")
	if len(splits) < 3 {
		return "", 0, nil
	}
	c := strings.TrimSpace(splits[2])
	bcAmount := basecamp.ExtractBasecampExpenseAmount(strings.TrimSpace(splits[1]))
	if c != model.VNDCurrency {
		tempAmount, _, err := h.service.Wise.Convert(float64(bcAmount), c, model.VNDCurrency)
		if err != nil {
			return "", 0, err
		}
		amount = model.NewVietnamDong(int64(tempAmount))
	} else {
		amount = model.NewVietnamDong(int64(basecamp.ExtractBasecampExpenseAmount(strings.TrimSpace(splits[1]))))
	}

	return strings.TrimSpace(splits[0]), amount.Format(), nil
}

// tryParticalCalculation calcuate the partial base salary and contract of those who did not work the full month (join late or leave early)
func tryParticalCalculation(batchDate, dueDate, startDate time.Time, leftDate *time.Time, baseSalary, contract int64) (int64, int64, string) {
	var explaination string
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
		temp, explaination = calculatePartialPayroll(partialStart, partialEnd, dueDate, batchDate, baseSalary+contract)
		baseSalary = temp - contract
		if baseSalary < 0 {
			contract += baseSalary
			baseSalary = 0
		}
	}

	return baseSalary, contract, explaination
}

func calculatePartialPayroll(startDate, endDate, dueDate, lastDueDate time.Time, totalSalary int64) (int64, string) {
	dayWorkOfMonth := int(dueDate.Sub(lastDueDate).Hours() / 24)

	weekendsOfMonth := utils.CountWeekendDays(lastDueDate, dueDate)

	// minus the weekends
	dayWorkOfMonth -= weekendsOfMonth

	// get day work in fist batch
	dayWorkOfFirstBatch := int(endDate.Sub(startDate).Hours() / 24)

	// sum up with end date if left
	if !utils.IsSameDay(endDate, dueDate) {
		dayWorkOfFirstBatch++
	}

	// minus the weekends
	weekendsOfFirstBatch := utils.CountWeekendDays(startDate, endDate)
	dayWorkOfFirstBatch -= weekendsOfFirstBatch

	if dayWorkOfFirstBatch == dayWorkOfMonth {
		return totalSalary, ""
	}

	// calculate salary per date
	// get the actual salary of first batch
	total := int64(dayWorkOfFirstBatch) * totalSalary / int64(dayWorkOfMonth)
	return total, fmt.Sprintf("Work from %s to %s", startDate.Format("2 Jan"), endDate.Format("2 Jan"))
}

func checkUserFirstBatch(startDate, batchDate time.Time) bool {
	if (startDate.Month() == batchDate.Month() && startDate.Year() == batchDate.Year()) ||
		(startDate.Month() == 12 && batchDate.Month() == 1 && startDate.Year() == batchDate.Year()-1) {
		return true
	}
	return false
}

// CommitPayroll godoc
// @Summary Commit employees' payroll.
// @Description Commit employees' payroll, store and mark payrolls as paid and send payroll email.
// @Tags Payroll
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param batch query int true "date of the specific batch (1/15)"
// @Param month query int true " month of the specific batch"
// @Param year query int true "year of the specific batch"
// @Success 200
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /payrolls [post]
func (h *handler) CommitPayroll(c *gin.Context) {
	// 1.1 parse year, month, batch from param
	query := GetListPayrollInput{}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "List",
		"query":   query,
	})

	// 1.2 validate query
	if err := query.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	year, month := utils.LastYearMonth(query.Year, query.Month)
	if month > time.Now().Month() && month != 12 && time.Now().Month() != 1 {
		l.Error(ErrInvalidPayrollDate, "invalid batch")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidPayrollDate, nil, ""))
		return
	}
	tax, err := strconv.ParseFloat(c.Query("tax"), 64)
	if err != nil {
		tax = 0
	}
	err = h.commitPayroll(year, month, query.Batch, tax)
	if err != nil {
		l.Error(err, "commit payroll failed")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	c.JSON(http.StatusOK, nil)
}

// commitPayroll mark payroll as paid
func (h *handler) commitPayroll(year int, month time.Month, batch int, tax float64) error {
	for _, b := range []int{1, 15} {
		if batch != b {
			continue
		}
		pq := payroll.PayrollInput{
			Year:        year,
			Month:       month,
			Day:         batch,
			IsNotCommit: true,
		}
		payrolls, err := h.store.Payroll.List(h.repo.DB(), pq)
		if err != nil {
			return err
		}
		if len(payrolls) == 0 {
			return ErrPayrollNotFound
		}
		for i := range payrolls {
			payrolls[i].IsPaid = true
			if err = h.markBonusAsDone(&payrolls[i]); err != nil {
				return err
			}
		}
		err = h.store.Payroll.Save(h.repo.DB(), payrolls)
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
				if h.config.Env == "prod" {
					c <- &p
				}
			}(payroll)
		}
		wg.Wait()
		c <- nil
		go h.activateGmailQueue(c, tax)

	}
	return nil
}

// markBonusAsDone close reimbursement, accounting and commission record
func (h *handler) markBonusAsDone(p *model.Payroll) error {
	var bonusExplains, commissionExplains []model.PayrollExplain
	err := json.Unmarshal(p.BonusExplain, &bonusExplains)
	if err != nil {
		return err
	}
	err = json.Unmarshal(p.CommissionExplain, &commissionExplains)
	if err != nil {
		return err
	}
	for _, explain := range bonusExplains {
		// reimbursement
		if explain.BasecampBucketID != 0 && explain.BasecampTodoID != 0 {
			woodlandID := bcModel.BasecampPlaygroundID
			accountingID := bcModel.BasecampPlaygroundID
			if h.config.Env == "prod" {
				woodlandID = bcModel.BasecampWoodlandID
				accountingID = bcModel.BasecampAccountingID
			}
			switch explain.BasecampBucketID {
			case woodlandID:
				err := h.service.Bc3.Todo.Complete(explain.BasecampBucketID,
					explain.BasecampTodoID)
				if err != nil {
					return err
				}
			case accountingID:
				mention := basecamp.BasecampMention(p.Employee.BasecampAttachableSGID)
				msg := fmt.Sprintf("Amount has been deposited in your payroll %v", mention)
				h.service.Bc3.Comment.Create(explain.BasecampTodoID,
					explain.BasecampBucketID,
					basecamp.BuildCompletedComment(msg))
			}
		}
	}
	for _, explain := range commissionExplains {
		err := h.store.Commission.CloseEmployeeCommission(h.repo.DB(), explain.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *handler) activateGmailQueue(p chan *model.Payroll, tax float64) {
	d := time.NewTicker(time.Second)
	for range d.C {
		c := <-p
		if c == nil {
			return
		}
		err := h.service.GoogleMail.SendPayrollPaidMail(c, tax)
		if err != nil {
			return

		}
	}
}
