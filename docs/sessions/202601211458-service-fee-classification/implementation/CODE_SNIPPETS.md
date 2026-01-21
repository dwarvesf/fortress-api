# Ready-to-Use Code Snippets

This document contains complete, copy-paste-ready code snippets for implementing the Service Fee classification fix.

---

## Table of Contents

1. [Change 1: determineSourceType() Function](#change-1-determinesourcetype-function)
2. [Change 2: Fee Section Grouping](#change-2-fee-section-grouping)
3. [Change 3: Extra Payment Section Grouping](#change-3-extra-payment-section-grouping)
4. [Change 4: Enhanced Debug Logging](#change-4-enhanced-debug-logging)
5. [Unit Tests](#unit-tests)
6. [Integration Tests](#integration-tests)

---

## Change 1: determineSourceType() Function

**File**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payouts.go`
**Lines**: 365-380

### Complete Replacement

Replace the entire `determineSourceType()` function with this:

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
	// Priority 1: TaskOrder (Development Work)
	if entry.TaskOrderID != "" {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: sourceType=ServiceFee (TaskOrderID=%s)", entry.TaskOrderID))
		return PayoutSourceTypeServiceFee
	}

	// Priority 2: InvoiceSplit (check Description for role-based fees)
	if entry.InvoiceSplitID != "" {
		// Add nil/empty check for safety
		if entry.Description != "" {
			desc := strings.ToLower(entry.Description)

			// Service Fee: Delivery Lead or Account Management roles
			// These keywords match the Notion formula logic
			if strings.Contains(desc, "delivery lead") || strings.Contains(desc, "account management") {
				s.logger.Debug(fmt.Sprintf(
					"[DEBUG] contractor_payouts: sourceType=ServiceFee (InvoiceSplitID=%s, keywords found in Description)",
					entry.InvoiceSplitID,
				))
				return PayoutSourceTypeServiceFee
			}
		}

		// Otherwise: Commission (no keywords found)
		s.logger.Debug(fmt.Sprintf(
			"[DEBUG] contractor_payouts: sourceType=Commission (InvoiceSplitID=%s, no keywords)",
			entry.InvoiceSplitID,
		))
		return PayoutSourceTypeCommission
	}

	// Priority 3: Refund
	if entry.RefundRequestID != "" {
		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: sourceType=Refund (RefundRequestID=%s)", entry.RefundRequestID))
		return PayoutSourceTypeRefund
	}

	// Default: Extra Payment
	s.logger.Debug("[DEBUG] contractor_payouts: sourceType=ExtraPayment (no relation set)")
	return PayoutSourceTypeExtraPayment
}
```

### Ensure Import

Add to imports at top of file if not already present:

```go
import (
	"strings"
	// ... other imports ...
)
```

---

## Change 2: Fee Section Grouping

**File**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go`
**Lines**: 977-994

### Complete Replacement

Replace the Fee section grouping block with this:

```go
	// Fee section: Service Fee items from InvoiceSplit only
	// These are items with "Delivery Lead" or "Account Management" roles
	// determined by Description content during payout classification.
	//
	// Note: Service Fee items from TaskOrder go to "Development Work" section.
	// This filtering ensures only InvoiceSplit-based fees appear here.
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

---

## Change 3: Extra Payment Section Grouping

**File**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go`
**Lines**: 996-1012

### Complete Replacement

Replace the Extra Payment section grouping block with this:

```go
	// Extra Payment section: Commission items + ExtraPayment items
	// - Commission: InvoiceSplit items without special role keywords
	// - ExtraPayment: Items with no task order, invoice split, or refund relation
	//
	// Note: "Bonus" is replaced with "Fee" in descriptions for consistency.
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

---

## Change 4: Enhanced Debug Logging

### Option A: Add to determineSourceType() (Comprehensive)

Add this at the end of `determineSourceType()` function before the final return:

```go
// Enhanced debug logging for classification tracing
func (s *ContractorPayoutsService) determineSourceType(entry PayoutEntry) PayoutSourceType {
	// ... existing classification logic ...

	// Determine the source type (use variable to capture result)
	var sourceType PayoutSourceType

	// ... classification logic that sets sourceType ...

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

### Option B: Add Section-Level Logging (Simpler)

Add after all sections are created in `groupIntoSections()` (around line 1014):

```go
	// Count items per section for logging
	var devWorkCount, feeCount, extraPaymentCount, refundCount int
	for _, section := range sections {
		switch section.Name {
		case "Development Work":
			devWorkCount = len(section.Items)
		case "Fee":
			feeCount = len(section.Items)
		case "Extra Payment":
			extraPaymentCount = len(section.Items)
		case "Expense Reimbursement":
			refundCount = len(section.Items)
		}
	}

	l.Debug(fmt.Sprintf(
		"[INVOICE_SECTIONS] Total items=%d | Sections: Development=%d Fee=%d ExtraPayment=%d Refund=%d",
		len(items),
		devWorkCount,
		feeCount,
		extraPaymentCount,
		refundCount,
	))
```

---

## Unit Tests

**File**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payouts_test.go`

### Complete Test Function

```go
package notion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetermineSourceType(t *testing.T) {
	// Create service with mock logger (adjust based on your test setup)
	service := &ContractorPayoutsService{
		logger: newTestLogger(t), // Use your existing test logger helper
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
				PageID:         "page-1",
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
				PageID:         "page-2",
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
				PageID:         "page-3",
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
				PageID:         "page-4",
				TaskOrderID:    "",
				InvoiceSplitID: "invoice-split-101",
				Description:    "DELIVERY LEAD for Project X",
			},
			expected:    PayoutSourceTypeServiceFee,
			description: "Case-insensitive matching",
		},
		{
			name: "InvoiceSplit with mixed case 'Account management' - returns ServiceFee",
			entry: PayoutEntry{
				PageID:         "page-5",
				TaskOrderID:    "",
				InvoiceSplitID: "invoice-split-102",
				Description:    "Account management role",
			},
			expected:    PayoutSourceTypeServiceFee,
			description: "Case-insensitive matching",
		},
		{
			name: "InvoiceSplit without keywords - returns Commission",
			entry: PayoutEntry{
				PageID:         "page-6",
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
				PageID:         "page-7",
				TaskOrderID:    "",
				InvoiceSplitID: "invoice-split-111",
				Description:    "Delivery of project milestone",
			},
			expected:    PayoutSourceTypeCommission,
			description: "Partial keyword should NOT match",
		},
		{
			name: "InvoiceSplit with partial keyword 'Lead Developer' - returns Commission",
			entry: PayoutEntry{
				PageID:         "page-8",
				TaskOrderID:    "",
				InvoiceSplitID: "invoice-split-112",
				Description:    "Lead Developer on team",
			},
			expected:    PayoutSourceTypeCommission,
			description: "Partial keyword should NOT match",
		},
		{
			name: "InvoiceSplit with empty Description - returns Commission",
			entry: PayoutEntry{
				PageID:         "page-9",
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
				PageID:          "page-10",
				TaskOrderID:     "",
				InvoiceSplitID:  "",
				RefundRequestID: "refund-123",
				Description:     "Refund for overpayment",
			},
			expected:    PayoutSourceTypeRefund,
			description: "Refund relation takes priority",
		},
		{
			name: "No relations - returns ExtraPayment",
			entry: PayoutEntry{
				PageID:          "page-11",
				TaskOrderID:     "",
				InvoiceSplitID:  "",
				RefundRequestID: "",
				Description:     "Special bonus payment",
			},
			expected:    PayoutSourceTypeExtraPayment,
			description: "Default when no relations",
		},
		{
			name: "TaskOrder takes priority over InvoiceSplit with keywords",
			entry: PayoutEntry{
				PageID:         "page-12",
				TaskOrderID:    "task-order-999",
				InvoiceSplitID: "invoice-split-999",
				Description:    "Delivery Lead - should still be TaskOrder priority",
			},
			expected:    PayoutSourceTypeServiceFee,
			description: "TaskOrder has highest priority regardless of Description",
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

---

## Integration Tests

**File**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice_test.go`

### Test Function for Section Grouping

```go
package invoice

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/dwarvesf/fortress-api/pkg/model/notion"
)

func TestGroupIntoSections_MixedPayoutTypes(t *testing.T) {
	tests := []struct {
		name              string
		items             []ContractorInvoiceLineItem
		expectedSections  []string // Section names expected
		sectionItemCounts map[string]int
	}{
		{
			name: "Mixed payout types - all sections present",
			items: []ContractorInvoiceLineItem{
				// Service Fee from TaskOrder (Development Work)
				{
					Type:          string(notion.PayoutSourceTypeServiceFee),
					TaskOrderID:   "task-order-123",
					ServiceRateID: "rate-456",
					Description:   "Development work",
					AmountUSD:     1000.0,
				},
				// Service Fee from InvoiceSplit with keywords (Fee section)
				{
					Type:          string(notion.PayoutSourceTypeServiceFee),
					TaskOrderID:   "",
					ServiceRateID: "",
					Description:   "Delivery Lead services",
					AmountUSD:     500.0,
				},
				// Commission (Extra Payment)
				{
					Type:        string(notion.PayoutSourceTypeCommission),
					Description: "Performance bonus",
					AmountUSD:   200.0,
				},
				// ExtraPayment (Extra Payment)
				{
					Type:        string(notion.PayoutSourceTypeExtraPayment),
					Description: "Special bonus",
					AmountUSD:   150.0,
				},
				// Refund (Expense Reimbursement)
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
				"Development Work":      1,
				"Fee":                   1,
				"Extra Payment":         2, // Commission + ExtraPayment
				"Expense Reimbursement": 1,
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
		{
			name: "Service Fee filtering - TaskOrder vs InvoiceSplit",
			items: []ContractorInvoiceLineItem{
				// Should go to Development Work
				{
					Type:          string(notion.PayoutSourceTypeServiceFee),
					TaskOrderID:   "task-123",
					ServiceRateID: "rate-456",
					Description:   "Dev work",
					AmountUSD:     1000.0,
				},
				// Should go to Fee
				{
					Type:          string(notion.PayoutSourceTypeServiceFee),
					TaskOrderID:   "",
					ServiceRateID: "",
					Description:   "Delivery Lead",
					AmountUSD:     500.0,
				},
			},
			expectedSections: []string{
				"Development Work",
				"Fee",
			},
			sectionItemCounts: map[string]int{
				"Development Work": 1,
				"Fee":              1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock logger (use your existing test logger helper)
			logger := newTestLogger(t)

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

func TestGroupIntoSections_BonusReplacement(t *testing.T) {
	items := []ContractorInvoiceLineItem{
		{
			Type:        string(notion.PayoutSourceTypeCommission),
			Description: "Bonus for Q4 performance",
			AmountUSD:   500.0,
		},
		{
			Type:        string(notion.PayoutSourceTypeExtraPayment),
			Description: "Special Bonus payment",
			AmountUSD:   300.0,
		},
	}

	logger := newTestLogger(t)
	sections := groupIntoSections(items, logger)

	// Find Extra Payment section
	var extraPaymentSection *ContractorInvoiceSection
	for i := range sections {
		if sections[i].Name == "Extra Payment" {
			extraPaymentSection = &sections[i]
			break
		}
	}

	assert.NotNil(t, extraPaymentSection, "Extra Payment section should exist")
	assert.Equal(t, 2, len(extraPaymentSection.Items), "Should have 2 items")

	// Verify "Bonus" is replaced with "Fee"
	for _, item := range extraPaymentSection.Items {
		assert.NotContains(t, item.Description, "Bonus",
			"'Bonus' should be replaced with 'Fee'")
		assert.Contains(t, item.Description, "Fee",
			"Description should contain 'Fee'")
	}
}
```

### Test Logger Helper

If you don't have a test logger helper, add this:

```go
// newTestLogger creates a simple logger for tests
func newTestLogger(t *testing.T) Logger {
	// Return your project's logger implementation
	// Example (adjust to your actual logger):
	return logger.NewLogger().With("test", t.Name())
}
```

---

## Quick Verification Commands

After implementing the changes, run these commands:

```bash
# Run unit tests
go test -v ./pkg/service/notion -run TestDetermineSourceType

# Run integration tests
go test -v ./pkg/controller/invoice -run TestGroupIntoSections

# Run all tests
make test

# Build the application
make build

# Run locally
make dev

# Generate test invoice (replace {contractor_id} and {token})
curl -X POST http://localhost:8080/api/v1/invoices/contractor/{contractor_id}/generate \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"month": 1, "year": 2026}' \
  -o test-invoice.pdf

# Check debug logs
tail -f logs/fortress-api.log | grep -E "\[CLASSIFICATION\]|\[INVOICE_SECTIONS\]"
```

---

## Diff Summary

### File 1: contractor_payouts.go

```diff
  func (s *ContractorPayoutsService) determineSourceType(entry PayoutEntry) PayoutSourceType {
  	if entry.TaskOrderID != "" {
- 		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: sourceType=ServiceFee (TaskOrderID=%s)", entry.TaskOrderID))
  		return PayoutSourceTypeServiceFee
  	}
+
+ 	// For InvoiceSplit items: check Description to determine type
  	if entry.InvoiceSplitID != "" {
- 		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: sourceType=Commission (InvoiceSplitID=%s)", entry.InvoiceSplitID))
- 		return PayoutSourceTypeCommission
+ 		if entry.Description != "" {
+ 			desc := strings.ToLower(entry.Description)
+
+ 			// Service Fee: Delivery Lead or Account Management roles
+ 			if strings.Contains(desc, "delivery lead") || strings.Contains(desc, "account management") {
+ 				return PayoutSourceTypeServiceFee
+ 			}
+ 		}
+
+ 		// Otherwise: Commission
+ 		return PayoutSourceTypeCommission
  	}
+
  	if entry.RefundRequestID != "" {
- 		s.logger.Debug(fmt.Sprintf("[DEBUG] contractor_payouts: sourceType=Refund (RefundRequestID=%s)", entry.RefundRequestID))
  		return PayoutSourceTypeRefund
  	}
- 	s.logger.Debug("[DEBUG] contractor_payouts: sourceType=Extra Payment (no relation set)")
+
  	return PayoutSourceTypeExtraPayment
  }
```

### File 2: contractor_invoice.go (Fee Section)

```diff
- 	// Group Bonus (Commission) items - individual display
+ 	// Fee section: Service Fee items from InvoiceSplit only
  	var feeItems []ContractorInvoiceLineItem
- 	for i, item := range items {
- 		items[i].Description = strings.Replace(items[i].Description, "Bonus", "Fee", -1)
- 		if item.Type == string(notion.PayoutSourceTypeCommission) {
+ 	for _, item := range items {
+ 		if item.Type == string(notion.PayoutSourceTypeServiceFee) {
+ 			// Verify it's from InvoiceSplit (not TaskOrder)
+ 			if item.TaskOrderID == "" && item.ServiceRateID == "" {
  				feeItems = append(feeItems, item)
+ 			}
  		}
  	}
```

### File 3: contractor_invoice.go (Extra Payment Section)

```diff
- 	// Group Extra Payment items - individual display
+ 	// Extra Payment section: Commission items + ExtraPayment items
  	var extraPaymentItems []ContractorInvoiceLineItem
- 	for _, item := range items {
- 		if item.Type == string(notion.PayoutSourceTypeExtraPayment) {
+ 	for i, item := range items {
+ 		// Replace "Bonus" with "Fee" in descriptions
+ 		items[i].Description = strings.Replace(items[i].Description, "Bonus", "Fee", -1)
+
+ 		// Include Commission and ExtraPayment types
+ 		if item.Type == string(notion.PayoutSourceTypeCommission) ||
+ 			item.Type == string(notion.PayoutSourceTypeExtraPayment) {
  			extraPaymentItems = append(extraPaymentItems, item)
  		}
  	}
```

---

## Validation Checklist

After implementing all code changes:

- [ ] All imports are correct (especially `strings` package)
- [ ] No syntax errors (`go build` succeeds)
- [ ] Unit tests compile and pass
- [ ] Integration tests compile and pass
- [ ] Debug logging statements compile
- [ ] No linter warnings (`make lint` passes)
- [ ] Code follows project conventions
- [ ] Comments are clear and accurate

---

**Last Updated**: 2026-01-21
