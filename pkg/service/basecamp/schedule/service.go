package schedule

import "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"

type Service interface {
	CreateScheduleEntry(projectID int64, scheduleID int64, scheduleEntry model.ScheduleEntry) (res *model.ScheduleEntry, err error)
	GetScheduleEntries(projectID int64, scheduleID int64) (res []*model.ScheduleEntry, err error)
	UpdateSheduleEntry(projectID int64, se *model.ScheduleEntry) (err error)
}
