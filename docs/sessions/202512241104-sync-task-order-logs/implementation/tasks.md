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
- [x] Add retry logic with model rotation (4 free models)
- [x] Updated prompt: scope-based format (max 2 scopes, 3-4 activities per scope, 2-3 words per activity)
- [x] Add DEBUG logging

## Task 3: Task Order Log Notion Service
**File**: `pkg/service/notion/task_order_log.go` (NEW)
- [x] Create `TaskOrderLogService` struct
- [x] Implement `QueryApprovedTimesheetsByMonth(ctx, month, contractorDiscord string) ([]*TimesheetEntry, error)`
- [x] Implement `GetDeploymentByContractor(ctx, contractorID string) (string, error)`
- [x] Implement `GetDeploymentByContractorAndProject(ctx, contractorID, projectID string) (string, error)` - NEW
- [x] Implement `CreateOrder(ctx, deploymentID, month string) (string, error)` - with Deployment field
- [x] Implement `CreateTimesheetLineItem(ctx, orderID, deploymentID, projectID, hours, proofOfWorks, timesheetIDs, month) (string, error)`
- [x] Implement `CheckOrderExists(ctx, deploymentID, month string) (bool, string, error)`
- [x] Implement `CheckOrderExistsByContractor(ctx, contractorID, month string) (bool, string, error)` - NEW
- [x] Implement `CheckLineItemExists(ctx, orderID, deploymentID string) (bool, string, error)` - NEW
- [x] Add DEBUG logging

## Task 4: Handler
**File**: `pkg/handler/notion/task_order_log.go` (NEW)
- [x] Create `SyncTaskOrderLogs(c *gin.Context)` handler
- [x] Parse `month` query param (required) with validation
- [x] Parse `contractor` query param (optional) - Discord username to filter by specific contractor
- [x] Implement processing logic:
  1. Query approved timesheets
  2. Group by Contractor → Project
  3. For each Contractor:
     - Get first project's deployment for Order
     - Create Order (if not exists) with first deployment
  4. For each Project:
     - Get deployment for contractor+project
     - Aggregate hours, summarize PoW
     - Check if line item exists
     - Create sub-item (if not exists)
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

## Post-Development Refinements

### Issue 1: Wrong Deployment for Subitems
**Problem**: Contractor working on 2 projects had both subitems with same deployment (first project's deployment)
**Solution**:
- Added `GetDeploymentByContractorAndProject` to query by contractor+project pair
- Updated handler to get deployment per project (not per contractor)

### Issue 2: Order Date Field
**Problem**: Date was set to 1st of month
**Solution**: Changed to use current date (`time.Now()`)

### Issue 3: Proof of Works Too Long
**Problem**: Summaries were too verbose with multiple entries
**Solution**: Updated OpenRouter prompt to use scope-based format:
- Max 2 scopes (most significant work areas)
- 3-4 activities per scope
- 2-3 words per activity
- Example: `• Backend Infrastructure: Upload optimization, data retention, search capabilities`

### Issue 4: Subitem Status
**Problem**: Status was "Approved" (should be pending)
**Solution**: Changed subitem status to "Pending Approval"

### Issue 5: Order Missing Deployment
**Problem**: Order (Type=Order) had no Deployment field
**Solution**:
- Get first project's deployment
- Set Order's Deployment to first project's deployment
- Each subitem still has its own project-specific deployment

### Issue 6: Duplicate Line Items on Re-run
**Problem**: Running sync twice created duplicate line items
**Solution**:
- Added `CheckLineItemExists(orderID, deploymentID)` function
- Skip creating line item if already exists for order+deployment

## Status: ✅ COMPLETED & REFINED
All tasks implemented, tested, and refined based on user feedback. Build successful.

## Key Features
- ✅ Prevents duplicate Orders (checks by contractor+month)
- ✅ Prevents duplicate line items (checks by order+deployment)
- ✅ Correct deployment per project for contractors working on multiple projects
- ✅ Concise PoW summaries using scope-based format
- ✅ Free OpenRouter models with rotation and retry
- ✅ DEBUG logging throughout
