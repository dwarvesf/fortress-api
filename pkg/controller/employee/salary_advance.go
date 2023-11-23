package employee

import (
	"errors"
	"math"
	"strconv"

	"github.com/Rhymond/go-money"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils"
)

type SalaryAdvanceResponse struct {
	EmployeeID      string `json:"employee_id"`
	AmountIcy       string `json:"amount_icy"`
	AmountUSD       string `json:"amount_usd"`
	TransactionID   string `json:"transaction_id"`
	TransactionHash string `json:"transaction_hash"`
}

func (r *controller) SalaryAdvance(discordID string, amount int64) (*SalaryAdvanceResponse, error) {
	icyUsdRateConfig, err := r.store.Config.OneByKey(r.repo.DB(), model.ConfigKeyIcyUSDRate)
	if err != nil {
		return nil, err
	}
	icyUsdRate, err := strconv.ParseFloat(icyUsdRateConfig.Value, 64)
	if err != nil {
		return nil, err
	}

	// Get employee by discord id
	employee, err := r.store.Employee.GetByDiscordID(r.repo.DB(), discordID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}

	err = r.checkFullTimeRole(employee)
	if err != nil {
		return nil, err
	}

	salaryAdvances, err := r.store.SalaryAdvance.ListNotPayBackByEmployeeID(r.repo.DB(), employee.ID.String())
	if err != nil {
		return nil, err
	}

	// calculate number of not paid back advance salary
	var notPaidBackAmount int64
	for _, salaryAdvance := range salaryAdvances {
		if !salaryAdvance.IsPaidBack {
			notPaidBackAmount = notPaidBackAmount + salaryAdvance.AmountIcy
		}
	}

	// Get employee's salary
	salary, err := r.store.BaseSalary.OneByEmployeeID(r.repo.DB(), employee.ID.String())
	if err != nil {
		return nil, err
	}

	// Calculate advance amount
	maxAdvanceAmountIcy, err := r.calculateMaxAdvanceAmountIcy(salary, icyUsdRate)
	if err != nil {
		return nil, err
	}

	// Check if amount exceed amount can be advanced
	if amount > maxAdvanceAmountIcy-notPaidBackAmount {
		return nil, errors.Join(ErrSalaryAdvanceExceedAmount, errors.New("amount can be advanced is "+utils.FormatNumber(maxAdvanceAmountIcy-notPaidBackAmount)))
	}

	// Create advance salary record
	amountUSD := float64(amount) * icyUsdRate
	baseAmount, rate, err := r.service.Wise.Convert(amountUSD, "USD", salary.Currency.Name)
	if err != nil {
		return nil, err
	}

	tx, done := r.repo.NewTransaction()
	salaryAdvance := &model.SalaryAdvance{
		EmployeeID:     employee.ID,
		AmountIcy:      amount,
		AmountUSD:      amountUSD,
		BaseAmount:     baseAmount,
		ConversionRate: rate,
		CurrencyID:     salary.CurrencyID,
	}
	if err := r.store.SalaryAdvance.Save(tx.DB(), salaryAdvance); err != nil {
		return nil, done(err)
	}

	// Make advance salary request
	txs, err := r.service.Mochi.SendFromAccountToUser(int(amount), discordID)
	if err != nil {
		return nil, done(err)
	}

	if len(txs) == 0 {
		return nil, done(ErrNoTransactionFound)
	}

	response := &SalaryAdvanceResponse{
		EmployeeID:      employee.ID.String(),
		AmountIcy:       utils.FormatNumber(amount),
		AmountUSD:       utils.FormatMoney(amountUSD, money.USD),
		TransactionID:   strconv.Itoa(int(txs[0].TransactionID)),
		TransactionHash: txs[0].RecipientID,
	}

	return response, done(nil)
}

func (r *controller) CheckSalaryAdvance(discordID string) (amountIcy string, amountUSD string, error error) {
	icyUsdRateConfig, err := r.store.Config.OneByKey(r.repo.DB(), model.ConfigKeyIcyUSDRate)
	if err != nil {
		return "", "", err
	}
	icyUsdRate, err := strconv.ParseFloat(icyUsdRateConfig.Value, 64)
	if err != nil {
		return "", "", err
	}

	// Get employee by discord id
	employee, err := r.store.Employee.GetByDiscordID(r.repo.DB(), discordID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", ErrEmployeeNotFound
		}
		return "", "", err
	}

	err = r.checkFullTimeRole(employee)
	if err != nil {
		return "", "", err
	}

	salaryAdvances, err := r.store.SalaryAdvance.ListNotPayBackByEmployeeID(r.repo.DB(), employee.ID.String())
	if err != nil {
		return "", "", err
	}

	// calculate number of not paid back advance salary
	var notPaidBackAmount int64
	for _, salaryAdvance := range salaryAdvances {
		if !salaryAdvance.IsPaidBack {
			notPaidBackAmount = notPaidBackAmount + salaryAdvance.AmountIcy
		}
	}

	// Get employee's salary
	salary, err := r.store.BaseSalary.OneByEmployeeID(r.repo.DB(), employee.ID.String())
	if err != nil {
		return "", "", err
	}

	// Calculate advance amount
	advanceAmountIcy, err := r.calculateMaxAdvanceAmountIcy(salary, icyUsdRate)
	if err != nil {
		return "", "", err
	}

	advanceAmountUSD := float64(advanceAmountIcy-notPaidBackAmount) * icyUsdRate

	return utils.FormatNumber(advanceAmountIcy - notPaidBackAmount),
		utils.FormatMoney(advanceAmountUSD, money.USD),
		nil
}

func (r *controller) checkFullTimeRole(employee *model.Employee) error {
	fullTimeRole, err := r.store.Role.GetByCode(r.repo.DB(), model.RoleFullTimeCode)
	if err != nil {
		return err
	}

	var highestEmployeeLevel int64 = 10000
	for _, role := range employee.Roles {
		if role.Level < highestEmployeeLevel {
			highestEmployeeLevel = role.Level
		}
	}

	// check if employee is full time or higher
	if highestEmployeeLevel > fullTimeRole.Level {
		return ErrEmployeeNotFullTime
	}
	return nil
}

func (r *controller) calculateMaxAdvanceAmountIcy(salary *model.BaseSalary, icyUsdRate float64) (int64, error) {
	// Get advance salary max cap
	salaryAdvanceMaxCap, err := r.store.Config.OneByKey(r.repo.DB(), model.ConfigKeySalaryAdvanceMaxCap)
	if err != nil {
		return 0, err
	}

	// Check if advance salary max cap is number and in range of 0 - 100
	maxCap, err := strconv.Atoi(salaryAdvanceMaxCap.Value)
	if err != nil || maxCap < 0 || maxCap > 100 {
		return 0, ErrSalaryAdvanceMaxCapInvalid
	}

	var advanceAmountUSD float64
	if salary.Currency.Name == "USD" {
		advanceAmountUSD = float64(salary.ContractAmount) * (float64(maxCap) / 100)
	} else {
		convertedValue, _, err := r.service.Wise.Convert(float64(salary.ContractAmount)*(float64(maxCap)/100), salary.Currency.Name, "USD")
		if err != nil {
			return 0, err
		}
		advanceAmountUSD = convertedValue
	}

	advanceAmountIcy := advanceAmountUSD / icyUsdRate

	return int64(math.Round(advanceAmountIcy/10) * 10), nil
}
