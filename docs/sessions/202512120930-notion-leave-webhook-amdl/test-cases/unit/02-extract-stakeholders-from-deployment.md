# Test Case Design: ExtractStakeholdersFromDeployment

**Function Under Test:** `LeaveService.ExtractStakeholdersFromDeployment(deploymentPage nt.Page) ([]string, error)`

**Location:** `pkg/service/notion/leave.go`

**Purpose:** Extract Account Manager (AM) and Delivery Lead (DL) contractor page IDs from a deployment page, following override priority logic.

## Test Strategy

### Approach
- **Pattern:** Table-driven tests with property extraction scenarios
- **Dependencies:** No external dependencies (pure property extraction)
- **Validation:** Assert correct stakeholder IDs extracted based on override logic
- **Focus:** Override vs rollup priority, property type handling, deduplication

### Test Data Requirements
- Mock `nt.Page` objects with various property configurations
- Notion property structures: relation, rollup (array of relations)
- Edge cases: empty properties, malformed data

## Priority Logic to Test

### Override Priority Rules
1. **AM Extraction:**
   - If "Override AM" has relation → use Override AM
   - Else if "Account Managers" rollup has data → use rollup
   - Else → no AM stakeholders

2. **DL Extraction:**
   - If "Override DL" has relation → use Override DL
   - Else if "Delivery Leads" rollup has data → use rollup
   - Else → no DL stakeholders

3. **Deduplication:**
   - Combine AM and DL lists
   - Remove duplicate contractor page IDs

## Test Cases

### TC-2.1: Override AM and DL Present - Use Overrides Only

**Given:**
- Deployment page with:
  - "Override AM" relation: `["am-override-123"]`
  - "Override DL" relation: `["dl-override-456"]`
  - "Account Managers" rollup: `["am-rollup-999"]` (should be ignored)
  - "Delivery Leads" rollup: `["dl-rollup-888"]` (should be ignored)

**When:**
- Call `ExtractStakeholdersFromDeployment(deploymentPage)`

**Then:**
- Should return: `["am-override-123", "dl-override-456"]`
- Should NOT include rollup values
- Should return nil error
- Should log DEBUG: "using override AM/DL"

**Test Data:**
```go
props := nt.DatabasePageProperties{
    "Override AM": nt.DatabasePageProperty{
        Relation: []nt.Relation{{ID: "am-override-123"}},
    },
    "Override DL": nt.DatabasePageProperty{
        Relation: []nt.Relation{{ID: "dl-override-456"}},
    },
    "Account Managers": nt.DatabasePageProperty{
        Rollup: nt.RollupResult{
            Type: "array",
            Array: []nt.DatabasePageProperty{
                {Relation: []nt.Relation{{ID: "am-rollup-999"}}},
            },
        },
    },
    "Delivery Leads": nt.DatabasePageProperty{
        Rollup: nt.RollupResult{
            Type: "array",
            Array: []nt.DatabasePageProperty{
                {Relation: []nt.Relation{{ID: "dl-rollup-888"}}},
            },
        },
    },
}
```

---

### TC-2.2: No Override - Use Rollup AM and DL

**Given:**
- Deployment page with:
  - "Override AM" relation: `[]` (empty)
  - "Override DL" relation: `[]` (empty)
  - "Account Managers" rollup: `["am-rollup-111", "am-rollup-222"]`
  - "Delivery Leads" rollup: `["dl-rollup-333"]`

**When:**
- Call `ExtractStakeholdersFromDeployment(deploymentPage)`

**Then:**
- Should return: `["am-rollup-111", "am-rollup-222", "dl-rollup-333"]`
- Should use rollup values since overrides are empty
- Should return nil error
- Should log DEBUG: "no override, using rollup AM/DL"

**Test Data:**
```go
props := nt.DatabasePageProperties{
    "Override AM": nt.DatabasePageProperty{
        Relation: []nt.Relation{}, // Empty
    },
    "Override DL": nt.DatabasePageProperty{
        Relation: []nt.Relation{}, // Empty
    },
    "Account Managers": nt.DatabasePageProperty{
        Rollup: nt.RollupResult{
            Type: "array",
            Array: []nt.DatabasePageProperty{
                {Relation: []nt.Relation{{ID: "am-rollup-111"}}},
                {Relation: []nt.Relation{{ID: "am-rollup-222"}}},
            },
        },
    },
    "Delivery Leads": nt.DatabasePageProperty{
        Rollup: nt.RollupResult{
            Type: "array",
            Array: []nt.DatabasePageProperty{
                {Relation: []nt.Relation{{ID: "dl-rollup-333"}}},
            },
        },
    },
}
```

---

### TC-2.3: Mixed - Override AM, Rollup DL

**Given:**
- Deployment page with:
  - "Override AM" relation: `["am-override-123"]`
  - "Override DL" relation: `[]` (empty)
  - "Account Managers" rollup: `["am-rollup-999"]` (ignored)
  - "Delivery Leads" rollup: `["dl-rollup-333"]`

**When:**
- Call `ExtractStakeholdersFromDeployment(deploymentPage)`

**Then:**
- Should return: `["am-override-123", "dl-rollup-333"]`
- Should use override for AM, rollup for DL
- Should return nil error

---

### TC-2.4: Mixed - Rollup AM, Override DL

**Given:**
- Deployment page with:
  - "Override AM" relation: `[]` (empty)
  - "Override DL" relation: `["dl-override-456"]`
  - "Account Managers" rollup: `["am-rollup-111"]`
  - "Delivery Leads" rollup: `["dl-rollup-999"]` (ignored)

**When:**
- Call `ExtractStakeholdersFromDeployment(deploymentPage)`

**Then:**
- Should return: `["am-rollup-111", "dl-override-456"]`
- Should use rollup for AM, override for DL
- Should return nil error

---

### TC-2.5: Deduplication - Same Stakeholder as AM and DL

**Given:**
- Deployment page with:
  - "Override AM" relation: `["stakeholder-xyz"]`
  - "Override DL" relation: `["stakeholder-xyz"]` (same person)

**When:**
- Call `ExtractStakeholdersFromDeployment(deploymentPage)`

**Then:**
- Should return: `["stakeholder-xyz"]` (deduplicated, appears once)
- Should NOT return: `["stakeholder-xyz", "stakeholder-xyz"]`
- Should return nil error
- Should log DEBUG: "deduplicated stakeholders: original count=2, unique count=1"

---

### TC-2.6: Rollup with Multiple Relations per Array Item

**Given:**
- Deployment page with:
  - "Account Managers" rollup array item containing multiple relations:
    ```
    Array[0].Relation = [{ID: "am-1"}, {ID: "am-2"}]
    ```

**When:**
- Call `ExtractStakeholdersFromDeployment(deploymentPage)`

**Then:**
- Should flatten and return: `["am-1", "am-2"]`
- Should handle multiple relations within a single rollup array item
- Should return nil error

**Test Data:**
```go
"Account Managers": nt.DatabasePageProperty{
    Rollup: nt.RollupResult{
        Type: "array",
        Array: []nt.DatabasePageProperty{
            {
                Relation: []nt.Relation{
                    {ID: "am-1"},
                    {ID: "am-2"},
                },
            },
        },
    },
}
```

---

### TC-2.7: Empty Deployment - No Stakeholders

**Given:**
- Deployment page with:
  - All override fields empty: `[]`
  - All rollup fields empty: `[]` or null

**When:**
- Call `ExtractStakeholdersFromDeployment(deploymentPage)`

**Then:**
- Should return: `[]` (empty array, not nil)
- Should return nil error (not an error condition)
- Should log INFO: "no stakeholders found for deployment"

---

### TC-2.8: Property Not Found - Graceful Handling

**Given:**
- Deployment page missing expected properties:
  - "Override AM" property does not exist in properties map
  - "Account Managers" property does not exist

**When:**
- Call `ExtractStakeholdersFromDeployment(deploymentPage)`

**Then:**
- Should return: `[]` (empty array)
- Should return nil error (graceful degradation)
- Should log WARNING: "expected property not found: Override AM"

**Rationale:** Notion schema may change; function should degrade gracefully

---

### TC-2.9: Invalid Property Type - Skip and Continue

**Given:**
- Deployment page with:
  - "Override AM" is not a relation (wrong type, e.g., text)
  - "Override DL" relation: `["dl-override-456"]` (valid)

**When:**
- Call `ExtractStakeholdersFromDeployment(deploymentPage)`

**Then:**
- Should skip invalid "Override AM" property
- Should still extract valid "Override DL": `["dl-override-456"]`
- Should return partial results, nil error
- Should log WARNING: "property type mismatch: Override AM expected relation, got text"

**Rationale:** Partial data is better than complete failure

---

### TC-2.10: Rollup Type is Not Array - Graceful Handling

**Given:**
- Deployment page with:
  - "Account Managers" rollup has type "number" instead of "array"

**When:**
- Call `ExtractStakeholdersFromDeployment(deploymentPage)`

**Then:**
- Should skip invalid rollup
- Should return empty array or other valid stakeholders
- Should log WARNING: "rollup type mismatch: expected array, got number"

**Test Data:**
```go
"Account Managers": nt.DatabasePageProperty{
    Rollup: nt.RollupResult{
        Type:   "number",
        Number: 5.0, // Not an array
    },
}
```

---

### TC-2.11: Nil Page Properties - Return Error

**Given:**
- Deployment page with properties that cannot be cast to `nt.DatabasePageProperties`

**When:**
- Call `ExtractStakeholdersFromDeployment(deploymentPage)`

**Then:**
- Should return: `nil` for stakeholders array
- Should return error: "failed to cast page properties"
- Should log ERROR with page ID

**Test Data:**
```go
page := nt.Page{
    ID:         "invalid-page",
    Properties: nil, // or wrong type
}
```

---

### TC-2.12: Complex Rollup - Multiple Array Items with Multiple Relations

**Given:**
- Deployment page with:
  - "Account Managers" rollup:
    ```
    Array[0].Relation = [{ID: "am-1"}, {ID: "am-2"}]
    Array[1].Relation = [{ID: "am-3"}]
    ```

**When:**
- Call `ExtractStakeholdersFromDeployment(deploymentPage)`

**Then:**
- Should flatten all relations: `["am-1", "am-2", "am-3"]`
- Should maintain order from nested structure
- Should return nil error

**Rationale:** Rollup aggregates multiple project assignments

---

### TC-2.13: Override with Multiple Relations - Extract All

**Given:**
- Deployment page with:
  - "Override AM" relation: `["am-1", "am-2"]` (multiple overrides)

**When:**
- Call `ExtractStakeholdersFromDeployment(deploymentPage)`

**Then:**
- Should extract all overrides: includes both `"am-1"` and `"am-2"`
- Should NOT extract only first relation
- Should return nil error

**Note:** While typically overrides have single value, API allows arrays

---

## Property Extraction Helper Functions

### Helper: extractFirstRelationID
**Usage:** Extract Override AM, Override DL
**Test Coverage:**
- Relation array with single item → return ID
- Relation array with multiple items → return first ID
- Empty relation array → return empty string
- Property not found → return empty string

### Helper: extractRollupRelations
**Usage:** Extract Account Managers, Delivery Leads rollups
**Test Coverage:**
- Rollup type "array" with relations → return all IDs
- Rollup type "array" with empty array → return empty array
- Rollup type not "array" → return empty array
- Nested relations in rollup → flatten correctly
- Property not found → return empty array

## Mock Setup Requirements

### Mock Page Structure
```go
func createMockDeploymentPage(overrideAM, overrideDL []string, rollupAM, rollupDL [][]string) nt.Page {
    // Helper to construct realistic nt.Page for testing
}
```

### Assertions
- Verify correct number of stakeholders returned
- Verify stakeholder IDs match expected values
- Verify deduplication logic works
- Verify order preservation (if required by spec)

## Error Handling Strategy

### Partial Extraction
- If one property fails, continue with others
- Return all successfully extracted stakeholders
- Log warnings for failed extractions
- Only return error if page structure is completely invalid

### Error Return Conditions
1. Cannot cast page.Properties to DatabasePageProperties
2. Page is nil (defensive check)

### Warning Log Conditions
1. Expected property not found
2. Property type mismatch
3. Rollup type is not array

## Logging Assertions

### DEBUG Level
- "extracting stakeholders: page_id=%s"
- "override AM found: %v"
- "override DL found: %v"
- "using rollup AM: count=%d"
- "using rollup DL: count=%d"
- "deduplicated stakeholders: original=%d unique=%d"

### WARNING Level
- "property not found: %s"
- "property type mismatch: %s expected %s, got %s"
- "rollup type mismatch: expected array, got %s"

### ERROR Level
- "failed to cast page properties: page_id=%s"

## Configuration Dependencies

**None** - This is a pure property extraction function with no config dependencies.

## Performance Considerations

### Expected Behavior
- In-memory property extraction (fast)
- No API calls
- Typical stakeholder count: 1-4 per deployment
- Deduplication using map for O(n) complexity

### No Optimization Required
- Small data sets (< 10 stakeholders per deployment)
- Simple iteration and extraction

## Integration Notes

### Caller Expectations
- Caller: `handleNotionLeaveCreated` after `GetActiveDeploymentsForContractor`
- Caller will iterate over multiple deployment pages
- Caller expects empty array as valid response
- Caller will aggregate stakeholders from all deployments

### Downstream Usage
- Returned stakeholder IDs will be passed to `GetDiscordUsernameFromContractor`
- IDs must be valid Notion page IDs (UUID format)

## Test Implementation Checklist

- [ ] Test override priority for AM (override > rollup)
- [ ] Test override priority for DL (override > rollup)
- [ ] Test mixed scenarios (override AM + rollup DL, etc.)
- [ ] Test deduplication of same stakeholder in AM and DL
- [ ] Test empty deployment (no stakeholders)
- [ ] Test missing properties (graceful handling)
- [ ] Test invalid property types (skip and continue)
- [ ] Test rollup with multiple array items
- [ ] Test rollup with multiple relations per item
- [ ] Test nil or invalid page properties (error case)
- [ ] Verify helper functions work correctly
- [ ] Verify logging at appropriate levels
- [ ] Test complex nested rollup structures
- [ ] Test property extraction edge cases (empty strings, nulls)
