# Unit Test Cases: Service Layer

**Feature**: Hourly Rate-Based Service Fee Display
**Scope**: `pkg/service/notion/`
**Target Files**: 
- `contractor_rates.go`
- `task_order_log.go`

## 1. Contractor Rates Service

### Method: `FetchContractorRateByPageID`

**Test Suite**: `TestContractorRatesService_FetchContractorRateByPageID`

#### Test Case 1.1: Successful Fetch (Hourly Rate USD)
- **Description**: Verify successful fetching and parsing of a valid hourly rate page.
- **Input**: `pageID = "valid-rate-id"`
- **Mock Setup**: 
    - `FindPageByID` returns `notion.Page` with:
        - Properties:
            - "Contractor": Relation to "contractor-1"
            - "Billing Type": Select "Hourly Rate"
            - "Hourly Rate": Number 50.0
            - "Currency": Select "USD"
- **Expected Output**:
    - `rateData` struct:
        - `BillingType`: "Hourly Rate"
        - `HourlyRate`: 50.0
        - `Currency`: "USD"
    - `error`: nil

#### Test Case 1.2: Successful Fetch (Monthly Fixed)
- **Description**: Verify fetching of a monthly fixed rate (valid page, different type).
- **Input**: `pageID = "monthly-rate-id"`
- **Mock Setup**: 
    - `FindPageByID` returns `notion.Page` with:
        - Properties:
            - "Billing Type": Select "Monthly Fixed"
            - "Monthly Fixed": Formula 2000.0
- **Expected Output**:
    - `rateData` struct:
        - `BillingType`: "Monthly Fixed"
        - `HourlyRate`: 0.0 (default)
    - `error`: nil

#### Test Case 1.3: Page Not Found
- **Description**: Verify error handling when page does not exist.
- **Input**: `pageID = "missing-id"`
- **Mock Setup**: `FindPageByID` returns error `object_not_found`
- **Expected Output**:
    - `rateData`: nil
    - `error`: "failed to fetch contractor rate page: object_not_found"

#### Test Case 1.4: Invalid Page Properties
- **Description**: Verify error handling when page exists but properties cast fails.
- **Input**: `pageID = "invalid-props"`
- **Mock Setup**: `FindPageByID` returns `notion.Page` with invalid/unexpected properties structure.
- **Expected Output**:
    - `rateData`: nil
    - `error`: "failed to cast page properties"

## 2. Task Order Log Service

### Method: `FetchTaskOrderHoursByPageID`

**Test Suite**: `TestTaskOrderLogService_FetchTaskOrderHoursByPageID`

#### Test Case 2.1: Successful Fetch (Valid Hours)
- **Description**: Verify successful fetching of computed hours.
- **Input**: `pageID = "valid-task-id"`
- **Mock Setup**:
    - `FindPageByID` returns `notion.Page` with:
        - Properties: "Final Hours Worked": Formula -> Number 10.5
- **Expected Output**:
    - `hours`: 10.5
    - `error`: nil

#### Test Case 2.2: Field Missing/Null (Graceful Degradation)
- **Description**: Verify that missing or null formula field returns 0 hours without error.
- **Input**: `pageID = "empty-hours-id"`
- **Mock Setup**:
    - `FindPageByID` returns `notion.Page` with:
        - Properties: "Final Hours Worked": Formula -> Number nil (or property missing)
- **Expected Output**:
    - `hours`: 0.0
    - `error`: nil

#### Test Case 2.3: Page Not Found
- **Description**: Verify error handling when page does not exist.
- **Input**: `pageID = "missing-id"`
- **Mock Setup**: `FindPageByID` returns error `object_not_found`
- **Expected Output**:
    - `hours`: 0.0
    - `error`: "failed to fetch task order page: object_not_found"

#### Test Case 2.4: Network Error
- **Description**: Verify error handling on network failure.
- **Input**: `pageID = "network-error"`
- **Mock Setup**: `FindPageByID` returns error `context deadline exceeded`
- **Expected Output**:
    - `hours`: 0.0
    - `error`: "failed to fetch task order page: context deadline exceeded"
