# ADR-002: Email Extraction from Notion Rollup Property

## Status
Proposed

## Context

The Notion Expense Request database has a **relation** to the Contractor database for the `Requestor` field, and a **rollup** property named `Email` that aggregates the email from the related contractor record.

### Database Structure

```
Expense Request Database
├── Requestor (relation → Contractor DB)
└── Email (rollup from Requestor.Email)

Contractor Database
├── Name (title)
└── Email (email property)
```

The payroll calculation requires the employee's email address to look up their `BasecampID` for expense attribution:

```go
// Critical flow for payroll integration
Email → Employee.TeamEmail → Employee.BasecampID → Todo.Assignees[0].ID
```

### Notion API Characteristics

1. **Relation Properties**: Store only page references (UUIDs), not actual data
2. **Rollup Properties**: Aggregate data from related pages (computed by Notion)
3. **Rollup Types**: Can be array, number, date, or string based on configuration
4. **Query Efficiency**: Rollup data included in database query response (no extra call)

### Existing Implementation Pattern

The NocoDB implementation accesses email directly as a flat field:

```go
// NocoDB: Direct field access
requesterEmail := extractString(record, "requester_team_email")
```

However, Notion requires extracting from the rollup property structure.

## Decision

We will use a **two-tier extraction strategy** that prioritizes the rollup property for efficiency, with a fallback to direct relation query if rollup extraction fails.

### Primary Strategy: Extract from Rollup Property

```go
func (s *NotionExpenseService) extractEmailFromRollup(props notion.DatabasePageProperties) (string, error) {
    emailProp, ok := props["Email"]
    if !ok {
        return "", fmt.Errorf("Email property not found")
    }

    if emailProp.Type != notion.DBPropTypeRollup {
        return "", fmt.Errorf("Email is not a rollup property")
    }

    rollup := emailProp.Rollup

    // Handle different rollup types
    switch rollup.Type {
    case notion.RollupTypeArray:
        // Most common: rollup returns array of property values
        if len(rollup.Array) > 0 {
            // Type assert to DatabasePageProperty
            if propVal, ok := rollup.Array[0].(notion.DatabasePageProperty); ok {
                if propVal.Type == notion.DBPropTypeEmail {
                    return propVal.Email, nil
                }
            }
        }
        return "", fmt.Errorf("rollup array is empty or invalid")

    case notion.RollupTypeString:
        // Some rollup configurations return direct string
        if rollup.String != "" {
            return rollup.String, nil
        }
        return "", fmt.Errorf("rollup string is empty")

    default:
        return "", fmt.Errorf("unsupported rollup type: %s", rollup.Type)
    }
}
```

### Fallback Strategy: Query Relation Directly

```go
func (s *NotionExpenseService) extractEmailFromRelation(ctx context.Context, props notion.DatabasePageProperties) (string, error) {
    requestorProp, ok := props["Requestor"]
    if !ok || requestorProp.Type != notion.DBPropTypeRelation {
        return "", fmt.Errorf("Requestor relation not found")
    }

    if len(requestorProp.Relation) == 0 {
        return "", fmt.Errorf("no contractor linked to expense")
    }

    // Get first related contractor page
    contractorPageID := requestorProp.Relation[0].ID

    // Query contractor page for email
    contractorPage, err := s.client.FindPageByID(ctx, contractorPageID)
    if err != nil {
        return "", fmt.Errorf("failed to fetch contractor page: %w", err)
    }

    // Extract email from contractor properties
    contractorProps := contractorPage.Properties.(notion.DatabasePageProperties)
    emailProp, ok := contractorProps["Email"]
    if !ok || emailProp.Type != notion.DBPropTypeEmail {
        return "", fmt.Errorf("email not found in contractor page")
    }

    return emailProp.Email, nil
}
```

### Combined Implementation

```go
func (s *NotionExpenseService) getEmail(ctx context.Context, props notion.DatabasePageProperties) (string, error) {
    // Try rollup first (efficient, single query)
    email, err := s.extractEmailFromRollup(props)
    if err == nil && email != "" {
        return email, nil
    }

    // Log rollup failure for monitoring
    s.logger.Warn("Rollup extraction failed, falling back to direct relation query", "error", err)

    // Fallback to direct relation query
    return s.extractEmailFromRelation(ctx, props)
}
```

## Alternatives Considered

### Option 1: Rollup Only (No Fallback)

Use only rollup extraction, fail if rollup is misconfigured.

**Rejected because**:
- Single point of failure (misconfigured rollup breaks all expense fetching)
- No graceful degradation for edge cases
- Harder to debug rollup configuration issues
- Risk of production failures due to Notion UI changes

### Option 2: Relation Query Only

Always query the Contractor database directly, ignore rollup.

**Rejected because**:
- N+1 query problem (one extra API call per expense)
- Slower performance (2× API calls vs 1× with rollup)
- Unnecessary load on Notion API (hits rate limits faster)
- Ignores efficiency benefits of Notion's rollup feature

### Option 3: Cache Contractor Database

Pre-fetch entire Contractor database, lookup emails in memory.

**Rejected because**:
- Additional complexity (cache invalidation, memory management)
- Large contractor database may not fit efficiently in memory
- Still requires API call to fetch contractors initially
- Rollup approach is simpler and more efficient

## Consequences

### Positive

1. **Performance Optimization**: Primary path uses single API call (rollup included in query response)
2. **Resilience**: Fallback ensures email extraction succeeds even if rollup misconfigured
3. **Debugging**: Fallback path helps identify rollup configuration issues
4. **Flexibility**: Handles different rollup configurations (array vs string types)
5. **Notion Best Practices**: Leverages Notion's built-in rollup computation

### Negative

1. **Complexity**: Two extraction paths increase code surface area
2. **Configuration Dependency**: Relies on correct Notion rollup setup for optimal performance
3. **Type Handling**: Must handle multiple rollup types (array, string)
4. **Error Messages**: More granular error handling required for debugging
5. **Testing Overhead**: Must test both rollup and fallback paths

### Monitoring and Debugging

1. **Log Metrics**:
   - Count of successful rollup extractions
   - Count of fallback to relation query
   - Failed extractions (no email found)

2. **Alert Triggers**:
   - High fallback rate (>10%) indicates rollup misconfiguration
   - Failed extraction rate (>1%) indicates data quality issues

3. **Debug Logging**:
   ```go
   s.logger.Debug("Email extraction",
       "method", "rollup",
       "rollup_type", rollup.Type,
       "email", email,
       "page_id", pageID,
   )
   ```

## Validation

### Rollup Configuration Requirements

For optimal performance, ensure the Notion database rollup is configured as:

- **Property Name**: `Email`
- **Relation**: `Requestor` (to Contractor database)
- **Property**: `Email` (from Contractor database)
- **Calculate**: `Show original` (returns array of email values)

### Testing Strategy

1. **Unit Tests**:
   - Mock rollup with array type containing email
   - Mock rollup with string type containing email
   - Mock rollup with empty array (trigger fallback)
   - Mock relation query returning contractor page
   - Test error handling for missing properties

2. **Integration Tests**:
   - Test with real Notion database (rollup configured)
   - Test with misconfigured rollup (verify fallback)
   - Test with multiple contractors (relation array)
   - Test with missing contractor (handle gracefully)

3. **Performance Tests**:
   - Measure API call count for 100 expenses
   - Compare rollup path vs fallback path latency
   - Verify no N+1 query pattern

### Acceptance Criteria

- [ ] Email extraction succeeds for all approved expenses (100% success rate)
- [ ] Rollup extraction used in >90% of cases (primary path)
- [ ] Fallback extraction used in <10% of cases (edge cases only)
- [ ] Email matches existing employee records in database
- [ ] Performance is comparable to NocoDB implementation (1 query vs N queries)
- [ ] Logs clearly indicate which extraction method succeeded

## Implementation Notes

### Rollup Type Detection

The rollup type depends on Notion UI configuration:

| Rollup Function | Rollup Type | Extraction Method |
|-----------------|-------------|-------------------|
| Show original   | Array       | `rollup.Array[0]` |
| Show unique     | Array       | `rollup.Array[0]` |
| Count all       | Number      | N/A (not applicable for email) |
| Count values    | Number      | N/A (not applicable for email) |

**Recommendation**: Configure rollup with "Show original" for consistent array-based extraction.

### Error Handling

```go
// Transformation should fail if email cannot be extracted
todo, err := s.transformPageToTodo(page)
if err != nil {
    s.logger.Error(err, "failed to transform page",
        "page_id", page.ID,
        "error_type", "email_extraction_failed",
    )
    continue  // Skip this expense, don't fail entire batch
}
```

### Employee Validation

After extracting email, validate it matches an existing employee:

```go
employee, err := s.store.Employee.OneByEmail(s.repo.DB(), email)
if err != nil {
    return nil, fmt.Errorf("employee not found for email %s: %w", email, err)
}

if employee.BasecampID == 0 {
    return nil, fmt.Errorf("employee %s has no basecamp_id", email)
}
```

## References

- **Research**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/research/notion-api-patterns.md` (Section 4: Handling Rollup Properties)
- **Technical Considerations**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/research/technical-considerations.md` (Section 3: Relation and Rollup Handling)
- **NocoDB Implementation**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/nocodb/expense.go` (transformRecordToTodo)
- **Notion API**: [Property values - Rollup](https://developers.notion.com/reference/page-property-values#rollup)

## Future Considerations

1. **Caching**: If fallback usage is high, consider caching contractor data
2. **Batch Queries**: Explore Notion API for batch relation lookups (if available)
3. **Monitoring Dashboard**: Track rollup vs fallback usage over time
4. **Configuration Validation**: Add startup check to verify rollup is properly configured
