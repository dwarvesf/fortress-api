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
	"github.com/dwarvesf/fortress-api/pkg/service/nocodb"
	expensestore "github.com/dwarvesf/fortress-api/pkg/store/expense"
	"github.com/dwarvesf/fortress-api/pkg/store/payroll"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// ExpenseSubmissionData represents an expense submission extracted from payroll
type ExpenseSubmissionData struct {
	RecordID     string
	EmployeeID   model.UUID
	EmployeeName string
	Amount       float64
	Currency     string
	Description  string
}

// AccountingTodoData represents an accounting todo extracted from payroll
type AccountingTodoData struct {
	TodoID       int
	EmployeeID   model.UUID
	EmployeeName string
	Amount       float64
	Currency     string
	Description  string
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
	emailOnly := c.Query("email_only") == "true"

	l.Info(fmt.Sprintf("commit payroll - email_only: %v", emailOnly))

	if emailOnly {
		err = h.resendPayrollEmailHandler(int(month), int(year), int(batch), email)
		if err != nil {
			l.Errorf(err, "failed to resend payroll emails for batch date", "date", batch)
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}

		err = h.controller.Discord.Log(model.LogDiscordInput{
			Type: "payroll_email_resent",
			Data: map[string]interface{}{
				"batch_number": strconv.Itoa(int(batch)),
				"month":        time.Month(month).String(),
				"year":         year,
				"email":        email,
			},
		})
		if err != nil {
			l.Error(err, "failed to log to discord")
		}

		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "Payroll emails resent successfully"))
		return
	}

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

		// Extract expense submissions from payroll
		expenseSubmissions := h.extractExpenseSubmissionsFromPayroll(payrolls)
		h.logger.Debug(fmt.Sprintf("Extracted %d expense submissions from payroll", len(expenseSubmissions)))

		// Extract accounting todos from payroll
		accountingTodos := h.extractAccountingTodosFromPayroll(payrolls)
		h.logger.Debug(fmt.Sprintf("Extracted %d accounting todos from payroll", len(accountingTodos)))

		// Store expense submissions (Expense records + AccountingTransactions)
		if len(expenseSubmissions) > 0 {
			err = h.storeExpenseSubmissions(expenseSubmissions, batchDate)
			if err != nil {
				h.logger.Error(err, "failed to store expense submissions")
				return fmt.Errorf("failed to store expense submissions: %w", err)
			}
			h.logger.Debug(fmt.Sprintf("Stored %d expense submissions", len(expenseSubmissions)))

			// Mark expense submissions as completed in NocoDB
			err = h.markExpenseSubmissionsAsCompleted(expenseSubmissions)
			if err != nil {
				h.logger.Error(err, "failed to mark expense submissions as completed in NocoDB")
				// Don't return error - log only, continue with processing
			}
		}

		// Store accounting todo transactions
		if len(accountingTodos) > 0 {
			err = h.storeAccountingTodoTransactions(accountingTodos, batchDate)
			if err != nil {
				h.logger.Error(err, "failed to store accounting todo transactions")
				return fmt.Errorf("failed to store accounting todo transactions: %w", err)
			}
			h.logger.Debug(fmt.Sprintf("Stored %d accounting todo transactions", len(accountingTodos)))

			// Mark todos as completed in NocoDB
			err = h.markAccountingTodosAsCompleted(accountingTodos)
			if err != nil {
				h.logger.Error(err, "failed to mark accounting todos as completed in NocoDB")
				// Don't return error - log only, continue with email sending
			}
		}

		// Simplify commission notes for email payslip
		// NOTE: Convert detailed notes (e.g., "2025104-KAFI-009 - Hiring - Nguyễn Hoàng Anh")
		// to simplified format (e.g., "2025104-KAFI-009 - Bonus") for email template
		for i := range payrolls {
			for j := range payrolls[i].CommissionExplains {
				// Extract invoice number (text before first " - ")
				name := payrolls[i].CommissionExplains[j].Name
				if strings.Contains(name, " - ") {
					parts := strings.SplitN(name, " - ", 2)
					payrolls[i].CommissionExplains[j].Name = parts[0] + " - Bonus"
				}
			}
		}

		// Using WaitGroup go routines to SendPayrollPaidEmail
		var wg sync.WaitGroup
		wg.Add(len(payrolls))
		c := make(chan *model.Payroll, len(payrolls)+1)
		for _, pr := range payrolls {
			go func(p model.Payroll) {
				defer wg.Done()
				if h.config.Env == "prod" || p.Employee.TeamEmail == "quang@d.foundation" {
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

	// Skip Basecamp bonus handling when Basecamp service is not available
	if h.service.Basecamp == nil {
		h.logger.Debug("skipping Basecamp bonus handling - Basecamp service unavailable")
		return nil
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

// extractExpenseSubmissionsFromPayroll extracts expense submissions from payroll bonus explains
func (h *handler) extractExpenseSubmissionsFromPayroll(payrolls []model.Payroll) []ExpenseSubmissionData {
	h.logger.Debug("Extracting expense submissions from payroll")

	var expenseSubmissions []ExpenseSubmissionData

	for _, p := range payrolls {
		var projectBonusExplains []model.ProjectBonusExplain
		err := json.Unmarshal(p.ProjectBonusExplain, &projectBonusExplains)
		if err != nil {
			h.logger.Error(err, fmt.Sprintf("failed to unmarshal ProjectBonusExplain for employee %s", p.EmployeeID))
			continue
		}

		h.logger.Debug(fmt.Sprintf("Employee %s has %d project bonus explains", p.Employee.FullName, len(projectBonusExplains)))
		for i, bonus := range projectBonusExplains {
			h.logger.Debug(fmt.Sprintf("  [%d] Name: %s, BasecampTodoID: %d, BasecampBucketID: %d", i, bonus.Name, bonus.BasecampTodoID, bonus.BasecampBucketID))
		}

		for _, bonus := range projectBonusExplains {
			// Identify expense submissions by checking if it's from expense_submissions table
			// For NocoDB: items will have title format "Description | Amount | Currency" or just description
			// BasecampTodoID stores the NocoDB record ID
			// BasecampBucketID may also be set (stores some NocoDB identifier)

			// Check if this is an accounting todo first (has specific keywords)
			isAccountingTodo := strings.Contains(bonus.Name, "Tiền điện") ||
				strings.Contains(bonus.Name, "CBRE") ||
				strings.Contains(bonus.Name, "Office Rental") ||
				strings.Contains(bonus.Name, "Rental")

			// Skip accounting todos in this extraction
			if isAccountingTodo {
				continue
			}

			// Check if this looks like an expense submission
			// Expense submissions from NocoDB may have simple names (not pipe-separated)
			// We'll rely on Amount field from bonus instead of parsing title
			if bonus.BasecampTodoID != 0 {
				description := bonus.Name
				amount := float64(bonus.Amount)
				currencyStr := "VND" // Default currency

				h.logger.Debug(fmt.Sprintf("Processing expense: %s, amount: %.0f, todoID: %d", description, amount, bonus.BasecampTodoID))

				// Append expense submission
				expenseSubmissions = append(expenseSubmissions, ExpenseSubmissionData{
					RecordID:     strconv.Itoa(bonus.BasecampTodoID),
					EmployeeID:   p.EmployeeID,
					EmployeeName: p.Employee.FullName,
					Amount:       amount,
					Currency:     currencyStr,
					Description:  description,
				})
			}
		}
	}

	h.logger.Debug(fmt.Sprintf("Extracted %d expense submissions from payroll", len(expenseSubmissions)))
	return expenseSubmissions
}

// extractAccountingTodosFromPayroll extracts accounting todos from payroll bonus explains
func (h *handler) extractAccountingTodosFromPayroll(payrolls []model.Payroll) []AccountingTodoData {
	h.logger.Debug("Extracting accounting todos from payroll")

	var accountingTodos []AccountingTodoData

	for _, p := range payrolls {
		var projectBonusExplains []model.ProjectBonusExplain
		err := json.Unmarshal(p.ProjectBonusExplain, &projectBonusExplains)
		if err != nil {
			h.logger.Error(err, fmt.Sprintf("failed to unmarshal ProjectBonusExplain for employee %s", p.EmployeeID))
			continue
		}

		h.logger.Debug(fmt.Sprintf("Employee %s has %d project bonus explains (checking for accounting todos)", p.Employee.FullName, len(projectBonusExplains)))

		for _, bonus := range projectBonusExplains {
			// Identify accounting todos by checking for specific keywords
			isAccountingTodo := strings.Contains(bonus.Name, "Tiền điện") ||
				strings.Contains(bonus.Name, "CBRE") ||
				strings.Contains(bonus.Name, "Office Rental") ||
				strings.Contains(bonus.Name, "Rental")

			if bonus.BasecampTodoID != 0 && isAccountingTodo {
				description := bonus.Name
				amount := float64(bonus.Amount)
				currencyStr := "VND" // Default currency

				h.logger.Debug(fmt.Sprintf("Processing accounting todo: %s, amount: %.0f, todoID: %d", description, amount, bonus.BasecampTodoID))

				accountingTodos = append(accountingTodos, AccountingTodoData{
					TodoID:       bonus.BasecampTodoID,
					EmployeeID:   p.EmployeeID,
					EmployeeName: p.Employee.FullName,
					Amount:       amount,
					Currency:     currencyStr,
					Description:  description,
				})
			}
		}
	}

	h.logger.Debug(fmt.Sprintf("Extracted %d accounting todos from payroll", len(accountingTodos)))
	return accountingTodos
}

// storeExpenseSubmissions creates Expense records and AccountingTransactions for expense submissions
func (h *handler) storeExpenseSubmissions(expenses []ExpenseSubmissionData, batchDate time.Time) error {
	h.logger.Debug(fmt.Sprintf("Storing %d expense submissions", len(expenses)))

	for _, expense := range expenses {
		// Check idempotency - skip if expense already exists
		existingExpense, err := h.store.Expense.GetByQuery(h.repo.DB(), &expensestore.ExpenseQuery{
			TaskProvider: "nocodb",
			TaskRef:      expense.RecordID,
		})
		if err == nil && existingExpense != nil {
			h.logger.Debug(fmt.Sprintf("Expense submission %s already exists (ID: %s), skipping", expense.RecordID, existingExpense.ID))
			continue
		}

		// Get currency by name
		cur, err := h.store.Currency.GetByName(h.repo.DB(), expense.Currency)
		if err != nil {
			h.logger.Error(err, fmt.Sprintf("failed to get currency %s for expense submission %s", expense.Currency, expense.RecordID))
			return fmt.Errorf("failed to get currency %s: %w", expense.Currency, err)
		}

		// For now, all expenses are in VND (already converted during payroll calculation)
		// ConversionAmount and ConversionRate are set to amount and 1.0 respectively
		conversionAmount := expense.Amount
		conversionRate := 1.0

		// Create metadata JSON
		metadata, err := json.Marshal(map[string]interface{}{
			"source":      "expense_submission",
			"expense_ref": expense.RecordID,
		})
		if err != nil {
			h.logger.Error(err, fmt.Sprintf("failed to marshal metadata for expense submission %s", expense.RecordID))
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		now := time.Now()
		organization := "Dwarves Foundation"

		// Create AccountingTransaction first
		transaction := &model.AccountingTransaction{
			Amount:           expense.Amount,
			ConversionAmount: model.VietnamDong(conversionAmount),
			Name:             fmt.Sprintf("%s - %s %s", expense.Description, expense.EmployeeName, batchDate.Format("02 January 2006")),
			Category:         model.AccountingOfficeSupply,
			Currency:         cur.Name,
			Date:             &now,
			CurrencyID:       &cur.ID,
			Organization:     organization,
			Metadata:         metadata,
			ConversionRate:   conversionRate,
			Type:             model.AccountingSE,
		}

		err = h.store.Accounting.CreateTransaction(h.repo.DB(), transaction)
		if err != nil {
			h.logger.Error(err, fmt.Sprintf("failed to create accounting transaction for expense submission %s", expense.RecordID))
			return fmt.Errorf("failed to create accounting transaction: %w", err)
		}

		h.logger.Debug(fmt.Sprintf("Created AccountingTransaction %s for expense submission %s", transaction.ID, expense.RecordID))

		// Create Expense record linked to transaction
		emptyJSON := []byte("[]")
		expenseRecord := &model.Expense{
			EmployeeID:              expense.EmployeeID,
			CurrencyID:              cur.ID,
			Amount:                  int(expense.Amount),
			Reason:                  expense.Description,
			IssuedDate:              batchDate,
			TaskProvider:            "nocodb",
			TaskRef:                 expense.RecordID,
			TaskBoard:               "expense_submissions",
			TaskAttachments:         emptyJSON,
			AccountingTransactionID: &transaction.ID,
		}

		_, err = h.store.Expense.Create(h.repo.DB(), expenseRecord)
		if err != nil {
			h.logger.Error(err, fmt.Sprintf("failed to create expense record for expense submission %s", expense.RecordID))
			return fmt.Errorf("failed to create expense record: %w", err)
		}

		h.logger.Debug(fmt.Sprintf("Created Expense record %s linked to transaction %s", expenseRecord.ID, transaction.ID))
	}

	h.logger.Debug(fmt.Sprintf("Successfully stored %d expense submissions", len(expenses)))
	return nil
}

// storeAccountingTodoTransactions creates individual transactions for accounting todos
func (h *handler) storeAccountingTodoTransactions(todos []AccountingTodoData, batchDate time.Time) error {
	h.logger.Debug(fmt.Sprintf("Storing %d accounting todo transactions", len(todos)))

	for _, todo := range todos {
		// Check idempotency - skip if transaction already exists
		metadataCheck, _ := json.Marshal(map[string]interface{}{
			"source":  "accounting_todo",
			"todo_id": todo.TodoID,
		})

		// Query for existing transaction by metadata
		// Note: This is a simplified check - in production you'd want a proper query method
		var existingTx model.AccountingTransaction
		err := h.repo.DB().Where("metadata = ?", metadataCheck).First(&existingTx).Error
		if err == nil {
			h.logger.Debug(fmt.Sprintf("Accounting todo transaction %d already exists (ID: %s), skipping", todo.TodoID, existingTx.ID))
			continue
		}

		// Get currency by name
		cur, err := h.store.Currency.GetByName(h.repo.DB(), todo.Currency)
		if err != nil {
			h.logger.Error(err, fmt.Sprintf("failed to get currency %s for accounting todo %d", todo.Currency, todo.TodoID))
			return fmt.Errorf("failed to get currency %s: %w", todo.Currency, err)
		}

		// For now, all accounting todos are in VND (already converted during payroll calculation)
		// ConversionAmount and ConversionRate are set to amount and 1.0 respectively
		conversionAmount := todo.Amount
		conversionRate := 1.0

		// Create metadata JSON
		metadata, err := json.Marshal(map[string]interface{}{
			"source":  "accounting_todo",
			"todo_id": todo.TodoID,
		})
		if err != nil {
			h.logger.Error(err, fmt.Sprintf("failed to marshal metadata for accounting todo %d", todo.TodoID))
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		now := time.Now()
		organization := "Dwarves Foundation"

		// Create AccountingTransaction
		transaction := &model.AccountingTransaction{
			Amount:           todo.Amount,
			ConversionAmount: model.VietnamDong(conversionAmount),
			Name:             fmt.Sprintf("%s - %s", todo.Description, todo.EmployeeName),
			Category:         model.AccountingOfficeServices,
			Currency:         cur.Name,
			Date:             &now,
			CurrencyID:       &cur.ID,
			Organization:     organization,
			Metadata:         metadata,
			ConversionRate:   conversionRate,
			Type:             model.AccountingSE,
		}

		err = h.store.Accounting.CreateTransaction(h.repo.DB(), transaction)
		if err != nil {
			h.logger.Error(err, fmt.Sprintf("failed to create accounting transaction for accounting todo %d", todo.TodoID))
			return fmt.Errorf("failed to create accounting transaction: %w", err)
		}

		h.logger.Debug(fmt.Sprintf("Created AccountingTransaction %s for accounting todo %d", transaction.ID, todo.TodoID))
	}

	h.logger.Debug(fmt.Sprintf("Successfully stored %d accounting todo transactions", len(todos)))
	return nil
}

// markAccountingTodosAsCompleted marks todos as completed in NocoDB
func (h *handler) markAccountingTodosAsCompleted(todos []AccountingTodoData) error {
	h.logger.Debug(fmt.Sprintf("Marking %d accounting todos as completed in NocoDB", len(todos)))

	// Check if PayrollAccountingTodoProvider is NocoDB service
	if h.service.PayrollAccountingTodoProvider == nil {
		h.logger.Debug("PayrollAccountingTodoProvider is nil, skipping NocoDB update")
		return nil
	}

	// Try to cast to NocoDB AccountingTodoService
	nocoService, ok := h.service.PayrollAccountingTodoProvider.(*nocodb.AccountingTodoService)
	if !ok {
		h.logger.Debug("PayrollAccountingTodoProvider is not NocoDB service (Basecamp flow), skipping NocoDB update")
		return nil
	}

	successCount := 0
	failedCount := 0
	var errors []string

	for _, todo := range todos {
		err := nocoService.MarkTodoAsCompleted(todo.TodoID)
		if err != nil {
			h.logger.Error(err, fmt.Sprintf("failed to mark accounting todo %d as completed in NocoDB", todo.TodoID))
			errors = append(errors, fmt.Sprintf("todo %d: %v", todo.TodoID, err))
			failedCount++
		} else {
			h.logger.Debug(fmt.Sprintf("Marked accounting todo %d as completed in NocoDB", todo.TodoID))
			successCount++
		}
	}

	h.logger.Debug(fmt.Sprintf("Completed marking accounting todos (success: %d, failed: %d)", successCount, failedCount))

	// Return aggregated error if any failed (non-fatal, logged only)
	if len(errors) > 0 {
		return fmt.Errorf("failed to mark %d todos as completed: %v", failedCount, strings.Join(errors, "; "))
	}

	return nil
}

// markExpenseSubmissionsAsCompleted marks expense submissions as completed in NocoDB
func (h *handler) markExpenseSubmissionsAsCompleted(expenses []ExpenseSubmissionData) error {
	h.logger.Debug(fmt.Sprintf("Marking %d expense submissions as completed in NocoDB", len(expenses)))

	if h.service.PayrollExpenseProvider == nil {
		h.logger.Debug("PayrollExpenseProvider is nil, skipping NocoDB update")
		return nil
	}

	// Try to cast to NocoDB ExpenseService
	nocoService, ok := h.service.PayrollExpenseProvider.(*nocodb.ExpenseService)
	if !ok {
		h.logger.Debug("PayrollExpenseProvider is not NocoDB service (Basecamp flow), skipping NocoDB update")
		return nil
	}

	successCount := 0
	failedCount := 0
	var errors []string

	for _, expense := range expenses {
		// Convert RecordID string to int
		expenseID, err := strconv.Atoi(expense.RecordID)
		if err != nil {
			h.logger.Error(err, fmt.Sprintf("failed to convert expense RecordID to int: %s", expense.RecordID))
			errors = append(errors, fmt.Sprintf("expense %s: invalid ID", expense.RecordID))
			failedCount++
			continue
		}

		err = nocoService.MarkExpenseAsCompleted(expenseID)
		if err != nil {
			h.logger.Error(err, fmt.Sprintf("failed to mark expense submission %d as completed in NocoDB", expenseID))
			errors = append(errors, fmt.Sprintf("expense %d: %v", expenseID, err))
			failedCount++
		} else {
			h.logger.Debug(fmt.Sprintf("Marked expense submission %d as completed in NocoDB", expenseID))
			successCount++
		}
	}

	h.logger.Debug(fmt.Sprintf("Completed marking expense submissions (success: %d, failed: %d)", successCount, failedCount))

	// Return aggregated error if any failed (non-fatal, logged only)
	if len(errors) > 0 {
		return fmt.Errorf("failed to mark %d expense submissions as completed: %v", failedCount, strings.Join(errors, "; "))
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
