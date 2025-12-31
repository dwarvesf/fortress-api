# ADR 001: Automated Contractor Fee Creation from Approved Task Order Logs

## Status

Proposed

## Context

Currently, the process of creating Contractor Fee entries from approved Task Order Logs is manual. This creates several problems:

1. **Manual Effort**: Operations team must manually check for approved orders and create corresponding fee entries
2. **Inconsistency**: Manual process leads to potential delays or missed entries
3. **Error-Prone**: Human error can result in incorrect fee calculations or missing entries
4. **Scalability**: As the number of contractors grows, manual processing becomes unsustainable

### Business Requirements

- Automatically process Task Order Log entries with Status="Approved"
- Find matching Contractor Rates based on contractor and date
- Create Contractor Fee entries linking the order and rate
- Update processed orders to Status="Completed"
- Ensure idempotency (no duplicate fees)
- Handle errors gracefully without stopping entire batch

### Technical Context

The system uses Notion as the data source with three key databases:
- **Task Order Log** (ID: `2b964b29-b84c-801e-ab9e-000b0662b987`) - Contains work orders
- **Contractor Rates** (ID: `2c464b29-b84c-80cf-bef6-000b42bce15e`) - Contains billing rates
- **Contractor Fees** (ID: `2c264b29-b84c-8037-807c-000bf6d0792c`) - Target for fee entries

The application follows a layered architecture:
```
Routes → Handlers → Services → Notion API
```

Existing patterns in the codebase:
- Cronjob endpoints under `/cronjobs` with auth/permission middleware
- Service layer handles Notion API interactions
- Handler orchestrates workflow and response formatting
- Continue-on-error pattern for batch operations (see: SyncTaskOrderLogs)

## Decision

We will implement an automated cronjob endpoint following these architectural decisions:

### Decision 1: Service Layer Design

**Chosen**: Extend existing Notion services with new methods

**Rationale**:
- Maintains clear service boundaries (one service per Notion database)
- Follows existing codebase pattern
- Enables method reuse across different handlers
- Keeps services focused and testable

**Implementation**:
- `TaskOrderLogService.QueryApprovedOrders()` - Query approved orders
- `TaskOrderLogService.UpdateOrderStatus()` - Update status after processing
- `ContractorRatesService.FindActiveRateByContractor()` - Find matching rate by contractor and date
- `ContractorFeesService.CheckFeeExistsByTaskOrder()` - Idempotency check
- `ContractorFeesService.CreateContractorFee()` - Create fee entry

### Decision 2: Idempotency Strategy

**Chosen**: Query Contractor Fees by Task Order Log relation before creating

**Rationale**:
- Notion does not support database-level unique constraints
- Application-level check is the standard pattern
- Allows safe re-execution of cronjob
- Prevents billing errors from duplicate fees

**Implementation**:
```
For each approved order:
  1. Check if fee exists (query by Task Order Log relation)
  2. If exists: Skip, log debug message
  3. If not exists: Create fee entry
```

### Decision 3: Error Handling Pattern

**Chosen**: Continue-on-error with detailed logging and statistics tracking

**Rationale**:
- Individual failures should not block processing of other orders
- Maximizes successful fee creation
- Matches existing cronjob pattern (SyncTaskOrderLogs)
- Provides visibility into partial failures

**Implementation**:
- Continue processing on individual order failures
- Track statistics (created, skipped, errors)
- Log warnings/errors for failed orders
- Return detailed response with per-order results

### Decision 4: Status Update Timing

**Chosen**: Update Task Order Log status AFTER creating Contractor Fee (create-then-update)

**Rationale**:
- Prevents orphaned status updates if fee creation fails
- Fee creation is the critical operation
- Status update failure is non-critical (fee still created)
- Allows retry of status update without recreating fee

**Implementation**:
```
1. Create Contractor Fee
2. If creation succeeds → Update status to "Completed"
3. If status update fails → Log error, continue (fee already created)
```

### Decision 5: Contractor Rate Matching

**Chosen**: Fetch active rates via Notion query, filter by date range in application

**Rationale**:
- Notion API has limited date filtering capabilities
- Application-level date logic provides more control
- Can handle nil End Date (ongoing rate) correctly
- Clear logic for date range validation

**Implementation**:
```
1. Query Contractor Rates: Contractor={contractor_id}, Status=Active
2. Filter in memory: Start Date <= Order Date <= End Date (or End Date is nil)
3. Return first matching rate
```

### Decision 6: Contractor Extraction

**Chosen**: Use Contractor rollup property directly from Task Order Log

**Rationale**:
- Simpler than querying Deployment separately
- Rollup is already computed by Notion
- Reduces API calls
- Matches data model design (rollup exists for this purpose)

**Implementation**:
- Extract Contractor page ID from rollup property (property ID: `q?kW`)
- Handle rollup array structure (extract first relation ID)
- Skip order if rollup is empty (log warning)

## Consequences

### Positive

1. **Automation**: Eliminates manual fee creation process
2. **Consistency**: All approved orders processed systematically
3. **Scalability**: Can handle increasing number of contractors
4. **Idempotency**: Safe to re-run without creating duplicates
5. **Resilience**: Individual failures don't block entire batch
6. **Observability**: Detailed logging and statistics tracking
7. **Maintainability**: Follows existing codebase patterns
8. **Testability**: Service methods can be unit tested independently

### Negative

1. **Complexity**: Additional code to maintain
2. **Dependencies**: Relies on Notion API availability
3. **Data Quality**: Requires accurate Contractor rollup and rate data
4. **No Rollback**: Status updates are not rolled back if subsequent operations fail
5. **Rate Matching**: Date range logic may have edge cases

### Mitigation Strategies

| Risk | Mitigation |
|------|------------|
| Notion property name changes | Use property IDs from schemas, extensive debug logging |
| Missing contractor data | Skip order with warning, continue processing |
| No matching rate found | Skip order with error, continue processing |
| Date range edge cases | Handle nil End Date explicitly, clear comparison logic |
| API rate limiting | Implement pagination, respect Notion API limits |
| Partial failures | Continue-on-error pattern, detailed response statistics |

## Alternatives Considered

### Alternative 1: Dedicated Orchestrator Service

**Approach**: Create a new service (e.g., `ContractorFeeOrchestrator`) that encapsulates the entire workflow.

**Pros**:
- Single responsibility
- Easier end-to-end testing
- Workflow logic in one place

**Cons**:
- Deviation from existing pattern (service-per-database)
- Less code reuse
- More coupling between Notion databases

**Rejected because**: Doesn't follow established codebase patterns.

### Alternative 2: Event-Driven Architecture

**Approach**: Trigger fee creation when Task Order Log status changes to "Approved" (webhook-based).

**Pros**:
- Real-time processing
- No polling needed
- Automatic triggering

**Cons**:
- Requires Notion webhook setup (not available in all plans)
- More complex error handling
- Harder to batch process existing approved orders

**Rejected because**: Cronjob pattern is simpler and matches existing infrastructure.

### Alternative 3: Update Status Before Creating Fee

**Approach**: Update status to "Completed" first, then create fee.

**Pros**:
- Marks order as processed earlier
- Slightly simpler logic

**Cons**:
- Orphaned status updates if fee creation fails
- Harder to retry failed fee creation (order already marked completed)
- Violates atomicity principle

**Rejected because**: Creates inconsistent state on fee creation failure.

## Implementation Notes

### Files to Modify/Create

1. **pkg/service/notion/task_order_log.go** - Add QueryApprovedOrders, UpdateOrderStatus
2. **pkg/service/notion/contractor_rates.go** - Add FindActiveRateByContractor
3. **pkg/service/notion/contractor_fees.go** - Add CheckFeeExistsByTaskOrder, CreateContractorFee
4. **pkg/handler/notion/contractor_fees.go** (NEW) - Handler implementation
5. **pkg/handler/notion/interface.go** - Add CreateContractorFees method
6. **pkg/routes/v1.go** - Register cronjob route

### Testing Requirements

- Unit tests for each service method
- Handler test with mocked services
- Integration test with test Notion database
- Manual testing with production data

### Rollback Plan

If issues are discovered post-deployment:
1. Comment out route in `pkg/routes/v1.go`
2. Deploy updated code
3. Cronjob will stop running
4. Service methods are additive (no breaking changes to existing code)

## References

- Requirements: `docs/sessions/202512311020-cronjob-create-contractor-fees/requirements/README.md`
- Task Order Log Schema: `docs/specs/notion/schema/task-order-log.md`
- Contractor Fees Schema: `docs/specs/notion/contractor-fees.md`
- Existing Pattern: `pkg/handler/notion/task_order_log.go` (SyncTaskOrderLogs)
