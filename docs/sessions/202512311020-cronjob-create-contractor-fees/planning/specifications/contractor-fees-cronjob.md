# Technical Specification: Contractor Fees Cronjob

## Overview

This specification defines the implementation for an automated cronjob endpoint that creates Contractor Fee entries from approved Task Order Logs.

**Endpoint**: `POST /cronjobs/contractor-fees`

**Purpose**: Process Task Order Logs with Status="Approved" and automatically create corresponding Contractor Fee entries with matching Contractor Rates.

## API Contract

### HTTP Method

`POST`

### URL Path

`/api/v1/cronjobs/contractor-fees`

### Authentication

- **Required**: Bearer token authentication
- **Permission**: `model.PermissionCronjobExecute`
- **Bypass**: Auth bypassed in local environment (per existing pattern)

### Request

**Headers**:
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Body**: None (no request body required)

**Query Parameters**: None (processes all approved orders)

### Response

**Success Response** (HTTP 200):

```json
{
  "data": {
    "contractor_fees_created": 5,
    "orders_processed": 8,
    "orders_skipped": 2,
    "errors": 1,
    "details": [
      {
        "order_page_id": "2b964b29-b84c-800c-bb63-fe28a4546f23",
        "contractor_name": "John Doe",
        "contractor_page_id": "2b864b29-b84c-8079-abcd-1234567890ab",
        "contractor_fee_page_id": "2c264b29-b84c-8037-xyz-abcdef123456",
        "status": "created",
        "reason": null
      },
      {
        "order_page_id": "2b964b29-b84c-800c-1111-222233334444",
        "contractor_name": "Jane Smith",
        "contractor_page_id": "2b864b29-b84c-8079-5555-666677778888",
        "contractor_fee_page_id": null,
        "status": "skipped",
        "reason": "contractor fee already exists"
      },
      {
        "order_page_id": "2b964b29-b84c-800c-9999-aaabbbcccddd",
        "contractor_name": "Bob Wilson",
        "contractor_page_id": "2b864b29-b84c-8079-eeee-fffggghhhjjj",
        "contractor_fee_page_id": null,
        "status": "error",
        "reason": "no active contractor rate found for date 2025-12-15"
      }
    ]
  },
  "error": null,
  "message": "ok"
}
```

**Field Descriptions**:

| Field | Type | Description |
|-------|------|-------------|
| `contractor_fees_created` | number | Count of new fees created |
| `orders_processed` | number | Total orders evaluated |
| `orders_skipped` | number | Orders skipped (fee exists, missing data) |
| `errors` | number | Orders with errors |
| `details` | array | Per-order processing results |
| `details[].order_page_id` | string | Task Order Log page ID |
| `details[].contractor_name` | string | Contractor name |
| `details[].contractor_page_id` | string | Contractor page ID |
| `details[].contractor_fee_page_id` | string | Created fee page ID (null if skipped/error) |
| `details[].status` | string | "created", "skipped", or "error" |
| `details[].reason` | string | Reason for skip/error (null if created) |

**Error Response** (HTTP 500):

```json
{
  "data": null,
  "error": "failed to query approved orders: notion API error",
  "message": ""
}
```

## Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ STEP 1: Query Approved Orders                                  │
│ Service: TaskOrderLogService.QueryApprovedOrders()             │
│ Filter: Type=Order AND Status=Approved                         │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────────┐
│ STEP 2: For Each Approved Order                                │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────────┐
│ STEP 2a: Extract Contractor Info                               │
│ Property: Contractor rollup (property ID: q?kW)                 │
│ Action: Get contractor page ID and name                        │
│ Skip if: Rollup is empty                                       │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────────┐
│ STEP 2b: Find Matching Contractor Rate                         │
│ Service: ContractorRatesService.FindActiveRateByContractor()   │
│ Filter: Contractor=<id> AND Status=Active                      │
│         AND Start Date <= Order Date <= End Date (or nil)      │
│ Skip if: No matching rate found                                │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────────┐
│ STEP 2c: Check If Fee Exists (Idempotency)                     │
│ Service: ContractorFeesService.CheckFeeExistsByTaskOrder()     │
│ Query: Task Order Log relation = <order_page_id>               │
│ Skip if: Fee already exists                                    │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────────┐
│ STEP 2d: Create Contractor Fee                                 │
│ Service: ContractorFeesService.CreateContractorFee()           │
│ Relations:                                                      │
│   - Task Order Log → <order_page_id>                           │
│   - Contractor Rate → <rate_page_id>                           │
│ Properties:                                                     │
│   - Payment Status = "New"                                     │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────────┐
│ STEP 2e: Update Order Status                                   │
│ Service: TaskOrderLogService.UpdateOrderStatus()               │
│ Update: Status = "Completed"                                   │
│ Note: Log error but don't fail if update fails                 │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────────┐
│ STEP 3: Return Response                                        │
│ Aggregate: Statistics and per-order details                    │
└─────────────────────────────────────────────────────────────────┘
```

## Service Layer Specification

### TaskOrderLogService Extensions

**File**: `pkg/service/notion/task_order_log.go`

#### New Data Structure

```go
// ApprovedOrderData represents an approved Task Order Log entry
type ApprovedOrderData struct {
    PageID           string    // Task Order Log page ID
    ContractorPageID string    // From Contractor rollup (property ID: q?kW)
    ContractorName   string    // From contractor page Full Name
    Date             time.Time // From Date property (property ID: Ri:O)
    FinalHoursWorked float64   // From Final Hours Worked formula (property ID: ;J>Y)
    ProofOfWorks     string    // From Proof of Works rich text (property ID: hlty)
}
```

#### Method: QueryApprovedOrders

```go
// QueryApprovedOrders queries all Task Order Log entries with Type=Order and Status=Approved
//
// Returns:
//   - []*ApprovedOrderData: Slice of approved orders with extracted data
//   - error: Error if query fails
//
// Filters:
//   - Type = "Order"
//   - Status = "Approved"
//
// Extracts:
//   - Contractor page ID from Contractor rollup
//   - Contractor name by fetching contractor page
//   - Date, Final Hours Worked, Proof of Works
//
// Handles:
//   - Pagination (page size 100)
//   - Empty rollup arrays
//   - Missing properties
func (s *TaskOrderLogService) QueryApprovedOrders(ctx context.Context) ([]*ApprovedOrderData, error)
```

**Implementation Notes**:
- Use Notion query filter with AND condition
- Handle rollup property extraction (array of relations)
- Fetch contractor name from contractor page (Full Name property)
- Parse date string to time.Time
- Extract formula number and rich text properties

**Notion Query**:
```go
query := &nt.DatabaseQuery{
    Filter: &nt.DatabaseQueryFilter{
        And: []nt.DatabaseQueryFilter{
            {
                Property: "Type",
                DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
                    Select: &nt.SelectDatabaseQueryFilter{
                        Equals: "Order",
                    },
                },
            },
            {
                Property: "Status",
                DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
                    Select: &nt.SelectDatabaseQueryFilter{
                        Equals: "Approved",
                    },
                },
            },
        },
    },
    PageSize: 100,
}
```

#### Method: UpdateOrderStatus

```go
// UpdateOrderStatus updates the status field of a Task Order Log entry
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - orderPageID: Page ID of the order to update
//   - newStatus: New status value (e.g., "Completed")
//
// Returns:
//   - error: Error if update fails, nil on success
//
// Updates:
//   - Status property (property ID: LnUY)
func (s *TaskOrderLogService) UpdateOrderStatus(ctx context.Context, orderPageID, newStatus string) error
```

**Implementation Notes**:
- Use Notion UpdatePage API
- Update only the Status property
- Return error if API call fails

**Notion Update**:
```go
updateParams := nt.UpdatePageParams{
    DatabasePageProperties: nt.DatabasePageProperties{
        "Status": nt.DatabasePageProperty{
            Type: nt.DBPropTypeSelect,
            Select: &nt.SelectOptions{
                Name: newStatus,
            },
        },
    },
}
```

### ContractorRatesService Extensions

**File**: `pkg/service/notion/contractor_rates.go`

#### Method: FindActiveRateByContractor

```go
// FindActiveRateByContractor finds the active contractor rate for a given contractor at a specific date
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - contractorPageID: Contractor page ID
//   - orderDate: Date to match against rate's date range
//
// Returns:
//   - *ContractorRateData: Matching rate data, nil if not found
//   - error: Error if query fails or no matching rate found
//
// Filters:
//   - Contractor relation = contractorPageID
//   - Status = "Active"
//   - Start Date <= orderDate <= End Date (or End Date is nil)
//
// Logic:
//   - Query for active rates for contractor
//   - Filter by date range in-memory
//   - Handle nil End Date (ongoing rate)
//   - Return first matching rate
func (s *ContractorRatesService) FindActiveRateByContractor(ctx context.Context, contractorPageID string, orderDate time.Time) (*ContractorRateData, error)
```

**Implementation Notes**:
- Query Notion with contractor and status filters
- Fetch all active rates for contractor
- Filter by date range in application code
- Nil End Date means rate is still active (no end date set)

**Date Comparison Logic**:
```go
// Check: Start Date <= orderDate
if startDate != nil && startDate.After(orderDate) {
    continue // Skip this rate
}

// Check: orderDate <= End Date OR End Date is nil
if endDate != nil && endDate.Before(orderDate) {
    continue // Skip this rate
}

// Rate is valid for this date
return rateData, nil
```

### ContractorFeesService Extensions

**File**: `pkg/service/notion/contractor_fees.go`

#### Method: CheckFeeExistsByTaskOrder

```go
// CheckFeeExistsByTaskOrder checks if a contractor fee entry exists for a given Task Order Log
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - taskOrderPageID: Task Order Log page ID
//
// Returns:
//   - bool: true if fee exists, false otherwise
//   - string: Existing fee page ID (empty if not found)
//   - error: Error if query fails
//
// Query:
//   - Task Order Log relation contains taskOrderPageID
func (s *ContractorFeesService) CheckFeeExistsByTaskOrder(ctx context.Context, taskOrderPageID string) (bool, string, error)
```

**Implementation Notes**:
- Query Contractor Fees database by Task Order Log relation
- Return first matching page ID if found
- Page size can be 1 (we only need to know if it exists)

**Notion Query**:
```go
query := &nt.DatabaseQuery{
    Filter: &nt.DatabaseQueryFilter{
        Property: "Task Order Log",
        DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
            Relation: &nt.RelationDatabaseQueryFilter{
                Contains: taskOrderPageID,
            },
        },
    },
    PageSize: 1,
}
```

#### Method: CreateContractorFee

```go
// CreateContractorFee creates a new Contractor Fee entry
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - taskOrderPageID: Task Order Log page ID to link
//   - contractorRatePageID: Contractor Rate page ID to link
//
// Returns:
//   - string: Created fee page ID
//   - error: Error if creation fails
//
// Creates entry with:
//   - Task Order Log relation (property ID: MyT\)
//   - Contractor Rate relation (property ID: a_@z)
//   - Payment Status = "New" (property ID: BkF\)
func (s *ContractorFeesService) CreateContractorFee(ctx context.Context, taskOrderPageID, contractorRatePageID string) (string, error)
```

**Implementation Notes**:
- Use Notion CreatePage API
- Set parent to Contractor Fees database
- Create relations to Task Order Log and Contractor Rate
- Set Payment Status to "New" by default

**Notion Create**:
```go
params := nt.CreatePageParams{
    ParentType: nt.ParentTypeDatabase,
    ParentID:   contractorFeesDBID,
    DatabasePageProperties: &nt.DatabasePageProperties{
        "Task Order Log": nt.DatabasePageProperty{
            Type: nt.DBPropTypeRelation,
            Relation: []nt.Relation{
                {ID: taskOrderPageID},
            },
        },
        "Contractor Rate": nt.DatabasePageProperty{
            Type: nt.DBPropTypeRelation,
            Relation: []nt.Relation{
                {ID: contractorRatePageID},
            },
        },
        "Payment Status": nt.DatabasePageProperty{
            Type: nt.DBPropTypeStatus,
            Status: &nt.SelectOptions{
                Name: "New",
            },
        },
    },
}
```

## Handler Implementation

**File**: `pkg/handler/notion/contractor_fees.go` (NEW FILE)

### Handler Signature

```go
// CreateContractorFees godoc
// @Summary Create contractor fees from approved task orders
// @Description Processes approved task order logs and creates contractor fee entries
// @Tags Cronjobs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} view.Response
// @Failure 500 {object} view.Response
// @Router /cronjobs/contractor-fees [post]
func (h *handler) CreateContractorFees(c *gin.Context)
```

### Handler Logic

```go
func (h *handler) CreateContractorFees(c *gin.Context) {
    l := h.logger.Fields(logger.Fields{
        "handler": "Notion",
        "method":  "CreateContractorFees",
    })
    ctx := c.Request.Context()

    // Step 1: Get services
    taskOrderLogService := h.service.Notion.TaskOrderLog
    contractorRatesService := h.service.Notion.ContractorRates
    contractorFeesService := h.service.Notion.ContractorFees

    // Validate services
    // ...

    // Step 2: Query approved orders
    l.Info("querying approved orders")
    approvedOrders, err := taskOrderLogService.QueryApprovedOrders(ctx)
    if err != nil {
        l.Error(err, "failed to query approved orders")
        c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
        return
    }

    l.Info(fmt.Sprintf("found %d approved orders", len(approvedOrders)))

    // Step 3: Process each order
    var (
        feesCreated   = 0
        ordersSkipped = 0
        errors        = 0
        details       = []map[string]any{}
    )

    for _, order := range approvedOrders {
        detail := map[string]any{
            "order_page_id":       order.PageID,
            "contractor_name":     order.ContractorName,
            "contractor_page_id":  order.ContractorPageID,
        }

        // Step 3a: Validate contractor
        if order.ContractorPageID == "" {
            l.Warn(fmt.Sprintf("skipping order %s: no contractor", order.PageID))
            detail["status"] = "skipped"
            detail["reason"] = "contractor not found"
            detail["contractor_fee_page_id"] = nil
            ordersSkipped++
            details = append(details, detail)
            continue
        }

        // Step 3b: Find matching rate
        rate, err := contractorRatesService.FindActiveRateByContractor(ctx, order.ContractorPageID, order.Date)
        if err != nil {
            l.Error(err, fmt.Sprintf("no active rate for contractor %s at date %s", order.ContractorPageID, order.Date))
            detail["status"] = "error"
            detail["reason"] = fmt.Sprintf("no active contractor rate found for date %s", order.Date.Format("2006-01-02"))
            detail["contractor_fee_page_id"] = nil
            errors++
            details = append(details, detail)
            continue
        }

        // Step 3c: Check if fee exists
        exists, existingFeeID, err := contractorFeesService.CheckFeeExistsByTaskOrder(ctx, order.PageID)
        if err != nil {
            l.Error(err, fmt.Sprintf("failed to check fee existence for order %s", order.PageID))
            detail["status"] = "error"
            detail["reason"] = "failed to check fee existence"
            detail["contractor_fee_page_id"] = nil
            errors++
            details = append(details, detail)
            continue
        }

        if exists {
            l.Debug(fmt.Sprintf("fee already exists for order %s: %s", order.PageID, existingFeeID))
            detail["status"] = "skipped"
            detail["reason"] = "contractor fee already exists"
            detail["contractor_fee_page_id"] = existingFeeID
            ordersSkipped++
            details = append(details, detail)
            continue
        }

        // Step 3d: Create fee
        feePageID, err := contractorFeesService.CreateContractorFee(ctx, order.PageID, rate.PageID)
        if err != nil {
            l.Error(err, fmt.Sprintf("failed to create fee for order %s", order.PageID))
            detail["status"] = "error"
            detail["reason"] = "failed to create contractor fee"
            detail["contractor_fee_page_id"] = nil
            errors++
            details = append(details, detail)
            continue
        }

        l.Info(fmt.Sprintf("created contractor fee: %s for order: %s", feePageID, order.PageID))

        // Step 3e: Update order status
        err = taskOrderLogService.UpdateOrderStatus(ctx, order.PageID, "Completed")
        if err != nil {
            l.Error(err, fmt.Sprintf("failed to update order status: %s (fee created: %s)", order.PageID, feePageID))
            // Don't fail - fee is already created
        } else {
            l.Info(fmt.Sprintf("updated order %s status to Completed", order.PageID))
        }

        detail["status"] = "created"
        detail["reason"] = nil
        detail["contractor_fee_page_id"] = feePageID
        feesCreated++
        details = append(details, detail)
    }

    // Step 4: Return response
    l.Info(fmt.Sprintf("processing complete: fees_created=%d skipped=%d errors=%d", feesCreated, ordersSkipped, errors))

    c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
        "contractor_fees_created": feesCreated,
        "orders_processed":        len(approvedOrders),
        "orders_skipped":          ordersSkipped,
        "errors":                  errors,
        "details":                 details,
    }, nil, nil, nil, "ok"))
}
```

## Error Scenarios

| Scenario | HTTP Status | Response | Handler Behavior | User Impact |
|----------|-------------|----------|------------------|-------------|
| No approved orders found | 200 | Success with zero counts | Return empty results | None (expected) |
| Notion API unavailable | 500 | Error response | Return error immediately | Cronjob fails, retry needed |
| Contractor rollup empty | 200 | Success with skip in details | Skip order, continue processing | Order not processed, logged |
| No matching contractor rate | 200 | Success with error in details | Skip order, continue processing | Order not processed, logged |
| Contractor fee already exists | 200 | Success with skip in details | Skip order (idempotent), continue | No duplicate created |
| Fee creation fails | 200 | Success with error in details | Skip order, continue processing | Order not processed, logged |
| Status update fails | 200 | Success (fee created) | Log error, continue | Fee created, status not updated |
| Service not configured | 500 | Error response | Return error immediately | Cronjob fails, config issue |

## Configuration

### Required Notion Database IDs

These should already be configured in `pkg/config/config.go`:

```go
type Notion struct {
    Secret    string
    Databases struct {
        TaskOrderLog     string // Should be: 2b964b29-b84c-801e-ab9e-000b0662b987
        ContractorRates  string // Should be: 2c464b29-b84c-80cf-bef6-000b42bce15e
        ContractorFees   string // Should be: 2c264b29-b84c-8037-807c-000bf6d0792c
        // ... other databases
    }
}
```

**Validation**: Services should validate database IDs are not empty in initialization.

## Logging Strategy

### Log Levels

**DEBUG**:
- Notion property extraction details
- Rollup array contents
- Date comparison logic
- Existing fee found (idempotency)

**INFO**:
- Cronjob started
- Number of approved orders found
- Each contractor fee created
- Order status updated
- Processing complete with statistics

**WARNING**:
- Order skipped (missing contractor, missing data)

**ERROR**:
- Failed to query approved orders
- Failed to find contractor rate
- Failed to create contractor fee
- Failed to update order status (non-critical)

### Example Log Output

```
[INFO] querying approved orders
[INFO] found 8 approved orders
[DEBUG] processing order: 2b964b29-b84c-800c-bb63-fe28a4546f23
[DEBUG] contractor rollup: [2b864b29-b84c-8079-abcd-1234567890ab]
[INFO] created contractor fee: 2c264b29-b84c-8037-xyz-abcdef123456 for order: 2b964b29-b84c-800c-bb63-fe28a4546f23
[INFO] updated order 2b964b29-b84c-800c-bb63-fe28a4546f23 status to Completed
[WARN] skipping order 2b964b29-b84c-800c-1111-222233334444: contractor fee already exists
[ERROR] no active rate for contractor 2b864b29-b84c-8079-eeee-fffggghhhjjj at date 2025-12-15
[INFO] processing complete: fees_created=5 skipped=2 errors=1
```

## Test Scenarios

### Unit Tests

**TaskOrderLogService**:
1. QueryApprovedOrders - returns orders with correct filter
2. QueryApprovedOrders - handles pagination
3. QueryApprovedOrders - extracts contractor rollup correctly
4. QueryApprovedOrders - handles empty rollup array
5. UpdateOrderStatus - updates status successfully
6. UpdateOrderStatus - returns error on API failure

**ContractorRatesService**:
1. FindActiveRateByContractor - finds rate within date range
2. FindActiveRateByContractor - finds rate with nil end date (ongoing)
3. FindActiveRateByContractor - skips rate before start date
4. FindActiveRateByContractor - skips rate after end date
5. FindActiveRateByContractor - returns error if no matching rate

**ContractorFeesService**:
1. CheckFeeExistsByTaskOrder - returns true if fee exists
2. CheckFeeExistsByTaskOrder - returns false if fee doesn't exist
3. CreateContractorFee - creates fee with correct relations
4. CreateContractorFee - sets Payment Status to "New"
5. CreateContractorFee - returns created page ID

### Integration Tests

**Handler**:
1. Happy path - creates fees for all approved orders
2. Idempotency - skips existing fees
3. Missing contractor - skips order with warning
4. No matching rate - skips order with error
5. Batch processing - processes multiple orders with mixed results
6. Status update failure - creates fee but logs error

### Manual Test Cases

**Test Case 1: Single Approved Order**
- Precondition: 1 approved order, matching active rate exists
- Expected: 1 fee created, order status updated to Completed
- Response: fees_created=1, orders_processed=1, orders_skipped=0, errors=0

**Test Case 2: Idempotency Check**
- Precondition: 1 approved order, fee already exists
- Expected: No duplicate fee created, order skipped
- Response: fees_created=0, orders_processed=1, orders_skipped=1, errors=0

**Test Case 3: Missing Contractor Rate**
- Precondition: 1 approved order, no active rate for contractor
- Expected: Order skipped with error
- Response: fees_created=0, orders_processed=1, orders_skipped=0, errors=1

**Test Case 4: Multiple Orders**
- Precondition: 5 approved orders (3 new, 1 existing fee, 1 missing rate)
- Expected: 3 fees created, 1 skipped, 1 error
- Response: fees_created=3, orders_processed=5, orders_skipped=1, errors=1

## Routes Configuration

**File**: `pkg/routes/v1.go`

Add to cronjob group:

```go
cronjob := r.Group("/cronjobs")
{
    // ... existing routes
    cronjob.POST("/contractor-fees", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Notion.CreateContractorFees)
}
```

**Line**: After line 62 (after sync-task-order-logs)

## Interface Updates

**File**: `pkg/handler/notion/interface.go`

Add method to IHandler:

```go
type IHandler interface {
    // ... existing methods
    SyncTaskOrderLogs(c *gin.Context)
    CreateContractorFees(c *gin.Context) // NEW
}
```

## Notion Property Reference

### Task Order Log Properties

| Property Name | Property ID | Type | Purpose |
|--------------|-------------|------|---------|
| Type | `pXKA` | Select | Filter by "Order" |
| Status | `LnUY` | Select | Filter by "Approved", update to "Completed" |
| Date | `Ri:O` | Date | For rate date range matching |
| Contractor | `q?kW` | Rollup | Extract contractor page ID |
| Final Hours Worked | `;J>Y` | Formula | Reference data |
| Proof of Works | `hlty` | Rich Text | Reference data |

### Contractor Rates Properties

| Property Name | Property ID | Type | Purpose |
|--------------|-------------|------|---------|
| Contractor | (relation) | Relation | Match by contractor ID |
| Status | (status) | Status | Filter by "Active" |
| Start Date | (date) | Date | Date range validation |
| End Date | (date) | Date | Date range validation (nil = ongoing) |

### Contractor Fees Properties

| Property Name | Property ID | Type | Purpose |
|--------------|-------------|------|---------|
| Task Order Log | `MyT\` | Relation | Link to order |
| Contractor Rate | `a_@z` | Relation | Link to rate |
| Payment Status | `BkF\` | Status | Set to "New" |

## Dependencies

### Internal Dependencies

- `pkg/service/notion/task_order_log.go`
- `pkg/service/notion/contractor_rates.go`
- `pkg/service/notion/contractor_fees.go`
- `pkg/handler/notion/interface.go`
- `pkg/routes/v1.go`
- `pkg/config/config.go`

### External Dependencies

- `github.com/dstotijn/go-notion` - Notion API client
- `github.com/gin-gonic/gin` - HTTP framework
- `github.com/dwarvesf/fortress-api/pkg/logger` - Logging
- `github.com/dwarvesf/fortress-api/pkg/view` - Response formatting

## Performance Considerations

### Notion API Rate Limits

- Notion imposes rate limits (typically 3 requests per second)
- Use pagination with reasonable page size (100)
- Process orders sequentially to avoid overwhelming API

### Expected Load

- Typical: 10-50 approved orders per execution
- Worst case: 100-200 orders
- Execution time: ~1-2 seconds per order (API latency)
- Total time: <5 minutes for 200 orders

### Optimization Opportunities

- Batch fetch contractor rates (reduce API calls)
- Cache contractor names (reduce page fetches)
- Parallel processing with rate limiting (future enhancement)

## Rollback Strategy

### Code Rollback

1. Comment out route in `pkg/routes/v1.go`:
   ```go
   // cronjob.POST("/contractor-fees", ...) // Disabled - rollback
   ```
2. Deploy updated code
3. Cronjob stops executing

### Data Rollback

- Contractor Fees created can be manually deleted from Notion if needed
- Task Order Log status can be manually reverted from "Completed" to "Approved"
- No database migrations or schema changes involved

## Future Enhancements

1. **Query Parameters**: Add optional filters for contractor or date range
2. **Dry Run Mode**: Add query param to preview without creating fees
3. **Batch Operations**: Use Notion batch API for better performance
4. **Notifications**: Send Discord/Slack notification on completion
5. **Retry Logic**: Automatic retry for failed fee creations
6. **Metrics**: Track execution time, success rate, error types
