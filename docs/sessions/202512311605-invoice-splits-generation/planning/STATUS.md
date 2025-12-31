# Planning Status: Invoice Splits Generation

**Session ID**: `202512311605-invoice-splits-generation`
**Date**: 2025-12-31
**Status**: COMPLETE

---

## Summary

Planning for automatic invoice splits generation when client invoice is marked as paid.

## ADRs Created

| ADR | Title | Status |
|-----|-------|--------|
| ADR-001 | Use Background Worker Queue for Invoice Splits Generation | Accepted |

## Specifications Created

| Spec | Description |
|------|-------------|
| invoice-splits-generation.md | Full specification for invoice splits generation flow |

## Key Decisions

1. **Background Processing**: Use existing Worker queue for async processing
2. **Trigger Point**: Enqueue job in `processNotionInvoicePaid()` after status update
3. **Idempotency**: Check `Splits Generated` flag before processing
4. **Error Handling**: Log errors, don't fail main flow

## Components Affected

- `pkg/model/worker_message.go` - New message type
- `pkg/service/notion/invoice.go` - Query line items, mark splits generated
- `pkg/service/notion/invoice_split.go` - Create split records
- `pkg/worker/worker.go` - Handle splits generation job
- `pkg/controller/invoice/mark_paid.go` - Enqueue job

## Next Steps

Proceed to task breakdown (Phase 3).
