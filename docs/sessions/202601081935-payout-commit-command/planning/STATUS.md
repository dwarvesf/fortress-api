# Planning Phase Status

## Session Information

- **Session ID**: 202601081935-payout-commit-command
- **Feature**: Discord command for committing contractor payables
- **Date**: 2026-01-08
- **Status**: COMPLETE

## Overview

Planning for the `?payout commit <month> <batch>` Discord command implementation across two repositories:
- **fortress-api**: API endpoints, business logic, Notion services
- **fortress-discord**: Discord command handling, user interaction

## Documents Created

### Architecture Decision Records

1. **ADR-001-cascade-status-update.md**
   - Decision: Sequential best-effort update strategy
   - Rationale: Notion API limitations, no transaction support
   - Approach: Continue-on-error with detailed logging
   - Trade-offs: Eventual consistency vs. strong consistency

### Specifications

1. **01-api-endpoints.md**
   - Handler layer: PreviewCommit, Commit endpoints
   - Controller layer: Business logic with cascade updates
   - Routes: Registration with permission middleware
   - Request/response models
   - Error handling patterns

2. **02-notion-services.md**
   - ContractorPayablesService: 3 new methods
   - ContractorPayoutsService: 2 new methods
   - InvoiceSplitService: 1 new method
   - RefundRequestsService: 1 new method
   - Property type mappings (Status vs. Select)
   - Query and update patterns

3. **03-discord-command.md**
   - Command structure (payout/commit subcommand)
   - Service layer for API calls
   - View layer for Discord embeds
   - Adapter layer for HTTP communication
   - Button interaction handling
   - Permission checking

## Key Technical Decisions

### Cascade Update Strategy

**Decision**: Sequential best-effort updates with continue-on-error

**Update Sequence**:
1. For each Contractor Payable:
   - Get linked Payout Items
   - For each Payout Item:
     - Update Invoice Split (if linked) → Paid
     - Update Refund Request (if linked) → Paid
     - Update Payout Item → Paid
   - Update Contractor Payable → Paid + Payment Date

**Error Handling**: Track success/failure counts, log all operations, return summary

### PayDay Filtering

**Decision**: Query all pending payables, then filter by PayDay in-memory

**Approach**:
1. Query all Pending payables for period
2. For each payable, fetch contractor's Service Rate
3. Extract PayDay from Service Rate
4. Filter to match batch parameter (1 or 15)

### API Design

**Endpoints**:
- `GET /api/v1/contractor-payables/preview-commit?month=YYYY-MM&batch=1|15`
- `POST /api/v1/contractor-payables/commit` with JSON body

**Permissions**:
- Preview: Requires `PayrollsRead`
- Commit: Requires `PayrollsCreate`

## Implementation Summary

### fortress-api Changes

**New Files** (15 files):
```
pkg/handler/contractorpayables/
  - interface.go
  - contractorpayables.go
  - request.go

pkg/controller/contractorpayables/
  - interface.go
  - contractorpayables.go
```

**Modified Files** (5 files):
```
pkg/service/notion/contractor_payables.go   (+3 methods)
pkg/service/notion/contractor_payouts.go    (+2 methods)
pkg/service/notion/invoice_split.go         (+1 method)
pkg/service/notion/refund_requests.go       (+1 method)
pkg/routes/v1.go                            (register routes)
```

### fortress-discord Changes

**New Files** (10 files):
```
pkg/discord/command/payout/
  - command.go
  - new.go

pkg/discord/service/payout/
  - service.go
  - interface.go

pkg/discord/view/payout/
  - payout.go

pkg/adapter/fortress/
  - payout.go

pkg/model/
  - payout.go
```

**Modified Files** (2 files):
```
pkg/discord/command/command.go     (register command)
pkg/adapter/fortress/fortress.go   (add payout adapter)
```

## Implementation Patterns

### Existing Patterns Followed

1. **Layered Architecture**: Routes → Handler → Controller → Service
2. **Notion Service Pattern**: Query/update methods with DEBUG logging
3. **Discord Command Pattern**: Command → Service → View → Adapter
4. **Error Handling**: Wrap errors with context, log before returning
5. **Permissions**: Middleware-based permission checking

### Code Style Conventions

- Extensive DEBUG logging with page IDs and context
- Structured logging with logger.Fields
- Error wrapping with fmt.Errorf
- Gin framework for HTTP handlers
- discordgo library for Discord interactions

## Database Schema References

### Notion Databases

| Database | ID | Key Properties |
|----------|-----|----------------|
| Contractor Payables | `2c264b29-b84c-8037-807c-000bf6d0792c` | Payment Status (Status), Period, Payment Date |
| Contractor Payouts | `2c564b29-b84c-8045-80ee-000bee2e3669` | Status (Status), 01 Refund, 02 Invoice Split |
| Invoice Split | `2c364b29-b84c-804f-9856-000b58702dea` | Status (Select) |
| Refund Request | `2cc64b29-b84c-8066-adf2-cc56171cedf4` | Status (Status) |
| Service Rate | `2c464b29-b84c-80cf-bef6-000b42bce15e` | PayDay (Select) |

### Property Type Mapping

**CRITICAL**: Different databases use different property types:
- Contractor Payables: `Payment Status` → Status type
- Contractor Payouts: `Status` → Status type
- Invoice Split: `Status` → **Select type** (different!)
- Refund Request: `Status` → Status type

## User Flow

```
User enters: ?payout commit 2025-01 15
         |
         v
[Command validates args]
         |
         v
[Service calls API preview endpoint]
         |
         v
[View displays confirmation embed with:]
  - Count: 3 payables
  - Total: $15,000.00
  - Contractor list
  - [Confirm] [Cancel] buttons
         |
         v
User clicks [Confirm]
         |
         v
[Service calls API commit endpoint]
         |
         v
[Controller executes cascade updates:]
  - Update Invoice Splits
  - Update Refunds
  - Update Payouts
  - Update Payables
         |
         v
[View displays result:]
  - Success: "Committed 3 payables"
  - Partial: "Committed 2/3 (1 failed)"
```

## Testing Strategy

### Unit Tests

**API Layer**:
- Handler request validation
- Controller business logic
- Service query/update methods
- Mock Notion client

**Discord Layer**:
- Command argument parsing
- Service API calls (mock adapter)
- View embed formatting
- Permission checking

### Integration Tests

**API Layer**:
- Full preview and commit flow
- Test with Notion test databases
- Idempotency (re-running commit)
- Partial failure scenarios

**Discord Layer**:
- Full command flow
- Button interactions
- Error handling
- Empty result handling

### Manual Testing

- [ ] Preview with no payables
- [ ] Preview with multiple payables
- [ ] Commit with all success
- [ ] Commit with partial failure
- [ ] Invalid month format
- [ ] Invalid batch value
- [ ] Permission check
- [ ] Button interactions

## Risk Assessment

### High Risk

1. **Notion Property Types**: Using wrong type (Status vs. Select) will cause updates to fail
   - **Mitigation**: Clearly documented in specs, validation in tests

2. **Partial Failures**: Some records updated, others not
   - **Mitigation**: Detailed logging, track success/failure counts, idempotent design

### Medium Risk

1. **PayDay Query Performance**: Querying Service Rate for each payable
   - **Mitigation**: Consider caching PayDay values (future optimization)

2. **API Rate Limits**: Notion API has rate limits
   - **Mitigation**: Add retry logic with exponential backoff (future enhancement)

### Low Risk

1. **Button Expiration**: Discord buttons expire after 15 minutes
   - **Mitigation**: Document this behavior, acceptable UX trade-off

2. **Embed Size Limits**: Discord embeds have size limits
   - **Mitigation**: Limit contractor list to 10 entries, show "... and N more"

## Dependencies

### External APIs

- Notion API (dstotijn/go-notion client)
- Discord API (bwmarrin/discordgo client)

### Internal Dependencies

- fortress-api config (database IDs, permissions)
- fortress-discord config (API URL, API key)

### Permissions

- Discord: Admin or ops role required
- API: PayrollsRead (preview), PayrollsCreate (commit)

## Success Criteria

### Functional Requirements

- [x] Command accepts month (YYYY-MM) and batch (1 or 15)
- [x] Preview shows count, total, contractor list
- [x] Confirmation flow with buttons
- [x] Cascade updates across 4 tables
- [x] Error handling with partial failure support
- [x] Permission checking

### Non-Functional Requirements

- [x] Detailed DEBUG logging throughout
- [x] Idempotent operation (re-running is safe)
- [x] User-friendly error messages
- [x] Clear documentation
- [x] Follows existing code patterns

## Next Steps

### Implementation Phase

1. **Create API endpoints** (fortress-api)
   - Implement handler layer
   - Implement controller layer
   - Register routes

2. **Extend Notion services** (fortress-api)
   - Add query methods
   - Add update methods
   - Test with Notion test database

3. **Create Discord command** (fortress-discord)
   - Implement command layer
   - Implement service layer
   - Implement view layer
   - Implement adapter layer

4. **Integration**
   - Test full flow
   - Handle button interactions
   - Test error scenarios

5. **Documentation**
   - Update user documentation
   - Add operation runbook
   - Document troubleshooting

### Handoff to Implementation

All specifications are ready for developers to implement. Each spec includes:
- Detailed method signatures
- Implementation notes
- Error handling patterns
- Logging requirements
- Testing considerations

## References

- **Requirements**: `/docs/sessions/202601081935-payout-commit-command/requirements/requirements.md`
- **Spec**: `/docs/specs/payout-commit-command.md`
- **ADR-001**: `./ADRs/ADR-001-cascade-status-update.md`
- **API Spec**: `./specifications/01-api-endpoints.md`
- **Notion Services Spec**: `./specifications/02-notion-services.md`
- **Discord Command Spec**: `./specifications/03-discord-command.md`

## Planning Phase Complete

All planning documents have been created. The specifications are comprehensive and ready for implementation by the development team.
