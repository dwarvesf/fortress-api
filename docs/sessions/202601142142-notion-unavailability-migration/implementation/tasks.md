# Implementation Tasks: Notion Unavailability Notices Migration

## Overview
Migrate from old Notion leave schema to Unavailability Notices schema. No database operations - all data stays in Notion.

**Plan Reference**: `/Users/quang/.claude/plans/shimmering-popping-valiant.md`

---

## Phase 1: Notion Service Layer Updates

### Task 1.1: Update LeaveRequest Struct
- **File(s)**: `pkg/service/notion/leave.go` (lines 26-45)
- **Description**: Update `LeaveRequest` struct fields:
  - `Reason` → `LeaveRequestTitle` (string)
  - `LeaveType` → `UnavailabilityType` (string)
  - `ApprovedByID` → `ReviewedByID` (string)
  - `ApprovedAt` → `DateApproved` (*time.Time)
  - Add `DateRequested` (*time.Time)
  - Add `AdditionalContext` (string)
- **Acceptance**:
  - Struct compiles without errors
  - All new fields have correct types
  - Comments updated to reflect new field names

### Task 1.2: Update Property Extraction - Title Field
- **File(s)**: `pkg/service/notion/leave.go` (line ~68)
- **Description**: Update title property extraction
  - Keep `case "title"` but populate `LeaveRequestTitle` instead of `Reason`
- **Acceptance**:
  - Title field extracts to `LeaveRequestTitle`
  - Code compiles without errors

### Task 1.3: Update Property Extraction - Unavailability Type
- **File(s)**: `pkg/service/notion/leave.go` (line ~85)
- **Description**: Update select property extraction
  - Change `case "Leave Type"` to `case "Unavailability Type"`
  - Populate `UnavailabilityType` instead of `LeaveType`
- **Acceptance**:
  - Unavailability Type values extracted correctly
  - All select options supported

### Task 1.4: Update Property Extraction - Additional Context
- **File(s)**: `pkg/service/notion/leave.go` (new case in property switch)
- **Description**: Add new rich_text property extraction
  - Add `case "Additional Context"`
  - Extract to `AdditionalContext` field
- **Acceptance**:
  - Additional Context text extracted correctly
  - Handles empty/null values gracefully

### Task 1.5: Update Property Extraction - Reviewed By
- **File(s)**: `pkg/service/notion/leave.go` (line ~95)
- **Description**: Update relation property extraction
  - Change `case "Approved By"` to `case "Reviewed By"`
  - Populate `ReviewedByID` instead of `ApprovedByID`
- **Acceptance**:
  - Reviewed By relation extracted correctly
  - Handles empty relations

### Task 1.6: Add Property Extraction - Date Requested
- **File(s)**: `pkg/service/notion/leave.go` (new case in property switch)
- **Description**: Add new date property extraction
  - Add `case "Date Requested"`
  - Parse and populate `DateRequested` field
  - Use RFC3339 format for parsing
- **Acceptance**:
  - Date Requested parsed correctly
  - Handles empty/null dates
  - Time zone handling correct

---

## Phase 2: Webhook Handler Updates

### Task 2.1: Remove Deprecated Handler Function
- **File(s)**: `pkg/handler/webhook/notion_leave.go` (lines 60-883)
- **Description**: Delete `HandleNotionLeave` function and supporting functions:
  - `HandleNotionLeave` (main handler)
  - `handleNotionLeaveCreated`
  - `handleNotionLeaveApproved`
  - `handleNotionLeaveRejected`
  - `NotionLeaveEventType` and related constants
- **Acceptance**:
  - All deprecated functions removed
  - No compilation errors
  - No references to removed functions remain

### Task 2.2: Remove Deprecated Route
- **File(s)**: `pkg/routes/v1.go` (line 82)
- **Description**: Remove deprecated webhook route
  - Delete `webhook.POST("/notion", h.Webhook.HandleNotionLeave)`
- **Acceptance**:
  - Route removed from router
  - Code compiles
  - Route tests updated/removed

### Task 2.3: Remove Handler from Interface
- **File(s)**: `pkg/handler/webhook/interface.go`
- **Description**: Remove `HandleNotionLeave` from interface definition
- **Acceptance**:
  - Interface definition updated
  - No compilation errors
  - All implementations satisfy interface

### Task 2.4: Update Status Check in HandleNotionOnLeave
- **File(s)**: `pkg/handler/webhook/notion_leave.go` (line 928)
- **Description**: Update status filtering logic
  - Change `if status != "Pending"` to `if status != "New"`
  - Update log messages to reference "New" status
- **Acceptance**:
  - Only "New" status processed
  - Other statuses ignored with appropriate log
  - Returns 200 OK with "ignored" message

### Task 2.5: Update Auto-Fill Logic
- **File(s)**: `pkg/handler/webhook/notion_leave.go` (lines 983-994)
- **Description**: Update auto-fill logic for unavailability type
  - Change `leave.LeaveType` to `leave.UnavailabilityType`
  - Change default from "Annual Leave" to "Personal Time"
  - Update `updateLeaveType` function name to `updateUnavailabilityType`
- **Acceptance**:
  - Auto-fills "Personal Time" when empty
  - Updates Notion correctly
  - Logs reflect new field name

### Task 2.6: Update Discord Embed Fields
- **File(s)**: `pkg/handler/webhook/notion_leave.go` (lines 1117-1119)
- **Description**: Update Discord notification embed fields
  - Replace `{Name: "Reason", Value: leave.Reason}` with:
    - `{Name: "Leave Request", Value: leave.LeaveRequestTitle, Inline: false}`
    - `{Name: "Type", Value: leave.UnavailabilityType, Inline: true}`
    - `{Name: "Details", Value: leave.AdditionalContext, Inline: false}`
- **Acceptance**:
  - Discord embed shows all three fields correctly
  - Field values populated from new schema
  - Formatting looks clean

---

## Phase 3: Discord Interaction Handler Updates

### Task 3.1: Update Approve Button Status
- **File(s)**: `pkg/handler/webhook/discord_interaction.go` (line 523)
- **Description**: Update approve handler to use "Acknowledged" status
  - Change `UpdateLeaveStatus(ctx, pageID, "Approved", approverPageID)`
  - To `UpdateLeaveStatus(ctx, pageID, "Acknowledged", approverPageID)`
- **Acceptance**:
  - Approve button updates Notion to "Acknowledged"
  - Reviewed By field populated in Notion
  - Discord message updates correctly

### Task 3.2: Update Reject Button Status
- **File(s)**: `pkg/handler/webhook/discord_interaction.go` (line 599)
- **Description**: Update reject handler to use "Not Applicable" status
  - Change `UpdateLeaveStatus(ctx, pageID, "Rejected", rejectorPageID)`
  - To `UpdateLeaveStatus(ctx, pageID, "Not Applicable", rejectorPageID)`
- **Acceptance**:
  - Reject button updates Notion to "Not Applicable"
  - Discord message updates correctly

---

## Phase 4: Configuration Updates

### Task 4.1: Update Environment Variables
- **File(s)**: `.env`, deployment configs
- **Description**: Update Notion database ID
  - Change `NOTION_LEAVE_DB_ID` to `2cc64b29-b84c-80ef-bb0e-000bf2c8bfcb`
  - Verify `NOTION_CONTRACTOR_DB_ID` is `ed2b9224-97d9-4dff-97f9-82598b61f65d`
- **Acceptance**:
  - Environment variable updated in all environments
  - Configuration loads correctly
  - Webhook points to new database

---

## Phase 5: Testing

### Task 5.1: Update Integration Tests
- **File(s)**: `pkg/handler/webhook/notion_leave_test.go` or create new test file
- **Description**: Update/create tests for new schema
  - Test "New" status webhook handling
  - Test field extraction (LeaveRequestTitle, UnavailabilityType, AdditionalContext)
  - Test auto-fill logic for UnavailabilityType
  - Test Discord notification formatting
  - Test approve/reject button status updates
- **Acceptance**:
  - All tests pass
  - Coverage for new fields
  - Mock Notion responses use new schema

### Task 5.2: Create Test Webhook Payloads
- **File(s)**: `testdata/fixtures/notion_unavailability_webhook.json` (new file)
- **Description**: Create sample webhook payloads
  - "New" status payload with all new fields
  - Payload with empty UnavailabilityType (for auto-fill test)
  - Payload with rollup fields (Team Email, Discord)
- **Acceptance**:
  - Valid JSON structure
  - Matches Notion Unavailability Notices schema
  - Can be used in integration tests

---

## Verification Checklist

After all tasks completed:

- [ ] All deprecated code removed (`HandleNotionLeave`, route, interface)
- [ ] Service layer extracts all new fields correctly
- [ ] Webhook processes "New" status only
- [ ] Discord notifications show new field names
- [ ] Approve button sets "Acknowledged" status
- [ ] Reject button sets "Not Applicable" status
- [ ] Environment variable updated
- [ ] Tests updated and passing
- [ ] No compilation errors
- [ ] No runtime errors in logs

---

## Dependencies

**Task Dependencies**:
- Task 2.5, 2.6 depend on Task 1.1-1.6 (need updated struct)
- Task 3.1, 3.2 depend on Task 1.5 (need ReviewedByID field)
- Task 5.1 depends on all implementation tasks
- All tasks can proceed after Phase 1 completion

**Recommended Order**:
1. Phase 1 (all tasks) - Foundation
2. Phase 2 (all tasks) - Handler updates
3. Phase 3 (all tasks) - Discord buttons
4. Phase 4 - Configuration
5. Phase 5 - Testing

---

## Estimated Time

- **Phase 1**: 1 hour
- **Phase 2**: 1 hour
- **Phase 3**: 30 minutes
- **Phase 4**: 15 minutes
- **Phase 5**: 1 hour

**Total**: ~3.5-4 hours
