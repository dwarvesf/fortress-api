# Planning Status: Commission Payout Processing

**Session ID**: `202512312037-commission-payout-processing`
**Date**: 2025-12-31
**Status**: COMPLETE

---

## Summary

Planning complete for adding commission payout processing to the `create-contractor-payouts` cronjob.

## Documents Created

### ADRs
- [x] `001-follow-existing-payout-pattern.md` - Decision to follow existing payout processing pattern

### Specifications
- [x] `commission-payout-processing.md` - Full technical specification

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Cronjob Endpoint                         │
│        POST /cronjobs/create-contractor-payouts?type=       │
└─────────────────────┬───────────────────────────────────────┘
                      │
         ┌────────────┼────────────┬─────────────┐
         ▼            ▼            ▼             ▼
   contractor    commission     refund        bonus
    _payroll     (NEW)        (existing)    (not impl)
         │            │            │
         ▼            ▼            ▼
   Contractor    Invoice       Refund
     Fees        Splits       Requests
         │            │            │
         └────────────┼────────────┘
                      ▼
              Contractor Payouts
```

## Key Decisions
1. Follow existing handler pattern (`processXxxPayouts`)
2. Use Invoice Split relation for idempotency check
3. Extract Person from Invoice Split's Person relation

## Next Steps
Proceed to Phase 3: Task Breakdown
