# Research Status

## Session: 202601071248-revise-contractor-invoice-pdf

**Status:** COMPLETED

**Last Updated:** 2026-01-07

## Research Summary

This research phase investigated best practices for Go HTML template-based invoice generation with focus on multi-currency support (VND and USD).

## Completed Research Documents

| Document | Description | Status |
|----------|-------------|--------|
| `go-html-template-best-practices.md` | Go template architecture, custom functions, PDF generation patterns | Completed |
| `currency-formatting-best-practices.md` | VND/USD formatting conventions, Go library recommendations | Completed |
| `multi-currency-invoice-patterns.md` | Exchange rate display, dual-currency patterns, IAS 21 standards | Completed |

## Key Findings

### 1. Go HTML Template Best Practices

- **Template Stack:** html/template + wkhtmltopdf (via go-wkhtmltopdf) - current implementation is correct
- **Critical Pattern:** Register FuncMap BEFORE parsing templates
- **Currency Libraries:** `bojanz/currency` (recommended) or `Rhymond/go-money` (currently used)
- **CSS:** Use HEX colors for PDF compatibility

### 2. Currency Formatting

**USD Format:**
- Symbol: `$` before amount, no space
- Thousands: comma (`,`)
- Decimal: period (`.`)
- Decimal places: 2
- Example: `$1,234.56`

**VND Format:**
- Symbol: `₫` after amount, with space
- Thousands: period (`.`) in Vietnam, comma (`,`) international
- Decimal places: 0 (no minor units)
- Example: `500.000 ₫` or `500,000 VND`

### 3. Multi-Currency Invoice Patterns

- **Exchange Rate Display:** Show rate, date, and source
- **Dual Currency:** Primary with secondary in parentheses or side-by-side columns
- **IAS 21 Compliance:** Use spot rate on transaction/invoice date
- **Grouping:** Logical sections by type, project, or period

### 4. Recommendations for Fortress-API

1. **Current Implementation Analysis:**
   - Using `Rhymond/go-money` - adequate for current needs
   - Template function registration is correct
   - Currency conversion via Wise API is appropriate

2. **Potential Improvements:**
   - Add exchange rate display on invoice footer
   - Consider dual-currency display for VND contractors
   - Implement section grouping for different payout types
   - Add VND symbol support (`₫`) for original amounts

## References Summary

### Go Templates & PDF Generation
- [Medium: Building PDF Generator in Go](https://medium.com/@yashbatra11111/building-a-pdf-generator-microservice-in-go-with-html-templates-167965e8b176)
- [Calhoun.io: Go Template Functions](https://www.calhoun.io/intro-to-templates-p3-functions/)
- [Go Template Documentation](https://pkg.go.dev/text/template)

### Currency Libraries
- [bojanz/currency](https://pkg.go.dev/github.com/bojanz/currency)
- [Rhymond/go-money](https://github.com/leekchan/accounting)
- [golang.org/x/text/currency](https://pkg.go.dev/golang.org/x/text/currency)

### Currency Formatting Standards
- [Vietnamese Dong - Wikipedia](https://en.wikipedia.org/wiki/Vietnamese_đồng)
- [ISO 4217 Standard](https://en.wikipedia.org/wiki/ISO_4217)
- [FastSpring Currency Guide](https://fastspring.com/blog/how-to-format-30-currencies-from-countries-all-over-the-world/)
- [CLDR Number/Currency Patterns](https://cldr.unicode.org/translation/number-currency-formats/number-and-currency-patterns)

### Multi-Currency & Accounting Standards
- [Stripe Multi-Currency Invoicing](https://docs.stripe.com/invoicing/multi-currency-customers)
- [IAS 21 - Foreign Currency](https://www.iasplus.com/en/standards/ias/ias21)
- [IMF: Patterns in Invoicing Currency](https://www.imf.org/en/Publications/WP/Issues/2020/07/17/Patterns-in-Invoicing-Currency-in-Global-Trade-49574)

## Next Steps

Research is complete. Findings should be used by @agent-project-manager to inform:
1. Technical specifications for invoice template improvements
2. Architecture decisions for currency formatting
3. Implementation planning for multi-currency display features
