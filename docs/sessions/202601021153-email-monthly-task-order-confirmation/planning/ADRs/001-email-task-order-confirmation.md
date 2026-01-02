# ADR 001: Email Monthly Task Order Confirmation to Contractors

## Status

Proposed

## Context

Currently, contractors are not receiving automated monthly confirmations of their work assignments (task orders) via email. This creates several operational challenges:

1. **Manual Process**: Operations team must manually notify contractors about their monthly assignments
2. **No Paper Trail**: No automated email record of work orders sent to contractors
3. **Inconsistency**: Delays or missing confirmations due to manual process
4. **Contractor Awareness**: Contractors may not have clear visibility into their upcoming month's assignments
5. **Compliance**: Email confirmations serve as work order documentation for contractor engagements

### Business Requirements

The system needs to:
- Send monthly task order confirmation emails to contractors
- Include contractor's active client assignments for the month
- Show client locations (headquarters/country) for tax/compliance purposes
- Support filtering by month and individual contractor (for testing/reruns)
- Track email sending status and provide detailed response

### Technical Context

The application has:
- **Notion Deployment Tracker** - Source of truth for contractor-project-client assignments
- **Gmail API Integration** - Existing email service (`pkg/service/googlemail`) using OAuth2 refresh tokens
- **TaskOrderLogService** - Existing service with methods for querying deployments
- **Cronjob Pattern** - Standard pattern for scheduled operations under `/cronjobs`

Key data sources:
- **Deployment Tracker** (Notion) - Contains contractor deployments with relations to:
  - Contractor (with Name, Team Email, Discord)
  - Project (with relation to Client)
  - Client (with Name, Country/Headquarters)

## Decision

We will implement a new cronjob endpoint that queries active deployments and sends confirmation emails to contractors.

### Decision 1: Data Source - Deployment Tracker

**Chosen**: Query Deployment Tracker directly for active deployments

**Rationale**:
- Deployment Tracker is the authoritative source for contractor-client assignments
- Contains all required information (contractor details, project, client)
- Active deployments represent current month's assignments
- Avoids dependency on Task Order Log (which is historical/billing-focused)

**Implementation**:
- Query Deployment Tracker with Status="Active"
- Follow Notion relation chains: Deployment → Project → Client
- Extract contractor info: Name, Team Email, Discord
- Extract client info: Name, Country

### Decision 2: Email Service - Gmail API with Accounting Refresh Token

**Chosen**: Use existing Gmail API service with accounting refresh token for email delivery

**Rationale**:
- Already integrated and configured (`pkg/service/googlemail`)
- Proven pattern in codebase (see `SendPayrollPaidMail`, `SendInvoiceMail`)
- Uses accounting@d.foundation as sender (matching existing accounting workflows)
- MIME template-based email with professional formatting
- User explicitly requested Gmail with accounting token

**Implementation**:
- Create `TaskOrderConfirmationEmail` model with contractor data
- Create MIME template `taskOrderConfirmation.tpl`
- Add `SendTaskOrderConfirmationMail` method to GoogleMail service
- Use `AccountingGoogleRefreshToken` for authentication
- Send from accounting@d.foundation alias

### Decision 3: Service Layer Design

**Chosen**: Extend TaskOrderLogService with new methods

**Rationale**:
- TaskOrderLogService already has deployment query methods
- Follows existing pattern (`GetDeploymentByContractor`, `getContractorInfo`)
- Can reuse Notion client configuration
- Logical extension of existing service responsibilities

**New Methods**:
- `QueryActiveDeploymentsByMonth(ctx, month, contractorDiscord)` - Get active deployments
- `getClientInfo(ctx, projectPageID)` - Fetch client name and country from project

### Decision 4: Month Parameter Default

**Chosen**: Default to current month if parameter not provided

**Rationale**:
- Simplifies cronjob scheduling (no parameters needed for monthly run)
- Supports manual reruns for specific months
- Matches existing cronjob patterns in codebase
- Allows testing with past/future months

**Implementation**:
```go
month := c.Query("month")
if month == "" {
    now := time.Now()
    month = now.Format("2006-01")
}
```

### Decision 5: Filtering Strategy

**Chosen**: Support optional Discord username filter

**Rationale**:
- Enables testing with specific contractor before mass send
- Allows reruns for individual contractors who didn't receive email
- Matches pattern in SyncTaskOrderLogs (contractor filter)
- Simple query parameter, no complex filtering needed

**Implementation**:
- Optional `discord` query parameter
- Filter deployments by contractor Discord username if provided
- Process all active contractors if parameter omitted

### Decision 6: Email Template Format

**Chosen**: HTML email with Vietnamese/English hybrid content

**Rationale**:
- Matches existing email format from requirements
- Clear subject line with month/year in Vietnamese
- Professional tone with client location emphasis
- Includes confirmation request for acknowledgment

**Template Structure**:
```
Subject: Monthly Task Order – [Tháng/Năm]

Body:
- Greeting with contractor name
- Month/year period
- Bulleted list of clients with headquarters
- Workflow reminder (Notion/Jira tracking)
- Confirmation request
- Professional signature
```

### Decision 7: Error Handling Pattern

**Chosen**: Continue-on-error with detailed per-contractor results

**Rationale**:
- Individual email failures shouldn't block other contractors
- Maximizes successful email deliveries
- Matches existing cronjob pattern (SyncTaskOrderLogs, CreateContractorFees)
- Provides visibility into partial failures

**Implementation**:
- Track statistics (emails_sent, emails_failed)
- Return detailed per-contractor results
- Log errors for failed emails
- Continue processing remaining contractors

### Decision 8: Grouping Strategy

**Chosen**: Group deployments by contractor, aggregate clients

**Rationale**:
- One email per contractor (not one per deployment)
- Contractor may have multiple client assignments
- Email should show all active clients in single message
- Reduces email volume and improves contractor experience

**Implementation**:
```go
// Group deployments by contractor
contractorGroups := groupDeploymentsByContractor(deployments)

// For each contractor:
for contractorID, contractorDeployments := range contractorGroups {
    // Collect all clients
    clients := extractClientsFromDeployments(contractorDeployments)

    // Send single email with all clients
    sendEmail(contractor, clients, month)
}
```

## Consequences

### Positive

1. **Automation**: Eliminates manual monthly notification process
2. **Consistency**: All active contractors receive timely confirmations
3. **Compliance**: Email trail for contractor work orders
4. **Scalability**: Handles growing number of contractors automatically
5. **Testability**: Discord filter enables safe testing before mass send
6. **Maintainability**: Follows established codebase patterns
7. **Observability**: Detailed logging and response statistics

### Negative

1. **Email Dependency**: Relies on Gmail API availability
2. **Data Quality**: Requires accurate Deployment Tracker data
3. **Notion API Calls**: Multiple API calls to follow relation chains
4. **No Retry Mechanism**: Failed emails require manual rerun
5. **Template in Code**: Email template stored in codebase (not external management)

### Mitigation Strategies

| Risk | Mitigation |
|------|------------|
| Gmail API failure | Continue-on-error pattern, detailed failure tracking |
| Missing contractor email | Skip with warning, log for manual follow-up |
| Missing client data | Skip deployment with warning, include partial data |
| Notion API rate limits | Pagination, sequential processing |
| Email bounce/reject | Gmail handles bounce tracking via accounting email |
| Wrong month sent | Month parameter validation, careful scheduling |

## Alternatives Considered

### Alternative 1: Store Email History in Database

**Approach**: Persist email sending records in fortress database

**Pros**:
- Permanent audit trail
- Can query historical sends
- Enables retry logic

**Cons**:
- Additional database schema/migration
- More complexity
- Not required for MVP

**Rejected because**: Cronjob response + logs sufficient for initial implementation. Can add persistence later if needed.

### Alternative 2: Use Notion Task Order Log

**Approach**: Query Task Order Log instead of Deployment Tracker

**Pros**:
- Already has contractor-project associations
- Billing-aligned data

**Cons**:
- Task Order Log is historical (past work)
- Not designed for future work assignments
- Missing active deployment status

**Rejected because**: Deployment Tracker is the authoritative source for current/future assignments.

### Alternative 3: Email Per Deployment

**Approach**: Send separate email for each client assignment

**Pros**:
- Simpler per-email logic
- Easier to track per-assignment

**Cons**:
- Email spam (contractor with 3 clients gets 3 emails)
- Poor user experience
- Higher SendGrid costs

**Rejected because**: Single consolidated email is clearer and less intrusive.

### Alternative 4: Use SendGrid Instead of Gmail

**Approach**: Use existing SendGrid service for email delivery

**Pros**:
- Already has `model.Email` and `SendEmail` method
- Different email infrastructure

**Cons**:
- User explicitly requested Gmail with accounting token
- Accounting emails should come from accounting@d.foundation
- Gmail patterns already established for similar emails (payroll, invoice)

**Rejected because**: User explicitly requested Gmail with accounting refresh token, matching existing accounting email patterns.

## Implementation Notes

### Files to Modify/Create

1. **pkg/service/notion/task_order_log.go** - Add QueryActiveDeploymentsByMonth, getClientInfo
2. **pkg/templates/taskOrderConfirmation.tpl** (NEW) - MIME email template
3. **pkg/model/email.go** - Add TaskOrderConfirmationEmail model
4. **pkg/service/googlemail/google_mail.go** - Add SendTaskOrderConfirmationMail method
5. **pkg/service/googlemail/utils.go** - Add template composition functions
6. **pkg/service/googlemail/interface.go** - Add method to interface
7. **pkg/handler/notion/task_order_log.go** - Add SendTaskOrderConfirmation handler
8. **pkg/handler/notion/interface.go** - Add SendTaskOrderConfirmation method
9. **pkg/routes/v1.go** - Register cronjob route

### Testing Requirements

- Unit tests for service methods (deployment query, client extraction)
- Handler test with mocked services
- Manual testing with test contractor (using discord filter)
- Dry run with current month to verify email content
- Validate email template rendering in inbox
- Verify Gmail sends from accounting@d.foundation

### Rollback Plan

If issues are discovered post-deployment:
1. Comment out route in `pkg/routes/v1.go`
2. Deploy updated code
3. Cronjob will stop executing
4. No database changes or schema impacts
5. Service methods are additive (no breaking changes)

## References

- Requirements: `docs/sessions/202601021153-email-monthly-task-order-confirmation/requirements/requirements.md`
- Existing GoogleMail Service: `pkg/service/googlemail/google_mail.go`
- Existing Email Template: `pkg/templates/paidPayroll.tpl`
- Existing Pattern: `pkg/handler/notion/task_order_log.go` (SyncTaskOrderLogs)
- Email Model: `pkg/model/email.go`
