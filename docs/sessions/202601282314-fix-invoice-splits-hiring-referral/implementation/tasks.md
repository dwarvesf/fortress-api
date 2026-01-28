# Implementation Tasks: Fix Invoice Splits for Milestone Invoices & Hiring Referral

**Session:** 202601282314-fix-invoice-splits-hiring-referral
**Branch:** fix/invoice-split-error-on-type-milestone
**Status:** Ready for implementation

## Problem Summary

Two related issues causing invoice splits not to be generated:

1. **Milestone Invoices**: No Deployment Tracker linked → person IDs are empty → 0 splits created
2. **Hiring Referral**: Property is a formula (returns name string), but code expects relation/rollup → always empty

## Solution Overview

Use the existing "Final" formula columns in Notion (`Final Sales`, `Final AM`, `Final DL`, `Hiring Referral`) which already implement the priority logic (Override → Deployment → Project). Since these return names as strings, perform a contractor name-to-ID lookup.

---

## Tasks

### Task 1: Add Contractors Database ID Constant

- **File(s)**: `pkg/service/notion/invoice.go`
- **Description**: Add constant for Contractors database ID to enable name lookups
- **Code Location**: Near top of file with other constants
- **Implementation**:
  ```go
  const (
      ContractorsDBID = "ed2b9224-97d9-4dff-97f9-82598b61f65d"
  )
  ```
- **Acceptance**: Constant is defined and accessible

---

### Task 2: Add Helper to Extract Formula String Value

- **File(s)**: `pkg/service/notion/invoice.go`
- **Description**: Create helper function to extract string value from Notion formula properties
- **Code Location**: Near existing `extractFormulaProp` helper (line ~920)
- **Implementation**:
  ```go
  // extractFormulaString extracts a string value from a formula property
  func (n *notionService) extractFormulaString(props nt.DatabasePageProperties, propName string) string {
      if prop, ok := props[propName]; ok && prop.Formula != nil && prop.Formula.String != nil {
          return *prop.Formula.String
      }
      return ""
  }
  ```
- **Acceptance**: Helper correctly extracts string from formula properties, returns empty string if not found

---

### Task 3: Add Contractor Name Search Function

- **File(s)**: `pkg/service/notion/invoice.go`
- **Description**: Create function to search Contractors database by Full Name and return page ID
- **Code Location**: After `getContractorIDsFromDeployment` function (line ~909)
- **Implementation**:
  - Query Contractors database with filter: `Full Name equals <name>`
  - Return first match's page ID
  - Return empty string if not found (with debug log)
  - Handle errors gracefully
- **Acceptance**:
  - Returns correct contractor ID for valid names
  - Returns empty string for unknown names
  - Logs search activity at debug level

---

### Task 4: Add Function to Resolve Contractor IDs from Final Formulas

- **File(s)**: `pkg/service/notion/invoice.go`
- **Description**: Create function that extracts names from Final formula columns and looks up contractor IDs
- **Code Location**: After `searchContractorByName` function
- **Implementation**:
  - Extract names from: `Final Sales`, `Final AM`, `Final DL`, `Hiring Referral`
  - Use name-to-ID cache to avoid duplicate lookups (same person in multiple roles)
  - Return `DeploymentContractorIDs` struct with resolved IDs
  - Log extracted names and resolved IDs at debug level
- **Acceptance**:
  - Correctly resolves all four role types
  - Deduplicates lookups for same person in multiple roles
  - Returns partial results if some names not found

---

### Task 5: Update QueryLineItemsWithCommissions to Use New Resolution Logic

- **File(s)**: `pkg/service/notion/invoice.go`
- **Description**: Replace current Deployment Tracker-based person ID extraction with Final formula-based resolution
- **Code Location**: Lines 779-790 in `QueryLineItemsWithCommissions`
- **Current Code**:
  ```go
  if data.DeploymentPageID != "" {
      contractorIDs, err := n.getContractorIDsFromDeployment(data.DeploymentPageID)
      // ...
  }
  ```
- **New Code**:
  - Always call `resolveContractorIDsFromFinalFormulas` (works for both Resource and Milestone invoices)
  - Remove dependency on `DeploymentPageID` for person resolution
  - Keep `DeploymentPageID` extraction for backward compatibility/logging
- **Acceptance**:
  - Person IDs are populated for Milestone invoices (no Deployment Tracker)
  - Person IDs are populated for Resource invoices (with Deployment Tracker)
  - Hiring Referral IDs are correctly resolved

---

### Task 6: Add Warning Log for Commission Without Persons

- **File(s)**: `pkg/service/notion/invoice.go`
- **Description**: Add warning-level log when line item has commission amounts but no persons could be resolved
- **Code Location**: After person ID resolution in `QueryLineItemsWithCommissions`
- **Implementation**:
  ```go
  hasCommissions := data.SalesAmount > 0 || data.AccountMgrAmount > 0 ||
                    data.DeliveryLeadAmount > 0 || data.HiringRefAmount > 0
  hasPersons := len(data.SalesPersonIDs) > 0 || len(data.AccountMgrIDs) > 0 ||
                len(data.DeliveryLeadIDs) > 0 || len(data.HiringRefIDs) > 0

  if hasCommissions && !hasPersons {
      l.Warnf("line item %s has commission amounts but no persons assigned", item.ID)
  }
  ```
- **Acceptance**: Warning logged when commissions exist but no persons found

---

### Task 7: Clean Up Deprecated Code (Optional)

- **File(s)**: `pkg/service/notion/invoice.go`
- **Description**: Remove or mark as deprecated the `getContractorIDsFromDeployment` function
- **Code Location**: Lines 827-909
- **Implementation**:
  - Option A: Delete function entirely (if no other callers)
  - Option B: Mark as deprecated with comment, keep for rollback safety
- **Acceptance**: Code is cleaned up or clearly marked for future removal

---

### Task 8: Add Unit Tests for New Functions

- **File(s)**: `pkg/service/notion/invoice_test.go` (create if needed)
- **Description**: Add tests for new helper functions
- **Test Cases**:
  1. `extractFormulaString` - returns string value, handles nil/empty
  2. `searchContractorByName` - mocked Notion client, returns ID or empty
  3. `resolveContractorIDsFromFinalFormulas` - end-to-end resolution with mocked client
- **Acceptance**: Tests pass, cover happy path and edge cases

---

## Task Dependencies

```
Task 1 ──┐
         ├──► Task 3 ──► Task 4 ──► Task 5 ──► Task 6
Task 2 ──┘                                      │
                                                ▼
                                            Task 7
                                                │
                                                ▼
                                            Task 8
```

## Testing Plan

1. **Unit Tests**: Task 8
2. **Manual Testing**:
   - Test with Resource invoice (has Deployment Tracker) → verify splits created
   - Test with Milestone invoice (no Deployment Tracker) → verify splits created
   - Test with Override fields set → verify overrides respected
   - Test with Hiring Referral configured → verify HR splits created

## Rollback Plan

If issues arise:
1. Keep `getContractorIDsFromDeployment` function available
2. Can switch back by modifying `QueryLineItemsWithCommissions` to use old logic
3. No database schema changes required

---

## Next Steps

Run `proceed` command to begin implementation.
