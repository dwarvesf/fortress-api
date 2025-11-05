package basecamp

import (
	"encoding/json"
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/thoas/go-funk"
	"gorm.io/datatypes"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/currency"
	"github.com/dwarvesf/fortress-api/pkg/store/expense"
)

const (
	defaultCurrencyType = "VND"
	thousandUnit        = 1000
	millionUnit         = 1000000
	amountPat           = "(\\d+(k|tr|m)\\d+|\\d+(k|tr|m)|\\d+)"
)

// BasecampExpenseData --
type BasecampExpenseData struct {
	Reason          string
	Amount          int
	CurrencyType    string
	CreatorEmail    string
	InvoiceImageURL string
	MetaData        datatypes.JSON
	BasecampID      int
	CreatorID       int
}

// ExtractBasecampExpenseAmount --
func (s *Service) ExtractBasecampExpenseAmount(source string) int {
	return getAmount(strings.Replace(source, ".", "", -1))
}

// CreateBasecampExpense --
func (s *Service) CreateBasecampExpense(
	data BasecampExpenseData,
) error {
	l := s.logger.Fields(logger.Fields{
		"service":     "basecamp",
		"method":      "CreateBasecampExpense",
		"basecampID":  data.BasecampID,
		"creatorID":   data.CreatorID,
		"amount":      data.Amount,
		"currency":    data.CurrencyType,
	})

	employee, err := s.store.Employee.OneByBasecampID(s.repo.DB(), data.CreatorID)
	if err != nil {
		l.AddField("creatorID", data.CreatorID).Error(err, "failed to get employee by basecampID")
		return errors.New("failed to get employee by basecampID: " + strconv.Itoa(data.BasecampID))
	}

	c, err := s.store.Currency.GetByName(s.repo.DB(), data.CurrencyType)
	if err != nil {
		l.AddField("currencyType", data.CurrencyType).Error(err, "failed to get currency by name")
		return errors.New("failed to get currency by name: " + data.CurrencyType)
	}

	date := time.Now()

	e, err := s.store.Expense.Create(s.repo.DB(), &model.Expense{
		Amount:          data.Amount,
		Reason:          data.Reason,
		EmployeeID:      employee.ID,
		CurrencyID:      c.ID,
		InvoiceImageURL: data.InvoiceImageURL,
		Metadata:        data.MetaData,
		BasecampID:      data.BasecampID,
	})
	if err != nil {
		l.Fields(logger.Fields{
			"employeeID": employee.ID,
			"currencyID": c.ID,
			"reason":     data.Reason,
		}).Error(err, "failed to create expense in database")
		return err
	}

	m := model.AccountingMetadata{
		Source: "expense",
		ID:     e.ID.String(),
	}
	bonusBytes, err := json.Marshal(&m)
	if err != nil {
		l.AddField("expenseID", e.ID).Error(err, "failed to marshal accounting metadata")
		return err
	}

	temp, rate, err := s.Wise.Convert(float64(data.Amount), c.Name, currency.VNDCurrency)
	if err != nil {
		l.Fields(logger.Fields{
			"amount":       data.Amount,
			"fromCurrency": c.Name,
			"toCurrency":   currency.VNDCurrency,
			"expenseID":    e.ID,
		}).Error(err, "CRITICAL: Wise API currency conversion failed - expense created but conversion rate is 0")
		return err
	}
	am := model.NewVietnamDong(int64(temp))

	transaction := &model.AccountingTransaction{
		Name:             "Expense - " + data.Reason,
		Amount:           float64(data.Amount),
		ConversionAmount: am.Format(),
		Date:             &date,
		Category:         model.AccountingOfficeSupply,
		CurrencyID:       &c.ID,
		Currency:         c.Name,
		ConversionRate:   rate,
		Metadata:         bonusBytes,
		Type:             model.AccountingOV,
	}

	if err = s.store.Accounting.CreateTransaction(
		s.repo.DB(),
		transaction,
	); err != nil {
		l.Fields(logger.Fields{
			"expenseID":        e.ID,
			"transactionName":  transaction.Name,
			"amount":           transaction.Amount,
			"conversionAmount": transaction.ConversionAmount,
		}).Error(err, "failed to create accounting transaction")
		return err
	}

	e.AccountingTransactionID = &transaction.ID

	if _, err = s.store.Expense.Update(s.repo.DB(), e); err != nil {
		l.Fields(logger.Fields{
			"expenseID":              e.ID,
			"accountingTransactionID": transaction.ID,
		}).Error(err, "failed to update expense with accounting transaction ID")
		return err
	}

	return nil
}

func (s *Service) UncheckBasecampExpenseHandler(
	data BasecampExpenseData,
) error {
	e, err := s.store.Expense.GetByQuery(s.repo.DB(), &expense.ExpenseQuery{BasecampID: data.BasecampID})
	if err != nil {
		return err
	}

	if _, err = s.store.Expense.Delete(s.repo.DB(), e); err != nil {
		return err
	}

	return nil
}

func getAmountStr(s string) string {
	c, _ := regexp.Compile(amountPat)
	return c.FindString(s)
}

// func getReason(s string) string {
// 	amount := getAmountStr(s)
// 	s = strings.Replace(s, amount, "", 1)
// 	return strings.TrimSpace(strings.Replace(s, "for", "", 1))
// }

func getAmount(source string) int {
	s := getAmountStr(source)
	if len(s) == 0 {
		return 0
	}

	switch {
	case isThousand(s):
		return thousand(s)
	case isMillion(s):
		return million(s)
	default:
		a, _ := strconv.Atoi(s)
		return a
	}
}

func isThousand(s string) bool {
	return funk.Contains(s, "k")
}

func thousand(s string) int {
	a := strings.Index(s, "k")
	if len(s[a+1:]) > 3 {
		return 0
	}
	prefix, _ := strconv.Atoi(s[0:a])
	suffix, _ := strconv.Atoi(s[a+1:])
	return prefix*thousandUnit + int(float64(suffix)/math.Pow10(len(s[a+1:])-1)*100)
}

func isMillion(s string) bool {
	return funk.Contains(s, "tr") || funk.Contains(s, "m")
}

func million(s string) int {
	newStr := strings.Replace(s, "tr", "m", -1)
	i := strings.Index(newStr, "m")
	if len(newStr[i+1:]) > 6 {
		return 0
	}
	pref, _ := strconv.Atoi(newStr[0:i])
	suf, _ := strconv.Atoi(newStr[i+1:])
	return (pref * millionUnit) + int(float64(suf)/math.Pow10(len(newStr[i+1:])-1)*thousandUnit*100)
}
