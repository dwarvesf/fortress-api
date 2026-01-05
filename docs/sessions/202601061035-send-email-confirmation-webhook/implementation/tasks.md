# Implementation Tasks

## Overview
Add webhook endpoint `POST /webhooks/notion/send-email-confirmation` that receives Notion button click from Order page, reads email content from page body, and sends it to contractor.

## Tasks

### Task 1: Add `GetOrderPageContent` service method ✅ COMPLETED
**File**: `pkg/service/notion/task_order_log.go` (lines 1647-1679)

**Description**: Read page body content from Order page using `FindBlockChildrenByID`, extract text from paragraph blocks.

```go
func (s *TaskOrderLogService) GetOrderPageContent(ctx context.Context, pageID string) (string, error)
```

**Steps**:
1. Call `s.client.FindBlockChildrenByID(ctx, pageID, nil)`
2. Iterate through blocks, extract text from paragraph rich_text
3. Concatenate and return content
4. Add DEBUG logging

---

### Task 2: Add `GetContractorFromOrder` service method ✅ COMPLETED
**File**: `pkg/service/notion/task_order_log.go` (lines 1681-1750)

**Description**: Get contractor info from Order via Sub-items → Deployment → Contractor chain.

```go
func (s *TaskOrderLogService) GetContractorFromOrder(ctx context.Context, orderID string) (contractorID, email, name string, err error)
```

**Steps**:
1. Query Sub-items relation to get first Line Item
2. Get Deployment from Line Item
3. Get Contractor from Deployment
4. Return contractor ID, team email, name
5. Add DEBUG logging

---

### Task 3: Create `notion_task_order.go` webhook handler ✅ COMPLETED
**File**: `pkg/handler/webhook/notion_task_order.go` (NEW)

**Description**: Add handler for send email confirmation webhook.

```go
func (h *handler) HandleNotionTaskOrderSendEmail(c *gin.Context)
```

**Steps**:
1. Parse webhook payload (reuse NotionInvoiceWebhookPayload structure)
2. Handle verification challenge
3. Verify signature
4. Fetch page from Notion
5. Validate Type = "Order"
6. Call `GetOrderPageContent` to read email content
7. Call `GetContractorFromOrder` to get contractor email
8. Build email data and call `SendTaskOrderRawContentMail`
9. Return success/error response
10. Add DEBUG logging throughout

**Additional changes**:
- Added `TaskOrderRawEmail` model in `pkg/model/email.go`
- Added `SendTaskOrderRawContentMail` method in `pkg/service/googlemail/google_mail.go`
- Added interface method in `pkg/service/googlemail/interface.go`

---

### Task 4: Update webhook interface ✅ COMPLETED
**File**: `pkg/handler/webhook/interface.go`

**Description**: Add `HandleNotionTaskOrderSendEmail(c *gin.Context)` to IHandler interface.

---

### Task 5: Add route ✅ COMPLETED
**File**: `pkg/routes/v1.go` (line 87)

**Description**: Register route in webhook group.

```go
webhook.POST("/notion/send-email-confirmation", h.Webhook.HandleNotionTaskOrderSendEmail)
```

---

### Task 6: Test compilation ✅ COMPLETED
**Description**: Verify code compiles without errors.

```bash
go build ./pkg/handler/webhook/... ./pkg/service/notion/... ./pkg/routes/...
```

## Execution Order
1. Task 1 (GetOrderPageContent)
2. Task 2 (GetContractorFromOrder)
3. Task 3 (Handler)
4. Task 4 (Interface)
5. Task 5 (Route)
6. Task 6 (Compile test)
