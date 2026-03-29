# Notion On-Leave Automation Webhook

## Overview

The Notion On-Leave Automation feature handles incoming webhooks from Notion's Unavailability Notices database automation. When a contractor creates a new leave request in Notion, this system automatically:

1. **Validates** the leave request (dates, contractor status)
2. **Auto-fills** the Contractor relation based on the creator's Notion user ID
3. **Notifies** relevant Account Managers (AM) and Delivery Leads (DL) via Discord
4. **Creates** Google Calendar events for approved leave

## System Architecture

### Components

| Component | File | Purpose |
|-----------|------|---------|
| Webhook Handler | `pkg/handler/webhook/notion_leave.go` | Entry point, validation, notification |
| Leave Service | `pkg/service/notion/leave.go` | Notion API operations |
| Discord Interaction | `pkg/handler/webhook/discord_interaction_leave_notion.go` | Approve/reject button handling |

### Databases Involved

| Database | Environment Variable | Purpose |
|----------|---------------------|---------|
| Unavailability Notices | `NOTION_LEAVE_DB_ID` | Leave request storage |
| Contractors | `NOTION_CONTRACTOR_DB_ID` | Employee profiles with Discord |
| Deployment Tracker | `NOTION_DEPLOYMENT_TRACKER_DB_ID` | Project-contractor mappings |

---

## Data Models

### LeaveRequest

```go
type LeaveRequest struct {
    PageID              string     // Notion page ID
    LeaveRequestTitle   string     // Title field (format: OOO-YYYY-USERNAME-CODE)
    EmployeeID          string     // Relation to Contractors DB
    Email               string     // Team email (from Contractor rollup)
    UnavailabilityType  string     // "Personal Time", "Health / Illness", etc.
    StartDate           *time.Time // Leave start
    EndDate             *time.Time // Leave end
    Status              string     // "New", "Acknowledged", "Not Applicable", "Withdrawn"
    ReviewedByID        string     // Relation to person who approved
    DateApproved        *time.Time // When request was approved
    DateRequested       *time.Time // When request was submitted
    AdditionalContext   string     // Reason for leave
    Assignees           []string   // Discord mentions for AM/DL
}
```

### ContractorDetails

```go
type ContractorDetails struct {
    PageID          string // Notion page ID
    FullName        string // Full name from title
    DiscordUsername string // Discord username
    TeamEmail       string // Team email
    Status          string // "Active", etc.
}
```

---

## Webhook Endpoint

```
POST /api/v1/webhooks/notion/onleave
```

**Handler:** `HandleNotionOnLeave` in `pkg/handler/webhook/notion_leave.go:422`

### Configuration

The webhook requires the following environment variables:

| Variable | Description |
|----------|-------------|
| `NOTION_SECRET` | Notion API secret |
| `NOTION_CONTRACTOR_DB_ID` | Contractors database ID |
| `NOTION_LEAVE_DB_ID` | Unavailability Notices database ID |
| `NOTION_DEPLOYMENT_TRACKER_DB_ID` | Deployment Tracker database ID |
| `DISCORD_CHANNEL_ONLEAVE` | Discord channel ID for leave notifications |
| `GOOGLE_SERVICE_ACCOUNT_JSON` | Google service account for Calendar API |

### Payload Structure

Notion automation sends this payload:

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
        "status": { "name": "New" }
      },
      "Team Email": {
        "email": "name@d.foundation"
      },
      "Contractor": {
        "relation": []
      },
      "Created by": {
        "created_by": {
          "id": "notion-user-id"
        }
      }
    },
    "url": "https://notion.so/..."
  }
}
```

---

## Processing Flow

### Main Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│            Notion Automation Triggers Webhook                    │
│         (New leave request created in Unavailability Notices)    │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ HandleNotionOnLeave()                                           │
│  1. Verify signature (HMAC-SHA256)                              │
│  2. Parse payload                                               │
│  3. Gate: Only process "New" status                             │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Step 1: Fetch Leave Request                                      │
│  - GetLeaveRequest(pageID)                                      │
│  - Extracts: dates, title, type, email                          │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Step 2: Lookup Contractor by Created By                         │
│  - lookupContractorByUserIDForOnLeave(userID)                   │
│  - Maps Notion user to Contractor page                          │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Step 3: Fetch Contractor Details                                 │
│  - GetContractorDetails(contractorPageID)                       │
│  - Gets: full name, Discord, email, status                      │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Step 4: Validate Contractor Status                              │
│  - Must be "Active"                                             │
│  - If inactive → send Discord error notification                │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Step 5: Validate Dates                                          │
│  - Start date must exist                                        │
│  - End date must be >= start date                               │
│  - If invalid → send Discord error notification                 │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Step 6: Send Discord Notification (PRIORITY)                    │
│  - getAMDLMentionsFromDeployments()                             │
│  - Get active deployments for contractor                        │
│  - Extract AM/DL from deployments                               │
│  - Convert to Discord mentions                                  │
│  - Fallback to default assignees if none found                  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Step 7: Send HTTP 200 Response (Immediate)                      │
│  - Notion webhooks timeout in 10 seconds                        │
│  - Remaining steps are best-effort only                         │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Step 8 (Best-Effort): Auto-fill Contractor Relation             │
│  - updateOnLeaveContractor(pageID, contractorPageID)            │
│  - Fills "Contractor" property on Notion page                   │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Step 9 (Best-Effort): Auto-fill Unavailability Type             │
│  - Set to "Personal Time" if empty                              │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Step 10 (Best-Effort): Complete Leave Request ID                │
│  - Notion generates "OOO-YYYY--CODE" (missing username)         │
│  - completeLeaveRequestID(title, discordUsername)               │
│  - Fills missing username: "OOO-YYYY-username-CODE"             │
└─────────────────────────────────────────────────────────────────┘
```

---

## Key Functions

### Notion Operations (`pkg/service/notion/leave.go`)

| Function | Purpose |
|----------|---------|
| `GetLeaveRequest(ctx, pageID)` | Fetch leave request details from Notion |
| `UpdateLeaveStatus(ctx, pageID, status, approverPageID)` | Update status to "Acknowledged" or "Not Applicable" |
| `GetContractorPageIDByEmail(ctx, email)` | Lookup contractor by email |
| `LookupContractorByEmail(ctx, email)` | Find contractor page ID by team email |
| `GetContractorDetails(ctx, pageID)` | Get contractor full name, Discord, status |
| `GetActiveDeploymentsForContractor(ctx, contractorPageID)` | Query active deployments |
| `ExtractStakeholdersFromDeployment(ctx, deployment)` | Extract AM/DL from deployment |
| `GetDiscordUsernameFromContractor(ctx, contractorPageID)` | Get Discord username |
| `QueryAcknowledgedLeaveDatesByContractorMonth(ctx, contractorID, year, month)` | Query approved leave for calendar |

### Webhook Handler (`pkg/handler/webhook/notion_leave.go`)

| Function | Purpose |
|----------|---------|
| `HandleNotionOnLeave(c)` | Main webhook entry point |
| `getAMDLMentionsFromDeployments(ctx, l, leaveService, email)` | Get Discord mentions for AM/DL |
| `sendLeaveNotification(ctx, l, leaveService, leave, contractor, mention)` | Send Discord embed with buttons |
| `lookupContractorByUserIDForOnLeave(ctx, l, userID)` | Find contractor by Notion user ID |
| `updateOnLeaveContractor(ctx, l, onLeavePageID, contractorPageID)` | Auto-fill Contractor relation |
| `getContractorEmail(ctx, l, contractorPageID)` | Get contractor team email |
| `updateUnavailabilityType(ctx, l, pageID, unavailabilityType)` | Auto-fill Unavailability Type |

### Discord Interaction (`pkg/handler/webhook/discord_interaction_leave_notion.go`)

| Function | Purpose |
|----------|---------|
| `handleNotionLeaveApproveButton(c, l, interaction, pageID)` | Handle Approve button click |
| `handleNotionLeaveRejectButton(c, l, interaction, pageID)` | Handle Reject button click |
| `createCalendarEventForLeave(ctx, l, leaveService, pageID)` | Create Google Calendar event |

---

## AM/DL Stakeholder Detection

### Logic Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ For each Active Deployment where Contractor = employee          │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Check Override AM (relation)                                    │
│   If exists → Use these contractor IDs                         │
│   ELSE → Use Account Managers rollup from Project              │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ Check Override DL (relation)                                    │
│   If exists → Use these contractor IDs                         │
│   ELSE → Use Delivery Leads rollup from Project                │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│ For each AM/DL contractor:                                      │
│   1. GetDiscordUsernameFromContractor()                         │
│   2. getDiscordMentionFromUsername()                            │
│   3. Add to mentions array                                      │
└─────────────────────────────────────────────────────────────────┘
```

### Fallback Logic

If no AM/DL mentions found (no active deployments or no stakeholders):
1. Use default assignees: `han@d.foundation`, `thanhpd@d.foundation`
2. Lookup employees in database
3. Convert to Discord mentions

---

## Discord Notification Format

### Embed Message

```
📋 Leave Request
─────────────────────────────────────────
@Request>  OOO-2026-nlk0211-GD5F
<Type>     Personal Time
<Dates>    Jan 15-20, 2026
<Details>  Taking time off for personal matters
─────────────────────────────────────────
🔔 Assignees: <@123456789> <@987654321>
```

### Interactive Buttons

- **Approve** (`notion_leave_approve_{pageID}`) - Updates status to "Acknowledged"
- **Reject** (`notion_leave_reject_{pageID}`) - Updates status to "Not Applicable"

---

## Calendar Integration

When a leave request is approved, a Google Calendar event is created:

- **Title**: "{Contractor Name} - Time Off"
- **Dates**: Start Date to End Date
- **Attendees**: Contractor + AM/DL stakeholders
- **Description**: Includes leave type and additional context

---

## Error Handling

### Validation Failures

| Error | Discord Notification |
|-------|---------------------|
| Contractor not active | ❌ Leave Request Validation Failed - Contractor is not active |
| Missing dates | ❌ Leave Request Validation Failed - missing start or end date |
| Invalid date range | ❌ Leave Request Validation Failed - End date must be after start date |

### Auto-Fill Failures

When auto-fill fails (e.g., contractor lookup by user ID fails):
- Send error notification to `fortress-logs` Discord channel
- Include: Page ID, Created By User, Error message
- Request manual intervention

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `NOTION_SECRET` | Notion API secret |
| `NOTION_CONTRACTOR_DB_ID` | Contractors database ID |
| `NOTION_LEAVE_DB_ID` | Unavailability Notices database ID |
| `NOTION_DEPLOYMENT_TRACKER_DB_ID` | Deployment Tracker database ID |
| `DISCORD_CHANNEL_ONLEAVE` | Leave notification channel ID |
| `DISCORD_WEBHOOK_AUDITLOG` | Audit log webhook URL |
| `DISCORD_WEBHOOK_FORTRESS_LOGS` | Error notification channel |
| `GOOGLE_SERVICE_ACCOUNT_JSON` | Google service account JSON |

---

## Related Files

### Core Implementation

- `pkg/handler/webhook/notion_leave.go` - Main webhook handler
- `pkg/handler/webhook/discord_interaction_leave_notion.go` - Discord button handlers
- `pkg/service/notion/leave.go` - Leave service operations
- `pkg/routes/v1.go` - Route registration (line 83)

### Tests

- `pkg/routes/v1_test.go` - Route tests for `/webhooks/notion/onleave`

### Database Schema

- `docs/specs/notion/schema/unavailability-notices.md` - Leave request schema
- `docs/specs/notion/schema/contractor.md` - Contractor schema
- `docs/specs/notion/schema/deployment-tracker.md` - Deployment schema

---

## API Summary

| Endpoint | Method | Handler Function |
|----------|--------|------------------|
| `/webhooks/notion/onleave` | POST | `HandleNotionOnLeave` |
| `/webhooks/discord/interaction` | POST | `HandleDiscordInteraction` |

### Leave Status Values

| Status | Description |
|--------|-------------|
| `New` | Newly submitted request, awaiting approval |
| `Acknowledged` | Approved by AM/DL |
| `Not Applicable` | Rejected by AM/DL |
| `Withdrawn` | Cancelled by requester |

### Unavailability Types

| Type | Description |
|------|-------------|
| `Personal Time` | Personal vacation/time off |
| `Health / Illness` | Sick leave |
| `Family / Emergency` | Family emergency |
| `Travel / Vacation` | Travel or vacation |
| `Other` | Other leave type |