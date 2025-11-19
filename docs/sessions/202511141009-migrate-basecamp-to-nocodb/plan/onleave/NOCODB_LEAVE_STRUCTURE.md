# NocoDB Leave Requests Table Structure

**Date:** 2025-01-19
**Base ID:** `pin7oroe7to3o1l`
**Table ID:** `myvvv4swtdflfwq`
**Table Name:** `leave_requests`

---

## Table Schema

### Fields

| Field Name | NocoDB Type | Required | Nullable | Notes |
|------------|-------------|----------|----------|-------|
| `Id` | AutoNumber | ✓ | No | Primary key, auto-generated |
| `employee_email` | Email | ✓ | No | Employee submitting the request |
| `type` | SingleSelect | ✓ | No | Options: "Off", "Remote" |
| `start_date` | Date | ✓ | No | Leave start date (YYYY-MM-DD) |
| `end_date` | Date | ✓ | No | Leave end date (YYYY-MM-DD) |
| `shift` | SingleSelect | - | Yes | Options: "Morning", "Afternoon", "Full Day" |
| `reason` | LongText | - | Yes | Optional description/reason |
| `status` | SingleSelect | ✓ | No | Options: "Pending", "Approved", "Rejected"<br/>Default: "Pending" |
| `approved_by` | Email | - | Yes | Email of approver (set on approval) |
| `approved_at` | DateTime | - | Yes | Timestamp when approved |
| `CreatedAt` | DateTime | ✓ | No | Auto-generated creation timestamp |
| `UpdatedAt` | DateTime | - | Yes | Auto-updated modification timestamp |
| `assignees` | LinkToAnotherRecord | - | Yes | Many-to-Many link to `employees` table |

### Relationships

**`assignees` → `employees` table**
- **Type:** Many-to-Many
- **Junction Table:** `_nc_m2m_leave_requests_employees`
- **Junction Fields:**
  - `leave_requests_id` - References `leave_requests.Id`
  - `employees_id` - References `employees.Id`
- **Purpose:** CC list of employees to notify about the leave request

---

## Sample Record Structure

### JSON Response from NocoDB API

```json
{
  "Id": 1,
  "employee_email": "quang@d.foundation",
  "type": "Off",
  "start_date": "2025-11-19",
  "end_date": "2025-11-22",
  "shift": "Full Day",
  "reason": "di kham benh",
  "status": "Pending",
  "approved_by": null,
  "approved_at": null,
  "CreatedAt": "2025-11-19 06:04:22+00:00",
  "UpdatedAt": "2025-11-19 06:04:22+00:00",
  "assignees": 2,
  "_nc_m2m_leave_requests_employees": [
    {
      "leave_requests_id": 1,
      "employees_id": 4
    },
    {
      "leave_requests_id": 1,
      "employees_id": 3
    }
  ]
}
```

---

## Webhook Payload Structure

### Record Created Event

```json
{
  "type": "record.created",
  "data": {
    "table_name": "leave_requests",
    "table_id": "myvvv4swtdflfwq",
    "row_id": "rec_xxxx",
    "record": {
      "Id": 1,
      "employee_email": "quang@d.foundation",
      "type": "Off",
      "start_date": "2025-11-19",
      "end_date": "2025-11-22",
      "shift": "Full Day",
      "reason": "di kham benh",
      "status": "Pending",
      "approved_by": null,
      "approved_at": null,
      "CreatedAt": "2025-11-19 06:04:22+00:00",
      "UpdatedAt": "2025-11-19 06:04:22+00:00",
      "assignees": 2,
      "_nc_m2m_leave_requests_employees": [
        {
          "leave_requests_id": 1,
          "employees_id": 4
        }
      ]
    },
    "old_record": null
  }
}
```

### Record Updated Event (Status → Approved)

```json
{
  "type": "record.updated",
  "data": {
    "table_name": "leave_requests",
    "table_id": "myvvv4swtdflfwq",
    "row_id": "rec_xxxx",
    "record": {
      "Id": 1,
      "employee_email": "quang@d.foundation",
      "type": "Off",
      "start_date": "2025-11-19",
      "end_date": "2025-11-22",
      "shift": "Full Day",
      "reason": "di kham benh",
      "status": "Approved",
      "approved_by": "nikki@d.foundation",
      "approved_at": "2025-11-19 10:30:00+00:00",
      "CreatedAt": "2025-11-19 06:04:22+00:00",
      "UpdatedAt": "2025-11-19 10:30:00+00:00",
      "assignees": 2,
      "_nc_m2m_leave_requests_employees": [...]
    },
    "old_record": {
      "Id": 1,
      "status": "Pending",
      "approved_by": null,
      "approved_at": null,
      "UpdatedAt": "2025-11-19 06:04:22+00:00"
    }
  }
}
```

---

## Mapping to Fortress DB (`on_leave_requests`)

### Field Mapping

| NocoDB Field | Fortress DB Field | Transformation |
|--------------|-------------------|----------------|
| `employee_email` | `creator_id` | Lookup employee by email → UUID |
| `type` | `type` | Direct copy ("Off", "Remote") |
| `start_date` | `start_date` | Parse "YYYY-MM-DD" → `*time.Time` |
| `end_date` | `end_date` | Parse "YYYY-MM-DD" → `*time.Time` |
| `shift` | `shift` | Direct copy (or empty string) |
| `reason` | `description` | Direct copy |
| `approved_by` | `approver_id` | Lookup approver by email → UUID |
| `_nc_m2m_leave_requests_employees` | `assignee_ids` | Extract `employees_id` array → `JSONArrayString` |
| Generated | `title` | Format: `{name} \| {type} \| {start} - {end} [\| {shift}]` |

### Fortress DB Model

```go
// pkg/model/onleave_request.go
type OnLeaveRequest struct {
    BaseModel

    Type        string          // "Off" or "Remote"
    StartDate   *time.Time      // Leave start date
    EndDate     *time.Time      // Leave end date
    Shift       string          // "Morning", "Afternoon", "Full Day", or ""
    Title       string          // Generated title
    Description string          // Reason field
    CreatorID   UUID            // Employee who submitted
    ApproverID  UUID            // Employee who approved
    AssigneeIDs JSONArrayString // Array of employee UUIDs (CC list)

    Creator  *Employee
    Approver *Employee
}
```

---

## Implementation Notes

### Employee Lookup

**NocoDB `employees` table must exist with:**
- `Id` (NocoDB AutoNumber)
- `fortress_id` (UUID from Fortress DB)
- `email` (Employee email)
- `full_name` (Display name)
- `working_status` (e.g., "full-time")

**Lookup Process:**
1. Query Fortress DB: `SELECT id, full_name FROM employees WHERE team_email = ? OR personal_email = ?`
2. If not found → validation error (reject webhook)
3. If found → use `id` for `creator_id`/`approver_id`

### Assignees Processing

```go
// Extract assignee IDs from junction table
assigneeIDs := model.JSONArrayString{}

// Add creator to assignees
assigneeIDs = append(assigneeIDs, employee.ID.String())

// Add linked employees from _nc_m2m_leave_requests_employees
for _, link := range payload.Data.Record.AssigneeLinks {
    // Lookup employee by employees_id (from NocoDB employees table)
    emp, err := h.store.Employee.OneByID(h.repo.DB(), link.EmployeesID)
    if err != nil {
        logger.L.Warnw("assignee not found", "employees_id", link.EmployeesID)
        continue
    }
    assigneeIDs = append(assigneeIDs, emp.ID.String())
}
```

### Title Generation

```go
title := fmt.Sprintf("%s | %s | %s - %s",
    employee.FullName,
    payload.Data.Record.Type,
    payload.Data.Record.StartDate,
    payload.Data.Record.EndDate,
)
if payload.Data.Record.Shift != "" {
    title += " | " + payload.Data.Record.Shift
}
```

---

## API Endpoints

### Fetch Records

```bash
curl -H "xc-token: ${NOCO_TOKEN}" \
  "https://app.nocodb.com/api/v1/db/data/noco/pin7oroe7to3o1l/myvvv4swtdflfwq?limit=10"
```

### Query by Status

```bash
curl -H "xc-token: ${NOCO_TOKEN}" \
  "https://app.nocodb.com/api/v1/db/data/noco/pin7oroe7to3o1l/myvvv4swtdflfwq?where=(status,eq,Pending)"
```

### Update Record

```bash
curl -X PATCH \
  -H "xc-token: ${NOCO_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"status": "Approved", "approved_by": "approver@d.foundation", "approved_at": "2025-11-19T10:30:00Z"}' \
  "https://app.nocodb.com/api/v1/db/data/noco/pin7oroe7to3o1l/myvvv4swtdflfwq/{record_id}"
```

---

## Webhook Events to Handle

1. **`record.created`** - Validation webhook
   - Validate employee exists
   - Validate date range (end >= start, start >= today)
   - Send Discord notification for pending approval

2. **`record.updated` (status → "Approved")** - Approval webhook
   - Lookup employee and approver
   - Create record in Fortress DB `on_leave_requests`
   - Send Discord approval notification

3. **`record.updated` (status → "Rejected")** - Rejection webhook
   - Send Discord rejection notification
   - No DB write needed

---

## Configuration Required

### Environment Variables

```bash
# NocoDB Leave Integration
NOCO_LEAVE_TABLE_ID=myvvv4swtdflfwq
NOCO_LEAVE_WEBHOOK_SECRET=xxxx

# Discord Integration
DISCORD_LEAVE_CHANNEL_ID=xxxx
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...
```

### Config Structure

```go
type LeaveIntegration struct {
    Noco    LeaveNocoIntegration
    Discord LeaveDiscordIntegration
}

type LeaveNocoIntegration struct {
    TableID       string
    WebhookSecret string
}

type LeaveDiscordIntegration struct {
    ChannelID  string
    WebhookURL string
}
```
