# Implementation Tasks

## Overview

Implementation tasks for Notion Leave Provider. Tests are skipped per user request.

## Task List

### Phase 1: Configuration

- [ ] **TASK-1**: Add Leave configuration to NotionConfig
  - File: `pkg/config/config.go`
  - Add `LeaveDBID` and `LeaveDataSourceID` fields
  - Update config loading

### Phase 2: Service Layer

- [ ] **TASK-2**: Create LeaveRequest struct
  - File: `pkg/service/notion/leave.go`
  - Define struct with all properties from Notion schema

- [ ] **TASK-3**: Implement GetLeaveRequest method
  - File: `pkg/service/notion/leave.go`
  - Fetch page by ID
  - Extract all properties using existing helpers
  - Return populated LeaveRequest struct

- [ ] **TASK-4**: Implement UpdateLeaveStatus method
  - File: `pkg/service/notion/leave.go`
  - Update Status select property
  - Set Approved By relation (if approving)
  - Set Approved at date (if approving)

- [ ] **TASK-5**: Add LeaveService to Notion service struct
  - File: `pkg/service/notion/notion.go`
  - Add LeaveService field
  - Initialize in constructor

### Phase 3: Model Updates

- [ ] **TASK-6**: Add NotionPageID to OnLeaveRequest model
  - File: `pkg/model/onleave_request.go`
  - Add `NotionPageID *string` field
  - Ensure GORM tags are correct

### Phase 4: Webhook Handler

- [ ] **TASK-7**: Create Notion leave webhook handler
  - File: `pkg/handler/webhook/notion_leave.go`
  - Implement `HandleNotionLeave` main handler
  - Parse webhook payload
  - Route to appropriate handler based on event type

- [ ] **TASK-8**: Implement handleLeaveCreated
  - File: `pkg/handler/webhook/notion_leave.go`
  - Fetch leave request from Notion
  - Validate employee exists
  - Validate date range
  - Send Discord notification with buttons

- [ ] **TASK-9**: Implement handleLeaveApproved
  - File: `pkg/handler/webhook/notion_leave.go`
  - Fetch leave request from Notion
  - Create OnLeaveRequest record in database
  - Send Discord confirmation

- [ ] **TASK-10**: Implement handleLeaveRejected
  - File: `pkg/handler/webhook/notion_leave.go`
  - Check if OnLeaveRequest exists
  - Delete if exists
  - Send Discord notification

### Phase 5: Discord Integration

- [ ] **TASK-11**: Implement Discord button handlers
  - File: `pkg/handler/webhook/notion_leave.go` or `pkg/handler/discord/`
  - Handle approve_leave button
  - Handle reject_leave button
  - Update Notion status via LeaveService

### Phase 6: Route Registration

- [ ] **TASK-12**: Register webhook route
  - File: `pkg/routes/v1.go` or webhook routes file
  - Add `POST /webhooks/notion/leave` route
  - Connect to handler

### Phase 7: Integration

- [ ] **TASK-13**: Wire up service dependencies
  - Ensure NotionService has access to LeaveService
  - Ensure handler has access to required services

## Dependencies

```
TASK-1 (config)
    │
    ▼
TASK-2, TASK-3, TASK-4, TASK-5 (service layer)
    │
    ▼
TASK-6 (model)
    │
    ▼
TASK-7, TASK-8, TASK-9, TASK-10 (webhook handler)
    │
    ▼
TASK-11 (discord integration)
    │
    ▼
TASK-12 (routes)
    │
    ▼
TASK-13 (integration)
```

## Estimated Complexity

| Task | Complexity | Notes |
|------|------------|-------|
| TASK-1 | Low | Config addition |
| TASK-2 | Low | Struct definition |
| TASK-3 | Medium | Property extraction |
| TASK-4 | Medium | Notion API update |
| TASK-5 | Low | Wiring |
| TASK-6 | Low | Model field addition |
| TASK-7 | Medium | Webhook parsing |
| TASK-8 | High | Validation + Discord |
| TASK-9 | Medium | DB persistence |
| TASK-10 | Low | Delete logic |
| TASK-11 | Medium | Button handling |
| TASK-12 | Low | Route registration |
| TASK-13 | Low | Wiring |

## Reference Files

- NocoDB leave handler: `pkg/handler/webhook/nocodb_leave.go`
- Notion expense service: `pkg/service/notion/expense.go`
- OnLeaveRequest model: `pkg/model/onleave_request.go`
- Existing Notion client: `pkg/service/notion/notion.go`
