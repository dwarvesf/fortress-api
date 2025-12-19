package notion

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

// TimesheetEntry represents a timesheet entry from Notion
type TimesheetEntry struct {
	PageID       string     `json:"page_id"`
	ProjectID    string     `json:"project_id"`    // Relation page ID
	ProjectName  string     `json:"project_name"`  // For display
	ContractorID string     `json:"contractor_id"` // Relation page ID
	Discord      string     `json:"discord"`       // Discord username
	Date         *time.Time `json:"date"`
	TaskType     string     `json:"task_type"` // Development, Design, Meeting
	Status       string     `json:"status"`    // Approved/Pending
	Hours        float64    `json:"hours"`
	ProofOfWorks string     `json:"proof_of_works"`
	TaskOrderID  string     `json:"task_order_id"` // Optional relation
}

// TimesheetService handles timesheet operations with Notion
type TimesheetService struct {
	client *nt.Client
	cfg    *config.Config
	store  *store.Store
	repo   store.DBRepo
	logger logger.Logger
}

// NewTimesheetService creates a new Notion timesheet service
func NewTimesheetService(cfg *config.Config, store *store.Store, repo store.DBRepo, logger logger.Logger) *TimesheetService {
	if cfg.Notion.Secret == "" {
		logger.Error(errors.New("notion secret not configured"), "notion secret is empty")
		return nil
	}

	logger.Debug("creating new TimesheetService")

	return &TimesheetService{
		client: nt.NewClient(cfg.Notion.Secret),
		cfg:    cfg,
		store:  store,
		repo:   repo,
		logger: logger,
	}
}

// CreateTimesheetEntry creates a new timesheet entry in Notion
func (s *TimesheetService) CreateTimesheetEntry(ctx context.Context, entry TimesheetEntry) (string, error) {
	if s.client == nil {
		return "", errors.New("notion client is nil")
	}

	timesheetDBID := s.cfg.Notion.Databases.Timesheet
	if timesheetDBID == "" {
		return "", errors.New("timesheet database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("creating timesheet entry: project=%s discord=%s date=%v hours=%.2f",
		entry.ProjectID, entry.Discord, entry.Date, entry.Hours))

	// Build properties for Notion
	props := nt.DatabasePageProperties{
		"Project": nt.DatabasePageProperty{
			Type:     nt.DBPropTypeRelation,
			Relation: []nt.Relation{{ID: entry.ProjectID}},
		},
		"Contractor": nt.DatabasePageProperty{
			Type:     nt.DBPropTypeRelation,
			Relation: []nt.Relation{{ID: entry.ContractorID}},
		},
		"Task Type": nt.DatabasePageProperty{
			Type:   nt.DBPropTypeSelect,
			Select: &nt.SelectOptions{Name: entry.TaskType},
		},
		"Hours": nt.DatabasePageProperty{
			Type:   nt.DBPropTypeNumber,
			Number: &entry.Hours,
		},
		"Proof of Works": nt.DatabasePageProperty{
			Type: nt.DBPropTypeRichText,
			RichText: []nt.RichText{{
				Type: nt.RichTextTypeText,
				Text: &nt.Text{Content: entry.ProofOfWorks},
			}},
		},
	}

	// Add date if provided
	if entry.Date != nil {
		props["Date"] = nt.DatabasePageProperty{
			Type: nt.DBPropTypeDate,
			Date: &nt.Date{Start: nt.NewDateTime(*entry.Date, false)},
		}
	}

	// Add optional Task Order relation
	if entry.TaskOrderID != "" {
		props["Task Order"] = nt.DatabasePageProperty{
			Type:     nt.DBPropTypeRelation,
			Relation: []nt.Relation{{ID: entry.TaskOrderID}},
		}
	}

	page, err := s.client.CreatePage(ctx, nt.CreatePageParams{
		ParentType:             nt.ParentTypeDatabase,
		ParentID:               timesheetDBID,
		DatabasePageProperties: &props,
	})
	if err != nil {
		s.logger.Error(err, "failed to create timesheet entry in Notion")
		return "", fmt.Errorf("failed to create timesheet entry: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("created timesheet entry: page_id=%s", page.ID))
	return page.ID, nil
}

// QueryTimesheetByDiscord queries timesheet entries by Discord username
func (s *TimesheetService) QueryTimesheetByDiscord(
	ctx context.Context,
	discordUsername string,
	startDate, endDate *time.Time,
) ([]TimesheetEntry, error) {
	if s.client == nil {
		return nil, errors.New("notion client is nil")
	}

	timesheetDBID := s.cfg.Notion.Databases.Timesheet
	if timesheetDBID == "" {
		return nil, errors.New("timesheet database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("querying timesheet entries: discord=%s start=%v end=%v",
		discordUsername, startDate, endDate))

	// Build filter for Discord username
	filters := []nt.DatabaseQueryFilter{
		{
			Property: "Discord",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				RichText: &nt.TextPropertyFilter{
					Contains: discordUsername,
				},
			},
		},
	}

	// Add date range filters
	if startDate != nil {
		filters = append(filters, nt.DatabaseQueryFilter{
			Property: "Date",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Date: &nt.DatePropertyFilter{
					OnOrAfter: startDate,
				},
			},
		})
	}

	if endDate != nil {
		filters = append(filters, nt.DatabaseQueryFilter{
			Property: "Date",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Date: &nt.DatePropertyFilter{
					OnOrBefore: endDate,
				},
			},
		})
	}

	filter := &nt.DatabaseQueryFilter{And: filters}

	resp, err := s.client.QueryDatabase(ctx, timesheetDBID, &nt.DatabaseQuery{
		Filter: filter,
		Sorts: []nt.DatabaseQuerySort{
			{Property: "Date", Direction: nt.SortDirDesc},
		},
	})
	if err != nil {
		s.logger.Error(err, "failed to query timesheet entries from Notion")
		return nil, fmt.Errorf("failed to query timesheet: %w", err)
	}

	// Parse results into TimesheetEntry structs
	entries := make([]TimesheetEntry, 0, len(resp.Results))
	for _, page := range resp.Results {
		props, ok := page.Properties.(nt.DatabasePageProperties)
		if !ok {
			continue
		}
		entry := s.parseTimesheetPage(page.ID, props)
		entries = append(entries, entry)
	}

	s.logger.Debug(fmt.Sprintf("found %d timesheet entries for discord=%s", len(entries), discordUsername))
	return entries, nil
}

// GetContractorPageIDByDiscordID looks up a contractor page ID by Discord ID
// It uses the existing pattern: discord_accounts -> employees -> Notion Contractor DB
func (s *TimesheetService) GetContractorPageIDByDiscordID(ctx context.Context, discordID string) (string, error) {
	s.logger.Debug(fmt.Sprintf("looking up contractor by Discord ID: %s", discordID))

	// 1. Look up discord_accounts table to get the employee's team email
	var discordAccount struct {
		ID string
	}
	err := s.repo.DB().WithContext(ctx).
		Table("discord_accounts").
		Select("id").
		Where("discord_id = ? AND deleted_at IS NULL", discordID).
		First(&discordAccount).Error
	if err != nil {
		s.logger.Debug(fmt.Sprintf("discord account not found for discord_id: %s, error: %v", discordID, err))
		return "", fmt.Errorf("discord account not found: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("found discord account: discord_id=%s account_id=%s", discordID, discordAccount.ID))

	// 2. Look up employees table by discord_account_id to get team_email
	var employee struct {
		TeamEmail string
	}
	err = s.repo.DB().WithContext(ctx).
		Table("employees").
		Select("team_email").
		Where("discord_account_id = ? AND deleted_at IS NULL", discordAccount.ID).
		First(&employee).Error
	if err != nil {
		s.logger.Debug(fmt.Sprintf("employee not found for discord_account_id: %s, error: %v", discordAccount.ID, err))
		return "", fmt.Errorf("employee not found for discord account: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("found employee: discord_account_id=%s team_email=%s", discordAccount.ID, employee.TeamEmail))

	// 3. Query Notion Contractor database by Team Email
	contractorDBID := s.cfg.Notion.Databases.Contractor
	if contractorDBID == "" {
		return "", errors.New("contractor database ID not configured")
	}

	filter := &nt.DatabaseQueryFilter{
		Property: "Team Email",
		DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
			Email: &nt.TextPropertyFilter{
				Equals: employee.TeamEmail,
			},
		},
	}

	resp, err := s.client.QueryDatabase(ctx, contractorDBID, &nt.DatabaseQuery{
		Filter:   filter,
		PageSize: 1,
	})
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to query contractor database: email=%s", employee.TeamEmail))
		return "", fmt.Errorf("failed to query contractor database: %w", err)
	}

	if len(resp.Results) == 0 {
		s.logger.Debug(fmt.Sprintf("no contractor found for email: %s", employee.TeamEmail))
		return "", fmt.Errorf("contractor not found for email: %s", employee.TeamEmail)
	}

	contractorPageID := resp.Results[0].ID
	s.logger.Debug(fmt.Sprintf("found contractor page: email=%s page_id=%s", employee.TeamEmail, contractorPageID))
	return contractorPageID, nil
}

// parseTimesheetPage parses a Notion page into a TimesheetEntry
func (s *TimesheetService) parseTimesheetPage(pageID string, props nt.DatabasePageProperties) TimesheetEntry {
	entry := TimesheetEntry{PageID: pageID}

	// Parse Hours
	if prop, ok := props["Hours"]; ok && prop.Number != nil {
		entry.Hours = *prop.Number
	}

	// Parse Date
	if prop, ok := props["Date"]; ok && prop.Date != nil {
		t := prop.Date.Start.Time
		if !t.IsZero() {
			entry.Date = &t
		}
	}

	// Parse Task Type
	if prop, ok := props["Task Type"]; ok && prop.Select != nil {
		entry.TaskType = prop.Select.Name
	}

	// Parse Status
	if prop, ok := props["Status"]; ok && prop.Status != nil {
		entry.Status = prop.Status.Name
	}

	// Parse Discord (rich text or rollup)
	if prop, ok := props["Discord"]; ok {
		if len(prop.RichText) > 0 {
			var parts []string
			for _, rt := range prop.RichText {
				parts = append(parts, rt.PlainText)
			}
			entry.Discord = strings.Join(parts, "")
		}
	}

	// Parse Proof of Works
	if prop, ok := props["Proof of Works"]; ok && len(prop.RichText) > 0 {
		var parts []string
		for _, rt := range prop.RichText {
			parts = append(parts, rt.PlainText)
		}
		entry.ProofOfWorks = strings.Join(parts, "")
	}

	// Parse Project relation
	if prop, ok := props["Project"]; ok && len(prop.Relation) > 0 {
		entry.ProjectID = prop.Relation[0].ID
	}

	// Parse Contractor relation
	if prop, ok := props["Contractor"]; ok && len(prop.Relation) > 0 {
		entry.ContractorID = prop.Relation[0].ID
	}

	// Parse Task Order relation
	if prop, ok := props["Task Order"]; ok && len(prop.Relation) > 0 {
		entry.TaskOrderID = prop.Relation[0].ID
	}

	return entry
}
