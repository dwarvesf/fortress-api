# Technical Specification: AM/DL Integration for Notion Leave Webhook

**Version:** 1.0
**Date:** 2025-12-12
**Status:** Draft

## Overview

This specification details the technical implementation for automatically fetching Account Managers (AM) and Delivery Leads (DL) from the Deployment Tracker when a leave request is created in Notion, replacing the manual "Assignees" multi-select field.

## Objectives

1. Automatically identify AM/DL from active deployments
2. Send Discord notifications with AM/DL mentions
3. Auto-fill "Approved/Rejected By" relation field on approval/rejection
4. Maintain backward compatibility during transition
5. Provide robust error handling and logging

## Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│                 Notion Leave Request Created                 │
│                    (Team Email provided)                     │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 1: Lookup Contractor by Team Email                     │
│                                                              │
│ Input:  leave.TeamEmail (email)                             │
│ Query:  Contractors DB                                      │
│ Filter: Team Email (email) = leave.TeamEmail                │
│ Output: contractor_page_id (string)                         │
│                                                              │
│ Error Handling: Return empty string if not found            │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 2: Query Active Deployments                            │
│                                                              │
│ Input:  contractor_page_id                                  │
│ Query:  Deployment Tracker DB                               │
│ Filter: AND                                                 │
│   - Contractor (relation) contains contractor_page_id       │
│   - Deployment Status (status) = "Active"                   │
│ Output: []deployment_page (array of pages)                  │
│                                                              │
│ Error Handling: Return empty array if none found            │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 3: Extract AM/DL from Each Deployment                  │
│                                                              │
│ For each deployment_page:                                   │
│   1. Check Override AM (relation)                           │
│      - If set: add to am_list                               │
│   2. If no Override AM, check Account Managers (rollup)     │
│      - Extract relation IDs from rollup array               │
│      - Add to am_list                                       │
│   3. Check Override DL (relation)                           │
│      - If set: add to dl_list                               │
│   4. If no Override DL, check Delivery Leads (rollup)       │
│      - Extract relation IDs from rollup array               │
│      - Add to dl_list                                       │
│                                                              │
│ Output: am_list []string, dl_list []string                  │
│                                                              │
│ Error Handling: Skip deployment if property extraction fails│
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 4: Deduplicate Stakeholders                            │
│                                                              │
│ Input:  am_list, dl_list                                    │
│ Process:                                                     │
│   - Combine am_list and dl_list into stakeholders           │
│   - Remove duplicates by contractor_page_id                 │
│ Output: unique_stakeholders []string                        │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 5: Get Discord Usernames from Notion                   │
│                                                              │
│ For each stakeholder_page_id in unique_stakeholders:        │
│   1. Fetch contractor page by ID                            │
│   2. Extract "Discord" property (rich_text)                 │
│   3. Concatenate RichText parts to get username             │
│   4. Add to discord_usernames list                          │
│                                                              │
│ Output: discord_usernames []string                          │
│                                                              │
│ Error Handling: Skip if page fetch fails or Discord empty   │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 6: Convert Discord Usernames to Mentions               │
│                                                              │
│ For each discord_username in discord_usernames:             │
│   1. Query fortress DB: discord_accounts table              │
│      WHERE discord_username = discord_username              │
│   2. Get discord_id from result                             │
│   3. Format as "<@discord_id>"                              │
│   4. Add to mentions list                                   │
│                                                              │
│ Output: mentions []string                                   │
│                                                              │
│ Error Handling: Skip if username not found in DB            │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 7: Send Discord Notification                           │
│                                                              │
│ Input:  mentions []string                                   │
│ Format: "Assignees: <@id1> <@id2> <@id3>"                   │
│ Send:   Discord message with embed + buttons + mentions     │
└─────────────────────────────────────────────────────────────┘
```

## Notion API Calls

### Call 1: Lookup Contractor by Email

**Method:** `QueryDatabase`
**Database:** Contractors (`NOTION_CONTRACTOR_DB_ID`)
**Filter:**
```go
filter := &nt.DatabaseQueryFilter{
    Property: "Team Email",
    DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
        Email: &nt.TextPropertyFilter{
            Equals: teamEmail,
        },
    },
}
```

**Expected Response:**
```json
{
  "results": [
    {
      "id": "contractor-page-id",
      "properties": {
        "Team Email": {
          "type": "email",
          "email": "employee@d.foundation"
        }
      }
    }
  ]
}
```

**Error Scenarios:**
- No results: Employee not in Contractors DB (log warning, skip AM/DL lookup)
- Multiple results: Take first result (log warning if count > 1)
- API error: Log error, skip AM/DL lookup

### Call 2: Query Active Deployments

**Method:** `QueryDatabase`
**Database:** Deployment Tracker (`NOTION_DEPLOYMENT_TRACKER_DB_ID`)
**Filter:**
```go
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
```

**Expected Response:**
```json
{
  "results": [
    {
      "id": "deployment-page-id-1",
      "properties": {
        "Contractor": {
          "type": "relation",
          "relation": [{"id": "contractor-page-id"}]
        },
        "Deployment Status": {
          "type": "status",
          "status": {"name": "Active"}
        },
        "Override AM": {
          "type": "relation",
          "relation": [{"id": "am-contractor-page-id"}]
        },
        "Override DL": {
          "type": "relation",
          "relation": []
        },
        "Account Managers": {
          "type": "rollup",
          "rollup": {
            "type": "array",
            "array": [
              {
                "type": "relation",
                "relation": [{"id": "default-am-page-id"}]
              }
            ]
          }
        },
        "Delivery Leads": {
          "type": "rollup",
          "rollup": {
            "type": "array",
            "array": [
              {
                "type": "relation",
                "relation": [{"id": "default-dl-page-id"}]
              }
            ]
          }
        }
      }
    }
  ]
}
```

**Error Scenarios:**
- No results: Employee has no active deployments (log info, use empty mentions)
- API error: Log error, use empty mentions
- Invalid property types: Log error, skip that property

### Call 3: Fetch Contractor Page for Discord Username

**Method:** `FindPageByID`
**Input:** Contractor page ID from AM/DL extraction
**Expected Response:**
```json
{
  "id": "contractor-page-id",
  "properties": {
    "Discord": {
      "type": "rich_text",
      "rich_text": [
        {
          "type": "text",
          "text": {"content": "username"},
          "plain_text": "username"
        }
      ]
    }
  }
}
```

**Error Scenarios:**
- Page not found: Log warning, skip this stakeholder
- Discord field empty: Log info, skip this stakeholder
- API error: Log error, skip this stakeholder

### Call 4: Update Approved/Rejected By Relation

**Method:** `UpdatePage`
**Input:** Leave request page ID, approver contractor page ID
**Update Params:**
```go
updateParams := nt.UpdatePageParams{
    DatabasePageProperties: nt.DatabasePageProperties{
        "Approved/Rejected By": nt.DatabasePageProperty{
            Relation: []nt.Relation{
                {ID: approverContractorPageID},
            },
        },
    },
}
```

**Error Scenarios:**
- Approver not found in Contractors DB: Log warning, skip update
- API error: Log error, skip update (approval still processed)

## Property Extraction Details

### Relation Property (Override AM/DL)

**Type:** `relation`
**Structure:** Array of relation objects
**Extraction Pattern:**
```go
func extractFirstRelationID(props nt.DatabasePageProperties, propName string) string {
    if prop, ok := props[propName]; ok && len(prop.Relation) > 0 {
        return prop.Relation[0].ID
    }
    return ""
}
```

**Usage:**
- Extract "Override AM" → first relation ID or empty string
- Extract "Override DL" → first relation ID or empty string

### Rollup Property (Account Managers/Delivery Leads)

**Type:** `rollup`
**Subtype:** `array` of relations
**Structure:** Nested array of relation objects
**Extraction Pattern:**
```go
func extractRollupRelations(props nt.DatabasePageProperties, propName string) []string {
    var relationIDs []string

    if prop, ok := props[propName]; ok {
        if prop.Rollup.Type == "array" && len(prop.Rollup.Array) > 0 {
            for _, item := range prop.Rollup.Array {
                // Each item should be a relation property
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
```

**Note:** Rollup array structure can vary. The rollup aggregates relations from linked pages, so each array item is itself a relation property.

### Rich Text Property (Discord Username)

**Type:** `rich_text`
**Structure:** Array of rich text objects
**Extraction Pattern:**
```go
func extractRichText(props nt.DatabasePageProperties, propName string) string {
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

**Usage:**
- Extract "Discord" → concatenated plain text, trimmed

## Database Schema Changes

### New Store Method: DiscordAccount.OneByUsername

**Location:** `pkg/store/discord_account/`

**Interface Addition:**
```go
// In pkg/store/discord_account/interface.go
type IDiscordAccountStore interface {
    // ... existing methods ...
    OneByUsername(db *gorm.DB, username string) (*model.DiscordAccount, error)
}
```

**Implementation Pattern:**
```go
// In pkg/store/discord_account/discord_account.go
func (s *store) OneByUsername(db *gorm.DB, username string) (*model.DiscordAccount, error) {
    var discordAccount model.DiscordAccount

    err := db.Where("discord_username = ?", username).First(&discordAccount).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil // Return nil, nil for not found (graceful)
        }
        return nil, err
    }

    return &discordAccount, nil
}
```

**Database Query:**
```sql
SELECT * FROM discord_accounts WHERE discord_username = $1 LIMIT 1;
```

**Index Recommendation:**
```sql
CREATE INDEX idx_discord_accounts_username ON discord_accounts(discord_username);
```

**Rationale:** Discord username lookups will happen on every leave request creation. Index improves query performance.

## Service Layer Functions

### Function 1: GetActiveDeploymentsForContractor

**Location:** `pkg/service/notion/leave.go`

**Signature:**
```go
func (s *LeaveService) GetActiveDeploymentsForContractor(
    ctx context.Context,
    contractorPageID string,
) ([]nt.Page, error)
```

**Purpose:** Query Deployment Tracker for active deployments of a contractor

**Returns:**
- Array of deployment pages
- Empty array if none found
- Error if API call fails

**Error Handling:**
- Log contractor page ID for debugging
- Log query filter for debugging
- Return empty array on not found (graceful)
- Return error on API failure

### Function 2: ExtractStakeholdersFromDeployment

**Location:** `pkg/service/notion/leave.go`

**Signature:**
```go
func (s *LeaveService) ExtractStakeholdersFromDeployment(
    deploymentPage nt.Page,
) ([]string, error)
```

**Purpose:** Extract AM/DL contractor page IDs from a deployment page

**Logic:**
1. Extract "Override AM" relation
2. If empty, extract "Account Managers" rollup
3. Extract "Override DL" relation
4. If empty, extract "Delivery Leads" rollup
5. Combine and return unique page IDs

**Returns:**
- Array of contractor page IDs (deduplicated within deployment)
- Empty array if no stakeholders found
- Error if property extraction fails critically

**Error Handling:**
- Log each property extraction attempt
- Continue if individual property fails (partial data)
- Return error only if page properties are completely invalid

### Function 3: GetDiscordUsernameFromContractor

**Location:** `pkg/service/notion/leave.go`

**Signature:**
```go
func (s *LeaveService) GetDiscordUsernameFromContractor(
    ctx context.Context,
    contractorPageID string,
) (string, error)
```

**Purpose:** Fetch Discord username from a contractor page

**Logic:**
1. Call `FindPageByID(contractorPageID)`
2. Extract "Discord" rich_text property
3. Return concatenated plain text

**Returns:**
- Discord username (string)
- Empty string if not set
- Error if API call fails

**Error Handling:**
- Log page fetch attempt
- Return empty string if Discord field not found (graceful)
- Return error on API failure

## Handler Layer Changes

### Modified Function: handleNotionLeaveCreated

**Location:** `pkg/handler/webhook/notion_leave.go`

**Current Flow:**
```
1. Validate employee exists
2. Validate dates
3. Extract assignees from multi-select
4. Convert to Discord mentions
5. Send notification
```

**New Flow:**
```
1. Validate employee exists
2. Validate dates
3. Get AM/DL from deployments (NEW)
   a. Lookup contractor by Team Email
   b. Query active deployments
   c. Extract stakeholders from deployments
   d. Get Discord usernames from Notion
   e. Convert usernames to Discord IDs
4. Send notification with AM/DL mentions
```

### New Handler Function: getDiscordMentionFromUsername

**Location:** `pkg/handler/webhook/notion_leave.go`

**Signature:**
```go
func (h *handler) getDiscordMentionFromUsername(
    l logger.Logger,
    discordUsername string,
) string
```

**Purpose:** Convert Discord username to Discord mention format

**Logic:**
1. Call `h.store.DiscordAccount.OneByUsername(db, discordUsername)`
2. Get DiscordID from result
3. Format as `<@discord_id>`

**Returns:**
- Discord mention string `<@id>` or empty string

**Error Handling:**
- Log username lookup
- Return empty string if not found (graceful)
- Log mismatch for manual correction

### New Handler Function: updateApprovedRejectedBy

**Location:** `pkg/handler/webhook/notion_leave.go`

**Signature:**
```go
func (h *handler) updateApprovedRejectedBy(
    ctx context.Context,
    l logger.Logger,
    leavePageID string,
    approverDiscordID string,
) error
```

**Purpose:** Update "Approved/Rejected By" relation when leave is approved/rejected

**Logic:**
1. Get Discord account by discord_id from fortress DB
2. Get discord_username from Discord account
3. Query Notion Contractors DB for contractor with matching Discord username
4. Update leave request page with contractor page ID relation

**Error Handling:**
- Log each step
- Skip update if approver not found (log warning)
- Don't fail approval/rejection if update fails

## Configuration Changes

### Environment Variables

**New Variable:**
```bash
NOTION_DEPLOYMENT_TRACKER_DB_ID=2b864b29b84c80799568dc17685f4f33
```

**Config Struct Update:**
```go
// In pkg/config/config.go
type NotionConfig struct {
    Secret    string
    Databases struct {
        Contractor         string `envconfig:"NOTION_CONTRACTOR_DB_ID" required:"true"`
        DeploymentTracker  string `envconfig:"NOTION_DEPLOYMENT_TRACKER_DB_ID" required:"true"`
        // ... other databases
    }
}
```

## Error Handling Strategy

### Graceful Degradation

**Philosophy:** Non-critical failures should not block the webhook

**Implementation:**
- Return empty mentions if AM/DL lookup fails
- Log warnings for manual follow-up
- Still send Discord notification (without mentions if necessary)
- Still process approval/rejection even if relation update fails

### Error Levels

**Critical (Return 500):**
- Cannot parse webhook payload
- Cannot verify webhook signature
- Database connection failure

**Warning (Log, Continue):**
- Contractor not found in Notion
- No active deployments found
- Discord username not in fortress DB
- Notion API call fails for optional fields

**Info (Log):**
- Empty Discord field on contractor page
- No Override AM/DL set (expected, use rollup)

### Logging Requirements

**Debug Level:**
- Each API call (input parameters)
- Each property extraction (property name, value)
- Each conversion step (email → ID → mention)

**Info Level:**
- Webhook received
- Notification sent
- Approval/rejection processed

**Warning Level:**
- Expected data missing (contractor not found, no deployments)
- Discord username mismatch
- Optional field updates failed

**Error Level:**
- API call failures
- Unexpected property types
- Database query errors

## Testing Considerations

### Unit Tests

**Test Cases:**

1. **Property Extraction**
   - Extract relation with value
   - Extract relation when empty
   - Extract rollup with array of relations
   - Extract rollup when empty
   - Extract rich_text with multiple parts
   - Extract rich_text when empty

2. **Store Method**
   - `OneByUsername` finds existing account
   - `OneByUsername` returns nil for non-existent username
   - `OneByUsername` handles database errors

3. **Stakeholder Extraction**
   - Override AM/DL present → use override
   - Override AM/DL absent → use rollup
   - Both override and rollup empty → return empty array
   - Mixed: some override, some rollup

### Integration Tests

**Test Scenarios:**

1. **Full Flow with Active Deployments**
   - Create leave request with Team Email
   - Verify contractor lookup
   - Verify deployment query
   - Verify stakeholder extraction
   - Verify Discord notification with mentions

2. **No Active Deployments**
   - Create leave request for contractor with no active deployments
   - Verify empty mentions
   - Verify notification still sent

3. **Contractor Not Found**
   - Create leave request with unknown Team Email
   - Verify graceful handling
   - Verify notification sent without mentions

4. **Approval Flow**
   - Approve leave request via Discord
   - Verify "Approved/Rejected By" relation updated
   - Verify approval processed even if relation update fails

### Test Data Requirements

**Notion Test Data:**
- Contractor with Team Email
- Active deployment with Override AM/DL
- Active deployment with rollup AM/DL only
- Contractor with Discord username
- Contractor without Discord username

**Fortress Test Data:**
- Employee with team email
- Discord account with username matching Notion
- Discord account with discord_id

## Migration and Rollout

### Phase 1: Implementation
- Add new store method
- Add service layer functions
- Update handler to use new flow
- Keep "Assignees" field code for comparison

### Phase 2: Testing
- Test in development with real Notion data
- Compare automated AM/DL with manual assignees
- Log discrepancies

### Phase 3: Deployment
- Deploy to production
- Monitor logs for errors
- Validate Discord notifications

### Phase 4: Cleanup
- Remove "Assignees" field code after validation period
- Update documentation

## Performance Considerations

### API Call Volume

Per leave request creation:
- 1 call: Lookup contractor by email
- 1 call: Query active deployments
- N calls: Fetch contractor pages for AM/DL (where N = unique stakeholders)

**Optimization:**
- Batch contractor page fetches if Notion API supports it (future)
- Cache contractor Discord usernames (future)

### Database Query Volume

Per leave request creation:
- 1 query: Employee by email (existing)
- M queries: Discord account by username (where M = unique Discord usernames)

**Optimization:**
- Add index on discord_username (recommended)
- Batch Discord account queries with IN clause (future)

### Expected Load

- Leave requests: ~10-50 per month
- API calls per request: ~5-10
- Database queries per request: ~5-10

**Conclusion:** Current approach is acceptable for expected load. Optimization can be deferred.

## Security Considerations

1. **Webhook Signature Verification:** Already implemented
2. **Database Query Sanitization:** GORM handles parameterization
3. **Discord ID Validation:** Discord API validates mention format
4. **Notion API Token:** Stored in environment variables

No new security concerns introduced.

## References

- ADR: `/docs/sessions/202512120930-notion-leave-webhook-amdl/planning/ADRs/001-am-dl-lookup-strategy.md`
- Requirements: `/docs/sessions/202512120930-notion-leave-webhook-amdl/requirements/requirements.md`
- Research: `/docs/sessions/202512120930-notion-leave-webhook-amdl/research/notion-patterns.md`
- Current Implementation: `/pkg/handler/webhook/notion_leave.go`
- Notion Leave Service: `/pkg/service/notion/leave.go`
- Discord Account Store: `/pkg/store/discord_account/`
