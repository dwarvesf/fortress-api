# Planning Status: Notion Leave Webhook AM/DL Integration

**Session:** 202512120930-notion-leave-webhook-amdl
**Date:** 2025-12-12
**Status:** Planning Complete

## Summary

Planning documentation has been completed for the Notion Leave Webhook AM/DL integration feature. This feature will replace the manual "Assignees" multi-select field with automatic lookup of Account Managers (AM) and Delivery Leads (DL) from the Deployment Tracker database.

## Documents Created

### 1. Architecture Decision Record (ADR)

**File:** `/docs/sessions/202512120930-notion-leave-webhook-amdl/planning/ADRs/001-am-dl-lookup-strategy.md`

**Key Decisions:**
- Dynamic lookup strategy from Deployment Tracker replaces manual assignees
- Override AM/DL takes precedence over rollup values
- Discord username-based lookup for mention conversion
- Graceful degradation on failures (don't block webhook)
- Auto-fill "Approved/Rejected By" relation on approval/rejection

**Consequences:**
- Positive: Automation, accuracy, maintainability, scalability
- Negative: Dependency on Deployment Tracker data quality, additional API calls
- Mitigation: Robust error handling, logging, graceful fallback

### 2. Technical Specification

**File:** `/docs/sessions/202512120930-notion-leave-webhook-amdl/planning/specifications/am-dl-integration-spec.md`

**Contents:**
- Complete data flow diagram (7 steps)
- Notion API call specifications with filters and responses
- Property extraction patterns (relation, rollup, rich_text)
- Service layer function signatures and logic
- Handler layer modifications
- Database schema changes (new store method)
- Configuration changes (environment variables)
- Error handling strategy
- Testing considerations
- Performance and security analysis

## Key Technical Decisions

### 1. Deployment Query Strategy

**Decision:** Use compound AND filter
```
Filter: AND
  - Contractor (relation) = contractor_page_id
  - Deployment Status (status) = "Active"
```

**Rationale:** Only active deployments are relevant for current assignments

### 2. AM/DL Extraction Logic

**Decision:** Two-tier priority system
- Priority 1: Override AM/DL (direct relation fields)
- Priority 2: Account Managers/Delivery Leads (rollup from project)

**Rationale:** Allows project-specific overrides while maintaining default automation

### 3. Discord Lookup Approach

**Decision:** New store method `DiscordAccount.OneByUsername()`

**Implementation:**
```go
func (s *store) OneByUsername(db *gorm.DB, username string) (*model.DiscordAccount, error)
```

**Database Index:**
```sql
CREATE INDEX idx_discord_accounts_username ON discord_accounts(discord_username);
```

**Rationale:** Discord usernames are stored in Notion, need to convert to Discord IDs for mentions

### 4. Approver Relation Update

**Decision:** Lookup contractor by Discord username, set relation on approval

**Flow:**
1. Discord approval action → Discord ID of approver
2. Query DiscordAccount by discord_id → discord_username
3. Query Notion Contractors by Discord username → contractor_page_id
4. Update leave request "Approved/Rejected By" relation

**Rationale:** Creates audit trail and links approval to contractor identity

## Implementation Components

### Service Layer (pkg/service/notion/leave.go)

**New Functions:**
1. `GetActiveDeploymentsForContractor(ctx, contractorPageID)` → Query Deployment Tracker
2. `ExtractStakeholdersFromDeployment(deploymentPage)` → Extract AM/DL with override logic
3. `GetDiscordUsernameFromContractor(ctx, contractorPageID)` → Fetch Discord username

### Handler Layer (pkg/handler/webhook/notion_leave.go)

**Modified Functions:**
1. `handleNotionLeaveCreated()` → Add AM/DL lookup flow

**New Functions:**
1. `getDiscordMentionFromUsername(username)` → Convert username to mention
2. `updateApprovedRejectedBy(leavePageID, approverDiscordID)` → Update relation on approval

### Store Layer (pkg/store/discord_account/)

**New Method:**
1. `OneByUsername(db, username)` → Query Discord account by username

### Configuration

**New Environment Variable:**
```bash
NOTION_DEPLOYMENT_TRACKER_DB_ID=2b864b29b84c80799568dc17685f4f33
```

## Error Handling Strategy

### Graceful Degradation

**Philosophy:** Non-critical failures should not block the webhook

**Critical Errors (500):**
- Cannot parse webhook payload
- Cannot verify signature
- Database connection failure

**Warnings (Log, Continue):**
- Contractor not found
- No active deployments
- Discord username not in DB
- Optional API calls fail

**Result:** Leave request notification always sent, even if AM/DL lookup fails

### Logging Levels

- **Debug:** Each API call, property extraction, conversion step
- **Info:** Webhook received, notification sent, approval processed
- **Warning:** Expected data missing, mismatches, optional failures
- **Error:** API failures, unexpected types, database errors

## Testing Strategy

### Unit Tests

**Focus Areas:**
- Property extraction (relation, rollup, rich_text)
- Store method (OneByUsername)
- Stakeholder extraction logic (override vs rollup)

### Integration Tests

**Scenarios:**
1. Full flow with active deployments
2. No active deployments (graceful handling)
3. Contractor not found (graceful handling)
4. Approval flow with relation update

**Test Data:**
- Notion: Contractors, deployments with override/rollup, Discord usernames
- Fortress: Employees, Discord accounts

## Migration Plan

### Phase 1: Implementation
- Add new store method
- Add service layer functions
- Update handler with new flow
- Keep "Assignees" code for comparison

### Phase 2: Testing
- Test with real Notion data
- Compare automated AM/DL with manual assignees
- Log discrepancies for validation

### Phase 3: Deployment
- Deploy to production
- Monitor logs for errors
- Validate Discord notifications

### Phase 4: Cleanup
- Remove "Assignees" code after validation
- Update documentation

## Performance Considerations

### API Call Volume (Per Leave Request)
- 1 call: Lookup contractor by email
- 1 call: Query active deployments
- N calls: Fetch contractor pages for AM/DL (N = unique stakeholders, typically 2-5)

### Database Query Volume (Per Leave Request)
- 1 query: Employee by email (existing)
- M queries: Discord account by username (M = unique usernames, typically 2-5)

### Expected Load
- Leave requests: 10-50 per month
- Total API calls: 50-250 per month
- Total DB queries: 50-250 per month

**Conclusion:** Performance impact is minimal. Current approach acceptable.

## Dependencies

### Notion Databases
- Leave Requests (existing)
- Contractors (existing)
- Deployment Tracker (existing)

### Fortress Database
- employees table (existing)
- discord_accounts table (existing, needs index)

### External APIs
- Notion API (existing integration)
- Discord API (existing integration)

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Deployment Tracker data incomplete | Medium | Graceful degradation, log warnings for manual follow-up |
| Discord usernames mismatch | Low | Log mismatches, manual correction, validation period |
| Additional API call latency | Low | Acceptable for current load, can optimize later |
| Notion API rate limits | Low | Current volume well within limits, implement retry if needed |
| Rollup property structure changes | Medium | Defensive parsing, log unexpected structures |

## Next Steps

### For Test-Case Designer
1. Review technical specification for test scenarios
2. Design unit tests for property extraction
3. Design unit tests for store method
4. Design integration tests for full flow
5. Define test data requirements for Notion and Fortress

### For Implementation
1. Implement `DiscordAccount.OneByUsername()` store method
2. Add database index for discord_username
3. Implement service layer functions
4. Update handler with new AM/DL lookup flow
5. Add comprehensive logging
6. Write unit tests
7. Write integration tests

### For QA
1. Validate against test scenarios
2. Compare automated AM/DL with manual assignees
3. Test error scenarios (missing data, API failures)
4. Verify Discord notifications
5. Test approval/rejection flow

## References

- Requirements: `/docs/sessions/202512120930-notion-leave-webhook-amdl/requirements/requirements.md`
- Research: `/docs/sessions/202512120930-notion-leave-webhook-amdl/research/notion-patterns.md`
- Existing Spec: `/docs/specs/notion-leave-request-webhook.md`
- ADR: `/docs/sessions/202512120930-notion-leave-webhook-amdl/planning/ADRs/001-am-dl-lookup-strategy.md`
- Technical Spec: `/docs/sessions/202512120930-notion-leave-webhook-amdl/planning/specifications/am-dl-integration-spec.md`

## Notes

- NO CODE IMPLEMENTATION in this planning phase
- All decisions documented with rationale
- Error handling strategy emphasizes graceful degradation
- Performance analysis shows minimal impact
- Migration plan includes validation phase before cleanup
- Comprehensive testing strategy defined
- Ready for handoff to test-case designer and implementation team
