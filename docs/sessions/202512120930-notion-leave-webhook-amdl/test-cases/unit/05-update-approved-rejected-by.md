# Test Case Design: UpdateApprovedRejectedBy

**Function Under Test:** `handler.updateApprovedRejectedBy(ctx context.Context, l logger.Logger, leavePageID string, approverDiscordID string) error`

**Location:** `pkg/handler/webhook/notion_leave.go`

**Purpose:** Update the "Approved/Rejected By" relation field on a Notion leave request page when a leave is approved or rejected via Discord button, linking it to the contractor who performed the action.

## Test Strategy

### Approach
- **Pattern:** Table-driven tests with multi-step flow scenarios
- **Dependencies:** Mock database, mock Notion API client
- **Validation:** Assert correct lookup chain and Notion update
- **Focus:** Error handling, graceful degradation, relation update correctness

### Test Data Requirements
- Mock `discord_accounts` table with test data
- Mock Notion Contractors database with Discord usernames
- Mock Notion API responses for UpdatePage
- Valid and invalid Discord IDs and usernames

## Workflow Steps to Test

1. **Get Discord Account by Discord ID** (fortress DB)
2. **Extract Discord Username** from account
3. **Query Notion Contractors DB** by Discord username
4. **Update Leave Request Page** with contractor page ID relation

## Test Cases

### TC-5.1: Successfully Update Approved/Rejected By - Full Flow

**Given:**
- Leave page ID: `leave-page-123`
- Approver Discord ID: `987654321012345678`
- Fortress DB has Discord account:
  - `discord_id`: `987654321012345678`
  - `discord_username`: `"approver_user"`
- Notion Contractors DB has contractor:
  - "Discord" property: `"approver_user"`
  - Page ID: `contractor-page-456`

**When:**
- Call `updateApprovedRejectedBy(ctx, logger, "leave-page-123", "987654321012345678")`

**Then:**
- **Step 1:** Should call `store.DiscordAccount.OneByDiscordID(db, "987654321012345678")`
  - Retrieves: Discord account with `discord_username = "approver_user"`
- **Step 2:** Should query Notion Contractors DB:
  - Filter: "Discord" (rich_text) equals `"approver_user"`
  - Finds: Contractor page ID `contractor-page-456`
- **Step 3:** Should call `client.UpdatePage()` with:
  - Page ID: `leave-page-123`
  - Properties: `{"Approved/Rejected By": {"relation": [{"id": "contractor-page-456"}]}}`
- Should return: `nil` error
- Should log INFO: "updated approved/rejected by: leave_page_id=leave-page-123 approver_contractor_id=contractor-page-456"

**Mock Data:**

Discord Account:
```go
&model.DiscordAccount{
    ID:              "uuid-1",
    DiscordID:       "987654321012345678",
    DiscordUsername: "approver_user",
}
```

Notion Contractor Query Response:
```json
{
  "results": [
    {
      "id": "contractor-page-456",
      "properties": {
        "Discord": {
          "type": "rich_text",
          "rich_text": [{"plain_text": "approver_user"}]
        }
      }
    }
  ]
}
```

Notion Update Call:
```go
updateParams := nt.UpdatePageParams{
    DatabasePageProperties: nt.DatabasePageProperties{
        "Approved/Rejected By": nt.DatabasePageProperty{
            Relation: []nt.Relation{
                {ID: "contractor-page-456"},
            },
        },
    },
}
```

---

### TC-5.2: Approver Discord Account Not Found - Skip Update

**Given:**
- Leave page ID: `leave-page-123`
- Approver Discord ID: `unknown-discord-id`
- Fortress DB has no matching Discord account

**When:**
- Call `updateApprovedRejectedBy(ctx, logger, "leave-page-123", "unknown-discord-id")`

**Then:**
- **Step 1:** Should call `store.DiscordAccount.OneByDiscordID(db, "unknown-discord-id")`
  - Returns: `nil` (not found)
- Should return: `nil` error (graceful degradation)
- Should NOT query Notion
- Should NOT update Notion page
- Should log WARNING: "discord account not found for approver: discord_id=unknown-discord-id, skipping approved/rejected by update"

**Rationale:** Don't fail approval/rejection if relation update fails

---

### TC-5.3: Approver Discord Username Empty - Skip Update

**Given:**
- Leave page ID: `leave-page-123`
- Approver Discord ID: `987654321012345678`
- Fortress DB has Discord account but `discord_username` is empty:
  ```go
  &model.DiscordAccount{
      DiscordID:       "987654321012345678",
      DiscordUsername: "", // Empty
  }
  ```

**When:**
- Call `updateApprovedRejectedBy(ctx, logger, "leave-page-123", "987654321012345678")`

**Then:**
- **Step 1:** Should retrieve Discord account
- **Step 2:** Should detect empty `discord_username`
- Should return: `nil` error (graceful)
- Should NOT query Notion
- Should log WARNING: "discord username is empty for approver: discord_id=987654321012345678"

---

### TC-5.4: Contractor Not Found in Notion - Skip Update

**Given:**
- Leave page ID: `leave-page-123`
- Approver Discord ID: `987654321012345678`
- Fortress DB has Discord account: `discord_username = "approver_user"`
- Notion Contractors DB has NO matching contractor with Discord username `"approver_user"`

**When:**
- Call `updateApprovedRejectedBy(ctx, logger, "leave-page-123", "987654321012345678")`

**Then:**
- **Step 1:** Should retrieve Discord account successfully
- **Step 2:** Should query Notion Contractors DB
  - Filter: "Discord" (rich_text) equals `"approver_user"`
  - Returns: Empty results `[]`
- Should return: `nil` error (graceful)
- Should NOT update Notion page
- Should log WARNING: "contractor not found in notion for discord username: approver_user"

**Rationale:** Approver may not be in Contractors DB (e.g., external approver)

---

### TC-5.5: Notion Query Error - Skip Update

**Given:**
- Leave page ID: `leave-page-123`
- Approver Discord ID: `987654321012345678`
- Fortress DB has Discord account
- Notion API returns error when querying Contractors DB (network error, rate limit)

**When:**
- Call `updateApprovedRejectedBy(ctx, logger, "leave-page-123", "987654321012345678")`

**Then:**
- **Step 1:** Should retrieve Discord account successfully
- **Step 2:** Should attempt to query Notion Contractors DB
  - Returns: Error
- Should return: `nil` error (graceful degradation)
- Should NOT update Notion page
- Should log ERROR: "failed to query notion contractors db: discord_username=approver_user error=%v"

**Mock Error:**
```go
errors.New("notion API: rate limit exceeded")
```

---

### TC-5.6: Notion Update Page Error - Return Error

**Given:**
- Leave page ID: `leave-page-123`
- Approver Discord ID: `987654321012345678`
- All lookups successful (Discord account found, contractor found)
- Notion `UpdatePage` API call fails (network error, API error)

**When:**
- Call `updateApprovedRejectedBy(ctx, logger, "leave-page-123", "987654321012345678")`

**Then:**
- **Step 1-2:** Should complete successfully
- **Step 3:** Should attempt `client.UpdatePage()`
  - Returns: Error
- Should return: `nil` error (graceful, don't fail approval)
- Should log ERROR: "failed to update approved/rejected by in notion: leave_page_id=leave-page-123 error=%v"

**Rationale:** Approval/rejection should still succeed even if Notion update fails

**Alternative Design Decision:**
- Could return error to indicate partial failure
- Recommended: Return nil to ensure approval processing continues

---

### TC-5.7: Empty Leave Page ID - Return Error

**Given:**
- Leave page ID: `""` (empty)
- Approver Discord ID: `987654321012345678`

**When:**
- Call `updateApprovedRejectedBy(ctx, logger, "", "987654321012345678")`

**Then:**
- Should return: Error "leave page ID is required"
- Should NOT perform any lookups
- Should log WARNING: "empty leave page ID provided"

**Rationale:** Input validation

---

### TC-5.8: Empty Approver Discord ID - Return Error

**Given:**
- Leave page ID: `leave-page-123`
- Approver Discord ID: `""` (empty)

**When:**
- Call `updateApprovedRejectedBy(ctx, logger, "leave-page-123", "")`

**Then:**
- Should return: Error "approver discord ID is required"
- Should NOT perform any lookups
- Should log WARNING: "empty approver discord ID provided"

**Rationale:** Input validation

---

### TC-5.9: Database Query Error - Skip Update

**Given:**
- Leave page ID: `leave-page-123`
- Approver Discord ID: `987654321012345678`
- Database query for Discord account returns error (connection lost)

**When:**
- Call `updateApprovedRejectedBy(ctx, logger, "leave-page-123", "987654321012345678")`

**Then:**
- **Step 1:** Should call `store.DiscordAccount.OneByDiscordID()`
  - Returns: Error
- Should return: `nil` error (graceful)
- Should NOT query Notion
- Should log ERROR: "failed to query discord account: discord_id=987654321012345678 error=%v"

---

### TC-5.10: Context Cancellation - Respect Context

**Given:**
- Leave page ID: `leave-page-123`
- Approver Discord ID: `987654321012345678`
- Context is cancelled before Notion update

**When:**
- Call `updateApprovedRejectedBy(cancelledCtx, logger, "leave-page-123", "987654321012345678")`

**Then:**
- Should attempt operations with cancelled context
- Should return: `nil` or context error (depending on design)
- Should NOT hang or block
- Should log context cancellation

**Note:** Context may be cancelled during Notion API call

---

### TC-5.11: Multiple Contractors Found - Use First Result

**Given:**
- Leave page ID: `leave-page-123`
- Approver Discord ID: `987654321012345678`
- Fortress DB has Discord account: `discord_username = "approver_user"`
- Notion Contractors DB returns multiple contractors with same Discord username (data issue)

**When:**
- Call `updateApprovedRejectedBy(ctx, logger, "leave-page-123", "987654321012345678")`

**Then:**
- **Step 2:** Should receive multiple contractor pages from Notion query
- Should use first result's page ID
- Should update leave request with first contractor's page ID
- Should log WARNING: "multiple contractors found with discord username: approver_user, using first result"

**Notion Response:**
```json
{
  "results": [
    {"id": "contractor-page-1", "properties": {...}},
    {"id": "contractor-page-2", "properties": {...}}
  ]
}
```

---

### TC-5.12: Nil Logger - Handle Gracefully

**Given:**
- Leave page ID: `leave-page-123`
- Approver Discord ID: `987654321012345678`
- Logger is nil

**When:**
- Call `updateApprovedRejectedBy(ctx, nil, "leave-page-123", "987654321012345678")`

**Then:**
- Should NOT panic
- Should still perform all lookups and updates
- Should skip logging (no-op if logger is nil)
- Should return result correctly

**Note:** Defensive programming

---

### TC-5.13: Notion Query Filter Correctness

**Given:**
- Discord username: `"approver_user"`

**When:**
- Querying Notion Contractors DB

**Then:**
- Should construct filter:
  ```go
  filter := &nt.DatabaseQueryFilter{
      Property: "Discord",
      DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
          RichText: &nt.TextPropertyFilter{
              Equals: "approver_user",
          },
      },
  }
  ```
- Should query correct database: `cfg.LeaveIntegration.Notion.ContractorDBID`
- Should set PageSize: 1 (only need first result)

**Note:** Rich text property uses TextPropertyFilter, not RichTextPropertyFilter

---

### TC-5.14: Relation Update Format Correctness

**Given:**
- Contractor page ID: `contractor-page-456`
- Leave page ID: `leave-page-123`

**When:**
- Updating Notion leave request page

**Then:**
- Should call `client.UpdatePage()` with correct structure:
  ```go
  updateParams := nt.UpdatePageParams{
      DatabasePageProperties: nt.DatabasePageProperties{
          "Approved/Rejected By": nt.DatabasePageProperty{
              Relation: []nt.Relation{
                  {ID: "contractor-page-456"},
              },
          },
      },
  }
  ```
- Relation array should contain single relation with contractor page ID
- Property name must match exactly: `"Approved/Rejected By"`

---

## Helper Function: lookupContractorByDiscordUsername

**Signature:** `lookupContractorByDiscordUsername(ctx context.Context, discordUsername string) (string, error)`

**Purpose:** Query Notion Contractors DB for contractor page ID by Discord username

### Test Cases for Helper

#### TC-5.15: Helper - Find Contractor by Discord Username

**Given:**
- Discord username: `"approver_user"`
- Notion has matching contractor

**When:**
- Call `lookupContractorByDiscordUsername(ctx, "approver_user")`

**Then:**
- Should query Contractors DB with Discord rich_text filter
- Should return: contractor page ID
- Should return: nil error

---

#### TC-5.16: Helper - Contractor Not Found

**Given:**
- Discord username: `"unknown_user"`
- Notion has no matching contractor

**When:**
- Call `lookupContractorByDiscordUsername(ctx, "unknown_user")`

**Then:**
- Should query Contractors DB
- Should return: `""` (empty string)
- Should return: nil error (not found is not an error)

---

## Mock Setup Requirements

### Mock Discord Account Store
```go
type MockDiscordAccountStore struct {
    OneByDiscordIDFunc func(db *gorm.DB, discordID string) (*model.DiscordAccount, error)
}
```

### Mock Notion Client
```go
type MockNotionClient struct {
    QueryDatabaseFunc func(ctx context.Context, dbID string, query *nt.DatabaseQuery) (*nt.DatabaseQueryResponse, error)
    UpdatePageFunc    func(ctx context.Context, pageID string, params nt.UpdatePageParams) (*nt.Page, error)
}
```

### Mock Assertions
- Verify `OneByDiscordID` called with correct Discord ID
- Verify Contractors DB query called with correct filter
- Verify `UpdatePage` called with correct relation structure
- Verify error handling at each step

## Error Handling Strategy

### Graceful Degradation (Return nil)
1. Discord account not found
2. Discord username is empty
3. Contractor not found in Notion
4. Notion query error
5. Notion update error

**Rationale:** The approval/rejection should still succeed even if we cannot update the "Approved/Rejected By" field. This is metadata enhancement, not critical functionality.

### Input Validation (Return error)
1. Empty leave page ID
2. Empty approver Discord ID

**Alternative Design:**
- Could treat input validation as warnings and gracefully skip
- Recommended: Return error for invalid inputs to catch caller bugs

## Logging Assertions

### DEBUG Level
- "updating approved/rejected by: leave_page_id=%s approver_discord_id=%s"
- "discord account found: discord_id=%s discord_username=%s"
- "contractor found in notion: discord_username=%s contractor_page_id=%s"

### INFO Level
- "updated approved/rejected by: leave_page_id=%s approver_contractor_id=%s"

### WARNING Level
- "discord account not found for approver: discord_id=%s, skipping update"
- "discord username is empty for approver: discord_id=%s"
- "contractor not found in notion for discord username: %s"
- "multiple contractors found with discord username: %s, using first result"

### ERROR Level
- "failed to query discord account: discord_id=%s error=%v"
- "failed to query notion contractors db: discord_username=%s error=%v"
- "failed to update approved/rejected by in notion: leave_page_id=%s error=%v"

## Configuration Dependencies

### Required Config Values
- `cfg.LeaveIntegration.Notion.ContractorDBID` - Contractors database ID

### Config Validation
- Function should check if contractor DB ID is configured
- Return error if not configured: "contractor database not configured"

## Performance Considerations

### Expected Operations
1. Database query: < 50ms (indexed lookup by discord_id)
2. Notion Contractors query: < 500ms
3. Notion UpdatePage: < 500ms
4. **Total:** ~1s per approval/rejection

### Optimization
- Sequential execution acceptable (low frequency operation)
- No batching needed (single approval at a time)

## Integration Notes

### Caller Context
- Caller: `handleNotionLeaveApproved`, `handleNotionLeaveRejected`
- Called AFTER approval/rejection status update in Notion
- Caller will log if this function returns error but won't fail the approval

### Discord Button Interaction Context
- Approver clicks "Approve" or "Reject" button in Discord
- Discord interaction includes `approverDiscordID` (user who clicked)
- This function links that Discord user back to Notion contractor

## Relation Field Details

### Notion Property
- **Property Name:** "Approved/Rejected By"
- **Property Type:** Relation
- **Target Database:** Contractors
- **Cardinality:** Single relation (one approver)

### Update Behavior
- Overwrites any existing value
- Setting empty relation `[]` would clear the field
- Notion validates that relation target exists in Contractors DB

## Test Implementation Checklist

- [ ] Test successful full flow (happy path)
- [ ] Test Discord account not found (graceful skip)
- [ ] Test Discord username empty (graceful skip)
- [ ] Test contractor not found in Notion (graceful skip)
- [ ] Test Notion query error (graceful skip)
- [ ] Test Notion update error (graceful skip)
- [ ] Test empty leave page ID (input validation)
- [ ] Test empty approver Discord ID (input validation)
- [ ] Test database query error (graceful skip)
- [ ] Test context cancellation
- [ ] Test multiple contractors found (use first)
- [ ] Test nil logger handling
- [ ] Verify Notion query filter correctness
- [ ] Verify relation update format correctness
- [ ] Verify logging at appropriate levels
- [ ] Test helper function for contractor lookup
- [ ] Verify graceful degradation doesn't fail approval
