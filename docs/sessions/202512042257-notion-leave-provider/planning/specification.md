# Notion Leave Provider Specification

## Overview

This specification details the implementation of Notion as the leave request provider.

## Components

### 1. LeaveRequest Struct

```go
// pkg/service/notion/leave.go

type LeaveRequest struct {
    PageID       string
    Reason       string
    EmployeeID   string    // Relation page ID
    Email        string    // Rollup value
    LeaveType    string    // "Off" or "Remote"
    StartDate    time.Time
    EndDate      time.Time
    Shift        string    // "Full day", "Morning", "Afternoon"
    Status       string    // "Pending", "Approved", "Rejected"
    ApprovedByID string    // Relation page ID
    ApprovedAt   *time.Time
}
```

### 2. LeaveService Interface

```go
// pkg/service/notion/leave.go

type LeaveService struct {
    client *notion.Client
    config *config.Config
    store  *store.Store
    logger *zap.Logger
}

// GetLeaveRequest fetches a leave request by page ID
func (s *LeaveService) GetLeaveRequest(ctx context.Context, pageID string) (*LeaveRequest, error)

// UpdateLeaveStatus updates the status and approval fields
func (s *LeaveService) UpdateLeaveStatus(ctx context.Context, pageID, status, approverEmail string) error

// GetEmployeeByEmail looks up employee by email from rollup
func (s *LeaveService) GetEmployeeByEmail(ctx context.Context, email string) (*model.Employee, error)
```

### 3. Webhook Handler

```go
// pkg/handler/webhook/notion_leave.go

type NotionLeaveWebhook struct {
    Type   string `json:"type"`   // "page.created", "page.updated"
    PageID string `json:"page_id"`
    // Additional fields as needed
}

// HandleNotionLeave processes incoming Notion webhooks
func (h *handler) HandleNotionLeave(c *gin.Context)

// handleLeaveCreated validates new leave request and sends Discord notification
func (h *handler) handleLeaveCreated(ctx context.Context, pageID string) error

// handleLeaveUpdated handles status changes
func (h *handler) handleLeaveUpdated(ctx context.Context, pageID string, previousStatus string) error
```

### 4. Discord Integration

Reuse existing Discord patterns from NocoDB handler:

```go
// Send notification with Approve/Reject buttons
func (h *handler) sendLeaveNotification(ctx context.Context, leave *LeaveRequest, employee *model.Employee) error

// Handle button interactions
func (h *handler) HandleLeaveApproveButton(c *gin.Context)
func (h *handler) HandleLeaveRejectButton(c *gin.Context)
```

Button payload format:
```
approve_leave|{page_id}
reject_leave|{page_id}
```

### 5. Database Persistence

On approval, create `OnLeaveRequest` record:

```go
onLeaveRequest := &model.OnLeaveRequest{
    EmployeeID:   employee.ID,
    Title:        leave.Reason,
    Type:         mapLeaveType(leave.LeaveType), // "off" or "remote"
    StartDate:    leave.StartDate,
    EndDate:      leave.EndDate,
    Shift:        mapShift(leave.Shift),
    NotionPageID: &leave.PageID,  // New field
}
```

### 6. Configuration

Environment variables:
```env
NOTION_LEAVE_DB_ID=2bfb69f8f5738101a121c4464e7a901b
NOTION_LEAVE_DATA_SOURCE_ID=2bfb69f8-f573-8194-b8fe-000b7b278e8d
```

Config struct addition:
```go
// pkg/config/config.go

type NotionConfig struct {
    // existing fields...
    LeaveDBID         string `env:"NOTION_LEAVE_DB_ID"`
    LeaveDataSourceID string `env:"NOTION_LEAVE_DATA_SOURCE_ID"`
}
```

## API Endpoints

### Webhook Endpoint

```
POST /webhooks/notion/leave
Content-Type: application/json

{
    "type": "page.created" | "page.updated",
    "page_id": "uuid-string",
    "timestamp": "ISO8601"
}
```

### Discord Button Callback

```
POST /webhooks/discord/interactions
Content-Type: application/json

// Discord interaction payload with custom_id: "approve_leave|{page_id}"
```

## Property Extraction

Use existing Notion client helpers:

```go
// Extract title property
reason := notion.ExtractTitleText(page.Properties["Reason"])

// Extract rollup email
email := notion.ExtractRollupText(page.Properties["Email"])

// Extract select property
status := notion.ExtractSelectName(page.Properties["Status"])
leaveType := notion.ExtractSelectName(page.Properties["Leave Type"])
shift := notion.ExtractSelectName(page.Properties["Shift"])

// Extract date property
startDate := notion.ExtractDate(page.Properties["Start Date"])
endDate := notion.ExtractDate(page.Properties["End Date"])

// Extract relation IDs
employeeIDs := notion.ExtractRelationIDs(page.Properties["Employee"])
approverIDs := notion.ExtractRelationIDs(page.Properties["Approved By"])
```

## Status Update

Update Notion page when approving/rejecting via Discord:

```go
func (s *LeaveService) UpdateLeaveStatus(ctx context.Context, pageID, status, approverEmail string) error {
    properties := map[string]interface{}{
        "Status": notion.SelectProperty{Name: status},
    }

    if status == "Approved" {
        // Look up approver by email to get their Contractor page ID
        approverPageID := s.getContractorPageIDByEmail(ctx, approverEmail)
        properties["Approved By"] = notion.RelationProperty{PageIDs: []string{approverPageID}}
        properties["Approved at"] = notion.DateProperty{Start: time.Now()}
    }

    return s.client.UpdatePage(ctx, pageID, properties)
}
```

## Mapping Functions

```go
// Map Notion leave type to internal type
func mapLeaveType(notionType string) string {
    switch notionType {
    case "Off":
        return "off"
    case "Remote":
        return "remote"
    default:
        return "off"
    }
}

// Map Notion shift to internal shift
func mapShift(notionShift string) string {
    switch notionShift {
    case "Full day":
        return "full-day"
    case "Morning":
        return "morning"
    case "Afternoon":
        return "afternoon"
    default:
        return "full-day"
    }
}
```

## Error Handling

1. **Employee not found**: Log warning, skip processing
2. **Invalid date range**: Reject with Discord notification
3. **Notion API errors**: Retry with exponential backoff
4. **Database errors**: Log error, return 500

## Validation Rules

1. **Start Date**: Must not be in the past (with 1-day grace period)
2. **End Date**: Must be >= Start Date
3. **Employee**: Must exist in employees table (matched by email)
4. **Leave Type**: Must be "Off" or "Remote"
5. **Shift**: Must be "Full day", "Morning", or "Afternoon"
