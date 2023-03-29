package accounting

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type accountingService struct {
}

// New create new pg service
func New() IStore {
	return &accountingService{}
}

func (s *accountingService) GetAccountingTransactions(db *gorm.DB) ([]model.AccountingTransaction, error) {
	transactions := []model.AccountingTransaction{}
	return transactions, db.
		Joins(`left join accounting_categories on accounting_transactions.category = accounting_categories.name`).
		Preload("CurrencyInfo").
		Preload("AccountingCategory").
		Order(`date DESC`).
		Order(`
		CASE accounting_categories.type
		WHEN 'IN' THEN 1
		WHEN 'SE' THEN 2
		WHEN 'OP' THEN 3
		WHEN 'OV' THEN 4
		WHEN 'CA' THEN 5
		ELSE 6
		END`).
		Find(&transactions).
		Error
}

func (s *accountingService) CreateTransaction(db *gorm.DB, transaction *model.AccountingTransaction) error {
	return db.Create(transaction).Error
}

func (s *accountingService) DeleteTransaction(db *gorm.DB, t *model.AccountingTransaction) error {
	return db.Delete(&t).Error
}

func (s *accountingService) GetAccountingCategories(db *gorm.DB) ([]model.AccountingCategory, error) {
	categories := []model.AccountingCategory{}
	return categories, db.Find(&categories).Error
}

func (s *accountingService) CreateMultipleTransaction(db *gorm.DB, transactions []*model.AccountingTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	valueStrings := []string{}
	valueArgs := []interface{}{}

	for _, v := range transactions {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

		valueArgs = append(valueArgs, v.Name)
		valueArgs = append(valueArgs, v.Amount)
		valueArgs = append(valueArgs, v.Category)
		valueArgs = append(valueArgs, v.Currency)
		valueArgs = append(valueArgs, v.CurrencyID)
		valueArgs = append(valueArgs, v.Date)
		valueArgs = append(valueArgs, v.ConversionAmount)
		valueArgs = append(valueArgs, v.Organization)
		valueArgs = append(valueArgs, v.ConversionRate)
		valueArgs = append(valueArgs, v.Type)
		valueArgs = append(valueArgs, v.Metadata)
	}

	smt := `
		INSERT INTO accounting_transactions(name,amount,category,currency,currency_id,date,conversion_amount,organization,conversion_rate, type, metadata)
		VALUES %s
		ON CONFLICT ON CONSTRAINT transaction_info_unique
		DO UPDATE SET
		amount = EXCLUDED.amount,
		category = EXCLUDED.category,
		currency_id = EXCLUDED.currency_id,
		currency = EXCLUDED.currency,
		conversion_amount = EXCLUDED.conversion_amount,
		type = EXCLUDED.type,
		metadata = EXCLUDED.metadata`

	smt = fmt.Sprintf(smt, strings.Join(valueStrings, ","))

	tx := db.Begin()
	if err := tx.Exec(smt, valueArgs...).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
