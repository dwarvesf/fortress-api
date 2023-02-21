package invoicenumbercaching

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) get(db *gorm.DB, key string) (next int, err error) {
	var maxNumber model.InvoiceNumberCaching
	err = db.Where(&model.InvoiceNumberCaching{Key: key}).Find(&maxNumber).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return maxNumber.Number + 1, nil
}

func (s *store) GetNext(db *gorm.DB, key string) (int, error) {
	return s.get(db, key)
}

func (s *store) Set(db *gorm.DB, key string) error {
	var maxNumber model.InvoiceNumberCaching
	err := db.Where(&model.InvoiceNumberCaching{Key: key}).FirstOrCreate(&maxNumber).Error
	if err != nil {
		return err
	}

	maxNumber.Number++
	return db.Model(&maxNumber).Update("number", maxNumber.Number).Error
}

func (s *store) NextProjectTemplateNumber(db *gorm.DB, pid string) (next int, err error) {
	key := fmt.Sprintf("%s%s", model.InvoiceCachingKey.ProjectTemplateNumberPrefix, pid)
	return s.GetNext(db, key)
}

func (s *store) Decrease(db *gorm.DB, key string) error {
	var maxNumber model.InvoiceNumberCaching
	err := db.Where(&model.InvoiceNumberCaching{Key: key}).FirstOrCreate(&maxNumber).Error
	if err != nil {
		return err
	}

	maxNumber.Number--
	return db.Model(&maxNumber).Update("number", maxNumber.Number).Error
}

func (s *store) UpdateInvoiceCachingNumber(db *gorm.DB, issuedAt time.Time, alias string) error {
	year := issuedAt.Year()
	err := s.Set(db, fmt.Sprintf("%s_%d", model.InvoiceCachingKey.YearInvoiceNumberPrefix, year))
	if err != nil {
		return err
	}

	err = s.Set(db, fmt.Sprintf("%s_%s_%d", model.InvoiceCachingKey.ProjectInvoiceNumberPrefix, strings.ToUpper(alias), year))
	if err != nil {
		return err
	}

	return nil
}

func (s *store) UnCountErrorInvoice(db *gorm.DB, issuedAt time.Time) error {
	year := issuedAt.Year()
	return s.Decrease(db, fmt.Sprint(model.InvoiceCachingKey.YearInvoiceNumberPrefix, year))
}
