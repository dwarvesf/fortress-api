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

// NotionRefundWebhookPayload represents the webhook payload from Notion automation
type NotionRefundWebhookPayload struct {
	Source NotionRefundSource `json:"source"`
	Data   NotionRefundData   `json:"data"`
}

// NotionRefundSource represents the source of the webhook
type NotionRefundSource struct {
	Type         string `json:"type"`
	AutomationID string `json:"automation_id"`
	ActionID     string `json:"action_id"`
	EventID      string `json:"event_id"`
	Attempt      int    `json:"attempt"`
}

// NotionRefundData represents the page data in the webhook
type NotionRefundData struct {
	Object     string                     `json:"object"`
	ID         string                     `json:"id"`
	Properties NotionRefundProperties     `json:"properties"`
	URL        string                     `json:"url"`
}

// NotionRefundProperties represents the properties of the refund page
type NotionRefundProperties struct {
	Status      NotionStatusProperty    `json:"Status"`
	WorkEmail   NotionEmailProperty     `json:"Work Email"`
	Amount      NotionNumberProperty    `json:"Amount"`
	Currency    NotionSelectProperty    `json:"Currency"`
	Reason      NotionSelectProperty    `json:"Reason"`
	Description NotionRichTextProperty  `json:"Description"`
	Contractor  NotionRelationProperty  `json:"Contractor"`
}

// NotionStatusProperty represents a status property
type NotionStatusProperty struct {
	ID     string              `json:"id"`
	Type   string              `json:"type"`
	Status *NotionSelectOption `json:"status"`
}

// NotionEmailProperty represents an email property
type NotionEmailProperty struct {
	ID    string  `json:"id"`
	Type  string  `json:"type"`
	Email *string `json:"email"`
}

// NotionNumberProperty represents a number property
type NotionNumberProperty struct {
	ID     string   `json:"id"`
	Type   string   `json:"type"`
	Number *float64 `json:"number"`
}

// NotionSelectProperty represents a select property
type NotionSelectProperty struct {
	ID     string              `json:"id"`
	Type   string              `json:"type"`
	Select *NotionSelectOption `json:"select"`
}

// NotionSelectOption represents a select option
type NotionSelectOption struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// NotionRichTextProperty represents a rich text property
type NotionRichTextProperty struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	RichText []NotionRichText  `json:"rich_text"`
}

// NotionRichText represents a rich text element
type NotionRichText struct {
	Type      string `json:"type"`
	PlainText string `json:"plain_text"`
}

// NotionRelationProperty represents a relation property
type NotionRelationProperty struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Relation []NotionRelation `json:"relation"`
	HasMore  bool             `json:"has_more"`
}

// NotionRelation represents a relation item
type NotionRelation struct {
	ID string `json:"id"`
}

// HandleNotionRefund handles refund webhook events from Notion
func (h *handler) HandleNotionRefund(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "webhook",
		"method":  "HandleNotionRefund",
	})

	l.Debug("received notion refund webhook request")

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
		l.Error(err, "failed to read notion refund webhook body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("received webhook body: %s", string(body)))

	// Parse payload
	var payload NotionRefundWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		l.Error(err, "failed to parse notion refund webhook payload")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Log parsed data
	l.Debug(fmt.Sprintf("parsed refund: page_id=%s status=%s email=%v amount=%v currency=%v",
		payload.Data.ID,
		getStatusName(payload.Data.Properties.Status),
		getEmailValue(payload.Data.Properties.WorkEmail),
		getNumberValue(payload.Data.Properties.Amount),
		getSelectName(payload.Data.Properties.Currency),
	))

	// Check if Contractor relation is empty and we have a Work Email
	if len(payload.Data.Properties.Contractor.Relation) == 0 {
		email := getEmailValue(payload.Data.Properties.WorkEmail)
		if email != "" {
			l.Debug(fmt.Sprintf("contractor relation is empty, looking up by email: %s", email))

			// Look up contractor page ID by email
			contractorPageID, err := h.lookupContractorByEmail(c.Request.Context(), l, email)
			if err != nil {
				l.Error(err, fmt.Sprintf("failed to lookup contractor by email: %s", email))
				// Continue without updating - don't fail the webhook
			} else if contractorPageID != "" {
				l.Debug(fmt.Sprintf("found contractor page: email=%s page_id=%s", email, contractorPageID))

				// Update the refund page with contractor relation
				if err := h.updateRefundContractor(c.Request.Context(), l, payload.Data.ID, contractorPageID); err != nil {
					l.Error(err, fmt.Sprintf("failed to update contractor relation: page_id=%s", payload.Data.ID))
				} else {
					l.Debug(fmt.Sprintf("successfully updated contractor relation: refund_page=%s contractor_page=%s", payload.Data.ID, contractorPageID))
				}
			}
		} else {
			l.Debug("contractor relation is empty but no work email provided")
		}
	} else {
		l.Debug(fmt.Sprintf("contractor relation already set: %v", payload.Data.Properties.Contractor.Relation))
	}

	// Return success to acknowledge receipt
	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "processed"))
}

// lookupContractorByEmail queries the contractor database to find a contractor by email
func (h *handler) lookupContractorByEmail(ctx context.Context, l logger.Logger, email string) (string, error) {
	contractorDBID := h.config.Notion.Databases.Contractor
	if contractorDBID == "" {
		return "", fmt.Errorf("NOTION_CONTRACTOR_DB_ID not configured")
	}

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

// updateRefundContractor updates the Contractor relation field on the refund page
func (h *handler) updateRefundContractor(ctx context.Context, l logger.Logger, refundPageID, contractorPageID string) error {
	l.Debug(fmt.Sprintf("updating refund contractor: refund_page=%s contractor_page=%s", refundPageID, contractorPageID))

	// Create Notion client
	client := nt.NewClient(h.config.Notion.Secret)

	// Build update params with Contractor relation
	updateParams := nt.UpdatePageParams{
		DatabasePageProperties: nt.DatabasePageProperties{
			"Contractor": nt.DatabasePageProperty{
				Relation: []nt.Relation{
					{ID: contractorPageID},
				},
			},
		},
	}

	l.Debug(fmt.Sprintf("update params: %+v", updateParams))

	updatedPage, err := client.UpdatePage(ctx, refundPageID, updateParams)
	if err != nil {
		l.Error(err, fmt.Sprintf("notion API error updating page: %s", refundPageID))
		return fmt.Errorf("failed to update page: %w", err)
	}

	l.Debug(fmt.Sprintf("notion API response - page ID: %s, URL: %s", updatedPage.ID, updatedPage.URL))
	l.Debug(fmt.Sprintf("successfully updated contractor relation on refund page: %s", refundPageID))
	return nil
}

// Helper functions for extracting property values

func getStatusName(prop NotionStatusProperty) string {
	if prop.Status != nil {
		return prop.Status.Name
	}
	return ""
}

func getEmailValue(prop NotionEmailProperty) string {
	if prop.Email != nil {
		return *prop.Email
	}
	return ""
}

func getNumberValue(prop NotionNumberProperty) *float64 {
	return prop.Number
}

func getSelectName(prop NotionSelectProperty) string {
	if prop.Select != nil {
		return prop.Select.Name
	}
	return ""
}
