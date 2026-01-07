# Go HTML Template Best Practices for Invoice Generation

## Overview

This document summarizes best practices for using Go HTML templates to generate professional invoices, with a focus on PDF generation workflows.

## 1. Template Architecture

### 1.1 Recommended PDF Generation Stack

Based on industry research, the recommended approach for Go-based invoice PDF generation:

| Component | Recommended Tool | Purpose |
|-----------|-----------------|---------|
| Templating | `html/template` | Native Go templating with security features |
| HTML to PDF | `wkhtmltopdf` | Industry-standard HTML-to-PDF conversion |
| Wrapper | `go-wkhtmltopdf` | Go bindings for wkhtmltopdf |
| Currency | `bojanz/currency` or `Rhymond/go-money` | Locale-aware currency formatting |

**Current Implementation:** The fortress-api correctly uses this stack.

### 1.2 Template Function Registration Pattern

**Critical Rule:** Custom functions MUST be registered before template parsing.

```go
// CORRECT: Register functions before parsing
funcMap := template.FuncMap{
    "formatMoney": formatMoneyFunc,
    "formatDate":  formatDateFunc,
}
tmpl, err := template.New("invoice").Funcs(funcMap).ParseFiles(templateFile)

// INCORRECT: Will cause "undefined function" errors
tmpl, err := template.ParseFiles(templateFile)
tmpl.Funcs(funcMap) // Too late - template already parsed
```

### 1.3 Template Data Structure Pattern

Use a dedicated struct for template data to ensure type safety and clarity:

```go
type InvoiceTemplateData struct {
    Invoice       *InvoiceData
    LineItems     []LineItem
    Sections      []Section          // For grouped items
    Summary       *InvoiceSummary
    FormatOptions *FormatOptions
}
```

## 2. Custom Template Functions

### 2.1 Essential Functions for Invoices

| Function | Purpose | Example Usage |
|----------|---------|---------------|
| `formatMoney` | Currency formatting | `{{formatMoney .Amount .Currency}}` |
| `formatDate` | Date formatting | `{{formatDate .DueDate}}` |
| `formatNumber` | Number with separators | `{{formatNumber .Quantity}}` |
| `add`/`sub` | Arithmetic operations | `{{add .Index 1}}` |
| `safeHTML` | Render trusted HTML | `{{safeHTML .Description}}` |

### 2.2 Currency Formatting Function Pattern

**Recommended Pattern Using `bojanz/currency`:**

```go
import "github.com/bojanz/currency"

func createCurrencyFormatter(currencyCode, locale string) func(float64) string {
    loc := currency.NewLocale(locale)
    formatter := currency.NewFormatter(loc)

    return func(amount float64) string {
        amt, _ := currency.NewAmount(fmt.Sprintf("%.2f", amount), currencyCode)
        return formatter.Format(amt)
    }
}
```

**Current Implementation Pattern (using `go-money`):**

```go
func(amount float64) string {
    pound := money.New(1, currencyCode)
    tmpValue := amount * math.Pow(10, float64(pound.Currency().Fraction))
    return pound.Multiply(int64(math.Round(tmpValue))).Display()
}
```

### 2.3 Safe HTML Rendering

For user-generated content that needs HTML rendering (like line breaks):

```go
"formatProofOfWork": func(text string) template.HTML {
    // Sanitize and format
    formatted := strings.ReplaceAll(text, "\n", "<br>")
    return template.HTML(formatted)
}
```

**Security Note:** Only use `template.HTML` for trusted content or after proper sanitization.

## 3. Template Structure Best Practices

### 3.1 Section Organization

Templates should follow a clear visual hierarchy:

```
1. Header (Invoice title, number, dates)
2. Parties (Bill To, Bill From)
3. Line Items Table
4. Summary (Subtotal, Tax, Total)
5. Footer (Payment terms, bank details)
```

### 3.2 Table Structure for Line Items

```html
<table class="invoice-table">
    <colgroup>
        <col style="width: 50%">  <!-- Description -->
        <col style="width: 15%">  <!-- Quantity -->
        <col style="width: 15%">  <!-- Unit Price -->
        <col style="width: 20%">  <!-- Total -->
    </colgroup>
    <thead>...</thead>
    <tbody>
        {{range .LineItems}}
        <tr>...</tr>
        {{end}}
    </tbody>
    <tfoot>...</tfoot>
</table>
```

### 3.3 Conditional Rendering

```html
{{if .Invoice.Description}}
    <strong>Description:</strong> {{.Invoice.Description}}
{{end}}

{{if eq .BillingType "Hourly"}}
    <!-- Show hours and rate columns -->
{{else}}
    <!-- Show fixed amount -->
{{end}}
```

## 4. CSS Considerations for PDF

### 4.1 Color Format

Use HEX colors for PDF compatibility:

```css
/* CORRECT */
color: #333333;
background: #F3F3F5;

/* AVOID - may cause issues */
color: rgb(51, 51, 51);
```

### 4.2 Font Considerations

Use web-safe fonts or embed fonts explicitly:

```css
body {
    font-family: 'Arial', 'Helvetica', sans-serif;
}
```

### 4.3 Page Break Control

```css
.page-break {
    page-break-after: always;
}

.no-break {
    page-break-inside: avoid;
}
```

## 5. Error Handling Patterns

### 5.1 Template Execution Errors

```go
var buf bytes.Buffer
if err := tmpl.ExecuteTemplate(&buf, templateName, data); err != nil {
    return nil, fmt.Errorf("failed to execute template: %w", err)
}
```

### 5.2 Graceful Fallbacks in Templates

```html
{{if .Rate}}{{formatMoney .Rate}}{{else}}-{{end}}
```

## 6. Performance Considerations

### 6.1 Template Caching

For production, parse templates once at startup:

```go
var invoiceTemplate *template.Template

func init() {
    funcMap := template.FuncMap{...}
    invoiceTemplate = template.Must(
        template.New("invoice").Funcs(funcMap).ParseFiles(templateFile),
    )
}
```

### 6.2 Buffer Pooling

For high-volume PDF generation:

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}
```

## References

- [Building a PDF Generator Microservice in Go](https://medium.com/@yashbatra11111/building-a-pdf-generator-microservice-in-go-with-html-templates-167965e8b176)
- [Using Functions Inside Go Templates](https://www.calhoun.io/intro-to-templates-p3-functions/)
- [Go Template Package Documentation](https://pkg.go.dev/text/template)
- [GitHub: anvilco/html-pdf-invoice-template](https://github.com/anvilco/html-pdf-invoice-template)
