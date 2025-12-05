package onleaverequest

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

type GetOnLeaveInput struct {
	Date string
}

// Create creates an on-leave request record in the database
func (s *store) Create(db *gorm.DB, r *model.OnLeaveRequest) (request *model.OnLeaveRequest, err error) {
	return r, db.Create(r).Error
}

func (s *store) All(db *gorm.DB, input GetOnLeaveInput) ([]*model.OnLeaveRequest, error) {
	var chapters []*model.OnLeaveRequest
	query := db.
		Preload("Creator").
		Preload("Creator.DiscordAccount").
		Preload("Approver").
		Preload("Approver.DiscordAccount")

	if input.Date != "" {
		query = query.Where("start_date <= ? AND ? <= end_date", input.Date, input.Date)
	}

	return chapters, query.Find(&chapters).Error
}

// GetByNocodbID retrieves an on-leave request by NocoDB ID (including soft-deleted)
func (s *store) GetByNocodbID(db *gorm.DB, nocodbID int) (*model.OnLeaveRequest, error) {
	var request model.OnLeaveRequest
	err := db.Unscoped().Where("nocodb_id = ?", nocodbID).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// GetByNotionPageID retrieves an on-leave request by Notion page ID (including soft-deleted)
func (s *store) GetByNotionPageID(db *gorm.DB, notionPageID string) (*model.OnLeaveRequest, error) {
	var request model.OnLeaveRequest
	err := db.Unscoped().Where("notion_page_id = ?", notionPageID).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// Delete permanently deletes an on-leave request by ID (hard delete)
func (s *store) Delete(db *gorm.DB, id string) error {
	return db.Unscoped().Delete(&model.OnLeaveRequest{}, "id = ?", id).Error
}
