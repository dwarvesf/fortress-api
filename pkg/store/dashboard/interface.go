package dashboard

import (
	"time"

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
	AverageEngineeringHealthByProjectNotionID(db *gorm.DB, projectID string) ([]*model.AverageEngineeringHealth, error)
	GroupEngineeringHealthByProjectNotionID(db *gorm.DB, projectID string) ([]*model.GroupEngineeringHealth, error)
	GetAverageAudit(db *gorm.DB) ([]*model.AverageAudit, error)
	GetGroupAudit(db *gorm.DB) ([]*model.GroupAudit, error)
	GetActionItemSquashReportsByProjectID(db *gorm.DB, projectID string) ([]*model.ActionItemSquashReport, error)
	GetAllActionItemSquashReports(db *gorm.DB) ([]*model.ActionItemSquashReport, error)
	GetAverageAuditByProjectNotionID(db *gorm.DB, projectID string) ([]*model.AverageAudit, error)
	GetGroupAuditByProjectNotionID(db *gorm.DB, projectID string) ([]*model.GroupAudit, error)
	GetAuditSummaries(db *gorm.DB) ([]*model.AuditSummary, error)
	GetProjectSizesByStartTime(db *gorm.DB, curr time.Time) ([]*model.ProjectSize, error)
	GetPendingSlots(db *gorm.DB) ([]*model.ProjectSlot, error)
	GetAvailableEmployees(db *gorm.DB) ([]*model.Employee, error)
	GetResourceUtilization(db *gorm.DB) ([]*model.ResourceUtilization, error)
	GetWorkUnitDistribution(db *gorm.DB, name string) ([]*model.WorkUnitDistribution, error)
}
