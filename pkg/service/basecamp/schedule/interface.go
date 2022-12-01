package schedule

import "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"

type ScheduleService interface {
	CreateScheduleEntry(projectID int64, scheduleID int64, scheduleEntry model.ScheduleEntry) (res *model.ScheduleEntry, err error)
}
