# Implementation Tasks: Update Timesheet Schema to Project Updates

## Overview

Update all Go code to reflect the Notion database schema changes from "Timesheet" to "Project Updates", including property name changes, status updates, and database ID updates.

## Tasks

### Task 1: Update Core Service Layer - TimesheetEntry Struct
- **File(s)**: `pkg/service/notion/timesheet.go`
- **Description**:
  - Update `TimesheetEntry` struct (lines 23-33):
    - Rename field `Hours float64` → `ApproxEffort float64`
    - Update struct comments to reference "project update" instead of "timesheet"
  - Update property extraction calls:
    - Line 80: Change `"(auto) Timesheet Entry"` → `"(auto) Entry"`
    - Line 84: Change `"Hours"` → `"Appx. effort"`
  - Update function comments and documentation strings to use "project update" terminology
- **Acceptance**:
  - Struct compiles without errors
  - Property names match new Notion schema
  - Comments accurately describe project updates

### Task 2: Update Core Service Layer - Property Extraction
- **File(s)**: `pkg/service/notion/timesheet.go`
- **Description**:
  - Update `extractNumber` method call at line 84 to use `"Appx. effort"` property name
  - Verify all helper methods (`extractTitle`, `extractStatus`, `extractNumber`, etc.) work with updated property names
  - Update log messages to reference "project update" instead of "timesheet"
- **Acceptance**:
  - All property extractions use correct property names
  - Debug logs reference correct terminology
  - No hardcoded old property names remain

### Task 3: Update Webhook Handler - Comments and Messages
- **File(s)**: `pkg/handler/webhook/notion_timesheet.go`
- **Description**:
  - Update struct comments (lines 19-40) to reference "Project Updates"
  - Update function comment (line 42-44): Replace "timesheet entry" → "project update entry"
  - Update log messages throughout to use "project update" terminology
  - Update Discord error notification (lines 220-265):
    - Title: `"⚠️ Timesheet Contractor Fill Failed"` → `"⚠️ Project Update Contractor Fill Failed"`
    - Description: Update to reference "project update entry"
    - Field values: Update to reference "project update"
- **Acceptance**:
  - All comments reference "project updates"
  - Discord notifications show updated terminology
  - Log messages are consistent with new naming

### Task 4: Update Task Order Log Service - Property Names
- **File(s)**: `pkg/service/notion/task_order_log.go`
- **Description**:
  - Replace all occurrences of `"Proof of Works"` → `"Key deliverables"`:
    - Line 158
    - Line 581
    - Line 732
    - Line 1073
    - Line 1276-1277
  - Update struct field comments:
    - Line 941: Update `ProofOfWork` comment
    - Line 1100: Update `ProofOfWorks` comment
- **Acceptance**:
  - All property references use "Key deliverables"
  - Struct comments accurately describe the fields
  - No "Proof of Works" references remain

### Task 5: Update Task Order Log Service - Status Values
- **File(s)**: `pkg/service/notion/task_order_log.go`
- **Description**:
  - Replace all occurrences of status `"Pending Approval"` → `"Pending Feedback"`:
    - Line 568
    - Line 750
    - Line 766
  - Update function comment at line 715
- **Acceptance**:
  - All status values use "Pending Feedback"
  - Status transitions work correctly
  - Function comments describe correct status flow

### Task 6: Update Scripts - Database ID and Property Names
- **File(s)**: `scripts/fill-timesheet-contractor/main.go`
- **Description**:
  - Line 21: Update database ID constant:
    - `timesheetDBID = "2c664b29b84c8089b304e9c5b5c70ac3"` → `timesheetDBID = "2c664b29b84c8048b7e2000bb8278044"`
  - Line 381: Update title property name:
    - `"(auto) Timesheet Entry"` → `"(auto) Entry"`
  - Update log messages and comments to reference "project update"
- **Acceptance**:
  - Script uses correct database ID
  - Property name matches new schema
  - Log messages use updated terminology

### Task 7: Search and Update Additional References
- **File(s)**: Multiple files
- **Description**:
  - Search `pkg/service/notion/contractor_fees.go` for "Proof of Works" → "Key deliverables"
  - Search `pkg/handler/notion/task_order_log.go` for "Proof of Works" → "Key deliverables"
  - Verify no other files reference old property names or status values
  - Check for any remaining "Pending Approval" references
- **Acceptance**:
  - All files use new property names
  - No old status values remain
  - Grep searches return no matches for old names

### Task 8: Update Documentation
- **File(s)**:
  - `docs/specs/notion/schema/timesheet.md` (already updated)
  - `docs/specs/notion/exploring-notion-schemas.md` (already created)
- **Description**:
  - Verify schema documentation is accurate
  - Ensure exploration guide is complete
  - Add migration notes if needed
- **Acceptance**:
  - Documentation matches actual implementation
  - All schema changes are documented
  - Examples use correct property names

### Task 9: Verification - Build and Test
- **File(s)**: All modified files
- **Description**:
  - Run `make build` to verify compilation
  - Run `make lint` to check code quality
  - Run `make test` to verify tests pass
  - Review any test failures related to schema changes
- **Acceptance**:
  - Code compiles without errors
  - Linter passes
  - All tests pass or failing tests are documented

### Task 10: Configuration Verification
- **File(s)**: Configuration files, environment variables
- **Description**:
  - Verify database ID configuration in deployment environments
  - Check if `NOTION_DATABASES_TIMESHEET` env var needs updating
  - Document configuration changes needed for deployment
  - Verify webhook URL is correctly configured
- **Acceptance**:
  - Configuration requirements documented
  - Deployment checklist created
  - Environment variables identified

## Notes

- **Breaking Changes**: Database ID change is breaking and requires coordinated deployment
- **Backward Compatibility**: None - old database ID will stop working
- **Monitoring**: Watch Discord #fortress-logs for webhook errors post-deployment
- **Rollback**: Requires reverting both code and Notion schema changes

## Success Criteria

1. All code compiles and tests pass
2. Property names match new Notion schema exactly
3. Status values use new naming ("Pending Feedback" not "Pending Approval")
4. Database ID points to new database
5. Documentation is up-to-date and accurate
6. No references to old property/status names remain
7. Webhook handler works with new schema
8. Scripts use correct database ID
