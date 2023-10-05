package employee

import (
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
