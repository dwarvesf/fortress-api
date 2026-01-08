# ADR-002: Non-Blocking Error Handling

## Status
Accepted

## Context
When creating the Contractor Payables record in Notion fails, we need to decide whether to:
1. Fail the entire invoice generation request
2. Log and continue with the response

## Decision
**Non-blocking approach**: Log the error but continue with successful response.

### Rationale
- The primary deliverable (PDF invoice + Google Drive upload) is already complete
- User can manually create the Notion record if needed
- Notion API failures are transient and recoverable
- Better UX to deliver the invoice than fail entirely

### Implementation
```go
payablePageID, err := h.service.Notion.ContractorPayables.CreatePayable(ctx, input)
if err != nil {
    l.Error(err, "failed to create contractor payables record")
    // Continue - don't return error
} else {
    l.Debug(fmt.Sprintf("contractor payables record created: pageID=%s", payablePageID))
}
```

## Consequences

### Positive
- Invoice generation always succeeds if upload succeeds
- User gets their invoice even with Notion issues
- Clear error logging for debugging

### Negative
- Potential data inconsistency (invoice exists but no Notion record)
- Manual intervention may be needed

## Future Consideration
Consider adding a retry mechanism or background job for failed Notion creations.
