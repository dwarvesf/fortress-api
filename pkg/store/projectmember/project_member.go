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

// OneByID return a project member by id
func (s *store) OneByID(db *gorm.DB, id string) (*model.ProjectMember, error) {
	var member *model.ProjectMember
	return member, db.Where("id = ?", id).First(&member).Error
}

// OneBySlotID return a project member by slotID
func (s *store) OneBySlotID(db *gorm.DB, slotID string) (*model.ProjectMember, error) {
	var member *model.ProjectMember
	return member, db.Where("project_slot_id = ? AND status = ?", slotID, model.ProjectMemberStatusActive).
		Preload("Employee").
		First(&member).Error
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

// UpdateEndDateByProjectID just update end_date by projectID
func (s *store) UpdateEndDateByProjectID(db *gorm.DB, projectID string) error {
	now := time.Now()
	return db.Model(&model.ProjectMember{}).
		Where("project_id = ? AND (end_date IS NULL OR end_date > ?)", projectID, now).
		Select("end_date").
		Updates(model.ProjectMember{EndDate: &now}).Error
}

// IsExistsByEmployeeID check ProjectMember existence by project id and employee id
func (s *store) IsExistsByEmployeeID(db *gorm.DB, projectID string, employeeID string) (bool, error) {
	var record struct {
		Result bool
	}

	query := db.Raw("SELECT EXISTS (SELECT * FROM project_members WHERE project_id  = ? and employee_id = ?) as result", projectID, employeeID)
	return record.Result, query.Scan(&record).Error
}

// GetActiveByProjectIDs get project member by projectID list
func (s *store) GetActiveByProjectIDs(db *gorm.DB, projectIDs []string) ([]*model.ProjectMember, error) {
	var members []*model.ProjectMember
	return members, db.Joins("JOIN employees ON project_members.employee_id = employees.id").Where("(project_members.end_date IS NULL OR project_members.end_date > ?) AND project_members.status = 'active' AND employees.working_status = 'full-time' AND project_members.project_id IN ?", time.Now(), projectIDs).Preload("Employee").Find(&members).Error
}

func (s *store) GetActiveMemberInProject(db *gorm.DB, projectID string, employeeID string) (*model.ProjectMember, error) {
	var member *model.ProjectMember
	return member, db.
		Where("project_id = ?", projectID).
		Where("employee_id = ?", employeeID).
		Where("(end_date IS NULL OR end_date > now())").
		Preload("Employee").
		First(&member).Error
}

func (s *store) GetActiveMembersBySlotID(db *gorm.DB, slotID string) ([]*model.ProjectMember, error) {
	var members []*model.ProjectMember
	return members, db.Where("project_slot_id = ? AND status = ?", slotID, model.ProjectMemberStatusActive).Find(&members).Error
}

func (s *store) GetAssignedMembers(db *gorm.DB, projectID string, status string, preload bool) ([]*model.ProjectMember, error) {
	timeNow := time.Now()

	query := db.Table("project_members").
		Joins("LEFT JOIN seniorities s ON project_members.seniority_id = s.id").
		Joins(`LEFT JOIN project_heads ph ON (project_members.end_date IS NULL OR project_members.end_date > ?)
			AND project_members.project_id = ph.project_id 
			AND project_members.employee_id = ph.employee_id 
			AND ph.deleted_at IS NULL
			AND (ph.end_date IS NULL OR ph.end_date > ?)
			AND ph.position = ?
		`, timeNow, timeNow, model.HeadPositionTechnicalLead).
		Where("project_members.deleted_at IS NULL AND project_members.project_id = ?", projectID).
		Order("project_members.end_date DESC, ph.created_at, s.level DESC").
		Preload("Employee", "deleted_at IS NULL").
		Preload("Employee.Referrer", "deleted_at IS NULL")

	switch status {
	case model.ProjectMemberStatusOnBoarding.String():
		query = query.Where("project_members.start_date > ?", timeNow)

	case model.ProjectMemberStatusActive.String():
		query = query.Where("project_members.start_date <= ?", timeNow).
			Where("(project_members.end_date IS NULL OR project_members.end_date > ?)", timeNow)

	case model.ProjectMemberStatusInactive.String():
		query = query.Where("project_members.end_date <= ?", timeNow)
	}

	if preload {
		query = query.Preload("Seniority", "deleted_at IS NULL").
			Preload("UpsellPerson", "deleted_at IS NULL").
			Preload("ProjectMemberPositions", "deleted_at IS NULL").
			Preload("ProjectMemberPositions.Position", "deleted_at IS NULL")
	}

	var members []*model.ProjectMember
	return members, query.Find(&members).Error
}

// UpdateEndDateOverdueMemberToInActive just update end_date by projectID
func (s *store) UpdateEndDateOverdueMemberToInActive(db *gorm.DB) error {
	sql := `
		UPDATE project_members
		SET  status   = 'inactive'
		WHERE status = 'active' 
			AND end_date >= (NOW() AT TIME ZONE 'ICT')::DATE;
	`
	return db.Exec(sql).Error
}

// UpdateMemberInClosedProjectToInActive just update if project is closed or paused
func (s *store) UpdateMemberInClosedProjectToInActive(db *gorm.DB) error {
	sql := `
		UPDATE project_members pm
		SET
			status = 'inactive',
			end_date = p.end_date
		FROM
			projects p
		WHERE
			pm.project_id = p.id
			AND pm.status <> 'inactive'
			AND p.status IN ('closed', 'paused');
	`
	return db.Exec(sql).Error
}

// UpdateLeftMemberToInActive just update if employee is left
func (s *store) UpdateLeftMemberToInActive(db *gorm.DB) error {
	sql := `
		UPDATE project_members pm
		SET
			status = 'inactive',
			end_date = e.left_date
		FROM
			employees e
		WHERE
			pm.employee_id = e.id
			and pm.status <> 'inactive'
			AND e.working_status IN ('left');
	`
	return db.Exec(sql).Error
}

func (s *store) UpdateMemberToInActiveByID(db *gorm.DB, id string, endDate *time.Time) error {
	sql := `
		UPDATE project_members pm
		SET
			status = 'inactive',
			end_date = ?
		WHERE
			employee_id = ?
			AND (status <> 'inactive' OR end_date IS NULL OR end_date > NOW())
			
	`
	return db.Exec(sql, endDate, id).Error
}
