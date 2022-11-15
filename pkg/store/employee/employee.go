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
	return employee, s.db.Where("id = ?", id).First(&employee).Error
}

// OneByTeamEmail get 1 employee by team email
func (s *store) OneByTeamEmail(teamEmail string) (*model.Employee, error) {
	var employee *model.Employee
	return employee, s.db.Where("team_email = ?", teamEmail).First(&employee).Error
}

type Preloadable interface {
	Connection() string
	PreloadParams() []interface{}
}

type PreloadString string

func (p PreloadString) Connection() string {
	return string(p)
}

func (p PreloadString) PreloadParams() []interface{} {
	return []interface{}{p}
}

type PreloadParam struct {
	Field     string
	Condition string
}

func (p PreloadParam) Connection() string {
	return p.Field
}

func (p PreloadParam) PreloadParams() []interface{} {
	return []interface{}{p.Connection, p.Condition}
}

func InitPreloadable(params ...interface{}) Preloadable {

	if len(params) == 0 {
		panic("params is empty")
	}

	if len(params) == 1 {
		return PreloadString(params[0].(string))
	}

	var rs []Preloadable
	for _, p := range params {
		switch p.(type) {
		case string:
			rs = append(rs, PreloadString(p.(string)))
		}
	}
	return PreloadParam{
		Field:     params[0].(string),
		Condition: params[1].(string),
	}
}

var preloadConfig = map[string]Preloadable{}

// Search get employees by filter and pagination
func (s *store) Search(filter SearchFilter, pagination model.Pagination, preloads ...Preloadable) ([]*model.Employee, int64, error) {
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

	query = PreloadData(query, []Preloadable{
		InitPreloadable("ProjectMembers", "deleted_at IS NULL"),
		InitPreloadable("ProjectMembers.Project"),
		InitPreloadable("ProjectMembers.Project.Heads"),
		InitPreloadable("EmployeePositions.Position"),
		InitPreloadable("EmployeePositions", "deleted_at IS NULL"),
	})

	// query = query.Preload("ProjectMembers", "deleted_at IS NULL").
	// 	Preload("ProjectMembers.Project").
	// 	Preload("ProjectMembers.Project.Heads").
	// 	Preload("EmployeePositions", "deleted_at IS NULL").
	// 	Preload("EmployeePositions.Position").
	// 	Offset(offset)

	query = query.Offset(offset)

	return employees, total, query.Find(&employees).Error
}

func PreloadData(db *gorm.DB, preloads []Preloadable) *gorm.DB {
	for _, preload := range preloads {
		db = db.Preload(preload.Connection(), preload.PreloadParams()...)
	}
	return db
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
		Preload("Chapter").Preload("Seniority").Preload("LineManager").Preload("EmployeePositions").
		Preload("EmployeePositions.Position").First(&employee).Error
}
