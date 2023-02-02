package dashboard

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) GetProjectSizes(db *gorm.DB) ([]*model.ProjectSize, error) {
	var ru []*model.ProjectSize

	query := `
		SELECT projects.id, projects.name, projects.code, count(*) AS size
			FROM (projects join project_members pm ON projects.id = pm.project_id)
			WHERE projects.function = 'development' AND pm.status = 'active' AND (pm.status = 'active' OR pm.status='on-boarding') AND projects.deleted_at IS NULL
			GROUP BY projects.id
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
				LEFT JOIN employee_event_topics eet ON feedback_events.id = eet.event_id
				LEFT JOIN employee_event_questions eeq ON feedback_events.id = eeq.event_id
		WHERE eet.project_id = ? AND feedback_events.deleted_at IS NULL AND eet.deleted_at IS NULL AND eeq.deleted_at IS NULL
		GROUP BY eeq.event_id, feedback_events.id
		ORDER BY feedback_events.end_date desc
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
			LEFT JOIN employee_event_topics eet ON feedback_events.id = eet.event_id
			JOIN projects p ON eet.project_id = p.id
			LEFT JOIN employee_event_questions eeq ON feedback_events.id = eeq.event_id
		WHERE p.function = 'development' AND feedback_events.deleted_at IS NULL AND eet.deleted_at IS NULL AND eeq.deleted_at IS NULL
		GROUP BY feedback_events.end_date
		ORDER BY feedback_events.end_date desc
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

func (s *store) GetActionItemReportsByProjectID(db *gorm.DB, projectID string) ([]*model.ActionItemReport, error) {
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
		WITH days AS (SELECT date_trunc('DAY', start_date) AS day
					FROM feedback_events
					WHERE subtype = 'work'
					GROUP BY date_trunc('DAY', start_date))
		SELECT (sum(high) + sum(medium) + sum(low)) as all,
			sum(high)                            as high,
			sum(medium)                          as medium,
			sum(low)                             as low,
			action_item_snapshots.created_at     as snap_date
		FROM action_item_snapshots
				LEFT JOIN projects on action_item_snapshots.project_id = projects.notion_id
				JOIN days ON date_trunc('DAY', action_item_snapshots.created_at) = date_trunc('DAY', days.day)
		WHERE projects.function = 'development'
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
		WITH days AS (SELECT date_trunc('DAY', start_date) AS day
					FROM feedback_events
					WHERE subtype = 'work'
					GROUP BY date_trunc('DAY', start_date))
		SELECT (sum(high) + sum(medium) + sum(low)) as all,
			sum(high)                            as high,
			sum(medium)                          as medium,
			sum(low)                             as low,
			action_item_snapshots.created_at     as snap_date
		FROM action_item_snapshots
				LEFT JOIN projects on action_item_snapshots.project_id = projects.notion_id
				JOIN days ON date_trunc('DAY', action_item_snapshots.created_at) = date_trunc('DAY', days.day)
		WHERE projects.function = 'development'
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
				JOIN projects AS p ON audit_cycles.project_id = p.notion_id
				JOIN project_members pm ON p.id = pm.project_id
		WHERE (pm.status = 'active' OR pm.status='on-boarding') AND audit_cycles.deleted_at IS NULL AND 
			a.deleted_at IS NULL AND pm.deleted_at IS NULL AND p.deleted_at IS NULL
		GROUP BY audit_cycles.id,audit_cycles.quarter, p.id
		ORDER BY audit_cycles.quarter DESC`

	return rs, db.Raw(query).Scan(&rs).Error
}

func (s *store) GetProjectSizesByStartTime(db *gorm.DB, curr time.Time) ([]*model.ProjectSize, error) {
	var ru []*model.ProjectSize

	query := `
		SELECT projects.id, projects.name, projects.code, count(*) AS size
			FROM (projects JOIN project_members pm ON projects.id = pm.project_id)
			WHERE projects.function = 'development' AND pm.status = 'active' AND 
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
			WHERE pm.end_date IS NOT NULL AND pm.end_date > now()
		)`).
		Preload("Seniority").
		Preload("Project").
		Preload("ProjectSlotPositions").
		Find(&slots).Error
}

func (s *store) GetAvailableEmployees(db *gorm.DB) ([]*model.Employee, error) {
	var employees []*model.Employee
	return employees, db.
		Where(`working_status != ? AND id NOT IN (
			SELECT pm.employee_id
			FROM project_members pm
			WHERE pm.end_date IS NOT NULL AND pm.end_date > now()
		)`, model.WorkingStatusLeft).
		Preload("Seniority").
		Preload("EmployeePositions").
		Preload("EmployeeStacks").
		Preload("ProjectMembers").
		Preload("ProjectMembers.Project").
		Find(&employees).Error
}
