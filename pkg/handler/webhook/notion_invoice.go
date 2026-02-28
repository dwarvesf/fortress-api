package webhook

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// NotionInvoiceWebhookPayload represents the webhook payload from Notion for invoice generation
type NotionInvoiceWebhookPayload struct {
	// Verification fields (for endpoint verification challenge)
	VerificationToken string `json:"verification_token"` // Verification token from Notion
	Challenge         string `json:"challenge"`          // Challenge string to respond with

	// Event fields
	Source *NotionInvoiceWebhookSource `json:"source"` // Automation source info
	Data   *NotionInvoiceWebhookData   `json:"data"`   // The page data
}

// NotionInvoiceWebhookSource represents the automation source
type NotionInvoiceWebhookSource struct {
	Type         string `json:"type"`          // "automation"
	AutomationID string `json:"automation_id"` // Automation ID
}

// NotionInvoiceWebhookData represents the page data in the webhook payload
type NotionInvoiceWebhookData struct {
	Object string `json:"object"` // "page"
	ID     string `json:"id"`     // Page ID
}

// HandleNotionInvoiceGenerate handles invoice generation webhook events from Notion
// This webhook is triggered when the "Generate Invoice" button is clicked in Notion
// It validates that the page type is "Invoice" (not "Line Item"), generates a PDF, and uploads it
func (h *handler) HandleNotionInvoiceGenerate(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "webhook",
		"method":  "HandleNotionInvoiceGenerate",
	})

	l.Debug("received notion invoice generation webhook request")

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		l.Error(err, "failed to read notion invoice webhook body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("received webhook body: %s", string(body)))

	// Parse payload
	var payload NotionInvoiceWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		l.Error(err, "failed to parse notion invoice webhook payload")
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
	l.Info(fmt.Sprintf("processing invoice generation for page: %s", pageID))

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

	// Validate Type property - must be "Invoice", not "Line Item"
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

	if pageType != "Invoice" {
		l.Info(fmt.Sprintf("rejecting invoice generation for non-invoice type: %s", pageType))
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil,
			fmt.Errorf("invoice generation only supported for Invoice records, got: %s", pageType), nil, ""))
		return
	}

	// Extract invoice data from Notion properties
	invoice, lineItems, err := h.extractInvoiceDataFromNotion(l, page, props)
	if err != nil {
		l.Error(err, "failed to extract invoice data from notion")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil,
			fmt.Errorf("failed to extract invoice data: %w", err), nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("invoice data extracted: number=%s, total=%.2f, items=%d", invoice.Number, invoice.Total, len(lineItems)))

	// Generate PDF using invoice controller
	if err := h.controller.Invoice.GenerateInvoicePDFForNotion(l, invoice, lineItems); err != nil {
		l.Error(err, "failed to generate invoice PDF")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil,
			fmt.Errorf("failed to generate PDF: %w", err), nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("PDF generated successfully: size=%d bytes", len(invoice.InvoiceFileContent)))

	// Upload PDF to Notion
	filename := fmt.Sprintf("%s.pdf", invoice.Number)
	fileUploadID, err := h.service.Notion.UploadFile(filename, "application/pdf", invoice.InvoiceFileContent)
	if err != nil {
		l.Error(err, "failed to upload PDF to notion")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil,
			fmt.Errorf("failed to upload PDF: %w", err), nil, ""))
		return
	}

	l.Info(fmt.Sprintf("PDF uploaded to notion successfully: fileUploadID=%s", fileUploadID))

	// Attach PDF to Notion page's Preview property
	// The go-notion library doesn't support file_upload type, so we use a custom method
	// that handles the raw JSON marshaling for file_upload type attachments
	if err := h.service.Notion.UpdatePagePropertiesWithFileUpload(pageID, "Preview", fileUploadID, filename); err != nil {
		l.Error(err, "failed to attach PDF to notion page")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil,
			fmt.Errorf("failed to attach PDF to page: %w", err), nil, ""))
		return
	}

	l.Info(fmt.Sprintf("invoice PDF generated and attached successfully: page_id=%s, invoice=%s, file_id=%s",
		pageID, invoice.Number, fileUploadID))

	c.JSON(http.StatusOK, view.CreateResponse[any](gin.H{
		"status":         "success",
		"message":        "Invoice PDF generated and attached successfully",
		"page_id":        pageID,
		"invoice_number": invoice.Number,
		"file_upload_id": fileUploadID,
	}, nil, nil, nil, ""))
}

// HandleNotionInvoiceSend handles invoice sending webhook events from Notion
// This webhook is triggered when the "Send invoice" button is clicked in Notion
// It downloads the PDF from Notion Preview, uploads to Google Drive, sends email, and updates Status
func (h *handler) HandleNotionInvoiceSend(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "webhook",
		"method":  "HandleNotionInvoiceSend",
	})

	l.Debug("received notion invoice send webhook request")

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		l.Error(err, "failed to read notion invoice send webhook body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("received webhook body: %s", string(body)))

	// Parse payload
	var payload NotionInvoiceWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		l.Error(err, "failed to parse notion invoice send webhook payload")
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
	l.Info(fmt.Sprintf("processing invoice send for page: %s", pageID))

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

	// Validate Type property - must be "Invoice", not "Line Item"
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

	if pageType != "Invoice" {
		l.Info(fmt.Sprintf("rejecting invoice send for non-invoice type: %s", pageType))
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil,
			fmt.Errorf("invoice send only supported for Invoice records, got: %s", pageType), nil, ""))
		return
	}

	// Extract invoice data from Notion properties
	invoice, lineItems, err := h.extractInvoiceDataFromNotion(l, page, props)
	if err != nil {
		l.Error(err, "failed to extract invoice data from notion")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil,
			fmt.Errorf("failed to extract invoice data: %w", err), nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("invoice data extracted: number=%s, total=%.2f, items=%d", invoice.Number, invoice.Total, len(lineItems)))

	// Extract Recipients from rollup property
	recipients, err := h.extractRecipientsFromNotion(l, props)
	if err != nil {
		l.Error(err, "failed to extract recipients from notion")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil,
			fmt.Errorf("failed to extract recipients: %w", err), nil, ""))
		return
	}

	if len(recipients) == 0 {
		l.Error(errors.New("no recipients found"), "recipients rollup is empty")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil,
			errors.New("no recipients found in Recipients rollup"), nil, ""))
		return
	}

	l.Debug(fmt.Sprintf("extracted %d recipients: %v", len(recipients), recipients))

	// Map recipients to invoice Email and CC
	invoice.Email = recipients[0]
	if len(recipients) > 1 {
		ccList := recipients[1:]
		ccJSON, err := json.Marshal(ccList)
		if err != nil {
			l.Error(err, "failed to marshal CC list")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil,
				fmt.Errorf("failed to marshal CC list: %w", err), nil, ""))
			return
		}
		invoice.CC = ccJSON
	}

	l.Debug(fmt.Sprintf("email recipients configured: to=%s, cc=%d", invoice.Email, len(recipients)-1))

	// Download PDF from Notion Preview property
	pdfBytes, err := h.downloadPDFFromNotionAttachment(l, props)
	if err != nil {
		l.Error(err, "failed to download PDF from notion attachment")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil,
			fmt.Errorf("failed to download PDF from attachment: %w", err), nil, ""))
		return
	}

	invoice.InvoiceFileContent = pdfBytes
	l.Debug(fmt.Sprintf("PDF downloaded from notion: size=%d bytes", len(invoice.InvoiceFileContent)))

	// Upload PDF to Google Drive
	if err := h.service.GoogleDrive.UploadInvoicePDF(invoice, "Sent"); err != nil {
		l.Error(err, "failed to upload PDF to google drive")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil,
			fmt.Errorf("failed to upload to Google Drive: %w", err), nil, ""))
		return
	}

	l.Info(fmt.Sprintf("PDF uploaded to Google Drive successfully: invoice=%s", invoice.Number))

	// Send invoice email
	threadID, err := h.service.GoogleMail.SendInvoiceMail(invoice)
	if err != nil {
		l.Error(err, "failed to send invoice email")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil,
			fmt.Errorf("failed to send email: %w", err), nil, ""))
		return
	}

	l.Info(fmt.Sprintf("invoice email sent successfully: invoice=%s, thread_id=%s", invoice.Number, threadID))

	// Update Status to "Sent" in Notion
	// Note: We don't update "Sent by" because button automations run as bots and People properties can't be set to bots
	l.Debug("updating invoice status to Sent in Notion")

	updatePayload := map[string]interface{}{
		"properties": map[string]interface{}{
			"Status": map[string]interface{}{
				"status": map[string]string{
					"name": "Sent",
				},
			},
			"Thread ID": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"type": "text",
						"text": map[string]string{
							"content": threadID,
						},
					},
				},
			},
		},
	}

	l.Debugf("storing Thread ID in Notion: %s", threadID)

	payloadBytes, err := json.Marshal(updatePayload)
	if err != nil {
		l.Error(err, "failed to marshal update payload")
		l.Warn(fmt.Sprintf("invoice sent successfully but failed to update Notion fields: %v", err))
	} else {
		l.Debug(fmt.Sprintf("update payload: %s", string(payloadBytes)))

		// Create raw HTTP request to Notion API
		notionURL := fmt.Sprintf("https://api.notion.com/v1/pages/%s", pageID)
		req, err := http.NewRequest("PATCH", notionURL, bytes.NewReader(payloadBytes))
		if err != nil {
			l.Error(err, "failed to create HTTP request for Notion update")
			l.Warn(fmt.Sprintf("invoice sent successfully but failed to update Notion fields: %v", err))
		} else {
			req.Header.Set("Authorization", "Bearer "+h.config.Notion.Secret)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Notion-Version", "2022-06-28")

			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				l.Error(err, "failed to send HTTP request to Notion")
				l.Warn(fmt.Sprintf("invoice sent successfully but failed to update Notion fields: %v", err))
			} else {
				defer resp.Body.Close()
				respBody, _ := io.ReadAll(resp.Body)

				if resp.StatusCode != http.StatusOK {
					l.Error(fmt.Errorf("notion update failed with status %d", resp.StatusCode),
						fmt.Sprintf("response body: %s", string(respBody)))
					l.Warn(fmt.Sprintf("invoice sent successfully but failed to update Notion Status: status=%d", resp.StatusCode))
				} else {
					l.Debug("updated Status to Sent in Notion successfully")
				}
			}
		}
	}

	// Update line items status to "Sent"
	if lineItemProp, ok := props["Line Item"]; ok && lineItemProp.Relation != nil {
		l.Debug(fmt.Sprintf("updating %d line items status to Sent", len(lineItemProp.Relation)))

		lineItemUpdatePayload := map[string]interface{}{
			"properties": map[string]interface{}{
				"Status": map[string]interface{}{
					"status": map[string]string{
						"name": "Sent",
					},
				},
			},
		}

		lineItemPayloadBytes, err := json.Marshal(lineItemUpdatePayload)
		if err != nil {
			l.Error(err, "failed to marshal line item update payload")
		} else {
			for _, rel := range lineItemProp.Relation {
				lineItemID := rel.ID
				l.Debug(fmt.Sprintf("updating line item %s status to Sent", lineItemID))

				notionURL := fmt.Sprintf("https://api.notion.com/v1/pages/%s", lineItemID)
				req, err := http.NewRequest("PATCH", notionURL, bytes.NewReader(lineItemPayloadBytes))
				if err != nil {
					l.Error(err, fmt.Sprintf("failed to create HTTP request for line item %s", lineItemID))
					continue
				}

				req.Header.Set("Authorization", "Bearer "+h.config.Notion.Secret)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Notion-Version", "2022-06-28")

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					l.Error(err, fmt.Sprintf("failed to update line item %s status", lineItemID))
					continue
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					respBody, _ := io.ReadAll(resp.Body)
					l.Error(fmt.Errorf("notion update failed with status %d", resp.StatusCode),
						fmt.Sprintf("failed to update line item %s: %s", lineItemID, string(respBody)))
				} else {
					l.Debug(fmt.Sprintf("updated line item %s status to Sent", lineItemID))
				}
			}
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](gin.H{
		"status":         "success",
		"message":        "Invoice sent successfully",
		"page_id":        pageID,
		"invoice_number": invoice.Number,
		"recipients":     recipients,
		"thread_id":      threadID,
	}, nil, nil, nil, ""))
}
