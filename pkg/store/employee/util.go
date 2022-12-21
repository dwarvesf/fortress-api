package employee

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

func getByWhereConditions(query *gorm.DB, filter EmployeeFilter) *gorm.DB {
	if len(filter.WorkingStatuses) > 0 {
		query = query.Where("working_status IN ?", filter.WorkingStatuses)
	}

	if filter.PositionID != "" {
		query = query.Joins("JOIN employee_positions ON employees.id = employee_positions.employee_id AND employee_positions.position_id = ?",
			filter.PositionID)
	}

	if filter.StackID != "" {
		query = query.Joins("JOIN employee_stacks ON employees.id = employee_stacks.employee_id AND employee_stacks.stack_id = ?",
			filter.StackID)
	}

	if filter.ProjectID != "" {
		query = query.Joins("JOIN project_members ON employees.id = project_members.employee_id AND project_members.project_id = ?",
			filter.ProjectID)
	}

	if filter.Keyword != "" {
		query = query.Where("employees.keyword_vector @@ plainto_tsquery('english_nostop', LOWER(?))", filter.Keyword)
	}
	return query
}

func getByFieldSort(query *gorm.DB, field string, sortOrder model.SortOrder) *gorm.DB {
	if len(sortOrder) > 0 {
		if sortOrder == model.SortOrderASC {
			query = query.Order(field + " asc")
		} else {
			query = query.Order(field + " desc")
		}
	}
	return query
}
