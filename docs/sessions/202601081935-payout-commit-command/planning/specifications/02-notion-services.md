# Specification: Notion Services for Payout Commit

## Overview

This specification defines the Notion service methods required to support payout commit functionality. All methods follow existing patterns in the codebase for querying and updating Notion databases.

## Services to Modify

```
fortress-api/pkg/service/notion/
├── contractor_payables.go      [MODIFY - Add 3 new methods]
├── contractor_payouts.go        [MODIFY - Add 2 new methods]
├── invoice_split.go             [MODIFY - Add 1 new method]
└── refund_requests.go           [MODIFY - Add 1 new method]
```

## 1. ContractorPayablesService

### File: `pkg/service/notion/contractor_payables.go`

### Method 1: QueryPendingPayablesByPeriod

**Purpose**: Query all contractor payables with Payment Status="Pending" for a given period.

**Signature**:
```go
func (s *ContractorPayablesService) QueryPendingPayablesByPeriod(ctx context.Context, period string) ([]PendingPayable, error)
```

**Parameters**:
- `ctx context.Context`: Request context
- `period string`: Date in YYYY-MM-DD format (e.g., "2025-01-01")

**Returns**:
- `[]PendingPayable`: List of pending payables
- `error`: Error if query fails

**Data Structure**:
```go
// PendingPayable represents a payable record with Pending status
type PendingPayable struct {
    PageID             string   // Payable page ID
    ContractorPageID   string   // From Contractor relation
    ContractorName     string   // Rollup from Contractor
    Total              float64  // Total amount
    Currency           string   // USD or VND
    Period             string   // YYYY-MM-DD
    PayoutItemPageIDs  []string // From Payout Items relation (multiple)
}
```

**Implementation Notes**:
```go
// Query filter
filter := &nt.DatabaseQueryFilter{
    And: []nt.DatabaseQueryFilter{
        {
            Property: "Payment Status",
            DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
                Status: &nt.StatusDatabaseQueryFilter{
                    Equals: "Pending",
                },
            },
        },
        {
            Property: "Period",
            DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
                Date: &nt.DatePropertyFilter{
                    Equals: parseDate(period),
                },
            },
        },
    },
}

// Handle pagination
query := &nt.DatabaseQuery{
    Filter:   filter,
    PageSize: 100,
}

// Loop through pages with cursor
for {
    resp, err := s.client.QueryDatabase(ctx, payablesDBID, query)
    // Extract data, handle pagination
    if !resp.HasMore {
        break
    }
    query.StartCursor = *resp.NextCursor
}
```

**Property Extraction**:
- `Payment Status`: Status property
- `Period`: Date property
- `Contractor`: Relation property (extract first ID)
- `Total`: Number property
- `Currency`: Select property
- `Payout Items`: Relation property (extract all IDs)

**Error Handling**:
- Return error if database ID not configured
- Return error if Notion API call fails
- Log all operations at DEBUG level
- Return empty slice if no results (not an error)

**Logging**:
```go
s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: querying pending payables period=%s", period))
s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: found %d pending payables", len(payables)))
```

---

### Method 2: UpdatePayableStatus

**Purpose**: Update a payable's Payment Status and Payment Date.

**Signature**:
```go
func (s *ContractorPayablesService) UpdatePayableStatus(ctx context.Context, pageID string, status string, paymentDate string) error
```

**Parameters**:
- `ctx context.Context`: Request context
- `pageID string`: Payable page ID to update
- `status string`: New status value (e.g., "Paid")
- `paymentDate string`: Payment date in YYYY-MM-DD format

**Returns**:
- `error`: Error if update fails

**Implementation Notes**:
```go
// Build update parameters
params := nt.UpdatePageParams{
    DatabasePageProperties: nt.DatabasePageProperties{
        "Payment Status": nt.DatabasePageProperty{
            Status: &nt.SelectOptions{
                Name: status,
            },
        },
        "Payment Date": nt.DatabasePageProperty{
            Date: &nt.Date{
                Start: nt.NewDateTime(parseDate(paymentDate), false),
            },
        },
    },
}

_, err := s.client.UpdatePage(ctx, pageID, params)
```

**Error Handling**:
- Return error if page ID is empty
- Return error if Notion API call fails
- Log operation at DEBUG level

**Idempotency**:
- Operation is idempotent: re-running with same values is safe
- Notion API will accept update even if values haven't changed

**Logging**:
```go
s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: updating payable pageID=%s status=%s paymentDate=%s", pageID, status, paymentDate))
s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: updated payable pageID=%s successfully", pageID))
```

---

### Method 3: GetContractorPayDay

**Purpose**: Get the PayDay value from a contractor's Service Rate relation.

**Signature**:
```go
func (s *ContractorPayablesService) GetContractorPayDay(ctx context.Context, contractorPageID string) (int, error)
```

**Parameters**:
- `ctx context.Context`: Request context
- `contractorPageID string`: Contractor page ID

**Returns**:
- `int`: PayDay value (1 or 15)
- `error`: Error if query fails or no Service Rate found

**Implementation Strategy**:

**Option A: Query Service Rate Database**
```go
// Query Service Rate database for records with Contractor relation
serviceRateDBID := s.cfg.Notion.Databases.ServiceRate

filter := &nt.DatabaseQueryFilter{
    Property: "Contractor",
    DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
        Relation: &nt.RelationDatabaseQueryFilter{
            Contains: contractorPageID,
        },
    },
}

query := &nt.DatabaseQuery{
    Filter:   filter,
    PageSize: 1,
}

resp, err := s.client.QueryDatabase(ctx, serviceRateDBID, query)
if err != nil || len(resp.Results) == 0 {
    return 0, fmt.Errorf("no service rate found for contractor %s", contractorPageID)
}

// Extract PayDay from first result
props := resp.Results[0].Properties.(nt.DatabasePageProperties)
payDayProp := props["PayDay"]
if payDayProp.Select == nil {
    return 0, errors.New("PayDay property not found")
}

// Parse PayDay value ("1" or "15" as string)
payDay, err := strconv.Atoi(payDayProp.Select.Name)
if err != nil {
    return 0, fmt.Errorf("invalid PayDay value: %s", payDayProp.Select.Name)
}

return payDay, nil
```

**Option B: Fetch Contractor Page and Follow Service Rate Relation**
```go
// Fetch contractor page
contractorPage, err := s.client.FindPageByID(ctx, contractorPageID)
if err != nil {
    return 0, fmt.Errorf("failed to fetch contractor: %w", err)
}

// Extract Service Rate relation
props := contractorPage.Properties.(nt.DatabasePageProperties)
serviceRateProp := props["Service Rate"]
if len(serviceRateProp.Relation) == 0 {
    return 0, errors.New("no service rate relation found")
}

serviceRatePageID := serviceRateProp.Relation[0].ID

// Fetch Service Rate page
serviceRatePage, err := s.client.FindPageByID(ctx, serviceRatePageID)
if err != nil {
    return 0, fmt.Errorf("failed to fetch service rate: %w", err)
}

// Extract PayDay
rateProps := serviceRatePage.Properties.(nt.DatabasePageProperties)
payDayProp := rateProps["PayDay"]
if payDayProp.Select == nil {
    return 0, errors.New("PayDay property not found")
}

payDay, err := strconv.Atoi(payDayProp.Select.Name)
if err != nil {
    return 0, fmt.Errorf("invalid PayDay value: %s", payDayProp.Select.Name)
}

return payDay, nil
```

**Recommendation**: Use Option A (Query Service Rate Database) - fewer API calls, more efficient.

**Error Handling**:
- Return error if contractor ID is empty
- Return error if no Service Rate found
- Return error if PayDay property is missing or invalid
- Log at DEBUG level

**Caching Consideration**:
- PayDay values rarely change
- Consider adding in-memory cache with TTL (optional optimization)

**Logging**:
```go
s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: fetching PayDay for contractor=%s", contractorPageID))
s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payables: contractor=%s PayDay=%d", contractorPageID, payDay))
```

---

## 2. ContractorPayoutsService

### File: `pkg/service/notion/contractor_payouts.go`

### Method 1: GetPayoutWithRelations

**Purpose**: Fetch a payout record with its related Invoice Split and Refund Request IDs.

**Signature**:
```go
func (s *ContractorPayoutsService) GetPayoutWithRelations(ctx context.Context, payoutPageID string) (*PayoutWithRelations, error)
```

**Parameters**:
- `ctx context.Context`: Request context
- `payoutPageID string`: Payout page ID

**Returns**:
- `*PayoutWithRelations`: Payout data with relation IDs
- `error`: Error if fetch fails

**Data Structure**:
```go
// PayoutWithRelations contains payout data with related record IDs
type PayoutWithRelations struct {
    PageID          string
    Status          string
    InvoiceSplitID  string // From "02 Invoice Split" relation (may be empty)
    RefundRequestID string // From "01 Refund" relation (may be empty)
}
```

**Implementation Notes**:
```go
page, err := s.client.FindPageByID(ctx, payoutPageID)
if err != nil {
    return nil, fmt.Errorf("failed to fetch payout: %w", err)
}

props, ok := page.Properties.(nt.DatabasePageProperties)
if !ok {
    return nil, errors.New("failed to cast payout page properties")
}

result := &PayoutWithRelations{
    PageID:          payoutPageID,
    Status:          s.extractStatus(props, "Status"),
    InvoiceSplitID:  s.extractFirstRelationID(props, "02 Invoice Split"),
    RefundRequestID: s.extractFirstRelationID(props, "01 Refund"),
}

return result, nil
```

**Property Extraction**:
- `Status`: Status property
- `02 Invoice Split`: Relation property (extract first ID, may be empty)
- `01 Refund`: Relation property (extract first ID, may be empty)

**Error Handling**:
- Return error if page ID is empty
- Return error if Notion API call fails
- Log at DEBUG level

**Logging**:
```go
s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: fetching payout with relations pageID=%s", payoutPageID))
s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: payout pageID=%s invoiceSplit=%s refund=%s",
    payoutPageID, result.InvoiceSplitID, result.RefundRequestID))
```

---

### Method 2: UpdatePayoutStatus

**Purpose**: Update a payout's Status to a new value.

**Signature**:
```go
func (s *ContractorPayoutsService) UpdatePayoutStatus(ctx context.Context, pageID string, status string) error
```

**Parameters**:
- `ctx context.Context`: Request context
- `pageID string`: Payout page ID to update
- `status string`: New status value (e.g., "Paid")

**Returns**:
- `error`: Error if update fails

**Implementation Notes**:
```go
params := nt.UpdatePageParams{
    DatabasePageProperties: nt.DatabasePageProperties{
        "Status": nt.DatabasePageProperty{
            Status: &nt.SelectOptions{
                Name: status,
            },
        },
    },
}

_, err := s.client.UpdatePage(ctx, pageID, params)
if err != nil {
    return fmt.Errorf("failed to update payout status: %w", err)
}

return nil
```

**Property Type**: Status (not Select) - use `.Status` field

**Error Handling**:
- Return error if page ID is empty
- Return error if Notion API call fails
- Log at DEBUG level

**Idempotency**: Operation is idempotent

**Logging**:
```go
s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: updating payout pageID=%s status=%s", pageID, status))
s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: updated payout pageID=%s successfully", pageID))
```

---

## 3. InvoiceSplitService

### File: `pkg/service/notion/invoice_split.go`

### Method: UpdateInvoiceSplitStatus

**Purpose**: Update an invoice split's Status to a new value.

**Signature**:
```go
func (s *InvoiceSplitService) UpdateInvoiceSplitStatus(ctx context.Context, pageID string, status string) error
```

**Parameters**:
- `ctx context.Context`: Request context
- `pageID string`: Invoice Split page ID to update
- `status string`: New status value (e.g., "Paid")

**Returns**:
- `error`: Error if update fails

**Implementation Notes**:
```go
params := nt.UpdatePageParams{
    DatabasePageProperties: nt.DatabasePageProperties{
        "Status": nt.DatabasePageProperty{
            Select: &nt.SelectOptions{
                Name: status,
            },
        },
    },
}

_, err := s.client.UpdatePage(ctx, pageID, params)
if err != nil {
    return fmt.Errorf("failed to update invoice split status: %w", err)
}

return nil
```

**IMPORTANT**: Invoice Split uses **Select** property type (not Status)

**Property Type Difference**:
```go
// Invoice Split (Select type)
Select: &nt.SelectOptions{Name: "Paid"}

// vs. Other tables (Status type)
Status: &nt.SelectOptions{Name: "Paid"}
```

**Error Handling**:
- Return error if page ID is empty
- Return error if Notion API call fails
- Log at DEBUG level

**Idempotency**: Operation is idempotent

**Logging**:
```go
s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: updating status pageID=%s status=%s", pageID, status))
s.logger.Debug(fmt.Sprintf("[DEBUG] invoice_split: updated pageID=%s successfully", pageID))
```

---

## 4. RefundRequestsService

### File: `pkg/service/notion/refund_requests.go`

### Method: UpdateRefundRequestStatus

**Purpose**: Update a refund request's Status to a new value.

**Signature**:
```go
func (s *RefundRequestsService) UpdateRefundRequestStatus(ctx context.Context, pageID string, status string) error
```

**Parameters**:
- `ctx context.Context`: Request context
- `pageID string`: Refund Request page ID to update
- `status string`: New status value (e.g., "Paid")

**Returns**:
- `error`: Error if update fails

**Implementation Notes**:
```go
params := nt.UpdatePageParams{
    DatabasePageProperties: nt.DatabasePageProperties{
        "Status": nt.DatabasePageProperty{
            Status: &nt.SelectOptions{
                Name: status,
            },
        },
    },
}

_, err := s.client.UpdatePage(ctx, pageID, params)
if err != nil {
    return fmt.Errorf("failed to update refund request status: %w", err)
}

return nil
```

**Property Type**: Status (not Select) - use `.Status` field

**Error Handling**:
- Return error if page ID is empty
- Return error if Notion API call fails
- Log at DEBUG level

**Idempotency**: Operation is idempotent

**Logging**:
```go
s.logger.Debug(fmt.Sprintf("[DEBUG] refund_requests: updating status pageID=%s status=%s", pageID, status))
s.logger.Debug(fmt.Sprintf("[DEBUG] refund_requests: updated pageID=%s successfully", pageID))
```

---

## Summary of Property Types by Database

**CRITICAL**: Different Notion databases use different property types for Status field:

| Database | Property Name | Property Type | Update Code |
|----------|--------------|---------------|-------------|
| Contractor Payables | Payment Status | Status | `.Status: &nt.SelectOptions{Name: "Paid"}` |
| Contractor Payouts | Status | Status | `.Status: &nt.SelectOptions{Name: "Paid"}` |
| Invoice Split | Status | **Select** | `.Select: &nt.SelectOptions{Name: "Paid"}` |
| Refund Request | Status | Status | `.Status: &nt.SelectOptions{Name: "Paid"}` |

**Why This Matters**:
- Using wrong property type will cause Notion API to reject the update
- Invoice Split is the only one using Select type
- All others use Status type

## Database IDs Configuration

All database IDs should be configured in `pkg/config/config.go`:

```go
type Databases struct {
    ContractorPayables string `env:"NOTION_DATABASE_CONTRACTOR_PAYABLES" envDefault:"2c264b29-b84c-8037-807c-000bf6d0792c"`
    ContractorPayouts  string `env:"NOTION_DATABASE_CONTRACTOR_PAYOUTS" envDefault:"2c564b29-b84c-8045-80ee-000bee2e3669"`
    InvoiceSplit       string `env:"NOTION_DATABASE_INVOICE_SPLIT" envDefault:"2c364b29-b84c-804f-9856-000b58702dea"`
    RefundRequest      string `env:"NOTION_DATABASE_REFUND_REQUEST" envDefault:"2cc64b29-b84c-8066-adf2-cc56171cedf4"`
    ServiceRate        string `env:"NOTION_DATABASE_SERVICE_RATE" envDefault:"2c464b29-b84c-80cf-bef6-000b42bce15e"`
}
```

## Testing Considerations

### Unit Tests

1. **Mock Notion Client**: Use gomock to mock `*nt.Client`
2. **Test Data Extraction**: Verify property parsing logic
3. **Test Error Handling**: Simulate API errors
4. **Test Idempotency**: Verify re-running updates is safe

### Integration Tests

1. Create test records in Notion test databases
2. Test full query and update flow
3. Verify status transitions
4. Test with missing relations (empty Invoice Split/Refund)

### Test Cases

**QueryPendingPayablesByPeriod**:
- No pending payables (return empty slice)
- Multiple pending payables
- Payables with different periods (filter correctly)
- Pagination handling (>100 records)

**UpdatePayableStatus**:
- Update Pending → Paid
- Update already Paid (idempotent)
- Invalid page ID (error)
- API error (error)

**GetContractorPayDay**:
- Contractor with PayDay=1
- Contractor with PayDay=15
- Contractor with no Service Rate (error)
- Invalid PayDay value (error)

**GetPayoutWithRelations**:
- Payout with Invoice Split only
- Payout with Refund only
- Payout with both
- Payout with neither

**Update Methods (Status)**:
- Update Pending → Paid
- Update already Paid (idempotent)
- Invalid page ID (error)
- API error (error)

## Performance Considerations

1. **Pagination**: Handle large result sets (>100 records)
2. **Caching**: Consider caching PayDay values (optional)
3. **Parallel Updates**: Current design is sequential; could parallelize payout updates within a payable
4. **Rate Limiting**: Notion API has rate limits; add retry logic with exponential backoff

## Error Handling Best Practices

1. **Wrap Errors**: Use `fmt.Errorf("context: %w", err)` for error chains
2. **Log Before Returning**: Always log errors at DEBUG or ERROR level
3. **Include Context**: Include page IDs in error messages
4. **Don't Panic**: Return errors, don't panic
5. **Check Empty Values**: Validate inputs before making API calls

## Logging Best Practices

1. **Use DEBUG Level**: All Notion operations should log at DEBUG level
2. **Include Page IDs**: Always include page IDs in logs
3. **Log Entry and Exit**: Log at start and end of each operation
4. **Log Query Sizes**: Log count of results from queries
5. **Structured Fields**: Use logger.Fields for structured logging

Example:
```go
l := s.logger.Fields(logger.Fields{
    "service":  "contractor_payables",
    "method":   "UpdatePayableStatus",
    "page_id":  pageID,
    "status":   status,
})
l.Debug("updating payable status")
// ... do work ...
l.Debug("payable status updated successfully")
```

## References

- ADR-001: Cascade Status Update Strategy
- Existing Service: `pkg/service/notion/contractor_payables.go`
- Existing Service: `pkg/service/notion/contractor_payouts.go`
- Notion API Docs: https://developers.notion.com/reference/intro
- Property Types Reference: Spec `/docs/specs/payout-commit-command.md`
