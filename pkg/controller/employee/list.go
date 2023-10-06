package employee

import (
	errors "errors"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
)

type GetListEmployeeInput struct {
	model.Pagination

	WorkingStatuses []string
	Preload         bool
	Positions       []string
	Stacks          []string
	Projects        []string
	Chapters        []string
	Seniorities     []string
	Organizations   []string
	LineManagers    []string
	Keyword         string
}

func (r *controller) List(workingStatuses []string, body GetListEmployeeInput, userInfo *model.CurrentLoggedUserInfo) ([]*model.Employee, int64, error) {
	filter := employee.EmployeeFilter{
		Preload:        body.Preload,
		Keyword:        body.Keyword,
		Positions:      body.Positions,
		Stacks:         body.Stacks,
		Chapters:       body.Chapters,
		Seniorities:    body.Seniorities,
		Organizations:  body.Organizations,
		LineManagers:   body.LineManagers,
		JoinedDateSort: model.SortOrderDESC,
		Projects:       body.Projects,
	}

	// If user don't have this permission, they can only see employees in the project that they are in
	if !authutils.HasPermission(userInfo.Permissions, model.PermissionEmployeesReadReadActive) {
		projectIDs := make([]string, 0)
		for _, p := range userInfo.Projects {
			projectIDs = append(projectIDs, p.Code)
		}

		filter.Projects = []string{""}
		if len(projectIDs) > 0 {
			filter.Projects = projectIDs
		}
	}

	filter.WorkingStatuses = workingStatuses

	employees, total, err := r.store.Employee.All(r.repo.DB(), filter, body.Pagination)
	if err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

func (r *controller) ListWithLocation() ([]*model.Employee, error) {
	employees, err := r.store.Employee.SimpleList(r.repo.DB())
	if err != nil {
		return nil, err
	}

	return employees, nil
}

func (r *controller) ListWithMMAScore() (employees []model.EmployeeMMAScoreData, err error) {
	rs, err := r.store.Employee.ListWithMMAScore(r.repo.DB())
	if err != nil {
		return nil, err
	}

	return rs, nil
}

func (r *controller) ListByDiscordRequest(discordID, email, key string, userInfo *model.CurrentLoggedUserInfo) ([]model.Employee, error) {
	in := employee.DiscordRequestFilter{
		Email: email,
	}

	discordIDs := make([]string, 0)
	if discordID != "" {
		discordIDs = append(discordIDs, discordID)
	}

	if key != "" {
		dt, err := r.service.Discord.SearchMember(key)
		if err != nil {
			return nil, err
		}

		if len(dt) <= 0 {
			in.Keyword = key
		} else {
			for _, d := range dt {
				discordIDs = append(discordIDs, d.User.ID)
			}
		}
	}

	in.DiscordID = discordIDs

	if len(in.DiscordID) > 0 || in.Email != "" || in.Keyword != "" {
		rs, err := r.store.Employee.ListByDiscordRequest(r.repo.DB(), in, true)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrEmployeeNotFound
			}
			return nil, err
		}
		return rs, nil
	}

	return nil, ErrEmployeeNotFound
}
