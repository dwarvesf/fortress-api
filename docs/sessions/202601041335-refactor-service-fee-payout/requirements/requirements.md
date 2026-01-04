# Requirements: Refactor Service Fee Payout

## Overview
Refactor `processContractorPayrollPayouts` endpoint to use Task Order Log as the data source instead of Contractor Fees.

## Background
- The payout system previously used Contractor Fees table with Payment Status = "New"
- Now, Service Fee payouts should be created directly from Task Order Log entries
- The "00 Task Order" relation in Contractor Payouts now references Task Order Log directly

## Requirements

### Functional Requirements

1. **Data Source Change**
   - Query Task Order Log where Type = "Order" AND Status = "Approved"
   - Replace Contractor Fees as the source of payout data

2. **Amount Calculation**
   - Get contractor rate from Contractor Rates table by contractor page ID and month
   - If BillingType = "Monthly Fixed": Amount = MonthlyFixed value
   - If BillingType = "Hourly Rate": Amount = HourlyRate × FinalHoursWorked

3. **Payout Creation**
   - Use Task Order Log PageID for "00 Task Order" relation
   - Use calculated amount and currency from Contractor Rates
   - Include ProofOfWorks as Description (optional)

4. **Status Update**
   - After payout creation, update Task Order Log Status from "Approved" to "Completed"

### Non-Functional Requirements

1. **Idempotency**: Check if payout already exists before creating
2. **Error Handling**: Log errors but continue processing other orders
3. **Debug Logging**: Include detailed debug logs for tracing

## Data Sources

| Source | Table | Filter | Fields Used |
|--------|-------|--------|-------------|
| Orders | Task Order Log | Type=Order, Status=Approved | PageID, ContractorPageID, Date, FinalHoursWorked, ProofOfWorks |
| Rates | Contractor Rates | ContractorPageID, Month | BillingType, MonthlyFixed, HourlyRate, Currency |

## Acceptance Criteria

1. Endpoint `cronjobs/create-contractor-payouts?type=contractor_payroll` creates payouts from Task Order Log
2. Amount is correctly calculated based on billing type
3. Task Order Log status is updated to "Pending" after payout creation
4. Existing payout check prevents duplicates
