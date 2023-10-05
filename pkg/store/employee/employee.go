package employee

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One get 1 employee by id
func (s *store) One(db *gorm.DB, id string, preload bool) (*model.Employee, error) {
	query := db
	if !model.IsUUIDFromString(id) {
		query = db.Where("username = ?", id)
	} else {
		query = db.Where("id = ?", id)
	}

	query = query.Preload("EmployeeRoles", func(db *gorm.DB) *gorm.DB {
		return db.Joins("employee_roles JOIN roles ON roles.id = employee_roles.role_id").
			Where("employee_roles.deleted_at IS NULL").
			Order("roles.level")
	}).
		Preload("EmployeeRoles.Role", "deleted_at IS NULL").
		Preload("SocialAccounts", "deleted_at IS NULL").
		Preload("DiscordAccount", "deleted_at IS NULL")

	if preload {
		query = query.
			Preload("ProjectMembers", "deleted_at IS NULL").
			Preload("ProjectMembers.Project", "deleted_at IS NULL").
			Preload("ProjectMembers.ProjectMemberPositions", "deleted_at IS NULL").
			Preload("ProjectMembers.ProjectMemberPositions.Position", "deleted_at IS NULL").
			Preload("EmployeePositions", "deleted_at IS NULL").
			Preload("EmployeePositions.Position", "deleted_at IS NULL").
			Preload("EmployeeStacks", "deleted_at IS NULL").
			Preload("EmployeeStacks.Stack", "deleted_at IS NULL").
			Preload("EmployeeChapters", "deleted_at IS NULL").
			Preload("EmployeeChapters.Chapter", "deleted_at IS NULL").
			Preload("EmployeeOrganizations", "deleted_at IS NULL").
			Preload("EmployeeOrganizations.Organization", "deleted_at IS NULL").
			Preload("Referrer", "deleted_at IS NULL").
			Preload("Seniority").
			Preload("LineManager").
			Preload("BaseSalary").
			Preload("BaseSalary.Currency")
	}

	var employee *model.Employee
	return employee, query.First(&employee).Error
}

// OneByEmail get 1 employee by team email or personal email
func (s *store) OneByEmail(db *gorm.DB, email string) (*model.Employee, error) {
	var employee *model.Employee

	return employee, db.Where("NULLIF(TRIM(team_email), '') = ? OR NULLIF(TRIM(personal_email), '') = ?", email, email).First(&employee).Error
}

// OneByNotionID get 1 employee by notion id
func (s *store) OneByNotionID(db *gorm.DB, notionID string) (*model.Employee, error) {
	var employee *model.Employee
	return employee, db.Where(`id IN (
			SELECT sa.employee_id
			FROM social_accounts sa
			WHERE sa.account_id = ? AND sa.type = ?
		)`, notionID, model.SocialAccountTypeNotion).First(&employee).Error
}

// OneByBasecampID get 1 employee by basecampID
func (s *store) OneByBasecampID(db *gorm.DB, basecampID int) (*model.Employee, error) {
	var employee *model.Employee
	return employee, db.Where("basecamp_id = ?", basecampID).First(&employee).Error
}

// All get employees by query and pagination
func (s *store) All(db *gorm.DB, filter EmployeeFilter, pagination model.Pagination) ([]*model.Employee, int64, error) {
	var total int64
	var employees []*model.Employee

	query := db.Table("employees").Distinct("ON(employees.id) employees.*")

	query = getByWhereConditions(query, filter)
	err := db.Raw("SELECT COUNT(*) FROM (?) res", query).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	query = getByFieldSort(db, query, "employees.joined_date", filter.JoinedDateSort).
		Preload("SocialAccounts", "deleted_at IS NULL").
		Preload("DiscordAccount", "deleted_at IS NULL")

	if filter.Preload {
		query = query.
			Preload("LineManager", "deleted_at IS NULL").
			Preload("Referrer", "deleted_at IS NULL").
			Preload("Seniority", "deleted_at IS NULL").
			Preload("ProjectMembers", func(db *gorm.DB) *gorm.DB {
				return db.Joins("JOIN projects ON projects.id = project_members.project_id").
					Where("project_members.deleted_at IS NULL").
					Where("project_members.start_date <= now()").
					Where("(project_members.end_date IS NULL OR project_members.end_date > now())").
					Where("projects.status IN ?", []string{
						model.ProjectStatusOnBoarding.String(),
						model.ProjectStatusActive.String(),
					}).
					Order("projects.name")
			}).
			Preload("ProjectMembers.Project", "deleted_at IS NULL").
			Preload("ProjectMembers.Project.Heads", "deleted_at IS NULL").
			Preload("EmployeePositions", "deleted_at IS NULL").
			Preload("EmployeePositions.Position", "deleted_at IS NULL").
			Preload("EmployeeRoles", "deleted_at IS NULL").
			Preload("EmployeeRoles.Role", "deleted_at IS NULL").
			Preload("EmployeeChapters", "deleted_at IS NULL").
			Preload("EmployeeChapters.Chapter", "deleted_at IS NULL").
			Preload("EmployeeStacks", "deleted_at IS NULL").
			Preload("EmployeeStacks.Stack", "deleted_at IS NULL").
			Preload("EmployeeOrganizations", func(db *gorm.DB) *gorm.DB {
				return db.Joins("JOIN organizations ON organizations.id = employee_organizations.organization_id").
					Where("employee_organizations.deleted_at IS NULL").
					Order("organizations.code DESC")
			}).
			Preload("EmployeeOrganizations.Organization", "deleted_at IS NULL").
			Preload("BaseSalary").
			Preload("BaseSalary.Currency")
	}

	limit, offset := pagination.ToLimitOffset()
	if pagination.Page > 0 {
		query = query.Limit(limit)
	}

	return employees, total, query.
		Limit(limit).
		Offset(offset).
		Find(&employees).Error
}

func (s *store) Create(db *gorm.DB, e *model.Employee) (employee *model.Employee, err error) {
	return e, db.Create(e).Error
}

// IsExist check the existence of employee
func (s *store) IsExist(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM employees WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}

// Update update all value (including nested model)
func (s *store) Update(db *gorm.DB, employee *model.Employee) (*model.Employee, error) {
	return employee, db.Model(&employee).Where("id = ?", employee.ID).Updates(&employee).First(&employee).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.Employee, updatedFields ...string) (*model.Employee, error) {
	employee := model.Employee{}
	return &employee, db.Model(&employee).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}

// GetByIDs return list employee by IDs
func (s *store) GetByIDs(db *gorm.DB, ids []model.UUID) ([]*model.Employee, error) {
	var employees []*model.Employee
	return employees, db.Where("id IN ?", ids).Order("created_at").Find(&employees).Error
}

// GetByWorkingStatus return list employee by working status
func (s *store) GetByWorkingStatus(db *gorm.DB, workingStatus model.WorkingStatus) ([]*model.Employee, error) {
	var employees []*model.Employee
	return employees, db.Where("working_status = ?", workingStatus).
		Preload("Organizations", "deleted_at IS NULL").
		Find(&employees).Error
}

func (s *store) GetLineManagers(db *gorm.DB) ([]*model.Employee, error) {
	var employees []*model.Employee
	return employees, db.Where(`id IN (
		SELECT e.line_manager_id
		FROM employees e
		WHERE e.deleted_at IS NULL
			AND e.working_status != ?
			AND (e.left_date IS NULL OR e.left_date >= now())
	)`, model.WorkingStatusLeft).Find(&employees).Error
}

func (s *store) GetLineManagersOfPeers(db *gorm.DB, employeeID string) ([]*model.Employee, error) {
	var employees []*model.Employee
	return employees, db.Where(`id IN (
		SELECT e.line_manager_id
		FROM employees e JOIN project_members pm ON pm.employee_id = e.id
		WHERE e.deleted_at IS NULL
			AND e.working_status != ?
			AND (e.left_date IS NULL OR e.left_date >= now())
			AND pm.project_id IN (
				SELECT pm2.project_id
				FROM project_members pm2
				WHERE pm2.employee_id = employees.id
			)
	)`, model.WorkingStatusLeft).Find(&employees).Error
}

func (s *store) GetMenteesByID(db *gorm.DB, employeeID string) ([]*model.Employee, error) {
	var employees []*model.Employee
	return employees, db.Where(`id IN (
		SELECT e.id
		FROM employees e
		WHERE e.deleted_at IS NULL
			AND e.line_manager_id = ?
			AND e.working_status <> ?
			AND (e.left_date IS NULL OR e.left_date >= now())
	)`, employeeID, model.WorkingStatusLeft).
		Preload("EmployeePositions", "deleted_at IS NULL").
		Preload("EmployeePositions.Position", "deleted_at IS NULL").
		Preload("Seniority").
		Find(&employees).Error
}

func (s *store) GetByEmails(db *gorm.DB, emails []string) ([]*model.Employee, error) {
	var employees []*model.Employee

	return employees, db.Where("NULLIF(TRIM(team_email), '') IN ? OR NULLIF(TRIM(personal_email), '') IN ?", emails, emails).Find(&employees).Error
}

func (s *store) GetByBasecampIDs(db *gorm.DB, basecampIDs []int) ([]*model.Employee, error) {
	var employees []*model.Employee

	return employees, db.Where("basecamp_id IN ?", basecampIDs).Find(&employees).Error
}

func (s *store) GetByDiscordID(db *gorm.DB, discordID string, preload bool) (*model.Employee, error) {
	var employee *model.Employee
	query := db.Joins("JOIN discord_accounts ON discord_accounts.id = employees.discord_account_id AND discord_accounts.discord_id = ?", discordID)
	if preload {
		query = query.
			Preload("SocialAccounts", "deleted_at IS NULL").
			Preload("DiscordAccount", "deleted_at IS NULL").
			Preload("ProjectMembers", "deleted_at IS NULL").
			Preload("ProjectMembers.Project", "deleted_at IS NULL").
			Preload("EmployeePositions", "deleted_at IS NULL").
			Preload("EmployeePositions.Position", "deleted_at IS NULL").
			Preload("EmployeeStacks", "deleted_at IS NULL").
			Preload("EmployeeStacks.Stack", "deleted_at IS NULL").
			Preload("EmployeeMMAScores", func(db *gorm.DB) *gorm.DB {
				return db.Order("rated_at DESC").Limit(1)
			}).
			Preload("Seniority")
	}
	return employee, query.First(&employee).Error
}

// SimpleList get employees by query and pagination
func (s *store) SimpleList(db *gorm.DB) ([]*model.Employee, error) {
	var employees []*model.Employee

	query := db.Where("deleted_at IS NULL AND working_status <> ?", model.WorkingStatusLeft).
		Order("created_at").
		Preload("DiscordAccount", "deleted_at IS NULL").
		Preload("EmployeeChapters", "deleted_at IS NULL").
		Preload("EmployeeChapters.Chapter", "deleted_at IS NULL")

	return employees, query.Find(&employees).Error
}

func (s *store) GetRawList(db *gorm.DB, filter EmployeeFilter) ([]model.Employee, error) {
	var employees []model.Employee

	query := db
	if filter.WorkingStatuses != nil {
		query = query.Where("working_status IN ?", filter.WorkingStatuses)
	}

	return employees, query.Find(&employees).Error
}

// ListWithMMAScore get employees with latest mma score
func (s *store) ListWithMMAScore(db *gorm.DB) ([]model.EmployeeMMAScoreData, error) {
	var employees []model.EmployeeMMAScoreData
	query := `
		SELECT e.id AS employee_id, e.full_name, m.id AS mma_id, m.mastery_score, m.autonomy_score, m.meaning_score, m.rated_at
		FROM employees e
		LEFT JOIN (
			SELECT id, employee_id, mastery_score, autonomy_score, meaning_score, rated_at
			FROM employee_mma_scores
			WHERE (employee_id, rated_at) IN (
				SELECT employee_id, MAX(rated_at) AS rated_at
				FROM employee_mma_scores
				GROUP BY employee_id
			)
		) m ON e.id = m.employee_id
		WHERE e.deleted_at IS NULL AND e.working_status <> ?
	`

	return employees, db.Raw(query, model.WorkingStatusLeft).Scan(&employees).Error
}
