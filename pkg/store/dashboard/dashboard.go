package dashboard

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) GetProjectSizes(db *gorm.DB) ([]*model.ProjectSize, error) {
	var ru []*model.ProjectSize

	query := `
		SELECT projects.id, projects.name, projects.code, projects.avatar, count(*) AS size 
			FROM (projects 
				JOIN project_members pm ON projects.id = pm.project_id
				JOIN organizations ON projects.organization_id = organizations.id)
			WHERE projects.function = 'development' 
				AND organizations.code = 'dwarves-foundation' 
				AND pm.status = 'active' AND (pm.status = 'active' OR pm.status='on-boarding') AND projects.deleted_at IS NULL
			GROUP BY projects.id
			ORDER BY size DESC
	`

	return ru, db.Raw(query).Scan(&ru).Error
}

func (s *store) GetWorkSurveysByProjectID(db *gorm.DB, projectID string) ([]*model.WorkSurvey, error) {
	var rs []*model.WorkSurvey

	query := `
		SELECT feedback_events.end_date,
			AVG(
				CASE
					WHEN "order" = 1 THEN 
						CASE
							When length(answer) = 0 THEN null
							ELSE cast(answer AS integer)
							END
					END) as workload,
			AVG(
				CASE
					WHEN "order" = 2 THEN 
						CASE
							When length(answer) = 0 THEN null
							ELSE cast(answer AS integer)
							END
					END) as deadline,
			AVG(
				CASE
					WHEN "order" = 3 THEN
						CASE
							When length(answer) = 0 THEN null
							ELSE cast(answer AS integer)
							END
					END) as learning
		FROM feedback_events
				JOIN employee_event_topics eet ON feedback_events.id = eet.event_id
				JOIN employee_event_questions eeq ON feedback_events.id = eeq.event_id
		WHERE eet.project_id = ? AND feedback_events.subtype='work' AND feedback_events.deleted_at IS NULL AND eet.deleted_at IS NULL AND eeq.deleted_at IS NULL
		GROUP BY feedback_events.end_date
		ORDER BY feedback_events.end_date
		LIMIT 6
	`

	return rs, db.Raw(query, projectID).Scan(&rs).Error
}

func (s *store) GetAllWorkSurveys(db *gorm.DB) ([]*model.WorkSurvey, error) {
	var rs []*model.WorkSurvey

	query := `
		SELECT feedback_events.end_date,
			AVG(
				CASE
					WHEN "order" = 1 THEN
						CASE
							When length(answer) = 0 THEN null
							ELSE cast(answer AS integer)
							END
					END) as workload,
			AVG(
				CASE
					WHEN "order" = 2 THEN
						CASE
							When length(answer) = 0 THEN null
							ELSE cast(answer AS integer)
							END
					END) as deadline,
			AVG(
				CASE
					WHEN "order" = 3 THEN
						CASE
							When length(answer) = 0 THEN null
							ELSE cast(answer AS integer)
							END
					END) as learning
		FROM feedback_events
			JOIN employee_event_topics eet ON feedback_events.id = eet.event_id
			JOIN projects p ON eet.project_id = p.id
			JOIN organizations ON p.organization_id = organizations.id
			JOIN employee_event_questions eeq ON feedback_events.id = eeq.event_id
		WHERE p.function = 'development' AND organizations.code = 'dwarves-foundation' AND feedback_events.subtype='work' AND feedback_events.deleted_at IS NULL AND eet.deleted_at IS NULL AND eeq.deleted_at IS NULL
		GROUP BY feedback_events.end_date
		ORDER BY feedback_events.end_date
		LIMIT 6
	`

	return rs, db.Raw(query).Scan(&rs).Error
}

func (s *store) GetAllActionItemReports(db *gorm.DB) ([]*model.ActionItemReport, error) {
	var rs []*model.ActionItemReport

	query := `
		SELECT sum(action_item_high) AS high, sum(action_item_medium) AS medium, sum(action_item_low) AS low, quarter AS quarter
		FROM audit_cycles
		WHERE audit_cycles.deleted_at IS NULL
		GROUP BY quarter 
		ORDER BY quarter desc 
		LIMIT 4
	`

	return rs, db.Raw(query).Scan(&rs).Error
}

func (s *store) GetActionItemReportsByProjectNotionID(db *gorm.DB, projectID string) ([]*model.ActionItemReport, error) {
	var rs []*model.ActionItemReport

	query := `
		SELECT sum(audit_cycles.action_item_high) AS high, sum(audit_cycles.action_item_medium) AS medium, sum(audit_cycles.action_item_low) AS low, audit_cycles.quarter AS quarter
		FROM audit_cycles
		WHERE audit_cycles.project_id = ? AND audit_cycles.deleted_at IS NULL
		GROUP BY quarter
		ORDER BY quarter desc
		LIMIT 4
	`

	return rs, db.Raw(query, projectID).Scan(&rs).Error
}

func (s *store) AverageEngineeringHealth(db *gorm.DB) ([]*model.AverageEngineeringHealth, error) {
	var rs []*model.AverageEngineeringHealth

	query := `
		SELECT audit_cycles.quarter,
			avg(CASE
					When a.score = 0 THEN NULL
					ELSE a.score
				END) as avg
		FROM audit_cycles
				LEFT JOIN audits AS a ON audit_cycles.health_audit_id = a.id
		WHERE audit_cycles.deleted_at IS NULL AND a.deleted_at IS NULL
		GROUP BY audit_cycles.quarter
		ORDER BY audit_cycles.quarter DESC
		LIMIT 4
	`

	return rs, db.Raw(query).Scan(&rs).Error
}

func (s *store) GroupEngineeringHealth(db *gorm.DB) ([]*model.GroupEngineeringHealth, error) {
	var rs []*model.GroupEngineeringHealth

	query := `
		SELECT audit_cycles.quarter, ai.area,
			avg(CASE
					When ai.grade = 0 THEN NULL
					ELSE ai.grade
				END) as avg
		FROM audit_cycles
				LEFT JOIN audits AS a ON audit_cycles.health_audit_id = a.id
				LEFT JOIN audit_items AS ai ON a.id = ai.audit_id
		WHERE audit_cycles.deleted_at IS NULL AND a.deleted_at IS NULL AND ai.deleted_at IS NULL
		GROUP BY audit_cycles.quarter, ai.area
		ORDER BY audit_cycles.quarter DESC
		LIMIT 16
	`

	return rs, db.Raw(query).Scan(&rs).Error
}

func (s *store) AverageEngineeringHealthByProjectNotionID(db *gorm.DB, projectID string) ([]*model.AverageEngineeringHealth, error) {
	var rs []*model.AverageEngineeringHealth

	query := `
		SELECT audit_cycles.quarter,
			avg(CASE
					When a.score = 0 THEN NULL
					ELSE a.score
				END) as avg
		FROM audit_cycles
				LEFT JOIN audits AS a ON audit_cycles.health_audit_id = a.id
		WHERE audit_cycles.project_id = ? AND audit_cycles.deleted_at IS NULL AND a.deleted_at IS NULL
		GROUP BY audit_cycles.quarter
		ORDER BY audit_cycles.quarter DESC
		LIMIT 4
	`

	return rs, db.Raw(query, projectID).Scan(&rs).Error
}

func (s *store) GroupEngineeringHealthByProjectNotionID(db *gorm.DB, projectID string) ([]*model.GroupEngineeringHealth, error) {
	var rs []*model.GroupEngineeringHealth

	query := `
		SELECT audit_cycles.quarter, ai.area,
			avg(CASE
					When ai.grade = 0 THEN NULL
					ELSE ai.grade
				END) as avg
		FROM audit_cycles
				LEFT JOIN audits AS a ON audit_cycles.health_audit_id = a.id
				LEFT JOIN audit_items AS ai ON a.id = ai.audit_id
		WHERE audit_cycles.project_id = ? AND audit_cycles.deleted_at IS NULL AND a.deleted_at IS NULL AND ai.deleted_at IS NULL
		GROUP BY audit_cycles.quarter, ai.area
		ORDER BY audit_cycles.quarter DESC
		LIMIT 16
	`

	return rs, db.Raw(query, projectID).Scan(&rs).Error
}

func (s *store) GetAverageAudit(db *gorm.DB) ([]*model.AverageAudit, error) {
	var rs []*model.AverageAudit

	query := `
		SELECT audit_cycles.quarter,
			avg(CASE
					When audit_cycles.average_score = 0 THEN null
					ELSE audit_cycles.average_score
				END) AS avg
		FROM audit_cycles
		WHERE audit_cycles.deleted_at IS NULL
		GROUP BY audit_cycles.quarter
		ORDER BY audit_cycles.quarter DESC
		LIMIT 4
	`

	return rs, db.Raw(query).Scan(&rs).Error
}

func (s *store) GetGroupAudit(db *gorm.DB) ([]*model.GroupAudit, error) {
	var rs []*model.GroupAudit

	query := `
		SELECT audit_cycles.quarter,
			avg(CASE
					When asy.score = 0 THEN null
					ELSE asy.score
				END) as system,
			avg(CASE
					When afe.score = 0 THEN null
					ELSE afe.score
				END) as frontend,
			avg(CASE
					When abe.score = 0 THEN null
					ELSE abe.score
				END) as backend,
			avg(CASE
					When ap.score = 0 THEN null
					ELSE ap.score
				END) as process,
			avg(CASE
					When ab.score = 0 THEN null
					ELSE ab.score
				END) as blockchain,
			avg(CASE
					When am.score = 0 THEN null
					ELSE am.score
				END) as mobile
		FROM audit_cycles
				LEFT JOIN audits AS afe ON audit_cycles.frontend_audit_id = afe.id
				LEFT JOIN audits AS abe ON audit_cycles.backend_audit_id = abe.id
				LEFT JOIN audits AS ap ON audit_cycles.process_audit_id = ap.id
				LEFT JOIN audits AS asy ON audit_cycles.system_audit_id = asy.id
				LEFT JOIN audits AS ab ON audit_cycles.blockchain_audit_id = ab.id
				LEFT JOIN audits AS am ON audit_cycles.mobile_audit_id = am.id
		WHERE audit_cycles.deleted_at IS NULL AND afe.deleted_at IS NULL AND 
			abe.deleted_at IS NULL AND ap.deleted_at IS NULL AND 
			asy.deleted_at IS NULL AND ab.deleted_at IS NULL AND am.deleted_at IS NULL
		GROUP BY audit_cycles.quarter
		ORDER BY audit_cycles.quarter DESC
		LIMIT 4
	`

	return rs, db.Raw(query).Scan(&rs).Error
}

func (s *store) GetAverageAuditByProjectNotionID(db *gorm.DB, projectID string) ([]*model.AverageAudit, error) {
	var rs []*model.AverageAudit

	query := `
		SELECT audit_cycles.quarter,
			avg(CASE
					When audit_cycles.average_score = 0 THEN null
					ELSE audit_cycles.average_score
				END) AS avg
		FROM audit_cycles
		WHERE audit_cycles.project_id = ? AND audit_cycles.deleted_at IS NULL
		GROUP BY audit_cycles.quarter
		ORDER BY audit_cycles.quarter DESC
		LIMIT 4
	`

	return rs, db.Raw(query, projectID).Scan(&rs).Error
}

func (s *store) GetGroupAuditByProjectNotionID(db *gorm.DB, projectID string) ([]*model.GroupAudit, error) {
	var rs []*model.GroupAudit

	query := `
		SELECT audit_cycles.quarter,
			avg(CASE
					When asy.score = 0 THEN null
					ELSE asy.score
				END) as system,
			avg(CASE
					When afe.score = 0 THEN null
					ELSE afe.score
				END) as frontend,
			avg(CASE
					When abe.score = 0 THEN null
					ELSE abe.score
				END) as backend,
			avg(CASE
					When ap.score = 0 THEN null
					ELSE ap.score
				END) as process,
			avg(CASE
					When ab.score = 0 THEN null
					ELSE ab.score
				END) as blockchain,
			avg(CASE
					When am.score = 0 THEN null
					ELSE am.score
				END) as mobile
		FROM audit_cycles
				LEFT JOIN audits AS afe ON audit_cycles.frontend_audit_id = afe.id
				LEFT JOIN audits AS abe ON audit_cycles.backend_audit_id = abe.id
				LEFT JOIN audits AS ap ON audit_cycles.process_audit_id = ap.id
				LEFT JOIN audits AS asy ON audit_cycles.system_audit_id = asy.id
				LEFT JOIN audits AS ab ON audit_cycles.blockchain_audit_id = ab.id
				LEFT JOIN audits AS am ON audit_cycles.mobile_audit_id = am.id
		WHERE audit_cycles.project_id = ? AND audit_cycles.deleted_at IS NULL AND afe.deleted_at IS NULL AND
			abe.deleted_at IS NULL AND ap.deleted_at IS NULL AND
			asy.deleted_at IS NULL AND ab.deleted_at IS NULL AND am.deleted_at IS NULL
		GROUP BY audit_cycles.quarter
		ORDER BY audit_cycles.quarter DESC
		LIMIT 4
	`

	return rs, db.Raw(query, projectID).Scan(&rs).Error
}

func (s *store) GetAllActionItemSquashReports(db *gorm.DB) ([]*model.ActionItemSquashReport, error) {
	var rs []*model.ActionItemSquashReport

	query := `
		WITH days AS (SELECT date_trunc('DAY', end_date) AS day
					FROM feedback_events
					WHERE subtype = 'work'
					GROUP BY date_trunc('DAY', end_date))
		SELECT (sum(high) + sum(medium) + sum(low)) as all,
			sum(high)                            as high,
			sum(medium)                          as medium,
			sum(low)                             as low,
			date_trunc('DAY', action_item_snapshots.created_at)     as snap_date
		FROM action_item_snapshots
				JOIN project_notions ON action_item_snapshots.project_id = project_notions.audit_notion_id
				JOIN projects ON project_notions.project_id = projects.id
				JOIN organizations ON projects.organization_id = organizations.id
				JOIN days ON date_trunc('DAY', action_item_snapshots.created_at) = date_trunc('DAY', days.day)
		WHERE projects.function = 'development'
			AND organizations.code = 'dwarves-foundation'
			AND action_item_snapshots.deleted_at IS NULL
			AND projects.deleted_at IS NULL
		GROUP BY snap_date
		ORDER BY snap_date desc
		LIMIT 12
	`

	return rs, db.Raw(query).Scan(&rs).Error
}

func (s *store) GetActionItemSquashReportsByProjectID(db *gorm.DB, projectID string) ([]*model.ActionItemSquashReport, error) {
	var rs []*model.ActionItemSquashReport

	query := `
		WITH days AS (SELECT date_trunc('DAY', end_date) AS day
					FROM feedback_events
					WHERE subtype = 'work'
					GROUP BY date_trunc('DAY', end_date))
		SELECT (sum(high) + sum(medium) + sum(low)) as all,
			sum(high)                            					as high,
			sum(medium)                          					as medium,
			sum(low)                             					as low,
			date_trunc('DAY', action_item_snapshots.created_at )    as snap_date
		FROM action_item_snapshots
				LEFT JOIN project_notions ON action_item_snapshots.project_id = project_notions.audit_notion_id
				JOIN projects ON project_notions.project_id = projects.id
				JOIN organizations ON projects.organization_id = organizations.id
				JOIN days ON date_trunc('DAY', action_item_snapshots.created_at) = date_trunc('DAY', days.day)
		WHERE projects.function = 'development'
			AND organizations.code = 'dwarves-foundation'
			AND action_item_snapshots.deleted_at IS NULL
			AND projects.deleted_at IS NULL
			AND projects.id = ?
		GROUP BY snap_date
		ORDER BY snap_date desc
		LIMIT 12
	`

	return rs, db.Raw(query, projectID).Scan(&rs).Error
}

func (s *store) GetAuditSummaries(db *gorm.DB) ([]*model.AuditSummary, error) {
	var rs []*model.AuditSummary

	query := `
		SELECT audit_cycles.quarter, p.id, p.name, p.code, p.avatar, SUM(audit_cycles.action_item_high) AS high, SUM(audit_cycles.action_item_medium) AS medium, SUM(audit_cycles.action_item_low) AS low,
			(SELECT count(*) FROM action_items WHERE status = 'done' AND audit_cycle_id = audit_cycles.id) as done,
			count(DISTINCT pm.id) as size,
			avg(CASE
					When a.score = 0 THEN NULL
					ELSE a.score
				END) as health,
			avg(CASE
					When audit_cycles.average_score = 0 THEN null
					ELSE audit_cycles.average_score
				END) AS audit
		FROM audit_cycles
				LEFT JOIN audits AS a ON audit_cycles.health_audit_id = a.id
				JOIN project_notions ON audit_cycles.project_id = project_notions.audit_notion_id
				JOIN projects AS p ON project_notions.project_id = p.id
				JOIN organizations ON p.organization_id = organizations.id
				JOIN project_members pm ON p.id = pm.project_id
		WHERE (pm.status = 'active' OR pm.status='on-boarding') AND audit_cycles.deleted_at IS NULL AND 
			a.deleted_at IS NULL AND pm.deleted_at IS NULL AND p.deleted_at IS NULL AND organizations.code = 'dwarves-foundation'
		GROUP BY audit_cycles.id,audit_cycles.quarter, p.id
		ORDER BY audit_cycles.quarter DESC`

	return rs, db.Raw(query).Scan(&rs).Error
}

func (s *store) GetProjectSizesByStartTime(db *gorm.DB, curr time.Time) ([]*model.ProjectSize, error) {
	var ru []*model.ProjectSize

	query := `
		SELECT projects.id, projects.name, projects.code, count(*) AS size
			FROM (projects 
				JOIN project_members pm ON projects.id = pm.project_id
				JOIN organizations ON projects.organization_id = organizations.id)
			WHERE projects.function = 'development' AND 
				organizations.code = 'dwarves-foundation' AND 
				pm.status = 'active' AND 
				projects.deleted_at IS NULL AND pm.deleted_at IS NULL AND
				(pm.status = 'active' OR pm.status='on-boarding') AND pm.start_date <= ?
			GROUP BY projects.id
	`

	return ru, db.Raw(query, curr.UTC()).Scan(&ru).Error
}

func (s *store) GetPendingSlots(db *gorm.DB) ([]*model.ProjectSlot, error) {
	var slots []*model.ProjectSlot
	return slots, db.
		Where(`id NOT IN (
			SELECT pm.project_slot_id
			FROM project_members pm
			WHERE (pm.end_date IS NULL OR pm.end_date > now()) AND pm.deleted_at IS NULL
		) AND project_id NOT IN (
			SELECT p.id
			FROM projects p
			WHERE p.status = ? AND p.deleted_at IS NULL
		) AND status = ?`, model.ProjectStatusClosed, model.ProjectMemberStatusPending).
		Order("updated_at").
		Preload("Seniority", "deleted_at IS NULL").
		Preload("Project", "deleted_at IS NULL").
		Preload("ProjectSlotPositions", "deleted_at IS NULL").
		Preload("ProjectSlotPositions.Position", "deleted_at IS NULL").
		Find(&slots).Error
}

func (s *store) GetAvailableEmployees(db *gorm.DB) ([]*model.Employee, error) {
	var employees []*model.Employee
	return employees, db.
		Where("working_status != ? ", model.WorkingStatusLeft).
		Where(`id IN (
			SELECT eo.employee_id
			FROM employee_organizations eo JOIN organizations o ON eo.organization_id = o.id
			WHERE o.deleted_at IS NULL AND eo.deleted_at IS NULL AND o.code = ?
		)`, model.OrganizationCodeDwarves).
		Where(`id NOT IN (
			SELECT pm.employee_id
			FROM project_members pm JOIN projects p ON pm.project_id = p.id
			WHERE p.type <> ?
				AND pm.deployment_type = ?
				AND p.status IN ?
				AND (pm.end_date IS NULL OR pm.end_date > now() + INTERVAL '2 months') 
				AND p.deleted_at IS NULL
				AND pm.deleted_at IS NULL
		)`,
			model.ProjectTypeDwarves,
			model.MemberDeploymentTypeOfficial,
			[]string{
				model.ProjectStatusOnBoarding.String(),
				model.ProjectStatusActive.String(),
			}).
		Where(`id IN (
			SELECT e2.id 
			FROM employees e2
				LEFT JOIN employee_chapters ec ON ec.employee_id = e2.id
				LEFT JOIN chapters c ON ec.chapter_id = c.id 
			WHERE ec.deleted_at IS NULL
				AND c.deleted_at IS NULL
				AND e2.deleted_at IS NULL
				AND (c.code IS NULL OR (c.code <> 'sales' AND c.code <> 'operations'))
		)`).
		Order("created_at, display_name").
		Preload("Seniority", "deleted_at IS NULL").
		Preload("EmployeePositions", "deleted_at IS NULL").
		Preload("EmployeePositions.Position", "deleted_at IS NULL").
		Preload("EmployeeStacks", "deleted_at IS NULL").
		Preload("EmployeeStacks.Stack", "deleted_at IS NULL").
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
		Find(&employees).Error
}

func (s *store) GetResourceUtilization(db *gorm.DB, currentDate time.Time) ([]*model.ResourceUtilization, error) {
	var ru []*model.ResourceUtilization

	query := `
	WITH resource_utilization AS (
		SELECT d.d AS "date",
				e.id AS employee_id,
				p.id AS project_id,
				p."type" AS "project_type",
				pm.deployment_type,
				pm.start_date,
				pm.end_date AS pm_end_date,
				p.end_date AS p_end_date,
				e.left_date
		FROM employees e
			LEFT JOIN project_members pm ON pm.employee_id = e.id
			LEFT JOIN projects p ON pm.project_id = p.id,
			generate_series(
				date_trunc('month', TO_DATE(?, 'YYYY-MM-DD')) - INTERVAL '3 month',
				date_trunc('month', TO_DATE(?, 'YYYY-MM-DD')) + INTERVAL '3 month',
				'1 month'
			) d
		WHERE e.deleted_at IS NULL
			AND pm.deleted_at IS NULL
			AND p.deleted_at IS NULL
			AND e.id IN (
				SELECT e2.id
				FROM employees e2
					LEFT JOIN employee_chapters ec ON ec.employee_id = e2.id
					LEFT JOIN chapters c ON ec.chapter_id = c.id
				WHERE ec.deleted_at IS NULL
					AND c.deleted_at IS NULL
					AND e2.deleted_at IS NULL
					AND (c.code IS NULL OR (c.code <> 'sales' AND c.code <> 'operations'))
			)
			AND e.id IN (
				SELECT eo.employee_id 
				FROM employee_organizations eo JOIN organizations o ON eo.organization_id = o.id 
				WHERE o.code = ? 
					AND eo.deleted_at IS NULL 
					AND o.deleted_at IS NULL 
			)
		ORDER BY d.d
	)

	SELECT "date" ,
		COUNT(DISTINCT(employee_id)) FILTER (
			WHERE deployment_type = 'official'
				AND "project_type" != 'dwarves'
				AND start_date <= "date"
				AND (pm_end_date IS NULL OR pm_end_date > "date")
				AND (p_end_date IS NULL OR p_end_date > "date")
				AND (left_date IS NULL OR left_date > "date")
		) AS staffed,

		COUNT(DISTINCT(employee_id)) FILTER (
			WHERE start_date <= "date"
				AND (pm_end_date IS NULL OR pm_end_date > "date")
				AND employee_id NOT IN (
					SELECT ru2.employee_id
					FROM resource_utilization ru2
					WHERE ru2.deployment_type = 'official'
						AND ru2."project_type" != 'dwarves'
						AND ru2.start_date <= ru."date"
						AND (ru2.pm_end_date IS NULL OR ru2.pm_end_date > ru."date")
						AND (ru2.p_end_date IS NULL OR ru2.p_end_date > ru."date")
						AND (ru2.left_date IS NULL OR ru2.left_date > ru."date")
				)
				AND (left_date IS NULL OR left_date > "date")
		) AS internal,

		COUNT(DISTINCT(employee_id)) FILTER (
			WHERE (project_id IS NULL
				OR employee_id NOT IN (
					SELECT ru3.employee_id
					FROM resource_utilization ru3
					WHERE ru3.start_date <= ru."date"
						AND (ru3.pm_end_date IS NULL OR ru3.pm_end_date > ru."date")
						AND (ru3.p_end_date IS NULL OR ru3.p_end_date > ru."date")
						AND (ru3.left_date IS NULL OR ru3.left_date > ru."date")
				))
				AND (left_date IS NULL OR left_date > "date")
		) AS available
	FROM resource_utilization ru
	GROUP BY "date"
	`

	return ru, db.Raw(query, currentDate, currentDate, model.OrganizationCodeDwarves).Scan(&ru).Error
}

func (s *store) TotalWorkUnitDistribution(db *gorm.DB) (*model.TotalWorkUnitDistribution, error) {
	var rs *model.TotalWorkUnitDistribution

	query := `
		WITH
		employee_data AS (
			SELECT
				e.id,
				e.full_name,
				e.display_name,
				e.username,
				e.avatar,
				e.line_manager_id,
				o.code as organization_code
			FROM employees e
				LEFT JOIN employee_organizations eo on e.id = eo.employee_id
				LEFT JOIN organizations o on eo.organization_id = o.id
			WHERE
				e.deleted_at IS NULL AND o.code = 'dwarves-foundation'
		),
		work_unit_info AS (
			SELECT
				wum.employee_id,
				wu.type,
				COUNT(*) AS work_unit_count
			FROM work_units wu
					JOIN work_unit_members wum ON wu.id = wum.work_unit_id
			WHERE
				wum.status = 'active'
				AND wu.status = 'active'
				AND wum.deleted_at IS NULL
				AND wu.deleted_at IS NULL
			GROUP BY wum.employee_id, wu.type),
			count_work_unit_distribution AS (
				SELECT
					e.id,
					e.full_name,
					e.display_name,
					e.username,
					e.avatar,
					(SELECT COUNT(*)
					FROM employee_data
					WHERE line_manager_id = e.id)                   AS line_manager_count,
					(SELECT COUNT(*)
					FROM project_heads ph
					WHERE (ph.position = 'technical-lead' OR
							ph.position = 'delivery-manager' OR
							ph.position = 'account-manager')
						AND ph.employee_id = e.id
						AND ph.end_date IS NULL)                 AS project_head_count,
					COALESCE(wui_learning.work_unit_count, 0)    AS learning,
					COALESCE(wui_development.work_unit_count, 0) AS development,
					COALESCE(wui_management.work_unit_count, 0)  AS management,
					COALESCE(wui_training.work_unit_count, 0)    AS training
				FROM
					employee_data e
					LEFT JOIN work_unit_info wui_learning
								ON e.id = wui_learning.employee_id AND wui_learning.type = 'learning'
					LEFT JOIN work_unit_info wui_development
								ON e.id = wui_development.employee_id AND
									wui_development.type = 'development'
					LEFT JOIN work_unit_info wui_management
								ON e.id = wui_management.employee_id AND wui_management.type = 'management'
					LEFT JOIN work_unit_info wui_training
								ON e.id = wui_training.employee_id AND wui_training.type = 'training')
		SELECT
			SUM(line_manager_count) AS total_line_manager,
			SUM(project_head_count) AS total_project_head,
			SUM(learning)           AS total_learning,
			SUM(development)        AS total_development,
			SUM(management)         AS total_management,
			SUM(training)           AS total_training
		FROM count_work_unit_distribution
	`
	return rs, db.Raw(query).Scan(&rs).Error
}

func (s *store) GetAllWorkReviews(db *gorm.DB, keyword string, pagination model.Pagination) ([]*model.EmployeeEventReviewer, error) {
	var eer []*model.EmployeeEventReviewer

	query := db.Table("employee_event_reviewers").
		Joins(`JOIN feedback_events fe ON employee_event_reviewers.event_id = fe.id 
			AND fe.deleted_at IS NULL
			AND fe.subtype = ?
			AND fe.start_date > now() - INTERVAL '70 DAYS'`, model.EventSubtypeWork). // last 10 weeks
		Joins("JOIN employee_event_topics eet ON employee_event_reviewers.employee_event_topic_id = eet.id").
		Joins("JOIN employees e ON employee_event_reviewers.reviewer_id = e.id AND e.deleted_at IS NULL")

	limit, offset := pagination.ToLimitOffset()

	if keyword != "" {
		query = query.Where("e.keyword_vector @@ plainto_tsquery('english_nostop', fn_remove_vietnamese_accents(LOWER(?)))", keyword)
	}

	query = query.Where("employee_event_reviewers.reviewer_status = ?", model.EventReviewerStatusDone)

	return eer, query.Order("fe.end_date, employee_event_reviewers.reviewer_id, eet.project_id").
		Preload("Event", "deleted_at IS NULL").
		Preload("Reviewer", "deleted_at IS NULL").
		Preload("EmployeeEventQuestions", "deleted_at IS NULL").
		Preload("EmployeeEventTopic", "deleted_at IS NULL").
		Preload("EmployeeEventTopic.Project", "deleted_at IS NULL").
		Limit(limit).
		Offset(offset).
		Find(&eer).Error
}

func (s *store) GetProjectHeadByEmployeeID(db *gorm.DB, employeeID string) ([]*model.ManagementInfo, error) {
	var rs []*model.ManagementInfo

	query := `
		SELECT projects.id, projects.name, projects.code, projects.type, projects.status, projects.avatar, ph.position
		FROM projects
				JOIN project_heads AS ph ON projects.id = ph.project_id
		WHERE ph.employee_id = ? AND projects.status IN (?, ?)
	`

	return rs, db.Raw(query, employeeID, model.ProjectStatusActive, model.ProjectStatusOnBoarding).Scan(&rs).Error
}

func (s *store) GetWorkUnitDistributionEmployees(db *gorm.DB, keyword string, workUnitType string) ([]*model.Employee, error) {
	var employees []*model.Employee

	query := db.Table("employees")

	if keyword != "" {
		query = query.Where("keyword_vector @@ plainto_tsquery('english_nostop', fn_remove_vietnamese_accents(LOWER(?)))", keyword)
	}

	query = query.Where("deleted_at IS NULL").
		Where("working_status IN ?", []string{
			model.WorkingStatusFullTime.String(),
			model.WorkingStatusContractor.String(),
		}).
		Order("display_name ASC")

	// preload mentees
	if workUnitType == "" || workUnitType == model.WorkUnitTypeTraining.String() {
		query = query.Preload("Mentees", "deleted_at IS NULL AND working_status IN ?", []string{
			model.WorkingStatusFullTime.String(),
			model.WorkingStatusContractor.String(),
		})
	}

	// preload project heads
	if workUnitType == "" || workUnitType == model.WorkUnitTypeManagement.String() {
		query = query.Preload("Heads", func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN projects p ON project_heads.project_id = p.id").
				Where("(project_heads.start_date IS NULL OR project_heads.start_date <= now())").
				Where("(project_heads.end_date IS NULL OR project_heads.end_date > now())").
				Where("p.status IN ?", []string{
					model.ProjectStatusActive.String(),
					model.ProjectStatusOnBoarding.String(),
				}).
				Where("project_heads.deleted_at IS NULL").
				Where("p.deleted_at IS NULL")
		}).Preload("Heads.Project", "deleted_at IS NULL")
	}

	// preload work units
	query = query.Preload("WorkUnitMembers", func(db *gorm.DB) *gorm.DB {
		db = db.
			Joins("JOIN work_units wu ON wu.id = work_unit_members.work_unit_id").
			Joins("JOIN projects p ON p.id = wu.project_id").
			Joins("JOIN project_members pm ON pm.project_id = work_unit_members.project_id AND pm.employee_id = work_unit_members.employee_id").
			Where("wu.status = ?", model.WorkUnitStatusActive).
			Where("p.status IN ?", []string{
				model.ProjectStatusActive.String(),
				model.ProjectStatusOnBoarding.String(),
			}).
			Where("pm.start_date <= now() AND (pm.end_date IS NULL OR pm.end_date > now())").
			Where("p.deleted_at IS NULL").
			Where("pm.deleted_at IS NULL").
			Where("wu.deleted_at IS NULL").
			Where("work_unit_members.deleted_at IS NULL")

		if workUnitType != "" {
			db = db.Where("wu.type = ?", workUnitType)
		}

		return db
	}).
		Preload("WorkUnitMembers.WorkUnit").
		Preload("WorkUnitMembers.WorkUnit.Project")

	return employees, query.Find(&employees).Error
}
