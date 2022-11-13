package employee

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
	db *gorm.DB
}

func New(db *gorm.DB) IStore {
	return &store{
		db: db,
	}
}

// One get 1 employee by id
func (s *store) One(id string) (*model.Employee, error) {
	var employee *model.Employee
	return employee, s.db.Where("id = ?", id).
		Preload("Roles").
		Preload("Chapter").
		Preload("Seniority").
		First(&employee).
		Error
}

// OneByTeamEmail get 1 employee by team email
func (s *store) OneByTeamEmail(teamEmail string) (*model.Employee, error) {
	var employee *model.Employee
	return employee, s.db.Where("team_email = ?", teamEmail).First(&employee).Error
}

// Search get employees by filter and pagination
func (s *store) Search(filter SearchFilter, pagination model.Pagination) ([]*model.Employee, int64, error) {
	var total int64
	var employees []*model.Employee

	query := s.db.Table("employees")

	if filter.WorkingStatus != "" {
		query = query.Where("working_status = ?", filter.WorkingStatus)
	}
	query = query.Count(&total)

	query = query.Order(pagination.Sort)
	limit, offset := pagination.ToLimitOffset()
	if pagination.Page > 0 {
		query = query.Limit(limit)
	}

	query = query.Preload("ProjectMembers", "deleted_at IS NULL").
		Preload("ProjectMembers.Project").
		Preload("ProjectMembers.Project.Heads").
		Preload("Positions", "deleted_at IS NULL").
		Preload("Roles", "deleted_at IS NULL").
		Offset(offset)

	return employees, total, query.Find(&employees).Error
}

func (s *store) UpdateEmployeeStatus(employeeID string, accountStatus model.AccountStatus) (*model.Employee, error) {
	employee := &model.Employee{}
	return employee, s.db.Model(&employee).Where("id = ?", employeeID).Update("account_status", string(accountStatus)).Find(&employee).Error
}

func (s *store) UpdateGeneralInfo(body EditGeneralInfo, id string) (*model.Employee, error) {
	employee := &model.Employee{}

	// 1.2 update infor
	employee.FullName = body.Fullname
	employee.TeamEmail = body.Email
	employee.PhoneNumber = body.Phone
	employee.DiscordID = body.DiscordID
	employee.GithubID = body.GithubID

	if body.LineManagerID != "" {
		employee.LineManagerID = model.MustGetUUIDFromString(body.LineManagerID)
	}

	// 1.3 save to DB
	return employee, s.db.Table("employees").Where("id = ?", id).Updates(&employee).
		Preload("Chapter").
		Preload("Seniority").
		Preload("LineManager").
		Preload("Positions", "deleted_at IS NULL").
		Preload("Roles", "deleted_at IS NULL").
		First(&employee).Error
}

func (s *store) Create(e *model.Employee) (employee *model.Employee, err error) {
	return e, s.db.Create(e).Error
}
