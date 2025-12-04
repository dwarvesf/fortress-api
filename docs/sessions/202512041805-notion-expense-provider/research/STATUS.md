# Research Phase Status

## Phase: COMPLETED ✅

**Completion Date**: 2025-12-04
**Duration**: Research session
**Status**: All research objectives completed

---

## Research Objectives

### ✅ Objective 1: Notion API Best Practices
**Status**: Completed

Researched the `go-notion` library and Notion API patterns:
- Status property filtering using `StatusDatabaseQueryFilter`
- Page property updates with `UpdatePage` API
- Relation and rollup property handling
- Pagination with cursor-based approach
- Error handling and rate limiting strategies
- Property type mapping (title, number, select, status, rollup)

**Deliverable**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/research/notion-api-patterns.md`

### ✅ Objective 2: Existing Codebase Patterns
**Status**: Completed

Analyzed existing implementations:
- **NocoDB ExpenseService**: HTTP REST API, status filtering, record transformation
- **NocoDB AccountingTodoService**: Title parsing, assignee filtering, employee lookup
- **Basecamp ExpenseAdapter**: Native API, comment-based approval
- **Payroll Calculator Integration**: Expense fetching, bonus calculation, title parsing
- **Service Initialization**: Provider selection, fallback logic, dual providers

**Deliverable**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/research/existing-implementation-analysis.md`

### ✅ Objective 3: Technical Considerations
**Status**: Completed

Documented technical challenges and solutions:
- **ID Mapping**: UUID→int conversion using hash-based strategy
- **Property Extraction**: Typed property access for title, amount, currency, email
- **Rollup Handling**: Email extraction from rollup with fallback to direct query
- **Status Management**: Filter by "Approved", update to "Paid"
- **Performance**: Batch lookups, caching, concurrent updates
- **Error Handling**: Graceful degradation, retry with backoff

**Deliverable**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/research/technical-considerations.md`

---

## Key Findings

### 1. Notion API Patterns

#### Status Property Filter
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

#### Status Property Update
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
```

#### Rollup Email Extraction
```go
rollup := emailProp.Rollup
if rollup.Type == notion.RollupTypeArray && len(rollup.Array) > 0 {
    if propVal, ok := rollup.Array[0].(notion.DatabasePageProperty); ok {
        email := propVal.Email
    }
}
```

### 2. Critical Mapping Requirements

| Aspect | NocoDB | Notion | Solution |
|--------|--------|--------|----------|
| Record ID | Integer | UUID string | Hash last 8 hex chars to int |
| Status Field | Text ("approved") | Status type ("Approved") | Use StatusDatabaseQueryFilter |
| Email Access | Direct field | Rollup property | Extract from rollup array |
| Title Format | `"desc \| amt \| cur"` | Combine 3 properties | fmt.Sprintf("%s \| %.0f \| %s") |
| API Pattern | HTTP REST | go-notion client | Use notion.Client methods |
| Update Method | PATCH with Id | UpdatePage with UUID | Store UUID in metadata |

### 3. Implementation Patterns from Existing Code

#### ExpenseProvider Interface (Must Implement)
```go
type ExpenseProvider interface {
    GetAllInList(todolistID, projectID int) ([]model.Todo, error)
    GetGroups(todosetID, projectID int) ([]model.TodoGroup, error)
    GetLists(projectID, todosetID int) ([]model.TodoList, error)
}
```

#### Service Structure (Follow NocoDB Pattern)
```go
type NotionExpenseService struct {
    client *notion.Client
    cfg    *config.Config
    store  *store.Store
    repo   store.DBRepo
    logger logger.Logger
}
```

#### Transformation Flow (Critical for Payroll)
```
Notion Page → transformPageToTodo() → bcModel.Todo
- Extract: Title (rich text), Amount (number), Currency (select), Email (rollup)
- Lookup: Email → Employee → BasecampID
- Format: "title | amount | currency"
- Create: Todo with ID, Title, Assignees, Bucket, Completed
```

#### Mark as Completed (After Payroll Commit)
```go
func (s *NotionExpenseService) MarkExpenseAsCompleted(pageID string) error {
    // Update Status: "Approved" → "Paid"
    // Store pageID in CommissionExplain metadata for reverse lookup
}
```

---

## Deliverables

### Research Documentation

1. **[notion-api-patterns.md](./notion-api-patterns.md)**
   - Notion API reference and best practices
   - go-notion library usage patterns
   - Status property filter and update syntax
   - Relation and rollup handling
   - Pagination and error handling
   - Property type mapping reference

2. **[existing-implementation-analysis.md](./existing-implementation-analysis.md)**
   - ExpenseProvider interface analysis
   - NocoDB ExpenseService implementation review
   - NocoDB AccountingTodoService patterns
   - Basecamp ExpenseAdapter approach
   - Payroll calculator integration points
   - Service initialization and configuration
   - Key patterns and learnings

3. **[technical-considerations.md](./technical-considerations.md)**
   - ID mapping strategies (UUID → int)
   - Property type extraction methods
   - Relation and rollup handling techniques
   - Status property filtering and updating
   - Pagination implementation
   - Error handling patterns
   - Performance optimization strategies
   - Testing approaches

---

## Key Recommendations

### 1. ID Mapping Strategy
**Use hash-based conversion with metadata storage**:
- Convert UUID to int using last 8 hex characters (deterministic)
- Store original UUID in CommissionExplain for reverse lookup
- No database schema changes required

### 2. Email Extraction Strategy
**Use rollup with direct query fallback**:
- Primary: Extract from Email rollup property (efficient, single query)
- Fallback: Query Requestor relation directly (if rollup fails)
- Validates against existing employee records

### 3. Property Extraction Pattern
**Follow typed property access**:
```go
titleProp := props["Title"]
if titleProp.Type == notion.DBPropTypeTitle {
    title := extractPlainText(titleProp.Title)
}
```

### 4. Title Format Contract
**Critical format for payroll parsing**:
```go
todoTitle := fmt.Sprintf("%s | %.0f | %s", title, amount, currency)
// Example: "Office supplies | 5000000 | VND"
```

### 5. Service Initialization
**Follow NocoDB provider pattern**:
```go
if cfg.TaskProvider == "notion" {
    payrollExpenseProvider = notion.NewExpenseService(notionClient, cfg, store, repo, logger)
}
```

### 6. Error Handling
**Graceful degradation required**:
- Log errors but continue processing
- Skip invalid records (don't fail entire batch)
- Default values for optional fields (currency defaults to "VND")

---

## Implementation Readiness Checklist

### Prerequisites
- [x] Notion API research completed
- [x] Existing codebase patterns analyzed
- [x] Technical considerations documented
- [x] ID mapping strategy defined
- [x] Property extraction methods researched
- [x] Status handling patterns identified

### Ready for Implementation
- [x] ExpenseProvider interface requirements understood
- [x] Data transformation logic documented
- [x] Employee lookup pattern identified
- [x] Title format requirements clarified
- [x] Status update method researched
- [x] Error handling strategy defined

### Configuration Requirements
- [x] NOTION_EXPENSE_DB_ID - Database ID for Expense Requests
- [x] NOTION_CONTRACTOR_DB_ID - Database ID for Contractors (optional)
- [x] NOTION_SECRET - API authentication token (already exists)
- [x] TASK_PROVIDER=notion - Provider selection flag

---

## Next Steps

### Phase 2: Planning (Ready to Start)

The research phase is complete. The next phase should focus on:

1. **Architecture Decision Records (ADRs)**
   - Document ID mapping decision (UUID → int hash)
   - Document email extraction strategy (rollup + fallback)
   - Document provider selection logic (config-based)

2. **Technical Specifications**
   - Define NotionExpenseService struct and methods
   - Specify property extraction functions
   - Define error handling flows
   - Specify configuration structure

3. **Implementation Plan**
   - Create task breakdown for implementation
   - Define file structure and locations
   - Identify test requirements
   - Plan service initialization changes

### Handoff to Planning Phase

**Research Complete**: All research objectives achieved ✅

**Key Artifacts**:
- Notion API patterns documented with examples
- Existing implementations analyzed with code references
- Technical considerations detailed with solutions
- Implementation checklist prepared

**Blockers**: None

**Risks Identified**:
1. UUID→int collision risk (mitigated with 32-bit hash space)
2. Rollup property configuration dependency (mitigated with fallback query)
3. Notion API rate limits (mitigated with backoff strategy)

**Ready for Planning**: Yes ✅

---

## Research Summary

### What We Learned

1. **Notion API**: Status filters use option names, property updates require correct types
2. **go-notion Library**: Provides typed property access, cursor-based pagination
3. **Existing Patterns**: NocoDB and Basecamp implementations follow ExpenseProvider interface
4. **Critical Flow**: Email → Employee → BasecampID is essential for payroll matching
5. **Title Format**: `"description | amount | currency"` is strictly required

### What We Validated

1. **Feasibility**: Notion can serve as expense provider (same interface as NocoDB/Basecamp)
2. **ID Mapping**: Hash-based conversion is viable for UUID→int requirement
3. **Email Access**: Rollup properties provide efficient access to related data
4. **Status Updates**: Notion API supports updating status from "Approved" to "Paid"
5. **Integration**: Minimal changes needed to payroll calculator (provider swap)

### What We Decided

1. **Use hash-based UUID→int conversion** (deterministic, no DB changes)
2. **Extract email from rollup property** (efficient, follows Notion best practices)
3. **Follow NocoDB service structure** (consistent with existing code)
4. **Implement graceful error handling** (skip invalid records, log errors)
5. **Support concurrent status updates** (optimize payroll commit performance)

---

## References

### External Resources
- [Notion API Reference](https://developers.notion.com/)
- [go-notion GitHub Repository](https://github.com/dstotijn/go-notion)
- [go-notion Package Documentation](https://pkg.go.dev/github.com/dstotijn/go-notion)
- [Notion API Filter Documentation](https://developers.notion.com/reference/post-database-query-filter)
- [Notion API Update Page](https://developers.notion.com/reference/patch-page)

### Internal Resources
- [Requirements Document](../requirements/requirements.md)
- [Notion Task Provider Spec](../../../specs/notion-task-provider.md)
- [CLAUDE.md](../../../../CLAUDE.md) - Project development guidelines

### Code References
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/basecamp/basecamp.go` - ExpenseProvider interface
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/nocodb/expense.go` - NocoDB implementation
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/payroll/payroll_calculator.go` - Payroll integration
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/notion.go` - Existing Notion service

---

**Research Phase**: COMPLETED ✅
**Next Phase**: Planning (ADRs, Specifications, Implementation Plan)
**Ready to Proceed**: Yes
