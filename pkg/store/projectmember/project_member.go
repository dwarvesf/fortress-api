package projectmember

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Delete delete ProjectMember by id
func (s *store) Delete(db *gorm.DB, id string) error {
	return db.Unscoped().Where("id = ?", id).Delete(&model.ProjectMember{}).Error
}

// IsExist check ProjectMember existence
func (s *store) IsExist(db *gorm.DB, id string) (bool, error) {
	var record struct {
		Result bool
	}

	query := db.Raw("SELECT EXISTS (SELECT * FROM project_members WHERE id = ?) as result", id)
	return record.Result, query.Scan(&record).Error
}

// One return a project member by projectID and employeeID
func (s *store) One(db *gorm.DB, projectID string, employeeID string, preload bool) (*model.ProjectMember, error) {
	query := db.Where("project_id = ? AND employee_id = ? AND status = ?",
		projectID,
		employeeID,
		model.ProjectMemberStatusActive)

	if preload {
		query = query.Preload("Employee")
	}

	var member *model.ProjectMember
	return member, query.First(&member).Error
}

// GetOneBySlotID return a project member by slotID
func (s *store) GetOneBySlotID(db *gorm.DB, slotID string) (*model.ProjectMember, error) {
	query := db.Where("project_slot_id = ? AND left_date IS NULL", slotID)

	member := &model.ProjectMember{}
	return member, query.Preload("Employee").First(&member).Error
}

// Create using for create new member
func (s *store) Create(db *gorm.DB, member *model.ProjectMember) error {
	return db.Create(&member).Preload("Employee").First(&member).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.ProjectMember, updatedFields ...string) (*model.ProjectMember, error) {
	member := model.ProjectMember{}
	return &member, db.Model(&member).Where("id = ?", id).Select(updatedFields).Updates(&updateModel).Error
}

// IsExistsByEmployeeID check ProjectMember existance by project id and employee id
func (s *store) IsExistsByEmployeeID(db *gorm.DB, projectID string, employeeID string) (bool, error) {
	var record struct {
		Result bool
	}

	query := db.Raw("SELECT EXISTS (SELECT * FROM project_members WHERE project_id  = ? and employee_id = ?) as result", projectID, employeeID)
	return record.Result, query.Scan(&record).Error
}

// GetByProjectIDs get project member by porjectID list
func (s *store) GetByProjectIDs(db *gorm.DB, projectIDs []string) ([]*model.ProjectMember, error) {
	var members []*model.ProjectMember
	return members, db.Where("left_date IS NULL AND status = 'active' AND project_id IN ?", projectIDs).Preload("Employee").Find(&members).Error
}
