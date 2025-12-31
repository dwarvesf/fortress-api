# Planning Status: Cronjob Refund Payouts

**Session**: `202512311436-cronjob-refund-payouts`
**Date**: 2025-12-31
**Status**: PLANNING COMPLETE

## Summary

Add refund payout type support to `POST /cronjobs/contractor-payouts?type=refund`

## Key Decisions

1. **Direction**: Outgoing (company pays contractor)
2. **Status Update**: Do NOT update Refund Request status after payout creation
3. **Idempotency**: Check via Refund Request relation before creating payout

## Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| `pkg/service/notion/refund_requests.go` | NEW | RefundRequestsService |
| `pkg/service/notion/contractor_payouts.go` | EXTEND | Add refund payout methods |
| `pkg/service/notion/notion_services.go` | EXTEND | Add RefundRequests field |
| `pkg/service/service.go` | EXTEND | Initialize RefundRequestsService |
| `pkg/handler/notion/contractor_payouts.go` | EXTEND | Add processRefundPayouts |

## Config

- `RefundRequest` database ID already configured in `pkg/config/config.go`
- Env var: `NOTION_REFUND_REQUEST_DB_ID`

## Notion Property Mappings

### Refund Request Properties
- `Refund ID` (title)
- `Amount` (number)
- `Currency` (select)
- `Status` (status) - filter by "Approved"
- `Contractor` (relation)
- `Reason` (select)
- `Date Approved` (date)

### Payout Properties to Set
- `Name`: Title from refund reason
- `Amount`: From refund
- `Currency`: From refund
- `Type`: "Refund"
- `Direction`: "Outgoing"
- `Status`: "Pending"
- `Person`: Contractor from refund
- `Refund Request`: Relation to source refund
