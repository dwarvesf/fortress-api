package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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

// NotionRefundSource represents the source of the refund webhook
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
	Status      NotionStatusProperty   `json:"Status"`
	WorkEmail   NotionEmailProperty    `json:"Work Email"`
	Amount      NotionNumberProperty   `json:"Amount"`
	Currency    NotionSelectProperty   `json:"Currency"`
	Reason      NotionSelectProperty   `json:"Reason"`
	Description NotionRichTextProperty `json:"Description"`
	Contractor  NotionRelationProperty `json:"Contractor"`
	CreatedBy   NotionCreatedByProperty `json:"Created by"`
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

// NotionCreatedByProperty represents a created_by property
type NotionCreatedByProperty struct {
	ID        string           `json:"id"`
	Type      string           `json:"type"`
	CreatedBy *NotionUserObject `json:"created_by"`
}

// NotionUserObject represents a Notion user
type NotionUserObject struct {
	Object string `json:"object"`
	ID     string `json:"id"`
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
	createdByUserID := getCreatedByUserID(payload.Data.Properties.CreatedBy)
	l.Debug(fmt.Sprintf("parsed refund: page_id=%s status=%s created_by_user_id=%s amount=%v currency=%v",
		payload.Data.ID,
		getStatusName(payload.Data.Properties.Status),
		createdByUserID,
		getNumberValue(payload.Data.Properties.Amount),
		getSelectName(payload.Data.Properties.Currency),
	))

	// Check if Contractor relation is empty and we have a Created By user
	if len(payload.Data.Properties.Contractor.Relation) == 0 {
		if createdByUserID != "" {
			l.Debug(fmt.Sprintf("contractor relation is empty, looking up by created_by user_id: %s", createdByUserID))

			// Look up contractor page ID by user ID
			contractorPageID, err := h.lookupContractorByUserID(c.Request.Context(), l, createdByUserID)
			if err != nil {
				l.Error(err, fmt.Sprintf("failed to lookup contractor by user_id: %s", createdByUserID))
				// Continue without updating - don't fail the webhook
			} else if contractorPageID != "" {
				l.Debug(fmt.Sprintf("found contractor page: user_id=%s page_id=%s", createdByUserID, contractorPageID))

				// Update the refund page with contractor relation
				if err := h.updateRefundContractor(c.Request.Context(), l, payload.Data.ID, contractorPageID); err != nil {
					l.Error(err, fmt.Sprintf("failed to update contractor relation: page_id=%s", payload.Data.ID))
				} else {
					l.Debug(fmt.Sprintf("successfully updated contractor relation: refund_page=%s contractor_page=%s", payload.Data.ID, contractorPageID))
				}
			}
		} else {
			l.Debug("contractor relation is empty but no created_by user provided")
		}
	} else {
		l.Debug(fmt.Sprintf("contractor relation already set: %v", payload.Data.Properties.Contractor.Relation))
	}

	// Generate and update refund details asynchronously via LLM
	go h.generateAndUpdateRefundDetails(payload.Data.ID)

	// Return success to acknowledge receipt
	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "processed"))
}

// lookupContractorByUserID queries the contractor database to find a contractor by Notion user ID
func (h *handler) lookupContractorByUserID(ctx context.Context, l logger.Logger, userID string) (string, error) {
	contractorDBID := h.config.Notion.Databases.Contractor
	if contractorDBID == "" {
		return "", fmt.Errorf("NOTION_CONTRACTOR_DB_ID not configured")
	}

	l.Debug(fmt.Sprintf("querying contractor database: db_id=%s user_id=%s", contractorDBID, userID))

	// Create Notion client
	client := nt.NewClient(h.config.Notion.Secret)

	// Query contractor database for matching Person (user ID)
	filter := &nt.DatabaseQueryFilter{
		Property: "Person",
		DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
			People: &nt.PeopleDatabaseQueryFilter{
				Contains: userID,
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
		l.Debug(fmt.Sprintf("no contractor found for user_id: %s", userID))
		return "", nil
	}

	pageID := resp.Results[0].ID
	l.Debug(fmt.Sprintf("found contractor: user_id=%s page_id=%s", userID, pageID))
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

func getNumberValue(prop NotionNumberProperty) *float64 {
	return prop.Number
}

func getSelectName(prop NotionSelectProperty) string {
	if prop.Select != nil {
		return prop.Select.Name
	}
	return ""
}

func getCreatedByUserID(prop NotionCreatedByProperty) string {
	if prop.CreatedBy != nil {
		return prop.CreatedBy.ID
	}
	return ""
}

// fetchDescriptionFormatted fetches the "Description Formatted" formula value from a refund page.
// It retries with exponential backoff (1s, 2s, 4s, 8s, 16s) waiting for Notion to compute the formula.
func (h *handler) fetchDescriptionFormatted(pageID string) (string, error) {
	l := h.logger.Fields(logger.Fields{
		"handler": "webhook",
		"method":  "fetchDescriptionFormatted",
		"page_id": pageID,
	})

	client := nt.NewClient(h.config.Notion.Secret)

	maxRetries := 5
	baseDelay := 1 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			retryDelay := baseDelay * (1 << (attempt - 1))
			l.Debug(fmt.Sprintf("retry %d/%d: waiting %v for formula computation", attempt, maxRetries, retryDelay))
			time.Sleep(retryDelay)
		}

		page, err := client.FindPageByID(context.Background(), pageID)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to fetch page on attempt %d", attempt))
			continue
		}

		props, ok := page.Properties.(nt.DatabasePageProperties)
		if !ok {
			l.Debug("failed to cast page properties")
			continue
		}

		prop, ok := props["Description Formatted"]
		if !ok || prop.Formula == nil {
			l.Debug(fmt.Sprintf("attempt %d: Description Formatted property not found or formula nil", attempt))
			continue
		}

		if prop.Formula.String != nil && *prop.Formula.String != "" {
			value := *prop.Formula.String
			l.Debug(fmt.Sprintf("fetched Description Formatted: %d characters", len(value)))
			return value, nil
		}

		l.Debug(fmt.Sprintf("attempt %d: formula string is empty", attempt))
	}

	return "", fmt.Errorf("Description Formatted still empty after %d retries", maxRetries)
}

// generateAndUpdateRefundDetails runs asynchronously to generate a short summary
// from the "Description Formatted" formula and write it back to the "Details" field.
func (h *handler) generateAndUpdateRefundDetails(pageID string) {
	l := h.logger.Fields(logger.Fields{
		"handler": "webhook",
		"method":  "generateAndUpdateRefundDetails",
		"page_id": pageID,
	})

	l.Debug("starting async refund details generation")

	// Check if OpenRouter service is available
	if h.service.OpenRouter == nil {
		l.Error(fmt.Errorf("OpenRouter service is nil"), "skipping refund details generation")
		return
	}

	// Fetch the formula value with retry
	descFormatted, err := h.fetchDescriptionFormatted(pageID)
	if err != nil {
		l.Warn(fmt.Sprintf("skipping details generation: %v", err))
		return
	}

	if descFormatted == "" {
		l.Warn("Description Formatted is empty, skipping details generation")
		return
	}

	l.Debug(fmt.Sprintf("generating details summary from %d characters of input", len(descFormatted)))

	// Call LLM to generate a concise summary
	systemPrompt := `Summarize the following refund request description into a concise phrase of 5-10 words. Be factual and specific. Output only the summary, nothing else.`
	summary, err := h.service.OpenRouter.GenerateText(
		context.Background(),
		systemPrompt,
		descFormatted,
		"google/gemini-2.5-flash-lite",
		150,
		0.0,
	)
	if err != nil {
		l.Error(err, "failed to generate refund details via LLM")
		return
	}

	l.Debug(fmt.Sprintf("LLM generated summary: %s", summary))

	// Update the refund page's Details rich text property
	client := nt.NewClient(h.config.Notion.Secret)
	updateParams := nt.UpdatePageParams{
		DatabasePageProperties: nt.DatabasePageProperties{
			"Details": nt.DatabasePageProperty{
				RichText: []nt.RichText{
					{
						Text: &nt.Text{Content: summary},
					},
				},
			},
		},
	}

	_, err = client.UpdatePage(context.Background(), pageID, updateParams)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to update Details on page: %s", pageID))
		return
	}

	l.Debug(fmt.Sprintf("successfully updated Details on refund page: %s", pageID))
}
