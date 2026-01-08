# ADR-001: Cascade Status Update Strategy for Payout Commit

## Status

Proposed

## Context

The `?payout commit` command needs to update the status of records across four related Notion databases when committing contractor payables:

1. **Contractor Payable**: Payment Status: Pending → Paid, Payment Date: set to now
2. **Contractor Payouts** (via "Payout Items" relation): Status: Pending → Paid
3. **Invoice Split** (via "02 Invoice Split" relation): Status: Pending → Paid
4. **Refund Request** (via "01 Refund" relation): Status: Approved → Paid

### Challenges

1. **No Transaction Support**: Notion API does not support multi-record transactions
2. **Partial Failures**: Any update operation can fail due to network issues, API limits, or invalid data
3. **Data Consistency**: Need to maintain reasonable consistency across related records
4. **User Feedback**: Need to communicate success/failure clearly
5. **Debugging**: Need visibility into which updates succeeded/failed

### Relation Structure

```
Contractor Payable
    |
    +--- Payout Items (1:N relation)
            |
            +--- Each Payout Item may have:
                    |
                    +--- 02 Invoice Split (0:1 relation) → Invoice Split record
                    |
                    +--- 01 Refund (0:1 relation) → Refund Request record
```

## Decision

We will implement a **sequential best-effort update strategy** with the following characteristics:

### Update Sequence

For each Contractor Payable being committed:

1. Query all linked Payout Items from "Payout Items" relation
2. For each Payout Item:
   - If "02 Invoice Split" relation exists: Update Invoice Split status to "Paid"
   - If "01 Refund" relation exists: Update Refund Request status to "Paid"
   - Update Payout Item status to "Paid"
3. Update Contractor Payable:
   - Set Payment Status to "Paid"
   - Set Payment Date to current timestamp

### Error Handling

- **Continue on Error**: If any individual update fails, log the error and continue with remaining updates
- **Track Results**: Maintain counters for successful and failed updates
- **Detailed Logging**: Log every update operation at DEBUG level with page IDs
- **Return Summary**: Return counts of successful and failed updates to user

### Response Format

**Success Response:**
```json
{
  "month": "2025-01",
  "batch": 15,
  "updated": 3,
  "failed": 0
}
```

**Partial Failure Response:**
```json
{
  "month": "2025-01",
  "batch": 15,
  "updated": 2,
  "failed": 1,
  "errors": [
    {
      "payable_id": "page-id-123",
      "error": "failed to update payout: network timeout"
    }
  ]
}
```

## Rationale

### Why Sequential Best-Effort?

1. **Notion API Limitations**: No support for transactions or batch updates with rollback
2. **Partial Success Value**: Committing some payables is better than all-or-nothing approach
3. **Retry Capability**: User can re-run command to process failed records (idempotent operation)
4. **Operational Reality**: In accounting workflows, partial completion with manual intervention is acceptable

### Why This Update Order?

1. **Dependency Flow**: Update leaf nodes (Invoice Split, Refund) before parent (Payout Item)
2. **Status Consistency**: Ensures child records are marked Paid before parent
3. **Audit Trail**: Payment Date on Payable is set last, serving as final confirmation

### Why Continue-on-Error?

1. **Resilience**: One failed record shouldn't block processing of valid records
2. **Visibility**: User gets immediate feedback on what succeeded/failed
3. **Debugging**: Errors are logged with context for investigation
4. **Recovery**: Failed records can be identified and retried

## Consequences

### Positive

- Simple implementation following existing Notion service patterns
- Resilient to individual record failures
- Clear feedback to users about what succeeded/failed
- Easy to debug with detailed logging
- Idempotent: re-running on partial failure is safe

### Negative

- **Inconsistent State Possible**: Some related records may be Paid while others remain Pending
- **No Atomicity**: Updates are not atomic across related records
- **Manual Intervention**: Partial failures require investigation and potential manual fixes
- **Eventual Consistency**: System relies on eventual consistency rather than strong consistency

### Mitigation Strategies

1. **Idempotency Check**: Before updating, check if record is already Paid (skip if so)
2. **Detailed Error Messages**: Include page IDs and error details in responses
3. **Retry Logic**: Allow users to re-run commit command on same month/batch
4. **Monitoring**: Log all operations for audit and debugging
5. **Documentation**: Clearly document the eventual consistency model in user docs

## Implementation Notes

### Notion API Property Types

Different Notion databases use different property types for status:

- **Contractor Payables**: `Payment Status` uses `Status` property type
- **Contractor Payouts**: `Status` uses `Status` property type
- **Invoice Split**: `Status` uses `Select` property type
- **Refund Request**: `Status` uses `Status` property type

Update code must use correct property type for each database:

```go
// Status property type
props["Status"] = nt.DatabasePageProperty{
    Status: &nt.SelectOptions{
        Name: "Paid",
    },
}

// Select property type
props["Status"] = nt.DatabasePageProperty{
    Select: &nt.SelectOptions{
        Name: "Paid",
    },
}
```

### Logging Requirements

Every update operation must log:
- Operation type (update payable, update payout, etc.)
- Page ID being updated
- Old status value (if available)
- New status value
- Success/failure result
- Error details (if failed)

Example:
```
[DEBUG] contractor_payables: updating payable pageID=abc123 status=Pending→Paid
[DEBUG] contractor_payables: updated payable pageID=abc123 successfully
```

## Alternatives Considered

### Alternative 1: All-or-Nothing with Pre-validation

**Approach**: Validate all records can be updated before starting any updates. Abort if any validation fails.

**Rejected Because**:
- Still no atomicity (Notion API limitation)
- More complex validation logic
- Delayed feedback to user
- Same inconsistent state risk if update fails after validation passes

### Alternative 2: Queue-Based Asynchronous Updates

**Approach**: Queue all updates for background processing with retry logic.

**Rejected Because**:
- Over-engineered for the use case
- Delayed feedback to user
- Adds infrastructure complexity (queue system)
- Doesn't solve fundamental consistency problem
- User still needs to check status asynchronously

### Alternative 3: Database-Backed State Machine

**Approach**: Store update state in fortress-api database, track each update operation.

**Rejected Because**:
- Adds significant complexity
- Requires new database schema
- Doesn't match existing architecture patterns
- Overkill for the problem scope
- Still doesn't provide true atomicity

## References

- Spec: `/docs/specs/payout-commit-command.md`
- Requirements: `/docs/sessions/202601081935-payout-commit-command/requirements/requirements.md`
- Notion API Docs: https://developers.notion.com/reference/page
- Existing Patterns: `pkg/service/notion/contractor_payables.go`
