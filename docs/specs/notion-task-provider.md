# Notion Task Provider Specification

## Overview

Replace NocoDB as the task provider with Notion databases for managing Expense Requests and Leave Requests.

> **Scope**: This document covers **Expense Request** implementation. Leave Request will be addressed separately.

## Notion Database Schemas

### Expense Request

- **Database URL**: `https://www.notion.so/2bfb69f8f57381cba2daf06d28896390`
- **Data Source ID**: `collection://2bfb69f8-f573-81af-ae31-000b24dacac1`

| Property         | Type     | Options/Details                                      |
|------------------|----------|------------------------------------------------------|
| Title            | title    | Expense title/description                            |
| Requestor        | relation | Links to Contractor database (configured via `NOTION_CONTRACTOR_DB_ID`) |
| Email            | rollup   | Rolled up from Requestor relation                    |
| Status           | status   | Pending (to_do), Approved (in_progress), Paid (complete) |
| Expense Category | select   | Expense category options                             |
| Request Date     | date     | Date of request                                      |
| Amount           | number   | Expense amount (float)                               |
| Currency         | select   | USD, VND                                             |
| Attachments      | file     | Uploaded files/receipts                              |

**Status Flow:**
- `Pending` → `Approved` → `Paid`

**Expense Request Flow:**

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           EXPENSE REQUEST FLOW                              │
│                         (No webhook required)                               │
└─────────────────────────────────────────────────────────────────────────────┘

1. SUBMISSION & APPROVAL (Manual in Notion)
   ┌──────────────┐
   │ Contractor   │──▶ Creates expense in Notion DB (Status: Pending)
   └──────────────┘
         │
         ▼
   ┌──────────────┐
   │ Operations   │──▶ Reviews and approves (Status: Pending → Approved)
   └──────────────┘

2. PAYROLL CALCULATION (Monthly payroll run)
   ┌──────────────────────────────────────────────────────────────────────────┐
   │ Fortress API: PayrollExpenseProvider.GetAllInList()                      │
   │ - Query Notion API for expenses with Status = Approved                   │
   │ - Transform to bcModel.Todo format                                       │
   │ - Include in contractor's payroll calculation                            │
   └──────────────────────────────────────────────────────────────────────────┘

3. PAYROLL COMMIT (After payroll is committed)
   ┌──────────────────────────────────────────────────────────────────────────┐
   │ Fortress API: Commit Payroll                                             │
   │ - Extract expenses from ProjectBonusExplain                              │
   │ - Create Expense record + AccountingTransaction                          │
   │ - Update Notion: Status: Approved → Paid                                 │
   └──────────────────────────────────────────────────────────────────────────┘
```

**Payroll Calculation Details:**

Expenses are counted via `PayrollExpenseProvider` interface (`basecamp.ExpenseProvider`):

```go
type ExpenseProvider interface {
    GetAllInList(todolistID, projectID int) ([]model.Todo, error)
    GetGroups(todosetID, projectID int) ([]model.TodoGroup, error)
    GetLists(projectID, todosetID int) ([]model.TodoList, error)
}
```

1. **Fetch Phase** (during payroll calculation):
   - `GetAllInList()` queries Notion for expenses with `Status = Approved`
   - Transforms to `bcModel.Todo` format with title: `"description | amount | currency"`
   - Links employee via `Requestor` relation → `Email` rollup → `employees.team_email`

2. **Commit Phase** (when payroll is committed):
   - `extractExpenseSubmissionsFromPayroll()` extracts expenses from `ProjectBonusExplain`
   - `storeExpenseSubmissions()` creates `Expense` record + `AccountingTransaction`
   - `markExpenseSubmissionsAsCompleted()` updates Notion status to `Paid`

3. **Data Storage**:
   - `expenses` table stores: `task_provider`, `task_ref` (Notion page ID), `task_board`
   - `accounting_transactions` table stores amount with metadata

### Leave Request

- **Database URL**: `https://www.notion.so/2bfb69f8f5738101a121c4464e7a901b`
- **Data Source ID**: `collection://2bfb69f8-f573-8194-b8fe-000b7b278e8d`

| Property       | Type   | Options              | Description                |
|----------------|--------|----------------------|----------------------------|
| Start Date     | title  |                      | Leave start date (title)   |
| Name           | text   |                      | Request name/title         |
| Employee Email | email  |                      | Employee email address     |
| Employee       | text   |                      | Employee name              |
| Email          | email  |                      | Alternate email field      |
| Status         | select | Approved             | Approval status            |
| End Date       | text   |                      | Leave end date             |
| Leave Type     | select | Off                  | Type of leave              |
| Shift          | select | full day             | Full day / half day        |
| Approved By    | text   |                      | Approver email/name        |
| Approved at    | text   |                      | Approval timestamp         |
| Files          | file   |                      | Supporting documents       |

**Status Flow:**
- `Pending` → `Approved` / `Rejected`

**Leave Request Flow:**

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           LEAVE REQUEST FLOW                                │
└─────────────────────────────────────────────────────────────────────────────┘

1. SUBMISSION (Employee submits via Notion form)
   ┌──────────────┐
   │ Employee     │──▶ Creates leave request in Notion DB
   └──────────────┘    Status: Pending
         │
         ▼
   ┌──────────────────────────────────────────────────────────────────────────┐
   │ Notion Webhook → POST /webhooks/notion/leave                             │
   │ Event: page.created                                                      │
   └──────────────────────────────────────────────────────────────────────────┘
         │
         ▼
   ┌──────────────────────────────────────────────────────────────────────────┐
   │ Fortress API: Validate Submission                                        │
   │ - Lookup Employee by email                                               │
   │ - Validate date range (start <= end, not in past)                        │
   │ - Send Discord notification with Approve/Reject buttons                  │
   └──────────────────────────────────────────────────────────────────────────┘

2. APPROVAL (Manager approves via Notion or Discord button)
   ┌──────────────┐
   │ Manager      │──▶ Changes Status: Pending → Approved
   └──────────────┘    Sets Approved By, Approved at
         │
         ▼
   ┌──────────────────────────────────────────────────────────────────────────┐
   │ Notion Webhook → POST /webhooks/notion/leave                             │
   │ Event: page.updated (Status changed to Approved)                         │
   └──────────────────────────────────────────────────────────────────────────┘
         │
         ▼
   ┌──────────────────────────────────────────────────────────────────────────┐
   │ Fortress API: Handle Approval                                            │
   │ - Create on_leave_requests record in DB                                  │
   │ - Store notion_page_id for reference                                     │
   │ - Send Discord confirmation                                              │
   └──────────────────────────────────────────────────────────────────────────┘

3. REJECTION (Manager rejects)
   ┌──────────────┐
   │ Manager      │──▶ Changes Status: Pending → Rejected
   └──────────────┘
         │
         ▼
   ┌──────────────────────────────────────────────────────────────────────────┐
   │ Fortress API: Handle Rejection                                           │
   │ - If previously approved, delete on_leave_requests record                │
   │ - Send Discord notification                                              │
   └──────────────────────────────────────────────────────────────────────────┘
```

## Current Architecture

### Task Provider Interfaces

Located in `pkg/service/taskprovider/`:

```go
// ExpenseProvider - pkg/service/taskprovider/expense.go
type ExpenseProvider interface {
    Type() ProviderType
    ParseExpenseWebhook(ctx context.Context, req ExpenseWebhookRequest) (*ExpenseWebhookPayload, error)
    ValidateSubmission(ctx context.Context, payload *ExpenseWebhookPayload) (*ExpenseValidationResult, error)
    CreateExpense(ctx context.Context, payload *ExpenseWebhookPayload) (*ExpenseTaskRef, error)
    CompleteExpense(ctx context.Context, payload *ExpenseWebhookPayload) error
    UncompleteExpense(ctx context.Context, payload *ExpenseWebhookPayload) error
    DeleteExpense(ctx context.Context, payload *ExpenseWebhookPayload) error
    PostFeedback(ctx context.Context, payload *ExpenseWebhookPayload, input ExpenseFeedbackInput) error
}
```

### Provider Selection

From `pkg/service/service.go`:
- Config: `TASK_PROVIDER` environment variable
- Values: `nocodb`, `basecamp`
- Fallback: Basecamp if provider unavailable

### Existing Webhook Handlers

- `pkg/handler/webhook/nocodb_expense.go` - NocoDB expense webhooks
- `pkg/handler/webhook/nocodb_leave.go` - NocoDB leave webhooks

### Database Models

- `pkg/model/onleave_request.go` - Contains `NocodbID *int` field for tracking source

## Implementation Plan

### Phase 1: Infrastructure

1. Add Notion database config to `pkg/config/`
2. Create `pkg/service/taskprovider/notion/` package
3. Add `ProviderNotion` type constant

### Phase 2: Expense Request Provider

1. Implement `ExpenseProvider` interface for Notion
2. Create webhook handler for Notion expense webhooks
3. Map Notion properties to `ExpenseWebhookPayload`

### Phase 3: Leave Request Provider

1. Create Notion leave service
2. Implement webhook handler
3. Map Notion properties to leave request model

### Phase 4: Integration

1. Add `TASK_PROVIDER=notion` support
2. Register Notion webhook routes
3. Add `NotionPageID` field to models

## Key Differences from NocoDB

| Aspect          | NocoDB                  | Notion                     |
|-----------------|-------------------------|----------------------------|
| Record ID       | Integer (`Id`)          | UUID (page ID)             |
| Webhook Format  | Custom NocoDB format    | Notion webhook format      |
| Authentication  | `X-NocoDB-Signature`    | Notion webhook secret      |
| Status Field    | Text field              | Select property            |
| API             | REST API                | Notion API                 |

## Configuration Required

```env
# Notion Task Provider
TASK_PROVIDER=notion
NOTION_EXPENSE_DB_ID=2bfb69f8-f573-81cb-a2da-f06d28896390
NOTION_CONTRACTOR_DB_ID=2bfb69f8-f573-805a-8915-000bc44ce188
```

## References

- Current NocoDB implementation: `pkg/service/taskprovider/nocodb/`
- Basecamp implementation: `pkg/service/taskprovider/basecamp/`
- Notion service: `pkg/service/notion/`
