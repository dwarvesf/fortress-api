# Implementation Status: Cronjob Refund Payouts

**Session ID**: `202512311436-cronjob-refund-payouts`
**Date**: 2025-12-31
**Status**: IMPLEMENTATION COMPLETE

---

## Summary

Added refund payout type support to the existing contractor payouts cronjob endpoint.

**Endpoint**: `POST /api/v1/cronjobs/contractor-payouts?type=refund`

**Build Status**: Passed (`go build ./...`)

---

## Completed Tasks

### Task 1: Create RefundRequestsService
- **File**: `pkg/service/notion/refund_requests.go` (NEW)
- **Status**: Complete
- **Changes**:
  - Created `RefundRequestsService` struct
  - Created `ApprovedRefundData` struct
  - Implemented `NewRefundRequestsService(cfg, logger)`
  - Implemented `QueryApprovedRefunds(ctx)` - queries refunds with Status=Approved
  - Added DEBUG logging

### Task 2: Extend ContractorPayoutsService
- **File**: `pkg/service/notion/contractor_payouts.go`
- **Status**: Complete
- **Changes**:
  - Added `CreateRefundPayoutInput` struct
  - Added `CheckPayoutExistsByRefundRequest(ctx, refundRequestPageID)` for idempotency
  - Added `CreateRefundPayout(ctx, input)` - creates payout with Type=Refund, Direction=Outgoing
  - Added DEBUG logging

### Task 3: Update Service Initialization
- **Files**: `pkg/service/notion/notion_services.go`, `pkg/service/service.go`
- **Status**: Complete
- **Changes**:
  - Added `RefundRequests *RefundRequestsService` field to NotionServices
  - Initialized RefundRequestsService in service.go

### Task 4: Add processRefundPayouts Handler
- **File**: `pkg/handler/notion/contractor_payouts.go`
- **Status**: Complete
- **Changes**:
  - Implemented `processRefundPayouts(c, l, payoutType)` method
  - Updated switch in `CreateContractorPayouts` to call processRefundPayouts for type=refund
  - Does NOT update refund status after payout creation
  - Added DEBUG logging

### Task 5: Build & Verify
- **Status**: Complete
- Build passes with no errors

---

## Files Modified

| File | Change Type |
|------|-------------|
| `pkg/service/notion/refund_requests.go` | NEW |
| `pkg/service/notion/contractor_payouts.go` | Extended |
| `pkg/service/notion/notion_services.go` | Extended |
| `pkg/service/service.go` | Extended |
| `pkg/handler/notion/contractor_payouts.go` | Extended |

---

## API Response Format

```json
{
  "data": {
    "payouts_created": 2,
    "refunds_processed": 3,
    "refunds_skipped": 1,
    "errors": 0,
    "details": [
      {
        "refund_page_id": "...",
        "refund_id": "RFD-2025-...",
        "contractor_id": "...",
        "contractor_name": "...",
        "amount": 500000,
        "currency": "VND",
        "reason": "Advance Return",
        "payout_page_id": "...",
        "status": "created",
        "error_reason": null
      }
    ],
    "type": "Refund"
  },
  "message": "ok"
}
```

---

## Key Implementation Details

1. **Direction**: Outgoing (company pays contractor)
2. **Status Update**: Refund Request status is NOT updated after payout creation
3. **Idempotency**: Checked via "Refund Request" relation before creating payout
4. **Payout Name**: Built from `{Reason} - {Description}` or just `{Reason}`

---

## Next Steps

1. **Manual Testing**: Test endpoint with real Notion data
2. **Deploy**: Deploy to staging environment
3. **Monitor**: Monitor first production runs
