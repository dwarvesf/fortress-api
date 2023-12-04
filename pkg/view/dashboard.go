package view

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"golang.org/x/exp/slices"
)

type EngagementDashboardQuestionStat struct {
	Title     string     `json:"title"`
	StartDate *time.Time `json:"startDate"`
	Point     float64    `json:"point"`
} // @name EngagementDashboardQuestionStat
type EngagementDashboard struct {
	Content    string                            `json:"content"`
	QuestionID string                            `json:"questionID"`
	Stats      []EngagementDashboardQuestionStat `json:"stats"`
} // @name EngagementDashboard

type EngagementDashboardQuestionDetailStat struct {
	Field     string     `json:"field"`
	StartDate *time.Time `json:"startDate"`
	Point     float64    `json:"point"`
} // @name EngagementDashboardQuestionDetailStat

type EngagementDashboardDetail struct {
	QuestionID string                                  `json:"questionID"`
	Stats      []EngagementDashboardQuestionDetailStat `json:"stats"`
} // @name EngagementDashboardDetail

type ProjectSizeResponse struct {
	Data []*ProjectSize `json:"data"`
} // @name ProjectSizeResponse

type ProjectSize struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Avatar string `json:"avatar"`
	Size   int64  `json:"size"`
} // @name ProjectSize

func ToProjectSizes(projects []*model.ProjectSize) []*ProjectSize {
	var rs []*ProjectSize
	for _, p := range projects {
		rs = append(rs, &ProjectSize{
			ID:     p.ID.String(),
			Name:   p.Name,
			Code:   p.Code,
			Avatar: p.Avatar,
			Size:   p.Size,
		})
	}

	return rs
}

type WorkSurveyResponse struct {
	Data WorkSurveysData `json:"data"`
} // @name WorkSurveyResponse

type Trend struct {
	Workload float64 `json:"workload"`
	Deadline float64 `json:"deadline"`
	Learning float64 `json:"learning"`
} // @name Trend

type WorkSurvey struct {
	EndDate  string  `json:"endDate"`
	Workload float64 `json:"workload"`
	Deadline float64 `json:"deadline"`
	Learning float64 `json:"learning"`
	Trend    Trend   `json:"trend"`
} // @name WorkSurvey

type ActionItemTrend struct {
	High   float64 `json:"high"`
	Medium float64 `json:"medium"`
	Low    float64 `json:"low"`
} // @name ActionItemTrend

type AuditActionItemReport struct {
	Quarter string          `json:"quarter"`
	High    int64           `json:"high"`
	Medium  int64           `json:"medium"`
	Low     int64           `json:"low"`
	Trend   ActionItemTrend `json:"trend"`
} // @name AuditActionItemReport

type WorkSurveysData struct {
	Project     *BasicProjectInfo `json:"project"`
	WorkSurveys []*WorkSurvey     `json:"workSurveys"`
} // @name WorkSurveysData

type ActionItemReportResponse struct {
	AuditActionItemReports []*AuditActionItemReport `json:"data"`
} // @name ActionItemReportResponse

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

	if len(workSurveys) > 1 {
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
			Medium:  ws.Medium,
			Low:     ws.Low,
		})
	}

	if len(actionItemReports) > 1 {
		for i := 1; i < len(actionItemReports); i++ {
			rs[i].Trend = calculateActionItemReportTrend(actionItemReports[i-1], actionItemReports[i])
		}
	}

	return rs
}

// calculateTrend calculate the trend for work survey
func calculateTrend(previous *model.WorkSurvey, current *model.WorkSurvey) Trend {
	rs := Trend{}

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
func calculateActionItemReportTrend(previous *model.ActionItemReport, current *model.ActionItemReport) ActionItemTrend {
	rs := ActionItemTrend{}

	// if previous or current value = 0 trend = 0
	if previous.High == 0 || current.High == 0 {
		rs.High = 0
	} else {
		rs.High = float64(current.High-previous.High) / float64(previous.High) * 100
	}

	if previous.Medium == 0 || current.Medium == 0 {
		rs.Medium = 0
	} else {
		rs.Medium = float64(current.Medium-previous.Medium) / float64(previous.Medium) * 100
	}

	if previous.Low == 0 || current.Low == 0 {
		rs.Low = 0
	} else {
		rs.Low = float64(current.Low-previous.Low) / float64(previous.Low) * 100
	}

	return rs
}

type EngineeringHealthResponse struct {
	Data EngineeringHealthData `json:"data"`
} // @name EngineeringHealthResponse

type EngineeringHealthData struct {
	Average []*EngineeringHealth      `json:"average"`
	Groups  []*GroupEngineeringHealth `json:"groups"`
} // @name EngineeringHealthData

type EngineeringHealth struct {
	Quarter string  `json:"quarter"`
	Value   float64 `json:"avg"`
	Trend   float64 `json:"trend"`
} // @name EngineeringHealth

type GroupEngineeringHealth struct {
	Quarter       string                 `json:"quarter"`
	Delivery      float64                `json:"delivery"`
	Quality       float64                `json:"quality"`
	Collaboration float64                `json:"collaboration"`
	Feedback      float64                `json:"feedback"`
	Trend         EngineeringHealthTrend `json:"trend"`
} // @name GroupEngineeringHealth

type EngineeringHealthTrend struct {
	Delivery      float64 `json:"delivery"`
	Quality       float64 `json:"quality"`
	Collaboration float64 `json:"collaboration"`
	Feedback      float64 `json:"feedback"`
} // @name EngineeringHealthTrend

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
	calculateEngineeringHealthGroupTrend(rs.Groups)

	return rs
}

func toGroupEngineeringHealth(groups []*model.GroupEngineeringHealth) []*GroupEngineeringHealth {
	var rs []*GroupEngineeringHealth
	count := 0
	quarter := ""
	i := 0

	for i < len(groups) {
		if quarter != groups[i].Quarter {
			count++
			quarter = groups[i].Quarter

			if count > 4 {
				break
			}
		}

		rs = append(rs, &GroupEngineeringHealth{
			Quarter: strings.Split(groups[i].Quarter, "/")[1] + "/" + strings.Split(groups[i].Quarter, "/")[0],
		})

		for quarter == groups[i].Quarter {
			switch groups[i].Area {
			case model.AuditItemAreaDelivery:
				rs[count-1].Delivery = groups[i].Avg
			case model.AuditItemAreaQuality:
				rs[count-1].Quality = groups[i].Avg
			case model.AuditItemAreaCollaborating:
				rs[count-1].Collaboration = groups[i].Avg
			case model.AuditItemAreaFeedback:
				rs[count-1].Feedback = groups[i].Avg
			}

			i++
			if i >= len(groups) {
				break
			}
		}
	}

	return rs
}

func calculateTrendForEngineeringHealthList(healths []*EngineeringHealth) {
	for i := 1; i < len(healths); i++ {
		healths[i].Trend = calculateEngineeringHealthTrend(healths[i-1], healths[i])
	}
}

func calculateEngineeringHealthGroupTrend(groups []*GroupEngineeringHealth) {
	for i := 1; i < len(groups); i++ {
		groups[i].Trend.Delivery = calculateEngineeringHealthTrend(&EngineeringHealth{Value: groups[i-1].Delivery}, &EngineeringHealth{Value: groups[i].Delivery})
		groups[i].Trend.Quality = calculateEngineeringHealthTrend(&EngineeringHealth{Value: groups[i-1].Quality}, &EngineeringHealth{Value: groups[i].Quality})
		groups[i].Trend.Collaboration = calculateEngineeringHealthTrend(&EngineeringHealth{Value: groups[i-1].Collaboration}, &EngineeringHealth{Value: groups[i].Collaboration})
		groups[i].Trend.Feedback = calculateEngineeringHealthTrend(&EngineeringHealth{Value: groups[i-1].Feedback}, &EngineeringHealth{Value: groups[i].Feedback})
	}
}

func calculateEngineeringHealthTrend(previous *EngineeringHealth, current *EngineeringHealth) float64 {
	// if previous or current value = 0 trend = 0
	if previous.Value == 0 || current.Value == 0 {
		return 0
	}

	// return the value fixed 2 decimal places
	return math.Round((current.Value-previous.Value)/previous.Value*100*100) / 100
}

type AuditResponse struct {
	Data AuditData `json:"data"`
} // @name AuditResponse

type AuditData struct {
	Average []*Audit      `json:"average"`
	Groups  []*GroupAudit `json:"groups"`
} // @name AuditData

type Audit struct {
	Quarter string  `json:"quarter"`
	Value   float64 `json:"avg"`
	Trend   float64 `json:"trend"`
} // @name Audit

type GroupAudit struct {
	Quarter    string          `json:"quarter"`
	Frontend   float64         `json:"frontend"`
	Backend    float64         `json:"backend"`
	System     float64         `json:"system"`
	Process    float64         `json:"process"`
	Mobile     float64         `json:"mobile"`
	Blockchain float64         `json:"blockchain"`
	Trend      GroupAuditTrend `json:"trend"`
} // @name GroupAudit

type GroupAuditTrend struct {
	Frontend   float64 `json:"frontend"`
	Backend    float64 `json:"backend"`
	System     float64 `json:"system"`
	Process    float64 `json:"process"`
	Mobile     float64 `json:"mobile"`
	Blockchain float64 `json:"blockchain"`
} // @name GroupAuditTrend

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
	calculateAuditGroupTrend(rs.Groups)

	return rs
}

func toGroupAudit(groups []*model.GroupAudit) []*GroupAudit {
	var rs []*GroupAudit

	for i := range groups {
		rs = append(rs, &GroupAudit{
			Quarter:    strings.Split(groups[i].Quarter, "/")[1] + "/" + strings.Split(groups[i].Quarter, "/")[0],
			Frontend:   groups[i].Frontend,
			Backend:    groups[i].Backend,
			System:     groups[i].System,
			Process:    groups[i].Process,
			Mobile:     groups[i].Mobile,
			Blockchain: groups[i].Blockchain,
		})
	}

	return rs
}

func calculateTrendForAuditList(healths []*Audit) {
	for i := 1; i < len(healths); i++ {
		healths[i].Trend = calculateAuditTrend(healths[i-1], healths[i])
	}
}

func calculateAuditGroupTrend(groups []*GroupAudit) {
	for i := 1; i < len(groups); i++ {
		groups[i].Trend.Frontend = calculateAuditTrend(&Audit{Value: groups[i-1].Frontend}, &Audit{Value: groups[i].Frontend})
		groups[i].Trend.Backend = calculateAuditTrend(&Audit{Value: groups[i-1].Backend}, &Audit{Value: groups[i].Backend})
		groups[i].Trend.System = calculateAuditTrend(&Audit{Value: groups[i-1].System}, &Audit{Value: groups[i].System})
		groups[i].Trend.Process = calculateAuditTrend(&Audit{Value: groups[i-1].Process}, &Audit{Value: groups[i].Process})
		groups[i].Trend.Mobile = calculateAuditTrend(&Audit{Value: groups[i-1].Mobile}, &Audit{Value: groups[i].Mobile})
		groups[i].Trend.Blockchain = calculateAuditTrend(&Audit{Value: groups[i-1].Blockchain}, &Audit{Value: groups[i].Blockchain})
	}
}

func calculateAuditTrend(previous *Audit, current *Audit) float64 {
	// if previous or current value = 0 trend = 0
	if previous.Value == 0 || current.Value == 0 {
		return 0
	}

	// return the value fixed 2 decimal places
	return math.Round((current.Value-previous.Value)/previous.Value*100*100) / 100
}

type ActionItemSquash struct {
	SnapDate string  `json:"snapDate"`
	Value    int64   `json:"value"`
	Trend    float64 `json:"trend"`
} // @name ActionItemSquash
type ActionItemSquashReport struct {
	All    []*ActionItemSquash `json:"all"`
	High   []*ActionItemSquash `json:"high"`
	Medium []*ActionItemSquash `json:"medium"`
	Low    []*ActionItemSquash `json:"low"`
} // @name ActionItemSquashReport

type ActionItemSquashReportResponse struct {
	Data *ActionItemSquashReport `json:"data"`
} // @name ActionItemSquashReportResponse

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

	if len(actionItemReports) > 1 {
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

		if float64(items[i-1].Value) != 0 {
			items[i].Trend = math.Floor(float64(items[i].Value-items[i-1].Value)/float64(items[i-1].Value)*100*100) / 100
		}
	}
}

type AuditValue struct {
	Value float64 `json:"value"`
	Trend float64 `json:"trend"`
} // @name AuditValue

type ItemValue struct {
	Value int64   `json:"value"`
	Trend float64 `json:"trend"`
} // @name ItemValue

type AuditSummary struct {
	ID           model.UUID `json:"id"`
	Name         string     `json:"name"`
	Code         string     `json:"code"`
	Avatar       string     `json:"avatar"`
	Size         ItemValue  `json:"size"`
	Health       AuditValue `json:"health"`
	Audit        AuditValue `json:"audit"`
	NewItem      ItemValue  `json:"newItem"`
	ResolvedItem ItemValue  `json:"resolvedItem"`
} // @name AuditSummary

type AuditSummaries struct {
	Summary []*AuditSummary `json:"summary"`
} // @name AuditSummaries

type AuditSummariesResponse struct {
	Data *AuditSummaries `json:"data"`
} // @name AuditSummariesResponse

func ToAuditSummary(summary []*model.AuditSummary, previousSize int) *AuditSummary {
	rs := &AuditSummary{
		ID:     summary[0].ID,
		Name:   summary[0].Name,
		Code:   summary[0].Code,
		Avatar: summary[0].Avatar,
	}

	// Size
	rs.Size.Value = summary[0].Size
	if previousSize != 0 {
		rs.Size.Trend = math.Round((float64(rs.Size.Value)-float64(previousSize))/float64(previousSize)*100*100) / 100
	}

	// Health and Audit value
	rs.Health.Value = summary[0].Health
	rs.Audit.Value = summary[0].Audit

	if len(summary) > 1 && summary[1].Audit != 0 && summary[0].Audit != 0 && summary[1].Health != 0 && summary[0].Health != 0 {
		rs.Health.Trend = math.Round((summary[0].Health-summary[1].Health)/summary[1].Health*100*100) / 100
		rs.Audit.Trend = math.Round((summary[0].Audit-summary[1].Audit)/summary[1].Audit*100*100) / 100
	}

	// New and Resolved item
	if summary[0].Size != 0 {
		rs.NewItem.Value = (summary[0].High + summary[0].Medium + summary[0].Low) / summary[0].Size
	}

	if len(summary) > 1 {
		var currentItem int64
		if summary[0].Size != 0 {
			currentItem = (summary[1].High + summary[1].Medium + summary[1].Low) / summary[0].Size
		}

		if currentItem != 0 {
			rs.NewItem.Trend = math.Round((float64(rs.NewItem.Value)-float64(currentItem))/float64(currentItem)*100*100) / 100
		}

		rs.ResolvedItem.Value = summary[1].Done
		if len(summary) > 2 && summary[2].Done != 0 {
			rs.ResolvedItem.Trend = math.Round((float64(summary[1].Done)-float64(summary[2].Done))/float64(summary[2].Done)*100*100) / 100
		}
	}

	return rs
}

func ToAuditSummaries(summaryMap map[model.UUID][]*model.AuditSummary, previousQuarterMap map[model.UUID]int64, allProjectsMap map[model.UUID]*model.ProjectSize) *AuditSummaries {
	rs := &AuditSummaries{}
	for _, summaries := range summaryMap {
		previouSize := 0

		delete(allProjectsMap, summaries[0].ID)
		if size, ok := previousQuarterMap[summaries[0].ID]; ok {
			previouSize = int(size)
		}

		rs.Summary = append(rs.Summary, ToAuditSummary(summaries, previouSize))
	}

	for projectID, project := range allProjectsMap {
		newAuditSummary := &AuditSummary{
			ID:     project.ID,
			Name:   project.Name,
			Code:   project.Code,
			Avatar: project.Avatar,
			Size:   ItemValue{Value: project.Size},
		}

		if size, ok := previousQuarterMap[projectID]; ok {
			previouSize := int(size)

			if previouSize != 0 && newAuditSummary.Size.Trend != 0 {
				newAuditSummary.Size.Trend = math.Round((float64(newAuditSummary.Size.Value)-float64(previouSize))/float64(previouSize)*100*100) / 100
			}
		}

		rs.Summary = append(rs.Summary, newAuditSummary)
	}

	sort.Slice(rs.Summary, func(i, j int) bool {
		return rs.Summary[i].Name < rs.Summary[j].Name
	})

	return rs
}

type AvailableSlot struct {
	ID        string           `json:"id"`
	Type      string           `json:"type"`
	Note      string           `json:"note"`
	CreatedAt string           `json:"createdAt"`
	Seniority Seniority        `json:"seniority"`
	Project   BasicProjectInfo `json:"project"`
	Positions []Position       `json:"positions"`
} // @name AvailableSlot

type AvailableEmployee struct {
	ID          string             `json:"id"`
	FullName    string             `json:"fullName"`
	DisplayName string             `json:"displayName"`
	Username    string             `json:"username"`
	Avatar      string             `json:"avatar"`
	Seniority   Seniority          `json:"seniority"`
	Positions   []Position         `json:"positions"`
	Stacks      []Stack            `json:"stacks"`
	Projects    []BasicProjectInfo `json:"projects"`
} // @name AvailableEmployee

type ResourceAvailability struct {
	Slots     []AvailableSlot     `json:"slots"`
	Employees []AvailableEmployee `json:"employees"`
} // @name ResourceAvailability

type ResourceAvailabilityResponse struct {
	Data ResourceAvailability `json:"data"`
} // @name ResourceAvailabilityResponse

func ToResourceAvailability(slots []*model.ProjectSlot, employees []*model.Employee) ResourceAvailability {
	var res ResourceAvailability

	for _, v := range slots {
		res.Slots = append(res.Slots, AvailableSlot{
			ID:        v.ID.String(),
			Type:      v.DeploymentType.String(),
			Note:      v.Note,
			CreatedAt: v.CreatedAt.String(),
			Seniority: ToSeniority(v.Seniority),
			Project:   *toBasicProjectInfo(v.Project),
			Positions: ToProjectSlotPositions(v.ProjectSlotPositions),
		})
	}

	for _, v := range employees {
		employee := AvailableEmployee{
			ID:          v.ID.String(),
			FullName:    v.FullName,
			DisplayName: v.DisplayName,
			Username:    v.Username,
			Avatar:      v.Avatar,
			Seniority:   ToSeniority(*v.Seniority),
			Positions:   ToEmployeePositions(v.EmployeePositions),
			Stacks:      ToEmployeeStacks(v.EmployeeStacks),
		}

		for _, pm := range v.ProjectMembers {
			project := toBasicProjectInfo(pm.Project)
			employee.Projects = append(employee.Projects, *project)
		}

		res.Employees = append(res.Employees, employee)
	}

	return res
}

func ToEngagementDashboard(statistic []*model.StatisticEngagementDashboard) []EngagementDashboard {
	questionMapper := make(map[string][]EngagementDashboardQuestionStat)
	questionIDMapper := make(map[string]string)
	for _, s := range statistic {
		questionMapper[s.Content] = append(questionMapper[s.Content], EngagementDashboardQuestionStat{
			Title:     strings.Replace(s.Title, ", ", "/", -1),
			StartDate: &s.StartDate,
			Point:     math.Floor(s.Point*100) / 100,
		})
		questionIDMapper[s.Content] = s.QuestionID.String()
	}

	dashboard := make([]EngagementDashboard, 0)

	for k, v := range questionMapper {
		sort.Slice(v, func(i, j int) bool {
			return v[i].StartDate.After(*v[j].StartDate)
		})
		dashboard = append(dashboard, EngagementDashboard{
			Content:    k,
			Stats:      v,
			QuestionID: questionIDMapper[k],
		})
	}

	sort.Slice(dashboard, func(i, j int) bool {
		return dashboard[i].Content < dashboard[j].Content
	})
	return dashboard
}

func ToEngagementDashboardDetails(statistic []*model.StatisticEngagementDashboard) []EngagementDashboardDetail {
	questionMapper := make(map[string][]EngagementDashboardQuestionDetailStat)
	for _, s := range statistic {
		questionMapper[s.QuestionID.String()] = append(questionMapper[s.QuestionID.String()], EngagementDashboardQuestionDetailStat{
			Field:     s.Name,
			StartDate: &s.StartDate,
			Point:     math.Floor(s.Point*100) / 100,
		})
	}

	dashboard := make([]EngagementDashboardDetail, 0)

	for k, v := range questionMapper {
		sort.Slice(v, func(i, j int) bool {
			return v[i].Point < v[j].Point
		})
		dashboard = append(dashboard, EngagementDashboardDetail{
			QuestionID: k,
			Stats:      v,
		})
	}

	sort.Slice(dashboard, func(i, j int) bool {
		return dashboard[i].QuestionID < dashboard[j].QuestionID
	})

	return dashboard
}

type GetEngagementDashboardResponse struct {
	Data []EngagementDashboard `json:"data"`
} // @name GetEngagementDashboardResponse

type GetEngagementDashboardDetailResponse struct {
	Data []EngagementDashboardDetail `json:"data"`
} // @name GetEngagementDashboardDetailResponse

type GetDashboardResourceUtilizationResponse struct {
	Data []ResourceUtilization `json:"data"`
} // @name GetDashboardResourceUtilizationResponse

type WorkUnitDistribution struct {
	Employee    BasicEmployeeInfo               `json:"employee"`
	Learning    WorkUnitDistributionLearning    `json:"learning"`
	Development WorkUnitDistributionDevelopment `json:"development"`
	Management  WorkUnitDistributionManagement  `json:"management"`
	Training    WorkUnitDistributionTraining    `json:"training"`
} // @name WorkUnitDistribution

func ManagementInfoToProjectHead(managementInfo *model.ManagementInfo) *WorkUnitDistributionWUProjectHead {
	return &WorkUnitDistributionWUProjectHead{
		Project: BasicProjectInfo{
			ID:     managementInfo.ID.String(),
			Name:   managementInfo.Name,
			Type:   managementInfo.Type.String(),
			Status: managementInfo.Status.String(),
			Avatar: managementInfo.Avatar,
			Code:   managementInfo.Code,
		},
		Position: managementInfo.Position.String(),
	}
}

func ToWorkUnitDistributionWU(workUnit *model.WorkUnit) *WorkUnitDistributionWU {
	return &WorkUnitDistributionWU{
		Project:      *toBasicProjectInfo(*workUnit.Project),
		WorkUnitName: workUnit.Name,
	}
}

type WorkUnitDistributionLearning struct {
	Total     int                       `json:"total"`
	WorkUnits []*WorkUnitDistributionWU `json:"workUnits"`
} // @name WorkUnitDistributionLearning

type WorkUnitDistributionDevelopment struct {
	Total     int                       `json:"total"`
	WorkUnits []*WorkUnitDistributionWU `json:"workUnits"`
} // @name WorkUnitDistributionDevelopment

type WorkUnitDistributionManagement struct {
	Total        int                                  `json:"total"`
	WorkUnits    []*WorkUnitDistributionWU            `json:"workUnits"`
	ProjectHeads []*WorkUnitDistributionWUProjectHead `json:"projectHeads"`
} // @name WorkUnitDistributionManagement

type WorkUnitDistributionTraining struct {
	Total     int                       `json:"total"`
	WorkUnits []*WorkUnitDistributionWU `json:"workUnits"`
	Mentees   []BasicEmployeeInfo       `json:"mentees"`
} // @name WorkUnitDistributionTraining

type WorkUnitDistributionWU struct {
	Project      BasicProjectInfo `json:"project"`
	WorkUnitName string           `json:"workUnitName"`
} // @name WorkUnitDistributionWU

type WorkUnitDistributionWUProjectHead struct {
	Project  BasicProjectInfo `json:"project"`
	Position string           `json:"position"`
} // @name WorkUnitDistributionWUProjectHead

type WorkUnitDistributionData struct {
	WorkUnitDistributions []*WorkUnitDistribution `json:"workUnitDistributions"`
} // @name WorkUnitDistributionData

func toWorkUnitDistributionWUProjectHeads(projectHeads []model.ProjectHead) []*WorkUnitDistributionWUProjectHead {
	rs := make([]*WorkUnitDistributionWUProjectHead, 0)

	for _, projectHead := range projectHeads {
		rs = append(rs, &WorkUnitDistributionWUProjectHead{
			Project:  *toBasicProjectInfo(projectHead.Project),
			Position: projectHead.Position.String(),
		})
	}

	return rs
}

func SortWorkUnitDistributionData(wudd *WorkUnitDistributionData, sortRequired model.SortOrder) *WorkUnitDistributionData {
	if sortRequired != "" {
		if sortRequired == model.SortOrderASC {
			sort.Slice(wudd.WorkUnitDistributions, func(i, j int) bool {
				return wudd.WorkUnitDistributions[i].Learning.Total+wudd.WorkUnitDistributions[i].Development.Total+wudd.WorkUnitDistributions[i].Management.Total+wudd.WorkUnitDistributions[i].Training.Total <
					wudd.WorkUnitDistributions[j].Learning.Total+wudd.WorkUnitDistributions[j].Development.Total+wudd.WorkUnitDistributions[j].Management.Total+wudd.WorkUnitDistributions[j].Training.Total
			})
		} else {
			sort.Slice(wudd.WorkUnitDistributions, func(i, j int) bool {
				return wudd.WorkUnitDistributions[i].Learning.Total+wudd.WorkUnitDistributions[i].Development.Total+wudd.WorkUnitDistributions[i].Management.Total+wudd.WorkUnitDistributions[i].Training.Total >
					wudd.WorkUnitDistributions[j].Learning.Total+wudd.WorkUnitDistributions[j].Development.Total+wudd.WorkUnitDistributions[j].Management.Total+wudd.WorkUnitDistributions[j].Training.Total
			})
		}
	}

	return wudd
}

func ToWorkUnitDistributionData(employees []*model.Employee, sort string) WorkUnitDistributionData {
	wuds := make([]*WorkUnitDistribution, 0)

	for _, employee := range employees {
		wud := &WorkUnitDistribution{
			Employee: *toBasicEmployeeInfo(*employee),
			Management: WorkUnitDistributionManagement{
				ProjectHeads: toWorkUnitDistributionWUProjectHeads(employee.Heads),
				Total:        len(employee.Heads),
			},
			Training: WorkUnitDistributionTraining{
				Mentees: ToBasicEmployees(employee.Mentees),
				Total:   len(employee.Mentees),
			},
		}

		for _, wum := range employee.WorkUnitMembers {
			switch wum.WorkUnit.Type {
			case model.WorkUnitTypeLearning:
				wud.Learning.WorkUnits = append(wud.Learning.WorkUnits, ToWorkUnitDistributionWU(&wum.WorkUnit))
				wud.Learning.Total++
			case model.WorkUnitTypeDevelopment:
				wud.Development.WorkUnits = append(wud.Development.WorkUnits, ToWorkUnitDistributionWU(&wum.WorkUnit))
				wud.Development.Total++
			case model.WorkUnitTypeManagement:
				wud.Management.WorkUnits = append(wud.Management.WorkUnits, ToWorkUnitDistributionWU(&wum.WorkUnit))
				wud.Management.Total++
			case model.WorkUnitTypeTraining:
				wud.Training.WorkUnits = append(wud.Training.WorkUnits, ToWorkUnitDistributionWU(&wum.WorkUnit))
				wud.Training.Total++
			}
		}
		wuds = append(wuds, wud)
	}

	return *SortWorkUnitDistributionData(&WorkUnitDistributionData{
		WorkUnitDistributions: wuds,
	}, model.SortOrder(sort))
}

type SummaryWorkUnitDistributionData struct {
	Learning    float64 `json:"learning"`
	Development float64 `json:"development"`
	Management  float64 `json:"management"`
	Training    float64 `json:"training"`
} // @name SummaryWorkUnitDistributionData

type SummaryWorkUnitDistributionResponse struct {
	Data *SummaryWorkUnitDistributionData `json:"data"`
} // @name SummaryWorkUnitDistributionResponse

func ToSummaryWorkUnitDistributionData(data *model.TotalWorkUnitDistribution) *SummaryWorkUnitDistributionData {
	total := data.TotalLineManagerCount + data.TotalProjectHead + data.TotalLearning + data.TotalDevelopment + data.TotalManagement + data.TotalTraining
	rs := &SummaryWorkUnitDistributionData{}

	rs.Learning = math.Round(float64(data.TotalLearning+data.TotalLineManagerCount)/float64(total)*100*100) / 100
	rs.Development = math.Round(float64(data.TotalDevelopment)/float64(total)*100*100) / 100
	rs.Management = math.Round(float64(data.TotalManagement+data.TotalProjectHead)/float64(total)*100*100) / 100
	rs.Training = math.Round(float64(data.TotalTraining)/float64(total)*100*100) / 100

	return rs
}

type WorkUnitDistributionsResponse struct {
	Data *WorkUnitDistributionData `json:"data"`
} // @name WorkUnitDistributionsResponse

type WorkSurveySummaryAnswer struct {
	Answer  string           `json:"answer"`
	Project BasicProjectInfo `json:"project"`
} // @name WorkSurveySummaryAnswer

type WorkSurveySummaryListAnswer struct {
	Date    string                    `json:"date"`
	Answers []WorkSurveySummaryAnswer `json:"answers"`
} // @name WorkSurveySummaryListAnswer

type WorkSurveySummaryEmployee struct {
	Reviewer    BasicEmployeeInfo             `json:"reviewer"`
	ListAnswers []WorkSurveySummaryListAnswer `json:"listAnswers"`
} // @name WorkSurveySummaryEmployee

type WorkSurveySummary struct {
	Type  string                      `json:"type"`
	Dates []string                    `json:"dates"`
	Data  []WorkSurveySummaryEmployee `json:"data"`
} // @name WorkSurveySummary

type WorkSurveySummaryResponse struct {
	PaginationResponse
	Data []WorkSurveySummary `json:"data"`
} // @name WorkSurveySummaryResponse

func ToWorkSummaries(eers []*model.EmployeeEventReviewer) []WorkSurveySummary {
	rs := []WorkSurveySummary{
		{
			Type: model.QuestionDomainWorkload.String(),
		},
		{
			Type: model.QuestionDomainDeadline.String(),
		},
		{
			Type: model.QuestionDomainLearning.String(),
		},
	}

	domainMap := map[model.QuestionDomain]*WorkSurveySummary{
		model.QuestionDomainWorkload: &rs[0],
		model.QuestionDomainDeadline: &rs[1],
		model.QuestionDomainLearning: &rs[2],
	}

	answerMap := map[model.QuestionDomain]map[model.UUID]map[string][]WorkSurveySummaryAnswer{
		model.QuestionDomainWorkload: make(map[model.UUID]map[string][]WorkSurveySummaryAnswer),
		model.QuestionDomainDeadline: make(map[model.UUID]map[string][]WorkSurveySummaryAnswer),
		model.QuestionDomainLearning: make(map[model.UUID]map[string][]WorkSurveySummaryAnswer),
	}

	employeeMap := make(map[model.UUID]model.Employee)

	listDate := make([]string, 0)
	for _, eer := range eers {
		employeeMap[eer.ReviewerID] = *eer.Reviewer

		// to order date
		date := eer.Event.EndDate.Format("02/01/2006")
		if !slices.Contains(listDate, date) {
			listDate = append(listDate, date)
		}

		for _, eeq := range eer.EmployeeEventQuestions {
			answer := WorkSurveySummaryAnswer{
				Answer:  model.AgreementLevelValueMap[eeq.Answer].String(),
				Project: *toBasicProjectInfo(*eer.EmployeeEventTopic.Project),
			}

			if answerMap[eeq.Domain][eer.ReviewerID] == nil {
				answerMap[eeq.Domain][eer.ReviewerID] = make(map[string][]WorkSurveySummaryAnswer)
			}
			answerMap[eeq.Domain][eer.ReviewerID][date] = append(answerMap[eeq.Domain][eer.ReviewerID][date], answer)
		}
	}

	for domain, eIDMap := range answerMap {
		domainMap[domain].Dates = listDate

		for eID, dateMap := range eIDMap {
			listAnswers := make([]WorkSurveySummaryListAnswer, 0)

			for _, date := range listDate {
				if dateMap[date] != nil && len(dateMap[date]) > 0 {
					listAnswers = append(listAnswers, WorkSurveySummaryListAnswer{
						Date:    date,
						Answers: dateMap[date],
					})
				}
			}

			employee := WorkSurveySummaryEmployee{
				Reviewer:    *toBasicEmployeeInfo(employeeMap[eID]),
				ListAnswers: listAnswers,
			}

			domainMap[domain].Data = append(domainMap[domain].Data, employee)
		}
	}

	return rs
}

type ResourceUtilization struct {
	Date      time.Time `json:"date"`
	Staffed   int       `json:"staffed"`
	Internal  int       `json:"internal"`
	Available int       `json:"available"`
} // @name ResourceUtilization

func ToReSourceUtilizations(data []*model.ResourceUtilization) []ResourceUtilization {
	rs := make([]ResourceUtilization, 0, len(data))

	for _, v := range data {
		rs = append(rs, ResourceUtilization{
			Date:      v.Date,
			Staffed:   v.Staffed,
			Internal:  v.Internal,
			Available: v.Available,
		})
	}

	return rs
}
