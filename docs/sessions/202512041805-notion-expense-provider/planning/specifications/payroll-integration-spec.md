# Payroll Integration Specification

## Overview

This document specifies how the Notion Expense Service integrates with the existing payroll calculation and commit workflow, including required changes to service initialization and commit handlers.

## Integration Points

### 1. Service Initialization

**Location**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/service.go`

#### Current Implementation Pattern

```go
// Existing provider initialization
var payrollExpenseProvider basecamp.ExpenseProvider

if cfg.TaskProvider == "nocodb" {
    payrollExpenseProvider = nocodb.NewExpenseService(nocoSvc, cfg, store, repo, logger.L)
}
if payrollExpenseProvider == nil {
    payrollExpenseProvider = basecamp.NewExpenseAdapter(basecampSvc)
}
```

#### Updated Implementation

```go
// Updated provider initialization with Notion support
var payrollExpenseProvider basecamp.ExpenseProvider

switch cfg.TaskProvider {
case "nocodb":
    // NocoDB provider
    payrollExpenseProvider = nocodb.NewExpenseService(nocoSvc, cfg, store, repo, logger.L)

case "notion":
    // NEW: Notion provider initialization
    if notionSvc == nil {
        logger.Fatal("Notion service not initialized but TASK_PROVIDER=notion")
    }
    payrollExpenseProvider = notion.NewExpenseService(notionSvc, cfg, store, repo, logger.L)
    logger.Info("Initialized Notion expense provider",
        "database_id", cfg.ExpenseIntegration.Notion.ExpenseDBID,
    )

default:
    // Fallback to Basecamp (empty string or "basecamp")
    if basecampSvc != nil {
        payrollExpenseProvider = basecamp.NewExpenseAdapter(basecampSvc)
        logger.Info("Initialized Basecamp expense provider (fallback)")
    }
}

// Validation
if payrollExpenseProvider == nil {
    logger.Fatal("No expense provider configured",
        "task_provider", cfg.TaskProvider,
    )
}
```

#### Service Structure Update

```go
// Service struct (no changes needed)
type Service struct {
    // Expense providers
    PayrollExpenseProvider       basecamp.ExpenseProvider  // Notion implements this
    PayrollAccountingTodoProvider basecamp.ExpenseProvider  // Not used with Notion

    // Other services...
    Notion *notion.Service  // Existing Notion service
}
```

### 2. Payroll Calculation

**Location**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/payroll/payroll_calculator.go`

#### Expense Fetching

**Existing Code** (Lines 39-136):

```go
func (h *handler) calculatePayrolls(users []*model.Employee, batchDate time.Time, simplifyNotes bool) {
    var expenses []bcModel.Todo

    if h.service.Basecamp != nil {
        // Basecamp-specific fetching logic
        opsTodoLists, err := h.service.PayrollExpenseProvider.GetAllInList(opsExpenseID, opsID)
        // ... fetch from multiple sources
    } else {
        // NocoDB flow
        allExpenses, err := h.service.PayrollExpenseProvider.GetAllInList(opsExpenseID, opsID)
        expenses = append(expenses, allExpenses...)

        accountingTodos, err := h.service.PayrollAccountingTodoProvider.GetAllInList(opsExpenseID, opsID)
        expenses = append(expenses, accountingTodos...)
    }

    // Process expenses
    for i, u := range users {
        bonus, commission, reimbursementAmount, bonusExplains, commissionExplains :=
            h.getBonus(*users[i], batchDate, expenses, simplifyNotes)
    }
}
```

#### No Changes Required

The payroll calculator already uses the `ExpenseProvider` interface, so Notion integration works transparently:

```go
// Notion provider will be called here
allExpenses, err := h.service.PayrollExpenseProvider.GetAllInList(opsExpenseID, opsID)
```

**Behavior with Notion**:
- `GetAllInList()` fetches approved expenses from Notion database
- Returns `[]bcModel.Todo` in standard format
- Title format: `"description | amount | currency"`
- Assignees contain employee BasecampID for matching

#### Expense Processing

**Existing Code** (Lines 252-308):

```go
func (h *handler) getBonus(u model.Employee, batchDate time.Time, expenses []bcModel.Todo, simplifyNotes bool) {
    for i := range expenses {
        hasReimbursement := false

        // Match expense to employee by BasecampID
        for j := range expenses[i].Assignees {
            if expenses[i].Assignees[j].ID == u.BasecampID {
                hasReimbursement = true
                break
            }
        }

        if hasReimbursement {
            // Parse expense title: "description | amount | currency"
            name, amount, err := h.getReimbursement(expenses[i].Title)

            // Add to bonus
            bonus += amount
            reimbursementAmount += amount

            // Store metadata
            bonusExplain = append(bonusExplain, model.CommissionExplain{
                Amount:           amount,
                Month:            int(batchDate.Month()),
                Year:             batchDate.Year(),
                Name:             name,
                BasecampTodoID:   expenses[i].ID,        // Hash-based int from UUID
                BasecampBucketID: expenses[i].Bucket.ID, // Same hash
            })
        }
    }

    return bonus, commission, reimbursementAmount, bonusExplain, commissionExplain
}
```

#### No Changes Required

Expense processing works with Notion provider because:
1. **Assignee Matching**: `expenses[i].Assignees[j].ID` matches `u.BasecampID` (populated from employee lookup)
2. **Title Parsing**: `getReimbursement()` parses `"description | amount | currency"` format (Notion provides this)
3. **ID Storage**: `BasecampTodoID` stores hash-based integer (Notion page UUID converted)

### 3. Payroll Commit and Status Update

**Location**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/payroll/commit.go`

#### Current Implementation (Lines 772-790)

```go
func (h *handler) markExpenseSubmissionsAsCompleted(payrolls []*model.Payroll) {
    if h.service.PayrollExpenseProvider == nil {
        h.logger.Debug("PayrollExpenseProvider is nil, skipping NocoDB update")
        return
    }

    // Type assertion for NocoDB provider
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

#### Updated Implementation

```go
func (h *handler) markExpenseSubmissionsAsCompleted(payrolls []*model.Payroll) {
    if h.service.PayrollExpenseProvider == nil {
        h.logger.Debug("PayrollExpenseProvider is nil, skipping expense status update")
        return
    }

    // Type assertion based on provider type
    switch provider := h.service.PayrollExpenseProvider.(type) {
    case *nocodb.ExpenseService:
        // NocoDB: Update using integer IDs
        h.logger.Debug("Marking NocoDB expenses as completed")
        expenseIDs := extractExpenseIDsFromPayrolls(payrolls)

        for _, expenseID := range expenseIDs {
            err := provider.MarkExpenseAsCompleted(expenseID)
            if err != nil {
                h.logger.Error(err, "Failed to mark NocoDB expense as completed",
                    "expense_id", expenseID,
                )
            }
        }

    case *notion.ExpenseService:
        // NEW: Notion - Update using page UUIDs
        h.logger.Debug("Marking Notion expenses as completed")
        pageIDs := extractNotionPageIDsFromPayrolls(payrolls)

        for _, pageID := range pageIDs {
            err := provider.MarkExpenseAsCompleted(pageID)
            if err != nil {
                h.logger.Error(err, "Failed to mark Notion expense as completed",
                    "page_id", pageID,
                )
            }
        }

    case *basecamp.ExpenseAdapter:
        // Basecamp: No status update needed (uses comments)
        h.logger.Debug("Basecamp provider does not support status updates, skipping")

    default:
        h.logger.Warn("Unknown expense provider type, skipping status update",
            "provider_type", fmt.Sprintf("%T", provider),
        )
    }
}
```

#### New Helper Function: extractNotionPageIDsFromPayrolls

```go
// extractNotionPageIDsFromPayrolls extracts Notion page UUIDs from payroll bonus explanations.
//
// The UUID must be stored in the CommissionExplain during payroll calculation.
// Options for storage location (to be determined during implementation):
//   1. Use existing text field (e.g., task_ref, notes)
//   2. Add new field in CommissionExplain model
//   3. Use JSON metadata field if available
//
// Parameters:
//   - payrolls: List of committed payrolls
//
// Returns:
//   - []string: List of unique Notion page UUIDs
func extractNotionPageIDsFromPayrolls(payrolls []*model.Payroll) []string {
    pageIDMap := make(map[string]bool)

    for _, payroll := range payrolls {
        for _, explain := range payroll.BonusExplain {
            // Extract UUID from stored location
            // TODO: Determine exact field during implementation
            // Option 1: Parse from explain.Name or explain.TaskRef
            // Option 2: Use dedicated field if added

            // Placeholder logic (to be implemented):
            pageID := extractPageIDFromExplain(explain)
            if pageID != "" {
                pageIDMap[pageID] = true
            }
        }
    }

    // Convert map to slice
    pageIDs := make([]string, 0, len(pageIDMap))
    for pageID := range pageIDMap {
        pageIDs = append(pageIDs, pageID)
    }

    return pageIDs
}
```

## Title Format Contract

### Format Specification

All expense providers must generate Todo titles in the exact format:

```
"<description> | <amount> | <currency>"
```

### Examples

```
"Office supplies | 5000000 | VND"
"Conference ticket | 500 | USD"
"Team lunch | 2000000 | VND"
```

### Format Requirements

1. **Separator**: Pipe character `|` with spaces on both sides (` | `)
2. **Description**: Any non-empty string (can contain spaces, no trailing spaces)
3. **Amount**: Numeric value formatted as integer (no decimals, no thousand separators)
   - For Notion: Use `fmt.Sprintf("%.0f", amount)`
4. **Currency**: Currency code (uppercase, e.g., "VND", "USD")

### Parsing Logic

**Location**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/payroll/payroll_calculator.go` (Lines 415-444)

```go
func (h *handler) getReimbursement(expense string) (string, model.VietnamDong, error) {
    splits := strings.Split(expense, "|")
    if len(splits) < 3 {
        return "", 0, fmt.Errorf("invalid expense format: expected 3 parts, got %d", len(splits))
    }

    name := strings.TrimSpace(splits[0])       // Description
    amountStr := strings.TrimSpace(splits[1])  // Amount
    c := strings.TrimSpace(splits[2])          // Currency

    // Default to VND if currency is empty
    if c == "" {
        c = currency.VNDCurrency
    }

    // Parse amount
    bcAmount := h.extractExpenseAmount(amountStr)

    // Convert to VND if necessary
    if c != currency.VNDCurrency {
        tempAmount, _, err := h.service.Wise.Convert(float64(bcAmount), c, currency.VNDCurrency)
        if err != nil {
            return "", 0, fmt.Errorf("currency conversion failed: %w", err)
        }
        amount = model.NewVietnamDong(int64(tempAmount))
    } else {
        amount = model.NewVietnamDong(int64(bcAmount))
    }

    return name, amount.Format(), nil
}
```

### Notion Implementation

**Location**: Notion ExpenseService

```go
func (s *ExpenseService) transformPageToTodo(ctx context.Context, page notion.Page) (*bcModel.Todo, error) {
    // Extract properties
    title := extractTitle(props)      // "Office supplies"
    amount := extractAmount(props)    // 5000000.0
    currency := extractCurrency(props) // "VND"

    // Build Todo title in required format
    todoTitle := fmt.Sprintf("%s | %.0f | %s", title, amount, currency)
    // Result: "Office supplies | 5000000 | VND"

    return &bcModel.Todo{
        Title: todoTitle,
        // ... other fields
    }, nil
}
```

## UUID Storage Strategy

### Challenge

After payroll commit, the system needs to mark expenses as completed in Notion using the original page UUID. However, `CommissionExplain` currently stores only the integer ID (`BasecampTodoID`).

### Solution Options

#### Option 1: Use Existing Text Field (Recommended for Phase 1)

Store UUID in an existing text field if available (e.g., `task_ref`, `notes`).

```go
bonusExplain := model.CommissionExplain{
    // ... existing fields
    BasecampTodoID:   notionPageIDToInt(page.ID),  // Hash-based int
    // Store UUID in existing text field
    TaskRef: page.ID,  // "2bfb69f8-f573-81cb-a2da-f06d28896390"
}
```

**Pros**:
- No schema migration required
- Quick implementation
- Works immediately

**Cons**:
- Field repurposing (may cause confusion)
- Need to validate field exists and is available

#### Option 2: Add Dedicated Field (Recommended for Phase 2)

Add a new field to `CommissionExplain` model specifically for Notion page ID.

**Migration**:

```sql
-- Add Notion page ID field
ALTER TABLE commission_explains
ADD COLUMN notion_page_id UUID;

CREATE INDEX idx_commission_explains_notion_page_id
ON commission_explains(notion_page_id);
```

**Model Update**:

```go
type CommissionExplain struct {
    // Existing fields
    BasecampTodoID   int
    BasecampBucketID int

    // New field
    NotionPageID     *string  // Nullable, only set for Notion expenses
}
```

**Pros**:
- Clean, explicit design
- No field repurposing
- Supports mixed providers (if needed in future)

**Cons**:
- Requires database migration
- More complex initial implementation

#### Option 3: JSON Metadata Field

Use a JSON metadata field to store provider-specific data.

```go
type CommissionExplain struct {
    // Existing fields
    BasecampTodoID   int
    BasecampBucketID int

    // Metadata field (JSONB)
    Metadata         map[string]interface{}
}

// Usage
metadata := map[string]interface{}{
    "notion_page_id": page.ID,
    "task_provider":  "notion",
}
```

**Pros**:
- Flexible for future provider-specific data
- Single field for all metadata

**Cons**:
- More complex querying
- Less type-safe
- Requires JSON serialization/deserialization

### Recommended Approach

**Phase 1**: Use Option 1 (existing field) for quick implementation and validation.

**Phase 2**: Migrate to Option 2 (dedicated field) for production quality and long-term maintainability.

## Testing Strategy

### Unit Tests

#### Service Initialization Tests

**Location**: `pkg/service/service_test.go`

```go
func TestServiceInitialization_NotionProvider(t *testing.T) {
    cfg := &config.Config{
        TaskProvider: "notion",
        ExpenseIntegration: config.ExpenseIntegration{
            Notion: config.ExpenseNotionIntegration{
                ExpenseDBID: "2bfb69f8-f573-81cb-a2da-f06d28896390",
            },
        },
    }

    svc := NewService(cfg, store, repo, logger)

    require.NotNil(t, svc.PayrollExpenseProvider)
    assert.IsType(t, &notion.ExpenseService{}, svc.PayrollExpenseProvider)
}
```

#### Payroll Calculation Tests

**Location**: `pkg/handler/payroll/payroll_calculator_test.go`

```go
func TestCalculatePayrolls_WithNotionExpenses(t *testing.T) {
    // Mock Notion provider
    mockProvider := &MockExpenseProvider{
        GetAllInListFunc: func(listID, projectID int) ([]bcModel.Todo, error) {
            return []bcModel.Todo{
                {
                    ID:    680141712,  // Hash of Notion UUID
                    Title: "Office supplies | 5000000 | VND",
                    Assignees: []bcModel.Assignee{
                        {ID: 123456, Name: "John Doe"},
                    },
                    Completed: true,
                },
            }, nil
        },
    }

    handler := &handler{
        service: &Service{
            PayrollExpenseProvider: mockProvider,
        },
    }

    // Test calculation
    payrolls := handler.calculatePayrolls(employees, batchDate, false)

    // Verify expense included in payroll
    assert.Equal(t, 1, len(payrolls[0].BonusExplain))
    assert.Equal(t, "Office supplies", payrolls[0].BonusExplain[0].Name)
}
```

#### Status Update Tests

**Location**: `pkg/handler/payroll/commit_test.go`

```go
func TestMarkExpenseSubmissionsAsCompleted_Notion(t *testing.T) {
    mockProvider := &MockNotionExpenseService{
        MarkExpenseAsCompletedFunc: func(pageID string) error {
            assert.Equal(t, "2bfb69f8-f573-81cb-a2da-f06d28896390", pageID)
            return nil
        },
    }

    handler := &handler{
        service: &Service{
            PayrollExpenseProvider: mockProvider,
        },
    }

    payrolls := []*model.Payroll{
        {
            BonusExplain: []model.CommissionExplain{
                {
                    TaskRef: "2bfb69f8-f573-81cb-a2da-f06d28896390",  // Notion UUID
                },
            },
        },
    }

    handler.markExpenseSubmissionsAsCompleted(payrolls)

    // Verify MarkExpenseAsCompleted was called
}
```

### Integration Tests

#### End-to-End Payroll Flow

**Location**: `pkg/handler/payroll/payroll_integration_test.go`

```go
func TestPayrollFlow_NotionProvider(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup test Notion database
    cfg := loadTestConfig()
    cfg.TaskProvider = "notion"

    svc := setupTestService(cfg)
    handler := setupTestHandler(svc)

    // Create test expenses in Notion (Status = "Approved")
    // ...

    // Calculate payroll
    payrolls := handler.calculatePayrolls(testEmployees, time.Now(), false)

    // Verify expenses included
    assert.Greater(t, len(payrolls[0].BonusExplain), 0)

    // Commit payroll
    err := handler.commitPayroll(payrolls)
    require.NoError(t, err)

    // Verify Notion expenses marked as "Paid"
    // ...
}
```

## Rollback Plan

### Quick Rollback (Configuration Change)

```bash
# Revert to NocoDB
TASK_PROVIDER=nocodb  # Changed from "notion"
```

**Impact**: Immediate (next server restart)

**Data**: No data loss, payrolls continue with NocoDB expenses

### Full Rollback (Database Cleanup)

If Notion-based payrolls were committed:

1. **Identify Notion Expenses**: Query `commission_explains` where `TaskRef` contains UUID format
2. **Validate Payrolls**: Ensure payroll calculations are correct
3. **Revert Status Updates**: Manually mark Notion expenses as "Approved" again if needed
4. **Switch Provider**: Change `TASK_PROVIDER` back to `nocodb`

## Monitoring and Observability

### Metrics to Track

```go
// Provider usage
expense_provider_type{provider="notion"} = 1

// Fetch operations
expense_fetch_count{provider="notion"} = 42
expense_fetch_duration_seconds{provider="notion"} = 2.5
expense_fetch_errors{provider="notion"} = 0

// Transformation
expense_transformation_success{provider="notion"} = 40
expense_transformation_errors{provider="notion"} = 2

// Status updates
expense_status_update_count{provider="notion"} = 40
expense_status_update_errors{provider="notion"} = 0
```

### Log Events

```
INFO: Fetching expenses from Notion: database_id=2bfb69f8...
INFO: Transformed Notion expenses: count=42, errors=0
INFO: Marking Notion expenses as completed: count=40
ERROR: Failed to mark Notion expense as completed: page_id=2bfb69f8..., error=...
```

### Alerts

1. **High Transformation Error Rate**: > 10% of expenses fail transformation
2. **Status Update Failures**: Any failure to mark expense as completed
3. **Provider Initialization Failure**: Service fails to start with Notion provider
4. **Rollup Fallback Rate**: > 20% of emails extracted via fallback (indicates rollup misconfiguration)

## Success Criteria

- [ ] Service initializes with `TASK_PROVIDER=notion`
- [ ] Approved expenses fetched from Notion database
- [ ] Expenses transformed to correct Todo format
- [ ] Employee matching works via email → BasecampID lookup
- [ ] Title format parsed correctly by payroll calculator
- [ ] Expenses included in payroll bonus calculations
- [ ] Payroll commit succeeds with Notion expenses
- [ ] Expense status updated from "Approved" to "Paid" after commit
- [ ] No changes required to existing Basecamp/NocoDB flows
- [ ] Can switch providers via configuration change
- [ ] Rollback to NocoDB works immediately

## References

- **Service Initialization**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/service.go`
- **Payroll Calculator**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/payroll/payroll_calculator.go`
- **Commit Handler**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/payroll/commit.go`
- **ExpenseProvider Interface**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/basecamp/basecamp.go`
- **NocoDB Reference**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/nocodb/expense.go`
- **ADR-001**: UUID to Int Mapping
- **ADR-003**: Provider Selection
- **Notion Service Spec**: `notion-expense-service-spec.md`
