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

// LeaveRequest represents a leave request from Notion Unavailability Notices database
type LeaveRequest struct {
	PageID              string
	LeaveRequestTitle   string     // Title field - leave request name
	EmployeeID          string     // Relation page ID - contractor
	Email               string     // Rollup value from Contractor relation (Team Email)
	UnavailabilityType  string     // "Personal Time", "Health / Illness", "Family / Emergency", "Travel / Vacation", "Other"
	StartDate           *time.Time
	EndDate             *time.Time
	Status              string     // "New", "Acknowledged", "Not Applicable", "Withdrawn"
	ReviewedByID        string     // Relation page ID - person who reviewed/approved
	DateApproved        *time.Time
	DateRequested       *time.Time // When the request was submitted
	AdditionalContext   string     // Detailed reason for leave
	Assignees           []string   // Notion user emails from People property
}

// LeaveService handles leave request operations with Notion
type LeaveService struct {
	*baseService
	store *store.Store
	repo  store.DBRepo
}

// NewLeaveService creates a new Notion leave service
func NewLeaveService(cfg *config.Config, store *store.Store, repo store.DBRepo, l logger.Logger) *LeaveService {
	base := newBaseService(cfg, l)
	if base == nil {
		return nil
	}

	l.Debug("creating new LeaveService")

	return &LeaveService{
		baseService: base,
		store:       store,
		repo:        repo,
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

	// Extract all properties from Unavailability Notices schema
	leave := &LeaveRequest{
		PageID:             pageID,
		LeaveRequestTitle:  ExtractTitle(props, "Leave Request"),
		UnavailabilityType: ExtractSelect(props, "Unavailability Type"),
		AdditionalContext:  ExtractRichText(props, "Additional Context"),
		Status:             ExtractSelect(props, "Status"),
	}

	// Extract dates
	leave.StartDate = ExtractDate(props, "Start Date")
	leave.EndDate = ExtractDate(props, "End Date")
	leave.DateApproved = ExtractDate(props, "Date Approved")
	leave.DateRequested = ExtractDate(props, "Date Requested")

	// Extract email from Team Email property (rollup type)
	leave.Email = ExtractEmailFromRollup(props, "Team Email")
	s.logger.Debug(fmt.Sprintf("extracted email from Team Email: %s", leave.Email))

	// Extract relation IDs - using Contractor instead of Employee
	leave.EmployeeID = ExtractFirstRelationID(props, "Contractor")
	leave.ReviewedByID = ExtractFirstRelationID(props, "Reviewed By")

	// Extract assignees from multi_select property (format: "Name (email@domain)")
	leave.Assignees = s.extractEmailsFromMultiSelect(props, "Assignees")
	s.logger.Debug(fmt.Sprintf("extracted assignees from multi_select: %v", leave.Assignees))

	s.logger.Debug(fmt.Sprintf("fetched leave request: page_id=%s title=%s email=%s status=%s unavailability_type=%s assignees=%v",
		pageID, leave.LeaveRequestTitle, leave.Email, leave.Status, leave.UnavailabilityType, leave.Assignees))

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
				Status: &nt.SelectOptions{
					Name: status,
				},
			},
		},
	}

	// If acknowledging/rejecting/withdrawing, also set Reviewed By and Date Approved
	if (status == "Acknowledged" || status == "Not Applicable" || status == "Withdrawn") && approverPageID != "" {
		updateParams.DatabasePageProperties["Reviewed By"] = nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: approverPageID},
			},
		}
		now := time.Now()
		updateParams.DatabasePageProperties["Date Approved"] = nt.DatabasePageProperty{
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
			PageID:             page.ID,
			LeaveRequestTitle:  ExtractTitle(props, "Leave Request"),
			UnavailabilityType: ExtractSelect(props, "Unavailability Type"),
			AdditionalContext:  ExtractRichText(props, "Additional Context"),
			Status:             ExtractSelect(props, "Status"),
			StartDate:          ExtractDate(props, "Start Date"),
			EndDate:            ExtractDate(props, "End Date"),
			Email:              ExtractEmail(props, "Team Email"),
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

// extractEmailsFromMultiSelect extracts emails from multi_select option names (format: "Name (email@domain)")
func (s *LeaveService) extractEmailsFromMultiSelect(props nt.DatabasePageProperties, propName string) []string {
	var emails []string
	if prop, ok := props[propName]; ok && len(prop.MultiSelect) > 0 {
		s.logger.Debug(fmt.Sprintf("extractEmailsFromMultiSelect: property %s has %d options", propName, len(prop.MultiSelect)))
		for _, opt := range prop.MultiSelect {
			s.logger.Debug(fmt.Sprintf("extractEmailsFromMultiSelect: option name: %s", opt.Name))
			email := s.parseEmailFromOptionName(opt.Name)
			if email != "" {
				emails = append(emails, email)
			}
		}
	}
	return emails
}

// parseEmailFromOptionName extracts email from option name format "Name (email@domain)"
func (s *LeaveService) parseEmailFromOptionName(optionName string) string {
	// Find the last occurrence of "(" and ")"
	start := strings.LastIndex(optionName, "(")
	end := strings.LastIndex(optionName, ")")
	if start != -1 && end != -1 && end > start {
		email := strings.TrimSpace(optionName[start+1 : end])
		if strings.Contains(email, "@") {
			return email
		}
	}
	return ""
}

// GetActiveDeploymentsForContractor queries Deployment Tracker for active deployments
// Returns empty array if none found (graceful handling)
// Returns error only on API failures
func (s *LeaveService) GetActiveDeploymentsForContractor(
	ctx context.Context,
	contractorPageID string,
) ([]nt.Page, error) {
	if contractorPageID == "" {
		s.logger.Debug("contractor page ID is empty, skipping deployment lookup")
		return []nt.Page{}, nil
	}

	s.logger.Debug(fmt.Sprintf("querying active deployments: contractor_id=%s", contractorPageID))

	filter := &nt.DatabaseQueryFilter{
		And: []nt.DatabaseQueryFilter{
			{
				Property: "Contractor",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Relation: &nt.RelationDatabaseQueryFilter{
						Contains: contractorPageID,
					},
				},
			},
			{
				Property: "Deployment Status",
				DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
					Status: &nt.StatusDatabaseQueryFilter{
						Equals: "Active",
					},
				},
			},
		},
	}

	query := &nt.DatabaseQuery{
		Filter: filter,
	}

	deploymentDBID := s.cfg.Notion.Databases.DeploymentTracker
	resp, err := s.client.QueryDatabase(ctx, deploymentDBID, query)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to query deployment tracker: contractor_id=%s", contractorPageID))
		return nil, err
	}

	s.logger.Debug(fmt.Sprintf("found %d active deployments for contractor: contractor_id=%s", len(resp.Results), contractorPageID))

	return resp.Results, nil
}

// ExtractStakeholdersFromDeployment extracts AM/DL contractor page IDs from deployment
// Fetches Project page directly for AM/DL relations (rollups are unreliable due to Notion sync issues)
// Returns unique stakeholder page IDs
func (s *LeaveService) ExtractStakeholdersFromDeployment(
	ctx context.Context,
	deploymentPage nt.Page,
) []string {
	props, ok := deploymentPage.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Error(errors.New("invalid properties type"), "failed to cast deployment properties")
		return []string{}
	}

	stakeholderMap := make(map[string]bool) // Use map for deduplication

	// Replicate "Final AM" formula logic: if(empty(Override AM), Account Managers rollup, Override AM)
	// 1. Check Override AM (relation) first
	if overrideAM, ok := props["Override AM"]; ok && len(overrideAM.Relation) > 0 {
		s.logger.Debug(fmt.Sprintf("found Override AM with %d items", len(overrideAM.Relation)))
		for _, rel := range overrideAM.Relation {
			s.logger.Debug(fmt.Sprintf("adding AM from Override AM: %s", rel.ID))
			stakeholderMap[rel.ID] = true
		}
	} else {
		// 2. Fallback to Account Managers rollup
		if amRollup, ok := props["Account Managers"]; ok && amRollup.Rollup != nil && amRollup.Rollup.Array != nil {
			s.logger.Debug(fmt.Sprintf("found Account Managers rollup with %d items", len(amRollup.Rollup.Array)))
			for _, item := range amRollup.Rollup.Array {
				if len(item.Relation) > 0 {
					for _, rel := range item.Relation {
						s.logger.Debug(fmt.Sprintf("adding AM from Account Managers rollup: %s", rel.ID))
						stakeholderMap[rel.ID] = true
					}
				}
			}
		} else {
			s.logger.Debug("no Override AM or Account Managers rollup found")
		}
	}

	// Replicate "Final Delivery Lead" formula logic: if(empty(Override DL), Delivery Leads rollup, Override DL)
	// 1. Check Override DL (relation) first
	if overrideDL, ok := props["Override DL"]; ok && len(overrideDL.Relation) > 0 {
		s.logger.Debug(fmt.Sprintf("found Override DL with %d items", len(overrideDL.Relation)))
		for _, rel := range overrideDL.Relation {
			s.logger.Debug(fmt.Sprintf("adding DL from Override DL: %s", rel.ID))
			stakeholderMap[rel.ID] = true
		}
	} else {
		// 2. Fallback to Delivery Leads rollup
		if dlRollup, ok := props["Delivery Leads"]; ok && dlRollup.Rollup != nil && dlRollup.Rollup.Array != nil {
			s.logger.Debug(fmt.Sprintf("found Delivery Leads rollup with %d items", len(dlRollup.Rollup.Array)))
			for _, item := range dlRollup.Rollup.Array {
				if len(item.Relation) > 0 {
					for _, rel := range item.Relation {
						s.logger.Debug(fmt.Sprintf("adding DL from Delivery Leads rollup: %s", rel.ID))
						stakeholderMap[rel.ID] = true
					}
				}
			}
		} else {
			s.logger.Debug("no Override DL or Delivery Leads rollup found")
		}
	}

	// Convert map to slice
	stakeholders := make([]string, 0, len(stakeholderMap))
	for id := range stakeholderMap {
		stakeholders = append(stakeholders, id)
	}

	s.logger.Debug(fmt.Sprintf("extracted %d unique stakeholders from Deployment page", len(stakeholders)))

	return stakeholders
}

// GetDiscordUsernameFromContractor fetches Discord username from contractor page
// Returns empty string if Discord field not set (graceful handling)
// Returns error only on API failures
func (s *LeaveService) GetDiscordUsernameFromContractor(
	ctx context.Context,
	contractorPageID string,
) (string, error) {
	if contractorPageID == "" {
		s.logger.Debug("contractor page ID is empty")
		return "", nil
	}

	s.logger.Debug(fmt.Sprintf("fetching contractor Discord username: page_id=%s", contractorPageID))

	page, err := s.client.FindPageByID(ctx, contractorPageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to fetch contractor page: page_id=%s", contractorPageID))
		return "", err
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Error(errors.New("invalid properties type"), "failed to cast contractor properties")
		return "", nil
	}

	username := ExtractRichText(props, "Discord")
	if username == "" {
		s.logger.Debug(fmt.Sprintf("Discord field is empty for contractor: page_id=%s", contractorPageID))
	} else {
		s.logger.Debug(fmt.Sprintf("found Discord username: %s (page_id=%s)", username, contractorPageID))
	}

	return username, nil
}

// ContractorDetails holds contractor information from Notion
type ContractorDetails struct {
	PageID          string
	FullName        string
	DiscordUsername string
	TeamEmail       string
	Status          string
}

// GetContractorDetails fetches contractor details from Notion by page ID
// Returns nil if contractor not found
func (s *LeaveService) GetContractorDetails(ctx context.Context, contractorPageID string) (*ContractorDetails, error) {
	if contractorPageID == "" {
		s.logger.Debug("contractor page ID is empty")
		return nil, nil
	}

	s.logger.Debug(fmt.Sprintf("fetching contractor details: page_id=%s", contractorPageID))

	page, err := s.client.FindPageByID(ctx, contractorPageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to fetch contractor page: page_id=%s", contractorPageID))
		return nil, err
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		s.logger.Error(errors.New("invalid properties type"), "failed to cast contractor properties")
		return nil, errors.New("failed to cast contractor properties")
	}

	details := &ContractorDetails{
		PageID:          contractorPageID,
		FullName:        ExtractTitle(props, "Full Name"),
		DiscordUsername: ExtractRichText(props, "Discord"),
		TeamEmail:       ExtractEmail(props, "Team Email"),
		Status:          ExtractSelect(props, "Status"),
	}

	s.logger.Debug(fmt.Sprintf("fetched contractor details: page_id=%s name=%s discord=%s status=%s",
		contractorPageID, details.FullName, details.DiscordUsername, details.Status))

	return details, nil
}

// LookupContractorByEmail finds contractor page ID by team email
// Returns empty string if not found (graceful handling)
// Returns error only on API failures
func (s *LeaveService) LookupContractorByEmail(
	ctx context.Context,
	teamEmail string,
) (string, error) {
	if teamEmail == "" {
		s.logger.Debug("team email is empty, skipping contractor lookup")
		return "", nil
	}

	s.logger.Debug(fmt.Sprintf("looking up contractor by email: %s", teamEmail))

	filter := &nt.DatabaseQueryFilter{
		Property: "Team Email",
		DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
			Email: &nt.TextPropertyFilter{
				Equals: teamEmail,
			},
		},
	}

	query := &nt.DatabaseQuery{
		Filter: filter,
	}

	contractorDBID := s.cfg.LeaveIntegration.Notion.ContractorDBID
	resp, err := s.client.QueryDatabase(ctx, contractorDBID, query)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to query contractors: email=%s", teamEmail))
		return "", err
	}

	if len(resp.Results) == 0 {
		s.logger.Info(fmt.Sprintf("contractor not found in Notion: email=%s", teamEmail))
		return "", nil
	}

	if len(resp.Results) > 1 {
		s.logger.Warn(fmt.Sprintf("multiple contractors found for email (taking first): email=%s count=%d", teamEmail, len(resp.Results)))
	}

	contractorPageID := resp.Results[0].ID
	s.logger.Debug(fmt.Sprintf("found contractor: email=%s page_id=%s", teamEmail, contractorPageID))

	return contractorPageID, nil
}

