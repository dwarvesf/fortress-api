package dashboard

import (
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
			WHERE projects.function = 'development' AND pm.status = 'active' 
			GROUP BY projects.id
	`

	return ru, db.Debug().Raw(query).Scan(&ru).Error
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
		WHERE eet.project_id = ?
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
					END) as dealine,
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
		WHERE p.function = 'development'
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
		WHERE audit_cycles.project_id = ?
		GROUP BY quarter
		ORDER BY quarter desc
		LIMIT 4
	`

	return rs, db.Raw(query, projectID).Scan(&rs).Error
}
