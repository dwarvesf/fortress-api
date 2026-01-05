package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// HandleNotionTaskOrderSendEmail handles send email confirmation webhook events from Notion
// This webhook is triggered when the "Send email" button is clicked on an Order record
// It reads email content from page body and sends it to contractor
func (h *handler) HandleNotionTaskOrderSendEmail(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "webhook",
		"method":  "HandleNotionTaskOrderSendEmail",
	})

	l.Debug("received notion task order send email webhook request")

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		l.Error(err, "failed to read notion task order webhook body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("received webhook body: %s", string(body)))

	// Parse payload (reuse NotionInvoiceWebhookPayload structure)
	var payload NotionInvoiceWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		l.Error(err, "failed to parse notion task order webhook payload")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Handle verification challenge
	if payload.VerificationToken != "" {
		l.Debug(fmt.Sprintf("responding to notion webhook verification: token=%s challenge=%s", payload.VerificationToken, payload.Challenge))
		if payload.Challenge != "" {
			c.JSON(http.StatusOK, gin.H{"challenge": payload.Challenge})
		} else {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
		return
	}

	// Verify webhook signature if verification token is configured
	verificationToken := h.config.Notion.Secret
	if verificationToken != "" {
		signature := c.GetHeader("X-Notion-Signature")
		l.Debug(fmt.Sprintf("signature header: %s", signature))
		if signature != "" {
			if !h.verifyNotionWebhookSignature(body, signature, verificationToken) {
				l.Error(errors.New("invalid signature"), "webhook signature verification failed")
				c.JSON(http.StatusUnauthorized, view.CreateResponse[any](nil, nil, errors.New("invalid signature"), nil, ""))
				return
			}
			l.Debug("webhook signature verified successfully")
		}
	}

	// Validate payload has data with page ID
	if payload.Data == nil || payload.Data.ID == "" {
		l.Error(errors.New("missing data.id"), "no data.id in webhook payload")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errors.New("missing data.id"), nil, ""))
		return
	}

	pageID := payload.Data.ID
	l.Info(fmt.Sprintf("processing task order send email for page: %s", pageID))

	// Fetch page from Notion to get page properties
	page, err := h.service.Notion.GetPage(pageID)
	if err != nil {
		l.Error(err, "failed to fetch notion page")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("fetched notion page: id=%s", page.ID))

	// Cast properties to DatabasePageProperties
	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		l.Error(errors.New("failed to cast page properties"), "page properties is not DatabasePageProperties")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errors.New("invalid page properties format"), nil, ""))
		return
	}

	// Validate Type property - must be "Order"
	typeProperty, ok := props["Type"]
	if !ok {
		l.Error(errors.New("type property not found"), "page does not have Type property")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errors.New("Type property not found on page"), nil, ""))
		return
	}

	// Access select value from the Type property
	var pageType string
	if typeProperty.Select != nil && typeProperty.Select.Name != "" {
		pageType = typeProperty.Select.Name
	} else {
		l.Error(errors.New("type property has no value"), "Type property is empty")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errors.New("Type property has no value"), nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("page type: %s", pageType))

	if pageType != "Order" {
		l.Info(fmt.Sprintf("rejecting task order send email for non-order type: %s", pageType))
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil,
			fmt.Errorf("task order send email only supported for Order records, got: %s", pageType), nil, ""))
		return
	}

	// Extract Month property for email subject (formula type)
	var month string
	if monthProp, ok := props["Month"]; ok && monthProp.Formula != nil {
		if monthProp.Formula.String != nil {
			month = *monthProp.Formula.String
		}
	}
	l.Debug(fmt.Sprintf("extracted month from formula: %s", month))
	if month == "" {
		l.Error(errors.New("month property not found"), "Order does not have Month property")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errors.New("Month property not found on Order"), nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("order month: %s", month))

	// Get TaskOrderLog service
	taskOrderLogService := h.service.Notion.TaskOrderLog

	ctx := context.Background()

	// Read email content from page body
	content, err := taskOrderLogService.GetOrderPageContent(ctx, pageID)
	if err != nil {
		l.Error(err, "failed to get order page content")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil,
			fmt.Errorf("failed to read email content from page: %w", err), nil, ""))
		return
	}

	if content == "" {
		l.Error(errors.New("empty content"), "order page has no content")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil,
			errors.New("order page has no email content"), nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("read email content from page: %d chars", len(content)))

	// Get contractor info from Order via Sub-items -> Deployment -> Contractor chain
	contractorID, email, name, err := taskOrderLogService.GetContractorFromOrder(ctx, pageID)
	if err != nil {
		l.Error(err, "failed to get contractor from order")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil,
			fmt.Errorf("failed to get contractor: %w", err), nil, ""))
		return
	}

	if email == "" {
		l.Error(errors.New("empty email"), "contractor has no team email")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil,
			fmt.Errorf("contractor %s has no team email", name), nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("contractor info: id=%s name=%s email=%s", contractorID, name, email))

	// Build email data and send
	emailData := &model.TaskOrderRawEmail{
		ContractorName: name,
		TeamEmail:      email,
		Month:          month,
		RawContent:     content,
	}

	l.Info(fmt.Sprintf("sending task order email to: %s for month: %s", email, month))

	if err := h.service.GoogleMail.SendTaskOrderRawContentMail(emailData); err != nil {
		l.Error(err, "failed to send task order email")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil,
			fmt.Errorf("failed to send email: %w", err), nil, ""))
		return
	}

	l.Info(fmt.Sprintf("task order email sent successfully: page_id=%s, contractor=%s, email=%s",
		pageID, name, email))

	c.JSON(http.StatusOK, view.CreateResponse[any](gin.H{
		"status":        "success",
		"message":       "Task order email sent successfully",
		"page_id":       pageID,
		"contractor":    name,
		"email":         email,
		"month":         month,
		"content_chars": len(content),
	}, nil, nil, nil, ""))
}
