# Specification: Task Order Log Service Updates

## Overview
Add new method to update Task Order Log status after payout creation.

## File
`pkg/service/notion/task_order_log.go`

## New Method

### UpdateOrderStatus

```go
// UpdateOrderStatus updates the Status of a Task Order Log entry
func (s *TaskOrderLogService) UpdateOrderStatus(ctx context.Context, pageID, status string) error
```

**Parameters:**
- `pageID`: Task Order Log page ID
- `status`: New status value (e.g., "Pending", "Completed")

**Behavior:**
1. Log debug message with pageID and status
2. Create UpdatePageParams with Status property
3. Call Notion API to update page
4. Return error if update fails

**Reference Implementation:**
Similar to `ContractorFeesService.UpdatePaymentStatus` in `contractor_fees.go:376`
