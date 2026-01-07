# Multi-Currency Invoice Display Patterns

## Overview

This document covers best practices for displaying invoices that involve multiple currencies, including exchange rate conventions, dual-currency display, and international accounting standards.

## 1. Multi-Currency Invoice Scenarios

### 1.1 Common Use Cases

| Scenario | Description |
|----------|-------------|
| Cross-border services | Services billed in client's currency, recorded in company's base currency |
| Currency conversion | Line items in original currency, totals converted to payment currency |
| Dual display | Show amounts in both original and converted currencies |
| Mixed currency items | Different line items in different currencies |

### 1.2 Fortress-API Context

The contractor invoice system handles:
- **Source currencies:** VND, USD (contractor's local currency)
- **Target currency:** USD (Dwarves LLC base currency)
- **Conversion:** All amounts converted to USD for final invoice

## 2. Exchange Rate Display Conventions

### 2.1 What to Display

| Element | Requirement | Example |
|---------|-------------|---------|
| Exchange rate | Show rate used for conversion | "1 USD = 25,450 VND" |
| Rate date | Date when rate was obtained | "Rate as of: Dec 15, 2025" |
| Source | Optional - rate provider | "Source: Wise" |

### 2.2 Rate Display Formats

**Direct Quote Format:**
```
1 USD = 25,450 VND
```

**Indirect Quote Format:**
```
1 VND = 0.0000393 USD
```

**Recommended:** Use direct quote for readability.

### 2.3 Placement on Invoice

**Standard Locations:**
1. **Footer/Notes section** - Most common
2. **Below line items table** - When conversion is applied to all items
3. **Per line item** - When different rates apply to different items

**Example Footer:**
```
Exchange Rate: 1 USD = 25,450 VND (as of December 15, 2025)
Amounts converted from VND to USD using the above rate.
```

## 3. Dual-Currency Display Pattern

### 3.1 When to Use

- Client/contractor prefers to see their local currency
- Regulatory requirements in certain jurisdictions
- Transparency in currency conversion

### 3.2 Display Approaches

**Approach 1: Primary with Secondary in Parentheses**
```
Total: $1,234.56 (31,419,048 VND)
```

**Approach 2: Side-by-Side Columns**
```
| Description | Amount (VND)  | Amount (USD) |
|-------------|---------------|--------------|
| Service Fee | 25,000,000    | $981.75      |
```

**Approach 3: Summary Section Only**
```
Subtotal (VND): 25,000,000 VND
Exchange Rate:  1 USD = 25,450 VND
Subtotal (USD): $981.75
```

### 3.3 Template Implementation

```html
<!-- Primary currency with conversion note -->
<tr>
    <td>{{.Description}}</td>
    <td>{{formatUSD .AmountUSD}}</td>
</tr>

<!-- Optional: Show original currency -->
{{if ne .OriginalCurrency "USD"}}
<tr class="conversion-note">
    <td colspan="2" class="note">
        Original: {{formatMoney .OriginalAmount .OriginalCurrency}}
        (Rate: 1 USD = {{formatRate .ExchangeRate}} {{.OriginalCurrency}})
    </td>
</tr>
{{end}}
```

## 4. International Accounting Standards

### 4.1 IAS 21 - Foreign Currency Transactions

**Key Requirements:**
1. Use spot rate on transaction date for initial recording
2. Monetary items translated at closing rate
3. Exchange differences recognized in profit/loss

**Invoice Implications:**
- Document the exchange rate used
- Use rate from invoice date (or payment date if different)
- Keep records for audit trail

### 4.2 Rate Selection Options

| Rate Type | Use Case | Pros | Cons |
|-----------|----------|------|------|
| Spot rate (invoice date) | Standard practice | Reflects actual value | May differ from payment |
| Average rate (period) | Monthly/quarterly invoicing | Smooths volatility | Less precise |
| Forward rate | Contracted future payment | Locks in rate | Requires hedging |

### 4.3 Documentation Requirements

Include on invoice or supporting documents:
- Exchange rate used
- Date of rate
- Source of rate (if relevant)
- Any rate lock agreement

## 5. Invoice Structure for Multi-Currency

### 5.1 Recommended Layout

```
INVOICE
========

Invoice #: CONTR-202512-ABCD
Date: December 15, 2025
Due Date: December 30, 2025

BILL TO:
Company Name
Address

-----------------------------------------
| Description          | Qty  | Amount   |
-----------------------------------------
| Service Fee - Dec    |   1  | $981.75  |
| Commission - Project |   1  | $150.00  |
-----------------------------------------
|                      | Subtotal: $1,131.75
|                      | Total:    $1,131.75
-----------------------------------------

EXCHANGE RATE INFORMATION:
Original amounts were in VND.
Rate applied: 1 USD = 25,450 VND
Rate date: December 15, 2025

PAYMENT DETAILS:
...
```

### 5.2 Section Order Priority

1. Invoice header (number, dates)
2. Parties (bill to, bill from)
3. **Line items (in payment currency)**
4. Summary/totals
5. **Exchange rate information**
6. Payment terms/bank details

## 6. Grouped/Sectioned Invoice Patterns

### 6.1 Grouping Strategies

| Strategy | Use Case | Example |
|----------|----------|---------|
| By type | Different service categories | "Consulting", "Development" |
| By project | Multi-project contractors | "Project A", "Project B" |
| By period | Time-based billing | "Week 1", "Week 2" |
| By currency | Mixed currency items | "USD Items", "VND Items" |

### 6.2 Implementation Pattern

```go
type InvoiceSection struct {
    Title     string
    Items     []LineItem
    Subtotal  float64
}

type InvoiceTemplateData struct {
    Sections []InvoiceSection
    Total    float64
}
```

**Template:**
```html
{{range .Sections}}
<div class="section">
    <h3>{{.Title}}</h3>
    <table>
        {{range .Items}}
        <tr>
            <td>{{.Description}}</td>
            <td>{{formatUSD .Amount}}</td>
        </tr>
        {{end}}
        <tr class="subtotal">
            <td>Section Subtotal</td>
            <td>{{formatUSD .Subtotal}}</td>
        </tr>
    </table>
</div>
{{end}}
```

### 6.3 Visual Hierarchy

- Use borders or spacing between sections
- Clear section headers
- Section subtotals
- Grand total at bottom

## 7. Best Practices Summary

### 7.1 Do's

- Display exchange rate used for conversion
- Show rate date for audit trail
- Use consistent currency formatting throughout
- Place totals prominently
- Keep line items clear and readable
- Group related items logically

### 7.2 Don'ts

- Don't mix currency formats inconsistently
- Don't hide conversion details
- Don't use outdated exchange rates without disclosure
- Don't overcomplicate with too many currencies
- Don't forget decimal precision for calculations

### 7.3 Accessibility Considerations

- Use sufficient contrast for currency symbols
- Ensure symbols are screen-reader friendly
- Provide text alternatives for special characters

## 8. Industry Examples

### 8.1 Stripe Invoicing

- Supports 135+ currencies
- Automatic exchange rate application
- Shows original and converted amounts
- Rate locked at invoice creation

### 8.2 Wise Business

- Real-time exchange rates
- Clear rate display on invoices
- Historical rate tracking
- Multi-currency accounts

## References

- [Stripe Multi-Currency Documentation](https://docs.stripe.com/invoicing/multi-currency-customers)
- [IAS 21 - Effects of Changes in Foreign Exchange Rates](https://www.iasplus.com/en/standards/ias/ias21)
- [Invoice Best Practices - InvoiceMojo](https://invoicemojo.com/invoicing/multi-currency-invoicing/)
- [IMF Working Paper: Patterns in Invoicing Currency](https://www.imf.org/en/Publications/WP/Issues/2020/07/17/Patterns-in-Invoicing-Currency-in-Global-Trade-49574)
- [Productive Help Center: Invoices in Multiple Currencies](https://help.productive.io/en/articles/6197486-invoices-in-multiple-currencies)
