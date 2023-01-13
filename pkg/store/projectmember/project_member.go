package projectmember

import (
	"time"

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
	query := db.Where("project_id = ? AND employee_id = ?", projectID, employeeID)

	if preload {
		query = query.Preload("Employee")
	}

	var member *model.ProjectMember
	return member, query.First(&member).Error
}

// OneBySlotID return a project member by slotID
func (s *store) OneBySlotID(db *gorm.DB, slotID string) (*model.ProjectMember, error) {
	var member *model.ProjectMember
	return member, db.Where("project_slot_id = ? AND (left_date IS NULL OR left_date > ?)", slotID, time.Now()).Preload("Employee").First(&member).Error
}

// Create using for create new member
func (s *store) Create(db *gorm.DB, member *model.ProjectMember) error {
	return db.Create(&member).Preload("Employee").First(&member).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.ProjectMember, updatedFields ...string) (*model.ProjectMember, error) {
	member := model.ProjectMember{}
	return &member, db.Model(&member).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}

// UpdateSelectedFieldByProjectID just update selected field by projectID
func (s *store) UpdateSelectedFieldByProjectID(db *gorm.DB, projectID string, updateModel model.ProjectMember, updatedField string) error {
	return db.Model(&model.ProjectMember{}).
		Where("project_id = ?", projectID).
		Select(updatedField).
		Updates(updateModel).Error
}

// UpdateLeftDateByProjectID just update left_date by projectID
func (s *store) UpdateLeftDateByProjectID(db *gorm.DB, projectID string) error {
	now := time.Now()
	return db.Model(&model.ProjectMember{}).
		Where("project_id = ? AND (left_date IS NULL OR left_date > ?)", projectID, now).
		Select("left_date").
		Updates(model.ProjectMember{LeftDate: &now}).Error
}

// IsExistsByEmployeeID check ProjectMember existance by project id and employee id
func (s *store) IsExistsByEmployeeID(db *gorm.DB, projectID string, employeeID string) (bool, error) {
	var record struct {
		Result bool
	}

	query := db.Raw("SELECT EXISTS (SELECT * FROM project_members WHERE project_id  = ? and employee_id = ?) as result", projectID, employeeID)
	return record.Result, query.Scan(&record).Error
}

// GetActiveByProjectIDs get project member by porjectID list
func (s *store) GetActiveByProjectIDs(db *gorm.DB, projectIDs []string) ([]*model.ProjectMember, error) {
	var members []*model.ProjectMember
	return members, db.Joins("JOIN employees ON project_members.employee_id = employees.id").Where("(project_members.left_date IS NULL OR project_members.left_date > ?) AND project_members.status = 'active' AND employees.working_status = 'full-time' AND project_members.project_id IN ?", time.Now(), projectIDs).Preload("Employee").Find(&members).Error
}

func (s *store) GetActiveByProjectIDAndEmployeeID(db *gorm.DB, projectID string, employeeID string) (*model.ProjectMember, error) {
	var member *model.ProjectMember
	return member, db.Where("project_id = ? AND employee_id = ? AND (left_date IS NULL OR left_date > now())", projectID, employeeID).First(&member).Error
}
