# Discord Username Filter for Contractor Payables Commit

## Discussion Summary

**Date**: 2026-01-09
**Session**: 202601081935-payout-commit-command
**Feature**: Add ability to commit payables for a specific contractor by Discord username

---

## Initial Question

**User Request**: "I want to commit for a specific contractor by Discord name, how can I do that?"

**Context**: The current implementation (Phase 2 complete) commits ALL contractors with matching PayDay for a given month/batch. Need to add filtering by Discord username.

---

## Requirements Clarification

Through discussion, the following requirements were established:

### 1. Scope
**Question**: Filter in API endpoints, Discord command, or both?
**Decision**: **Both API and Discord command**

The Discord filter will be added to the API endpoints first, then used by the Discord command in Phase 3.

### 2. Filter Mode
**Question**: Should Discord filter be optional or required?
**Decision**: **Optional - can commit all or specific contractor**

- If `discord` parameter is provided → commit only that contractor
- If `discord` parameter is omitted → commit all contractors (existing behavior)
- Maintains backward compatibility with existing API consumers

### 3. Batch Parameter Behavior
**Question**: How should batch parameter work when filtering by contractor?
**Decision**: **Keep batch parameter as validation**

- Both `batch` and `discord` parameters are required
- When Discord is specified, validate that contractor's PayDay matches the batch parameter
- If PayDay doesn't match → return 400 error
- Prevents accidentally committing with wrong batch number

---

## Current Implementation Analysis

### Existing API Endpoints

**Preview**: `GET /api/v1/contractor-payables/preview-commit`
**Commit**: `POST /api/v1/contractor-payables/commit`

**Current Parameters**:
- `month` (required): YYYY-MM format
- `batch` (required): 1 or 15 (PayDay)

**Current Behavior**:
1. Query ALL pending payables for the given month
2. Filter by contractor's PayDay matching batch
3. Preview shows contractors that would be committed
4. Commit updates all matching payables to "Paid"

### Existing Data Flow

```
Request (month, batch)
    ↓
QueryPendingPayablesByPeriod(period)
    ↓
For each payable:
    GetContractorPayDay(contractorPageID)
    If payDay == batch → include in commit
    ↓
Update payables + cascade to related entities
```

### Available Service Methods

**For Discord Lookup**:
- `ContractorRatesService.QueryRatesByDiscordAndMonth(ctx, discord, month)`
  - Searches Contractor Rates by Discord username (rollup property)
  - Returns: `ContractorRateData` with `ContractorPageID` and `ContractorName`
  - Filters by: Discord contains username AND Status = "Active"

**For Payables Query**:
- `ContractorPayablesService.QueryPendingPayablesByPeriod(ctx, period)`
  - Returns: `[]PendingPayable` with `ContractorPageID`, `ContractorName`, `Total`, etc.
  - Current filters: Payment Status = "Pending" AND Period = date

**For PayDay Lookup**:
- `ContractorPayablesService.GetContractorPayDay(ctx, contractorPageID)`
  - Returns: PayDay (1 or 15) from Service Rate database

### Data Structures

**PendingPayable** (already contains contractor info):
```go
type PendingPayable struct {
    PageID            string   // Payable page ID
    ContractorPageID  string   // Contractor relation - use for filtering!
    ContractorName    string   // Rollup from Contractor
    Total             float64
    Currency          string
    Period            string
    PayoutItemPageIDs []string
}
```

---

## Implementation Design

### Architecture Decision

**Chosen Approach**: Filter in controller layer, reuse existing service methods

**Rationale**:
- Controller already performs PayDay filtering
- Adding Discord filtering follows the same pattern
- No new service methods needed (reuse existing Discord lookup)
- Service layer stays focused on data retrieval

### New Data Flow with Discord Filter

```
Request (month, batch, discord)
    ↓
IF discord provided:
    QueryRatesByDiscordAndMonth(discord, month)
        → Get ContractorPageID
    GetContractorPayDay(contractorPageID)
        → Validate PayDay == batch
    ↓
QueryPendingPayablesByPeriod(period)
    ↓
For each payable:
    IF discord provided:
        Skip if payable.ContractorPageID != contractorPageID
    ELSE:
        GetContractorPayDay(contractorPageID)
        Skip if payDay != batch
    ↓
Update matching payables
```

### Error Handling Strategy

| Scenario | HTTP Status | Error Message | Notes |
|----------|-------------|---------------|-------|
| Discord username not found | 404 | "no contractor found with Discord username '{discord}'" | QueryRatesByDiscordAndMonth returns error |
| PayDay mismatch | 400 | "contractor {name} has PayDay {actual}, but batch {requested} was specified" | Validation fails |
| No payables found | 200 (preview), 404 (commit) | Empty list or "no pending payables" | Existing behavior |
| Service/API errors | 500 | Generic internal error | Network, Notion API failures |

---

## Implementation Plan

### Files to Modify

1. **`pkg/handler/contractorpayables/request.go`**
   - Add `Discord string` field to `PreviewCommitRequest` and `CommitRequest`
   - Optional field (no `binding:"required"`)

2. **`pkg/controller/contractorpayables/interface.go`**
   - Update `PreviewCommit(ctx, month, batch, discord)` signature
   - Update `CommitPayables(ctx, month, batch, discord)` signature
   - Add `Discord string` field to `PreviewCommitResponse` and `CommitResponse` (with `json:",omitempty"`)

3. **`pkg/controller/contractorpayables/contractorpayables.go`**
   - Implement Discord lookup and validation at method start
   - Add contractor filtering in payables loop
   - Optimize: skip PayDay lookup when Discord filter active
   - Add DEBUG logging with discord parameter

4. **`pkg/handler/contractorpayables/contractorpayables.go`**
   - Pass `req.Discord` to controller methods
   - Add discord to debug logs
   - Add HTTP status code handling for Discord errors (404, 400)
   - Update Swagger annotations

### Validation Flow

**Handler Layer**:
- Parameter binding (format validation)
- Extract discord from query params (GET) or JSON body (POST)

**Controller Layer**:
- If discord != "":
  1. Lookup contractor by Discord username → error if not found (404)
  2. Get contractor's PayDay → error if lookup fails (500)
  3. Validate PayDay == batch → error if mismatch (400)
  4. Store contractorPageID for filtering
- Query pending payables (all)
- Filter by contractorPageID (if provided) and/or PayDay

**Service Layer**:
- No changes (reuse existing methods)

### Performance Optimizations

1. **Single Discord Lookup**: Call `QueryRatesByDiscordAndMonth` once at the beginning
2. **Skip PayDay Lookup**: When Discord filter active, don't fetch PayDay for each payable (already validated)
3. **In-Memory Filtering**: Filter payables in controller (acceptable for expected volume)

---

## API Examples

### Without Discord Filter (Existing Behavior)

**Preview**:
```bash
GET /api/v1/contractor-payables/preview-commit?month=2025-01&batch=1
```

**Response**:
```json
{
  "month": "2025-01",
  "batch": 1,
  "count": 5,
  "total_amount": 15000.00,
  "contractors": [
    {"name": "John Doe", "amount": 3000.00, "currency": "USD", "payable_id": "..."},
    {"name": "Jane Smith", "amount": 5000.00, "currency": "USD", "payable_id": "..."}
  ]
}
```

**Commit**:
```bash
POST /api/v1/contractor-payables/commit
Content-Type: application/json

{
  "month": "2025-01",
  "batch": 1
}
```

### With Discord Filter (New Feature)

**Preview**:
```bash
GET /api/v1/contractor-payables/preview-commit?month=2025-01&batch=1&discord=johndoe
```

**Response** (success - PayDay matches):
```json
{
  "month": "2025-01",
  "batch": 1,
  "discord": "johndoe",
  "count": 1,
  "total_amount": 3000.00,
  "contractors": [
    {"name": "John Doe", "amount": 3000.00, "currency": "USD", "payable_id": "..."}
  ]
}
```

**Response** (error - Discord not found):
```json
{
  "error": "no contractor found with Discord username 'johndoe'"
}
```
HTTP Status: 404

**Response** (error - PayDay mismatch):
```json
{
  "error": "contractor John Doe has PayDay 15, but batch 1 was specified"
}
```
HTTP Status: 400

**Commit**:
```bash
POST /api/v1/contractor-payables/commit
Content-Type: application/json

{
  "month": "2025-01",
  "batch": 1,
  "discord": "johndoe"
}
```

---

## Testing Strategy

### Unit Tests (Controller Layer)

**Preview Tests**:
1. `TestPreviewCommit_NoDiscord` - Backward compatibility
2. `TestPreviewCommit_WithDiscord_Success` - Valid Discord, correct PayDay
3. `TestPreviewCommit_WithDiscord_NotFound` - Invalid Discord username
4. `TestPreviewCommit_WithDiscord_PayDayMismatch` - Valid Discord, wrong batch
5. `TestPreviewCommit_WithDiscord_NoPayables` - Valid Discord, no pending payables

**Commit Tests**:
6. `TestCommitPayables_NoDiscord` - Backward compatibility
7. `TestCommitPayables_WithDiscord_Success` - Successful commit for one contractor
8. `TestCommitPayables_WithDiscord_NotFound` - Discord lookup fails
9. `TestCommitPayables_WithDiscord_PayDayMismatch` - Validation fails

**Mock Services**:
- `ContractorRatesService.QueryRatesByDiscordAndMonth`
- `ContractorPayablesService.QueryPendingPayablesByPeriod`
- `ContractorPayablesService.GetContractorPayDay`

### Integration Tests (Handler Layer)

1. End-to-end preview with Discord parameter
2. End-to-end commit with Discord parameter
3. HTTP status code validation (200, 400, 404, 500)
4. Response format validation
5. Backward compatibility without Discord parameter

### Manual Testing Scenarios

1. **Happy Path**: Discord exists, PayDay matches, payables exist → Success
2. **Discord Not Found**: Invalid username → 404 error
3. **PayDay Mismatch**: Discord valid but wrong batch → 400 error
4. **No Payables**: Discord valid, correct batch, but no pending payables → Empty list
5. **Backward Compatibility**: No discord parameter → Works as before

---

## Backward Compatibility

**Guarantees**:
1. ✅ Discord parameter is optional (not required)
2. ✅ Existing API calls work unchanged (empty discord = existing behavior)
3. ✅ Response format compatible (new fields use `omitempty`)
4. ✅ No breaking changes to API contract
5. ✅ Same endpoints, same HTTP methods
6. ✅ Existing error responses unchanged

**Migration**: None required - feature is additive only

---

## Future Integration (Phase 3)

### Discord Command Usage

Once implemented, the Discord command will use this API:

**User Action**: `/payout commit` (or similar command)

**Discord Bot Flow**:
1. Get user's Discord username from interaction
2. Call API: `POST /api/v1/contractor-payables/commit`
   ```json
   {
     "month": "2025-01",
     "batch": 1,
     "discord": "<user_discord_username>"
   }
   ```
3. Show result to user
4. Handle errors (not found, PayDay mismatch, etc.)

**Benefits**:
- Contractors can commit their own payables
- No need to commit all contractors at once
- Self-service reduces admin overhead
- Clear error messages guide users

---

## Design Decisions Summary

| Decision Point | Choice | Rationale |
|----------------|--------|-----------|
| Where to filter | Controller layer | Follows existing pattern, no service changes needed |
| Discord parameter | Optional | Backward compatible, flexible for different use cases |
| Batch validation | Validate PayDay matches | Prevents errors, ensures correct batch |
| Error handling | HTTP status codes (404, 400) | Clear error types for clients |
| Service method | Reuse QueryRatesByDiscordAndMonth | Already exists, proven pattern |
| Performance | Lookup once, filter in loop | Acceptable for expected volume, simple logic |
| Response changes | Add discord field with omitempty | Shows what filter was applied, backward compatible |

---

## Implementation Checklist

- [ ] Update request structs with discord field
- [ ] Update controller interface signatures
- [ ] Update response structs with discord field
- [ ] Implement Discord lookup and validation in controller
- [ ] Implement contractor filtering in controller
- [ ] Update handler to pass discord parameter
- [ ] Add HTTP status code handling for Discord errors
- [ ] Update Swagger annotations
- [ ] Add DEBUG logging throughout
- [ ] Write unit tests (controller)
- [ ] Write integration tests (handler)
- [ ] Manual testing with real API
- [ ] Verify backward compatibility
- [ ] Update documentation

---

## References

- **Current Session**: `docs/sessions/202601081935-payout-commit-command/`
- **Implementation Plan**: `/Users/quang/.claude/plans/squishy-riding-hennessy.md`
- **Session Status**: `docs/sessions/202601081935-payout-commit-command/implementation/STATUS.md`
- **Phase 2 Completion**: Handler/Controller layer (100% complete)
- **Next Phase**: Phase 3 - fortress-discord command

---

## Notes

- This feature enables self-service payable commits for contractors via Discord
- Maintains 100% backward compatibility with existing API consumers
- No database schema changes required
- No new Notion queries needed - reuses existing service methods
- Clear separation of concerns: handler (I/O) → controller (business logic) → service (data access)
- Performance optimized: single Discord lookup, skip redundant PayDay lookups
- Comprehensive error handling with appropriate HTTP status codes
