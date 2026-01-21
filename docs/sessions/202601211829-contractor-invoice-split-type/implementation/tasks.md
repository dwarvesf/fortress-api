# Implementation Tasks: Split Contractor Invoice by Type + Dry Run + Payout-Based Deduplication

**Session**: 202601211829-contractor-invoice-split-type
**Date Created**: 2026-01-21
**Status**: Ready for Implementation
**Priority**: High

## Overview

This document breaks down the implementation of:
1. **invoiceType**: Split invoice generation into two types (`service_and_refund`, `extra_payment`)
2. **dryRun**: Generate PDF without creating Contractor Payables record in Notion
3. **Payout-based deduplication**: Use payout item IDs to identify existing payables for update vs create

### Problem Statement

Current `findExistingPayable()` queries by `(Contractor, Period)` only. With two invoice types per month:
- The second invoice would overwrite the first (if status "New") or skip creation (if "Pending")

### Solution

Use payout items as the "hash" to identify the correct payable:
- If payout items overlap with an existing payable → update that specific payable
- If no overlap → create new payable

---

## Implementation Tasks

### Task 1: Update `findExistingPayable` Function Signature and Logic

**File(s)**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payables.go`

**Description**: Modify `findExistingPayable` to accept payout item IDs and use overlap detection to find the correct payable.

**Implementation Notes**:
- **Location**: Lines 62-130 (function `findExistingPayable`)
- **Current behavior**: Queries by `(Contractor, Period)` with `PageSize: 1`
- **New behavior**:
  1. Remove `PageSize: 1` to get ALL matching payables
  2. For each payable, extract its `Payout Items` relation IDs
  3. Check if any input `payoutItemIDs` overlap with the payable's items
  4. Return the first matching payable (overlap found)

**Code Changes**:
```go
// Change signature to accept payout item IDs
func (s *ContractorPayablesService) findExistingPayable(
    ctx context.Context,
    contractorPageID string,
    periodStart string,
    payoutItemIDs []string,  // NEW: payout items to match
) (*ExistingPayable, error)

// Query all payables for contractor + period (remove PageSize: 1)
resp, err := s.client.QueryDatabase(ctx, payablesDBID, &nt.DatabaseQuery{
    Filter: filter,
    // Remove PageSize: 1 to get all matches
})

// Check each payable for payout item overlap
for _, page := range resp.Results {
    props, ok := page.Properties.(nt.DatabasePageProperties)
    if !ok {
        continue
    }

    existingPayoutIDs := s.extractAllRelationIDs(props, "Payout Items")

    // Check for overlap
    if hasOverlap(payoutItemIDs, existingPayoutIDs) {
        status := ""
        if statusProp, exists := props["Payment Status"]; exists && statusProp.Status != nil {
            status = statusProp.Status.Name
        }
        return &ExistingPayable{PageID: page.ID, Status: status}, nil
    }
}

// No overlap found
return nil, nil
```

**Acceptance Criteria**:
- ✅ Function accepts `payoutItemIDs` parameter
- ✅ Queries all payables for contractor/period (not just first)
- ✅ Extracts payout item IDs from each payable
- ✅ Returns payable with overlapping payout items
- ✅ Returns nil when no overlap found

---

### Task 2: Add `hasOverlap` Helper Function

**File(s)**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payables.go`

**Description**: Create a helper function to check if any element in one slice exists in another.

**Code**:
```go
// hasOverlap checks if any element in slice a exists in slice b
func hasOverlap(a, b []string) bool {
    bSet := make(map[string]bool, len(b))
    for _, id := range b {
        bSet[id] = true
    }
    for _, id := range a {
        if bSet[id] {
            return true
        }
    }
    return false
}
```

**Acceptance Criteria**:
- ✅ Returns true when any element overlaps
- ✅ Returns false when no overlap
- ✅ Handles empty slices correctly
- ✅ Efficient O(n+m) implementation using map

---

### Task 3: Update `CreatePayable` to Pass Payout Item IDs

**File(s)**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payables.go`

**Description**: Update `CreatePayable` to pass payout item IDs to `findExistingPayable`.

**Implementation Notes**:
- **Location**: Lines 147-170
- **Changes**: Pass `input.PayoutItemIDs` to `findExistingPayable`

**Code Changes**:
```go
// Pass payout item IDs to findExistingPayable
existing, err := s.findExistingPayable(ctx, input.ContractorPageID, input.PeriodStart, input.PayoutItemIDs)

// Status handling (unchanged from current behavior):
if existing != nil {
    if existing.Status == "New" {
        // Update existing record with status "New"
        return s.updatePayable(ctx, existing.PageID, input)
    }
    // Status is "Pending" or "Paid" - skip
    return existing.PageID, nil
}
// No existing record found, create new
```

**Acceptance Criteria**:
- ✅ Payout item IDs passed to findExistingPayable
- ✅ Update behavior unchanged for "New" status
- ✅ Skip behavior unchanged for "Pending"/"Paid" status
- ✅ Create behavior when no overlap found

---

### Task 4: Ensure `PayoutItemIDs` is Available in CreatePayableInput

**File(s)**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payables.go`

**Description**: Verify or add `PayoutItemIDs` field to the input struct for `CreatePayable`.

**Implementation Notes**:
- Check if `CreatePayableInput` struct has `PayoutItemIDs []string` field
- If not, add it and update callers

**Acceptance Criteria**:
- ✅ `CreatePayableInput` has `PayoutItemIDs` field
- ✅ All callers of `CreatePayable` populate the field

---

### Task 5: Update Handler to Pass Payout Item IDs to Service

**File(s)**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/invoice/invoice.go`

**Description**: Ensure the invoice handler passes payout item IDs to the CreatePayable service call.

**Implementation Notes**:
- Locate where `CreatePayable` is called
- Ensure payout item IDs from the invoice generation are passed

**Acceptance Criteria**:
- ✅ Handler extracts payout item IDs from generated invoice
- ✅ Payout item IDs passed to service call
- ✅ IDs match the actual items included in the invoice

---

### Task 6: Add Unit Tests for `hasOverlap` Function

**File(s)**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payables_test.go`

**Description**: Write unit tests for the `hasOverlap` helper function.

**Test Cases**:
1. Both slices have overlapping elements → true
2. No overlapping elements → false
3. First slice empty → false
4. Second slice empty → false
5. Both slices empty → false
6. Single element overlap → true
7. Multiple overlaps → true

**Acceptance Criteria**:
- ✅ All test cases pass
- ✅ Edge cases covered (empty slices)
- ✅ Tests follow table-driven pattern

---

### Task 7: Add Unit Tests for Updated `findExistingPayable`

**File(s)**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payables_test.go`

**Description**: Write unit tests for the updated `findExistingPayable` logic.

**Test Cases**:
1. Multiple payables exist, one has overlapping payout items → returns that payable
2. Multiple payables exist, none have overlapping payout items → returns nil
3. No payables exist for contractor/period → returns nil
4. Payable with "New" status and overlap → returns it
5. Payable with "Pending" status and overlap → returns it

**Acceptance Criteria**:
- ✅ All test cases pass
- ✅ Mock Notion client responses
- ✅ Tests verify overlap detection works correctly

---

### Task 8: Integration Test for Dual Invoice Type Scenario

**File(s)**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/invoice/invoice_test.go` or similar

**Description**: Write integration test verifying the main scenario: generating two different invoice types for the same contractor/month creates separate payables.

**Test Scenario**:
1. Generate `service_and_refund` invoice with payout items [A, B, C]
   - Verify: Creates payable P1 with items [A, B, C]
2. Generate `extra_payment` invoice with payout items [X, Y]
   - Verify: Creates payable P2 with items [X, Y] (NOT updates P1)
3. Re-generate `service_and_refund` invoice with items [A, B, C, D]
   - Verify: Updates P1 (overlap detected), does NOT create new

**Acceptance Criteria**:
- ✅ Dual invoice types create separate payables
- ✅ Re-generation updates correct payable
- ✅ No cross-type overwrites

---

## Behavior Matrix

| Scenario | Payout Overlap | Status | Action |
|----------|---------------|--------|--------|
| New contractor/period | None | N/A | Create new payable |
| Same/overlapping payout items | Yes | New | Update existing payable |
| Same/overlapping payout items | Yes | Pending | Skip (return existing ID) |
| Same/overlapping payout items | Yes | Paid | Skip (return existing ID) |
| Different payout items (other type) | None | N/A | Create new payable |

---

## Task Dependencies

```
Task 1 (findExistingPayable signature)
    ↓
Task 2 (hasOverlap helper)
    ↓
Task 3 (CreatePayable update)
    ↓
Task 4 (Verify PayoutItemIDs in input)
    ↓
Task 5 (Handler update)
    ↓
Task 6 (hasOverlap tests) ← Can start after Task 2
Task 7 (findExistingPayable tests) ← Can start after Task 3
    ↓
Task 8 (Integration test) ← Requires all code changes
```

**Critical Path**: Tasks 1 → 2 → 3 → 4 → 5 → 8

---

## Pre-Implementation Checklist

Before starting implementation:
- [ ] Review existing `findExistingPayable` implementation
- [ ] Identify where payout item IDs are available in the flow
- [ ] Verify `extractAllRelationIDs` function exists or needs creation
- [ ] Confirm Notion database schema has "Payout Items" relation field

---

## Post-Implementation Checklist

After completing all tasks:
- [ ] All unit tests pass (`make test`)
- [ ] Integration test passes
- [ ] Manual verification with dual invoice types
- [ ] Debug logging verified
- [ ] PR created with detailed description
- [ ] Code review requested

---

## Rollback Plan

If issues are discovered after deployment:

**Immediate Rollback**:
```bash
git revert <commit-hash>
git push origin develop
```

**Partial Rollback**:
- Revert to previous `findExistingPayable` behavior (query by contractor/period only)
- Keep other improvements

---

**End of Implementation Tasks Document**
