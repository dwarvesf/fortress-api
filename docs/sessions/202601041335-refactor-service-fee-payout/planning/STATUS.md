# Planning Status: Refactor Service Fee Payout

## Status: IMPLEMENTATION COMPLETE ✅

## Summary
Refactored `processContractorPayrollPayouts` to use Task Order Log as data source instead of Contractor Fees.

## Implementation Summary

### Changes Made
1. **Handler refactored** (`pkg/handler/notion/contractor_payouts.go`)
   - Replaced ContractorFees service with TaskOrderLog + ContractorRates services
   - Query Task Order Log (Type=Order, Status=Approved) instead of Contractor Fees
   - Calculate amount from Contractor Rates based on billing type
   - Update Task Order Log AND subitems status to "Completed" after payout creation
   - Updated PayoutType map: "Contractor Payroll" → "Service Fee"
   - Added `contractor` query parameter to filter by discord, name, or page ID
   - Added `pay_day` query parameter to filter by contractor's pay day (1-31)
   - Description set to empty string for Service Fee payout type

2. **ContractorPayouts service updated** (`pkg/service/notion/contractor_payouts.go`)
   - Added `ServiceRateID` field to `CreatePayoutInput` struct
   - CreatePayout now fills "00 Service Rate" relation column

3. **ContractorRates service updated** (`pkg/service/notion/contractor_rates.go`)
   - Added `PayDay` field to `ContractorRateData` struct
   - `FindActiveRateByContractor()` now extracts Pay Day from "Pay Day" property

4. **TaskOrderLog service updated** (`pkg/service/notion/task_order_log.go`)
   - Added `UpdateOrderAndSubitemsStatus()` method
   - Updated `QueryOrderSubitems()` to extract project via Subitem → Deployment → Project chain

### Build Status
- ✅ `go build ./...` - Passed
- ✅ `golangci-lint run ./...` - 0 issues

## Key Decisions
1. Query Task Order Log (Type=Order, Status=Approved) instead of Contractor Fees
2. Fetch rates from Contractor Rates by contractor page ID and order date
3. Calculate amount in Go based on billing type (Monthly Fixed vs Hourly Rate)
4. Update Task Order Log AND subitems status to "Completed" after payout creation
5. Description set to empty for Service Fee payout type (not JSON from subitems)
6. Fill "00 Service Rate" relation with rate.PageID
7. Support filtering by pay_day parameter matching Contractor Rates PayDay

## Files Modified
1. `pkg/handler/notion/contractor_payouts.go` - Refactored processContractorPayrollPayouts
2. `pkg/service/notion/contractor_payouts.go` - Added ServiceRateID to CreatePayoutInput
3. `pkg/service/notion/contractor_rates.go` - Added PayDay to ContractorRateData
4. `pkg/service/notion/task_order_log.go` - Added UpdateOrderAndSubitemsStatus, updated project extraction

## API Usage

```
GET /api/v1/cronjobs/create-contractor-payouts?type=contractor_payroll

Optional parameters:
- contractor: Filter by discord, name, or page ID
- pay_day: Filter by pay day (1-31) matching contractor's rate
- test_email: Send notifications to test email instead

Examples:
GET /api/v1/cronjobs/create-contractor-payouts?type=contractor_payroll&pay_day=15
GET /api/v1/cronjobs/create-contractor-payouts?type=contractor_payroll&contractor=john.doe
```

## Next Step
Manual testing with live Notion data
