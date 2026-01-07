# Integration Test Cases: Invoice Generation Flow

**Feature**: Hourly Rate-Based Service Fee Display
**Scope**: `pkg/controller/invoice/`
**Target Method**: `GenerateContractorInvoice`

## Overview

These tests verify the end-to-end logic within the controller, mocking the external services (Notion) but executing the full `GenerateContractorInvoice` method logic, including the new hourly rate fetching, fallback mechanisms, and aggregation.

## Test Suite: `TestGenerateContractorInvoice_HourlyIntegration`

### Test Case 1: End-to-End Success (Single Hourly Item)
- **Scenario**: A standard contractor with one Service Fee payout configured for hourly billing.
- **Setup**:
    - **Payouts**: 1 Service Fee payout.
        - `ServiceRateID`: "rate-1"
        - `TaskOrderID`: "task-1"
        - `Amount`: 500
    - **Mock Services**:
        - `FetchContractorRateByPageID`: Returns Hourly Rate ($50).
        - `FetchTaskOrderHoursByPageID`: Returns 10 hours.
- **Execution**: Call `GenerateContractorInvoice`.
- **Verify**:
    - Resulting `ContractorInvoiceData` contains 1 line item.
    - Line Item Title: "Service Fee (Development work from ...)"
    - Line Item Quantity: 10
    - Line Item Rate: 50
    - Line Item Amount: 500
    - Logs contain `[SUCCESS] payout ... applying hourly rate display`.
    - Logs contain `[AGGREGATE] created aggregated item`.

### Test Case 2: Fallback - Rate Fetch Failure
- **Scenario**: Notion API fails when fetching rate. System should fallback to default display.
- **Setup**:
    - **Payouts**: 1 Service Fee payout with `ServiceRateID`.
    - **Mock Services**:
        - `FetchContractorRateByPageID`: Returns error.
- **Execution**: Call `GenerateContractorInvoice`.
- **Verify**:
    - Resulting `ContractorInvoiceData` contains 1 line item.
    - Line Item Title: "" (empty, uses default description logic)
    - Line Item Quantity: 1
    - Line Item Rate: 500 (Total Amount)
    - Line Item Amount: 500
    - Logs contain `[FALLBACK] ... failed to fetch rate`.

### Test Case 3: Fallback - Not Hourly Rate
- **Scenario**: Contractor is set to "Monthly Fixed".
- **Setup**:
    - **Payouts**: 1 Service Fee payout.
    - **Mock Services**:
        - `FetchContractorRateByPageID`: Returns "Monthly Fixed".
- **Execution**: Call `GenerateContractorInvoice`.
- **Verify**:
    - Resulting `ContractorInvoiceData` contains 1 line item.
    - Line Item Quantity: 1
    - Logs contain `[INFO] ... billingType=Monthly Fixed`.

### Test Case 4: Graceful Degradation - Hours Fetch Failure
- **Scenario**: Rate is hourly, but Task Order hours cannot be fetched.
- **Setup**:
    - **Payouts**: 1 Service Fee payout.
    - **Mock Services**:
        - `FetchContractorRateByPageID`: Returns Hourly Rate ($50).
        - `FetchTaskOrderHoursByPageID`: Returns error.
- **Execution**: Call `GenerateContractorInvoice`.
- **Verify**:
    - Resulting `ContractorInvoiceData` contains 1 line item.
    - Line Item Quantity: 0 (Graceful degradation per ADR)
    - Line Item Rate: 50
    - Line Item Amount: 500 (Original amount preserved)
    - Logs contain `[FALLBACK] ... failed to fetch hours`.

### Test Case 5: Complex Aggregation (Multiple Items + Commission)
- **Scenario**: Contractor has multiple hourly task orders and a commission.
- **Setup**:
    - **Payouts**:
        1. Service Fee (Task A): $500, Rate $50 (implies 10h).
        2. Service Fee (Task B): $250, Rate $50 (implies 5h).
        3. Commission: $100.
    - **Mock Services**:
        - Return valid data for both Service Fees.
- **Execution**: Call `GenerateContractorInvoice`.
- **Verify**:
    - Resulting `ContractorInvoiceData` contains 2 line items.
    - **Item 1 (Commission)**: $100, Qty 1.
    - **Item 2 (Service Fee)**: 
        - Quantity: 15 (10+5)
        - Rate: 50
        - Amount: 750
        - Description: Contains descriptions from Task A and Task B.
    - Sorting check: Commission appears before Service Fee (based on standard sorting rules).

### Test Case 6: Multi-Currency Handling
- **Scenario**: Payout is in VND.
- **Setup**:
    - **Payouts**: 1 Service Fee payout in VND (Amount 12,000,000).
    - **Mock Services**:
        - `FetchContractorRateByPageID`: Returns Hourly Rate 1,000,000 VND.
        - `FetchTaskOrderHoursByPageID`: Returns 12 hours.
- **Execution**: Call `GenerateContractorInvoice`.
- **Verify**:
    - Line Item preserved original currency VND.
    - Quantity: 12
    - Rate: 1,000,000
    - Amount: 12,000,000
