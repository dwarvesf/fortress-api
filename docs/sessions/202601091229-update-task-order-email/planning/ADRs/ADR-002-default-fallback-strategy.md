# ADR-002: Default Fallback Strategy for Missing Payday Data

## Status
Accepted

## Context
The updated task order confirmation email requires a dynamic invoice due date based on contractor Payday. However, there are scenarios where Payday data may be unavailable or invalid:

### Failure Scenarios
1. **Service Rate Not Found**: Contractor has no active Service Rate record
2. **Missing Payday Property**: Service Rate record exists but Payday field is empty
3. **Invalid Payday Value**: Payday contains unexpected value (not "01" or "15")
4. **Query Failure**: Notion API error during Service Rate query
5. **Network Issues**: Temporary connectivity problems with Notion API

### Business Requirements
From the approved requirements:
- Email must always send successfully
- Missing Payday data should not block email delivery
- Provide reasonable default for contractors without configured Payday
- Maintain delivery success rate >95%

### Stakeholder Expectations
- Operations team expects reliable email delivery even with incomplete data
- Contractors expect timely receipt of work order confirmations
- Leadership prefers graceful degradation over hard failures

## Decision
**When Payday data is unavailable or invalid, the system will default to "10th" as the invoice due date.**

This applies to all failure scenarios:
- Payday fetch returns error → Use "10th"
- Payday value is 0 or nil → Use "10th"
- Payday value is invalid (not 1 or 15) → Use "10th"
- Service Rate query fails → Use "10th"

### Business Rule Mapping
```
IF Payday = 1 THEN DueDate = "10th"
IF Payday = 15 THEN DueDate = "25th"
IF Payday = 0 OR Payday = null OR Payday = invalid THEN DueDate = "10th" (DEFAULT)
```

## Rationale

### Why "10th" as Default
1. **Earlier Deadline**: 10th is the earlier of the two deadlines, encouraging prompt invoice submission
2. **Majority Case**: Most contractors (Payday=1) submit by the 10th, making it the common case
3. **Business Safety**: Earlier deadline reduces risk of late payment processing
4. **Conservative Approach**: Better to request earlier submission than later

### Why Not "25th" as Default
- Would encourage delayed invoice submission
- Less conservative from business operations perspective
- Doesn't align with majority contractor schedule

### Why Not "15th" (Middle Ground)
- Not a valid deadline in current business rules (only 10th or 25th)
- Would confuse contractors with arbitrary deadline
- Doesn't match either established payment schedule

### Why Not Block Email
- Violates "email must always send" requirement
- Creates operational burden for manual follow-up
- Reduces system reliability and user trust
- Email purpose is broader than just invoice due date

### Why Not Generic Message
- Email loses value by omitting specific deadline
- Contractors may not know when to submit invoice
- Defeats purpose of invoice reminder section
- "10th" is better than no deadline information

## Consequences

### Positive
1. **High Reliability**: Email delivery not dependent on Service Rate data quality
2. **Graceful Degradation**: System continues to function with incomplete data
3. **Operational Simplicity**: No manual intervention required for missing data
4. **Conservative Approach**: Earlier deadline protects business interests
5. **Clear Behavior**: Deterministic fallback logic easy to understand and test

### Negative
1. **Incorrect Deadline Risk**: Contractors with Payday=15 might receive wrong deadline if data missing
2. **Data Quality Masking**: Default hides underlying configuration problems
3. **User Confusion**: Contractor might receive "10th" but expect "25th" payment schedule
4. **Support Burden**: May require manual correction communication

### Mitigation Strategies

#### 1. Debug Logging
Log every fallback occurrence for monitoring:
```go
l.Debug(fmt.Sprintf("using default due date '10th' for contractor %s: payday=%d (fetch failed or missing)", contractorName, payday))
```

#### 2. Metrics Tracking
Monitor fallback usage rate:
- Track percentage of emails using default vs. fetched Payday
- Alert if default usage exceeds 10% threshold
- Dashboard showing Payday data coverage

#### 3. Data Quality Monitoring
- Regular audits of Service Rate Payday configuration
- Automated alerts for contractors missing Payday data
- Operations team dashboard showing configuration gaps

#### 4. Documentation
- Document default behavior in email template comments
- Include in operational runbooks
- Train support team on handling questions about due dates

#### 5. Future Enhancement Path
- Phase 2: Add Payday configuration validation webhook
- Phase 3: Automated reminders to operations team for missing configurations
- Phase 4: Self-service Payday configuration for contractors

## Implementation Notes

### Error Handling Pattern
```go
// Fetch Payday with graceful fallback
payday, err := contractorPayablesService.GetContractorPayDay(ctx, contractorPageID)
if err != nil {
    l.Debug(fmt.Sprintf("payday fetch failed for %s: %v - using default", contractorName, err))
    payday = 0 // 0 triggers default "10th" logic
}

// Calculate due date with fallback
var invoiceDueDay string
if payday == 15 {
    invoiceDueDay = "25th"
} else {
    // Default to "10th" for payday=1, payday=0, or any invalid value
    invoiceDueDay = "10th"
    if payday == 0 {
        l.Debug(fmt.Sprintf("using default due date for %s: no valid payday configured", contractorName))
    }
}
```

### Logging Levels
- **Debug**: Payday fetch success with value
- **Debug**: Fallback usage with reason
- **Error**: Only if email send fails (not for Payday fetch failures)

### Testing Requirements
Unit tests must cover:
1. Payday=1 → "10th"
2. Payday=15 → "25th"
3. Payday=0 → "10th" (fallback)
4. Payday fetch error → "10th" (fallback)
5. Invalid Payday value → "10th" (fallback)

## Monitoring and Alerts

### Success Metrics
- Email delivery success rate >95%
- Payday fetch success rate >90%
- Default fallback usage <10%

### Alert Conditions
- Default fallback usage >10% for 24 hours
- Email delivery success rate <95%
- Payday fetch error rate >10%

### Dashboard Metrics
- Total emails sent per day
- Payday fetch success rate
- Due date distribution ("10th" vs "25th")
- Default fallback count

## Rollback Considerations
If default fallback causes issues:
1. **Quick Fix**: Change default to "25th" (one-line code change)
2. **Alternative**: Send generic message without specific due date
3. **Full Rollback**: Revert to old email template

Default fallback behavior is isolated to invoice due date calculation, making it easy to modify without affecting other email functionality.

## References
- Requirements: `docs/sessions/202601091229-update-task-order-email/requirements/overview.md` (Error Handling section)
- Approved Plan: `/Users/quang/.claude/plans/glistening-roaming-fiddle.md` (Error Handling)
- Related: ADR-001 (Payday Data Source Selection)

## Related Decisions
- ADR-001: Payday Data Source Selection (source of Payday data)
- SPEC-002: Payday Fetching Service (implementation of fetch with fallback)
- SPEC-003: Invoice Due Date Calculation (business logic for date mapping)
