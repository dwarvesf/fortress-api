package companyinfo

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string) (companyInfo *model.CompanyInfo, err error)
	IsExist(db *gorm.DB, id string) (exists bool, err error)
	All(db *gorm.DB) (companyInfos []*model.CompanyInfo, err error)
}
