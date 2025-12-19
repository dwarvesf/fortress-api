package notion

import (
	"context"
	"errors"
	"fmt"
	"strings"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// TimesheetEntry represents a timesheet entry from Notion
type TimesheetEntry struct {
	PageID            string
	Title             string
	ContractorPageID  string // Relation page ID
	CreatedByUserID   string // Created by user ID
	CreatedByUserName string // Created by user name
	ProjectPageID     string // Relation page ID
	Date              string // Date field
	Hours             float64 // Number field
	Status            string  // Status field
}

// TimesheetService handles timesheet operations with Notion
type TimesheetService struct {
	client *nt.Client
	cfg    *config.Config
	logger logger.Logger
}

// NewTimesheetService creates a new Notion timesheet service
func NewTimesheetService(cfg *config.Config, logger logger.Logger) *TimesheetService {
	if cfg.Notion.Secret == "" {
		logger.Error(errors.New("notion secret not configured"), "notion secret is empty")
		return nil
	}

	logger.Debug("creating new TimesheetService")

	return &TimesheetService{
		client: nt.NewClient(cfg.Notion.Secret),
		cfg:    cfg,
		logger: logger,
	}
}

// GetTimesheetEntry fetches a timesheet entry by page ID from Notion
func (s *TimesheetService) GetTimesheetEntry(ctx context.Context, pageID string) (*TimesheetEntry, error) {
	if s.client == nil {
		return nil, errors.New("notion client is nil")
	}

	s.logger.Debug(fmt.Sprintf("fetching timesheet entry from Notion: page_id=%s", pageID))

	page, err := s.client.FindPageByID(ctx, pageID)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to fetch timesheet entry page: page_id=%s", pageID))
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}

	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		return nil, errors.New("failed to cast page properties")
	}

	// Extract all properties
	entry := &TimesheetEntry{
		PageID:           pageID,
		Title:            s.extractTitle(props, "(auto) Timesheet Entry"),
		ContractorPageID: s.extractFirstRelationID(props, "Contractor"),
		ProjectPageID:    s.extractFirstRelationID(props, "Project"),
		Date:             s.extractDateString(props, "Date"),
		Hours:            s.extractNumber(props, "Hours"),
		Status:           s.extractStatus(props, "Status"),
	}

	// Extract created by user info
	entry.CreatedByUserID = page.CreatedBy.ID
	entry.CreatedByUserName = "" // Name not available in API response

	s.logger.Debug(fmt.Sprintf("fetched timesheet entry: page_id=%s created_by_id=%s contractor=%s",
		pageID, entry.CreatedByUserID, entry.ContractorPageID))

	return entry, nil
}

// UpdateTimesheetEntry updates the contractor field of a timesheet entry in Notion
func (s *TimesheetService) UpdateTimesheetEntry(ctx context.Context, pageID, contractorPageID string) error {
	if s.client == nil {
		return errors.New("notion client is nil")
	}

	s.logger.Debug(fmt.Sprintf("updating timesheet entry in Notion: page_id=%s contractor=%s",
		pageID, contractorPageID))

	// Build update params
	updateParams := nt.UpdatePageParams{
		DatabasePageProperties: nt.DatabasePageProperties{},
	}

	// Add contractor relation if provided
	if contractorPageID != "" {
		updateParams.DatabasePageProperties["Contractor"] = nt.DatabasePageProperty{
			Relation: []nt.Relation{
				{ID: contractorPageID},
			},
		}
	}

	_, err := s.client.UpdatePage(ctx, pageID, updateParams)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to update timesheet entry in Notion: page_id=%s", pageID))
		return fmt.Errorf("failed to update page: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("successfully updated timesheet entry in Notion: page_id=%s", pageID))
	return nil
}

// FindContractorByPersonID finds a contractor page ID by Notion person ID
// Queries the Contractor database for a contractor where the Person field contains the given user ID
func (s *TimesheetService) FindContractorByPersonID(ctx context.Context, personID string) (string, string, error) {
	if s.client == nil {
		return "", "", errors.New("notion client is nil")
	}

	contractorDBID := s.cfg.Notion.Databases.Contractor
	if contractorDBID == "" {
		return "", "", errors.New("contractor database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("looking up contractor by person ID: person_id=%s db_id=%s", personID, contractorDBID))

	// Query contractor database for matching person ID
	// Note: Notion doesn't support filtering by People property directly via API
	// We need to fetch all contractors and filter in-memory
	resp, err := s.client.QueryDatabase(ctx, contractorDBID, &nt.DatabaseQuery{
		PageSize: 100,
	})
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to query contractor database: person_id=%s", personID))
		return "", "", fmt.Errorf("failed to query contractor database: %w", err)
	}

	// Search through results for matching person ID
	for _, page := range resp.Results {
		props, ok := page.Properties.(nt.DatabasePageProperties)
		if !ok {
			continue
		}

		// Check if Person field contains the target person ID
		if personProp, ok := props["Person"]; ok && len(personProp.People) > 0 {
			for _, person := range personProp.People {
				if person.ID == personID {
					// Found a match! Extract Discord username
					discordUsername := s.extractRichText(props, "Discord")
					s.logger.Debug(fmt.Sprintf("found contractor: person_id=%s contractor_id=%s discord=%s",
						personID, page.ID, discordUsername))
					return page.ID, discordUsername, nil
				}
			}
		}
	}

	// Handle pagination if needed
	if resp.HasMore && resp.NextCursor != nil {
		cursor := *resp.NextCursor
		for cursor != "" {
			resp, err = s.client.QueryDatabase(ctx, contractorDBID, &nt.DatabaseQuery{
				StartCursor: cursor,
				PageSize:    100,
			})
			if err != nil {
				break
			}

			for _, page := range resp.Results {
				props, ok := page.Properties.(nt.DatabasePageProperties)
				if !ok {
					continue
				}

				if personProp, ok := props["Person"]; ok && len(personProp.People) > 0 {
					for _, person := range personProp.People {
						if person.ID == personID {
							discordUsername := s.extractRichText(props, "Discord")
							s.logger.Debug(fmt.Sprintf("found contractor: person_id=%s contractor_id=%s discord=%s",
								personID, page.ID, discordUsername))
							return page.ID, discordUsername, nil
						}
					}
				}
			}

			if !resp.HasMore || resp.NextCursor == nil {
				break
			}
			cursor = *resp.NextCursor
		}
	}

	s.logger.Debug(fmt.Sprintf("no contractor found for person ID: %s", personID))
	return "", "", fmt.Errorf("contractor not found for person ID: %s", personID)
}

// Property extraction helpers

// extractTitle extracts a title property value
func (s *TimesheetService) extractTitle(props nt.DatabasePageProperties, propName string) string {
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
func (s *TimesheetService) extractSelect(props nt.DatabasePageProperties, propName string) string {
	if prop, ok := props[propName]; ok && prop.Select != nil {
		return prop.Select.Name
	}
	return ""
}

// extractStatus extracts a status property value
func (s *TimesheetService) extractStatus(props nt.DatabasePageProperties, propName string) string {
	if prop, ok := props[propName]; ok && prop.Status != nil {
		return prop.Status.Name
	}
	return ""
}

// extractFirstRelationID extracts the first relation page ID
func (s *TimesheetService) extractFirstRelationID(props nt.DatabasePageProperties, propName string) string {
	if prop, ok := props[propName]; ok && len(prop.Relation) > 0 {
		return prop.Relation[0].ID
	}
	return ""
}

// extractRichText concatenates rich text parts into a single string
func (s *TimesheetService) extractRichText(props nt.DatabasePageProperties, propName string) string {
	if prop, ok := props[propName]; ok && len(prop.RichText) > 0 {
		var parts []string
		for _, rt := range prop.RichText {
			parts = append(parts, rt.PlainText)
		}
		result := strings.TrimSpace(strings.Join(parts, ""))
		s.logger.Debug(fmt.Sprintf("extractRichText: property %s has value: %s", propName, result))
		return result
	}
	s.logger.Debug(fmt.Sprintf("extractRichText: property %s not found or empty", propName))
	return ""
}

// extractDateString extracts a date property value as string
func (s *TimesheetService) extractDateString(props nt.DatabasePageProperties, propName string) string {
	if prop, ok := props[propName]; ok && prop.Date != nil {
		return prop.Date.Start.String()
	}
	return ""
}

// extractNumber extracts a number property value
func (s *TimesheetService) extractNumber(props nt.DatabasePageProperties, propName string) float64 {
	if prop, ok := props[propName]; ok && prop.Number != nil {
		return *prop.Number
	}
	return 0
}
