package projectnotion

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// OneByProjectID get projectNotion by project id
func (s *store) OneByProjectID(db *gorm.DB, projectID string) (*model.ProjectNotion, error) {
	var projectNotion *model.ProjectNotion
	return projectNotion, db.Where("project_id = ?", projectID).Preload("Project", "deleted_at IS NULL").First(&projectNotion).Error
}

// OneByAuditNotionID get projectNotion by audit notion id
func (s *store) OneByAuditNotionID(db *gorm.DB, auditNotionID string) (*model.ProjectNotion, error) {
	var projectNotion *model.ProjectNotion
	return projectNotion, db.Where("audit_notion_id = ?", auditNotionID).Preload("Project", "deleted_at IS NULL").First(&projectNotion).Error
}

// Create creates a new projectNotion
func (s *store) Create(db *gorm.DB, e *model.ProjectNotion) (projectNotion *model.ProjectNotion, err error) {
	return e, db.Create(e).Error
}

// Update update all value (including nested model)
func (s *store) Update(db *gorm.DB, projectNotion *model.ProjectNotion) (*model.ProjectNotion, error) {
	return projectNotion, db.Model(&projectNotion).Where("id = ?", projectNotion.ID).Updates(&projectNotion).First(&projectNotion).Error
}

// IsExistByAuditNotionID check project notion existence by audit notion id
func (s *store) IsExistByAuditNotionID(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM project_notions WHERE audit_notion_id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.ProjectNotion, updatedFields ...string) (*model.ProjectNotion, error) {
	projectNotion := model.ProjectNotion{}
	return &projectNotion, db.Model(&projectNotion).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}
