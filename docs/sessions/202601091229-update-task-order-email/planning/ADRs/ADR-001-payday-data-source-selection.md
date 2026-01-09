# ADR-001: Payday Data Source Selection

## Status
Accepted

## Context
The updated task order confirmation email requires a dynamic invoice due date based on each contractor's payment schedule (Payday). The email needs to display either "10th" or "25th" as the invoice submission deadline.

### Business Rule
- Contractors with Payday = 1 should submit invoices by the 10th of the month
- Contractors with Payday = 15 should submit invoices by the 25th of the month

### Available Data Sources
During requirements analysis, several potential data sources were evaluated:

1. **Service Rate Database (ContractorRates)** - Notion database containing contractor rate configurations
2. **Task Order Log Database** - Notion database tracking work order confirmations
3. **Contractor Database** - Notion database with contractor profile information
4. **New Database/Property** - Create a dedicated Payday tracking system

### Service Rate Database Schema
The ContractorRates database in Notion contains:
- **Contractor** (Relation): Link to contractor profile
- **Status** (Status): "Active", "Inactive", etc.
- **Payday** (Select): Values "01" or "15" indicating payment schedule
- **Rate** (Number): Hourly or monthly rate
- **Type** (Select): "Hourly" or "Monthly"

### Existing Implementation Precedent
The contractor_payables service already implements Payday fetching from Service Rate database:
- Method: `ContractorPayablesService.GetContractorPayDay(ctx, contractorPageID)`
- Location: `pkg/service/notion/contractor_payables.go` (lines 555-630)
- Pattern: Query with Contractor relation + Status="Active" filter
- Returns: Integer value (1 or 15) or error

## Decision
**We will use the Service Rate database (ContractorRates) as the source for Payday information.**

The implementation will:
1. Reuse the existing `GetContractorPayDay` method from ContractorPayablesService
2. Query the ContractorRates database with filters:
   - Contractor relation contains the contractor's page ID
   - Status equals "Active"
3. Extract the "Payday" Select property value
4. Convert to invoice due date using business logic

## Rationale

### Advantages of Service Rate Database
1. **Data Already Exists**: The Payday field is already configured and maintained in Service Rate database
2. **Proven Implementation**: contractor_payables service has battle-tested code for querying this data
3. **Code Reuse**: Can leverage existing `GetContractorPayDay` method with minimal modifications
4. **Single Source of Truth**: Service Rate is the canonical source for contractor payment configuration
5. **No Schema Changes**: No database modifications or new properties required
6. **Consistency**: Aligns with existing patterns in contractor_payables workflow

### Why Not Task Order Log
- Task Order Log does not contain Payday information
- Would require adding new properties and migration of data
- Task Order Log is transactional (per-month records), while Payday is a stable configuration

### Why Not Contractor Database
- Contractor database is for profile information, not payment configuration
- Service Rate already contains payment-related configuration
- Would duplicate data and create synchronization issues

### Why Not New Database
- Unnecessary complexity when data already exists
- Would require new integration code and configuration
- Increases maintenance burden

## Consequences

### Positive
1. **Fast Implementation**: Existing method can be reused with minimal adaptation
2. **Data Quality**: Service Rate database is actively maintained for payroll operations
3. **Reliability**: Proven query patterns reduce risk of bugs
4. **Maintainability**: Follows established codebase conventions
5. **Performance**: Service Rate queries are already optimized

### Negative
1. **Cross-Database Dependency**: Email service depends on Service Rate database configuration
2. **Requires Join Query**: Must query both Timesheet and Service Rate databases
3. **Data Availability Risk**: Missing Service Rate configuration will require fallback handling

### Mitigation Strategies
1. **Graceful Fallback**: Default to "10th" if Payday fetch fails (see ADR-002)
2. **Debug Logging**: Log all Payday fetch attempts for monitoring
3. **Configuration Validation**: Ensure ContractorRates database ID is configured
4. **Error Handling**: Non-blocking errors to ensure email always sends

## Implementation Notes

### Service Access Pattern
The TaskOrderLogService will need access to ContractorPayablesService to call `GetContractorPayDay`:

```go
// In handler/notion/task_order_log.go
contractorPayablesService := // injected dependency
payday, err := contractorPayablesService.GetContractorPayDay(ctx, contractorPageID)
if err != nil {
    // Log error and use default (see ADR-002)
    l.Debug(fmt.Sprintf("failed to fetch payday for contractor %s: %v", contractorPageID, err))
    payday = 0 // Will map to "10th"
}
```

### Configuration Requirements
- `NOTION_CONTRACTOR_RATES_DB_ID` must be configured (already required for contractor_payables)
- No new environment variables needed

### Data Quality Expectations
Based on contractor_payables usage:
- Expected: >90% of active contractors have configured Payday in Service Rate
- Service Rate records are maintained by operations team for payroll processing
- Data quality is monitored for contractor_payables workflow

## References
- Requirements: `docs/sessions/202601091229-update-task-order-email/requirements/overview.md`
- Approved Plan: `/Users/quang/.claude/plans/glistening-roaming-fiddle.md`
- Existing Implementation: `pkg/service/notion/contractor_payables.go` (GetContractorPayDay method)
- Configuration: `pkg/config/config.go` (NotionDatabase.ContractorRates)

## Related Decisions
- ADR-002: Default Fallback Strategy (handles missing Payday data)
- SPEC-002: Payday Fetching Service (implementation details)
