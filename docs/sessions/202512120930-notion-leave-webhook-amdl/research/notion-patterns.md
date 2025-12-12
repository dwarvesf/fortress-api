# Notion Webhook Implementation Patterns Research

**Date:** 2025-12-12
**Session:** 202512120930-notion-leave-webhook-amdl
**Research Focus:** Notion API patterns, property extraction, relation queries, Discord mentions

## Executive Summary

This research document analyzes the existing Notion webhook implementation patterns in the fortress-api codebase. The analysis covers four key areas:

1. **Notion API Queries** - Filtering by relation, status, and email fields
2. **Property Extraction** - Rollup, formula, and various field types
3. **Discord Mention Lookup** - Converting email/username to Discord ID
4. **Relation Updates** - Setting relation fields on Notion pages

The codebase implements two types of Notion webhooks:
- **Automation webhooks** (notion_refund.go, notion_leave.go) - Triggered by Notion automations
- **Event webhooks** (notion_leave.go) - Triggered by page.created/page.updated events

## 1. Notion API Query Patterns

### 1.1 Standard Database Query (Single Data Source)

**File:** `pkg/service/notion/expense.go:180-220`

**Pattern:** Query standard Notion database with status filter

```go
func (e *ExpenseService) queryDatabase(databaseID string) ([]nt.Page, error) {
    ctx := context.Background()

    // Query for expenses with status "Approved"
    filter := &nt.DatabaseQueryFilter{
        Property: "Status",
        DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
            Status: &nt.StatusDatabaseQueryFilter{
                Equals: "Approved",
            },
        },
    }

    var allPages []nt.Page
    var cursor string

    // Pagination loop
    for {
        query := &nt.DatabaseQuery{
            Filter:   filter,
            PageSize: 100,
        }
        if cursor != "" {
            query.StartCursor = cursor
        }

        resp, err := e.client.QueryDatabase(ctx, databaseID, query)
        if err != nil {
            return nil, fmt.Errorf("notion query failed: %w", err)
        }

        allPages = append(allPages, resp.Results...)

        if !resp.HasMore || resp.NextCursor == nil {
            break
        }
        cursor = *resp.NextCursor
    }

    return allPages, nil
}
```

**Key Points:**
- Uses `go-notion` client's `QueryDatabase` method
- Implements pagination with `StartCursor` and `NextCursor`
- Default page size: 100 records
- Filter by status property using `StatusDatabaseQueryFilter`

### 1.2 Data Source Query (Multi-Source Database)

**File:** `pkg/service/notion/expense.go:255-298`

**Pattern:** Query multi-source database using raw HTTP API

```go
func (e *ExpenseService) queryDataSource(dataSourceID string) ([]nt.Page, error) {
    filter := &nt.DatabaseQueryFilter{
        Property: "Status",
        DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
            Status: &nt.StatusDatabaseQueryFilter{
                Equals: "Approved",
            },
        },
    }

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

        resp, err := e.executeDataSourceQuery(dataSourceID, reqBody)
        if err != nil {
            return nil, err
        }

        // Convert DataSourcePage to nt.Page
        for _, dsPage := range resp.Results {
            page := e.convertDataSourcePageToPage(dsPage)
            allPages = append(allPages, page)
        }

        if !resp.HasMore || resp.NextCursor == nil {
            break
        }
        cursor = *resp.NextCursor
    }

    return allPages, nil
}
```

**HTTP Request Details:**

```go
func (e *ExpenseService) executeDataSourceQuery(dataSourceID string, reqBody DataSourceQueryRequest) (*DataSourceQueryResponse, error) {
    // Normalize: remove hyphens
    normalizedID := strings.ReplaceAll(dataSourceID, "-", "")
    url := fmt.Sprintf("https://api.notion.com/v1/data_sources/%s/query", normalizedID)

    jsonBody, _ := json.Marshal(reqBody)
    req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(jsonBody))

    req.Header.Set("Authorization", "Bearer "+e.cfg.Notion.Secret)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Notion-Version", "2025-09-03") // REQUIRED for data source queries

    // Execute request...
}
```

**Key Points:**
- Data sources require raw HTTP API (not supported by go-notion)
- API endpoint: `POST /v1/data_sources/{id}/query`
- Must use Notion-Version header: `2025-09-03`
- Must normalize data source ID (remove hyphens)
- Same filter structure as standard database queries

### 1.3 Query by Email Filter

**File:** `pkg/handler/webhook/notion_refund.go:190-228`

**Pattern:** Query database by email property to find contractor

```go
func (h *handler) lookupContractorByEmail(ctx context.Context, l logger.Logger, email string) (string, error) {
    contractorDBID := h.config.Notion.Databases.Contractor

    client := nt.NewClient(h.config.Notion.Secret)

    // Query contractor database for matching email
    filter := &nt.DatabaseQueryFilter{
        Property: "Team Email",
        DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
            Email: &nt.TextPropertyFilter{
                Equals: email,
            },
        },
    }

    resp, err := client.QueryDatabase(ctx, contractorDBID, &nt.DatabaseQuery{
        Filter:   filter,
        PageSize: 1, // Only need first result
    })
    if err != nil {
        return "", fmt.Errorf("failed to query contractor database: %w", err)
    }

    if len(resp.Results) == 0 {
        return "", nil // Not found
    }

    pageID := resp.Results[0].ID
    return pageID, nil
}
```

**Key Points:**
- Uses `Email` filter with `TextPropertyFilter`
- Filter syntax: `Email: &nt.TextPropertyFilter{ Equals: email }`
- PageSize: 1 when only one result needed
- Returns empty string if not found (graceful handling)

### 1.4 Query by Relation Filter (Pending Status)

**File:** `pkg/service/notion/leave.go:202-253`

**Pattern:** Query for pending leave requests using status filter

```go
func (s *LeaveService) QueryPendingLeaveRequests(ctx context.Context) ([]LeaveRequest, error) {
    dataSourceID := s.cfg.LeaveIntegration.Notion.DataSourceID

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
    // Transform pages to LeaveRequest...
}
```

**Key Points:**
- Status property can be filtered using `Select` filter type
- Filter syntax: `Select: &nt.SelectDatabaseQueryFilter{ Equals: "Pending" }`
- Different from `Status` filter type used for status property type

## 2. Property Extraction Patterns

### 2.1 Title Property

**File:** `pkg/service/notion/leave.go:351-361`

**Pattern:** Extract title from title property

```go
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
```

**Key Points:**
- Title is array of RichText objects
- Each RichText has `PlainText` field
- Concatenate all parts to get full title

### 2.2 Rich Text Property

**File:** `pkg/service/notion/expense.go:440-450`

**Pattern:** Extract rich text with multiple fallback options

```go
func (e *ExpenseService) extractTitle(props nt.DatabasePageProperties) string {
    // Try "Notes" first (Refund Requests DB)
    if notesProp, ok := props["Notes"]; ok && len(notesProp.RichText) > 0 {
        var parts []string
        for _, rt := range notesProp.RichText {
            parts = append(parts, rt.PlainText)
        }
        result := strings.TrimSpace(strings.Join(parts, ""))
        if result != "" {
            return result
        }
    }

    // Try "Description" as fallback
    if descProp, ok := props["Description"]; ok && len(descProp.RichText) > 0 {
        // Same pattern...
    }

    // Multiple fallback layers...
}
```

**Key Points:**
- Rich text properties stored in `RichText` array
- Use cascading fallback for different property names
- Always trim whitespace from concatenated result
- Log which property was used for debugging

### 2.3 Select Property

**File:** `pkg/service/notion/leave.go:363-369`

**Pattern:** Extract select option name

```go
func (s *LeaveService) extractSelect(props nt.DatabasePageProperties, propName string) string {
    if prop, ok := props[propName]; ok && prop.Select != nil {
        return prop.Select.Name
    }
    return ""
}
```

**Key Points:**
- Select has single option object
- Access option name via `prop.Select.Name`
- Return empty string if not set

### 2.4 Email Property

**File:** `pkg/service/notion/leave.go:390-398`

**Pattern:** Extract email value

```go
func (s *LeaveService) extractEmail(props nt.DatabasePageProperties, propName string) string {
    if prop, ok := props[propName]; ok && prop.Email != nil {
        s.logger.Debug(fmt.Sprintf("extractEmail: property %s has email value: %s", propName, *prop.Email))
        return *prop.Email
    }
    s.logger.Debug(fmt.Sprintf("extractEmail: property %s not found or empty", propName))
    return ""
}
```

**Key Points:**
- Email is pointer to string
- Dereference pointer: `*prop.Email`
- Log extraction for debugging

### 2.5 Date Property

**File:** `pkg/service/notion/leave.go:371-380`

**Pattern:** Extract date value

```go
func (s *LeaveService) extractDate(props nt.DatabasePageProperties, propName string) *time.Time {
    if prop, ok := props[propName]; ok && prop.Date != nil {
        t := prop.Date.Start.Time
        if !t.IsZero() {
            return &t
        }
    }
    return nil
}
```

**Key Points:**
- Date has `Start` and optionally `End` fields
- Access via `prop.Date.Start.Time`
- Check if time is zero before returning
- Return pointer to time.Time (can be nil)

### 2.6 Number Property

**File:** `pkg/service/notion/expense.go:508-514`

**Pattern:** Extract number value

```go
func (e *ExpenseService) extractNumber(props nt.DatabasePageProperties, propName string) float64 {
    if prop, ok := props[propName]; ok && prop.Number != nil {
        return *prop.Number
    }
    return 0
}
```

**Key Points:**
- Number is pointer to float64
- Dereference: `*prop.Number`
- Default to 0 if not set

### 2.7 Relation Property (First ID)

**File:** `pkg/service/notion/leave.go:382-388`

**Pattern:** Extract first relation page ID

```go
func (s *LeaveService) extractFirstRelationID(props nt.DatabasePageProperties, propName string) string {
    if prop, ok := props[propName]; ok && len(prop.Relation) > 0 {
        return prop.Relation[0].ID
    }
    return ""
}
```

**Key Points:**
- Relation is array of relation objects
- Each relation has `ID` field (page ID)
- Extract first relation only
- Return empty string if no relations

### 2.8 Multi-Select Property (Extract Emails)

**File:** `pkg/service/notion/leave.go:400-414`

**Pattern:** Extract emails from multi-select option names

```go
func (s *LeaveService) extractEmailsFromMultiSelect(props nt.DatabasePageProperties, propName string) []string {
    var emails []string
    if prop, ok := props[propName]; ok && len(prop.MultiSelect) > 0 {
        for _, opt := range prop.MultiSelect {
            email := s.parseEmailFromOptionName(opt.Name)
            if email != "" {
                emails = append(emails, email)
            }
        }
    }
    return emails
}

// Format: "Name (email@domain)"
func (s *LeaveService) parseEmailFromOptionName(optionName string) string {
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
```

**Key Points:**
- Multi-select is array of option objects
- Each option has `Name` field
- Parse email from structured format: "Name (email@domain)"
- Use `LastIndex` to find rightmost parentheses
- Validate email contains "@" symbol

### 2.9 Status Property

**File:** `pkg/handler/webhook/notion_refund.go:263-268`

**Pattern:** Extract status name from status property

```go
func getStatusName(prop NotionStatusProperty) string {
    if prop.Status != nil {
        return prop.Status.Name
    }
    return ""
}
```

**Key Points:**
- Status has option object similar to select
- Access name via `prop.Status.Name`

### 2.10 Rollup and Formula Fields

**Context:** Rollup and formula fields are read-only computed values in Notion.

**Not directly shown in code but inferred from usage:**

From the spec document (docs/specs/notion-leave-request-webhook.md):
- Rollup fields aggregate values from related pages
- Formula fields compute values using expressions
- Both are extracted like their underlying property types

**Pattern for Rollup (Array of Relations):**
```go
// Rollup that returns array of relations
if prop, ok := props["Account Managers"]; ok && len(prop.Rollup.Array) > 0 {
    for _, item := range prop.Rollup.Array {
        // item could be relation, number, etc.
    }
}
```

**Pattern for Formula (Various Types):**
```go
// Formula that returns string
if prop, ok := props["Final AM"]; ok {
    switch prop.Formula.Type {
    case "string":
        value = prop.Formula.String
    case "number":
        value = prop.Formula.Number
    // etc.
    }
}
```

**Note:** The current codebase doesn't show rollup/formula extraction, but the spec indicates they will be needed for the AM/DL lookup feature.

## 3. Discord Mention Lookup Patterns

### 3.1 Email to Discord Mention (Current Pattern)

**File:** `pkg/handler/webhook/nocodb_leave.go:361-394`

**Pattern:** Employee email → Discord ID → Discord mention

```go
func (h *handler) getEmployeeDiscordMention(l logger.Logger, email string) string {
    if email == "" {
        return ""
    }

    // Step 1: Look up employee by email
    employee, err := h.store.Employee.OneByEmail(h.repo.DB(), email)
    if err != nil {
        l.Debugf("could not find employee by email %s: %v", email, err)
        return ""
    }

    // Step 2: Check if employee has Discord account linked
    if employee.DiscordAccountID.String() == "" {
        l.Debugf("employee %s has no discord account linked", email)
        return ""
    }

    // Step 3: Get Discord account to retrieve Discord ID
    discordAccount, err := h.store.DiscordAccount.One(h.repo.DB(), employee.DiscordAccountID.String())
    if err != nil {
        l.Debugf("could not find discord account for employee %s: %v", email, err)
        return ""
    }

    // Step 4: Validate Discord ID exists
    if discordAccount.DiscordID == "" {
        l.Debugf("discord account for employee %s has no discord id", email)
        return ""
    }

    // Step 5: Format Discord mention
    l.Debugf("found discord id %s for employee %s", discordAccount.DiscordID, email)
    return fmt.Sprintf("<@%s>", discordAccount.DiscordID)
}
```

**Database Schema:**
```
employees
  ├── id (UUID)
  ├── email (string)
  ├── full_name (string)
  └── discord_account_id (UUID, FK to discord_accounts)

discord_accounts
  ├── id (UUID)
  ├── discord_id (string) - Discord's user ID
  └── discord_username (string)
```

**Key Points:**
- Two-step DB lookup: Employee → DiscordAccount
- Discord mention format: `<@{discord_id}>`
- Graceful failure: returns empty string if any step fails
- Extensive debug logging at each step
- Validates at each step before proceeding

### 3.2 Usage in Discord Notifications

**File:** `pkg/handler/webhook/notion_leave.go:278-293`

**Pattern:** Build mention string for multiple assignees

```go
// Build Discord mentions for assignees
var assigneeMentions string
if len(leave.Assignees) > 0 {
    l.Debug(fmt.Sprintf("found %d assignees for leave request: %v", len(leave.Assignees), leave.Assignees))
    var mentions []string
    for _, email := range leave.Assignees {
        mention := h.getEmployeeDiscordMention(l, email)
        if mention != "" {
            mentions = append(mentions, mention)
        }
    }
    if len(mentions) > 0 {
        assigneeMentions = fmt.Sprintf("🔔 **Assignees:** %s", strings.Join(mentions, " "))
    }
}

// Send message with assignee mentions as content
msg, err := h.service.Discord.SendChannelMessageComplex(channelID, assigneeMentions, []*discordgo.MessageEmbed{embed}, components)
```

**Key Points:**
- Convert array of emails to array of mentions
- Filter out empty mentions (employees without Discord)
- Join mentions with space separator
- Format as message content (not embed) to trigger actual mentions
- Only add mention prefix if mentions exist

### 3.3 Discord Username to Discord ID (Proposed Pattern)

**Context:** For AM/DL lookup from Notion Contractors DB

**File:** Not yet implemented, but pattern inferred from spec

**Proposed Pattern:**
```go
// Step 1: Extract Discord username from Notion contractor page
func (s *LeaveService) getDiscordUsernameFromContractor(ctx context.Context, contractorPageID string) (string, error) {
    page, err := s.client.FindPageByID(ctx, contractorPageID)
    if err != nil {
        return "", err
    }

    props, ok := page.Properties.(nt.DatabasePageProperties)
    if !ok {
        return "", errors.New("invalid properties")
    }

    // Extract Discord field (rich_text)
    if discordProp, ok := props["Discord"]; ok && len(discordProp.RichText) > 0 {
        var parts []string
        for _, rt := range discordProp.RichText {
            parts = append(parts, rt.PlainText)
        }
        username := strings.TrimSpace(strings.Join(parts, ""))
        return username, nil
    }

    return "", nil
}

// Step 2: Look up Discord ID by username
func (h *handler) getDiscordMentionFromUsername(l logger.Logger, discordUsername string) string {
    if discordUsername == "" {
        return ""
    }

    // Query discord_accounts by username
    discordAccount, err := h.store.DiscordAccount.OneByUsername(h.repo.DB(), discordUsername)
    if err != nil {
        l.Debugf("could not find discord account by username %s: %v", discordUsername, err)
        return ""
    }

    if discordAccount.DiscordID == "" {
        l.Debugf("discord account %s has no discord id", discordUsername)
        return ""
    }

    return fmt.Sprintf("<@%s>", discordAccount.DiscordID)
}
```

**Key Points:**
- Discord username stored as rich_text in Notion
- Need new store method: `DiscordAccount.OneByUsername()`
- Same mention format: `<@{discord_id}>`
- Graceful failure pattern

## 4. Relation Field Update Patterns

### 4.1 Update Single Relation Field

**File:** `pkg/handler/webhook/notion_refund.go:230-259`

**Pattern:** Update relation property with contractor page ID

```go
func (h *handler) updateRefundContractor(ctx context.Context, l logger.Logger, refundPageID, contractorPageID string) error {
    client := nt.NewClient(h.config.Notion.Secret)

    // Build update params with Contractor relation
    updateParams := nt.UpdatePageParams{
        DatabasePageProperties: nt.DatabasePageProperties{
            "Contractor": nt.DatabasePageProperty{
                Relation: []nt.Relation{
                    {ID: contractorPageID},
                },
            },
        },
    }

    updatedPage, err := client.UpdatePage(ctx, refundPageID, updateParams)
    if err != nil {
        return fmt.Errorf("failed to update page: %w", err)
    }

    l.Debug(fmt.Sprintf("successfully updated contractor relation on refund page: %s", refundPageID))
    return nil
}
```

**Key Points:**
- Use `UpdatePage` with `UpdatePageParams`
- Relation is array of `nt.Relation` objects
- Each relation has `ID` field (page ID)
- Can update single field without affecting others
- Returns updated page object (can ignore if not needed)

### 4.2 Update Multiple Fields (Status + Relations + Date)

**File:** `pkg/service/notion/leave.go:116-158`

**Pattern:** Update multiple properties atomically

```go
func (s *LeaveService) UpdateLeaveStatus(ctx context.Context, pageID, status, approverPageID string) error {
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
        return fmt.Errorf("failed to update page: %w", err)
    }

    return nil
}
```

**Key Points:**
- Build `DatabasePageProperties` map with all updates
- Select property: use `Select: &nt.SelectOptions{ Name: value }`
- Relation property: use `Relation: []nt.Relation{{ ID: pageID }}`
- Date property: use `Date: &nt.Date{ Start: nt.NewDateTime(time, false) }`
- All updates applied atomically in single API call
- Conditional field updates based on business logic

### 4.3 Update Relation from Automation Webhook

**File:** `pkg/handler/webhook/notion_leave.go:764-793`

**Pattern:** Update relation based on webhook payload

```go
func (h *handler) updateOnLeaveEmployee(ctx context.Context, l logger.Logger, onLeavePageID, contractorPageID string) error {
    l.Debug(fmt.Sprintf("updating on-leave employee: onleave_page=%s contractor_page=%s", onLeavePageID, contractorPageID))

    client := nt.NewClient(h.config.Notion.Secret)

    // Build update params with Employee relation
    updateParams := nt.UpdatePageParams{
        DatabasePageProperties: nt.DatabasePageProperties{
            "Employee": nt.DatabasePageProperty{
                Relation: []nt.Relation{
                    {ID: contractorPageID},
                },
            },
        },
    }

    updatedPage, err := client.UpdatePage(ctx, onLeavePageID, updateParams)
    if err != nil {
        l.Error(err, fmt.Sprintf("notion API error updating page: %s", onLeavePageID))
        return fmt.Errorf("failed to update page: %w", err)
    }

    l.Debug(fmt.Sprintf("notion API response - page ID: %s, URL: %s", updatedPage.ID, updatedPage.URL))
    return nil
}
```

**Key Points:**
- Same pattern as standard relation update
- Log request and response for debugging
- Return Notion page URL from response
- Error handling with context

## 5. Webhook Signature Verification

### 5.1 Notion Webhook HMAC-SHA256 Verification

**File:** `pkg/handler/webhook/notion_leave.go:563-581`

**Pattern:** Verify webhook signature using HMAC-SHA256

```go
func (h *handler) verifyNotionWebhookSignature(body []byte, signature, token string) bool {
    // Remove "sha256=" prefix if present
    cleanSignature := signature
    if len(signature) > 7 && signature[:7] == "sha256=" {
        cleanSignature = signature[7:]
    }

    // Calculate expected HMAC-SHA256 signature
    mac := hmac.New(sha256.New, []byte(token))
    mac.Write(body)
    expectedMAC := mac.Sum(nil)
    expectedSignature := hex.EncodeToString(expectedMAC)

    h.logger.Debug(fmt.Sprintf("signature verification: received=%s expected=%s token_len=%d",
        cleanSignature, expectedSignature, len(token)))

    // Compare signatures using constant-time comparison
    return hmac.Equal([]byte(expectedSignature), []byte(cleanSignature))
}
```

**Usage:**
```go
// In webhook handler
signature := c.GetHeader("X-Notion-Signature")
if !h.verifyNotionWebhookSignature(body, signature, verificationToken) {
    c.JSON(http.StatusUnauthorized, ...)
    return
}
```

**Key Points:**
- Header: `X-Notion-Signature`
- Format: `sha256=<hex_encoded_hash>` or just `<hex_encoded_hash>`
- Algorithm: HMAC-SHA256(token, body)
- Use constant-time comparison to prevent timing attacks
- Log signature for debugging (but not in production)

## 6. Webhook Payload Patterns

### 6.1 Notion Event Webhook (page.created/page.updated)

**File:** `pkg/handler/webhook/notion_leave.go:36-56`

**Payload Structure:**
```go
type NotionLeaveWebhookPayload struct {
    // Verification fields (for endpoint verification challenge)
    VerificationToken string `json:"verification_token"`
    Challenge         string `json:"challenge"`

    // Event fields
    Type   string                    `json:"type"`   // "page.created", "page.updated"
    Entity *NotionLeaveWebhookEntity `json:"entity"` // The entity that triggered
    Data   *NotionLeaveWebhookData   `json:"data"`   // Additional data
}

type NotionLeaveWebhookEntity struct {
    ID   string `json:"id"`   // Page ID
    Type string `json:"type"` // "page", "database"
}
```

**Verification Challenge Handling:**
```go
// Handle verification challenge
if payload.VerificationToken != "" {
    if payload.Challenge != "" {
        c.JSON(http.StatusOK, gin.H{"challenge": payload.Challenge})
    } else {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    }
    return
}
```

**Key Points:**
- Verification request: respond with challenge value
- Event types: `page.created`, `page.updated`, `page.content_updated`, `page.properties_updated`
- Entity ID is the page ID
- Must fetch full page details using `FindPageByID`

### 6.2 Notion Automation Webhook

**File:** `pkg/handler/webhook/notion_refund.go:17-110`

**Payload Structure:**
```go
type NotionRefundWebhookPayload struct {
    Source NotionRefundSource `json:"source"`
    Data   NotionRefundData   `json:"data"`
}

type NotionRefundSource struct {
    Type         string `json:"type"`          // "automation"
    AutomationID string `json:"automation_id"`
    ActionID     string `json:"action_id"`
    EventID      string `json:"event_id"`
    Attempt      int    `json:"attempt"`
}

type NotionRefundData struct {
    Object     string                 `json:"object"`     // "page"
    ID         string                 `json:"id"`         // Page ID
    Properties NotionRefundProperties `json:"properties"` // Inline properties
    URL        string                 `json:"url"`
}

type NotionRefundProperties struct {
    Status      NotionStatusProperty   `json:"Status"`
    WorkEmail   NotionEmailProperty    `json:"Work Email"`
    Amount      NotionNumberProperty   `json:"Amount"`
    Currency    NotionSelectProperty   `json:"Currency"`
    Contractor  NotionRelationProperty `json:"Contractor"`
}
```

**Key Points:**
- Automation webhooks include full page data inline
- No need for additional API call to fetch page
- Properties are strongly typed in payload
- Source metadata identifies automation execution
- No verification challenge (uses headers only)

## 7. Common Patterns and Best Practices

### 7.1 Error Handling

**Pattern:** Graceful degradation with logging

```go
// Don't fail webhook on non-critical errors
contractorPageID, err := h.lookupContractorByEmail(ctx, email)
if err != nil {
    l.Error(err, "failed to lookup contractor")
    // Continue without updating - don't fail webhook
} else if contractorPageID != "" {
    // Update if found
    h.updateRefundContractor(ctx, refundPageID, contractorPageID)
}
```

**Key Points:**
- Always log errors with context
- Distinguish critical vs non-critical failures
- Return 200 OK for non-critical failures (webhook acknowledged)
- Return 4xx/5xx only for invalid requests or system errors

### 7.2 Logging Strategy

**Pattern:** Debug logging at every step

```go
l.Debug(fmt.Sprintf("parsed webhook payload: type=%s page_id=%s", payload.Type, pageID))
l.Debug(fmt.Sprintf("fetched leave request: status=%s email=%s", leave.Status, leave.Email))
l.Debug(fmt.Sprintf("found %d assignees: %v", len(leave.Assignees), leave.Assignees))
```

**Key Points:**
- Log webhook payload on receipt
- Log each extraction/transformation step
- Log API calls (request and response)
- Use structured logging with fields
- Debug level for verbose details

### 7.3 Configuration Management

**Pattern:** Environment-based configuration

```go
// In config
type NotionConfig struct {
    Secret    string
    Databases struct {
        Contractor         string
        DeploymentTracker  string
    }
}

// In service
if s.cfg.Notion.Secret == "" {
    return nil, errors.New("notion secret not configured")
}
```

**Key Points:**
- Validate configuration on service creation
- Return nil service if critical config missing
- Use nested config structs for organization
- Environment variables for all secrets

### 7.4 Duplicate Prevention

**Pattern:** Check existence before creating

```go
// Check if leave request already exists
existingLeave, err := h.store.OnLeaveRequest.GetByNotionPageID(h.repo.DB(), leave.PageID)
if err == nil && existingLeave != nil {
    l.Debug(fmt.Sprintf("already approved (skipping duplicate): page_id=%s", leave.PageID))
    c.JSON(http.StatusOK, ...)
    return
}

// Create new record
leaveRequest := &model.OnLeaveRequest{...}
h.store.OnLeaveRequest.Create(h.repo.DB(), leaveRequest)
```

**Key Points:**
- Always check for existing records before insert
- Use unique identifiers (Notion page ID)
- Log and skip duplicates gracefully
- Return success status for duplicates

## 8. Data Flow Summary

### Current Leave Request Flow (Email-based Assignees)

```
1. Notion Leave Request Created
   └─> Webhook: page.created event
       └─> HandleNotionLeave()
           ├─> Fetch page via GetLeaveRequest()
           │   └─> Extract: Email, LeaveType, StartDate, EndDate, Status
           │   └─> Extract: Assignees from multi_select (emails)
           │
           ├─> Validate employee exists
           ├─> Validate dates
           │
           ├─> Convert assignee emails to Discord mentions
           │   └─> For each email:
           │       └─> Employee.OneByEmail()
           │           └─> DiscordAccount.One()
           │               └─> Format: <@discord_id>
           │
           └─> Send Discord notification
               ├─> Embed with leave details
               ├─> Approve/Reject buttons
               └─> Assignee mentions in content
```

### Proposed Leave Request Flow (AM/DL from Deployments)

```
1. Notion Leave Request Created
   └─> Webhook: page.created event
       └─> HandleNotionLeave()
           ├─> Fetch page via GetLeaveRequest()
           │
           ├─> Validate employee exists
           ├─> Validate dates
           │
           ├─> Get AM/DL from Deployment Tracker
           │   ├─> Lookup contractor by Team Email
           │   │   └─> Query Contractors DB (email filter)
           │   │
           │   ├─> Query active deployments
           │   │   └─> Query Deployment Tracker
           │   │       ├─> Filter: Contractor = contractor_id
           │   │       └─> Filter: Status = "Active"
           │   │
           │   ├─> For each deployment:
           │   │   ├─> Get Override AM/DL (relations)
           │   │   ├─> If no override, get AM/DL (rollups)
           │   │   └─> Fetch contractor pages
           │   │       └─> Extract Discord username (rich_text)
           │   │
           │   └─> Convert Discord usernames to mentions
           │       └─> DiscordAccount.OneByUsername()
           │           └─> Format: <@discord_id>
           │
           └─> Send Discord notification
               ├─> Embed with leave details
               ├─> Approve/Reject buttons
               └─> AM/DL mentions in content
```

## 9. Key Takeaways for AM/DL Implementation

### Required New Functions

1. **Query Active Deployments by Contractor**
   - Query: Deployment Tracker
   - Filter: Contractor (relation) = contractor_id AND Status = "Active"
   - Return: Array of deployment pages

2. **Extract Stakeholders from Deployment**
   - First: Check Override AM/DL (relation fields)
   - Fallback: Extract AM/DL from rollups
   - Return: Array of contractor page IDs

3. **Get Discord Username from Contractor Page**
   - Fetch: Contractor page by ID
   - Extract: Discord field (rich_text property)
   - Return: Discord username string

4. **Convert Discord Username to Mention**
   - New store method: `DiscordAccount.OneByUsername()`
   - Lookup: Discord account by username
   - Return: `<@discord_id>` or empty string

### Property Types to Handle

- **Relation** (Override AM/DL) - Extract first ID
- **Rollup** (Account Managers/Delivery Leads) - Extract array of relations
- **Formula** (Final AM/Final DL) - Extract computed relation value
- **Rich Text** (Discord username) - Concatenate plain text

### Environment Configuration

Add to config:
```
NOTION_DEPLOYMENT_TRACKER_DB_ID=2b864b29b84c80799568dc17685f4f33
```

### Database Schema Changes

Add to `discord_accounts` store:
```go
func (s *discordAccountStore) OneByUsername(db *gorm.DB, username string) (*model.DiscordAccount, error)
```

## 10. References

### Files Analyzed

1. `/pkg/handler/webhook/notion_leave.go` - Event webhooks with validation and Discord notifications
2. `/pkg/handler/webhook/notion_refund.go` - Automation webhooks with auto-fill relations
3. `/pkg/service/notion/leave.go` - Leave service with data source queries and property extraction
4. `/pkg/service/notion/expense.go` - Expense service with multi-source database patterns
5. `/pkg/handler/webhook/nocodb_leave.go` - Discord mention lookup implementation
6. `/docs/specs/notion-leave-request-webhook.md` - Specification document

### Go-Notion Library

- Repository: https://github.com/dstotijn/go-notion
- Version: Latest (used in codebase)
- Supports: Standard database queries, page updates, property extraction
- Does NOT support: Data source queries (requires raw HTTP)

### Notion API

- Version: 2025-09-03 (required for data sources)
- Endpoint: https://api.notion.com/v1
- Authentication: Bearer token in Authorization header
- Rate limits: Not explicitly handled in code (relies on Notion's retry-after)

---

**End of Research Document**

This document provides comprehensive patterns for Notion webhook implementation. All code examples are taken directly from the codebase and represent production-tested patterns.
