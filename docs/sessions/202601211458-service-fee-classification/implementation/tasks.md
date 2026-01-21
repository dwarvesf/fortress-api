# Implementation Tasks: Service Fee Classification Fix

**Session**: 202601211458-service-fee-classification
**Date Created**: 2026-01-21
**Status**: Ready for Implementation
**Priority**: Medium

## Overview

This document breaks down the implementation of the Service Fee classification fix into sequential, actionable tasks. The fix ensures that items from Invoice Split with "Delivery Lead" or "Account Management" in their Description are properly classified as Service Fee and grouped in the "Fee" section, while other Invoice Split items remain as Commission and appear in the "Extra Payment" section.

**Reference**: See `/Users/quang/workspace/dwarvesf/fortress-api/docs/issues/service-fee-invoice-split-classification.md` for detailed investigation and solution design.

---

## Implementation Tasks

### Task 1: Update Source Type Determination Logic

**File(s)**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payouts.go`

**Description**: Modify the `determineSourceType()` function to check Description content for items with InvoiceSplitID and return the appropriate source type based on keyword matching.

**Implementation Notes**:
- **Location**: Lines 365-380 (function `determineSourceType`)
- **Current behavior**: Returns `PayoutSourceTypeCommission` for ALL items with `InvoiceSplitID`
- **New behavior**:
  - Check if Description contains "delivery lead" OR "account management" (case-insensitive)
  - Return `PayoutSourceTypeServiceFee` if keywords found
  - Return `PayoutSourceTypeCommission` otherwise
- **Required changes**:
  1. Add Description content check in the InvoiceSplitID condition block
  2. Use `strings.ToLower()` for case-insensitive matching
  3. Use `strings.Contains()` for keyword detection
  4. Ensure `strings` package is imported at top of file
  5. Add nil/empty Description safety check

**Code Changes**:
```go
// Current code (lines 370-372):
if entry.InvoiceSplitID != "" {
    s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: sourceType=Commission (InvoiceSplitID=%s)", entry.InvoiceSplitID))
    return PayoutSourceTypeCommission
}

// New code:
// For InvoiceSplit items: check Description to determine type
if entry.InvoiceSplitID != "" {
    // Add nil/empty check for safety
    if entry.Description != "" {
        desc := strings.ToLower(entry.Description)

        // Service Fee: Delivery Lead or Account Management roles
        // These keywords match the Notion formula logic
        if strings.Contains(desc, "delivery lead") || strings.Contains(desc, "account management") {
            s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: sourceType=ServiceFee (InvoiceSplitID=%s, keywords found in Description)", entry.InvoiceSplitID))
            return PayoutSourceTypeServiceFee
        }
    }

    // Otherwise: Commission
    s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: sourceType=Commission (InvoiceSplitID=%s, no keywords)", entry.InvoiceSplitID))
    return PayoutSourceTypeCommission
}
```

**Acceptance Criteria**:
- ✅ Items with `InvoiceSplitID` AND "Delivery Lead" in Description → `PayoutSourceTypeServiceFee`
- ✅ Items with `InvoiceSplitID` AND "Account Management" in Description → `PayoutSourceTypeServiceFee`
- ✅ Items with `InvoiceSplitID` without keywords → `PayoutSourceTypeCommission`
- ✅ Case-insensitive matching works ("DELIVERY LEAD", "delivery lead", etc.)
- ✅ Empty/nil Description doesn't cause errors
- ✅ Debug logging shows classification decision with Description content
- ✅ Existing behavior for TaskOrderID, RefundRequestID unchanged

**Testing Commands**:
```bash
# Run unit tests for contractor_payouts
go test -v ./pkg/service/notion -run TestDetermineSourceType

# Manual verification with debug logs
grep "determineSourceType" logs/fortress-api.log | grep "InvoiceSplitID"
```

---

### Task 2: Update Fee Section Grouping Logic

**File(s)**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go`

**Description**: Change the Fee section grouping from filtering Commission items to filtering Service Fee items that originate from InvoiceSplit (not TaskOrder).

**Implementation Notes**:
- **Location**: Lines 977-994 (Fee section grouping)
- **Current behavior**: Groups items where `Type == PayoutSourceTypeCommission`
- **New behavior**: Group items where `Type == PayoutSourceTypeServiceFee` AND item is from InvoiceSplit (not TaskOrder)
- **Required changes**:
  1. Change condition from `item.Type == string(notion.PayoutSourceTypeCommission)` to `item.Type == string(notion.PayoutSourceTypeServiceFee)`
  2. Add filter to exclude TaskOrder items (they go to Development Work section)
  3. Use `TaskOrderID == ""` and `ServiceRateID == ""` to identify InvoiceSplit items
  4. Remove "Bonus" → "Fee" replacement (move to Task 3)
  5. Update debug logging to reflect new logic

**Code Changes**:
```go
// Current code (lines 977-994):
// Group Bonus (Commission) items - individual display
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

// New code:
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

**Acceptance Criteria**:
- ✅ Fee section contains ONLY Service Fee items from InvoiceSplit
- ✅ Fee section DOES NOT contain Commission items
- ✅ Fee section DOES NOT contain Service Fee items from TaskOrder
- ✅ Debug logging shows correct count of Service Fee items
- ✅ Section is not created if no qualifying items exist

**Testing Commands**:
```bash
# Run invoice controller tests
go test -v ./pkg/controller/invoice -run TestGroupIntoSections

# Manual verification with debug logs
grep "created Fee section" logs/fortress-api.log
```

---

### Task 3: Update Extra Payment Section Grouping Logic

**File(s)**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go`

**Description**: Update the Extra Payment section to include both Commission items (moved from Fee section) and ExtraPayment source type items, with "Bonus" → "Fee" description replacement.

**Implementation Notes**:
- **Location**: Lines 996-1012 (Extra Payment section grouping)
- **Current behavior**: Groups items where `Type == PayoutSourceTypeExtraPayment` only
- **New behavior**: Group items where `Type == PayoutSourceTypeCommission` OR `Type == PayoutSourceTypeExtraPayment`
- **Required changes**:
  1. Add Commission type check to condition
  2. Keep existing ExtraPayment type check
  3. Move "Bonus" → "Fee" replacement from Fee section here (applies to both types)
  4. Update debug logging to show combined count and breakdown

**Code Changes**:
```go
// Current code (lines 996-1012):
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

// New code:
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

**Acceptance Criteria**:
- ✅ Extra Payment section contains Commission items
- ✅ Extra Payment section contains ExtraPayment source type items
- ✅ "Bonus" is replaced with "Fee" in all item descriptions
- ✅ Debug logging shows correct total count
- ✅ Section is not created if no qualifying items exist

**Testing Commands**:
```bash
# Run invoice controller tests
go test -v ./pkg/controller/invoice -run TestGroupIntoSections

# Manual verification with debug logs
grep "created Extra Payment section" logs/fortress-api.log
```

---

### Task 4: Add Enhanced Debug Logging for Classification

**File(s)**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payouts.go`

**Description**: Add comprehensive debug logging to trace classification decisions and help troubleshoot misclassification issues.

**Implementation Notes**:
- **Location**: Within `determineSourceType()` function (after Task 1 changes)
- **Logging should include**:
  1. PageID for tracking specific entries
  2. All relation IDs (TaskOrder, InvoiceSplit, Refund)
  3. Description content (first 100 chars to avoid log bloat)
  4. Keywords found (if applicable)
  5. Resulting source type
- **Use existing logger**: `s.logger.Debug()`

**Code Changes**:
```go
// Add at the end of determineSourceType() function, before returning:
func (s *ContractorPayoutsService) determineSourceType(entry PayoutEntry) PayoutSourceType {
    // ... existing logic with changes from Task 1 ...

    // Enhanced debug logging at function end
    sourceType := // determined type based on logic above

    // Truncate Description for logging (avoid log bloat)
    descPreview := entry.Description
    if len(descPreview) > 100 {
        descPreview = descPreview[:100] + "..."
    }

    s.logger.Debug(fmt.Sprintf(
        "[CLASSIFICATION] pageID=%s | Relations: TaskOrder=%s InvoiceSplit=%s Refund=%s | Description=%q | Result=%s",
        entry.PageID,
        entry.TaskOrderID,
        entry.InvoiceSplitID,
        entry.RefundRequestID,
        descPreview,
        sourceType,
    ))

    return sourceType
}
```

**Additional Logging in Invoice Controller**:
Add section-level debug logs to show item distribution:

```go
// After all sections are created (around line 1014):
l.Debug(fmt.Sprintf(
    "[INVOICE_SECTIONS] contractor=%s | Total items=%d | Sections: Development=%d Fee=%d ExtraPayment=%d Refund=%d",
    contractorID,
    totalItemCount,
    devWorkCount,
    feeCount,
    extraPaymentCount,
    refundCount,
))
```

**Acceptance Criteria**:
- ✅ Every payout entry classification is logged with all relevant data
- ✅ Logs include PageID for cross-referencing with Notion
- ✅ Description content is truncated to prevent log bloat
- ✅ Section creation logs show item distribution
- ✅ Logs are at DEBUG level (not INFO/ERROR)
- ✅ Log format is consistent and parseable

**Testing Commands**:
```bash
# Generate invoice and check logs
tail -f logs/fortress-api.log | grep "\[CLASSIFICATION\]"
tail -f logs/fortress-api.log | grep "\[INVOICE_SECTIONS\]"

# Filter logs for specific contractor
grep "contractor=<contractor-id>" logs/fortress-api.log | grep CLASSIFICATION
```

---

### Task 5: Handle GroupFeeByProject Option (Decision Required)

**File(s)**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go`

**Description**: Decide on the fate of the `GroupFeeByProject` functionality (lines 436-491) which currently groups Commission items by project, but will become misaligned after Fee section now contains Service Fee items.

**Implementation Notes**:
- **Location**: Lines 436-491 (`GroupFeeByProject` logic block)
- **Current behavior**: Groups Commission items by project name before adding to Fee section
- **Problem**: After Task 2, Fee section contains Service Fee items (not Commission), making this grouping logic incorrect
- **Options**:

  **Option A (RECOMMENDED): Remove Grouping Logic**
  - Delete lines 436-491 entirely
  - Remove `GroupFeeByProject` from options struct (if not used elsewhere)
  - Simplifies codebase
  - Assumes Service Fee items don't need project grouping

  **Option B: Adapt for Service Fee Items**
  - Rename to `GroupServiceFeeByProject`
  - Update logic to group Service Fee items instead
  - Extract project information from Service Fee descriptions
  - More complex, requires understanding Service Fee project structure

  **Option C: Keep But Disable**
  - Set default to `false`
  - Add comment explaining it's deprecated
  - Keep code for potential rollback

**Code Changes (Option A - Recommended)**:
```go
// DELETE lines 436-491:
// 4.6 Group Commission items by Project (if enabled)
if opts.GroupFeeByProject {
    // ... entire block ...
}

// Verify GroupFeeByProject is not used elsewhere:
# grep -r "GroupFeeByProject" pkg/
```

**Decision Required**: Stakeholder input needed on whether Service Fee items should be grouped by project. Until decision is made, recommend **Option C** as interim solution.

**Acceptance Criteria (Option A)**:
- ✅ `GroupFeeByProject` code block removed
- ✅ No references to `GroupFeeByProject` in codebase (or marked deprecated)
- ✅ Tests updated to remove grouping expectations
- ✅ Invoice generation works without grouping logic

**Acceptance Criteria (Option C - Interim)**:
- ✅ `GroupFeeByProject` defaults to `false`
- ✅ Code block has deprecation comment
- ✅ Function still works if explicitly enabled (for rollback)

**Testing Commands**:
```bash
# Check for all references to GroupFeeByProject
grep -r "GroupFeeByProject" pkg/ docs/

# Generate invoice with and without option
# (ensure both work correctly)
```

---

### Task 6: Write Unit Tests for determineSourceType

**File(s)**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payouts_test.go` (create if doesn't exist)

**Description**: Create comprehensive unit tests for the updated `determineSourceType()` function to verify all classification scenarios.

**Implementation Notes**:
- **Test file location**: Same directory as `contractor_payouts.go`
- **Test function name**: `TestDetermineSourceType`
- **Test approach**: Table-driven tests with multiple scenarios
- **Mock requirements**: May need mock logger if not already available

**Test Cases**:
1. **TaskOrder present** → ServiceFee (existing behavior)
2. **InvoiceSplit + "Delivery Lead" in Description** → ServiceFee (new)
3. **InvoiceSplit + "Account Management" in Description** → ServiceFee (new)
4. **InvoiceSplit + "DELIVERY LEAD" (uppercase)** → ServiceFee (case-insensitive)
5. **InvoiceSplit + "delivery lead" (lowercase)** → ServiceFee (case-insensitive)
6. **InvoiceSplit + "Account management" (mixed case)** → ServiceFee (case-insensitive)
7. **InvoiceSplit + no keywords** → Commission (new)
8. **InvoiceSplit + "Delivery" (partial keyword)** → Commission (should NOT match)
9. **InvoiceSplit + "Lead Developer" (partial keyword)** → Commission (should NOT match)
10. **InvoiceSplit + empty Description** → Commission (safe handling)
11. **RefundRequestID present** → Refund (existing behavior)
12. **No relations** → ExtraPayment (existing behavior)

**Code Template**:
```go
package notion

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestDetermineSourceType(t *testing.T) {
    // Create service with mock logger
    service := &ContractorPayoutsService{
        logger: newMockLogger(), // implement or use existing mock
    }

    tests := []struct {
        name        string
        entry       PayoutEntry
        expected    PayoutSourceType
        description string
    }{
        {
            name: "TaskOrder present - returns ServiceFee",
            entry: PayoutEntry{
                TaskOrderID:    "task-order-123",
                InvoiceSplitID: "",
                Description:    "Development work",
            },
            expected:    PayoutSourceTypeServiceFee,
            description: "TaskOrder has highest priority",
        },
        {
            name: "InvoiceSplit with 'Delivery Lead' - returns ServiceFee",
            entry: PayoutEntry{
                TaskOrderID:    "",
                InvoiceSplitID: "invoice-split-456",
                Description:    "[FEE :: MUDAH :: ics3rd] Delivery Lead - Project Y",
            },
            expected:    PayoutSourceTypeServiceFee,
            description: "Keywords trigger ServiceFee classification",
        },
        {
            name: "InvoiceSplit with 'Account Management' - returns ServiceFee",
            entry: PayoutEntry{
                TaskOrderID:    "",
                InvoiceSplitID: "invoice-split-789",
                Description:    "Account Management services for client",
            },
            expected:    PayoutSourceTypeServiceFee,
            description: "Keywords trigger ServiceFee classification",
        },
        {
            name: "InvoiceSplit with uppercase 'DELIVERY LEAD' - returns ServiceFee",
            entry: PayoutEntry{
                TaskOrderID:    "",
                InvoiceSplitID: "invoice-split-101",
                Description:    "DELIVERY LEAD for Project X",
            },
            expected:    PayoutSourceTypeServiceFee,
            description: "Case-insensitive matching",
        },
        {
            name: "InvoiceSplit without keywords - returns Commission",
            entry: PayoutEntry{
                TaskOrderID:    "",
                InvoiceSplitID: "invoice-split-999",
                Description:    "[BONUS :: PROJECT] Performance bonus Q4",
            },
            expected:    PayoutSourceTypeCommission,
            description: "No keywords = Commission",
        },
        {
            name: "InvoiceSplit with partial keyword 'Delivery' - returns Commission",
            entry: PayoutEntry{
                TaskOrderID:    "",
                InvoiceSplitID: "invoice-split-111",
                Description:    "Delivery of project milestone",
            },
            expected:    PayoutSourceTypeCommission,
            description: "Partial keyword should NOT match",
        },
        {
            name: "InvoiceSplit with empty Description - returns Commission",
            entry: PayoutEntry{
                TaskOrderID:    "",
                InvoiceSplitID: "invoice-split-222",
                Description:    "",
            },
            expected:    PayoutSourceTypeCommission,
            description: "Empty Description handled safely",
        },
        {
            name: "RefundRequestID present - returns Refund",
            entry: PayoutEntry{
                TaskOrderID:      "",
                InvoiceSplitID:   "",
                RefundRequestID:  "refund-123",
                Description:      "Refund for overpayment",
            },
            expected:    PayoutSourceTypeRefund,
            description: "Refund relation takes priority",
        },
        {
            name: "No relations - returns ExtraPayment",
            entry: PayoutEntry{
                TaskOrderID:     "",
                InvoiceSplitID:  "",
                RefundRequestID: "",
                Description:     "Special bonus payment",
            },
            expected:    PayoutSourceTypeExtraPayment,
            description: "Default when no relations",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := service.determineSourceType(tt.entry)
            assert.Equal(t, tt.expected, result, tt.description)
        })
    }
}
```

**Acceptance Criteria**:
- ✅ All 12 test cases pass
- ✅ Tests cover all classification paths
- ✅ Tests verify case-insensitive matching
- ✅ Tests verify partial keyword rejection
- ✅ Tests verify empty/nil Description handling
- ✅ Tests follow Go testing conventions
- ✅ Test output is clear and descriptive

**Testing Commands**:
```bash
# Run the new unit tests
go test -v ./pkg/service/notion -run TestDetermineSourceType

# Run with coverage
go test -cover ./pkg/service/notion -run TestDetermineSourceType
```

---

### Task 7: Write Integration Tests for Invoice Section Grouping

**File(s)**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice_test.go`

**Description**: Create integration tests that verify the complete flow from payout classification to invoice section grouping, ensuring items appear in the correct sections.

**Implementation Notes**:
- **Test file location**: Same directory as `contractor_invoice.go`
- **Test function name**: `TestGroupIntoSections_MixedPayoutTypes`
- **Test approach**: Create mock line items with different types and verify section grouping
- **Dependencies**: May need test fixtures or mock data

**Test Scenarios**:
1. **Mixed payout types invoice** - All section types present
2. **Service Fee from TaskOrder** - Appears in Development Work
3. **Service Fee from InvoiceSplit** - Appears in Fee section
4. **Commission items** - Appear in Extra Payment
5. **ExtraPayment items** - Appear in Extra Payment
6. **Refund items** - Appear in Refund section
7. **Empty sections** - Not created when no items

**Code Template**:
```go
func TestGroupIntoSections_MixedPayoutTypes(t *testing.T) {
    tests := []struct {
        name            string
        items           []ContractorInvoiceLineItem
        expectedSections []string // Section names expected
        sectionItemCounts map[string]int
    }{
        {
            name: "Mixed payout types - all sections present",
            items: []ContractorInvoiceLineItem{
                // Service Fee from TaskOrder
                {
                    Type:          string(notion.PayoutSourceTypeServiceFee),
                    TaskOrderID:   "task-order-123",
                    ServiceRateID: "rate-456",
                    Description:   "Development work",
                    AmountUSD:     1000.0,
                },
                // Service Fee from InvoiceSplit (Delivery Lead)
                {
                    Type:           string(notion.PayoutSourceTypeServiceFee),
                    TaskOrderID:    "",
                    ServiceRateID:  "",
                    Description:    "Delivery Lead services",
                    AmountUSD:      500.0,
                },
                // Commission (no keywords)
                {
                    Type:        string(notion.PayoutSourceTypeCommission),
                    Description: "Performance bonus",
                    AmountUSD:   200.0,
                },
                // ExtraPayment
                {
                    Type:        string(notion.PayoutSourceTypeExtraPayment),
                    Description: "Special bonus",
                    AmountUSD:   150.0,
                },
                // Refund
                {
                    Type:        string(notion.PayoutSourceTypeRefund),
                    Description: "Refund for error",
                    AmountUSD:   -100.0,
                },
            },
            expectedSections: []string{
                "Development Work",
                "Fee",
                "Extra Payment",
                "Expense Reimbursement",
            },
            sectionItemCounts: map[string]int{
                "Development Work":       1,
                "Fee":                    1,
                "Extra Payment":          2, // Commission + ExtraPayment
                "Expense Reimbursement":  1,
            },
        },
        {
            name: "Only Service Fee from InvoiceSplit",
            items: []ContractorInvoiceLineItem{
                {
                    Type:          string(notion.PayoutSourceTypeServiceFee),
                    TaskOrderID:   "",
                    ServiceRateID: "",
                    Description:   "Account Management services",
                    AmountUSD:     500.0,
                },
            },
            expectedSections: []string{"Fee"},
            sectionItemCounts: map[string]int{
                "Fee": 1,
            },
        },
        {
            name: "Commission and ExtraPayment mix",
            items: []ContractorInvoiceLineItem{
                {
                    Type:        string(notion.PayoutSourceTypeCommission),
                    Description: "Commission Q4",
                    AmountUSD:   300.0,
                },
                {
                    Type:        string(notion.PayoutSourceTypeExtraPayment),
                    Description: "Holiday bonus",
                    AmountUSD:   200.0,
                },
            },
            expectedSections: []string{"Extra Payment"},
            sectionItemCounts: map[string]int{
                "Extra Payment": 2,
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Mock logger
            logger := newMockLogger()

            // Call groupIntoSections
            sections := groupIntoSections(tt.items, logger)

            // Verify expected sections exist
            sectionNames := make([]string, len(sections))
            for i, section := range sections {
                sectionNames[i] = section.Name
            }

            assert.ElementsMatch(t, tt.expectedSections, sectionNames,
                "Section names should match expected")

            // Verify item counts per section
            for _, section := range sections {
                expectedCount, ok := tt.sectionItemCounts[section.Name]
                if ok {
                    assert.Equal(t, expectedCount, len(section.Items),
                        fmt.Sprintf("Section %s should have %d items", section.Name, expectedCount))
                }
            }
        })
    }
}
```

**Additional Test Cases**:
```go
func TestGroupIntoSections_ServiceFeeFiltering(t *testing.T) {
    // Test that verifies:
    // 1. Service Fee from TaskOrder goes to Development Work (not Fee)
    // 2. Service Fee from InvoiceSplit goes to Fee (not Development Work)
}

func TestGroupIntoSections_BonusReplacement(t *testing.T) {
    // Test that verifies "Bonus" is replaced with "Fee" in Extra Payment items
}
```

**Acceptance Criteria**:
- ✅ All integration tests pass
- ✅ Tests verify correct section grouping for each payout type
- ✅ Tests verify Service Fee items are correctly filtered by origin
- ✅ Tests verify Commission items appear in Extra Payment
- ✅ Tests verify "Bonus" → "Fee" replacement works
- ✅ Tests verify empty sections are not created
- ✅ Tests follow table-driven pattern

**Testing Commands**:
```bash
# Run integration tests
go test -v ./pkg/controller/invoice -run TestGroupIntoSections

# Run with coverage
go test -cover ./pkg/controller/invoice -run TestGroupIntoSections
```

---

### Task 8: Manual End-to-End Verification

**File(s)**: N/A (Manual testing)

**Description**: Perform manual testing with actual contractor data to verify the fix works end-to-end in a real environment.

**Prerequisites**:
- Local development environment running (`make dev`)
- Access to Notion API with contractor payout data
- Test contractor with mixed payout types (Service Fee, Commission, ExtraPayment)
- Valid JWT token for API authentication

**Test Steps**:

**Step 1: Identify Test Contractor**
```bash
# Find contractor with mixed payout types in current month
# Check Notion database or use API to list contractors
curl -X GET http://localhost:8080/api/v1/contractors \
  -H "Authorization: Bearer {token}"

# Select contractor with InvoiceSplit items containing "Delivery Lead" or "Account Management"
```

**Step 2: Generate Invoice**
```bash
# Generate contractor invoice for current month
curl -X POST http://localhost:8080/api/v1/invoices/contractor/{contractor_id}/generate \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "month": 1,
    "year": 2026
  }' \
  -o test-invoice.pdf

# Check response status (should be 200 OK)
```

**Step 3: Review Generated PDF**
- Open `test-invoice.pdf`
- Verify sections present:
  - **Development Work**: Contains Service Fee from TaskOrder (if any)
  - **Fee**: Contains ONLY Service Fee items from InvoiceSplit
    - Check for items with "Delivery Lead" or "Account Management"
  - **Extra Payment**: Contains Commission AND ExtraPayment items
    - Verify "Bonus" is replaced with "Fee" in descriptions
  - **Expense Reimbursement**: Contains Refund items (if any)

**Step 4: Check Debug Logs**
```bash
# View classification decisions
grep "\[CLASSIFICATION\]" logs/fortress-api.log | tail -20

# View section grouping
grep "\[INVOICE_SECTIONS\]" logs/fortress-api.log | tail -5

# Check for InvoiceSplit items with keywords
grep "InvoiceSplit=" logs/fortress-api.log | grep -i "delivery lead\|account management"
```

**Step 5: Verify Notion Data**
- Open Notion Contractor Payouts database
- Find the specific entries that appeared in the invoice
- Verify:
  - Items in Fee section have `02 Invoice Split` relation
  - Their Description contains "Delivery Lead" OR "Account Management"
  - Type formula shows "Service Fee" in Notion

**Step 6: Cross-Reference**
Create a comparison table:

| PageID (Notion) | Description (Notion) | Type (Notion) | Section (PDF) | Match? |
|-----------------|----------------------|---------------|---------------|--------|
| page-123 | "Delivery Lead..." | Service Fee | Fee | ✅ |
| page-456 | "Commission..." | Commission | Extra Payment | ✅ |

**Expected Results**:
- ✅ Invoice PDF generated successfully
- ✅ Fee section contains ONLY Service Fee items from InvoiceSplit
- ✅ No Commission items in Fee section
- ✅ Commission items appear in Extra Payment section
- ✅ ExtraPayment items appear in Extra Payment section
- ✅ Service Fee from TaskOrder appears in Development Work (not Fee)
- ✅ "Bonus" replaced with "Fee" in descriptions
- ✅ Debug logs show correct classifications
- ✅ All sections have correct item counts
- ✅ Invoice totals are correct

**Regression Checks**:
- ✅ Development Work section unchanged (Service Fee from TaskOrder)
- ✅ Refund section unchanged (if applicable)
- ✅ Invoice totals match sum of all items
- ✅ No items missing or duplicated
- ✅ Currency conversions correct (VND/USD)

**Edge Cases to Test**:
1. **Case sensitivity**: Create test with "DELIVERY LEAD", "delivery lead", "Delivery Lead"
2. **Partial keywords**: Create test with "Delivery" (no "Lead") - should be Commission
3. **Empty Description**: Create InvoiceSplit item with empty Description - should be Commission
4. **Multiple keywords**: Create item with both "Delivery Lead" AND "Account Management" - should be Service Fee

**Acceptance Criteria**:
- ✅ Manual test completed for at least 3 different contractors
- ✅ All expected results verified
- ✅ All regression checks passed
- ✅ Edge cases tested and behaving correctly
- ✅ Screenshots captured for documentation
- ✅ Test results documented in test report

**Documentation**:
Create a test report in `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601211458-service-fee-classification/implementation/manual-test-report.md` with:
- Contractor IDs tested
- Screenshots of PDF sections
- Debug log excerpts
- Notion data verification
- Issues found (if any)

---

### Task 9: Update Code Documentation and Comments

**File(s)**:
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payouts.go`
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go`

**Description**: Update code comments and documentation to reflect the new classification logic and section grouping behavior.

**Implementation Notes**:

**File 1: contractor_payouts.go**

Update function comment for `determineSourceType()`:
```go
// determineSourceType determines the source type based on which relation is set
// and Description content for InvoiceSplit items.
//
// Classification rules:
// 1. TaskOrderID present → ServiceFee (Development Work)
// 2. InvoiceSplitID present:
//    - Description contains "delivery lead" OR "account management" (case-insensitive) → ServiceFee
//    - Otherwise → Commission
// 3. RefundRequestID present → Refund
// 4. No relations → ExtraPayment
//
// Note: For InvoiceSplit items, this logic replicates the Notion Type formula
// to ensure consistency between Notion and the application.
// Keywords: "delivery lead", "account management" (must match Notion formula)
func (s *ContractorPayoutsService) determineSourceType(entry PayoutEntry) PayoutSourceType {
    // ... implementation ...
}
```

Add comment explaining keyword maintenance:
```go
// IMPORTANT: Keyword list must match Notion Type formula
// If Notion formula changes, update these keywords accordingly:
// - "delivery lead"
// - "account management"
// Case-insensitive matching is used to handle variations.
const (
    keywordDeliveryLead      = "delivery lead"
    keywordAccountManagement = "account management"
)
```

**File 2: contractor_invoice.go**

Update comment for Fee section grouping:
```go
// Fee section: Service Fee items from InvoiceSplit only
// These are items with "Delivery Lead" or "Account Management" roles
// determined by Description content during payout classification.
//
// Note: Service Fee items from TaskOrder go to "Development Work" section.
// This filtering ensures only InvoiceSplit-based fees appear here.
```

Update comment for Extra Payment section grouping:
```go
// Extra Payment section: Commission items + ExtraPayment items
// - Commission: InvoiceSplit items without special role keywords
// - ExtraPayment: Items with no task order, invoice split, or refund relation
//
// Note: "Bonus" is replaced with "Fee" in descriptions for consistency.
```

Update comment for Development Work section (if not already clear):
```go
// Development Work section: Service Fee items from TaskOrder
// Aggregated by hourly rate or displayed individually for fixed-price tasks.
// This section represents billable development work.
```

**File 3: Update Type Enum Documentation**

If PayoutSourceType enum has comments, update them:
```go
// PayoutSourceType represents the source/category of a contractor payout
type PayoutSourceType string

const (
    // PayoutSourceTypeServiceFee: Development work from TaskOrder OR
    // role-based fees from InvoiceSplit (Delivery Lead, Account Management)
    PayoutSourceTypeServiceFee PayoutSourceType = "ServiceFee"

    // PayoutSourceTypeCommission: Bonuses/commissions from InvoiceSplit
    // without special role keywords
    PayoutSourceTypeCommission PayoutSourceType = "Commission"

    // PayoutSourceTypeRefund: Expense reimbursements from RefundRequest
    PayoutSourceTypeRefund PayoutSourceType = "Refund"

    // PayoutSourceTypeExtraPayment: Ad-hoc payments without relation
    PayoutSourceTypeExtraPayment PayoutSourceType = "ExtraPayment"
)
```

**Acceptance Criteria**:
- ✅ Function comments clearly explain classification logic
- ✅ Keyword list is documented with maintenance instructions
- ✅ Section grouping comments are accurate and detailed
- ✅ Comments explain the relationship between Notion and code logic
- ✅ Enum documentation is updated
- ✅ Comments follow Go documentation conventions
- ✅ No outdated comments remain (e.g., old "Commission in Fee section" references)

**Testing Commands**:
```bash
# Generate godoc and review
go doc pkg/service/notion.ContractorPayoutsService.determineSourceType
go doc pkg/controller/invoice

# Check for outdated "Commission" references in comments
grep -n "Commission.*Fee section" pkg/controller/invoice/contractor_invoice.go
```

---

### Task 10: Create Architecture Decision Record (Optional)

**File(s)**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601211458-service-fee-classification/planning/ADRs/001-service-fee-classification-logic.md`

**Description**: Document the architectural decision to replicate Notion formula logic in code rather than extracting the formula value from Notion API.

**Template**:
```markdown
# ADR 001: Service Fee Classification Logic for InvoiceSplit Items

## Status

Accepted

## Context

Contractor payouts from InvoiceSplit can represent different types of payments:
1. Role-based fees (Delivery Lead, Account Management) - should be Service Fee
2. Bonuses and commissions - should be Commission

Previously, all InvoiceSplit items were classified as Commission, causing incorrect invoice section grouping.

Notion has a Type formula field that determines classification based on Description content, but the application was not using it.

### Options Considered

1. **Extract Type formula value from Notion API**
   - Pros: Single source of truth, automatic sync
   - Cons: Additional API call overhead, formula parsing complexity, API limitations

2. **Replicate formula logic in code** (SELECTED)
   - Pros: No API dependency, fast execution, full control
   - Cons: Manual sync required if formula changes

3. **Store classification in database**
   - Pros: Fast lookup, no computation
   - Cons: Data duplication, sync complexity, migration required

## Decision

We will replicate the Notion Type formula logic in the `determineSourceType()` function:
- Check Description content for keywords: "delivery lead", "account management"
- Use case-insensitive matching
- Return ServiceFee if keywords found, Commission otherwise

### Keywords

The following keywords trigger Service Fee classification:
- "delivery lead" (case-insensitive)
- "account management" (case-insensitive)

These MUST match the Notion formula logic.

## Consequences

### Positive

- Fast classification (no API calls)
- Full control over logic and edge cases
- Easy to debug and test
- No additional database schema changes

### Negative

- Keyword list is hardcoded in code
- Manual sync required if Notion formula changes
- Keyword additions require code deployment

### Mitigation

1. Document keyword list in code comments with reference to Notion formula
2. Add integration tests that can detect drift between code and Notion
3. Include keyword sync in PR review checklist
4. Consider configuration-based keywords in future if changes become frequent

## Related

- Investigation: `/docs/issues/service-fee-invoice-split-classification.md`
- Implementation: `/docs/sessions/202601211458-service-fee-classification/implementation/tasks.md`

## Notes

If keyword list becomes dynamic or changes frequently, consider moving to configuration file or database table. Current business requirements indicate these role types are stable.
```

**Acceptance Criteria**:
- ✅ ADR follows standard format (Status, Context, Decision, Consequences)
- ✅ All options considered are documented
- ✅ Decision rationale is clear
- ✅ Consequences (positive and negative) are listed
- ✅ Mitigation strategies are included
- ✅ Related documents are linked
- ✅ Status is set to "Accepted"

---

## Task Dependencies

```
Task 1 (determineSourceType)
    ↓
Task 4 (Debug Logging) ← Can be done in parallel with Task 1
    ↓
Task 2 (Fee Section)
    ↓
Task 3 (Extra Payment Section)
    ↓
Task 5 (GroupFeeByProject) ← Decision required, can block deployment
    ↓
Task 6 (Unit Tests) ← Can start after Task 1
Task 7 (Integration Tests) ← Can start after Tasks 2-3
    ↓
Task 8 (Manual Testing) ← Requires all code changes complete
    ↓
Task 9 (Documentation) ← Can be done alongside implementation
Task 10 (ADR) ← Can be done anytime, recommended before deployment
```

**Critical Path**: Tasks 1 → 2 → 3 → 6 → 7 → 8

**Can be parallelized**:
- Task 4 (logging) with Task 1
- Task 9 (documentation) during implementation
- Task 10 (ADR) at any time

---

## Pre-Implementation Checklist

Before starting implementation:
- [ ] Review investigation document thoroughly
- [ ] Confirm keyword list with stakeholders ("delivery lead", "account management")
- [ ] Decide on GroupFeeByProject fate (Task 5)
- [ ] Ensure development environment is set up (`make dev`)
- [ ] Verify access to Notion API for testing
- [ ] Identify test contractors with mixed payout types
- [ ] Review existing test patterns in codebase

---

## Post-Implementation Checklist

After completing all tasks:
- [ ] All unit tests pass (`make test`)
- [ ] All integration tests pass
- [ ] Manual end-to-end testing completed
- [ ] Debug logging verified in logs
- [ ] Code documentation updated
- [ ] ADR created (if applicable)
- [ ] PR created with detailed description
- [ ] Code review requested from @huynguyenh @lmquang
- [ ] Staging deployment tested with real data
- [ ] Production deployment planned
- [ ] Monitoring dashboard prepared for classification metrics

---

## Rollback Plan

If issues are discovered after deployment:

**Immediate Rollback** (within 1 hour):
```bash
# Revert commits
git revert <commit-hash>
git push origin develop

# Redeploy previous version
# No database changes to reverse
```

**Partial Rollback** (keep logging, revert logic):
- Revert Tasks 1-3 only
- Keep Task 4 (debug logging) for troubleshooting

**Investigation Required**:
- Check debug logs for misclassification patterns
- Compare with Notion data
- Identify edge cases not covered in testing

---

## Success Metrics

After deployment, monitor:
- ✅ No increase in invoice generation errors
- ✅ Fee section contains only Service Fee items
- ✅ Extra Payment section contains Commission + ExtraPayment items
- ✅ Debug logs show correct classifications
- ✅ No user reports of incorrect invoice sections
- ✅ Performance metrics unchanged (invoice generation time)

---

## Related Documents

- **Investigation**: `/Users/quang/workspace/dwarvesf/fortress-api/docs/issues/service-fee-invoice-split-classification.md`
- **Code Files**:
  - `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payouts.go`
  - `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go`
- **Testing**:
  - Unit tests: `pkg/service/notion/contractor_payouts_test.go`
  - Integration tests: `pkg/controller/invoice/contractor_invoice_test.go`

---

**End of Implementation Tasks Document**
