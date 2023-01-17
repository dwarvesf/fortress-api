package dashboard

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	GetProjectSizes(db *gorm.DB) (res []*model.ProjectSize, err error)
	GetWorkSurveysByProjectID(db *gorm.DB, projectID string) ([]*model.WorkSurvey, error)
	GetAllWorkSurveys(db *gorm.DB) ([]*model.WorkSurvey, error)
	GetActionItemReportsByProjectID(db *gorm.DB, projectID string) ([]*model.ActionItemReport, error)
	GetAllActionItemReports(db *gorm.DB) ([]*model.ActionItemReport, error)
	AverageEngineeringHealth(db *gorm.DB) ([]*model.AverageEngineeringHealth, error)
	GroupEngineeringHealth(db *gorm.DB) ([]*model.GroupEngineeringHealth, error)
	AverageEngineeringHealthByProjectID(db *gorm.DB, projectID string) ([]*model.AverageEngineeringHealth, error)
	GroupEngineeringHealthByProjectID(db *gorm.DB, projectID string) ([]*model.GroupEngineeringHealth, error)
	GetAverageAudit(db *gorm.DB) ([]*model.AverageAudit, error)
	GetGroupAudit(db *gorm.DB) ([]*model.GroupAudit, error)
	GetAverageAuditByProjectID(db *gorm.DB, projectID string) ([]*model.AverageAudit, error)
	GetGroupAuditByProjectID(db *gorm.DB, projectID string) ([]*model.GroupAudit, error)
	GetActionItemSquashReportsByProjectID(db *gorm.DB, projectID string) ([]*model.ActionItemSquashReport, error)
	GetAllActionItemSquashReports(db *gorm.DB) ([]*model.ActionItemSquashReport, error)
}
