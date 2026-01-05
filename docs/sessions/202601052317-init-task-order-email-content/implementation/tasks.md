# Implementation Tasks

## Overview
Add email confirmation content generation to `InitTaskOrderLogs` endpoint.

## Tasks

### Task 1: Add `AppendBlocksToPage` service method ✅ COMPLETED
**File**: `pkg/service/notion/task_order_log.go`

**Description**: Add method to append text content as paragraph blocks to a Notion page using go-notion `AppendBlockChildren`.

```go
func (s *TaskOrderLogService) AppendBlocksToPage(ctx context.Context, pageID string, content string) error
```

**Steps**:
1. Split content by newlines
2. Create paragraph blocks with rich text for each line
3. Call `s.client.AppendBlockChildren(ctx, pageID, children)`
4. Add DEBUG logging

---

### Task 2: Add `GenerateConfirmationContent` service method ✅ COMPLETED
**File**: `pkg/service/notion/task_order_log.go`

**Description**: Generate plain text confirmation content for a contractor.

```go
func (s *TaskOrderLogService) GenerateConfirmationContent(contractorName, month string, clients []model.TaskOrderClient) string
```

**Steps**:
1. Format month as "January 2006"
2. Calculate period end day
3. Build plain text content with:
   - Greeting
   - Month/period info
   - Clients list
   - Confirmation request
4. Return as string

---

### Task 3: Update `InitTaskOrderLogs` handler ✅ COMPLETED
**File**: `pkg/handler/notion/task_order_log.go`

**Description**: Modify handler to generate and append email content after creating Line Items.

**Steps**:
1. After Line Item loop for each contractor, collect client info from deployments
2. Use existing `GetClientInfo` method (already used in `SendTaskOrderConfirmation`)
3. Apply Vietnam → "Dwarves LLC (USA)" replacement
4. Deduplicate clients
5. Call `GenerateConfirmationContent`
6. Call `AppendBlocksToPage` with Order page ID
7. Add DEBUG logging
8. Track `contentGenerated` count in response

---

### Task 4: Test compilation ✅ COMPLETED
**Description**: Verify code compiles without errors.

```bash
go build ./pkg/handler/notion/... ./pkg/service/notion/...
```

## Execution Order
1. Task 1 (AppendBlocksToPage)
2. Task 2 (GenerateConfirmationContent)
3. Task 3 (Update handler)
4. Task 4 (Compile test)
