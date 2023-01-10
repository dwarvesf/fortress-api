// please edit this file only with approval from hnh
package valuation

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) GetAccountReceivable(db *gorm.DB, year string) (*model.CurrencyView, error) {
	res := &model.CurrencyView{}
	return res, db.Raw("select * from vw_account_receivables where year = ?", year).Find(&res).Error
}

func (s *store) GetRevenue(db *gorm.DB, year string) (*model.CurrencyView, error) {
	res := &model.CurrencyView{}
	return res, db.Raw(`SELECT sum(vnd) AS vnd, sum(usd) AS usd, sum(gbp) AS gbp, sum(eur) AS eur, sum(sgd) as sgd
			FROM vw_incomes
			WHERE year = ?`, year).
		Find(&res).Error
}

func (s *store) GetInvestment(db *gorm.DB, year string) (*model.CurrencyView, error) {
	res := &model.CurrencyView{}
	return res, db.Raw(`SELECT * FROM vw_investments WHERE year = ?`, year).
		Find(&res).Error
}

func (s *store) GetLiabilities(db *gorm.DB, year string) (res []model.Liability, total *model.CurrencyView, err error) {
	err = db.Find(&res, "date_part('year', created_at) = ?", year).Error
	if err != nil {
		return nil, nil, err
	}

	return res, total, db.Raw(`SELECT sum(vnd) AS vnd, sum(usd) AS usd, sum(gbp) AS gbp, sum(eur) AS eur, sum(sgd) as sgd
	FROM vw_liabilities
	WHERE year = ?`, year).
		Find(&total).Error
}

func (s *store) GetAssetAmount(db *gorm.DB, year string) (float64, error) {
	return 0, nil
}

func (s *store) GetExpense(db *gorm.DB, year string) (*model.CurrencyView, error) {
	res := &model.CurrencyView{}
	return res, db.Raw("select * from vw_expenses where year = ?", year).Find(&res).Error
}

func (s *store) GetPayroll(db *gorm.DB, year string) (*model.CurrencyView, error) {
	res := &model.CurrencyView{}
	return res, db.Raw("select total as vnd from vw_payrolls where year = ?", year).Scan(&res).Error
}
