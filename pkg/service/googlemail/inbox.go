package googlemail

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"

	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// ListInboxMessages lists unread messages matching the query
func (g *googleService) ListInboxMessages(ctx context.Context, query string, maxResults int64) ([]InboxMessage, error) {
	l := logger.NewLogrusLogger("debug").Fields(logger.Fields{
		"service": "googlemail",
		"method":  "ListInboxMessages",
		"query":   query,
	})
	l.Debug("listing inbox messages")

	if g.service == nil {
		l.Debug("gmail service not initialized, preparing service")
		if err := g.ensureToken(g.appConfig.InvoiceListener.RefreshToken); err != nil {
			l.Error(err, "failed to ensure token")
			return nil, fmt.Errorf("failed to ensure token: %w", err)
		}
		if err := g.prepareService(); err != nil {
			l.Error(err, "failed to prepare service")
			return nil, fmt.Errorf("failed to prepare service: %w", err)
		}
	}

	call := g.service.Users.Messages.List("me").Context(ctx).Q(query)
	if maxResults > 0 {
		call = call.MaxResults(maxResults)
	}

	resp, err := call.Do()
	if err != nil {
		l.Error(err, "failed to list messages")
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	l.Debugf("found %d messages", len(resp.Messages))

	var messages []InboxMessage
	for _, msg := range resp.Messages {
		messages = append(messages, InboxMessage{
			ID:       msg.Id,
			ThreadID: msg.ThreadId,
		})
	}

	return messages, nil
}

// GetMessage retrieves a full message with headers and attachment info
func (g *googleService) GetMessage(ctx context.Context, messageID string) (*InboxMessage, error) {
	l := logger.NewLogrusLogger("debug").Fields(logger.Fields{
		"service":   "googlemail",
		"method":    "GetMessage",
		"messageID": messageID,
	})
	l.Debug("getting message")

	if g.service == nil {
		l.Debug("gmail service not initialized, preparing service")
		if err := g.ensureToken(g.appConfig.InvoiceListener.RefreshToken); err != nil {
			l.Error(err, "failed to ensure token")
			return nil, fmt.Errorf("failed to ensure token: %w", err)
		}
		if err := g.prepareService(); err != nil {
			l.Error(err, "failed to prepare service")
			return nil, fmt.Errorf("failed to prepare service: %w", err)
		}
	}

	msg, err := g.service.Users.Messages.Get("me", messageID).Context(ctx).Format("full").Do()
	if err != nil {
		l.Error(err, "failed to get message")
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	inboxMsg := &InboxMessage{
		ID:       msg.Id,
		ThreadID: msg.ThreadId,
		Snippet:  msg.Snippet,
		LabelIDs: msg.LabelIds,
	}

	// Parse headers
	for _, header := range msg.Payload.Headers {
		switch strings.ToLower(header.Name) {
		case "subject":
			inboxMsg.Subject = header.Value
		case "from":
			inboxMsg.From = header.Value
		case "to":
			inboxMsg.To = header.Value
		case "date":
			if t, err := time.Parse(time.RFC1123Z, header.Value); err == nil {
				inboxMsg.Date = t
			} else if t, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", header.Value); err == nil {
				inboxMsg.Date = t
			}
		}
	}

	// Find PDF attachment
	inboxMsg.PDFPartID = findPDFAttachment(msg.Payload)
	inboxMsg.HasPDF = inboxMsg.PDFPartID != ""

	l.Debugf("message subject: %s, hasPDF: %v", inboxMsg.Subject, inboxMsg.HasPDF)

	return inboxMsg, nil
}

// findPDFAttachment recursively searches for PDF attachments in message parts
func findPDFAttachment(part *gmail.MessagePart) string {
	if part == nil {
		return ""
	}

	// Check if this part is a PDF
	if part.MimeType == "application/pdf" && part.Body != nil && part.Body.AttachmentId != "" {
		return part.Body.AttachmentId
	}

	// Check filename for PDF extension
	if part.Filename != "" && strings.HasSuffix(strings.ToLower(part.Filename), ".pdf") {
		if part.Body != nil && part.Body.AttachmentId != "" {
			return part.Body.AttachmentId
		}
	}

	// Recursively check child parts
	for _, child := range part.Parts {
		if attachmentID := findPDFAttachment(child); attachmentID != "" {
			return attachmentID
		}
	}

	return ""
}

// GetAttachment retrieves the raw bytes of an attachment
func (g *googleService) GetAttachment(ctx context.Context, messageID, attachmentID string) ([]byte, error) {
	l := logger.NewLogrusLogger("debug").Fields(logger.Fields{
		"service":      "googlemail",
		"method":       "GetAttachment",
		"messageID":    messageID,
		"attachmentID": attachmentID,
	})
	l.Debug("getting attachment")

	if g.service == nil {
		l.Debug("gmail service not initialized, preparing service")
		if err := g.ensureToken(g.appConfig.InvoiceListener.RefreshToken); err != nil {
			l.Error(err, "failed to ensure token")
			return nil, fmt.Errorf("failed to ensure token: %w", err)
		}
		if err := g.prepareService(); err != nil {
			l.Error(err, "failed to prepare service")
			return nil, fmt.Errorf("failed to prepare service: %w", err)
		}
	}

	attachment, err := g.service.Users.Messages.Attachments.Get("me", messageID, attachmentID).Context(ctx).Do()
	if err != nil {
		l.Error(err, "failed to get attachment")
		return nil, fmt.Errorf("failed to get attachment: %w", err)
	}

	// Gmail API returns base64url encoded data
	data, err := base64.URLEncoding.DecodeString(attachment.Data)
	if err != nil {
		// Try standard base64 as fallback
		data, err = base64.StdEncoding.DecodeString(attachment.Data)
		if err != nil {
			l.Error(err, "failed to decode attachment data")
			return nil, fmt.Errorf("failed to decode attachment data: %w", err)
		}
	}

	l.Debugf("attachment size: %d bytes", len(data))

	return data, nil
}

// AddLabel adds a label to a message
func (g *googleService) AddLabel(ctx context.Context, messageID, labelID string) error {
	l := logger.NewLogrusLogger("debug").Fields(logger.Fields{
		"service":   "googlemail",
		"method":    "AddLabel",
		"messageID": messageID,
		"labelID":   labelID,
	})
	l.Debug("adding label to message")

	if g.service == nil {
		l.Debug("gmail service not initialized, preparing service")
		if err := g.ensureToken(g.appConfig.InvoiceListener.RefreshToken); err != nil {
			l.Error(err, "failed to ensure token")
			return fmt.Errorf("failed to ensure token: %w", err)
		}
		if err := g.prepareService(); err != nil {
			l.Error(err, "failed to prepare service")
			return fmt.Errorf("failed to prepare service: %w", err)
		}
	}

	modifyRequest := &gmail.ModifyMessageRequest{
		AddLabelIds: []string{labelID},
	}

	_, err := g.service.Users.Messages.Modify("me", messageID, modifyRequest).Context(ctx).Do()
	if err != nil {
		l.Error(err, "failed to add label")
		return fmt.Errorf("failed to add label: %w", err)
	}

	l.Debug("label added successfully")

	return nil
}

// GetOrCreateLabel gets an existing label or creates a new one
func (g *googleService) GetOrCreateLabel(ctx context.Context, labelName string) (string, error) {
	l := logger.NewLogrusLogger("debug").Fields(logger.Fields{
		"service":   "googlemail",
		"method":    "GetOrCreateLabel",
		"labelName": labelName,
	})
	l.Debug("getting or creating label")

	if g.service == nil {
		l.Debug("gmail service not initialized, preparing service")
		if err := g.ensureToken(g.appConfig.InvoiceListener.RefreshToken); err != nil {
			l.Error(err, "failed to ensure token")
			return "", fmt.Errorf("failed to ensure token: %w", err)
		}
		if err := g.prepareService(); err != nil {
			l.Error(err, "failed to prepare service")
			return "", fmt.Errorf("failed to prepare service: %w", err)
		}
	}

	// List existing labels
	labels, err := g.service.Users.Labels.List("me").Context(ctx).Do()
	if err != nil {
		l.Error(err, "failed to list labels")
		return "", fmt.Errorf("failed to list labels: %w", err)
	}

	// Check if label already exists
	for _, label := range labels.Labels {
		if label.Name == labelName {
			l.Debugf("found existing label with ID: %s", label.Id)
			return label.Id, nil
		}
	}

	// Create new label
	newLabel := &gmail.Label{
		Name:                  labelName,
		LabelListVisibility:   "labelShow",
		MessageListVisibility: "show",
	}

	created, err := g.service.Users.Labels.Create("me", newLabel).Context(ctx).Do()
	if err != nil {
		l.Error(err, "failed to create label")
		return "", fmt.Errorf("failed to create label: %w", err)
	}

	l.Debugf("created new label with ID: %s", created.Id)

	return created.Id, nil
}
