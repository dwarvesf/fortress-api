# Service Fee Items from Invoice Split Classification Issue

**Date**: 2026-01-21
**Status**: Investigation Complete - Implementation Pending
**Priority**: Medium
**Area**: Contractor Invoice PDF Generation

## Table of Contents

1. [Problem Statement](#problem-statement)
2. [Evidence](#evidence)
3. [Investigation Process](#investigation-process)
4. [Root Cause Analysis](#root-cause-analysis)
5. [Current Behavior](#current-behavior)
6. [Expected Behavior](#expected-behavior)
7. [Solution Design](#solution-design)
8. [Implementation Plan](#implementation-plan)
9. [Testing Strategy](#testing-strategy)
10. [Additional Considerations](#additional-considerations)
11. [References](#references)

---

## Problem Statement

Service Fee source type payouts from Invoice Split are incorrectly appearing in the "Fee" section of contractor invoice PDFs. Based on the code description and business requirements, these items should be classified differently based on their role/description content.

### Key Issue

- **Observed**: Service Fee items with "Delivery Lead" or "Account Management" in Description appear with label "Service Fee" but are grouped in the Fee section alongside Commission items
- **Expected**: These items should be properly classified and grouped based on their role type
- **Impact**: Incorrect invoice section grouping leads to confusion about payout types

### Example Case

**Rows 300-301** in contractor payouts table:
- **Type Label**: "Service Fee"
- **Title**: Contains "[FEE :: MUDAH :: ics3rd]"
- **Notion Relations**:
  - ✅ `02 Invoice Split` has a value
  - ❌ `00 Task Order` does NOT have a value
  - ❌ `01 Refund` does NOT have a value

---

## Evidence

### Screenshot Analysis

From contractor invoice PDF:
```
Row 300: Type = "Service Fee", Title = "[FEE :: MUDAH :: ics3rd] ..."
Row 301: Type = "Service Fee", Title = "[FEE :: MUDAH :: ics3rd] ..."
```

Both items appear in the **"Fee" section** of the invoice.

### Notion Data Structure

Items with Invoice Split relation (`02 Invoice Split`) can have different Type values:
1. **"Service Fee"** with Description containing "Delivery Lead" or "Account Management"
2. **"Commission"** for other invoice split items

---

## Investigation Process

### Phase 1: Initial Hypothesis

**Questions**:
1. How are items classified into the "Fee" section?
2. Is there logic that transforms/reclassifies Service Fee items as Commission?
3. What determines the source type of a payout?
4. Is there a mismatch between Notion data and how it's read?

### Phase 2: Code Analysis

**Files Investigated**:
- `pkg/controller/invoice/contractor_invoice.go` - Section grouping logic (lines 436-491, 980-994)
- `pkg/service/notion/contractor_payouts.go` - Payout data fetching and source type determination
- `pkg/handler/invoice/contractor_invoice.go` - API handler that triggers invoice generation

### Phase 3: Key Findings

#### Finding 1: Source Type Determination Logic

**File**: `pkg/service/notion/contractor_payouts.go:365-380`

```go
func (s *ContractorPayoutsService) determineSourceType(entry PayoutEntry) PayoutSourceType {
	if entry.TaskOrderID != "" {
		return PayoutSourceTypeServiceFee  // Highest priority
	}
	if entry.InvoiceSplitID != "" {
		return PayoutSourceTypeCommission  // ❌ ALWAYS returns Commission
	}
	if entry.RefundRequestID != "" {
		return PayoutSourceTypeRefund
	}
	return PayoutSourceTypeExtraPayment
}
```

**Critical Observation**: The function **infers** source type from relations only. It does NOT read an explicit Type field from Notion. Items with `InvoiceSplitID` are ALWAYS classified as Commission, regardless of Description content.

#### Finding 2: Fee Section Grouping Logic

**File**: `pkg/controller/invoice/contractor_invoice.go:977-984`

```go
// Group Commission items
var feeItems []ContractorInvoiceLineItem
for i, item := range items {
	items[i].Description = strings.Replace(items[i].Description, "Bonus", "Fee", -1)
	if item.Type == string(notion.PayoutSourceTypeCommission) {
		feeItems = append(feeItems, item)
	}
}
```

**Critical Observation**: Fee section **only** includes items where `Type == "Commission"`. This is why Service Fee items from Invoice Split appear here - they are misclassified as Commission by `determineSourceType()`.

#### Finding 3: Notion Type Field is a Formula

**Code Comments** (lines 518, 675, 813 in `contractor_payouts.go`):
> "Note: Type is now a formula (auto-calculated from relations)"

**Current State**:
- Type formula field exists in Notion but is NOT being extracted by the code
- Notion formula checks Description content to determine type
- Code uses relation-based inference instead of reading the formula result

**Notion Type Formula Logic**:
- If Description contains "Delivery Lead" OR "Account Management" → Type = "Service Fee"
- Otherwise (for InvoiceSplit items) → Type = "Commission"

---

## Root Cause Analysis

### The Core Problem

The code uses a **priority-based hierarchy** to infer source type from relations, but this doesn't match the actual Notion Type field logic:

```
Code Logic:
InvoiceSplitID present → ALWAYS Commission

Notion Formula Logic:
InvoiceSplitID present + Description contains keywords → Service Fee
InvoiceSplitID present + No keywords → Commission
```

### Why Items Appear in Fee Section

**Sequence of Events**:
1. Notion payout entry has `InvoiceSplitID` and Description with "Delivery Lead"
2. `determineSourceType()` sees `InvoiceSplitID` → returns Commission (ignores Description)
3. Invoice grouping logic sees `Type == Commission` → adds to Fee section
4. Result: Service Fee item appears in Fee section

### Mismatch Between Notion and Code

| Aspect | Notion | Code | Match? |
|--------|--------|------|--------|
| Type Field | Formula (reads Description) | Inferred from relations only | ❌ No |
| Service Fee Items | Can come from InvoiceSplit | Only from TaskOrder | ❌ No |
| Commission Items | From InvoiceSplit (no keywords) | From InvoiceSplit (all) | ⚠️ Partial |

---

## Current Behavior

### Source Type Classification

```
TaskOrderID present → ServiceFee
InvoiceSplitID present → Commission (regardless of Description)
RefundRequestID present → Refund
None of above → ExtraPayment
```

### Invoice Section Grouping

| Section | Contains |
|---------|----------|
| Development Work | ServiceFee from TaskOrder |
| **Fee** | Commission from InvoiceSplit |
| Extra Payment | ExtraPayment source type |
| Refund | Refund source type |

---

## Expected Behavior

### Business Requirements (Clarified)

**Items with Invoice Split relation** should be classified into two types:

1. **Service Fee Items**:
   - Criteria: Description contains "Delivery Lead" OR "Account Management"
   - Invoice Section: **"Fee"**
   - Example: "[FEE :: MUDAH :: ics3rd] Delivery Lead - Project X"

2. **Commission Items**:
   - Criteria: Description does NOT contain special keywords
   - Invoice Section: **"Extra Payment"**
   - Example: "[BONUS :: PROJECT] Performance bonus Q4"

### Updated Source Type Classification

```
TaskOrderID present → ServiceFee (Development Work section)
InvoiceSplitID present + Description contains "Delivery Lead" OR "Account Management" → ServiceFee (Fee section)
InvoiceSplitID present + No special keywords → Commission (Extra Payment section)
RefundRequestID present → Refund
None of above → ExtraPayment
```

### Updated Invoice Section Grouping

| Section | Contains |
|---------|----------|
| Development Work | ServiceFee from TaskOrder |
| **Fee** | ServiceFee from InvoiceSplit (Delivery Lead, Account Management) |
| **Extra Payment** | Commission from InvoiceSplit + ExtraPayment source type |
| Refund | Refund source type |

---

## Solution Design

### Approach

Replicate the Notion formula logic in the code instead of trying to extract the formula value. This ensures consistency and avoids additional API calls.

### Design Principles

1. **Single Source of Truth**: Description content determines type for InvoiceSplit items
2. **Case-Insensitive Matching**: Keywords checked with `strings.ToLower()` and `strings.Contains()`
3. **Explicit Classification**: Clear logic flow with documented decision points
4. **Backward Compatible**: Changes to grouping logic should not affect other payout types

### Architecture Impact

**Minimal**:
- ✅ No database schema changes
- ✅ No API contract changes
- ✅ No Notion integration changes
- ✅ Only logic updates in two files

---

## Implementation Plan

### Change 1: Update Source Type Determination

**File**: `pkg/service/notion/contractor_payouts.go`
**Location**: Lines 365-380
**Function**: `determineSourceType()`

**Current Code**:
```go
func (s *ContractorPayoutsService) determineSourceType(entry PayoutEntry) PayoutSourceType {
	if entry.TaskOrderID != "" {
		return PayoutSourceTypeServiceFee
	}
	if entry.InvoiceSplitID != "" {
		return PayoutSourceTypeCommission  // ❌ Always returns Commission
	}
	if entry.RefundRequestID != "" {
		return PayoutSourceTypeRefund
	}
	return PayoutSourceTypeExtraPayment
}
```

**New Code**:
```go
func (s *ContractorPayoutsService) determineSourceType(entry PayoutEntry) PayoutSourceType {
	if entry.TaskOrderID != "" {
		return PayoutSourceTypeServiceFee
	}

	// For InvoiceSplit items: check Description to determine type
	if entry.InvoiceSplitID != "" {
		desc := strings.ToLower(entry.Description)

		// Service Fee: Delivery Lead or Account Management roles
		// These keywords match the Notion formula logic
		if strings.Contains(desc, "delivery lead") || strings.Contains(desc, "account management") {
			return PayoutSourceTypeServiceFee
		}

		// Otherwise: Commission
		return PayoutSourceTypeCommission
	}

	if entry.RefundRequestID != "" {
		return PayoutSourceTypeRefund
	}

	return PayoutSourceTypeExtraPayment
}
```

**Changes Summary**:
- ✅ Add Description content check for items with `InvoiceSplitID`
- ✅ Case-insensitive keyword matching ("delivery lead", "account management")
- ✅ Return ServiceFee if keywords found, Commission otherwise
- ✅ Ensure `strings` package is imported at top of file

**Testing Verification**:
```go
// Add debug logging to trace classification
s.logger.Debug(fmt.Sprintf(
	"[DEBUG] contractor_payouts: determineSourceType pageID=%s InvoiceSplitID=%s Description=%q → Type=%s",
	entry.PageID,
	entry.InvoiceSplitID,
	entry.Description,
	sourceType,
))
```

---

### Change 2: Update Fee Section Grouping

**File**: `pkg/controller/invoice/contractor_invoice.go`
**Location**: Lines 977-994
**Section**: Fee section grouping logic

**Current Code**:
```go
// Group Commission items - individual display OR grouped by project
var feeItems []ContractorInvoiceLineItem
for i, item := range items {
	items[i].Description = strings.Replace(items[i].Description, "Bonus", "Fee", -1)
	if item.Type == string(notion.PayoutSourceTypeCommission) {
		feeItems = append(feeItems, item)
	}
}

if len(feeItems) > 0 {
	sections = append(sections, ContractorInvoiceSection{
		Name:         "Fee",
		IsAggregated: false,
		Items:        feeItems,
	})
	l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: created Fee section with %d items", len(feeItems)))
}
```

**New Code**:
```go
// Fee section: Service Fee items from InvoiceSplit only
// These are items with "Delivery Lead" or "Account Management" in Description
var feeItems []ContractorInvoiceLineItem
for _, item := range items {
	// Include only Service Fee items that are NOT from TaskOrder
	// (TaskOrder items go to Development Work section)
	if item.Type == string(notion.PayoutSourceTypeServiceFee) {
		// Verify it's from InvoiceSplit by checking TaskOrderID is empty
		// and ServiceRateID is empty (ServiceRateID indicates Development Work)
		if item.TaskOrderID == "" && item.ServiceRateID == "" {
			feeItems = append(feeItems, item)
		}
	}
}

if len(feeItems) > 0 {
	sections = append(sections, ContractorInvoiceSection{
		Name:         "Fee",
		IsAggregated: false,
		Items:        feeItems,
	})
	l.Debug(fmt.Sprintf(
		"[DEBUG] contractor_invoice: created Fee section with %d Service Fee items from InvoiceSplit",
		len(feeItems),
	))
}
```

**Changes Summary**:
- ❌ Remove Commission type check
- ✅ Add ServiceFee type check
- ✅ Filter to include only items from InvoiceSplit (not TaskOrder)
- ✅ Update debug logging to reflect new logic

---

### Change 3: Update Extra Payment Section Grouping

**File**: `pkg/controller/invoice/contractor_invoice.go`
**Location**: Lines 996-1012
**Section**: Extra Payment section grouping logic

**Current Code**:
```go
// Group Extra Payment items - individual display
var extraPaymentItems []ContractorInvoiceLineItem
for _, item := range items {
	if item.Type == string(notion.PayoutSourceTypeExtraPayment) {
		extraPaymentItems = append(extraPaymentItems, item)
	}
}

if len(extraPaymentItems) > 0 {
	sections = append(sections, ContractorInvoiceSection{
		Name:         "Extra Payment",
		IsAggregated: false,
		Items:        extraPaymentItems,
	})
	l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: created Extra Payment section with %d items", len(extraPaymentItems)))
}
```

**New Code**:
```go
// Extra Payment section: Commission items + ExtraPayment items
var extraPaymentItems []ContractorInvoiceLineItem
for i, item := range items {
	// Replace "Bonus" with "Fee" in descriptions for both types
	items[i].Description = strings.Replace(items[i].Description, "Bonus", "Fee", -1)

	// Include:
	// 1. Commission items (from InvoiceSplit without special Description keywords)
	// 2. ExtraPayment source type items
	if item.Type == string(notion.PayoutSourceTypeCommission) ||
		item.Type == string(notion.PayoutSourceTypeExtraPayment) {
		extraPaymentItems = append(extraPaymentItems, item)
	}
}

if len(extraPaymentItems) > 0 {
	sections = append(sections, ContractorInvoiceSection{
		Name:         "Extra Payment",
		IsAggregated: false,
		Items:        extraPaymentItems,
	})
	l.Debug(fmt.Sprintf(
		"[DEBUG] contractor_invoice: created Extra Payment section with %d items (Commission + ExtraPayment)",
		len(extraPaymentItems),
	))
}
```

**Changes Summary**:
- ✅ Add Commission type check (moved from Fee section)
- ✅ Keep ExtraPayment type check (existing)
- ✅ Move "Bonus" → "Fee" replacement here (was in Fee section loop)
- ✅ Update debug logging to reflect new logic

---

### Change 4: Verify Required Fields

**File**: `pkg/controller/invoice/contractor_invoice.go`
**Location**: Around lines 60-85
**Struct**: `ContractorInvoiceLineItem`

**Need to Verify**:
- ✅ `TaskOrderID` field exists for filtering
- ✅ `ServiceRateID` field exists for filtering
- ✅ `Type` field matches `notion.PayoutSourceType` values

**If fields are missing**: They are available during line item building (lines 378, 414) and can be used directly in filtering logic without struct changes.

---

## Testing Strategy

### Unit Test Cases

#### Test 1: Service Fee from TaskOrder → Development Work Section
**Input**:
- `TaskOrderID`: "task-order-123"
- `InvoiceSplitID`: ""
- `Description`: "Development work on Project X"

**Expected**:
- `determineSourceType()` → `PayoutSourceTypeServiceFee`
- Appears in: **Development Work** section

**Status**: ✓ Existing behavior (no change)

---

#### Test 2: Service Fee from InvoiceSplit with "Delivery Lead" → Fee Section
**Input**:
- `TaskOrderID`: ""
- `InvoiceSplitID`: "invoice-split-456"
- `Description`: "[FEE :: MUDAH :: ics3rd] Delivery Lead - Project Y"

**Expected**:
- `determineSourceType()` → `PayoutSourceTypeServiceFee`
- Appears in: **Fee** section

**Status**: ✅ New behavior (fix target)

---

#### Test 3: Service Fee from InvoiceSplit with "Account Management" → Fee Section
**Input**:
- `TaskOrderID`: ""
- `InvoiceSplitID`: "invoice-split-789"
- `Description`: "[FEE :: CLIENT :: project] Account Management services"

**Expected**:
- `determineSourceType()` → `PayoutSourceTypeServiceFee`
- Appears in: **Fee** section

**Status**: ✅ New behavior (fix target)

---

#### Test 4: Commission from InvoiceSplit → Extra Payment Section
**Input**:
- `TaskOrderID`: ""
- `InvoiceSplitID`: "invoice-split-999"
- `Description`: "[BONUS :: PROJECT] Performance bonus Q4 2024"

**Expected**:
- `determineSourceType()` → `PayoutSourceTypeCommission`
- Appears in: **Extra Payment** section

**Status**: ✅ New behavior (moved from Fee)

---

#### Test 5: ExtraPayment Source Type → Extra Payment Section
**Input**:
- `TaskOrderID`: ""
- `InvoiceSplitID`: ""
- `RefundRequestID`: ""
- `Description`: "Special bonus payment"

**Expected**:
- `determineSourceType()` → `PayoutSourceTypeExtraPayment`
- Appears in: **Extra Payment** section

**Status**: ✓ Existing behavior (no change)

---

#### Test 6: Refund → Refund Section
**Input**:
- `TaskOrderID`: ""
- `InvoiceSplitID`: ""
- `RefundRequestID`: "refund-123"
- `Description`: "Refund for overpayment"

**Expected**:
- `determineSourceType()` → `PayoutSourceTypeRefund`
- Appears in: **Refund** section

**Status**: ✓ Existing behavior (no change)

---

### Integration Test Cases

#### Integration Test 1: Mixed Payout Types Invoice
**Setup**:
- Contractor with 6 payout entries:
  1. Service Fee from TaskOrder
  2. Service Fee from InvoiceSplit with "Delivery Lead"
  3. Service Fee from InvoiceSplit with "Account Management"
  4. Commission from InvoiceSplit
  5. ExtraPayment source type
  6. Refund

**Expected Sections**:
```
Development Work Section:
  - Item 1 (Service Fee from TaskOrder)

Fee Section:
  - Item 2 (Service Fee with "Delivery Lead")
  - Item 3 (Service Fee with "Account Management")

Extra Payment Section:
  - Item 4 (Commission)
  - Item 5 (ExtraPayment)

Refund Section:
  - Item 6 (Refund)
```

---

#### Integration Test 2: Case-Insensitive Keyword Matching
**Setup**:
- Three InvoiceSplit items with variations:
  1. Description: "DELIVERY LEAD - Project A" (uppercase)
  2. Description: "Delivery lead - Project B" (mixed case)
  3. Description: "delivery lead - Project C" (lowercase)

**Expected**:
- All three items classified as ServiceFee
- All three appear in **Fee** section

---

#### Integration Test 3: Edge Cases
**Setup**:
- InvoiceSplit items with edge cases:
  1. Description: "Account Management" (exact match)
  2. Description: "Senior Account Management Lead" (keyword in middle)
  3. Description: "Delivery" (partial keyword - should NOT match)
  4. Description: "Lead Developer" (contains "Lead" but not "Delivery Lead")

**Expected**:
- Item 1: ServiceFee → Fee section ✅
- Item 2: ServiceFee → Fee section ✅
- Item 3: Commission → Extra Payment section ✅
- Item 4: Commission → Extra Payment section ✅

---

### Manual Verification Steps

1. **Generate Test Invoice**:
   ```bash
   # Use actual contractor with mixed payout types
   curl -X POST http://localhost:8080/api/v1/invoices/contractor/{contractor_id}/generate \
     -H "Authorization: Bearer {token}" \
     -d '{"month": 1, "year": 2026}'
   ```

2. **Review Generated PDF**:
   - ✅ Fee section contains only Service Fee items from InvoiceSplit
   - ✅ Extra Payment section contains Commission + ExtraPayment items
   - ✅ No Commission items in Fee section
   - ✅ Development Work section unchanged

3. **Check Debug Logs**:
   ```bash
   # Filter logs for classification decisions
   grep "determineSourceType" logs/fortress-api.log
   grep "created Fee section" logs/fortress-api.log
   grep "created Extra Payment section" logs/fortress-api.log
   ```

4. **Verify Notion Data**:
   - Check original Notion entries for rows 300-301
   - Confirm Description contains "Delivery Lead" or "Account Management"
   - Verify they now appear in Fee section (not Extra Payment)

---

### Regression Testing

**Ensure no impact to**:
- ✅ Development Work section items (Service Fee from TaskOrder)
- ✅ Refund section items
- ✅ Invoice totals and calculations
- ✅ Other invoice sections (Summary, Banking, etc.)
- ✅ Non-invoice payout flows (payroll calculation, webhook handling)

---

## Additional Considerations

### 1. GroupFeeByProject Option

**Current Location**: `pkg/controller/invoice/contractor_invoice.go:436-491`

**Current Function**: Groups Commission items by project before displaying in Fee section.

**Impact of Changes**:
- Fee section will now contain ServiceFee items (not Commission)
- GroupFeeByProject logic currently operates on Commission items
- These functions become mismatched

**Options**:

#### Option A: Remove Grouping Logic (Recommended)
- Delete lines 436-491 completely
- Simplifies code
- Assumes Fee section items (ServiceFee) don't need grouping
- **Pros**: Cleaner, less complexity
- **Cons**: If grouping is needed later, requires reimplementation

#### Option B: Adapt Grouping for ServiceFee Items
- Update `GroupFeeByProject()` to group ServiceFee items instead
- Rename function to `GroupServiceFeeByProject()`
- Apply grouping to Fee section items
- **Pros**: Maintains grouping capability
- **Cons**: More complex, requires testing grouping logic

#### Option C: Keep But Disable
- Set default to `false` for GroupFeeByProject option
- Make function a no-op
- Keep code for potential future use
- **Pros**: Easy rollback
- **Cons**: Dead code in codebase

**Recommendation**: **Option A** - Remove grouping unless there's a business requirement to group Service Fee items by project.

**Decision Required**: Confirm with stakeholders if Service Fee items in Fee section need to be grouped by project.

---

### 2. Keyword Maintenance

**Current Keywords**:
- "delivery lead"
- "account management"

**Considerations**:
- **Hardcoded in function**: Keywords are in `determineSourceType()`
- **Case sensitivity**: Currently case-insensitive (good)
- **Exact match**: Uses `strings.Contains()` (allows flexibility)

**Potential Issues**:
- ❌ New role types require code changes
- ❌ Typos in Notion descriptions cause misclassification
- ❌ No validation of Description format

**Improvement Options**:

#### Option A: Configuration-Based Keywords
```go
// In config file or environment variable
SERVICE_FEE_KEYWORDS=delivery lead,account management,project lead
```
**Pros**: Easy to add new keywords without code changes
**Cons**: Requires config management, restart to apply changes

#### Option B: Database Table for Keywords
```sql
CREATE TABLE payout_classification_keywords (
  id uuid PRIMARY KEY,
  keyword text NOT NULL UNIQUE,
  classification text NOT NULL,
  created_at timestamp DEFAULT now()
);
```
**Pros**: Dynamic management via admin UI
**Cons**: Additional database calls, more complexity

#### Option C: Keep Hardcoded (Current Approach)
**Pros**: Simple, fast, no dependencies
**Cons**: Requires code changes for new keywords

**Recommendation**: **Option C** (Keep hardcoded) for now. Revisit if keyword changes become frequent.

---

### 3. Notion Formula Dependency

**Current State**:
- Notion has a Type formula field
- Code replicates formula logic

**Risk**: If Notion formula changes but code doesn't update, classifications diverge.

**Mitigation Strategies**:

#### Strategy A: Document Sync Requirements
- Add comment in code linking to Notion formula
- Document keyword list in both places
- Include in PR review checklist

#### Strategy B: Automated Tests with Notion API
- Periodic integration tests that fetch from Notion
- Compare Notion formula result with code classification
- Alert on mismatches

#### Strategy C: Extract Formula from Notion (Future Enhancement)
- Fetch formula definition via Notion API
- Parse and execute formula in code
- **Pros**: Single source of truth
- **Cons**: Complex parsing, formula language differences

**Recommendation**: **Strategy A** (documentation) with future consideration of Strategy B for regression detection.

---

### 4. Performance Considerations

**Impact Analysis**:

| Operation | Before | After | Impact |
|-----------|--------|-------|--------|
| `determineSourceType()` | 4 checks | 4 checks + 2 string operations | ⚠️ Minimal |
| Fee section grouping | Loop all items | Loop all items | ✅ None |
| Extra Payment grouping | Loop all items | Loop all items | ✅ None |

**String Operations Added**:
```go
desc := strings.ToLower(entry.Description)  // O(n) where n = Description length
strings.Contains(desc, "delivery lead")     // O(n*m) where m = keyword length
strings.Contains(desc, "account management") // O(n*m)
```

**Performance Impact**: Negligible
- Typical Description length: < 200 characters
- Keyword length: < 20 characters
- Operation count: Same number of items processed
- Total added time: < 1ms per 1000 items

**Optimization**: Not needed unless processing 100,000+ items per request.

---

### 5. Backwards Compatibility

**Database**:
- ✅ No schema changes
- ✅ Existing data compatible

**API**:
- ✅ No endpoint changes
- ✅ No request/response format changes
- ✅ Same invoice generation parameters

**Invoice Format**:
- ⚠️ Section grouping changes (expected behavior change)
- ⚠️ Commission items move from Fee to Extra Payment
- ⚠️ Service Fee items appear in Fee section

**Migration Concerns**:
- **Historical invoices**: Already generated PDFs unchanged
- **New invoice generation**: Will use new logic
- **User expectation**: May need communication about section changes

**Rollback Plan**:
- Git revert commits
- No database migrations to reverse
- Previous invoices stored as PDFs (unchanged)

---

### 6. Error Handling

**Current Error Scenarios**:

#### Scenario A: Empty Description
**Input**: `InvoiceSplitID` present, `Description` is empty string

**Current Behavior**:
```go
desc := strings.ToLower("") // Returns ""
strings.Contains("", "delivery lead") // Returns false
// Result: Classified as Commission
```

**Expected**: Classified as Commission (no keywords found)
**Status**: ✅ Correct behavior

#### Scenario B: Nil Description (Unlikely)
**Current Code**: No nil check before `strings.ToLower()`

**Risk**: Potential panic if Description is nil

**Mitigation**:
```go
if entry.InvoiceSplitID != "" {
	desc := strings.ToLower(entry.Description)
	// ... rest of logic
}
```

**Status**: ⚠️ Add nil/empty check for safety

#### Scenario C: Malformed Description
**Input**: Description with special characters, newlines, etc.

**Current Behavior**: `strings.Contains()` handles all string content

**Status**: ✅ No issue

**Recommendation**: Add defensive check:
```go
if entry.InvoiceSplitID != "" && entry.Description != "" {
	desc := strings.ToLower(entry.Description)
	// ... keyword checks
}
```

---

### 7. Debug Logging Strategy

**Add logging at key decision points**:

```go
// In determineSourceType()
s.logger.Debug(fmt.Sprintf(
	"[PAYOUT_CLASSIFICATION] pageID=%s relations=[TaskOrder:%s InvoiceSplit:%s Refund:%s] desc=%q → type=%s",
	entry.PageID,
	entry.TaskOrderID,
	entry.InvoiceSplitID,
	entry.RefundRequestID,
	entry.Description,
	sourceType,
))

// In Fee section grouping
l.Debug(fmt.Sprintf(
	"[INVOICE_SECTION_FEE] contractor=%s items_processed=%d service_fee_items=%d",
	contractorID,
	len(items),
	len(feeItems),
))

// In Extra Payment section grouping
l.Debug(fmt.Sprintf(
	"[INVOICE_SECTION_EXTRA] contractor=%s commission_items=%d extra_payment_items=%d total=%d",
	contractorID,
	commissionCount,
	extraPaymentCount,
	len(extraPaymentItems),
))
```

**Benefits**:
- ✅ Trace classification decisions
- ✅ Verify keyword matching
- ✅ Audit section grouping
- ✅ Troubleshoot misclassification issues

---

### 8. Documentation Updates

**Files to Update**:

1. **Code Comments**:
   - `pkg/service/notion/contractor_payouts.go` - Update `determineSourceType()` comment
   - `pkg/controller/invoice/contractor_invoice.go` - Update section grouping comments

2. **API Documentation**:
   - Update Swagger comments if invoice structure described
   - Document section meanings in API docs

3. **Architecture Decision Record** (if applicable):
   - Create ADR documenting this classification logic change
   - Explain rationale for keyword-based classification

4. **User-Facing Documentation** (if exists):
   - Update contractor invoice documentation
   - Explain what appears in each section
   - Clarify Service Fee vs Commission types

---

## References

### Code Locations

| File | Function/Section | Lines | Purpose |
|------|-----------------|-------|---------|
| `pkg/service/notion/contractor_payouts.go` | `determineSourceType()` | 365-380 | Source type classification |
| `pkg/controller/invoice/contractor_invoice.go` | Fee section grouping | 977-994 | Groups Commission items |
| `pkg/controller/invoice/contractor_invoice.go` | Extra Payment section | 996-1012 | Groups ExtraPayment items |
| `pkg/controller/invoice/contractor_invoice.go` | `GroupFeeByProject()` | 436-491 | Commission grouping logic |

### Related Notion Properties

| Property Name | Field Type | Purpose |
|---------------|-----------|---------|
| `00 Task Order` | Relation | Links to Task Order (Development Work) |
| `02 Invoice Split` | Relation | Links to Invoice Split (Commissions/Fees) |
| `01 Refund` | Relation | Links to Refund Request |
| Type | Formula | Auto-calculated based on relations and Description |
| Description | Text | Contains role/purpose of payout |

### Payout Source Type Enum

**File**: `pkg/model/notion.go`

```go
type PayoutSourceType string

const (
	PayoutSourceTypeServiceFee    PayoutSourceType = "ServiceFee"
	PayoutSourceTypeCommission    PayoutSourceType = "Commission"
	PayoutSourceTypeRefund        PayoutSourceType = "Refund"
	PayoutSourceTypeExtraPayment  PayoutSourceType = "ExtraPayment"
)
```

### Invoice Section Structure

**File**: `pkg/controller/invoice/contractor_invoice.go`

```go
type ContractorInvoiceSection struct {
	Name         string                       // "Development Work", "Fee", "Extra Payment", "Refund"
	IsAggregated bool                         // Whether items are grouped/summarized
	Items        []ContractorInvoiceLineItem // Individual line items
}
```

---

## Next Steps

### Pre-Implementation

- [ ] **Review with stakeholders**: Confirm section grouping changes align with business requirements
- [ ] **Decide on GroupFeeByProject**: Remove, adapt, or keep disabled?
- [ ] **Verify keywords**: Confirm "Delivery Lead" and "Account Management" are complete list
- [ ] **Test data preparation**: Identify contractors with mixed payout types for testing

### Implementation

- [ ] **Implement Change 1**: Update `determineSourceType()` in `contractor_payouts.go`
- [ ] **Implement Change 2**: Update Fee section grouping in `contractor_invoice.go`
- [ ] **Implement Change 3**: Update Extra Payment section grouping in `contractor_invoice.go`
- [ ] **Add debug logging**: Implement classification and grouping debug logs
- [ ] **Add error handling**: Add nil/empty Description checks

### Testing

- [ ] **Write unit tests**: Test cases 1-6 in Testing Strategy section
- [ ] **Integration tests**: Test mixed payout invoice generation
- [ ] **Edge case tests**: Case sensitivity, partial keywords, etc.
- [ ] **Manual verification**: Generate actual invoice PDF and verify sections

### Documentation

- [ ] **Update code comments**: Document keyword-based classification logic
- [ ] **Update ADR**: Create decision record for classification change (if needed)
- [ ] **Update user docs**: Explain invoice sections (if user-facing docs exist)

### Deployment

- [ ] **Code review**: Review with @huynguyenh @lmquang per CODEOWNERS
- [ ] **Staging deployment**: Test with real Notion data
- [ ] **Monitor logs**: Check classification decisions with debug logs
- [ ] **Production deployment**: Roll out with invoice generation monitoring

### Post-Deployment

- [ ] **Monitor for misclassifications**: Watch for unexpected items in wrong sections
- [ ] **User feedback**: Collect feedback on new section groupings
- [ ] **Performance monitoring**: Verify no performance impact
- [ ] **Documentation sync**: Ensure Notion formula matches code logic

---

## Approval

**Documented by**: Claude Code
**Date**: 2026-01-21
**Status**: Awaiting review and approval

**Reviewers**:
- [ ] Technical Lead - Code approach and architecture
- [ ] Product Owner - Business requirements and section grouping
- [ ] QA Lead - Testing strategy

**Approved by**:
- [ ] ___________________ (Name, Date)

---

## Revision History

| Date | Version | Changes | Author |
|------|---------|---------|--------|
| 2026-01-21 | 1.0 | Initial investigation and solution design | Claude Code |

---

**End of Document**
