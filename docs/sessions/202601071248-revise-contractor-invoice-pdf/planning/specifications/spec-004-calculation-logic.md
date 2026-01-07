# Specification: Calculation Logic

**Version:** 1.0
**Date:** 2026-01-07
**Status:** Ready for Implementation

## Overview

This specification details the exact calculation logic for multi-currency subtotals, conversions, and totals in contractor invoices.

## Calculation Flow

```
Line Items (Original Currency)
    ↓
Group by Currency
    ↓
Calculate Subtotals
    ├── Subtotal VND (sum of VND items)
    └── Subtotal USD Items (sum of USD items)
    ↓
Convert VND to USD
    └── Subtotal USD from VND (via Wise API)
    ↓
Calculate Combined Subtotal
    └── Subtotal USD = Subtotal USD from VND + Subtotal USD Items
    ↓
Add FX Support
    └── Total USD = Subtotal USD + FX Support ($8)
```

## Implementation Location

**File:** `pkg/controller/invoice/contractor_invoice.go`

**Function:** `GenerateContractorInvoice`

**Insert After:** Line item grouping and sorting (around line 310)

## Step-by-Step Algorithm

### Step 1: Sum Line Items by Currency

**Purpose:** Calculate the total amount for VND items and USD items separately.

**Implementation:**

```go
// Calculate subtotals for multi-currency display
var subtotalVND float64       // Sum of all VND-denominated items
var subtotalUSDItems float64  // Sum of all USD-denominated items
var vndItemCount int
var usdItemCount int

l.Debug("calculating subtotals by currency")

for _, item := range lineItems {
    // Validate currency
    if item.OriginalCurrency != "VND" && item.OriginalCurrency != "USD" {
        l.Error(nil, fmt.Sprintf("invalid currency: %s for item: %s", item.OriginalCurrency, item.Description))
        return nil, fmt.Errorf("invalid currency: %s (must be VND or USD)", item.OriginalCurrency)
    }

    // Sum by currency
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

l.Debug(fmt.Sprintf("subtotals calculated - VND: %.0f (%d items), USD: %.2f (%d items)",
    subtotalVND, vndItemCount, subtotalUSDItems, usdItemCount))
```

**Validation Rules:**
- Currency must be "VND" or "USD" (case-sensitive)
- OriginalAmount must be non-negative
- If invalid currency, return error immediately

### Step 2: Round Subtotals

**Purpose:** Round to appropriate decimal places for each currency.

**Implementation:**

```go
// Round subtotals to appropriate decimal places
subtotalVND = math.Round(subtotalVND) // VND has no minor units (0 decimals)
subtotalUSDItems = math.Round(subtotalUSDItems*100) / 100 // USD has 2 decimals

l.Debug(fmt.Sprintf("rounded subtotals - VND: %.0f, USD: %.2f", subtotalVND, subtotalUSDItems))
```

**Rounding Rules:**
- **VND:** Round to nearest whole number (no decimal places)
- **USD:** Round to 2 decimal places using banker's rounding

### Step 3: Convert VND Subtotal to USD

**Purpose:** Use Wise API to convert VND subtotal to USD and capture exchange rate.

**Implementation:**

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
    // No VND items, exchange rate is 1 (or could be 0)
    exchangeRate = 1.0
    subtotalUSDFromVND = 0.0

    l.Debug("no VND items, skipping conversion")
}
```

**Validation Rules:**
- If subtotalVND = 0, skip conversion (set exchangeRate = 1.0)
- If Wise API fails, return error (don't use fallback rate)
- Exchange rate must be > 0
- Round converted USD amount to 2 decimals

**Error Handling:**
- API timeout → return error with clear message
- API returns invalid rate → return error
- Network failure → return error (don't generate invoice without accurate rate)

### Step 4: Calculate Combined USD Subtotal

**Purpose:** Sum the USD amounts from VND conversion and direct USD items.

**Implementation:**

```go
// Calculate combined USD subtotal
subtotalUSD := subtotalUSDFromVND + subtotalUSDItems
subtotalUSD = math.Round(subtotalUSD*100) / 100 // Round to 2 decimals

l.Debug(fmt.Sprintf("combined USD subtotal: %.2f (%.2f from VND + %.2f direct USD)",
    subtotalUSD, subtotalUSDFromVND, subtotalUSDItems))
```

**Formula:**
```
Subtotal USD = Subtotal USD from VND + Subtotal USD Items
```

**Rounding:** Always round result to 2 decimal places.

### Step 5: Add FX Support Fee

**Purpose:** Add the FX support fee to calculate final total.

**Implementation:**

```go
// Add FX support fee (hardcoded for now)
fxSupport := 8.0 // TODO: Implement dynamic calculation based on business rules

l.Debug(fmt.Sprintf("FX support fee: %.2f", fxSupport))

// Calculate final total
totalUSD := subtotalUSD + fxSupport
totalUSD = math.Round(totalUSD*100) / 100 // Round to 2 decimals

l.Debug(fmt.Sprintf("final total USD: %.2f (%.2f subtotal + %.2f FX support)",
    totalUSD, subtotalUSD, fxSupport))
```

**Formula:**
```
Total USD = Subtotal USD + FX Support
```

**Current Business Rule:**
- FX Support is hardcoded at $8.00
- TODO: Future implementation will calculate based on transaction volume/complexity

### Step 6: Populate InvoiceData

**Purpose:** Store all calculated values in the invoice data structure.

**Implementation:**

```go
invoiceData := &ContractorInvoiceData{
    // ... existing fields ...

    // Currency and exchange rate
    Currency:          "USD", // Invoice currency for payment is always USD
    ExchangeRate:      exchangeRate,

    // Totals (legacy fields)
    Total:             totalUSD,
    TotalUSD:          totalUSD,

    // NEW: Multi-currency subtotals
    SubtotalVND:        subtotalVND,
    SubtotalUSDFromVND: subtotalUSDFromVND,
    SubtotalUSDItems:   subtotalUSDItems,
    SubtotalUSD:        subtotalUSD,
    FXSupport:          fxSupport,
}

l.Debug("invoice data populated with calculated totals")
```

## Calculation Examples

### Example 1: All VND Items

**Input:**
- Service Fee: 45,000,000 VND
- Refund: 500,000 VND
- Exchange Rate: 26,269 VND/USD

**Calculation:**
```
Subtotal VND = 45,000,000 + 500,000 = 45,500,000 VND
Subtotal USD from VND = 45,500,000 / 26,269 = 1,731.89 USD
Subtotal USD Items = 0 USD
Subtotal USD = 1,731.89 + 0 = 1,731.89 USD
FX Support = 8.00 USD
Total USD = 1,731.89 + 8.00 = 1,739.89 USD
```

**Output:**
```
SubtotalVND: 45,500,000
SubtotalUSDFromVND: 1,731.89
SubtotalUSDItems: 0.00
SubtotalUSD: 1,731.89
FXSupport: 8.00
TotalUSD: 1,739.89
ExchangeRate: 26,269
```

### Example 2: All USD Items

**Input:**
- Service Fee: $1,500.00
- Bonus: $100.00

**Calculation:**
```
Subtotal VND = 0
Subtotal USD from VND = 0
Subtotal USD Items = 1,500.00 + 100.00 = 1,600.00 USD
Subtotal USD = 0 + 1,600.00 = 1,600.00 USD
FX Support = 8.00 USD
Total USD = 1,600.00 + 8.00 = 1,608.00 USD
```

**Output:**
```
SubtotalVND: 0
SubtotalUSDFromVND: 0.00
SubtotalUSDItems: 1,600.00
SubtotalUSD: 1,600.00
FXSupport: 8.00
TotalUSD: 1,608.00
ExchangeRate: 1.0
```

### Example 3: Mixed Currencies

**Input:**
- Service Fee: 45,000,000 VND
- Refund: 500,000 VND
- Bonus: $100.00
- Exchange Rate: 26,269 VND/USD

**Calculation:**
```
Subtotal VND = 45,000,000 + 500,000 = 45,500,000 VND
Subtotal USD from VND = 45,500,000 / 26,269 = 1,731.89 USD
Subtotal USD Items = 100.00 USD
Subtotal USD = 1,731.89 + 100.00 = 1,831.89 USD
FX Support = 8.00 USD
Total USD = 1,831.89 + 8.00 = 1,839.89 USD
```

**Output:**
```
SubtotalVND: 45,500,000
SubtotalUSDFromVND: 1,731.89
SubtotalUSDItems: 100.00
SubtotalUSD: 1,831.89
FXSupport: 8.00
TotalUSD: 1,839.89
ExchangeRate: 26,269
```

## Edge Cases

### Edge Case 1: Zero Amounts

**Input:** All line items have 0 amount

**Handling:**
```
Subtotal VND = 0
Subtotal USD = 0
FX Support = 8.00
Total USD = 8.00
```

**Note:** This is a valid scenario (e.g., adjustment invoice).

### Edge Case 2: Negative Amounts

**Input:** Line item has negative OriginalAmount

**Handling:**
- Validation should reject negative amounts
- Return error: "negative amount for line item"

**Rationale:** Negative amounts should be represented as separate Refund items.

### Edge Case 3: Very Small Amounts

**Input:** USD item with $0.001

**Handling:**
```
Amount = 0.001
Rounded = 0.00 (rounds to 2 decimals)
```

**Note:** Amounts < $0.01 will round to $0.00.

### Edge Case 4: Very Large Amounts

**Input:** VND item with 1,000,000,000 (1 billion VND)

**Handling:**
```
Amount = 1,000,000,000 VND
Conversion = 38,063.61 USD (at rate 26,269)
```

**Note:** Ensure number formatting handles billions correctly.

### Edge Case 5: Exchange Rate API Failure

**Input:** Wise API returns error

**Handling:**
- Do NOT use fallback rate
- Return error to caller
- Log detailed error message
- Invoice generation fails

**Rationale:** Using incorrect exchange rate could lead to payment discrepancies.

## Validation Checklist

Before populating invoiceData, validate:

- [ ] All OriginalCurrency values are "VND" or "USD"
- [ ] All OriginalAmount values are >= 0
- [ ] All AmountUSD values are >= 0
- [ ] ExchangeRate > 0 (if VND items exist)
- [ ] SubtotalVND rounded to 0 decimals
- [ ] All USD amounts rounded to 2 decimals
- [ ] Total USD = Subtotal USD + FX Support

## Testing Requirements

### Unit Tests

**Test File:** `pkg/controller/invoice/contractor_invoice_calculations_test.go` (new)

```go
func TestCalculateSubtotals(t *testing.T) {
    tests := []struct {
        name                   string
        lineItems              []ContractorInvoiceLineItem
        exchangeRate           float64
        expectedSubtotalVND    float64
        expectedSubtotalUSD    float64
        expectedFXSupport      float64
        expectedTotal          float64
    }{
        {
            name: "All VND items",
            lineItems: []ContractorInvoiceLineItem{
                {OriginalAmount: 45000000, OriginalCurrency: "VND", AmountUSD: 1713.41},
                {OriginalAmount: 500000, OriginalCurrency: "VND", AmountUSD: 19.03},
            },
            exchangeRate:        26269,
            expectedSubtotalVND: 45500000,
            expectedSubtotalUSD: 1732.44,
            expectedFXSupport:   8.00,
            expectedTotal:       1740.44,
        },
        {
            name: "All USD items",
            lineItems: []ContractorInvoiceLineItem{
                {OriginalAmount: 1500, OriginalCurrency: "USD", AmountUSD: 1500},
                {OriginalAmount: 100, OriginalCurrency: "USD", AmountUSD: 100},
            },
            exchangeRate:        1.0,
            expectedSubtotalVND: 0,
            expectedSubtotalUSD: 1600.00,
            expectedFXSupport:   8.00,
            expectedTotal:       1608.00,
        },
        {
            name: "Mixed currencies",
            lineItems: []ContractorInvoiceLineItem{
                {OriginalAmount: 45000000, OriginalCurrency: "VND", AmountUSD: 1713.41},
                {OriginalAmount: 100, OriginalCurrency: "USD", AmountUSD: 100},
            },
            exchangeRate:        26269,
            expectedSubtotalVND: 45000000,
            expectedSubtotalUSD: 1813.41,
            expectedFXSupport:   8.00,
            expectedTotal:       1821.41,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Integration Tests

Test with real Wise API:
- Verify exchange rate is current
- Test API error handling
- Test timeout scenarios

## Related Documents

- ADR-001: Data Structure for Multi-Currency Support
- spec-001-data-structure-changes.md: Data structure definitions
- spec-002-template-functions.md: Formatting functions
- Requirements: Currency display rules
