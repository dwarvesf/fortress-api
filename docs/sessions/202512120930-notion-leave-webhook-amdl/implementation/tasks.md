# Implementation Tasks: Notion Leave Webhook AM/DL Integration

**Version:** 1.0
**Date:** 2025-12-12
**Status:** Ready for Implementation

## Overview

This document provides a detailed breakdown of implementation tasks for the Notion Leave Webhook AM/DL integration feature. Tasks are organized in dependency order with clear acceptance criteria.

---

## Task 1: Database Store Layer - Add OneByUsername Method

**Priority:** High
**Estimated Effort:** 1-2 hours
**Dependencies:** None

### Description
Add a new method to the Discord Account store to lookup accounts by Discord username. This is required to convert Discord usernames (from Notion) to Discord IDs (for mentions).

### Files to Modify
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/store/discordaccount/discord_account.go`

### Implementation Details

#### Add Method to Store
```go
// OneByUsername gets a discord account by discord username
// Returns nil, nil if not found (graceful handling)
// Returns nil, error if database error occurs
func (r *store) OneByUsername(db *gorm.DB, username string) (*model.DiscordAccount, error) {
    var res model.DiscordAccount
    err := db.Where("discord_username = ?", username).First(&res).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil // Not found is not an error
        }
        return nil, err
    }
    return &res, nil
}
```

### Acceptance Criteria
- [ ] Method `OneByUsername` added to store implementation
- [ ] Returns `nil, nil` when username not found (graceful)
- [ ] Returns `nil, error` on database errors
- [ ] Uses parameterized query for SQL injection protection
- [ ] Follows existing store method patterns

### Testing Requirements
- Unit test: Username exists - returns account
- Unit test: Username not found - returns nil, nil
- Unit test: Database error - returns nil, error
- Unit test: Empty username - returns nil, nil
- Unit test: Case sensitivity handling

---

## Task 2: Database Migration - Add Index on discord_username

**Priority:** Medium
**Estimated Effort:** 30 minutes
**Dependencies:** None

### Description
Add database index on `discord_accounts.discord_username` to optimize lookups. This will be called on every leave request creation.

### Files to Create
- `/Users/quang/workspace/dwarvesf/fortress-api/migrations/schemas/YYYYMMDDHHMMSS-add-discord-username-index.sql`

### Implementation Details

#### Migration File
```sql
-- +migrate Up
CREATE INDEX IF NOT EXISTS idx_discord_accounts_username
ON discord_accounts(discord_username);

-- +migrate Down
DROP INDEX IF EXISTS idx_discord_accounts_username;
```

### Acceptance Criteria
- [ ] Migration file created with proper timestamp format
- [ ] Up migration creates index
- [ ] Down migration drops index
- [ ] Index created with `IF NOT EXISTS` for safety
- [ ] Migration tested locally with `make migrate-up` and `make migrate-down`

### Testing Requirements
- Test: `make migrate-up` succeeds
- Test: `make migrate-down` succeeds
- Test: Verify index exists in database after up migration
- Test: Verify index removed after down migration

---

## Task 3: Configuration - Add Deployment Tracker DB ID

**Priority:** High
**Estimated Effort:** 15 minutes
**Dependencies:** None

### Description
Add environment variable and config struct field for Notion Deployment Tracker database ID.

### Files to Modify
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/config/config.go`
- `/Users/quang/workspace/dwarvesf/fortress-api/.env.sample` (if exists)

### Implementation Details

#### Config Struct Update
```go
type NotionConfig struct {
    Secret    string
    Databases struct {
        Contractor         string `envconfig:"NOTION_CONTRACTOR_DB_ID" required:"true"`
        DeploymentTracker  string `envconfig:"NOTION_DEPLOYMENT_TRACKER_DB_ID" required:"true"`
        LeaveRequest       string `envconfig:"NOTION_LEAVE_REQUEST_DB_ID" required:"true"`
        // ... other databases
    }
}
```

#### Environment Variable
```bash
NOTION_DEPLOYMENT_TRACKER_DB_ID=2b864b29b84c80799568dc17685f4f33
```

### Acceptance Criteria
- [ ] Config field added with proper envconfig tag
- [ ] Field marked as `required:"true"`
- [ ] Environment variable documented in `.env.sample`
- [ ] Config loads successfully with new field

### Testing Requirements
- Test: Application starts with valid config
- Test: Application fails gracefully with missing config

---

## Task 4: Service Layer - Helper Functions for Property Extraction

**Priority:** High
**Estimated Effort:** 2-3 hours
**Dependencies:** None

### Description
Add helper functions to extract Notion properties (relation, rollup, rich_text). These will be reused across multiple service methods.

### Files to Modify
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/leave.go`

### Implementation Details

#### Helper Functions
```go
// extractFirstRelationID extracts the first relation ID from a property
func (s *LeaveService) extractFirstRelationID(props nt.DatabasePageProperties, propName string) string {
    if prop, ok := props[propName]; ok && len(prop.Relation) > 0 {
        return prop.Relation[0].ID
    }
    return ""
}

// extractRollupRelations extracts all relation IDs from a rollup array property
func (s *LeaveService) extractRollupRelations(props nt.DatabasePageProperties, propName string) []string {
    var relationIDs []string

    if prop, ok := props[propName]; ok {
        if prop.Rollup.Type == "array" && len(prop.Rollup.Array) > 0 {
            for _, item := range prop.Rollup.Array {
                if len(item.Relation) > 0 {
                    for _, rel := range item.Relation {
                        relationIDs = append(relationIDs, rel.ID)
                    }
                }
            }
        }
    }

    return relationIDs
}

// extractRichText concatenates rich text parts into a single string
func (s *LeaveService) extractRichText(props nt.DatabasePageProperties, propName string) string {
    if prop, ok := props[propName]; ok && len(prop.RichText) > 0 {
        var parts []string
        for _, rt := range prop.RichText {
            parts = append(parts, rt.PlainText)
        }
        return strings.TrimSpace(strings.Join(parts, ""))
    }
    return ""
}
```

### Acceptance Criteria
- [ ] `extractFirstRelationID` handles empty relations gracefully
- [ ] `extractRollupRelations` flattens nested rollup arrays correctly
- [ ] `extractRichText` concatenates multiple parts and trims whitespace
- [ ] All functions handle missing properties without panicking
- [ ] Functions are private (lowercase first letter)

### Testing Requirements
- Unit test: Extract relation with value
- Unit test: Extract relation when empty
- Unit test: Extract rollup with nested array
- Unit test: Extract rollup when empty
- Unit test: Extract rich text with multiple parts
- Unit test: Extract rich text when empty
- Unit test: Missing properties return empty/zero values

---

## Task 5: Service Layer - GetActiveDeploymentsForContractor

**Priority:** High
**Estimated Effort:** 2-3 hours
**Dependencies:** Task 3 (Config), Task 4 (Helpers)

### Description
Implement service method to query Notion Deployment Tracker for active deployments of a contractor.

### Files to Modify
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/leave.go`

### Implementation Details

#### Service Method
```go
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
```

### Acceptance Criteria
- [ ] Queries Deployment Tracker with correct filter (AND condition)
- [ ] Filters by Contractor relation and Deployment Status = "Active"
- [ ] Returns empty array when no deployments found
- [ ] Returns error on API failures
- [ ] Logs query parameters for debugging
- [ ] Logs result count
- [ ] Handles empty contractor ID gracefully

### Testing Requirements
- Unit test: Single active deployment found
- Unit test: Multiple active deployments found
- Unit test: No active deployments (empty array)
- Unit test: Empty contractor ID (empty array)
- Unit test: API error (returns error)
- Unit test: Context cancellation
- Unit test: Filter construction is correct

---

## Task 6: Service Layer - ExtractStakeholdersFromDeployment

**Priority:** High
**Estimated Effort:** 3-4 hours
**Dependencies:** Task 4 (Helpers)

### Description
Extract AM/DL contractor page IDs from a deployment page with override priority logic.

### Files to Modify
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/leave.go`

### Implementation Details

#### Service Method
```go
// ExtractStakeholdersFromDeployment extracts AM/DL contractor page IDs from deployment
// Override AM/DL takes precedence over rollup AM/DL
// Returns unique stakeholder page IDs
func (s *LeaveService) ExtractStakeholdersFromDeployment(
    deploymentPage nt.Page,
) []string {
    props, ok := deploymentPage.Properties.(nt.DatabasePageProperties)
    if !ok {
        s.logger.Error(errors.New("invalid properties type"), "failed to cast deployment properties")
        return []string{}
    }

    stakeholderMap := make(map[string]bool) // Use map for deduplication

    // Extract AM: Override AM takes precedence over Account Managers rollup
    overrideAM := s.extractFirstRelationID(props, "Override AM")
    if overrideAM != "" {
        s.logger.Debug(fmt.Sprintf("found override AM: %s", overrideAM))
        stakeholderMap[overrideAM] = true
    } else {
        // Fallback to rollup
        rollupAMs := s.extractRollupRelations(props, "Account Managers")
        s.logger.Debug(fmt.Sprintf("found %d AMs from rollup", len(rollupAMs)))
        for _, amID := range rollupAMs {
            stakeholderMap[amID] = true
        }
    }

    // Extract DL: Override DL takes precedence over Delivery Leads rollup
    overrideDL := s.extractFirstRelationID(props, "Override DL")
    if overrideDL != "" {
        s.logger.Debug(fmt.Sprintf("found override DL: %s", overrideDL))
        stakeholderMap[overrideDL] = true
    } else {
        // Fallback to rollup
        rollupDLs := s.extractRollupRelations(props, "Delivery Leads")
        s.logger.Debug(fmt.Sprintf("found %d DLs from rollup", len(rollupDLs)))
        for _, dlID := range rollupDLs {
            stakeholderMap[dlID] = true
        }
    }

    // Convert map to slice
    stakeholders := make([]string, 0, len(stakeholderMap))
    for id := range stakeholderMap {
        stakeholders = append(stakeholders, id)
    }

    s.logger.Debug(fmt.Sprintf("extracted %d unique stakeholders from deployment", len(stakeholders)))

    return stakeholders
}
```

### Acceptance Criteria
- [ ] Override AM takes precedence over rollup
- [ ] Override DL takes precedence over rollup
- [ ] Falls back to rollup when override is empty
- [ ] Deduplicates stakeholders (same person as AM and DL)
- [ ] Returns empty array if no stakeholders found
- [ ] Logs each extraction step for debugging
- [ ] Handles invalid properties gracefully

### Testing Requirements
- Unit test: Override AM + Override DL present (use overrides)
- Unit test: No overrides (use rollups)
- Unit test: Mixed: Override AM, rollup DL
- Unit test: Mixed: Rollup AM, Override DL
- Unit test: Same person as AM and DL (deduplication)
- Unit test: Empty AM and DL (empty array)
- Unit test: Invalid properties (empty array)
- Unit test: Multiple AMs/DLs from rollup

---

## Task 7: Service Layer - GetDiscordUsernameFromContractor

**Priority:** High
**Estimated Effort:** 2 hours
**Dependencies:** Task 4 (Helpers)

### Description
Fetch Discord username from a Notion contractor page.

### Files to Modify
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/leave.go`

### Implementation Details

#### Service Method
```go
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

    username := s.extractRichText(props, "Discord")
    if username == "" {
        s.logger.Debug(fmt.Sprintf("Discord field is empty for contractor: page_id=%s", contractorPageID))
    } else {
        s.logger.Debug(fmt.Sprintf("found Discord username: %s (page_id=%s)", username, contractorPageID))
    }

    return username, nil
}
```

### Acceptance Criteria
- [ ] Fetches contractor page by ID
- [ ] Extracts Discord rich_text property
- [ ] Returns empty string when Discord field empty (not error)
- [ ] Returns error on API failures
- [ ] Logs fetch attempt and result
- [ ] Handles empty page ID gracefully

### Testing Requirements
- Unit test: Discord username exists (returns username)
- Unit test: Discord field empty (returns "")
- Unit test: Discord field missing (returns "")
- Unit test: Multi-part rich text (concatenates)
- Unit test: Page not found (returns error)
- Unit test: API error (returns error)
- Unit test: Empty contractor ID (returns "")
- Unit test: Invalid properties (returns "")

---

## Task 8: Service Layer - LookupContractorByEmail

**Priority:** High
**Estimated Effort:** 2 hours
**Dependencies:** Task 3 (Config)

### Description
Lookup contractor page ID by team email in Notion Contractors database.

### Files to Modify
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/leave.go`

### Implementation Details

#### Service Method
```go
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
```

### Acceptance Criteria
- [ ] Queries Contractors DB with email filter
- [ ] Returns first result's page ID
- [ ] Returns empty string when not found (not error)
- [ ] Logs warning if multiple results found
- [ ] Returns error on API failures
- [ ] Handles empty email gracefully

### Testing Requirements
- Unit test: Contractor found (returns page ID)
- Unit test: Contractor not found (returns "")
- Unit test: Multiple contractors (takes first, logs warning)
- Unit test: Empty email (returns "")
- Unit test: API error (returns error)
- Unit test: Context cancellation

---

## Task 9: Handler Layer - GetDiscordMentionFromUsername

**Priority:** High
**Estimated Effort:** 2 hours
**Dependencies:** Task 1 (Store method)

### Description
Convert Discord username to Discord mention format by looking up Discord ID in database.

### Files to Modify
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/webhook/notion_leave.go`

### Implementation Details

#### Handler Method
```go
// getDiscordMentionFromUsername converts Discord username to mention format
// Returns empty string if username not found in database (graceful handling)
func (h *handler) getDiscordMentionFromUsername(
    l logger.Logger,
    discordUsername string,
) string {
    if discordUsername == "" {
        return ""
    }

    username := strings.TrimSpace(discordUsername)
    l.Debug(fmt.Sprintf("looking up Discord ID for username: %s", username))

    db := h.store.DBRepo.DB()
    account, err := h.store.DiscordAccount.OneByUsername(db, username)
    if err != nil {
        l.Error(err, fmt.Sprintf("failed to lookup Discord account: username=%s", username))
        return ""
    }

    if account == nil {
        l.Info(fmt.Sprintf("Discord username not found in database: %s", username))
        return ""
    }

    mention := fmt.Sprintf("<@%s>", account.DiscordID)
    l.Debug(fmt.Sprintf("converted username to mention: %s -> %s", username, mention))

    return mention
}
```

### Acceptance Criteria
- [ ] Calls store `OneByUsername` method
- [ ] Formats mention as `<@discord_id>`
- [ ] Returns empty string if not found
- [ ] Returns empty string on database errors
- [ ] Trims whitespace from username
- [ ] Logs lookup attempt and result

### Testing Requirements
- Unit test: Username exists (returns mention)
- Unit test: Username not found (returns "")
- Unit test: Database error (returns "")
- Unit test: Empty username (returns "")
- Unit test: Username with whitespace (trims)
- Unit test: Mention format is correct

---

## Task 10: Handler Layer - GetAMDLMentionsFromDeployments

**Priority:** High
**Estimated Effort:** 3-4 hours
**Dependencies:** Task 5, 6, 7, 8, 9

### Description
Orchestrate the full flow: email → contractor → deployments → stakeholders → usernames → mentions.

### Files to Modify
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/webhook/notion_leave.go`

### Implementation Details

#### Handler Method
```go
// getAMDLMentionsFromDeployments gets Discord mentions for AM/DL from active deployments
// Returns array of Discord mentions (may be empty if no stakeholders found)
func (h *handler) getAMDLMentionsFromDeployments(
    ctx context.Context,
    l logger.Logger,
    teamEmail string,
) []string {
    // Step 1: Lookup contractor by email
    contractorPageID, err := h.service.NotionLeave.LookupContractorByEmail(ctx, teamEmail)
    if err != nil {
        l.Error(err, fmt.Sprintf("failed to lookup contractor: email=%s", teamEmail))
        return []string{}
    }
    if contractorPageID == "" {
        l.Info(fmt.Sprintf("contractor not found for email: %s", teamEmail))
        return []string{}
    }

    // Step 2: Get active deployments
    deployments, err := h.service.NotionLeave.GetActiveDeploymentsForContractor(ctx, contractorPageID)
    if err != nil {
        l.Error(err, fmt.Sprintf("failed to get active deployments: contractor_id=%s", contractorPageID))
        return []string{}
    }
    if len(deployments) == 0 {
        l.Info(fmt.Sprintf("no active deployments found: contractor_id=%s", contractorPageID))
        return []string{}
    }

    // Step 3: Extract stakeholders from all deployments
    stakeholderMap := make(map[string]bool)
    for _, deployment := range deployments {
        stakeholders := h.service.NotionLeave.ExtractStakeholdersFromDeployment(deployment)
        for _, stakeholderID := range stakeholders {
            stakeholderMap[stakeholderID] = true
        }
    }

    // Step 4: Get Discord usernames from Notion
    var mentions []string
    for stakeholderID := range stakeholderMap {
        username, err := h.service.NotionLeave.GetDiscordUsernameFromContractor(ctx, stakeholderID)
        if err != nil {
            l.Error(err, fmt.Sprintf("failed to get Discord username: contractor_id=%s", stakeholderID))
            continue // Skip this stakeholder
        }
        if username == "" {
            l.Debug(fmt.Sprintf("Discord username not set for stakeholder: contractor_id=%s", stakeholderID))
            continue // Skip this stakeholder
        }

        // Step 5: Convert username to mention
        mention := h.getDiscordMentionFromUsername(l, username)
        if mention != "" {
            mentions = append(mentions, mention)
        }
    }

    l.Debug(fmt.Sprintf("extracted %d Discord mentions from %d deployments", len(mentions), len(deployments)))

    return mentions
}
```

### Acceptance Criteria
- [ ] Calls all service methods in correct order
- [ ] Deduplicates stakeholders across deployments
- [ ] Skips stakeholders on individual failures (graceful degradation)
- [ ] Returns empty array if any step returns no data
- [ ] Returns empty array on critical errors
- [ ] Logs each step for debugging
- [ ] Continues processing remaining stakeholders on individual failures

### Testing Requirements
- Unit test: Full flow with active deployments (returns mentions)
- Unit test: Contractor not found (empty array)
- Unit test: No active deployments (empty array)
- Unit test: No stakeholders in deployments (empty array)
- Unit test: Discord username not set (skips stakeholder)
- Unit test: Discord username not in DB (skips stakeholder)
- Unit test: Multiple deployments with same stakeholder (deduplicates)
- Unit test: Service method error (continues with others)

---

## Task 11: Handler Layer - Update handleNotionLeaveCreated

**Priority:** High
**Estimated Effort:** 2-3 hours
**Dependencies:** Task 10

### Description
Modify the leave created handler to use new AM/DL lookup instead of assignees multi-select.

### Files to Modify
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/webhook/notion_leave.go`

### Implementation Details

#### Modify Handler
Replace the current assignee extraction logic with the new AM/DL lookup:

```go
// In handleNotionLeaveCreated function:

// OLD CODE (remove or comment out):
// assignees := leave.Assignees
// var assigneeMentions []string
// for _, email := range assignees {
//     // ... existing assignee lookup logic
// }

// NEW CODE:
// Get AM/DL mentions from deployments
assigneeMentions := h.getAMDLMentionsFromDeployments(ctx, l, leave.Email)

if len(assigneeMentions) == 0 {
    l.Info(fmt.Sprintf("no AM/DL mentions found for leave request: email=%s", leave.Email))
    // Still send notification without mentions
}

// Format assignee list for Discord message
assigneesList := "None"
if len(assigneeMentions) > 0 {
    assigneesList = strings.Join(assigneeMentions, " ")
}
```

### Acceptance Criteria
- [ ] Remove old assignee multi-select logic
- [ ] Call `getAMDLMentionsFromDeployments` with team email
- [ ] Send notification even if mentions array is empty
- [ ] Format mentions correctly in Discord message
- [ ] Preserve existing notification structure
- [ ] Log when no mentions found

### Testing Requirements
- Unit test: Leave with active deployments (includes mentions)
- Unit test: Leave with no deployments (no mentions, still sends)
- Unit test: Leave with contractor not found (no mentions, still sends)
- Unit test: Discord message format is correct
- Unit test: Multiple mentions formatted correctly

---

## Task 12: Handler Layer - UpdateApprovedRejectedBy

**Priority:** Medium
**Estimated Effort:** 2-3 hours
**Dependencies:** Task 8

### Description
Update "Approved/Rejected By" relation field when leave is approved/rejected via Discord button.

### Files to Modify
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/webhook/notion_leave.go`

### Implementation Details

#### New Handler Method
```go
// updateApprovedRejectedBy updates the Approved/Rejected By relation on leave request
// Does not fail the approval/rejection if update fails (graceful degradation)
func (h *handler) updateApprovedRejectedBy(
    ctx context.Context,
    l logger.Logger,
    leavePageID string,
    approverDiscordID string,
) error {
    l.Debug(fmt.Sprintf("updating approved/rejected by: page_id=%s discord_id=%s", leavePageID, approverDiscordID))

    // Step 1: Get Discord account by ID
    db := h.store.DBRepo.DB()
    account, err := h.store.DiscordAccount.OneByDiscordID(db, approverDiscordID)
    if err != nil {
        l.Error(err, fmt.Sprintf("failed to get Discord account: discord_id=%s", approverDiscordID))
        return err
    }
    if account == nil {
        l.Warn(fmt.Sprintf("Discord account not found: discord_id=%s", approverDiscordID))
        return nil // Graceful: don't fail approval
    }

    // Step 2: Get Discord username
    username := account.DiscordUsername
    if username == "" {
        l.Warn(fmt.Sprintf("Discord username is empty: discord_id=%s", approverDiscordID))
        return nil // Graceful: don't fail approval
    }

    // Step 3: Lookup contractor by Discord username
    contractorPageID, err := h.service.NotionLeave.LookupContractorByDiscordUsername(ctx, username)
    if err != nil {
        l.Error(err, fmt.Sprintf("failed to lookup contractor: username=%s", username))
        return err
    }
    if contractorPageID == "" {
        l.Warn(fmt.Sprintf("contractor not found for Discord username: %s", username))
        return nil // Graceful: don't fail approval
    }

    // Step 4: Update leave request with contractor relation
    updateParams := nt.UpdatePageParams{
        DatabasePageProperties: nt.DatabasePageProperties{
            "Approved/Rejected By": nt.DatabasePageProperty{
                Relation: []nt.Relation{
                    {ID: contractorPageID},
                },
            },
        },
    }

    _, err = h.service.NotionLeave.Client().UpdatePage(ctx, leavePageID, updateParams)
    if err != nil {
        l.Error(err, fmt.Sprintf("failed to update approved/rejected by: page_id=%s", leavePageID))
        return err
    }

    l.Info(fmt.Sprintf("updated approved/rejected by: page_id=%s contractor_id=%s", leavePageID, contractorPageID))

    return nil
}
```

#### Add Helper Method to Service
```go
// In pkg/service/notion/leave.go

// LookupContractorByDiscordUsername finds contractor page ID by Discord username
func (s *LeaveService) LookupContractorByDiscordUsername(
    ctx context.Context,
    discordUsername string,
) (string, error) {
    if discordUsername == "" {
        return "", nil
    }

    filter := &nt.DatabaseQueryFilter{
        Property: "Discord",
        DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
            RichText: &nt.TextPropertyFilter{
                Equals: discordUsername,
            },
        },
    }

    query := &nt.DatabaseQuery{
        Filter: filter,
    }

    contractorDBID := s.cfg.LeaveIntegration.Notion.ContractorDBID
    resp, err := s.client.QueryDatabase(ctx, contractorDBID, query)
    if err != nil {
        return "", err
    }

    if len(resp.Results) == 0 {
        return "", nil
    }

    return resp.Results[0].ID, nil
}

// Client exposes the Notion client for direct API calls
func (s *LeaveService) Client() *nt.Client {
    return s.client
}
```

#### Integrate into Approval Handler
```go
// In handleNotionLeaveApproved/Rejected functions:

// After approval/rejection is processed:
if err := h.updateApprovedRejectedBy(ctx, l, leavePageID, approverDiscordID); err != nil {
    l.Error(err, "failed to update approved/rejected by relation (non-critical)")
    // Don't fail the approval - this is metadata only
}
```

### Acceptance Criteria
- [ ] Multi-step lookup: Discord ID → username → contractor page ID
- [ ] Updates Notion relation field correctly
- [ ] Does not fail approval/rejection if update fails
- [ ] Logs each step for debugging
- [ ] Handles missing data gracefully
- [ ] Returns error but doesn't propagate to approval flow

### Testing Requirements
- Unit test: Full flow succeeds (updates relation)
- Unit test: Discord account not found (graceful, logs warning)
- Unit test: Contractor not found (graceful, logs warning)
- Unit test: Notion API error (graceful, logs error)
- Unit test: Empty Discord username (graceful)
- Unit test: Relation format is correct

---

## Task 13: Unit Tests - Store Method

**Priority:** High
**Estimated Effort:** 2 hours
**Dependencies:** Task 1

### Description
Write comprehensive unit tests for the new `OneByUsername` store method.

### Files to Create
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/store/discordaccount/discord_account_test.go` (or add to existing)

### Implementation Details

#### Test Cases
1. Username exists - returns account
2. Username not found - returns nil, nil
3. Database error - returns nil, error
4. Empty username - returns nil, nil
5. Whitespace handling
6. Case sensitivity

### Acceptance Criteria
- [ ] All test cases pass
- [ ] Use testhelper for database setup
- [ ] Mock database with test data
- [ ] Test happy path and error paths
- [ ] Code coverage >90%

### Testing Requirements
- Follow fortress-api testing patterns
- Use table-driven tests
- Mock database properly

---

## Task 14: Unit Tests - Service Layer Functions

**Priority:** High
**Estimated Effort:** 6-8 hours
**Dependencies:** Task 4, 5, 6, 7, 8

### Description
Write comprehensive unit tests for all new service layer functions.

### Files to Create/Modify
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/leave_test.go`

### Implementation Details

#### Test Functions
1. Property extraction helpers (Task 4)
2. GetActiveDeploymentsForContractor (Task 5)
3. ExtractStakeholdersFromDeployment (Task 6)
4. GetDiscordUsernameFromContractor (Task 7)
5. LookupContractorByEmail (Task 8)
6. LookupContractorByDiscordUsername (Task 12)

#### Mock Structures
- Mock Notion client
- Mock Notion API responses
- Mock database

### Acceptance Criteria
- [ ] All service functions have unit tests
- [ ] Mock Notion API client
- [ ] Test happy paths and error paths
- [ ] Test edge cases (empty, nil, invalid)
- [ ] Code coverage >90%
- [ ] Use table-driven test pattern

### Testing Requirements
- Follow test plan in `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512120930-notion-leave-webhook-amdl/test-cases/test-plans/overall-test-plan.md`
- Reference individual test case documents in `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512120930-notion-leave-webhook-amdl/test-cases/unit/`

---

## Task 15: Unit Tests - Handler Layer Functions

**Priority:** High
**Estimated Effort:** 6-8 hours
**Dependencies:** Task 9, 10, 11, 12

### Description
Write comprehensive unit tests for handler layer functions.

### Files to Create/Modify
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/webhook/notion_leave_test.go`

### Implementation Details

#### Test Functions
1. getDiscordMentionFromUsername (Task 9)
2. getAMDLMentionsFromDeployments (Task 10)
3. handleNotionLeaveCreated (Task 11)
4. updateApprovedRejectedBy (Task 12)

#### Mock Structures
- Mock store
- Mock service layer
- Mock Discord client
- Mock Notion client

### Acceptance Criteria
- [ ] All handler functions have unit tests
- [ ] Mock all external dependencies
- [ ] Test happy paths and error paths
- [ ] Test graceful degradation
- [ ] Code coverage >90%
- [ ] Use table-driven test pattern

### Testing Requirements
- Follow test plan documentation
- Test Discord message formatting
- Test error handling and logging

---

## Task 16: Documentation - Update ADR and Specs

**Priority:** Low
**Estimated Effort:** 1 hour
**Dependencies:** Task 11, 12

### Description
Update architecture decision records and specifications to reflect final implementation.

### Files to Modify
- `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512120930-notion-leave-webhook-amdl/planning/ADRs/001-am-dl-lookup-strategy.md`
- `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512120930-notion-leave-webhook-amdl/planning/specifications/am-dl-integration-spec.md`

### Implementation Details

#### Updates
- Mark ADR as "Accepted"
- Update spec with any implementation deviations
- Document any edge cases discovered during development

### Acceptance Criteria
- [ ] ADR status updated
- [ ] Spec reflects actual implementation
- [ ] Any deviations documented with rationale

---

## Task 17: Documentation - Update Implementation Status

**Priority:** Low
**Estimated Effort:** 30 minutes
**Dependencies:** All tasks

### Description
Create implementation status document tracking progress.

### Files to Create
- `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512120930-notion-leave-webhook-amdl/implementation/STATUS.md`

### Implementation Details

#### Status Document Structure
```markdown
# Implementation Status

**Last Updated:** YYYY-MM-DD
**Status:** In Progress / Complete

## Progress Summary
- [ ] Task 1: Database Store Layer
- [ ] Task 2: Database Migration
- ...

## Completed Tasks
- Task X: Description (YYYY-MM-DD)

## In Progress
- Task Y: Description

## Blocked Tasks
- None

## Issues Found
- Issue description and resolution
```

### Acceptance Criteria
- [ ] Status document created
- [ ] Updated as tasks complete
- [ ] Tracks blockers and issues

---

## Task 18: Manual Testing - Local Environment

**Priority:** High
**Estimated Effort:** 4 hours
**Dependencies:** Task 11, 12

### Description
Manually test the full integration in local development environment with real Notion data.

### Testing Steps

#### Setup
1. Set `NOTION_DEPLOYMENT_TRACKER_DB_ID` in `.env`
2. Run migrations: `make migrate-up`
3. Verify index created on `discord_accounts.discord_username`
4. Start application: `make dev`

#### Test Cases

**Test Case 1: Leave Request with Active Deployment**
1. Create contractor in Notion with Team Email and Discord username
2. Create active deployment with AM/DL
3. Create leave request with Team Email
4. Trigger webhook
5. Verify Discord notification includes AM/DL mentions

**Test Case 2: Leave Request without Deployments**
1. Create contractor with no active deployments
2. Create leave request
3. Trigger webhook
4. Verify notification sent without mentions

**Test Case 3: Leave Approval Flow**
1. Create leave request
2. Approve via Discord button
3. Verify "Approved/Rejected By" relation updated in Notion

**Test Case 4: Edge Cases**
- Contractor not in Notion Contractors DB
- Discord username not in fortress DB
- Multiple deployments for same contractor
- Same person as both AM and DL

### Acceptance Criteria
- [ ] All test cases pass
- [ ] Discord notifications formatted correctly
- [ ] Mentions work in Discord (clickable)
- [ ] Notion relations updated correctly
- [ ] Logs show expected behavior
- [ ] No errors in application logs

### Manual Testing Checklist
- [ ] Setup complete
- [ ] Test case 1: Passed
- [ ] Test case 2: Passed
- [ ] Test case 3: Passed
- [ ] Test case 4: All edge cases passed
- [ ] Performance acceptable (<5s per request)

---

## Task 19: Code Review Preparation

**Priority:** Medium
**Estimated Effort:** 2 hours
**Dependencies:** All implementation tasks

### Description
Prepare code for review by running linters, checking test coverage, and cleaning up.

### Checklist

#### Code Quality
- [ ] Run `make lint` - all checks pass
- [ ] Run `make test` - all tests pass
- [ ] Code coverage >90% for new code
- [ ] No commented-out code
- [ ] No debug logging left in code
- [ ] Error messages are clear and actionable

#### Documentation
- [ ] All functions have comments
- [ ] Complex logic has inline comments
- [ ] README updated if needed
- [ ] Migration documented

#### Testing
- [ ] All unit tests pass
- [ ] Manual testing complete
- [ ] Edge cases covered
- [ ] Error handling tested

### Acceptance Criteria
- [ ] Lint checks pass
- [ ] All tests pass
- [ ] Coverage meets target
- [ ] Code is clean and documented

---

## Task 20: Deployment Preparation

**Priority:** Medium
**Estimated Effort:** 1 hour
**Dependencies:** Task 2, 19

### Description
Prepare for production deployment with environment variables and migration plan.

### Deployment Checklist

#### Environment Variables
- [ ] `NOTION_DEPLOYMENT_TRACKER_DB_ID` added to production config
- [ ] Value verified with Notion workspace
- [ ] Environment variable documented

#### Migration Plan
- [ ] Migration file tested locally
- [ ] Migration rollback tested
- [ ] Migration will run on deployment
- [ ] Index creation time estimated (should be fast)

#### Rollout Strategy
- [ ] Monitor logs for errors during first hour
- [ ] Validate Discord notifications are sent
- [ ] Check Notion relation updates work
- [ ] Have rollback plan ready

#### Monitoring
- [ ] Log queries to identify issues
- [ ] Monitor Notion API rate limits
- [ ] Monitor database query performance
- [ ] Check Discord API errors

### Acceptance Criteria
- [ ] Environment variables configured
- [ ] Migration ready to deploy
- [ ] Rollout plan documented
- [ ] Monitoring plan in place

---

## Summary

### Task Dependencies Visualization

```
Task 1 (Store)      Task 2 (Migration)    Task 3 (Config)       Task 4 (Helpers)
    |                                           |                     |
    |                                           +---------+-----------+
    |                                                     |
    |                                                     v
    |                                          +----------+-----------+
    |                                          |                      |
    |                                          v                      v
    |                                      Task 5              Task 6, 7, 8
    |                                   (Deployments)         (Extraction)
    |                                          |                      |
    v                                          +----------+-----------+
Task 9                                                    |
(Mention)                                                 v
    |                                                 Task 10
    +------------------------------------------------>(Orchestrate)
                                                          |
                                                          v
                                                      Task 11
                                                     (Handler)
                                                          |
                                                          v
                                           Task 12 <------+
                                        (Update Relation)
                                                          |
                                                          v
                                        Task 13, 14, 15 (Tests)
                                                          |
                                                          v
                                              Task 16, 17 (Docs)
                                                          |
                                                          v
                                             Task 18 (Manual Testing)
                                                          |
                                                          v
                                              Task 19 (Code Review)
                                                          |
                                                          v
                                              Task 20 (Deployment)
```

### Estimated Timeline

**Week 1 (Core Implementation)**
- Day 1-2: Tasks 1-4 (Foundation)
- Day 3-4: Tasks 5-8 (Service Layer)
- Day 5: Tasks 9-10 (Handler Layer)

**Week 2 (Integration & Testing)**
- Day 1-2: Tasks 11-12 (Integration)
- Day 3-4: Tasks 13-15 (Unit Tests)
- Day 5: Tasks 16-18 (Documentation & Manual Testing)

**Week 3 (Review & Deploy)**
- Day 1: Task 19 (Code Review)
- Day 2: Task 20 (Deployment Prep)
- Day 3-5: Review cycles and deployment

### Total Estimated Effort
- Implementation: 35-45 hours
- Testing: 20-24 hours
- Documentation: 5-7 hours
- Total: 60-76 hours (approximately 2-3 weeks)

### Critical Path
1. Task 3 (Config) - blocks service layer
2. Task 4 (Helpers) - blocks property extraction
3. Tasks 5-8 (Service Layer) - blocks orchestration
4. Task 10 (Orchestrate) - blocks handler
5. Task 11 (Handler) - blocks manual testing
6. Tasks 13-15 (Tests) - blocks code review

### Risk Mitigation
- **Risk**: Notion API changes property structure
  - **Mitigation**: Extensive unit tests with real property structures
- **Risk**: Database performance issues
  - **Mitigation**: Index on discord_username (Task 2)
- **Risk**: Complex rollup extraction logic
  - **Mitigation**: Detailed unit tests (Task 14)
- **Risk**: Silent failures in production
  - **Mitigation**: Comprehensive logging at all levels

---

## Notes

- **NO IMPLEMENTATION**: This is a task breakdown only. Do not execute these tasks.
- **Task Order**: Follow dependency order for implementation
- **Testing**: Write tests as you implement each function
- **Logging**: Add comprehensive logging at DEBUG, INFO, WARN, ERROR levels
- **Error Handling**: Implement graceful degradation - don't fail webhooks on optional features
- **Code Review**: Follow fortress-api coding standards and patterns
