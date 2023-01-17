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
		if filter.Projects[0] == "-" {
			query = query.Where(`employees.id NOT IN (
				SELECT project_members.employee_id 
				FROM project_members
				WHERE project_members.deleted_at IS NULL AND project_members.status = 'active')`)
		} else {
			query = query.Where(`employees.id IN (
				SELECT pm.employee_id
				FROM project_members pm JOIN projects p ON pm.project_id = p.id AND p.code IN ?
				WHERE pm.deleted_at IS NULL AND pm.status = 'active'
			)`, filter.Projects)
		}
	}

	if len(filter.LineManagers) > 0 {
		if filter.LineManagers[0] == "-" {
			query = query.Where("employees.line_manager_id IS NULL")
		} else {
			query = query.Where("employees.line_manager_id IN ?", filter.LineManagers)
		}
	}

	if len(filter.Chapters) > 0 {
		query = query.Joins("JOIN employee_chapters ON employees.id = employee_chapters.employee_id JOIN chapters ON employee_chapters.chapter_id = chapters.id AND chapters.code IN ?",
			filter.Chapters)
	}

	if len(filter.Seniorities) > 0 {
		query = query.Joins("JOIN seniorities ON employees.seniority_id = seniorities.id AND seniorities.code IN ?",
			filter.Seniorities)
	}

	if len(filter.Organizations) > 0 {
		if filter.Organizations[0] == "-" {
			query = query.Joins(`LEFT JOIN employee_organizations ON employees.id = employee_organizations.employee_id AND employee_organizations.id IS NULL`)
		} else {
			query = query.Joins(`JOIN employee_organizations ON employees.id = employee_organizations.employee_id JOIN organizations ON employee_organizations.organization_id = organizations.id AND organizations.code IN ?`,
				filter.Organizations)
		}
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
