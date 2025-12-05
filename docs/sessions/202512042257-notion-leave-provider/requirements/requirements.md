# Notion Leave Provider Requirements

## Overview

Implement Notion as the leave request provider to replace NocoDB for managing Leave Requests.

## Source Reference

- Spec: `docs/specs/notion-task-provider.md`
- Database URL: `https://www.notion.so/2bfb69f8f5738101a121c4464e7a901b`
- Data Source ID: `collection://2bfb69f8-f573-8194-b8fe-000b7b278e8d`

## Notion Leave Request Schema

| Property     | Type     | Options/Details                          |
|--------------|----------|------------------------------------------|
| Reason       | title    | Leave reason (title field)               |
| Employee     | relation | Links to Contractor database             |
| Email        | rollup   | Rolled up from Employee relation         |
| Leave Type   | select   | `Off`, `Remote`                          |
| Start Date   | date     | Leave start date                         |
| End Date     | date     | Leave end date                           |
| Shift        | select   | `Full day`, `Morning`, `Afternoon`       |
| Status       | select   | `Pending`, `Approved`, `Rejected`                    |
| Approved By  | relation | Links to Contractor database (approver)  |
| Approved at  | date     | Approval timestamp                       |

## Functional Requirements

### FR-1: Leave Request Validation (on create)
- When a new leave request is created in Notion (Status: Pending)
- Validate employee exists by email (from Email rollup)
- Validate date range (start <= end, start not in past)
- Send Discord notification with Approve/Reject buttons to onleave channel

### FR-2: Leave Request Approval
- When Status changes from Pending to Approved in Notion
- Create `on_leave_requests` record in database
- Store `notion_page_id` for reference
- Send Discord confirmation notification

### FR-3: Leave Request Rejection
- When Status changes to Rejected in Notion
- If previously approved, delete `on_leave_requests` record from database
- Send Discord rejection notification

### FR-4: Discord Button Interaction
- Approve button updates Notion status to Approved and sets Approved By, Approved at
- Reject button updates Notion status to Rejected

## Non-Functional Requirements

### NFR-1: Provider Selection
- Use `TASK_PROVIDER=notion` environment variable
- Fallback gracefully if Notion not configured

### NFR-2: API Version
- Use Notion API version 2025-09-03 for data source queries (multi-source database support)

### NFR-3: Logging
- Debug logging for all operations to help trace issues

## Configuration Required

```env
TASK_PROVIDER=notion
NOTION_LEAVE_DB_ID=2bfb69f8f5738101a121c4464e7a901b
NOTION_LEAVE_DATA_SOURCE_ID=2bfb69f8-f573-8194-b8fe-000b7b278e8d
```

## Out of Scope

- Webhook implementation (Notion webhooks require separate setup)
- Migration from NocoDB data
- Test implementation (as per user request)
