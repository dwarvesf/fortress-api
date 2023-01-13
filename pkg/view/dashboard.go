package view

import (
	"fmt"
	"strings"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type ProjectSizeResponse struct {
	Data []*model.ProjectSize `json:"data"`
}

type WorkSurveyResponse struct {
	Data WorkSurveysData `json:"data"`
}

type Trend struct {
	Workload float64 `json:"workload"`
	Deadline float64 `json:"deadline"`
	Learning float64 `json:"learning"`
}

type WorkSurvey struct {
	EndDate  string  `json:"endDate"`
	Workload float64 `json:"workload"`
	Deadline float64 `json:"deadline"`
	Learning float64 `json:"learning"`
	Trend    *Trend  `json:"trend"`
}

type ActionItemTrend struct {
	High   float64 `json:"high"`
	Medium float64 `json:"medium"`
	Low    float64 `json:"low"`
}

type AuditActionItemReport struct {
	Quarter string           `json:"quarter"`
	High    int64            `json:"high"`
	Medium  int64            `json:"medium"`
	Low     int64            `json:"low"`
	Trend   *ActionItemTrend `json:"trend"`
}

type WorkSurveysData struct {
	Project     *BasicProjectInfo `json:"project"`
	WorkSurveys []*WorkSurvey     `json:"workSurveys"`
}

type ActionItemReportData struct {
	AuditActionItemReports []*AuditActionItemReport `json:"auditActionItemReports"`
}

func ToWorkSurveyData(project *model.Project, workSurveys []*model.WorkSurvey) *WorkSurveysData {
	rs := &WorkSurveysData{}

	for _, ws := range workSurveys {
		rs.WorkSurveys = append(rs.WorkSurveys, &WorkSurvey{
			EndDate:  ws.EndDate.Format("02/01"),
			Workload: ws.Workload,
			Deadline: ws.Deadline,
			Learning: ws.Learning,
		})
	}

	if project != nil {
		rs.Project = toBasicProjectInfo(*project)
	}

	if workSurveys != nil && len(workSurveys) > 1 {
		for i := 1; i < len(workSurveys); i++ {
			rs.WorkSurveys[i].Trend = calculateTrend(workSurveys[i-1], workSurveys[i])
		}
	}

	return rs
}

func ToActionItemReportData(actionItemReports []*model.ActionItemReport) *ActionItemReportData {
	rs := &ActionItemReportData{}
	// reverse the order to correct timeline
	for i, j := 0, len(actionItemReports)-1; i < j; i, j = i+1, j-1 {
		actionItemReports[i], actionItemReports[j] = actionItemReports[j], actionItemReports[i]
	}
	for _, ws := range actionItemReports {
		rs.AuditActionItemReports = append(rs.AuditActionItemReports, &AuditActionItemReport{
			Quarter: strings.Split(ws.Quarter, "/")[1] + "/" + strings.Split(ws.Quarter, "/")[0],
			High:    ws.High,
			Medium:  ws.Low,
			Low:     ws.Low,
		})
	}

	if actionItemReports != nil && len(actionItemReports) > 1 {
		for i := 1; i < len(actionItemReports); i++ {
			rs.AuditActionItemReports[i].Trend = calculateActionItemReportTrend(actionItemReports[i-1], actionItemReports[i])
			fmt.Print("----quarter: " + rs.AuditActionItemReports[i].Quarter)
		}
	}

	return rs
}

// calculateTrend calculate the trend for work survey
func calculateTrend(previous *model.WorkSurvey, current *model.WorkSurvey) *Trend {
	rs := &Trend{}

	// if previous or current value = 0 trend = 0
	if previous.Workload == 0 || current.Workload == 0 {
		rs.Workload = 0
	} else {
		rs.Workload = (current.Workload - previous.Workload) / previous.Workload * 100
	}

	if previous.Deadline == 0 || current.Deadline == 0 {
		rs.Deadline = 0
	} else {
		rs.Deadline = (current.Deadline - previous.Deadline) / previous.Deadline * 100
	}

	if previous.Learning == 0 || current.Learning == 0 {
		rs.Learning = 0
	} else {
		rs.Learning = (current.Learning - previous.Learning) / previous.Learning * 100
	}

	return rs
}

// calculateTrend calculate the trend for action item report
func calculateActionItemReportTrend(previous *model.ActionItemReport, current *model.ActionItemReport) *ActionItemTrend {
	rs := &ActionItemTrend{}

	// if previous or current value = 0 trend = 0
	if previous.High == 0 || current.High == 0 {
		rs.High = 0
	} else {
		rs.High = float64(float64(current.High-previous.High) / float64(previous.High) * 100)
	}

	if previous.Medium == 0 || current.Medium == 0 {
		rs.Medium = 0
	} else {
		rs.Medium = float64(float64(current.Medium-previous.Medium) / float64(previous.Medium) * 100)
	}

	if previous.Low == 0 || current.Low == 0 {
		rs.Low = 0
	} else {
		rs.Low = float64(float64(current.Low-previous.Low) / float64(previous.Low) * 100)
	}

	return rs
}
