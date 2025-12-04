# Notion Expense Provider Requirements

## Overview

Replace NocoDB as the task provider with Notion databases for managing Expense Requests during payroll calculation.

## Scope

- **In Scope**: Expense Request implementation
- **Out of Scope**: Leave Request (to be addressed separately)

## Notion Database Information

### Expense Request Database

- **Database URL**: `https://www.notion.so/2bfb69f8f57381cba2daf06d28896390`
- **Data Source ID**: `collection://2bfb69f8-f573-81af-ae31-000b24dacac1`

### Contractor Database (for Requestor relation)

- **Database ID**: `2bfb69f8-f573-805a-8915-000bc44ce188`

## Database Schema

| Property         | Type     | Options/Details                                      |
|------------------|----------|------------------------------------------------------|
| Title            | title    | Expense title/description                            |
| Requestor        | relation | Links to Contractor database                         |
| Email            | rollup   | Rolled up from Requestor relation                    |
| Status           | status   | Pending (to_do), Approved (in_progress), Paid (complete) |
| Expense Category | select   | Expense category options                             |
| Request Date     | date     | Date of request                                      |
| Amount           | number   | Expense amount (float)                               |
| Currency         | select   | USD, VND                                             |
| Attachments      | file     | Uploaded files/receipts                              |

## Status Flow

```
Pending → Approved → Paid
```

## Functional Requirements

### FR1: Fetch Approved Expenses

- The system must query Notion API for expenses with `Status = Approved`
- Transform Notion page data to `bcModel.Todo` format for payroll compatibility
- Link employee via `Requestor` relation → `Email` rollup → `employees.team_email`

### FR2: Transform Expense Data

- Build Todo title in format: `"description | amount | currency"`
- This matches the format expected by payroll calculator's `getReimbursement` function

### FR3: Mark Expenses as Paid

- After payroll commit, update Notion expense status from `Approved` to `Paid`
- Use Notion API to update page properties

### FR4: No Webhook Required

- Expenses are fetched during payroll calculation, not via webhook
- No real-time notification needed for expense approval

## Integration Points

### Payroll Calculation (Fetch Phase)

- `PayrollExpenseProvider.GetAllInList()` queries Notion for approved expenses
- `PayrollExpenseProvider.GetGroups()` returns expense groups
- `PayrollExpenseProvider.GetLists()` returns expense lists

### Payroll Commit (Update Phase)

- `markExpenseSubmissionsAsCompleted()` updates Notion status to `Paid`

## Configuration Requirements

```env
TASK_PROVIDER=notion
NOTION_EXPENSE_DB_ID=2bfb69f8-f573-81cb-a2da-f06d28896390
NOTION_CONTRACTOR_DB_ID=2bfb69f8-f573-805a-8915-000bc44ce188
```

## Interface to Implement

```go
type ExpenseProvider interface {
    GetAllInList(todolistID, projectID int) ([]model.Todo, error)
    GetGroups(todosetID, projectID int) ([]model.TodoGroup, error)
    GetLists(projectID, todosetID int) ([]model.TodoList, error)
}
```

Plus additional method for marking expenses as completed:

```go
MarkExpenseAsCompleted(expenseID string) error  // Uses Notion page ID (UUID)
```

## Key Differences from NocoDB

| Aspect          | NocoDB                  | Notion                     |
|-----------------|-------------------------|----------------------------|
| Record ID       | Integer (`Id`)          | UUID (page ID)             |
| Status Field    | Text field              | Status property type       |
| API             | REST API                | go-notion client           |
| Authentication  | API Token               | Notion Secret              |

## Constraints

- Must use existing `go-notion` library (`github.com/dstotijn/go-notion`)
- Must maintain backward compatibility with existing payroll flow
- Must match existing `bcModel.Todo` format for seamless integration
