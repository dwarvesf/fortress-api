package nocodb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

// LeaveService handles leave request operations in NocoDB
type LeaveService struct {
	client *Service
	cfg    *config.Config
	store  *store.Store
	repo   store.DBRepo
	logger logger.Logger
}

// NewLeaveService creates a new NocoDB leave service
func NewLeaveService(client *Service, cfg *config.Config, store *store.Store, repo store.DBRepo, logger logger.Logger) *LeaveService {
	return &LeaveService{
		client: client,
		cfg:    cfg,
		store:  store,
		repo:   repo,
		logger: logger,
	}
}

// UpdateLeaveStatus updates the status of a leave request in NocoDB
func (l *LeaveService) UpdateLeaveStatus(leaveID int, status string, approvedBy string) error {
	if l.client == nil {
		return errors.New("nocodb client is nil")
	}

	tableID := l.cfg.LeaveIntegration.Noco.TableID
	if tableID == "" {
		l.logger.Error(errors.New("leave table id not configured"), "nocodb leave table id is empty")
		return errors.New("nocodb leave table id not configured")
	}

	l.logger.Debugf("Updating leave request %d status to %s in NocoDB table %s", leaveID, status, tableID)

	ctx := context.Background()
	path := fmt.Sprintf("/tables/%s/records", tableID)

	// Build PATCH payload
	payload := map[string]interface{}{
		"Id":     leaveID,
		"status": status,
	}

	// Add approved_by if provided
	if approvedBy != "" {
		payload["approved_by"] = approvedBy
	}

	payloadJSON, _ := json.Marshal(payload)
	l.logger.Debugf("Sending PATCH request to NocoDB: %s with payload: %s", path, string(payloadJSON))

	resp, err := l.client.makeRequest(ctx, http.MethodPatch, path, nil, payload)
	if err != nil {
		l.logger.Errorf(err, "failed to update leave request %d status in nocodb", leaveID)
		return fmt.Errorf("nocodb update leave status failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	l.logger.Debugf("NocoDB PATCH response status: %s, body: %s", resp.Status, string(body))

	if resp.StatusCode >= 300 {
		l.logger.Errorf(fmt.Errorf("nocodb returned status %s", resp.Status), "nocodb update leave %d status failed: %s", leaveID, string(body))
		return fmt.Errorf("nocodb update leave %d status failed: %s - %s", leaveID, resp.Status, string(body))
	}

	l.logger.Debugf("Successfully updated leave request %d status to %s in NocoDB", leaveID, status)
	return nil
}

// GetLeaveAssigneeEmails fetches the assignee emails for a leave request from NocoDB
func (l *LeaveService) GetLeaveAssigneeEmails(leaveID int) ([]string, error) {
	if l.client == nil {
		return nil, errors.New("nocodb client is nil")
	}

	tableID := l.cfg.LeaveIntegration.Noco.TableID
	if tableID == "" {
		l.logger.Error(errors.New("leave table id not configured"), "nocodb leave table id is empty")
		return nil, errors.New("nocodb leave table id not configured")
	}

	l.logger.Debugf("Fetching assignees for leave request %d from NocoDB table %s", leaveID, tableID)

	ctx := context.Background()

	// Fetch the leave request with linked employees
	path := fmt.Sprintf("/tables/%s/records/%d", tableID, leaveID)
	query := url.Values{}

	resp, err := l.client.makeRequest(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		l.logger.Errorf(err, "failed to fetch leave request %d from nocodb", leaveID)
		return nil, fmt.Errorf("nocodb get leave request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	l.logger.Debugf("NocoDB GET response status: %s, body: %s", resp.Status, string(body))

	if resp.StatusCode >= 300 {
		l.logger.Errorf(fmt.Errorf("nocodb returned status %s", resp.Status), "nocodb get leave %d failed: %s", leaveID, string(body))
		return nil, fmt.Errorf("nocodb get leave %d failed: %s - %s", leaveID, resp.Status, string(body))
	}

	// Parse response to get linked employees
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		l.logger.Errorf(err, "failed to parse leave request response")
		return nil, fmt.Errorf("failed to parse nocodb response: %w", err)
	}

	// Try to get linked employees from the response
	// NocoDB returns linked records in a field like "nc_m2m_<table>_<linked_table>" or custom field name
	var emails []string

	// Check for the m2m linked field
	for key, value := range result {
		l.logger.Debugf("checking field %s for assignees", key)

		// Look for the _nc_m2m_leave_requests_nc_employees field
		if arr, ok := value.([]interface{}); ok {
			for _, item := range arr {
				if record, ok := item.(map[string]interface{}); ok {
					// Check nested nc_employees object (NocoDB naming convention)
					if ncEmployees, ok := record["nc_employees"].(map[string]interface{}); ok {
						if email, ok := ncEmployees["email"].(string); ok && email != "" {
							l.logger.Debugf("found assignee email from nc_employees.email: %s", email)
							emails = append(emails, email)
						}
					}
					// Also check employees object as fallback
					if employees, ok := record["employees"].(map[string]interface{}); ok {
						if email, ok := employees["email"].(string); ok && email != "" {
							l.logger.Debugf("found assignee email from employees.email: %s", email)
							emails = append(emails, email)
						}
					}
					// Also check direct fields as fallback
					if email, ok := record["email"].(string); ok && email != "" {
						l.logger.Debugf("found assignee email from email: %s", email)
						emails = append(emails, email)
					}
				}
			}
		}
	}

	l.logger.Debugf("Found %d assignee emails for leave request %d", len(emails), leaveID)
	return emails, nil
}
