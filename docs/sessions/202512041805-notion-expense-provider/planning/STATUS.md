# Planning Phase Status

## Phase: COMPLETED ✅

**Completion Date**: 2025-12-04
**Duration**: Planning session
**Status**: All planning objectives completed

---

## Planning Objectives

### ✅ Objective 1: Architecture Decision Records (ADRs)
**Status**: Completed

Created three comprehensive ADRs documenting critical architectural decisions:

1. **ADR-001: UUID to Integer ID Mapping Strategy**
   - Decision: Hash-based conversion using last 8 hex characters
   - Rationale: Minimal code changes, deterministic mapping, backward compatible
   - Trade-offs: Collision risk (mitigated), UUID storage required for reverse lookup

2. **ADR-002: Email Extraction from Notion Rollup Property**
   - Decision: Two-tier strategy (rollup first, relation query fallback)
   - Rationale: Performance optimization with resilience
   - Trade-offs: Added complexity vs single-path approaches

3. **ADR-003: Task Provider Selection and Configuration**
   - Decision: Use `TASK_PROVIDER=notion` environment variable
   - Rationale: Clean separation, explicit configuration, easy rollback
   - Trade-offs: No gradual migration, single configuration point

**Deliverables**:
- `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/ADRs/ADR-001-uuid-to-int-mapping.md`
- `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/ADRs/ADR-002-email-extraction-strategy.md`
- `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/ADRs/ADR-003-provider-selection.md`

### ✅ Objective 2: Technical Specifications
**Status**: Completed

Created two detailed technical specifications:

1. **Notion Expense Service Specification**
   - Service structure and constructor
   - Interface implementation (GetAllInList, GetGroups, GetLists)
   - Property extraction helpers (title, amount, currency, email)
   - ID conversion method (notionPageIDToInt)
   - Status update method (MarkExpenseAsCompleted)
   - Error handling patterns
   - Configuration requirements
   - Testing strategy

2. **Payroll Integration Specification**
   - Service initialization updates
   - Provider selection logic
   - Payroll calculation integration points
   - Commit handler updates with type assertions
   - Title format contract: `"description | amount | currency"`
   - UUID storage strategy
   - Testing requirements
   - Rollback plan

**Deliverables**:
- `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/notion-expense-service-spec.md`
- `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/payroll-integration-spec.md`

### ✅ Objective 3: Planning Summary
**Status**: Completed

**Deliverable**: This STATUS.md document

---

## Key Decisions Summary

### 1. UUID to Integer Mapping (ADR-001)

**Problem**: Notion uses UUID page IDs, but `bcModel.Todo.ID` expects integers.

**Solution**: Hash-based conversion using last 8 hex characters of UUID.

```go
func notionPageIDToInt(pageID string) int {
    cleanID := strings.ReplaceAll(pageID, "-", "")
    hashStr := cleanID[len(cleanID)-8:]
    hash, _ := strconv.ParseInt(hashStr, 16, 64)
    return int(hash)
}
```

**Trade-offs**:
- Pros: Deterministic, no DB changes, simple implementation
- Cons: Collision risk (low probability), requires UUID storage for reverse lookup

**Implementation**: Store original UUID in metadata for status updates after payroll commit.

### 2. Email Extraction Strategy (ADR-002)

**Problem**: Notion stores email in rollup property from Requestor relation.

**Solution**: Two-tier extraction (rollup first, fallback to direct query).

```go
func (s *ExpenseService) getEmail(ctx, props) (string, error) {
    // Try rollup first (efficient)
    email, err := s.extractEmailFromRollup(props)
    if err == nil && email != "" {
        return email, nil
    }

    // Fallback to relation query
    return s.extractEmailFromRelation(ctx, props)
}
```

**Trade-offs**:
- Pros: Single API call in normal case, resilient to misconfiguration
- Cons: Two code paths, depends on rollup configuration

**Implementation**: Primary path uses rollup array extraction, fallback queries Contractor page directly.

### 3. Provider Selection Strategy (ADR-003)

**Problem**: Support multiple task providers (Basecamp, NocoDB, Notion).

**Solution**: Config-based provider selection using `TASK_PROVIDER` environment variable.

```go
switch cfg.TaskProvider {
case "nocodb":
    payrollExpenseProvider = nocodb.NewExpenseService(...)
case "notion":
    payrollExpenseProvider = notion.NewExpenseService(...)
default:
    payrollExpenseProvider = basecamp.NewExpenseAdapter(...)
}
```

**Trade-offs**:
- Pros: Clean separation, explicit config, easy rollback
- Cons: No gradual migration, requires restart to switch

**Implementation**: Single provider active at a time, selected at service initialization.

---

## Files to be Created/Modified

### New Files to Create

1. **`pkg/service/notion/expense.go`**
   - NotionExpenseService struct
   - ExpenseProvider interface implementation
   - Property extraction methods
   - ID conversion logic
   - Status update method
   - Estimated: ~500 lines

### Files to Modify

1. **`pkg/service/service.go`**
   - Add Notion provider initialization in switch statement
   - Update provider selection logic
   - Add configuration validation
   - Changes: ~20 lines (additions to existing switch)

2. **`pkg/handler/payroll/commit.go`**
   - Update `markExpenseSubmissionsAsCompleted()` method
   - Add type assertion for Notion provider
   - Add `extractNotionPageIDsFromPayrolls()` helper
   - Changes: ~50 lines (refactor existing method)

3. **`pkg/config/config.go`**
   - Already has `ExpenseNotionIntegration` struct
   - Verify `ExpenseDBID` and `ContractorDBID` fields exist
   - Changes: None required (already configured)

### Files Referenced (No Changes)

1. **`pkg/handler/payroll/payroll_calculator.go`**
   - No changes needed (uses ExpenseProvider interface)
   - Existing logic works with Notion provider

2. **`pkg/service/basecamp/basecamp.go`**
   - ExpenseProvider interface definition (unchanged)

3. **`pkg/service/nocodb/expense.go`**
   - Reference implementation (patterns to follow)

---

## Dependencies and Integration Points

### Internal Dependencies

1. **Notion Client Service**
   - Package: `pkg/service/notion/notion.go`
   - Status: Already exists
   - Usage: Wrap existing client for database queries

2. **Employee Store**
   - Package: `pkg/store/employee.go`
   - Method: `OneByEmail(db, email) (*model.Employee, error)`
   - Usage: Lookup employee by email for BasecampID

3. **bcModel Package**
   - Package: `pkg/model/basecamp/todo.go`
   - Types: `Todo`, `Assignee`, `Bucket`
   - Usage: Return format for ExpenseProvider interface

4. **Configuration**
   - Package: `pkg/config/config.go`
   - Fields: `TaskProvider`, `ExpenseIntegration.Notion.ExpenseDBID`
   - Usage: Provider selection and database configuration

### External Dependencies

1. **go-notion Library**
   - Package: `github.com/dstotijn/go-notion`
   - Version: Already in use
   - Usage: Notion API client for database queries and updates

2. **Context Package**
   - Package: `context`
   - Usage: API call context management

3. **Standard Library**
   - `strings`: String manipulation (UUID cleaning)
   - `strconv`: String to int conversion (hex parsing)
   - `fmt`: String formatting and error messages
   - `hash/crc32`: Fallback hash function

---

## Implementation Phases

### Phase 1: Core Service Implementation
**Files**: `pkg/service/notion/expense.go`

1. Create `ExpenseService` struct and constructor
2. Implement `GetAllInList()` method
   - `fetchApprovedExpenses()` with pagination
   - `transformPageToTodo()` with property extraction
3. Implement property extraction helpers
   - `extractTitle()`
   - `extractAmount()`
   - `extractCurrency()`
   - `getEmail()` with rollup and fallback
   - `extractAttachmentURL()`
4. Implement ID conversion
   - `notionPageIDToInt()`
5. Implement stub methods
   - `GetGroups()` (returns empty)
   - `GetLists()` (returns empty)

**Estimated Effort**: Core implementation (~400 lines)

### Phase 2: Status Update Implementation
**Files**: `pkg/service/notion/expense.go`

1. Implement `MarkExpenseAsCompleted()` method
2. Implement error handling
   - `handleNotionError()` helper
3. Add logging throughout

**Estimated Effort**: Status updates and error handling (~100 lines)

### Phase 3: Service Integration
**Files**: `pkg/service/service.go`, `pkg/handler/payroll/commit.go`

1. Update service initialization
   - Add Notion case to provider switch
   - Add configuration validation
2. Update commit handler
   - Refactor `markExpenseSubmissionsAsCompleted()`
   - Add type assertion for Notion provider
   - Implement `extractNotionPageIDsFromPayrolls()`
3. Determine UUID storage location in `CommissionExplain`

**Estimated Effort**: Integration changes (~70 lines)

### Phase 4: Testing
**Files**: Test files

1. Unit tests for property extraction
2. Unit tests for page transformation
3. Mock Notion client for testing
4. Integration tests with test database
5. End-to-end payroll flow testing

**Estimated Effort**: Comprehensive test coverage

---

## Testing Strategy

### Unit Tests

**Location**: `pkg/service/notion/expense_test.go`

#### Property Extraction Tests
- [ ] Extract title from rich text property
- [ ] Extract amount from number property
- [ ] Extract currency from select property (with default)
- [ ] Extract email from rollup array
- [ ] Extract email from rollup string
- [ ] Fallback to relation query when rollup fails
- [ ] Extract attachment URL from files property

#### Transformation Tests
- [ ] Transform valid Notion page to Todo
- [ ] Handle missing title (error)
- [ ] Handle zero amount (error)
- [ ] Handle missing employee (error)
- [ ] Handle employee without BasecampID (error)
- [ ] Verify title format: "description | amount | currency"

#### ID Conversion Tests
- [ ] Convert UUID to deterministic integer
- [ ] Same UUID produces same integer
- [ ] Handle malformed UUID (fallback to CRC32)

#### Interface Tests
- [ ] GetAllInList returns approved expenses
- [ ] GetAllInList handles partial transformation failures
- [ ] GetGroups returns empty list
- [ ] GetLists returns empty list

#### Status Update Tests
- [ ] MarkExpenseAsCompleted updates status to "Paid"
- [ ] Handle Notion API errors gracefully

### Integration Tests

**Location**: `pkg/service/notion/expense_integration_test.go`

- [ ] Query test Notion database for approved expenses
- [ ] Transform real pages to Todos
- [ ] Validate employee lookup works
- [ ] Update expense status to "Paid"
- [ ] Verify round-trip: fetch → transform → update

### End-to-End Tests

**Location**: `pkg/handler/payroll/payroll_integration_test.go`

- [ ] Initialize service with Notion provider
- [ ] Fetch expenses during payroll calculation
- [ ] Verify expenses included in payroll
- [ ] Commit payroll with Notion expenses
- [ ] Verify status updates after commit

---

## Risk Assessment and Mitigation

### Risk 1: UUID to Integer Collision
**Probability**: Low (0.01% with 1000 expenses)
**Impact**: Medium (duplicate expense ID in payroll)

**Mitigation**:
- Use last 8 hex characters (32-bit space)
- Log original UUID alongside integer ID
- Store UUID in metadata for validation
- Plan Phase 2 migration to proper UUID field

### Risk 2: Rollup Misconfiguration
**Probability**: Low (validated in research)
**Impact**: Medium (email extraction fails)

**Mitigation**:
- Implement fallback to direct relation query
- Log rollup failures for monitoring
- Document rollup configuration requirements
- Alert on high fallback usage rate

### Risk 3: Missing Employee Records
**Probability**: Medium (data quality dependent)
**Impact**: High (expense not included in payroll)

**Mitigation**:
- Validate email against employee database
- Log missing employees with details
- Skip invalid expenses, don't fail entire batch
- Provide admin dashboard to fix mismatches

### Risk 4: Notion API Rate Limits
**Probability**: Low (fetch once per payroll)
**Impact**: Low (payroll calculation delayed)

**Mitigation**:
- Implement exponential backoff on rate limit errors
- Pagination handles large datasets efficiently
- Status updates can be concurrent with semaphore
- Monitor API usage metrics

### Risk 5: UUID Storage Strategy Unclear
**Probability**: High (requires implementation decision)
**Impact**: Medium (status updates may not work)

**Mitigation**:
- Phase 1: Use existing text field (quick solution)
- Phase 2: Add dedicated UUID field (proper solution)
- Document field usage clearly
- Test round-trip storage and retrieval

---

## Blockers and Open Questions

### Open Questions

1. **UUID Storage Location**
   - **Question**: Which field in `CommissionExplain` should store Notion page UUID?
   - **Options**:
     - Existing text field (e.g., `task_ref`, `notes`)
     - New dedicated field (requires migration)
     - JSON metadata field
   - **Decision Required**: During implementation
   - **Impact**: Affects status update functionality

2. **Employee Matching Edge Cases**
   - **Question**: How to handle contractors not in employee database?
   - **Options**:
     - Skip expense (log error)
     - Create temporary employee record
     - Manual admin intervention
   - **Decision Required**: During implementation
   - **Impact**: Affects expense coverage

3. **Concurrent Status Updates**
   - **Question**: Should status updates be concurrent or sequential?
   - **Options**:
     - Sequential (simple, slower)
     - Concurrent with semaphore (complex, faster)
   - **Decision Required**: During implementation
   - **Impact**: Affects commit performance

### No Blockers

All prerequisites are satisfied:
- [x] Research phase completed
- [x] Notion API patterns documented
- [x] Existing implementation analyzed
- [x] Technical considerations mapped
- [x] ADRs created
- [x] Specifications written
- [x] go-notion library available
- [x] Configuration structure exists
- [x] ExpenseProvider interface defined

---

## Next Steps

### Phase 3: Implementation (Ready to Start)

The planning phase is complete. The next phase should focus on:

1. **Create Notion Expense Service**
   - Implement `pkg/service/notion/expense.go`
   - Follow specifications exactly
   - Use NocoDB implementation as reference

2. **Integrate with Service Layer**
   - Update `pkg/service/service.go`
   - Update `pkg/handler/payroll/commit.go`
   - Determine UUID storage location

3. **Write Comprehensive Tests**
   - Unit tests for all methods
   - Integration tests with test database
   - End-to-end payroll flow tests

4. **Documentation**
   - Code comments and godoc
   - Update project README if needed
   - Deployment instructions

### Handoff to Implementation Phase

**Planning Complete**: All planning objectives achieved ✅

**Key Artifacts**:
- 3 Architecture Decision Records (ADRs)
- 2 Technical Specifications
- Implementation roadmap
- Testing strategy
- Risk assessment

**Blockers**: None (open questions to be resolved during implementation)

**Ready for Implementation**: Yes ✅

---

## Summary

### What We Planned

1. **Architecture Decisions**:
   - UUID→int mapping via hash of last 8 hex characters
   - Email extraction via rollup with fallback to relation query
   - Provider selection via `TASK_PROVIDER=notion` configuration

2. **Technical Design**:
   - NotionExpenseService structure and methods
   - Property extraction patterns
   - Error handling strategy
   - Integration with payroll calculator
   - Status update workflow

3. **Implementation Approach**:
   - Follow NocoDB patterns for consistency
   - Implement ExpenseProvider interface exactly
   - Maintain backward compatibility
   - Enable easy rollback to NocoDB

### What We Validated

1. **Feasibility**: Notion can fully replace NocoDB as expense provider
2. **Interface Compliance**: Design satisfies ExpenseProvider interface
3. **Data Flow**: Email → Employee → BasecampID lookup works
4. **Title Format**: Can build required `"description | amount | currency"` format
5. **Status Updates**: Can mark expenses as "Paid" after payroll commit

### What We Decided

1. **Use hash-based UUID→int conversion** (ADR-001)
2. **Use rollup-first email extraction** (ADR-002)
3. **Use TASK_PROVIDER config selection** (ADR-003)
4. **Follow NocoDB implementation patterns** (Specification)
5. **Store UUID in metadata for reverse lookup** (Specification)

---

## References

### Planning Documents

- **ADR-001**: UUID to Int Mapping Strategy
- **ADR-002**: Email Extraction Strategy
- **ADR-003**: Provider Selection
- **Notion Service Spec**: Technical specification for ExpenseService
- **Payroll Integration Spec**: Integration points and changes

### Research Documents

- **Requirements**: `requirements/requirements.md`
- **Research STATUS**: `research/STATUS.md`
- **Notion API Patterns**: `research/notion-api-patterns.md`
- **Existing Implementation Analysis**: `research/existing-implementation-analysis.md`
- **Technical Considerations**: `research/technical-considerations.md`

### Code References

- **ExpenseProvider Interface**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/basecamp/basecamp.go`
- **NocoDB Implementation**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/nocodb/expense.go`
- **Service Initialization**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/service.go`
- **Payroll Calculator**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/payroll/payroll_calculator.go`
- **Commit Handler**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/payroll/commit.go`

---

**Planning Phase**: COMPLETED ✅
**Next Phase**: Implementation (Create NotionExpenseService)
**Ready to Proceed**: Yes

**Total Planning Time**: Single session (2025-12-04)
**Documentation Created**: 5 documents (3 ADRs, 2 Specs, 1 Status)
**Lines of Documentation**: ~3000 lines
