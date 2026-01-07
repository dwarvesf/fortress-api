# Currency Formatting Best Practices in Go

## Overview

This document covers best practices for currency formatting in Go, with specific focus on VND (Vietnamese Dong) and USD (US Dollar) display patterns for invoice generation.

## 1. Foundational Principles

### 1.1 Never Use Floats for Currency

**Critical Rule:** Avoid using `float64` for currency calculations due to precision issues.

**Recommended Approaches:**
1. Store amounts in minor units (cents) as integers
2. Use `big.Rat` (Go < 1.5) or `big.Float` (Go >= 1.5)
3. Use dedicated currency libraries

```go
// BAD: Float precision issues
price := 19.99
total := price * 100 // May result in 1998.9999999...

// GOOD: Store in minor units
priceInCents := int64(1999)
totalInCents := priceInCents * 100 // Exact: 199900

// GOOD: Use currency library
amount, _ := currency.NewAmount("19.99", "USD")
```

### 1.2 Formatting is Locale-Specific

Currency formatting varies by locale, not currency:
- **US locale:** "$1,234.56" (comma thousands, period decimal)
- **European locales:** "1.234,56 $" or "1 234,56 $" (varies by country)
- **Vietnamese locale:** "1.234.567 ₫" (period thousands, no decimal)

## 2. USD Formatting Standards

### 2.1 Standard Format

| Element | Convention |
|---------|-----------|
| Symbol | `$` |
| Position | Before amount (no space) |
| Thousands separator | Comma `,` |
| Decimal separator | Period `.` |
| Decimal places | 2 (cents) |
| ISO Code | USD |

**Examples:**
- `$1,234.56`
- `$0.99`
- `$1,000,000.00`

### 2.2 Accounting Format for Negative Values

For accounting/financial documents:
- Standard: `-$3.00`
- Accounting: `($3.00)`

```go
formatter := currency.NewFormatter(currency.NewLocale("en"))
formatter.AccountingStyle = true
```

### 2.3 Implementation Example

```go
func formatUSD(amount float64) string {
    locale := currency.NewLocale("en-US")
    formatter := currency.NewFormatter(locale)
    amt, _ := currency.NewAmount(fmt.Sprintf("%.2f", amount), "USD")
    return formatter.Format(amt) // "$1,234.56"
}
```

## 3. VND Formatting Standards

### 3.1 Standard Format

| Element | Convention |
|---------|-----------|
| Symbol | `₫` (U+20AB) or `d` |
| Position | After amount (with space) |
| Thousands separator | Period `.` or Comma `,` |
| Decimal separator | Comma `,` (rarely used) |
| Decimal places | 0 (VND has no minor units) |
| ISO Code | VND |

**Vietnam Local Standard:**
- Thousands: Period `.`
- Decimal: Comma `,`
- Example: `500.000 ₫` (five hundred thousand dong)

**International Standard:**
- May use comma for thousands
- Example: `500,000 ₫`

### 3.2 Symbol Variations

| Symbol | Usage |
|--------|-------|
| `₫` | Official Unicode symbol (U+20AB) |
| `d` | ASCII fallback |
| `VND` | ISO code (formal documents) |
| `dong` | Written form |

### 3.3 Implementation Example

```go
func formatVND(amount float64) string {
    // VND has no decimal places (fraction = 0)
    rounded := int64(math.Round(amount))

    // Format with period as thousands separator
    formatted := formatWithThousandsSeparator(rounded, ".")

    return fmt.Sprintf("%s ₫", formatted)
}

// Alternative using library
func formatVNDWithLibrary(amount float64) string {
    locale := currency.NewLocale("vi-VN")
    formatter := currency.NewFormatter(locale)
    amt, _ := currency.NewAmount(fmt.Sprintf("%.0f", amount), "VND")
    return formatter.Format(amt)
}
```

### 3.4 VND Display Examples

| Amount | Vietnam Format | International Format |
|--------|---------------|---------------------|
| 500000 | 500.000 ₫ | 500,000 VND |
| 1234567 | 1.234.567 ₫ | 1,234,567 VND |
| 50000000 | 50.000.000 ₫ | 50,000,000 VND |

## 4. Go Currency Libraries Comparison

### 4.1 bojanz/currency (Recommended)

**Strengths:**
- CLDR v47 locale data
- Proper amount handling with `big.Rat`
- Formatting and parsing
- Custom currency registration

```go
import "github.com/bojanz/currency"

amount, _ := currency.NewAmount("1234.56", "USD")
formatter := currency.NewFormatter(currency.NewLocale("en"))
formatted := formatter.Format(amount) // "$1,234.56"
```

### 4.2 Rhymond/go-money

**Strengths:**
- Fowler's Money pattern
- 181 currencies supported
- Simple API

```go
import "github.com/Rhymond/go-money"

m := money.New(123456, money.USD) // 1234.56 USD
formatted := m.Display() // "$1,234.56"
```

### 4.3 golang.org/x/text/currency (Official)

**Note:** Still under development, API may change.

```go
import "golang.org/x/text/currency"

unit := currency.USD
// Formatting requires integration with message package
```

### 4.4 Recommendation Matrix

| Use Case | Recommended Library |
|----------|-------------------|
| Production invoice generation | `bojanz/currency` |
| Simple formatting needs | `Rhymond/go-money` |
| Maximum locale support | `bojanz/currency` |
| Minimal dependencies | Custom implementation |

## 5. Template Function Implementation

### 5.1 Multi-Currency Formatter

```go
func createCurrencyFuncMap() template.FuncMap {
    return template.FuncMap{
        "formatUSD": func(amount float64) string {
            return fmt.Sprintf("$ %.2f", amount)
        },
        "formatVND": func(amount float64) string {
            rounded := int64(math.Round(amount))
            return fmt.Sprintf("%s ₫", formatThousands(rounded, "."))
        },
        "formatMoney": func(amount float64, currencyCode string) string {
            switch currencyCode {
            case "VND":
                return formatVND(amount)
            case "USD":
                return formatUSD(amount)
            default:
                return fmt.Sprintf("%.2f %s", amount, currencyCode)
            }
        },
    }
}
```

### 5.2 Thousands Separator Helper

```go
func formatThousands(n int64, sep string) string {
    s := strconv.FormatInt(n, 10)
    if len(s) <= 3 {
        return s
    }

    var result strings.Builder
    start := len(s) % 3
    if start > 0 {
        result.WriteString(s[:start])
    }

    for i := start; i < len(s); i += 3 {
        if result.Len() > 0 {
            result.WriteString(sep)
        }
        result.WriteString(s[i : i+3])
    }

    return result.String()
}
```

## 6. Common Pitfalls

### 6.1 Rounding Issues

```go
// BAD: Truncation instead of rounding
int64(amount * 100) // 19.99 * 100 = 1998 (truncated)

// GOOD: Proper rounding
int64(math.Round(amount * 100)) // 19.99 * 100 = 1999 (rounded)
```

### 6.2 Currency Symbol Encoding

Ensure UTF-8 encoding for currency symbols:

```html
<meta charset="UTF-8">
```

### 6.3 Locale Mismatch

Don't format VND with US locale conventions:

```go
// BAD: US format for VND
"$1,234" // Wrong symbol, wrong context

// GOOD: Appropriate format
"1.234 ₫" // Vietnam convention
"1,234 VND" // International convention
```

## References

- [ISO 4217 - Currency Codes](https://en.wikipedia.org/wiki/ISO_4217)
- [Vietnamese Dong - Wikipedia](https://en.wikipedia.org/wiki/Vietnamese_đồng)
- [bojanz/currency Documentation](https://pkg.go.dev/github.com/bojanz/currency)
- [Rhymond/go-money Documentation](https://github.com/leekchan/accounting)
- [CLDR Number/Currency Patterns](https://cldr.unicode.org/translation/number-currency-formats/number-and-currency-patterns)
- [FastSpring Currency Formatting Guide](https://fastspring.com/blog/how-to-format-30-currencies-from-countries-all-over-the-world/)
