# Specification 002: Service Methods

**Feature**: Hourly Rate-Based Service Fee Display
**Date**: 2026-01-07
**Status**: Draft

## Overview

This specification defines new service layer methods required to fetch Contractor Rate and Task Order hours data from Notion. All methods follow existing service patterns in the codebase.

## Service Method 1: FetchContractorRateByPageID

### Signature

```go
// File: pkg/service/notion/contractor_rates.go
// Location: After FindActiveRateByContractor method (line ~420)

// FetchContractorRateByPageID fetches a single Contractor Rate by its page ID.
// Used for hourly rate detection in invoice generation.
//
// This method differs from QueryRatesByDiscordAndMonth:
// - Direct page ID lookup (no filtering by discord/month)
// - No date range validation
// - Used when ServiceRateID already known from Contractor Payouts
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - pageID: Contractor Rate page ID from PayoutEntry.ServiceRateID
//
// Returns:
//   - *ContractorRateData: Rate data including BillingType, HourlyRate, Currency
//   - error: Error if fetch fails or page not found
func (s *ContractorRatesService) FetchContractorRateByPageID(
    ctx context.Context,
    pageID string,
) (*ContractorRateData, error)
```

### Implementation

```go
func (s *ContractorRatesService) FetchContractorRateByPageID(ctx context.Context, pageID string) (*ContractorRateData, error) {
    s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] fetching contractor rate: pageID=%s", pageID))

    // Step 1: Fetch the page by ID using Notion client
    page, err := s.client.FindPageByID(ctx, pageID)
    if err != nil {
        s.logger.Error(err, fmt.Sprintf("failed to fetch contractor rate page: %s", pageID))
        return nil, fmt.Errorf("failed to fetch contractor rate page: %w", err)
    }

    // Step 2: Cast page properties to database properties
    props, ok := page.Properties.(nt.DatabasePageProperties)
    if !ok {
        return nil, fmt.Errorf("failed to cast page properties for contractor rate: %s", pageID)
    }

    // Step 3: Extract contractor page ID from relation
    contractorPageID := s.extractFirstRelationID(props, "Contractor")

    // Step 4: Fetch contractor name if contractor page ID available
    contractorName := ""
    if contractorPageID != "" {
        contractorName = s.getContractorName(ctx, contractorPageID)
    }

    // Step 5: Extract all rate data fields
    rateData := &ContractorRateData{
        PageID:           page.ID,
        ContractorPageID: contractorPageID,
        ContractorName:   contractorName,
        Discord:          s.extractRollupRichText(props, "Discord"),
        BillingType:      s.extractSelect(props, "Billing Type"),
        MonthlyFixed:     s.extractFormulaNumber(props, "Monthly Fixed"),
        HourlyRate:       s.extractNumber(props, "Hourly Rate"),
        GrossFixed:       s.extractFormulaNumber(props, "Gross Fixed"),
        Currency:         s.extractSelect(props, "Currency"),
        StartDate:        s.extractDate(props, "Start Date"),
        EndDate:          s.extractDate(props, "End Date"),
    }

    // Step 6: Log extracted data for debugging
    s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] fetched rate: billingType=%s hourlyRate=%.2f currency=%s",
        rateData.BillingType, rateData.HourlyRate, rateData.Currency))

    return rateData, nil
}
```

### Input Parameters

#### pageID
- **Type**: `string`
- **Format**: Notion page ID (UUID with hyphens)
- **Example**: `"e3d8a2b0-1234-5678-90ab-cdef12345678"`
- **Source**: `PayoutEntry.ServiceRateID` from Contractor Payouts
- **Validation**: No validation (passed directly to Notion API)

### Output

#### Success Case

```go
&ContractorRateData{
    PageID:           "e3d8a2b0-1234-5678-90ab-cdef12345678",
    ContractorPageID: "contractor-page-id",
    ContractorName:   "John Doe",
    Discord:          "johndoe#1234",
    BillingType:      "Hourly Rate",
    MonthlyFixed:     0.0,              // Not used for hourly rate
    HourlyRate:       50.0,             // USD per hour
    GrossFixed:       0.0,              // Not used for hourly rate
    Currency:         "USD",
    StartDate:        &time.Time{...},  // 2026-01-01
    EndDate:          nil,              // Active, no end date
}
```

#### Error Cases

**1. Page Not Found**
```go
// Notion API returns 404
// Error: "failed to fetch contractor rate page: object_not_found"
return nil, fmt.Errorf("failed to fetch contractor rate page: %w", err)
```

**2. Invalid Page Type**
```go
// Page exists but not in Contractor Rates database
// Properties type assertion fails
return nil, fmt.Errorf("failed to cast page properties for contractor rate: %s", pageID)
```

**3. Network Error**
```go
// Notion API timeout or connection error
// Error: "failed to fetch contractor rate page: context deadline exceeded"
return nil, fmt.Errorf("failed to fetch contractor rate page: %w", err)
```

**4. Permissions Error**
```go
// Notion API returns 403
// Error: "failed to fetch contractor rate page: unauthorized"
return nil, fmt.Errorf("failed to fetch contractor rate page: %w", err)
```

### Notion API Mapping

| Field | Notion Property | Property Type | Extraction Method |
|-------|----------------|---------------|-------------------|
| PageID | Page ID | UUID | `page.ID` |
| ContractorPageID | Contractor | Relation | `extractFirstRelationID(props, "Contractor")` |
| ContractorName | (from Contractor page) | Title | `getContractorName(ctx, contractorPageID)` |
| Discord | Discord | Rollup (Rich Text) | `extractRollupRichText(props, "Discord")` |
| BillingType | Billing Type | Select | `extractSelect(props, "Billing Type")` |
| MonthlyFixed | Monthly Fixed | Formula (Number) | `extractFormulaNumber(props, "Monthly Fixed")` |
| HourlyRate | Hourly Rate | Number | `extractNumber(props, "Hourly Rate")` |
| GrossFixed | Gross Fixed | Formula (Number) | `extractFormulaNumber(props, "Gross Fixed")` |
| Currency | Currency | Select | `extractSelect(props, "Currency")` |
| StartDate | Start Date | Date | `extractDate(props, "Start Date")` |
| EndDate | End Date | Date | `extractDate(props, "End Date")` |

### Logging

```go
// Entry log
s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] fetching contractor rate: pageID=%s", pageID))

// Success log
s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] fetched rate: billingType=%s hourlyRate=%.2f currency=%s",
    rateData.BillingType, rateData.HourlyRate, rateData.Currency))

// Error log
s.logger.Error(err, fmt.Sprintf("failed to fetch contractor rate page: %s", pageID))
```

### Testing

#### Unit Test Cases

```go
func TestFetchContractorRateByPageID(t *testing.T) {
    tests := []struct {
        name           string
        pageID         string
        mockPage       *nt.Page
        mockError      error
        expectedRate   *ContractorRateData
        expectedError  string
    }{
        {
            name:   "successful fetch - hourly rate USD",
            pageID: "rate-123",
            mockPage: &nt.Page{
                ID: "rate-123",
                Properties: nt.DatabasePageProperties{
                    "Contractor":    {Relation: []nt.Relation{{ID: "contractor-456"}}},
                    "Billing Type":  {Select: &nt.SelectOptions{Name: "Hourly Rate"}},
                    "Hourly Rate":   {Number: toPtr(50.0)},
                    "Currency":      {Select: &nt.SelectOptions{Name: "USD"}},
                },
            },
            expectedRate: &ContractorRateData{
                PageID:      "rate-123",
                BillingType: "Hourly Rate",
                HourlyRate:  50.0,
                Currency:    "USD",
            },
        },
        {
            name:          "page not found",
            pageID:        "invalid-id",
            mockError:     errors.New("object_not_found"),
            expectedError: "failed to fetch contractor rate page",
        },
        {
            name:   "monthly fixed rate (not hourly)",
            pageID: "rate-789",
            mockPage: &nt.Page{
                Properties: nt.DatabasePageProperties{
                    "Billing Type": {Select: &nt.SelectOptions{Name: "Monthly Fixed"}},
                    "Monthly Fixed": {Formula: &nt.FormulaResult{Number: toPtr(5000.0)}},
                },
            },
            expectedRate: &ContractorRateData{
                BillingType:  "Monthly Fixed",
                MonthlyFixed: 5000.0,
                HourlyRate:   0.0,  // Not set for monthly fixed
            },
        },
    }
}
```

### Performance

- **API Calls**: 1 call to Notion (FindPageByID)
- **Additional Calls**: +1 if contractor name fetch needed (getContractorName)
- **Total**: 1-2 Notion API calls per invocation
- **Latency**: ~100-300ms per call (Notion API typical)

---

## Service Method 2: FetchTaskOrderHoursByPageID

### Signature

```go
// File: pkg/service/notion/task_order_log.go
// Location: After GetContractorFromOrder method (line ~2025)

// FetchTaskOrderHoursByPageID fetches the Final Hours Worked from a Task Order Log page.
// Used for hourly rate invoice line item display.
//
// The "Final Hours Worked" field is a Notion formula that calculates total hours
// from sub-items. This method reads the computed value.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - pageID: Task Order Log page ID from PayoutEntry.TaskOrderID
//
// Returns:
//   - float64: Final Hours Worked value (0.0 if field not found or empty)
//   - error: Error if fetch fails or page not found (not returned for missing field)
func (s *TaskOrderLogService) FetchTaskOrderHoursByPageID(
    ctx context.Context,
    pageID string,
) (float64, error)
```

### Implementation

```go
func (s *TaskOrderLogService) FetchTaskOrderHoursByPageID(ctx context.Context, pageID string) (float64, error) {
    s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] fetching task order hours: pageID=%s", pageID))

    // Step 1: Fetch the page by ID using Notion client
    page, err := s.client.FindPageByID(ctx, pageID)
    if err != nil {
        s.logger.Error(err, fmt.Sprintf("failed to fetch task order page: %s", pageID))
        return 0, fmt.Errorf("failed to fetch task order page: %w", err)
    }

    // Step 2: Cast page properties to database properties
    props, ok := page.Properties.(nt.DatabasePageProperties)
    if !ok {
        return 0, fmt.Errorf("failed to cast page properties for task order: %s", pageID)
    }

    // Step 3: Extract Final Hours Worked (formula field)
    // Returns 0.0 if field not found or empty (graceful degradation)
    var hours float64
    if prop, ok := props["Final Hours Worked"]; ok && prop.Formula != nil && prop.Formula.Number != nil {
        hours = *prop.Formula.Number
        s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] fetched hours: %.2f", hours))
    } else {
        // Field not found or empty - log but don't error
        // This allows graceful degradation (display with 0 hours)
        s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] Final Hours Worked not found or empty for pageID=%s, returning 0", pageID))
        return 0, nil
    }

    return hours, nil
}
```

### Input Parameters

#### pageID
- **Type**: `string`
- **Format**: Notion page ID (UUID with hyphens)
- **Example**: `"task-order-123"`
- **Source**: `PayoutEntry.TaskOrderID` from Contractor Payouts
- **Validation**: No validation (passed directly to Notion API)

### Output

#### Success Cases

**1. Hours Found**
```go
// Returns: 10.5, nil
hours, err := service.FetchTaskOrderHoursByPageID(ctx, "task-123")
// hours = 10.5
// err = nil
```

**2. Hours Field Empty (Not an Error)**
```go
// Final Hours Worked formula returns null or field missing
// Returns: 0.0, nil (graceful degradation)
hours, err := service.FetchTaskOrderHoursByPageID(ctx, "task-456")
// hours = 0.0
// err = nil
```

#### Error Cases

**1. Page Not Found**
```go
// Notion API returns 404
// Returns: 0, error
return 0, fmt.Errorf("failed to fetch task order page: object_not_found")
```

**2. Invalid Page Type**
```go
// Page exists but not in Task Order Log database
// Returns: 0, error
return 0, fmt.Errorf("failed to cast page properties for task order: %s", pageID)
```

**3. Network Error**
```go
// Notion API timeout
// Returns: 0, error
return 0, fmt.Errorf("failed to fetch task order page: context deadline exceeded")
```

### Notion API Mapping

| Field | Notion Property | Property Type | Extraction |
|-------|----------------|---------------|------------|
| Hours | Final Hours Worked | Formula (Number) | `props["Final Hours Worked"].Formula.Number` |

**Formula Field Details**:
- **Property Name**: "Final Hours Worked"
- **Type**: Formula (returns Number)
- **Calculation**: Sums "Line Item Hours" from all sub-items (Timesheet entries)
- **Can be**: null (if no sub-items or calculation error)
- **Valid Range**: >= 0 (non-negative hours)

### Logging

```go
// Entry log
s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] fetching task order hours: pageID=%s", pageID))

// Success log
s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] fetched hours: %.2f", hours))

// Missing field log (not error)
s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] Final Hours Worked not found or empty for pageID=%s, returning 0", pageID))

// Error log
s.logger.Error(err, fmt.Sprintf("failed to fetch task order page: %s", pageID))
```

### Testing

#### Unit Test Cases

```go
func TestFetchTaskOrderHoursByPageID(t *testing.T) {
    tests := []struct {
        name          string
        pageID        string
        mockPage      *nt.Page
        mockError     error
        expectedHours float64
        expectedError string
    }{
        {
            name:   "successful fetch - 10.5 hours",
            pageID: "task-123",
            mockPage: &nt.Page{
                ID: "task-123",
                Properties: nt.DatabasePageProperties{
                    "Final Hours Worked": {
                        Formula: &nt.FormulaResult{
                            Number: toPtr(10.5),
                        },
                    },
                },
            },
            expectedHours: 10.5,
        },
        {
            name:   "hours field empty - return 0",
            pageID: "task-456",
            mockPage: &nt.Page{
                Properties: nt.DatabasePageProperties{
                    "Final Hours Worked": {
                        Formula: &nt.FormulaResult{
                            Number: nil,  // Formula returned null
                        },
                    },
                },
            },
            expectedHours: 0.0,  // Graceful degradation
        },
        {
            name:   "hours field missing - return 0",
            pageID: "task-789",
            mockPage: &nt.Page{
                Properties: nt.DatabasePageProperties{
                    // "Final Hours Worked" property missing
                },
            },
            expectedHours: 0.0,  // Graceful degradation
        },
        {
            name:          "page not found",
            pageID:        "invalid-id",
            mockError:     errors.New("object_not_found"),
            expectedError: "failed to fetch task order page",
        },
    }
}
```

### Performance

- **API Calls**: 1 call to Notion (FindPageByID)
- **Latency**: ~100-300ms per call (Notion API typical)

---

## Service Layer Integration

### Existing Helper Methods Reused

Both new methods use existing helper methods from their respective services:

#### ContractorRatesService
- `extractFirstRelationID(props, propName)` - Extract first relation ID
- `extractSelect(props, propName)` - Extract select field
- `extractNumber(props, propName)` - Extract number field
- `extractFormulaNumber(props, propName)` - Extract formula number
- `extractRollupRichText(props, propName)` - Extract rollup rich text
- `extractDate(props, propName)` - Extract date field
- `getContractorName(ctx, pageID)` - Fetch contractor name

#### TaskOrderLogService
- (No existing helpers needed, direct property access)

### Service Instantiation

Services are instantiated in the controller:

```go
// In GenerateContractorInvoice method
ratesService := notion.NewContractorRatesService(c.config, c.logger)
taskOrderService := notion.NewTaskOrderLogService(c.config, c.logger)

// Pass to helper function
hourlyData := fetchHourlyRateData(ctx, payout, ratesService, taskOrderService, l)
```

---

## Error Handling Strategy

### Method-Level Error Handling

Both methods follow this pattern:

1. **Log entry**: DEBUG log at method start
2. **Attempt fetch**: Call Notion API
3. **On error**: Log error, return error (let caller decide fallback)
4. **On success**: Extract data, log success, return data

### Caller Responsibility

The controller (fetchHourlyRateData helper) handles errors:

```go
// Fetch rate
rateData, err := ratesService.FetchContractorRateByPageID(ctx, serviceRateID)
if err != nil {
    l.Error(err, "[FALLBACK] failed to fetch rate, using default display")
    return nil  // Caller uses default display
}

// Fetch hours (graceful degradation for 0 hours)
hours, err := taskOrderService.FetchTaskOrderHoursByPageID(ctx, taskOrderID)
if err != nil {
    l.Error(err, "[FALLBACK] failed to fetch hours, using 0 hours")
    hours = 0  // Continue with 0 hours
}
```

---

## Success Criteria

1. Both methods compile without errors
2. Methods follow existing service patterns
3. All API calls logged at DEBUG level
4. Errors properly wrapped and returned
5. Unit tests cover success and error cases
6. Performance within acceptable range (< 500ms per method)
7. Graceful degradation for missing data (hours = 0)
8. No breaking changes to existing service code

## Related Documents

- ADR-001: Data Fetching Strategy
- ADR-003: Error Handling Strategy
- ADR-004: Code Organization
- Specification: spec-001-data-structures.md
- Specification: spec-003-detection-logic.md
