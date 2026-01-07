# Specification: Template Functions

**Version:** 1.0
**Date:** 2026-01-07
**Status:** Ready for Implementation

## Overview

This specification defines the template helper functions required to format multi-currency amounts and exchange rates in the contractor invoice PDF template.

## File: pkg/controller/invoice/contractor_invoice.go

### Location: GenerateContractorInvoicePDF Function

The template FuncMap is created around line 414. This specification details new functions to add and modifications to existing functions.

## New Template Functions

### 1. formatCurrency

**Purpose:** Format an amount in the specified currency with proper symbol, separators, and decimal places.

**Signature:**
```go
"formatCurrency": func(amount float64, currency string) string
```

**Parameters:**
- `amount` (float64): The numeric amount to format
- `currency` (string): Currency code ("VND" or "USD")

**Returns:** Formatted string with currency symbol and proper separators

**Implementation:**
```go
"formatCurrency": func(amount float64, currency string) string {
    switch strings.ToUpper(currency) {
    case "VND":
        return formatVND(amount)
    case "USD":
        return formatUSD(amount)
    default:
        // Fallback for unknown currencies
        return fmt.Sprintf("%.2f %s", amount, currency)
    }
},
```

**Template Usage:**
```html
<!-- Display line item in original currency -->
<td>{{formatCurrency .OriginalAmount .OriginalCurrency}}</td>

<!-- Display VND subtotal -->
<td>{{formatCurrency .Invoice.SubtotalVND "VND"}}</td>

<!-- Display USD total -->
<td>{{formatCurrency .Invoice.TotalUSD "USD"}}</td>
```

**Examples:**
```
formatCurrency(45000000, "VND")   → "45.000.000 ₫"
formatCurrency(1234.56, "USD")     → "$1,234.56"
formatCurrency(100, "USD")         → "$100.00"
formatCurrency(500000, "VND")      → "500.000 ₫"
```

### 2. formatExchangeRate

**Purpose:** Format exchange rate for display in invoice footnote.

**Signature:**
```go
"formatExchangeRate": func(rate float64) string
```

**Parameters:**
- `rate` (float64): VND per 1 USD exchange rate (e.g., 26269.5)

**Returns:** Formatted string "1 USD = X VND"

**Implementation:**
```go
"formatExchangeRate": func(rate float64) string {
    // Round to nearest whole number (VND has no minor units)
    rounded := math.Round(rate)

    // Format VND amount with period separators
    vndStr := formatVNDNumber(rounded)

    return fmt.Sprintf("1 USD = %s VND", vndStr)
},
```

**Template Usage:**
```html
<p class="exchange-rate">*FX Rate {{formatExchangeRate .Invoice.ExchangeRate}}</p>
```

**Examples:**
```
formatExchangeRate(26269.5)  → "1 USD = 26.270 VND"
formatExchangeRate(25000)    → "1 USD = 25.000 VND"
formatExchangeRate(26269.45) → "1 USD = 26.269 VND"
```

### 3. isSectionHeader

**Purpose:** Determine if a line item should be rendered as a section header (Refund, Bonus).

**Signature:**
```go
"isSectionHeader": func(itemType string) bool
```

**Parameters:**
- `itemType` (string): The Type field from ContractorInvoiceLineItem

**Returns:** Boolean indicating if this is a section header type

**Implementation:**
```go
"isSectionHeader": func(itemType string) bool {
    // Section headers: Refund and Commission (Bonus)
    // Service Fee is NOT a section header (it's aggregated differently)
    return itemType == "Refund" || itemType == "Commission"
},
```

**Template Usage:**
```html
{{if isSectionHeader .Type}}
    <!-- Render as section header row -->
{{else}}
    <!-- Render as regular item row -->
{{end}}
```

### 4. isServiceFee

**Purpose:** Determine if a line item is a Service Fee (for aggregated display).

**Signature:**
```go
"isServiceFee": func(itemType string) bool
```

**Parameters:**
- `itemType` (string): The Type field from ContractorInvoiceLineItem

**Returns:** Boolean indicating if this is a Service Fee type

**Implementation:**
```go
"isServiceFee": func(itemType string) bool {
    return itemType == "Service Fee"
},
```

**Template Usage:**
```html
{{if isServiceFee .Type}}
    <!-- Render aggregated Service Fee row with descriptions below -->
{{end}}
```

## Helper Functions (Internal)

These functions are called by the template functions but are not directly exposed to the template.

### formatVND

**Signature:**
```go
func formatVND(amount float64) string
```

**Implementation:**
```go
func formatVND(amount float64) string {
    // Round to nearest whole number (VND has no minor units)
    rounded := math.Round(amount)

    // Format with period separators
    formatted := formatVNDNumber(rounded)

    // Add symbol after with space
    return fmt.Sprintf("%s ₫", formatted)
}
```

**Examples:**
```
formatVND(45000000)  → "45.000.000 ₫"
formatVND(500000)    → "500.000 ₫"
formatVND(1234567)   → "1.234.567 ₫"
formatVND(100)       → "100 ₫"
```

### formatUSD

**Signature:**
```go
func formatUSD(amount float64) string
```

**Implementation:**
```go
func formatUSD(amount float64) string {
    // Round to 2 decimal places
    rounded := math.Round(amount*100) / 100

    // Format with comma separators
    formatted := formatUSDNumber(rounded)

    // Add symbol before with no space
    return fmt.Sprintf("$%s", formatted)
}
```

**Examples:**
```
formatUSD(1234.56)   → "$1,234.56"
formatUSD(100)       → "$100.00"
formatUSD(1000000)   → "$1,000,000.00"
formatUSD(0.99)      → "$0.99"
```

### formatVNDNumber

**Signature:**
```go
func formatVNDNumber(num float64) string
```

**Implementation Option A: Using golang.org/x/text/message (Recommended)**
```go
import (
    "golang.org/x/text/language"
    "golang.org/x/text/message"
)

func formatVNDNumber(num float64) string {
    p := message.NewPrinter(language.Vietnamese)
    // Vietnamese locale uses period as thousands separator
    return p.Sprintf("%.0f", num)
}
```

**Implementation Option B: Manual String Manipulation**
```go
func formatVNDNumber(num float64) string {
    // Convert to string with no decimals
    str := fmt.Sprintf("%.0f", num)

    // Add period separators every 3 digits from right
    var result strings.Builder
    n := len(str)

    for i, digit := range str {
        if i > 0 && (n-i)%3 == 0 {
            result.WriteRune('.')
        }
        result.WriteRune(digit)
    }

    return result.String()
}
```

### formatUSDNumber

**Signature:**
```go
func formatUSDNumber(num float64) string
```

**Implementation Option A: Using golang.org/x/text/message (Recommended)**
```go
import (
    "golang.org/x/text/language"
    "golang.org/x/text/message"
)

func formatUSDNumber(num float64) string {
    p := message.NewPrinter(language.AmericanEnglish)
    // US locale uses comma as thousands separator
    return p.Sprintf("%.2f", num)
}
```

**Implementation Option B: Manual String Manipulation**
```go
func formatUSDNumber(num float64) string {
    // Convert to string with 2 decimals
    str := fmt.Sprintf("%.2f", num)
    parts := strings.Split(str, ".")

    // Add comma separators to integer part
    intPart := parts[0]
    var result strings.Builder
    n := len(intPart)

    for i, digit := range intPart {
        if i > 0 && (n-i)%3 == 0 {
            result.WriteRune(',')
        }
        result.WriteRune(digit)
    }

    // Add decimal part
    return result.String() + "." + parts[1]
}
```

## Updated FuncMap

**Complete FuncMap Definition:**
```go
funcMap := template.FuncMap{
    // NEW: Multi-currency formatting
    "formatCurrency": func(amount float64, currency string) string {
        switch strings.ToUpper(currency) {
        case "VND":
            return formatVND(amount)
        case "USD":
            return formatUSD(amount)
        default:
            return fmt.Sprintf("%.2f %s", amount, currency)
        }
    },

    // NEW: Exchange rate formatting
    "formatExchangeRate": func(rate float64) string {
        rounded := math.Round(rate)
        vndStr := formatVNDNumber(rounded)
        return fmt.Sprintf("1 USD = %s VND", vndStr)
    },

    // NEW: Section header detection
    "isSectionHeader": func(itemType string) bool {
        return itemType == "Refund" || itemType == "Commission"
    },

    // NEW: Service Fee detection
    "isServiceFee": func(itemType string) bool {
        return itemType == "Service Fee"
    },

    // EXISTING: Keep for backward compatibility
    "formatMoney": func(amount float64) string {
        tmpValue := amount * math.Pow(10, float64(pound.Currency().Fraction))
        return pound.Multiply(int64(math.Round(tmpValue))).Display()
    },

    "formatDate": func(t time.Time) string {
        return timeutil.FormatDatetime(t)
    },

    "isMonthlyFixed": func() bool {
        return data.BillingType == "Monthly Fixed"
    },

    "isHourlyRate": func() bool {
        return data.BillingType == "Hourly Rate"
    },

    "add": func(a, b int) int {
        return a + b
    },

    "float": func(n float64) string {
        return fmt.Sprintf("%.2f", n)
    },

    "formatProofOfWork": func(text string) template.HTML {
        formatted := strings.ReplaceAll(text, " • ", "\n• ")
        formatted = strings.ReplaceAll(formatted, " •", "\n•")
        formatted = strings.ReplaceAll(formatted, "\n", "<br>")
        return template.HTML(strings.TrimSpace(formatted))
    },
}
```

## Testing Requirements

### Unit Tests for Helper Functions

**Test File:** `pkg/controller/invoice/contractor_invoice_formatters_test.go` (new file)

```go
func TestFormatVND(t *testing.T) {
    tests := []struct {
        name     string
        amount   float64
        expected string
    }{
        {"Zero", 0, "0 ₫"},
        {"Small amount", 100, "100 ₫"},
        {"Thousands", 500000, "500.000 ₫"},
        {"Millions", 45000000, "45.000.000 ₫"},
        {"With decimals (should round)", 1234.56, "1.235 ₫"},
        {"Negative (edge case)", -1000, "-1.000 ₫"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := formatVND(tt.amount)
            if result != tt.expected {
                t.Errorf("formatVND(%f) = %s; want %s", tt.amount, result, tt.expected)
            }
        })
    }
}

func TestFormatUSD(t *testing.T) {
    tests := []struct {
        name     string
        amount   float64
        expected string
    }{
        {"Zero", 0, "$0.00"},
        {"Cents only", 0.99, "$0.99"},
        {"Whole dollars", 100, "$100.00"},
        {"Thousands", 1234.56, "$1,234.56"},
        {"Millions", 1000000, "$1,000,000.00"},
        {"Rounding needed", 1234.567, "$1,234.57"},
        {"Negative (edge case)", -100.50, "$-100.50"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := formatUSD(tt.amount)
            if result != tt.expected {
                t.Errorf("formatUSD(%f) = %s; want %s", tt.amount, result, tt.expected)
            }
        })
    }
}

func TestFormatExchangeRate(t *testing.T) {
    tests := []struct {
        name     string
        rate     float64
        expected string
    }{
        {"Current rate", 26269.5, "1 USD = 26.270 VND"},
        {"Round number", 25000, "1 USD = 25.000 VND"},
        {"Needs rounding down", 26269.45, "1 USD = 26.269 VND"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := formatExchangeRate(tt.rate)
            if result != tt.expected {
                t.Errorf("formatExchangeRate(%f) = %s; want %s", tt.rate, result, tt.expected)
            }
        })
    }
}
```

### Integration Tests

Test template rendering with various data combinations:
- All VND items
- All USD items
- Mixed VND and USD items
- Zero amounts
- Large amounts

## Related Documents

- ADR-002: Currency Formatting Approach
- spec-003-html-template-restructure.md: Template usage
- Research: currency-formatting-best-practices.md
