# Implementation Tasks: Payout Commit Command

## Overview

This document provides an ordered, actionable task breakdown for implementing the `?payout commit` feature across fortress-api and fortress-discord repositories. Tasks are organized by phase and include file paths, dependencies, and implementation guidelines.

**Total Estimated Tasks**: 24
**Repositories**: 2 (fortress-api, fortress-discord)
**Phases**: 4

---

## Phase 1: fortress-api Notion Services (7 tasks)

### Task 1.1: Add QueryPendingPayablesByPeriod Method
**Description**: Implement method to query all pending contractor payables for a given period.

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payables.go`

**Dependencies**: None

**Implementation Details**:
- Create `PendingPayable` struct with fields: PageID, ContractorPageID, ContractorName, Total, Currency, Period, PayoutItemPageIDs
- Build Notion filter for Payment Status="Pending" AND Period=parameter
- Handle pagination (page size 100, cursor-based)
- Extract properties: Payment Status, Period, Contractor (relation), Total, Currency, Payout Items (relation array)
- Return empty slice (not error) if no results found
- Log at DEBUG level: query start, result count

**Test Coverage**:
- Multiple results with pagination
- Single result
- Empty results
- Missing contractor relation
- Empty payout items array
- Notion API errors

---

### Task 1.2: Add UpdatePayableStatus Method
**Description**: Implement method to update payable's Payment Status and Payment Date.

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payables.go`

**Dependencies**: None

**Implementation Details**:
- Accept parameters: pageID, status, paymentDate (YYYY-MM-DD format)
- Validate pageID is not empty
- Build UpdatePageParams with:
  - `Payment Status`: Status property type (NOT Select)
  - `Payment Date`: Date property type with nt.NewDateTime
- Call client.UpdatePage
- Log at DEBUG level: operation start, success/failure

**Test Coverage**:
- Successful update
- Empty page ID (error)
- Notion API error
- Idempotency (re-running with same values)
- Correct property types used

---

### Task 1.3: Add GetContractorPayDay Method
**Description**: Implement method to retrieve PayDay from contractor's Service Rate.

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payables.go`

**Dependencies**: None

**Implementation Details**:
- Accept contractorPageID parameter
- Validate contractorPageID is not empty
- Query Service Rate database with filter: Contractor relation contains contractorPageID
- Use page size 1 (only need first result)
- Extract PayDay Select property value
- Parse PayDay string ("1" or "15") to integer
- Return error if no Service Rate found or PayDay missing
- Log at DEBUG level: query, PayDay value found

**Test Coverage**:
- PayDay=15 found
- PayDay=1 found
- No Service Rate found (error)
- Missing PayDay property (error)
- Invalid PayDay value (error)
- Empty contractor ID (error)

---

### Task 1.4: Add GetPayoutWithRelations Method
**Description**: Implement method to fetch payout with Invoice Split and Refund relations.

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payouts.go`

**Dependencies**: None

**Implementation Details**:
- Create `PayoutWithRelations` struct with fields: PageID, Status, InvoiceSplitID, RefundRequestID
- Accept payoutPageID parameter
- Validate payoutPageID is not empty
- Call client.FindPageByID
- Cast page.Properties to DatabasePageProperties
- Extract Status (Status property)
- Extract "02 Invoice Split" relation (first ID or empty)
- Extract "01 Refund" relation (first ID or empty)
- Return PayoutWithRelations struct
- Log at DEBUG level: fetch operation, relation IDs found

**Test Coverage**:
- Payout with Invoice Split only
- Payout with Refund only
- Payout with both relations
- Payout with no relations
- Empty page ID (error)
- Property type cast error

---

### Task 1.5: Add UpdatePayoutStatus Method
**Description**: Implement method to update payout's Status.

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payouts.go`

**Dependencies**: None

**Implementation Details**:
- Accept parameters: pageID, status
- Validate pageID is not empty
- Build UpdatePageParams with Status property type (NOT Select)
- Call client.UpdatePage
- Log at DEBUG level: operation start, success/failure

**Test Coverage**:
- Successful update
- Empty page ID (error)
- Notion API error
- Correct property type (Status, not Select)

---

### Task 1.6: Add UpdateInvoiceSplitStatus Method
**Description**: Implement method to update invoice split's Status.

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/invoice_split.go`

**Dependencies**: None

**Implementation Details**:
- Accept parameters: pageID, status
- Validate pageID is not empty
- Build UpdatePageParams with **Select** property type (CRITICAL: NOT Status)
- Call client.UpdatePage
- Log at DEBUG level: operation start, success/failure

**Test Coverage**:
- Successful update
- Empty page ID (error)
- Notion API error
- **CRITICAL**: Verify Select property type used (not Status)

**IMPORTANT NOTE**: Invoice Split uses Select type, unlike other databases that use Status type.

---

### Task 1.7: Add UpdateRefundRequestStatus Method
**Description**: Implement method to update refund request's Status.

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/refund_requests.go` (if doesn't exist)

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/refund_requests.go` (if exists)

**Dependencies**: None

**Implementation Details**:
- Accept parameters: pageID, status
- Validate pageID is not empty
- Build UpdatePageParams with Status property type (NOT Select)
- Call client.UpdatePage
- Log at DEBUG level: operation start, success/failure

**Test Coverage**:
- Successful update (Approved → Paid)
- Empty page ID (error)
- Notion API error
- Correct property type (Status, not Select)

---

## Phase 2: fortress-api Handler/Controller (7 tasks)

### Task 2.1: Create Controller Interface
**Description**: Define controller interface for contractor payables operations.

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/contractorpayables/interface.go`

**Dependencies**: None

**Implementation Details**:
- Define IController interface with methods:
  - PreviewCommit(ctx context.Context, month string, batch int) (*contractorpayables.PreviewCommitResponse, error)
  - CommitPayables(ctx context.Context, month string, batch int) (*contractorpayables.CommitResponse, error)
- Import handler types for response structs

---

### Task 2.2: Create Controller Implementation
**Description**: Implement business logic for preview and commit operations with cascade updates.

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/contractorpayables/contractorpayables.go`

**Dependencies**: Tasks 1.1-1.7 (all Notion service methods)

**Implementation Details**:
- Create controller struct with config, logger, service dependencies
- Implement PreviewCommit method:
  - Convert month (YYYY-MM) to period (YYYY-MM-01)
  - Query pending payables for period
  - For each payable, get contractor's PayDay
  - Filter by matching batch parameter
  - Calculate total amount
  - Build ContractorPreview array
  - Return PreviewCommitResponse
- Implement CommitPayables method:
  - Query and filter payables (same as preview)
  - Return error if no payables found
  - For each payable, call commitSinglePayable
  - Track success/failure counts
  - Collect error details
  - Return CommitResponse with counts and errors
- Implement commitSinglePayable helper:
  - For each payout item, call commitPayoutItem
  - Update payable status to "Paid" with current date
  - Return error if payable update fails
- Implement commitPayoutItem helper:
  - Get payout with relations
  - If InvoiceSplitID exists, update Invoice Split status
  - If RefundRequestID exists, update Refund Request status
  - Update payout status to "Paid"
  - Log errors but continue (best-effort)
- Add PayableToCommit struct for internal data
- Log at DEBUG level throughout cascade updates

**Test Coverage**:
- Preview: multiple payables, filtering by PayDay, empty results, query errors
- Commit: full success, partial failure, PayDay filtering, cascade updates
- Cascade: Invoice Split only, Refund only, both, neither
- Error handling: GetPayDay error, payout update error, payable update error

---

### Task 2.3: Register Controller in Main Controller
**Description**: Add contractor payables controller to main controller struct.

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/controller.go`

**Dependencies**: Task 2.1, 2.2

**Implementation Details**:
- Import contractorpayables controller package
- Add ContractorPayables field to Controller struct
- Initialize in New() function with config, logger, service

---

### Task 2.4: Create Handler Request/Response Types
**Description**: Define request/response structs for API endpoints.

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/contractorpayables/request.go`

**Dependencies**: None

**Implementation Details**:
- Create structs:
  - PreviewCommitRequest (query params): Month (required, form binding), Batch (required, oneof=1 15, form binding)
  - CommitRequest (JSON body): Month (required, json binding), Batch (required, oneof=1 15, json binding)
  - PreviewCommitResponse: Month, Batch, Count, TotalAmount, Contractors array
  - ContractorPreview: Name, Amount, Currency, PayableID
  - CommitResponse: Month, Batch, Updated, Failed, Errors array
  - CommitError: PayableID, Error
- Use Gin validation tags

---

### Task 2.5: Create Handler Interface
**Description**: Define handler interface for HTTP operations.

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/contractorpayables/interface.go`

**Dependencies**: None

**Implementation Details**:
- Define IHandler interface with methods:
  - PreviewCommit(c *gin.Context)
  - Commit(c *gin.Context)

---

### Task 2.6: Create Handler Implementation
**Description**: Implement HTTP handlers with validation and response formatting.

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/contractorpayables/contractorpayables.go`

**Dependencies**: Tasks 2.1-2.5

**Implementation Details**:
- Create handler struct with controller, logger, config
- Implement PreviewCommit handler:
  - Bind query parameters
  - Validate month format (YYYY-MM)
  - Call controller.PreviewCommit
  - Handle "no pending payables" error (return 200 with count=0)
  - Return 200 OK with preview data
  - Return 500 on controller errors
  - Add Swagger annotations
- Implement Commit handler:
  - Bind JSON request body
  - Validate month format
  - Call controller.CommitPayables
  - Handle "no pending payables" error (return 404)
  - Return 200 OK if all succeeded
  - Return 207 Multi-Status if partial failure
  - Return 500 on complete failure
  - Add Swagger annotations
- Add isValidMonthFormat helper function
- Log at handler level with request context

**Test Coverage**:
- PreviewCommit: valid request, empty results, invalid month format, invalid batch, missing parameters
- Commit: valid request, partial success, invalid body, no payables (404)
- Validation: month format, batch values, required fields

---

### Task 2.7: Register Routes and Handlers
**Description**: Register API endpoints with middleware and permissions.

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/routes/v1.go`
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/handler.go`

**Dependencies**: Tasks 2.5, 2.6

**Implementation Details**:
- In handler.go:
  - Import contractorpayables handler package
  - Add ContractorPayables field to Handler struct
  - Initialize in New() function
- In v1.go:
  - Create contractor-payables route group
  - Add authMiddleware
  - Register GET /preview-commit with PayrollsRead permission
  - Register POST /commit with PayrollsCreate permission

---

## Phase 3: fortress-discord Command (7 tasks)

### Task 3.1: Create Discord Model Types
**Description**: Define data models for payout operations.

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/model/payout.go`

**Dependencies**: None

**Implementation Details**:
- Create structs matching API response format:
  - PayoutPreview: Month, Batch, Count, TotalAmount, Contractors array
  - ContractorPreview: Name, Amount, Currency, PayableID
  - PayoutCommitResult: Month, Batch, Updated, Failed, Errors array
  - CommitError: PayableID, Error
- Use JSON tags for serialization

---

### Task 3.2: Create Adapter Interface and Implementation
**Description**: Implement HTTP client adapter to call fortress-api endpoints.

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/adapter/fortress/payout.go`

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/adapter/fortress/fortress.go` (add interface)

**Dependencies**: Task 3.1

**Implementation Details**:
- In fortress.go:
  - Add IPayoutAdapter interface with PreviewCommit, ExecuteCommit methods
  - Add Payout() method to main IAdapter interface
  - Initialize payout adapter in adapter struct
- In payout.go:
  - Create payoutAdapter struct with config, logger, http.Client
  - Implement PreviewCommit:
    - Build GET URL with query parameters
    - Add Bearer token authorization
    - Parse JSON response into PayoutPreview
    - Handle HTTP errors
  - Implement ExecuteCommit:
    - Build POST URL with JSON body
    - Add Bearer token authorization
    - Parse JSON response into PayoutCommitResult
    - Accept both 200 OK and 207 Multi-Status
- Add timeout to HTTP client (30 seconds recommended)

**Test Coverage**:
- Successful API calls
- HTTP errors (4xx, 5xx)
- JSON parsing errors
- Network timeout

---

### Task 3.3: Create Service Interface and Implementation
**Description**: Implement service layer calling API adapter.

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/service/payout/interface.go`
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/service/payout/service.go`

**Dependencies**: Task 3.2

**Implementation Details**:
- In interface.go:
  - Define IService interface with PreviewCommit, ExecuteCommit methods
- In service.go:
  - Create service struct with config, logger, adapter
  - Implement PreviewCommit: call adapter, log, wrap errors
  - Implement ExecuteCommit: call adapter, log, wrap errors
- Log at DEBUG level: API calls, result counts

**Test Coverage**:
- Successful operations
- Adapter errors
- Error wrapping

---

### Task 3.4: Create Discord View
**Description**: Implement Discord embed views for command responses.

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/view/payout/payout.go`

**Dependencies**: Task 3.1

**Implementation Details**:
- Create View struct with logger, Discord session
- Implement Help() - show command usage
- Implement NoPayables() - show info message when count=0
- Implement ShowConfirmation():
  - Build embed with preview data
  - Limit contractor list to 10 (avoid embed size limit)
  - Show "... and N more" if >10 contractors
  - Create Confirm/Cancel buttons
  - CustomID format: `payout_commit_confirm:{month}:{batch}`
  - Use ColorOrange for confirmation
- Implement ShowResult():
  - Show success embed (ColorGreen) if Failed=0
  - Show partial success embed (ColorOrange) if Failed>0
  - Include error details (limit to 5 errors)
  - Show "... and N more errors" if >5 errors
- Use consistent color scheme and formatting

**Test Coverage**:
- Help message formatting
- Empty results message
- Confirmation embed with various contractor counts
- Success result
- Partial failure result with errors

---

### Task 3.5: Create Command Implementation
**Description**: Implement Discord command handler.

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/command/payout/command.go`
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/command/payout/new.go`

**Dependencies**: Tasks 3.3, 3.4

**Implementation Details**:
- Create Command struct with config, logger, service, view
- Implement Prefix() - return ["payout"]
- Implement Execute():
  - Parse subcommand
  - Route to commit() or Help()
- Implement PermissionCheck() - require admin or ops role
- Implement commit():
  - Validate arguments (need month and batch)
  - Validate month format (YYYY-MM)
  - Validate batch (1 or 15)
  - Call service.PreviewCommit
  - Handle empty results (show NoPayables view)
  - Show confirmation view
- Implement ExecuteCommitConfirmation():
  - Call service.ExecuteCommit
  - Show result view
- Add isValidMonthFormat helper
- Log at DEBUG level: command execution, user ID

**Test Coverage**:
- Argument parsing
- Validation (month format, batch values)
- Permission checking
- Service call handling
- Error handling

---

### Task 3.6: Register Command
**Description**: Register payout command in command list.

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/command/command.go`

**Dependencies**: Task 3.5

**Implementation Details**:
- Import payout command package
- Add to NewCommands() function
- Initialize with config, logger, service, view

---

### Task 3.7: Implement Button Interaction Handler
**Description**: Handle Confirm/Cancel button clicks.

**Files to Modify**:
- Main bot interaction handler (location varies, typically in main.go or interaction handler)

**Dependencies**: Task 3.5

**Implementation Details**:
- In handleInteractionCreate or similar:
  - Check for customID prefix `payout_commit_confirm:`
  - Parse month and batch from customID
  - Acknowledge interaction (DeferredMessageUpdate)
  - Call command.ExecuteCommitConfirmation
  - Handle payout_commit_cancel:
    - Update message to show "Payout commit cancelled"
    - Clear buttons
- Log interaction handling

**Test Coverage**:
- Confirm button click
- Cancel button click
- Invalid customID format

---

## Phase 4: Integration & Testing (3 tasks)

### Task 4.1: Write Unit Tests for fortress-api
**Description**: Create comprehensive unit tests for all new code.

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payables_test.go` (if doesn't exist, add tests)
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payouts_test.go` (add tests)
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/invoice_split_test.go` (add tests)
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/refund_requests_test.go` (add tests)
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/contractorpayables/contractorpayables_test.go`
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/contractorpayables/contractorpayables_test.go`

**Dependencies**: All Phase 1 and Phase 2 tasks

**Implementation Details**:
- Follow test case specifications in:
  - `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601081935-payout-commit-command/test-cases/unit/notion-service-tests.md`
  - `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601081935-payout-commit-command/test-cases/unit/controller-tests.md`
  - `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601081935-payout-commit-command/test-cases/unit/handler-tests.md`
- Mock Notion client for service tests
- Mock controller for handler tests
- Mock Notion services for controller tests
- Aim for 100% line coverage
- Use table-driven tests where appropriate
- Test all error paths
- **CRITICAL**: Test property type differences (Select vs Status)

**Test Execution**:
```bash
go test ./pkg/service/notion -v
go test ./pkg/controller/contractorpayables -v
go test ./pkg/handler/contractorpayables -v
```

---

### Task 4.2: Write Unit Tests for fortress-discord
**Description**: Create unit tests for Discord command components.

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/command/payout/command_test.go`
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/service/payout/service_test.go`
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/view/payout/payout_test.go`
- `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/adapter/fortress/payout_test.go`

**Dependencies**: All Phase 3 tasks

**Implementation Details**:
- Mock HTTP client for adapter tests
- Mock adapter for service tests
- Mock service for command tests
- Mock Discord session for view tests
- Test argument parsing and validation
- Test permission checking
- Test button interaction handling
- Test embed formatting

---

### Task 4.3: Integration Testing and Documentation
**Description**: Perform end-to-end testing and update documentation.

**Files to Create**:
- `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601081935-payout-commit-command/implementation/STATUS.md`

**Files to Modify**:
- `/Users/quang/workspace/dwarvesf/fortress-api/README.md` (if API changes need documenting)
- Swagger documentation (regenerate with `make gen-swagger`)

**Dependencies**: All previous tasks

**Implementation Details**:
- Manual testing checklist:
  - [ ] ?payout help shows correct usage
  - [ ] ?payout commit with invalid args shows errors
  - [ ] ?payout commit 2025-01 15 with no payables shows count=0
  - [ ] ?payout commit 2025-01 15 with payables shows preview
  - [ ] Cancel button dismisses confirmation
  - [ ] Confirm button executes commit and shows result
  - [ ] Partial failure shows error details
  - [ ] Permission check works (non-admin blocked)
  - [ ] API endpoints return correct response formats
  - [ ] Cascade updates work correctly in Notion
  - [ ] Idempotency: re-running commit is safe
- Test with test Notion workspace (not production)
- Verify database IDs in configuration match test environment
- Document any environment variables needed
- Update STATUS.md with completion status
- Run `make gen-swagger` to update API docs
- Verify CI/CD passes (make test, make lint)

---

## Configuration Requirements

### fortress-api Environment Variables
```bash
# Add to .env if not already present
NOTION_DATABASE_CONTRACTOR_PAYABLES=2c264b29-b84c-8037-807c-000bf6d0792c
NOTION_DATABASE_CONTRACTOR_PAYOUTS=2c564b29-b84c-8045-80ee-000bee2e3669
NOTION_DATABASE_INVOICE_SPLIT=2c364b29-b84c-804f-9856-000b58702dea
NOTION_DATABASE_REFUND_REQUEST=2cc64b29-b84c-8066-adf2-cc56171cedf4
NOTION_DATABASE_SERVICE_RATE=2c464b29-b84c-80cf-bef6-000b42bce15e
```

### fortress-discord Environment Variables
```bash
# Add to .env if not already present
FORTRESS_API_URL=https://api.fortress.example.com
FORTRESS_API_KEY=your-api-key-here
```

---

## Development Order Recommendations

1. **Start with Phase 1** (Notion services) - foundational layer, can be tested independently
2. **Then Phase 2** (Controller/Handler) - builds on Notion services, can test with mocked services
3. **Then Phase 3** (Discord command) - requires API to be complete for integration testing
4. **Finally Phase 4** (Testing/Documentation) - comprehensive validation

**Parallel Work Opportunities**:
- Phase 1 tasks (1.1-1.7) can be worked on in parallel by different developers
- Phase 3 tasks (3.1-3.4) can start while Phase 2 is in progress (using mocked API responses)

---

## Critical Implementation Notes

### Property Type Differences (CRITICAL)
**Invoice Split uses Select type, all others use Status type**:
```go
// Invoice Split - MUST use Select
props["Status"] = nt.DatabasePageProperty{
    Select: &nt.SelectOptions{Name: "Paid"},
}

// All others - use Status
props["Status"] = nt.DatabasePageProperty{
    Status: &nt.SelectOptions{Name: "Paid"},
}
```

### Error Handling Strategy
- **Best-effort updates**: Continue processing on individual failures
- **Track failures**: Maintain counts and error details
- **Detailed logging**: Every update operation logged with page IDs
- **Idempotency**: Re-running commit is safe

### Testing Priority
1. Property type tests (prevent Notion API rejections)
2. Pagination handling (prevent data loss)
3. Empty relation handling (prevent nil pointer errors)
4. Cascade update sequence (ensure correct order)
5. Error propagation (ensure proper handling)

---

## References

- **Full Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/specs/payout-commit-command.md`
- **Requirements**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601081935-payout-commit-command/requirements/requirements.md`
- **API Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601081935-payout-commit-command/planning/specifications/01-api-endpoints.md`
- **Notion Services Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601081935-payout-commit-command/planning/specifications/02-notion-services.md`
- **Discord Command Spec**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601081935-payout-commit-command/planning/specifications/03-discord-command.md`
- **ADR**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601081935-payout-commit-command/planning/ADRs/ADR-001-cascade-status-update.md`
- **Test Cases**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601081935-payout-commit-command/test-cases/`

---

## Completion Checklist

### Phase 1: fortress-api Notion Services
- [ ] Task 1.1: QueryPendingPayablesByPeriod implemented and tested
- [ ] Task 1.2: UpdatePayableStatus implemented and tested
- [ ] Task 1.3: GetContractorPayDay implemented and tested
- [ ] Task 1.4: GetPayoutWithRelations implemented and tested
- [ ] Task 1.5: UpdatePayoutStatus implemented and tested
- [ ] Task 1.6: UpdateInvoiceSplitStatus implemented and tested (Select type!)
- [ ] Task 1.7: UpdateRefundRequestStatus implemented and tested

### Phase 2: fortress-api Handler/Controller
- [ ] Task 2.1: Controller interface created
- [ ] Task 2.2: Controller implementation with cascade updates
- [ ] Task 2.3: Controller registered
- [ ] Task 2.4: Handler request/response types created
- [ ] Task 2.5: Handler interface created
- [ ] Task 2.6: Handler implementation with validation
- [ ] Task 2.7: Routes and handlers registered

### Phase 3: fortress-discord Command
- [ ] Task 3.1: Discord model types created
- [ ] Task 3.2: Adapter interface and implementation
- [ ] Task 3.3: Service interface and implementation
- [ ] Task 3.4: Discord view created
- [ ] Task 3.5: Command implementation
- [ ] Task 3.6: Command registered
- [ ] Task 3.7: Button interaction handler

### Phase 4: Integration & Testing
- [ ] Task 4.1: fortress-api unit tests (100% coverage)
- [ ] Task 4.2: fortress-discord unit tests
- [ ] Task 4.3: Integration testing and documentation

### Final Validation
- [ ] All unit tests passing
- [ ] Manual testing completed
- [ ] Swagger documentation regenerated
- [ ] CI/CD pipeline passing
- [ ] STATUS.md updated
- [ ] Code reviewed and approved
