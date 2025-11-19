# On-Leave Flow Implementation Status

**Date:** 2025-11-19
**Session:** 202511141009-migrate-basecamp-to-nocodb
**Phase:** ✅ Complete

---

## Summary

**Status:** ✅ Implementation complete (7/7 tasks)

All webhook handlers successfully implemented with:
- Single unified endpoint following expense pattern
- Discord embed notifications to auditlog channel
- Hard delete on rejection
- Duplicate prevention with nocodb_id tracking
- Complete error handling and validation

**Build Status:** ✅ Successful

---

## Completed Tasks (7/7)

### 1. ✅ Database Migration
**File:** `migrations/schemas/20251119140000-add-nocodb-id-to-onleave-requests.sql`
- Added `nocodb_id INTEGER` column
- Created index `idx_on_leave_requests_nocodb_id`
- Reversible with proper `Down` migration

### 2. ✅ Model Update
**File:** `pkg/model/onleave_request.go:19`
- Added `NocodbID *int` field with GORM tags

### 3. ✅ Store Methods
**File:** `pkg/store/onleaverequest/`
- `GetByNocodbID()` - Lookup by NocoDB ID (including deleted)
- `Delete()` - Hard delete with `Unscoped()`
- Removed `Undelete()` method (hard delete only)

### 4. ✅ Configuration
**File:** `pkg/config/config.go`
- Uses existing `NOCO_WEBHOOK_SECRET` (no new env vars)
- Discord webhook: `DISCORD_WEBHOOK_AUDIT` (full URL required)

### 5. ✅ Routes Registration
**File:** `pkg/routes/v1.go:74`
- Single endpoint: `POST /webhooks/nocodb/leave`
- Consolidated from 3 separate endpoints

### 6. ✅ Interface Update
**File:** `pkg/handler/webhook/interface.go:17`
- Single method: `HandleNocodbLeave(c *gin.Context)`
- Removed: `ValidateNocodbLeave`, `ApproveNocodbLeave`, `RejectNocodbLeave`

### 7. ✅ Webhook Handler Implementation
**File:** `pkg/handler/webhook/nocodb_leave.go`

**Main Handler:** `HandleNocodbLeave()` (lines 62-157)
- Event routing based on `payload.Type` and status transitions
- HMAC signature verification
- Handles: `records.after.insert`, `records.after.update`

**Validation Handler:** `handleLeaveValidation()` (lines 167-269)
- Employee lookup by email
- Date range validation (start >= today, end >= start)
- Discord notification for pending approval
- ✅ Uses embed format (blue color)

**Approval Handler:** `handleLeaveApproval()` (lines 271-387)
- Employee and approver lookup
- Duplicate prevention with `GetByNocodbID()`
- Title generation with YYYY/MM/DD format
- Persist to `on_leave_requests` table
- Discord notification with approval details
- ✅ Uses embed format (green color)

**Rejection Handler:** `handleLeaveRejection()` (lines 389-434)
- Employee lookup for full name
- Hard delete if previously approved
- Discord notification with rejection details
- ✅ Uses embed format (red color)
- ✅ Uses employee full name instead of email

**Discord Helper:** `sendLeaveDiscordNotification()` (lines 436-455)
- ✅ Sends embed messages to auditlog channel
- Parameters: title, description, color, fields
- Timestamp formatting
- Error handling

---

## Implementation Details

### Event Routing Logic

```go
switch payload.Type {
case "records.after.insert":
    eventType = LeaveEventValidate
case "records.after.update":
    if old.Status != "Approved" && new.Status == "Approved":
        eventType = LeaveEventApprove
    else if old.Status != "Rejected" && new.Status == "Rejected":
        eventType = LeaveEventReject
    else:
        // Ignore other updates
```

### Duplicate Prevention

```go
// Check if leave request already exists
existingLeave, err := h.store.OnLeaveRequest.GetByNocodbID(h.repo.DB(), nocodbID)
if err == nil && existingLeave != nil {
    // Skip duplicate - already exists
    return
}
// Only create if doesn't exist
```

### Hard Delete Behavior

```go
// Delete permanently (hard delete)
func (s *store) Delete(db *gorm.DB, id string) error {
    return db.Unscoped().Delete(&model.OnLeaveRequest{}, "id = ?", id).Error
}
```

**Behavior:**
- Rejected → record permanently removed from database
- Re-approval after rejection → creates new record with new ID
- No restoration capability

### Discord Embed Format

**Colors:**
- Red (15158332) - Validation errors, approval failures, rejections
- Blue (3447003) - Pending approval
- Green (3066993) - Approved

**Fields:**
- Employee (full name or email)
- Type (inline)
- Shift (inline)
- Dates
- Reason

---

## Configuration

### Environment Variables

```bash
# NocoDB
NOCO_BASE_URL=https://app.nocodb.com/api/v2
NOCO_TOKEN=<token>
NOCO_WEBHOOK_SECRET=<secret>
NOCO_WORKSPACE_ID=<workspace>
NOCO_BASE_ID=<base>
NOCO_LEAVE_TABLE_ID=<table>

# Discord (full webhook URL required)
DISCORD_WEBHOOK_AUDIT=https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN
```

### Webhook Endpoint

| URL | Method | Events | Purpose |
|-----|--------|--------|---------|
| `/webhooks/nocodb/leave` | POST | After Insert, After Update | Unified handler with internal routing |

### NocoDB Webhook Configuration

**URL:** `https://your-api.com/webhooks/nocodb/leave`
**Events:** After Insert, After Update
**Condition:** None (routing handled in code)
**Secret:** Set in `NOCO_WEBHOOK_SECRET`

---

## Testing Status

### Manual Testing
- ✅ Build successful
- ✅ Signature verification working
- ✅ Employee lookup working
- ✅ Date validation working
- ✅ Duplicate prevention working
- ✅ Hard delete working
- ✅ Discord embeds configured

### Unit Tests
⏳ Pending - Can be added later if needed

---

## Migration Notes

### For New Deployments
1. Run migration: `make migrate-up`
2. Set environment variables
3. Create webhook in NocoDB with single endpoint

### For Existing Deployments
1. Run migration: `make migrate-up`
2. Delete old webhooks (validate/approve/reject)
3. Create new unified webhook
4. Update `DISCORD_WEBHOOK_AUDIT` to full URL

---

## References

- **Refactor Guide**: `implementation/REFACTOR_SINGLE_ENDPOINT.md`
- **Setup Guide**: `implementation/SETUP_AND_TEST_GUIDE.md`
- **Tasks**: `implementation/ONLEAVE_FLOW_TASKS.md`
- **Schema**: `plan/onleave/NOCODB_LEAVE_STRUCTURE.md`
- **Migration Plan**: `plan/onleave/ONLEAVE_MIGRATION_PLAN.md`

---

## ✅ Feature Complete

**Implementation Date:** 2025-11-19
**Status:** Ready for deployment
**Build:** ✅ Successful
**Pattern:** Follows expense handler pattern
**Notifications:** Discord embeds to auditlog channel
