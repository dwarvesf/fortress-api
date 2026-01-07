# Implementation Tasks: Hourly Rate Service Fee

**Session**: 202601072141-hourly-rate-service-fee
**Feature**: Hourly Rate-Based Service Fee Display in Contractor Invoices
**Status**: Planned

## 1. Service Layer Implementation

- [ ] **Task 1.1: Update PayoutEntry Struct**
    - **File**: `pkg/service/notion/contractor_payouts.go`
    - **Action**: Add `ServiceRateID` string field to `PayoutEntry` struct.
    - **Action**: Update `QueryPendingPayoutsByContractor` to extract "00 Service Rate" relation ID.
    - **Verification**: Run existing tests for contractor payouts.

- [ ] **Task 1.2: Implement Contractor Rate Fetch**
    - **File**: `pkg/service/notion/contractor_rates.go`
    - **Action**: Add `FetchContractorRateByPageID` method.
    - **Logic**: Fetch page by ID, extract properties (Billing Type, Hourly Rate, Currency).
    - **Reference**: `spec-002-service-methods.md`

- [ ] **Task 1.3: Implement Task Order Hours Fetch**
    - **File**: `pkg/service/notion/task_order_log.go`
    - **Action**: Add `FetchTaskOrderHoursByPageID` method.
    - **Logic**: Fetch page by ID, extract "Final Hours Worked" formula (graceful handling for missing/null).
    - **Reference**: `spec-002-service-methods.md`

- [ ] **Task 1.4: Unit Tests for Services**
    - **File**: `pkg/service/notion/contractor_rates_test.go`, `pkg/service/notion/task_order_log_test.go`
    - **Action**: Implement unit tests defined in `test-cases/unit/service-layer.md`.

## 2. Controller Layer Implementation

- [ ] **Task 2.1: Update ContractorInvoiceLineItem Struct**
    - **File**: `pkg/controller/invoice/contractor_invoice.go`
    - **Action**: Add fields: `IsHourlyRate` (bool), `ServiceRateID` (string), `TaskOrderID` (string).
    - **Reference**: `spec-001-data-structures.md`

- [ ] **Task 2.2: Add Helper Structs & Functions**
    - **File**: `pkg/controller/invoice/contractor_invoice.go` (or new `hourly_helpers.go` if preferred, but usually same file in this codebase)
    - **Action**: Add `hourlyRateData` and `hourlyRateAggregation` structs.
    - **Action**: Implement `fetchHourlyRateData` helper function (with error handling/fallback logic).
    - **Action**: Implement `aggregateHourlyServiceFees` helper function.
    - **Action**: Implement `generateServiceFeeTitle` and `concatenateDescriptions`.
    - **Reference**: `spec-003-detection-logic.md`

- [ ] **Task 2.3: Unit Tests for Helpers**
    - **File**: `pkg/controller/invoice/contractor_invoice_test.go`
    - **Action**: Implement unit tests defined in `test-cases/unit/controller-helpers.md`.

## 3. Integration & Logic

- [ ] **Task 3.1: Integrate into GenerateContractorInvoice**
    - **File**: `pkg/controller/invoice/contractor_invoice.go`
    - **Action**: Instantiate `ContractorRatesService` and `TaskOrderLogService` inside `GenerateContractorInvoice`.
    - **Action**: Update the payout loop to call `fetchHourlyRateData` when SourceType is ServiceFee.
    - **Action**: Implement fallback logic (if nil, use default).
    - **Action**: Call `aggregateHourlyServiceFees` after the payout loop.
    - **Reference**: `spec-004-integration.md`

- [ ] **Task 3.2: Integration Tests**
    - **File**: `pkg/controller/invoice/contractor_invoice_test.go`
    - **Action**: Implement integration tests defined in `test-cases/integration/invoice-generation.md`.
    - **Action**: Verify "Graceful Degradation" scenarios (fallback behavior).

## 4. Verification & Cleanup

- [ ] **Task 4.1: Code Review & Lint**
    - **Action**: Run `golangci-lint run`.
    - **Action**: Ensure all DEBUG logs are in place as per ADR-003.

- [ ] **Task 4.2: Build Check**
    - **Action**: Run `go build ./...` to ensure no compilation errors.
    - **Action**: Run all tests `go test ./...`.
