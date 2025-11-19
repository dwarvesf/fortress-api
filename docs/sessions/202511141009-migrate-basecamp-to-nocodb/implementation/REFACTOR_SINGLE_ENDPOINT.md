# Refactor: Unified NocoDB Leave Webhook Endpoint

**Date:** 2025-11-19
**Change Type:** Implementation Refactor
**Status:** ✅ Complete

---

## Summary

Consolidated three separate webhook endpoints into a single unified endpoint following the expense handler pattern. The handler now routes internally based on event type and status transitions.

---

## Changes Made

### 1. Handler Implementation

**File:** `pkg/handler/webhook/nocodb_leave.go`

**Added:**
- `LeaveEventType` enum (validate, approve, reject)
- `HandleNocodbLeave()` - Main webhook handler with event routing
- `handleLeaveValidation()` - Private handler for validation logic
- `handleLeaveApproval()` - Private handler for approval logic
- `handleLeaveRejection()` - Private handler for rejection logic

**Event Routing Logic:**
```go
switch payload.Type {
case "record.created", "records.after.insert":
    eventType = LeaveEventValidate
case "record.updated", "records.after.update":
    if old.status != "Approved" && new.status == "Approved":
        eventType = LeaveEventApprove
    else if old.status != "Rejected" && new.status == "Rejected":
        eventType = LeaveEventReject
    else:
        // Ignore other updates
```

**Removed:**
- `ValidateNocodbLeave()` (replaced by `handleLeaveValidation()`)
- `ApproveNocodbLeave()` (replaced by `handleLeaveApproval()`)
- `RejectNocodbLeave()` (replaced by `handleLeaveRejection()`)

### 2. Interface Update

**File:** `pkg/handler/webhook/interface.go`

**Before:**
```go
ValidateNocodbLeave(c *gin.Context)
ApproveNocodbLeave(c *gin.Context)
RejectNocodbLeave(c *gin.Context)
```

**After:**
```go
HandleNocodbLeave(c *gin.Context)
```

### 3. Routes Update

**File:** `pkg/routes/v1.go`

**Before:**
```go
nocodbLeaveGroup := webhook.Group("/nocodb/leave")
{
    nocodbLeaveGroup.POST("/validate", h.Webhook.ValidateNocodbLeave)
    nocodbLeaveGroup.POST("/approve", h.Webhook.ApproveNocodbLeave)
    nocodbLeaveGroup.POST("/reject", h.Webhook.RejectNocodbLeave)
}
```

**After:**
```go
webhook.POST("/nocodb/leave", h.Webhook.HandleNocodbLeave)
```

### 4. Documentation Updates

**Updated Files:**
- `SETUP_AND_TEST_GUIDE.md` - Simplified webhook configuration from 3 webhooks to 1
- `ONLEAVE_STATUS.md` - Updated endpoint table

---

## Benefits

### 1. **Simpler NocoDB Configuration**
- **Before:** 3 separate webhooks with complex conditions
- **After:** 1 webhook with no conditions (routing handled in code)

### 2. **Easier Maintenance**
- Single point of entry for all leave events
- Centralized signature verification
- Consistent error handling

### 3. **Follows Existing Pattern**
- Matches `HandleNocoExpense()` implementation
- Consistent with project architecture

### 4. **Better Debugging**
- All leave events logged in one handler
- Easier to trace event flow
- Single webhook to monitor in NocoDB

---

## Migration Guide

### For New Deployments

1. Create single webhook in NocoDB:
   - **URL:** `https://your-api.com/webhooks/nocodb/leave`
   - **Events:** After Insert, After Update
   - **Condition:** None

### For Existing Deployments

1. Delete old webhooks:
   - `/webhooks/nocodb/leave/validate`
   - `/webhooks/nocodb/leave/approve`
   - `/webhooks/nocodb/leave/reject`

2. Create new webhook:
   - **URL:** `https://your-api.com/webhooks/nocodb/leave`
   - **Events:** After Insert, After Update
   - **Condition:** None

---

## Testing

**Build Status:** ✅ Successful
```bash
go build -o bin/fortress-api cmd/server/main.go
```

**Verification:**
- Code compiles without errors
- Follows existing expense handler pattern
- DEBUG logs added for tracing

---

## Technical Details

### Event Detection

**Record Created:**
- Triggers when `payload.Type == "record.created"` or `"records.after.insert"`
- Routes to validation handler

**Record Updated (Approval):**
- Triggers when `payload.Type == "record.updated"`
- AND `old.status != "Approved"` AND `new.status == "Approved"`
- Routes to approval handler

**Record Updated (Rejection):**
- Triggers when `payload.Type == "record.updated"`
- AND `old.status != "Rejected"` AND `new.status == "Rejected"`
- Routes to rejection handler

**Other Updates:**
- Logged as "ignored"
- Returns 200 OK with `"ignored"` message

### Signature Verification

- Performed once at entry point
- Uses `verifyNocoSignature()` helper
- Checks `X-NocoDB-Signature` header
- Same secret for all event types

### Deletion Behavior

- **Hard Delete:** Rejected leave requests are permanently deleted from the database
- Uses `db.Unscoped().Delete()` to bypass GORM's soft delete
- Re-approval after rejection creates a new record (original is permanently removed)
- No `Undelete()` method - hard-deleted records cannot be restored

---

## Code References

- **Handler:** `pkg/handler/webhook/nocodb_leave.go:59-371`
- **Interface:** `pkg/handler/webhook/interface.go:17`
- **Routes:** `pkg/routes/v1.go:74`
- **Setup Guide:** `docs/sessions/.../implementation/SETUP_AND_TEST_GUIDE.md`
