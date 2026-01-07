# ADR-004: Code Organization and Placement

**Status:** Proposed
**Date:** 2026-01-07
**Context:** Hourly Rate-Based Service Fee Display Implementation

## Context

The hourly rate feature requires changes across multiple layers:
- Service layer: New methods for fetching rate and hours data
- Controller layer: Detection, aggregation, and orchestration logic
- Data structures: New fields and helper structs

The challenge is organizing code to maintain:
- Clear separation of concerns
- Testability
- Maintainability
- Consistency with existing codebase patterns
- Minimal disruption to existing code

## Decision

**We will follow the existing layered architecture with minimal new abstractions.**

### Code Organization Principles

1. **Service Layer = Data Access**: Pure data fetching, no business logic
2. **Controller Layer = Business Logic**: Orchestration, decisions, aggregation
3. **Private Helpers in Controller**: Implementation details not exposed
4. **Existing Patterns**: Follow established conventions in codebase

### File Modification Map

```
pkg/service/notion/
├── contractor_payouts.go      [MODIFY]  Add ServiceRateID field extraction
├── contractor_rates.go        [MODIFY]  Add FetchContractorRateByPageID method
└── task_order_log.go          [MODIFY]  Add FetchTaskOrderHoursByPageID method

pkg/controller/invoice/
└── contractor_invoice.go      [MODIFY]  Add helper functions + integrate
```

### Architectural Layers

```
┌─────────────────────────────────────────────────────────────┐
│ HTTP Handler Layer                                          │
│ (pkg/handler/invoice/)                                      │
│ - Routes requests to controller                             │
│ - No changes needed                                         │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      v
┌─────────────────────────────────────────────────────────────┐
│ Controller Layer                                            │
│ (pkg/controller/invoice/contractor_invoice.go)             │
│                                                             │
│ PUBLIC:                                                     │
│ - GenerateContractorInvoice()      [MODIFY]                │
│                                                             │
│ PRIVATE (NEW):                                              │
│ - fetchHourlyRateData()                                     │
│ - aggregateHourlyServiceFees()                              │
│ - generateServiceFeeTitle()                                 │
│ - concatenateDescriptions()                                 │
│                                                             │
│ PRIVATE (NEW STRUCTS):                                      │
│ - hourlyRateData                                            │
│ - hourlyRateAggregation                                     │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      v
┌─────────────────────────────────────────────────────────────┐
│ Service Layer                                               │
│ (pkg/service/notion/)                                       │
│                                                             │
│ contractor_payouts.go:                                      │
│ - PayoutEntry struct         [ADD FIELD]                    │
│   + ServiceRateID string                                    │
│ - QueryPendingPayoutsByContractor()  [MODIFY]              │
│   + Extract ServiceRateID                                   │
│                                                             │
│ contractor_rates.go:                                        │
│ - FetchContractorRateByPageID()     [NEW METHOD]           │
│                                                             │
│ task_order_log.go:                                          │
│ - FetchTaskOrderHoursByPageID()     [NEW METHOD]           │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      v
┌─────────────────────────────────────────────────────────────┐
│ External API Layer                                          │
│ (Notion API via go-notion client)                           │
│ - No changes needed                                         │
└─────────────────────────────────────────────────────────────┘
```

## Rationale

### Why Modify Existing Files?

1. **Service Layer**: Adding methods to existing services (ContractorRatesService, TaskOrderLogService)
   - Follows single responsibility principle
   - Reuses existing Notion client and configuration
   - Consistent with other methods in same service
   - No need for new service just for one method

2. **Controller Layer**: Adding helpers to contractor_invoice.go
   - Keeps related logic together
   - Helpers are implementation details, not public API
   - Easy to find and maintain
   - Follows existing pattern (e.g., validateLineItemCurrencies, formatCurrency)

### Why Not Create New Files?

**Considered**: Create `hourly_rate_service.go` in pkg/service/notion/

**Rejected Because**:
- Only 2 new methods (FetchContractorRateByPageID, FetchTaskOrderHoursByPageID)
- These methods belong naturally to ContractorRatesService and TaskOrderLogService
- New service would need to duplicate Notion client setup
- Would fragment related code across multiple files
- Over-engineering for small feature

**Considered**: Create `hourly_rate_helpers.go` in pkg/controller/invoice/

**Rejected Because**:
- Only ~150 lines of new code in controller
- Helpers are specific to hourly rate aggregation, not reusable
- Better to keep in contractor_invoice.go near usage
- Would make code review harder (changes split across files)

### Why Private Helpers in Controller?

1. **Not Public API**: These are implementation details
2. **Single Use**: Only used by GenerateContractorInvoice
3. **Testable**: Can still unit test via test files in same package
4. **Encapsulation**: Hides complexity from callers

## Implementation Details

### Service Layer Changes

#### contractor_payouts.go

```go
// ADD FIELD to existing struct (line ~23)
type PayoutEntry struct {
    PageID          string
    Name            string
    Description     string
    PersonPageID    string
    SourceType      PayoutSourceType
    Amount          float64
    Currency        string
    Status          string
    TaskOrderID     string
    InvoiceSplitID  string
    RefundRequestID string
    WorkDetails     string
    CommissionRole    string
    CommissionProject string

    // NEW: Service Rate relation for hourly rate billing detection
    ServiceRateID   string // From "00 Service Rate" relation
}

// MODIFY existing method (line ~133)
func (s *ContractorPayoutsService) QueryPendingPayoutsByContractor(...) {
    // ... existing code ...

    entry := PayoutEntry{
        // ... existing field extractions ...

        // NEW: Extract ServiceRateID
        ServiceRateID:    s.extractFirstRelationID(props, "00 Service Rate"),
    }

    // ... rest of existing code ...
}
```

**Lines changed**: ~10 lines
**Risk**: Low (additive change, no breaking changes)

#### contractor_rates.go

```go
// ADD NEW METHOD after existing methods (after FindActiveRateByContractor)
// Line ~420

// FetchContractorRateByPageID fetches a single Contractor Rate by its page ID.
// Used for hourly rate detection in invoice generation.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - pageID: Contractor Rate page ID from PayoutEntry.ServiceRateID
//
// Returns:
//   - *ContractorRateData: Rate data including BillingType, HourlyRate, Currency
//   - error: Error if fetch fails or page not found
func (s *ContractorRatesService) FetchContractorRateByPageID(ctx context.Context, pageID string) (*ContractorRateData, error) {
    s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] fetching contractor rate: pageID=%s", pageID))

    // Fetch the page by ID
    page, err := s.client.FindPageByID(ctx, pageID)
    if err != nil {
        s.logger.Error(err, fmt.Sprintf("failed to fetch contractor rate page: %s", pageID))
        return nil, fmt.Errorf("failed to fetch contractor rate page: %w", err)
    }

    props, ok := page.Properties.(nt.DatabasePageProperties)
    if !ok {
        return nil, fmt.Errorf("failed to cast page properties for contractor rate: %s", pageID)
    }

    // Extract contractor page ID
    contractorPageID := s.extractFirstRelationID(props, "Contractor")

    // Fetch contractor name if we have the page ID
    contractorName := ""
    if contractorPageID != "" {
        contractorName = s.getContractorName(ctx, contractorPageID)
    }

    // Extract rate data
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

    s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] fetched rate: billingType=%s hourlyRate=%.2f currency=%s",
        rateData.BillingType, rateData.HourlyRate, rateData.Currency))

    return rateData, nil
}
```

**Lines added**: ~40 lines
**Risk**: Low (new method, no impact on existing code)

#### task_order_log.go

```go
// ADD NEW METHOD after existing methods
// Line ~2025+

// FetchTaskOrderHoursByPageID fetches the Final Hours Worked from a Task Order Log page.
// Used for hourly rate invoice line item display.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - pageID: Task Order Log page ID from PayoutEntry.TaskOrderID
//
// Returns:
//   - float64: Final Hours Worked value
//   - error: Error if fetch fails or page not found
func (s *TaskOrderLogService) FetchTaskOrderHoursByPageID(ctx context.Context, pageID string) (float64, error) {
    s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] fetching task order hours: pageID=%s", pageID))

    // Fetch the page by ID
    page, err := s.client.FindPageByID(ctx, pageID)
    if err != nil {
        s.logger.Error(err, fmt.Sprintf("failed to fetch task order page: %s", pageID))
        return 0, fmt.Errorf("failed to fetch task order page: %w", err)
    }

    props, ok := page.Properties.(nt.DatabasePageProperties)
    if !ok {
        return 0, fmt.Errorf("failed to cast page properties for task order: %s", pageID)
    }

    // Extract Final Hours Worked (formula field)
    var hours float64
    if prop, ok := props["Final Hours Worked"]; ok && prop.Formula != nil && prop.Formula.Number != nil {
        hours = *prop.Formula.Number
        s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] fetched hours: %.2f", hours))
    } else {
        s.logger.Debug(fmt.Sprintf("[HOURLY_RATE] Final Hours Worked not found or empty for pageID=%s", pageID))
        // Return 0 instead of error - allows graceful degradation
        return 0, nil
    }

    return hours, nil
}
```

**Lines added**: ~35 lines
**Risk**: Low (new method, no impact on existing code)

### Controller Layer Changes

#### contractor_invoice.go - New Structs

```go
// ADD NEW STRUCTS near top of file after ContractorInvoiceSection (line ~82)

// hourlyRateData holds fetched data for hourly rate Service Fee display
type hourlyRateData struct {
    HourlyRate    float64
    Hours         float64
    Currency      string
    BillingType   string
    ServiceRateID string
    TaskOrderID   string
}

// hourlyRateAggregation holds aggregated data for multiple hourly Service Fees
type hourlyRateAggregation struct {
    TotalHours    float64
    HourlyRate    float64
    TotalAmount   float64
    Currency      string
    Descriptions  []string
    TaskOrderIDs  []string
}
```

**Lines added**: ~20 lines

#### contractor_invoice.go - Modified Struct

```go
// MODIFY ContractorInvoiceLineItem (line ~56)
type ContractorInvoiceLineItem struct {
    Title       string
    Description string
    Hours       float64
    Rate        float64
    Amount      float64
    AmountUSD   float64
    Type        string

    CommissionRole    string
    CommissionProject string

    OriginalAmount   float64
    OriginalCurrency string

    // NEW: Hourly rate metadata for aggregation
    IsHourlyRate  bool   // Mark as hourly-rate Service Fee
    ServiceRateID string // For logging/debugging
    TaskOrderID   string // For logging/debugging
}
```

**Lines modified**: 3 new fields added

#### contractor_invoice.go - New Helper Functions

```go
// ADD NEW FUNCTIONS after GenerateContractorInvoice (end of file)

// fetchHourlyRateData attempts to fetch hourly rate and hours data for a Service Fee payout.
// Returns nil if any step fails (caller should use default display).
func fetchHourlyRateData(
    ctx context.Context,
    payout notion.PayoutEntry,
    ratesService *notion.ContractorRatesService,
    taskOrderService *notion.TaskOrderLogService,
    l logger.Logger,
) *hourlyRateData {
    // Implementation per ADR-003 error handling strategy
}

// aggregateHourlyServiceFees consolidates all hourly-rate Service Fee items
// into a single line item with summed hours and amounts.
func aggregateHourlyServiceFees(
    lineItems []ContractorInvoiceLineItem,
    month string,
    l logger.Logger,
) []ContractorInvoiceLineItem {
    // Implementation per ADR-002 aggregation approach
}

// generateServiceFeeTitle creates title for aggregated Service Fee line item.
func generateServiceFeeTitle(month string) string {
    // Implementation per spec-003
}

// concatenateDescriptions joins descriptions with double line breaks.
func concatenateDescriptions(descriptions []string) string {
    // Implementation per spec-003
}

// createDefaultLineItem creates line item with default display (Qty=1).
func createDefaultLineItem(
    payout notion.PayoutEntry,
    amountUSD float64,
    description string,
) ContractorInvoiceLineItem {
    // Implementation per ADR-003 fallback strategy
}
```

**Lines added**: ~150 lines for all helper functions

#### contractor_invoice.go - Integration Point

```go
// MODIFY GenerateContractorInvoice at line ~286

// 4. Build line items from payouts
var lineItems []ContractorInvoiceLineItem
var total float64

// Create service references for helpers
ratesService := notion.NewContractorRatesService(c.config, c.logger)
taskOrderService := notion.NewTaskOrderLogService(c.config, c.logger)

for i, payout := range payouts {
    amountUSD := amountsUSD[i]

    // ... existing description logic (lines 208-253) ...

    // NEW: For Service Fees, attempt hourly rate display
    var lineItem ContractorInvoiceLineItem
    if payout.SourceType == notion.PayoutSourceTypeServiceFee {
        hourlyData := fetchHourlyRateData(ctx, payout, ratesService, taskOrderService, l)
        if hourlyData != nil {
            // Use hourly rate display
            l.Debug(fmt.Sprintf("[SUCCESS] payout %s: applying hourly rate display", payout.PageID))
            lineItem = ContractorInvoiceLineItem{
                Title:            "",
                Description:      description,
                Hours:            hourlyData.Hours,
                Rate:             hourlyData.HourlyRate,
                Amount:           payout.Amount,
                AmountUSD:        amountUSD,
                Type:             string(payout.SourceType),
                OriginalAmount:   payout.Amount,
                OriginalCurrency: payout.Currency,
                IsHourlyRate:     true,
                ServiceRateID:    payout.ServiceRateID,
                TaskOrderID:      payout.TaskOrderID,
            }
        } else {
            // Fallback to default display
            lineItem = createDefaultLineItem(payout, amountUSD, description)
        }
    } else {
        // Non-Service Fee: use existing default logic
        lineItem = createDefaultLineItem(payout, amountUSD, description)
    }

    lineItems = append(lineItems, lineItem)
    total += amountUSD
}

l.Debug(fmt.Sprintf("built %d line items with total=%.2f USD", len(lineItems), total))

// NEW: 4.5 Aggregate hourly-rate Service Fees
l.Debug("aggregating hourly-rate Service Fee items")
lineItems = aggregateHourlyServiceFees(lineItems, month, l)

// EXISTING: 4.6 Group Commission items by Project (line ~288+)
// ... rest of existing code unchanged ...
```

**Lines modified/added**: ~40 lines at integration point

## Alternatives Considered

### Alternative 1: Separate Service Files

**Approach**: Create new service files
- `pkg/service/notion/hourly_rate_service.go`
- `pkg/controller/invoice/hourly_rate_helpers.go`

**Rejected Because**:
- Fragments related code across multiple files
- Increases complexity for small feature
- Makes code review harder
- Adds more files to maintain
- Methods belong naturally to existing services

### Alternative 2: Middleware Pattern

**Approach**: Create hourly rate processing middleware that wraps invoice generation.

**Rejected Because**:
- Over-engineering for feature size
- Adds abstraction layer with no clear benefit
- Doesn't fit existing codebase patterns
- Harder to understand control flow

### Alternative 3: Strategy Pattern

**Approach**: Create LineItemBuilder interface with implementations for default and hourly rate.

**Rejected Because**:
- Over-engineering for 2 display types
- Adds complexity without clear benefit
- Doesn't match existing codebase style
- Makes simple logic harder to follow

### Alternative 4: Embedded Services

**Approach**: Add hourly rate methods directly to invoice controller struct.

**Rejected Because**:
- Violates separation of concerns
- Controller shouldn't do data access
- Can't reuse methods for other features
- Harder to unit test

## Consequences

### Positive

1. **Consistency**: Follows existing codebase patterns
2. **Simplicity**: Minimal new abstractions
3. **Maintainability**: Related code stays together
4. **Testability**: Each layer testable independently
5. **Reusability**: Service methods reusable for future features
6. **Discoverability**: Easy to find related code
7. **Code Review**: Changes localized, easier to review

### Negative

1. **File Size**: contractor_invoice.go grows by ~150 lines
   - Mitigation: Still reasonable size (~700 lines total)
2. **Service Methods**: Add methods to existing services
   - Mitigation: Follows single responsibility, methods belong there

### Neutral

1. **Private Helpers**: Only accessible within package
   - Acceptable: Not needed elsewhere

## Testing Strategy

### Service Layer Tests

Each new service method gets its own test file section:

```go
// pkg/service/notion/contractor_rates_test.go
func TestFetchContractorRateByPageID(t *testing.T) { ... }

// pkg/service/notion/task_order_log_test.go
func TestFetchTaskOrderHoursByPageID(t *testing.T) { ... }
```

### Controller Tests

Test helper functions and integration:

```go
// pkg/controller/invoice/contractor_invoice_test.go
func TestFetchHourlyRateData(t *testing.T) { ... }
func TestAggregateHourlyServiceFees(t *testing.T) { ... }
func TestGenerateServiceFeeTitle(t *testing.T) { ... }
func TestConcatenateDescriptions(t *testing.T) { ... }
func TestGenerateContractorInvoice_HourlyRate(t *testing.T) { ... }
```

## File Impact Summary

```
SERVICE LAYER (pkg/service/notion/):
  contractor_payouts.go:    +10 lines  (add field + extraction)
  contractor_rates.go:      +40 lines  (new method)
  task_order_log.go:        +35 lines  (new method)

CONTROLLER LAYER (pkg/controller/invoice/):
  contractor_invoice.go:    +200 lines total
    - New structs:           +20 lines
    - Modified struct:       +3 fields
    - Helper functions:      +150 lines
    - Integration point:     +40 lines (modified existing code)

TOTAL:                      +285 lines across 4 files
```

## Code Review Checklist

- [ ] Service methods follow existing service patterns
- [ ] Helper functions are properly private (lowercase names)
- [ ] All new code has DEBUG logging
- [ ] Error handling follows ADR-003 strategy
- [ ] Unit tests cover all new functions
- [ ] Integration tests cover end-to-end flow
- [ ] No breaking changes to existing code
- [ ] Documentation comments follow Go conventions
- [ ] No new exported functions (all internal to package)

## Success Criteria

1. All new code follows existing codebase patterns
2. No new exported types or functions (except service methods)
3. Clear separation: services = data access, controller = logic
4. Helper functions are private implementation details
5. Code review takes < 2 hours (localized changes)
6. Merge doesn't require changes to other packages
7. New code integrates seamlessly with existing flow

## Related Documents

- Requirements: Code Quality (NFR-4)
- ADR-001: Data Fetching Strategy
- ADR-002: Aggregation Approach
- ADR-003: Error Handling Strategy
- Codebase patterns: `CLAUDE.md`
