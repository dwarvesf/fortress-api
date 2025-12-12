# ADR-001: Account Manager and Delivery Lead Lookup Strategy

**Status:** Proposed

**Date:** 2025-12-12

**Context:**

The current Notion leave request webhook implementation relies on a manually maintained "Assignees" multi-select field to determine who should be notified via Discord when a leave request is created. This approach has several limitations:

1. **Manual Maintenance:** Team members must manually add assignees to each leave request, which is error-prone and inconsistent
2. **No Source of Truth:** The assignees are not derived from the actual project deployments, leading to potential mismatches
3. **Stale Data:** As project assignments change, the assignees may become outdated
4. **Scalability:** As the team grows, manually managing assignees becomes increasingly difficult

The Deployment Tracker database in Notion already contains the authoritative mapping of contractors to their Account Managers (AM) and Delivery Leads (DL). This data includes:
- Active deployment status
- Override AM/DL for special cases
- Rollup AM/DL from project assignments

**Decision:**

We will replace the manual "Assignees" multi-select approach with a dynamic lookup strategy that fetches AM/DL from the Deployment Tracker based on the employee's active deployments.

## Lookup Strategy

### 1. Contractor Identification

When a leave request is created:
- Extract `Team Email` from the leave request
- Query the Contractors database to find the contractor page by email
- Use this contractor page ID for subsequent queries

**Rationale:** Email is the unique identifier that links leave requests to contractor records. This is already a proven pattern in the codebase (see `lookupContractorByEmail` in refund webhook).

### 2. Active Deployment Query

Query the Deployment Tracker database with compound filter:
```
Filter: AND
  - Contractor (relation) = contractor_page_id
  - Deployment Status (status) = "Active"
```

**Rationale:**
- Only active deployments are relevant for current AM/DL assignments
- Compound AND filter ensures we only get deployments for the specific contractor
- This matches the existing query pattern for status filters (see research on `StatusDatabaseQueryFilter`)

### 3. AM/DL Extraction with Override Logic

For each active deployment, extract stakeholders using priority logic:

**Priority 1: Override Fields (Direct Relations)**
- Check "Override AM" relation field
- Check "Override DL" relation field
- If set, use these values

**Priority 2: Rollup Fields (From Project)**
- If no override, use "Account Managers" rollup field
- If no override, use "Delivery Leads" rollup field

**Rationale:**
- Override fields allow project-specific AM/DL assignments that differ from project defaults
- Rollups provide the default AM/DL from the project
- This two-tier approach balances flexibility with automation
- Formula fields "Final AM" and "Final Delivery Lead" already implement this logic in Notion, but we need to handle it in the application layer for reliability

### 4. Discord Username to Discord ID Lookup

For each AM/DL contractor page:
- Fetch the contractor page from Notion
- Extract "Discord" property (rich_text field)
- Query fortress database for Discord account by username
- Retrieve Discord ID for mention formatting

**Rationale:**
- Discord usernames are stored in Notion Contractors database
- Discord IDs are stored in fortress database
- This requires a new store method: `DiscordAccount.OneByUsername()`
- Follows the existing pattern of email-to-Discord lookup but uses username instead

### 5. Deduplication

- Collect all AM/DL across all active deployments
- Deduplicate by contractor page ID
- Deduplicate by Discord ID (in case multiple contractor pages have same Discord account)

**Rationale:**
- An employee may have multiple deployments with the same AM/DL
- Avoid spamming the same person multiple times
- Reduces Discord notification noise

## Approver Relation Auto-fill

When a leave request is approved or rejected via Discord button:

1. Extract Discord username from the approver's Discord account
2. Query Notion Contractors database for contractor page with matching Discord username
3. Update the leave request page's "Approved/Rejected By" relation field with the contractor page ID

**Rationale:**
- Automates record-keeping of who approved/rejected the request
- Creates audit trail in Notion
- Links approval action to contractor identity
- Follows the existing pattern for relation updates (see `updateRefundContractor`)

**Consequences:**

### Positive

1. **Automation:** No manual assignee management required
2. **Accuracy:** AM/DL derived from authoritative deployment data
3. **Maintainability:** Single source of truth in Deployment Tracker
4. **Scalability:** Works automatically as deployments are added/updated
5. **Flexibility:** Override mechanism allows special cases
6. **Audit Trail:** Approved/Rejected By relation creates clear accountability

### Negative

1. **Data Quality Dependency:** Relies on accurate Deployment Tracker data
   - Mitigation: Deployment Tracker is already maintained for other purposes
   - Fallback: If no deployments found, could fall back to default approvers (future enhancement)

2. **Additional API Calls:** Requires multiple Notion API queries
   - Mitigation: Queries are only made on leave request creation (low frequency)
   - Mitigation: Notion API rate limits are generous for this use case

3. **New Database Method Required:** Need to add `OneByUsername` to DiscordAccount store
   - Mitigation: Simple query method, follows existing patterns
   - Mitigation: Can add database index on discord_username for performance

4. **Failure Scenarios:** If Deployment Tracker or Contractors data is incomplete
   - Mitigation: Graceful degradation with logging
   - Mitigation: Return empty mentions rather than failing the webhook
   - Mitigation: Log warnings for manual follow-up

5. **Discord Username Consistency:** Depends on Discord usernames being accurate in Notion
   - Mitigation: Discord usernames are already used for other integrations
   - Mitigation: Log mismatches for manual correction

### Migration Path

**Phase 1: Implementation (This ADR)**
- Implement new lookup strategy
- Keep "Assignees" field for backward compatibility (read-only)
- Test with real deployments

**Phase 2: Validation**
- Run in parallel with manual assignees for validation period
- Compare automated AM/DL with manual assignees
- Log discrepancies for investigation

**Phase 3: Cleanup**
- Remove "Assignees" field references from code
- Archive or remove "Assignees" field from Notion (optional)

**Alternatives Considered:**

### Alternative 1: Continue with Manual Assignees
**Rejected:** Does not solve the core problem of manual maintenance and staleness.

### Alternative 2: Use Formula Fields (Final AM/Final DL) Directly
**Rejected:** Formula fields in Notion can return complex types that are harder to parse. Direct extraction of Override/Rollup gives us more control and better error handling.

### Alternative 3: Store AM/DL in Fortress Database
**Rejected:** Would create data duplication. Notion Deployment Tracker is already the source of truth.

### Alternative 4: Email-based Lookup for AM/DL
**Rejected:** AM/DL emails may not be in fortress database. Discord usernames are more reliable for the Discord notification use case.

**References:**

- Requirements: `/docs/sessions/202512120930-notion-leave-webhook-amdl/requirements/requirements.md`
- Research: `/docs/sessions/202512120930-notion-leave-webhook-amdl/research/notion-patterns.md`
- Existing Spec: `/docs/specs/notion-leave-request-webhook.md`
- Existing Implementation: `/pkg/handler/webhook/notion_leave.go`
- Refund Webhook Pattern: `/pkg/handler/webhook/notion_refund.go`
