# Research Documentation: Notion Expense Provider

## Overview

This directory contains comprehensive research findings for implementing the Notion Expense Provider to replace NocoDB as the task provider for managing Expense Requests during payroll calculation in the fortress-api Go application.

## Research Status

**Phase**: COMPLETED ✅
**Date**: 2025-12-04
**Scope**: Expense Request implementation (Leave Request out of scope)

## Documentation Structure

### 1. [notion-api-patterns.md](./notion-api-patterns.md)
**Notion API Best Practices and Patterns**

Research on the `go-notion` library and Notion API:
- Status property filtering and updates
- Database query patterns with filters
- Relation and rollup property handling
- Pagination implementation (cursor-based)
- Error handling and rate limiting
- Property type mapping reference
- Testing strategies

**Key Sections**:
- Querying databases with status filters
- Updating page properties (status updates)
- Handling relation and rollup properties
- Pagination best practices
- Property type mapping table
- Error handling patterns

### 2. [existing-implementation-analysis.md](./existing-implementation-analysis.md)
**Analysis of Existing Codebase Patterns**

Deep dive into NocoDB, Basecamp, and payroll calculator implementations:
- ExpenseProvider interface requirements
- NocoDB ExpenseService implementation
- NocoDB AccountingTodoService patterns
- Basecamp ExpenseAdapter approach
- Payroll calculator integration
- Service initialization and configuration
- Key patterns and learnings

**Key Sections**:
- ExpenseProvider interface definition
- NocoDB implementation (query, transform, update)
- Payroll calculator integration points
- Service initialization patterns
- Employee-BasecampID linkage flow
- Title format contract (`"desc | amt | cur"`)
- ID mapping challenge (integer vs UUID)

### 3. [technical-considerations.md](./technical-considerations.md)
**Technical Challenges and Solutions**

Detailed technical considerations and mapping strategies:
- ID mapping (UUID → int conversion)
- Property type extraction methods
- Relation and rollup handling
- Status property management
- Pagination strategy
- Error handling patterns
- Performance optimization
- Testing approaches

**Key Sections**:
- ID mapping strategy (hash-based conversion)
- Property type mapping (9 types covered)
- Rollup extraction with fallback
- Status filter and update syntax
- Performance optimizations (batching, caching)
- Graceful error handling
- Unit and integration testing

### 4. [STATUS.md](./STATUS.md)
**Research Phase Status Report**

Executive summary of research completion:
- Research objectives and completion status
- Key findings and recommendations
- Implementation readiness checklist
- Next steps for planning phase
- References and resources

## Quick Reference

### Critical Findings

#### 1. ID Mapping Solution
```go
// Convert Notion UUID to int using last 8 hex characters
func notionPageIDToInt(pageID string) int {
    cleanID := strings.ReplaceAll(pageID, "-", "")
    hashStr := cleanID[len(cleanID)-8:]
    hash, _ := strconv.ParseInt(hashStr, 16, 64)
    return int(hash)
}
```

#### 2. Status Filter Pattern
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

#### 3. Email Extraction from Rollup
```go
rollup := emailProp.Rollup
if rollup.Type == notion.RollupTypeArray && len(rollup.Array) > 0 {
    if propVal, ok := rollup.Array[0].(notion.DatabasePageProperty); ok {
        email := propVal.Email
    }
}
```

#### 4. Title Format Contract
```go
// Critical format for payroll parsing
todoTitle := fmt.Sprintf("%s | %.0f | %s", title, amount, currency)
// Example: "Office supplies | 5000000 | VND"
```

## Key Recommendations

### 1. Implementation Strategy
- **Follow NocoDB pattern**: Similar service structure and transformation flow
- **Use go-notion library**: Leverage existing Notion integration
- **Hash-based ID mapping**: Convert UUID to int deterministically
- **Rollup with fallback**: Use rollup for email, fallback to direct query

### 2. Critical Requirements
- **Interface compliance**: Must implement `ExpenseProvider` exactly
- **Title format**: Must produce `"description | amount | currency"` format
- **Employee linking**: Must maintain Email → BasecampID lookup
- **Status updates**: Must update from "Approved" to "Paid" after payroll

### 3. Configuration Needs
```bash
TASK_PROVIDER=notion
NOTION_EXPENSE_DB_ID=2bfb69f8-f573-81cb-a2da-f06d28896390
NOTION_CONTRACTOR_DB_ID=2bfb69f8-f573-805a-8915-000bc44ce188
NOTION_SECRET=secret_xxxxxxxxxxxxxxxxxxxx  # Already exists
```

### 4. Performance Priorities
- Use rollups to avoid N+1 queries
- Batch employee lookups with caching
- Implement concurrent status updates
- Handle pagination for large datasets

## Comparison Table

| Aspect | NocoDB | Notion | Solution |
|--------|--------|--------|----------|
| Record ID | Integer | UUID string | Hash last 8 hex chars to int |
| Status Field | Text ("approved") | Status type ("Approved") | Use StatusDatabaseQueryFilter |
| Email Access | Direct field | Rollup property | Extract from rollup array |
| Title Format | Pre-formatted | Combine 3 properties | fmt.Sprintf("%s \| %.0f \| %s") |
| API Client | HTTP REST | go-notion library | Use notion.Client methods |
| Query Syntax | URL query params | DatabaseQueryFilter struct | Typed filter objects |
| Update Method | PATCH /tables/{id}/records | UpdatePage(uuid, params) | Store UUID in metadata |
| Pagination | Offset-based | Cursor-based | Implement HasMore/NextCursor |

## Implementation Checklist

Based on research findings, the Notion implementation requires:

### Core Implementation
- [ ] Create `NotionExpenseService` struct with client, config, store, repo, logger
- [ ] Implement `ExpenseProvider` interface (3 methods)
- [ ] Implement query filter for `Status = "Approved"`
- [ ] Extract properties: Title, Amount, Currency, Email (rollup)
- [ ] Transform Notion page to `bcModel.Todo` format
- [ ] Build title format: `"description | amount | currency"`
- [ ] Map employee email to BasecampID via database lookup
- [ ] Map Notion page UUID to integer for Todo.ID
- [ ] Implement pagination for large result sets

### Status Management
- [ ] Implement `MarkExpenseAsCompleted()` method
- [ ] Update Status property from "Approved" to "Paid"
- [ ] Store UUID in CommissionExplain for reverse lookup
- [ ] Handle errors gracefully (log but don't fail payroll)

### Service Integration
- [ ] Add Notion expense service initialization in `service.go`
- [ ] Check `cfg.TaskProvider == "notion"` for provider selection
- [ ] Load configuration from environment variables
- [ ] Fallback to Basecamp if Notion not configured

### Testing
- [ ] Unit tests for property extraction
- [ ] Unit tests for page transformation
- [ ] Mock Notion client for testing
- [ ] Test pagination logic
- [ ] Test error handling scenarios
- [ ] Integration test with test database

## Migration Path

### Phase 1: Research (COMPLETED ✅)
- [x] Research Notion API and go-notion library
- [x] Analyze existing NocoDB and Basecamp implementations
- [x] Document technical considerations
- [x] Identify critical patterns and requirements

### Phase 2: Planning (Next)
- [ ] Write Architecture Decision Records (ADRs)
- [ ] Create technical specifications
- [ ] Define implementation tasks
- [ ] Plan testing strategy

### Phase 3: Implementation
- [ ] Implement NotionExpenseService
- [ ] Update service initialization
- [ ] Add configuration support
- [ ] Write unit tests
- [ ] Integration testing

### Phase 4: Deployment
- [ ] Code review and approval
- [ ] Deploy to test environment
- [ ] Validate with test data
- [ ] Deploy to production
- [ ] Monitor and verify

## Resources

### External Documentation
- [Notion API Reference](https://developers.notion.com/)
- [go-notion GitHub](https://github.com/dstotijn/go-notion)
- [go-notion Package Docs](https://pkg.go.dev/github.com/dstotijn/go-notion)
- [Notion API Filter Reference](https://developers.notion.com/reference/post-database-query-filter)
- [Notion API Update Page](https://developers.notion.com/reference/patch-page)

### Internal Documentation
- [Requirements Document](../requirements/requirements.md)
- [Notion Task Provider Spec](../../../specs/notion-task-provider.md)
- [Project Guidelines](../../../../CLAUDE.md)

### Code References
- `pkg/service/basecamp/basecamp.go` - ExpenseProvider interface
- `pkg/service/nocodb/expense.go` - NocoDB implementation pattern
- `pkg/handler/payroll/payroll_calculator.go` - Payroll integration
- `pkg/service/notion/notion.go` - Existing Notion service

## Questions & Answers

### Q: Why use hash-based ID conversion instead of database mapping?
**A**: Deterministic conversion with minimal code changes. No database migrations required. Low collision probability with 32-bit hash space.

### Q: Why use rollup instead of querying the relation directly?
**A**: Rollups are pre-computed by Notion and returned in the same query (1 API call vs N+1). More efficient for bulk operations.

### Q: How to handle UUID storage for reverse lookups?
**A**: Store original UUID in `CommissionExplain` metadata when creating bonus entries. This enables status updates after payroll commit.

### Q: What if an employee doesn't have a BasecampID?
**A**: Log error and skip that expense (same as NocoDB behavior). Payroll continues for other employees.

### Q: How to handle Notion API rate limits?
**A**: Implement exponential backoff with max 3 retries. Rate limit is 3 requests/second per integration.

### Q: Can we change status option names in Notion?
**A**: Yes, but must update code to match new names. Status options are configured in Notion UI and referenced by name in API.

## Contact & Support

For questions about this research or implementation:
- Review research documents in this directory
- Check [existing-implementation-analysis.md](./existing-implementation-analysis.md) for code patterns
- Refer to [technical-considerations.md](./technical-considerations.md) for solutions
- Consult [STATUS.md](./STATUS.md) for executive summary

---

**Research Phase**: COMPLETED ✅
**Ready for Planning**: Yes
**Blockers**: None
**Next Step**: Create ADRs and specifications in planning phase
