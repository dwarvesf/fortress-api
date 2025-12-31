# Implementation Tasks: Cronjob Refund Payouts

## Task 1: Create RefundRequestsService

**File**: `pkg/service/notion/refund_requests.go`

- [x] Create `RefundRequestsService` struct
- [x] Create `ApprovedRefundData` struct
- [x] Implement `NewRefundRequestsService(cfg, logger)`
- [x] Implement `QueryApprovedRefunds(ctx) ([]*ApprovedRefundData, error)`
  - Query Refund Requests with Status=Approved
  - Extract: PageID, RefundID, Amount, Currency, ContractorPageID, Reason, DateApproved
- [x] Add DEBUG logging

## Task 2: Extend ContractorPayoutsService

**File**: `pkg/service/notion/contractor_payouts.go`

- [x] Create `CreateRefundPayoutInput` struct
- [x] Implement `CheckPayoutExistsByRefundRequest(ctx, refundRequestPageID) (bool, string, error)`
  - Query by "Refund Request" relation
- [x] Implement `CreateRefundPayout(ctx, input) (string, error)`
  - Set Type: "Refund"
  - Set Direction: "Outgoing"
  - Set Refund Request relation
- [x] Add DEBUG logging

## Task 3: Update Service Initialization

**Files**:
- `pkg/service/notion/notion_services.go`
- `pkg/service/service.go`

- [x] Add `RefundRequests *RefundRequestsService` field to NotionServices
- [x] Initialize RefundRequestsService in service.go

## Task 4: Add processRefundPayouts Handler

**File**: `pkg/handler/notion/contractor_payouts.go`

- [x] Implement `processRefundPayouts(c, l, payoutType)`
  - Query approved refund requests
  - For each: validate, check idempotency, create payout
  - Do NOT update refund status
- [x] Update switch in `CreateContractorPayouts` to call `processRefundPayouts`
- [x] Add DEBUG logging

## Task 5: Build & Verify

- [x] Run `go build ./...`
- [x] Verify no compilation errors
