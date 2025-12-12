# Notion Leave Request Webhook

## Overview

When a leave request is created in Notion, the webhook:
1. Auto-fills the `Contractor` relation based on `Team Email`
2. Finds active deployments for the employee
3. Gets Account Managers (AM) and Delivery Leads (DL) from those deployments
4. Notifies AM/DL via Discord mentions

## Databases Involved

| Database | Environment Variable | Purpose |
|----------|---------------------|---------|
| Leave Requests | `NOTION_LEAVE_REQUEST_DB_ID` | Stores leave requests |
| Deployment Tracker | `NOTION_DEPLOYMENT_TRACKER_DB_ID` | Maps contractors to projects |
| Contractors | `NOTION_CONTRACTOR_DB_ID` | Employee profiles with Discord usernames |

## Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    Leave Request Created                         │
│                    (Team Email provided)                         │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Step 1: Lookup Contractor by Team Email                          │
│ - Query Contractors DB where Team Email = leave.Team Email       │
│ - Get contractor page ID                                         │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Step 2: Update Leave Request                                     │
│ - Set Contractor relation to found contractor page               │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Step 3: Find Active Deployments                                  │
│ - Query Deployment Tracker where:                                │
│   - Contractor = found contractor page ID                        │
│   - Deployment Status = "Active"                                 │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Step 4: Get Stakeholders for Each Deployment                     │
│ - For each active deployment:                                    │
│   - Get Account Managers (rollup from Project)                   │
│   - Get Delivery Leads (rollup from Project)                     │
│   - Check Override AM / Override DL (if set, use these instead)  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Step 5: Get Discord Usernames                                    │
│ - For each AM/DL contractor page:                                │
│   - Fetch Discord field (rich_text)                              │
│ - Deduplicate usernames                                          │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Step 6: Return Discord Usernames                                 │
│ - Store in Leave Request (Submitted Discord field?)              │
│ - Or use for notification purposes                               │
└─────────────────────────────────────────────────────────────────┘
```

## Database Schema Details

### Leave Requests

| Property | Type | Description |
|----------|------|-------------|
| Team Email | email | Employee's team email |
| Contractor | relation | Link to Contractors DB |
| Status | status | Pending, Approved, Rejected, Cancelled |
| Leave Type | select | Annual, Sick, Unpaid, etc. |
| Start Date | date | Leave start |
| End Date | date | Leave end |
| Submitted Discord | rich_text | Discord usernames to notify |

### Deployment Tracker

| Property | Type | Description |
|----------|------|-------------|
| Contractor | relation | Link to Contractors DB |
| Deployment Status | status | Active, Done, Not started |
| Account Managers | rollup | From Project relation |
| Delivery Leads | rollup | From Project relation |
| Override AM | relation | Override Account Manager |
| Override DL | relation | Override Delivery Lead |
| Final AM | formula | Uses Override AM if set, else Account Managers |
| Final Delivery Lead | formula | Uses Override DL if set, else Delivery Leads |

### Contractors

| Property | Type | Description |
|----------|------|-------------|
| Team Email | email | Employee's team email |
| Discord | rich_text | Discord username |
| Full Name | title | Employee name |

## API Calls Required

1. **Query Contractors** - Find contractor by Team Email
2. **Update Leave Request** - Set Contractor relation
3. **Query Deployment Tracker** - Find active deployments for contractor
4. **Fetch Contractor Pages** - Get Discord usernames from AM/DL relations

## Environment Variables

| Variable | Description |
|----------|-------------|
| `NOTION_SECRET` | Notion API secret |
| `NOTION_CONTRACTOR_DB_ID` | Contractors database ID |
| `NOTION_DEPLOYMENT_TRACKER_DB_ID` | Deployment Tracker database ID |

## Webhook Endpoint

```
POST /api/v1/webhooks/notion/onleave
```

Handler: `HandleNotionOnLeave` in `pkg/handler/webhook/notion_leave.go`

## Payload Structure

```json
{
  "source": {
    "type": "automation",
    "automation_id": "...",
    "action_id": "...",
    "event_id": "...",
    "attempt": 1
  },
  "data": {
    "object": "page",
    "id": "page-id",
    "properties": {
      "Status": {
        "status": { "name": "Pending" }
      },
      "Team Email": {
        "email": "employee@d.foundation"
      },
      "Contractor": {
        "relation": []
      }
    },
    "url": "https://notion.so/..."
  }
}
```

## Response

Updates the Leave Request page with:
- `Contractor` relation filled
- `Submitted Discord` field with AM/DL Discord usernames (comma-separated)

---

## Current Implementation

### Endpoint

```
POST /api/v1/webhooks/notion
```

Handler: `HandleNotionLeave` in `pkg/handler/webhook/notion_leave.go`

### Current Flow (page.created event)

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Receive webhook from Notion (page.created / page.updated)    │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. Fetch leave request details via leaveService.GetLeaveRequest │
│    - Extracts: Email, LeaveType, StartDate, EndDate, Status     │
│    - Extracts: Assignees from "Assignees" multi_select property │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. handleNotionLeaveCreated()                                   │
│    - Validates employee exists in DB                            │
│    - Validates dates                                            │
│    - Gets leave.Assignees (emails from Notion multi_select)     │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. Convert assignee emails to Discord mentions                  │
│    - getEmployeeDiscordMention(email)                           │
│    - Looks up employee by email in DB                           │
│    - Gets DiscordAccountID → Discord ID → <@discord_id>         │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. Send Discord message to OnLeaveChannel                       │
│    - Embed with leave details                                   │
│    - Approve/Reject buttons                                     │
│    - Assignee mentions in message content                       │
└─────────────────────────────────────────────────────────────────┘
```

### Current Assignee Source

The current implementation uses the `Assignees` multi_select property from the Leave Request page itself:

```go
// In pkg/service/notion/leave.go
leave.Assignees = s.extractEmailsFromMultiSelect(props, "Assignees")
```

Format: `"Name (email@domain)"` → extracts email

### Current Discord Mention Lookup

```go
// In pkg/handler/webhook/nocodb_leave.go
func (h *handler) getEmployeeDiscordMention(l logger.Logger, email string) string {
    employee, err := h.store.Employee.OneByEmail(h.repo.DB(), email)
    // ...
    discordAccount, err := h.store.DiscordAccount.One(h.repo.DB(), employee.DiscordAccountID.String())
    // ...
    return fmt.Sprintf("<@%s>", discordAccount.DiscordID)
}
```

---

## Proposed Change

### Replace Assignees Source

Instead of using the "Assignees" multi_select from Leave Request, fetch AM/DL from Deployment Tracker:

```
┌─────────────────────────────────────────────────────────────────┐
│ NEW Step 3.5: Get AM/DL from Deployment Tracker                 │
└─────────────────────────────────────────────────────────────────┘
                                │
        ┌───────────────────────┴───────────────────────┐
        ▼                                               ▼
┌───────────────────────┐                 ┌───────────────────────┐
│ a. Find contractor    │                 │ b. Query Deployment   │
│    by Team Email      │                 │    Tracker where:     │
│    in Contractors DB  │                 │    - Contractor = ID  │
│                       │                 │    - Status = Active  │
└───────────────────────┘                 └───────────────────────┘
                                                      │
                                                      ▼
                                          ┌───────────────────────┐
                                          │ c. For each deployment│
                                          │    get Final AM and   │
                                          │    Final Delivery Lead│
                                          │    (formula fields)   │
                                          └───────────────────────┘
                                                      │
                                                      ▼
                                          ┌───────────────────────┐
                                          │ d. Fetch Discord      │
                                          │    username from each │
                                          │    AM/DL contractor   │
                                          │    page               │
                                          └───────────────────────┘
                                                      │
                                                      ▼
                                          ┌───────────────────────┐
                                          │ e. Convert Discord    │
                                          │    username to mention│
                                          │    via DB lookup      │
                                          └───────────────────────┘
```

### New Functions Required

1. **`getActiveDeploymentsForContractor(contractorPageID)`**
   - Query Deployment Tracker where Contractor = contractorPageID AND Status = Active
   - Return list of deployment pages

2. **`getStakeholdersFromDeployment(deploymentPage)`**
   - Extract Override AM / Override DL relations
   - If empty, extract Account Managers / Delivery Leads rollups
   - Return list of contractor page IDs

3. **`getDiscordUsernameFromContractor(contractorPageID)`**
   - Fetch contractor page
   - Extract Discord field (rich_text)
   - Return Discord username

4. **`getDiscordMentionFromUsername(discordUsername)`**
   - Query DB for employee with matching Discord username
   - Return `<@discord_id>` mention

### Key Properties to Extract

| Database | Property | Type | Notes |
|----------|----------|------|-------|
| Deployment Tracker | Contractor | relation | To match employee |
| Deployment Tracker | Deployment Status | status | Filter by "Active" |
| Deployment Tracker | Override AM | relation | Direct contractor relation |
| Deployment Tracker | Override DL | relation | Direct contractor relation |
| Deployment Tracker | Account Managers | rollup | From Project (use if no override) |
| Deployment Tracker | Delivery Leads | rollup | From Project (use if no override) |
| Contractors | Discord | rich_text | Discord username |

### Environment Variables Required

| Variable | Description |
|----------|-------------|
| `NOTION_DEPLOYMENT_TRACKER_DB_ID` | Deployment Tracker database ID (`2b864b29b84c80799568dc17685f4f33`) |
