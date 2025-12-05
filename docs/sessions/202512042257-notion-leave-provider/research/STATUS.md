# Research Status

## Status: COMPLETE (Reused from Expense Provider)

Research reused from previous session: `docs/sessions/202512041805-notion-expense-provider/research/`

## Applicable Research Documents

The following research from the expense provider session applies directly to leave provider implementation:

### 1. Notion API Patterns (`../202512041805-notion-expense-provider/research/notion-api-patterns.md`)
- Status property filtering (Status = "Approved")
- Status property updates (Status → "Approved")
- Relation + Rollup property handling
- Pagination best practices
- Error handling patterns

### 2. Technical Considerations (`../202512041805-notion-expense-provider/research/technical-considerations.md`)
- ID mapping strategy (UUID → int hash)
- Property type mapping (title, date, select, relation, rollup)
- Rollup extraction for email
- Configuration management

### 3. Existing Implementation Analysis (`../202512041805-notion-expense-provider/research/existing-implementation-analysis.md`)
- NocoDB leave service patterns (`pkg/service/nocodb/leave.go`)
- Webhook handler patterns (`pkg/handler/webhook/nocodb_leave.go`)
- OnLeaveRequest model structure

## Leave-Specific Considerations

### Differences from Expense Provider

| Aspect | Expense | Leave |
|--------|---------|-------|
| Status Flow | Pending → Approved → Paid | Pending → Approved/Rejected |
| Trigger | Payroll calculation (pull) | Webhook (push) |
| DB Storage | Only on payroll commit | On approval |
| Update Action | Mark as Paid | None (just persist) |

### Key Leave Request Properties

| Property | Type | Notes |
|----------|------|-------|
| Reason | title | Leave reason |
| Employee | relation | Links to Contractor |
| Email | rollup | From Employee relation |
| Leave Type | select | `Off`, `Remote` |
| Start Date | date | Leave start |
| End Date | date | Leave end |
| Shift | select | `Full day`, `Morning`, `Afternoon` |
| Status | select | `Pending`, `Approved`, `Rejected` |
| Approved By | relation | Links to Contractor |
| Approved at | date | Approval timestamp |

### Webhook Flow (Unlike Expense)

```
Notion → Webhook → Validate → Discord Notification
                        ↓
              Status Change to Approved
                        ↓
              Create on_leave_requests record
```

## Completed: 2024-12-04
