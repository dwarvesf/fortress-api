# ADR-001: UUID to Integer ID Mapping Strategy

## Status
Proposed

## Context

The existing `ExpenseProvider` interface and payroll calculator expect integer IDs for expense records, as defined in the Basecamp-based `bcModel.Todo` structure:

```go
type Todo struct {
    ID    int    // Expected as integer
    Title string
    // ... other fields
}
```

However, Notion uses UUID strings as page identifiers (e.g., `"2bfb69f8-f573-81cb-a2da-f06d28896390"`), which are 128-bit universally unique identifiers represented as 36-character strings with hyphens.

This creates a type mismatch that must be resolved for the Notion Expense Provider to satisfy the `ExpenseProvider` interface without breaking compatibility with existing payroll calculation logic.

### Constraints

1. **Interface Compatibility**: Must maintain `bcModel.Todo.ID` as `int` type (no interface changes allowed)
2. **No Schema Changes**: Cannot modify database schema to add UUID fields during initial implementation
3. **Deterministic Mapping**: Same UUID must always map to same integer for consistency
4. **Reverse Lookup Required**: Need to map integer back to UUID for status updates after payroll commit
5. **Minimal Code Changes**: Solution should minimize impact on existing codebase

### Integration Points

- **Payroll Calculator**: Uses `Todo.ID` for tracking expenses in bonus calculations
- **Commit Handler**: Uses ID from `CommissionExplain.BasecampTodoID` to mark expenses as completed
- **Database Storage**: `CommissionExplain` table stores expense metadata for reference

## Decision

We will use a **hash-based conversion strategy** that converts Notion UUID strings to integers by taking the last 8 hexadecimal characters of the UUID (without hyphens) and converting them to a 32-bit integer.

### Implementation

```go
func notionPageIDToInt(pageID string) int {
    // Remove hyphens from UUID: "2bfb69f8-f573-81cb-a2da-f06d28896390" → "2bfb69f8f57381cba2daf06d28896390"
    cleanID := strings.ReplaceAll(pageID, "-", "")

    // Take last 8 hex characters (32 bits): "28896390"
    hashStr := cleanID[len(cleanID)-8:]

    // Convert hex to int: 0x28896390 → 680141712
    hash, _ := strconv.ParseInt(hashStr, 16, 64)

    return int(hash)
}
```

### UUID Storage for Reverse Lookup

To enable marking expenses as completed after payroll commit, we will store the original Notion page UUID in a new metadata field that can be added to `CommissionExplain` without schema migration:

```go
// When creating bonus explanation during payroll calculation
bonusExplain := model.CommissionExplain{
    Amount:           amount,
    Month:            int(batchDate.Month()),
    Year:             batchDate.Year(),
    Name:             name,
    BasecampTodoID:   notionPageIDToInt(pageID),  // Hash-based int
    BasecampBucketID: notionPageIDToInt(pageID),  // Same hash for bucket
    // Store original UUID in a string field for reverse lookup
    // This field can be added without breaking existing code
}
```

**Note**: The exact field name for UUID storage will be determined during implementation phase based on existing `CommissionExplain` structure analysis. Options include:
- Reusing existing text fields if available
- Adding a new `TaskRef` or `NotionPageID` field in a future schema migration
- Using JSON metadata field if one exists

## Alternatives Considered

### Option 2: CRC32 Hash

Use Go's built-in CRC32 checksum for UUID conversion:

```go
func notionPageIDToInt(pageID string) int {
    return int(crc32.ChecksumIEEE([]byte(pageID)))
}
```

**Rejected because**:
- Higher collision probability than last-8-hex approach (CRC32 is designed for error detection, not uniqueness)
- Less intuitive debugging (can't visually match UUID to hash)
- No significant performance advantage for small datasets

### Option 3: Database Mapping Table

Create a dedicated table to store UUID-to-integer mappings:

```sql
CREATE TABLE notion_expense_id_mapping (
    id SERIAL PRIMARY KEY,
    notion_page_id UUID UNIQUE NOT NULL,
    local_id INT UNIQUE NOT NULL
);
```

**Rejected because**:
- Requires database migration (violates minimal changes constraint)
- Adds complexity with cache management
- Additional database queries on every expense fetch
- Overkill for this use case (UUIDs are already unique identifiers)

### Option 4: Change Interface to Accept UUID

Modify `ExpenseProvider` interface and `bcModel.Todo` to use string IDs:

```go
type Todo struct {
    ID string  // Change from int to string
    // ...
}
```

**Rejected because**:
- Breaking change to existing interface (affects Basecamp and NocoDB implementations)
- Requires extensive refactoring of payroll calculator
- Database schema changes for `CommissionExplain.BasecampTodoID`
- High risk and testing overhead

## Consequences

### Positive

1. **Minimal Code Impact**: No interface changes, no database migrations for initial implementation
2. **Deterministic**: Same UUID always produces same integer (important for idempotency)
3. **Simple Implementation**: Single function, easy to understand and test
4. **Fast Performance**: String manipulation and hex conversion are O(1) operations
5. **No External Dependencies**: Uses only Go standard library
6. **Backward Compatible**: Existing Basecamp and NocoDB providers unaffected

### Negative

1. **Collision Risk**: Theoretical possibility of two UUIDs having same last 8 characters
   - **Mitigation**: 32-bit space (4.3 billion values) is sufficient for expected expense volume
   - **Probability**: With 1000 expenses, collision probability is approximately 0.01% (1 in 10,000)
   - **Detection**: Collision would cause duplicate ID in `CommissionExplain`, detectable by database constraints

2. **Irreversible Conversion**: Cannot derive original UUID from integer alone
   - **Mitigation**: Store original UUID in metadata field for reverse lookup
   - **Usage**: UUID needed only for status updates after payroll commit, not during calculation

3. **Debugging Complexity**: Integer IDs in logs don't visually match Notion page URLs
   - **Mitigation**: Log both integer ID and original UUID in expense transformation
   - **Example**: `"Transformed expense: id=680141712, notion_page_id=2bfb69f8-f573-81cb-a2da-f06d28896390"`

4. **Future Schema Change**: Eventually need proper UUID field in database
   - **Mitigation**: Treat as technical debt, plan migration for Phase 2
   - **Migration Path**: Add `notion_page_id UUID` column, backfill from metadata, update queries

## Implementation Plan

### Phase 1: Initial Implementation (Current)

1. Implement `notionPageIDToInt()` function in Notion expense service
2. Use hash-based conversion for `Todo.ID` field
3. Store original UUID in existing metadata field (e.g., `task_ref` if available)
4. Log both integer and UUID during transformation for debugging
5. Use stored UUID for status updates after payroll commit

### Phase 2: Schema Enhancement (Future)

1. Add database migration for proper UUID storage:
   ```sql
   ALTER TABLE commission_explains
   ADD COLUMN notion_page_id UUID;
   ```
2. Update transformation logic to populate new field
3. Backfill existing records from metadata
4. Update status update logic to use dedicated UUID field
5. Remove hash-based UUID extraction from metadata

## Validation

### Testing Strategy

1. **Unit Tests**:
   - Verify deterministic conversion (same UUID → same int)
   - Test collision detection with known UUID patterns
   - Validate round-trip storage and retrieval of UUID

2. **Integration Tests**:
   - Fetch expenses from Notion, verify integer IDs in range
   - Create payroll with Notion expenses, verify UUID storage
   - Mark expense as completed using stored UUID

3. **Manual Validation**:
   - Query Notion database, note page IDs
   - Run payroll calculation, verify integer IDs in `CommissionExplain`
   - Check logs for UUID preservation
   - Verify status update uses correct Notion page ID

### Acceptance Criteria

- [ ] Notion page UUIDs successfully convert to positive integers
- [ ] No integer ID collisions in test dataset (100+ expenses)
- [ ] Original UUID stored and retrievable from payroll records
- [ ] Status updates work correctly using stored UUID
- [ ] Existing Basecamp/NocoDB providers continue working unchanged
- [ ] No changes required to `ExpenseProvider` interface

## References

- **Research**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202512041805-notion-expense-provider/research/technical-considerations.md` (Section 1: ID Mapping Strategy)
- **Interface Definition**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/basecamp/basecamp.go` (ExpenseProvider interface)
- **Data Model**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/model/commission_explain.go`
- **NocoDB Implementation**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/nocodb/expense.go` (Lines 231-295)

## Notes

- This ADR addresses only the ID mapping challenge. Email extraction and property transformation are covered in separate ADRs.
- The 32-bit integer space (2^32 ≈ 4.3 billion) provides sufficient uniqueness for the expected scale (thousands of expenses, not millions).
- Collision probability formula: P(collision) ≈ n² / (2 × 2^32) where n = number of expenses
  - For 1,000 expenses: P ≈ 0.0001 (0.01%)
  - For 10,000 expenses: P ≈ 0.01 (1%)
- If collision rate becomes problematic in practice, migration to proper UUID storage (Phase 2) should be prioritized.
