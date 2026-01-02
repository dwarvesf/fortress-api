# Planning Status: Email Monthly Task Order Confirmation

## Status: Complete

**Date**: 2026-01-02

## Summary

Planning phase complete for the Email Monthly Task Order Confirmation feature. All architectural decisions have been documented and detailed specifications have been created.

## Completed Deliverables

### 1. Architecture Decision Record (ADR)

**File**: `ADRs/001-email-task-order-confirmation.md`

**Key Decisions**:
- Data Source: Query Deployment Tracker directly for active deployments
- Email Service: Use Gmail API with accounting refresh token (accounting@d.foundation)
- Service Layer: Extend TaskOrderLogService with new methods
- Month Parameter: Default to current month if not provided
- Filtering: Support optional Discord username filter for testing
- Email Format: MIME template with Vietnamese/English hybrid content
- Error Handling: Continue-on-error pattern with detailed per-contractor results
- Grouping: One email per contractor aggregating all clients

**Alternatives Considered**:
- Store email history in database (deferred for MVP)
- Query Task Order Log instead of Deployment Tracker (rejected)
- Send separate email per client assignment (rejected)
- Use SendGrid instead of Gmail (rejected - user requested Gmail with accounting token)

### 2. Technical Specification

**File**: `specifications/send-task-order-confirmation.md`

**Specifications Include**:

#### API Contract
- Endpoint: `POST /api/v1/cronjobs/send-task-order-confirmation`
- Query Parameters: `month` (optional), `discord` (optional)
- Response format with detailed per-contractor results
- Error scenarios and status codes

#### Service Layer
- `TaskOrderLogService.QueryActiveDeploymentsByMonth()` - Query active deployments from Notion
- `TaskOrderLogService.getClientInfo()` - Extract client name and country from project
- Data structures: `DeploymentData`, `ContractorWithClients`, `ClientInfo`

#### Handler Implementation
- Parameter parsing and validation
- Deployment query and grouping by contractor
- Client information extraction
- Email generation and sending
- Result aggregation and response

#### Email Template
- Subject: "Monthly Task Order – [Tháng/Năm]"
- Body: Contractor name, period, client list with headquarters, confirmation request
- Helper functions: `formatMonthVietnamese()`, `calculatePeriodEndDate()`, `generateTaskOrderEmail()`

#### Supporting Specifications
- Error scenarios and handling
- Configuration requirements
- Logging strategy
- Test scenarios (unit and manual)
- Route configuration
- Dependencies

## Technical Approach

### Data Flow

```
1. Parse Parameters (month, discord)
   ↓
2. Query Active Deployments (Deployment Tracker)
   ↓
3. Group Deployments by Contractor
   ↓
4. For Each Contractor:
   a. Extract contractor info (name, email, discord)
   b. Extract client info from deployments
   c. Generate email HTML
   d. Send email via SendGrid
   ↓
5. Aggregate Results and Return Response
```

### Key Components

**New Service Methods**:
- `QueryActiveDeploymentsByMonth(ctx, month, discord)` - Get active deployments
- `getClientInfo(ctx, projectPageID)` - Fetch client from project

**New Handler**:
- `pkg/handler/notion/task_order_confirmation.go`
- Method: `SendTaskOrderConfirmation(c *gin.Context)`

**Email Template**:
- HTML format with contractor name, month/year, client list
- Vietnamese month format: "Tháng MM/YYYY"
- Confirmation request included

### Files to Create/Modify

**New Files**:
1. `pkg/templates/taskOrderConfirmation.tpl` - MIME email template

**Modified Files**:
1. `pkg/service/notion/task_order_log.go` - Add new query methods
2. `pkg/model/email.go` - Add TaskOrderConfirmationEmail model
3. `pkg/service/googlemail/google_mail.go` - Add SendTaskOrderConfirmationMail method
4. `pkg/service/googlemail/utils.go` - Add template composition functions
5. `pkg/service/googlemail/interface.go` - Add method to interface
6. `pkg/handler/notion/task_order_log.go` - Add SendTaskOrderConfirmation handler
7. `pkg/handler/notion/interface.go` - Add method to IHandler interface
8. `pkg/routes/v1.go` - Register cronjob route

## Dependencies

### External Services
- Notion API (Deployment Tracker, Contractor, Project, Client databases)
- Gmail API (email delivery via accounting refresh token)

### Internal Services
- TaskOrderLogService (Notion queries)
- GoogleMail Service (email sending via accounting@d.foundation)

### Configuration Required
- `NOTION_DATABASE_DEPLOYMENT_TRACKER` - Deployment Tracker database ID
- `Google.AccountingGoogleRefreshToken` - Gmail OAuth2 refresh token for accounting
- `Google.AccountingEmailID` - Gmail user ID for sending emails
- `Invoice.TemplatePath` - Path to email templates directory

## Risk Assessment

### Low Risk
- Gmail integration already proven and working (same pattern as SendPayrollPaidMail)
- Notion query patterns well-established in codebase
- Continue-on-error pattern prevents total failures

### Medium Risk
- Data quality in Deployment Tracker (contractor email, client info)
- Notion API rate limits for large contractor base
- Email deliverability (bounces, spam filters)

### Mitigation
- Skip contractors with missing data, log warnings
- Sequential processing to respect rate limits
- Gmail handles bounce tracking via accounting email
- Discord filter enables testing before mass send

## Next Steps

### Test Case Design Phase
- [ ] Create unit test specifications for service methods
- [ ] Create integration test plan for handler
- [ ] Define test data fixtures for Notion mocking
- [ ] Create golden file test cases for email generation

### Implementation Phase
- [ ] Implement service methods in TaskOrderLogService
- [ ] Create handler in task_order_confirmation.go
- [ ] Update interface and routes
- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Manual testing with test contractor

## Questions Resolved

1. **Q**: Should we store email history in database?
   - **A**: Not for MVP. Cronjob response + logs sufficient. Can add later if needed.

2. **Q**: Should we query Task Order Log or Deployment Tracker?
   - **A**: Deployment Tracker - it's the source of truth for current/future assignments.

3. **Q**: One email per client or one email per contractor?
   - **A**: One consolidated email per contractor with all clients listed.

4. **Q**: How to handle missing contractor email?
   - **A**: Skip with warning, log for manual follow-up, continue processing others.

5. **Q**: Should month parameter be required?
   - **A**: Optional - default to current month for scheduled cronjob convenience.

## Notes

- Email template follows format from requirements specification
- Vietnamese month format ("Tháng MM/YYYY") matches requirement
- Client headquarters/country emphasized for tax/compliance purposes
- Confirmation request included as specified
- Discord filter enables safe testing before production use
- Continue-on-error ensures partial success scenarios handled gracefully

## References

- Requirements: `../requirements/requirements.md`
- ADR: `ADRs/001-email-task-order-confirmation.md`
- Specification: `specifications/send-task-order-confirmation.md`
- Similar Pattern: `pkg/handler/notion/task_order_log.go` (SyncTaskOrderLogs)
- GoogleMail Service: `pkg/service/googlemail/google_mail.go`
- Email Template Pattern: `pkg/templates/paidPayroll.tpl`

---

**Planning Phase**: Complete
**Ready for**: Test Case Design and Implementation
