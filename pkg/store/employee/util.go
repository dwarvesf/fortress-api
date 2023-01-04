package employee

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

func getByWhereConditions(query *gorm.DB, filter EmployeeFilter) *gorm.DB {
	if len(filter.WorkingStatuses) > 0 {
		query = query.Where("working_status IN ?", filter.WorkingStatuses)
	}

	if len(filter.Positions) > 0 {
		query = query.Joins("JOIN employee_positions ON employees.id = employee_positions.employee_id JOIN positions ON employee_positions.position_id = positions.id AND positions.code IN ?",
			filter.Positions)
	}

	if len(filter.Stacks) > 0 {
		query = query.Joins("JOIN employee_stacks ON employees.id = employee_stacks.employee_id JOIN stacks ON employee_stacks.stack_id = stacks.id AND stacks.code IN ?",
			filter.Stacks)
	}

	if len(filter.Projects) > 0 {
		query = query.Joins("JOIN project_members ON employees.id = project_members.employee_id JOIN projects ON project_members.project_id = projects.id AND projects.code IN ?",
			filter.Projects)
	}

	if len(filter.Chapters) > 0 {
		query = query.Joins("JOIN employee_chapters ON employees.id = employee_chapters.employee_id JOIN chapters ON employee_chapters.chapter_id = chapters.id AND chapters.code IN ?",
			filter.Chapters)
	}

	if len(filter.Seniorities) > 0 {
		query = query.Joins("JOIN seniorities ON employees.seniority_id = seniorities.id AND seniorities.code IN ?",
			filter.Seniorities)
	}

	if filter.Keyword != "" {
		query = query.Where("employees.keyword_vector @@ plainto_tsquery('english_nostop', fn_remove_vietnamese_accents(LOWER(?)))", filter.Keyword)
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
