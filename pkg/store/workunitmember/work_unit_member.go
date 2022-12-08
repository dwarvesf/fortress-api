package workunitmember

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One return a work member by workUnitID and employeeID
func (s *store) One(db *gorm.DB, workUnitID string, employeeID string, status string) (*model.WorkUnitMember, error) {
	var member *model.WorkUnitMember
	query := db.Where("work_unit_id = ? AND employee_id = ? and status = ?", workUnitID, employeeID, status)

	return member, query.First(&member).Error
}

// Create create new WorkUnitMember
func (s *store) Create(db *gorm.DB, wum *model.WorkUnitMember) error {
	return db.Create(&wum).Error
}

// GetByWorkUnitID return list member of a work unit
func (s *store) GetByWorkUnitID(db *gorm.DB, wuID string) (wuMembers []model.WorkUnitMember, err error) {
	var members []model.WorkUnitMember
	return members, db.Where("work_unit_id = ?", wuID).Find(&members).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.WorkUnitMember, updatedFields ...string) (*model.WorkUnitMember, error) {
	member := model.WorkUnitMember{}
	return &member, db.Debug().Model(&member).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}

// DeleteByWorkUnitID delete many workUnitMember by workUnitID
func (s *store) DeleteByWorkUnitID(db *gorm.DB, workUnitID string) error {
	return db.Unscoped().Where("work_unit_id = ?", workUnitID).Delete(&model.WorkUnitMember{}).Error
}

// SoftDeleteByWorkUnitID delete one workUnitMember by EmployeeID and workUnitID
func (s *store) SoftDeleteByWorkUnitID(db *gorm.DB, workUnitID string, employeeID string) error {
	return db.Where("work_unit_id = ? and employee_id = ?", workUnitID, employeeID).Delete(&model.WorkUnitMember{}).Error
}

// All get all active members of a work unit
func (s *store) All(db *gorm.DB, workUnitID string) ([]*model.WorkUnitMember, error) {
	var members []*model.WorkUnitMember
	return members, db.Where("work_unit_id = ? and status = 'active'", workUnitID).Find(&members).Error
}

func (s *store) GetPeerReviewerInTimeRange(db *gorm.DB, from *time.Time, to *time.Time) ([]model.WorkUnitPeer, error) {
	var peers []model.WorkUnitPeer

	query := db.Raw(`
	WITH peer AS (
			SELECT
				employee_id,
				reviewer_id
			FROM (
				SELECT
					w1.employee_id,
					w1.work_unit_id,
					w2.employee_id AS reviewer_id,
					w1.project_id
				FROM
					work_unit_members w1
					JOIN work_unit_members w2 ON w1.work_unit_id = w2.work_unit_id
				WHERE (w1.joined_date < ?
					AND(w1.left_date > ?
						OR w1.left_date IS NULL))
				AND(w2.joined_date < ?
					AND(w2.left_date > ?
						OR w2.left_date IS NULL))) a
			JOIN employees e ON a.employee_id = e.id
		WHERE
			employee_id <> reviewer_id
			AND e.working_status = 'full-time'
		GROUP BY
			employee_id,
			reviewer_id
		HAVING
			count(*) = 1
		)
		SELECT
			p.employee_id,
			p.reviewer_id
		FROM
			peer p
			JOIN employees e ON e.id = p.reviewer_id
		WHERE
			e.line_manager_id <> p.reviewer_id
			AND e.working_status = 'full-time';	
	`, to, from, to, from)
	return peers, query.Scan(&peers).Error
}
