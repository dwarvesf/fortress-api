package employee

import (
	"time"

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

type EditEmployeeRequest struct {
	Fullname    string     `form:"fullName" json:"fullName" binding:"required,max=99"`
	Email       string     `form:"email" json:"email" binding:"required,email"`
	Phone       string     `form:"phone" json:"phone" binding:"required,max=12,min=10"`
	LineManager model.UUID `form:"lineManager" json:"lineManager"`
	DiscordID   string     `form:"discordID" json:"discordID"`
	GithubID    string     `form:"githubID" json:"githubID"`

	Role      model.UUID   `form:"role" json:"role" binding:"required"`
	Chapter   model.UUID   `form:"chapter" json:"chapter"`
	Seniority model.UUID   `form:"seniority" json:"seniority" binding:"required"`
	Stack     []model.UUID `form:"stack" json:"stack" binding:"required"`

	DoB           *time.Time `form:"DoB" json:"DoB" binding:"required"`
	Gender        string     `form:"gender" json:"gender" binding:"required"`
	Address       string     `form:"address" json:"address" binding:"required,max=200"`
	PersonalEmail string     `form:"personalEmail" json:"personalEmail" binding:"required,email"`
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

// Search get employees by filter and pagination
func (s *store) Search(filter SearchFilter, pagination model.Pagination) ([]*model.Employee, int64, error) {
	db := s.db.Table("employees")
	var total int64

	if filter.WorkingStatus != "" {
		db = db.Where("working_status = ?", filter.WorkingStatus)
	}
	db = db.Count(&total)

	if pagination.Page > 1 {
		db = db.Offset(int((pagination.Page - 1) * pagination.Size))
	}
	db = db.Limit(int(pagination.Size))
	db = db.Order(pagination.Sort)

	var employees []*model.Employee
	return employees, total, db.Find(&employees).Error
}

func (s *store) Save(employee *model.Employee) error {
	return s.db.Save(&employee).Error
}

func (s *store) UpdateGeneralInfo(body EditGeneralInfo, id string) (*model.Employee, error) {
	employee := &model.Employee{}
	// 1.2 update infor
	employee.FullName = body.Fullname
	employee.TeamEmail = body.Email
	employee.PhoneNumber = body.Phone
	employee.LineManagerID = body.LineManager
	employee.DiscordID = body.DiscordID
	employee.GithubID = body.GithubID
	// 1.3 save to DB
	return employee, s.db.Table("employees").Where("id = ?", id).Updates(&employee).Preload("Chapter").Preload("Seniority").Preload("Position").Preload("LineManager").Preload("EmployeeRoles").Preload("EmployeeRoles.Role").Find(&employee).Error
}

func (s *store) UpdateSkills(body EditSkills, id string) (*model.Employee, error) {

	var employee *model.Employee

	return employee, s.db.Transaction(func(tx *gorm.DB) error {
		// get employee by employee id
		err := tx.Where("id = ?", id).First(&employee).Error
		if err != nil {
			return err
		}

		// 1.1 delete all roles of the employee
		employeeRole := model.EmployeeRoles{}
		tx.Table("employee_roles").Where("employee_id = ?", id).Delete(&employeeRole)

		// 1.2 create role for employee
		employeeRole.ID = model.NewUUID()
		employeeRole.EmployeeID = model.MustGetUUIDFromString(id)
		employeeRole.RoleID = body.Role

		err = tx.Table("employee_roles").Create(&employeeRole).Error
		if err != nil {
			return err
		}

		// 2 update tech stacks for employee

		// 2.1 delete all employee_tech_stacks for employee id
		employeeTechStack := model.EmployeeTechStack{}
		tx.Table("employee_tech_stacks").Where("employee_id = ?", id).Delete(&employeeTechStack)

		// 2.2 create tech_stacks for employee
		for _, techID := range body.Stack {
			employeeTechStack.ID = model.NewUUID()
			employeeTechStack.EmployeeID = model.MustGetUUIDFromString(id)
			employeeTechStack.TechStackID = techID

			err = tx.Table("employee_tech_stacks").Create(&employeeTechStack).Error
			if err != nil {
				return err
			}
		}

		// 3 update employee table

		// 3.2 update infor
		employee.ChapterID = body.Chapter
		employee.SeniorityID = body.Seniority

		// 3.3 save to DB
		err = tx.Table("employees").Updates(&employee).Preload("Chapter").Preload("Seniority").Preload("Position").Preload("LineManager").Preload("EmployeeRoles").Preload("EmployeeRoles.Role").Find(&employee).Error
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *store) UpdatePersonalInfo(body EditPersonalInfo, id string) (*model.Employee, error) {
	employee := &model.Employee{}
	// 1.2 update infor
	employee.DateOfBirth = body.DoB
	employee.Gender = body.Gender
	employee.Address = body.Address
	employee.PersonalEmail = body.PersonalEmail

	// 1.3 save to DB
	return employee, s.db.Table("employees").Where("id = ?", id).Updates(&employee).Preload("Chapter").Preload("Seniority").Preload("Position").Preload("LineManager").Preload("EmployeeRoles").Preload("EmployeeRoles.Role").Find(&employee).Error
}

func (s *store) UpdateEmployeeStatus(employeeID string, accountStatus model.AccountStatus) (*model.Employee, error) {
	employee := &model.Employee{}
	return employee, s.db.Model(&employee).Where("id = ?", employeeID).Update("account_status", string(accountStatus)).Find(&employee).Error
}
