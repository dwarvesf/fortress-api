# Implementation Tasks: Notion Expense Provider

**Session**: 202512041805-notion-expense-provider
**Status**: Ready for Implementation
**Date Created**: 2025-12-04

---

## Overview

This document provides a detailed task breakdown for implementing the Notion Expense Provider, which replaces NocoDB as the expense provider for payroll calculations.

**Total Tasks**: 15
**Estimated Total Effort**: ~3-4 days

---

## Task Organization

Tasks are organized into 4 implementation phases:

1. **Phase 1: Core Service Implementation** (Tasks 1-7) - Core functionality
2. **Phase 2: Service Integration** (Tasks 8-10) - Connect to existing systems
3. **Phase 3: Configuration** (Tasks 11-12) - Environment setup
4. **Phase 4: Testing & Validation** (Tasks 13-15) - Verification and quality assurance

---

## Phase 1: Core Service Implementation

### Task 1: Create NotionExpenseService Structure and Constructor

**ID**: NOTION-001
**Phase**: 1
**Dependencies**: None
**Estimated Effort**: 1 hour

#### Description

Create the `ExpenseService` struct and constructor function in a new file `pkg/service/notion/expense.go`.

#### Files to Create/Modify

- **CREATE**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/expense.go`

#### Implementation Details

1. Create new file `pkg/service/notion/expense.go`
2. Define `ExpenseService` struct:
   ```go
   type ExpenseService struct {
       client *Service          // Notion client wrapper
       cfg    *config.Config    // Application configuration
       store  *store.Store      // Database store
       repo   store.DBRepo      // Database repository
       logger logger.Logger     // Structured logger
   }
   ```
3. Implement `NewExpenseService` constructor:
   - Accept parameters: client, cfg, store, repo, logger
   - Validate `cfg.ExpenseIntegration.Notion.ExpenseDBID` is not empty
   - Validate database ID format (32 hex characters)
   - Return initialized `*ExpenseService`
   - Use `logger.Fatal()` for configuration errors

#### Acceptance Criteria

- [ ] File `pkg/service/notion/expense.go` created
- [ ] `ExpenseService` struct defined with all required fields
- [ ] `NewExpenseService` constructor implemented
- [ ] Configuration validation present (empty check, format check)
- [ ] Constructor panics with clear message for invalid configuration
- [ ] Code compiles without errors

#### References

- **Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/notion-expense-service-spec.md` (Lines 22-77)
- **Reference**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/nocodb/expense.go` (Constructor pattern)

---

### Task 2: Implement GetAllInList Method

**ID**: NOTION-002
**Phase**: 1
**Dependencies**: NOTION-001, NOTION-003
**Estimated Effort**: 2 hours

#### Description

Implement the main method that fetches approved expenses from Notion and transforms them to `bcModel.Todo` format.

#### Files to Create/Modify

- **MODIFY**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/expense.go`

#### Implementation Details

1. Implement `GetAllInList(todolistID, projectID int) ([]bcModel.Todo, error)` method
2. Call `fetchApprovedExpenses()` to get Notion pages
3. Transform each page using `transformPageToTodo()`
4. Handle errors gracefully:
   - Log individual page transformation errors
   - Continue processing other pages
   - Return partial results if some pages succeed
   - Return error only if query fails or all transformations fail
5. Log summary with counts of successful/failed transformations

#### Acceptance Criteria

- [ ] `GetAllInList` method signature matches `ExpenseProvider` interface
- [ ] Method ignores `todolistID` and `projectID` parameters (uses config)
- [ ] Calls `fetchApprovedExpenses()` to fetch pages
- [ ] Iterates through pages and transforms each one
- [ ] Logs errors for individual failures but continues processing
- [ ] Returns partial results if some transformations succeed
- [ ] Returns error if query fails or all transformations fail
- [ ] Logs summary with successful/failed counts
- [ ] Code compiles without errors

#### References

- **Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/notion-expense-service-spec.md` (Lines 93-157)

---

### Task 3: Implement fetchApprovedExpenses Helper

**ID**: NOTION-003
**Phase**: 1
**Dependencies**: NOTION-001
**Estimated Effort**: 1.5 hours

#### Description

Implement the private helper method that queries Notion database for expenses with Status = "Approved", handling pagination.

#### Files to Create/Modify

- **MODIFY**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/expense.go`

#### Implementation Details

1. Implement `fetchApprovedExpenses(ctx context.Context) ([]notion.Page, error)` method
2. Build Notion database query filter:
   ```go
   filter := &notion.DatabaseQueryFilter{
       Property: "Status",
       DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
           Status: &notion.StatusDatabaseQueryFilter{
               Equals: "Approved",
           },
       },
   }
   ```
3. Implement cursor-based pagination loop:
   - Query with page size = 100
   - Append results to `allPages`
   - Continue if `result.HasMore` is true
   - Update `startCursor` for next page
4. Use `handleNotionError()` for error wrapping

#### Acceptance Criteria

- [ ] Method signature: `fetchApprovedExpenses(ctx context.Context) ([]notion.Page, error)`
- [ ] Builds status filter for "Approved" status
- [ ] Uses database ID from `cfg.ExpenseIntegration.Notion.ExpenseDBID`
- [ ] Implements pagination loop with cursor
- [ ] Sets page size to 100 (maximum allowed)
- [ ] Accumulates all pages across pagination
- [ ] Handles Notion API errors via `handleNotionError()`
- [ ] Returns all approved expense pages
- [ ] Code compiles without errors

#### References

- **Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/notion-expense-service-spec.md` (Lines 210-263)

---

### Task 4: Implement Property Extraction Helpers

**ID**: NOTION-004
**Phase**: 1
**Dependencies**: NOTION-001
**Estimated Effort**: 2 hours

#### Description

Implement helper methods to extract title, amount, currency, and attachment URL from Notion page properties.

#### Files to Create/Modify

- **MODIFY**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/expense.go`

#### Implementation Details

Implement four property extraction methods:

1. **extractTitle**:
   - Extract from `Title` property (type: `DBPropTypeTitle`)
   - Concatenate all rich text segments
   - Return error if missing or empty

2. **extractAmount**:
   - Extract from `Amount` property (type: `DBPropTypeNumber`)
   - Return error if missing or zero

3. **extractCurrency**:
   - Extract from `Currency` property (type: `DBPropTypeSelect`)
   - Default to "VND" if missing or empty
   - Log warning when using default
   - Never return error (always provides default)

4. **extractAttachmentURL**:
   - Extract from `Attachments` property (type: `DBPropTypeFiles`)
   - Return first file URL (empty string if none)
   - Handle both `FileTypeFile` and `FileTypeExternal`

#### Acceptance Criteria

- [ ] `extractTitle(props) (string, error)` implemented
- [ ] `extractAmount(props) (float64, error)` implemented
- [ ] `extractCurrency(props) string` implemented (no error return)
- [ ] `extractAttachmentURL(props) string` implemented
- [ ] Title extraction concatenates all rich text segments
- [ ] Amount extraction rejects zero values
- [ ] Currency extraction defaults to "VND" with warning log
- [ ] Attachment extraction returns first file URL
- [ ] All methods validate property type
- [ ] All methods handle missing properties appropriately
- [ ] Code compiles without errors

#### References

- **Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/notion-expense-service-spec.md` (Lines 357-646)

---

### Task 5: Implement Email Extraction with Rollup and Fallback

**ID**: NOTION-005
**Phase**: 1
**Dependencies**: NOTION-001
**Estimated Effort**: 2.5 hours

#### Description

Implement the two-tier email extraction strategy: primary extraction from rollup property, fallback to direct relation query.

#### Files to Create/Modify

- **MODIFY**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/expense.go`

#### Implementation Details

Implement three email extraction methods:

1. **getEmail** (orchestrator):
   - Try `extractEmailFromRollup()` first
   - If successful and non-empty, return email
   - Otherwise, log warning and fall back to `extractEmailFromRelation()`

2. **extractEmailFromRollup**:
   - Extract from `Email` property (type: `DBPropTypeRollup`)
   - Handle `RollupTypeArray`: Extract from first array element
   - Handle `RollupTypeString`: Use string value directly
   - Return error for missing, wrong type, or empty rollup

3. **extractEmailFromRelation**:
   - Extract from `Requestor` property (type: `DBPropTypeRelation`)
   - Get first relation ID
   - Query contractor page using `client.FindPageByID()`
   - Extract email from contractor page properties
   - Return error if relation empty or email not found

#### Acceptance Criteria

- [ ] `getEmail(ctx, props) (string, error)` implemented
- [ ] `extractEmailFromRollup(props) (string, error)` implemented
- [ ] `extractEmailFromRelation(ctx, props) (string, error)` implemented
- [ ] `getEmail` tries rollup first, falls back to relation query
- [ ] Rollup extraction handles array and string types
- [ ] Relation extraction queries contractor page
- [ ] Warning logged when fallback is used
- [ ] All error cases properly handled
- [ ] Code compiles without errors

#### References

- **Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/notion-expense-service-spec.md` (Lines 465-603)
- **ADR**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/ADRs/ADR-002-email-extraction-strategy.md`

---

### Task 6: Implement UUID to Integer Conversion

**ID**: NOTION-006
**Phase**: 1
**Dependencies**: NOTION-001
**Estimated Effort**: 1 hour

#### Description

Implement the hash-based conversion method that converts Notion page UUIDs to deterministic integer IDs.

#### Files to Create/Modify

- **MODIFY**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/expense.go`

#### Implementation Details

1. Implement `notionPageIDToInt(pageID string) int` method
2. Algorithm:
   - Remove hyphens from UUID: `strings.ReplaceAll(pageID, "-", "")`
   - Take last 8 hex characters
   - Parse hex string to int64: `strconv.ParseInt(hashStr, 16, 64)`
   - Return as int
3. Fallback handling:
   - If UUID too short (< 8 chars): Use CRC32 checksum
   - If parsing fails: Use CRC32 checksum
   - Log warning when using fallback

#### Acceptance Criteria

- [ ] Method signature: `notionPageIDToInt(pageID string) int`
- [ ] Removes hyphens from UUID
- [ ] Extracts last 8 hex characters
- [ ] Parses hex to int using base 16
- [ ] Returns deterministic integer (same UUID -> same int)
- [ ] Falls back to CRC32 for malformed UUIDs
- [ ] Logs warning when fallback is used
- [ ] Imports `hash/crc32` for fallback
- [ ] Code compiles without errors

#### References

- **Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/notion-expense-service-spec.md` (Lines 648-697)
- **ADR**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/ADRs/ADR-001-uuid-to-int-mapping.md`

---

### Task 7: Implement Page Transformation and Status Update Methods

**ID**: NOTION-007
**Phase**: 1
**Dependencies**: NOTION-004, NOTION-005, NOTION-006
**Estimated Effort**: 2 hours

#### Description

Implement the core transformation method that converts Notion pages to `bcModel.Todo`, plus status update method and error handling helper.

#### Files to Create/Modify

- **MODIFY**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/expense.go`

#### Implementation Details

Implement three methods:

1. **transformPageToTodo**:
   - Extract properties using helper methods
   - Lookup employee by email: `store.Employee.OneByEmail(repo.DB(), email)`
   - Validate employee has `BasecampID != 0`
   - Convert page ID to int using `notionPageIDToInt()`
   - Build title in format: `fmt.Sprintf("%s | %.0f | %s", title, amount, currency)`
   - Create `bcModel.Todo` with all fields
   - Log transformation details

2. **MarkExpenseAsCompleted**:
   - Update page status from "Approved" to "Paid"
   - Use `client.UpdatePage()` with status property
   - Log success

3. **handleNotionError**:
   - Type assert to `*notion.Error`
   - Match error code and wrap with context
   - Handle: ObjectNotFound, Unauthorized, RestrictedResource, ValidationError, RateLimited, InternalServerError

#### Acceptance Criteria

- [ ] `transformPageToTodo(ctx, page) (*bcModel.Todo, error)` implemented
- [ ] `MarkExpenseAsCompleted(pageID string) error` implemented
- [ ] `handleNotionError(err error, context string) error` implemented
- [ ] Transformation extracts all properties
- [ ] Employee lookup via email works
- [ ] Employee validation checks BasecampID
- [ ] Title format: "description | %.0f | currency"
- [ ] Todo includes assignee with BasecampID
- [ ] Status update sets status to "Paid"
- [ ] Error handler matches all Notion error codes
- [ ] Code compiles without errors

#### References

- **Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/notion-expense-service-spec.md` (Lines 266-354, 699-777)

---

### Task 8: Implement Stub Methods (GetGroups, GetLists)

**ID**: NOTION-008
**Phase**: 1
**Dependencies**: NOTION-001
**Estimated Effort**: 15 minutes

#### Description

Implement stub methods for `GetGroups` and `GetLists` that return empty slices (not used with Notion provider).

#### Files to Create/Modify

- **MODIFY**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/expense.go`

#### Implementation Details

1. Implement `GetGroups(todosetID, projectID int) ([]bcModel.TodoGroup, error)`:
   - Log debug message: "GetGroups called (not used with Notion provider)"
   - Return empty slice and nil error

2. Implement `GetLists(projectID, todosetID int) ([]bcModel.TodoList, error)`:
   - Log debug message: "GetLists called (not used with Notion provider)"
   - Return empty slice and nil error

#### Acceptance Criteria

- [ ] `GetGroups` method implemented
- [ ] `GetLists` method implemented
- [ ] Both methods return empty slices
- [ ] Both methods return nil error
- [ ] Debug logs include method name and parameters
- [ ] Method signatures match `ExpenseProvider` interface
- [ ] Code compiles without errors

#### References

- **Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/notion-expense-service-spec.md` (Lines 160-205)

---

## Phase 2: Service Integration

### Task 9: Update Service Initialization with Notion Provider

**ID**: NOTION-009
**Phase**: 2
**Dependencies**: NOTION-001, NOTION-002, NOTION-007, NOTION-008
**Estimated Effort**: 1 hour

#### Description

Update the service initialization logic in `pkg/service/service.go` to support Notion provider selection via configuration.

#### Files to Create/Modify

- **MODIFY**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/service.go`

#### Implementation Details

1. Locate the `payrollExpenseProvider` initialization code
2. Replace if-statement pattern with switch statement:
   ```go
   switch cfg.TaskProvider {
   case "nocodb":
       payrollExpenseProvider = nocodb.NewExpenseService(...)
   case "notion":
       if notionSvc == nil {
           logger.Fatal("Notion service not initialized but TASK_PROVIDER=notion")
       }
       payrollExpenseProvider = notion.NewExpenseService(notionSvc, cfg, store, repo, logger.L)
       logger.Info("Initialized Notion expense provider", "database_id", cfg.ExpenseIntegration.Notion.ExpenseDBID)
   default:
       if basecampSvc != nil {
           payrollExpenseProvider = basecamp.NewExpenseAdapter(basecampSvc)
           logger.Info("Initialized Basecamp expense provider (fallback)")
       }
   }
   ```
3. Add validation after switch:
   ```go
   if payrollExpenseProvider == nil {
       logger.Fatal("No expense provider configured", "task_provider", cfg.TaskProvider)
   }
   ```
4. Import notion package: `"github.com/dwarvesf/fortress-api/pkg/service/notion"`

#### Acceptance Criteria

- [ ] Switch statement replaces if-statement for provider selection
- [ ] "notion" case added to switch
- [ ] Notion service nil check present
- [ ] NewExpenseService called with correct parameters
- [ ] Info log when Notion provider initialized
- [ ] Validation ensures provider is not nil
- [ ] Fatal log if no provider configured
- [ ] Notion package imported
- [ ] Code compiles without errors
- [ ] Service starts successfully with `TASK_PROVIDER=notion`

#### References

- **Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/payroll-integration-spec.md` (Lines 9-62)
- **ADR**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/ADRs/ADR-003-provider-selection.md`

---

### Task 10: Update Payroll Commit Handler for Notion Status Updates

**ID**: NOTION-010
**Phase**: 2
**Dependencies**: NOTION-007, NOTION-009
**Estimated Effort**: 2 hours

#### Description

Update the `markExpenseSubmissionsAsCompleted` method in the commit handler to support Notion provider status updates using UUIDs.

#### Files to Create/Modify

- **MODIFY**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/payroll/commit.go`

#### Implementation Details

1. Locate `markExpenseSubmissionsAsCompleted` method (around line 772)
2. Replace single type assertion with switch statement on provider type:
   ```go
   switch provider := h.service.PayrollExpenseProvider.(type) {
   case *nocodb.ExpenseService:
       // NocoDB flow (existing)
   case *notion.ExpenseService:
       // NEW: Notion flow
   case *basecamp.ExpenseAdapter:
       // Basecamp flow (skip)
   default:
       // Unknown provider
   }
   ```
3. Implement Notion case:
   - Call `extractNotionPageIDsFromPayrolls(payrolls)` helper
   - Iterate through page IDs
   - Call `provider.MarkExpenseAsCompleted(pageID)`
   - Log errors for individual failures
4. Create helper function `extractNotionPageIDsFromPayrolls`:
   - Iterate through `payrolls[].BonusExplain[]`
   - Extract UUID from appropriate field (determined during implementation)
   - Deduplicate using map
   - Return slice of unique UUIDs
5. Import notion package if not already imported

#### Open Question (To Resolve During Implementation)

**UUID Storage Location**: Determine which field in `CommissionExplain` will store the Notion page UUID.

**Options**:
- Option 1: Use existing field like `TaskRef` (quick, requires no migration)
- Option 2: Add new `NotionPageID` field (clean, requires migration)
- Option 3: Use JSON metadata field (flexible, more complex)

**Decision Point**: Review `CommissionExplain` model and choose best option based on available fields.

#### Acceptance Criteria

- [ ] Switch statement replaces single type assertion
- [ ] Notion case added to switch
- [ ] `extractNotionPageIDsFromPayrolls` helper implemented
- [ ] UUID extraction logic implemented (field TBD)
- [ ] Deduplication of page IDs works
- [ ] Status update called for each unique page ID
- [ ] Errors logged but don't fail entire operation
- [ ] Notion package imported
- [ ] Code compiles without errors
- [ ] Integration with `transformPageToTodo` stores UUID in chosen field

#### Notes

- **IMPORTANT**: The UUID storage field must be set during `transformPageToTodo` in Task 7
- Update Task 7 code after determining storage location in this task

#### References

- **Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/payroll-integration-spec.md` (Lines 175-300)
- **Spec**: UUID Storage Strategy (Lines 390-489)

---

## Phase 3: Configuration

### Task 11: Verify and Document Configuration Requirements

**ID**: NOTION-011
**Phase**: 3
**Dependencies**: None
**Estimated Effort**: 30 minutes

#### Description

Verify that the configuration structure in `pkg/config/config.go` has the required Notion expense configuration fields, and update `.env.example` with documentation.

#### Files to Create/Modify

- **READ**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/config/config.go`
- **MODIFY**: `/Users/quang/workspace/dwarvesf/fortress-api/.env.example`

#### Implementation Details

1. Verify `ExpenseNotionIntegration` struct exists in `config.go`
2. Verify it has these fields:
   - `ExpenseDBID string`
   - `ContractorDBID string` (optional)
3. If missing, add the struct and wire to main config
4. Update `.env.example` with:
   ```bash
   # Task Provider Selection
   # Options: basecamp, nocodb, notion
   TASK_PROVIDER=notion

   # Notion Expense Configuration (required when TASK_PROVIDER=notion)
   NOTION_SECRET=secret_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
   NOTION_EXPENSE_DB_ID=2bfb69f8-f573-81cb-a2da-f06d28896390

   # Notion Contractor Database (optional, only used for fallback email extraction)
   NOTION_CONTRACTOR_DB_ID=2bfb69f8-f573-805a-8915-000bc44ce188
   ```

#### Acceptance Criteria

- [ ] Configuration structure verified in `config.go`
- [ ] `ExpenseDBID` field present
- [ ] `ContractorDBID` field present (optional)
- [ ] `.env.example` updated with Notion configuration
- [ ] Comments explain when each config is required
- [ ] Example values provided in `.env.example`
- [ ] Documentation notes that ContractorDBID is optional

#### References

- **Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/notion-expense-service-spec.md` (Lines 779-802)
- **Requirements**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/requirements/requirements.md` (Lines 77-84)

---

### Task 12: Create Provider Constant in taskprovider Package

**ID**: NOTION-012
**Phase**: 3
**Dependencies**: None
**Estimated Effort**: 15 minutes

#### Description

Add `ProviderNotion` constant to the taskprovider package for consistent provider identification (if such a package exists).

#### Files to Create/Modify

- **CHECK**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/taskprovider/` (if exists)
- **MODIFY**: Provider constants file (if applicable)

#### Implementation Details

1. Check if `pkg/service/taskprovider/` package exists
2. If yes:
   - Locate constants file (e.g., `constants.go`, `provider.go`)
   - Add: `const ProviderNotion = "notion"`
   - Update service initialization to use constant
3. If no:
   - Skip this task (use string literal "notion" directly)
   - Document decision in task notes

#### Acceptance Criteria

- [ ] taskprovider package checked for existence
- [ ] If exists: `ProviderNotion` constant added
- [ ] If exists: Service initialization uses constant
- [ ] If not exists: Task skipped with documentation
- [ ] Code compiles without errors

#### References

- **Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/payroll-integration-spec.md` (Lines 29-62)

---

## Phase 4: Testing & Validation

### Task 13: Create Unit Tests for Notion Expense Service

**ID**: NOTION-013
**Phase**: 4
**Dependencies**: NOTION-001 through NOTION-008
**Estimated Effort**: 4 hours

#### Description

Create comprehensive unit tests for all Notion expense service methods, including property extraction, transformation, and error handling.

#### Files to Create/Modify

- **CREATE**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/expense_test.go`

#### Implementation Details

Create test cases for:

1. **Constructor Tests**:
   - Test with valid configuration
   - Test with missing ExpenseDBID (should fatal)
   - Test with invalid ExpenseDBID format (should fatal)

2. **Property Extraction Tests**:
   - `TestExtractTitle`: Valid title, empty title, missing property
   - `TestExtractAmount`: Valid amount, zero amount, missing property
   - `TestExtractCurrency`: Valid currency, missing currency (defaults to VND)
   - `TestExtractAttachmentURL`: With file, with external URL, no attachments

3. **Email Extraction Tests**:
   - `TestExtractEmailFromRollup`: Array type, string type, empty rollup
   - `TestExtractEmailFromRelation`: Valid relation, empty relation, missing email
   - `TestGetEmail`: Successful rollup, fallback to relation

4. **ID Conversion Tests**:
   - `TestNotionPageIDToInt`: Valid UUID, deterministic (same UUID -> same int), malformed UUID (fallback to CRC32)

5. **Transformation Tests**:
   - `TestTransformPageToTodo`: Valid page, missing employee, employee without BasecampID
   - Verify title format: "description | amount | currency"

6. **Interface Tests**:
   - `TestGetAllInList`: Single expense, multiple expenses, partial failures, all failures
   - `TestGetGroups`: Returns empty slice
   - `TestGetLists`: Returns empty slice

7. **Status Update Tests**:
   - `TestMarkExpenseAsCompleted`: Success, API error

#### Test Helpers

Create mock structures:
- Mock Notion client
- Mock employee store
- Sample Notion page data
- Sample database page properties

#### Acceptance Criteria

- [ ] Test file created with all test functions
- [ ] All property extraction methods tested
- [ ] Email extraction (rollup and fallback) tested
- [ ] ID conversion tested (including determinism)
- [ ] Page transformation tested
- [ ] GetAllInList tested with various scenarios
- [ ] Status update tested
- [ ] Mock structures created for dependencies
- [ ] All tests pass
- [ ] Test coverage > 80% for expense.go

#### References

- **Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/notion-expense-service-spec.md` (Lines 822-877)

---

### Task 14: Manual Integration Testing with Notion Database

**ID**: NOTION-014
**Phase**: 4
**Dependencies**: NOTION-009, NOTION-011
**Estimated Effort**: 2 hours

#### Description

Perform manual integration testing with the real Notion Expense Request database to verify end-to-end functionality.

#### Prerequisites

- Notion database configured with test data
- Test environment configuration:
  - `TASK_PROVIDER=notion`
  - `NOTION_SECRET` set
  - `NOTION_EXPENSE_DB_ID` set
- At least 2-3 test expenses with Status = "Approved"
- Test employees in database with matching emails

#### Testing Steps

1. **Setup Test Data**:
   - Create test expenses in Notion with Status = "Approved"
   - Ensure expenses have valid Title, Amount, Currency, Email
   - Link to test employees in system

2. **Test Service Initialization**:
   - Start service with `TASK_PROVIDER=notion`
   - Verify Notion provider initialized
   - Check logs for initialization message

3. **Test Expense Fetching**:
   - Call `GetAllInList()` via payroll calculator
   - Verify expenses fetched from Notion
   - Verify transformation to Todo format
   - Check logs for fetch count and transformation success

4. **Test Payroll Calculation**:
   - Run payroll calculation for test batch
   - Verify expenses included in employee bonuses
   - Verify amounts match Notion data
   - Check BonusExplain records created

5. **Test Status Update**:
   - Commit test payroll
   - Verify `markExpenseSubmissionsAsCompleted` called
   - Check Notion database: Status should be "Paid"
   - Verify all test expenses updated

6. **Test Error Handling**:
   - Test with expense missing email (should skip)
   - Test with expense for non-existent employee (should skip)
   - Verify errors logged but process continues

#### Acceptance Criteria

- [ ] Test data created in Notion database
- [ ] Service starts successfully with Notion provider
- [ ] Expenses fetched correctly from Notion
- [ ] Transformation to Todo format works
- [ ] Employee matching via email works
- [ ] Title format correct: "description | amount | currency"
- [ ] Payroll calculation includes Notion expenses
- [ ] Payroll commit succeeds
- [ ] Expense status updated to "Paid" in Notion
- [ ] Errors handled gracefully (invalid expenses skipped)
- [ ] No critical errors in logs

#### References

- **Requirements**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/requirements/requirements.md`
- **Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/payroll-integration-spec.md` (Lines 591-624)

---

### Task 15: End-to-End Payroll Flow Validation

**ID**: NOTION-015
**Phase**: 4
**Dependencies**: NOTION-014
**Estimated Effort**: 2 hours

#### Description

Validate the complete payroll flow with Notion expenses, including calculation accuracy, data persistence, and rollback capability.

#### Testing Scenarios

1. **Complete Payroll Flow**:
   - Fetch approved expenses from Notion
   - Calculate payroll for batch of employees
   - Verify bonus amounts correct
   - Commit payroll to database
   - Verify expense status updated to "Paid"

2. **Accuracy Validation**:
   - Compare calculated amounts with Notion data
   - Verify currency conversion (if applicable)
   - Verify title parsing works correctly
   - Check employee assignment matches BasecampID

3. **Data Persistence**:
   - Verify payroll records created in database
   - Verify CommissionExplain records include expense metadata
   - Verify UUID stored in chosen field (from Task 10)
   - Check BasecampTodoID is hash-based integer

4. **Provider Switching**:
   - Test with `TASK_PROVIDER=notion` (expenses from Notion)
   - Switch to `TASK_PROVIDER=nocodb` (expenses from NocoDB)
   - Verify service restarts successfully
   - Verify no cross-contamination of data

5. **Rollback Capability**:
   - Document steps to revert to NocoDB
   - Verify configuration change is sufficient
   - Test that NocoDB flow still works

6. **Edge Cases**:
   - Multiple expenses for same employee
   - Expenses in different currencies (USD, VND)
   - Large batch (>100 expenses, test pagination)
   - Zero approved expenses (should succeed with empty list)

#### Acceptance Criteria

- [ ] Complete payroll flow works end-to-end
- [ ] Calculation accuracy verified
- [ ] All data persisted correctly
- [ ] UUID storage verified in database
- [ ] Provider switching works (notion <-> nocodb)
- [ ] Rollback to NocoDB tested and documented
- [ ] All edge cases handled correctly
- [ ] No data corruption or loss
- [ ] Performance acceptable (< 5 seconds for 100 expenses)

#### References

- **Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/payroll-integration-spec.md` (Lines 625-698)
- **Planning Status**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/STATUS.md` (Lines 686-698)

---

## Task Dependencies Graph

```
Phase 1 (Core Service):
NOTION-001 (Create Service)
    ├── NOTION-002 (GetAllInList) [requires NOTION-003]
    ├── NOTION-003 (fetchApprovedExpenses)
    ├── NOTION-004 (Property Extractors)
    ├── NOTION-005 (Email Extraction)
    ├── NOTION-006 (UUID to Int)
    └── NOTION-008 (Stub Methods)

NOTION-007 (Transformation) [requires NOTION-004, NOTION-005, NOTION-006]

Phase 2 (Integration):
NOTION-009 (Service Init) [requires NOTION-001, NOTION-002, NOTION-007, NOTION-008]
NOTION-010 (Commit Handler) [requires NOTION-007, NOTION-009]

Phase 3 (Config):
NOTION-011 (Config Verification) [no dependencies]
NOTION-012 (Provider Constant) [no dependencies]

Phase 4 (Testing):
NOTION-013 (Unit Tests) [requires NOTION-001 through NOTION-008]
NOTION-014 (Integration Testing) [requires NOTION-009, NOTION-011]
NOTION-015 (E2E Validation) [requires NOTION-014]
```

---

## Implementation Order Recommendation

### Day 1: Core Service Foundation
1. NOTION-001 - Create service structure (1h)
2. NOTION-004 - Property extractors (2h)
3. NOTION-006 - UUID to int conversion (1h)
4. NOTION-005 - Email extraction (2.5h)
5. NOTION-008 - Stub methods (15m)

**Total**: ~6.75 hours

### Day 2: Core Logic and Integration
1. NOTION-003 - Fetch approved expenses (1.5h)
2. NOTION-007 - Transformation and status update (2h)
3. NOTION-002 - GetAllInList (2h)
4. NOTION-009 - Service initialization (1h)

**Total**: ~6.5 hours

### Day 3: Integration and Configuration
1. NOTION-010 - Commit handler update (2h)
2. NOTION-011 - Config verification (30m)
3. NOTION-012 - Provider constant (15m)
4. NOTION-013 - Unit tests (4h)

**Total**: ~6.75 hours

### Day 4: Testing and Validation
1. NOTION-014 - Integration testing (2h)
2. NOTION-015 - E2E validation (2h)
3. Bug fixes and adjustments (2h)
4. Documentation updates (1h)

**Total**: ~7 hours

---

## Critical Path

The critical path for implementation is:

```
NOTION-001 → NOTION-004 → NOTION-005 → NOTION-007 → NOTION-002 → NOTION-009 → NOTION-010 → NOTION-014 → NOTION-015
```

Any delays in these tasks will delay the overall project completion.

---

## Risk Mitigation

### High-Risk Tasks

1. **NOTION-005 (Email Extraction)**: Complex two-tier strategy
   - **Mitigation**: Implement and test rollup first, then add fallback
   - **Fallback Plan**: If rollup unreliable, use relation query only

2. **NOTION-010 (Commit Handler)**: UUID storage decision required
   - **Mitigation**: Review model early, make quick decision
   - **Fallback Plan**: Use existing text field initially, migrate later

3. **NOTION-014 (Integration Testing)**: Requires real Notion database
   - **Mitigation**: Set up test database early
   - **Fallback Plan**: Use mocked integration tests if database unavailable

### Low-Risk Tasks

- NOTION-001, NOTION-006, NOTION-008: Straightforward implementations
- NOTION-011, NOTION-012: Configuration tasks, minimal code

---

## Success Metrics

### Code Quality
- [ ] All files compile without errors
- [ ] No linter warnings
- [ ] Test coverage > 80%
- [ ] All tests pass

### Functionality
- [ ] Service implements `ExpenseProvider` interface
- [ ] Expenses fetched from Notion database
- [ ] Transformation to Todo format correct
- [ ] Employee matching works via email
- [ ] Status updates work after payroll commit

### Integration
- [ ] Provider selection via configuration
- [ ] Payroll calculation includes Notion expenses
- [ ] Payroll commit succeeds
- [ ] No changes required to existing Basecamp/NocoDB flows

### Deployment
- [ ] Configuration documented
- [ ] Rollback plan tested
- [ ] No breaking changes to existing functionality

---

## Post-Implementation Tasks

After all implementation tasks are complete, consider these follow-up items:

1. **Performance Optimization** (optional):
   - Add employee caching for repeated lookups
   - Implement concurrent status updates with semaphore
   - Optimize property extraction (reduce allocations)

2. **Monitoring** (recommended):
   - Add metrics for fetch count, errors, duration
   - Add alerts for high error rates
   - Track rollup fallback usage rate

3. **Documentation** (required):
   - Update project README if needed
   - Add godoc comments to all exported methods
   - Document deployment procedure
   - Create runbook for troubleshooting

4. **Phase 2 Improvements** (future):
   - Add dedicated UUID field to CommissionExplain model
   - Migrate existing records to new field
   - Remove hash-based ID conversion (use UUID directly)

---

## References

### Planning Documents
- **Requirements**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/requirements/requirements.md`
- **Planning Status**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/STATUS.md`
- **Notion Service Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/notion-expense-service-spec.md`
- **Payroll Integration Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/planning/specifications/payroll-integration-spec.md`

### Architecture Decision Records
- **ADR-001**: UUID to Integer Mapping Strategy
- **ADR-002**: Email Extraction Strategy
- **ADR-003**: Provider Selection

### Code References
- **ExpenseProvider Interface**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/basecamp/basecamp.go`
- **NocoDB Reference Implementation**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/nocodb/expense.go`
- **Service Initialization**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/service.go`
- **Payroll Calculator**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/payroll/payroll_calculator.go`
- **Commit Handler**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/payroll/commit.go`

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2025-12-04 | Initial task breakdown created | Planning Phase |

---

**End of Implementation Tasks Document**
