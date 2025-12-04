# Technical Considerations and Mapping Strategies

## Overview

This document outlines the technical challenges and mapping strategies for implementing the Notion Expense Provider, with specific focus on data type conversions, API patterns, and integration considerations.

## 1. ID Mapping Strategy

### Challenge

**NocoDB**: Integer IDs (e.g., `123`, `456`)
```go
ID: 123  // Simple integer
```

**Notion**: UUID strings (e.g., `"2bfb69f8-f573-81cb-a2da-f06d28896390"`)
```go
ID: "2bfb69f8-f573-81cb-a2da-f06d28896390"  // 36-character UUID
```

**Problem**: The `bcModel.Todo.ID` field expects an `int`, but Notion provides UUID strings.

### Solution Options

#### Option 1: Hash-Based Conversion (Recommended)

Convert Notion UUID to stable integer representation:

```go
func notionPageIDToInt(pageID string) int {
    // Remove hyphens from UUID: "2bfb69f8-f573-81cb-a2da-f06d28896390" → "2bfb69f8f57381cba2daf06d28896390"
    cleanID := strings.ReplaceAll(pageID, "-", "")

    // Take last 8 hex characters (32 bits): "28896390"
    hashStr := cleanID[len(cleanID)-8:]

    // Convert hex to int: 0x28896390 → 680141712
    hash, _ := strconv.ParseInt(hashStr, 16, 64)

    return int(hash)
}
```

**Pros**:
- Deterministic (same UUID always produces same int)
- No database storage needed
- Simple implementation
- Minimal code changes

**Cons**:
- Potential collisions (very low probability with last 8 chars)
- Loses original UUID for reverse lookup
- May need to store original UUID in metadata

#### Option 2: CRC32 Hash

Use CRC32 checksum for more distributed hash values:

```go
func notionPageIDToInt(pageID string) int {
    return int(crc32.ChecksumIEEE([]byte(pageID)))
}
```

**Pros**:
- Built-in Go library
- Good distribution
- Fast computation

**Cons**:
- Higher collision probability than last-8-hex
- Still loses original UUID

#### Option 3: Database Mapping Table

Store UUID→Int mapping in database:

```sql
CREATE TABLE notion_expense_id_mapping (
    id SERIAL PRIMARY KEY,
    notion_page_id UUID UNIQUE NOT NULL,
    local_id INT UNIQUE NOT NULL
);
```

```go
func (s *NotionExpenseService) getOrCreateLocalID(pageID string) (int, error) {
    // Check cache/database for existing mapping
    localID, exists := s.idMapping[pageID]
    if exists {
        return localID, nil
    }

    // Create new mapping with auto-increment ID
    localID = s.store.NotionExpenseIDMapping.Create(pageID)
    s.idMapping[pageID] = localID
    return localID, nil
}
```

**Pros**:
- No collision risk (guaranteed unique)
- Preserves UUID for reverse lookup
- Can regenerate mapping if needed

**Cons**:
- Database migration required
- Additional database queries
- More complex implementation
- Cache management needed

#### Option 4: Store UUID in Metadata Field

Add new field to track original Notion page ID:

```go
type CommissionExplain struct {
    // Existing fields
    BasecampTodoID   int
    BasecampBucketID int

    // New fields
    NotionPageID     string  // Store original UUID
}
```

**Pros**:
- Preserves full UUID for future reference
- No collision concerns
- Enables reverse lookup for updates

**Cons**:
- Database schema changes
- Migration of existing data
- Larger storage footprint

### Recommendation

**Use Option 1 (Hash-Based Conversion) with Option 4 (Metadata Storage)**:

```go
func (s *NotionExpenseService) transformPageToTodo(page notion.Page) (*bcModel.Todo, error) {
    // Convert UUID to int for compatibility
    intID := notionPageIDToInt(page.ID)

    // Store original UUID for future reference (in CommissionExplain when used)
    // This allows reverse lookup for status updates

    return &bcModel.Todo{
        ID: intID,
        // ... other fields
    }, nil
}
```

**Rationale**:
- Minimal code changes (no new tables, simple conversion)
- Deterministic mapping (same UUID = same int)
- Original UUID preserved in metadata for updates
- Low collision probability (32-bit space from 128-bit UUID)

## 2. Property Type Mapping

### Title Property (Title Type)

**Notion Structure**:
```json
{
  "Title": {
    "type": "title",
    "title": [
      {
        "type": "text",
        "text": { "content": "Office supplies expense" },
        "plain_text": "Office supplies expense"
      }
    ]
  }
}
```

**Extraction Code**:
```go
titleProp := props["Title"]
if titleProp.Type == notion.DBPropTypeTitle && len(titleProp.Title) > 0 {
    title := titleProp.Title[0].PlainText
}
```

**Challenge**: Title is an array of rich text objects, need to extract plain text.

**Solution**: Concatenate all rich text segments (usually just one):
```go
func extractTitle(titleProp notion.DatabasePageProperty) string {
    var parts []string
    for _, rt := range titleProp.Title {
        parts = append(parts, rt.PlainText)
    }
    return strings.Join(parts, "")
}
```

### Amount Property (Number Type)

**Notion Structure**:
```json
{
  "Amount": {
    "type": "number",
    "number": 5000000
  }
}
```

**Extraction Code**:
```go
amountProp := props["Amount"]
if amountProp.Type == notion.DBPropTypeNumber {
    amount := amountProp.Number  // float64
}
```

**Challenge**: Notion stores as float64, payroll expects integer amount in formatted string.

**Solution**: Format as integer without decimals:
```go
func extractAmount(amountProp notion.DatabasePageProperty) float64 {
    if amountProp.Type == notion.DBPropTypeNumber {
        return amountProp.Number
    }
    return 0
}

// In title formatting
todoTitle := fmt.Sprintf("%s | %.0f | %s", title, amount, currency)
```

### Currency Property (Select Type)

**Notion Structure**:
```json
{
  "Currency": {
    "type": "select",
    "select": {
      "id": "abc123",
      "name": "VND",
      "color": "blue"
    }
  }
}
```

**Extraction Code**:
```go
currencyProp := props["Currency"]
if currencyProp.Type == notion.DBPropTypeSelect && currencyProp.Select != nil {
    currency := currencyProp.Select.Name
}
```

**Challenge**: Handle null/empty select values.

**Solution**: Default to VND if empty:
```go
func extractCurrency(currencyProp notion.DatabasePageProperty) string {
    if currencyProp.Type == notion.DBPropTypeSelect && currencyProp.Select != nil {
        return currencyProp.Select.Name
    }
    return "VND"  // Default currency
}
```

### Status Property (Status Type)

**Notion Structure**:
```json
{
  "Status": {
    "type": "status",
    "status": {
      "id": "in_progress",
      "name": "Approved",
      "color": "yellow"
    }
  }
}
```

**Extraction Code**:
```go
statusProp := props["Status"]
if statusProp.Type == notion.DBPropTypeStatus && statusProp.Status != nil {
    status := statusProp.Status.Name  // "Approved", "Paid", "Pending"
}
```

**Challenge**: Map Notion status names to internal workflow states.

**Solution**: Use exact status option names from Notion UI:
```go
func isApproved(statusProp notion.DatabasePageProperty) bool {
    if statusProp.Type == notion.DBPropTypeStatus && statusProp.Status != nil {
        return statusProp.Status.Name == "Approved"
    }
    return false
}
```

### Email Property (Rollup Type)

**Notion Structure**:
```json
{
  "Email": {
    "type": "rollup",
    "rollup": {
      "type": "array",
      "array": [
        {
          "type": "email",
          "email": "employee@d.foundation"
        }
      ]
    }
  }
}
```

**Extraction Code**:
```go
emailProp := props["Email"]
if emailProp.Type == notion.DBPropTypeRollup {
    rollup := emailProp.Rollup

    switch rollup.Type {
    case notion.RollupTypeArray:
        if len(rollup.Array) > 0 {
            // First item in array should be email property
            if emailVal, ok := rollup.Array[0].(notion.DatabasePageProperty); ok {
                if emailVal.Type == notion.DBPropTypeEmail {
                    email := emailVal.Email
                }
            }
        }
    }
}
```

**Challenge**: Rollup type varies based on computation (array vs string vs number).

**Solution**: Handle multiple rollup types with type switch:
```go
func extractEmail(emailProp notion.DatabasePageProperty) (string, error) {
    if emailProp.Type != notion.DBPropTypeRollup {
        return "", fmt.Errorf("not a rollup property")
    }

    rollup := emailProp.Rollup

    switch rollup.Type {
    case notion.RollupTypeArray:
        // Email in array format (most common for rollups)
        if len(rollup.Array) > 0 {
            // Type assertion to DatabasePageProperty
            if emailProp, ok := rollup.Array[0].(notion.DatabasePageProperty); ok {
                return emailProp.Email, nil
            }
        }
    case notion.RollupTypeString:
        // Direct string value (if rollup uses "Show original")
        return rollup.String, nil
    }

    return "", fmt.Errorf("unable to extract email from rollup")
}
```

### Attachments Property (Files Type)

**Notion Structure**:
```json
{
  "Attachments": {
    "type": "files",
    "files": [
      {
        "name": "receipt.pdf",
        "type": "file",
        "file": {
          "url": "https://...",
          "expiry_time": "2024-01-01T00:00:00.000Z"
        }
      }
    ]
  }
}
```

**Extraction Code**:
```go
attachmentsProp := props["Attachments"]
if attachmentsProp.Type == notion.DBPropTypeFiles && len(attachmentsProp.Files) > 0 {
    // Get first attachment URL
    firstFile := attachmentsProp.Files[0]
    if firstFile.Type == notion.FileTypeFile {
        attachmentURL := firstFile.File.URL
    }
}
```

**Challenge**: Notion file URLs expire after 1 hour.

**Solution**: Store URL immediately, don't rely on it for long-term access:
```go
func extractAttachmentURL(attachmentsProp notion.DatabasePageProperty) string {
    if attachmentsProp.Type == notion.DBPropTypeFiles && len(attachmentsProp.Files) > 0 {
        firstFile := attachmentsProp.Files[0]
        if firstFile.Type == notion.FileTypeFile {
            return firstFile.File.URL  // Valid for ~1 hour
        } else if firstFile.Type == notion.FileTypeExternal {
            return firstFile.External.URL  // Permanent external URL
        }
    }
    return ""
}
```

## 3. Relation and Rollup Handling

### Understanding Relations

**Notion Database Structure**:
```
Expense Request DB ─┐
                    │ Relation: Requestor
                    ↓
Contractor DB       │
  - Name           │
  - Email  ←───────┘ Rollup: Email from Requestor
```

**Relation Property**:
```json
{
  "Requestor": {
    "type": "relation",
    "relation": [
      {
        "id": "contractor-page-id-uuid"
      }
    ]
  }
}
```

**Rollup Property (Email from Requestor)**:
```json
{
  "Email": {
    "type": "rollup",
    "rollup": {
      "type": "array",
      "array": [
        {
          "type": "email",
          "email": "contractor@d.foundation"
        }
      ],
      "function": "show_original"
    }
  }
}
```

### Rollup Extraction Strategy

#### Strategy 1: Use Rollup Directly (Recommended)

Rollup already aggregates data from the relation, no need for separate query:

```go
func (s *NotionExpenseService) extractEmail(props notion.DatabasePageProperties) (string, error) {
    emailProp, ok := props["Email"]
    if !ok {
        return "", fmt.Errorf("Email property not found")
    }

    if emailProp.Type != notion.DBPropTypeRollup {
        return "", fmt.Errorf("Email is not a rollup property")
    }

    rollup := emailProp.Rollup

    switch rollup.Type {
    case notion.RollupTypeArray:
        if len(rollup.Array) > 0 {
            // Rollup array contains property values
            if propVal, ok := rollup.Array[0].(notion.DatabasePageProperty); ok {
                if propVal.Type == notion.DBPropTypeEmail {
                    return propVal.Email, nil
                }
            }
        }
    case notion.RollupTypeString:
        // Some rollup configurations return direct string
        return rollup.String, nil
    }

    return "", fmt.Errorf("unable to extract email from rollup")
}
```

**Pros**:
- Single API call (no secondary query needed)
- Efficient (Notion pre-computes rollup)
- Follows existing Notion UI configuration

**Cons**:
- Depends on correct rollup configuration in Notion
- Rollup type may vary based on configuration

#### Strategy 2: Query Relation Directly (Fallback)

If rollup fails, query the related contractor page:

```go
func (s *NotionExpenseService) extractEmailFromRelation(ctx context.Context, props notion.DatabasePageProperties) (string, error) {
    requestorProp, ok := props["Requestor"]
    if !ok || requestorProp.Type != notion.DBPropTypeRelation {
        return "", fmt.Errorf("Requestor relation not found")
    }

    if len(requestorProp.Relation) == 0 {
        return "", fmt.Errorf("no contractor linked")
    }

    // Get first related contractor
    contractorPageID := requestorProp.Relation[0].ID

    // Query contractor page
    contractorPage, err := s.client.FindPageByID(ctx, contractorPageID)
    if err != nil {
        return "", fmt.Errorf("failed to fetch contractor page: %w", err)
    }

    // Extract email from contractor page
    contractorProps := contractorPage.Properties.(notion.DatabasePageProperties)
    emailProp, ok := contractorProps["Email"]
    if !ok || emailProp.Type != notion.DBPropTypeEmail {
        return "", fmt.Errorf("email not found in contractor page")
    }

    return emailProp.Email, nil
}
```

**Pros**:
- Independent of rollup configuration
- Direct access to source data
- More control over error handling

**Cons**:
- Requires additional API call per expense
- Slower (N+1 query problem)
- More complex code

### Recommendation

**Use Strategy 1 (Rollup) as primary, Strategy 2 (Direct Query) as fallback**:

```go
func (s *NotionExpenseService) getEmail(ctx context.Context, props notion.DatabasePageProperties) (string, error) {
    // Try rollup first (efficient)
    email, err := s.extractEmail(props)
    if err == nil && email != "" {
        return email, nil
    }

    s.logger.Warn("Rollup extraction failed, falling back to direct relation query")

    // Fallback to direct relation query
    return s.extractEmailFromRelation(ctx, props)
}
```

## 4. Status Property Handling

### Status Options in Notion

**Notion UI Configuration**:
```
Status property type: Status
Options:
  - Pending (to_do group, gray color)
  - Approved (in_progress group, yellow color)
  - Paid (complete group, green color)
```

### Query Filter for Approved Status

```go
filter := &notion.DatabaseQueryFilter{
    Property: "Status",
    DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
        Status: &notion.StatusDatabaseQueryFilter{
            Equals: "Approved",  // Exact status option name
        },
    },
}
```

**Important Notes**:
- Status name must match exactly (case-sensitive)
- Use option **name** ("Approved"), not ID ("in_progress")
- Status filter is different from Select filter (don't use `Select` filter type)

### Update Status from Approved to Paid

```go
func (s *NotionExpenseService) MarkExpenseAsCompleted(pageID string) error {
    ctx := context.Background()

    updateParams := notion.UpdatePageParams{
        Properties: notion.DatabasePageProperties{
            "Status": notion.DatabasePageProperty{
                Type: notion.DBPropTypeStatus,
                Status: &notion.SelectOptions{
                    Name: "Paid",  // Target status option name
                },
            },
        },
    }

    _, err := s.client.UpdatePage(ctx, pageID, updateParams)
    if err != nil {
        return fmt.Errorf("failed to update expense status: %w", err)
    }

    return nil
}
```

**Key Points**:
- Use `DBPropTypeStatus` (not `DBPropTypeSelect`)
- Provide status option **name** in `SelectOptions.Name`
- Cannot create new status options via API (must exist in Notion UI)
- Must have "update content" permission for integration

### Status Mapping

| Notion Status | Workflow State | Payroll Action |
|---------------|----------------|----------------|
| Pending       | New request    | Skip (not ready) |
| Approved      | Ready for payout | Include in payroll |
| Paid          | Completed      | Skip (already paid) |

**Filter Logic**:
```go
// Fetch only approved expenses
filter := &notion.DatabaseQueryFilter{
    Property: "Status",
    DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
        Status: &notion.StatusDatabaseQueryFilter{
            Equals: "Approved",
        },
    },
}

// After payroll commit, mark as paid
updateParams := notion.UpdatePageParams{
    Properties: notion.DatabasePageProperties{
        "Status": notion.DatabasePageProperty{
            Type: notion.DBPropTypeStatus,
            Status: &notion.SelectOptions{
                Name: "Paid",
            },
        },
    },
}
```

## 5. Pagination Strategy

### Notion API Pagination

Notion uses cursor-based pagination with a maximum page size of 100:

```go
func (s *NotionExpenseService) fetchAllApprovedExpenses(ctx context.Context) ([]notion.Page, error) {
    var allPages []notion.Page
    var startCursor string
    hasMore := true

    filter := &notion.DatabaseQueryFilter{
        Property: "Status",
        DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
            Status: &notion.StatusDatabaseQueryFilter{
                Equals: "Approved",
            },
        },
    }

    for hasMore {
        query := &notion.DatabaseQuery{
            Filter:      filter,
            PageSize:    100,  // Maximum allowed
            StartCursor: startCursor,
        }

        result, err := s.client.QueryDatabase(ctx, s.cfg.ExpenseIntegration.Notion.ExpenseDBID, query)
        if err != nil {
            return nil, fmt.Errorf("query database failed: %w", err)
        }

        allPages = append(allPages, result.Results...)

        hasMore = result.HasMore
        if hasMore {
            startCursor = *result.NextCursor
        }
    }

    return allPages, nil
}
```

**Pagination Characteristics**:
- Maximum 100 entries per request
- Cursor-based (not offset-based)
- `HasMore` indicates if more results exist
- `NextCursor` provides cursor for next page
- First request: omit `StartCursor` (defaults to start)

### Rate Limiting Considerations

**Notion API Rate Limits**:
- 3 requests per second per integration
- Burst allowance for short spikes
- HTTP 429 status when rate limited

**Backoff Strategy**:
```go
func (s *NotionExpenseService) queryWithBackoff(ctx context.Context, dbID string, query *notion.DatabaseQuery) (notion.DatabaseQueryResponse, error) {
    maxRetries := 3
    baseDelay := 1 * time.Second

    for attempt := 0; attempt < maxRetries; attempt++ {
        result, err := s.client.QueryDatabase(ctx, dbID, query)

        if err != nil {
            // Check if rate limited
            if notionErr, ok := err.(*notion.Error); ok {
                if notionErr.Code == notion.ErrorCodeRateLimited {
                    // Exponential backoff
                    delay := baseDelay * time.Duration(1<<attempt)
                    s.logger.Warnf("Rate limited, retrying in %v (attempt %d/%d)", delay, attempt+1, maxRetries)
                    time.Sleep(delay)
                    continue
                }
            }
            return notion.DatabaseQueryResponse{}, err
        }

        return result, nil
    }

    return notion.DatabaseQueryResponse{}, fmt.Errorf("max retries exceeded")
}
```

## 6. Error Handling Patterns

### Notion API Errors

```go
func (s *NotionExpenseService) handleNotionError(err error, context string) error {
    if notionErr, ok := err.(*notion.Error); ok {
        switch notionErr.Code {
        case notion.ErrorCodeObjectNotFound:
            return fmt.Errorf("%s: database or page not found: %w", context, err)
        case notion.ErrorCodeUnauthorized:
            return fmt.Errorf("%s: invalid or missing Notion API token: %w", context, err)
        case notion.ErrorCodeRestrictedResource:
            return fmt.Errorf("%s: integration lacks required permissions: %w", context, err)
        case notion.ErrorCodeValidationError:
            return fmt.Errorf("%s: invalid request parameters: %w", context, err)
        case notion.ErrorCodeConflictError:
            return fmt.Errorf("%s: concurrent modification conflict: %w", context, err)
        case notion.ErrorCodeRateLimited:
            return fmt.Errorf("%s: rate limit exceeded: %w", context, err)
        case notion.ErrorCodeInternalServerError:
            return fmt.Errorf("%s: Notion service error: %w", context, err)
        default:
            return fmt.Errorf("%s: unknown Notion error (code: %s): %w", context, notionErr.Code, err)
        }
    }
    return fmt.Errorf("%s: %w", context, err)
}
```

### Graceful Degradation

```go
func (s *NotionExpenseService) transformPageToTodo(page notion.Page) (*bcModel.Todo, error) {
    props := page.Properties.(notion.DatabasePageProperties)

    // Extract properties with fallbacks
    title, err := s.extractTitle(props)
    if err != nil {
        s.logger.Error(err, "failed to extract title", "page_id", page.ID)
        return nil, fmt.Errorf("missing title: %w", err)
    }

    amount := s.extractAmount(props)  // Returns 0 on error
    if amount == 0 {
        s.logger.Warn("amount is zero or missing", "page_id", page.ID)
        // Continue processing (may be intentional zero amount)
    }

    currency := s.extractCurrency(props)  // Returns "VND" on error

    email, err := s.getEmail(context.Background(), props)
    if err != nil {
        s.logger.Error(err, "failed to extract email", "page_id", page.ID)
        return nil, fmt.Errorf("missing email: %w", err)
    }

    // Continue with transformation...
}
```

### Batch Processing with Error Continuation

```go
func (s *NotionExpenseService) GetAllInList(todolistID, projectID int) ([]bcModel.Todo, error) {
    pages, err := s.fetchAllApprovedExpenses(context.Background())
    if err != nil {
        return nil, fmt.Errorf("failed to fetch expenses: %w", err)
    }

    todos := make([]bcModel.Todo, 0, len(pages))
    errorCount := 0

    for _, page := range pages {
        todo, err := s.transformPageToTodo(page)
        if err != nil {
            s.logger.Error(err, "failed to transform page", "page_id", page.ID)
            errorCount++
            continue  // Skip invalid page, don't fail entire batch
        }
        todos = append(todos, *todo)
    }

    if errorCount > 0 {
        s.logger.Warnf("Processed %d expenses with %d errors", len(todos)+errorCount, errorCount)
    }

    return todos, nil
}
```

## 7. Configuration Management

### Configuration Structure

```go
type ExpenseNotionIntegration struct {
    ExpenseDBID    string  // Notion database ID for Expense Requests
    ContractorDBID string  // Notion database ID for Contractors (optional, for reference)
}
```

### Environment Variables

```bash
# Required
NOTION_EXPENSE_DB_ID=2bfb69f8-f573-81cb-a2da-f06d28896390

# Optional (for reference, not directly used if rollup configured)
NOTION_CONTRACTOR_DB_ID=2bfb69f8-f573-805a-8915-000bc44ce188

# Existing Notion config (already in use)
NOTION_SECRET=secret_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### Configuration Validation

```go
func (s *NotionExpenseService) validateConfig() error {
    if s.cfg.ExpenseIntegration.Notion.ExpenseDBID == "" {
        return fmt.Errorf("NOTION_EXPENSE_DB_ID is required but not configured")
    }

    // Validate database ID format (UUID with or without hyphens)
    dbID := s.cfg.ExpenseIntegration.Notion.ExpenseDBID
    cleanID := strings.ReplaceAll(dbID, "-", "")
    if len(cleanID) != 32 {
        return fmt.Errorf("NOTION_EXPENSE_DB_ID has invalid format (expected 32-char UUID): %s", dbID)
    }

    return nil
}
```

## 8. Performance Considerations

### Optimization Strategies

#### 1. Minimize API Calls

**Use Rollups Instead of Relations**:
```go
// Good: Single query with rollup (1 API call)
email := extractEmail(props["Email"])  // Rollup already contains email

// Avoid: Separate queries for each relation (N+1 problem)
for _, expense := range expenses {
    contractorPage := queryContractorPage(expense.Requestor)  // N additional calls
}
```

#### 2. Batch Employee Lookups

**Pre-fetch All Employees**:
```go
func (s *NotionExpenseService) GetAllInList(todolistID, projectID int) ([]bcModel.Todo, error) {
    pages, err := s.fetchAllApprovedExpenses(ctx)

    // Extract all unique emails
    emails := make(map[string]bool)
    for _, page := range pages {
        email, _ := s.extractEmail(page.Properties)
        emails[email] = true
    }

    // Batch fetch employees
    employeeMap := make(map[string]*model.Employee)
    for email := range emails {
        employee, err := s.store.Employee.OneByEmail(s.repo.DB(), email)
        if err == nil {
            employeeMap[email] = employee
        }
    }

    // Transform pages using cached employees
    for _, page := range pages {
        email, _ := s.extractEmail(page.Properties)
        employee := employeeMap[email]
        // ... use employee data
    }
}
```

#### 3. Concurrent Status Updates

**Update Multiple Expenses in Parallel**:
```go
func (h *handler) markExpenseSubmissionsAsCompleted(payrolls []*model.Payroll) {
    expenseIDs := extractExpenseIDsFromPayrolls(payrolls)

    // Use worker pool for concurrent updates
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 5)  // Limit to 5 concurrent updates

    for _, expenseID := range expenseIDs {
        wg.Add(1)
        go func(id string) {
            defer wg.Done()
            semaphore <- struct{}{}        // Acquire
            defer func() { <-semaphore }() // Release

            err := notionService.MarkExpenseAsCompleted(id)
            if err != nil {
                h.logger.Error(err, "failed to mark expense as completed", "expense_id", id)
            }
        }(expenseID)
    }

    wg.Wait()
}
```

### Caching Strategy

**Cache Employee Lookups**:
```go
type NotionExpenseService struct {
    client        *notion.Client
    cfg           *config.Config
    store         *store.Store
    repo          store.DBRepo
    logger        logger.Logger
    employeeCache map[string]*model.Employee  // Email → Employee
    cacheMu       sync.RWMutex
}

func (s *NotionExpenseService) getEmployeeByEmail(email string) (*model.Employee, error) {
    // Check cache first
    s.cacheMu.RLock()
    if employee, ok := s.employeeCache[email]; ok {
        s.cacheMu.RUnlock()
        return employee, nil
    }
    s.cacheMu.RUnlock()

    // Fetch from database
    employee, err := s.store.Employee.OneByEmail(s.repo.DB(), email)
    if err != nil {
        return nil, err
    }

    // Update cache
    s.cacheMu.Lock()
    s.employeeCache[email] = employee
    s.cacheMu.Unlock()

    return employee, nil
}
```

## 9. Testing Strategy

### Unit Test Structure

```go
func TestNotionExpenseService_GetAllInList(t *testing.T) {
    tests := []struct {
        name          string
        mockPages     []notion.Page
        mockError     error
        expectedTodos int
        expectedError bool
    }{
        {
            name: "success - single approved expense",
            mockPages: []notion.Page{
                createMockExpensePage("page-1", "Office supplies", 5000000, "VND", "employee@d.foundation", "Approved"),
            },
            expectedTodos: 1,
        },
        {
            name: "success - multiple approved expenses",
            mockPages: []notion.Page{
                createMockExpensePage("page-1", "Expense 1", 1000000, "VND", "emp1@d.foundation", "Approved"),
                createMockExpensePage("page-2", "Expense 2", 2000000, "VND", "emp2@d.foundation", "Approved"),
            },
            expectedTodos: 2,
        },
        {
            name:          "error - database query failed",
            mockError:     &notion.Error{Code: notion.ErrorCodeObjectNotFound},
            expectedError: true,
        },
        {
            name: "partial success - skip invalid page",
            mockPages: []notion.Page{
                createMockExpensePage("page-1", "Valid expense", 1000000, "VND", "emp@d.foundation", "Approved"),
                createInvalidExpensePage("page-2"),  // Missing required fields
            },
            expectedTodos: 1,  // Only 1 valid expense
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockClient := &MockNotionClient{
                QueryDatabaseFunc: func(ctx context.Context, id string, query *notion.DatabaseQuery) (notion.DatabaseQueryResponse, error) {
                    if tt.mockError != nil {
                        return notion.DatabaseQueryResponse{}, tt.mockError
                    }
                    return notion.DatabaseQueryResponse{
                        Results: tt.mockPages,
                        HasMore: false,
                    }, nil
                },
            }

            service := NewNotionExpenseService(mockClient, mockCfg, mockStore, mockRepo, mockLogger)

            todos, err := service.GetAllInList(0, 0)

            if tt.expectedError {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.Len(t, todos, tt.expectedTodos)
            }
        })
    }
}
```

### Integration Test

```go
func TestNotionExpenseService_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Use test Notion database
    testDBID := os.Getenv("NOTION_TEST_EXPENSE_DB_ID")
    if testDBID == "" {
        t.Skip("NOTION_TEST_EXPENSE_DB_ID not set")
    }

    client := notion.NewClient(os.Getenv("NOTION_SECRET"))
    service := NewNotionExpenseService(client, testCfg, testStore, testRepo, testLogger)

    // Test fetch
    todos, err := service.GetAllInList(0, 0)
    require.NoError(t, err)
    assert.GreaterOrEqual(t, len(todos), 0)

    // Verify todo format
    if len(todos) > 0 {
        todo := todos[0]
        assert.Greater(t, todo.ID, 0)
        assert.NotEmpty(t, todo.Title)
        assert.Contains(t, todo.Title, "|")  // Check format "desc | amount | currency"
    }
}
```

## Summary

### Critical Mapping Requirements

1. **ID Conversion**: Use hash-based UUID→int conversion with last 8 hex chars
2. **Email Extraction**: Use rollup property with fallback to direct relation query
3. **Property Extraction**: Handle typed Notion properties with defaults
4. **Title Format**: Build `"description | amount | currency"` exactly
5. **Status Filtering**: Use `StatusDatabaseQueryFilter` with "Approved" name
6. **Status Update**: Use `UpdatePage` with `DBPropTypeStatus` and "Paid" name
7. **Pagination**: Implement cursor-based pagination with 100-item pages
8. **Error Handling**: Continue on individual record errors, log and skip

### Performance Priorities

1. Use rollups to avoid N+1 queries
2. Batch employee lookups with caching
3. Implement concurrent status updates with rate limiting
4. Cache employee data per request cycle

### Testing Focus

1. Unit test property extraction with various types
2. Unit test pagination logic
3. Mock Notion client for isolated tests
4. Integration test with test database
5. Test error handling for all API error codes
