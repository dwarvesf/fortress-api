# Test Data Requirements

**Version:** 1.0
**Date:** 2026-01-07
**Component:** Contractor Invoice PDF Generation
**Purpose:** Define test data structures and fixtures needed for all test cases

## Overview

This document specifies all test data requirements for unit and integration testing of the contractor invoice multi-currency feature.

## Mock Data Structures

### 1. PayoutEntry Mock Data

```go
type MockPayoutEntry struct {
    Amount            float64
    Currency          string
    SourceType        string
    CommissionRole    string
    CommissionProject string
    ProofOfWork       string
    // ... other fields as needed
}
```

### Pre-defined Test Payouts

#### VND Payouts
```go
var TestPayoutVNDService = MockPayoutEntry{
    Amount:       45000000,
    Currency:     "VND",
    SourceType:   "Service Fee",
    ProofOfWork:  "Backend development for Project Alpha",
}

var TestPayoutVNDRefund = MockPayoutEntry{
    Amount:       500000,
    Currency:     "VND",
    SourceType:   "Refund",
    ProofOfWork:  "Equipment reimbursement",
}

var TestPayoutVNDSmall = MockPayoutEntry{
    Amount:       100000,
    Currency:     "VND",
    SourceType:   "Service Fee",
    ProofOfWork:  "Code review",
}

var TestPayoutVNDLarge = MockPayoutEntry{
    Amount:       1000000000,
    Currency:     "VND",
    SourceType:   "Service Fee",
    ProofOfWork:  "Major project completion",
}
```

#### USD Payouts
```go
var TestPayoutUSDService = MockPayoutEntry{
    Amount:       1500,
    Currency:     "USD",
    SourceType:   "Service Fee",
    ProofOfWork:  "Consulting services",
}

var TestPayoutUSDBonus = MockPayoutEntry{
    Amount:       100,
    Currency:     "USD",
    SourceType:   "Commission",
    ProofOfWork:  "Performance bonus",
}

var TestPayoutUSDSmall = MockPayoutEntry{
    Amount:       0.50,
    Currency:     "USD",
    SourceType:   "Service Fee",
    ProofOfWork:  "Minor adjustment",
}

var TestPayoutUSDLarge = MockPayoutEntry{
    Amount:       100000,
    Currency:     "USD",
    SourceType:   "Service Fee",
    ProofOfWork:  "Annual contract",
}
```

#### Edge Case Payouts
```go
var TestPayoutZeroVND = MockPayoutEntry{
    Amount:       0,
    Currency:     "VND",
    SourceType:   "Service Fee",
}

var TestPayoutZeroUSD = MockPayoutEntry{
    Amount:       0,
    Currency:     "USD",
    SourceType:   "Service Fee",
}

var TestPayoutInvalidCurrency = MockPayoutEntry{
    Amount:       1000,
    Currency:     "EUR",
    SourceType:   "Service Fee",
}

var TestPayoutNegativeAmount = MockPayoutEntry{
    Amount:       -500,
    Currency:     "USD",
    SourceType:   "Service Fee",
}
```

### 2. Exchange Rate Mock Data

```go
type MockExchangeRate struct {
    Rate  float64
    Error error
}
```

### Pre-defined Exchange Rates
```go
var TestExchangeRateTypical = MockExchangeRate{
    Rate:  26269.5,
    Error: nil,
}

var TestExchangeRateLow = MockExchangeRate{
    Rate:  25000.0,
    Error: nil,
}

var TestExchangeRateHigh = MockExchangeRate{
    Rate:  27000.0,
    Error: nil,
}

var TestExchangeRateRound = MockExchangeRate{
    Rate:  26000.0,
    Error: nil,
}

var TestExchangeRateAPIError = MockExchangeRate{
    Rate:  0,
    Error: errors.New("Wise API timeout"),
}

var TestExchangeRateZero = MockExchangeRate{
    Rate:  0,
    Error: nil,
}

var TestExchangeRateNegative = MockExchangeRate{
    Rate:  -26269,
    Error: nil,
}
```

### 3. Invoice Scenario Fixtures

#### Scenario: All VND Invoice
```go
var ScenarioAllVND = InvoiceTestScenario{
    Name: "All VND items",
    Payouts: []MockPayoutEntry{
        TestPayoutVNDService,
        TestPayoutVNDRefund,
    },
    ExchangeRate: TestExchangeRateTypical,
    Expected: ExpectedInvoiceData{
        SubtotalVND:        45500000,
        SubtotalUSDFromVND: 1731.89,
        SubtotalUSDItems:   0,
        SubtotalUSD:        1731.89,
        FXSupport:          8.0,
        TotalUSD:           1739.89,
        ExchangeRate:       26269.5,
    },
}
```

#### Scenario: All USD Invoice
```go
var ScenarioAllUSD = InvoiceTestScenario{
    Name: "All USD items",
    Payouts: []MockPayoutEntry{
        TestPayoutUSDService,
        TestPayoutUSDBonus,
    },
    ExchangeRate: MockExchangeRate{Rate: 1.0},
    Expected: ExpectedInvoiceData{
        SubtotalVND:        0,
        SubtotalUSDFromVND: 0,
        SubtotalUSDItems:   1600,
        SubtotalUSD:        1600,
        FXSupport:          8.0,
        TotalUSD:           1608,
        ExchangeRate:       1.0,
    },
}
```

#### Scenario: Mixed Currencies
```go
var ScenarioMixed = InvoiceTestScenario{
    Name: "Mixed VND and USD",
    Payouts: []MockPayoutEntry{
        TestPayoutVNDService,
        TestPayoutVNDRefund,
        TestPayoutUSDBonus,
    },
    ExchangeRate: TestExchangeRateTypical,
    Expected: ExpectedInvoiceData{
        SubtotalVND:        45500000,
        SubtotalUSDFromVND: 1731.89,
        SubtotalUSDItems:   100,
        SubtotalUSD:        1831.89,
        FXSupport:          8.0,
        TotalUSD:           1839.89,
        ExchangeRate:       26269.5,
    },
}
```

#### Scenario: Empty Invoice
```go
var ScenarioEmpty = InvoiceTestScenario{
    Name:         "No line items",
    Payouts:      []MockPayoutEntry{},
    ExchangeRate: MockExchangeRate{Rate: 1.0},
    Expected: ExpectedInvoiceData{
        SubtotalVND:        0,
        SubtotalUSDFromVND: 0,
        SubtotalUSDItems:   0,
        SubtotalUSD:        0,
        FXSupport:          8.0,
        TotalUSD:           8.0,
        ExchangeRate:       1.0,
    },
}
```

#### Scenario: Single Item VND
```go
var ScenarioSingleVND = InvoiceTestScenario{
    Name: "Single VND item",
    Payouts: []MockPayoutEntry{
        TestPayoutVNDService,
    },
    ExchangeRate: TestExchangeRateTypical,
    Expected: ExpectedInvoiceData{
        SubtotalVND:        45000000,
        SubtotalUSDFromVND: 1713.41,
        SubtotalUSDItems:   0,
        SubtotalUSD:        1713.41,
        FXSupport:          8.0,
        TotalUSD:           1721.41,
        ExchangeRate:       26269.5,
    },
}
```

#### Scenario: Single Item USD
```go
var ScenarioSingleUSD = InvoiceTestScenario{
    Name: "Single USD item",
    Payouts: []MockPayoutEntry{
        TestPayoutUSDService,
    },
    ExchangeRate: MockExchangeRate{Rate: 1.0},
    Expected: ExpectedInvoiceData{
        SubtotalVND:        0,
        SubtotalUSDFromVND: 0,
        SubtotalUSDItems:   1500,
        SubtotalUSD:        1500,
        FXSupport:          8.0,
        TotalUSD:           1508,
        ExchangeRate:       1.0,
    },
}
```

#### Scenario: Large Amounts
```go
var ScenarioLargeAmounts = InvoiceTestScenario{
    Name: "Large VND and USD amounts",
    Payouts: []MockPayoutEntry{
        TestPayoutVNDLarge,
        TestPayoutUSDLarge,
    },
    ExchangeRate: TestExchangeRateTypical,
    Expected: ExpectedInvoiceData{
        SubtotalVND:        1000000000,
        SubtotalUSDFromVND: 38063.61,
        SubtotalUSDItems:   100000,
        SubtotalUSD:        138063.61,
        FXSupport:          8.0,
        TotalUSD:           138071.61,
        ExchangeRate:       26269.5,
    },
}
```

#### Scenario: Small Amounts
```go
var ScenarioSmallAmounts = InvoiceTestScenario{
    Name: "Small VND and USD amounts",
    Payouts: []MockPayoutEntry{
        TestPayoutVNDSmall,
        TestPayoutUSDSmall,
    },
    ExchangeRate: TestExchangeRateTypical,
    Expected: ExpectedInvoiceData{
        SubtotalVND:        100000,
        SubtotalUSDFromVND: 3.81,
        SubtotalUSDItems:   0.50,
        SubtotalUSD:        4.31,
        FXSupport:          8.0,
        TotalUSD:           12.31,
        ExchangeRate:       26269.5,
    },
}
```

## Formatting Test Data

### VND Formatting Test Cases
```go
var VNDFormattingTests = []FormattingTestCase{
    {Input: 0, Expected: "0 ₫"},
    {Input: 100, Expected: "100 ₫"},
    {Input: 500000, Expected: "500.000 ₫"},
    {Input: 45000000, Expected: "45.000.000 ₫"},
    {Input: 1234567, Expected: "1.234.567 ₫"},
    {Input: 1000000000, Expected: "1.000.000.000 ₫"},
    {Input: 1234.56, Expected: "1.235 ₫"},  // Rounding
    {Input: -1000, Expected: "-1.000 ₫"},   // Negative
}
```

### USD Formatting Test Cases
```go
var USDFormattingTests = []FormattingTestCase{
    {Input: 0, Expected: "$0.00"},
    {Input: 0.99, Expected: "$0.99"},
    {Input: 100, Expected: "$100.00"},
    {Input: 1234.56, Expected: "$1,234.56"},
    {Input: 1000000, Expected: "$1,000,000.00"},
    {Input: 1234.567, Expected: "$1,234.57"},  // Rounding
    {Input: -100.50, Expected: "$-100.50"},    // Negative
}
```

### Exchange Rate Formatting Test Cases
```go
var ExchangeRateFormattingTests = []FormattingTestCase{
    {Input: 26269.5, Expected: "1 USD = 26.270 VND"},
    {Input: 25000, Expected: "1 USD = 25.000 VND"},
    {Input: 26269.4, Expected: "1 USD = 26.269 VND"},
    {Input: 1000000, Expected: "1 USD = 1.000.000 VND"},
}
```

## Helper Functions for Test Data

### Test Data Builders
```go
// Build invoice with specific line items
func BuildTestInvoice(payouts []MockPayoutEntry) *ContractorInvoiceData

// Build line item from payout
func BuildLineItem(payout MockPayoutEntry) ContractorInvoiceLineItem

// Build invoice data with custom values
func BuildInvoiceData(opts InvoiceDataOptions) *ContractorInvoiceData

// Generate random valid payout
func RandomValidPayout() MockPayoutEntry

// Generate random edge case payout
func RandomEdgeCasePayout() MockPayoutEntry
```

### Mock Service Builders
```go
// Create mock Wise service with specific behavior
func MockWiseService(rate float64, err error) *MockWiseService

// Create mock Wise service with call counting
func MockWiseServiceWithCounter() *MockWiseServiceCounter

// Create mock Notion service
func MockNotionService(payouts []MockPayoutEntry) *MockNotionService
```

## Test Fixture Files

### JSON Fixtures (Optional)

Create JSON files for complex test scenarios:

```
testdata/
├── invoices/
│   ├── all-vnd-invoice.json
│   ├── all-usd-invoice.json
│   ├── mixed-invoice.json
│   ├── empty-invoice.json
│   └── large-invoice.json
├── payouts/
│   ├── vnd-payouts.json
│   ├── usd-payouts.json
│   └── mixed-payouts.json
└── exchange-rates/
    ├── typical-rates.json
    └── edge-case-rates.json
```

### SQL Fixtures (For Integration Tests)

```sql
-- testdata/payouts.sql
INSERT INTO payouts (id, amount, currency, source_type, contractor_id) VALUES
('payout-1', 45000000, 'VND', 'Service Fee', 'contractor-1'),
('payout-2', 500000, 'VND', 'Refund', 'contractor-1'),
('payout-3', 100, 'USD', 'Commission', 'contractor-1');
```

## Mock Object Interfaces

### Wise API Mock
```go
type MockWiseService struct {
    ConvertFunc func(amount float64, from, to string) (float64, float64, error)
    CallCount   int
}

func (m *MockWiseService) Convert(amount float64, from, to string) (float64, float64, error) {
    m.CallCount++
    if m.ConvertFunc != nil {
        return m.ConvertFunc(amount, from, to)
    }
    // Default behavior
    return amount / 26269.5, 26269.5, nil
}
```

### Notion API Mock
```go
type MockNotionService struct {
    Payouts []MockPayoutEntry
    Error   error
}

func (m *MockNotionService) QueryPayouts(discord, month string) ([]PayoutEntry, error) {
    if m.Error != nil {
        return nil, m.Error
    }
    return m.Payouts, nil
}
```

## Test Assertion Helpers

### Float Comparison
```go
// Assert float equality with tolerance
func AssertFloatEqual(t *testing.T, actual, expected, tolerance float64) {
    if math.Abs(actual-expected) > tolerance {
        t.Errorf("Expected %.4f, got %.4f (tolerance %.4f)", expected, actual, tolerance)
    }
}

// Assert float equality for currency (2 decimal places)
func AssertCurrencyEqual(t *testing.T, actual, expected float64) {
    AssertFloatEqual(t, actual, expected, 0.01)
}
```

### String Comparison
```go
// Assert formatted currency output
func AssertFormattedCurrency(t *testing.T, actual, expected string) {
    if actual != expected {
        t.Errorf("Expected %q, got %q", expected, actual)
    }
}
```

## Data Validation Rules

All test data must comply with:

1. **Currency Codes:** Only "VND" or "USD" (uppercase)
2. **Amounts:** Non-negative float64 values
3. **Exchange Rates:** Positive float64 values (> 0)
4. **Source Types:** Valid payout source types from model
5. **Precision:**
   - VND: 0 decimal places
   - USD: 2 decimal places
   - Exchange Rate: 4 decimal places

## Test Data Management

### Best Practices

1. **Centralize:** Keep all test data in one package/file
2. **Reuse:** Use pre-defined fixtures across test files
3. **Document:** Comment expected outcomes for each fixture
4. **Version:** Update fixtures when business rules change
5. **Isolate:** Each test should have independent data (no shared state)

### Test Data Lifecycle

1. **Setup:** Load fixtures before test
2. **Execute:** Run test with fixture data
3. **Assert:** Verify results match expected values
4. **Teardown:** Clean up any modified state

## Success Criteria

- All test scenarios have corresponding fixtures
- Test data covers happy path and edge cases
- Mock objects behave consistently
- Test data is well-documented
- Easy to add new test scenarios
