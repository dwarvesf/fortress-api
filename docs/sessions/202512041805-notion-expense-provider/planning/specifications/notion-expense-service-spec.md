# Notion Expense Service Technical Specification

## Overview

This document provides the detailed technical specification for the `NotionExpenseService` implementation, which integrates Notion as an expense provider for the fortress-api payroll system.

## Package Structure

```
pkg/service/notion/
├── expense.go           # ExpenseService implementation (NEW)
└── notion.go            # Existing Notion client wrapper
```

## Service Structure

### Type Definition

```go
// Location: pkg/service/notion/expense.go

type ExpenseService struct {
    client *Service          // Notion client wrapper (pkg/service/notion/notion.go)
    cfg    *config.Config    // Application configuration
    store  *store.Store      // Database store
    repo   store.DBRepo      // Database repository
    logger logger.Logger     // Structured logger
}
```

### Constructor

```go
// NewExpenseService creates a new Notion expense service for payroll integration.
//
// Parameters:
//   - client: Notion API client wrapper
//   - cfg: Application configuration (must include ExpenseIntegration.Notion)
//   - store: Database store for employee lookups
//   - repo: Database repository
//   - logger: Structured logger
//
// Returns:
//   - *ExpenseService: Initialized service
//
// Panics if:
//   - cfg.ExpenseIntegration.Notion.ExpenseDBID is empty
//   - ExpenseDBID is not a valid Notion database ID
func NewExpenseService(
    client *Service,
    cfg *config.Config,
    store *store.Store,
    repo store.DBRepo,
    logger logger.Logger,
) *ExpenseService {
    // Validate configuration
    if cfg.ExpenseIntegration.Notion.ExpenseDBID == "" {
        logger.Fatal("NOTION_EXPENSE_DB_ID is required but not configured")
    }

    // Validate database ID format (32 hex chars with or without hyphens)
    dbID := strings.ReplaceAll(cfg.ExpenseIntegration.Notion.ExpenseDBID, "-", "")
    if len(dbID) != 32 {
        logger.Fatal("NOTION_EXPENSE_DB_ID has invalid format",
            "expected", "32-character UUID",
            "got", cfg.ExpenseIntegration.Notion.ExpenseDBID,
        )
    }

    return &ExpenseService{
        client: client,
        cfg:    cfg,
        store:  store,
        repo:   repo,
        logger: logger,
    }
}
```

## Interface Implementation

### ExpenseProvider Interface

```go
// From pkg/service/basecamp/basecamp.go
type ExpenseProvider interface {
    GetAllInList(todolistID, projectID int) ([]model.Todo, error)
    GetGroups(todosetID, projectID int) ([]model.TodoGroup, error)
    GetLists(projectID, todosetID int) ([]model.TodoList, error)
}
```

### Method: GetAllInList

```go
// GetAllInList fetches all approved expenses from the Notion Expense Request database
// and transforms them into bcModel.Todo format for payroll calculation.
//
// Parameters:
//   - todolistID: Ignored (Notion uses database ID from config)
//   - projectID: Ignored (Notion uses database ID from config)
//
// Returns:
//   - []model.Todo: List of approved expenses in Basecamp Todo format
//   - error: Error if query fails or transformation fails for all records
//
// Behavior:
//   - Queries Notion database for expenses with Status = "Approved"
//   - Transforms each page to bcModel.Todo format
//   - Logs errors for individual pages but continues processing
//   - Returns partial results if some pages fail transformation
//   - Returns error only if query fails or all transformations fail
func (s *ExpenseService) GetAllInList(todolistID, projectID int) ([]bcModel.Todo, error) {
    ctx := context.Background()

    // Fetch all approved expenses from Notion
    pages, err := s.fetchApprovedExpenses(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch approved expenses: %w", err)
    }

    s.logger.Info("Fetched approved expenses from Notion",
        "count", len(pages),
        "database_id", s.cfg.ExpenseIntegration.Notion.ExpenseDBID,
    )

    // Transform each page to Todo format
    todos := make([]bcModel.Todo, 0, len(pages))
    errorCount := 0

    for _, page := range pages {
        todo, err := s.transformPageToTodo(ctx, page)
        if err != nil {
            s.logger.Error(err, "Failed to transform Notion page to Todo",
                "page_id", page.ID,
            )
            errorCount++
            continue
        }

        todos = append(todos, *todo)
    }

    if errorCount > 0 {
        s.logger.Warn("Completed expense transformation with errors",
            "total_pages", len(pages),
            "successful", len(todos),
            "failed", errorCount,
        )
    }

    if len(todos) == 0 && len(pages) > 0 {
        return nil, fmt.Errorf("failed to transform all %d expense pages", len(pages))
    }

    return todos, nil
}
```

### Method: GetGroups

```go
// GetGroups returns empty list (not used with Notion provider).
//
// Notion database structure does not use hierarchical groups like Basecamp.
// All expenses are fetched via GetAllInList.
//
// Parameters:
//   - todosetID: Ignored
//   - projectID: Ignored
//
// Returns:
//   - []model.TodoGroup: Empty list
//   - error: nil
func (s *ExpenseService) GetGroups(todosetID, projectID int) ([]bcModel.TodoGroup, error) {
    s.logger.Debug("GetGroups called (not used with Notion provider)",
        "todosetID", todosetID,
        "projectID", projectID,
    )
    return []bcModel.TodoGroup{}, nil
}
```

### Method: GetLists

```go
// GetLists returns empty list (not used with Notion provider).
//
// Notion database structure does not use lists like Basecamp.
// All expenses are fetched via GetAllInList.
//
// Parameters:
//   - projectID: Ignored
//   - todosetID: Ignored
//
// Returns:
//   - []model.TodoList: Empty list
//   - error: nil
func (s *ExpenseService) GetLists(projectID, todosetID int) ([]bcModel.TodoList, error) {
    s.logger.Debug("GetLists called (not used with Notion provider)",
        "projectID", projectID,
        "todosetID", todosetID,
    )
    return []bcModel.TodoList{}, nil
}
```

## Core Methods

### fetchApprovedExpenses

```go
// fetchApprovedExpenses queries the Notion Expense Request database for all
// expenses with Status = "Approved" using cursor-based pagination.
//
// Returns:
//   - []notion.Page: List of approved expense pages
//   - error: Error if query fails
//
// Behavior:
//   - Queries with status filter: Status equals "Approved"
//   - Handles pagination automatically (max 100 per page)
//   - Retries on rate limit errors with exponential backoff
func (s *ExpenseService) fetchApprovedExpenses(ctx context.Context) ([]notion.Page, error) {
    dbID := s.cfg.ExpenseIntegration.Notion.ExpenseDBID

    // Build status filter
    filter := &notion.DatabaseQueryFilter{
        Property: "Status",
        DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
            Status: &notion.StatusDatabaseQueryFilter{
                Equals: "Approved",
            },
        },
    }

    // Paginate through all results
    var allPages []notion.Page
    var startCursor *notion.Cursor

    for {
        query := &notion.DatabaseQuery{
            Filter:      filter,
            PageSize:    100,  // Maximum allowed
            StartCursor: startCursor,
        }

        result, err := s.client.QueryDatabase(ctx, dbID, query)
        if err != nil {
            return nil, s.handleNotionError(err, "query database")
        }

        allPages = append(allPages, result.Results...)

        if !result.HasMore {
            break
        }

        startCursor = result.NextCursor
    }

    return allPages, nil
}
```

### transformPageToTodo

```go
// transformPageToTodo converts a Notion expense page to bcModel.Todo format.
//
// Parameters:
//   - ctx: Context for database queries
//   - page: Notion page from Expense Request database
//
// Returns:
//   - *bcModel.Todo: Transformed expense todo
//   - error: Error if transformation fails (missing required fields, employee not found)
//
// Transformation steps:
//   1. Extract page properties (Title, Amount, Currency, Email)
//   2. Lookup employee by email → BasecampID
//   3. Build Todo title in format: "description | amount | currency"
//   4. Create Todo with assignee and metadata
func (s *ExpenseService) transformPageToTodo(ctx context.Context, page notion.Page) (*bcModel.Todo, error) {
    props := page.Properties.(notion.DatabasePageProperties)

    // Extract title
    title, err := s.extractTitle(props)
    if err != nil {
        return nil, fmt.Errorf("failed to extract title: %w", err)
    }

    // Extract amount
    amount, err := s.extractAmount(props)
    if err != nil {
        return nil, fmt.Errorf("failed to extract amount: %w", err)
    }

    // Extract currency (with default)
    currency := s.extractCurrency(props)

    // Extract email from rollup
    email, err := s.getEmail(ctx, props)
    if err != nil {
        return nil, fmt.Errorf("failed to extract email: %w", err)
    }

    // Lookup employee by email
    employee, err := s.store.Employee.OneByEmail(s.repo.DB(), email)
    if err != nil {
        return nil, fmt.Errorf("employee not found for email %s: %w", email, err)
    }

    // Validate employee has BasecampID
    if employee.BasecampID == 0 {
        return nil, fmt.Errorf("employee %s has no basecamp_id", email)
    }

    // Convert Notion page UUID to integer ID
    intID := s.notionPageIDToInt(page.ID)

    // Build Todo title in payroll format
    todoTitle := fmt.Sprintf("%s | %.0f | %s", title, amount, currency)

    // Extract attachment URL (first file if available)
    attachmentURL := s.extractAttachmentURL(props)

    // Create Todo
    todo := &bcModel.Todo{
        ID:    intID,
        Title: todoTitle,
        Assignees: []bcModel.Assignee{
            {
                ID:   employee.BasecampID,
                Name: employee.FullName,
            },
        },
        Bucket: bcModel.Bucket{
            ID:   intID,  // Use same hash-based ID
            Name: "Notion Expenses",  // Static bucket name
        },
        Completed: true,  // Approved expenses are considered "completed"
    }

    s.logger.Debug("Transformed Notion page to Todo",
        "page_id", page.ID,
        "int_id", intID,
        "employee", employee.TeamEmail,
        "amount", amount,
        "currency", currency,
    )

    return todo, nil
}
```

## Property Extraction Methods

### extractTitle

```go
// extractTitle extracts the plain text title from a Notion title property.
//
// Parameters:
//   - props: Page properties map
//
// Returns:
//   - string: Concatenated plain text from all rich text segments
//   - error: Error if Title property is missing or invalid
func (s *ExpenseService) extractTitle(props notion.DatabasePageProperties) (string, error) {
    titleProp, ok := props["Title"]
    if !ok {
        return "", fmt.Errorf("Title property not found")
    }

    if titleProp.Type != notion.DBPropTypeTitle {
        return "", fmt.Errorf("Title property has wrong type: %s", titleProp.Type)
    }

    if len(titleProp.Title) == 0 {
        return "", fmt.Errorf("Title property is empty")
    }

    // Concatenate all rich text segments
    var parts []string
    for _, rt := range titleProp.Title {
        parts = append(parts, rt.PlainText)
    }

    title := strings.Join(parts, "")
    if title == "" {
        return "", fmt.Errorf("Title text is empty")
    }

    return title, nil
}
```

### extractAmount

```go
// extractAmount extracts the numeric amount from a Notion number property.
//
// Parameters:
//   - props: Page properties map
//
// Returns:
//   - float64: Expense amount
//   - error: Error if Amount property is missing or zero
func (s *ExpenseService) extractAmount(props notion.DatabasePageProperties) (float64, error) {
    amountProp, ok := props["Amount"]
    if !ok {
        return 0, fmt.Errorf("Amount property not found")
    }

    if amountProp.Type != notion.DBPropTypeNumber {
        return 0, fmt.Errorf("Amount property has wrong type: %s", amountProp.Type)
    }

    amount := amountProp.Number
    if amount == 0 {
        return 0, fmt.Errorf("Amount is zero")
    }

    return amount, nil
}
```

### extractCurrency

```go
// extractCurrency extracts the currency code from a Notion select property.
//
// Parameters:
//   - props: Page properties map
//
// Returns:
//   - string: Currency code (e.g., "VND", "USD"), defaults to "VND" if missing
//
// Note: Never returns error, always provides default value
func (s *ExpenseService) extractCurrency(props notion.DatabasePageProperties) string {
    currencyProp, ok := props["Currency"]
    if !ok {
        s.logger.Warn("Currency property not found, using default", "default", "VND")
        return "VND"
    }

    if currencyProp.Type != notion.DBPropTypeSelect {
        s.logger.Warn("Currency property has wrong type, using default",
            "type", currencyProp.Type,
            "default", "VND",
        )
        return "VND"
    }

    if currencyProp.Select == nil || currencyProp.Select.Name == "" {
        s.logger.Warn("Currency select is empty, using default", "default", "VND")
        return "VND"
    }

    return currencyProp.Select.Name
}
```

### getEmail

```go
// getEmail extracts the employee email using rollup-first strategy with fallback.
//
// Parameters:
//   - ctx: Context for potential relation query
//   - props: Page properties map
//
// Returns:
//   - string: Employee email address
//   - error: Error if email cannot be extracted from rollup or relation
//
// Strategy:
//   1. Try extracting from Email rollup property (primary, efficient)
//   2. Fall back to querying Requestor relation directly (secondary, slower)
func (s *ExpenseService) getEmail(ctx context.Context, props notion.DatabasePageProperties) (string, error) {
    // Try rollup first
    email, err := s.extractEmailFromRollup(props)
    if err == nil && email != "" {
        return email, nil
    }

    s.logger.Warn("Rollup extraction failed, falling back to relation query",
        "rollup_error", err,
    )

    // Fallback to direct relation query
    return s.extractEmailFromRelation(ctx, props)
}
```

### extractEmailFromRollup

```go
// extractEmailFromRollup extracts email from the Email rollup property.
//
// Handles multiple rollup types:
//   - Array: Extract from first array element (most common)
//   - String: Use string value directly (some configurations)
//
// Returns:
//   - string: Email address
//   - error: Error if rollup is missing, wrong type, or empty
func (s *ExpenseService) extractEmailFromRollup(props notion.DatabasePageProperties) (string, error) {
    emailProp, ok := props["Email"]
    if !ok {
        return "", fmt.Errorf("Email property not found")
    }

    if emailProp.Type != notion.DBPropTypeRollup {
        return "", fmt.Errorf("Email property is not a rollup (type: %s)", emailProp.Type)
    }

    rollup := emailProp.Rollup

    switch rollup.Type {
    case notion.RollupTypeArray:
        if len(rollup.Array) == 0 {
            return "", fmt.Errorf("rollup array is empty")
        }

        // Type assert first element to DatabasePageProperty
        propVal, ok := rollup.Array[0].(notion.DatabasePageProperty)
        if !ok {
            return "", fmt.Errorf("rollup array element is not a property")
        }

        if propVal.Type == notion.DBPropTypeEmail {
            if propVal.Email == "" {
                return "", fmt.Errorf("email property is empty")
            }
            return propVal.Email, nil
        }

        return "", fmt.Errorf("rollup array element is not an email (type: %s)", propVal.Type)

    case notion.RollupTypeString:
        if rollup.String == "" {
            return "", fmt.Errorf("rollup string is empty")
        }
        return rollup.String, nil

    default:
        return "", fmt.Errorf("unsupported rollup type: %s", rollup.Type)
    }
}
```

### extractEmailFromRelation

```go
// extractEmailFromRelation queries the Requestor relation to fetch email directly.
//
// This is a fallback method when rollup extraction fails.
//
// Returns:
//   - string: Email address from contractor page
//   - error: Error if relation is missing, empty, or contractor page lacks email
func (s *ExpenseService) extractEmailFromRelation(ctx context.Context, props notion.DatabasePageProperties) (string, error) {
    requestorProp, ok := props["Requestor"]
    if !ok {
        return "", fmt.Errorf("Requestor property not found")
    }

    if requestorProp.Type != notion.DBPropTypeRelation {
        return "", fmt.Errorf("Requestor property is not a relation (type: %s)", requestorProp.Type)
    }

    if len(requestorProp.Relation) == 0 {
        return "", fmt.Errorf("Requestor relation is empty (no contractor linked)")
    }

    // Get first related contractor page ID
    contractorPageID := requestorProp.Relation[0].ID

    // Query contractor page
    contractorPage, err := s.client.FindPageByID(ctx, contractorPageID)
    if err != nil {
        return "", fmt.Errorf("failed to fetch contractor page %s: %w", contractorPageID, err)
    }

    // Extract email from contractor properties
    contractorProps := contractorPage.Properties.(notion.DatabasePageProperties)
    emailProp, ok := contractorProps["Email"]
    if !ok {
        return "", fmt.Errorf("Email property not found in contractor page")
    }

    if emailProp.Type != notion.DBPropTypeEmail {
        return "", fmt.Errorf("contractor Email property has wrong type: %s", emailProp.Type)
    }

    if emailProp.Email == "" {
        return "", fmt.Errorf("contractor Email property is empty")
    }

    return emailProp.Email, nil
}
```

### extractAttachmentURL

```go
// extractAttachmentURL extracts the URL of the first attachment file.
//
// Parameters:
//   - props: Page properties map
//
// Returns:
//   - string: URL of first attachment, empty string if none
//
// Note: Notion file URLs expire after ~1 hour, use immediately
func (s *ExpenseService) extractAttachmentURL(props notion.DatabasePageProperties) string {
    attachmentsProp, ok := props["Attachments"]
    if !ok {
        return ""
    }

    if attachmentsProp.Type != notion.DBPropTypeFiles {
        return ""
    }

    if len(attachmentsProp.Files) == 0 {
        return ""
    }

    firstFile := attachmentsProp.Files[0]

    switch firstFile.Type {
    case notion.FileTypeFile:
        if firstFile.File != nil {
            return firstFile.File.URL
        }
    case notion.FileTypeExternal:
        if firstFile.External != nil {
            return firstFile.External.URL
        }
    }

    return ""
}
```

## ID Conversion Method

### notionPageIDToInt

```go
// notionPageIDToInt converts a Notion page UUID to an integer using hash-based conversion.
//
// Conversion algorithm:
//   1. Remove hyphens from UUID: "2bfb69f8-f573-81cb-a2da-f06d28896390" → "2bfb69f8f57381cba2daf06d28896390"
//   2. Take last 8 hex characters: "28896390"
//   3. Convert hex to int: 0x28896390 → 680141712
//
// Parameters:
//   - pageID: Notion page UUID (with or without hyphens)
//
// Returns:
//   - int: Deterministic integer representation
//
// Note: Same UUID always produces same int (deterministic)
func (s *ExpenseService) notionPageIDToInt(pageID string) int {
    // Remove hyphens
    cleanID := strings.ReplaceAll(pageID, "-", "")

    // Take last 8 hex chars (32 bits)
    if len(cleanID) < 8 {
        s.logger.Warn("Page ID too short for hash conversion",
            "page_id", pageID,
            "clean_length", len(cleanID),
        )
        // Fallback: use CRC32 if UUID is malformed
        return int(crc32.ChecksumIEEE([]byte(pageID)))
    }

    hashStr := cleanID[len(cleanID)-8:]

    // Convert hex to int
    hash, err := strconv.ParseInt(hashStr, 16, 64)
    if err != nil {
        s.logger.Error(err, "Failed to parse page ID hash",
            "page_id", pageID,
            "hash_str", hashStr,
        )
        // Fallback: use CRC32 if parsing fails
        return int(crc32.ChecksumIEEE([]byte(pageID)))
    }

    return int(hash)
}
```

## Status Update Method

### MarkExpenseAsCompleted

```go
// MarkExpenseAsCompleted updates the status of an expense from "Approved" to "Paid".
//
// This method is called after payroll commit to mark expenses as completed.
//
// Parameters:
//   - pageID: Notion page UUID (stored in CommissionExplain metadata)
//
// Returns:
//   - error: Error if status update fails
//
// Note: This method expects the UUID, not the integer ID
func (s *ExpenseService) MarkExpenseAsCompleted(pageID string) error {
    ctx := context.Background()

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

    _, err := s.client.UpdatePage(ctx, pageID, updateParams)
    if err != nil {
        return s.handleNotionError(err, fmt.Sprintf("update status for page %s", pageID))
    }

    s.logger.Info("Marked expense as completed",
        "page_id", pageID,
        "status", "Paid",
    )

    return nil
}
```

## Error Handling Method

### handleNotionError

```go
// handleNotionError wraps Notion API errors with additional context.
//
// Parameters:
//   - err: Original error from Notion API
//   - context: Description of the operation that failed
//
// Returns:
//   - error: Wrapped error with context
func (s *ExpenseService) handleNotionError(err error, context string) error {
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
        case notion.ErrorCodeRateLimited:
            return fmt.Errorf("%s: rate limit exceeded: %w", context, err)
        case notion.ErrorCodeInternalServerError:
            return fmt.Errorf("%s: Notion service error: %w", context, err)
        default:
            return fmt.Errorf("%s: Notion error (code: %s): %w", context, notionErr.Code, err)
        }
    }
    return fmt.Errorf("%s: %w", context, err)
}
```

## Configuration Requirements

### Environment Variables

```bash
# Required
TASK_PROVIDER=notion
NOTION_SECRET=secret_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
NOTION_EXPENSE_DB_ID=2bfb69f8-f573-81cb-a2da-f06d28896390

# Optional (for reference only, not used if rollup configured)
NOTION_CONTRACTOR_DB_ID=2bfb69f8-f573-805a-8915-000bc44ce188
```

### Configuration Structure

```go
// In pkg/config/config.go

type ExpenseNotionIntegration struct {
    ExpenseDBID    string  // NOTION_EXPENSE_DB_ID
    ContractorDBID string  // NOTION_CONTRACTOR_DB_ID (optional)
}
```

## Dependencies

### External Packages

- `github.com/dstotijn/go-notion` - Notion API client
- `context` - Context for API calls
- `fmt` - Error formatting
- `strings` - String manipulation
- `strconv` - String to int conversion
- `hash/crc32` - Fallback hash function

### Internal Packages

- `pkg/model` - Data models (bcModel.Todo, Employee, CommissionExplain)
- `pkg/store` - Database stores (Employee store)
- `pkg/config` - Configuration management
- `pkg/logger` - Structured logging

## Testing Requirements

### Unit Tests

Location: `pkg/service/notion/expense_test.go`

Required test cases:

1. **Property Extraction**:
   - Test extractTitle with valid title
   - Test extractTitle with empty title (error)
   - Test extractAmount with valid amount
   - Test extractAmount with zero amount (error)
   - Test extractCurrency with valid currency
   - Test extractCurrency with missing currency (default to VND)

2. **Email Extraction**:
   - Test extractEmailFromRollup with array type
   - Test extractEmailFromRollup with string type
   - Test extractEmailFromRelation (fallback)
   - Test getEmail with successful rollup
   - Test getEmail with fallback to relation

3. **ID Conversion**:
   - Test notionPageIDToInt with valid UUID
   - Test notionPageIDToInt produces deterministic results
   - Test notionPageIDToInt with malformed UUID (fallback)

4. **Transformation**:
   - Test transformPageToTodo with valid page
   - Test transformPageToTodo with missing employee (error)
   - Test transformPageToTodo with employee lacking BasecampID (error)
   - Test title format: "description | amount | currency"

5. **Fetch and Transform**:
   - Test GetAllInList with single expense
   - Test GetAllInList with multiple expenses
   - Test GetAllInList with partial failures (continue processing)
   - Test GetAllInList with all failures (return error)

6. **Status Update**:
   - Test MarkExpenseAsCompleted success
   - Test MarkExpenseAsCompleted with invalid page ID (error)

### Integration Tests

Location: `pkg/service/notion/expense_integration_test.go`

Required test cases:

1. Query test Notion database for approved expenses
2. Transform real pages to Todos
3. Validate employee lookup works
4. Update expense status to "Paid"
5. Verify round-trip: fetch → transform → update

## Performance Considerations

### API Call Optimization

1. **Single Query**: All expenses fetched in one paginated query (not N separate queries)
2. **Rollup Usage**: Email extracted from rollup (avoids N+1 relation queries)
3. **Batch Employee Lookup**: Consider caching employees by email for repeated lookups

### Pagination

- Notion max page size: 100 items
- Use cursor-based pagination for large datasets
- Expected: <100 approved expenses per payroll cycle (single page)

### Rate Limiting

- Notion API: 3 requests per second
- Fetch operation: 1 request for expenses + 0-N requests for relations (if rollup fails)
- Status updates: Can be concurrent with semaphore limiting

## Error Handling Strategy

### Graceful Degradation

- **Individual Page Errors**: Log and skip, continue processing other pages
- **Query Errors**: Fail entire operation, return error
- **Employee Not Found**: Skip expense, log error (don't fail payroll)
- **Missing Properties**: Skip expense, log error with details

### Logging Levels

- **Debug**: Property extraction details, transformation steps
- **Info**: Fetch count, transformation summary, status updates
- **Warn**: Rollup fallback, missing optional properties, partial failures
- **Error**: Query failures, transformation failures, API errors

## Migration Path

### Phase 1: Initial Implementation

- Implement ExpenseService with all methods
- Use hash-based ID conversion
- Store UUID in metadata for reverse lookup
- Deploy with `TASK_PROVIDER=notion`

### Phase 2: Production Validation

- Monitor error rates and transformation failures
- Validate all expenses processed correctly
- Verify status updates work after payroll commit
- Compare results with NocoDB (parallel run if possible)

### Phase 3: Optimization

- Add employee caching if needed
- Implement concurrent status updates
- Optimize property extraction (reduce allocations)
- Add performance metrics and dashboards

## References

- **ADR-001**: UUID to Int Mapping Strategy
- **ADR-002**: Email Extraction Strategy
- **ADR-003**: Provider Selection
- **Research**: Notion API Patterns (`research/notion-api-patterns.md`)
- **Research**: Technical Considerations (`research/technical-considerations.md`)
- **NocoDB Reference**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/nocodb/expense.go`
