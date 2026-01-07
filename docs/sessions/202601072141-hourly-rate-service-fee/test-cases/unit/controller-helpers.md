# Unit Test Cases: Controller Helpers

**Feature**: Hourly Rate-Based Service Fee Display
**Scope**: `pkg/controller/invoice/`
**Target File**: `contractor_invoice.go` (helper functions)

## 1. Helper: `fetchHourlyRateData`

**Test Suite**: `TestHelper_FetchHourlyRateData`

#### Test Case 1.1: Success Flow
- **Description**: Verify complete success path returning valid data.
- **Input**: 
    - `PayoutEntry`: `{ServiceRateID: "rate-1", TaskOrderID: "task-1"}`
- **Mock Setup**:
    - `ratesService.FetchContractorRateByPageID`: Returns `{BillingType: "Hourly Rate", HourlyRate: 50.0, Currency: "USD"}`
    - `taskOrderService.FetchTaskOrderHoursByPageID`: Returns `10.5`
- **Expected Output**:
    - Returns `*hourlyRateData` with:
        - `HourlyRate`: 50.0
        - `Hours`: 10.5
        - `BillingType`: "Hourly Rate"

#### Test Case 1.2: Missing ServiceRateID
- **Description**: Verify immediate return if ServiceRateID is empty.
- **Input**: 
    - `PayoutEntry`: `{ServiceRateID: ""}`
- **Expected Output**: `nil`

#### Test Case 1.3: Rate Fetch Failure
- **Description**: Verify nil return on rate fetch error.
- **Input**: `PayoutEntry`: `{ServiceRateID: "rate-fail"}`
- **Mock Setup**: `ratesService` returns error.
- **Expected Output**: `nil` (triggers fallback in caller)

#### Test Case 1.4: Non-Hourly Billing Type
- **Description**: Verify nil return if billing type is not "Hourly Rate".
- **Input**: `PayoutEntry`: `{ServiceRateID: "rate-monthly"}`
- **Mock Setup**: `ratesService` returns `{BillingType: "Monthly Fixed"}`.
- **Expected Output**: `nil`

#### Test Case 1.5: Hours Fetch Failure (Graceful Degradation)
- **Description**: Verify hours defaults to 0 if hours fetch fails, but struct is returned.
- **Input**: `PayoutEntry`: `{ServiceRateID: "rate-1", TaskOrderID: "task-fail"}`
- **Mock Setup**: 
    - `ratesService` returns valid hourly rate.
    - `taskOrderService` returns error.
- **Expected Output**:
    - Returns `*hourlyRateData` with:
        - `Hours`: 0.0
        - `HourlyRate`: 50.0
    - **Note**: This is distinct from rate fetch failure; we still want to show "Hourly Rate" format but with 0 hours, rather than fallback to fixed amount display? 
    - **Correction per ADR-003**: 
        - The ADR says: "Use 0 hours, continue with hourly rate display".
        - So `fetchHourlyRateData` should return the struct, not nil.

#### Test Case 1.6: Missing TaskOrderID (Graceful Degradation)
- **Description**: Verify hours defaults to 0 if TaskOrderID is empty.
- **Input**: `PayoutEntry`: `{ServiceRateID: "rate-1", TaskOrderID: ""}`
- **Mock Setup**: `ratesService` valid. `taskOrderService` not called.
- **Expected Output**:
    - Returns `*hourlyRateData` with `Hours`: 0.0.

## 2. Helper: `aggregateHourlyServiceFees`

**Test Suite**: `TestHelper_AggregateHourlyServiceFees`

#### Test Case 2.1: Single Hourly Item
- **Description**: Verify a single hourly item is converted to the aggregated format (title change).
- **Input**:
    - `lineItems`: `[{IsHourlyRate: true, Hours: 10, Rate: 50, Amount: 500, Description: "Work"}]`
    - `month`: "2026-01"
- **Expected Output**:
    - List with 1 item:
        - `Title`: "Service Fee (Development work from 2026-01-01 to 2026-01-31)"
        - `Hours`: 10
        - `Amount`: 500
        - `IsHourlyRate`: `false` (processed)

#### Test Case 2.2: Multiple Hourly Items
- **Description**: Verify multiple hourly items are summed correctly.
- **Input**:
    - `lineItems`: 
        - `[0]`: `{IsHourlyRate: true, Hours: 10, Rate: 50, Amount: 500, Description: "Desc A"}`
        - `[1]`: `{IsHourlyRate: true, Hours: 5, Rate: 50, Amount: 250, Description: "Desc B"}`
- **Expected Output**:
    - List with 1 item:
        - `Hours`: 15 (10+5)
        - `Amount`: 750 (500+250)
        - `Description`: "Desc A\n\nDesc B"
        - `Rate`: 50

#### Test Case 2.3: Mixed Items
- **Description**: Verify only hourly items are aggregated, others preserved.
- **Input**:
    - `lineItems`:
        - `[0]`: `{IsHourlyRate: true, Hours: 10, ...}`
        - `[1]`: `{IsHourlyRate: false, Type: "Commission", Amount: 100}`
- **Expected Output**:
    - List with 2 items:
        - `[0]`: Commission item (unchanged)
        - `[1]`: Aggregated Service Fee item

#### Test Case 2.4: No Hourly Items
- **Description**: Verify list is returned unchanged if no hourly items.
- **Input**: `[{IsHourlyRate: false}, {IsHourlyRate: false}]`
- **Expected Output**: Same as input.

#### Test Case 2.5: Conflicting Rates (Edge Case)
- **Description**: Verify behavior when rates differ (should use first).
- **Input**:
    - `[0]`: `{IsHourlyRate: true, Rate: 50, Amount: 500}`
    - `[1]`: `{IsHourlyRate: true, Rate: 60, Amount: 600}`
- **Expected Output**:
    - Aggregated Item:
        - `Rate`: 50 (from first)
        - `Amount`: 1100 (sum)
        - Log warning generated.

## 3. Helper: `generateServiceFeeTitle`

**Test Suite**: `TestHelper_GenerateServiceFeeTitle`

#### Test Case 3.1: Valid Month
- **Input**: "2026-01"
- **Expected**: "Service Fee (Development work from 2026-01-01 to 2026-01-31)"

#### Test Case 3.2: Leap Year
- **Input**: "2024-02"
- **Expected**: "Service Fee (Development work from 2024-02-01 to 2024-02-29)"

#### Test Case 3.3: Invalid Format
- **Input**: "invalid-date"
- **Expected**: "Service Fee" (fallback)

## 4. Helper: `concatenateDescriptions`

**Test Suite**: `TestHelper_ConcatenateDescriptions`

#### Test Case 4.1: Normal Strings
- **Input**: `["A", "B"]`
- **Expected**: "A\n\nB"

#### Test Case 4.2: Empty/Whitespace Strings
- **Input**: `["A", "", "  ", "B"]`
- **Expected**: "A\n\nB"

#### Test Case 4.3: All Empty
- **Input**: `["", " "]`
- **Expected**: ""
