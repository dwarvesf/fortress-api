# Invoice Splits Not Generated - Root Cause Analysis

**Date:** 2026-01-28
**Invoice:** INV-202601-BOTTS-0MIVR
**Status:** Documented

## Summary

When marking invoice `INV-202601-BOTTS-0MIVR` as paid via Discord command, **0 commission splits were created** despite commission percentages and amounts being configured on the line item.

## Observed Behavior

From the logs:

```json
{"msg":"commission percentages: sales=5.00%, am=2.00%, dl=0.00%, hr=0.00%"}
{"msg":"commission amounts: sales=125.00, am=50.00, dl=0.00, hr=0.00"}
{"msg":"person IDs: sales=[], am=[], dl=[], hr=[]"}
{"msg":"found 1 line items with commissions"}
{"msg":"created 0 commission splits"}
```

**Key observation:** Commission amounts exist ($125 for Sales, $50 for Account Manager), but all person ID arrays are empty.

## Root Cause

The Invoice Line Item does not have a **"Deployment Tracker" relation** linked.

### Code Flow Analysis

**File:** `pkg/service/notion/invoice.go`

```go
// Lines 753-758: Extract Deployment Tracker relation
if deploymentProp, ok := props["Deployment Tracker"]; ok && deploymentProp.Relation != nil {
    if len(deploymentProp.Relation) > 0 {
        data.DeploymentPageID = deploymentProp.Relation[0].ID
        l.Debugf("extracted deployment page ID: %s", data.DeploymentPageID)  // This log never appears
    }
}

// Lines 780-790: Only fetch contractor IDs if DeploymentPageID exists
if data.DeploymentPageID != "" {
    contractorIDs, err := n.getContractorIDsFromDeployment(data.DeploymentPageID)
    // ... populate person IDs
}
// If DeploymentPageID is empty, person IDs remain empty arrays
```

**File:** `pkg/worker/worker.go`

```go
// Lines 188-189: Create splits for each person
for _, personID := range role.personIDs {  // Empty array = 0 iterations
    // Create split...
}
```

## Notion Schema Relationships

### Client Invoices Database

| Property | Type | Source |
|----------|------|--------|
| `Deployment Tracker` | relation | → Project Deployment database |
| `% Sales` | number | Manual input (percentage) |
| `% Account Mgr` | number | Manual input (percentage) |
| `% Delivery Lead` | number | Manual input (percentage) |
| `% Hiring Referral` | number | Manual input (percentage) |
| `Sales Amount` | formula | `Line Total × % Sales` |
| `Account Amount` | formula | `Line Total × % Account Mgr` |
| `Sales Person` | rollup | → `Deployment Tracker.Final Sales Credit` |
| `Account Manager` | formula | → `Deployment Tracker.Account Managers` |
| `Delivery Lead` | formula | → `Deployment Tracker.Delivery Leads` |
| `Hiring Referral` | formula | → `Deployment Tracker.Hiring Referral` |
| `Override Sales` | relation | → Contractors (manual override) |
| `Override AM` | relation | → Contractors (manual override) |
| `Override DL` | relation | → Contractors (manual override) |
| `Override Hiring Referral (LI)` | relation | → Contractors (manual override) |

### Project Deployment Database

| Property | Type | Source |
|----------|------|--------|
| `Project` | relation | → Projects database |
| `Contractor` | relation | → Contractors database |
| `Original Sales` | rollup | → `Project.Sales` |
| `Account Managers` | rollup | → `Project.Account Managers` |
| `Delivery Leads` | rollup | → `Project.Delivery Leads` |
| `Final Sales Credit` | formula | `Upsell Person` if set, else `Original Sales` |
| `Hiring Referral` | formula | Derived from contractor's candidate referrer |
| `Upsell Person` | relation | → Contractors (override) |
| `Override AM` | relation | → Contractors (override) |
| `Override DL` | relation | → Contractors (override) |

## Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     Invoice Line Item                            │
│  ┌─────────────────┐    ┌─────────────────┐                     │
│  │ % Sales: 5%     │    │ Sales Amount:   │                     │
│  │ % AM: 2%        │    │ $125 (calc)     │                     │
│  │ % DL: 0%        │    │ AM Amount:      │                     │
│  │ % HR: 0%        │    │ $50 (calc)      │                     │
│  └─────────────────┘    └─────────────────┘                     │
│                                                                  │
│  ┌─────────────────────────────────────────┐                    │
│  │ Deployment Tracker: [EMPTY] ❌           │ ← ROOT CAUSE      │
│  └─────────────────────────────────────────┘                    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼ (No link, cannot traverse)
┌─────────────────────────────────────────────────────────────────┐
│                    Project Deployment                            │
│  ┌─────────────────┐    ┌─────────────────┐                     │
│  │ Project: BOTTS  │───▶│ Sales Person    │                     │
│  │                 │    │ Account Mgr     │                     │
│  │                 │    │ Delivery Lead   │                     │
│  └─────────────────┘    └─────────────────┘                     │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Project                                  │
│  ┌─────────────────────────────────────────┐                    │
│  │ Sales: [Contractor IDs]                 │                    │
│  │ Account Managers: [Contractor IDs]      │                    │
│  │ Delivery Leads: [Contractor IDs]        │                    │
│  └─────────────────────────────────────────┘                    │
└─────────────────────────────────────────────────────────────────┘
```

## Expected vs Actual

| Step | Expected | Actual |
|------|----------|--------|
| 1. Extract Deployment Tracker | Get deployment page ID | Empty (no relation) |
| 2. Fetch contractor IDs | Call `getContractorIDsFromDeployment()` | Skipped |
| 3. Populate person IDs | `sales=[id1], am=[id2]` | `sales=[], am=[]` |
| 4. Create splits | 2 splits (Sales + AM) | 0 splits |

## Solutions

### Immediate Fix (Data)

Link the Invoice Line Item to the appropriate **Project Deployment** record in Notion:

1. Open the Invoice Line Item for `INV-202601-BOTTS-0MIVR`
2. Set the **"Deployment Tracker"** relation to the correct deployment record
3. Ensure the deployment record is linked to the BOTTS project
4. Verify the BOTTS project has Sales and Account Manager persons assigned

### Code Enhancement (Recommended)

Add fallback to "Override" fields when Deployment Tracker is not linked:

**File:** `pkg/service/notion/invoice.go`

```go
// After line 790, add fallback logic:
if data.DeploymentPageID == "" {
    l.Debug("no deployment tracker linked, checking override fields")

    // Check Override Sales relation
    if salesProp, ok := props["Override Sales"]; ok && salesProp.Relation != nil {
        for _, rel := range salesProp.Relation {
            data.SalesPersonIDs = append(data.SalesPersonIDs, rel.ID)
        }
    }

    // Check Override AM relation
    if amProp, ok := props["Override AM"]; ok && amProp.Relation != nil {
        for _, rel := range amProp.Relation {
            data.AccountMgrIDs = append(data.AccountMgrIDs, rel.ID)
        }
    }

    // Similar for Override DL and Override Hiring Referral (LI)
}

// Add warning if commission amounts exist but no persons assigned
if data.DeploymentPageID == "" &&
   (data.SalesAmount > 0 || data.AccountMgrAmount > 0 || data.DeliveryLeadAmount > 0 || data.HiringRefAmount > 0) &&
   len(data.SalesPersonIDs) == 0 && len(data.AccountMgrIDs) == 0 &&
   len(data.DeliveryLeadIDs) == 0 && len(data.HiringRefIDs) == 0 {
    l.Warnf("line item %s has commission amounts but no persons assigned - no splits will be created", item.ID)
}
```

### Validation Enhancement

Add pre-flight validation before marking invoice as paid:

```go
// In controller/invoice, before marking as paid:
lineItems, _ := notionService.QueryLineItemsWithCommissions(invoicePageID)
for _, item := range lineItems {
    hasCommissions := item.SalesAmount > 0 || item.AccountMgrAmount > 0 ||
                      item.DeliveryLeadAmount > 0 || item.HiringRefAmount > 0
    hasPersons := len(item.SalesPersonIDs) > 0 || len(item.AccountMgrIDs) > 0 ||
                  len(item.DeliveryLeadIDs) > 0 || len(item.HiringRefIDs) > 0

    if hasCommissions && !hasPersons {
        return fmt.Errorf("line item has commissions but no persons assigned - please link Deployment Tracker or set Override fields")
    }
}
```

## Related Files

- `pkg/service/notion/invoice.go:753-810` - Commission data extraction
- `pkg/service/notion/invoice.go:827-910` - `getContractorIDsFromDeployment()`
- `pkg/worker/worker.go:170-255` - Split creation logic
- `pkg/controller/invoice/invoice.go` - Mark as paid controller

## Prevention

1. **Notion Automation:** Add a validation rule in Notion that warns when an Invoice Line Item has commission percentages > 0 but no Deployment Tracker linked
2. **API Validation:** Add pre-flight check before marking invoice as paid
3. **Logging:** Add warning-level log when commission amounts exist but no splits can be created
4. **Documentation:** Update invoice creation documentation to require Deployment Tracker linkage

---

## Solution: Use Final Formula Columns with Name-to-ID Lookup

### Problem Context

Milestone invoices typically don't have a Deployment Tracker linked because they're not tied to a specific resource deployment. The current code only extracts person IDs from Deployment Tracker, causing 0 splits for Milestone invoices even when commission percentages are configured.

### Chosen Approach: Final Formula Columns

The Client Invoices database already has **Final formula columns** that resolve the correct person using built-in priority logic:

| Formula Column | Notion Logic | Returns |
|----------------|--------------|---------|
| `Final Sales` | Override Sales → Sales Person (Deployment) → Project Sales | Person name (string) |
| `Final AM` | Override AM → Account Manager (Deployment) → Project AM | Person name (string) |
| `Final DL` | Override DL → Delivery Lead (Deployment) → Project DL | Person name (string) |
| `Hiring Referral` | From Deployment → Contractor → Candidate → Referrer | Person name (string) |

**Key Insight:** These formulas already implement the priority logic we need. The only issue is they return **names as strings**, not page IDs.

### Solution: Name-to-ID Lookup

Since Final formulas return names, we search the Contractors database by `Full Name` to get the page ID.

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Invoice Line Item                                │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │ Final Sales    = "Ngô Ngọc Trường Hân"                      │    │
│  │ Final AM       = "Ngô Ngọc Trường Hân"                      │    │
│  │ Final DL       = "Vòng Tiểu Hùng"                           │    │
│  │ Hiring Referral = "Tiêu Quang Huy"                          │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                              │                                       │
│                              ▼                                       │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │ searchContractorByName(name)                                │    │
│  │   → Query Contractors DB: Full Name == name                 │    │
│  │   → Return contractor page ID                               │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                              │                                       │
│                              ▼                                       │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │ SalesPersonIDs    = ["d5385930-f2cf-4ee5-879c-8c82e7d486ed"]│    │
│  │ AccountMgrIDs     = ["d5385930-f2cf-4ee5-879c-8c82e7d486ed"]│    │
│  │ DeliveryLeadIDs   = ["abc12345-..."]                        │    │
│  │ HiringRefIDs      = ["def67890-..."]                        │    │
│  └─────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────┘
```

### Notion Schema Details

#### Final Formula Columns (Client Invoices)

| Property | ID | Type | Logic |
|----------|-----|------|-------|
| `Final Sales` | `%5D%40lf` | formula | `Override Sales.first()` → `Sales Person.first()` → `Project.Sales` |
| `Final AM` | `DzwI` | formula | `Override AM.first()` → `Account Manager.first()` → `Project.Account Managers` |
| `Final DL` | `%3D%5D%60M` | formula | `Override DL.first()` → `Delivery Lead.first()` → `Project.Delivery Leads` |
| `Hiring Referral` | `nYwg` | formula | `Deployment.Hiring Referral` (checks Active status) |

#### Contractors Database

| Property | ID | Type | Description |
|----------|-----|------|-------------|
| Database ID | `ed2b9224-97d9-4dff-97f9-82598b61f65d` | - | Contractors database |
| `Full Name` | `title` | title | Contractor's full name (searchable) |
| `Discord` | `l%60p%5E` | rich_text | Discord username |
| `Status` | `Xfmt` | select | Active/Inactive/Intern/Apprentices |

### API Response Examples

#### Final Formula Response (Line Item)

```json
{
  "Final Sales": {
    "id": "%5D%40lf",
    "type": "formula",
    "formula": {
      "type": "string",
      "string": "Ngô Ngọc Trường Hân"
    }
  },
  "Final AM": {
    "id": "DzwI",
    "type": "formula",
    "formula": {
      "type": "string",
      "string": "Ngô Ngọc Trường Hân"
    }
  },
  "Final DL": {
    "id": "%3D%5D%60M",
    "type": "formula",
    "formula": {
      "type": "string",
      "string": "Vòng Tiểu Hùng"
    }
  },
  "Hiring Referral": {
    "id": "nYwg",
    "type": "formula",
    "formula": {
      "type": "string",
      "string": "Tiêu Quang Huy"
    }
  }
}
```

#### Contractor Search Query

```json
{
  "filter": {
    "property": "Full Name",
    "title": {
      "equals": "Ngô Ngọc Trường Hân"
    }
  },
  "page_size": 1
}
```

#### Contractor Search Response

```json
{
  "results": [
    {
      "id": "d5385930-f2cf-4ee5-879c-8c82e7d486ed",
      "properties": {
        "Full Name": {
          "title": [{"plain_text": "Ngô Ngọc Trường Hân"}]
        },
        "Discord": {
          "rich_text": [{"plain_text": "baddeed"}]
        },
        "Status": {
          "select": {"name": "Active"}
        }
      }
    }
  ]
}
```

### Implementation Plan

**File:** `pkg/service/notion/invoice.go`

#### 1. Add constants for Contractors database

```go
const (
    // ContractorsDBID is the Notion database ID for Contractors
    ContractorsDBID = "ed2b9224-97d9-4dff-97f9-82598b61f65d"
)
```

#### 2. Add helper to extract formula string value

```go
// extractFormulaString extracts a string value from a formula property
func (n *notionService) extractFormulaString(props nt.DatabasePageProperties, propName string) string {
    if prop, ok := props[propName]; ok && prop.Formula != nil && prop.Formula.String != nil {
        return *prop.Formula.String
    }
    return ""
}
```

#### 3. Add function to search contractor by name

```go
// searchContractorByName searches the Contractors database by Full Name and returns the page ID
func (n *notionService) searchContractorByName(ctx context.Context, name string) (string, error) {
    if name == "" {
        return "", nil
    }

    l := n.l.Fields(logger.Fields{
        "service": "notion",
        "method":  "searchContractorByName",
        "name":    name,
    })

    l.Debug("searching contractor by name")

    filter := &nt.DatabaseQueryFilter{
        Property: "Full Name",
        DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
            Title: &nt.TextPropertyFilter{
                Equals: name,
            },
        },
    }

    results, err := n.notionClient.QueryDatabase(ctx, ContractorsDBID, &nt.DatabaseQuery{
        Filter:   filter,
        PageSize: 1,
    })
    if err != nil {
        l.Error(err, "failed to search contractor by name")
        return "", fmt.Errorf("failed to search contractor: %w", err)
    }

    if len(results.Results) == 0 {
        l.Debugf("no contractor found with name: %s", name)
        return "", nil
    }

    contractorID := results.Results[0].ID
    l.Debugf("found contractor: %s", contractorID)
    return contractorID, nil
}
```

#### 4. Add function to resolve contractor IDs from Final formulas

```go
// resolveContractorIDsFromFinalFormulas extracts names from Final formula columns
// and looks up the corresponding contractor page IDs
func (n *notionService) resolveContractorIDsFromFinalFormulas(
    ctx context.Context,
    props nt.DatabasePageProperties,
) (*DeploymentContractorIDs, error) {
    l := n.l.Fields(logger.Fields{
        "service": "notion",
        "method":  "resolveContractorIDsFromFinalFormulas",
    })

    result := &DeploymentContractorIDs{}

    // Extract names from Final formula columns
    finalSalesName := n.extractFormulaString(props, "Final Sales")
    finalAMName := n.extractFormulaString(props, "Final AM")
    finalDLName := n.extractFormulaString(props, "Final DL")
    hiringRefName := n.extractFormulaString(props, "Hiring Referral")

    l.Debugf("extracted final names: sales=%q, am=%q, dl=%q, hr=%q",
        finalSalesName, finalAMName, finalDLName, hiringRefName)

    // Look up contractor IDs by name (with deduplication)
    nameToID := make(map[string]string)

    // Helper to get ID, using cache to avoid duplicate lookups
    lookupID := func(name string) string {
        if name == "" {
            return ""
        }
        if id, ok := nameToID[name]; ok {
            return id
        }
        id, err := n.searchContractorByName(ctx, name)
        if err != nil {
            l.Debugf("failed to lookup contractor %q: %v", name, err)
            return ""
        }
        nameToID[name] = id
        return id
    }

    // Resolve each role
    if salesID := lookupID(finalSalesName); salesID != "" {
        result.SalesIDs = []string{salesID}
    }
    if amID := lookupID(finalAMName); amID != "" {
        result.AccountMgrIDs = []string{amID}
    }
    if dlID := lookupID(finalDLName); dlID != "" {
        result.DeliveryLeadIDs = []string{dlID}
    }
    if hrID := lookupID(hiringRefName); hrID != "" {
        result.HiringRefIDs = []string{hrID}
    }

    l.Debugf("resolved contractor IDs: sales=%v, am=%v, dl=%v, hr=%v",
        result.SalesIDs, result.AccountMgrIDs, result.DeliveryLeadIDs, result.HiringRefIDs)

    return result, nil
}
```

#### 5. Update QueryLineItemsWithCommissions

Replace the current person ID extraction logic with:

```go
// In QueryLineItemsWithCommissions, replace lines 779-790 with:

// Resolve contractor IDs from Final formula columns
// This handles all priority logic (Override → Deployment → Project) via Notion formulas
contractorIDs, err := n.resolveContractorIDsFromFinalFormulas(ctx, props)
if err != nil {
    l.Debugf("failed to resolve contractor IDs from final formulas: %v", err)
} else {
    data.SalesPersonIDs = contractorIDs.SalesIDs
    data.AccountMgrIDs = contractorIDs.AccountMgrIDs
    data.DeliveryLeadIDs = contractorIDs.DeliveryLeadIDs
    data.HiringRefIDs = contractorIDs.HiringRefIDs
}

l.Debugf("person IDs: sales=%v, am=%v, dl=%v, hr=%v",
    data.SalesPersonIDs, data.AccountMgrIDs, data.DeliveryLeadIDs, data.HiringRefIDs)

// Add warning if commission amounts exist but no persons found
hasCommissions := data.SalesAmount > 0 || data.AccountMgrAmount > 0 ||
                  data.DeliveryLeadAmount > 0 || data.HiringRefAmount > 0
hasPersons := len(data.SalesPersonIDs) > 0 || len(data.AccountMgrIDs) > 0 ||
              len(data.DeliveryLeadIDs) > 0 || len(data.HiringRefIDs) > 0

if hasCommissions && !hasPersons {
    l.Warnf("line item %s has commission amounts but no persons assigned - no splits will be created", item.ID)
}
```

### Performance Considerations

#### API Calls per Line Item

| Scenario | API Calls | Notes |
|----------|-----------|-------|
| All same person | 1 | Cached after first lookup |
| All different persons | 4 | One per unique name |
| Some empty names | 1-3 | Only non-empty names queried |

#### Optimization: Name-to-ID Cache

The implementation uses a per-request cache (`nameToID map`) to avoid duplicate lookups when the same person appears in multiple roles (e.g., Sales and AM are the same person).

### Edge Cases

| Case | Handling |
|------|----------|
| Empty name (formula returns `""`) | Skip lookup, return empty ID |
| Name not found in Contractors | Log debug, return empty ID |
| Multiple contractors with same name | Returns first match (unlikely in practice) |
| Unicode/special characters in name | Exact match required |
| Contractor marked Inactive | Still returned (status not filtered) |

### Testing Scenarios

| Scenario | Final Sales | Final AM | Expected |
|----------|-------------|----------|----------|
| Resource invoice | "Person A" | "Person B" | Lookup both, create 2 splits |
| Milestone + Override | "Person A" | "Person A" | Lookup once, create 2 splits |
| No assignments | `""` | `""` | Skip lookups, warning log |
| Unknown name | "Unknown Person" | - | Log not found, 0 sales splits |

### Advantages of This Approach

1. **Leverages Existing Logic**: Notion formulas already implement the priority logic (Override → Deployment → Project)
2. **No Code Changes for Formula Updates**: If Notion formula logic changes, no code updates needed
3. **Simpler Code**: Single resolution path instead of multiple fallback checks
4. **Consistent with UI**: Same person shown in Notion UI will be used for splits

### Potential Risks

1. **Name Matching**: Depends on exact name match; typos or variations will fail
2. **Performance**: 1-4 extra API calls per line item
3. **Name Uniqueness**: If two contractors have identical names, wrong person might be selected

### Mitigation Strategies

1. **Logging**: Comprehensive debug logs for troubleshooting
2. **Warning on Mismatch**: Log warning when name lookup fails
3. **Monitoring**: Track cases where names are not found

---

## Backward Compatibility

The new implementation replaces the current `getContractorIDsFromDeployment` approach. The old code can be removed since:

1. Final formulas already pull from Deployment Tracker when available
2. The name lookup approach works for all invoice types (Resource and Milestone)
3. Override fields are respected via the Final formula logic

### Code to Remove

```go
// Can be removed after implementing new approach:
// - getContractorIDsFromDeployment() function (lines 827-909)
// - DeploymentContractorIDs struct can be repurposed or renamed
```

### Migration Checklist

- [ ] Add `ContractorsDBID` constant
- [ ] Add `extractFormulaString()` helper
- [ ] Add `searchContractorByName()` function
- [ ] Add `resolveContractorIDsFromFinalFormulas()` function
- [ ] Update `QueryLineItemsWithCommissions()` to use new approach
- [ ] Add comprehensive logging
- [ ] Test with Resource invoice (has Deployment Tracker)
- [ ] Test with Milestone invoice (no Deployment Tracker)
- [ ] Test with Override fields set
- [ ] Verify splits are created correctly

---

## User Actions

With this implementation, users don't need to do anything special. The system will automatically:

1. Read the `Final Sales`, `Final AM`, `Final DL`, `Hiring Referral` formula values
2. Look up the corresponding contractor by name
3. Create splits with the correct contractor IDs

### For Milestone Invoices

Ensure one of these is set (Notion formula priority):

1. **Override Fields** on Line Item (highest priority)
2. **Project Assignments** - Sales/Account Managers/Delivery Leads on the Project

### For Resource Invoices

No change needed - Deployment Tracker will be used via the formula logic.

---

## References

### Notion Databases

| Database | ID | Purpose |
|----------|-----|---------|
| Client Invoices | `2bf64b29-b84c-80e2-8cc7-000bfe534203` | Invoices and Line Items |
| Project Deployment | `2b864b29-b84c-80f8-b16e-000b8e8ad2b4` | Resource deployments |
| Projects | `2988f9de-9886-4c6f-a3ff-7f7ef74b3732` | Project definitions |
| Contractors | `ed2b9224-97d9-4dff-97f9-82598b61f65d` | Contractor profiles |
| Invoice Splits | `2c364b29-b84c-804f-9856-000b58702dea` | Commission splits |

### Key Property IDs

| Property | Database | ID | Type |
|----------|----------|-----|------|
| Final Sales | Client Invoices | `%5D%40lf` | formula |
| Final AM | Client Invoices | `DzwI` | formula |
| Final DL | Client Invoices | `%3D%5D%60M` | formula |
| Hiring Referral | Client Invoices | `nYwg` | formula |
| Full Name | Contractors | `title` | title |

### Related Files

| File | Purpose |
|------|---------|
| `pkg/service/notion/invoice.go` | Commission data extraction |
| `pkg/worker/worker.go` | Split creation logic |
| `pkg/controller/invoice/invoice.go` | Mark as paid controller |

### Discord

- Invoice notifications channel: `1440178115610411068`
