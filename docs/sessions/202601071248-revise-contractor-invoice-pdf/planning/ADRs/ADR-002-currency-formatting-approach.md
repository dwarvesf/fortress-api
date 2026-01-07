# ADR-002: Currency Formatting Approach

**Status:** Proposed

**Date:** 2026-01-07

**Context:** The invoice PDF must display amounts in two different currency formats (VND and USD) with different conventions for symbols, separators, and decimal places. The current implementation uses `go-money` library with a single currency formatter.

## Decision

We will implement **custom template functions** using Go's `text/template.FuncMap` with locale-specific formatting rules, while continuing to use the existing `go-money` library for currency arithmetic where applicable.

### Formatting Rules

#### VND (Vietnamese Dong)
- **Symbol:** `₫` (after amount, with space)
- **Thousands separator:** `.` (period)
- **Decimal separator:** None (VND has no minor units)
- **Decimal places:** 0
- **Format:** `45.000.000 ₫`

#### USD (US Dollar)
- **Symbol:** `$` (before amount, no space)
- **Thousands separator:** `,` (comma)
- **Decimal separator:** `.` (period)
- **Decimal places:** 2
- **Format:** `$1,234.56`

#### Exchange Rate Display
- **Format:** `1 USD = 26,269 VND`
- **Location:** Footer of invoice as footnote
- **Style:** `*FX Rate USD 1 ~ 26269 VND`

### Template Function Design

```go
funcMap := template.FuncMap{
    // NEW: Multi-currency formatter
    "formatCurrency": func(amount float64, currency string) string {
        switch currency {
        case "VND":
            return formatVND(amount)
        case "USD":
            return formatUSD(amount)
        default:
            return fmt.Sprintf("%.2f %s", amount, currency)
        }
    },

    // NEW: Exchange rate formatter
    "formatExchangeRate": func(rate float64) string {
        // Format: "1 USD = 26,269 VND"
        return fmt.Sprintf("1 USD = %s VND", formatVNDNumber(rate))
    },

    // EXISTING: Keep for backward compatibility
    "formatMoney": func(amount float64) string {
        // Existing implementation using go-money
    },

    // EXISTING: Other functions remain unchanged
    "formatDate": func(t time.Time) string { ... },
    "float": func(n float64) string { ... },
    "formatProofOfWork": func(text string) template.HTML { ... },
}
```

### Helper Functions (Internal)

```go
// formatVND formats amount as VND with period separator and đ symbol
func formatVND(amount float64) string {
    // Round to nearest whole number (VND has no minor units)
    rounded := math.Round(amount)

    // Format with period as thousands separator
    formatted := formatNumberWithSeparator(rounded, ".")

    // Add symbol after with space
    return fmt.Sprintf("%s ₫", formatted)
}

// formatUSD formats amount as USD with comma separator and $ symbol
func formatUSD(amount float64) string {
    // Round to 2 decimal places
    rounded := math.Round(amount*100) / 100

    // Format with comma as thousands separator
    formatted := formatNumberWithSeparator(rounded, ",")

    // Add symbol before with no space
    return fmt.Sprintf("$%s", formatted)
}

// formatNumberWithSeparator formats number with specified thousands separator
func formatNumberWithSeparator(num float64, separator string) string {
    // Implementation using golang.org/x/text/language and message.Printer
    // OR manual implementation with string manipulation
}
```

## Alternatives Considered

### Alternative 1: Use Multiple go-money Instances
**Rejected:** `go-money` library would require creating separate instances for VND and USD, adding complexity without significant benefit.

### Alternative 2: Use golang.org/x/text/currency Package
**Rejected:** Overkill for two currencies. The `text/currency` package is comprehensive but adds unnecessary dependency weight for this use case.

### Alternative 3: Format Everything in Template with Custom Filters
**Rejected:** Go templates have limited string manipulation capabilities. Better to implement in Go code.

### Alternative 4: Use bojanz/currency Library
**Considered:** Research indicated this is a robust library, but it would require replacing existing `go-money` usage. Decision: Keep existing library, add custom formatters.

## Consequences

### Positive
- Precise control over formatting for each currency
- No additional external dependencies beyond existing `go-money`
- Template syntax remains simple: `{{formatCurrency .Amount .Currency}}`
- Easy to add more currencies in the future
- Formatting logic is testable in Go (not in template)
- Performance: Simple string formatting is fast

### Negative
- Custom implementation requires manual testing of edge cases
- Need to maintain custom formatting code
- Potential for formatting bugs if not thoroughly tested

### Neutral
- Replaces some `go-money` usage but keeps library for arithmetic
- Template functions must be registered before template parsing (existing pattern)

## Implementation Guidelines

### 1. Number Formatting

For thousands separator formatting, we recommend using one of:

**Option A: golang.org/x/text/message (Recommended)**
```go
import "golang.org/x/text/message"

p := message.NewPrinter(language.Vietnamese) // For VND
formatted := p.Sprintf("%.0f", amount)

p := message.NewPrinter(language.AmericanEnglish) // For USD
formatted := p.Sprintf("%.2f", amount)
```

**Option B: Manual String Manipulation**
```go
func formatWithSeparator(num float64, sep string, decimals int) string {
    // Convert to string with fixed decimals
    str := fmt.Sprintf("%.*f", decimals, num)
    parts := strings.Split(str, ".")

    // Add thousand separators
    intPart := parts[0]
    // ... reverse, insert separators, reverse back

    if decimals > 0 && len(parts) > 1 {
        return intPart + "." + parts[1]
    }
    return intPart
}
```

### 2. Testing Requirements

- Unit tests for formatVND with various amounts (0, 1000, 1000000, negative)
- Unit tests for formatUSD with various amounts (0.01, 1234.56, negative)
- Unit tests for formatExchangeRate with various rates
- Integration tests with template rendering

### 3. Template Usage

```html
<!-- Line item amount in original currency -->
<td>{{formatCurrency .OriginalAmount .OriginalCurrency}}</td>

<!-- USD total -->
<td>{{formatCurrency .Invoice.TotalUSD "USD"}}</td>

<!-- VND subtotal -->
<td>{{formatCurrency .Invoice.SubtotalVND "VND"}}</td>

<!-- Exchange rate footnote -->
<p>*FX Rate {{formatExchangeRate .Invoice.ExchangeRate}}</p>
```

## Related Documents

- ADR-001: Data Structure for Multi-Currency Support
- Specification: spec-002-template-functions.md
- Research: currency-formatting-best-practices.md
- Research: go-html-template-best-practices.md
