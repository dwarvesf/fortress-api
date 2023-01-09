package view

import "github.com/dwarvesf/fortress-api/pkg/model"

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

type WorkSurveysData struct {
	Project     *BasicProjectInfo `json:"project"`
	WorkSurveys []*WorkSurvey     `json:"workSurveys"`
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
