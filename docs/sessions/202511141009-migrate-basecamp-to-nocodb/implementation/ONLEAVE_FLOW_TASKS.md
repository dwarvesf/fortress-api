# On-Leave Flow Implementation Tasks

> NOTE: Implementation in progress. This document tracks all tasks for migrating leave requests from Basecamp to NocoDB.

## 1. Prerequisites & Database Schema
- [x] 1.1 Create migration to add `nocodb_id` column to `on_leave_requests` table with index.
- [x] 1.2 Update `pkg/model/onleave_request.go` to include `NocodbID *int` field with proper GORM tags.
- [ ] 1.3 Run migration locally and verify schema changes.

## 2. Configuration
- [x] 2.1 Add `LeaveIntegration` config struct in `pkg/config/config.go`.
- [x] 2.2 Add `LeaveNocoIntegration` struct with `TableID` and `WebhookSecret` fields.
- [x] 2.3 Wire environment variables `NOCO_LEAVE_TABLE_ID` and `NOCO_LEAVE_WEBHOOK_SECRET` in `Generate()` function.
- [ ] 2.4 Validate config on startup; fail fast if required values missing.
- [ ] 2.5 Document new env vars in `.env.example`.

## 3. Webhook Handlers
- [x] 3.1 Create `pkg/handler/webhook/nocodb_leave.go` with handler struct and interface.
- [x] 3.2 Define webhook payload structs (`NocodbLeaveWebhookPayload`, `NocodbLeaveRecord`, `NocodbEmployeeLink`).
- [x] 3.3 Implement `ValidateNocodbLeave()` handler for `record.created` events:
  - [x] 3.3.1 Parse and validate webhook payload.
  - [x] 3.3.2 Verify HMAC signature using `NOCO_LEAVE_WEBHOOK_SECRET`.
  - [x] 3.3.3 Lookup employee by `employee_email`.
  - [x] 3.3.4 Validate date range (start >= today, end >= start).
  - [x] 3.3.5 Send Discord notification for pending approval.
  - [x] 3.3.6 Return appropriate status/error responses.
- [x] 3.4 Implement `ApproveNocodbLeave()` handler for `record.updated` (status=Approved) events:
  - [x] 3.4.1 Parse and validate webhook payload.
  - [x] 3.4.2 Verify HMAC signature.
  - [x] 3.4.3 Check status transition (old_record.status != "Approved").
  - [x] 3.4.4 Lookup employee by `employee_email`.
  - [x] 3.4.5 Lookup approver by `approved_by` email.
  - [x] 3.4.6 Parse assignees from `_nc_m2m_leave_requests_employees` junction table.
  - [x] 3.4.7 Generate leave request title.
  - [x] 3.4.8 Create `OnLeaveRequest` record with `NocodbID` reference.
  - [x] 3.4.9 Send Discord approval notification.
  - [x] 3.4.10 Return success response with created record ID.
- [x] 3.5 Implement `RejectNocodbLeave()` handler for `record.updated` (status=Rejected) events:
  - [x] 3.5.1 Parse and validate webhook payload.
  - [x] 3.5.2 Verify HMAC signature.
  - [x] 3.5.3 Send Discord rejection notification.
  - [x] 3.5.4 Log rejection (no DB write needed).
  - [x] 3.5.5 Return success response.
- [x] 3.6 Implement `sendLeaveDiscordNotification()` helper using existing Discord service.

## 4. Routes
- [x] 4.1 Add webhook routes in `pkg/routes/v1.go`:
  - [x] 4.1.1 Create `/webhooks/nocodb/leave` group.
  - [x] 4.1.2 Register `POST /webhooks/nocodb/leave/validate` route.
  - [x] 4.1.3 Register `POST /webhooks/nocodb/leave/approve` route.
  - [x] 4.1.4 Register `POST /webhooks/nocodb/leave/reject` route.
- [x] 4.2 Ensure routes are public (no auth middleware required for webhooks).

## 5. Store Layer (Optional Enhancements)
- [ ] 5.1 Add `GetByNocodbID(db *gorm.DB, nocodbID int)` method to `OnLeaveRequestStore` for idempotency checks (if needed).
- [ ] 5.2 Verify existing `Create()` method handles `NocodbID` field properly.

## 6. Testing & Validation
- [ ] 6.1 Write unit tests for `ValidateNocodbLeave()`:
  - [ ] 6.1.1 Test successful validation with valid payload.
  - [ ] 6.1.2 Test employee not found error.
  - [ ] 6.1.3 Test invalid date format errors.
  - [ ] 6.1.4 Test start date in past validation.
  - [ ] 6.1.5 Test end date before start date validation.
  - [ ] 6.1.6 Test HMAC signature verification.
- [ ] 6.2 Write unit tests for `ApproveNocodbLeave()`:
  - [ ] 6.2.1 Test successful approval flow.
  - [ ] 6.2.2 Test employee not found error.
  - [ ] 6.2.3 Test approver lookup (found/not found fallback).
  - [ ] 6.2.4 Test assignees parsing from junction table.
  - [ ] 6.2.5 Test DB record creation with NocodbID.
  - [ ] 6.2.6 Test status transition guard (already approved).
- [ ] 6.3 Write unit tests for `RejectNocodbLeave()`:
  - [ ] 6.3.1 Test successful rejection flow.
  - [ ] 6.3.2 Test Discord notification sent.
  - [ ] 6.3.3 Test no DB write occurs.
- [ ] 6.4 Create test fixtures for webhook payloads in `pkg/handler/webhook/testdata/`.
- [ ] 6.5 Run unit tests: `make test`.
- [ ] 6.6 Manual integration testing:
  - [ ] 6.6.1 Submit leave request via NocoDB form.
  - [ ] 6.6.2 Verify validation webhook received and Discord notification sent.
  - [ ] 6.6.3 Approve request in NocoDB.
  - [ ] 6.6.4 Verify approval webhook creates DB record.
  - [ ] 6.6.5 Verify Discord approval notification sent.
  - [ ] 6.6.6 Reject request in NocoDB.
  - [ ] 6.6.7 Verify Discord rejection notification sent.

## 7. NocoDB Setup (Manual Steps)
- [ ] 7.1 Create `assignees` link field in `leave_requests` table:
  - [ ] 7.1.1 Link to `nc_employees` table.
  - [ ] 7.1.2 Set relationship type to Many-to-Many.
  - [ ] 7.1.3 Display field: `full_name` (with email).
- [ ] 7.2 Configure NocoDB webhooks:
  - [ ] 7.2.1 Create webhook for `record.created` → `https://fortress-api/webhooks/nocodb/leave/validate`.
  - [ ] 7.2.2 Create webhook for `record.updated` (filter: `status=Approved` AND `old.status!=Approved`) → `https://fortress-api/webhooks/nocodb/leave/approve`.
  - [ ] 7.2.3 Create webhook for `record.updated` (filter: `status=Rejected` AND `old.status!=Rejected`) → `https://fortress-api/webhooks/nocodb/leave/reject`.
  - [ ] 7.2.4 Configure webhook secret for signature verification.
- [ ] 7.3 Create NocoDB views:
  - [ ] 7.3.1 "Pending Requests" view (filter: `status=Pending`).
  - [ ] 7.3.2 "Approved Leaves" view (filter: `status=Approved`).
  - [ ] 7.3.3 Calendar view (map `start_date`/`end_date`, filter: `status=Approved`).

## 8. Deployment & Environment
- [ ] 8.1 Update `.env` with `NOCO_LEAVE_TABLE_ID=myvvv4swtdflfwq`.
- [ ] 8.2 Generate and set `NOCO_LEAVE_WEBHOOK_SECRET`.
- [ ] 8.3 Deploy to staging environment.
- [ ] 8.4 Verify staging environment configuration.
- [ ] 8.5 Monitor logs and metrics for webhook processing.

## 9. Documentation
- [ ] 9.1 Update `NOCODB_LEAVE_STRUCTURE.md` with final implementation details.
- [ ] 9.2 Document webhook endpoints in Swagger/OpenAPI.
- [ ] 9.3 Update `CLAUDE.md` with leave workflow context.
- [ ] 9.4 Create runbook for troubleshooting leave webhook failures.

## 10. Rollout & Monitoring
- [ ] 10.1 Announce new leave request workflow to team.
- [ ] 10.2 Provide training/documentation for NocoDB form submission.
- [ ] 10.3 Monitor webhook success rate (target: >99%).
- [ ] 10.4 Monitor Discord notification delivery rate (target: >99%).
- [ ] 10.5 Monitor for validation/approval errors.
- [ ] 10.6 Collect user feedback for 1 week.
- [ ] 10.7 Address any issues or edge cases discovered.

## Status Summary
- **Total Tasks**: 87
- **Completed**: 0
- **In Progress**: 0
- **Remaining**: 87
- **Blocked**: 0

## Dependencies
- NocoDB base `pin7oroe7to3o1l` with `leave_requests` table (ID: `myvvv4swtdflfwq`)
- NocoDB base `pin7oroe7to3o1l` with `nc_employees` table (already created and synced)
- Existing Discord service and configuration
- Fortress DB with `on_leave_requests` table

## Risk Mitigation
- **Webhook Failures**: Implement retry logic and alerting for webhook processing failures.
- **Employee Lookup Failures**: Clear error messages and Discord notifications for missing employees.
- **Date Validation**: Comprehensive validation with user-friendly error messages.
- **Idempotency**: Use `NocodbID` to prevent duplicate records on webhook retries.
- **Rollback**: Keep Basecamp workflow intact until NocoDB proven stable in production.
