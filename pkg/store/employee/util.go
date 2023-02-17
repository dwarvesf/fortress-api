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
		if len(filter.Positions) == 1 {
			if filter.Positions[0] == "-" {
				query = query.Joins(`LEFT JOIN employee_positions ep ON employees.id = ep.employee_id`).Where("ep.id IS NULL")
			} else {
				query = query.Joins(`JOIN employee_positions ep ON employees.id = ep.employee_id`).Joins(`JOIN positions po ON ep.position_id = po.id`).
					Where("po.code = ? AND ep.deleted_at IS NULL", filter.Positions[0])
			}
		}
		if len(filter.Positions) > 1 {
			if filter.Positions[0] == "-" {
				query = query.Joins(`LEFT JOIN employee_positions ep ON employees.id = ep.employee_id`).
					Joins(`LEFT JOIN positions po ON ep.position_id = po.id`).
					Where(`ep.id IS NULL OR (po.code IN ? AND ep.deleted_at IS NULL)`, filter.Positions[1:])
			} else {
				query = query.Joins(`JOIN employee_positions ep ON employees.id = ep.employee_id`).
					Joins(`JOIN positions po ON ep.position_id = po.id`).
					Where(`po.code IN ? AND ep.deleted_at IS NULL`, filter.Positions)
			}
		}
	}

	if len(filter.Stacks) > 0 {
		if len(filter.Stacks) == 1 {
			if filter.Stacks[0] == "-" {
				query = query.Joins(`LEFT JOIN employee_stacks es ON employees.id = es.employee_id`).Where("es.id IS NULL")
			} else {
				query = query.Joins(`JOIN employee_stacks es ON employees.id = es.employee_id`).Joins(`JOIN stacks s ON es.stack_id = s.id`).
					Where("s.code = ? AND es.deleted_at IS NULL", filter.Stacks[0])
			}
		}
		if len(filter.Stacks) > 1 {
			if filter.Stacks[0] == "-" {
				query = query.Joins(`LEFT JOIN employee_stacks es ON employees.id = es.employee_id`).
					Joins(`LEFT JOIN stacks s ON es.project_id = s.id`).
					Where(`es.id IS NULL OR (s.code IN ? AND es.deleted_at IS NULL)`, filter.Stacks[1:])
			} else {
				query = query.Joins(`JOIN employee_stacks es ON employees.id = es.employee_id`).
					Joins(`JOIN stacks s ON es.stack_id = s.id`).
					Where(`s.code IN ? AND es.deleted_at IS NULL`, filter.Stacks)
			}
		}
	}

	if len(filter.Projects) > 0 {
		if len(filter.Projects) == 1 {
			if filter.Projects[0] == "-" {
				query = query.Joins(`LEFT JOIN project_members pm ON employees.id = pm.employee_id`).
					Where("pm.id IS NULL OR pm.status = ? OR pm.start_date > now() OR pm.end_date <= now()", model.ProjectMemberStatusInactive)
			} else {
				query = query.Joins(`JOIN project_members pm ON employees.id = pm.employee_id`).Joins(`JOIN projects p ON pm.project_id = p.id`).
					Where("p.code = ? AND pm.deleted_at IS NULL AND pm.status = 'active'", filter.Projects[0])
			}
		}
		if len(filter.Projects) > 1 {
			if filter.Projects[0] == "-" {
				query = query.Joins(`LEFT JOIN project_members pm ON employees.id = pm.employee_id`).
					Joins(`LEFT JOIN projects p ON pm.project_id = p.id`).
					Where(`pm.id IS NULL OR (p.code IN ? AND pm.deleted_at IS NULL AND pm.status = 'active')`, filter.Projects[1:])
			} else {
				query = query.Joins(`JOIN project_members pm ON employees.id = pm.employee_id`).
					Joins(`JOIN projects p ON pm.project_id = p.id`).
					Where(`p.code IN ? AND pm.deleted_at IS NULL AND pm.status = 'active'`, filter.Projects)
			}
		}
	}

	if len(filter.LineManagers) > 0 {
		if len(filter.LineManagers) == 1 {
			if filter.LineManagers[0] == "-" {
				query = query.Where("employees.line_manager_id IS NULL")
			} else {
				query = query.Where("employees.line_manager_id = ?", filter.LineManagers[0])
			}
		} else {
			if filter.LineManagers[0] == "-" {
				query = query.Where(`employees.line_manager_id IS NULL AND employees.line_manager_id IN ?`, filter.LineManagers[1:])
			} else {
				query = query.Where("employees.line_manager_id IN ?", filter.LineManagers)
			}
		}

	}

	if len(filter.Chapters) > 0 {
		if len(filter.Chapters) == 1 {
			if filter.Chapters[0] == "-" {
				query = query.Joins(`LEFT JOIN employee_chapters ec ON employees.id = ec.employee_id`).Where("ec.id IS NULL")
			} else {
				query = query.Joins(`JOIN employee_chapters ec ON employees.id = ec.employee_id`).Joins(`JOIN chapters c ON ec.chapter_id = c.id`).
					Where("c.code = ? AND ec.deleted_at IS NULL", filter.Chapters[0])
			}
		}
		if len(filter.Chapters) > 1 {
			if filter.Chapters[0] == "-" {
				query = query.Joins(`LEFT JOIN employee_chapters ec ON employees.id = ec.employee_id`).
					Joins(`LEFT JOIN chapters c ON ec.chapter_id = c.id`).
					Where(`ec.id IS NULL OR (c.code IN ? AND ec.deleted_at IS NULL)`, filter.Chapters[1:])
			} else {
				query = query.Joins(`JOIN employee_chapters ec ON employees.id = ec.employee_id`).
					Joins(`JOIN chapters c ON ec.chapter_id = c.id`).
					Where(`c.code IN ? AND ec.deleted_at IS NULL`, filter.Chapters)
			}
		}
	}

	if len(filter.Seniorities) > 0 {
		query = query.Joins("JOIN seniorities ON employees.seniority_id = seniorities.id AND seniorities.code IN ?",
			filter.Seniorities)
	}

	if len(filter.Organizations) > 0 {
		if filter.Organizations[0] == "-" {
			query = query.Where(`employees.id NOT IN (SELECT DISTINCT employee_id FROM employee_organizations)`)
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

func getByFieldSort(db *gorm.DB, query *gorm.DB, field string, sortOrder model.SortOrder) *gorm.DB {
	if len(sortOrder) > 0 {
		if sortOrder == model.SortOrderASC {
			return db.Table("(?) as employees", query).Order(field + " asc")
		} else {
			return db.Table("(?) as employees", query).Order(field + " desc")
		}
	}
	return query
}
