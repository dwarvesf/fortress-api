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
	ParentID       int64
	Name           string
	Type           string
	StartDate      time.Time
	EndDate        time.Time
	Shift          string
	Title          string
	CreatorEmail   string
	ApproverEmail  string
	AssigneeEmails []string
}

func parseOnLeaveDataFromMessage(msg model.BasecampWebhookMessage) (OnLeaveData, error) {
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
	data.CreatorEmail = msg.Recording.Creator.Email
	data.ApproverEmail = msg.Creator.Email

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
    
                I'm not smart enough to understand your on-leave request. Please ensure the following fortmat before handing it to someone 😊
    
                Your Name | Off/Remote | Date (in range, format dd/mm/yyyy) | Shift (if any)
    
                e.g:
                Huy Nguyen | Off | 29/01/2019 | Morning
                Minh Luu | Remote | 29/01/2019`, err.Error())

		commentMsg = h.service.Basecamp.BuildCommentMessage(projectID, recordingID, errMsg, "")
		return err
	}

	commentMsg = h.service.Basecamp.BuildCommentMessage(projectID, recordingID, "Your format looks good to go 👍", "")
	return nil
}

func (h *handler) validateOnLeaveData(msg model.BasecampWebhookMessage) error {
	data, err := parseOnLeaveDataFromMessage(msg)
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
	if h.config.Env == "prod" {
		if msg.Recording.Bucket.Name != "Woodland" {
			return nil
		}
	}

	data, err := parseOnLeaveDataFromMessage(msg)
	if err != nil {
		return err
	}

	r := model.OnLeaveRequest{
		Type:      data.Type,
		StartDate: &data.StartDate,
		EndDate:   &data.EndDate,
		Shift:     data.Shift,
		Title:     data.Title,
	}

	todo, err := h.service.Basecamp.Todo.Get(msg.Recording.URL)
	if err != nil {
		return fmt.Errorf("cannot get todo: %v", err.Error())
	}

	r.Description = strings.TrimSuffix(strings.TrimPrefix(todo.Description, "<div>"), "</div>")

	// assign assignees id
	for _, assignee := range todo.Assignees {
		data.AssigneeEmails = append(data.AssigneeEmails, assignee.EmailAddress)
	}

	assignees, err := h.store.Employee.GetByEmails(h.repo.DB(), data.AssigneeEmails)
	if err != nil {
		return fmt.Errorf("cannot get assignees with email %v: %v", data.AssigneeEmails, err.Error())
	}

	for _, assignee := range assignees {
		r.AssigneeIDs = append(r.AssigneeIDs, assignee.ID.String())
	}

	// assign creator id
	creator, err := h.store.Employee.OneByEmail(h.repo.DB(), data.CreatorEmail)
	if err != nil {
		return fmt.Errorf("cannot get creator with email %v: %v", data.CreatorEmail, err.Error())
	}
	r.CreatorID = creator.ID

	// assign appprover id
	approver, err := h.store.Employee.OneByEmail(h.repo.DB(), data.ApproverEmail)
	if err != nil {
		return fmt.Errorf("cannot get approver with email %v: %v", data.ApproverEmail, err.Error())
	}
	r.ApproverID = approver.ID

	_, err = h.store.OnLeaveRequest.Create(h.repo.DB(), &r)
	if err != nil {
		return fmt.Errorf("cannot create onLeaveRequest: %v", err.Error())
	}

	return nil
}
