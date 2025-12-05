# ADR-001: Notion Leave Provider Architecture

## Status
Accepted

## Context

The application currently uses NocoDB as the leave request provider via `pkg/service/nocodb/leave.go` and `pkg/handler/webhook/nocodb_leave.go`. We need to migrate to Notion to unify task management on a single platform.

Key constraints:
- Expense provider already implemented for Notion (reuse patterns)
- Leave requests require webhook-driven flow (unlike expense which is pull-based)
- Must support Discord approve/reject buttons
- Must persist approved requests to `on_leave_requests` table

## Decision

### 1. Service Layer Architecture

Create `pkg/service/notion/leave.go` following the same patterns as expense provider:

```go
type LeaveService struct {
    client    *notion.Client
    config    *config.Config
    store     *store.Store
    logger    *zap.Logger
}

// Core methods
func (s *LeaveService) GetLeaveRequest(ctx context.Context, pageID string) (*LeaveRequest, error)
func (s *LeaveService) UpdateLeaveStatus(ctx context.Context, pageID string, status string, approverEmail string) error
```

### 2. Webhook Handler

Create `pkg/handler/webhook/notion_leave.go` with handler methods:

```go
func (h *handler) HandleNotionLeave(c *gin.Context)
func (h *handler) handleLeaveCreated(ctx context.Context, pageID string) error
func (h *handler) handleLeaveUpdated(ctx context.Context, pageID string) error
```

### 3. Property Mapping

Map Notion properties to internal LeaveRequest struct:

| Notion Property | Go Field | Extraction Method |
|-----------------|----------|-------------------|
| Reason | Title | `extractTitleText()` |
| Employee | EmployeeRelation | `extractRelationIDs()` |
| Email | Email | `extractRollupText()` |
| Leave Type | LeaveType | `extractSelectName()` |
| Start Date | StartDate | `extractDate()` |
| End Date | EndDate | `extractDate()` |
| Shift | Shift | `extractSelectName()` |
| Status | Status | `extractSelectName()` |
| Approved By | ApprovedBy | `extractRelationIDs()` |
| Approved at | ApprovedAt | `extractDate()` |

### 4. Webhook Flow

```
Notion Webhook → POST /webhooks/notion/leave
                        │
                        ▼
                Parse webhook payload
                        │
        ┌───────────────┴───────────────┐
        │                               │
        ▼                               ▼
  page.created                    page.updated
        │                               │
        ▼                               │
  Validate submission              Check status change
        │                               │
        ▼                       ┌───────┴───────┐
  Send Discord notification     │               │
  (Approve/Reject buttons)      ▼               ▼
                          Status=Approved  Status=Rejected
                                │               │
                                ▼               ▼
                          Create record    Delete record
                          in DB            from DB (if exists)
                                │               │
                                ▼               ▼
                          Discord confirm  Discord notify
```

### 5. ID Mapping Strategy

Use hash-based mapping (same as expense provider):
- `NotionPageID` (UUID string) stored in model
- Hash to int for backward compatibility with existing code paths

### 6. Configuration

Add to existing config:
```go
type NotionConfig struct {
    // ... existing fields
    LeaveDBID         string `env:"NOTION_LEAVE_DB_ID"`
    LeaveDataSourceID string `env:"NOTION_LEAVE_DATA_SOURCE_ID"`
}
```

## Consequences

### Positive
- Consistent architecture with expense provider
- Reuses proven Notion API patterns
- Single source of truth for leave requests

### Negative
- Webhook setup required in Notion (manual configuration)
- No automatic migration of existing NocoDB data

### Neutral
- NocoDB leave handler remains for potential rollback
