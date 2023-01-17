package view

import (
	"math"
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

type ActionItemReportResponse struct {
	AuditActionItemReports []*AuditActionItemReport `json:"data"`
}

type ActionItemSquash struct {
	SnapDate string  `json:"snapDate"`
	Value    int64   `json:"value"`
	Trend    float64 `json:"trend"`
}

type ActionItemSquashReport struct {
	All    []*ActionItemSquash `json:"all"`
	High   []*ActionItemSquash `json:"high"`
	Medium []*ActionItemSquash `json:"medium"`
	Low    []*ActionItemSquash `json:"low"`
}

type ActionItemSquashReportResponse struct {
	Data *ActionItemSquashReport `json:"data"`
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

func ToActionItemReportData(actionItemReports []*model.ActionItemReport) []*AuditActionItemReport {
	var rs []*AuditActionItemReport
	// reverse the order to correct timeline
	for i, j := 0, len(actionItemReports)-1; i < j; i, j = i+1, j-1 {
		actionItemReports[i], actionItemReports[j] = actionItemReports[j], actionItemReports[i]
	}
	for _, ws := range actionItemReports {
		rs = append(rs, &AuditActionItemReport{
			Quarter: strings.Split(ws.Quarter, "/")[1] + "/" + strings.Split(ws.Quarter, "/")[0],
			High:    ws.High,
			Medium:  ws.Low,
			Low:     ws.Low,
		})
	}

	if actionItemReports != nil && len(actionItemReports) > 1 {
		for i := 1; i < len(actionItemReports); i++ {
			rs[i].Trend = calculateActionItemReportTrend(actionItemReports[i-1], actionItemReports[i])
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

type EngineringHealthResponse struct {
	Data EngineeringHealthData `json:"data"`
}

type EngineeringHealthData struct {
	Average []*EngineeringHealth    `json:"average"`
	Groups  *GroupEngineeringHealth `json:"groups"`
}

type EngineeringHealth struct {
	Quarter string  `json:"quarter"`
	Value   float64 `json:"avg"`
	Trend   float64 `json:"trend"`
}

type GroupEngineeringHealth struct {
	Delivery      []*EngineeringHealth `json:"delivery"`
	Quality       []*EngineeringHealth `json:"quality"`
	Collaboration []*EngineeringHealth `json:"collaboration"`
	Feedback      []*EngineeringHealth `json:"feedback"`
}

func ToEngineeringHealthData(average []*model.AverageEngineeringHealth, groups []*model.GroupEngineeringHealth) *EngineeringHealthData {
	rs := &EngineeringHealthData{}

	// Reverse quarter order
	for i, j := 0, len(average)-1; i < j; i, j = i+1, j-1 {
		average[i], average[j] = average[j], average[i]
	}

	for i, j := 0, len(groups)-1; i < j; i, j = i+1, j-1 {
		groups[i], groups[j] = groups[j], groups[i]
	}

	for _, a := range average {
		rs.Average = append(rs.Average, &EngineeringHealth{
			Quarter: strings.Split(a.Quarter, "/")[1] + "/" + strings.Split(a.Quarter, "/")[0],
			Value:   a.Avg,
		})
	}

	calculateTrendForEngineeringHealthList(rs.Average)

	rs.Groups = toGroupEngineeringHealth(groups)

	calculateTrendForEngineeringHealthList(rs.Groups.Delivery)
	calculateTrendForEngineeringHealthList(rs.Groups.Collaboration)
	calculateTrendForEngineeringHealthList(rs.Groups.Quality)
	calculateTrendForEngineeringHealthList(rs.Groups.Feedback)

	return rs
}

func toGroupEngineeringHealth(groups []*model.GroupEngineeringHealth) *GroupEngineeringHealth {
	rs := &GroupEngineeringHealth{}
	count := 0
	quarter := ""

	for _, g := range groups {
		if quarter != g.Quarter {
			count++
			quarter = g.Quarter

			if count > 4 {
				break
			}
		}

		switch g.Area {
		case model.AuditItemAreaDelivery:
			rs.Delivery = append(rs.Delivery, &EngineeringHealth{
				Quarter: strings.Split(g.Quarter, "/")[1] + "/" + strings.Split(g.Quarter, "/")[0],
				Value:   g.Avg,
			})
		case model.AuditItemAreaQuality:
			rs.Quality = append(rs.Quality, &EngineeringHealth{
				Quarter: strings.Split(g.Quarter, "/")[1] + "/" + strings.Split(g.Quarter, "/")[0],
				Value:   g.Avg,
			})
		case model.AuditItemAreaCollaborating:
			rs.Collaboration = append(rs.Collaboration, &EngineeringHealth{
				Quarter: strings.Split(g.Quarter, "/")[1] + "/" + strings.Split(g.Quarter, "/")[0],
				Value:   g.Avg,
			})
		case model.AuditItemAreaFeedback:
			rs.Feedback = append(rs.Feedback, &EngineeringHealth{
				Quarter: strings.Split(g.Quarter, "/")[1] + "/" + strings.Split(g.Quarter, "/")[0],
				Value:   g.Avg,
			})
		}

	}

	return rs
}

func calculateTrendForEngineeringHealthList(healths []*EngineeringHealth) {
	for i := 1; i < len(healths); i++ {
		healths[i].Trend = calculateEngineeringHealthTrend(healths[i-1], healths[i])
	}
}

func calculateEngineeringHealthTrend(previous *EngineeringHealth, current *EngineeringHealth) float64 {
	// if previous or current value = 0 trend = 0
	if previous.Value == 0 || current.Value == 0 {
		return 0
	}

	// return the value fixed 2 decimal places
	return float64(math.Trunc((current.Value-previous.Value)/previous.Value*100*100)) / 100
}

type AuditResponse struct {
	Data AuditData `json:"data"`
}

type AuditData struct {
	Average []*Audit    `json:"average"`
	Groups  *GroupAudit `json:"groups"`
}

type Audit struct {
	Quarter string  `json:"quarter"`
	Value   float64 `json:"avg"`
	Trend   float64 `json:"trend"`
}

type GroupAudit struct {
	Frontend   []*Audit `json:"frontend"`
	Backend    []*Audit `json:"backend"`
	System     []*Audit `json:"system"`
	Process    []*Audit `json:"process"`
	Mobile     []*Audit `json:"mobile"`
	Blockchain []*Audit `json:"blockchain"`
}

func ToAuditData(average []*model.AverageAudit, groups []*model.GroupAudit) *AuditData {
	rs := &AuditData{}

	// Reverse quarter order
	for i, j := 0, len(average)-1; i < j; i, j = i+1, j-1 {
		average[i], average[j] = average[j], average[i]
	}

	for i, j := 0, len(groups)-1; i < j; i, j = i+1, j-1 {
		groups[i], groups[j] = groups[j], groups[i]
	}

	for _, a := range average {
		rs.Average = append(rs.Average, &Audit{
			Quarter: strings.Split(a.Quarter, "/")[1] + "/" + strings.Split(a.Quarter, "/")[0],
			Value:   a.Avg,
		})
	}

	calculateTrendForAuditList(rs.Average)

	rs.Groups = toGroupAudit(groups)

	calculateTrendForAuditList(rs.Groups.Frontend)
	calculateTrendForAuditList(rs.Groups.Backend)
	calculateTrendForAuditList(rs.Groups.Process)
	calculateTrendForAuditList(rs.Groups.System)
	calculateTrendForAuditList(rs.Groups.Mobile)
	calculateTrendForAuditList(rs.Groups.Blockchain)

	return rs
}

func toGroupAudit(groups []*model.GroupAudit) *GroupAudit {
	rs := &GroupAudit{}

	for _, g := range groups {
		rs.Frontend = append(rs.Frontend, &Audit{
			Quarter: strings.Split(g.Quarter, "/")[1] + "/" + strings.Split(g.Quarter, "/")[0],
			Value:   g.Frontend,
		})

		rs.Backend = append(rs.Backend, &Audit{
			Quarter: strings.Split(g.Quarter, "/")[1] + "/" + strings.Split(g.Quarter, "/")[0],
			Value:   g.Backend,
		})

		rs.System = append(rs.System, &Audit{
			Quarter: strings.Split(g.Quarter, "/")[1] + "/" + strings.Split(g.Quarter, "/")[0],
			Value:   g.System,
		})

		rs.Process = append(rs.Process, &Audit{
			Quarter: strings.Split(g.Quarter, "/")[1] + "/" + strings.Split(g.Quarter, "/")[0],
			Value:   g.Process,
		})

		rs.Mobile = append(rs.Mobile, &Audit{
			Quarter: strings.Split(g.Quarter, "/")[1] + "/" + strings.Split(g.Quarter, "/")[0],
			Value:   g.Mobile,
		})

		rs.Blockchain = append(rs.Blockchain, &Audit{
			Quarter: strings.Split(g.Quarter, "/")[1] + "/" + strings.Split(g.Quarter, "/")[0],
			Value:   g.Blockchain,
		})
	}

	return rs
}

func calculateTrendForAuditList(healths []*Audit) {
	for i := 1; i < len(healths); i++ {
		healths[i].Trend = calculateAuditTrend(healths[i-1], healths[i])
	}
}

func calculateAuditTrend(previous *Audit, current *Audit) float64 {
	// if previous or current value = 0 trend = 0
	if previous.Value == 0 || current.Value == 0 {
		return 0
	}

	// return the value fixed 2 decimal places
	return float64(math.Trunc((current.Value-previous.Value)/previous.Value*100*100)) / 100
}

func ToActionItemSquashReportData(actionItemReports []*model.ActionItemSquashReport) *ActionItemSquashReport {
	rs := &ActionItemSquashReport{}
	// reverse the order to correct timeline
	for i, j := 0, len(actionItemReports)-1; i < j; i, j = i+1, j-1 {
		actionItemReports[i], actionItemReports[j] = actionItemReports[j], actionItemReports[i]
	}

	for _, item := range actionItemReports {
		date := item.SnapDate.Format("02/01")
		rs.All = append(rs.All, &ActionItemSquash{
			SnapDate: date,
			Value:    item.All,
		})
		rs.High = append(rs.High, &ActionItemSquash{
			SnapDate: date,
			Value:    item.High,
		})
		rs.Medium = append(rs.Medium, &ActionItemSquash{
			SnapDate: date,
			Value:    item.Medium,
		})
		rs.Low = append(rs.Low, &ActionItemSquash{
			SnapDate: date,
			Value:    item.Low,
		})
	}

	if actionItemReports != nil && len(actionItemReports) > 1 {
		calculateTrendForActionItemSquash(rs.All)
		calculateTrendForActionItemSquash(rs.High)
		calculateTrendForActionItemSquash(rs.Medium)
		calculateTrendForActionItemSquash(rs.Low)
	}

	return rs
}
func calculateTrendForActionItemSquash(items []*ActionItemSquash) {
	for i := 1; i < len(items); i++ {
		if items[i-1].Value == 0 || items[i].Value == 0 {
			items[i].Trend = 0
		}

		items[i].Trend = math.Floor(float64(items[i].Value-items[i-1].Value)/float64(items[i-1].Value)*100*100) / 100
	}
}
