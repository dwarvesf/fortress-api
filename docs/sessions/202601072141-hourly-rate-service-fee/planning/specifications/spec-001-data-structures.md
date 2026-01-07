# Specification 001: Data Structures

**Feature**: Hourly Rate-Based Service Fee Display
**Date**: 2026-01-07
**Status**: Draft

## Overview

This specification defines all data structures (structs, fields, types) required to support hourly rate-based Service Fee display in contractor invoices. All structures follow Go conventions and existing codebase patterns.

## Modified Existing Structures

### 1. PayoutEntry (pkg/service/notion/contractor_payouts.go)

**Purpose**: Add ServiceRateID field to capture the `00 Service Rate` relation from Contractor Payouts.

**Location**: Line ~23

**Modification**:
```go
// PayoutEntry represents a single payout entry from the Contractor Payouts database
type PayoutEntry struct {
	PageID          string
	Name            string           // Title/Name of the payout
	Description     string           // From Description rich_text field
	PersonPageID    string           // From Person relation
	SourceType      PayoutSourceType // Determined by which relation is set
	Amount          float64
	Currency        string
	Status          string
	TaskOrderID     string // From "00 Task Order" relation (was ContractorFeesID/Billing)
	InvoiceSplitID  string // From "02 Invoice Split" relation
	RefundRequestID string // From "01 Refund" relation
	WorkDetails     string // From "00 Work Details" formula (proof of works for Service Fee)

	// Commission-specific fields (populated from Invoice Split relation)
	CommissionRole    string // From Invoice Split "Role" select (Sales, Account Manager, etc.)
	CommissionProject string // From Invoice Split "Project" rollup (via Deployment)

	// NEW: Service Rate relation for hourly rate billing detection
	// Populated from "00 Service Rate" relation in Contractor Payouts
	// Used to fetch Contractor Rate data (Billing Type, Hourly Rate, Currency)
	ServiceRateID string // From "00 Service Rate" relation
}
```

**Field Details**:
- **Name**: `ServiceRateID`
- **Type**: `string`
- **Source**: Notion relation property `00 Service Rate` in Contractor Payouts database
- **Usage**: Page ID of related Contractor Rate record
- **Empty when**: Service Fee has no associated rate configuration
- **Example**: `"e3d8a2b0-1234-5678-90ab-cdef12345678"`

**Extraction Code**:
```go
// In QueryPendingPayoutsByContractor method (line ~133)
entry := PayoutEntry{
	PageID:          page.ID,
	Name:            s.extractTitle(props, "Name"),
	Description:     s.extractRichText(props, "Description"),
	PersonPageID:    s.extractFirstRelationID(props, "Person"),
	Amount:          s.extractNumber(props, "Amount"),
	Currency:        s.extractSelect(props, "Currency"),
	Status:          s.extractStatus(props, "Status"),
	TaskOrderID:     s.extractFirstRelationID(props, "00 Task Order"),
	InvoiceSplitID:  s.extractFirstRelationID(props, "02 Invoice Split"),
	RefundRequestID: s.extractFirstRelationID(props, "01 Refund"),
	WorkDetails:     s.extractFormulaString(props, "00 Work Details"),

	// NEW: Extract ServiceRateID
	ServiceRateID:   s.extractFirstRelationID(props, "00 Service Rate"),
}
```

**Impact**:
- Backward compatible: Empty string for payouts without service rate
- No breaking changes to existing code
- Field logged at DEBUG level for traceability

### 2. ContractorInvoiceLineItem (pkg/controller/invoice/contractor_invoice.go)

**Purpose**: Add metadata fields to mark and track hourly-rate Service Fees during aggregation.

**Location**: Line ~56

**Modification**:
```go
// ContractorInvoiceLineItem represents a line item in a contractor invoice
type ContractorInvoiceLineItem struct {
	Title       string
	Description string  // Proof of Work
	Hours       float64 // Only for Hourly Rate
	Rate        float64 // Only for Hourly Rate
	Amount      float64 // Only for Hourly Rate
	AmountUSD   float64 // Amount converted to USD (Only for Hourly Rate)
	Type        string  // Payout source type (Contractor Payroll, Commission, Refund, etc.)

	// Commission-specific fields (for grouping)
	CommissionRole    string
	CommissionProject string

	// Original currency fields (added for multi-currency support)
	OriginalAmount   float64 // Amount in original currency (VND or USD)
	OriginalCurrency string  // "VND" or "USD"

	// NEW: Hourly rate metadata for aggregation and debugging
	IsHourlyRate  bool   // Mark as hourly-rate Service Fee (for aggregation)
	ServiceRateID string // Contractor Rate page ID (for logging/debugging)
	TaskOrderID   string // Task Order Log page ID (for logging/debugging)
}
```

**Field Details**:

#### IsHourlyRate
- **Type**: `bool`
- **Purpose**: Flag to identify hourly-rate Service Fee items for aggregation
- **Set when**: BillingType = "Hourly Rate" and hourly data successfully fetched
- **Default**: `false` (all non-hourly items, fallback cases)
- **Usage**: `aggregateHourlyServiceFees()` scans for items with `IsHourlyRate=true`

#### ServiceRateID
- **Type**: `string`
- **Purpose**: Store Contractor Rate page ID for logging and debugging
- **Set when**: ServiceRateID present in PayoutEntry
- **Default**: `""` (empty string)
- **Usage**: Logged in DEBUG messages to trace rate lookups

#### TaskOrderID
- **Type**: `string`
- **Purpose**: Store Task Order Log page ID for logging and debugging
- **Set when**: TaskOrderID present in PayoutEntry
- **Default**: `""` (empty string)
- **Usage**: Logged in DEBUG messages to trace hours lookups

**Field Interpretation**:
```go
// Example interpretations for different scenarios:

// Hourly-rate Service Fee (successfully processed)
ContractorInvoiceLineItem{
    Title:            "",  // Set during aggregation
    Description:      "Work on Project X",
    Hours:            10.5,
    Rate:             50.0,
    Amount:           525.0,
    Type:             "Contractor Payroll",
    OriginalCurrency: "USD",
    IsHourlyRate:     true,   // Mark for aggregation
    ServiceRateID:    "rate-page-id",
    TaskOrderID:      "task-page-id",
}

// Service Fee with fallback (missing rate data)
ContractorInvoiceLineItem{
    Title:            "",
    Description:      "Work on Project Y",
    Hours:            1,      // Default quantity
    Rate:             500.0,  // Total amount
    Amount:           500.0,
    Type:             "Contractor Payroll",
    OriginalCurrency: "USD",
    IsHourlyRate:     false,  // Not marked (fallback used)
    ServiceRateID:    "rate-page-id",  // Still logged for debugging
    TaskOrderID:      "task-page-id",
}

// Commission (non-Service Fee)
ContractorInvoiceLineItem{
    Title:            "",
    Description:      "Bonus for Project Z",
    Hours:            1,
    Rate:             100.0,
    Amount:           100.0,
    Type:             "Commission",
    OriginalCurrency: "USD",
    IsHourlyRate:     false,  // Not a Service Fee
    ServiceRateID:    "",     // N/A for Commission
    TaskOrderID:      "",
}
```

**Impact**:
- Backward compatible: All new fields default to zero/false/empty
- No changes to existing line item creation for non-Service Fees
- Fields only populated during hourly rate processing

## New Helper Structures

### 3. hourlyRateData (pkg/controller/invoice/contractor_invoice.go)

**Purpose**: Temporary struct to hold fetched hourly rate and hours data.

**Location**: After ContractorInvoiceLineItem definition (line ~82)

**Definition**:
```go
// hourlyRateData holds fetched data for hourly rate Service Fee display.
// Returned by fetchHourlyRateData() when all conditions met:
// - ServiceRateID present
// - Contractor Rate fetch successful
// - BillingType = "Hourly Rate"
// - Task Order hours fetch attempted (may be 0 if failed)
type hourlyRateData struct {
	HourlyRate    float64 // From Contractor Rates "Hourly Rate" field
	Hours         float64 // From Task Order Log "Final Hours Worked" formula
	Currency      string  // From Contractor Rates "Currency" select (USD or VND)
	BillingType   string  // From Contractor Rates "Billing Type" select (for validation)
	ServiceRateID string  // Contractor Rate page ID (for logging)
	TaskOrderID   string  // Task Order Log page ID (for logging)
}
```

**Field Details**:

#### HourlyRate
- **Type**: `float64`
- **Source**: Contractor Rates database, "Hourly Rate" number property
- **Valid range**: > 0 (typically $25-$200 USD or equivalent VND)
- **Example**: `50.0` (USD), `1200000.0` (VND)
- **Used as**: Unit Cost in invoice line item

#### Hours
- **Type**: `float64`
- **Source**: Task Order Log database, "Final Hours Worked" formula property
- **Valid range**: >= 0 (0 if fetch failed, typically 1-200 per month)
- **Example**: `10.5`, `160.0`, `0.0` (if fetch failed)
- **Used as**: Quantity in invoice line item

#### Currency
- **Type**: `string`
- **Source**: Contractor Rates database, "Currency" select property
- **Valid values**: `"USD"`, `"VND"`
- **Used to**: Ensure consistency with payout currency
- **Validation**: Logged warning if mismatches payout currency

#### BillingType
- **Type**: `string`
- **Source**: Contractor Rates database, "Billing Type" select property
- **Valid values**: `"Hourly Rate"`, `"Monthly Fixed"`, etc.
- **Must be**: `"Hourly Rate"` for hourly rate display
- **Used for**: Validation that rate is hourly (not monthly fixed)

#### ServiceRateID, TaskOrderID
- **Type**: `string`
- **Purpose**: Pass through for logging in calling function
- **Used in**: DEBUG log messages to trace data sources

**Usage Example**:
```go
hourlyData := fetchHourlyRateData(ctx, payout, ratesService, taskOrderService, l)
if hourlyData != nil {
    // Successfully fetched hourly rate data
    l.Debug(fmt.Sprintf("[SUCCESS] hours=%.2f rate=%.2f currency=%s",
        hourlyData.Hours, hourlyData.HourlyRate, hourlyData.Currency))

    // Create line item with hourly display
    lineItem := ContractorInvoiceLineItem{
        Hours:  hourlyData.Hours,
        Rate:   hourlyData.HourlyRate,
        Amount: payout.Amount,  // Use original payout amount
        // ...
    }
} else {
    // Fetch failed or not hourly rate, use default display
    lineItem := createDefaultLineItem(payout, amountUSD, description)
}
```

### 4. hourlyRateAggregation (pkg/controller/invoice/contractor_invoice.go)

**Purpose**: Accumulator struct for aggregating multiple hourly Service Fee items.

**Location**: After hourlyRateData definition

**Definition**:
```go
// hourlyRateAggregation holds aggregated data for multiple hourly Service Fees.
// Used internally by aggregateHourlyServiceFees() to accumulate values
// from multiple hourly-rate line items into a single consolidated item.
type hourlyRateAggregation struct {
	TotalHours   float64  // Sum of all Hours fields from hourly items
	HourlyRate   float64  // Hourly rate (same for all items, use first)
	TotalAmount  float64  // Sum of all Amount fields from hourly items
	Currency     string   // Currency (same for all items, use first)
	Descriptions []string // All descriptions (concatenated with line breaks)
	TaskOrderIDs []string // All task order IDs (for logging)
}
```

**Field Details**:

#### TotalHours
- **Type**: `float64`
- **Calculation**: `sum(hourlyItems[i].Hours)` for all hourly items
- **Purpose**: Quantity for aggregated line item
- **Example**: If 3 items with 10, 5, 8 hours → `23.0`

#### HourlyRate
- **Type**: `float64`
- **Source**: First hourly item's Rate field
- **Assumption**: All hourly Service Fees use same hourly rate (per requirements)
- **Validation**: Log warning if multiple rates found
- **Purpose**: Unit Cost for aggregated line item

#### TotalAmount
- **Type**: `float64`
- **Calculation**: `sum(hourlyItems[i].Amount)` for all hourly items
- **Purpose**: Total Amount for aggregated line item
- **Note**: Uses original payout amounts (not recalculated from hours * rate)

#### Currency
- **Type**: `string`
- **Source**: First hourly item's OriginalCurrency field
- **Assumption**: All hourly Service Fees use same currency (per requirements)
- **Validation**: Log warning if multiple currencies found
- **Values**: `"USD"` or `"VND"`

#### Descriptions
- **Type**: `[]string`
- **Source**: All Description fields from hourly items
- **Purpose**: Concatenated into single description with line breaks
- **Processing**:
  - Filter out empty strings
  - Trim whitespace
  - Join with `"\n\n"` (double line break)

#### TaskOrderIDs
- **Type**: `[]string`
- **Source**: All TaskOrderID fields from hourly items
- **Purpose**: Logging to show which task orders were aggregated
- **Example**: `["task-1", "task-2", "task-3"]`

**Usage Example**:
```go
// Internal use in aggregateHourlyServiceFees function
agg := &hourlyRateAggregation{}

for _, item := range hourlyItems {
    agg.TotalHours += item.Hours
    agg.TotalAmount += item.Amount

    if agg.HourlyRate == 0 {
        agg.HourlyRate = item.Rate
        agg.Currency = item.OriginalCurrency
    }

    if item.Description != "" {
        agg.Descriptions = append(agg.Descriptions, item.Description)
    }

    if item.TaskOrderID != "" {
        agg.TaskOrderIDs = append(agg.TaskOrderIDs, item.TaskOrderID)
    }
}

l.Debug(fmt.Sprintf("[AGGREGATE] totalHours=%.2f rate=%.2f totalAmount=%.2f taskOrders=%v",
    agg.TotalHours, agg.HourlyRate, agg.TotalAmount, agg.TaskOrderIDs))
```

## Data Structure Relationships

```
┌─────────────────────────────────────────────────────────────┐
│ Notion: Contractor Payouts                                  │
│ ├─ PageID                                                    │
│ ├─ 00 Service Rate → ServiceRateID                          │
│ ├─ 00 Task Order   → TaskOrderID                            │
│ └─ Amount, Currency                                          │
└───────────────┬─────────────────────────────────────────────┘
                │ extracted into
                v
┌─────────────────────────────────────────────────────────────┐
│ Go: PayoutEntry                                             │
│ ├─ ServiceRateID  (NEW)                                     │
│ ├─ TaskOrderID    (existing)                                │
│ ├─ Amount         (existing)                                │
│ └─ Currency       (existing)                                │
└───────────────┬─────────────────────────────────────────────┘
                │ used to fetch
                v
┌─────────────────────────────────────────────────────────────┐
│ Go: hourlyRateData (helper)                                 │
│ ├─ HourlyRate  (from Contractor Rates)                      │
│ ├─ Hours       (from Task Order Log)                        │
│ ├─ Currency                                                  │
│ └─ BillingType                                               │
└───────────────┬─────────────────────────────────────────────┘
                │ used to create
                v
┌─────────────────────────────────────────────────────────────┐
│ Go: ContractorInvoiceLineItem                               │
│ ├─ Hours          (from hourlyRateData.Hours)               │
│ ├─ Rate           (from hourlyRateData.HourlyRate)          │
│ ├─ Amount         (from PayoutEntry.Amount)                 │
│ ├─ IsHourlyRate   (NEW: true)                               │
│ ├─ ServiceRateID  (NEW: for logging)                        │
│ └─ TaskOrderID    (NEW: for logging)                        │
└───────────────┬─────────────────────────────────────────────┘
                │ multiple items aggregated into
                v
┌─────────────────────────────────────────────────────────────┐
│ Go: hourlyRateAggregation (helper)                          │
│ ├─ TotalHours     (sum of Hours)                            │
│ ├─ HourlyRate     (first item's Rate)                       │
│ ├─ TotalAmount    (sum of Amounts)                          │
│ └─ Descriptions   (all descriptions)                        │
└───────────────┬─────────────────────────────────────────────┘
                │ converted to
                v
┌─────────────────────────────────────────────────────────────┐
│ Go: ContractorInvoiceLineItem (aggregated)                  │
│ ├─ Title: "Service Fee (Development work from ...)"         │
│ ├─ Hours: TotalHours                                         │
│ ├─ Rate: HourlyRate                                          │
│ ├─ Amount: TotalAmount                                       │
│ └─ Description: Concatenated descriptions                    │
└─────────────────────────────────────────────────────────────┘
```

## Validation Rules

### PayoutEntry.ServiceRateID
- **Empty is valid**: Service Fee may not have service rate configured
- **Format**: Notion page ID (UUID format with hyphens)
- **No validation needed**: Used as-is for fetching

### hourlyRateData
- **HourlyRate**: Must be > 0 (logged warning if <= 0)
- **Hours**: Can be 0 (fetch failed) or >= 0 (valid hours)
- **Currency**: Must be "USD" or "VND" (logged warning if other)
- **BillingType**: Must be "Hourly Rate" for hourly display

### ContractorInvoiceLineItem (hourly)
- **Hours**: >= 0 (can be 0 if fetch failed)
- **Rate**: > 0 (from validated HourlyRate)
- **Amount**: > 0 (from payout amount)
- **IsHourlyRate**: true only if all conditions met

### hourlyRateAggregation
- **TotalHours**: >= 0 (sum of non-negative values)
- **HourlyRate**: > 0 (from first item, validated)
- **TotalAmount**: > 0 (sum of positive amounts)
- **Descriptions**: Can be empty array (no descriptions provided)

## Edge Cases

### 1. Missing ServiceRateID
```go
// PayoutEntry
ServiceRateID: ""  // Empty string

// Result: Skip hourly rate processing, use default display
```

### 2. Zero Hours (Fetch Failed)
```go
// hourlyRateData
Hours: 0.0  // Fetch failed, defaulted to 0

// ContractorInvoiceLineItem
IsHourlyRate: true  // Still marked as hourly
Hours: 0.0          // Display "0" hours
Rate: 50.0          // Show hourly rate
Amount: 500.0       // Use actual payout amount

// Result: Displays as "0 hours @ $50/hr = $500" (amount from payout)
```

### 3. Single Hourly Item
```go
// hourlyRateAggregation
TotalHours: 10.0
HourlyRate: 50.0
TotalAmount: 500.0
Descriptions: ["Work on Project X"]

// Result: Still aggregates (for consistent title format)
```

### 4. Multiple Currencies (Shouldn't Happen)
```go
// hourlyItems[0].OriginalCurrency: "USD"
// hourlyItems[1].OriginalCurrency: "VND"

// Result:
// - Log warning: "multiple currencies found in hourly items: [USD VND]"
// - Use first currency: "USD"
// - Continue processing
```

## Testing Data Structures

### Mock Data for Unit Tests

```go
// Test PayoutEntry with ServiceRateID
testPayout := notion.PayoutEntry{
    PageID:        "payout-123",
    SourceType:    notion.PayoutSourceTypeServiceFee,
    Amount:        500.0,
    Currency:      "USD",
    TaskOrderID:   "task-456",
    ServiceRateID: "rate-789",  // NEW field
}

// Test hourlyRateData
testHourlyData := &hourlyRateData{
    HourlyRate:    50.0,
    Hours:         10.0,
    Currency:      "USD",
    BillingType:   "Hourly Rate",
    ServiceRateID: "rate-789",
    TaskOrderID:   "task-456",
}

// Test ContractorInvoiceLineItem with hourly metadata
testLineItem := ContractorInvoiceLineItem{
    Description:      "Work on Project X",
    Hours:            10.0,
    Rate:             50.0,
    Amount:           500.0,
    Type:             "Contractor Payroll",
    OriginalAmount:   500.0,
    OriginalCurrency: "USD",
    IsHourlyRate:     true,     // NEW field
    ServiceRateID:    "rate-789",  // NEW field
    TaskOrderID:      "task-456",  // NEW field
}

// Test hourlyRateAggregation
testAggregation := &hourlyRateAggregation{
    TotalHours:   25.0,
    HourlyRate:   50.0,
    TotalAmount:  1250.0,
    Currency:     "USD",
    Descriptions: []string{"Work A", "Work B", "Work C"},
    TaskOrderIDs: []string{"task-1", "task-2", "task-3"},
}
```

## Success Criteria

1. All structures compile without errors
2. Zero values are safe defaults (no nil pointer panics)
3. Fields have clear, documented purposes
4. Validation rules enforced at appropriate points
5. Edge cases handled gracefully
6. Test data provided for all structures
7. Backward compatibility maintained

## Related Documents

- ADR-001: Data Fetching Strategy
- ADR-002: Aggregation Approach
- ADR-004: Code Organization
- Specification: spec-002-service-methods.md
- Specification: spec-003-detection-logic.md
