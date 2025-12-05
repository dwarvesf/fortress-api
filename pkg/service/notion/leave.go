package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

// LeaveRequest represents a leave request from Notion
type LeaveRequest struct {
	PageID       string
	Reason       string     // Title field
	EmployeeID   string     // Relation page ID
	Email        string     // Rollup value from Employee relation
	LeaveType    string     // "Off" or "Remote"
	StartDate    *time.Time
	EndDate      *time.Time
	Shift        string     // "Full day", "Morning", "Afternoon"
	Status       string     // "Pending", "Approved", "Rejected"
	ApprovedByID string     // Relation page ID
	ApprovedAt   *time.Time
	Assignees    []string   // Notion user emails from People property
}

// LeaveService handles leave request operations with Notion
type LeaveService struct {
	client *nt.Client
	cfg    *config.Config
	store  *store.Store
	repo   store.DBRepo
	logger logger.Logger
}

// NewLeaveService creates a new Notion leave service
func NewLeaveService(cfg *config.Config, store *store.Store, repo store.DBRepo, logger logger.Logger) *LeaveService {
	if cfg.Notion.Secret == "" {
		logger.Error(errors.New("notion secret not configured"), "notion secret is empty")
		return nil
	}

	logger.Debug("creating new LeaveService")

	return &LeaveService{
		client: nt.NewClient(cfg.Notion.Secret),
		cfg:    cfg,
		store:  store,
		repo:   repo,
		logger: logger,
	}
}

// GetLeaveRequest fetches a leave request by page ID from Notion
func (s *LeaveService) GetLeaveRequest(ctx context.Context, pageID string) (*LeaveRequest, error) {
	if s.client == nil {
		return nil, errors.New("notion client is nil")
	}

	s.logger.Debug(fmt.Sprintf("fetching leave request from Notion: page_id=%s", pageID))

	page, err := s.client.FindPageByID(ctx, pageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to fetch leave request page: page_id=%s", pageID))
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		return nil, errors.New("failed to cast page properties")
	}

	// Extract all properties
	leave := &LeaveRequest{
		PageID:    pageID,
		Reason:    s.extractTitle(props, "Reason"),
		LeaveType: s.extractSelect(props, "Leave Type"),
		Shift:     s.extractSelect(props, "Shift"),
		Status:    s.extractSelect(props, "Status"),
	}

	// Extract dates
	leave.StartDate = s.extractDate(props, "Start Date")
	leave.EndDate = s.extractDate(props, "End Date")
	leave.ApprovedAt = s.extractDate(props, "Approved at")

	// Extract email from rollup
	leave.Email = s.extractRollupEmail(props, "Email")

	// Extract relation IDs
	leave.EmployeeID = s.extractFirstRelationID(props, "Employee")
	leave.ApprovedByID = s.extractFirstRelationID(props, "Approved By")

	// Extract assignees from Relation property (links to contractor pages)
	leave.Assignees = s.extractAssigneeEmails(ctx, props, "Assignees")

	s.logger.Debug(fmt.Sprintf("fetched leave request: page_id=%s reason=%s email=%s status=%s leave_type=%s assignees=%v",
		pageID, leave.Reason, leave.Email, leave.Status, leave.LeaveType, leave.Assignees))

	return leave, nil
}

// UpdateLeaveStatus updates the status and approval fields of a leave request in Notion
func (s *LeaveService) UpdateLeaveStatus(ctx context.Context, pageID, status, approverPageID string) error {
	if s.client == nil {
		return errors.New("notion client is nil")
	}

	s.logger.Debug(fmt.Sprintf("updating leave status in Notion: page_id=%s status=%s approver_page_id=%s",
		pageID, status, approverPageID))

	// Build update params
	updateParams := nt.UpdatePageParams{
		DatabasePageProperties: nt.DatabasePageProperties{
			"Status": nt.DatabasePageProperty{
				Select: &nt.SelectOptions{
					Name: status,
				},
			},
		},
	}

	// If approving, also set Approved By and Approved at
	if status == "Approved" && approverPageID != "" {
		updateParams.DatabasePageProperties["Approved By"] = nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: approverPageID},
			},
		}
		now := time.Now()
		updateParams.DatabasePageProperties["Approved at"] = nt.DatabasePageProperty{
			Date: &nt.Date{
				Start: nt.NewDateTime(now, false),
			},
		}
	}

	_, err := s.client.UpdatePage(ctx, pageID, updateParams)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to update leave status in Notion: page_id=%s", pageID))
		return fmt.Errorf("failed to update page: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("successfully updated leave status in Notion: page_id=%s status=%s", pageID, status))
	return nil
}

// GetContractorPageIDByEmail looks up a contractor page ID by email
func (s *LeaveService) GetContractorPageIDByEmail(ctx context.Context, email string) (string, error) {
	if s.client == nil {
		return "", errors.New("notion client is nil")
	}

	contractorDBID := s.cfg.LeaveIntegration.Notion.ContractorDBID
	if contractorDBID == "" {
		return "", errors.New("contractor database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("looking up contractor by email: email=%s db_id=%s", email, contractorDBID))

	// Query contractor database for matching email
	filter := &nt.DatabaseQueryFilter{
		Property: "Team Email",
		DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
			Email: &nt.TextPropertyFilter{
				Equals: email,
			},
		},
	}

	resp, err := s.client.QueryDatabase(ctx, contractorDBID, &nt.DatabaseQuery{
		Filter:   filter,
		PageSize: 1,
	})
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to query contractor database: email=%s", email))
		return "", fmt.Errorf("failed to query contractor database: %w", err)
	}

	if len(resp.Results) == 0 {
		s.logger.Debug(fmt.Sprintf("no contractor found for email: %s", email))
		return "", fmt.Errorf("contractor not found for email: %s", email)
	}

	pageID := resp.Results[0].ID
	s.logger.Debug(fmt.Sprintf("found contractor page: email=%s page_id=%s", email, pageID))
	return pageID, nil
}

// QueryPendingLeaveRequests fetches all pending leave requests from Notion
func (s *LeaveService) QueryPendingLeaveRequests(ctx context.Context) ([]LeaveRequest, error) {
	if s.client == nil {
		return nil, errors.New("notion client is nil")
	}

	dataSourceID := s.cfg.LeaveIntegration.Notion.DataSourceID
	if dataSourceID == "" {
		return nil, errors.New("leave data source ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("querying pending leave requests from data source: %s", dataSourceID))

	// Query for pending leave requests
	filter := &nt.DatabaseQueryFilter{
		Property: "Status",
		DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
			Select: &nt.SelectDatabaseQueryFilter{
				Equals: "Pending",
			},
		},
	}

	pages, err := s.queryDataSource(ctx, dataSourceID, filter)
	if err != nil {
		return nil, err
	}

	// Transform pages to LeaveRequest
	var requests []LeaveRequest
	for _, page := range pages {
		props, ok := page.Properties.(nt.DatabasePageProperties)
		if !ok {
			continue
		}

		leave := LeaveRequest{
			PageID:    page.ID,
			Reason:    s.extractTitle(props, "Reason"),
			LeaveType: s.extractSelect(props, "Leave Type"),
			Shift:     s.extractSelect(props, "Shift"),
			Status:    s.extractSelect(props, "Status"),
			StartDate: s.extractDate(props, "Start Date"),
			EndDate:   s.extractDate(props, "End Date"),
			Email:     s.extractRollupEmail(props, "Email"),
		}
		requests = append(requests, leave)
	}

	s.logger.Debug(fmt.Sprintf("found %d pending leave requests", len(requests)))
	return requests, nil
}

// queryDataSource queries a Notion data source directly using HTTP
// This is required for multi-source databases that aren't supported by go-notion
func (s *LeaveService) queryDataSource(ctx context.Context, dataSourceID string, filter *nt.DatabaseQueryFilter) ([]nt.Page, error) {
	var allPages []nt.Page
	var cursor string

	for {
		reqBody := DataSourceQueryRequest{
			Filter:   filter,
			PageSize: 100,
		}
		if cursor != "" {
			reqBody.StartCursor = cursor
		}

		resp, err := s.executeDataSourceQuery(ctx, dataSourceID, reqBody)
		if err != nil {
			return nil, err
		}

		// Convert DataSourcePage to nt.Page
		for _, dsPage := range resp.Results {
			page := s.convertDataSourcePageToPage(dsPage)
			allPages = append(allPages, page)
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}
		cursor = *resp.NextCursor
	}

	return allPages, nil
}

// convertDataSourcePageToPage converts a DataSourcePage to nt.Page
func (s *LeaveService) convertDataSourcePageToPage(dsPage DataSourcePage) nt.Page {
	return nt.Page{
		ID:         dsPage.ID,
		URL:        dsPage.URL,
		Archived:   dsPage.Archived,
		Properties: dsPage.Properties,
	}
}

// executeDataSourceQuery makes the HTTP request to query a data source
func (s *LeaveService) executeDataSourceQuery(ctx context.Context, dataSourceID string, reqBody DataSourceQueryRequest) (*DataSourceQueryResponse, error) {
	// Normalize data source ID - remove hyphens if present
	normalizedID := strings.ReplaceAll(dataSourceID, "-", "")
	url := fmt.Sprintf("https://api.notion.com/v1/data_sources/%s/query", normalizedID)
	s.logger.Debug(fmt.Sprintf("data source query URL: %s", url))

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("querying data source %s with body: %s", dataSourceID, string(jsonBody)))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.cfg.Notion.Secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2025-09-03") // Required for data source queries

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error(err, "failed to execute data source query")
		return nil, fmt.Errorf("data source query failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Error(errors.New(string(body)), fmt.Sprintf("data source query returned status %d", resp.StatusCode))
		return nil, fmt.Errorf("data source query failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result DataSourceQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// Property extraction helpers

// extractTitle extracts a title property value
func (s *LeaveService) extractTitle(props nt.DatabasePageProperties, propName string) string {
	if prop, ok := props[propName]; ok && len(prop.Title) > 0 {
		var parts []string
		for _, rt := range prop.Title {
			parts = append(parts, rt.PlainText)
		}
		return strings.Join(parts, "")
	}
	return ""
}

// extractSelect extracts a select property value
func (s *LeaveService) extractSelect(props nt.DatabasePageProperties, propName string) string {
	if prop, ok := props[propName]; ok && prop.Select != nil {
		return prop.Select.Name
	}
	return ""
}

// extractDate extracts a date property value
func (s *LeaveService) extractDate(props nt.DatabasePageProperties, propName string) *time.Time {
	if prop, ok := props[propName]; ok && prop.Date != nil {
		t := prop.Date.Start.Time
		if !t.IsZero() {
			return &t
		}
	}
	return nil
}

// extractRollupEmail extracts email from a rollup property
func (s *LeaveService) extractRollupEmail(props nt.DatabasePageProperties, propName string) string {
	if prop, ok := props[propName]; ok {
		// Check if it's a rollup type with array results
		if prop.Rollup != nil && len(prop.Rollup.Array) > 0 {
			for _, result := range prop.Rollup.Array {
				if result.Email != nil && *result.Email != "" {
					return *result.Email
				}
			}
		}
	}
	return ""
}

// extractFirstRelationID extracts the first relation page ID
func (s *LeaveService) extractFirstRelationID(props nt.DatabasePageProperties, propName string) string {
	if prop, ok := props[propName]; ok && len(prop.Relation) > 0 {
		return prop.Relation[0].ID
	}
	return ""
}

// extractAssigneeEmails extracts email addresses from a Relation property by fetching each contractor page
func (s *LeaveService) extractAssigneeEmails(ctx context.Context, props nt.DatabasePageProperties, propName string) []string {
	var emails []string
	if prop, ok := props[propName]; ok {
		s.logger.Debug(fmt.Sprintf("extractAssigneeEmails: found property %s with %d relations", propName, len(prop.Relation)))
		for i, rel := range prop.Relation {
			s.logger.Debug(fmt.Sprintf("extractAssigneeEmails: relation[%d] id=%s", i, rel.ID))
			// Fetch contractor page to get email
			email, err := s.getContractorEmailByPageID(ctx, rel.ID)
			if err != nil {
				s.logger.Debug(fmt.Sprintf("extractAssigneeEmails: failed to get email for relation[%d] id=%s: %v", i, rel.ID, err))
				continue
			}
			if email != "" {
				s.logger.Debug(fmt.Sprintf("extractAssigneeEmails: found email for relation[%d]: %s", i, email))
				emails = append(emails, email)
			}
		}
	} else {
		s.logger.Debug(fmt.Sprintf("extractAssigneeEmails: property %s not found in props", propName))
	}
	return emails
}

// getContractorEmailByPageID fetches the email from a contractor page
func (s *LeaveService) getContractorEmailByPageID(ctx context.Context, pageID string) (string, error) {
	page, err := s.client.FindPageByID(ctx, pageID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch contractor page: %w", err)
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		return "", errors.New("failed to cast page properties")
	}

	// Try "Team Email" property first (email type)
	if prop, ok := props["Team Email"]; ok && prop.Email != nil && *prop.Email != "" {
		return *prop.Email, nil
	}

	return "", nil
}

