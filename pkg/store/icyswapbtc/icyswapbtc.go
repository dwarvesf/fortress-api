package icyswapbtc

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create creates a new request
func (s *store) Create(db *gorm.DB, e *model.IcySwapBtcRequest) (request *model.IcySwapBtcRequest, err error) {
	return e, db.Create(e).Error
}

// Update update all value
func (s *store) Update(db *gorm.DB, request *model.IcySwapBtcRequest) (*model.IcySwapBtcRequest, error) {
	return request, db.Model(&request).Where("id = ?", request.ID).Updates(&request).First(&request).Error
}

func (s *store) One(db *gorm.DB, query *Query) (*model.IcySwapBtcRequest, error) {
	var request *model.IcySwapBtcRequest
	if query.ID != "" {
		db = db.Where("id = ?", query.ID)
	}
	if query.SwapRequestStatus != "" {
		db = db.Where("request_status = ?", query.SwapRequestStatus)
	}
	return request, db.First(&request).Error
}

func (s *store) IsExist(db *gorm.DB, requestCode string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM icy_swap_btc_requests WHERE request_code = ?) as result", requestCode)

	return result.Result, query.Scan(&result).Error
}
