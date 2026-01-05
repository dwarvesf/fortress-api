# Implementation Status

## Status: COMPLETED

## Completed Tasks

### Task 1: Add `GetOrderPageContent` service method
- **File**: `pkg/service/notion/task_order_log.go` (lines 1647-1679)
- Reads page body content from Order page using `FindBlockChildrenByID`
- Extracts text from paragraph blocks
- Returns concatenated content string

### Task 2: Add `GetContractorFromOrder` service method
- **File**: `pkg/service/notion/task_order_log.go` (lines 1681-1750)
- Gets contractor info via Sub-items → Deployment → Contractor chain
- Returns contractor ID, team email, and name

### Task 3: Create `notion_task_order.go` webhook handler
- **File**: `pkg/handler/webhook/notion_task_order.go` (NEW)
- Handles `POST /webhooks/notion/send-email-confirmation`
- Parses Notion button webhook payload
- Validates Type = "Order"
- Reads email content from page body
- Gets contractor via relation chain
- Sends email using `SendTaskOrderRawContentMail`

### Task 4: Update webhook interface
- **File**: `pkg/handler/webhook/interface.go`
- Added `HandleNotionTaskOrderSendEmail(c *gin.Context)` to interface

### Task 5: Add route
- **File**: `pkg/routes/v1.go` (line 87)
- Registered `POST /webhooks/notion/send-email-confirmation`

### Task 6: Test compilation
- `go build ./pkg/handler/webhook/... ./pkg/service/notion/... ./pkg/routes/...` - Success

## Additional Changes

### New model for raw email content
- **File**: `pkg/model/email.go`
- Added `TaskOrderRawEmail` struct with `ContractorName`, `TeamEmail`, `Month`, `RawContent` fields

### New email service method
- **File**: `pkg/service/googlemail/google_mail.go` (lines 471-518)
- Added `SendTaskOrderRawContentMail` method
- Sends email with raw content from Order page body (converted to HTML)
- Uses accounting alias for sending

### Updated interface
- **File**: `pkg/service/googlemail/interface.go`
- Added `SendTaskOrderRawContentMail(data *model.TaskOrderRawEmail) error`

## Summary

The webhook endpoint `POST /webhooks/notion/send-email-confirmation` now:
1. Receives Notion button click payload with Order page ID
2. Validates page Type = "Order"
3. Reads email content from page body (paragraph blocks)
4. Gets contractor via: Order → Sub-items → Deployment → Contractor
5. Sends email to contractor's team email with page content
6. Returns success response with details

## Flow

```
Webhook received (button click on Order page)
  ↓
Parse payload → get page ID
  ↓
Fetch page from Notion
  ↓
Validate Type = "Order"
  ↓
Read page body content (email text)
  ↓
Get Sub-items → first Line Item → Deployment → Contractor
  ↓
Get Contractor Team Email
  ↓
Send email with page body content
  ↓
Return success/failure
```
