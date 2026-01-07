# Implementation Tasks: Contractor Invoice PDF Multi-Currency Revisions

**Session:** 202601071248-revise-contractor-invoice-pdf
**Status:** Ready for Implementation
**Last Updated:** 2026-01-07

## Overview

This document provides a detailed breakdown of implementation tasks for revising the contractor invoice PDF to support multi-currency display (VND and USD) with section headers, subtotals, and proper totals calculation.

## Task Organization

Tasks are organized into phases with clear dependencies. Each task includes:
- **ID**: Unique task identifier
- **Description**: What needs to be done
- **Complexity**: S (Small), M (Medium), L (Large)
- **Estimated Time**: Approximate implementation time
- **Dependencies**: Tasks that must be completed first
- **Acceptance Criteria**: How to verify completion
- **Related Files**: Files to modify or create
- **Test Cases**: Related test specifications

## Prerequisites

Before starting implementation:
- [ ] Review all planning documents in `docs/sessions/202601071248-revise-contractor-invoice-pdf/planning/`
- [ ] Review all test specifications in `docs/sessions/202601071248-revise-contractor-invoice-pdf/test-cases/`
- [ ] Ensure development environment is set up (`make init`)
- [ ] Ensure tests can run (`make test`)

---

## Phase 1: Data Structure Changes

### TASK-001: Add Original Currency Fields to ContractorInvoiceLineItem

**Complexity:** S
**Estimated Time:** 15 minutes
**Dependencies:** None
**Priority:** P0 (Critical - Blocking)

**Description:**
Add `OriginalAmount` and `OriginalCurrency` fields to the `ContractorInvoiceLineItem` struct to preserve currency information from Notion payouts.

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go` (lines 48-61)

**Changes Required:**
```go
type ContractorInvoiceLineItem struct {
    Title       string
    Description string
    Hours       float64
    Rate        float64
    Amount      float64
    AmountUSD   float64
    Type        string

    CommissionRole    string
    CommissionProject string

    // NEW: Original currency fields
    OriginalAmount   float64 // Amount in original currency (VND or USD)
    OriginalCurrency string  // "VND" or "USD"
}
```

**Acceptance Criteria:**
- [ ] Struct compiles without errors
- [ ] New fields have descriptive comments
- [ ] Struct maintains backward compatibility

**Related Test Cases:**
- TC-006: Data Structure Population (tests 1-6)

---

### TASK-002: Add Subtotal Fields to ContractorInvoiceData

**Complexity:** S
**Estimated Time:** 15 minutes
**Dependencies:** None
**Priority:** P0 (Critical - Blocking)

**Description:**
Add multi-currency subtotal fields to the `ContractorInvoiceData` struct to support the new totals calculation and display.

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go` (lines 25-46)

**Changes Required:**
```go
type ContractorInvoiceData struct {
    // ... existing fields ...

    // NEW: Multi-currency subtotal fields
    SubtotalVND        float64 // Sum of all VND-denominated items
    SubtotalUSDFromVND float64 // SubtotalVND converted to USD at ExchangeRate
    SubtotalUSDItems   float64 // Sum of all USD-denominated items
    SubtotalUSD        float64 // SubtotalUSDFromVND + SubtotalUSDItems
    FXSupport          float64 // FX support fee (hardcoded $8 for now)
}
```

**Acceptance Criteria:**
- [ ] Struct compiles without errors
- [ ] All new fields have descriptive comments
- [ ] Comments explain what each subtotal represents
- [ ] FXSupport field includes TODO comment about future dynamic calculation

**Related Test Cases:**
- TC-006: Data Structure Population (tests 7-12)

---

### TASK-003: Update Line Item Population to Preserve Original Currency

**Complexity:** M
**Estimated Time:** 30 minutes
**Dependencies:** TASK-001
**Priority:** P0 (Critical - Blocking)

**Description:**
Update the line item creation logic in `GenerateContractorInvoice` to populate the new `OriginalAmount` and `OriginalCurrency` fields from Notion payout data.

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go` (around line 186-230)

**Changes Required:**
Locate the `lineItem := ContractorInvoiceLineItem{...}` construction and add:
```go
lineItem := ContractorInvoiceLineItem{
    Title:             title,
    Description:       description,
    Hours:             1,
    Rate:              amountUSD,
    Amount:            amountUSD,
    AmountUSD:         amountUSD,
    Type:              string(payout.SourceType),
    CommissionRole:    payout.CommissionRole,
    CommissionProject: payout.CommissionProject,

    // NEW: Preserve original currency
    OriginalAmount:   payout.Amount,
    OriginalCurrency: payout.Currency,
}
```

**Acceptance Criteria:**
- [ ] All line items have OriginalAmount set to payout.Amount
- [ ] All line items have OriginalCurrency set to payout.Currency
- [ ] Existing fields (AmountUSD, Hours, Rate, Amount) remain unchanged
- [ ] Code compiles and existing tests pass
- [ ] Add debug logging to verify field population

**Related Test Cases:**
- TC-006: Data Structure Population (tests 1-6)

---

## Phase 2: Calculation Logic Implementation

### TASK-004: Implement Currency Validation Helper Function

**Complexity:** S
**Estimated Time:** 20 minutes
**Dependencies:** TASK-001, TASK-003
**Priority:** P0 (Critical - Blocking)

**Description:**
Create a helper function to validate currency codes and amounts in line items before calculation.

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go`

**Changes Required:**
Add new helper function (around line 500):
```go
// validateLineItemCurrencies validates that all line items have valid currencies and amounts
func validateLineItemCurrencies(lineItems []ContractorInvoiceLineItem, l logger.Logger) error {
    for i, item := range lineItems {
        // Validate currency code
        if item.OriginalCurrency != "VND" && item.OriginalCurrency != "USD" {
            return fmt.Errorf("invalid currency for line item %d: %s (must be VND or USD)", i, item.OriginalCurrency)
        }

        // Validate amount is non-negative
        if item.OriginalAmount < 0 {
            return fmt.Errorf("negative amount for line item %d: %.2f %s", i, item.OriginalAmount, item.OriginalCurrency)
        }
    }
    return nil
}
```

**Acceptance Criteria:**
- [ ] Function rejects non-VND/USD currencies
- [ ] Function rejects negative amounts
- [ ] Function returns nil for valid line items
- [ ] Error messages are clear and include item index
- [ ] Function is unit tested

**Related Test Cases:**
- TC-005: Calculation Logic (validation tests 13-17)

---

### TASK-005: Implement Subtotal Calculation by Currency

**Complexity:** M
**Estimated Time:** 45 minutes
**Dependencies:** TASK-003, TASK-004
**Priority:** P0 (Critical - Blocking)

**Description:**
Implement the logic to calculate subtotals grouped by currency (VND and USD) with proper rounding.

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go` (after line item grouping, around line 310)

**Changes Required:**
Insert calculation block after line item sorting:
```go
// Validate currencies before calculations
if err := validateLineItemCurrencies(lineItems, l); err != nil {
    l.Error(err, "currency validation failed")
    return nil, err
}

// Calculate subtotals for multi-currency display
var subtotalVND float64       // Sum of all VND-denominated items
var subtotalUSDItems float64  // Sum of all USD-denominated items
var vndItemCount int
var usdItemCount int

l.Debug("calculating subtotals by currency")

for _, item := range lineItems {
    switch item.OriginalCurrency {
    case "VND":
        subtotalVND += item.OriginalAmount
        vndItemCount++
        l.Debug(fmt.Sprintf("VND item: %.0f (running total: %.0f)", item.OriginalAmount, subtotalVND))
    case "USD":
        subtotalUSDItems += item.AmountUSD
        usdItemCount++
        l.Debug(fmt.Sprintf("USD item: %.2f (running total: %.2f)", item.AmountUSD, subtotalUSDItems))
    }
}

// Round subtotals to appropriate decimal places
subtotalVND = math.Round(subtotalVND) // VND has no minor units (0 decimals)
subtotalUSDItems = math.Round(subtotalUSDItems*100) / 100 // USD has 2 decimals

l.Debug(fmt.Sprintf("subtotals calculated - VND: %.0f (%d items), USD: %.2f (%d items)",
    subtotalVND, vndItemCount, subtotalUSDItems, usdItemCount))
```

**Acceptance Criteria:**
- [ ] VND subtotal sums all VND items correctly
- [ ] USD subtotal sums all USD items correctly
- [ ] VND rounded to 0 decimal places
- [ ] USD rounded to 2 decimal places
- [ ] Item counts logged for debugging
- [ ] Handles empty line items (0 total)
- [ ] Handles all VND, all USD, and mixed scenarios

**Related Test Cases:**
- TC-005: Calculation Logic (tests 1-7)

---

### TASK-006: Implement VND to USD Conversion with Exchange Rate Capture

**Complexity:** M
**Estimated Time:** 45 minutes
**Dependencies:** TASK-005
**Priority:** P0 (Critical - Blocking)

**Description:**
Implement the logic to convert VND subtotal to USD using Wise API and capture the exchange rate.

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go` (immediately after TASK-005 code)

**Changes Required:**
```go
var exchangeRate float64
var subtotalUSDFromVND float64

if subtotalVND > 0 {
    l.Debug(fmt.Sprintf("converting VND subtotal to USD: %.0f VND", subtotalVND))

    // Convert VND to USD using Wise API
    amountUSD, rate, err := c.service.Wise.Convert(subtotalVND, "VND", "USD")
    if err != nil {
        l.Error(err, "failed to convert VND subtotal to USD")
        return nil, fmt.Errorf("failed to convert VND subtotal to USD: %w", err)
    }

    // Validate rate
    if rate <= 0 {
        l.Error(nil, fmt.Sprintf("invalid exchange rate from Wise: %.4f", rate))
        return nil, fmt.Errorf("invalid exchange rate: %.4f (must be > 0)", rate)
    }

    exchangeRate = rate
    subtotalUSDFromVND = math.Round(amountUSD*100) / 100 // Round to 2 decimals

    l.Debug(fmt.Sprintf("VND conversion - %.0f VND = %.2f USD at rate %.4f",
        subtotalVND, subtotalUSDFromVND, exchangeRate))
} else {
    // No VND items, set default values
    exchangeRate = 1.0
    subtotalUSDFromVND = 0.0

    l.Debug("no VND items, skipping conversion")
}
```

**Acceptance Criteria:**
- [ ] Calls Wise API when VND subtotal > 0
- [ ] Captures exchange rate from API response
- [ ] Validates exchange rate is positive
- [ ] Returns error if API call fails (no fallback rate)
- [ ] Rounds converted USD to 2 decimal places
- [ ] Sets exchangeRate to 1.0 when no VND items
- [ ] Comprehensive error logging
- [ ] Handles API timeout gracefully

**Related Test Cases:**
- TC-005: Calculation Logic (tests 8-12, edge cases)
- TEST-PLAN-001: Edge Cases (exchange rate scenarios)

---

### TASK-007: Implement Combined Subtotal and Total Calculation

**Complexity:** S
**Estimated Time:** 20 minutes
**Dependencies:** TASK-005, TASK-006
**Priority:** P0 (Critical - Blocking)

**Description:**
Calculate the combined USD subtotal and final total with FX support fee.

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go` (immediately after TASK-006 code)

**Changes Required:**
```go
// Calculate combined USD subtotal
subtotalUSD := subtotalUSDFromVND + subtotalUSDItems
subtotalUSD = math.Round(subtotalUSD*100) / 100 // Round to 2 decimals

l.Debug(fmt.Sprintf("combined USD subtotal: %.2f (%.2f from VND + %.2f direct USD)",
    subtotalUSD, subtotalUSDFromVND, subtotalUSDItems))

// Add FX support fee (hardcoded for now)
fxSupport := 8.0 // TODO: Implement dynamic calculation based on business rules

l.Debug(fmt.Sprintf("FX support fee: %.2f", fxSupport))

// Calculate final total
totalUSD := subtotalUSD + fxSupport
totalUSD = math.Round(totalUSD*100) / 100 // Round to 2 decimals

l.Debug(fmt.Sprintf("final total USD: %.2f (%.2f subtotal + %.2f FX support)",
    totalUSD, subtotalUSD, fxSupport))
```

**Acceptance Criteria:**
- [ ] Subtotal USD = Subtotal USD from VND + Subtotal USD Items
- [ ] FX Support is hardcoded to 8.0
- [ ] Total USD = Subtotal USD + FX Support
- [ ] All values rounded to 2 decimal places
- [ ] TODO comment added for future dynamic FX calculation
- [ ] Debug logging shows calculation breakdown

**Related Test Cases:**
- TC-005: Calculation Logic (tests 1-7)

---

### TASK-008: Update InvoiceData Population with Calculated Values

**Complexity:** S
**Estimated Time:** 15 minutes
**Dependencies:** TASK-002, TASK-007
**Priority:** P0 (Critical - Blocking)

**Description:**
Update the `invoiceData` struct initialization to include all calculated subtotals and totals.

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go` (around lines 349-369)

**Changes Required:**
Update the invoiceData initialization:
```go
invoiceData := &ContractorInvoiceData{
    InvoiceNumber:     invoiceNumber,
    ContractorName:    rateData.ContractorName,
    Month:             month,
    Date:              now,
    DueDate:           dueDate,
    Description:       description,
    BillingType:       rateData.BillingType,
    Currency:          "USD", // Invoice currency for payment is always USD
    LineItems:         lineItems,
    MonthlyFixed:      0,
    MonthlyFixedUSD:   0,
    Total:             totalUSD,      // UPDATED: Use calculated total
    TotalUSD:          totalUSD,      // UPDATED: Use calculated total
    ExchangeRate:      exchangeRate,  // UPDATED: Use actual rate from Wise
    BankAccountHolder: bankAccount.AccountHolderName,
    BankName:          bankAccount.BankName,
    BankAccountNumber: bankAccount.AccountNumber,
    BankSwiftBIC:      bankAccount.SwiftBIC,
    BankBranch:        bankAccount.BranchAddress,

    // NEW: Populate subtotal fields
    SubtotalVND:        subtotalVND,
    SubtotalUSDFromVND: subtotalUSDFromVND,
    SubtotalUSDItems:   subtotalUSDItems,
    SubtotalUSD:        subtotalUSD,
    FXSupport:          fxSupport,
}

l.Debug("invoice data populated with calculated totals")
```

**Acceptance Criteria:**
- [ ] All new subtotal fields are populated
- [ ] Total and TotalUSD use calculated totalUSD
- [ ] ExchangeRate uses actual rate from Wise API
- [ ] Currency remains "USD" for payment processing
- [ ] Code compiles without errors
- [ ] Debug logging confirms population

**Related Test Cases:**
- TC-006: Data Structure Population (tests 7-12)

---

## Phase 3: Currency Formatting Functions

### TASK-009: Implement formatVND Helper Function

**Complexity:** M
**Estimated Time:** 30 minutes
**Dependencies:** None
**Priority:** P0 (Critical - Blocking)

**Description:**
Create a helper function to format VND amounts according to Vietnamese currency conventions (period separator, symbol after).

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go` (add new helper function around line 500)

**Changes Required:**
```go
// formatVND formats a VND amount with Vietnamese conventions
// Example: 45000000 -> "45.000.000 ₫"
func formatVND(amount float64) string {
    // Round to 0 decimal places (VND has no minor units)
    rounded := math.Round(amount)

    // Convert to string without decimals
    str := fmt.Sprintf("%.0f", rounded)

    // Add period separators for thousands
    var result []rune
    for i, char := range str {
        if i > 0 && (len(str)-i)%3 == 0 {
            result = append(result, '.')
        }
        result = append(result, char)
    }

    // Add currency symbol after amount
    return string(result) + " ₫"
}
```

**Acceptance Criteria:**
- [ ] Formats with period (.) as thousands separator
- [ ] Appends ₫ symbol after amount with space
- [ ] Rounds to 0 decimal places
- [ ] Handles negative numbers (should not occur, but test)
- [ ] Handles zero correctly: "0 ₫"
- [ ] Handles large amounts (billions)
- [ ] Unit tested with all test cases from TC-001

**Related Test Cases:**
- TC-001: formatVND Function (14 test cases)

---

### TASK-010: Implement formatUSD Helper Function

**Complexity:** M
**Estimated Time:** 30 minutes
**Dependencies:** None
**Priority:** P0 (Critical - Blocking)

**Description:**
Create a helper function to format USD amounts according to US currency conventions (comma separator, $ before).

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go` (add new helper function)

**Changes Required:**
```go
// formatUSD formats a USD amount with US conventions
// Example: 1234.56 -> "$1,234.56"
func formatUSD(amount float64) string {
    // Round to 2 decimal places
    rounded := math.Round(amount*100) / 100

    // Split into integer and decimal parts
    intPart := int64(rounded)
    decPart := int((rounded - float64(intPart)) * 100)

    // Format integer part with comma separators
    intStr := fmt.Sprintf("%d", intPart)
    var result []rune
    for i, char := range intStr {
        if i > 0 && (len(intStr)-i)%3 == 0 {
            result = append(result, ',')
        }
        result = append(result, char)
    }

    // Add $ symbol before amount and decimal part
    return fmt.Sprintf("$%s.%02d", string(result), decPart)
}
```

**Acceptance Criteria:**
- [ ] Formats with comma (,) as thousands separator
- [ ] Prepends $ symbol before amount (no space)
- [ ] Always shows 2 decimal places
- [ ] Handles negative numbers (should not occur, but test)
- [ ] Handles zero correctly: "$0.00"
- [ ] Handles large amounts (millions)
- [ ] Handles small amounts (< $1.00)
- [ ] Unit tested with all test cases from TC-002

**Related Test Cases:**
- TC-002: formatUSD Function (17 test cases)

---

### TASK-011: Implement formatCurrency Template Function

**Complexity:** S
**Estimated Time:** 20 minutes
**Dependencies:** TASK-009, TASK-010
**Priority:** P0 (Critical - Blocking)

**Description:**
Create a template function that dispatches to the appropriate currency formatter based on currency code.

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go` (add new helper function)

**Changes Required:**
```go
// formatCurrency formats an amount according to its currency code
func formatCurrency(amount float64, currency string) string {
    switch strings.ToUpper(currency) {
    case "VND":
        return formatVND(amount)
    case "USD":
        return formatUSD(amount)
    default:
        // Fallback to USD formatting for unknown currencies
        return formatUSD(amount)
    }
}
```

**Acceptance Criteria:**
- [ ] Routes to formatVND for "VND" currency
- [ ] Routes to formatUSD for "USD" currency
- [ ] Case-insensitive currency matching
- [ ] Defaults to USD for unknown currencies
- [ ] Unit tested with all test cases from TC-003

**Related Test Cases:**
- TC-003: formatCurrency Function (12 test cases)

---

### TASK-012: Implement formatExchangeRate Template Function

**Complexity:** S
**Estimated Time:** 15 minutes
**Dependencies:** None
**Priority:** P1 (High)

**Description:**
Create a template function to format the exchange rate for the footnote display.

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go` (add new helper function)

**Changes Required:**
```go
// formatExchangeRate formats an exchange rate for display
// Example: 26269.123 -> "1 USD = 26,269 VND"
func formatExchangeRate(rate float64) string {
    // Round to nearest whole number for VND
    rounded := math.Round(rate)

    // Format with comma separators
    intStr := fmt.Sprintf("%.0f", rounded)
    var result []rune
    for i, char := range intStr {
        if i > 0 && (len(intStr)-i)%3 == 0 {
            result = append(result, ',')
        }
        result = append(result, char)
    }

    return fmt.Sprintf("1 USD = %s VND", string(result))
}
```

**Acceptance Criteria:**
- [ ] Format: "1 USD = X VND"
- [ ] VND amount rounded to whole number
- [ ] VND amount has comma separators
- [ ] Handles large exchange rates (> 20,000)
- [ ] Unit tested with all test cases from TC-004

**Related Test Cases:**
- TC-004: formatExchangeRate Function (14 test cases)

---

### TASK-013: Register Template Functions in FuncMap

**Complexity:** S
**Estimated Time:** 15 minutes
**Dependencies:** TASK-011, TASK-012
**Priority:** P0 (Critical - Blocking)

**Description:**
Register all new currency formatting functions in the Go template FuncMap so they're available in the HTML template.

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go` (locate template initialization, around line 380-390)

**Changes Required:**
Find the template.New().Funcs() call and add:
```go
tmpl, err := template.New("invoice").Funcs(template.FuncMap{
    "formatDate":         formatDate,
    "formatMoney":        formatMoney,
    "formatProofOfWork":  formatProofOfWork,
    "float":              float64ToString,

    // NEW: Multi-currency formatting functions
    "formatCurrency":     formatCurrency,
    "formatVND":          formatVND,
    "formatUSD":          formatUSD,
    "formatExchangeRate": formatExchangeRate,
}).ParseFiles(templatePath)
```

**Acceptance Criteria:**
- [ ] All 4 new functions registered in FuncMap
- [ ] Template compiles without errors
- [ ] Functions are callable from template
- [ ] Existing functions remain unchanged

---

## Phase 4: Section Grouping (Optional Enhancement)

**Note:** Based on requirements review, section grouping may be handled in the template using conditional logic rather than pre-processing. This phase can be deferred to post-MVP if needed.

### TASK-014: Implement Section Helper Functions (Optional)

**Complexity:** M
**Estimated Time:** 30 minutes
**Dependencies:** TASK-003
**Priority:** P2 (Medium - Enhancement)

**Description:**
Create helper functions to identify section headers and service fee items for template logic.

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice.go` (add new helper functions)

**Changes Required:**
```go
// isServiceFee returns true if the line item is a service fee type
func isServiceFee(itemType string) bool {
    return strings.Contains(strings.ToLower(itemType), "service fee")
}

// isSectionHeader returns true if this is the first item of a section
func isSectionHeader(itemType string, prevType string) bool {
    return itemType != prevType
}

// getSectionTitle returns a display title for the section
// For contractor payroll, returns "Development work from [start] to [end]"
func getSectionTitle(itemType string, invoiceMonth time.Time) string {
    switch strings.ToLower(itemType) {
    case "contractor payroll":
        startDate := invoiceMonth.Format("Jan 2")
        endDate := time.Date(invoiceMonth.Year(), invoiceMonth.Month()+1, 0, 0, 0, 0, 0, time.UTC).Format("Jan 2")
        return fmt.Sprintf("Development work from %s to %s", startDate, endDate)
    case "refund":
        return "Refund"
    case "commission":
        return "Bonus"
    default:
        return itemType
    }
}
```

**Acceptance Criteria:**
- [ ] isServiceFee correctly identifies service fee items
- [ ] isSectionHeader detects type changes
- [ ] getSectionTitle returns proper display names
- [ ] Functions registered in template FuncMap
- [ ] Unit tested per TC-007

**Related Test Cases:**
- TC-007: Section Helper Functions (20 test cases)

---

## Phase 5: HTML Template Updates

### TASK-015: Update Template to Display Original Currency in Line Items

**Complexity:** L
**Estimated Time:** 60 minutes
**Dependencies:** TASK-011, TASK-013
**Priority:** P0 (Critical - Blocking)

**Description:**
Update the HTML template to display line items in their original currency (VND or USD) instead of always showing USD.

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/templates/contractor-invoice-template.html` (tbody section, lines 216-227)

**Changes Required:**
Replace the tbody section:
```html
<tbody>
    {{range $index, $item := .LineItems}}
    <tr>
        <td style="word-wrap: break-word;">
            {{if $item.Title}}<div class="item-project">{{$item.Title}}</div>{{end}}
            <div class="item-details">{{formatProofOfWork $item.Description}}</div>
        </td>
        <td>{{if $item.Hours}}{{float $item.Hours}}{{else}}-{{end}}</td>
        <td>{{formatCurrency $item.OriginalAmount $item.OriginalCurrency}}</td>
        <td>{{formatCurrency $item.OriginalAmount $item.OriginalCurrency}}</td>
    </tr>
    {{end}}

    {{/* Keep MergedItems logic for backward compatibility */}}
    {{if .MergedItems}}
    <tr>
        <td style="word-wrap: break-word;">
            {{range $index, $item := .MergedItems}}
            {{if $item.Title}}<div class="item-project">{{$item.Title}}</div>{{end}}
            {{if $item.Description}}<div class="item-details" style="margin-bottom: 5px;">{{formatProofOfWork $item.Description}}</div>{{end}}
            {{end}}
        </td>
        <td>1</td>
        <td>{{if .MergedTotal}}$ {{printf "%.2f" .MergedTotal}}{{else}}-{{end}}</td>
        <td>{{if .MergedTotal}}$ {{printf "%.2f" .MergedTotal}}{{else}}-{{end}}</td>
    </tr>
    {{end}}
</tbody>
```

**Acceptance Criteria:**
- [ ] UNIT COST column displays original currency amount
- [ ] TOTAL column displays original currency amount
- [ ] formatCurrency function used for both columns
- [ ] MergedItems section preserved for backward compatibility
- [ ] Template compiles and renders without errors
- [ ] Visual testing confirms proper display

**Related Test Cases:**
- Manual visual testing with sample data

---

### TASK-016: Update Template Totals Section with Multi-Currency Subtotals

**Complexity:** L
**Estimated Time:** 60 minutes
**Dependencies:** TASK-012, TASK-013, TASK-015
**Priority:** P0 (Critical - Blocking)

**Description:**
Completely restructure the tfoot section to display VND subtotal, USD subtotal, FX support, and total according to the new requirements.

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/templates/contractor-invoice-template.html` (tfoot section, lines 242-259)

**Changes Required:**
Replace the entire tfoot section:
```html
<tfoot>
    <tr>
        <td colspan="4" style="padding-top:8px;"></td>
    </tr>

    {{/* Show VND Subtotal only if there are VND items */}}
    {{if gt .Invoice.SubtotalVND 0}}
    <tr style="line-height: 1.4">
        <td colspan="2"></td>
        <td style="font-weight: bold; text-align: right; font-size: 13px;">Subtotal</td>
        <td style="font-size: 13px;">{{formatVND .Invoice.SubtotalVND}}</td>
    </tr>
    {{end}}

    {{/* Show USD Subtotal */}}
    <tr style="line-height: 1.4">
        <td colspan="2"></td>
        {{if gt .Invoice.SubtotalVND 0}}
        <td style="text-align: right; font-size: 13px;"></td>
        {{else}}
        <td style="font-weight: bold; text-align: right; font-size: 13px;">Subtotal</td>
        {{end}}
        <td style="font-size: 13px;">{{formatUSD .Invoice.SubtotalUSD}}</td>
    </tr>

    {{/* FX Support Fee */}}
    <tr style="line-height: 1.4">
        <td colspan="2"></td>
        <td style="text-align: right; font-size: 13px;">FX support</td>
        <td style="font-size: 13px;">{{formatUSD .Invoice.FXSupport}}</td>
    </tr>

    {{/* Separator line */}}
    <tr>
        <td colspan="4" style="border-top: 2px solid #333; padding: 5px 0;"></td>
    </tr>

    {{/* Total */}}
    <tr style="line-height: 1.4">
        <td colspan="2"></td>
        <td style="font-weight: bold; text-align: right; font-size: 14px;">Total</td>
        <td style="font-weight: bold; font-size: 14px;">{{formatUSD .Invoice.TotalUSD}}</td>
    </tr>
</tfoot>
```

**Acceptance Criteria:**
- [ ] VND subtotal only shown when SubtotalVND > 0
- [ ] USD subtotal always shown
- [ ] FX support fee displayed
- [ ] Total displayed in bold
- [ ] All amounts use appropriate formatting functions
- [ ] Spacing and alignment matches requirements
- [ ] Template compiles without errors
- [ ] Visual testing confirms layout matches specification

**Related Test Cases:**
- Manual visual testing with:
  - All VND items
  - All USD items
  - Mixed VND and USD items

---

### TASK-017: Add Exchange Rate Footnote to Template

**Complexity:** S
**Estimated Time:** 15 minutes
**Dependencies:** TASK-012, TASK-016
**Priority:** P1 (High)

**Description:**
Add the exchange rate footnote below the invoice table, showing the VND to USD exchange rate.

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/templates/contractor-invoice-template.html` (after table, before footer, around line 261)

**Changes Required:**
Add after the closing `</table>` tag and before the footer section:
```html
        </table>

        {{/* Exchange Rate Footnote - only show if VND items exist */}}
        {{if gt .Invoice.SubtotalVND 0}}
        <div style="margin-top: 10px; font-size: 11px; color: #666; font-style: italic;">
            *FX Rate: {{formatExchangeRate .Invoice.ExchangeRate}}
        </div>
        {{end}}

        <!-- Footer Section -->
        <div class="footer">
```

**Acceptance Criteria:**
- [ ] Footnote only displayed when SubtotalVND > 0
- [ ] Uses formatExchangeRate function
- [ ] Styled as small, gray, italic text
- [ ] Positioned between table and footer
- [ ] Template compiles without errors

---

### TASK-018: Update Template CSS for Section Headers (Optional)

**Complexity:** M
**Estimated Time:** 30 minutes
**Dependencies:** TASK-014
**Priority:** P2 (Medium - Enhancement)

**Description:**
Add CSS styles for section headers if implementing section grouping.

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/templates/contractor-invoice-template.html` (style section)

**Changes Required:**
```css
.invoice-table .section-header {
    font-weight: bold;
    font-size: 12px;
    background-color: #E8E8EA !important;
    border-bottom: 2px solid #D8D8D8 !important;
    padding-top: 8px !important;
    padding-bottom: 8px !important;
}

.invoice-table .section-header td {
    background-color: #E8E8EA !important;
}

.invoice-table .service-fee-detail {
    padding-left: 20px;
    font-style: italic;
    color: #666;
}
```

**Acceptance Criteria:**
- [ ] Section headers visually distinct from line items
- [ ] Service fee details indented and styled
- [ ] Styles don't break existing layout
- [ ] Visual testing confirms proper appearance

---

## Phase 6: Unit Testing

### TASK-019: Write Unit Tests for Currency Formatting Functions

**Complexity:** M
**Estimated Time:** 90 minutes
**Dependencies:** TASK-009, TASK-010, TASK-011, TASK-012
**Priority:** P0 (Critical - Blocking)

**Description:**
Implement comprehensive unit tests for all currency formatting functions following the test specifications.

**Files to Create:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice_formatters_test.go` (new file)

**Changes Required:**
Create new test file with table-driven tests for:
1. formatVND (14 test cases from TC-001)
2. formatUSD (17 test cases from TC-002)
3. formatCurrency (12 test cases from TC-003)
4. formatExchangeRate (14 test cases from TC-004)

**Test Structure:**
```go
func TestFormatVND(t *testing.T) {
    tests := []struct {
        name     string
        amount   float64
        expected string
    }{
        // Test cases from TC-001
        {"Standard amount", 45000000, "45.000.000 ₫"},
        {"Zero amount", 0, "0 ₫"},
        {"Small amount", 100, "100 ₫"},
        // ... remaining cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := formatVND(tt.amount)
            require.Equal(t, tt.expected, result)
        })
    }
}

// Similar structure for other format functions
```

**Acceptance Criteria:**
- [ ] All 57 test cases implemented (TC-001 through TC-004)
- [ ] Tests use table-driven pattern
- [ ] All tests pass
- [ ] Test file follows project conventions
- [ ] Tests cover happy path and edge cases
- [ ] Code coverage > 95% for formatter functions

**Related Test Cases:**
- TC-001: formatVND (14 tests)
- TC-002: formatUSD (17 tests)
- TC-003: formatCurrency (12 tests)
- TC-004: formatExchangeRate (14 tests)

---

### TASK-020: Write Unit Tests for Calculation Logic

**Complexity:** L
**Estimated Time:** 120 minutes
**Dependencies:** TASK-005, TASK-006, TASK-007, TASK-008
**Priority:** P0 (Critical - Blocking)

**Description:**
Implement comprehensive unit tests for the calculation logic including subtotals, conversions, and totals.

**Files to Create:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice_calculations_test.go` (new file)

**Test Coverage:**
1. Subtotal calculations by currency (7 test cases from TC-005)
2. Exchange rate conversion scenarios (5 test cases from TC-005)
3. Validation logic (5 test cases from TC-005)
4. Edge cases from TEST-PLAN-001

**Acceptance Criteria:**
- [ ] All 17 calculation test cases implemented (TC-005)
- [ ] Tests use table-driven pattern with mocked Wise API
- [ ] All tests pass
- [ ] Tests cover all VND, all USD, and mixed scenarios
- [ ] Tests verify proper rounding
- [ ] Tests verify error handling
- [ ] Code coverage > 90% for calculation functions

**Related Test Cases:**
- TC-005: Calculation Logic (17 tests)
- TEST-PLAN-001: Edge Cases (calculation scenarios)

---

### TASK-021: Write Unit Tests for Data Structure Population

**Complexity:** M
**Estimated Time:** 60 minutes
**Dependencies:** TASK-003, TASK-008
**Priority:** P1 (High)

**Description:**
Implement unit tests to verify that data structures are correctly populated with original currency and calculated values.

**Files to Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice_test.go` (update existing tests)

**Test Coverage:**
1. LineItem field population (6 test cases from TC-006)
2. InvoiceData field population (6 test cases from TC-006)

**Acceptance Criteria:**
- [ ] All 12 data structure tests implemented (TC-006)
- [ ] Tests verify OriginalAmount and OriginalCurrency fields
- [ ] Tests verify all subtotal fields
- [ ] Tests verify ExchangeRate field
- [ ] All tests pass
- [ ] Existing tests still pass (no regression)

**Related Test Cases:**
- TC-006: Data Structure Population (12 tests)

---

### TASK-022: Write Unit Tests for Section Helper Functions (Optional)

**Complexity:** M
**Estimated Time:** 45 minutes
**Dependencies:** TASK-014
**Priority:** P2 (Medium - Enhancement)

**Description:**
Implement unit tests for section identification and grouping helper functions.

**Files to Create/Modify:**
- `/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/contractor_invoice_sections_test.go` (new file)

**Test Coverage:**
- All 20 test cases from TC-007

**Acceptance Criteria:**
- [ ] All 20 section helper tests implemented (TC-007)
- [ ] Tests use table-driven pattern
- [ ] All tests pass
- [ ] Code coverage > 95% for section functions

**Related Test Cases:**
- TC-007: Section Helper Functions (20 tests)

---

## Phase 7: Integration Testing

### TASK-023: Integration Test with Real Notion and Wise Data

**Complexity:** M
**Estimated Time:** 45 minutes
**Dependencies:** All previous tasks
**Priority:** P1 (High)

**Description:**
Perform end-to-end integration testing with actual Notion payout data and Wise API to verify the complete flow.

**Test Scenarios:**
1. Generate invoice for contractor with all VND items
2. Generate invoice for contractor with all USD items
3. Generate invoice for contractor with mixed VND and USD items
4. Verify PDF output visually matches specifications

**Acceptance Criteria:**
- [ ] Invoice generation succeeds for all scenarios
- [ ] All subtotals calculated correctly
- [ ] Exchange rate captured from Wise API
- [ ] PDF renders correctly with proper formatting
- [ ] All sections display correctly
- [ ] No errors in logs

**Test Data Required:**
- Contractor with VND payouts
- Contractor with USD payouts
- Contractor with mixed currency payouts

---

### TASK-024: Visual Testing of PDF Output

**Complexity:** M
**Estimated Time:** 30 minutes
**Dependencies:** TASK-015, TASK-016, TASK-017
**Priority:** P0 (Critical - Blocking)

**Description:**
Manually review generated PDF invoices to ensure visual appearance matches requirements and specifications.

**Test Checklist:**
- [ ] VND amounts formatted with period separators: "45.000.000 ₫"
- [ ] USD amounts formatted with comma separators: "$1,234.56"
- [ ] Section headers appear (if implemented)
- [ ] Service fee items aggregated correctly (if implemented)
- [ ] Refund and bonus items show individual amounts
- [ ] VND subtotal appears only when VND items exist
- [ ] USD subtotal appears
- [ ] FX support fee appears: "$8.00"
- [ ] Total appears in bold
- [ ] Exchange rate footnote appears (format: "*FX Rate: 1 USD = 26,269 VND")
- [ ] Overall layout is clean and readable
- [ ] No overlapping text or alignment issues
- [ ] Bank info section unchanged

**Test Scenarios:**
1. All VND items invoice
2. All USD items invoice
3. Mixed currency invoice
4. Single item invoice
5. Many items invoice (10+)

---

## Phase 8: Documentation and Cleanup

### TASK-025: Update Implementation STATUS.md

**Complexity:** S
**Estimated Time:** 15 minutes
**Dependencies:** All previous tasks
**Priority:** P1 (High)

**Description:**
Create or update the implementation STATUS.md to document completion of tasks and any issues encountered.

**Files to Create:**
- `/Users/quang/workspace/dwarvesf/fortress-api/docs/sessions/202601071248-revise-contractor-invoice-pdf/implementation/STATUS.md`

**Content Required:**
- List of completed tasks
- Test results summary
- Known issues or limitations
- Next steps or future enhancements

**Acceptance Criteria:**
- [ ] STATUS.md created and complete
- [ ] All tasks marked as complete or blocked
- [ ] Test results documented
- [ ] Any deviations from plan explained

---

### TASK-026: Code Review and Cleanup

**Complexity:** M
**Estimated Time:** 30 minutes
**Dependencies:** All implementation tasks
**Priority:** P1 (High)

**Description:**
Review all changed code for quality, consistency, and adherence to project conventions.

**Review Checklist:**
- [ ] All debug logging statements appropriate
- [ ] No commented-out code
- [ ] All TODO comments have context
- [ ] Error messages are clear and actionable
- [ ] Function and variable names follow conventions
- [ ] Comments are accurate and helpful
- [ ] No unnecessary code duplication
- [ ] Proper error handling throughout
- [ ] All magic numbers explained or moved to constants

**Acceptance Criteria:**
- [ ] Code passes `make lint`
- [ ] All tests pass: `make test`
- [ ] Code review checklist completed
- [ ] No technical debt introduced

---

### TASK-027: Update Project Documentation

**Complexity:** S
**Estimated Time:** 20 minutes
**Dependencies:** TASK-025, TASK-026
**Priority:** P2 (Medium)

**Description:**
Update any relevant project documentation to reflect the new multi-currency invoice functionality.

**Files to Update:**
- Project README (if invoice generation is documented)
- API documentation (if invoice endpoint is public)
- Developer guide (if exists)

**Acceptance Criteria:**
- [ ] Documentation mentions multi-currency support
- [ ] Exchange rate capture documented
- [ ] FX support fee documented as hardcoded $8
- [ ] Any breaking changes noted (should be none)

---

## Task Dependencies Summary

```
Phase 1 (Data Structures):
TASK-001 → TASK-003
TASK-002 → TASK-008

Phase 2 (Calculations):
TASK-003 → TASK-004 → TASK-005 → TASK-006 → TASK-007 → TASK-008

Phase 3 (Formatting):
TASK-009, TASK-010 → TASK-011 → TASK-013
TASK-012 → TASK-013

Phase 4 (Section Grouping - Optional):
TASK-003 → TASK-014

Phase 5 (Template):
TASK-011, TASK-013 → TASK-015
TASK-012, TASK-013, TASK-015 → TASK-016
TASK-012, TASK-016 → TASK-017
TASK-014 → TASK-018 (optional)

Phase 6 (Unit Tests):
TASK-009, TASK-010, TASK-011, TASK-012 → TASK-019
TASK-005, TASK-006, TASK-007, TASK-008 → TASK-020
TASK-003, TASK-008 → TASK-021
TASK-014 → TASK-022 (optional)

Phase 7 (Integration):
All previous → TASK-023, TASK-024

Phase 8 (Documentation):
All previous → TASK-025, TASK-026, TASK-027
```

## Estimated Total Time

**Critical Path (P0 Tasks Only):**
- Phase 1: 1.0 hours
- Phase 2: 2.5 hours
- Phase 3: 1.75 hours
- Phase 5: 2.0 hours
- Phase 6: 4.5 hours
- Phase 7: 1.25 hours
- Phase 8: 1.0 hour

**Total (Critical Path):** ~14 hours

**With Optional Tasks (P1-P2):** ~16-18 hours

## Success Criteria

The implementation is complete when:
- [ ] All P0 tasks completed
- [ ] All P0 unit tests passing
- [ ] Integration tests passing
- [ ] Visual testing confirms PDF matches specification
- [ ] Code passes linting and review
- [ ] Documentation updated
- [ ] No regression in existing invoice functionality

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Exchange rate API failure | Comprehensive error handling, fail fast with clear error |
| Rounding errors | Use banker's rounding, extensive test coverage |
| Template rendering errors | Incremental testing after each template change |
| Performance impact of formatting | Benchmark if needed, optimize later |
| Breaking existing invoices | Maintain backward compatibility, test old template |

## Notes

- **Section Grouping:** Phase 4 (TASK-014, TASK-018, TASK-022) is marked as optional and can be deferred if timeline is tight. The multi-currency display and totals calculation (Phases 1-3, 5) are the core requirements.

- **FX Support Fee:** Currently hardcoded to $8.00 as per requirements. Dynamic calculation to be implemented in future iteration.

- **Testing Priority:** Focus on P0 tests (formatters and calculations) first, as these are critical for correctness. P1 and P2 tests can be completed after core functionality is working.

- **Backward Compatibility:** All existing fields and template variables remain unchanged to ensure old invoices/templates continue to work.

## Related Documents

- **Requirements:** `docs/sessions/202601071248-revise-contractor-invoice-pdf/requirements/requirements.md`
- **Planning:** `docs/sessions/202601071248-revise-contractor-invoice-pdf/planning/STATUS.md`
- **Test Cases:** `docs/sessions/202601071248-revise-contractor-invoice-pdf/test-cases/STATUS.md`
- **Specifications:**
  - `spec-001-data-structure-changes.md`
  - `spec-002-template-functions.md`
  - `spec-003-html-template-restructure.md`
  - `spec-004-calculation-logic.md`
- **ADRs:**
  - `ADR-001-data-structure-multi-currency.md`
  - `ADR-002-currency-formatting-approach.md`
