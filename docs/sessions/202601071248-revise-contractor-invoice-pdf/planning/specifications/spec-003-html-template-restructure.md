# Specification: HTML Template Restructure

**Version:** 1.0
**Date:** 2026-01-07
**Status:** Ready for Implementation

## Overview

This specification details the changes required to restructure the contractor invoice HTML template to display section headers, group Service Fee items, and show multi-currency totals.

## File: pkg/templates/contractor-invoice-template.html

### Current Template Structure (Lines 200-260)

The current template has a simple table with all line items displayed uniformly, followed by a basic totals section.

### Required Changes

## 1. Table Body Restructure

**Location:** `<tbody>` section (lines 216-240)

**Current Implementation:**
```html
<tbody>
    {{range $index, $item := .LineItems}}
    <tr>
        <td style="word-wrap: break-word;">
            {{if $item.Title}}<div class="item-project">{{$item.Title}}</div>{{end}}
            <div class="item-details">{{formatProofOfWork $item.Description}}</div>
        </td>
        <td>{{if $item.Hours}}{{float $item.Hours}}{{else}}-{{end}}</td>
        <td>{{if $item.Rate}}{{formatMoney $item.Rate}}{{else}}-{{end}}</td>
        <td>{{if $item.AmountUSD}}$ {{printf "%.2f" $item.AmountUSD}}{{else}}-{{end}}</td>
    </tr>
    {{end}}
</tbody>
```

**New Implementation:**

The new implementation needs to handle three different display patterns:

### Pattern 1: Service Fee Items (Aggregated)

Service Fee items should be aggregated into ONE row showing total amount, with descriptions listed below (no individual amounts).

```html
{{/* Group Service Fee items */}}
{{$serviceFeeItems := list}}
{{$serviceFeeTotal := 0.0}}
{{$serviceFeeCurrency := ""}}
{{range .LineItems}}
    {{if isServiceFee .Type}}
        {{$serviceFeeItems = append $serviceFeeItems .}}
        {{$serviceFeeTotal = add $serviceFeeTotal .OriginalAmount}}
        {{$serviceFeeCurrency = .OriginalCurrency}}
    {{end}}
{{end}}

{{/* Render aggregated Service Fee row if exists */}}
{{if $serviceFeeItems}}
<tr>
    <td style="word-wrap: break-word;">
        <div class="item-project"><strong>Development work from {{.Invoice.Month | formatMonthStart}} to {{.Invoice.Month | formatMonthEnd}}</strong></div>
        {{range $serviceFeeItems}}
        <div class="item-details">{{formatProofOfWork .Description}}</div>
        {{end}}
    </td>
    <td>1</td>
    <td>{{formatCurrency $serviceFeeTotal $serviceFeeCurrency}}</td>
    <td>{{formatCurrency $serviceFeeTotal $serviceFeeCurrency}}</td>
</tr>
<tr><td colspan="4" style="padding: 2px; background-color: transparent;"></td></tr>
{{end}}
```

**Note:** Since Go html/template doesn't support variable assignment and list operations like this, we need to restructure the approach. See "Alternative Implementation" below.

### Pattern 2: Refund Section (Header + Individual Items)

Refund items should have a section header row (bold, no amounts) followed by individual item rows.

```html
{{/* Refund Section */}}
{{$hasRefund := false}}
{{range .LineItems}}
    {{if eq .Type "Refund"}}
        {{if not $hasRefund}}
            {{/* Section header row */}}
            <tr>
                <td><strong>Refund</strong></td>
                <td></td>
                <td></td>
                <td></td>
            </tr>
            {{$hasRefund = true}}
        {{end}}
        {{/* Individual refund item row */}}
        <tr>
            <td style="padding-left: 24px;">{{formatProofOfWork .Description}}</td>
            <td>1</td>
            <td>{{formatCurrency .OriginalAmount .OriginalCurrency}}</td>
            <td>{{formatCurrency .OriginalAmount .OriginalCurrency}}</td>
        </tr>
    {{end}}
{{end}}
{{if $hasRefund}}
<tr><td colspan="4" style="padding: 2px; background-color: transparent;"></td></tr>
{{end}}
```

### Pattern 3: Bonus Section (Header + Individual Items)

Bonus (Commission) items should have a section header row followed by individual item rows.

```html
{{/* Bonus Section */}}
{{$hasBonus := false}}
{{range .LineItems}}
    {{if eq .Type "Commission"}}
        {{if not $hasBonus}}
            {{/* Section header row */}}
            <tr>
                <td><strong>Bonus</strong></td>
                <td></td>
                <td></td>
                <td></td>
            </tr>
            {{$hasBonus = true}}
        {{end}}
        {{/* Individual bonus item row */}}
        <tr>
            <td style="padding-left: 24px;">{{formatProofOfWork .Description}}</td>
            <td>1</td>
            <td>{{formatCurrency .OriginalAmount .OriginalCurrency}}</td>
            <td>{{formatCurrency .OriginalAmount .OriginalCurrency}}</td>
        </tr>
    {{end}}
{{end}}
```

## Alternative Implementation (Recommended)

Since Go html/template has limited programming capabilities, we should pre-process the line items in the controller to create grouped sections.

### Controller Changes (pkg/controller/invoice/contractor_invoice.go)

Add a new struct for template data:

```go
// ContractorInvoiceSection represents a section of line items in the invoice
type ContractorInvoiceSection struct {
    Name           string                        // "Development work from [start] to [end]", "Refund", "Bonus"
    IsAggregated   bool                          // true for Development Work
    Total          float64                       // Total amount for aggregated sections
    Currency       string                        // Currency for aggregated sections
    Items          []ContractorInvoiceLineItem   // Individual items
}

// In GenerateContractorInvoicePDF, prepare template data:
templateData := struct {
    Invoice  *ContractorInvoiceData
    Sections []ContractorInvoiceSection
}{
    Invoice:  data,
    Sections: groupLineItemsIntoSections(data.LineItems),
}
```

Add helper function:

```go
func groupLineItemsIntoSections(items []ContractorInvoiceLineItem) []ContractorInvoiceSection {
    var sections []ContractorInvoiceSection

    // Group Service Fee items
    var serviceFeeItems []ContractorInvoiceLineItem
    for _, item := range items {
        if item.Type == "Service Fee" {
            serviceFeeItems = append(serviceFeeItems, item)
        }
    }
    if len(serviceFeeItems) > 0 {
        // Calculate total (all Service Fee items should be in same currency)
        total := 0.0
        currency := serviceFeeItems[0].OriginalCurrency
        for _, item := range serviceFeeItems {
            total += item.OriginalAmount
        }
        // Format section name with invoice month dates
        startDate := month.Format("Jan 2")
        endDate := time.Date(month.Year(), month.Month()+1, 0, 0, 0, 0, 0, time.UTC).Format("Jan 2")
        sections = append(sections, ContractorInvoiceSection{
            Name:         fmt.Sprintf("Development work from %s to %s", startDate, endDate),
            IsAggregated: true,
            Total:        total,
            Currency:     currency,
            Items:        serviceFeeItems,
        })
    }

    // Group Refund items
    var refundItems []ContractorInvoiceLineItem
    for _, item := range items {
        if item.Type == "Refund" {
            refundItems = append(refundItems, item)
        }
    }
    if len(refundItems) > 0 {
        sections = append(sections, ContractorInvoiceSection{
            Name:         "Refund",
            IsAggregated: false,
            Items:        refundItems,
        })
    }

    // Group Bonus (Commission) items
    var bonusItems []ContractorInvoiceLineItem
    for _, item := range items {
        if item.Type == "Commission" {
            bonusItems = append(bonusItems, item)
        }
    }
    if len(bonusItems) > 0 {
        sections = append(sections, ContractorInvoiceSection{
            Name:         "Bonus",
            IsAggregated: false,
            Items:        bonusItems,
        })
    }

    return sections
}
```

### Simplified Template with Sections

```html
<tbody>
    {{range $section := .Sections}}
        {{if $section.IsAggregated}}
            {{/* Aggregated section (Service Fee) */}}
            <tr>
                <td style="word-wrap: break-word;">
                    <div class="item-project"><strong>{{$section.Name}}</strong></div>
                    {{range $section.Items}}
                    <div class="item-details">{{formatProofOfWork .Description}}</div>
                    {{end}}
                </td>
                <td>1</td>
                <td>{{formatCurrency $section.Total $section.Currency}}</td>
                <td>{{formatCurrency $section.Total $section.Currency}}</td>
            </tr>
        {{else}}
            {{/* Section with individual items (Refund, Bonus) */}}
            <tr>
                <td><strong>{{$section.Name}}</strong></td>
                <td></td>
                <td></td>
                <td></td>
            </tr>
            {{range $section.Items}}
            <tr>
                <td style="word-wrap: break-word; padding-left: 24px;">
                    {{formatProofOfWork .Description}}
                </td>
                <td>1</td>
                <td>{{formatCurrency .OriginalAmount .OriginalCurrency}}</td>
                <td>{{formatCurrency .OriginalAmount .OriginalCurrency}}</td>
            </tr>
            {{end}}
        {{end}}
        {{/* Spacing row after each section */}}
        <tr><td colspan="4" style="padding: 2px; background-color: transparent; border: none;"></td></tr>
    {{end}}
</tbody>
```

## 2. Totals Section Restructure

**Location:** `<tfoot>` section (lines 242-259)

**Current Implementation:**
```html
<tfoot>
    <tr>
        <td colspan="4" style="padding-top:4px;"></td>
    </tr>
    <tr style="line-height: 1">
        <td colspan="2"></td>
        <td style="font-weight: bold; text-align: right; font-size: 14px;">Subtotal</td>
        <td style="font-size: 14px;">$ {{printf "%.2f" .Invoice.TotalUSD}}</td>
    </tr>
    <tr>
        <td colspan="4" style="border-top: 2px solid #333; padding: 5px 0;"></td>
    </tr>
    <tr style="line-height: 1">
        <td colspan="2"></td>
        <td style="font-weight: bold; text-align: right; font-size: 14px;">Total</td>
        <td style="font-weight: bold; font-size: 14px;">$ {{printf "%.2f" .Invoice.TotalUSD}}</td>
    </tr>
</tfoot>
```

**New Implementation:**

```html
<tfoot>
    <tr>
        <td colspan="4" style="padding-top:8px;"></td>
    </tr>

    {{/* VND Subtotal (if any VND items) */}}
    {{if gt .Invoice.SubtotalVND 0}}
    <tr style="line-height: 1.4">
        <td colspan="2"></td>
        <td style="text-align: right; font-size: 12px;">Subtotal</td>
        <td style="font-size: 12px;">{{formatCurrency .Invoice.SubtotalVND "VND"}}</td>
    </tr>
    {{end}}

    {{/* USD Subtotal (conversion + USD items) */}}
    <tr style="line-height: 1.4">
        <td colspan="2"></td>
        <td style="text-align: right; font-size: 12px;">
            {{if gt .Invoice.SubtotalVND 0}}{{else}}Subtotal{{end}}
        </td>
        <td style="font-size: 12px;">{{formatCurrency .Invoice.SubtotalUSD "USD"}}</td>
    </tr>

    {{/* FX Support */}}
    <tr style="line-height: 1.4">
        <td colspan="2"></td>
        <td style="text-align: right; font-size: 12px;">FX support</td>
        <td style="font-size: 12px;">{{formatCurrency .Invoice.FXSupport "USD"}}</td>
    </tr>

    {{/* Separator line */}}
    <tr>
        <td colspan="4" style="border-top: 2px solid #333; padding: 5px 0;"></td>
    </tr>

    {{/* Total */}}
    <tr style="line-height: 1">
        <td colspan="2"></td>
        <td style="font-weight: bold; text-align: right; font-size: 14px;">Total</td>
        <td style="font-weight: bold; font-size: 14px;">{{formatCurrency .Invoice.TotalUSD "USD"}}</td>
    </tr>
</tfoot>
```

## 3. Exchange Rate Footnote

**Location:** After `</table>`, before footer section (after line 260)

**Add New Section:**

```html
        </table>

        {{/* Exchange Rate Footnote */}}
        {{if gt .Invoice.ExchangeRate 1}}
        <div class="exchange-rate-note" style="margin-top: 10px; margin-bottom: 20px;">
            <p style="font-size: 11px; color: #666; font-style: italic;">
                *FX Rate {{formatExchangeRate .Invoice.ExchangeRate}}
            </p>
        </div>
        {{end}}

        <!-- Footer Section -->
        <div class="footer">
```

## 4. CSS Style Updates

**Location:** `<style>` section (lines 7-176)

**Add New Styles:**

```css
/* Section header rows */
.invoice-table tbody tr.section-header td {
    font-weight: bold;
    font-size: 12px;
    background-color: transparent;
    border-bottom: none;
    padding-top: 8px;
}

/* Indented section items */
.invoice-table tbody td.section-item {
    padding-left: 24px;
}

/* Exchange rate footnote */
.exchange-rate-note {
    margin-top: 10px;
    margin-bottom: 20px;
}

.exchange-rate-note p {
    font-size: 11px;
    color: #666;
    font-style: italic;
    margin: 0;
}

/* Totals section styling */
.invoice-table tfoot tr {
    line-height: 1.4;
}

.invoice-table tfoot td {
    padding: 3px 5px;
    font-size: 12px;
}

.invoice-table tfoot tr:last-child td {
    font-weight: bold;
    font-size: 14px;
    padding-top: 8px;
}
```

## Complete Updated Template Body

**Full `<tbody>` Section:**

```html
<tbody>
    {{range $section := .Sections}}
        {{if $section.IsAggregated}}
            {{/* Aggregated section (Development Work) */}}
            <tr>
                <td style="word-wrap: break-word;">
                    <div class="item-project"><strong>{{$section.Name}}</strong></div>
                    {{range $section.Items}}
                    <div class="item-details">{{formatProofOfWork .Description}}</div>
                    {{end}}
                </td>
                <td>1</td>
                <td>{{formatCurrency $section.Total $section.Currency}}</td>
                <td>{{formatCurrency $section.Total $section.Currency}}</td>
            </tr>
        {{else}}
            {{/* Section with header and individual items (Refund, Bonus) */}}
            <tr class="section-header">
                <td><strong>{{$section.Name}}</strong></td>
                <td></td>
                <td></td>
                <td></td>
            </tr>
            {{range $section.Items}}
            <tr>
                <td class="section-item" style="word-wrap: break-word;">
                    {{formatProofOfWork .Description}}
                </td>
                <td>1</td>
                <td>{{formatCurrency .OriginalAmount .OriginalCurrency}}</td>
                <td>{{formatCurrency .OriginalAmount .OriginalCurrency}}</td>
            </tr>
            {{end}}
        {{end}}
        {{/* Spacing row after each section */}}
        <tr><td colspan="4" style="padding: 4px; background-color: transparent; border: none;"></td></tr>
    {{end}}
</tbody>
```

## Testing Scenarios

### Visual Testing Required

1. **Service Fee Only:** One aggregated row with descriptions
2. **Refund Only:** Section header + individual rows
3. **Bonus Only:** Section header + individual rows
4. **All Sections:** Service Fee + Refund + Bonus in correct order
5. **Mixed Currencies:** VND Service Fee + USD Bonus
6. **VND Only:** Display VND subtotal and USD conversion
7. **USD Only:** Display only USD subtotal
8. **Exchange Rate:** Footnote appears when ExchangeRate > 1

### Edge Cases

- Empty sections (no items of a type)
- Single item in section
- Long descriptions with line breaks
- Large amounts (formatting with separators)
- Zero amounts

## Related Documents

- spec-001-data-structure-changes.md: Data structure for sections
- spec-002-template-functions.md: Template functions used
- ADR-001: Data structure decisions
- Requirements: Expected PDF format
