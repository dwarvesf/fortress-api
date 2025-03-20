package icyswapbtc

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, e *model.IcySwapBtcRequest) (request *model.IcySwapBtcRequest, err error)
	One(db *gorm.DB, query *Query) (request *model.IcySwapBtcRequest, err error)
	All(db *gorm.DB, query *Query) (request []model.IcySwapBtcRequest, err error)
	IsExist(db *gorm.DB, requestCode string) (exists bool, err error)
	Update(db *gorm.DB, request *model.IcySwapBtcRequest) (a *model.IcySwapBtcRequest, err error)
}

type Query struct {
	ID                string
	RequestCode       string
	SwapRequestStatus string
	RevertStatus      string
	OrderDirection    string // "ASC" or "DESC"
}
