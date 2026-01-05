# Specification: Handler Layer Changes

## Overview

Add new endpoint to initialize Task Order Logs for all active deployments.

## New Endpoint

### `POST /cronjobs/init-task-order-logs`

**File**: `pkg/handler/notion/task_order_log.go`

**Parameters**:
- `month` (query, required): Target month in YYYY-MM format

**Process**:
1. Validate month parameter
2. Query active deployments via `QueryActiveDeploymentsByMonth(ctx, month, "")`
3. Group deployments by contractor using `groupDeploymentsByContractor()`
4. For each contractor:
   a. Check if Order exists via `CheckOrderExistsByContractor()`
   b. If not, create Order via `CreateOrder(ctx, month)`
   c. For each deployment of that contractor:
      - Check if Line Item exists via `CheckLineItemExists(ctx, orderID, deploymentID)`
      - If not, create empty Line Item via `CreateEmptyTimesheetLineItem()`
5. Return response with counts

**Response**:
```json
{
  "data": {
    "month": "2025-12",
    "orders_created": 5,
    "line_items_created": 12,
    "deployments_processed": 20,
    "skipped": 8,
    "details": [
      {
        "contractor_id": "abc123",
        "order_page_id": "xyz789",
        "line_items_created": 3,
        "deployments": ["dep1", "dep2", "dep3"]
      }
    ]
  }
}
```

## Route Registration

**File**: `pkg/routes/v1.go`

Add route:
```go
cronjobsGroup.POST("/init-task-order-logs", notionHandler.InitTaskOrderLogs)
```

## Helper Function

Reuse existing `groupDeploymentsByContractor()` from task_order_log.go handler.
