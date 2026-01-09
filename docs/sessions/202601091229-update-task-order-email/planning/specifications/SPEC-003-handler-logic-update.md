# SPEC-003: Handler Logic Update

**Status**: Ready for Implementation
**Priority**: High
**Estimated Effort**: 1.5 hours
**Dependencies**: SPEC-001, SPEC-002

## Overview

Update the task order confirmation handler to fetch Payday data, calculate invoice due dates, create milestone data, and populate the new email model fields.

## Objectives

1. Integrate Payday fetching service call
2. Calculate invoice due day based on Payday value
3. Create mock milestone data for initial release
4. Populate new email model fields
5. Maintain existing email sending functionality

## Technical Details

### File to Modify

`pkg/handler/notion/task_order_log.go`

### Method to Update

`SendTaskOrderConfirmation` (approximately line 493)

### Current Logic Flow

```go
func (h *handler) SendTaskOrderConfirmation(c *gin.Context) {
    // 1. Parse query parameters
    // 2. Fetch task order log from Notion
    // 3. Extract contractor and project data
    // 4. Build TaskOrderConfirmationEmail struct
    // 5. Send email via GoogleMail service
    // 6. Return success response
}
```

### Required Changes

Add logic between step 3 and 4 to:
1. Fetch contractor Payday from Service Rate
2. Calculate invoice due day
3. Create milestone array
4. Add new fields to email struct

### Implementation Details

#### 1. Fetch Payday Value

Insert after contractor data extraction (around line 580):

```go
// Fetch contractor payday from Service Rate database
contractorPageID := contractorRelation[0].ID // Extract from existing contractor relation
payday, err := h.service.GetContractorPayday(c.Request.Context(), contractorPageID)
if err != nil {
    // Error already logged in service layer
    // Continue with default payday
    payday = 0
}

h.logger.Debug(c.Request.Context(), "fetched contractor payday",
    "contractor_id", contractorPageID,
    "payday", payday,
)
```

**Key Points**:
- Extract `contractorPageID` from existing contractor relation
- Call service method from SPEC-002
- Log the result for monitoring
- Continue execution even if fetch fails (graceful fallback)

#### 2. Calculate Invoice Due Day

```go
// Calculate invoice due day based on payday
invoiceDueDay := "10th" // Default for Payday 1
if payday == 15 {
    invoiceDueDay = "25th"
}

h.logger.Debug(c.Request.Context(), "calculated invoice due day",
    "payday", payday,
    "invoice_due_day", invoiceDueDay,
)
```

**Logic Table**:
| Payday Value | Invoice Due Day | Scenario |
|--------------|-----------------|----------|
| 0 | "10th" | Missing/invalid/error |
| 1 | "10th" | Payday = "01" |
| 15 | "25th" | Payday = "15" |

#### 3. Create Milestone Data

```go
// TODO: Replace with real data source when available
// For now, use hardcoded milestones as placeholder
milestones := []string{
    "Completed sprint deliverables on schedule",
    "Contributed to team knowledge sharing sessions",
}

h.logger.Debug(c.Request.Context(), "prepared milestones for email",
    "milestone_count", len(milestones),
)
```

**Key Points**:
- Use hardcoded array initially
- Add TODO comment for future enhancement
- Keep milestones generic and positive
- Log count for monitoring

**Alternative Empty Array**:
```go
// If no milestones should be shown initially:
milestones := []string{}
```

#### 4. Update Email Struct Population

Locate the `TaskOrderConfirmationEmail` struct creation and add new fields:

```go
emailData := model.TaskOrderConfirmationEmail{
    ContractorName:     contractorName,
    ContractorLastName: contractorLastName,
    PeriodStartDate:    periodStartDate,
    PeriodEndDate:      periodEndDate,
    ConfirmationLink:   confirmationLink,
    // ... existing fields ...

    // New fields for invoice reminder email
    InvoiceDueDay:      invoiceDueDay,
    Milestones:         milestones,
}
```

### Code Location Reference

**Approximate Line Numbers** (may vary):
- Line 493: `SendTaskOrderConfirmation` method starts
- Line 580: After contractor relation extraction
- Line 620: `TaskOrderConfirmationEmail` struct creation

### Integration Points

**Calls**:
- `h.service.GetContractorPayday()` - From SPEC-002
- `h.logger.Debug()` - Existing logger

**Called By**:
- Cron job endpoint: `/api/v1/cronjobs/send-task-order-confirmation`
- Manual trigger via API

**Dependencies**:
- Service layer must implement `GetContractorPayday` (SPEC-002)
- Model must have new fields (SPEC-001)
- Template must support new fields (SPEC-004)

## Testing Requirements

### Unit Tests

**Test File**: `pkg/handler/notion/task_order_log_test.go`

**New Test Cases**:

1. **TestSendTaskOrderConfirmation_Payday01**
   ```go
   // Setup: Mock GetContractorPayday returns 1
   // Expected: Email struct has InvoiceDueDay = "10th"
   // Verifies: Correct due day calculation for Payday 1
   ```

2. **TestSendTaskOrderConfirmation_Payday15**
   ```go
   // Setup: Mock GetContractorPayday returns 15
   // Expected: Email struct has InvoiceDueDay = "25th"
   // Verifies: Correct due day calculation for Payday 15
   ```

3. **TestSendTaskOrderConfirmation_PaydayFallback**
   ```go
   // Setup: Mock GetContractorPayday returns 0
   // Expected: Email struct has InvoiceDueDay = "10th"
   // Verifies: Default fallback works correctly
   ```

4. **TestSendTaskOrderConfirmation_MilestonesPopulated**
   ```go
   // Setup: Standard email send
   // Expected: Email struct has Milestones array with items
   // Verifies: Milestones array is populated
   ```

**Update Existing Tests**:
- Update `TestSendTaskOrderConfirmation_Success` to mock `GetContractorPayday`
- Verify new fields don't break existing assertions

### Integration Tests

**Manual Testing**:

1. **Test Email with Payday 1 Contractor**:
   ```bash
   curl -X POST "http://localhost:8080/api/v1/cronjobs/send-task-order-confirmation?month=2026-01&discord=contractor_with_payday_1&test_email=your@email.com"
   ```
   - Verify: Email received with "10th" due date

2. **Test Email with Payday 15 Contractor**:
   ```bash
   curl -X POST "http://localhost:8080/api/v1/cronjobs/send-task-order-confirmation?month=2026-01&discord=contractor_with_payday_15&test_email=your@email.com"
   ```
   - Verify: Email received with "25th" due date

3. **Test Email with Missing Payday**:
   ```bash
   curl -X POST "http://localhost:8080/api/v1/cronjobs/send-task-order-confirmation?month=2026-01&discord=contractor_without_payday&test_email=your@email.com"
   ```
   - Verify: Email received with "10th" due date (fallback)

4. **Check Logs**:
   - Verify debug logs for Payday fetch
   - Verify debug logs for due day calculation
   - Verify no error logs for normal operation

### Test Data Requirements

Mock setup for unit tests:

```go
// Mock successful Payday fetch
mockService.EXPECT().
    GetContractorPayday(gomock.Any(), contractorPageID).
    Return(1, nil)

// Mock Payday fetch failure
mockService.EXPECT().
    GetContractorPayday(gomock.Any(), contractorPageID).
    Return(0, nil)

// Mock Payday 15
mockService.EXPECT().
    GetContractorPayday(gomock.Any(), contractorPageID).
    Return(15, nil)
```

## Acceptance Criteria

- [ ] Payday fetching integrated in handler
- [ ] Invoice due day calculated correctly for all scenarios
- [ ] Milestone array created and populated
- [ ] Email struct includes new fields
- [ ] All unit tests pass
- [ ] Integration tests successful
- [ ] Debug logging added for monitoring
- [ ] Existing email functionality unchanged
- [ ] Code review approved

## Implementation Notes

### Contractor Page ID Extraction

The contractor page ID should already be available from existing logic:

```go
// Existing code (approximately line 570)
contractorRelation := pageData.Properties["Contractor"].(*notionapi.RelationProperty).Relation
if len(contractorRelation) == 0 {
    // Handle error
}
contractorPageID := contractorRelation[0].ID
```

Reuse this `contractorPageID` for the Payday fetch.

### Error Handling Strategy

**Never block email sending**:
- If `GetContractorPayday` returns error → use default "10th"
- If `GetContractorPayday` returns 0 → use default "10th"
- If contractor page ID not found → use default "10th"
- Log all scenarios for monitoring

### Logging Best Practices

Use structured logging with context:
```go
h.logger.Debug(c.Request.Context(), "message",
    "key1", value1,
    "key2", value2,
)
```

Include in logs:
- Contractor page ID
- Payday value returned
- Calculated invoice due day
- Milestone count

### Mock Milestone Content

Use professional, generic milestones:
- "Completed sprint deliverables on schedule"
- "Contributed to team knowledge sharing sessions"
- "Maintained high code quality standards"
- "Collaborated effectively with project stakeholders"

Choose 2-3 milestones to avoid overwhelming the email.

### Code Style

Follow existing patterns in the file:
- Use existing error handling style
- Match existing logging format
- Follow naming conventions
- Maintain consistent indentation

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Service method not implemented | High | Implement SPEC-002 first |
| Contractor page ID extraction fails | Medium | Add nil checks, use default |
| Template doesn't support new fields | High | Implement SPEC-004 in parallel |
| Breaking existing tests | Medium | Update mocks carefully |
| Performance impact | Low | Single additional API call |

## Future Enhancements

### Real Milestone Data

When real milestone source is available:

1. **Add service method**:
   ```go
   milestones, err := h.service.GetClientMilestones(c.Request.Context(), projectPageID, month)
   if err != nil || len(milestones) == 0 {
       milestones = []string{} // Empty array if none
   }
   ```

2. **Update TODO comment**:
   - Remove hardcoded array
   - Call service method
   - Handle empty results gracefully

3. **Possible data sources**:
   - Notion project properties
   - Project management API
   - Configuration file per project

### Caching

Add caching for Payday values:
```go
// Check cache first
if cached, found := h.cache.GetPayday(contractorPageID); found {
    payday = cached
} else {
    payday, _ = h.service.GetContractorPayday(ctx, contractorPageID)
    h.cache.SetPayday(contractorPageID, payday, 24*time.Hour)
}
```

### Metrics

Add Prometheus metrics:
```go
metrics.PaydayFetchTotal.Inc()
metrics.PaydayFallbackTotal.WithLabelValues("missing").Inc()
metrics.InvoiceDueDayDistribution.WithLabelValues(invoiceDueDay).Inc()
```

## Documentation Updates

Update inline comments:
```go
// Fetch contractor payday from Service Rate database to calculate invoice due date
// Falls back to default (10th) if payday is not found or invalid
```

Reference ADR documents:
```go
// See ADR-001 for data source selection rationale
// See ADR-002 for fallback strategy details
```

## Related Specifications

- **SPEC-001**: Data Model Update (provides email struct fields)
- **SPEC-002**: Payday Fetching Service (provides service method)
- **SPEC-004**: Email Template Update (consumes new fields)
- **SPEC-005**: Template Function Update (supports template rendering)
- **ADR-001**: Payday Data Source Selection
- **ADR-002**: Default Fallback Strategy
- **ADR-003**: Milestone Data Approach

## Sign-off

- [ ] Implementation completed
- [ ] Unit tests pass
- [ ] Integration tests successful
- [ ] Mock service method working
- [ ] Logging verified
- [ ] Code review approved
- [ ] Documentation updated
