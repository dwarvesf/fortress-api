# Specification: Service Layer Changes

## File: `pkg/service/notion/task_order_log.go`

### 1. New Method: `GetLineItemDetails`

**Purpose**: Fetch existing line item data for comparison

**Signature**:
```go
func (s *TaskOrderLogService) GetLineItemDetails(ctx context.Context, lineItemID string) (*LineItemDetails, error)
```

**Return struct**:
```go
type LineItemDetails struct {
    PageID       string
    Hours        float64
    TimesheetIDs []string
    Status       string
}
```

**Logic**:
1. Query Notion page by ID
2. Extract `Line Item Hours` number property
3. Extract `Timesheet` relation IDs
4. Extract `Status` select property
5. Return structured data

### 2. New Method: `UpdateTimesheetLineItem`

**Purpose**: Update existing line item with new data

**Signature**:
```go
func (s *TaskOrderLogService) UpdateTimesheetLineItem(
    ctx context.Context,
    lineItemID string,
    orderID string,
    hours float64,
    proofOfWorks string,
    timesheetIDs []string,
) error
```

**Logic**:
1. Build update properties:
   - `Line Item Hours`: new hours value
   - `Proof of Works`: new summarized text
   - `Timesheet`: new relation IDs
   - `Status`: "Pending Approval"
2. Call Notion UpdatePage API
3. Update parent Order status to "Pending Approval"
4. Add DEBUG logging throughout

### 3. Modify Method: `CreateOrder`

**Current signature**:
```go
func (s *TaskOrderLogService) CreateOrder(ctx context.Context, deploymentID, month string) (string, error)
```

**New signature**:
```go
func (s *TaskOrderLogService) CreateOrder(ctx context.Context, month string) (string, error)
```

**Changes**:
- Remove `deploymentID` parameter
- Remove Deployment relation from properties
- Keep Type, Status, Date properties

### 4. Existing Method: `UpdateOrderStatus`

**Check if exists**: Line 1006 shows this method exists

**Ensure it sets**: Status to "Pending Approval"
