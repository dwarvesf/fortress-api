# Existing Implementation Analysis

## Overview

This document analyzes the existing codebase patterns for expense providers, focusing on NocoDB and Basecamp implementations to inform the Notion Expense Provider implementation.

## 1. ExpenseProvider Interface

### Interface Definition

**Location**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/basecamp/basecamp.go` (lines 50-55)

```go
// ExpenseProvider defines methods for fetching expense todos for payroll calculation.
type ExpenseProvider interface {
    GetAllInList(todolistID, projectID int) ([]model.Todo, error)
    GetGroups(todosetID, projectID int) ([]model.TodoGroup, error)
    GetLists(projectID, todosetID int) ([]model.TodoList, error)
}
```

**Key Characteristics**:
- Interface uses integer IDs (legacy from Basecamp API)
- Returns `bcModel.Todo` format for payroll compatibility
- Three methods for hierarchical data fetching (Lists → Groups → Todos)
- No direct "mark as completed" method in interface (implementation-specific)

**Usage Context**:
- Used in `pkg/service/service.go` as `PayrollExpenseProvider basecamp.ExpenseProvider`
- Distinct from webhook `ExpenseProvider` in `pkg/service/taskprovider/expense.go`
- Called during payroll calculation, not during webhook processing

## 2. NocoDB Implementation

### Service Structure

**Location**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/nocodb/expense.go`

```go
type ExpenseService struct {
    client *Service          // NocoDB API client
    cfg    *config.Config    // Configuration with table IDs
    store  *store.Store      // Database access
    repo   store.DBRepo      // Database repository
    logger logger.Logger     // Structured logging
}
```

**Key Features**:
1. **Query Pattern**: Uses NocoDB REST API with `where` clause filters
2. **Status Filtering**: `(status,eq,approved)` for approved expenses
3. **Data Transformation**: Maps NocoDB records to `bcModel.Todo` format
4. **Employee Linking**: Fetches employee by email from `requester_team_email` field
5. **Title Format**: Builds `"title | amount | currency"` for payroll parsing

### GetAllInList Implementation (Lines 43-102)

```go
func (e *ExpenseService) GetAllInList(todolistID, projectID int) ([]bcModel.Todo, error) {
    // 1. Get table ID from config
    tableID := e.cfg.ExpenseIntegration.Noco.TableID

    // 2. Build query with status filter
    path := fmt.Sprintf("/tables/%s/records", tableID)
    query := url.Values{}
    query.Set("where", "(status,eq,approved)")
    query.Set("limit", "100")

    // 3. Make HTTP request to NocoDB
    resp, err := e.client.makeRequest(ctx, http.MethodGet, path, query, nil)

    // 4. Parse JSON response
    var result struct {
        List []map[string]interface{} `json:"list"`
    }
    json.NewDecoder(resp.Body).Decode(&result)

    // 5. Transform each record to bcModel.Todo
    for _, record := range result.List {
        todo, err := e.transformRecordToTodo(record)
        todos = append(todos, *todo)
    }

    return todos, nil
}
```

**Key Patterns**:
- Table ID from config: `cfg.ExpenseIntegration.Noco.TableID`
- HTTP-based API calls with query parameters
- Generic `map[string]interface{}` for record parsing
- Transformation layer (`transformRecordToTodo`)
- Error handling with logging and continuation

### Record Transformation (Lines 231-295)

```go
func (e *ExpenseService) transformRecordToTodo(record map[string]interface{}) (*bcModel.Todo, error) {
    // Extract fields
    recordID := extractRecordID(record)            // Integer ID
    requesterEmail := extractString(record, "requester_team_email")
    amount := extractFloat(record, "amount")
    currency := extractString(record, "currency")
    title := extractString(record, "title")
    taskBoard := extractString(record, "task_board")

    // Fetch employee by email
    employee, err := e.store.Employee.OneByEmail(e.repo.DB(), requesterEmail)
    if err != nil {
        return nil, fmt.Errorf("employee not found for email %s: %w", requesterEmail, err)
    }

    // Validate employee has basecamp_id
    if employee.BasecampID == 0 {
        return nil, fmt.Errorf("employee %s has no basecamp_id", requesterEmail)
    }

    // Build Todo title in payroll format
    todoTitle := fmt.Sprintf("%s | %.0f | %s", title, amount, currency)

    // Create Todo object
    return &bcModel.Todo{
        ID:    int(recordID),
        Title: todoTitle,
        Assignees: []bcModel.Assignee{
            {
                ID:   employee.BasecampID,
                Name: employee.FullName,
            },
        },
        Bucket: bcModel.Bucket{
            ID:   int(recordID),
            Name: taskBoard,
        },
        Completed: true,  // Approved expenses are "completed"
    }, nil
}
```

**Critical Transformation Logic**:
1. **ID Mapping**: NocoDB integer ID → Todo.ID
2. **Email Lookup**: `requester_team_email` → Employee → BasecampID
3. **Title Format**: `"description | amount | currency"` (exact format required for payroll)
4. **Assignee Mapping**: Employee.BasecampID and FullName
5. **Bucket Mapping**: Record ID and task_board name
6. **Completed Flag**: Always `true` for approved expenses

### Helper Functions (Lines 309-344)

```go
func extractString(record map[string]interface{}, key string) string {
    if val, ok := record[key]; ok {
        if str, ok := val.(string); ok {
            return str
        }
    }
    return ""
}

func extractFloat(record map[string]interface{}, key string) float64 {
    if val, ok := record[key]; ok {
        switch v := val.(type) {
        case float64:
            return v
        case int:
            return float64(v)
        case string:
            if f, err := strconv.ParseFloat(v, 64); err == nil {
                return f
            }
        }
    }
    return 0
}

func extractRecordID(record map[string]interface{}) string {
    // NocoDB uses "Id" or "id" for primary key
    if id, ok := record["Id"]; ok {
        return fmt.Sprintf("%v", id)
    }
    if id, ok := record["id"]; ok {
        return fmt.Sprintf("%v", id)
    }
    return ""
}
```

**Helper Patterns**:
- Graceful type conversion with fallbacks
- Multiple type support (string, int, float64 for amount)
- Case-insensitive field access ("Id" vs "id")
- Default values on extraction failure (empty string, 0)

### Mark as Completed (Lines 346-389)

```go
func (e *ExpenseService) MarkExpenseAsCompleted(expenseID int) error {
    tableID := e.cfg.ExpenseIntegration.Noco.TableID

    path := fmt.Sprintf("/tables/%s/records", tableID)
    payload := map[string]interface{}{
        "Id":     expenseID,
        "status": "completed",  // Update status field
    }

    resp, err := e.client.makeRequest(ctx, http.MethodPatch, path, nil, payload)
    // Error handling...

    return nil
}
```

**Update Pattern**:
- HTTP PATCH to `/tables/{tableID}/records`
- Payload includes `Id` and updated fields
- Status update: `"approved"` → `"completed"`

## 3. Accounting Todo Service (NocoDB)

### Service Structure

**Location**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/nocodb/accounting_todo.go`

Similar to ExpenseService but focuses on accounting todos:

```go
type AccountingTodoService struct {
    client *Service
    cfg    *config.Config
    store  *store.Store
    repo   store.DBRepo
    logger logger.Logger
}
```

**Key Differences**:
1. **Status Filter**: `(task_group,eq,out)~and(status,neq,completed)` (lines 61)
2. **Assignee Filtering**: Excludes Han (approver) from assignees (lines 273-280)
3. **Title Parsing**: Parses format `"Description | Amount | Currency"` (lines 185-210)
4. **Assignee Resolution**: Fetches employees by `basecamp_id` (lines 272-299)

**Title Parsing Logic** (Lines 185-210):
```go
func (a *AccountingTodoService) parseTodoTitle(title string) (string, float64, string, error) {
    parts := strings.Split(title, "|")
    if len(parts) != 3 {
        return "", 0, "", fmt.Errorf("invalid title format")
    }

    description := strings.TrimSpace(parts[0])
    amountStr := strings.TrimSpace(parts[1])
    currency := strings.TrimSpace(parts[2])

    // Default to VND if currency is empty
    if currency == "" {
        currency = "VND"
    }

    // Parse amount (remove thousand separators)
    amountStr = strings.ReplaceAll(amountStr, ",", "")
    amountStr = strings.ReplaceAll(amountStr, ".", "")
    amount, err := strconv.ParseFloat(amountStr, 64)

    return description, amount, currency, nil
}
```

**Assignee ID Parsing** (Lines 216-269):
```go
func (a *AccountingTodoService) parseAssigneeIDs(assigneeIDsRaw interface{}) ([]int, error) {
    // Handle multiple formats:
    // - JSON array: ["123", "456"]
    // - Comma-separated string: "123,456"
    // - Array of interface{}: []interface{}{123, 456}

    switch v := assigneeIDsRaw.(type) {
    case []interface{}:
        // Parse array items
    case string:
        // Try JSON, fallback to comma-separated
        if err := json.Unmarshal([]byte(str), &jsonIDs); err == nil {
            // JSON array
        } else {
            // Comma-separated
            parts := strings.Split(str, ",")
        }
    }
}
```

## 4. Basecamp Implementation

### Service Structure

**Location**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/basecamp/basecamp.go`

Basecamp service has native support for todos through its API client:

```go
type Service struct {
    store    *store.Store
    repo     store.DBRepo
    config   *config.Config
    logger   logger.Logger
    Basecamp *model.Basecamp
    Client   client.Service
    Todo     todo.Service  // Basecamp todo service
    // ... other services
}
```

**Basecamp Todo Service Methods** (Lines 50-55):
- `GetAllInList(todolistID, projectID int) ([]model.Todo, error)`
- `GetGroups(todosetID, projectID int) ([]model.TodoGroup, error)`
- `GetLists(projectID, todosetID int) ([]model.TodoList, error)`

**Native Implementation**:
- Direct API calls to Basecamp endpoints
- Native `bcModel.Todo` format (no transformation needed)
- Handles approval via comments (checks for approver comment)

### Expense Adapter Pattern

**Location**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/basecamp/expense_adapter.go`

```go
// ExpenseAdapter wraps the Basecamp service to implement the ExpenseProvider interface.
type ExpenseAdapter struct {
    svc *Service
}

func NewExpenseAdapter(svc *Service) *ExpenseAdapter {
    return &ExpenseAdapter{svc: svc}
}

func (a *ExpenseAdapter) GetAllInList(todolistID, projectID int) ([]model.Todo, error) {
    return a.svc.Todo.GetAllInList(todolistID, projectID)
}

func (a *ExpenseAdapter) GetGroups(todosetID, projectID int) ([]model.TodoGroup, error) {
    return a.svc.Todo.GetGroups(todosetID, projectID)
}

func (a *ExpenseAdapter) GetLists(projectID, todosetID int) ([]model.TodoList, error) {
    return a.svc.Todo.GetLists(projectID, todosetID)
}
```

**Adapter Pattern Benefits**:
- Wraps Basecamp.Todo service to satisfy `ExpenseProvider` interface
- No data transformation required (native format)
- Simple pass-through implementation
- Maintains separation between webhook and payroll providers

## 5. Payroll Calculator Integration

### Usage in Payroll Calculation

**Location**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/payroll/payroll_calculator.go`

#### Fetching Expenses (Lines 39-136)

```go
func (h *handler) calculatePayrolls(users []*model.Employee, batchDate time.Time, simplifyNotes bool) {
    var expenses []bcModel.Todo

    if h.service.Basecamp != nil {
        // Basecamp flow: fetch ops and team expenses separately
        opsTodoLists, err := h.service.PayrollExpenseProvider.GetAllInList(opsExpenseID, opsID)
        // ... approval check via comments

        todolists, err := h.service.PayrollExpenseProvider.GetGroups(expenseID, woodlandID)
        for i := range todolists {
            e, err := h.service.PayrollExpenseProvider.GetAllInList(todolists[i].ID, woodlandID)
            // ... approval check via comments
        }
    } else {
        // NocoDB flow: fetch from expense_submissions table
        allExpenses, err := h.service.PayrollExpenseProvider.GetAllInList(opsExpenseID, opsID)
        expenses = append(expenses, allExpenses...)

        // Also fetch accounting todos
        accountingTodos, err := h.service.PayrollAccountingTodoProvider.GetAllInList(opsExpenseID, opsID)
        expenses = append(expenses, accountingTodos...)
    }

    // Process expenses in bonus calculation
    for i, u := range users {
        bonus, commission, reimbursementAmount, bonusExplains, commissionExplains :=
            h.getBonus(*users[i], batchDate, expenses, simplifyNotes)
    }
}
```

**Key Integration Points**:
1. **Provider Check**: `if h.service.Basecamp != nil` determines provider type
2. **Multiple Sources**: Ops expenses + team expenses (Basecamp) OR single table (NocoDB)
3. **Approval Logic**: Basecamp checks comments, NocoDB filters by status
4. **Dual Providers**: `PayrollExpenseProvider` + `PayrollAccountingTodoProvider` for NocoDB

#### Processing Expenses in Bonus Calculation (Lines 252-308)

```go
func (h *handler) getBonus(u model.Employee, batchDate time.Time, expenses []bcModel.Todo, simplifyNotes bool) {
    for i := range expenses {
        hasReimbursement := false
        // Check if user is assignee
        for j := range expenses[i].Assignees {
            if expenses[i].Assignees[j].ID == u.BasecampID {
                hasReimbursement = true
                break
            }
        }

        if hasReimbursement {
            // Parse expense title
            name, amount, err := h.getReimbursement(expenses[i].Title)

            bonus += amount
            reimbursementAmount += amount
            bonusExplain = append(bonusExplain, model.CommissionExplain{
                Amount:           amount,
                Month:            int(batchDate.Month()),
                Year:             batchDate.Year(),
                Name:             name,
                BasecampTodoID:   expenses[i].ID,
                BasecampBucketID: expenses[i].Bucket.ID,
            })
        }
    }
}
```

**Processing Logic**:
1. **Assignee Matching**: Compare `expenses[i].Assignees[j].ID` with `u.BasecampID`
2. **Title Parsing**: Extract name and amount via `getReimbursement()`
3. **Bonus Accumulation**: Add expense amount to employee's bonus
4. **Metadata Tracking**: Store `BasecampTodoID` and `BasecampBucketID` for reference

#### Parsing Expense Title (Lines 415-444)

```go
func (h *handler) getReimbursement(expense string) (string, model.VietnamDong, error) {
    splits := strings.Split(expense, "|")
    if len(splits) < 3 {
        return "", 0, nil
    }

    name := strings.TrimSpace(splits[0])       // Description
    amountStr := strings.TrimSpace(splits[1])  // Amount
    c := strings.TrimSpace(splits[2])          // Currency

    // Default to VND if currency is empty
    if c == "" {
        c = currency.VNDCurrency
    }

    // Parse amount from expense title
    bcAmount := h.extractExpenseAmount(amountStr)

    // Convert to VND if necessary
    if c != currency.VNDCurrency {
        tempAmount, _, err := h.service.Wise.Convert(float64(bcAmount), c, currency.VNDCurrency)
        amount = model.NewVietnamDong(int64(tempAmount))
    } else {
        amount = model.NewVietnamDong(int64(bcAmount))
    }

    return name, amount.Format(), nil
}
```

**Title Format Requirements**:
- **Format**: `"description | amount | currency"`
- **Parts**: Exactly 3 parts separated by `|`
- **Amount Parsing**: Provider-specific (Basecamp uses complex parsing, NocoDB uses simple float)
- **Currency Default**: VND if empty
- **Currency Conversion**: Uses Wise API for non-VND amounts

### Marking Expenses as Completed

**Location**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/payroll/commit.go` (Lines 772-790)

```go
func (h *handler) markExpenseSubmissionsAsCompleted(payrolls []*model.Payroll) {
    if h.service.PayrollExpenseProvider == nil {
        h.logger.Debug("PayrollExpenseProvider is nil, skipping NocoDB update")
        return
    }

    // Check if provider is NocoDB
    nocoService, ok := h.service.PayrollExpenseProvider.(*nocodb.ExpenseService)
    if !ok {
        h.logger.Debug("PayrollExpenseProvider is not NocoDB service (Basecamp flow), skipping")
        return
    }

    // Extract expense IDs from payroll bonus explanations
    expenseIDs := extractExpenseIDsFromPayrolls(payrolls)

    // Mark each expense as completed
    for _, expenseID := range expenseIDs {
        err := nocoService.MarkExpenseAsCompleted(expenseID)
        if err != nil {
            h.logger.Error(err, fmt.Sprintf("Failed to mark expense %d as completed", expenseID))
        }
    }
}
```

**Completion Flow**:
1. **Provider Check**: Type assertion to verify NocoDB provider
2. **ID Extraction**: Parse expense IDs from `bonusExplain.BasecampTodoID`
3. **Status Update**: Call provider-specific `MarkExpenseAsCompleted()`
4. **Error Handling**: Log errors but don't fail payroll commit

## 6. Service Initialization

### Service Configuration

**Location**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/service.go` (Lines 243-279)

```go
// Initialize expense provider
var expenseProvider taskprovider.ExpenseProvider  // For webhooks
var payrollExpenseProvider basecamp.ExpenseProvider  // For payroll

// Payroll expense provider initialization
if cfg.TaskProvider == "nocodb" {
    payrollExpenseProvider = nocodb.NewExpenseService(nocoSvc, cfg, store, repo, logger.L)
}
if payrollExpenseProvider == nil {
    // Fallback to Basecamp adapter
    payrollExpenseProvider = basecamp.NewExpenseAdapter(basecampSvc)
}

// Payroll accounting todo provider initialization
var payrollAccountingTodoProvider basecamp.ExpenseProvider
if cfg.TaskProvider == "nocodb" {
    payrollAccountingTodoProvider = nocodb.NewAccountingTodoService(nocoSvc, cfg, store, repo, logger.L)
}

return &Service{
    ExpenseProvider:              expenseProvider,  // Webhook provider
    PayrollExpenseProvider:       payrollExpenseProvider,  // Payroll expenses
    PayrollAccountingTodoProvider: payrollAccountingTodoProvider,  // Accounting todos
}
```

**Configuration Pattern**:
1. **Dual Providers**: Separate providers for webhooks vs payroll
2. **Provider Selection**: Based on `cfg.TaskProvider` ("nocodb" or default to Basecamp)
3. **Fallback Logic**: Always defaults to Basecamp if NocoDB not configured
4. **Multiple Sources**: Separate providers for expenses and accounting todos (NocoDB only)

### Configuration Structure

**Location**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/config/config.go`

```go
type ExpenseIntegration struct {
    Noco            ExpenseNocoIntegration
    Notion          ExpenseNotionIntegration
    ApproverMapping map[string]string
}

type ExpenseNocoIntegration struct {
    WorkspaceID   string  // NOCO_EXPENSE_WORKSPACE_ID
    TableID       string  // NOCO_EXPENSE_TABLE_ID
    WebhookSecret string  // NOCO_EXPENSE_WEBHOOK_SECRET
}

type ExpenseNotionIntegration struct {
    ExpenseDBID    string  // NOTION_EXPENSE_DB_ID
    ContractorDBID string  // NOTION_CONTRACTOR_DB_ID
}
```

**Configuration Loading** (Lines 398-399):
```go
notionExpenseDBID := v.GetString("NOTION_EXPENSE_DB_ID")
notionContractorDBID := v.GetString("NOTION_CONTRACTOR_DB_ID")
```

## 7. Key Patterns and Learnings

### Pattern 1: Interface-Based Provider Abstraction

**Benefits**:
- Swappable implementations (Basecamp, NocoDB, Notion)
- Consistent payroll calculator integration
- No changes to consuming code when switching providers

**Implementation Strategy**:
```go
type ExpenseProvider interface {
    GetAllInList(todolistID, projectID int) ([]model.Todo, error)
    GetGroups(todosetID, projectID int) ([]model.TodoGroup, error)
    GetLists(projectID, todosetID int) ([]model.TodoList, error)
}
```

### Pattern 2: Data Transformation Layer

**NocoDB Example**:
```
NocoDB Record → transformRecordToTodo() → bcModel.Todo
```

**Key Transformations**:
1. Record ID (int/string) → Todo.ID (int)
2. Email field → Employee lookup → BasecampID
3. Raw amount/currency → Formatted title `"desc | amount | currency"`
4. Task metadata → Bucket and Assignee structures

**Notion Adaptation**:
```
Notion Page → transformPageToTodo() → bcModel.Todo
```

**Required Transformations**:
1. Page ID (UUID) → Todo.ID (convert to int or use hash)
2. Rollup Email → Employee lookup → BasecampID
3. Number + Select properties → Formatted title
4. Status property → Completed flag mapping

### Pattern 3: Graceful Degradation and Fallbacks

**Config Fallback**:
```go
if payrollExpenseProvider == nil {
    payrollExpenseProvider = basecamp.NewExpenseAdapter(basecampSvc)
}
```

**Default Values**:
```go
if currency == "" {
    currency = "VND"
}
```

**Error Handling**:
```go
for _, record := range result.List {
    todo, err := transformRecordToTodo(record)
    if err != nil {
        logger.Error(err, "failed to transform record")
        continue  // Skip invalid record, don't fail entire batch
    }
}
```

### Pattern 4: Employee-BasecampID Linkage

**Critical Flow**:
```
Email (from provider) → Employee.TeamEmail → Employee.BasecampID → Todo.Assignees[0].ID
```

**Requirements**:
1. Email must match `employees.team_email` exactly
2. Employee must have `basecamp_id` set (non-zero)
3. BasecampID is used for assignee matching in payroll calculation

**Notion Adaptation**:
- Rollup Email property → Extract email string
- Same employee lookup logic
- Same BasecampID requirement

### Pattern 5: Title Format Contract

**Critical Format**: `"description | amount | currency"`

**Parsing Logic**:
```go
splits := strings.Split(title, "|")
description := strings.TrimSpace(splits[0])
amount := parseAmount(strings.TrimSpace(splits[1]))
currency := strings.TrimSpace(splits[2])
```

**Notion Requirements**:
- Combine Title, Amount, Currency properties
- Format: `fmt.Sprintf("%s | %.0f | %s", title, amount, currency)`
- Amount: Float without decimals (e.g., "5000000" not "5,000,000.00")
- Currency: Select option name (e.g., "VND", "USD")

### Pattern 6: ID Mapping Challenge

**NocoDB**: Integer IDs (easy mapping)
```go
ID: int(recordID)  // Direct conversion
```

**Notion**: UUID IDs (requires mapping strategy)

**Options**:
1. **Hash to Int**: Convert UUID to integer (may lose uniqueness)
2. **Store Mapping**: Maintain UUID→Int mapping table
3. **Use Hash**: Use stable hash function for deterministic conversion
4. **Encode in Metadata**: Store Notion page ID in new field

**Recommendation**: Use hash-based conversion for minimal code changes

```go
func notionPageIDToInt(pageID string) int {
    // Remove hyphens from UUID
    cleanID := strings.ReplaceAll(pageID, "-", "")
    // Take last 8 hex characters and convert to int
    hashStr := cleanID[len(cleanID)-8:]
    hash, _ := strconv.ParseInt(hashStr, 16, 64)
    return int(hash)
}
```

## 8. Notion-Specific Considerations

### Database Query Pattern

**NocoDB**:
```go
query.Set("where", "(status,eq,approved)")
```

**Notion**:
```go
filter := &notion.DatabaseQueryFilter{
    Property: "Status",
    DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
        Status: &notion.StatusDatabaseQueryFilter{
            Equals: "Approved",  // Status option name
        },
    },
}
```

### Relation + Rollup Access

**NocoDB**: Direct field access
```go
email := extractString(record, "requester_team_email")
```

**Notion**: Rollup property extraction
```go
emailProp := page.Properties.(notion.DatabasePageProperties)["Email"]
if emailProp.Type == notion.DBPropTypeRollup {
    rollup := emailProp.Rollup
    // Extract email from rollup.Array or rollup.String
}
```

### Property Extraction

**NocoDB**: Generic map extraction
```go
amount := extractFloat(record, "amount")
currency := extractString(record, "currency")
```

**Notion**: Typed property access
```go
amountProp := props["Amount"]
amount := amountProp.Number

currencyProp := props["Currency"]
currency := currencyProp.Select.Name
```

### Status Update Pattern

**NocoDB**:
```go
payload := map[string]interface{}{
    "Id":     expenseID,
    "status": "completed",
}
// PATCH /tables/{tableID}/records
```

**Notion**:
```go
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
// UpdatePage(ctx, pageID, updateParams)
```

## 9. Implementation Checklist for Notion

Based on existing patterns, the Notion implementation needs:

### Core Requirements

- [ ] Implement `ExpenseProvider` interface with three methods
- [ ] Create `NotionExpenseService` struct with client, config, store, repo, logger
- [ ] Implement query filter for `Status = "Approved"`
- [ ] Extract properties: Title, Amount, Currency, Email (rollup)
- [ ] Transform Notion page to `bcModel.Todo` format
- [ ] Build title in format: `"description | amount | currency"`
- [ ] Map employee email to BasecampID via database lookup
- [ ] Map Notion page ID (UUID) to integer ID for Todo.ID
- [ ] Implement pagination for large result sets

### Transformation Layer

- [ ] Create `transformPageToTodo()` function
- [ ] Extract title from Title property (rich text array)
- [ ] Extract amount from Number property (float64)
- [ ] Extract currency from Select property (option name)
- [ ] Extract email from Rollup property (array or string)
- [ ] Handle employee lookup and validation
- [ ] Build assignee structure with BasecampID
- [ ] Create bucket structure with metadata

### Status Update

- [ ] Implement `MarkExpenseAsCompleted()` method
- [ ] Update page Status property from "Approved" to "Paid"
- [ ] Use `notion.UpdatePage()` with `DBPropTypeStatus`
- [ ] Handle errors gracefully (log but don't fail payroll)

### Service Integration

- [ ] Add Notion expense service initialization in `service.go`
- [ ] Check `cfg.TaskProvider == "notion"` for provider selection
- [ ] Load `cfg.ExpenseIntegration.Notion.ExpenseDBID` configuration
- [ ] Fallback to Basecamp if Notion not configured

### Error Handling

- [ ] Validate Notion database ID configuration
- [ ] Handle missing/invalid page properties
- [ ] Log detailed errors for debugging
- [ ] Continue processing on individual record errors
- [ ] Handle Notion API rate limits and errors

### Testing

- [ ] Unit tests for property extraction
- [ ] Unit tests for page transformation
- [ ] Mock Notion client for testing
- [ ] Test pagination logic
- [ ] Test error handling scenarios
- [ ] Integration test with test database

## Summary

The existing NocoDB and Basecamp implementations provide a solid foundation for the Notion provider:

1. **Interface Compliance**: Must implement `ExpenseProvider` interface exactly
2. **Data Format**: Must produce `bcModel.Todo` with exact title format
3. **Employee Linking**: Must maintain email → BasecampID lookup pattern
4. **ID Mapping**: Must handle UUID → integer conversion strategy
5. **Status Updates**: Must update Notion status after payroll commit
6. **Error Handling**: Must gracefully handle errors without failing payroll
7. **Configuration**: Must integrate with existing config structure

The Notion implementation will follow the NocoDB pattern closely, with adaptations for:
- go-notion client usage vs HTTP API
- UUID page IDs vs integer record IDs
- Typed property access vs generic map extraction
- Rollup property extraction vs direct field access
- Status property type vs text field
