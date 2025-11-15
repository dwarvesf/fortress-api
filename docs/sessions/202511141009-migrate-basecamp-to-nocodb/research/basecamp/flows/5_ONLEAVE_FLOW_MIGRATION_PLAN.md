# On-Leave Flow Migration Plan

## 1. Inventory Current Workflow
- Study `pkg/handler/webhook/onleave.go` to map validation (todo_created) and approval (todo_completed) branches, including calendar creation + comment posting.
- List Basecamp services used: Todo (create/update), Schedule (create events), People (subscriber lookup), plus constants (OnLeave bucket/list IDs, ops approvers).
- Capture dependencies on stores/models (`pkg/store/onleaverequest`, `pkg/model/onleave_request`) and any worker jobs triggered.
- Document parsing rules (pipe-delimited title, allowed leave types, date-range handling, shift field, multi-month chunking) and environment gating.

## 2. Design NocoDB Leave Request Schema
- Create tables for `LeaveRequests` (fields: employee, type, start_date, end_date, shift, notes, status, approver_ids) and `LeaveCalendarEntries` (normalized per day/month if needed).
- Plan how validation status and feedback are represented (status column + comment log). Ensure schema tracks metadata currently deduced from Basecamp (creator ID, bucket, todo ID).
- Decide comment/notification strategy (NocoDB activity log, email integration, or bridging to Slack) to replace Basecamp comments/assignments.

## 3. Build Provider Layer
- Add `LeaveIntegration` interface with methods like `ValidateRequest`, `CreateRequest`, `ApproveRequest`, `CreateCalendarEntry`, `PostFeedback`.
- Implement Basecamp + NocoDB adapters behind the interface so the webhook handler switches provider via configuration.
- Ensure provider handles subscriber/assignee propagation; for NocoDB, map employee records to workspace users or store references on the row.

## 4. Calendar & Scheduling Replacement
- Define how approved leaves appear on calendars (NocoDB calendar view, Google Calendar sync, or existing internal calendar service). Implement function to generate multi-month entries akin to Basecamp schedule events.
- Maintain chunking logic to avoid cross-month issues; store generated event IDs for rollback when requests are reverted.
- Provide migration for existing upcoming events: export from Basecamp, import into chosen calendar system.

## 5. Workflow Automation & Webhooks
- Configure NocoDB automations to trigger validation webhooks on row creation and approval webhooks on status changes.
- Update webhook handler to parse NocoDB payloads (who created, status, comments) and enforce same business rules (allowed types, date constraints, environment gating).
- Integrate approval routing: tie `approver_id` fields to HR/Ops roles rather than hardcoded Basecamp user IDs.

## 6. Testing & Cutover
- Add tests covering parsing edge cases (single-day, multi-day, shift optional, invalid formats) using provider mocks.
- Simulate calendar creation in tests to confirm monthly chunking + ID storage works for both providers.
- Rollout plan: run dual notifications (Basecamp + NocoDB) for a sprint, verify Ops approvals + calendar sync accuracy, then disable Basecamp webhook and remove OnLeave bucket automation.
