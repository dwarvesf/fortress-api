# Implementation Tasks - Sync Task Order Logs

## Task 1: Config Updates
**File**: `pkg/config/config.go`
- [x] Add `Timesheet`, `TaskOrderLog` to `NotionDatabase` struct
- [x] Add `OpenRouter` config struct with `APIKey` and `Model`
- [x] Add env var bindings in config initialization

## Task 2: OpenRouter Service
**File**: `pkg/service/openrouter/openrouter.go` (NEW)
- [x] Create `OpenRouterService` struct
- [x] Implement `SummarizeProofOfWorks(ctx, texts []string) (string, error)`
- [x] Add retry logic with exponential backoff (max 3 retries)
- [x] Add DEBUG logging

## Task 3: Task Order Log Notion Service
**File**: `pkg/service/notion/task_order_log.go` (NEW)
- [x] Create `TaskOrderLogService` struct
- [x] Implement `QueryApprovedTimesheetsByMonth(ctx, month, contractorDiscord string) ([]*TimesheetEntry, error)`
- [x] Implement `GetDeploymentByContractor(ctx, contractorID string) (string, error)`
- [x] Implement `CreateOrder(ctx, deploymentID, month string) (string, error)`
- [x] Implement `CreateTimesheetLineItem(ctx, orderID, projectID string, hours float64, proofOfWorks string, timesheetIDs []string, month string) (string, error)`
- [x] Implement `CheckOrderExists(ctx, deploymentID, month string) (bool, string, error)`
- [x] Add DEBUG logging

## Task 4: Handler
**File**: `pkg/handler/notion/task_order_log.go` (NEW)
- [x] Create `SyncTaskOrderLogs(c *gin.Context)` handler
- [x] Parse `month` query param (required) with validation
- [x] Parse `contractor` query param (optional) - Discord username to filter by specific contractor
- [x] Implement processing logic:
  1. Query approved timesheets
  2. Group by Contractor → Project
  3. For each Contractor: create Order (if not exists)
  4. For each Project: aggregate hours, summarize PoW, create sub-item
- [x] Return response with counts and details
- [x] Add DEBUG logging

## Task 5: Handler Interface
**File**: `pkg/handler/notion/interface.go`
- [x] Add `SyncTaskOrderLogs` to `IHandler` interface

## Task 6: Routes
**File**: `pkg/routes/v1.go`
- [x] Add `cronjob.POST("/sync-task-order-logs", ...)` with auth and permission middleware

## Task 7: Service Initialization
**File**: `pkg/service/service.go`
- [x] Add `OpenRouter` service to service struct
- [x] Create `notion.Services` struct with `TaskOrderLog` and `Timesheet` services
- [x] Initialize in `New()` function

## Status: ✅ COMPLETED
All tasks have been implemented and verified. The build is successful.

## Dependencies
- Task 1 must complete first (config)
- Task 2 and Task 3 can run in parallel after Task 1
- Task 4 depends on Task 2 and Task 3
- Task 5 and Task 6 depend on Task 4
- Task 7 depends on Task 2 and Task 3
