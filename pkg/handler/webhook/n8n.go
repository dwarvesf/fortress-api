package webhook

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

func (h *handler) N8n(c *gin.Context) {
	var req n8nEvent

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	req.CalendarData.ShouldSyncDiscord = req.ShouldSyncDiscord

	var err error
	switch req.Kind {
	case KindGcalEvent:
		err = h.handleGcalEvent(req.CalendarData)
	}
	if err != nil {
		h.logger.Fields(logger.Fields{
			"kind": req.Kind,
			"data": req,
		}).Error(err, "can't execute webhook")
	}
}

func (h *handler) handleGcalEvent(event *n8nCalendarEvent) error {
	// first, we get event from db
	existedEv, err := h.store.Schedule.GetOneByGcalID(h.repo.DB(), event.ID)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			h.logger.Error(err, "can't get schedule from db")
			return err
		}
		// if the schedule doesn't exists in db, we create a new one
		return h.handleNewEvent(event)
	}

	// in case the event already exists, we check if the event is updated
	if existedEv.GoogleCalendar.HangoutLink != event.HangoutLink ||
		existedEv.Name != event.Summary ||
		*existedEv.StartTime != event.Start.DateTime ||
		*existedEv.EndTime != event.End.DateTime {
		return h.handleUpdateEvent(existedEv, event)
	}

	return nil
}

func (h *handler) handleUpdateEvent(existedEv *model.Schedule, event *n8nCalendarEvent) error {
	// update schedule
	existedEv.Name = event.Summary
	existedEv.StartTime = &event.Start.DateTime
	existedEv.EndTime = &event.End.DateTime
	existedEv.GoogleCalendar.Description = event.Description
	existedEv.GoogleCalendar.HangoutLink = event.HangoutLink

	schedule, err := h.store.Schedule.Update(h.repo.DB(), existedEv)
	if err != nil {
		h.logger.Error(err, "can't update schedule")
		return err
	}

	// update discord event
	if schedule.DiscordEvent != nil {
		_, err = h.service.Discord.UpdateEvent(schedule)
		if err != nil {
			h.logger.Error(err, "can't update discord event")
			// return err
		}
	}

	return nil
}

func (h *handler) handleNewEvent(event *n8nCalendarEvent) error {
	// first, we create a schedule record in db
	sch := &model.Schedule{
		GoogleCalendar: &model.ScheduleGoogleCalendar{
			GoogleCalendarID: event.ID,
			Description:      event.Description,
			HangoutLink:      event.HangoutLink,
		},
		Name:      event.Summary,
		StartTime: &event.Start.DateTime,
		EndTime:   &event.End.DateTime,
	}
	schedule, err := h.store.Schedule.Create(h.repo.DB(), sch)
	if err != nil {
		h.logger.Error(err, "can't create schedule")
		return err
	}

	// this case, we create a discord event
	// check if time in the past
	if event.ShouldSyncDiscord == "true" {
		if !schedule.StartTime.Before(time.Now()) {
			discordEvent, err := h.service.Discord.CreateEvent(schedule)
			if err != nil {
				h.logger.Error(err, "can't create discord event")
				// return err
			} else {
				_, err = h.store.Schedule.CreateDiscord(h.repo.DB(), &model.ScheduleDiscordEvent{
					ScheduleID:     schedule.ID,
					DiscordEventID: discordEvent.ID,
					VoiceChannelID: discordEvent.ChannelID,
				})
				if err != nil {
					h.logger.Error(err, "can't create discord event")
				}
			}
		}
	}

	return nil
}
