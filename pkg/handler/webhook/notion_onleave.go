package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	nt "github.com/dstotijn/go-notion"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// NotionOnLeaveWebhookPayload represents the webhook payload from Notion automation for on-leave
type NotionOnLeaveWebhookPayload struct {
	Source NotionOnLeaveSource `json:"source"`
	Data   NotionOnLeaveData   `json:"data"`
}

// NotionOnLeaveSource represents the source of the webhook
type NotionOnLeaveSource struct {
	Type         string `json:"type"`
	AutomationID string `json:"automation_id"`
	ActionID     string `json:"action_id"`
	EventID      string `json:"event_id"`
	Attempt      int    `json:"attempt"`
}

// NotionOnLeaveData represents the page data in the webhook
type NotionOnLeaveData struct {
	Object     string                   `json:"object"`
	ID         string                   `json:"id"`
	Properties NotionOnLeaveProperties  `json:"properties"`
	URL        string                   `json:"url"`
}

// NotionOnLeaveProperties represents the properties of the on-leave page
type NotionOnLeaveProperties struct {
	Status    NotionStatusProperty   `json:"Status"`
	TeamEmail NotionEmailProperty    `json:"Team Email"`
	Employee  NotionRelationProperty `json:"Employee"`
}

// HandleNotionOnLeave handles on-leave webhook events from Notion automation
// This endpoint auto-fills the Employee relation based on Team Email
func (h *handler) HandleNotionOnLeave(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "webhook",
		"method":  "HandleNotionOnLeave",
	})

	l.Debug("received notion on-leave webhook request")

	// Log all headers for debugging
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	headersJSON, _ := json.Marshal(headers)
	l.Debug(fmt.Sprintf("request headers: %s", string(headersJSON)))

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		l.Error(err, "failed to read notion on-leave webhook body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("received webhook body: %s", string(body)))

	// Parse payload
	var payload NotionOnLeaveWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		l.Error(err, "failed to parse notion on-leave webhook payload")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Log parsed data
	l.Debug(fmt.Sprintf("parsed on-leave: page_id=%s status=%s email=%v",
		payload.Data.ID,
		getStatusName(payload.Data.Properties.Status),
		getEmailValue(payload.Data.Properties.TeamEmail),
	))

	// Check if Employee relation is empty and we have a Team Email
	if len(payload.Data.Properties.Employee.Relation) == 0 {
		email := getEmailValue(payload.Data.Properties.TeamEmail)
		if email != "" {
			l.Debug(fmt.Sprintf("employee relation is empty, looking up contractor by email: %s", email))

			// Look up contractor page ID by email
			contractorPageID, err := h.lookupContractorByEmailForOnLeave(c.Request.Context(), l, email)
			if err != nil {
				l.Error(err, fmt.Sprintf("failed to lookup contractor by email: %s", email))
				// Continue without updating - don't fail the webhook
			} else if contractorPageID != "" {
				l.Debug(fmt.Sprintf("found contractor page: email=%s page_id=%s", email, contractorPageID))

				// Update the on-leave page with Employee relation
				if err := h.updateOnLeaveEmployee(c.Request.Context(), l, payload.Data.ID, contractorPageID); err != nil {
					l.Error(err, fmt.Sprintf("failed to update employee relation: page_id=%s", payload.Data.ID))
				} else {
					l.Debug(fmt.Sprintf("successfully updated employee relation: onleave_page=%s contractor_page=%s", payload.Data.ID, contractorPageID))
				}
			}
		} else {
			l.Debug("employee relation is empty but no team email provided")
		}
	} else {
		l.Debug(fmt.Sprintf("employee relation already set: %v", payload.Data.Properties.Employee.Relation))
	}

	// Return success to acknowledge receipt
	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "processed"))
}

// lookupContractorByEmailForOnLeave queries the contractor database to find a contractor by email
func (h *handler) lookupContractorByEmailForOnLeave(ctx context.Context, l logger.Logger, email string) (string, error) {
	// Use hardcoded contractor DB ID for now
	contractorDBID := "9d468753ebb44977a8dc156428398a6b"

	l.Debug(fmt.Sprintf("querying contractor database: db_id=%s email=%s", contractorDBID, email))

	// Create Notion client
	client := nt.NewClient(h.config.Notion.Secret)

	// Query contractor database for matching email
	filter := &nt.DatabaseQueryFilter{
		Property: "Team Email",
		DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
			Email: &nt.TextPropertyFilter{
				Equals: email,
			},
		},
	}

	resp, err := client.QueryDatabase(ctx, contractorDBID, &nt.DatabaseQuery{
		Filter:   filter,
		PageSize: 1,
	})
	if err != nil {
		return "", fmt.Errorf("failed to query contractor database: %w", err)
	}

	if len(resp.Results) == 0 {
		l.Debug(fmt.Sprintf("no contractor found for email: %s", email))
		return "", nil
	}

	pageID := resp.Results[0].ID
	l.Debug(fmt.Sprintf("found contractor: email=%s page_id=%s", email, pageID))
	return pageID, nil
}

// updateOnLeaveEmployee updates the Employee relation field on the on-leave page
func (h *handler) updateOnLeaveEmployee(ctx context.Context, l logger.Logger, onLeavePageID, contractorPageID string) error {
	l.Debug(fmt.Sprintf("updating on-leave employee: onleave_page=%s contractor_page=%s", onLeavePageID, contractorPageID))

	// Create Notion client
	client := nt.NewClient(h.config.Notion.Secret)

	// Build update params with Employee relation
	updateParams := nt.UpdatePageParams{
		DatabasePageProperties: nt.DatabasePageProperties{
			"Employee": nt.DatabasePageProperty{
				Relation: []nt.Relation{
					{ID: contractorPageID},
				},
			},
		},
	}

	l.Debug(fmt.Sprintf("update params: %+v", updateParams))

	updatedPage, err := client.UpdatePage(ctx, onLeavePageID, updateParams)
	if err != nil {
		l.Error(err, fmt.Sprintf("notion API error updating page: %s", onLeavePageID))
		return fmt.Errorf("failed to update page: %w", err)
	}

	l.Debug(fmt.Sprintf("notion API response - page ID: %s, URL: %s", updatedPage.ID, updatedPage.URL))
	l.Debug(fmt.Sprintf("successfully updated employee relation on on-leave page: %s", onLeavePageID))
	return nil
}
