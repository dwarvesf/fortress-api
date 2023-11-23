package webhook

import (
	"fmt"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

type OnLeaveData struct {
	ParentID                int64
	Name                    string
	Type                    string
	StartDate               time.Time
	EndDate                 time.Time
	Shift                   string
	Title                   string
	Description             string
	AssigneeIDs             []int
	CreatorBasecampID       int
	ApproverBasecampID      int
	AssigneeBasecampIDs     []int
	CompletionSubscriberIDs []int
}

func parseOnLeaveDataFromMessage(todo *bcModel.Todo, msg model.BasecampWebhookMessage) (OnLeaveData, error) {
	data := OnLeaveData{}
	data.ParentID = msg.Recording.Parent.ID
	data.Title = msg.Recording.Title

	split := strings.Split(msg.Recording.Title, "|")
	if len(split) < 3 {
		return data, nil
	}

	// set name and off type
	data.Name = strings.TrimSpace(split[0])
	data.Type = strings.TrimSpace(split[1])

	// set start and end date
	times, err := timeutil.GetTimeRange(strings.TrimSpace(split[2]))
	if err != nil {
		return data, err
	}

	data.StartDate = *times[0]
	data.EndDate = *times[0]
	if len(times) == 2 {
		data.EndDate = *times[1]
	}

	// set shift
	if len(split) > 3 {
		data.Shift = strings.TrimSpace(split[3])
	}

	// set email infos
	data.CreatorBasecampID = msg.Recording.Creator.ID
	data.ApproverBasecampID = msg.Creator.ID

	assigneeIds := []int{
		msg.Creator.ID,
		msg.Recording.Creator.ID,
		consts.AutoBotID,
	}

	if todo != nil {
		data.Description = strings.TrimSuffix(strings.TrimPrefix(todo.Description, "<div>"), "</div>")
		data.CompletionSubscriberIDs = getSubscriberIDs(todo.CompletionSubscribers)

		for _, v := range todo.Assignees {
			assigneeIds = append(assigneeIds, v.ID)
		}
	}

	data.AssigneeIDs = assigneeIds

	return data, nil
}

func (h *handler) handleOnLeaveValidation(msg model.BasecampWebhookMessage) error {
	if h.config.Env == "prod" {
		if msg.Recording.Bucket.Name != "Woodland" {
			return nil
		}
	}

	projectID := msg.Recording.Bucket.ID
	recordingID := msg.Recording.ID
	var commentMsg bcModel.BasecampCommentMessage

	defer func() {
		h.worker.Enqueue(bcModel.BasecampCommentMsg, commentMsg)
	}()

	err := h.validateOnLeaveData(msg)
	if err != nil {
		errMsg := fmt.Sprintf(
			`   Your request has encountered an error: %v
    
                Your Name | Off | Date (in range, format dd/mm/yyyy) | Shift (if any)
    
                e.g:
                Huy Nguyen | Off | 29/01/2019 | Morning
                Nam Nguyen | Off | 29/01/2019 - 01/02/2020`, err.Error())

		commentMsg = h.service.Basecamp.BuildCommentMessage(projectID, recordingID, errMsg, "")
		return err
	}

	commentMsg = h.service.Basecamp.BuildCommentMessage(projectID, recordingID, "Your format looks good to go 👍", "")
	return nil
}

func (h *handler) validateOnLeaveData(msg model.BasecampWebhookMessage) error {
	data, err := parseOnLeaveDataFromMessage(nil, msg)
	if err != nil {
		return err
	}

	// Validate if request belongs to the onLeave group
	onLeaveID := consts.OnleavePlaygroundID
	if h.config.Env == "prod" {
		onLeaveID = consts.OnleaveID
	}

	if data.ParentID != int64(onLeaveID) {
		list, err := h.service.Basecamp.Todo.GetList(msg.Recording.Parent.URL)
		if err != nil {
			return fmt.Errorf("cannot get todo list: %v", err.Error())
		}
		if list.Parent.ID != onLeaveID {
			return fmt.Errorf("invalid group id: %v %v", list.Parent.ID, list.Parent.Title)
		}
	}

	// Validate title format
	split := strings.Split(msg.Recording.Title, "|")
	if len(split) < 3 {
		return fmt.Errorf("invalid title format: %v", msg.Recording.Title)
	}

	// Validate off type
	offtype := strings.ToLower(data.Type)
	if offtype != "off" && offtype != "remote" {
		return fmt.Errorf("invalid off type: %v (needs to be off or remote)", offtype)
	}

	// Validate time range
	if data.StartDate.Before(time.Now()) && !timeutil.IsSameDay(data.StartDate, time.Now()) {
		return fmt.Errorf("start date cannot be in the past: %v", timeutil.ParseTimeToDateFormat(&data.StartDate))
	}

	if data.EndDate.Before(data.StartDate) {
		return fmt.Errorf("end date must be after start date: start date is %v - end date is %v", timeutil.ParseTimeToDateFormat(&data.StartDate), timeutil.ParseTimeToDateFormat(&data.EndDate))
	}

	return nil
}

func (h *handler) handleApproveOnLeaveRequest(msg model.BasecampWebhookMessage) error {
	todo, err := h.service.Basecamp.Todo.Get(msg.Recording.URL)
	if err != nil {
		h.logger.Errorf(err, "failed to get basecamp todo: %v", err.Error())
		return fmt.Errorf("failed to get basecamp todo: %v", err.Error())
	}

	data, err := parseOnLeaveDataFromMessage(todo, msg)
	if err != nil {
		return err
	}

	r := model.OnLeaveRequest{
		Type:        data.Type,
		StartDate:   &data.StartDate,
		EndDate:     &data.EndDate,
		Shift:       data.Shift,
		Title:       data.Title,
		Description: data.Description,
	}

	// assign assignees id
	for _, assignee := range todo.Assignees {
		data.AssigneeBasecampIDs = append(data.AssigneeBasecampIDs, assignee.ID)
	}

	assignees, err := h.store.Employee.GetByBasecampIDs(h.repo.DB(), data.AssigneeBasecampIDs)
	if err != nil {
		h.logger.Errorf(err, "failed to get assignees with basecamp_ids", err.Error())
		return fmt.Errorf("failed to get assignees with basecamp_ids %v: %v", data.AssigneeBasecampIDs, err.Error())
	}

	for _, assignee := range assignees {
		r.AssigneeIDs = append(r.AssigneeIDs, assignee.ID.String())
	}

	// assign creator id
	creator, err := h.store.Employee.OneByBasecampID(h.repo.DB(), data.CreatorBasecampID)
	if err != nil {
		h.logger.Errorf(err, "failed to get creator with basecamp_id", err.Error())
		return fmt.Errorf("failed to get creator with basecamp_id %v: %v", data.CreatorBasecampID, err.Error())
	}
	r.CreatorID = creator.ID

	// assign approver id
	approver, err := h.store.Employee.OneByBasecampID(h.repo.DB(), data.ApproverBasecampID)
	if err != nil {
		h.logger.Errorf(err, "failed to get approver with basecamp_id", err.Error())
		return fmt.Errorf("failed to get approver with basecamp_id %v: %v", data.ApproverBasecampID, err.Error())
	}
	r.ApproverID = approver.ID

	_, err = h.store.OnLeaveRequest.Create(h.repo.DB(), &r)
	if err != nil {
		h.logger.Errorf(err, "failed to create onLeaveRequest", err.Error())
		return fmt.Errorf("failed to create onLeaveRequest: %v", err.Error())
	}

	dateChunks := timeutil.ChunkDateRange(data.StartDate, data.EndDate)
	for _, dateChunk := range dateChunks {
		startDate := dateChunk[0]
		endDate := dateChunk[1]

		basecampSchedule := bcModel.ScheduleEntry{
			Summary:     fmt.Sprintf("⚠️ %s", data.Title),
			StartsAt:    startDate.Format(time.RFC3339),
			EndsAt:      endDate.Format(time.RFC3339),
			AllDay:      true,
			Description: r.Description,
		}

		woodlandID := consts.PlaygroundID
		woodlandScheduleID := consts.PlaygroundScheduleID

		opsTeamIDs := []int{consts.NamNguyenBasecampID}
		if h.config.Env == "prod" {
			woodlandID = consts.WoodlandID
			woodlandScheduleID = consts.WoodlandScheduleID
			opsTeamIDs = []int{consts.HuyNguyenBasecampID, consts.GiangThanBasecampID}
			if msg.Recording.Bucket.Name != "Woodland" {
				return nil
			}
		}
		var subscriberIDs []int
		subscriberIDs = append(subscriberIDs, data.AssigneeIDs...)
		subscriberIDs = append(subscriberIDs, data.CompletionSubscriberIDs...)
		subscriberIDs = append(subscriberIDs, opsTeamIDs...)

		se, err := h.service.Basecamp.Schedule.CreateScheduleEntry(int64(woodlandID), woodlandScheduleID, basecampSchedule)
		if err != nil {
			h.logger.Errorf(err, "failed to  create basecamp schedule", err.Error())
			return fmt.Errorf("failed to create basecamp schedule: %v", err.Error())
		}

		if len(data.AssigneeIDs) > 0 {
			err = h.service.Basecamp.Subscription.Subscribe(
				se.SubscriptionUrl,
				&bcModel.SubscriptionList{Subscriptions: subscriberIDs},
			)
			if err != nil {
				h.logger.Errorf(err, "failed to set basecamp event subscriber", err.Error())
				return err
			}
		}
	}

	return nil
}

// Get list IDs from list subscribers
func getSubscriberIDs(list []bcModel.Subscriber) []int {
	res := make([]int, len(list))
	for i := range list {
		res[i] = list[i].ID
	}
	return res
}
