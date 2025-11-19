# On-Leave Flow Migration Plan - NocoDB Simplified Approach

**Date:** 2025-01-19
**Status:** Planning
**Approach:** NocoDB-only (No Basecamp dual-write)

---

## Overview

Migrate on-leave request workflow from Basecamp to NocoDB using a simplified approach:
- NocoDB table as single source of truth
- Form-based submission (no title parsing)
- Discord notifications for approval workflow
- NocoDB calendar view (no external calendar sync)
- Direct webhook integration to Fortress API

---

## 1. NocoDB Schema Design

### Table: `leave_requests`

**NocoDB Configuration:**

| Field Name | Field Type | Required | Config | Notes |
|------------|-----------|----------|---------|-------|
| `Id` | AutoNumber | ✓ | Primary key | NocoDB internal ID |
| `employee_email` | Email | ✓ | Validation: email format | Used to lookup employee in DB |
| `employee_name` | SingleLineText | - | Display only | Auto-filled from email lookup |
| `type` | SingleSelect | ✓ | Options: "Off", "Remote" | Leave type |
| `start_date` | Date | ✓ | Date picker | Cannot be in past |
| `end_date` | Date | ✓ | Date picker | Must be >= start_date |
| `shift` | SingleSelect | - | Options: "Morning", "Afternoon", "Full Day" | Optional field |
| `reason` | LongText | - | Multi-line text | Optional description |
| `status` | SingleSelect | ✓ | Options: "Pending", "Approved", "Rejected"<br/>Default: "Pending" | Workflow status |
| `approved_by_email` | Email | - | Auto-filled on approval | Approver email |
| `approved_at` | DateTime | - | Auto timestamp | When approved |
| `assignees` | LinkToAnotherRecord | - | Link to `employees` table (Many-to-Many) | Optional CC list |
| `created_at` | DateTime | ✓ | Auto timestamp | Submission time |

**NocoDB Table Relationships:**

**Link: `assignees` → `employees` table**
- **Type:** Many-to-Many
- **Display Field:** `full_name` (with email)
- **Lookup Fields:** `id`, `email`, `full_name`
- **Purpose:** Allow selecting multiple employees as assignees/CC

**Note:** Requires `employees` table in NocoDB with columns:
- `id` (UUID) - mapped to Fortress DB employee UUID
- `email` (Email) - employee email
- `full_name` (SingleLineText) - employee display name

**NocoDB Views:**

1. **Pending Requests** - Filter: `status = "Pending"` (for approvers)
2. **Approved Leaves** - Filter: `status = "Approved"` (for calendar view)
3. **Calendar View** - Map `start_date`/`end_date` to calendar, filter by "Approved"
4. **My Requests** - Filter by `employee_email = {current_user_email}`

---

## 2. Database Schema (Fortress API)

### Existing Table: `on_leave_requests`

**Migration Required** - Add `nocodb_id` column to track NocoDB record reference.

**Migration:** `YYYYMMDDHHMMSS-add-nocodb-id-to-onleave-requests.sql`

```sql
-- +migrate Up
ALTER TABLE on_leave_requests ADD COLUMN nocodb_id INTEGER;
CREATE INDEX idx_on_leave_requests_nocodb_id ON on_leave_requests(nocodb_id);

-- +migrate Down
DROP INDEX idx_on_leave_requests_nocodb_id;
ALTER TABLE on_leave_requests DROP COLUMN nocodb_id;
```

**Updated Model:**

```go
// pkg/model/onleave_request.go
type OnLeaveRequest struct {
    BaseModel

    Type        string
    StartDate   *time.Time
    EndDate     *time.Time
    Shift       string
    Title       string
    Description string
    CreatorID   UUID
    ApproverID  UUID
    AssigneeIDs JSONArrayString
    NocodbID    *int            // NEW: NocoDB record ID for cross-reference

    Creator  *Employee
    Approver *Employee
}
```

**Purpose:** Enables bidirectional sync between NocoDB and Fortress DB for status updates and data reconciliation.

---

## 3. Workflow Design

### 3.1 Submission Flow

```
1. Employee fills NocoDB form
   ├─ employee_email (required)
   ├─ type: "Off" or "Remote" (required)
   ├─ start_date / end_date (required)
   ├─ shift (optional)
   └─ reason (optional)
   ↓
2. NocoDB creates record
   └─ status = "Pending"
   ↓
3. NocoDB webhook: POST /webhooks/nocodb/leave/validate
   ├─ Event: "record.created"
   ├─ Fortress validates employee exists
   ├─ Validates date range
   └─ Sends Discord notification
   ↓
4. Discord message posted to #leave-requests channel
   ├─ Employee: {name}
   ├─ Type: {Off/Remote}
   ├─ Dates: {start} - {end}
   ├─ Shift: {shift}
   ├─ Reason: {reason}
   └─ Link to NocoDB record
```

### 3.2 Approval Flow

```
1. Approver reviews in NocoDB (or Discord link)
   ↓
2. Approver changes status field
   ├─ "Approved" → Creates DB record
   └─ "Rejected" → Discord notification only
   ↓
3. NocoDB webhook: POST /webhooks/nocodb/leave/approve
   ├─ Event: "record.updated"
   ├─ Condition: status changed to "Approved"
   └─ Fortress API processes
   ↓
4. Fortress API creates on_leave_request record
   ├─ Lookup employee by email → creator_id
   ├─ Lookup approver by approved_by_email → approver_id
   ├─ Parse assignee_emails → assignee_ids (JSONB)
   ├─ Store nocodb_id for reference
   └─ Set status = "approved"
   ↓
5. Discord notification: "Leave approved for {employee}"
   ↓
6. NocoDB calendar view automatically shows approved leave
```

### 3.3 Rejection Flow

```
1. Approver changes status to "Rejected"
   ↓
2. NocoDB webhook: POST /webhooks/nocodb/leave/reject
   ├─ Event: "record.updated"
   ├─ Condition: status changed to "Rejected"
   └─ Fortress API logs (no DB write)
   ↓
3. Discord notification: "Leave rejected for {employee}"
```

---

## 4. Webhook Integration

### 4.1 NocoDB Webhook Configuration

**Webhook 1: Validation (On Create)**
- **Event:** `record.created`
- **URL:** `https://fortress-api.example.com/webhooks/nocodb/leave/validate`
- **Method:** POST
- **Headers:** `x-nocodb-signature: {secret}`

**Webhook 2: Approval (On Update - Status = "Approved")**
- **Event:** `record.updated`
- **Filter:** `status = "Approved" AND old.status != "Approved"`
- **URL:** `https://fortress-api.example.com/webhooks/nocodb/leave/approve`
- **Method:** POST
- **Headers:** `x-nocodb-signature: {secret}`

**Webhook 3: Rejection (On Update - Status = "Rejected")**
- **Event:** `record.updated`
- **Filter:** `status = "Rejected" AND old.status != "Rejected"`
- **URL:** `https://fortress-api.example.com/webhooks/nocodb/leave/reject`
- **Method:** POST
- **Headers:** `x-nocodb-signature: {secret}`

### 4.2 Webhook Payload Structure

```json
{
  "type": "record.created",
  "data": {
    "table_name": "leave_requests",
    "table_id": "tbl_xxxx",
    "row_id": "rec_yyyy",
    "record": {
      "Id": 123,
      "employee_email": "john@dwarves.foundation",
      "employee_name": "John Doe",
      "type": "Off",
      "start_date": "2025-02-01",
      "end_date": "2025-02-05",
      "shift": "Full Day",
      "reason": "Personal leave",
      "status": "Pending",
      "assignees": [
        {
          "id": "uuid-1",
          "email": "user1@dwarves.foundation",
          "full_name": "User One"
        },
        {
          "id": "uuid-2",
          "email": "user2@dwarves.foundation",
          "full_name": "User Two"
        }
      ],
      "created_at": "2025-01-19T10:00:00Z"
    },
    "old_record": null
  }
}
```

---

## 5. API Implementation

### 5.1 File Structure

```
pkg/
├─ handler/webhook/
│  └─ nocodb_leave.go               (NEW)
├─ service/nocodb/
│  └─ leave.go                      (NEW - optional helper)
├─ store/onleaverequest/
│  └─ onleave_request.go            (UPDATE - add GetByNocodbID)
└─ routes/v1.go                     (UPDATE - add routes)
```

### 5.2 Webhook Handler

**File:** `pkg/handler/webhook/nocodb_leave.go`

```go
package webhook

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/dwarvesf/fortress-api/pkg/logger"
    "github.com/dwarvesf/fortress-api/pkg/model"
    "github.com/gin-gonic/gin"
)

// NocodbLeaveWebhookPayload represents the webhook payload from NocoDB
type NocodbLeaveWebhookPayload struct {
    Type string `json:"type"`
    Data struct {
        TableName string                 `json:"table_name"`
        TableID   string                 `json:"table_id"`
        RowID     string                 `json:"row_id"`
        Record    NocodbLeaveRecord      `json:"record"`
        OldRecord *NocodbLeaveRecord     `json:"old_record"`
    } `json:"data"`
}

type NocodbLeaveRecord struct {
    ID               int        `json:"Id"`
    EmployeeEmail    string     `json:"employee_email"`
    EmployeeName     string     `json:"employee_name"`
    Type             string     `json:"type"`
    StartDate        string     `json:"start_date"`
    EndDate          string     `json:"end_date"`
    Shift            string     `json:"shift"`
    Reason           string     `json:"reason"`
    Status           string     `json:"status"`
    ApprovedByEmail  string     `json:"approved_by_email"`
    ApprovedAt       *time.Time `json:"approved_at"`
    Assignees        []NocodbEmployee `json:"assignees"`
    CreatedAt        time.Time  `json:"created_at"`
}

type NocodbEmployee struct {
    ID       string `json:"id"`
    Email    string `json:"email"`
    FullName string `json:"full_name"`
}

// ValidateNocodbLeave handles leave request validation on creation
func (h *handler) ValidateNocodbLeave(c *gin.Context) {
    ctx := context.Background()
    var payload NocodbLeaveWebhookPayload

    if err := c.BindJSON(&payload); err != nil {
        logger.L.Debugw("failed to parse nocodb leave webhook", "error", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
        return
    }

    logger.L.Debugw("received nocodb leave validation webhook",
        "type", payload.Type,
        "row_id", payload.Data.RowID,
        "employee_email", payload.Data.Record.EmployeeEmail,
        "start_date", payload.Data.Record.StartDate,
        "end_date", payload.Data.Record.EndDate,
    )

    // Validate employee exists
    employee, err := h.store.Employee.OneByEmail(h.repo.DB(), payload.Data.Record.EmployeeEmail)
    if err != nil {
        logger.L.Errorw("employee not found", "email", payload.Data.Record.EmployeeEmail, "error", err)
        h.sendDiscordNotification(ctx, fmt.Sprintf(
            "❌ Leave request validation failed\n"+
                "Employee: %s\n"+
                "Reason: Employee not found in database",
            payload.Data.Record.EmployeeEmail,
        ))
        c.JSON(http.StatusOK, gin.H{"status": "validation_failed", "error": "employee_not_found"})
        return
    }

    // Validate date range
    startDate, err := time.Parse("2006-01-02", payload.Data.Record.StartDate)
    if err != nil {
        logger.L.Errorw("invalid start date", "start_date", payload.Data.Record.StartDate, "error", err)
        h.sendDiscordNotification(ctx, fmt.Sprintf(
            "❌ Leave request validation failed\n"+
                "Employee: %s\n"+
                "Reason: Invalid start date format",
            payload.Data.Record.EmployeeEmail,
        ))
        c.JSON(http.StatusOK, gin.H{"status": "validation_failed", "error": "invalid_start_date"})
        return
    }

    endDate, err := time.Parse("2006-01-02", payload.Data.Record.EndDate)
    if err != nil {
        logger.L.Errorw("invalid end date", "end_date", payload.Data.Record.EndDate, "error", err)
        h.sendDiscordNotification(ctx, fmt.Sprintf(
            "❌ Leave request validation failed\n"+
                "Employee: %s\n"+
                "Reason: Invalid end date format",
            payload.Data.Record.EmployeeEmail,
        ))
        c.JSON(http.StatusOK, gin.H{"status": "validation_failed", "error": "invalid_end_date"})
        return
    }

    // Validate date range (end >= start, start >= today)
    now := time.Now().Truncate(24 * time.Hour)
    if startDate.Before(now) {
        logger.L.Debugw("start date in past", "start_date", startDate, "now", now)
        h.sendDiscordNotification(ctx, fmt.Sprintf(
            "❌ Leave request validation failed\n"+
                "Employee: %s\n"+
                "Reason: Start date cannot be in the past",
            payload.Data.Record.EmployeeEmail,
        ))
        c.JSON(http.StatusOK, gin.H{"status": "validation_failed", "error": "start_date_in_past"})
        return
    }

    if endDate.Before(startDate) {
        logger.L.Debugw("end date before start date", "start_date", startDate, "end_date", endDate)
        h.sendDiscordNotification(ctx, fmt.Sprintf(
            "❌ Leave request validation failed\n"+
                "Employee: %s\n"+
                "Reason: End date must be after start date",
            payload.Data.Record.EmployeeEmail,
        ))
        c.JSON(http.StatusOK, gin.H{"status": "validation_failed", "error": "invalid_date_range"})
        return
    }

    // Send Discord notification for pending approval
    h.sendDiscordNotification(ctx, fmt.Sprintf(
        "📋 **New Leave Request - Pending Approval**\n\n"+
            "**Employee:** %s (%s)\n"+
            "**Type:** %s\n"+
            "**Dates:** %s to %s\n"+
            "**Shift:** %s\n"+
            "**Reason:** %s\n\n"+
            "🔗 [View in NocoDB](https://nocodb.example.com/nc/leave_requests/%s)",
        employee.FullName, payload.Data.Record.EmployeeEmail,
        payload.Data.Record.Type,
        payload.Data.Record.StartDate, payload.Data.Record.EndDate,
        payload.Data.Record.Shift,
        payload.Data.Record.Reason,
        payload.Data.RowID,
    ))

    logger.L.Infow("leave request validated successfully",
        "employee_id", employee.ID,
        "row_id", payload.Data.RowID,
    )

    c.JSON(http.StatusOK, gin.H{"status": "validated"})
}

// ApproveNocodbLeave handles leave request approval
func (h *handler) ApproveNocodbLeave(c *gin.Context) {
    ctx := context.Background()
    var payload NocodbLeaveWebhookPayload

    if err := c.BindJSON(&payload); err != nil {
        logger.L.Debugw("failed to parse nocodb leave webhook", "error", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
        return
    }

    logger.L.Debugw("received nocodb leave approval webhook",
        "type", payload.Type,
        "row_id", payload.Data.RowID,
        "employee_email", payload.Data.Record.EmployeeEmail,
        "status", payload.Data.Record.Status,
        "approved_by", payload.Data.Record.ApprovedByEmail,
    )

    // Validate status transition
    if payload.Data.OldRecord != nil && payload.Data.OldRecord.Status == "Approved" {
        logger.L.Debugw("leave already approved, skipping", "row_id", payload.Data.RowID)
        c.JSON(http.StatusOK, gin.H{"status": "already_approved"})
        return
    }

    // Note: No idempotency check needed - records only inserted on first approval

    // Lookup employee by email
    employee, err := h.store.Employee.OneByEmail(h.repo.DB(), payload.Data.Record.EmployeeEmail)
    if err != nil {
        logger.L.Errorw("employee not found", "email", payload.Data.Record.EmployeeEmail, "error", err)
        h.sendDiscordNotification(ctx, fmt.Sprintf(
            "❌ Leave approval failed\n"+
                "Employee: %s\n"+
                "Reason: Employee not found in database",
            payload.Data.Record.EmployeeEmail,
        ))
        c.JSON(http.StatusBadRequest, gin.H{"error": "employee_not_found"})
        return
    }

    // Lookup approver by email
    var approverID model.UUID
    if payload.Data.Record.ApprovedByEmail != "" {
        approver, err := h.store.Employee.OneByEmail(h.repo.DB(), payload.Data.Record.ApprovedByEmail)
        if err != nil {
            logger.L.Warnw("approver not found, using creator as approver",
                "approved_by_email", payload.Data.Record.ApprovedByEmail,
                "error", err,
            )
            approverID = employee.ID
        } else {
            approverID = approver.ID
        }
    } else {
        approverID = employee.ID
    }

    // Parse dates
    startDate, _ := time.Parse("2006-01-02", payload.Data.Record.StartDate)
    endDate, _ := time.Parse("2006-01-02", payload.Data.Record.EndDate)

    // Parse assignees from NocoDB linked records
    assigneeIDs := model.JSONArrayString{employee.ID.String()} // Include creator
    for _, assignee := range payload.Data.Record.Assignees {
        // Validate UUID format
        if _, err := uuid.Parse(assignee.ID); err != nil {
            logger.L.Warnw("invalid assignee UUID", "id", assignee.ID, "error", err)
            continue
        }
        assigneeIDs = append(assigneeIDs, assignee.ID)
    }

    // Generate title from record data
    title := fmt.Sprintf("%s | %s | %s - %s",
        employee.FullName,
        payload.Data.Record.Type,
        payload.Data.Record.StartDate,
        payload.Data.Record.EndDate,
    )
    if payload.Data.Record.Shift != "" {
        title += " | " + payload.Data.Record.Shift
    }

    // Create on_leave_request record
    nocodbID := payload.Data.Record.ID
    leaveRequest := &model.OnLeaveRequest{
        Type:        payload.Data.Record.Type,
        StartDate:   &startDate,
        EndDate:     &endDate,
        Shift:       payload.Data.Record.Shift,
        Title:       title,
        Description: payload.Data.Record.Reason,
        CreatorID:   employee.ID,
        ApproverID:  approverID,
        AssigneeIDs: assigneeIDs,
        NocodbID:    &nocodbID, // Store NocoDB record ID for cross-reference
    }

    err = h.store.OnLeaveRequest.Create(h.repo.DB(), leaveRequest)
    if err != nil {
        logger.L.Errorw("failed to create leave request", "error", err)
        h.sendDiscordNotification(ctx, fmt.Sprintf(
            "❌ Leave approval failed\n"+
                "Employee: %s\n"+
                "Reason: Database error",
            employee.FullName,
        ))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed_to_create_record"})
        return
    }

    // Send Discord notification
    h.sendDiscordNotification(ctx, fmt.Sprintf(
        "✅ **Leave Request Approved**\n\n"+
            "**Employee:** %s\n"+
            "**Type:** %s\n"+
            "**Dates:** %s to %s\n"+
            "**Shift:** %s\n"+
            "**Approved by:** %s",
        employee.FullName,
        payload.Data.Record.Type,
        payload.Data.Record.StartDate, payload.Data.Record.EndDate,
        payload.Data.Record.Shift,
        payload.Data.Record.ApprovedByEmail,
    ))

    logger.L.Infow("leave request approved and persisted",
        "id", leaveRequest.ID,
        "employee_id", employee.ID,
        "assignee_count", len(assigneeIDs),
    )

    c.JSON(http.StatusOK, gin.H{"status": "approved", "id": leaveRequest.ID})
}

// RejectNocodbLeave handles leave request rejection
func (h *handler) RejectNocodbLeave(c *gin.Context) {
    ctx := context.Background()
    var payload NocodbLeaveWebhookPayload

    if err := c.BindJSON(&payload); err != nil {
        logger.L.Debugw("failed to parse nocodb leave webhook", "error", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
        return
    }

    logger.L.Debugw("received nocodb leave rejection webhook",
        "type", payload.Type,
        "row_id", payload.Data.RowID,
        "employee_email", payload.Data.Record.EmployeeEmail,
        "status", payload.Data.Record.Status,
    )

    // No DB write needed - just send notification
    h.sendDiscordNotification(ctx, fmt.Sprintf(
        "❌ **Leave Request Rejected**\n\n"+
            "**Employee:** %s\n"+
            "**Type:** %s\n"+
            "**Dates:** %s to %s\n"+
            "**Reason:** %s",
        payload.Data.Record.EmployeeEmail,
        payload.Data.Record.Type,
        payload.Data.Record.StartDate, payload.Data.Record.EndDate,
        payload.Data.Record.Reason,
    ))

    logger.L.Infow("leave request rejected",
        "row_id", payload.Data.RowID,
        "employee_email", payload.Data.Record.EmployeeEmail,
    )

    c.JSON(http.StatusOK, gin.H{"status": "rejected"})
}

// sendDiscordNotification sends a message to Discord
func (h *handler) sendDiscordNotification(ctx context.Context, message string) {
    if h.service.Discord == nil {
        logger.L.Debugw("discord service not configured, skipping notification")
        return
    }

    // TODO: Use actual Discord service
    // For now, just log
    logger.L.Infow("discord notification", "message", message)
}
```

### 5.3 Store Layer

**File:** `pkg/store/onleaverequest/onleave_request.go`

**No changes needed** - existing `Create()` method is sufficient.

### 5.4 Routes Update

**File:** `pkg/routes/v1.go`

Add routes:

```go
// NocoDB webhooks
nocodbGroup := r.Group("/webhooks/nocodb")
{
    leaveGroup := nocodbGroup.Group("/leave")
    {
        leaveGroup.POST("/validate", h.Webhook.ValidateNocodbLeave)
        leaveGroup.POST("/approve", h.Webhook.ApproveNocodbLeave)
        leaveGroup.POST("/reject", h.Webhook.RejectNocodbLeave)
    }
}
```

---

## 6. Discord Notification Design

### 6.1 Channel Setup

**Channel:** `#leave-requests` (or existing channel)

**Bot Permissions:**
- Send Messages
- Embed Links
- Mention Users (optional)

### 6.2 Message Format

**Validation (Pending Approval):**
```
📋 **New Leave Request - Pending Approval**

**Employee:** John Doe (john@dwarves.foundation)
**Type:** Off
**Dates:** 2025-02-01 to 2025-02-05
**Shift:** Full Day
**Reason:** Personal leave

🔗 [View in NocoDB](https://nocodb.example.com/...)
```

**Approved:**
```
✅ **Leave Request Approved**

**Employee:** John Doe
**Type:** Off
**Dates:** 2025-02-01 to 2025-02-05
**Shift:** Full Day
**Approved by:** approver@dwarves.foundation
```

**Rejected:**
```
❌ **Leave Request Rejected**

**Employee:** John Doe
**Type:** Off
**Dates:** 2025-02-01 to 2025-02-05
**Reason:** Personal leave
```

**Validation Failed:**
```
❌ Leave request validation failed
Employee: john@dwarves.foundation
Reason: Start date cannot be in the past
```

---

## 7. Configuration

### 7.1 Environment Variables

```bash
# NocoDB Leave Integration
NOCO_LEAVE_TABLE_ID=myvvv4swtdflfwq
NOCO_LEAVE_WEBHOOK_SECRET=xxxx

# Note: Reuses existing Discord config (DISCORD_WEBHOOK_URL, DISCORD_CHANNEL_ID, etc.)
```

### 7.2 Config Structure

**File:** `pkg/config/config.go`

```go
type LeaveIntegration struct {
    Noco LeaveNocoIntegration
    // Note: Discord integration reuses existing cfg.Discord config
}

type LeaveNocoIntegration struct {
    TableID       string // NOCO_LEAVE_TABLE_ID
    WebhookSecret string // NOCO_LEAVE_WEBHOOK_SECRET
}
```

---

## 8. Testing Strategy

### 8.1 Unit Tests

**File:** `pkg/handler/webhook/nocodb_leave_test.go`

Test cases:
- [ ] Validate webhook payload parsing
- [ ] Employee lookup by email (found/not found)
- [ ] Date validation (past dates, invalid range)
- [ ] Approval creates DB record correctly
- [ ] Rejection doesn't create DB record
- [ ] Idempotency (duplicate approval)
- [ ] Discord notification called

### 8.2 Integration Tests

- [ ] NocoDB form submission → webhook received
- [ ] Validation webhook → Discord notification sent
- [ ] Approval webhook → DB record created
- [ ] Rejection webhook → Discord notification only
- [ ] Calendar view displays approved leaves

### 8.3 Manual Testing Checklist

- [ ] Submit leave request via NocoDB form
- [ ] Verify Discord notification received
- [ ] Verify validation errors shown for invalid data
- [ ] Approve request in NocoDB
- [ ] Verify DB record created in `on_leave_requests`
- [ ] Verify Discord approval notification
- [ ] Check NocoDB calendar view shows approved leave
- [ ] Reject request in NocoDB
- [ ] Verify Discord rejection notification
- [ ] Verify no DB record created

---

## 9. Migration & Rollout

### 9.1 Migration Steps

**Week 1: Setup**
1. ✅ Create `nc_employees` table in NocoDB
   - ✅ Sync employee data from Fortress DB (`scripts/local/sync_employees_to_nocodb.sh`)
   - ✅ Fields: `fortress_id` (UUID), `email`, `full_name`, `working_status`
2. ✅ Create `leave_requests` table with schema (`scripts/local/create_nocodb_leave_tables.sh`)
   - 🔲 Configure `assignees` link to `nc_employees` table (manual in NocoDB UI)
3. 🔲 Configure NocoDB webhooks (validate, approve, reject)
4. ✅ Discord channel already exists (reuse existing config)
5. 🔲 Update environment variables (NOCO_LEAVE_TABLE_ID, NOCO_LEAVE_WEBHOOK_SECRET)

**Week 2: Development**
1. 🔲 Create migration for `nocodb_id` column in `on_leave_requests`
2. 🔲 Update `pkg/model/onleave_request.go` with `NocodbID` field
3. 🔲 Update `pkg/config/config.go` with `LeaveIntegration` struct
4. 🔲 Implement webhook handlers (`pkg/handler/webhook/nocodb_leave.go`)
5. 🔲 Add routes (`pkg/routes/v1.go`)

**Week 3: Testing**
1. Unit tests
2. Integration tests with NocoDB test instance
3. Manual testing in dev environment

**Week 4: Rollout**
1. Deploy to staging
2. Test with real employees (small group)
3. Deploy to production
4. Monitor for 1 week
5. Disable Basecamp on-leave webhook

### 9.2 Rollback Strategy

**If NocoDB fails:**
1. Re-enable Basecamp on-leave webhook
2. Revert route changes
3. Keep DB records created via NocoDB (no data loss)

**Rollback triggers:**
- Webhook failures > 10% for 1 hour
- Discord notifications not sending
- Employee lookup failures

---

## 10. Success Metrics

### 10.1 Technical Metrics

- [ ] Webhook success rate > 99%
- [ ] Response time < 500ms (p95)
- [ ] Discord notification delivery rate > 99%
- [ ] Zero data loss during migration
- [ ] Calendar view accuracy 100%

### 10.2 User Experience Metrics

- [ ] Form submission success rate > 99%
- [ ] Approval latency < 24 hours (business metric)
- [ ] User satisfaction with NocoDB form vs Basecamp
- [ ] Reduction in validation errors (better UX)

---

## 11. Advantages Over Basecamp

| Aspect | Basecamp | NocoDB |
|--------|----------|--------|
| **Form Entry** | Title parsing (error-prone) | Form fields (validated) |
| **Validation** | Manual comment posting | Instant feedback + Discord |
| **Approval** | Todo completion | Status field change |
| **Calendar** | Basecamp schedule (chunking) | NocoDB calendar view (native) |
| **Notifications** | Comments only | Discord (better visibility) |
| **Data Quality** | Regex parsing errors | Form validation |
| **Mobile UX** | Basecamp app | NocoDB mobile app |
| **Reporting** | Manual | NocoDB views/filters |

---

## 12. Future Enhancements

**Phase 2 (Optional):**
1. Email notifications (in addition to Discord)
2. Slack integration
3. Google Calendar sync (for external stakeholders)
4. Leave balance tracking in NocoDB
5. Bulk import historical leaves
6. Automatic approver assignment (based on team)
7. Multi-level approval workflow
8. Leave type policies (max days per year)

---

## Summary

**Complexity:** Low (compared to expense/accounting flows)
**Estimated Effort:** 3-4 weeks
**Risk Level:** Low (NocoDB-only, no dual-write)
**Dependencies:** NocoDB setup, Discord bot

**Ready to proceed:** Yes, after NocoDB table creation and webhook configuration.
